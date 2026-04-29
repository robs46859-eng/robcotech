"""
Orchestration Service

Event-driven workflow execution for FullStackArkham.
Manages multi-step jobs with:
- Resumability
- Branching
- Retries
- Human checkpoints
- Persistent state
"""

import logging
from contextlib import asynccontextmanager
from typing import Dict, Any
from uuid import uuid4

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from app.settings import settings
from app.flows.registry import FlowRegistry
from app.checkpoints import CheckpointStore
from app.queues import TaskQueue

logger = logging.getLogger(__name__)


class WorkflowStartRequest(BaseModel):
    """Request to start a workflow"""
    workflow_type: str
    tenant_id: str
    input_data: Dict[str, Any] = Field(default_factory=dict)
    metadata: Dict[str, Any] = Field(default_factory=dict)


class WorkflowResponse(BaseModel):
    """Workflow response"""
    workflow_id: str
    status: str
    current_step: str | None = None
    message: str


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan handler"""
    # Startup
    logger.info("Starting Orchestration Service")
    
    # Initialize components
    app.state.flow_registry = FlowRegistry()
    app.state.checkpoint_store = CheckpointStore(settings.database_url)
    app.state.task_queue = TaskQueue(settings.redis_url)
    
    # Register built-in flows
    await app.state.flow_registry.register_built_in_flows()
    
    yield
    
    # Shutdown
    logger.info("Shutting down Orchestration Service")
    await app.state.checkpoint_store.close()
    await app.state.task_queue.close()


app = FastAPI(
    title="Orchestration Service",
    description="Event-driven workflow execution for FullStackArkham",
    version="0.1.0",
    lifespan=lifespan,
)


@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {"status": "healthy", "service": "orchestration"}


@app.get("/ready")
async def ready_check():
    """Readiness check endpoint"""
    if not hasattr(app.state, "checkpoint_store"):
        return {"status": "not ready", "reason": "checkpoint store not initialized"}
    return {"status": "ready"}


@app.post("/api/v1/workflows", response_model=WorkflowResponse)
async def start_workflow(request: WorkflowStartRequest):
    """
    Start a new workflow execution
    
    Creates a workflow instance and queues the first step.
    Workflow state is persisted for recovery.
    """
    flow_registry = app.state.flow_registry
    checkpoint_store = app.state.checkpoint_store
    task_queue = app.state.task_queue
    
    # Validate workflow type exists
    if not flow_registry.has_flow(request.workflow_type):
        raise HTTPException(
            status_code=400,
            detail=f"Unknown workflow type: {request.workflow_type}"
        )
    
    # Create workflow instance
    workflow_id = str(uuid4())
    
    # Initialize workflow state (separated per architecture)
    workflow_state = {
        "id": workflow_id,
        "tenant_id": request.tenant_id,
        "workflow_type": request.workflow_type,
        "status": "pending",
        "current_step": None,
        "workflow_state": {},  # What step the flow is in
        "operational_state": {  # Execution guarantees and recovery metadata
            "retry_count": 0,
            "idempotency_key": str(uuid4()),
            "timeout_budget": 3600,  # 1 hour default
            "last_checkpoint": None,
        },
        "cognitive_state": {  # Context available to models
            "retrieved_memory": [],
            "domain_records": [],
            "user_preferences": {},
        },
        "input_data": request.input_data,
        "metadata": request.metadata,
    }
    
    # Persist initial state
    await checkpoint_store.save_workflow(workflow_state)
    
    # Queue first task
    flow_def = flow_registry.get_flow(request.workflow_type)
    first_step = flow_def.steps[0] if flow_def.steps else None
    
    if first_step:
        await task_queue.enqueue(
            queue_name="workflow_tasks",
            task_type=first_step.task_type,
            workflow_id=workflow_id,
            step_name=first_step.name,
            step_data=first_step.config,
        )
        
        workflow_state["status"] = "running"
        workflow_state["current_step"] = first_step.name
        await checkpoint_store.save_workflow(workflow_state)
    
    logger.info(f"Started workflow {workflow_id} type={request.workflow_type}")
    
    return WorkflowResponse(
        workflow_id=workflow_id,
        status=workflow_state["status"],
        current_step=workflow_state["current_step"],
        message=f"Workflow {request.workflow_type} started",
    )


@app.get("/api/v1/workflows/{workflow_id}")
async def get_workflow(workflow_id: str):
    """Get workflow status and state"""
    checkpoint_store = app.state.checkpoint_store
    
    workflow = await checkpoint_store.get_workflow(workflow_id)
    if not workflow:
        raise HTTPException(status_code=404, detail="Workflow not found")
    
    return {
        "workflow_id": workflow_id,
        "status": workflow.get("status"),
        "workflow_type": workflow.get("workflow_type"),
        "current_step": workflow.get("current_step"),
        "created_at": workflow.get("created_at"),
        "updated_at": workflow.get("updated_at"),
    }


@app.post("/api/v1/workflows/{workflow_id}/cancel")
async def cancel_workflow(workflow_id: str):
    """Cancel a running workflow"""
    checkpoint_store = app.state.checkpoint_store
    task_queue = app.state.task_queue
    
    workflow = await checkpoint_store.get_workflow(workflow_id)
    if not workflow:
        raise HTTPException(status_code=404, detail="Workflow not found")
    
    # Update status
    workflow["status"] = "cancelled"
    await checkpoint_store.save_workflow(workflow)
    
    # Remove from queue
    await task_queue.dequeue_workflow_tasks(workflow_id)
    
    return {"workflow_id": workflow_id, "status": "cancelled"}


@app.post("/api/v1/workflows/{workflow_id}/retry")
async def retry_workflow(workflow_id: str):
    """Retry a failed workflow from last checkpoint"""
    checkpoint_store = app.state.checkpoint_store
    task_queue = app.state.task_queue
    
    workflow = await checkpoint_store.get_workflow(workflow_id)
    if not workflow:
        raise HTTPException(status_code=404, detail="Workflow not found")
    
    if workflow["status"] != "failed":
        raise HTTPException(
            status_code=400,
            detail="Can only retry failed workflows"
        )
    
    # Get last checkpoint
    last_checkpoint = workflow["operational_state"].get("last_checkpoint")
    if not last_checkpoint:
        raise HTTPException(
            status_code=400,
            detail="No checkpoint available for retry"
        )
    
    # Reset state and requeue
    workflow["status"] = "pending"
    workflow["operational_state"]["retry_count"] += 1
    
    # Requeue from checkpoint
    await task_queue.enqueue(
        queue_name="workflow_tasks",
        task_type=last_checkpoint["task_type"],
        workflow_id=workflow_id,
        step_name=last_checkpoint["step_name"],
        step_data=last_checkpoint["step_data"],
    )
    
    await checkpoint_store.save_workflow(workflow)
    
    return {"workflow_id": workflow_id, "status": "retrying"}


@app.get("/api/v1/flows")
async def list_flows():
    """List available workflow types"""
    flow_registry = app.state.flow_registry
    return {
        "flows": flow_registry.list_flows(),
    }


@app.get("/api/v1/flows/{flow_type}")
async def get_flow_definition(flow_type: str):
    """Get workflow definition"""
    flow_registry = app.state.flow_registry
    
    if not flow_registry.has_flow(flow_type):
        raise HTTPException(status_code=404, detail="Flow not found")
    
    flow_def = flow_registry.get_flow(flow_type)
    return {
        "flow_type": flow_type,
        "description": flow_def.description,
        "steps": [
            {"name": step.name, "task_type": step.task_type}
            for step in flow_def.steps
        ],
    }


@app.post("/api/v1/tasks/complete")
async def complete_task(
    workflow_id: str,
    step_name: str,
    output: Dict[str, Any],
    status: str = "completed",
):
    """
    Mark a task as complete and trigger next step
    
    Called by task workers after executing a workflow step.
    Handles state transitions and checkpointing.
    """
    checkpoint_store = app.state.checkpoint_store
    task_queue = app.state.task_queue
    flow_registry = app.state.flow_registry
    
    workflow = await checkpoint_store.get_workflow(workflow_id)
    if not workflow:
        raise HTTPException(status_code=404, detail="Workflow not found")
    
    if workflow["status"] == "cancelled":
        return {"status": "ignored", "reason": "workflow cancelled"}
    
    flow_def = flow_registry.get_flow(workflow["workflow_type"])
    
    # Find current step and next step
    current_idx = None
    for i, step in enumerate(flow_def.steps):
        if step.name == step_name:
            current_idx = i
            break
    
    if current_idx is None:
        raise HTTPException(status_code=400, detail="Unknown step")
    
    # Save step output
    workflow["workflow_state"][step_name] = output
    workflow["current_step"] = step_name
    workflow["updated_at"] = str(uuid4())  # timestamp
    
    # Determine next action
    if status == "failed":
        workflow["status"] = "failed"
        workflow["operational_state"]["last_checkpoint"] = {
            "task_type": flow_def.steps[current_idx].task_type,
            "step_name": step_name,
            "step_data": flow_def.steps[current_idx].config,
        }
        await checkpoint_store.save_workflow(workflow)
        return {"status": "failed", "workflow_id": workflow_id}
    
    # Check if there's a next step
    if current_idx + 1 < len(flow_def.steps):
        next_step = flow_def.steps[current_idx + 1]
        
        # Queue next task
        await task_queue.enqueue(
            queue_name="workflow_tasks",
            task_type=next_step.task_type,
            workflow_id=workflow_id,
            step_name=next_step.name,
            step_data=next_step.config,
        )
        
        workflow["current_step"] = next_step.name
        workflow["status"] = "running"
    else:
        # Workflow complete
        workflow["status"] = "completed"
        workflow["completed_at"] = str(uuid4())
    
    # Save checkpoint
    await checkpoint_store.save_workflow(workflow)
    
    return {
        "status": workflow["status"],
        "workflow_id": workflow_id,
        "next_step": workflow["current_step"],
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "app.main:app",
        host="0.0.0.0",
        port=8083,
        reload=True,
    )
