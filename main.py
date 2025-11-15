"""Simple CLI for building docs index using the project's indexer, embeddings and vectorstore.

This replaces the previous crew-based entrypoint and directly calls
`indexer.index_documentation` using an EmbeddingProvider implementation.
"""

import sys
import os

# Ensure the package src/ is on the import path so relative imports work.
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "src"))

from dotenv import load_dotenv
load_dotenv()
from indexer import index_documentation

from embeddings.openai_provider import OpenAIEmbeddingProvider



def main():
    url = os.environ.get("DOCS_URL", "https://docs.crewai.com/en/introduction")

    if OpenAIEmbeddingProvider is None:
        print("OpenAI embedding provider is not available. Install 'openai' package or check imports.")
        return 2

    try:
        provider = OpenAIEmbeddingProvider(api_key=os.environ.get("OPENAI_API_KEY"))
    except ValueError as e:
        print(f"OpenAI provider initialization failed: {e}")
        print("Set OPENAI_API_KEY environment variable or provide a valid api_key.")
        return 3

    print(f"Starting documentation indexing for: {url}")
    print("-" * 50)

    collection = index_documentation(url, provider,max_pages=100)

    print("-" * 50)
    if collection is not None:
        print("Indexing completed!")
        # Run a sample query against the newly-built index
        try:
            query_text = "how to create a crew"
            print(f"\nRunning sample query: '{query_text}'")
            q_vecs = provider.embed([query_text])
            if not q_vecs:
                print("No embedding returned for query; skipping search.")
            else:
                q_vec = q_vecs[0]
                results = collection.query(query_embedding=q_vec, n_results=5)

                # Chroma returns lists-of-lists when querying multiple queries.
                # Normalize fields to lists of results for the single query.
                def _first_or_list(x):
                    if x is None:
                        return []
                    if isinstance(x, list) and len(x) > 0 and isinstance(x[0], list):
                        return x[0]
                    return x

                ids = _first_or_list(results.get("ids"))
                docs = _first_or_list(results.get("documents"))
                dists = _first_or_list(results.get("distances"))
                metas = _first_or_list(results.get("metadatas"))

                if not docs:
                    print("No documents returned for the query.")
                else:
                    for i, doc in enumerate(docs):
                        meta = metas[i] if metas and i < len(metas) else {}
                        dist = dists[i] if dists and i < len(dists) else None
                        idv = ids[i] if ids and i < len(ids) else None
                        print("\n---")
                        print(f"Rank {i+1} (id={idv}, distance={dist}):")
                        print(doc[:1000])
                        if meta:
                            print(f"metadata: {meta}")
        except Exception as e:
            print(f"Sample query failed: {e}")
    else:
        print("Indexing finished with no results.")

    return 0


if __name__ == "__main__":
    exit_code = main()
    sys.exit(exit_code)
