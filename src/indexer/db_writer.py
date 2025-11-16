# src/docs_mcp/indexer/db_writer.py

from src.vectorstores.chrome_store import ChromaStore
from typing import List


def save_vector_db(url: str, chunks: List[str], embeddings: List[List[float]], output_dir: str | None = None):
    """Persist chunks+embeddings to the project's ChromaStore.

    Args:
        url: base URL used to name/locate the collection (used by ChromaStore)
        chunks: list of document text chunks
        embeddings: list of embedding vectors matching chunks
        output_dir: optional override (not currently used by ChromaStore)

    Returns:
        The ChromaStore instance used to persist data.
    """
    store = ChromaStore(url)

    # If user passed an explicit output_dir, note it (not currently used by ChromaStore).
    if output_dir is not None:
        print(f"Note: output_dir parameter is ignored; ChromaStore derives storage path from the URL. Received: {output_dir}")

    ids = [str(i) for i in range(len(chunks))]
    metadatas = [{"source": url} for _ in chunks]

    store.add(
        ids=ids,
        documents=chunks,
        embeddings=embeddings,
        metadatas=metadatas,
    )

    return store
