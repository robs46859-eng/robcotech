"""
BIM Ingestion Service

Handles IFC file parsing, normalization, and storage for the FullStackArkham platform.
This service is responsible for:
- Accepting uploaded IFC and related project files
- Validating format and metadata
- Extracting core entities
- Creating normalized records
- Pushing heavy parsing into queues when needed
"""

import logging
from contextlib import asynccontextmanager
from pathlib import Path
from typing import List, Optional

from fastapi import FastAPI, File, UploadFile, HTTPException, BackgroundTasks
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field

from app.settings import settings
from app.parsers.ifc import IFCParser
from app.storage import BIMStorage
from app.jobs import parse_ifc_job

logger = logging.getLogger(__name__)


class ProjectInfo(BaseModel):
    """Project information extracted from IFC"""
    name: str
    description: Optional[str] = None
    author: Optional[str] = None
    organization: Optional[str] = None
    timestamp: Optional[str] = None


class BuildingElement(BaseModel):
    """Normalized building element"""
    id: str
    type: str
    name: Optional[str] = None
    description: Optional[str] = None
    properties: dict = Field(default_factory=dict)
    quantities: dict = Field(default_factory=dict)
    materials: List[str] = Field(default_factory=list)
    spatial_container: Optional[str] = None


class Issue(BaseModel):
    """Detected issue in BIM data"""
    id: str
    type: str
    severity: str = "low"
    element_id: Optional[str] = None
    description: str
    location: Optional[str] = None


class IngestionResult(BaseModel):
    """Result of BIM ingestion"""
    project_id: str
    elements_count: int
    issues_count: int
    status: str
    message: str


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan handler"""
    # Startup
    logger.info("Starting BIM Ingestion Service")
    storage = BIMStorage(settings.storage_path)
    await storage.initialize()
    app.state.storage = storage
    app.state.parser = IFCParser()
    yield
    # Shutdown
    logger.info("Shutting down BIM Ingestion Service")
    await app.state.storage.close()


app = FastAPI(
    title="BIM Ingestion Service",
    description="IFC parsing and normalization for FullStackArkham",
    version="0.1.0",
    lifespan=lifespan,
)


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {"status": "healthy", "service": "bim-ingestion"}


@app.get("/ready")
async def ready_check():
    """Readiness check endpoint"""
    # Check if storage is available
    if not hasattr(app.state, "storage"):
        return JSONResponse(
            status_code=503,
            content={"status": "not ready", "reason": "storage not initialized"}
        )
    return {"status": "ready"}


@app.post("/api/v1/projects", response_model=IngestionResult)
async def ingest_project(
    files: List[UploadFile] = File(...),
    background_tasks: BackgroundTasks = None,
):
    """
    Ingest BIM project files (IFC, PDFs, schedules, markups)
    
    Accepts uploaded files and processes them:
    - IFC files are parsed and normalized
    - Heavy parsing can be queued for async processing
    - Results are stored in the BIM domain store
    """
    storage = app.state.storage
    parser = app.state.parser
    
    project_id = None
    elements_count = 0
    issues_count = 0
    
    for file in files:
        if not file.filename:
            continue
            
        # Save uploaded file
        file_path = await storage.save_upload(file)
        
        # Process based on file type
        if file.filename.endswith('.ifc'):
            try:
                # Parse IFC file
                ifc_data = await parser.parse(file_path)
                
                # Store elements
                elements = ifc_data.get('elements', [])
                elements_count += len(elements)
                
                # Detect issues
                issues = ifc_data.get('issues', [])
                issues_count += len(issues)
                
                # Store in database
                # await storage.store_elements(project_id, elements)
                # await storage.store_issues(project_id, issues)
                
                logger.info(f"Processed IFC: {file.filename}, {len(elements)} elements")
                
            except Exception as e:
                logger.error(f"Failed to parse IFC {file.filename}: {e}")
                raise HTTPException(
                    status_code=400,
                    detail=f"Failed to parse IFC file: {str(e)}"
                )
        
        elif file.filename.endswith('.pdf'):
            # Queue PDF processing for later
            if background_tasks:
                background_tasks.add_task(
                    parse_ifc_job,
                    file_path=str(file_path),
                    file_type='pdf'
                )
    
    return IngestionResult(
        project_id=project_id or "pending",
        elements_count=elements_count,
        issues_count=issues_count,
        status="processed",
        message=f"Successfully processed {len(files)} files"
    )


@app.get("/api/v1/projects/{project_id}/elements")
async def get_project_elements(project_id: str, limit: int = 100, offset: int = 0):
    """Get elements for a project"""
    # TODO: Implement element retrieval from database
    return {
        "project_id": project_id,
        "elements": [],
        "total": 0,
        "limit": limit,
        "offset": offset
    }


@app.get("/api/v1/projects/{project_id}/issues")
async def get_project_issues(project_id: str, limit: int = 100, offset: int = 0):
    """Get issues for a project"""
    # TODO: Implement issue retrieval from database
    return {
        "project_id": project_id,
        "issues": [],
        "total": 0,
        "limit": limit,
        "offset": offset
    }


@app.post("/api/v1/projects/{project_id}/analyze")
async def analyze_project(project_id: str):
    """
    Run analysis workflow on a project
    
    This triggers an orchestration flow that:
    1. Retrieves all elements for the project
    2. Runs issue detection
    3. Generates compliance summary
    4. Creates project status artifact
    """
    # TODO: Trigger orchestration workflow
    return {
        "workflow_id": "pending",
        "status": "queued",
        "message": "Analysis workflow queued"
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "app.main:app",
        host="0.0.0.0",
        port=8082,
        reload=True,
    )
