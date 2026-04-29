"""
Plan Manager

Manages subscription plans and tenant quotas.
"""

import logging
from typing import Dict, Any, List, Optional

import asyncpg

logger = logging.getLogger(__name__)


class PlanManager:
    """Manages subscription plans and tenant quotas"""
    
    def __init__(self, database_url: str):
        self.database_url = database_url
        self.pool: Optional[asyncpg.Pool] = None
    
    async def initialize(self):
        """Initialize database connection pool"""
        self.pool = await asyncpg.create_pool(self.database_url)
        
        # Ensure default plans exist
        await self._ensure_default_plans()
        
        logger.info("Plan manager initialized")
    
    async def close(self):
        """Close database connections"""
        if self.pool:
            await self.pool.close()
    
    async def _ensure_default_plans(self):
        """Create default plans if they don't exist"""
        default_plans = [
            {
                "id": "free",
                "name": "Free",
                "price_monthly": 0,
                "quota_monthly": 10000,
                "features": ["basic_inference", "semantic_cache"],
            },
            {
                "id": "basic",
                "name": "Basic",
                "price_monthly": 29,
                "quota_monthly": 100000,
                "features": ["basic_inference", "semantic_cache", "memory", "orchestration"],
            },
            {
                "id": "pro",
                "name": "Pro",
                "price_monthly": 99,
                "quota_monthly": 1000000,
                "features": ["all_models", "semantic_cache", "memory", "orchestration", "priority_support"],
            },
            {
                "id": "enterprise",
                "name": "Enterprise",
                "price_monthly": 499,
                "quota_monthly": 10000000,
                "features": ["all_models", "unlimited_cache", "custom_memory", "dedicated_support", "sla"],
            },
        ]
        
        async with self.pool.acquire() as conn:
            for plan in default_plans:
                await conn.execute("""
                    INSERT INTO subscription_plans (id, name, price_monthly, quota_monthly, features)
                    VALUES ($1, $2, $3, $4, $5)
                    ON CONFLICT (id) DO NOTHING
                """,
                    plan["id"],
                    plan["name"],
                    plan["price_monthly"],
                    plan["quota_monthly"],
                    plan["features"],
                )
    
    async def list_plans(self) -> List[Dict[str, Any]]:
        """List all available plans"""
        async with self.pool.acquire() as conn:
            rows = await conn.fetch("SELECT * FROM subscription_plans ORDER BY price_monthly")
            
            return [
                {
                    "id": row["id"],
                    "name": row["name"],
                    "price_monthly": float(row["price_monthly"]),
                    "quota_monthly": row["quota_monthly"],
                    "features": list(row["features"]) if row["features"] else [],
                }
                for row in rows
            ]
    
    async def get_plan(self, plan_id: str) -> Optional[Dict[str, Any]]:
        """Get plan by ID"""
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow(
                "SELECT * FROM subscription_plans WHERE id = $1", plan_id
            )
            
            if not row:
                return None
            
            return {
                "id": row["id"],
                "name": row["name"],
                "price_monthly": float(row["price_monthly"]),
                "quota_monthly": row["quota_monthly"],
                "features": list(row["features"]) if row["features"] else [],
            }
    
    async def get_tenant(self, tenant_id: str) -> Optional[Dict[str, Any]]:
        """Get tenant with plan info"""
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow("""
                SELECT t.*, sp.name as plan_name, sp.quota_monthly
                FROM tenants t
                LEFT JOIN subscription_plans sp ON t.plan = sp.id
                WHERE t.id = $1
            """, tenant_id)
            
            if not row:
                return None
            
            return {
                "id": str(row["id"]),
                "name": row["name"],
                "plan": row["plan"],
                "plan_name": row["plan_name"],
                "quota_monthly": row["quota_monthly"] or 10000,
                "quota_used": row["quota_used"],
                "stripe_customer_id": row.get("stripe_customer_id"),
            }
    
    async def update_tenant_plan(
        self,
        tenant_id: str,
        plan_id: str,
        stripe_subscription_id: Optional[str] = None,
    ):
        """Update tenant's subscription plan"""
        async with self.pool.acquire() as conn:
            await conn.execute("""
                UPDATE tenants SET
                    plan = $2,
                    stripe_subscription_id = $3,
                    updated_at = NOW()
                WHERE id = $1
            """, tenant_id, plan_id, stripe_subscription_id)
    
    async def update_quota_usage(
        self,
        tenant_id: str,
        quota_used: int,
    ):
        """Update tenant's quota usage"""
        async with self.pool.acquire() as conn:
            await conn.execute("""
                UPDATE tenants SET
                    quota_used = $2,
                    updated_at = NOW()
                WHERE id = $1
            """, tenant_id, quota_used)
    
    async def check_quota(
        self,
        tenant_id: str,
    ) -> tuple[bool, int]:
        """
        Check if tenant has remaining quota
        
        Returns (has_quota, remaining_tokens)
        """
        tenant = await self.get_tenant(tenant_id)
        
        if not tenant:
            return False, 0
        
        remaining = tenant["quota_monthly"] - tenant["quota_used"]
        return remaining > 0, max(0, remaining)
