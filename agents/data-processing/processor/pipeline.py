"""
Document processing pipeline.
Handles the complete flow from Kafka message to LightRAG storage.
"""

import asyncio
import logging
import time
from typing import Optional

from app.config import Settings
from models.messages import IndexingJobMessage, ProcessingResult
from processor.html_cleaner import decompress_html, clean_html_to_text
from rag.pool import LightRAGPool
from observability.metrics import Metrics

logger = logging.getLogger(__name__)


class DocumentProcessor:
    """
    Processes indexing jobs from Kafka and stores them in LightRAG.
    
    Pipeline:
    1. Decompress HTML (if compressed)
    2. Clean HTML to text (or use pre-extracted text)
    3. Insert into LightRAG with collection_id as workspace
    
    Note: We don't update individual job status to avoid excessive DB calls.
    Progress is tracked in-memory via metrics.
    """
    
    def __init__(
        self,
        settings: Settings,
        rag_pool: LightRAGPool,
        metrics: Optional[Metrics] = None,
    ):
        self.settings = settings
        self.rag_pool = rag_pool
        self.metrics = metrics
        self._semaphore = asyncio.Semaphore(settings.max_workers)
        
        # In-memory metrics
        self._jobs_processed = 0
        self._jobs_failed = 0
        
    async def process(self, job: IndexingJobMessage) -> ProcessingResult:
        """
        Process a single indexing job.
        
        Uses semaphore to limit concurrent processing.
        
        Args:
            job: Indexing job message from Kafka
            
        Returns:
            ProcessingResult with success/failure status
        """
        async with self._semaphore:
            return await self._process_internal(job)
    
    async def _process_internal(self, job: IndexingJobMessage) -> ProcessingResult:
        """Internal processing logic."""
        start_time = time.time()
        
        logger.info(
            f"Processing job {job.job_id}",
            extra={
                "job_id": job.job_id,
                "collection_id": job.collection_id,
                "url": job.url,
            }
        )
        
        try:
            # 1. Extract text content
            text = await self._extract_text(job)
            
            if not text:
                raise ValueError("No text content extracted from document")
            
            # 2. Get LightRAG instance for this collection
            rag = await self.rag_pool.get(job.collection_id)
            
            # 3. Insert into LightRAG
            # Use URL as document identifier for deduplication
            doc_id = job.url
            
            logger.debug(f"Inserting document into LightRAG: {doc_id}")
            await rag.ainsert(doc_id, text)
            
            processing_time_ms = int((time.time() - start_time) * 1000)
            self._jobs_processed += 1
            
            logger.info(
                f"Successfully processed job {job.job_id}",
                extra={
                    "job_id": job.job_id,
                    "collection_id": job.collection_id,
                    "url": job.url,
                    "text_length": len(text),
                    "processing_time_ms": processing_time_ms,
                    "total_processed": self._jobs_processed,
                }
            )
            
            return ProcessingResult(
                job_id=job.job_id,
                collection_id=job.collection_id,
                url=job.url,
                success=True,
                text_length=len(text),
                processing_time_ms=processing_time_ms,
            )
            
        except Exception as e:
            processing_time_ms = int((time.time() - start_time) * 1000)
            error_msg = str(e)
            self._jobs_failed += 1
            
            logger.error(
                f"Failed to process job {job.job_id}: {error_msg}",
                extra={
                    "job_id": job.job_id,
                    "collection_id": job.collection_id,
                    "url": job.url,
                    "error": error_msg,
                    "total_failed": self._jobs_failed,
                }
            )
            
            return ProcessingResult(
                job_id=job.job_id,
                collection_id=job.collection_id,
                url=job.url,
                success=False,
                error_message=error_msg,
                processing_time_ms=processing_time_ms,
            )
    
    async def _extract_text(self, job: IndexingJobMessage) -> str:
        """
        Extract text content from job.
        
        Prefers pre-extracted text from Go backend.
        Falls back to HTML cleaning if needed.
        
        Args:
            job: Indexing job with content
            
        Returns:
            Extracted text content
        """
        if not job.content:
            raise ValueError("Job has no content")
        
        # Prefer pre-extracted text (already cleaned by Go backend)
        if job.content.text:
            logger.debug(f"Using pre-extracted text for job {job.job_id}")
            return job.content.text
        
        # Fall back to HTML cleaning
        if job.content.html:
            logger.debug(f"Extracting text from HTML for job {job.job_id}")
            
            # Decompress if needed
            html = decompress_html(
                job.content.html,
                job.content.encoding,
            )
            
            # Clean HTML to text
            text = clean_html_to_text(html)
            return text
        
        raise ValueError("Job has neither text nor HTML content")
    
    @property
    def active_workers(self) -> int:
        """Return number of currently processing workers."""
        return self.settings.max_workers - self._semaphore._value
    
    @property
    def stats(self) -> dict:
        """Return processing statistics."""
        return {
            "jobs_processed": self._jobs_processed,
            "jobs_failed": self._jobs_failed,
            "active_workers": self.active_workers,
        }
