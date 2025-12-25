"""
HTTP client for updating job status in Go backend.
"""

import logging
from typing import Optional

import aiohttp

from app.config import Settings
from models.messages import JobStatusUpdate

logger = logging.getLogger(__name__)


class StatusClient:
    """
    HTTP client for communicating with Go backend.
    Used to update job statuses after processing.
    """
    
    def __init__(self, settings: Settings):
        self.settings = settings
        self.base_url = settings.go_backend_url.rstrip("/")
        self.timeout = aiohttp.ClientTimeout(total=settings.go_backend_timeout)
        self.internal_api_key = settings.internal_api_key
        self._session: Optional[aiohttp.ClientSession] = None
        
    async def _get_session(self) -> aiohttp.ClientSession:
        """Get or create HTTP session."""
        if self._session is None or self._session.closed:
            headers = {}
            if self.internal_api_key:
                headers["X-Internal-API-Key"] = self.internal_api_key
            self._session = aiohttp.ClientSession(
                timeout=self.timeout,
                headers=headers,
            )
        return self._session
    
    async def close(self) -> None:
        """Close HTTP session."""
        if self._session and not self._session.closed:
            await self._session.close()
            
    async def update_job_status(
        self,
        job_id: int,
        status: str,
        error_msg: Optional[str] = None,
    ) -> bool:
        """
        Update job status in Go backend.
        
        Args:
            job_id: Job ID to update
            status: New status (completed, failed)
            error_msg: Error message if status is failed
            
        Returns:
            True if update was successful
        """
        url = f"{self.base_url}/api/indexing/jobs/{job_id}/status"
        
        payload = JobStatusUpdate(
            status=status,
            error_msg=error_msg,
        )
        
        try:
            session = await self._get_session()
            
            async with session.put(url, json=payload.model_dump(exclude_none=True)) as resp:
                if resp.status == 200:
                    logger.debug(f"Updated job {job_id} status to {status}")
                    return True
                else:
                    response_text = await resp.text()
                    logger.error(
                        f"Failed to update job status",
                        extra={
                            "job_id": job_id,
                            "status_code": resp.status,
                            "response": response_text,
                        }
                    )
                    return False
                    
        except aiohttp.ClientError as e:
            logger.error(f"HTTP error updating job {job_id} status: {e}")
            return False
        except Exception as e:
            logger.error(f"Unexpected error updating job {job_id} status: {e}")
            return False
    
    async def increment_completed_jobs(self, request_id: int) -> bool:
        """
        Increment completed jobs counter for a request.
        
        Args:
            request_id: Request ID to update
            
        Returns:
            True if update was successful
        """
        url = f"{self.base_url}/api/indexing/requests/{request_id}/status"
        
        payload = {
            "increment_completed_jobs": True,
        }
        
        try:
            session = await self._get_session()
            
            async with session.put(url, json=payload) as resp:
                if resp.status == 200:
                    logger.debug(f"Incremented completed jobs for request {request_id}")
                    return True
                else:
                    response_text = await resp.text()
                    logger.error(
                        f"Failed to increment completed jobs",
                        extra={
                            "request_id": request_id,
                            "status_code": resp.status,
                            "response": response_text,
                        }
                    )
                    return False
                    
        except Exception as e:
            logger.error(f"Error incrementing completed jobs for request {request_id}: {e}")
            return False
    
    async def health_check(self) -> bool:
        """
        Check if Go backend is healthy.
        
        Returns:
            True if backend is healthy
        """
        url = f"{self.base_url}/health"
        
        try:
            session = await self._get_session()
            
            async with session.get(url) as resp:
                return resp.status == 200
                
        except Exception as e:
            logger.error(f"Health check failed: {e}")
            return False
