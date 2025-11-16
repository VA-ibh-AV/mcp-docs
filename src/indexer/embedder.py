# src/docs_mcp/indexer/embedder.py

from src.embeddings.provider import EmbeddingProvider
from typing import List


def embed_chunks(provider: EmbeddingProvider, chunks: List[str]):
    """Embed a list of text chunks using the provided EmbeddingProvider.

    Args:
        provider: an object implementing EmbeddingProvider
        chunks: list of strings to embed

    Returns:
        List of embeddings (list of float lists) matching chunks order.
    """
    return provider.embed(chunks)
