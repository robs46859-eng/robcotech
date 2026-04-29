"""
Attacker Fingerprinter

Generates unique behavioral fingerprints for attackers.
"""

import hashlib
import json
from typing import Dict, Any, List


class AttackerFingerprinter:
    """
    Behavioral fingerprinting for attackers
    
    Creates unique hashes based on:
    - Request patterns
    - Timing behavior
    - Header anomalies
    - Payload characteristics
    """
    
    def __init__(self):
        self.fingerprint_cache: Dict[str, str] = {}
    
    async def generate_fingerprint(
        self,
        features: Dict[str, Any],
        behavior_pattern: Dict[str, Any],
    ) -> str:
        """
        Generate unique fingerprint hash for attacker
        
        Combines multiple signals into a stable identifier
        that persists across session and IP changes.
        """
        # Extract fingerprint components
        components = {
            # Request characteristics
            "methods": self._normalize_methods(features),
            "header_order": self._extract_header_order(features),
            "header_anomalies": self._detect_header_anomalies(features),
            
            # Behavioral patterns
            "timing_pattern": self._extract_timing_pattern(features),
            "path_pattern": self._extract_path_pattern(features),
            
            # Payload characteristics
            "payload_signature": self._extract_payload_signature(features),
        }
        
        # Add behavior scores
        components.update({
            "scan_score": behavior_pattern.get("scan_score", 0),
            "auth_score": behavior_pattern.get("auth_score", 0),
            "enum_score": behavior_pattern.get("enum_score", 0),
        })
        
        # Create stable hash
        fingerprint_data = json.dumps(components, sort_keys=True)
        fingerprint_hash = hashlib.sha256(
            fingerprint_data.encode()
        ).hexdigest()[:32]
        
        return fingerprint_hash
    
    def _normalize_methods(self, features: Dict[str, Any]) -> List[str]:
        """Extract and normalize HTTP methods used"""
        method = features.get("method", "GET")
        return [method.upper()]
    
    def _extract_header_order(self, features: Dict[str, Any]) -> List[str]:
        """Extract header ordering (fingerprintable characteristic)"""
        headers = features.get("headers", {})
        # Return header names in order they appear
        return list(headers.keys())
    
    def _detect_header_anomalies(self, features: Dict[str, Any]) -> List[str]:
        """Detect anomalous header patterns"""
        anomalies = []
        headers = features.get("headers", {})
        
        # Missing standard headers
        standard_headers = ["user-agent", "accept", "accept-language"]
        for header in standard_headers:
            if header not in headers:
                anomalies.append(f"missing_{header}")
        
        # Unusual header values
        user_agent = headers.get("user-agent", "")
        if len(user_agent) > 500:
            anomalies.append("unusually_long_user_agent")
        
        # Contradictory headers
        if "content-length" in headers and headers.get("method") == "GET":
            anomalies.append("content_length_on_get")
        
        return anomalies
    
    def _extract_timing_pattern(self, features: Dict[str, Any]) -> str:
        """Extract timing characteristics"""
        request_time = features.get("request_time", 0)
        
        # Bin into time windows (simplified - would use more history in production)
        time_bin = int(request_time) % 60
        return f"bin_{time_bin}"
    
    def _extract_path_pattern(self, features: Dict[str, Any]) -> str:
        """Extract path characteristics"""
        path = features.get("path", "")
        
        # Categorize path
        if "/api/" in path:
            category = "api"
        elif "/admin" in path or "/wp-" in path:
            category = "admin"
        elif "/." in path:
            category = "hidden"
        else:
            category = "normal"
        
        # Count path segments
        segments = len(path.strip("/").split("/"))
        
        return f"{category}_{segments}"
    
    def _extract_payload_signature(self, features: Dict[str, Any]) -> str:
        """Extract payload characteristics signature"""
        content_length = features.get("content_length", 0)
        content_type = features.get("headers", {}).get("content-type", "")
        
        # Categorize payload
        if content_length == 0:
            return "empty"
        elif "json" in content_type.lower():
            return f"json_{self._bucket_size(content_length)}"
        elif "form" in content_type.lower():
            return f"form_{self._bucket_size(content_length)}"
        else:
            return f"other_{self._bucket_size(content_length)}"
    
    def _bucket_size(self, size: int) -> str:
        """Bucket payload sizes for stable fingerprinting"""
        if size < 100:
            return "tiny"
        elif size < 1000:
            return "small"
        elif size < 10000:
            return "medium"
        elif size < 100000:
            return "large"
        else:
            return "huge"
    
    async def match_known_attacker(
        self,
        fingerprint_hash: str,
        known_fingerprints: List[Dict[str, Any]],
    ) -> bool:
        """Check if fingerprint matches known attacker"""
        return fingerprint_hash in [
            fp.get("fingerprint_hash") for fp in known_fingerprints
        ]
