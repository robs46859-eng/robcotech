"""
Artifact Storage Task Executor

Store workflow outputs to object storage.
"""

import logging
from typing import Dict, Any
from datetime import datetime
import hashlib

from app.tasks import TaskExecutor

logger = logging.getLogger(__name__)


class ArtifactStorageExecutor(TaskExecutor):
    """Store workflow artifacts to object storage"""
    
    @property
    def task_type(self) -> str:
        return "artifact_storage"
    
    async def execute(
        self,
        workflow_id: str,
        step_name: str,
        step_data: Dict[str, Any],
        workflow_state: Dict[str, Any],
    ) -> Dict[str, Any]:
        """
        Store workflow output as artifact
        
        Step data should contain:
        - storage: storage type (object_store, database, local)
        - path_template: template for artifact path
        - content_field: field containing content to store
        - metadata: optional metadata to store with artifact
        """
        storage_type = step_data.get("storage", "object_store")
        content_field = step_data.get("content_field", "content")
        path_template = step_data.get("path_template", "artifacts/{workflow_id}/{timestamp}.json")
        
        # Get content to store
        previous_step = self._get_previous_step_name(step_name)
        previous_result = workflow_state.get(previous_step, {})
        content = previous_result.get(content_field)
        
        if content is None:
            return {
                "success": False,
                "error": f"No content found in field '{content_field}'",
            }
        
        # Generate artifact path
        timestamp = datetime.utcnow().strftime("%Y%m%d_%H%M%S")
        artifact_path = path_template.format(
            workflow_id=workflow_id,
            timestamp=timestamp,
            tenant_id=workflow_state.get("tenant_id", "default"),
        )
        
        # Store based on storage type
        if storage_type == "local":
            result = await self._store_local(artifact_path, content)
        elif storage_type == "database":
            result = await self._store_database(workflow_id, artifact_path, content)
        else:  # object_store
            result = await self._store_object_store(artifact_path, content)
        
        return result
    
    async def _store_local(self, path: str, content: Any) -> Dict[str, Any]:
        """Store to local filesystem"""
        import os
        import json
        
        # Create directory if needed
        dir_path = os.path.dirname(path)
        os.makedirs(dir_path, exist_ok=True)
        
        # Write content
        try:
            with open(path, "w") as f:
                if isinstance(content, (dict, list)):
                    json.dump(content, f, indent=2)
                else:
                    f.write(str(content))
            
            logger.info(f"Stored artifact locally: {path}")
            
            return {
                "success": True,
                "storage_type": "local",
                "path": path,
                "size_bytes": os.path.getsize(path),
            }
            
        except Exception as e:
            logger.error(f"Local storage error: {e}")
            return {"success": False, "error": str(e)}
    
    async def _store_object_store(self, path: str, content: Any) -> Dict[str, Any]:
        """
        Store to object storage (S3, GCS, etc.)
        
        In production, would use boto3 or google-cloud-storage.
        For now, mock the operation.
        """
        # Mock object storage - in production would actually upload
        content_str = str(content) if not isinstance(content, str) else content
        content_hash = hashlib.sha256(content_str.encode()).hexdigest()[:16]
        
        logger.info(f"Stored artifact in object store: {path} (hash: {content_hash})")
        
        return {
            "success": True,
            "storage_type": "object_store",
            "path": path,
            "object_hash": content_hash,
            "mock": True,  # Indicates this is mocked
        }
    
    async def _store_database(
        self,
        workflow_id: str,
        path: str,
        content: Any,
    ) -> Dict[str, Any]:
        """
        Store to database
        
        In production, would insert into artifacts table.
        For now, mock the operation.
        """
        logger.info(f"Stored artifact in database: {path} for workflow {workflow_id}")
        
        return {
            "success": True,
            "storage_type": "database",
            "path": path,
            "workflow_id": workflow_id,
            "mock": True,  # Indicates this is mocked
        }
    
    def _get_previous_step_name(self, current_step: str) -> str:
        """Get the name of the previous step"""
        parts = current_step.rsplit("_", 1)
        if len(parts) > 1:
            return parts[0]
        return "previous"
