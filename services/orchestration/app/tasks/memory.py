"""
Memory Task Executors

Execute memory-related workflow tasks:
- memory_retrieval: Retrieve relevant memory notes
- memory_creation: Create new memory notes from workflow
"""

import logging
from typing import Dict, Any
import httpx

from app.tasks import TaskExecutor

logger = logging.getLogger(__name__)


class MemoryRetrievalExecutor(TaskExecutor):
    """Retrieve relevant memory notes"""
    
    @property
    def task_type(self) -> str:
        return "memory_retrieval"
    
    async def execute(
        self,
        workflow_id: str,
        step_name: str,
        step_data: Dict[str, Any],
        workflow_state: Dict[str, Any],
    ) -> Dict[str, Any]:
        """
        Retrieve memory notes relevant to current workflow
        
        Step data should contain:
        - scope: memory scope (user, task, workflow, domain)
        - query: optional search query
        """
        scope = step_data.get("scope", "workflow")
        tenant_id = workflow_state.get("tenant_id")
        workflow_id_state = workflow_state.get("id")
        
        # Build query from workflow context
        query = step_data.get("query", f"workflow {workflow_id_state}")
        
        # Call memory service
        memory_url = step_data.get("memory_service_url", "http://localhost:8085")
        
        try:
            async with httpx.AsyncClient() as client:
                response = await client.post(
                    f"{memory_url}/api/v1/retrieve",
                    json={
                        "query": query,
                        "tenant_id": tenant_id,
                        "note_types": [scope] if scope != "all" else None,
                        "max_results": 10,
                    },
                    timeout=10.0,
                )
                
                if response.status_code != 200:
                    logger.warning(f"Memory retrieval failed: {response.status_code}")
                    return {"success": False, "notes": []}
                
                data = response.json()
                
                return {
                    "success": True,
                    "notes": data.get("notes", []),
                    "total_found": data.get("total", 0),
                    "scope": scope,
                }
                
        except Exception as e:
            logger.error(f"Memory retrieval error: {e}")
            return {"success": False, "notes": [], "error": str(e)}


class MemoryCreationExecutor(TaskExecutor):
    """Create memory notes from workflow"""
    
    @property
    def task_type(self) -> str:
        return "memory_creation"
    
    async def execute(
        self,
        workflow_id: str,
        step_name: str,
        step_data: Dict[str, Any],
        workflow_state: Dict[str, Any],
    ) -> Dict[str, Any]:
        """
        Create memory note from workflow results
        
        Step data should contain:
        - note_type: type of memory note
        - title_template: template for note title
        - content_fields: fields to include in content
        """
        note_type = step_data.get("note_type", "workflow")
        tenant_id = workflow_state.get("tenant_id")
        
        # Build note content from workflow state
        title = f"Workflow {workflow_id[:8]} - {note_type}"
        
        # Extract key results from workflow state
        results_summary = self._extract_results_summary(workflow_state)
        
        content = f"""
Workflow: {workflow_id}
Type: {note_type}
Tenant: {tenant_id}

Results Summary:
{results_summary}
        """.strip()
        
        # Call memory service to create note
        memory_url = step_data.get("memory_service_url", "http://localhost:8085")
        
        try:
            async with httpx.AsyncClient() as client:
                response = await client.post(
                    f"{memory_url}/api/v1/notes",
                    json={
                        "note_type": note_type,
                        "title": title,
                        "content": content,
                        "tenant_id": tenant_id,
                        "workflow_id": workflow_id,
                        "tags": [note_type, "auto-generated"],
                        "importance": 0.5,
                    },
                    timeout=10.0,
                )
                
                if response.status_code not in [200, 201]:
                    logger.warning(f"Memory creation failed: {response.status_code}")
                    return {"success": False}
                
                data = response.json()
                
                return {
                    "success": True,
                    "note_id": data.get("id"),
                    "title": data.get("title"),
                }
                
        except Exception as e:
            logger.error(f"Memory creation error: {e}")
            return {"success": False, "error": str(e)}
    
    def _extract_results_summary(self, workflow_state: Dict[str, Any]) -> str:
        """Extract a summary of results from workflow state"""
        summaries = []
        
        for key, value in workflow_state.items():
            if isinstance(value, dict) and value.get("success"):
                if "elements" in value:
                    summaries.append(f"- {key}: {len(value['elements'])} elements")
                elif "issues" in value:
                    summaries.append(f"- {key}: {len(value['issues'])} issues")
                elif "notes" in value:
                    summaries.append(f"- {key}: {len(value['notes'])} notes")
        
        return "\n".join(summaries) if summaries else "No structured results found"
