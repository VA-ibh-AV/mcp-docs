# src/docs_mcp/embeddings/provider.py
"""
EmbeddingProvider ABC for docs_mcp.

All embedding backends (OpenAI, HuggingFace, Local models, etc.)
must implement this interface so the rest of the system works
without caring about vendor differences.

The goal is strict SOLID:
- Interface Segregation
- Open/Closed (add providers w/o modifying core)
- Dependency Inversion
"""

from __future__ import annotations
from abc import ABC, abstractmethod
from typing import List, Dict, Optional


class EmbeddingProvider(ABC):
    """
    Abstract base class defining the interface for embedding providers.
    """

    @abstractmethod
    def info(self) -> Dict:
        """
        Return metadata about the embedding provider.
        Example:
        {
            "name": "openai",
            "model": "text-embedding-3-small",
            "batch_size": 128
        }
        """
        raise NotImplementedError

    @abstractmethod
    def embed(self, texts: List[str]) -> List[List[float]]:
        """
        Generate embeddings for a list of texts.
        Must preserve input ordering.

        Args:
            texts (List[str]): List of strings to embed.

        Returns:
            List[List[float]]: List of embedding vectors.
        """
        raise NotImplementedError

    @abstractmethod
    def embed_batch(
        self,
        texts: List[str],
        batch_size: Optional[int] = None
    ) -> List[List[float]]:
        """
        Optional batching wrapper. Default behavior can reuse embed(), but
        providers can override for performance.

        Args:
            texts: text list
            batch_size: override batch size for this call

        Returns:
            list of embedding vectors
        """
        raise NotImplementedError
