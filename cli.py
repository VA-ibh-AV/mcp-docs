#!/usr/bin/env python3
"""
CLI for MCP docs projects.

Commands:
  - add_project <name> <url>
  - index <project_name>
  - start <project_name> [--port]
"""

from typing import Optional
import json
import os
import subprocess
import sys
from pathlib import Path

import typer

from src.indexer import index_documentation
from src.embeddings import OpenAIEmbeddingProvider


APP = typer.Typer()
ROOT = Path.cwd()
PROJECTS_DIR = ROOT / "projects"
PROJECTS_DIR.mkdir(exist_ok=True)


def _project_dir(name: str) -> Path:
    return PROJECTS_DIR / name


def _project_config_path(name: str) -> Path:
    return _project_dir(name) / "project.json"


def _save_config(name: str, data: dict):
    p = _project_dir(name)
    p.mkdir(parents=True, exist_ok=True)
    with open(_project_config_path(name), "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2)


def _load_config(name: str) -> dict:
    cfgp = _project_config_path(name)
    if not cfgp.exists():
        typer.echo(f"Project '{name}' not found. Please run `add_project` first.")
        raise typer.Exit(1)
    return json.loads(cfgp.read_text(encoding="utf-8"))


SERVER_TEMPLATE = r'''# Auto-generated MCP server for project: {project_name}
# Do not edit if you want to regenerate via the CLI. Generated from template.

import os
import json
import asyncio
from typing import Any, List

# MCP server import - adjust if your environment uses a different mcp package
try:
    from mcp.server.fastmcp import FastMCP
except Exception:
    # fallback if FastMCP isn't available
    from mcp.server import Server as FastMCP  # type: ignore

import chromadb
from openai import OpenAI
import numpy as np

PROJECT_DIR = os.path.dirname(__file__)
# Load config to get the correct chroma_path
CONFIG_PATH = os.path.join(PROJECT_DIR, "project.json")
with open(CONFIG_PATH, "r", encoding="utf-8") as f:
    config = json.load(f)
CHROMA_PATH = config.get("chroma_path", os.path.join(PROJECT_DIR, "data", "chroma"))
COLLECTION_NAME = config.get("collection_name", "{collection_name}")
OPENAI_API_KEY = os.getenv("OPENAI_API_KEY")

# initialize OpenAI
import sys
if not OPENAI_API_KEY:
    print("ERROR: OpenAI API key not provided. Set OPENAI_API_KEY env var.", file=sys.stderr)
    raise RuntimeError("OpenAI API key not provided. Set OPENAI_API_KEY env var.")

openai_client = OpenAI(api_key=OPENAI_API_KEY)
print("✓ OpenAI API key configured", file=sys.stderr)

# init chroma client
print("Initializing ChromaDB client at: " + CHROMA_PATH, file=sys.stderr)
client = chromadb.PersistentClient(path=CHROMA_PATH)
try:
    collection = client.get_collection(COLLECTION_NAME)
    print("✓ Loaded existing collection: " + COLLECTION_NAME, file=sys.stderr)
except Exception:
    # create empty collection if it doesn't exist
    collection = client.create_collection(name=COLLECTION_NAME)
    print("✓ Created new collection: " + COLLECTION_NAME, file=sys.stderr)

# create MCP server
try:
    mcp = FastMCP("{project_name} MCP Server")
except Exception:
    # If FastMCP is actually a Server class fallback, wrap minimal decorator
    mcp = FastMCP("{project_name} MCP Server")  # type: ignore


async def _embed_query(text: str) -> List[float]:
    """Embed a single text query using OpenAI's new API."""
    response = openai_client.embeddings.create(
        model="text-embedding-3-small",
        input=[text],
        encoding_format="float"
    )
    return response.data[0].embedding


@mcp.tool()
async def search_docs(query: Any, top_k: int = 5) -> Any:
    """{tool_description}
    
    Search the project's chroma collection.
    
    Args:
        query: Text query (string) or pre-computed embedding (list/tuple of floats)
        top_k: Number of results to return (default: 5)
    
    Returns:
        JSON-like dict with 'results': list of {{id, score, document, metadata}}
    """
    # compute embedding if query is text
    if isinstance(query, (list, tuple)):
        q_emb = np.array(query, dtype=float)
    elif isinstance(query, str):
        q_emb = np.array(await _embed_query(query), dtype=float)
    else:
        return {{"error": "invalid_query", "details": "Query must be text or embedding vector."}}

    # perform search via chroma collection
    try:
        # chroma expects nested lists for embeddings
        res = collection.query(query_embeddings=[q_emb.tolist()], n_results=top_k,
                               include=["documents", "metadatas", "distances"])
    except Exception as e:
        return {{"error": "chromadb_error", "details": str(e)}}

    # format results
    results = []
    # res fields: ids, distances, documents, metadatas
    ids = res.get("ids", [[]])[0]
    docs = res.get("documents", [[]])[0]
    metas = res.get("metadatas", [[]])[0]
    dists = res.get("distances", [[]])[0]

    for idx, doc, meta, dist in zip(ids, docs, metas, dists):
        results.append({{
            "id": idx,
            "score": float(dist),
            "document": doc,
            "metadata": meta
        }})

    return {{"results": results}}


def main():
    # run MCP as SSE server for VSCode integration / clients
    print("Starting MCP server...", file=sys.stderr)
    print("Collection: " + COLLECTION_NAME, file=sys.stderr)
    print("ChromaDB path: " + CHROMA_PATH, file=sys.stderr)
    print("Server ready. Waiting for requests...", file=sys.stderr)
    print("=" * 50, file=sys.stderr)
    mcp.run(transport="sse")


if __name__ == "__main__":
    main()
'''

@APP.command()
def add_project(name: str = typer.Argument(..., help="collection name"),
                url: str = typer.Argument(..., help="root docs URL")):
    """
    Create a new project folder and basic config.
    """
    import hashlib

    pdir = _project_dir(name)
    if pdir.exists():
        typer.echo(f"Project {name} already exists at {pdir}")
        raise typer.Exit(1)

    typer.echo(f"Creating project '{name}'...")

    # create structure
    (pdir / "data").mkdir(parents=True, exist_ok=True)
    (pdir / "logs").mkdir(parents=True, exist_ok=True)
    typer.echo(f"✓ Created project directory structure at {pdir}")

    # Calculate collection name the same way ChromaStore does (MD5 hash of URL)
    collection_name = hashlib.md5(url.encode()).hexdigest()
    
    # ChromaStore creates indexes in ./indexes/<collection_name> based on cwd
    # We need to use the same path structure
    indexes_dir = ROOT / "indexes" / collection_name
    chroma_path = str(indexes_dir)

    # Save minimal config
    cfg = {
        "name": name,
        "url": url,
        "collection_name": collection_name,
        "chroma_path": chroma_path
    }
    _save_config(name, cfg)
    typer.echo(f"✓ Saved project configuration")

    # create an empty chroma collection so start() can load
    try:
        import chromadb
        typer.echo(f"Initializing ChromaDB at {chroma_path}...")
        indexes_dir.mkdir(parents=True, exist_ok=True)
        client = chromadb.PersistentClient(path=chroma_path)
        # ensure collection exists (create if missing)
        try:
            collection = client.create_collection(name=collection_name)
            typer.echo(f"✓ Created ChromaDB collection '{collection_name}'")
        except Exception as e:
            # Check if collection already exists
            try:
                collection = client.get_collection(name=collection_name)
                typer.echo(f"✓ ChromaDB collection '{collection_name}' already exists")
            except Exception:
                typer.secho(f"⚠ Warning: Failed to create/get collection: {e}", fg=typer.colors.YELLOW)
    except ImportError:
        typer.secho("⚠ Warning: chromadb not installed. Install chromadb if you want local vectorstore.", fg=typer.colors.YELLOW)
        typer.secho("  Run: pip install chromadb", fg=typer.colors.YELLOW)
    except Exception as e:
        typer.secho(f"⚠ Warning: Failed to initialize ChromaDB: {e}", fg=typer.colors.YELLOW)

    typer.echo(f"\n✅ Project '{name}' added successfully!")
    typer.echo(f"   URL: {url}")
    typer.echo(f"   Project directory: {pdir}")
    typer.echo(f"   Collection name: {collection_name}")
    typer.echo(f"   ChromaDB path: {chroma_path}")
    typer.echo(f"\nNext steps:")
    typer.echo(f"   1. Run 'index {name}' to index the documentation")
    typer.echo(f"   2. Run 'start {name}' to start the MCP server")


@APP.command()
def index(project_name: str = typer.Argument(..., help="project name to index"),
          max_pages: int = typer.Option(200, help="max pages to scrape"),
          max_depth: int = typer.Option(5, help="max depth to crawl")):
    """
    Run the indexer for the given project.
    """
    cfg = _load_config(project_name)
    if index_documentation is None:
        typer.echo("Indexer not found (index_documentation import failed). Ensure your indexer is available at src.docs_mcp.indexer.index_documentation")
        raise typer.Exit(1)

    output_dir = Path(cfg["chroma_path"])
    url = cfg['url']
    
    # Warn if URL looks like a specific page rather than root docs
    from urllib.parse import urlparse
    parsed = urlparse(url)
    if parsed.query or len(parsed.path.split('/')) > 3:
        typer.secho(f"⚠ Warning: URL looks like a specific page. For better crawling, consider using the root docs URL.", fg=typer.colors.YELLOW)
        typer.secho(f"   Current: {url}", fg=typer.colors.YELLOW)
        typer.secho(f"   Suggested: {parsed.scheme}://{parsed.netloc}/docs", fg=typer.colors.YELLOW)
    
    typer.echo(f"Indexing project {project_name} from {url} into {output_dir}")
    
    # Create embedding provider (default to OpenAI)
    try:
        provider = OpenAIEmbeddingProvider()
        typer.echo(f"Using embedding provider: {provider.info()['name']} (model: {provider.info()['model']})")
    except ValueError as e:
        typer.secho(f"Error: {e}", fg=typer.colors.RED)
        typer.echo("Please set OPENAI_API_KEY environment variable or configure an embedding provider.")
        raise typer.Exit(1)
    except Exception as e:
        typer.secho(f"Failed to initialize embedding provider: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)
    
    # call your indexer (synchronous)
    try:
        index_documentation(cfg["url"], provider=provider, output_dir=str(output_dir),
                           max_pages=max_pages, max_depth=max_depth)
    except TypeError:
        # earlier index_documentation signature may differ; attempt a minimal call
        index_documentation(cfg["url"], str(output_dir))
    except Exception as e:
        typer.echo(f"Indexing failed: {e}")
        raise typer.Exit(1)

    typer.echo("Indexing complete.")


@APP.command()
def start(project_name: str = typer.Argument(..., help="project name to start"),
          port: Optional[int] = typer.Option(None, help="port used by some MCP clients (optional)")):
    """
    Generate the project server.py and start it as a subprocess.
    """
    cfg = _load_config(project_name)
    pdir = _project_dir(project_name)

    # Build tool description deterministically
    tool_description = (
        f"Semantic search tool for the '{project_name}' documentation (root URL: {cfg.get('url')}). "
        "Accepts a text query (string) or a pre-computed embedding (list/tuple of floats). "
        "Returns top matching document chunks from the indexed docs collection."
    )

    server_py = SERVER_TEMPLATE.format(
        project_name=project_name,
        collection_name=cfg.get("collection_name", project_name),
        tool_description=tool_description.replace('"', '\\"')
    )

    server_file = pdir / "server.py"
    server_file.write_text(server_py, encoding="utf-8")
    typer.echo(f"Generated server at {server_file}")
    typer.echo(f"Starting server in interactive mode...")
    typer.echo(f"Press Ctrl+C to stop the server.\n")

    # Start the server in interactive mode (foreground)
    cmd = [sys.executable, str(server_file)]
    env = os.environ.copy()
    # keep existing env so OPENAI_API_KEY flows through if present
    try:
        # Use subprocess.run() to run in foreground/interactive mode
        subprocess.run(cmd, cwd=str(pdir), env=env)
    except KeyboardInterrupt:
        typer.echo("\n\nServer stopped by user.")
    except Exception as e:
        typer.echo(f"Failed to start server: {e}")
        raise typer.Exit(1)


if __name__ == "__main__":
    APP()
