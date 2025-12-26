"""
LightRAG instance pool with LRU eviction.
Manages LightRAG instances per collection_id (workspace).
"""

import asyncio
import logging
import time
from collections import OrderedDict
from dataclasses import dataclass, field
from typing import Optional

from lightrag import LightRAG
from lightrag.llm.openai import gpt_4o_mini_complete, openai_embed

from app.config import Settings

logger = logging.getLogger(__name__)


@dataclass
class LightRAGWrapper:
    """Wrapper for LightRAG instance with metadata."""
    rag: LightRAG
    collection_id: str
    created_at: float = field(default_factory=time.time)
    last_used: float = field(default_factory=time.time)
    

class LightRAGPool:
    """
    Pool of LightRAG instances with LRU eviction.
    
    Each collection_id maps to a separate LightRAG workspace,
    providing isolation for different indexing requests.
    """
    
    def __init__(self, settings: Settings):
        self.settings = settings
        self.instances: OrderedDict[str, LightRAGWrapper] = OrderedDict()
        self.max_instances = settings.max_rag_instances
        self.ttl_seconds = settings.rag_instance_ttl
        self._lock = asyncio.Lock()
        self._initialized = False
        
    async def get(self, collection_id: str) -> LightRAG:
        """
        Get or create a LightRAG instance for the given collection_id.
        Uses LRU eviction when max_instances is reached.
        
        Args:
            collection_id: UUID identifying the LightRAG workspace
            
        Returns:
            Initialized LightRAG instance
        """
        async with self._lock:
            if collection_id in self.instances:
                # Move to end (most recently used)
                self.instances.move_to_end(collection_id)
                wrapper = self.instances[collection_id]
                wrapper.last_used = time.time()
                logger.debug(f"Reusing existing LightRAG instance for collection: {collection_id}")
                return wrapper.rag
            
            # Evict oldest instances if at capacity
            while len(self.instances) >= self.max_instances:
                oldest_id, oldest_wrapper = self.instances.popitem(last=False)
                logger.info(f"Evicting LightRAG instance for collection: {oldest_id}")
                try:
                    await oldest_wrapper.rag.finalize_storages()
                except Exception as e:
                    logger.warning(f"Error finalizing evicted RAG instance: {e}")
            
            # Create new instance
            logger.info(f"Creating new LightRAG instance for collection: {collection_id}")
            rag = await self._create_instance(collection_id)
            self.instances[collection_id] = LightRAGWrapper(
                rag=rag,
                collection_id=collection_id,
            )
            return rag
    
    async def _create_instance(self, collection_id: str) -> LightRAG:
        """
        Create and initialize a new LightRAG instance.
        
        Args:
            collection_id: UUID to use as workspace name
            
        Returns:
            Initialized LightRAG instance
        """
        # Set environment variables for LightRAG PostgreSQL connection
        import os
        os.environ.setdefault("POSTGRES_HOST", self.settings.postgres_host)
        os.environ.setdefault("POSTGRES_PORT", str(self.settings.postgres_port))
        os.environ.setdefault("POSTGRES_USER", self.settings.postgres_user)
        os.environ.setdefault("POSTGRES_PASSWORD", self.settings.postgres_password)
        os.environ.setdefault("POSTGRES_DB", self.settings.postgres_db)
        
        # Prefix workspace with 'ws_' to ensure valid PostgreSQL/AGE graph names.
        # AGE graph names cannot start with a number, and UUIDs may start with digits.
        workspace_name = f"ws_{collection_id.replace('-', '_')}"
        
        rag = LightRAG(
            embedding_func=openai_embed,
            llm_model_func=gpt_4o_mini_complete,
            kv_storage="PGKVStorage",
            vector_storage="PGVectorStorage",
            graph_storage="PGGraphStorage",
            doc_status_storage="PGDocStatusStorage",
            workspace=workspace_name,
        )
        
        # Initialize storage backends
        await rag.initialize_storages()
        
        logger.info(f"LightRAG instance initialized for workspace: {collection_id}")
        return rag
    
    async def release(self, collection_id: str) -> None:
        """
        Explicitly release a LightRAG instance.
        
        Args:
            collection_id: Collection to release
        """
        async with self._lock:
            if collection_id in self.instances:
                wrapper = self.instances.pop(collection_id)
                try:
                    await wrapper.rag.finalize_storages()
                    logger.info(f"Released LightRAG instance for collection: {collection_id}")
                except Exception as e:
                    logger.warning(f"Error finalizing RAG instance: {e}")
    
    async def cleanup_expired(self) -> int:
        """
        Remove instances that haven't been used within TTL.
        
        Returns:
            Number of instances cleaned up
        """
        async with self._lock:
            now = time.time()
            expired = [
                cid for cid, wrapper in self.instances.items()
                if now - wrapper.last_used > self.ttl_seconds
            ]
            
            for collection_id in expired:
                wrapper = self.instances.pop(collection_id)
                try:
                    await wrapper.rag.finalize_storages()
                    logger.info(f"Cleaned up expired LightRAG instance: {collection_id}")
                except Exception as e:
                    logger.warning(f"Error finalizing expired RAG instance: {e}")
            
            return len(expired)
    
    async def close_all(self) -> None:
        """Close all LightRAG instances gracefully."""
        async with self._lock:
            logger.info(f"Closing {len(self.instances)} LightRAG instances...")
            
            for collection_id, wrapper in list(self.instances.items()):
                try:
                    await wrapper.rag.finalize_storages()
                    logger.debug(f"Closed LightRAG instance: {collection_id}")
                except Exception as e:
                    logger.warning(f"Error closing RAG instance {collection_id}: {e}")
            
            self.instances.clear()
            logger.info("All LightRAG instances closed")
    
    @property
    def active_instances(self) -> int:
        """Return number of active LightRAG instances."""
        return len(self.instances)
    
    def get_stats(self) -> dict:
        """Return pool statistics."""
        return {
            "active_instances": len(self.instances),
            "max_instances": self.max_instances,
            "ttl_seconds": self.ttl_seconds,
            "collections": list(self.instances.keys()),
        }
