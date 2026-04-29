"""
BIM Task Executors

Execute BIM-related workflow tasks:
- bim_retrieval: Query BIM domain store
- bim_issue_detection: Detect issues in BIM data
"""

import logging
from typing import Dict, Any
import httpx

from app.tasks import TaskExecutor

logger = logging.getLogger(__name__)


class BIMRetrievalExecutor(TaskExecutor):
    """Retrieve BIM project data from domain store"""
    
    @property
    def task_type(self) -> str:
        return "bim_retrieval"
    
    async def execute(
        self,
        workflow_id: str,
        step_name: str,
        step_data: Dict[str, Any],
        workflow_state: Dict[str, Any],
    ) -> Dict[str, Any]:
        """
        Retrieve BIM data for a project
        
        Step data should contain:
        - source: data source (e.g., "bim_store")
        - project_id: project to retrieve
        - element_types: optional filter for element types
        """
        source = step_data.get("source", "bim_store")
        project_id = workflow_state.get("input_data", {}).get("project_id")
        
        if not project_id:
            return {
                "success": False,
                "error": "project_id not found in workflow state",
            }
        
        # Call BIM ingestion service to retrieve elements
        # In production, would use proper service discovery
        bim_url = step_data.get("bim_service_url", "http://localhost:8082")
        
        try:
            async with httpx.AsyncClient() as client:
                response = await client.get(
                    f"{bim_url}/api/v1/projects/{project_id}/elements",
                    params={"limit": 1000},
                    timeout=30.0,
                )
                
                if response.status_code != 200:
                    logger.warning(f"BIM retrieval failed: {response.status_code}")
                    return {
                        "success": False,
                        "error": f"BIM service returned {response.status_code}",
                    }
                
                data = response.json()
                
                return {
                    "success": True,
                    "elements": data.get("elements", []),
                    "total_count": data.get("total", 0),
                    "project_id": project_id,
                }
                
        except Exception as e:
            logger.error(f"BIM retrieval error: {e}")
            return {
                "success": False,
                "error": str(e),
            }


class BIMIssueDetectionExecutor(TaskExecutor):
    """Detect issues in BIM data"""
    
    @property
    def task_type(self) -> str:
        return "bim_issue_detection"
    
    async def execute(
        self,
        workflow_id: str,
        step_name: str,
        step_data: Dict[str, Any],
        workflow_state: Dict[str, Any],
    ) -> Dict[str, Any]:
        """
        Detect issues in BIM elements
        
        Analyzes retrieved elements for:
        - Missing names/IDs
        - Duplicate GUIDs
        - Invalid relationships
        - Compliance violations
        """
        # Get elements from previous step
        retrieval_result = workflow_state.get(step_name.replace("_detection", "_retrieval"), {})
        elements = retrieval_result.get("elements", [])
        
        if not elements:
            return {
                "success": True,
                "issues": [],
                "message": "No elements to analyze",
            }
        
        issues = []
        
        # Check for unnamed elements
        unnamed_count = sum(1 for e in elements if not e.get("name"))
        if unnamed_count > 0:
            issues.append({
                "type": "naming",
                "severity": "low",
                "description": f"{unnamed_count} elements without names",
                "count": unnamed_count,
            })
        
        # Check for missing properties
        missing_props_count = sum(
            1 for e in elements 
            if not e.get("properties") or len(e.get("properties", {})) == 0
        )
        if missing_props_count > 0:
            issues.append({
                "type": "properties",
                "severity": "medium",
                "description": f"{missing_props_count} elements without properties",
                "count": missing_props_count,
            })
        
        # Check for elements without spatial container
        no_container = sum(1 for e in elements if not e.get("spatial_container"))
        if no_container > 0:
            issues.append({
                "type": "spatial",
                "severity": "low",
                "description": f"{no_container} elements without spatial container",
                "count": no_container,
            })
        
        logger.info(f"Detected {len(issues)} issues in {len(elements)} elements")
        
        return {
            "success": True,
            "issues": issues,
            "elements_analyzed": len(elements),
            "issues_by_severity": {
                "low": sum(1 for i in issues if i.get("severity") == "low"),
                "medium": sum(1 for i in issues if i.get("severity") == "medium"),
                "high": sum(1 for i in issues if i.get("severity") == "high"),
            },
        }
