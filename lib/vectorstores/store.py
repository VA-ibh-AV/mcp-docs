# src/vectorstores/store.py
"""
VectorStore ABC for docs_mcp.

All vector store backends (Chroma, Pinecone, Weaviate, etc.)
must implement this interface so the rest of the system works
without caring about vendor differences.

The goal is strict SOLID:
- Interface Segregation
- Open/Closed (add stores w/o modifying core)
- Dependency Inversion
"""

from __future__ import annotations
from abc import ABC, abstractmethod
from typing import List, Dict, Optional, Any


class VectorStore(ABC):
    """
    Abstract base class defining the interface for vector stores.
    """

    @abstractmethod
    def info(self) -> Dict:
        """
        Return metadata about the vector store.
        Example:
        {
            "name": "chroma",
            "collection": "docs_collection",
        }
        """
        raise NotImplementedError

    @abstractmethod
    def add(
        self,
        ids: List[str],
        documents: List[str],
        embeddings: List[List[float]],
        metadatas: List[Dict[str, Any]],
    ) -> None:
        """
        Add documents with embeddings to the vector store.

        Args:
            ids: List of unique document IDs.
            documents: List of document texts.
            embeddings: List of embedding vectors.
            metadatas: List of metadata dictionaries for each document.
        """
        raise NotImplementedError

    @abstractmethod
    def query(
        self,
        query_embedding: List[float],
        n_results: int = 5,
    ) -> Dict[str, Any]:
        """
        Query the vector store for similar documents.

        Args:
            query_embedding: Query embedding vector.
            n_results: Number of results to return.

        Returns:
            Dictionary with 'ids', 'distances', 'documents', and 'metadatas'.
        """
        raise NotImplementedError

    @abstractmethod
    def delete(self, ids: List[str]) -> None:
        """
        Delete documents from the vector store by IDs.

        Args:
            ids: List of document IDs to delete.
        """
        raise NotImplementedError

    @abstractmethod
    def clear(self) -> None:
        """
        Clear all documents from the vector store.
        """
        raise NotImplementedError
