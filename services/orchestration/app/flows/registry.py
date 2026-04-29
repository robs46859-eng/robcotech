"""
Flow Registry

Defines and manages workflow types available in the system.
Each flow is a sequence of steps with defined task types.
"""

from dataclasses import dataclass, field
from typing import List, Dict, Any, Callable
import logging

logger = logging.getLogger(__name__)


@dataclass
class FlowStep:
    """A single step in a workflow"""
    name: str
    task_type: str
    config: Dict[str, Any] = field(default_factory=dict)
    timeout: int = 300
    retries: int = 3


@dataclass
class FlowDefinition:
    """Definition of a workflow type"""
    flow_type: str
    description: str
    steps: List[FlowStep]
    version: str = "1.0.0"


class FlowRegistry:
    """Registry of available workflow types"""
    
    def __init__(self):
        self._flows: Dict[str, FlowDefinition] = {}
    
    def register(self, flow_def: FlowDefinition):
        """Register a flow definition"""
        self._flows[flow_def.flow_type] = flow_def
        logger.info(f"Registered flow: {flow_def.flow_type}")
    
    def has_flow(self, flow_type: str) -> bool:
        """Check if a flow type exists"""
        return flow_type in self._flows
    
    def get_flow(self, flow_type: str) -> FlowDefinition:
        """Get a flow definition"""
        if flow_type not in self._flows:
            raise ValueError(f"Unknown flow type: {flow_type}")
        return self._flows[flow_type]
    
    def list_flows(self) -> List[Dict[str, str]]:
        """List all registered flows"""
        return [
            {
                "flow_type": flow.flow_type,
                "description": flow.description,
                "steps_count": len(flow.steps),
            }
            for flow in self._flows.values()
        ]
    
    async def register_built_in_flows(self):
        """Register built-in workflow types"""
        
        # BIM Project Analysis Flow
        self.register(FlowDefinition(
            flow_type="bim_project_analysis",
            description="Analyze BIM project and generate status report",
            steps=[
                FlowStep(
                    name="retrieve_project_data",
                    task_type="bim_retrieval",
                    config={"source": "bim_store"},
                    timeout=60,
                    retries=2,
                ),
                FlowStep(
                    name="retrieve_memory_context",
                    task_type="memory_retrieval",
                    config={"scope": "project"},
                    timeout=30,
                    retries=2,
                ),
                FlowStep(
                    name="run_cheap_classification",
                    task_type="model_inference",
                    config={
                        "model_tier": "cheap",
                        "task": "classification",
                    },
                    timeout=60,
                    retries=3,
                ),
                FlowStep(
                    name="detect_issues",
                    task_type="bim_issue_detection",
                    config={},
                    timeout=120,
                    retries=2,
                ),
                FlowStep(
                    name="evaluate_confidence",
                    task_type="policy_evaluation",
                    config={
                        "confidence_threshold": 0.8,
                        "escalation_policy": "mid_cost",
                    },
                    timeout=30,
                    retries=1,
                ),
                FlowStep(
                    name="generate_report",
                    task_type="model_inference",
                    config={
                        "model_tier": "mid_cost",
                        "task": "report_generation",
                        "output_schema": "project_status_report",
                    },
                    timeout=120,
                    retries=3,
                ),
                FlowStep(
                    name="validate_output",
                    task_type="schema_validation",
                    config={"schema": "project_status_report"},
                    timeout=30,
                    retries=1,
                ),
                FlowStep(
                    name="write_memory_note",
                    task_type="memory_creation",
                    config={"note_type": "workflow"},
                    timeout=30,
                    retries=2,
                ),
                FlowStep(
                    name="commit_artifact",
                    task_type="artifact_storage",
                    config={"storage": "object_store"},
                    timeout=60,
                    retries=3,
                ),
            ],
        ))
        
        # IFC Ingestion Flow
        self.register(FlowDefinition(
            flow_type="ifc_ingestion",
            description="Ingest and parse IFC file",
            steps=[
                FlowStep(
                    name="validate_file",
                    task_type="file_validation",
                    config={"allowed_types": [".ifc"]},
                    timeout=30,
                    retries=1,
                ),
                FlowStep(
                    name="parse_ifc",
                    task_type="ifc_parsing",
                    config={},
                    timeout=600,
                    retries=2,
                ),
                FlowStep(
                    name="normalize_elements",
                    task_type="data_normalization",
                    config={},
                    timeout=120,
                    retries=2,
                ),
                FlowStep(
                    name="store_elements",
                    task_type="database_insert",
                    config={"table": "bim_elements"},
                    timeout=60,
                    retries=3,
                ),
                FlowStep(
                    name="create_embeddings",
                    task_type="embedding_generation",
                    config={"fields": ["name", "description"]},
                    timeout=120,
                    retries=2,
                ),
            ],
        ))
        
        # Inference Request Flow (gateway-triggered)
        self.register(FlowDefinition(
            flow_type="inference_request",
            description="Route and execute inference request",
            steps=[
                FlowStep(
                    name="check_semantic_cache",
                    task_type="cache_lookup",
                    config={"threshold": 0.95},
                    timeout=30,
                    retries=2,
                ),
                FlowStep(
                    name="classify_request",
                    task_type="model_inference",
                    config={
                        "model_tier": "cheap",
                        "task": "request_classification",
                    },
                    timeout=30,
                    retries=2,
                ),
                FlowStep(
                    name="select_model",
                    task_type="policy_evaluation",
                    config={"policy": "cost_ladder"},
                    timeout=10,
                    retries=1,
                ),
                FlowStep(
                    name="execute_inference",
                    task_type="model_inference",
                    config={},
                    timeout=120,
                    retries=3,
                ),
                FlowStep(
                    name="validate_response",
                    task_type="schema_validation",
                    config={},
                    timeout=30,
                    retries=1,
                ),
                FlowStep(
                    name="record_usage",
                    task_type="billing_record",
                    config={},
                    timeout=30,
                    retries=3,
                ),
            ],
        ))
        
        logger.info(f"Registered {len(self._flows)} built-in flows")
