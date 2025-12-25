"""
Configuration settings for the Python RAG Agent.
Uses pydantic-settings for environment variable management.
"""

from pydantic_settings import BaseSettings, SettingsConfigDict
from pydantic import Field
from functools import lru_cache


class Settings(BaseSettings):
    """Application settings loaded from environment variables."""
    
    # Kafka Configuration
    kafka_bootstrap_servers: str = "localhost:9092"
    kafka_group_id: str = "rag-processor"
    kafka_topic: str = "indexing_jobs"
    kafka_auto_offset_reset: str = "earliest"
    kafka_enable_auto_commit: bool = False
    
    # Processing Configuration
    max_workers: int = 5  # Concurrent message processors
    max_rag_instances: int = 10  # Max LightRAG pools in memory
    rag_instance_ttl: int = 300  # TTL for inactive RAG instances (seconds)
    
    # Retry Configuration
    max_retries: int = 3
    retry_initial_backoff: float = 1.0
    retry_max_backoff: float = 30.0
    retry_backoff_multiplier: float = 2.0
    
    # PostgreSQL Configuration (for LightRAG)
    postgres_host: str = "localhost"
    postgres_port: int = 5432
    postgres_user: str = "postgres"
    postgres_password: str = "postgres"
    postgres_db: str = Field(default="mcpdocs", validation_alias="POSTGRES_DB")
    
    # OpenAI Configuration
    openai_api_key: str = ""
    openai_embedding_model: str = "text-embedding-3-small"
    openai_llm_model: str = "gpt-4o-mini"
    
    # Go Backend Configuration (for status updates)
    go_backend_url: str = "http://localhost:8005"
    go_backend_timeout: int = 30  # seconds
    internal_api_key: str = ""  # Internal API key for service-to-service auth
    
    # Logging
    log_level: str = "INFO"
    
    @property
    def postgres_dsn(self) -> str:
        """Generate PostgreSQL connection string."""
        return f"postgresql://{self.postgres_user}:{self.postgres_password}@{self.postgres_host}:{self.postgres_port}/{self.postgres_db}"
    
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
        extra="ignore",  # Ignore extra environment variables
    )


@lru_cache()
def get_settings() -> Settings:
    """Get cached settings instance."""
    return Settings()
