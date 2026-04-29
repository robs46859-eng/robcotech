"""
Inference Task Executor

Execute model inference tasks through the gateway.
"""

import logging
from typing import Dict, Any
import httpx

from app.tasks import TaskExecutor

logger = logging.getLogger(__name__)


class ModelInferenceExecutor(TaskExecutor):
    """Execute model inference through gateway"""
    
    @property
    def task_type(self) -> str:
        return "model_inference"
    
    async def execute(
        self,
        workflow_id: str,
        step_name: str,
        step_data: Dict[str, Any],
        workflow_state: Dict[str, Any],
    ) -> Dict[str, Any]:
        """
        Execute model inference
        
        Step data should contain:
        - model_tier: cheap, mid, premium
        - task: task type for routing
        - output_schema: expected output schema
        - messages: prompt messages (or template to build them)
        """
        model_tier = step_data.get("model_tier", "mid")
        task_type = step_data.get("task", "general")
        messages = step_data.get("messages", [])
        
        # Build messages from template if needed
        if not messages and "template" in step_data:
            messages = self._build_messages_from_template(
                step_data["template"],
                workflow_state,
            )
        
        # Call gateway inference endpoint
        gateway_url = step_data.get("gateway_url", "http://localhost:8080")
        tenant_id = workflow_state.get("tenant_id", "default")
        
        try:
            async with httpx.AsyncClient() as client:
                response = await client.post(
                    f"{gateway_url}/v1/ai",
                    json={
                        "messages": messages,
                        "temperature": step_data.get("temperature", 0.7),
                        "max_tokens": step_data.get("max_tokens", 1000),
                    },
                    headers={
                        "X-API-Key": step_data.get("api_key", "dev-key"),
                        "X-Tenant-ID": tenant_id,
                    },
                    timeout=60.0,
                )
                
                if response.status_code != 200:
                    logger.error(f"Inference failed: {response.status_code}")
                    return {
                        "success": False,
                        "error": f"Inference returned {response.status_code}",
                    }
                
                data = response.json()
                
                # Extract content from response
                choices = data.get("choices", [])
                content = choices[0]["message"]["content"] if choices else ""
                
                return {
                    "success": True,
                    "content": content,
                    "model": data.get("model"),
                    "usage": data.get("usage", {}),
                    "model_tier": model_tier,
                }
                
        except Exception as e:
            logger.error(f"Inference error: {e}")
            return {"success": False, "error": str(e)}
    
    def _build_messages_from_template(
        self,
        template: str,
        workflow_state: Dict[str, Any],
    ) -> list[dict]:
        """Build messages from template using workflow state"""
        # Simple template substitution
        # In production, would use proper templating engine
        
        content = template
        
        # Substitute workflow state values
        for key, value in workflow_state.items():
            if isinstance(value, str):
                content = content.replace(f"{{{{{key}}}}}", value)
        
        return [{"role": "user", "content": content}]
