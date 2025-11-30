# Project Structure

This document describes the organization of the MCP-Docs platform codebase.

## Directory Structure

```
mcp-docs/
├── lib/                    # Shared Python library code
│   ├── embeddings/         # Embedding providers (OpenAI, Azure)
│   ├── indexer/            # Document indexing logic
│   ├── utils/             # Utility functions
│   ├── vectorstores/      # Vector database integrations
│   ├── cli.py             # CLI interface
│   ├── config.py          # Configuration management
│   ├── test_lib.py        # Test script for library functionality
│   ├── pyproject.toml     # Python package configuration
│   ├── requirements.txt   # Python dependencies
│   ├── readme.md         # Library documentation
│   ├── test/              # Test directory
│   ├── dist/              # Distribution files
│   └── mcp_docs.egg-info/ # Package metadata
│
├── frontend/              # Next.js frontend application
│   └── (to be initialized)
│
├── go-backend/            # Go backend API server
│   └── (to be initialized)
│
├── agents/                # Python agent workers
│   └── test_server.py     # MCP server test/example
│
├── DESIGN_DOCUMENT.md     # Full-stack design document
├── UI_DESIGN_README.md    # UI design specifications
├── PROJECT_STRUCTURE.md   # This file
└── LICENSE                # License file
```

## Directory Descriptions

### `lib/`
Contains all reusable Python library code that can be shared across:
- CLI tools (`main.py`)
- Agent workers (`agents/`)
- Future backend services

**Key Modules:**
- `embeddings/` - Embedding provider implementations (OpenAI, Azure OpenAI)
- `indexer/` - Document scraping, chunking, and indexing logic
- `vectorstores/` - Vector database integrations (ChromaDB, etc.)
- `utils/` - Shared utility functions (URL filtering, etc.)

### `frontend/`
Next.js 14+ application with App Router. Will contain:
- Landing page
- User dashboard
- Project management UI
- Subscription management
- MCP endpoint management

**Technology Stack:**
- Next.js 14+ (App Router)
- shadcn/ui components
- Tailwind CSS
- TypeScript
- lucide-react icons

### `go-backend/`
Go backend API server. Will provide:
- REST API endpoints
- WebSocket server for real-time updates
- Authentication & authorization
- Database access (PostgreSQL)
- Kafka integration
- Vector DB client

**Technology Stack:**
- Go 1.21+
- Gin or Echo framework
- GORM or sqlx
- PostgreSQL
- Kafka client

### `agents/`
Python agent workers that process background jobs:
- Indexing workers (consume Kafka messages)
- MCP server agents (manage MCP server lifecycle)
- Usage tracking workers

**Files:**
- `test_server.py` - Example MCP server implementation

## Import Paths

After restructuring, all imports use the `lib.` prefix:

```python
# Example imports
from lib.embeddings.openai_provider import OpenAIEmbeddingProvider
from lib.indexer import index_documentation
from lib.vectorstores.chrome_store import ChromaStore
from lib.utils.url_filter import is_relevant_docs_url
```

## Development Workflow

1. **Library Code** (`lib/`): Shared across all Python components
2. **Frontend** (`frontend/`): Independent Next.js application
3. **Backend** (`go-backend/`): Independent Go service
4. **Agents** (`agents/`): Python workers that use `lib/` code

## Next Steps

1. Initialize Next.js project in `frontend/`
2. Initialize Go project in `go-backend/`
3. Set up agent workers in `agents/` (beyond test_server.py)
4. Update build/deployment scripts for new structure

