"""
Task Executors

Execute workflow tasks for the orchestration layer.
Each executor handles a specific task type.
"""

import logging
from typing import Dict, Any, Optional
from abc import ABC, abstractmethod

logger = logging.getLogger(__name__)


class TaskExecutor(ABC):
    """Base class for task executors"""
    
    @property
    @abstractmethod
    def task_type(self) -> str:
        """Return the task type this executor handles"""
        pass
    
    @abstractmethod
    async def execute(
        self,
        workflow_id: str,
        step_name: str,
        step_data: Dict[str, Any],
        workflow_state: Dict[str, Any],
    ) -> Dict[str, Any]:
        """Execute the task and return output"""
        pass


class TaskExecutorRegistry:
    """Registry of task executors"""
    
    def __init__(self):
        self._executors: Dict[str, TaskExecutor] = {}
    
    def register(self, executor: TaskExecutor):
        """Register an executor"""
        self._executors[executor.task_type] = executor
        logger.info(f"Registered executor: {executor.task_type}")
    
    def get_executor(self, task_type: str) -> Optional[TaskExecutor]:
        """Get executor for task type"""
        return self._executors.get(task_type)
    
    def list_executors(self) -> list[str]:
        """List registered executor types"""
        return list(self._executors.keys())


# Import all executors to register them
from app.tasks.bim import BIMRetrievalExecutor, BIMIssueDetectionExecutor
from app.tasks.memory import MemoryRetrievalExecutor, MemoryCreationExecutor
from app.tasks.inference import ModelInferenceExecutor
from app.tasks.policy import PolicyEvaluationExecutor
from app.tasks.validation import SchemaValidationExecutor
from app.tasks.artifact import ArtifactStorageExecutor


def create_executor_registry() -> TaskExecutorRegistry:
    """Create and populate executor registry"""
    registry = TaskExecutorRegistry()
    
    # BIM executors
    registry.register(BIMRetrievalExecutor())
    registry.register(BIMIssueDetectionExecutor())
    
    # Memory executors
    registry.register(MemoryRetrievalExecutor())
    registry.register(MemoryCreationExecutor())
    
    # Inference executors
    registry.register(ModelInferenceExecutor())
    
    # Policy executors
    registry.register(PolicyEvaluationExecutor())
    
    # Validation executors
    registry.register(SchemaValidationExecutor())
    
    # Artifact executors
    registry.register(ArtifactStorageExecutor())
    
    return registry
