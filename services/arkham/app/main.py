"""
Arkham Security Service

Active defense and AI deception layer for FullStackArkham.

Capabilities:
- Threat classification (benign, probe, attack, scanner)
- Behavioral fingerprinting
- Novel trap generation (no two traps identical)
- Creative block responses
- Cross-tenant attacker intelligence sharing

Aligned with MITRE ATT&CK framework.
"""

import logging
from contextlib import asynccontextmanager
from typing import Dict, Any, List, Optional
from uuid import uuid4
import hashlib

from fastapi import FastAPI, Request, HTTPException
from pydantic import BaseModel, Field

from app.settings import settings
from app.detector import ThreatDetector
from app.fingerprint import AttackerFingerprinter
from app.deception import DeceptionGenerator
from app.audit import SecurityAuditor

logger = logging.getLogger(__name__)


class ThreatClassification(BaseModel):
    """Threat classification result"""
    request_id: str
    classification: str  # benign, probe, attack, scanner
    threat_score: float
    fingerprint_hash: Optional[str] = None
    recommended_action: str  # pass, deceive, block
    metadata: Dict[str, Any] = Field(default_factory=dict)


class DeceptionResponse(BaseModel):
    """Deception trap response"""
    request_id: str
    trap_type: str
    deception_payload: Dict[str, Any]
    engagement_id: str


class SecurityEvent(BaseModel):
    """Security event for audit"""
    event_id: str
    tenant_id: str
    event_type: str
    request_id: str
    source_ip: str
    fingerprint_hash: Optional[str]
    threat_score: float
    deception_engaged: bool
    blocked: bool
    cross_tenant_shared: bool


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan handler"""
    # Startup
    logger.info("Starting Arkham Security Service")
    
    app.state.detector = ThreatDetector()
    app.state.fingerprinter = AttackerFingerprinter()
    app.state.deception = DeceptionGenerator()
    app.state.auditor = SecurityAuditor(settings.database_url)
    
    await app.state.auditor.initialize()
    
    yield
    
    # Shutdown
    logger.info("Shutting down Arkham Security Service")
    await app.state.auditor.close()


app = FastAPI(
    title="Arkham Security Service",
    description="Active defense and AI deception layer",
    version="0.1.0",
    lifespan=lifespan,
)


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {"status": "healthy", "service": "arkham"}


@app.get("/ready")
async def ready_check():
    """Readiness check endpoint"""
    if not hasattr(app.state, "detector"):
        return {"status": "not ready", "reason": "detector not initialized"}
    return {"status": "ready"}


@app.post("/api/v1/classify", response_model=ThreatClassification)
async def classify_request(request: Request, tenant_id: str):
    """
    Classify an incoming request for threats
    
    Analyzes request patterns and returns:
    - Classification: benign, probe, attack, scanner
    - Threat score: 0.0-1.0
    - Recommended action: pass, deceive, block
    """
    detector = app.state.detector
    fingerprinter = app.state.fingerprinter
    auditor = app.state.auditor
    
    request_id = str(uuid4())
    
    # Extract request features
    features = await extract_request_features(request)
    
    # Classify threat
    classification = await detector.classify(features, tenant_id)
    
    # Generate fingerprint if suspicious
    fingerprint_hash = None
    if classification["classification"] != "benign":
        fingerprint_hash = await fingerprinter.generate_fingerprint(
            features=features,
            behavior_pattern=classification.get("behavior_pattern", {}),
        )
    
    # Log security event
    await auditor.log_event(
        tenant_id=tenant_id,
        event_type=classification["classification"],
        request_id=request_id,
        source_ip=features["source_ip"],
        fingerprint_hash=fingerprint_hash,
        threat_score=classification["threat_score"],
        deception_engaged=False,
        blocked=False,
    )
    
    return ThreatClassification(
        request_id=request_id,
        classification=classification["classification"],
        threat_score=classification["threat_score"],
        fingerprint_hash=fingerprint_hash,
        recommended_action=classification["recommended_action"],
        metadata=classification.get("metadata", {}),
    )


@app.post("/api/v1/deceive", response_model=DeceptionResponse)
async def generate_deception(
    request: Request,
    tenant_id: str,
    fingerprint_hash: str,
):
    """
    Generate a novel deception trap
    
    Creates a unique trap configuration for this attacker session.
    No two traps are identical for the same signature.
    """
    deception = app.state.deception
    auditor = app.state.auditor
    
    request_id = str(uuid4())
    engagement_id = str(uuid4())
    
    # Generate novel trap
    trap = await deception.generate_trap(
        tenant_id=tenant_id,
        fingerprint_hash=fingerprint_hash,
        request_context=await extract_request_features(request),
    )
    
    # Log deception engagement
    await auditor.log_event(
        tenant_id=tenant_id,
        event_type="deception_engaged",
        request_id=request_id,
        source_ip=trap.get("source_ip", "unknown"),
        fingerprint_hash=fingerprint_hash,
        threat_score=0.8,
        deception_engaged=True,
        blocked=False,
        cross_tenant_shared=settings.cross_tenant_share,
    )
    
    return DeceptionResponse(
        request_id=request_id,
        trap_type=trap["trap_type"],
        deception_payload=trap["payload"],
        engagement_id=engagement_id,
    )


@app.post("/api/v1/block")
async def creative_block(
    tenant_id: str,
    fingerprint_hash: str,
    engagement_id: Optional[str] = None,
):
    """
    Apply creative block response
    
    Returns a plausible but false API response that appears legitimate
    while silently blocking the attacker.
    """
    deception = app.state.deception
    auditor = app.state.auditor
    
    # Generate creative block
    block_response = await deception.generate_block(
        tenant_id=tenant_id,
        fingerprint_hash=fingerprint_hash,
        engagement_id=engagement_id,
    )
    
    # Log block event
    await auditor.log_event(
        tenant_id=tenant_id,
        event_type="blocked",
        request_id=str(uuid4()),
        source_ip=block_response.get("source_ip", "unknown"),
        fingerprint_hash=fingerprint_hash,
        threat_score=0.9,
        deception_engaged=engagement_id is not None,
        blocked=True,
        cross_tenant_shared=settings.cross_tenant_share,
    )
    
    # Share fingerprint across tenants if enabled
    if settings.cross_tenant_share:
        await deception.share_fingerprint(fingerprint_hash)
    
    return block_response


@app.get("/api/v1/fingerprint/{fingerprint_hash}")
async def get_fingerprint_info(fingerprint_hash: str):
    """Get information about an attacker fingerprint"""
    auditor = app.state.auditor
    
    info = await auditor.get_fingerprint_info(fingerprint_hash)
    
    if not info:
        raise HTTPException(status_code=404, detail="Fingerprint not found")
    
    return info


@app.get("/api/v1/events")
async def get_security_events(
    tenant_id: str,
    event_type: Optional[str] = None,
    limit: int = 100,
    offset: int = 0,
):
    """Get security events for a tenant"""
    auditor = app.state.auditor
    
    events = await auditor.get_events(
        tenant_id=tenant_id,
        event_type=event_type,
        limit=limit,
        offset=offset,
    )
    
    return {
        "events": events,
        "total": len(events),
    }


@app.get("/api/v1/stats")
async def get_security_stats(tenant_id: str):
    """Get security statistics for a tenant"""
    auditor = app.state.auditor
    
    stats = await auditor.get_stats(tenant_id)
    
    return {
        "tenant_id": tenant_id,
        "total_events": stats.get("total_events", 0),
        "by_type": stats.get("by_type", {}),
        "blocked_count": stats.get("blocked_count", 0),
        "deception_engagements": stats.get("deception_engagements", 0),
        "unique_fingerprints": stats.get("unique_fingerprints", 0),
        "cross_tenant_blocks": stats.get("cross_tenant_blocks", 0),
    }


async def extract_request_features(request: Request) -> Dict[str, Any]:
    """Extract features from request for analysis"""
    import time
    
    return {
        "source_ip": str(request.client.host) if request.client else "unknown",
        "method": request.method,
        "path": request.url.path,
        "headers": dict(request.headers),
        "user_agent": request.headers.get("user-agent", ""),
        "content_type": request.headers.get("content-type", ""),
        "content_length": int(request.headers.get("content-length", 0)),
        "timestamp": time.time(),
        # Add timing analysis
        "request_time": time.time(),
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "app.main:app",
        host="0.0.0.0",
        port=8081,
        reload=True,
    )
