"""Billing Service Settings"""

from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Application settings"""
    
    # Service
    service_name: str = "billing"
    host: str = "0.0.0.0"
    port: int = 8086
    log_level: str = "info"
    
    # Database
    database_url: str = "postgresql://postgres:postgres@localhost:5432/fullstackarkham"
    database_pool_size: int = 10
    
    # Stripe
    stripe_secret_key: str = ""
    stripe_webhook_secret: str = ""
    stripe_basic_price_id: str = ""
    stripe_basic_yearly_price_id: str = ""
    stripe_pro_price_id: str = ""
    stripe_pro_yearly_price_id: str = ""
    
    # Billing
    currency: str = "USD"
    billing_day: int = 1  # Day of month for billing cycle
    
    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"


settings = Settings()
