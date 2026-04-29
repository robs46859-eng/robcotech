"""
Usage Meter

Tracks and aggregates usage data for billing.
"""

import logging
from typing import Dict, Any, Optional
from datetime import date, datetime
from decimal import Decimal

import asyncpg
from orjson import dumps

logger = logging.getLogger(__name__)


class UsageMeter:
    """PostgreSQL-backed usage metering"""
    
    def __init__(self, database_url: str):
        self.database_url = database_url
        self.pool: Optional[asyncpg.Pool] = None
    
    async def initialize(self):
        """Initialize database connection pool"""
        self.pool = await asyncpg.create_pool(
            self.database_url,
            min_size=5,
            max_size=20,
        )
        logger.info("Usage meter initialized")
    
    async def close(self):
        """Close database connections"""
        if self.pool:
            await self.pool.close()
    
    async def record(
        self,
        tenant_id: str,
        request_id: str,
        model: str,
        provider: str,
        input_tokens: int,
        output_tokens: int,
        total_tokens: int,
        cost_usd: Decimal,
        task_type: Optional[str] = None,
        cache_hit: bool = False,
    ):
        """Record usage for a request"""
        async with self.pool.acquire() as conn:
            await conn.execute("""
                INSERT INTO inference_logs (
                    id, tenant_id, request_id, model, provider,
                    input_tokens, output_tokens, total_tokens,
                    cost_usd, cache_hit, task_type, status,
                    latency_ms, created_at
                ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'completed', 0, NOW())
            """,
                str(__import__("uuid").uuid4()),
                tenant_id,
                request_id,
                model,
                provider,
                input_tokens,
                output_tokens,
                total_tokens,
                cost_usd,
                cache_hit,
                task_type,
            )
    
    async def get_usage(
        self,
        tenant_id: str,
        period_start: str,
        period_end: str,
    ) -> Dict[str, Any]:
        """Get aggregated usage for a period"""
        async with self.pool.acquire() as conn:
            # Total usage
            total_row = await conn.fetchrow("""
                SELECT
                    SUM(total_tokens) as total_tokens,
                    COUNT(*) as requests_count,
                    SUM(CASE WHEN cache_hit THEN 1 ELSE 0 END) as cache_hits,
                    SUM(cost_usd) as total_cost
                FROM inference_logs
                WHERE tenant_id = $1
                AND created_at >= $2
                AND created_at < $3
            """, tenant_id, period_start, period_end)
            
            # By model
            model_rows = await conn.fetch("""
                SELECT model, SUM(total_tokens) as tokens
                FROM inference_logs
                WHERE tenant_id = $1
                AND created_at >= $2
                AND created_at < $3
                GROUP BY model
            """, tenant_id, period_start, period_end)
            
            # By provider
            provider_rows = await conn.fetch("""
                SELECT provider, SUM(total_tokens) as tokens
                FROM inference_logs
                WHERE tenant_id = $1
                AND created_at >= $2
                AND created_at < $3
                GROUP BY provider
            """, tenant_id, period_start, period_end)
            
            return {
                "total_tokens": int(total_row["total_tokens"]) if total_row["total_tokens"] else 0,
                "requests_count": int(total_row["requests_count"]) if total_row["requests_count"] else 0,
                "cache_hits": int(total_row["cache_hits"]) if total_row["cache_hits"] else 0,
                "total_cost": Decimal(str(total_row["total_cost"])) if total_row["total_cost"] else Decimal("0"),
                "by_model": {r["model"]: int(r["tokens"]) for r in model_rows},
                "by_provider": {r["provider"]: int(r["tokens"]) for r in provider_rows},
            }
    
    async def get_billing_record(
        self,
        tenant_id: str,
        period_start: date,
    ) -> Optional[Dict[str, Any]]:
        """Get billing record for a period"""
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow("""
                SELECT * FROM billing_records
                WHERE tenant_id = $1
                AND period_start = $2
            """, tenant_id, period_start)
            
            if not row:
                return None
            
            return {
                "id": str(row["id"]),
                "tenant_id": str(row["tenant_id"]),
                "period_start": row["period_start"],
                "period_end": row["period_end"],
                "tokens_used": row["tokens_used"],
                "total_cost": row["total_cost"],
                "status": row["status"],
                "stripe_invoice_id": row["stripe_invoice_id"],
            }
    
    async def create_billing_record(
        self,
        tenant_id: str,
        period_start: date,
        period_end: date,
        tokens_used: int,
        total_cost: Decimal,
        stripe_invoice_id: Optional[str] = None,
    ):
        """Create or update billing record"""
        async with self.pool.acquire() as conn:
            await conn.execute("""
                INSERT INTO billing_records (
                    id, tenant_id, period_start, period_end,
                    tokens_used, total_cost, status, stripe_invoice_id,
                    created_at
                ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
                ON CONFLICT (tenant_id, period_start) DO UPDATE SET
                    tokens_used = EXCLUDED.tokens_used,
                    total_cost = EXCLUDED.total_cost,
                    stripe_invoice_id = EXCLUDED.stripe_invoice_id,
                    status = EXCLUDED.status
            """,
                str(__import__("uuid").uuid4()),
                tenant_id,
                period_start,
                period_end,
                tokens_used,
                total_cost,
                "pending" if not stripe_invoice_id else "invoiced",
                stripe_invoice_id,
            )
