"""
Pydantic models for Kafka messages and internal data structures.
"""

from datetime import datetime
from typing import Optional
from pydantic import BaseModel, Field


class PageContent(BaseModel):
    """Content extracted from a crawled page."""
    html: str = ""  # Gzip + Base64 encoded HTML
    text: str = ""  # Plain text from body
    title: str = ""  # Page title
    content_type: str = "text/html"  # MIME type
    encoding: str = "gzip+base64"  # e.g., "gzip+base64" or "plain"
    html_size: int = 0  # Original HTML size in bytes


class IndexingJobMessage(BaseModel):
    """
    Kafka message received from Go backend for document processing.
    Maps to IndexingJobMessage in Go indexer/types.go
    """
    job_id: int = Field(..., description="Unique job identifier")
    request_id: int = Field(..., description="Parent indexing request ID")
    project_id: int = Field(..., description="Project ID")
    user_id: str = Field(..., description="User ID who initiated the request")
    collection_id: str = Field(..., description="UUID for LightRAG workspace isolation")
    url: str = Field(..., description="URL of the crawled page")
    depth: int = Field(0, description="Crawl depth from base URL")
    parent_url: str = Field("", description="Parent page URL")
    
    # Content
    content: Optional[PageContent] = None
    
    # Metadata
    discovered_at: datetime = Field(default_factory=datetime.utcnow)
    metadata: dict = Field(default_factory=dict)


class JobStatusUpdate(BaseModel):
    """Request body for updating job status in Go backend."""
    status: str = Field(..., description="Job status: pending, in_progress, completed, failed")
    error_msg: Optional[str] = Field(None, description="Error message if status is failed")


class ProcessingResult(BaseModel):
    """Result of processing a single document."""
    job_id: int
    collection_id: str
    url: str
    success: bool
    error_message: Optional[str] = None
    
    # Metrics
    text_length: int = 0
    processing_time_ms: int = 0
