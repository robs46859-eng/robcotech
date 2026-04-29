"""Checkpoint Store

Persistent storage for workflow state and checkpoints.
Enables recovery after crashes, retries, and partial execution.
"""

import logging
from typing import Dict, Any, Optional
from datetime import datetime

import asyncpg
from orjson import dumps, loads

logger = logging.getLogger(__name__)


class CheckpointStore:
    """PostgreSQL-backed checkpoint storage"""
    
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
        logger.info("Checkpoint store initialized")
    
    async def close(self):
        """Close database connections"""
        if self.pool:
            await self.pool.close()
            logger.info("Checkpoint store closed")
    
    async def save_workflow(self, workflow: Dict[str, Any]):
        """Save or update workflow state"""
        async with self.pool.acquire() as conn:
            await conn.execute("""
                INSERT INTO workflows (
                    id, tenant_id, name, status, current_step,
                    state, workflow_state, operational_state, cognitive_state,
                    retry_count, max_retries, started_at, completed_at,
                    created_at, updated_at
                ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
                ON CONFLICT (id) DO UPDATE SET
                    status = EXCLUDED.status,
                    current_step = EXCLUDED.current_step,
                    state = EXCLUDED.state,
                    workflow_state = EXCLUDED.workflow_state,
                    operational_state = EXCLUDED.operational_state,
                    cognitive_state = EXCLUDED.cognitive_state,
                    retry_count = EXCLUDED.retry_count,
                    completed_at = EXCLUDED.completed_at,
                    updated_at = NOW()
            """,
                workflow["id"],
                workflow["tenant_id"],
                workflow.get("workflow_type", "unknown"),
                workflow["status"],
                workflow.get("current_step"),
                dumps(workflow.get("input_data", {})),
                dumps(workflow.get("workflow_state", {})),
                dumps(workflow.get("operational_state", {})),
                dumps(workflow.get("cognitive_state", {})),
                workflow["operational_state"].get("retry_count", 0),
                workflow["operational_state"].get("max_retries", 3),
                workflow.get("started_at"),
                workflow.get("completed_at"),
            )
    
    async def get_workflow(self, workflow_id: str) -> Optional[Dict[str, Any]]:
        """Get workflow by ID"""
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow("""
                SELECT * FROM workflows WHERE id = $1
            """, workflow_id)
            
            if not row:
                return None
            
            return {
                "id": str(row["id"]),
                "tenant_id": str(row["tenant_id"]),
                "workflow_type": row["name"],
                "status": row["status"],
                "current_step": row["current_step"],
                "input_data": loads(row["state"]) if row["state"] else {},
                "workflow_state": loads(row["workflow_state"]) if row["workflow_state"] else {},
                "operational_state": loads(row["operational_state"]) if row["operational_state"] else {},
                "cognitive_state": loads(row["cognitive_state"]) if row["cognitive_state"] else {},
                "created_at": row["created_at"],
                "updated_at": row["updated_at"],
                "completed_at": row["completed_at"],
            }
    
    async def get_workflows_by_tenant(
        self,
        tenant_id: str,
        status: Optional[str] = None,
        limit: int = 50,
        offset: int = 0,
    ) -> list[Dict[str, Any]]:
        """Get workflows for a tenant"""
        async with self.pool.acquire() as conn:
            if status:
                rows = await conn.fetch("""
                    SELECT * FROM workflows
                    WHERE tenant_id = $1 AND status = $2
                    ORDER BY created_at DESC
                    LIMIT $3 OFFSET $4
                """, tenant_id, status, limit, offset)
            else:
                rows = await conn.fetch("""
                    SELECT * FROM workflows
                    WHERE tenant_id = $1
                    ORDER BY created_at DESC
                    LIMIT $2 OFFSET $3
                """, tenant_id, limit, offset)
            
            return [
                {
                    "id": str(row["id"]),
                    "workflow_type": row["name"],
                    "status": row["status"],
                    "current_step": row["current_step"],
                    "created_at": row["created_at"],
                }
                for row in rows
            ]
    
    async def save_checkpoint(
        self,
        workflow_id: str,
        step_name: str,
        step_output: Dict[str, Any],
        operational_state: Dict[str, Any],
    ):
        """Save a workflow step checkpoint"""
        async with self.pool.acquire() as conn:
            await conn.execute("""
                INSERT INTO workflow_steps (
                    workflow_id, step_name, step_order, status,
                    output, created_at
                ) VALUES ($1, $2, $3, $4, $5, NOW())
            """,
                workflow_id,
                step_name,
                0,  # step_order - would need to be calculated
                "completed",
                dumps(step_output),
            )
            
            # Update workflow operational state
            await conn.execute("""
                UPDATE workflows SET
                    operational_state = $1,
                    updated_at = NOW()
                WHERE id = $2
            """, dumps(operational_state), workflow_id)
    
    async def get_failed_workflows(
        self,
        limit: int = 100,
    ) -> list[Dict[str, Any]]:
        """Get failed workflows for retry processing"""
        async with self.pool.acquire() as conn:
            rows = await conn.fetch("""
                SELECT * FROM workflows
                WHERE status = 'failed'
                AND operational_state->>'retry_count' < '3'
                ORDER BY updated_at ASC
                LIMIT $1
            """, limit)
            
            return [
                {
                    "id": str(row["id"]),
                    "tenant_id": str(row["tenant_id"]),
                    "workflow_type": row["name"],
                    "operational_state": loads(row["operational_state"]),
                }
                for row in rows
            ]
