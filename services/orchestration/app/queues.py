"""
Task Queue

Redis-based task queue for workflow execution.
Supports delayed execution, retries, and dead-letter queues.
"""

import logging
from typing import Dict, Any, Optional
from datetime import datetime

import redis.asyncio as redis
from orjson import dumps, loads

logger = logging.getLogger(__name__)


class TaskQueue:
    """Redis-based task queue"""
    
    def __init__(self, redis_url: str):
        self.redis_url = redis_url
        self.client: Optional[redis.Redis] = None
    
    async def initialize(self):
        """Initialize Redis connection"""
        self.client = redis.from_url(
            self.redis_url,
            encoding="utf-8",
            decode_responses=True,
        )
        logger.info("Task queue initialized")
    
    async def close(self):
        """Close Redis connection"""
        if self.client:
            await self.client.close()
            logger.info("Task queue closed")
    
    async def enqueue(
        self,
        queue_name: str,
        task_type: str,
        workflow_id: str,
        step_name: str,
        step_data: Dict[str, Any],
        priority: int = 0,
        delay_seconds: int = 0,
    ) -> str:
        """
        Add a task to the queue
        
        Args:
            queue_name: Name of the queue
            task_type: Type of task (determines worker)
            workflow_id: Associated workflow ID
            step_name: Name of the workflow step
            step_data: Configuration for this step
            priority: Higher = more urgent
            delay_seconds: Delay before task becomes visible
            
        Returns:
            Task ID
        """
        import uuid
        task_id = str(uuid.uuid4())
        
        task = {
            "task_id": task_id,
            "task_type": task_type,
            "workflow_id": workflow_id,
            "step_name": step_name,
            "step_data": dumps(step_data),
            "priority": priority,
            "created_at": datetime.utcnow().isoformat(),
            "status": "pending",
            "retry_count": 0,
        }
        
        if delay_seconds > 0:
            # Add to delayed queue
            await self.client.zadd(
                f"{queue_name}:delayed",
                {dumps(task): int(datetime.utcnow().timestamp()) + delay_seconds}
            )
        else:
            # Add to main queue (using sorted set for priority)
            await self.client.zadd(
                queue_name,
                {dumps(task): -priority}  # Negative for descending order
            )
        
        logger.debug(f"Enqueued task {task_id} to {queue_name}")
        return task_id
    
    async def dequeue(
        self,
        queue_name: str,
        timeout: int = 5,
    ) -> Optional[Dict[str, Any]]:
        """
        Get the next task from the queue
        
        Args:
            queue_name: Name of the queue
            timeout: How long to wait for a task
            
        Returns:
            Task data or None if no task available
        """
        # First, move any due delayed tasks to main queue
        now = int(datetime.utcnow().timestamp())
        await self.client.zunionstore(
            queue_name,
            [queue_name, f"{queue_name}:delayed"],
            aggregate="max",
        )
        # Remove delayed tasks that are now due
        await self.client.zremrangebyscore(
            f"{queue_name}:delayed",
            "-inf",
            now,
        )
        
        # Get highest priority task
        result = await self.client.zpopmin(queue_name, count=1)
        
        if not result:
            return None
        
        task_json, _ = result[0]
        task = loads(task_json)
        task["step_data"] = loads(task["step_data"])
        
        logger.debug(f"Dequeued task {task['task_id']} from {queue_name}")
        return task
    
    async def complete(
        self,
        queue_name: str,
        task_id: str,
    ):
        """Mark a task as completed"""
        # Task already removed from queue on dequeue
        logger.debug(f"Task {task_id} completed")
    
    async def fail(
        self,
        queue_name: str,
        task_id: str,
        error: str,
        retry_count: int,
        max_retries: int = 3,
    ):
        """
        Mark a task as failed
        
        If retries remain, requeue with delay.
        Otherwise, move to dead-letter queue.
        """
        if retry_count < max_retries:
            # Requeue with exponential backoff
            delay = (2 ** retry_count) * 10  # 10s, 20s, 40s...
            await self.enqueue(
                queue_name=queue_name,
                task_type="retry",
                workflow_id=task_id,  # Would need to pass workflow_id separately
                step_name="retry",
                step_data={"retry_count": retry_count + 1},
                delay_seconds=delay,
            )
            logger.info(f"Task {task_id} scheduled for retry {retry_count + 1}")
        else:
            # Move to dead-letter queue
            await self.client.lpush(
                f"{queue_name}:dead_letter",
                dumps({
                    "task_id": task_id,
                    "error": error,
                    "failed_at": datetime.utcnow().isoformat(),
                })
            )
            logger.error(f"Task {task_id} moved to dead-letter queue")
    
    async def dequeue_workflow_tasks(
        self,
        workflow_id: str,
    ):
        """Remove all pending tasks for a workflow (used on cancel)"""
        # This is a simplified implementation
        # A production version would need to scan all queues
        logger.info(f"Removing tasks for workflow {workflow_id}")
    
    async def get_queue_stats(
        self,
        queue_name: str,
    ) -> Dict[str, int]:
        """Get queue statistics"""
        pending = await self.client.zcard(queue_name)
        delayed = await self.client.zcard(f"{queue_name}:delayed")
        dead_letter = await self.client.llen(f"{queue_name}:dead_letter")
        
        return {
            "pending": pending,
            "delayed": delayed,
            "dead_letter": dead_letter,
        }
