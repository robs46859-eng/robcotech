"""
Threat Detector

Classifies incoming requests using behavioral analysis and MITRE ATT&CK patterns.
"""

import logging
from typing import Dict, Any, List
from datetime import datetime, timedelta
import hashlib

logger = logging.getLogger(__name__)


class ThreatDetector:
    """
    Threat classification engine
    
    Detects:
    - Active scanning (T1595)
    - Vulnerability scanning (T1595.002)
    - Brute force (T1110)
    - Credential stuffing (T1110.004)
    - API enumeration (T1592)
    """
    
    def __init__(self):
        # Thresholds for detection
        self.thresholds = {
            "scan_requests_per_minute": 20,
            "auth_failures_per_minute": 10,
            "enumeration_paths_per_minute": 50,
            "payload_anomaly_score": 0.7,
        }
        
        # Request history for rate-based detection
        self._request_history: Dict[str, List[datetime]] = {}
    
    async def classify(
        self,
        features: Dict[str, Any],
        tenant_id: str,
    ) -> Dict[str, Any]:
        """
        Classify a request as benign, probe, attack, or scanner
        
        Returns:
            {
                "classification": str,
                "threat_score": float,
                "recommended_action": str,
                "behavior_pattern": dict,
                "metadata": dict,
            }
        """
        source_ip = features.get("source_ip", "unknown")
        
        # Track request timing
        self._track_request(source_ip)
        
        # Run detection checks
        scan_score = self._detect_scanning(source_ip, features)
        auth_score = self._detect_auth_attack(source_ip, features)
        enum_score = self._detect_enumeration(features)
        payload_score = self._detect_payload_anomalies(features)
        
        # Aggregate scores
        max_score = max(scan_score, auth_score, enum_score, payload_score)
        
        # Classify based on score
        if max_score >= 0.9:
            classification = "attack"
            action = "block"
        elif max_score >= 0.7:
            classification = "probe"
            action = "deceive"
        elif max_score >= 0.4:
            classification = "scanner"
            action = "deceive"
        else:
            classification = "benign"
            action = "pass"
        
        return {
            "classification": classification,
            "threat_score": max_score,
            "recommended_action": action,
            "behavior_pattern": {
                "scan_score": scan_score,
                "auth_score": auth_score,
                "enum_score": enum_score,
                "payload_score": payload_score,
            },
            "metadata": {
                "source_ip": source_ip,
                "request_count": len(self._request_history.get(source_ip, [])),
            },
        }
    
    def _track_request(self, source_ip: str):
        """Track request timing for rate-based detection"""
        now = datetime.utcnow()
        
        if source_ip not in self._request_history:
            self._request_history[source_ip] = []
        
        # Add current request
        self._request_history[source_ip].append(now)
        
        # Clean old entries (keep last 5 minutes)
        cutoff = now - timedelta(minutes=5)
        self._request_history[source_ip] = [
            t for t in self._request_history[source_ip]
            if t > cutoff
        ]
    
    def _detect_scanning(
        self,
        source_ip: str,
        features: Dict[str, Any],
    ) -> float:
        """
        Detect active scanning behavior (T1595)
        
        Indicators:
        - High request frequency
        - Sequential path probing
        - Common vulnerability paths
        """
        score = 0.0
        
        # Check request rate
        recent_requests = self._request_history.get(source_ip, [])
        requests_per_minute = len(recent_requests) / 5.0  # 5 minute window
        
        if requests_per_minute > self.thresholds["scan_requests_per_minute"]:
            score = min(1.0, requests_per_minute / 100.0)
        
        # Check for common scan paths
        scan_paths = [
            "/admin", "/wp-admin", "/phpmyadmin", "/.env",
            "/.git", "/api/swagger", "/graphql", "/actuator",
            "/.well-known", "/robots.txt", "/sitemap.xml",
        ]
        
        path = features.get("path", "").lower()
        if any(scan_path in path for scan_path in scan_paths):
            score = max(score, 0.5)
        
        # Check for automated scanner user agents
        scanner_agents = [
            "nikto", "nmap", "masscan", "zgrab", "gobuster",
            "dirbuster", "wfuzz", "sqlmap", "nuclei",
        ]
        
        user_agent = features.get("user_agent", "").lower()
        if any(agent in user_agent for agent in scanner_agents):
            score = max(score, 0.8)
        
        return score
    
    def _detect_auth_attack(
        self,
        source_ip: str,
        features: Dict[str, Any],
    ) -> float:
        """
        Detect authentication attacks (T1110)
        
        Indicators:
        - Multiple auth failures
        - Credential stuffing patterns
        - Brute force timing
        """
        score = 0.0
        
        # Check if this is an auth endpoint
        auth_paths = ["/login", "/auth", "/api/auth", "/token", "/oauth"]
        path = features.get("path", "").lower()
        
        if not any(auth_path in path for auth_path in auth_paths):
            return score
        
        # Check request rate to auth endpoints
        recent_requests = self._request_history.get(source_ip, [])
        auth_requests_per_minute = len(recent_requests) / 5.0
        
        if auth_requests_per_minute > self.thresholds["auth_failures_per_minute"]:
            score = min(1.0, auth_requests_per_minute / 50.0)
        
        # Check for credential stuffing patterns
        # (many different usernames from same IP in short time)
        content_length = features.get("content_length", 0)
        if 100 < content_length < 500 and auth_requests_per_minute > 5:
            # Typical credential payload size with high frequency
            score = max(score, 0.7)
        
        return score
    
    def _detect_enumeration(self, features: Dict[str, Any]) -> float:
        """
        Detect API enumeration (T1592)
        
        Indicators:
        - Probing many different endpoints
        - Schema discovery attempts
        - Version enumeration
        """
        score = 0.0
        
        path = features.get("path", "").lower()
        
        # Check for enumeration patterns
        enum_patterns = [
            "/api/v",  # Version enumeration
            "/api/",   # API root probing
            "/docs",   # Documentation discovery
            "/openapi", "/swagger", "/spec",  # Schema discovery
            "/health", "/ready", "/metrics",  # Infrastructure probing
            "/debug", "/trace", "/actuator",  # Debug endpoints
        ]
        
        matches = sum(1 for pattern in enum_patterns if pattern in path)
        
        if matches >= 3:
            score = 0.5
        if matches >= 5:
            score = 0.8
        
        # Check for OPTIONS requests (schema probing)
        if features.get("method") == "OPTIONS":
            score = max(score, 0.3)
        
        return score
    
    def _detect_payload_anomalies(self, features: Dict[str, Any]) -> float:
        """
        Detect anomalous payloads
        
        Indicators:
        - SQL injection patterns
        - XSS patterns
        - Path traversal
        - Command injection
        """
        score = 0.0
        
        # Get headers and check for anomalies
        headers = features.get("headers", {})
        content_type = headers.get("content-type", "")
        
        # Check for missing content-type on POST/PUT
        if features.get("method") in ["POST", "PUT"] and not content_type:
            score = 0.2
        
        # Check for suspicious patterns in user agent
        user_agent = features.get("user_agent", "")
        
        # Empty or suspicious user agents
        if not user_agent or len(user_agent) < 10:
            score = max(score, 0.3)
        
        # Check for known attack patterns in path
        attack_patterns = [
            "../", "..\\",  # Path traversal
            "<script", "javascript:",  # XSS
            "union select", "or 1=1", "' or '",  # SQL injection
            "; cat ", "| ls", "`whoami`",  # Command injection
            "${jndi:", "%00",  # Log4j, null byte
        ]
        
        path = features.get("path", "")
        for pattern in attack_patterns:
            if pattern.lower() in path.lower():
                score = max(score, 0.9)
                break
        
        return score
    
    def get_attacker_profile(self, source_ip: str) -> Dict[str, Any]:
        """Get behavioral profile for an IP"""
        requests = self._request_history.get(source_ip, [])
        
        return {
            "source_ip": source_ip,
            "total_requests": len(requests),
            "first_seen": min(requests) if requests else None,
            "last_seen": max(requests) if requests else None,
            "requests_per_minute": len(requests) / 5.0,
        }
