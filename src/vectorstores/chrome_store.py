# src/vectorstores/chrome_store.py
"""
Chroma Vector Store implementation for docs_mcp.

Uses Chroma's persistent client to store and retrieve embeddings.
"""

from __future__ import annotations
import chromadb
import hashlib
import os
from typing import List, Dict, Any, Optional

from .store import VectorStore


class ChromaStore(VectorStore):
    """
    Chroma-based vector store implementation.

    Example:
        store = ChromaStore(url="https://example.com/docs")
        store.add(
            ids=["doc1", "doc2"],
            documents=["text1", "text2"],
            embeddings=[[...], [...]],
            metadatas=[{"source": "page1"}, {"source": "page2"}]
        )
    """

    def __init__(self, url: str):
        """
        Initialize Chroma store with a URL.

        Args:
            url: Base URL for generating collection name and storage path.
        """
        collection_name = hashlib.md5(url.encode()).hexdigest()

        # Store the index in a relative ./indexes/<collection_name> directory
        # based on the current working directory. This makes it easy to keep
        # indexes next to the project while you iterate locally.
        path = os.path.abspath(os.path.join(os.getcwd(), "indexes", collection_name))
        os.makedirs(path, exist_ok=True)

        self.client = chromadb.PersistentClient(path=path)
        self.collection = self.client.get_or_create_collection(collection_name)
        self.url = url
        self.collection_name = collection_name

    def info(self) -> Dict:
        """Return store metadata."""
        return {
            "name": "chroma",
            "collection": self.collection_name,
            "url": self.url,
        }

    def add(
        self,
        ids: List[str],
        documents: List[str],
        embeddings: List[List[float]],
        metadatas: List[Dict[str, Any]],
    ) -> None:
        """
        Add documents with embeddings to the store.

        Args:
            ids: List of unique document IDs.
            documents: List of document texts.
            embeddings: List of embedding vectors.
            metadatas: List of metadata dictionaries.
        """
        self.collection.add(
            ids=ids,
            documents=documents,
            embeddings=embeddings,
            metadatas=metadatas,
        )

    def query(
        self,
        query_embedding: List[float],
        n_results: int = 5,
    ) -> Dict[str, Any]:
        """
        Query the store for similar documents.

        Args:
            query_embedding: Query embedding vector.
            n_results: Number of results to return.

        Returns:
            Dictionary with 'ids', 'distances', 'documents', and 'metadatas'.
        """
        results = self.collection.query(
            query_embeddings=[query_embedding],
            n_results=n_results,
        )
        return results

    def delete(self, ids: List[str]) -> None:
        """
        Delete documents from the store.

        Args:
            ids: List of document IDs to delete.
        """
        self.collection.delete(ids=ids)

    def clear(self) -> None:
        """
        Clear all documents from the store.
        """
        # Get all IDs and delete them
        all_items = self.collection.get()
        if all_items["ids"]:
            self.collection.delete(ids=all_items["ids"])
