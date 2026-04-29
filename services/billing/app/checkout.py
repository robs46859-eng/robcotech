"""
Stripe Checkout API

Creates checkout sessions for subscription purchases.
"""

from fastapi import APIRouter, HTTPException, Request, Header
from pydantic import BaseModel
from typing import Optional, Dict, Any
import logging

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/checkout", tags=["checkout"])


class CheckoutRequest(BaseModel):
    """Request to create a checkout session"""
    price_id: str
    customer_email: str
    customer_name: Optional[str] = None
    success_url: str
    cancel_url: str
    metadata: Optional[Dict[str, str]] = None


class CheckoutResponse(BaseModel):
    """Checkout session response"""
    session_id: str
    url: str
    customer_id: str


@router.post("/create", response_model=CheckoutResponse)
async def create_checkout_session(
    request: CheckoutRequest,
    stripe_api_key: str = Header(..., description="Stripe API Key"),
):
    """
    Create a Stripe Checkout session for subscription purchase.
    
    This redirects the customer to Stripe's hosted checkout page.
    """
    import stripe
    stripe.api_key = stripe_api_key
    
    try:
        # Create customer if email provided
        customer = stripe.Customer.create(
            email=request.customer_email,
            name=request.customer_name or "",
            metadata=request.metadata or {},
        )
        
        # Create checkout session
        session = stripe.checkout.Session.create(
            customer=customer.id,
            mode="subscription",
            payment_method_types=["card"],
            line_items=[
                {
                    "price": request.price_id,
                    "quantity": 1,
                }
            ],
            success_url=request.success_url,
            cancel_url=request.cancel_url,
            allow_promotion_codes=True,
            automatic_tax={"enabled": True},
            metadata=request.metadata or {},
        )
        
        logger.info(f"Created checkout session: {session.id} for customer {customer.id}")
        
        return CheckoutResponse(
            session_id=session.id,
            url=session.url,
            customer_id=customer.id,
        )
        
    except stripe.error.StripeError as e:
        logger.error(f"Stripe error: {e}")
        raise HTTPException(status_code=400, detail=f"Stripe error: {str(e)}")
    except Exception as e:
        logger.error(f"Failed to create checkout session: {e}")
        raise HTTPException(status_code=500, detail="Failed to create checkout session")


@router.get("/session/{session_id}")
async def get_checkout_session(
    session_id: str,
    stripe_api_key: str = Header(..., description="Stripe API Key"),
):
    """Get checkout session status"""
    import stripe
    stripe.api_key = stripe_api_key
    
    try:
        session = stripe.checkout.Session.retrieve(session_id)
        
        return {
            "id": session.id,
            "status": session.status,
            "customer_email": session.customer_details.email if session.customer_details else None,
            "amount_total": session.amount_total,
            "currency": session.currency,
        }
        
    except Exception as e:
        logger.error(f"Failed to get session: {e}")
        raise HTTPException(status_code=404, detail="Session not found")
