# src/docs_mcp/embeddings/openai_provider.py
"""
OpenAI Embedding Provider for docs_mcp.

Implements a simple EmbeddingProvider for OpenAI's Embeddings API.
- Uses environment variable OPENAI_API_KEY by default.
- Configurable model (default: text-embedding-3-small).
- Provides embed() and embed_batch() methods.
- Small retry/backoff for transient network/errors.
"""

from __future__ import annotations
import os
import time
from typing import List, Dict, Optional

try:
    from openai import OpenAI
except Exception as e:
    raise RuntimeError(
        "openai package is required for OpenAI embeddings. "
        "Install with `pip install openai`."
    ) from e

# Import the EmbeddingProvider ABC from your package
# Adjust import path if your package layout differs.
from .provider import EmbeddingProvider


class OpenAIEmbeddingProvider(EmbeddingProvider):
    """
    OpenAI embedding provider.

    Example:
        provider = OpenAIEmbeddingProvider(model="text-embedding-3-small", api_key=None)
        vecs = provider.embed(["hello", "world"])
    """

    def __init__(
        self,
        api_key: Optional[str] = None,
        model: str = "text-embedding-3-small",
        max_retries: int = 3,
        backoff_base: float = 1.0,
        timeout: int = 30,
        batch_size: int = 128,
    ):
        api_key = api_key or os.environ.get("OPENAI_API_KEY")
        if not api_key:
            raise ValueError(
                "OpenAI API key not provided. Set OPENAI_API_KEY env var or pass api_key."
            )

        self.client = OpenAI(api_key=api_key)
        self.model = model
        self.max_retries = max_retries
        self.backoff_base = backoff_base
        self.timeout = timeout
        self.batch_size = batch_size

    def info(self) -> Dict:
        """Return provider metadata."""
        return {
            "name": "openai",
            "model": self.model,
            "batch_size": self.batch_size,
        }

    def _call_embeddings(self, texts: List[str]) -> List[List[float]]:
        """Call OpenAI embeddings endpoint with simple retries & backoff."""
        attempt = 0
        while True:
            try:
                response = self.client.embeddings.create(
                    model=self.model,
                    input=texts,
                    encoding_format="float",
                )
                # response.data is a list of embeddings aligned with inputs
                vectors = [item.embedding for item in response.data]
                return vectors
            except Exception as e:
                attempt += 1
                if attempt > self.max_retries:
                    raise
                sleep_for = self.backoff_base * (2 ** (attempt - 1))
                time.sleep(sleep_for)
                print(f"Retrying OpenAI embeddings (attempt {attempt}) after error: {e}")

    def embed(self, texts: List[str]) -> List[List[float]]:
        """
        Embed a list of texts and return list of vectors (preserves order).

        This method will batch inputs to respect provider limits and to improve throughput.
        """
        if not isinstance(texts, list):
            raise TypeError("texts must be a list of strings")

        vectors: List[List[float]] = []
        n = len(texts)
        if n == 0:
            return vectors

        # Batch the requests
        for i in range(0, n, self.batch_size):
            batch = texts[i : i + self.batch_size]
            batch_vecs = self._call_embeddings(batch)
            # Basic validation
            if not isinstance(batch_vecs, list) or len(batch_vecs) != len(batch):
                raise RuntimeError("OpenAI returned unexpected embedding shape")
            vectors.extend(batch_vecs)

        return vectors

    def embed_batch(self, texts: List[str], batch_size: Optional[int] = None) -> List[List[float]]:
        """
        Alias for embed with configurable batch_size for this call.
        """
        if batch_size is not None:
            original = self.batch_size
            try:
                self.batch_size = batch_size
                return self.embed(texts)
            finally:
                self.batch_size = original
        return self.embed(texts)
