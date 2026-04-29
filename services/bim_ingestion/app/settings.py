"""BIM Ingestion Service Settings"""

import os
from pydantic_settings import BaseSettings
from typing import Optional


class Settings(BaseSettings):
    """Application settings"""
    
    # Service
    service_name: str = "bim-ingestion"
    host: str = "0.0.0.0"
    port: int = 8082
    log_level: str = "info"
    
    # Database
    database_url: str = "postgresql://postgres:postgres@localhost:5433/fullstackarkham_bim"
    database_pool_size: int = 10
    
    # Redis
    redis_url: str = "redis://localhost:6380/0"
    
    # Storage
    storage_path: str = "/tmp/bim-storage"
    max_upload_size: int = 100 * 1024 * 1024  # 100MB
    
    # Orchestration
    gateway_url: str = "http://localhost:8080"
    orchestration_url: str = "http://localhost:8083"
    
    # Feature flags
    async_parsing: bool = True
    issue_detection: bool = True
    
    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"


settings = Settings()
