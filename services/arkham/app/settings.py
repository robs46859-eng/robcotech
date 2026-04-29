"""Arkham Security Service Settings"""

from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Application settings"""
    
    # Service
    service_name: str = "arkham"
    host: str = "0.0.0.0"
    port: int = 8081
    log_level: str = "info"
    
    # Database
    database_url: str = "postgresql://postgres:postgres@localhost:5432/fullstackarkham"
    database_pool_size: int = 10
    
    # Redis
    redis_url: str = "redis://localhost:6379/0"
    
    # Security
    deception_enabled: bool = True
    cross_tenant_share: bool = True
    
    # Detection thresholds
    scan_threshold: float = 0.7
    attack_threshold: float = 0.9
    
    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"


settings = Settings()
