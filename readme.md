# MCP-Docs — Turn any docs site into an MCP server

A Python package that scrapes documentation sites, indexes them with embeddings, and exposes them as MCP (Model Context Protocol) servers for LLM integration.

## Features

- **Web Scraping**: Uses Playwright to scrape JavaScript-rendered documentation sites
- **Semantic Search**: Indexes documentation with OpenAI embeddings for semantic search
- **Vector Storage**: Uses ChromaDB for persistent vector storage
- **MCP Server**: Automatically generates and runs MCP servers for each documentation project
- **Project-Based**: Organizes documentation sites as separate projects
- **CLI Tool**: Simple command-line interface for managing projects

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourorg/mcp-docs.git
cd mcp-docs

# Create virtual environment
python -m venv .venv

# Activate virtual environment
# On Windows (PowerShell):
.\.venv\Scripts\Activate.ps1
# On macOS/Linux:
source .venv/bin/activate

# Install in editable mode
pip install -e ".[dev]"

# Install Playwright browsers
playwright install chromium
```

### Dependencies

The project requires:
- Python 3.8+
- OpenAI API key (for embeddings)
- Playwright (for web scraping)
- ChromaDB (for vector storage)

## Quick Start

### 1. Configure API Key

Set up your OpenAI API key:

```bash
mcp-docs configure
```

Or set it as an environment variable:

```bash
# Windows (PowerShell)
$env:OPENAI_API_KEY="sk-your-key-here"

# macOS/Linux
export OPENAI_API_KEY="sk-your-key-here"
```

### 2. Add a Documentation Project

Create a new project for a documentation site:

```bash
mcp-docs add-project my-docs https://docs.example.com
```

This creates a project directory at `projects/my-docs/` with:
- `project.json` - Project configuration
- `data/` - Data directory
- `logs/` - Log files directory

### 3. Index the Documentation

Scrape and index the documentation:

```bash
mcp-docs index my-docs --max-pages 200 --max-depth 5
```

This will:
- Scrape pages from the documentation site (up to `max-pages`)
- Clean and chunk the content
- Generate embeddings using OpenAI
- Store embeddings in ChromaDB

### 4. Start the MCP Server

Start the MCP server for your project:

```bash
mcp-docs start my-docs
```

The server will run and expose a `search_docs` tool that can be used by MCP clients.

## CLI Commands

### `add-project <name> <url>`

Create a new documentation project.

```bash
mcp-docs add-project react-docs https://react.dev
```

### `index <project_name> [options]`

Index a documentation project.

**Options:**
- `--max-pages <N>`: Maximum number of pages to scrape (default: 200)
- `--max-depth <N>`: Maximum crawl depth (default: 5)

```bash
mcp-docs index react-docs --max-pages 100 --max-depth 3
```

### `start <project_name> [options]`

Start the MCP server for a project.

**Options:**
- `--port <PORT>`: Port for MCP server (optional, depends on client)

```bash
mcp-docs start react-docs
```

### `configure [options]`

Configure API keys and settings.

**Options:**
- `--api-key <KEY>`: Set API key directly
- `--project <NAME>`: Configure for specific project
- `--global`: Save to global config (default)
- `--show`: Show current configuration
- `--unset`: Remove stored API key

```bash
# Interactive configuration
mcp-docs configure

# Set API key directly
mcp-docs configure --api-key sk-...

# Configure for specific project
mcp-docs configure --project my-docs --api-key sk-...

# Show current config
mcp-docs configure --show
```

## Project Structure

```
mcp-docs/
├── projects/
│   ├── my-docs/
│   │   ├── project.json      # Project configuration
│   │   ├── server.py          # Auto-generated MCP server
│   │   ├── data/              # Data directory
│   │   └── logs/              # Log files
│   └── ...
├── indexes/                    # ChromaDB indexes
│   └── <collection-hash>/
└── src/                        # Source code
    ├── cli.py                  # CLI implementation
    ├── config.py               # Configuration management
    ├── indexer/                # Indexing pipeline
    │   ├── scrapper.py         # Web scraper
    │   ├── cleaner.py          # Text cleaning
    │   ├── chunker.py          # Text chunking
    │   ├── embedder.py         # Embedding generation
    │   └── db_writer.py        # Database writing
    ├── embeddings/             # Embedding providers
    │   └── openai_provider.py  # OpenAI provider
    └── vectorstores/           # Vector store implementations
        └── chrome_store.py     # ChromaDB store
```

## Project Configuration

Each project has a `project.json` file:

```json
{
  "name": "my-docs",
  "url": "https://docs.example.com",
  "collection_name": "abc123...",
  "chroma_path": "/path/to/indexes/abc123..."
}
```

## API Key Configuration

API keys can be configured at multiple levels (in priority order):

1. **Environment variable**: `OPENAI_API_KEY`
2. **Project-specific**: `projects/<project>/.env`
3. **Global config**: `~/.config/mcp-docs/config.json` (Linux/macOS) or `%APPDATA%/mcp-docs/config.json` (Windows)

Use `mcp-docs configure` to manage API keys.

## How It Works

1. **Scraping**: Uses Playwright to render JavaScript-heavy documentation sites and extract content
2. **Cleaning**: Removes navigation, headers, and other non-content elements
3. **Chunking**: Splits content into manageable chunks for embedding
4. **Embedding**: Generates embeddings using OpenAI's `text-embedding-3-small` model
5. **Storage**: Stores embeddings and metadata in ChromaDB
6. **MCP Server**: Generates an MCP server with a `search_docs` tool for semantic search

## MCP Server Integration

The generated MCP server exposes a `search_docs` tool that:

- Accepts a text query or pre-computed embedding
- Returns top-k matching document chunks with metadata
- Provides semantic search over the indexed documentation

The server uses FastMCP and runs in SSE (Server-Sent Events) mode for compatibility with MCP clients.

## Development

### Running Tests

```bash
pytest
```

### Code Style

The project uses:
- `black` for code formatting
- `ruff` for linting

```bash
black src/
ruff check src/
```

## Requirements

- Python 3.8+
- OpenAI API key
- Playwright (with Chromium browser)
- ChromaDB

See `pyproject.toml` for complete dependency list.

## Limitations

- Currently supports OpenAI embeddings only
- Uses ChromaDB as the only vector store
- Requires JavaScript rendering (Playwright) for scraping
- No built-in BM25/hybrid search (semantic search only)

## Contributing

Contributions are welcome! Please:

1. Open an issue for feature requests or bug reports
2. Submit pull requests with tests
3. Follow code style (black + ruff)

## License

MIT License — see `LICENSE` file.
