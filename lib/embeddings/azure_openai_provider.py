# src/embeddings/azure_openai_provider.py
"""
Azure OpenAI Embedding Provider for docs_mcp.

Implements EmbeddingProvider for Azure OpenAI's Embeddings API.
- Requires Azure OpenAI endpoint URL, deployment ID, and API key
- Compatible with OpenAI's embedding models deployed on Azure
"""

from __future__ import annotations
import os
import time
from typing import List, Dict, Optional

try:
    from openai import AzureOpenAI
except Exception as e:
    raise RuntimeError(
        "openai package is required for Azure OpenAI embeddings. "
        "Install with `pip install openai`."
    ) from e

from .provider import EmbeddingProvider


class AzureOpenAIEmbeddingProvider(EmbeddingProvider):
    """
    Azure OpenAI embedding provider.
    
    Example:
        provider = AzureOpenAIEmbeddingProvider(
            api_key="your-key",
            endpoint="https://your-resource.openai.azure.com",
            deployment_id="text-embedding-ada-002",
            api_version="2024-02-15-preview"
        )
        vecs = provider.embed(["hello", "world"])
    """

    def __init__(
        self,
        api_key: Optional[str] = None,
        endpoint: Optional[str] = None,
        deployment_id: Optional[str] = None,
        api_version: str = "2024-02-15-preview",
        max_retries: int = 3,
        backoff_base: float = 1.0,
        timeout: int = 30,
        batch_size: int = 128,
    ):
        # Get from environment variables if not provided
        api_key = api_key or os.environ.get("AZURE_OPENAI_API_KEY")
        endpoint = endpoint or os.environ.get("AZURE_OPENAI_ENDPOINT")
        deployment_id = deployment_id or os.environ.get("AZURE_OPENAI_DEPLOYMENT_ID")
        
        if not api_key:
            raise ValueError(
                "Azure OpenAI API key not provided. Set AZURE_OPENAI_API_KEY env var or pass api_key."
            )
        if not endpoint:
            raise ValueError(
                "Azure OpenAI endpoint not provided. Set AZURE_OPENAI_ENDPOINT env var or pass endpoint."
            )
        if not deployment_id:
            raise ValueError(
                "Azure OpenAI deployment ID not provided. Set AZURE_OPENAI_DEPLOYMENT_ID env var or pass deployment_id."
            )
        
        # Ensure endpoint doesn't have trailing slash
        endpoint = endpoint.rstrip('/')
        
        self.client = AzureOpenAI(
            api_key=api_key,
            azure_endpoint=endpoint,
            api_version=api_version,
        )
        self.deployment_id = deployment_id
        self.api_version = api_version
        self.max_retries = max_retries
        self.backoff_base = backoff_base
        self.timeout = timeout
        self.batch_size = batch_size

    def info(self) -> Dict:
        """Return provider metadata."""
        return {
            "name": "azure-openai",
            "deployment_id": self.deployment_id,
            "endpoint": self.client._client._base_url.__str__().replace('/openai', ''),
            "api_version": self.api_version,
            "batch_size": self.batch_size,
        }

    def _call_embeddings(self, texts: List[str]) -> List[List[float]]:
        """Call Azure OpenAI embeddings endpoint with simple retries & backoff."""
        attempt = 0
        while True:
            try:
                response = self.client.embeddings.create(
                    model=self.deployment_id,  # Azure uses deployment_id instead of model name
                    input=texts,
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
                print(f"Retrying Azure OpenAI embeddings (attempt {attempt}) after error: {e}")

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
                raise RuntimeError("Azure OpenAI returned unexpected embedding shape")
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

