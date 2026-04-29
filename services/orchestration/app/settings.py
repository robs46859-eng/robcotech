"""Orchestration Service Settings"""

import os
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Application settings"""
    
    # Service
    service_name: str = "orchestration"
    host: str = "0.0.0.0"
    port: int = 8083
    log_level: str = "info"
    
    # Database
    database_url: str = "postgresql://postgres:postgres@localhost:5432/fullstackarkham"
    database_pool_size: int = 10
    
    # Redis
    redis_url: str = "redis://localhost:6380/0"
    
    # Gateway
    gateway_url: str = "http://localhost:8080"
    
    # Memory service
    memory_url: str = "http://localhost:8085"
    
    # Semantic cache
    cache_url: str = "http://localhost:8084"
    
    # Task queues
    default_queue: str = "workflow_tasks"
    max_retries: int = 3
    task_timeout: int = 300  # 5 minutes
    
    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"


settings = Settings()
