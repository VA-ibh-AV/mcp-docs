"""
Configuration management for mcp-docs.

Handles API key resolution from multiple sources:
1. Environment variables (highest priority)
2. Project-specific .env files
3. Global config file
4. Interactive prompts (fallback)
"""

import os
import json
from pathlib import Path
from typing import Optional
import getpass

# Try to import python-dotenv for .env file support
try:
    from dotenv import load_dotenv
    DOTENV_AVAILABLE = True
except ImportError:
    DOTENV_AVAILABLE = False


def _get_config_dir() -> Path:
    """Get the global config directory."""
    if os.name == 'nt':  # Windows
        config_dir = Path(os.environ.get('APPDATA', Path.home() / 'AppData' / 'Roaming'))
    else:  # Unix-like
        config_dir = Path.home() / '.config'
    
    config_dir = config_dir / 'mcp-docs'
    config_dir.mkdir(parents=True, exist_ok=True)
    return config_dir


def _get_global_config_path() -> Path:
    """Get path to global config file."""
    return _get_config_dir() / 'config.json'


def _get_project_env_path(project_name: str, projects_dir: Path) -> Path:
    """Get path to project-specific .env file."""
    return projects_dir / project_name / '.env'


def get_api_key(project_name: Optional[str] = None, projects_dir: Optional[Path] = None) -> Optional[str]:
    """
    Resolve OpenAI API key from multiple sources in priority order.
    
    Priority:
    1. Environment variable OPENAI_API_KEY
    2. Project-specific .env file (if project_name provided)
    3. Global config file
    
    Args:
        project_name: Optional project name to check project-specific config
        projects_dir: Optional path to projects directory
    
    Returns:
        API key string or None if not found
    """
    # Priority 1: Environment variable
    api_key = os.getenv("OPENAI_API_KEY")
    if api_key:
        return api_key
    
    # Priority 2: Project-specific .env file
    if project_name and projects_dir:
        env_path = _get_project_env_path(project_name, projects_dir)
        if env_path.exists() and DOTENV_AVAILABLE:
            load_dotenv(env_path, override=False)  # Don't override existing env vars
            api_key = os.getenv("OPENAI_API_KEY")
            if api_key:
                return api_key
    
    # Priority 3: Global config file
    config_path = _get_global_config_path()
    if config_path.exists():
        try:
            with open(config_path, 'r', encoding='utf-8') as f:
                config = json.load(f)
                api_key = config.get('openai_api_key') or config.get('OPENAI_API_KEY')
                if api_key:
                    return api_key
        except (json.JSONDecodeError, IOError):
            pass
    
    return None


def save_api_key(api_key: str, scope: str = 'global', project_name: Optional[str] = None, 
                 projects_dir: Optional[Path] = None) -> Path:
    """
    Save API key to specified location.
    
    Args:
        api_key: The API key to save
        scope: 'global' or 'project'
        project_name: Required if scope is 'project'
        projects_dir: Required if scope is 'project'
    
    Returns:
        Path to the file where key was saved
    """
    if scope == 'project':
        if not project_name or not projects_dir:
            raise ValueError("project_name and projects_dir required for project scope")
        
        env_path = _get_project_env_path(project_name, projects_dir)
        env_path.parent.mkdir(parents=True, exist_ok=True)
        
        # Read existing .env if it exists
        existing_vars = {}
        if env_path.exists():
            if DOTENV_AVAILABLE:
                from dotenv import dotenv_values
                existing_vars = dotenv_values(env_path)
        
        # Update with new API key
        existing_vars['OPENAI_API_KEY'] = api_key
        
        # Write .env file
        with open(env_path, 'w', encoding='utf-8') as f:
            for key, value in existing_vars.items():
                f.write(f"{key}={value}\n")
        
        # Set file permissions to 600 (read/write for owner only)
        os.chmod(env_path, 0o600)
        
        return env_path
    
    else:  # global
        config_path = _get_global_config_path()
        config = {}
        
        # Read existing config if it exists
        if config_path.exists():
            try:
                with open(config_path, 'r', encoding='utf-8') as f:
                    config = json.load(f)
            except (json.JSONDecodeError, IOError):
                pass
        
        # Update with new API key
        config['openai_api_key'] = api_key
        
        # Write config file
        with open(config_path, 'w', encoding='utf-8') as f:
            json.dump(config, f, indent=2)
        
        # Set file permissions to 600
        os.chmod(config_path, 0o600)
        
        return config_path


def prompt_api_key() -> str:
    """
    Interactively prompt user for API key.
    
    Returns:
        The API key entered by user
    """
    print("OpenAI API key is required for embedding generation.")
    print("You can get your API key from: https://platform.openai.com/api-keys")
    print()
    
    while True:
        api_key = getpass.getpass("Enter your OpenAI API key (input is hidden): ").strip()
        if api_key:
            if not api_key.startswith('sk-'):
                print("⚠ Warning: API key should start with 'sk-'. Continue anyway? (y/n): ", end='')
                response = input().strip().lower()
                if response != 'y':
                    continue
            return api_key
        else:
            print("API key cannot be empty. Please try again.")


def get_or_prompt_api_key(project_name: Optional[str] = None, 
                          projects_dir: Optional[Path] = None,
                          interactive: bool = True) -> Optional[str]:
    """
    Get API key from config or prompt user if not found.
    
    Args:
        project_name: Optional project name
        projects_dir: Optional projects directory
        interactive: If True, prompt user if key not found
    
    Returns:
        API key or None if not found and not interactive
    """
    api_key = get_api_key(project_name, projects_dir)
    
    if not api_key and interactive:
        api_key = prompt_api_key()
        if api_key:
            # Save to the most appropriate location
            if project_name and projects_dir:
                save_api_key(api_key, scope='project', project_name=project_name, projects_dir=projects_dir)
                print(f"✓ API key saved to project configuration")
            else:
                save_api_key(api_key, scope='global')
                print(f"✓ API key saved to global configuration")
    
    return api_key


def load_project_config(project_name: str, projects_dir: Path) -> dict:
    """Load project configuration from project.json."""
    config_path = projects_dir / project_name / "project.json"
    if not config_path.exists():
        raise FileNotFoundError(f"Project '{project_name}' not found. Run 'add-project' first.")
    
    with open(config_path, 'r', encoding='utf-8') as f:
        return json.load(f)


def save_project_config(project_name: str, projects_dir: Path, config: dict) -> None:
    """Save project configuration to project.json."""
    config_path = projects_dir / project_name / "project.json"
    config_path.parent.mkdir(parents=True, exist_ok=True)
    
    with open(config_path, 'w', encoding='utf-8') as f:
        json.dump(config, f, indent=2)

