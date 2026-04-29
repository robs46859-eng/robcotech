"""
End-to-End Test: BIM Ingestion Workflow

Tests the complete workflow:
1. Upload IFC file (BIMS_Structural.ifc)
2. Parse and extract elements
3. Detect issues
4. Generate project status report

Requires:
- Docker Compose running all services
- BIMS_Structural.ifc file in agent_init/
"""

import asyncio
import httpx
import pytest
import os
from pathlib import Path

# Service URLs
GATEWAY_URL = os.getenv("GATEWAY_URL", "http://localhost:8080")
BIM_INGESTION_URL = os.getenv("BIM_INGESTION_URL", "http://localhost:8082")
ORCHESTRATION_URL = os.getenv("ORCHESTRATION_URL", "http://localhost:8083")
ARKHAM_URL = os.getenv("ARKHAM_URL", "http://localhost:8081")
MEMORY_URL = os.getenv("MEMORY_URL", "http://localhost:8085")

# Test file path
BIM_FILE_PATH = Path(__file__).parent / "BIMS_Structural.ifc"


@pytest.fixture
async def client():
    """Create async HTTP client"""
    async with httpx.AsyncClient(timeout=60.0) as client:
        yield client


@pytest.fixture
def bim_file():
    """Get path to test BIM file"""
    if not BIM_FILE_PATH.exists():
        pytest.skip(f"BIM file not found: {BIM_FILE_PATH}")
    return BIM_FILE_PATH


class TestHealthChecks:
    """Test service health endpoints"""
    
    @pytest.mark.asyncio
    async def test_gateway_health(self, client):
        """Test gateway health endpoint"""
        response = await client.get(f"{GATEWAY_URL}/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
    
    @pytest.mark.asyncio
    async def test_bim_ingestion_health(self, client):
        """Test BIM ingestion service health"""
        response = await client.get(f"{BIM_INGESTION_URL}/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
    
    @pytest.mark.asyncio
    async def test_orchestration_health(self, client):
        """Test orchestration service health"""
        response = await client.get(f"{ORCHESTRATION_URL}/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"


class TestBIMIngestion:
    """Test BIM ingestion workflow"""
    
    @pytest.mark.asyncio
    async def test_upload_ifc_file(self, client, bim_file):
        """Test uploading an IFC file"""
        with open(bim_file, "rb") as f:
            files = {"files": ("BIMS_Structural.ifc", f, "application/x-step")}
            response = await client.post(
                f"{BIM_INGESTION_URL}/api/v1/projects",
                files=files,
            )
        
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "processed"
        assert "project_id" in data
        assert data["elements_count"] > 0
    
    @pytest.mark.asyncio
    async def test_get_project_elements(self, client):
        """Test retrieving project elements"""
        # First create a project (mock for now)
        project_id = "test-project"
        
        response = await client.get(
            f"{BIM_INGESTION_URL}/api/v1/projects/{project_id}/elements",
            params={"limit": 10},
        )
        
        # Should return 200 even if project doesn't exist (returns empty)
        assert response.status_code == 200
        data = response.json()
        assert "elements" in data
        assert "total" in data


class TestOrchestrationWorkflow:
    """Test orchestration workflow execution"""
    
    @pytest.mark.asyncio
    async def test_start_bim_analysis_workflow(self, client):
        """Test starting a BIM analysis workflow"""
        workflow_request = {
            "workflow_type": "bim_project_analysis",
            "tenant_id": "test-tenant",
            "input_data": {
                "project_id": "test-project",
            },
        }
        
        response = await client.post(
            f"{ORCHESTRATION_URL}/api/v1/workflows",
            json=workflow_request,
        )
        
        assert response.status_code == 200
        data = response.json()
        assert "workflow_id" in data
        assert data["status"] in ["pending", "running"]
    
    @pytest.mark.asyncio
    async def test_get_workflow_status(self, client):
        """Test getting workflow status"""
        # Create workflow first
        workflow_request = {
            "workflow_type": "bim_project_analysis",
            "tenant_id": "test-tenant",
            "input_data": {},
        }
        
        create_response = await client.post(
            f"{ORCHESTRATION_URL}/api/v1/workflows",
            json=workflow_request,
        )
        workflow_id = create_response.json()["workflow_id"]
        
        # Get status
        response = await client.get(
            f"{ORCHESTRATION_URL}/api/v1/workflows/{workflow_id}",
        )
        
        assert response.status_code == 200
        data = response.json()
        assert data["workflow_id"] == workflow_id
        assert "status" in data
    
    @pytest.mark.asyncio
    async def test_list_available_flows(self, client):
        """Test listing available workflow types"""
        response = await client.get(f"{ORCHESTRATION_URL}/api/v1/flows")
        
        assert response.status_code == 200
        data = response.json()
        assert "flows" in data
        
        # Check for expected flows
        flow_types = [f["flow_type"] for f in data["flows"]]
        assert "bim_project_analysis" in flow_types
        assert "ifc_ingestion" in flow_types


class TestGatewayInference:
    """Test gateway inference with model routing"""
    
    @pytest.mark.asyncio
    async def test_inference_request(self, client):
        """Test inference request through gateway"""
        inference_request = {
            "messages": [
                {"role": "user", "content": "Classify this building element: load-bearing wall"}
            ],
            "temperature": 0.7,
            "max_tokens": 100,
        }
        
        response = await client.post(
            f"{GATEWAY_URL}/v1/ai",
            json=inference_request,
            headers={"X-API-Key": "test-key"},
        )
        
        # Should return 200 (may use local mock provider)
        assert response.status_code == 200
        data = response.json()
        assert "choices" in data
        assert len(data["choices"]) > 0
    
    @pytest.mark.asyncio
    async def test_model_routing_headers(self, client):
        """Test that model routing headers are present"""
        inference_request = {
            "messages": [
                {"role": "user", "content": "Summarize this text"}
            ],
        }
        
        response = await client.post(
            f"{GATEWAY_URL}/v1/ai",
            json=inference_request,
            headers={"X-API-Key": "test-key"},
        )
        
        assert response.status_code == 200
        # Check for routing headers
        assert "X-Model-Tier" in response.headers
        assert "X-Cache-Hit" in response.headers


class TestArkhamSecurity:
    """Test Arkham security integration"""
    
    @pytest.mark.asyncio
    async def test_benign_request_passes(self, client):
        """Test that benign requests pass through"""
        inference_request = {
            "messages": [
                {"role": "user", "content": "Hello, how are you?"}
            ],
        }
        
        response = await client.post(
            f"{GATEWAY_URL}/v1/ai",
            json=inference_request,
            headers={"X-API-Key": "test-key"},
        )
        
        # Should pass through (200) or be blocked by auth (401)
        assert response.status_code in [200, 401]
    
    @pytest.mark.asyncio
    async def test_scanner_detection(self, client):
        """Test that scanner patterns are detected"""
        # Simulate scanner behavior with multiple rapid requests to admin paths
        scanner_paths = ["/admin", "/.env", "/.git", "/wp-admin"]
        
        for path in scanner_paths:
            response = await client.get(f"{GATEWAY_URL}{path}")
            # Should get 404 or security block
            assert response.status_code in [404, 403, 401]


class TestEndToEndWorkflow:
    """Complete end-to-end workflow test"""
    
    @pytest.mark.asyncio
    async def test_complete_bim_workflow(self, client, bim_file):
        """
        Test complete BIM workflow:
        1. Upload IFC file
        2. Start analysis workflow
        3. Wait for completion
        4. Verify results
        """
        # Step 1: Upload IFC file
        with open(bim_file, "rb") as f:
            files = {"files": ("BIMS_Structural.ifc", f, "application/x-step")}
            upload_response = await client.post(
                f"{BIM_INGESTION_URL}/api/v1/projects",
                files=files,
            )
        
        assert upload_response.status_code == 200
        project_data = upload_response.json()
        project_id = project_data.get("project_id", "test-project")
        
        # Step 2: Start analysis workflow
        workflow_request = {
            "workflow_type": "bim_project_analysis",
            "tenant_id": "test-tenant",
            "input_data": {"project_id": project_id},
        }
        
        workflow_response = await client.post(
            f"{ORCHESTRATION_URL}/api/v1/workflows",
            json=workflow_request,
        )
        
        assert workflow_response.status_code == 200
        workflow_id = workflow_response.json()["workflow_id"]
        
        # Step 3: Poll for completion
        max_attempts = 30
        for _ in range(max_attempts):
            await asyncio.sleep(1)
            
            status_response = await client.get(
                f"{ORCHESTRATION_URL}/api/v1/workflows/{workflow_id}",
            )
            
            status = status_response.json()["status"]
            if status in ["completed", "failed"]:
                break
        
        # Step 4: Verify results
        final_status = await client.get(
            f"{ORCHESTRATION_URL}/api/v1/workflows/{workflow_id}",
        )
        
        assert final_status.status_code == 200
        workflow_result = final_status.json()
        assert workflow_result["status"] == "completed"


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--asyncio-mode=auto"])
