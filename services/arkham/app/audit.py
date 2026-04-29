"""
Security Auditor

Audit logging and fingerprint tracking for Arkham security events.
"""

import logging
from typing import Dict, Any, List, Optional
from datetime import datetime, date

import asyncpg
from orjson import dumps, loads

logger = logging.getLogger(__name__)


class SecurityAuditor:
    """PostgreSQL-backed security audit logging"""
    
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
        logger.info("Security auditor initialized")
    
    async def close(self):
        """Close database connections"""
        if self.pool:
            await self.pool.close()
    
    async def log_event(
        self,
        tenant_id: str,
        event_type: str,
        request_id: str,
        source_ip: str,
        fingerprint_hash: Optional[str],
        threat_score: float,
        deception_engaged: bool,
        blocked: bool,
        cross_tenant_shared: bool = False,
        metadata: Optional[Dict[str, Any]] = None,
    ):
        """Log a security event"""
        async with self.pool.acquire() as conn:
            await conn.execute("""
                INSERT INTO arkham_events (
                    id, tenant_id, event_type, request_id, source_ip,
                    fingerprint_hash, threat_score, deception_engaged,
                    blocked, cross_tenant_shared, metadata, created_at
                ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
            """,
                str(__import__("uuid").uuid4()),
                tenant_id,
                event_type,
                request_id,
                source_ip,
                fingerprint_hash,
                threat_score,
                deception_engaged,
                blocked,
                cross_tenant_shared,
                dumps(metadata or {}),
            )
            
            # Update attacker fingerprint if provided
            if fingerprint_hash:
                await self._update_fingerprint(
                    conn, fingerprint_hash, event_type, tenant_id
                )
    
    async def _update_fingerprint(
        self,
        conn: asyncpg.Connection,
        fingerprint_hash: str,
        event_type: str,
        tenant_id: str,
    ):
        """Update or create attacker fingerprint record"""
        now = datetime.utcnow()
        
        # Check if fingerprint exists
        existing = await conn.fetchrow("""
            SELECT * FROM attacker_fingerprints
            WHERE fingerprint_hash = $1
        """, fingerprint_hash)
        
        if existing:
            # Update existing
            blocked_tenants = list(existing["blocked_tenants"])
            if tenant_id not in blocked_tenants:
                blocked_tenants.append(tenant_id)
            
            await conn.execute("""
                UPDATE attacker_fingerprints SET
                    last_seen = NOW(),
                    attack_count = attack_count + 1,
                    blocked_tenants = $2
                WHERE fingerprint_hash = $1
            """, fingerprint_hash, blocked_tenants)
        else:
            # Create new
            await conn.execute("""
                INSERT INTO attacker_fingerprints (
                    id, fingerprint_hash, behavior_patterns,
                    first_seen, last_seen, attack_count,
                    blocked_tenants, threat_level
                ) VALUES ($1, $2, $3, NOW(), NOW(), 1, $4, $5)
            """,
                str(__import__("uuid").uuid4()),
                fingerprint_hash,
                dumps({"event_types": [event_type]}),
                [tenant_id],
                self._calculate_threat_level(1),
            )
    
    def _calculate_threat_level(self, attack_count: int) -> str:
        """Calculate threat level based on attack count"""
        if attack_count >= 100:
            return "critical"
        elif attack_count >= 50:
            return "high"
        elif attack_count >= 10:
            return "medium"
        elif attack_count >= 1:
            return "low"
        return "unknown"
    
    async def get_events(
        self,
        tenant_id: str,
        event_type: Optional[str] = None,
        limit: int = 100,
        offset: int = 0,
    ) -> List[Dict[str, Any]]:
        """Get security events for a tenant"""
        async with self.pool.acquire() as conn:
            if event_type:
                rows = await conn.fetch("""
                    SELECT * FROM arkham_events
                    WHERE tenant_id = $1 AND event_type = $2
                    ORDER BY created_at DESC
                    LIMIT $3 OFFSET $4
                """, tenant_id, event_type, limit, offset)
            else:
                rows = await conn.fetch("""
                    SELECT * FROM arkham_events
                    WHERE tenant_id = $1
                    ORDER BY created_at DESC
                    LIMIT $2 OFFSET $3
                """, tenant_id, limit, offset)
            
            return [
                {
                    "id": str(row["id"]),
                    "event_type": row["event_type"],
                    "request_id": row["request_id"],
                    "source_ip": str(row["source_ip"]),
                    "fingerprint_hash": row["fingerprint_hash"],
                    "threat_score": row["threat_score"],
                    "deception_engaged": row["deception_engaged"],
                    "blocked": row["blocked"],
                    "created_at": row["created_at"],
                }
                for row in rows
            ]
    
    async def get_stats(self, tenant_id: str) -> Dict[str, Any]:
        """Get security statistics for a tenant"""
        async with self.pool.acquire() as conn:
            # Total events
            total_row = await conn.fetchrow("""
                SELECT COUNT(*) as count FROM arkham_events
                WHERE tenant_id = $1
            """, tenant_id)
            
            # By type
            type_rows = await conn.fetch("""
                SELECT event_type, COUNT(*) as count
                FROM arkham_events
                WHERE tenant_id = $1
                GROUP BY event_type
            """, tenant_id)
            
            # Blocked count
            blocked_row = await conn.fetchrow("""
                SELECT COUNT(*) as count FROM arkham_events
                WHERE tenant_id = $1 AND blocked = true
            """, tenant_id)
            
            # Deception engagements
            deception_row = await conn.fetchrow("""
                SELECT COUNT(*) as count FROM arkham_events
                WHERE tenant_id = $1 AND deception_engaged = true
            """, tenant_id)
            
            # Unique fingerprints
            fingerprint_row = await conn.fetchrow("""
                SELECT COUNT(DISTINCT fingerprint_hash) as count
                FROM arkham_events
                WHERE tenant_id = $1 AND fingerprint_hash IS NOT NULL
            """, tenant_id)
            
            # Cross-tenant blocks
            cross_tenant_row = await conn.fetchrow("""
                SELECT COUNT(*) as count FROM arkham_events
                WHERE tenant_id = $1 AND cross_tenant_shared = true
            """, tenant_id)
            
            return {
                "total_events": total_row["count"] if total_row else 0,
                "by_type": {r["event_type"]: r["count"] for r in type_rows},
                "blocked_count": blocked_row["count"] if blocked_row else 0,
                "deception_engagements": deception_row["count"] if deception_row else 0,
                "unique_fingerprints": fingerprint_row["count"] if fingerprint_row else 0,
                "cross_tenant_blocks": cross_tenant_row["count"] if cross_tenant_row else 0,
            }
    
    async def get_fingerprint_info(
        self,
        fingerprint_hash: str,
    ) -> Optional[Dict[str, Any]]:
        """Get information about an attacker fingerprint"""
        async with self.pool.acquire() as conn:
            row = await conn.fetchrow("""
                SELECT * FROM attacker_fingerprints
                WHERE fingerprint_hash = $1
            """, fingerprint_hash)
            
            if not row:
                return None
            
            return {
                "fingerprint_hash": row["fingerprint_hash"],
                "first_seen": row["first_seen"],
                "last_seen": row["last_seen"],
                "attack_count": row["attack_count"],
                "threat_level": row["threat_level"],
                "blocked_tenants": list(row["blocked_tenants"]),
                "behavior_patterns": loads(row["behavior_patterns"]) if row["behavior_patterns"] else {},
            }
