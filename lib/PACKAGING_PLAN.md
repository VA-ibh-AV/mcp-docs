# Packaging & Distribution Plan for mcp-docs

## Overview
This document outlines the approach for packaging `mcp-docs` as a distributable Python package that handles API keys, virtual environments, and server configuration elegantly.

## Problems Identified

1. **API Key Management**: Users need to provide OPENAI_API_KEY
2. **Virtual Environment**: Need to create venv on user's machine to run server
3. **Key Storage**: Need to store keys securely for the server

## Proposed Solutions

### 1. API Key Management

#### Approach: Multi-tier Key Management
- **Priority 1**: Environment variables (standard, secure)
- **Priority 2**: Project-specific `.env` files (convenient)
- **Priority 3**: Interactive CLI prompts (user-friendly)
- **Priority 4**: System keychain (optional, most secure)

#### Implementation:

**A. CLI Command for Key Setup**
```bash
mcp-docs configure --api-key <key>
# Or interactive:
mcp-docs configure
# Prompts: "Enter your OpenAI API key: [hidden input]"
```

**B. Key Storage Options:**
1. **Global config** (`~/.mcp-docs/config.json` or `~/.config/mcp-docs/config.json`)
2. **Project-specific** (`projects/<name>/.env` or `projects/<name>/config.json`)
3. **Environment variables** (highest priority, always checked first)

**C. Key Resolution Order:**
```python
1. Check environment variable: OPENAI_API_KEY
2. Check project .env file: projects/<name>/.env
3. Check global config: ~/.config/mcp-docs/config.json
4. Prompt user interactively (if --interactive flag)
```

### 2. Virtual Environment Handling

#### Approach: No Forced Venv Creation
- **Option A**: Run server in the same environment where package is installed
- **Option B**: Generate standalone server script with embedded dependencies check
- **Option C**: Use `pipx` for isolated installation (recommended for end users)

#### Implementation:

**A. Server Script Enhancement:**
- Add dependency check at startup
- Provide clear error messages if dependencies are missing
- Option to install missing dependencies automatically

**B. Installation Methods:**

1. **Standard pip install** (for developers):
   ```bash
   pip install mcp-docs
   mcp-docs start <project>
   ```

2. **pipx install** (recommended for end users, isolated):
   ```bash
   pipx install mcp-docs
   mcp-docs start <project>
   ```

3. **Standalone server generation** (for deployment):
   ```bash
   mcp-docs generate-server <project> --standalone
   # Creates a self-contained server.py with dependency checks
   ```

### 3. Key Storage for Server

#### Approach: Project-Level Configuration
- Store API keys in project directory (`.env` file)
- Never commit keys to git (already in .gitignore)
- Support both project-specific and global keys

#### Implementation:

**A. Project Structure:**
```
projects/
  <project-name>/
    project.json          # Public config (no secrets)
    .env                  # Secrets (OPENAI_API_KEY=...)
    server.py             # Generated server
    data/                 # ChromaDB data
    logs/                 # Logs
```

**B. Server Template Update:**
- Load API key from multiple sources (env var → project .env → global config)
- Provide clear error messages if key is missing
- Support key rotation without regenerating server

## Detailed Implementation Plan

### Phase 1: Configuration Management

#### 1.1 Create Config Module (`src/config.py`)
```python
- get_api_key() -> str | None  # Resolves key from all sources
- save_api_key(key: str, scope: 'global' | 'project') -> None
- load_project_config(project_name: str) -> dict
- save_project_config(project_name: str, config: dict) -> None
```

#### 1.2 Add CLI Commands
```bash
mcp-docs configure [--api-key KEY] [--global] [--project PROJECT]
mcp-docs config show [--project PROJECT]
mcp-docs config unset [--api-key] [--project PROJECT]
```

### Phase 2: Enhanced Server Generation

#### 2.1 Update Server Template
- Add key resolution logic (env → .env → config)
- Add dependency checking
- Add better error messages
- Support key from project .env file

#### 2.2 Server Startup Flow
```python
1. Check for OPENAI_API_KEY in environment
2. If not found, check projects/<name>/.env
3. If not found, check global config
4. If not found, show helpful error with setup instructions
5. Initialize OpenAI client
6. Start server
```

### Phase 3: Package Structure

#### 3.1 Update `pyproject.toml`
```toml
[project]
name = "mcp-docs"
version = "0.1.0"
description = "Turn any documentation site into an MCP server"
requires-python = ">=3.8"
dependencies = [
    "typer>=0.9.0",
    "chromadb>=0.5.5",
    "openai>=1.28.0",
    "playwright>=1.40.0",
    "beautifulsoup4>=4.12.3",
    "requests>=2.32.3",
    "python-dotenv>=1.0.0",  # For .env support
    "mcp>=0.1.0",
    # ... other deps
]

[project.scripts]
mcp-docs = "src.cli:APP"

[project.optional-dependencies]
dev = ["pytest", "black", "ruff"]
```

#### 3.2 Entry Point
- Move `cli.py` to `src/cli.py` or create `src/__main__.py`
- Add `mcp-docs` command via `[project.scripts]`

### Phase 4: Documentation

#### 4.1 Quick Start Guide
```markdown
# Installation
pip install mcp-docs
# or
pipx install mcp-docs

# Setup API Key
mcp-docs configure --api-key sk-...

# Create Project
mcp-docs add-project mydocs https://docs.example.com

# Index
mcp-docs index mydocs

# Start Server
mcp-docs start mydocs
```

#### 4.2 Configuration Guide
- Explain all key storage options
- Security best practices
- Troubleshooting guide

## File Structure Changes

```
mcp-docs/
├── src/
│   ├── __init__.py
│   ├── cli.py              # Main CLI (moved/refactored)
│   ├── config.py           # NEW: Config management
│   ├── indexer/
│   ├── embeddings/
│   ├── vectorstores/
│   └── utils/
├── pyproject.toml          # Updated with scripts
├── README.md               # Updated installation guide
├── setup.py                # Optional (if needed)
└── requirements.txt        # Keep for development
```

## Security Considerations

1. **Never commit secrets**: `.env` files already in `.gitignore`
2. **Key storage**: Use system keychain (keyring) as optional secure storage
3. **Permissions**: Ensure `.env` files have restricted permissions (600)
4. **Key rotation**: Support easy key updates without server regeneration

## Migration Path

1. **For existing users**: 
   - Keep current behavior (env vars)
   - Add new commands as opt-in
   - Backward compatible

2. **For new users**:
   - Interactive setup on first run
   - Guided configuration
   - Clear documentation

## Next Steps

1. ✅ Create `src/config.py` module
2. ✅ Add `configure` CLI command
3. ✅ Update server template with key resolution
4. ✅ Update `pyproject.toml` with entry points
5. ✅ Create comprehensive README
6. ✅ Add setup instructions
7. ✅ Test installation from PyPI (test PyPI first)

## Example Usage Flow

```bash
# 1. Install package
pip install mcp-docs

# 2. Configure (first time)
mcp-docs configure
# Interactive: "Enter OpenAI API key: [hidden]"
# Saves to ~/.config/mcp-docs/config.json

# 3. Create project
mcp-docs add-project react-docs https://react.dev

# 4. Index (uses stored API key)
mcp-docs index react-docs

# 5. Start server (uses stored API key)
mcp-docs start react-docs
```

## Alternative: Project-Specific Keys

```bash
# Store key per-project
mcp-docs configure --project react-docs --api-key sk-...

# This creates: projects/react-docs/.env
# Server automatically loads from there
```

