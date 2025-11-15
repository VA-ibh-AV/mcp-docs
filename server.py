from typing import Any, List
import os
from mcp.server.fastmcp import FastMCP
import chromadb

from dotenv import load_dotenv
load_dotenv()

# Initialize FastMCP server
mcp = FastMCP("CrewAI MCP Docs")


def _find_first_collection_index_dir() -> str:
    """Return the absolute path and collection name for the first collection found in ./indexes.

    This helper picks the first directory under ./indexes and treats its name as the
    Chroma collection name (this matches how `ChromaStore` creates collections).
    """
    base = os.path.abspath(os.path.join(os.getcwd(), "indexes"))
    if not os.path.isdir(base):
        raise FileNotFoundError(f"indexes directory not found: {base}")

    for name in os.listdir(base):
        candidate = os.path.join(base, name)
        if os.path.isdir(candidate):
            return candidate, name

    raise FileNotFoundError(f"no collection directories found in indexes: {base}")


@mcp.tool()
async def search_docs(query: Any) -> Any:
    """Search the Chroma vector DB and return the top 20 responses.

    Args:
        query: either a string (text query) or a list/tuple of floats (pre-computed embedding).

    Returns:
        A dict with the raw results from Chroma (ids, distances, documents, metadatas),
        or an error dictionary if something goes wrong.
    """
    # Prepare embedding: accept raw embedding or compute using OpenAI provider if available.
    embedding: List[float]
    if isinstance(query, (list, tuple)):
        # Assume user passed a pre-computed embedding vector
        embedding = list(query)
    elif isinstance(query, str):
        # Try to compute embedding using OpenAI provider (if available/configured)
        try:
            from src.embeddings.openai_provider import OpenAIEmbeddingProvider

            provider = OpenAIEmbeddingProvider()  # will raise if API key not set
            vecs = provider.embed([query])
            if not vecs or not isinstance(vecs[0], list):
                return {"error": "embedding_failure", "details": "provider returned unexpected shape"}
            embedding = vecs[0]
        except Exception as e:
            return {"error": "embedding_error", "details": str(e)}
    else:
        return {"error": "invalid_query_type", "details": "query must be text or embedding list"}

    # Ensure chromadb is available
    if chromadb is None:
        return {"error": "chromadb_missing", "details": "chromadb package not available"}

    # Find the first collection under ./indexes and query it
    try:
        index_path, collection_name = _find_first_collection_index_dir()
    except Exception as e:
        return {"error": "no_index_found", "details": str(e)}

    try:
        client = chromadb.PersistentClient(path=index_path)
        collection = client.get_or_create_collection(collection_name)
        results = collection.query(query_embeddings=[embedding], n_results=20)
        return results
    except Exception as e:
        return {"error": "query_failed", "details": str(e)}


def main():
    # Initialize and run the server
    mcp.run(transport='sse')

if __name__ == "__main__":
    main()