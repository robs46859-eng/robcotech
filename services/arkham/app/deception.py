"""
Deception Generator

Generates novel AI deception traps and creative block responses.
"""

import hashlib
import random
import uuid
from typing import Dict, Any, List, Optional
from datetime import datetime, timedelta


class DeceptionGenerator:
    """
    AI-adaptive deception generation
    
    Key properties:
    - No two traps identical for same attacker signature
    - Plausible but false API responses
    - Evolves mid-session faster than attacker adapts
    """
    
    def __init__(self):
        self.trap_templates = self._load_trap_templates()
        self.block_templates = self._load_block_templates()
        self._session_traps: Dict[str, set] = {}  # Track traps used per session
    
    def _load_trap_templates(self) -> Dict[str, List[Dict[str, Any]]]:
        """Load deception trap templates"""
        return {
            "honeypot_endpoint": [
                {
                    "type": "fake_admin",
                    "response": {
                        "status": "success",
                        "admin_panel": True,
                        "available_actions": ["users", "config", "logs", "export"],
                        "session_token": "{fake_token}",
                    },
                },
                {
                    "type": "fake_debug",
                    "response": {
                        "debug_mode": True,
                        "environment": "development",
                        "database": "postgres://admin:***@db.internal:5432/prod",
                        "api_keys": ["sk_live_{fake_key}"],
                    },
                },
                {
                    "type": "fake_schema",
                    "response": {
                        "openapi": "3.0.0",
                        "info": {"title": "Internal Admin API", "version": "2.0"},
                        "paths": {
                            "/internal/users": {"get": {"summary": "List all users"}},
                            "/internal/config": {"get": {"summary": "Get config"}},
                            "/internal/export": {"post": {"summary": "Export data"}},
                        },
                    },
                },
            ],
            "auth_success": [
                {
                    "type": "fake_login_success",
                    "response": {
                        "status": "authenticated",
                        "user": {"id": "admin", "role": "administrator"},
                        "token": "{fake_jwt}",
                        "expires_in": 3600,
                    },
                },
                {
                    "type": "fake_rate_limit_bypass",
                    "response": {
                        "status": "approved",
                        "message": "Rate limit exception granted",
                        "new_limit": 10000,
                        "valid_until": "{future_timestamp}",
                    },
                },
            ],
            "data_response": [
                {
                    "type": "fake_user_list",
                    "response": {
                        "users": [
                            {"id": "u_001", "email": "admin@company.com", "role": "admin"},
                            {"id": "u_002", "email": "dev@company.com", "role": "developer"},
                            {"id": "u_003", "email": "test@company.com", "role": "tester"},
                        ],
                        "total": 3,
                    },
                },
                {
                    "type": "fake_config_dump",
                    "response": {
                        "config": {
                            "feature_flags": {"new_ui": True, "beta_api": False},
                            "integrations": {"stripe": "enabled", "sendgrid": "enabled"},
                            "limits": {"max_upload_mb": 100, "api_rate_limit": 1000},
                        },
                    },
                },
            ],
        }
    
    def _load_block_templates(self) -> List[Dict[str, Any]]:
        """Load creative block response templates"""
        return [
            {
                "type": "fake_success",
                "http_status": 200,
                "body": {
                    "status": "processing",
                    "message": "Request accepted and queued for processing",
                    "job_id": "{fake_job_id}",
                    "estimated_completion": "{future_timestamp}",
                },
            },
            {
                "type": "fake_rate_limit",
                "http_status": 429,
                "body": {
                    "error": "rate_limit_exceeded",
                    "message": "Rate limit temporarily exceeded. Please retry after {seconds}s",
                    "retry_after": "{seconds}",
                    "current_limit": 100,
                },
            },
            {
                "type": "fake_maintenance",
                "http_status": 503,
                "body": {
                    "error": "service_unavailable",
                    "message": "Service undergoing scheduled maintenance",
                    "maintenance_window": "{time_range}",
                    "expected_resolution": "{future_timestamp}",
                },
            },
            {
                "type": "fake_timeout",
                "http_status": 504,
                "body": {
                    "error": "gateway_timeout",
                    "message": "Upstream service timed out",
                    "upstream": "auth-service",
                    "request_id": "{fake_request_id}",
                },
            },
        ]
    
    async def generate_trap(
        self,
        tenant_id: str,
        fingerprint_hash: str,
        request_context: Dict[str, Any],
    ) -> Dict[str, Any]:
        """
        Generate a novel deception trap
        
        Ensures no two traps are identical for the same fingerprint.
        """
        session_key = f"{tenant_id}:{fingerprint_hash}"
        
        # Track used traps for this session
        if session_key not in self._session_traps:
            self._session_traps[session_key] = set()
        
        used_traps = self._session_traps[session_key]
        
        # Select trap type based on request context
        trap_type = self._select_trap_type(request_context)
        templates = self.trap_templates.get(trap_type, [])
        
        if not templates:
            templates = self.trap_templates["honeypot_endpoint"]
        
        # Find unused template
        available = [
            (i, t) for i, t in enumerate(templates)
            if f"{trap_type}_{i}" not in used_traps
        ]
        
        if available:
            # Use unused template
            idx, template = random.choice(available)
        else:
            # All used - pick random and vary it
            idx = random.randint(0, len(templates) - 1)
            template = templates[idx]
            # All templates used, reset for this session
            used_traps.clear()
        
        # Mark as used
        used_traps.add(f"{trap_type}_{idx}")
        
        # Generate fake values
        payload = self._vary_payload(template["response"])
        
        return {
            "trap_type": template["type"],
            "payload": payload,
            "source_ip": request_context.get("source_ip", "unknown"),
            "generated_at": datetime.utcnow().isoformat(),
            "ttl_seconds": 300,  # Trap valid for 5 minutes
        }
    
    def _select_trap_type(self, request_context: Dict[str, Any]) -> str:
        """Select appropriate trap type based on request"""
        path = request_context.get("path", "")
        method = request_context.get("method", "GET")
        
        # Match trap to attacker intent
        if "/admin" in path or "/login" in path:
            return "auth_success"
        elif "/api" in path or "/graphql" in path:
            return "data_response"
        elif method == "OPTIONS":
            return "honeypot_endpoint"
        else:
            return "honeypot_endpoint"
    
    def _vary_payload(self, response: Dict[str, Any]) -> Dict[str, Any]:
        """Vary payload to ensure uniqueness"""
        import json
        
        payload_str = json.dumps(response)
        
        # Replace placeholders
        payload_str = payload_str.replace("{fake_token}", self._generate_token(32))
        payload_str = payload_str.replace("{fake_key}", self._generate_token(24))
        payload_str = payload_str.replace("{fake_jwt}", self._generate_jwt())
        payload_str = payload_str.replace("{fake_job_id}", str(uuid.uuid4()))
        payload_str = payload_str.replace("{fake_request_id}", str(uuid.uuid4()))
        
        # Time-based variations
        future = datetime.utcnow() + timedelta(hours=1)
        payload_str = payload_str.replace("{future_timestamp}", future.isoformat())
        payload_str = payload_str.replace("{seconds}", str(random.randint(60, 300)))
        payload_str = payload_str.replace(
            "{time_range}",
            f"{datetime.utcnow().strftime('%H:%M')}-{(datetime.utcnow() + timedelta(hours=2)).strftime('%H:%M')} UTC"
        )
        
        return json.loads(payload_str)
    
    def _generate_token(self, length: int) -> str:
        """Generate random token"""
        import secrets
        return secrets.token_hex(length // 2)
    
    def _generate_jwt(self) -> str:
        """Generate fake JWT-like token"""
        header = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
        payload = self._generate_token(50)
        signature = self._generate_token(32)
        return f"{header}.{payload}.{signature}"
    
    async def generate_block(
        self,
        tenant_id: str,
        fingerprint_hash: str,
        engagement_id: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        Generate creative block response
        
        Returns plausible response that appears legitimate while blocking.
        """
        template = random.choice(self.block_templates)
        
        # Vary the response
        body = self._vary_payload(template["body"])
        
        return {
            "http_status": template["http_status"],
            "headers": {
                "Content-Type": "application/json",
                "X-Request-ID": str(uuid.uuid4()),
                "X-Block-Reason": "security_policy" if engagement_id else "rate_limit",
            },
            "body": body,
            "source_ip": "unknown",
            "blocked_at": datetime.utcnow().isoformat(),
            "engagement_id": engagement_id,
        }
    
    async def share_fingerprint(
        self,
        fingerprint_hash: str,
    ) -> bool:
        """
        Share attacker fingerprint across tenants
        
        In production, this would update a shared database.
        """
        # Placeholder - would update cross-tenant block list
        logger.info(f"Sharing fingerprint {fingerprint_hash} across tenants")
        return True


# Import logger for the module
import logging
logger = logging.getLogger(__name__)
