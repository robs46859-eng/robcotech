"""
Policy and Validation Task Executors

- policy_evaluation: Evaluate routing and escalation policies
- schema_validation: Validate outputs against JSON schemas
"""

import logging
from typing import Dict, Any, Optional

from app.tasks import TaskExecutor

logger = logging.getLogger(__name__)


class PolicyEvaluationExecutor(TaskExecutor):
    """Evaluate routing and escalation policies"""
    
    @property
    def task_type(self) -> str:
        return "policy_evaluation"
    
    async def execute(
        self,
        workflow_id: str,
        step_name: str,
        step_data: Dict[str, Any],
        workflow_state: Dict[str, Any],
    ) -> Dict[str, Any]:
        """
        Evaluate policy for escalation decisions
        
        Step data should contain:
        - confidence_threshold: minimum confidence to proceed
        - escalation_policy: policy for escalation (cheap, mid, premium)
        - validation_result: optional previous validation result
        """
        confidence_threshold = step_data.get("confidence_threshold", 0.8)
        escalation_policy = step_data.get("escalation_policy", "mid_cost")
        
        # Get confidence from previous step result
        previous_step = self._get_previous_step_name(step_name)
        previous_result = workflow_state.get(previous_step, {})
        
        confidence = previous_result.get("confidence", 0.5)
        
        # Evaluate if escalation is needed
        should_escalate = confidence < confidence_threshold
        
        # Determine target tier based on policy
        if should_escalate:
            target_tier = self._get_escalation_tier(escalation_policy)
        else:
            target_tier = "current"
        
        logger.info(
            f"Policy evaluation: confidence={confidence:.2f}, "
            f"threshold={confidence_threshold:.2f}, escalate={should_escalate}"
        )
        
        return {
            "success": True,
            "should_escalate": should_escalate,
            "confidence": confidence,
            "target_tier": target_tier,
            "reason": (
                f"Confidence {confidence:.2f} below threshold {confidence_threshold:.2f}"
                if should_escalate
                else "Confidence meets threshold"
            ),
        }
    
    def _get_previous_step_name(self, current_step: str) -> str:
        """Get the name of the previous step"""
        # Simple heuristic - would need proper workflow graph traversal in production
        parts = current_step.rsplit("_", 1)
        if len(parts) > 1:
            return parts[0]
        return "previous"
    
    def _get_escalation_tier(self, policy: str) -> str:
        """Get target tier based on escalation policy"""
        policies = {
            "cheap": "cheap",
            "mid_cost": "mid",
            "premium": "premium",
            "aggressive": "premium",
            "conservative": "mid",
        }
        return policies.get(policy, "mid")


class SchemaValidationExecutor(TaskExecutor):
    """Validate outputs against JSON schemas"""
    
    @property
    def task_type(self) -> str:
        return "schema_validation"
    
    async def execute(
        self,
        workflow_id: str,
        step_name: str,
        step_data: Dict[str, Any],
        workflow_state: Dict[str, Any],
    ) -> Dict[str, Any]:
        """
        Validate output against schema
        
        Step data should contain:
        - schema: JSON schema to validate against
        - schema_name: name of schema (for error messages)
        - input_field: field containing data to validate
        """
        schema = step_data.get("schema")
        schema_name = step_data.get("schema_name", "output")
        input_field = step_data.get("input_field", "content")
        
        # Get data to validate
        previous_step = self._get_previous_step_name(step_name)
        previous_result = workflow_state.get(previous_step, {})
        data = previous_result.get(input_field)
        
        if data is None:
            return {
                "success": False,
                "error": f"No data found in field '{input_field}'",
                "valid": False,
            }
        
        # Validate against schema
        try:
            errors = self._validate_schema(data, schema)
            
            if errors:
                logger.warning(f"Schema validation failed: {errors}")
                return {
                    "success": True,
                    "valid": False,
                    "errors": errors,
                    "schema": schema_name,
                }
            
            return {
                "success": True,
                "valid": True,
                "schema": schema_name,
            }
            
        except Exception as e:
            logger.error(f"Validation error: {e}")
            return {
                "success": False,
                "error": str(e),
                "valid": False,
            }
    
    def _validate_schema(self, data: Any, schema: Optional[Dict]) -> list[str]:
        """
        Validate data against schema
        
        In production, would use jsonschema library.
        For now, performs basic type checking.
        """
        errors = []
        
        if schema is None:
            # No schema - assume valid
            return errors
        
        # Basic type checking
        schema_type = schema.get("type")
        
        if schema_type == "object" and not isinstance(data, dict):
            errors.append(f"Expected object, got {type(data).__name__}")
        
        elif schema_type == "array" and not isinstance(data, list):
            errors.append(f"Expected array, got {type(data).__name__}")
        
        elif schema_type == "string" and not isinstance(data, str):
            errors.append(f"Expected string, got {type(data).__name__}")
        
        # Check required fields
        if isinstance(data, dict):
            required = schema.get("required", [])
            for field in required:
                if field not in data:
                    errors.append(f"Missing required field: {field}")
        
        return errors
    
    def _get_previous_step_name(self, current_step: str) -> str:
        """Get the name of the previous step"""
        parts = current_step.rsplit("_", 1)
        if len(parts) > 1:
            return parts[0]
        return "previous"
