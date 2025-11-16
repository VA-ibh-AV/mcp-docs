#!/usr/bin/env python3
"""
CLI for MCP docs projects.

Commands:
  - add-project <name> <url>                    Create a new project
  - index <project_name> [--max-pages N] [--max-depth N]  Index documentation
  - start <project_name> [--port PORT]         Start MCP server
  - configure [options]                         Configure API keys
  - list                                        List all projects
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
from src.config import (
    get_api_key, save_api_key, get_or_prompt_api_key,
    load_project_config, save_project_config
)


APP = typer.Typer()
ROOT = Path.cwd()
PROJECTS_DIR = ROOT / "projects"
PROJECTS_DIR.mkdir(exist_ok=True)


def _project_dir(name: str) -> Path:
    return PROJECTS_DIR / name


def _project_config_path(name: str) -> Path:
    return _project_dir(name) / "project.json"


def _save_config(name: str, data: dict):
    """Save project config using the config module."""
    save_project_config(name, PROJECTS_DIR, data)


def _load_config(name: str) -> dict:
    """Load project config using the config module."""
    return load_project_config(name, PROJECTS_DIR)


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

# Resolve API key from multiple sources (env var -> project .env -> global config)
import sys
OPENAI_API_KEY = os.getenv("OPENAI_API_KEY")

# Try loading from project .env file
if not OPENAI_API_KEY:
    env_path = os.path.join(PROJECT_DIR, ".env")
    if os.path.exists(env_path):
        try:
            from dotenv import load_dotenv
            load_dotenv(env_path, override=False)
            OPENAI_API_KEY = os.getenv("OPENAI_API_KEY")
        except ImportError:
            pass

# Try loading from global config
if not OPENAI_API_KEY:
    try:
        if os.name == 'nt':  # Windows
            config_dir = os.path.join(os.environ.get('APPDATA', os.path.expanduser('~')), 'mcp-docs')
        else:  # Unix-like
            config_dir = os.path.join(os.path.expanduser('~'), '.config', 'mcp-docs')
        global_config_path = os.path.join(config_dir, 'config.json')
        if os.path.exists(global_config_path):
            with open(global_config_path, 'r', encoding='utf-8') as f:
                global_config = json.load(f)
                OPENAI_API_KEY = global_config.get('openai_api_key') or global_config.get('OPENAI_API_KEY')
    except Exception:
        pass

# Initialize OpenAI
if not OPENAI_API_KEY:
    print("ERROR: OpenAI API key not found.", file=sys.stderr)
    print("Please set OPENAI_API_KEY environment variable or run 'mcp-docs configure'", file=sys.stderr)
    raise RuntimeError("OpenAI API key not provided. Run 'mcp-docs configure' to set it up.")

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
# Get port from environment variable if set
mcp_port = os.getenv("MCP_PORT")
if mcp_port:
    try:
        mcp = FastMCP("{project_name} MCP Server", port=int(mcp_port))
    except Exception:
        # If FastMCP is actually a Server class fallback, wrap minimal decorator
        mcp = FastMCP("{project_name} MCP Server", port=int(mcp_port))  # type: ignore
else:
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
    import sys
    port = os.getenv("MCP_PORT")
    print("Starting MCP server...", file=sys.stderr)
    print("Collection: " + COLLECTION_NAME, file=sys.stderr)
    print("ChromaDB path: " + CHROMA_PATH, file=sys.stderr)
    if port:
        print("Port: " + port, file=sys.stderr)
    print("Server ready. Waiting for requests...", file=sys.stderr)
    print("=" * 50, file=sys.stderr)
    mcp.run(transport="sse")


if __name__ == "__main__":
    main()
'''

@APP.command()
def add_project(name: str = typer.Argument(..., help="Project name"),
                url: str = typer.Argument(..., help="Root documentation URL")):
    """
    Create a new project folder and basic config.
    
    Examples:
        mcp-docs add-project react-docs https://react.dev
        mcp-docs add-project mydocs https://docs.example.com
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
    typer.echo(f"   1. Run 'mcp-docs configure' to set up your API key (if not already done)")
    typer.echo(f"   2. Run 'mcp-docs index {name}' to index the documentation")
    typer.echo(f"   3. Run 'mcp-docs start {name}' to start the MCP server")


@APP.command()
def index(project_name: str = typer.Argument(..., help="project name to index"),
          max_pages: int = typer.Option(200, "--max-pages", help="Maximum number of pages to scrape (default: 200)"),
          max_depth: int = typer.Option(5, "--max-depth", help="Maximum crawl depth (default: 5)")):
    """
    Run the indexer for the given project.
    
    Examples:
        mcp-docs index mydocs
        mcp-docs index mydocs --max-pages 100
        mcp-docs index mydocs --max-pages 500 --max-depth 3
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
    
    # Get API key using config module
    api_key = get_or_prompt_api_key(project_name=project_name, projects_dir=PROJECTS_DIR, interactive=True)
    if not api_key:
        typer.secho("Error: OpenAI API key is required.", fg=typer.colors.RED)
        typer.echo("Run 'mcp-docs configure' to set up your API key.")
        raise typer.Exit(1)
    
    # Create embedding provider (default to OpenAI)
    try:
        provider = OpenAIEmbeddingProvider(api_key=api_key)
        typer.echo(f"Using embedding provider: {provider.info()['name']} (model: {provider.info()['model']})")
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
          port: Optional[int] = typer.Option(None, "--port", help="Port for MCP server (optional, for SSE transport)")):
    """
    Generate the project server.py and start it as a subprocess.
    
    Examples:
        mcp-docs start mydocs
        mcp-docs start mydocs --port 8080
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
    # Pass port via environment variable if provided
    if port:
        env["MCP_PORT"] = str(port)
        typer.echo(f"Starting server on port {port}...")
    # keep existing env so OPENAI_API_KEY flows through if present
    try:
        # Use subprocess.run() to run in foreground/interactive mode
        subprocess.run(cmd, cwd=str(pdir), env=env)
    except KeyboardInterrupt:
        typer.echo("\n\nServer stopped by user.")
    except Exception as e:
        typer.echo(f"Failed to start server: {e}")
        raise typer.Exit(1)


@APP.command("list")
def list_projects():
    """
    List all available projects and their status.
    
    Shows all projects with their configuration, indexing status, and document counts.
    """
    if not PROJECTS_DIR.exists():
        typer.echo("No projects directory found.")
        return
    
    projects = [p for p in PROJECTS_DIR.iterdir() if p.is_dir() and (p / "project.json").exists()]
    
    if not projects:
        typer.echo("No projects found. Create one with 'mcp-docs add-project <name> <url>'")
        return
    
    typer.echo(f"Found {len(projects)} project(s):\n")
    typer.echo("=" * 70)
    
    for project_dir in sorted(projects):
        try:
            cfg = load_project_config(project_dir.name, PROJECTS_DIR)
            typer.echo(f"\n📁 Project: {cfg.get('name', project_dir.name)}")
            typer.echo(f"   URL: {cfg.get('url', 'N/A')}")
            typer.echo(f"   Collection: {cfg.get('collection_name', 'N/A')}")
            typer.echo(f"   Path: {project_dir}")
            
            # Check if indexed
            chroma_path = Path(cfg.get('chroma_path', ''))
            if chroma_path.exists():
                try:
                    import chromadb
                    client = chromadb.PersistentClient(path=str(chroma_path))
                    collection = client.get_collection(cfg.get('collection_name'))
                    count = collection.count()
                    typer.echo(f"   Status: ✓ Indexed ({count} documents)")
                except Exception:
                    typer.echo(f"   Status: ⚠ Index may be incomplete")
            else:
                typer.echo(f"   Status: ✗ Not indexed")
                
        except Exception as e:
            typer.echo(f"\n📁 Project: {project_dir.name}")
            typer.secho(f"   ⚠ Error loading config: {e}", fg=typer.colors.YELLOW)
    
    typer.echo("\n" + "=" * 70)


@APP.command()
def configure(
    api_key: Optional[str] = typer.Option(None, "--api-key", "-k", help="OpenAI API key to save"),
    project: Optional[str] = typer.Option(None, "--project", "-p", help="Save key for specific project"),
    global_scope: bool = typer.Option(False, "--global", "-g", help="Save to global config (default: project-specific)"),
    show: bool = typer.Option(False, "--show", help="Show current configuration"),
    unset: bool = typer.Option(False, "--unset", help="Remove stored API key")
):
    """
    Configure API keys for mcp-docs.
    
    Examples:
        mcp-docs configure                    # Interactive setup
        mcp-docs configure --api-key sk-...   # Save key directly
        mcp-docs configure --project mydocs   # Save for specific project
        mcp-docs configure --global           # Save to global config
        mcp-docs configure --show             # Show current config
        mcp-docs configure --unset            # Remove stored key
    """
    if show:
        # Show current configuration
        typer.echo("Current API key configuration:")
        typer.echo("=" * 50)
        
        # Check environment variable
        env_key = os.getenv("OPENAI_API_KEY")
        if env_key:
            typer.echo("✓ Environment variable: OPENAI_API_KEY (set)")
        else:
            typer.echo("✗ Environment variable: OPENAI_API_KEY (not set)")
        
        # Check global config
        global_key = get_api_key()
        if global_key and not env_key:  # Only show if not overridden by env
            typer.echo("✓ Global config: API key found")
        else:
            typer.echo("✗ Global config: No API key stored")
        
        # Check project configs
        if project:
            project_key = get_api_key(project, PROJECTS_DIR)
            if project_key and not env_key:
                typer.echo(f"✓ Project '{project}': API key found")
            else:
                typer.echo(f"✗ Project '{project}': No API key stored")
        
        return
    
    if unset:
        # Remove API key
        if project:
            env_path = PROJECTS_DIR / project / ".env"
            if env_path.exists():
                env_path.unlink()
                typer.echo(f"✓ Removed API key for project '{project}'")
            else:
                typer.echo(f"No API key found for project '{project}'")
        else:
            from src.config import _get_global_config_path
            config_path = _get_global_config_path()
            if config_path.exists():
                try:
                    with open(config_path, 'r', encoding='utf-8') as f:
                        config = json.load(f)
                    config.pop('openai_api_key', None)
                    config.pop('OPENAI_API_KEY', None)
                    with open(config_path, 'w', encoding='utf-8') as f:
                        json.dump(config, f, indent=2)
                    typer.echo("✓ Removed API key from global config")
                except Exception as e:
                    typer.echo(f"Error removing key: {e}")
            else:
                typer.echo("No global API key found")
        return
    
    # Save API key
    if api_key:
        key = api_key
    else:
        # Interactive prompt
        from src.config import prompt_api_key
        key = prompt_api_key()
    
    if not key:
        typer.echo("No API key provided.")
        raise typer.Exit(1)
    
    # Determine scope
    if project:
        scope = 'project'
        scope_name = f"project '{project}'"
    elif global_scope:
        scope = 'global'
        scope_name = "global configuration"
    else:
        # Default: project if specified, otherwise global
        if project:
            scope = 'project'
            scope_name = f"project '{project}'"
        else:
            scope = 'global'
            scope_name = "global configuration"
    
    # Save the key
    try:
        saved_path = save_api_key(key, scope=scope, project_name=project, projects_dir=PROJECTS_DIR)
        typer.echo(f"✓ API key saved to {scope_name}")
        typer.echo(f"  Location: {saved_path}")
    except Exception as e:
        typer.secho(f"Error saving API key: {e}", fg=typer.colors.RED)
        raise typer.Exit(1)


if __name__ == "__main__":
    APP()

