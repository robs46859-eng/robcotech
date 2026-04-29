"""
Unit Tests for Arkham Security Detector

Tests threat classification, fingerprinting, and detection logic.
"""

import pytest
from app.detector import ThreatDetector


class TestThreatDetector:
    """Test threat classification"""
    
    @pytest.fixture
    def detector(self):
        """Create detector instance"""
        return ThreatDetector()
    
    @pytest.mark.asyncio
    async def test_benign_request(self, detector):
        """Test benign request classification"""
        features = {
            "source_ip": "192.168.1.1",
            "method": "GET",
            "path": "/api/v1/ai",
            "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0",
            "headers": {
                "content-type": "application/json",
                "user-agent": "Mozilla/5.0...",
            },
            "request_time": 1234567890,
        }
        
        result = await detector.classify(features, "test-tenant")
        
        assert result["classification"] == "benign"
        assert result["threat_score"] < 0.4
        assert result["recommended_action"] == "pass"
    
    @pytest.mark.asyncio
    async def test_scanner_detection(self, detector):
        """Test scanner detection"""
        features = {
            "source_ip": "10.0.0.1",
            "method": "GET",
            "path": "/admin",
            "user_agent": "Nikto/2.1.6",
            "headers": {},
            "request_time": 1234567890,
        }
        
        result = await detector.classify(features, "test-tenant")
        
        assert result["classification"] in ["scanner", "probe"]
        assert result["threat_score"] >= 0.4
    
    @pytest.mark.asyncio
    async def test_auth_attack_detection(self, detector):
        """Test authentication attack detection"""
        # Simulate rapid auth requests
        for i in range(50):
            detector._track_request("10.0.0.2")
        
        features = {
            "source_ip": "10.0.0.2",
            "method": "POST",
            "path": "/api/auth/login",
            "user_agent": "python-requests/2.28.0",
            "headers": {"content-type": "application/json"},
            "content_length": 150,
            "request_time": 1234567890,
        }
        
        result = await detector.classify(features, "test-tenant")
        
        assert result["classification"] in ["probe", "attack"]
        assert result["behavior_pattern"]["auth_score"] > 0
    
    @pytest.mark.asyncio
    async def test_path_traversal_detection(self, detector):
        """Test path traversal attack detection"""
        features = {
            "source_ip": "10.0.0.3",
            "method": "GET",
            "path": "/api/../../../etc/passwd",
            "user_agent": "curl/7.68.0",
            "headers": {},
            "request_time": 1234567890,
        }
        
        result = await detector.classify(features, "test-tenant")
        
        assert result["classification"] == "attack"
        assert result["threat_score"] >= 0.9
        assert result["recommended_action"] == "block"
    
    @pytest.mark.asyncio
    async def test_sql_injection_detection(self, detector):
        """Test SQL injection detection"""
        features = {
            "source_ip": "10.0.0.4",
            "method": "GET",
            "path": "/api/users?id=1' OR '1'='1",
            "user_agent": "Mozilla/5.0",
            "headers": {},
            "request_time": 1234567890,
        }
        
        result = await detector.classify(features, "test-tenant")
        
        assert result["classification"] == "attack"
        assert result["threat_score"] >= 0.9
    
    @pytest.mark.asyncio
    async def test_enumeration_detection(self, detector):
        """Test API enumeration detection"""
        features = {
            "source_ip": "10.0.0.5",
            "method": "GET",
            "path": "/api/v1/openapi.json",
            "user_agent": "Mozilla/5.0",
            "headers": {},
            "request_time": 1234567890,
        }
        
        result = await detector.classify(features, "test-tenant")
        
        # Single enumeration request might be benign
        assert result["behavior_pattern"]["enum_score"] >= 0
    
    @pytest.mark.asyncio
    async def test_empty_user_agent(self, detector):
        """Test empty user agent detection"""
        features = {
            "source_ip": "10.0.0.6",
            "method": "POST",
            "path": "/api/v1/ai",
            "user_agent": "",
            "headers": {},
            "request_time": 1234567890,
        }
        
        result = await detector.classify(features, "test-tenant")
        
        assert result["behavior_pattern"]["payload_score"] > 0
    
    @pytest.mark.asyncio
    async def test_attacker_profile(self, detector):
        """Test attacker profile generation"""
        # Track some requests
        for i in range(10):
            detector._track_request("10.0.0.7")
        
        profile = detector.get_attacker_profile("10.0.0.7")
        
        assert profile["source_ip"] == "10.0.0.7"
        assert profile["total_requests"] == 10
        assert profile["first_seen"] is not None
        assert profile["last_seen"] is not None


class TestFingerprinter:
    """Test attacker fingerprinting"""
    
    @pytest.fixture
    def fingerprinter(self):
        """Create fingerprinter instance"""
        from app.fingerprint import AttackerFingerprinter
        return AttackerFingerprinter()
    
    @pytest.mark.asyncio
    async def test_fingerprint_generation(self, fingerprinter):
        """Test fingerprint generation"""
        features = {
            "source_ip": "10.0.0.1",
            "method": "GET",
            "path": "/api/v1/ai",
            "headers": {
                "user-agent": "Mozilla/5.0",
                "accept": "application/json",
            },
            "content_length": 100,
            "request_time": 1234567890,
        }
        
        behavior_pattern = {
            "scan_score": 0.1,
            "auth_score": 0.0,
            "enum_score": 0.2,
        }
        
        fingerprint = await fingerprinter.generate_fingerprint(features, behavior_pattern)
        
        assert len(fingerprint) == 32  # SHA256 hex truncated
        assert isinstance(fingerprint, str)
    
    @pytest.mark.asyncio
    async def test_fingerprint_stability(self, fingerprinter):
        """Test that same features produce same fingerprint"""
        features = {
            "source_ip": "10.0.0.1",
            "method": "GET",
            "path": "/api/v1/ai",
            "headers": {"user-agent": "Mozilla/5.0"},
            "content_length": 100,
            "request_time": 1234567890,
        }
        
        behavior_pattern = {"scan_score": 0.1}
        
        fp1 = await fingerprinter.generate_fingerprint(features, behavior_pattern)
        fp2 = await fingerprinter.generate_fingerprint(features, behavior_pattern)
        
        assert fp1 == fp2


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
