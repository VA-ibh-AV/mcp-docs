from .scrapper import scrape_site
from .cleaner import clean_page_text
from .chunker import chunk_text
from .embedder import embed_chunks
from .db_writer import save_vector_db


def index_documentation(base_url: str, provider, output_dir: str | None = None,
                        max_pages: int = 50, max_depth: int = 5):
    """High-level indexer entrypoint.

    - Scrapes the given base_url
    - Cleans and chunks the pages
    - Uses the provided EmbeddingProvider to embed chunks
    - Persists embeddings into the project's VectorStore implementation
    """

    print("📥 Scraping...")
    raw_pages = scrape_site(base_url, max_pages, max_depth)

    print("🧹 Cleaning...")
    cleaned = [clean_page_text(p["text"]) for p in raw_pages]

    print("✂️ Chunking...")
    chunks = []
    for page in cleaned:
        chunks.extend(chunk_text(page))

    if not chunks:
        print("No content found to index.")
        return None

    print("🧠 Embedding...")
    embeddings = embed_chunks(provider, chunks)

    print("💾 Saving DB using configured VectorStore...")
    collection = save_vector_db(base_url, chunks, embeddings, output_dir)

    print(f"✅ Index built successfully: {getattr(collection, 'collection_name', output_dir)}")
    return collection
