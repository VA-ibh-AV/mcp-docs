# MCP-Docs — turn any docs site into an MCP server

Lightweight, extensible Python package that scrapes any documentation site, normalizes pages, builds searchable indexes (BM25 + vector search), and exposes an MCP-style API so LLMs / agent crews can query docs with provenance.

> Note: this project intentionally does not include a hosted embedding service. Supported embedding providers: OpenAI, HuggingFace (inference API), and local sentence-transformers. Add more providers via the EmbeddingProvider interface.

## Features (short)

-   Scrape & normalize docs (Docusaurus, GitBook, generic static HTML, GitHub READMEs)
-   Store canonical page JSONs in SQLite/DuckDB
-   Keyword search (BM25) + semantic search using embeddings + vector DB
-   Extensible embedding providers (OpenAI, HF, local ST)
-   VectorStore adapters (Qdrant, Weaviate, Chroma, Pinecone, Redis)
-   FastAPI MCP endpoints (list_pages, get_page, search, examples, ask_docs, …)
-   Typer-based CLI (`mcp-docs`) for lifecycle: init, add, scrape, index, serve, search, export
-   Designed with SOLID principles — easy to extend & test
-   Local-first defaults so you can run without external keys

## Quickstart (5 minutes)

POSIX / macOS / WSL

```bash
# 1. clone
git clone https://github.com/yourorg/mcp-docs.git
cd mcp-docs

# 2. create virtualenv and install (editable)
python -m venv .venv
source .venv/bin/activate
pip install -e ".[dev]"

# 3. init project
mcp-docs init --dir my-mcp
cd my-mcp

# 4. add a docs site
mcp-docs add https://react.dev --name react-docs --adapter auto

# 5. scrape
mcp-docs scrape react-docs --concurrency 6

# 6. index (semantic using local sentence-transformers by default)
mcp-docs index react-docs --method hybrid --embed-provider local_st

# 7. serve
mcp-docs serve --port 8080
# Open http://localhost:8080/docs for FastAPI swagger UI
```

Windows (PowerShell)

```powershell
# 1. clone
git clone https://github.com/yourorg/mcp-docs.git
cd mcp-docs

# 2. create virtualenv and install (editable)
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -e ".[dev]"

# proceed as above
```

## Installation

Install from PyPI (when published) or directly from source:

```bash
# from PyPI
pip install mcp-docs

# or from git (dev)
pip install -e "git+https://github.com/yourorg/mcp-docs.git#egg=mcp-docs"

# development extras (tests, linting)
pip install -e ".[dev]"
```

## Example `mcp.yaml` (project config)

```yaml
project:
    name: my-mcp
    data_dir: data
    db: sqlite
    db_path: data/mcp.db

embeddings:
    provider: local_st # openai | hf | local_st
    providers:
        openai:
            api_key_env: OPENAI_API_KEY
            model: text-embedding-3-small
        hf:
            api_token_env: HF_API_TOKEN
            model: sentence-transformers/all-MiniLM-L6-v2
        local_st:
            model: all-MiniLM-L6-v2
            cache_dir: ~/.cache/mcp-docs

vector_store:
    provider: qdrant # qdrant | weaviate | chroma | pinecone | redis
    qdrant:
        url: 'http://localhost:6333'
        collection: 'default'
```

Important: do not store API keys in the config file. Use environment variables (OPENAI_API_KEY, HF_API_TOKEN) or a secret manager.

## CLI reference (high level)

```
mcp-docs init [--dir DIR] [--force]
mcp-docs add <URL> [--name NAME] [--adapter ADAPTER]
mcp-docs remove <name_or_id> [--purge]
mcp-docs list
mcp-docs scrape <site> [--concurrency N] [--depth D]
mcp-docs index <site> [--method bm25|embed|hybrid] [--embed-provider local_st|openai|hf]
mcp-docs serve [--host HOST] [--port PORT] [--site SITE] [--reload]
mcp-docs search <query> [--site SITE] [--topk N] [--semantic]
mcp-docs get <slug_or_url> [--site SITE] [--format json|html|text]
mcp-docs examples <query> [--language LANG] [--site SITE]
mcp-docs summarize <slug_or_url> [--site SITE] [--length short|medium|long]
mcp-docs compare <slug1> <slug2> [--site SITE]
mcp-docs export <site> --format [json|ndjson|duckdb|sqlite] --out PATH
mcp-docs config [--show|--edit]
mcp-docs version
```

Use `--json` on most commands for machine-friendly output.

## Package API (Python)

Programmatic usage via the `MCP` class:

```python
from mcp_docs import MCP

m = MCP.load_project("path/to/my-mcp")
m.add_site("https://react.dev", name="react-docs")
m.scrape("react-docs")
m.index("react-docs", method="hybrid", embed_provider="local_st")

# query programmatically
results = m.search("useMemo vs useCallback", topk=8, semantic=True)
page = m.get_page("/learn/state")
examples = m.find_code_examples("debounce", language="js")
answer = m.ask_docs("How do I debounce in React?", site="react-docs", topk=6)
```

### How `ask_docs` works (RAG workflow)

1. Embed the user query via the configured `EmbeddingProvider`.
2. Use the configured `VectorStore` to perform ANN query (top-k).
3. Optionally rerank candidates (cross-encoder) or re-score with BM25 hybrid.
4. Build a provenance-aware context (snippets + page metadata).
5. Call the user-configured LLM with a RAG prompt and return `{answer, citations}`.

This workflow is modular: embedding, vector store, and LLM calls are pluggable via interfaces so you can swap providers without changing retrieval logic.

## Key code snippets

EmbeddingProvider interface (ABC)

```python
# mcp_docs/embeddings.py
from abc import ABC, abstractmethod
from typing import List, Dict

class EmbeddingProvider(ABC):
    @abstractmethod
    def embed(self, texts: List[str]) -> List[List[float]]:
        """Return list of vectors for given texts."""
        raise NotImplementedError

    @abstractmethod
    def embed_batch(self, texts: List[str], batch_size: int = 32) -> List[List[float]]:
        """Optional optimized batch method."""
        raise NotImplementedError

    @abstractmethod
    def info(self) -> Dict:
        """Return metadata: name, model, dims."""
        raise NotImplementedError
```

Provider registry (factory)

```python
# mcp_docs/provider_registry.py
from typing import Dict, Type
from .embeddings import EmbeddingProvider

class ProviderRegistry:
    def __init__(self):
        self._impls: Dict[str, Type[EmbeddingProvider]] = {}

    def register(self, name: str, impl_cls: Type[EmbeddingProvider]):
        self._impls[name] = impl_cls

    def get(self, name: str, **kwargs) -> EmbeddingProvider:
        impl = self._impls.get(name)
        if not impl:
            raise KeyError(f"Embedding provider '{name}' not registered")
        return impl(**kwargs)
```

VectorStore abstraction

```python
# mcp_docs/vector_store.py
from abc import ABC, abstractmethod
from typing import List, Dict

class VectorStore(ABC):
    @abstractmethod
    def upsert(self, ids: List[str], vectors: List[List[float]], metas: List[Dict]):
        pass

    @abstractmethod
    def query(self, vector: List[float], topk: int = 10) -> List[Dict]:
        """Return list of hits: {id, score, meta, payload_text}"""
        pass

    @abstractmethod
    def delete_collection(self):
        pass
```

Concrete adapters implement Qdrant, Weaviate, Chroma, etc.

Example: simple `ask_docs` sketch

```python
# mcp_docs/ask.py
def ask_docs(mcp: MCP, question: str, site: str, topk=8, embed_provider="local_st"):
    provider = mcp.provider_registry.get(embed_provider)
    q_vec = provider.embed([question])[0]
    store = mcp.vectorstore_for_site(site)
    hits = store.query(q_vec, topk=topk)

    # build context
    context = "\n\n".join([f"{h['meta']['title']}\n{h['payload_text']}" for h in hits])

    # call user-configured LLM (mcp.llm.ask(...)) with prompt + context
    answer = mcp.llm.ask(question, context)
    return {"answer": answer, "sources": [h['meta'] for h in hits]}
```

## Storage & Indexing notes

Pages are stored in canonical JSON with fields: `id`, `url`, `slug`, `title`, `headings`, `content_text`, `content_html`, `code_blocks`, `last_scraped`.

Indexing pipeline:

-   Clean and chunk page text (configurable chunk size & overlap).
-   Compute embeddings for chunks via `EmbeddingProvider`.
-   Upsert vectors into `VectorStore` with chunk-level metadata (page id, heading, offset).
-   Optionally build a BM25 index for hybrid retrieval.

## Running a local vector DB (Qdrant example)

docker-compose.yml (quick snippet):

```yaml
version: '3.8'
services:
    qdrant:
        image: qdrant/qdrant:latest
        ports:
            - '6333:6333'
        volumes:
            - qdrant_data:/qdrant/storage
volumes:
    qdrant_data:
```

Start:

```bash
docker compose up -d
```

Then set `vector_store.provider = qdrant` and `url: http://localhost:6333` in `mcp.yaml`.

## Security & ethics

-   Respect `robots.txt` and the target site's terms of service.
-   Never commit API keys to the repo. Use environment variables.
-   Secure MCP servers serving private docs (TLS, API keys, internal network).
-   Provide clear user consent and privacy statements before offering optional hosted services.

## Testing & CI

-   Unit tests mock embedding and vector store providers.
-   Integration tests use Dockerized Qdrant/Weaviate and small static docs fixtures.

Run tests:

```bash
pytest
```

## Roadmap

-   Add more adapters (Docusaurus/Next.js docs app optimizations)
-   Cross-encoder re-ranker option
-   Plugin system for custom normalizers
-   Official PyPI release & GitHub Actions for builds
-   Explorer UI (optional web-based dashboard)

## Contributing

Contributions are welcome — open issues for feature requests and bug reports. PRs should include tests and adhere to code style (black + isort + flake8).

## License

MIT License — see `LICENSE` file.
