"""
Billing Service

Usage metering and Stripe integration for FullStackArkham.

Responsibilities:
- Track token usage per tenant
- Calculate costs based on model and provider
- Manage subscriptions and plans
- Process Stripe payments
- Generate invoices
- Enforce quotas and limits
"""

import logging
from contextlib import asynccontextmanager
from typing import Dict, Any, List, Optional
from datetime import datetime, date
from decimal import Decimal

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from app.settings import settings
from app.metering import UsageMeter
from app.plans import PlanManager
from app.stripe import StripeService
from app.checkout import router as checkout_router

logger = logging.getLogger(__name__)


class UsageRecord(BaseModel):
    """Record usage for billing"""
    tenant_id: str
    request_id: str
    model: str
    provider: str
    input_tokens: int
    output_tokens: int
    total_tokens: int
    task_type: Optional[str] = None
    cache_hit: bool = False


class UsageResponse(BaseModel):
    """Usage query response"""
    tenant_id: str
    period_start: str
    period_end: str
    tokens_used: int
    requests_count: int
    cache_hits: int
    total_cost: float
    currency: str
    by_model: Dict[str, int]
    by_provider: Dict[str, int]


class BillingRecord(BaseModel):
    """Billing record for a period"""
    tenant_id: str
    period_start: str
    period_end: str
    tokens_used: int
    requests_count: int
    cache_hits: int
    total_cost: float
    currency: str
    status: str
    stripe_invoice_id: Optional[str] = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan handler"""
    # Startup
    logger.info("Starting Billing Service")
    
    app.state.usage_meter = UsageMeter(settings.database_url)
    app.state.plan_manager = PlanManager(settings.database_url)
    
    if settings.stripe_secret_key:
        app.state.stripe_service = StripeService(
            secret_key=settings.stripe_secret_key,
            webhook_secret=settings.stripe_webhook_secret,
        )
    else:
        logger.warning("Stripe not configured - billing will be local only")
        app.state.stripe_service = None
    
    await app.state.usage_meter.initialize()
    await app.state.plan_manager.initialize()
    
    yield
    
    # Shutdown
    logger.info("Shutting down Billing Service")
    await app.state.usage_meter.close()
    await app.state.plan_manager.close()


app = FastAPI(
    title="Billing Service",
    description="Usage metering and Stripe integration",
    version="0.1.0",
    lifespan=lifespan,
)

# Include checkout router
app.include_router(checkout_router)


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {"status": "healthy", "service": "billing"}


@app.get("/ready")
async def ready_check():
    """Readiness check endpoint"""
    if not hasattr(app.state, "usage_meter"):
        return {"status": "not ready", "reason": "usage meter not initialized"}
    return {"status": "ready"}


@app.post("/api/v1/usage/record")
async def record_usage(usage_data: UsageRecord):
    """
    Record usage for billing
    
    Called by the gateway after each inference request.
    Tracks tokens, calculates cost, updates tenant quota.
    """
    usage_meter = app.state.usage_meter
    plan_manager = app.state.plan_manager
    
    # Calculate cost based on model and provider
    cost = calculate_cost(
        model=usage_data.model,
        provider=usage_data.provider,
        input_tokens=usage_data.input_tokens,
        output_tokens=usage_data.output_tokens,
        cache_hit=usage_data.cache_hit,
    )
    
    # Record usage
    await usage_meter.record(
        tenant_id=usage_data.tenant_id,
        request_id=usage_data.request_id,
        model=usage_data.model,
        provider=usage_data.provider,
        input_tokens=usage_data.input_tokens,
        output_tokens=usage_data.output_tokens,
        total_tokens=usage_data.total_tokens,
        cost_usd=cost,
        task_type=usage_data.task_type,
        cache_hit=usage_data.cache_hit,
    )
    
    # Update tenant quota
    tenant = await plan_manager.get_tenant(usage_data.tenant_id)
    if tenant:
        new_usage = tenant["quota_used"] + usage_data.total_tokens
        await plan_manager.update_quota_usage(usage_data.tenant_id, new_usage)
        
        # Check if quota exceeded
        if new_usage > tenant["quota_monthly"]:
            logger.warning(
                f"Tenant {usage_data.tenant_id} exceeded quota: "
                f"{new_usage}/{tenant['quota_monthly']}"
            )
    
    return {
        "status": "recorded",
        "tokens": usage_data.total_tokens,
        "cost_usd": cost,
    }


@app.get("/api/v1/usage/{tenant_id}", response_model=UsageResponse)
async def get_usage(
    tenant_id: str,
    period_start: Optional[str] = None,
    period_end: Optional[str] = None,
):
    """Get usage for a tenant"""
    usage_meter = app.state.usage_meter
    
    # Default to current billing period
    if not period_start:
        period_start = date.today().replace(day=1).isoformat()
    if not period_end:
        period_end = date.today().isoformat()
    
    usage = await usage_meter.get_usage(
        tenant_id=tenant_id,
        period_start=period_start,
        period_end=period_end,
    )
    
    return UsageResponse(
        tenant_id=tenant_id,
        period_start=period_start,
        period_end=period_end,
        tokens_used=usage.get("total_tokens", 0),
        requests_count=usage.get("requests_count", 0),
        cache_hits=usage.get("cache_hits", 0),
        total_cost=float(usage.get("total_cost", Decimal("0"))),
        currency="USD",
        by_model=usage.get("by_model", {}),
        by_provider=usage.get("by_provider", {}),
    )


@app.get("/api/v1/billing/{tenant_id}", response_model=BillingRecord)
async def get_billing(
    tenant_id: str,
    period: Optional[str] = None,
):
    """Get billing record for a tenant"""
    usage_meter = app.state.usage_meter
    
    # Default to current period
    if not period:
        period_start = date.today().replace(day=1)
        period_end = date.today()
    else:
        # Parse period (YYYY-MM format)
        period_start = datetime.strptime(f"{period}-01", "%Y-%m-%d").date()
        if period_start.month == 12:
            period_end = period_start.replace(year=period_start.year + 1, month=1, day=1)
        else:
            period_end = period_start.replace(month=period_start.month + 1, day=1)
    
    usage = await usage_meter.get_usage(
        tenant_id=tenant_id,
        period_start=period_start.isoformat(),
        period_end=period_end.isoformat(),
    )
    
    # Get or create billing record
    billing = await usage_meter.get_billing_record(tenant_id, period_start)
    
    return BillingRecord(
        tenant_id=tenant_id,
        period_start=period_start.isoformat(),
        period_end=period_end.isoformat(),
        tokens_used=usage.get("total_tokens", 0),
        requests_count=usage.get("requests_count", 0),
        cache_hits=usage.get("cache_hits", 0),
        total_cost=float(usage.get("total_cost", Decimal("0"))),
        currency="USD",
        status=billing.get("status", "pending") if billing else "pending",
        stripe_invoice_id=billing.get("stripe_invoice_id") if billing else None,
    )


@app.post("/api/v1/billing/{tenant_id}/invoice")
async def create_invoice(tenant_id: str, period: Optional[str] = None):
    """
    Create an invoice for a tenant
    
    If Stripe is configured, creates invoice in Stripe.
    Otherwise, creates local billing record.
    """
    stripe_service = app.state.stripe_service
    usage_meter = app.state.usage_meter
    plan_manager = app.state.plan_manager
    
    # Get tenant
    tenant = await plan_manager.get_tenant(tenant_id)
    if not tenant:
        raise HTTPException(status_code=404, detail="Tenant not found")
    
    # Get usage for period
    if not period:
        period_start = date.today().replace(day=1)
    else:
        period_start = datetime.strptime(f"{period}-01", "%Y-%m-%d").date()
    
    period_end = date.today()
    
    usage = await usage_meter.get_usage(
        tenant_id=tenant_id,
        period_start=period_start.isoformat(),
        period_end=period_end.isoformat(),
    )
    
    total_cost = usage.get("total_cost", Decimal("0"))
    
    if stripe_service and tenant.get("stripe_customer_id"):
        # Create Stripe invoice
        invoice = await stripe_service.create_invoice(
            customer_id=tenant["stripe_customer_id"],
            amount=float(total_cost),
            description=f"FullStackArkham usage {period_start} to {period_end}",
            metadata={
                "tenant_id": tenant_id,
                "period_start": period_start.isoformat(),
                "period_end": period_end.isoformat(),
                "tokens_used": usage.get("total_tokens", 0),
            },
        )
        
        # Update billing record
        await usage_meter.create_billing_record(
            tenant_id=tenant_id,
            period_start=period_start,
            period_end=period_end,
            tokens_used=usage.get("total_tokens", 0),
            total_cost=total_cost,
            stripe_invoice_id=invoice["id"],
        )
        
        return {"status": "created", "stripe_invoice_id": invoice["id"]}
    
    else:
        # Create local billing record
        await usage_meter.create_billing_record(
            tenant_id=tenant_id,
            period_start=period_start,
            period_end=period_end,
            tokens_used=usage.get("total_tokens", 0),
            total_cost=total_cost,
        )
        
        return {"status": "created", "local": True}


@app.get("/api/v1/plans")
async def list_plans():
    """List available subscription plans"""
    plan_manager = app.state.plan_manager
    plans = await plan_manager.list_plans()
    
    return {"plans": plans}


@app.get("/api/v1/plans/{plan_id}")
async def get_plan(plan_id: str):
    """Get plan details"""
    plan_manager = app.state.plan_manager
    plan = await plan_manager.get_plan(plan_id)
    
    if not plan:
        raise HTTPException(status_code=404, detail="Plan not found")
    
    return plan


@app.post("/api/v1/tenants/{tenant_id}/subscription")
async def update_subscription(
    tenant_id: str,
    plan_id: str,
    stripe_customer_id: Optional[str] = None,
):
    """Update tenant subscription plan"""
    plan_manager = app.state.plan_manager
    stripe_service = app.state.stripe_service
    
    plan = await plan_manager.get_plan(plan_id)
    if not plan:
        raise HTTPException(status_code=404, detail="Plan not found")
    
    if stripe_service and stripe_customer_id:
        # Create Stripe subscription
        subscription = await stripe_service.create_subscription(
            customer_id=stripe_customer_id,
            price_id=plan.get("stripe_price_id"),
        )
        
        await plan_manager.update_tenant_plan(
            tenant_id=tenant_id,
            plan_id=plan_id,
            stripe_subscription_id=subscription["id"],
        )
    else:
        # Update local plan
        await plan_manager.update_tenant_plan(
            tenant_id=tenant_id,
            plan_id=plan_id,
        )
    
    return {"status": "updated", "plan_id": plan_id}


@app.post("/api/v1/stripe/webhook")
async def stripe_webhook(request: Any):
    """Handle Stripe webhooks"""
    stripe_service = app.state.stripe_service
    
    if not stripe_service:
        raise HTTPException(status_code=503, detail="Stripe not configured")
    
    # Verify webhook signature
    # payload = await request.body()
    # sig_header = request.headers.get("stripe-signature")
    # event = stripe_service.verify_webhook(payload, sig_header)
    
    # Process webhook event
    # - invoice.payment_succeeded: Update billing record
    # - invoice.payment_failed: Notify tenant, suspend service
    # - customer.subscription.updated: Update tenant plan
    
    return {"status": "received"}


def calculate_cost(
    model: str,
    provider: str,
    input_tokens: int,
    output_tokens: int,
    cache_hit: bool = False,
) -> Decimal:
    """
    Calculate cost for an inference request
    
    Uses cost ladder pricing:
    - Cache hits: $0 (already paid)
    - Local models: ~$0.0001/1K tokens
    - Cheap API: ~$0.0005/1K tokens
    - Mid-tier: ~$0.003/1K tokens
    - Premium: ~$0.03/1K tokens
    """
    if cache_hit:
        return Decimal("0")
    
    # Pricing per 1K tokens (input + output weighted)
    pricing = {
        # Local models (running on our infrastructure)
        "local": {"input": 0.0001, "output": 0.0001},
        
        # Cheap API models
        "haiku": {"input": 0.00025, "output": 0.00125},
        "gpt-3.5": {"input": 0.0005, "output": 0.0015},
        
        # Mid-tier models
        "sonnet": {"input": 0.003, "output": 0.015},
        "gpt-4": {"input": 0.01, "output": 0.03},
        
        # Premium models
        "opus": {"input": 0.015, "output": 0.075},
        "gpt-4-turbo": {"input": 0.01, "output": 0.03},
    }
    
    # Get pricing for model
    model_lower = model.lower()
    rates = pricing.get("local")  # Default to local
    
    for key, rate in pricing.items():
        if key in model_lower:
            rates = rate
            break
    
    # Calculate cost
    input_cost = Decimal(str(rates["input"])) * Decimal(input_tokens) / Decimal(1000)
    output_cost = Decimal(str(rates["output"])) * Decimal(output_tokens) / Decimal(1000)
    
    return input_cost + output_cost


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "app.main:app",
        host="0.0.0.0",
        port=8086,
        reload=True,
    )
