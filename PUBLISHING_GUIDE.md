# Publishing mcp-docs to PyPI

## Prerequisites

1. **PyPI Account**: Create accounts on both TestPyPI and PyPI
   - TestPyPI: https://test.pypi.org/account/register/
   - PyPI: https://pypi.org/account/register/

2. **Install build tools**:
   ```bash
   pip install build twine
   ```

3. **API Tokens** (recommended over passwords):
   - Go to https://test.pypi.org/manage/account/token/ (for TestPyPI)
   - Go to https://pypi.org/manage/account/token/ (for PyPI)
   - Create a new API token with scope: "Entire account" or "Project: mcp-docs"

## Step-by-Step Publishing Process

### Step 1: Prepare the Package

1. **Update version** in `pyproject.toml`:
   ```toml
   version = "0.1.0"  # Update for each release
   ```

2. **Ensure all files are ready**:
   - ✅ `pyproject.toml` - Package configuration
   - ✅ `readme.md` - Package description
   - ✅ `LICENSE` - License file (create if missing)
   - ✅ `src/` - Source code
   - ✅ `.gitignore` - Excludes build artifacts

3. **Clean previous builds**:
   ```bash
   # Remove old build artifacts
   rm -rf build/ dist/ *.egg-info/
   # Or on Windows:
   rmdir /s /q build dist *.egg-info
   ```

### Step 2: Build the Package

```bash
# Build source distribution and wheel
python -m build
```

This creates:
- `dist/mcp-docs-0.1.0.tar.gz` (source distribution)
- `dist/mcp_docs-0.1.0-py3-none-any.whl` (wheel)

### Step 3: Test on TestPyPI (Recommended)

**Always test on TestPyPI first!**

1. **Upload to TestPyPI**:
   ```bash
   python -m twine upload --repository testpypi dist/*
   ```
   
   When prompted:
   - Username: `__token__`
   - Password: Your TestPyPI API token (starts with `pypi-`)

2. **Test installation from TestPyPI**:
   ```bash
   # Create a fresh virtual environment
   python -m venv test_env
   test_env\Scripts\activate  # Windows
   # or: source test_env/bin/activate  # Linux/Mac
   
   # Install from TestPyPI (CRITICAL: Use BOTH --index-url AND --extra-index-url)
   # --index-url: Look for mcp-docs on TestPyPI
   # --extra-index-url: Look for dependencies (typer, chromadb, etc.) on PyPI
   pip install --index-url https://test.pypi.org/simple/ --extra-index-url https://pypi.org/simple/ mcp-docs==0.1.1
   
   # Or using uv (same flags):
   uv pip install --index-url https://test.pypi.org/simple/ --extra-index-url https://pypi.org/simple/ mcp-docs==0.1.1
   
   # Test the command
   mcp-docs --help
   ```
   
   **⚠️ IMPORTANT**: 
   - DO NOT use `-i` alone (it only checks TestPyPI for everything)
   - MUST use `--index-url` + `--extra-index-url` together
   - TestPyPI doesn't have all packages, so dependencies come from PyPI
   - There may be a 1-2 minute delay after upload before package is available

3. **Verify everything works**:
   - Test `mcp-docs configure`
   - Test `mcp-docs add-project`
   - Test `mcp-docs index`
   - Test `mcp-docs start`

### Step 4: Publish to PyPI

Once tested on TestPyPI:

```bash
# Upload to real PyPI
python -m twine upload dist/*
```

When prompted:
- Username: `__token__`
- Password: Your PyPI API token (starts with `pypi-`)

### Step 5: Verify Publication

1. **Check PyPI page**: https://pypi.org/project/mcp-docs/
2. **Test installation**:
   ```bash
   pip install mcp-docs
   mcp-docs --help
   ```

## Updating the Package

For future releases:

1. **Update version** in `pyproject.toml`:
   ```toml
   version = "0.1.1"  # Increment version
   ```

2. **Update CHANGELOG** (if you have one)

3. **Build and upload**:
   ```bash
   python -m build
   python -m twine upload dist/*
   ```

## Version Numbering

Follow [Semantic Versioning](https://semver.org/):
- **MAJOR.MINOR.PATCH** (e.g., 1.2.3)
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes

Examples:
- `0.1.0` → `0.1.1` (bug fix)
- `0.1.1` → `0.2.0` (new feature)
- `0.2.0` → `1.0.0` (major release)

## Troubleshooting

### "Package already exists"
- Version already published. Increment version number.

### "Invalid distribution"
- Check `pyproject.toml` syntax
- Ensure all required fields are present

### "File not found"
- Make sure you're in the project root directory
- Run `python -m build` first

### Authentication errors
- Use API tokens instead of passwords
- Ensure token has correct scope
- Check token hasn't expired

### "No solution found when resolving dependencies" (TestPyPI)
- **Problem**: TestPyPI doesn't have all packages. When using `-i` alone, it only checks TestPyPI for everything.
- **Solution**: Use BOTH `--index-url` AND `--extra-index-url`:
  ```bash
  pip install --index-url https://test.pypi.org/simple/ --extra-index-url https://pypi.org/simple/ mcp-docs==0.1.1
  ```
  This tells pip to:
  1. Look for `mcp-docs` on TestPyPI (primary index)
  2. Look for dependencies (typer, chromadb, etc.) on PyPI (extra index)

### "Could not find a version that satisfies the requirement typer>=0.9.0"
- **Problem**: Using `-i` or `--index-url` alone makes pip only check TestPyPI for ALL packages
- **Solution**: Add `--extra-index-url https://pypi.org/simple/` to also check PyPI for dependencies

### "No version of mcp-docs==0.1.1" (after upload)
- **Problem**: Package may not be immediately available (1-2 minute delay)
- **Solution**: Wait a minute and try again, or use regular `pip` instead of `uv pip`

## Security Best Practices

1. **Never commit API tokens** to git
2. **Use API tokens** instead of passwords
3. **Use TestPyPI** for testing before real PyPI
4. **Set token scope** to specific project when possible
5. **Rotate tokens** periodically

## Additional Resources

- [PyPI User Guide](https://packaging.python.org/en/latest/guides/distributing-packages-using-setuptools/)
- [Twine Documentation](https://twine.readthedocs.io/)
- [Python Packaging Guide](https://packaging.python.org/)

## Quick Reference Commands

```bash
# Build package
python -m build

# Upload to TestPyPI
python -m twine upload --repository testpypi dist/*

# Upload to PyPI
python -m twine upload dist/*

# Check package contents
python -m twine check dist/*

# Install locally for testing
pip install -e .

# Install from TestPyPI (with PyPI fallback for dependencies)
pip install --index-url https://test.pypi.org/simple/ --extra-index-url https://pypi.org/simple/ mcp-docs

# Install from PyPI (after publishing)
pip install mcp-docs
```

