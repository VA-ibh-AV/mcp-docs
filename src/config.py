"""
Configuration management for mcp-docs.

Handles embedding provider configuration and API key resolution from multiple sources:
1. Environment variables (highest priority)
2. Project-specific .env files
3. Global config file
4. Interactive prompts (fallback)

Supports multiple embedding providers:
- openai: OpenAI API
- azure-openai: Azure OpenAI
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


def get_provider_config(project_name: Optional[str] = None, projects_dir: Optional[Path] = None) -> Optional[dict]:
    """
    Get embedding provider configuration from multiple sources.
    
    Priority:
    1. Environment variables
    2. Project-specific .env file (if project_name provided)
    3. Global config file
    
    Args:
        project_name: Optional project name to check project-specific config
        projects_dir: Optional path to projects directory
    
    Returns:
        Dict with provider configuration or None if not found
    """
    config_path = _get_global_config_path()
    project_config = None
    
    # Priority 2: Project-specific .env file
    if project_name and projects_dir:
        env_path = _get_project_env_path(project_name, projects_dir)
        if env_path.exists() and DOTENV_AVAILABLE:
            load_dotenv(env_path, override=False)
        
        # Also check project.json for provider preference
        project_json = projects_dir / project_name / "project.json"
        if project_json.exists():
            try:
                with open(project_json, 'r', encoding='utf-8') as f:
                    project_config = json.load(f)
            except (json.JSONDecodeError, IOError):
                pass
    
    # Priority 3: Global config file
    global_config = {}
    if config_path.exists():
        try:
            with open(config_path, 'r', encoding='utf-8') as f:
                global_config = json.load(f)
        except (json.JSONDecodeError, IOError):
            pass
    
    # Determine provider (project config > global config > default to openai)
    provider_name = None
    if project_config and 'embedding_provider' in project_config:
        provider_name = project_config['embedding_provider']
    elif global_config.get('embedding_provider'):
        provider_name = global_config['embedding_provider']
    else:
        # Default to openai if no preference set
        provider_name = 'openai'
    
    # Build config based on provider
    if provider_name == 'azure-openai':
        config = {
            'provider': 'azure-openai',
            'api_key': os.getenv("AZURE_OPENAI_API_KEY") or project_config.get('azure_openai_api_key') or global_config.get('azure_openai_api_key'),
            'endpoint': os.getenv("AZURE_OPENAI_ENDPOINT") or project_config.get('azure_openai_endpoint') or global_config.get('azure_openai_endpoint'),
            'deployment_id': os.getenv("AZURE_OPENAI_DEPLOYMENT_ID") or project_config.get('azure_openai_deployment_id') or global_config.get('azure_openai_deployment_id'),
            'api_version': os.getenv("AZURE_OPENAI_API_VERSION") or project_config.get('azure_openai_api_version') or global_config.get('azure_openai_api_version', '2024-02-15-preview'),
        }
    else:  # openai
        config = {
            'provider': 'openai',
            'api_key': os.getenv("OPENAI_API_KEY") or project_config.get('openai_api_key') or global_config.get('openai_api_key') or global_config.get('OPENAI_API_KEY'),
        }
    
    # Return None if required fields are missing
    if provider_name == 'azure-openai':
        if not all([config.get('api_key'), config.get('endpoint'), config.get('deployment_id')]):
            return None
    else:
        if not config.get('api_key'):
            return None
    
    return config


def get_api_key(project_name: Optional[str] = None, projects_dir: Optional[Path] = None) -> Optional[str]:
    """
    Resolve OpenAI API key from multiple sources (backward compatibility).
    
    This function is kept for backward compatibility. Use get_provider_config() for new code.
    """
    config = get_provider_config(project_name, projects_dir)
    if config and config.get('provider') == 'openai':
        return config.get('api_key')
    return None


def save_provider_config(provider_config: dict, scope: str = 'global', 
                        project_name: Optional[str] = None, 
                        projects_dir: Optional[Path] = None) -> Path:
    """
    Save embedding provider configuration to specified location.
    
    Args:
        provider_config: Dict with provider configuration
        scope: 'global' or 'project'
        project_name: Required if scope is 'project'
        projects_dir: Required if scope is 'project'
    
    Returns:
        Path to the file where config was saved
    """
    provider_name = provider_config.get('provider', 'openai')
    
    if scope == 'project':
        if not project_name or not projects_dir:
            raise ValueError("project_name and projects_dir required for project scope")
        
        # Save to project.json
        project_json = projects_dir / project_name / "project.json"
        project_json.parent.mkdir(parents=True, exist_ok=True)
        
        # Read existing config
        existing_config = {}
        if project_json.exists():
            try:
                with open(project_json, 'r', encoding='utf-8') as f:
                    existing_config = json.load(f)
            except (json.JSONDecodeError, IOError):
                pass
        
        # Update with provider config
        existing_config['embedding_provider'] = provider_name
        
        if provider_name == 'azure-openai':
            existing_config['azure_openai_api_key'] = provider_config.get('api_key')
            existing_config['azure_openai_endpoint'] = provider_config.get('endpoint')
            existing_config['azure_openai_deployment_id'] = provider_config.get('deployment_id')
            if provider_config.get('api_version'):
                existing_config['azure_openai_api_version'] = provider_config.get('api_version')
        else:
            existing_config['openai_api_key'] = provider_config.get('api_key')
        
        # Write project.json
        with open(project_json, 'w', encoding='utf-8') as f:
            json.dump(existing_config, f, indent=2)
        
        os.chmod(project_json, 0o600)
        return project_json
    
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
        
        # Update with provider config
        config['embedding_provider'] = provider_name
        
        if provider_name == 'azure-openai':
            config['azure_openai_api_key'] = provider_config.get('api_key')
            config['azure_openai_endpoint'] = provider_config.get('endpoint')
            config['azure_openai_deployment_id'] = provider_config.get('deployment_id')
            if provider_config.get('api_version'):
                config['azure_openai_api_version'] = provider_config.get('api_version')
        else:
            config['openai_api_key'] = provider_config.get('api_key')
        
        # Write config file
        with open(config_path, 'w', encoding='utf-8') as f:
            json.dump(config, f, indent=2)
        
        os.chmod(config_path, 0o600)
        return config_path


def save_api_key(api_key: str, scope: str = 'global', project_name: Optional[str] = None, 
                 projects_dir: Optional[Path] = None) -> Path:
    """
    Save API key to specified location (backward compatibility).
    """
    provider_config = {'provider': 'openai', 'api_key': api_key}
    return save_provider_config(provider_config, scope, project_name, projects_dir)


def prompt_provider_config(provider_name: Optional[str] = None) -> dict:
    """
    Interactively prompt user for provider configuration.
    
    Args:
        provider_name: Optional provider name to prompt for (if None, will ask user to choose)
    
    Returns:
        Dict with provider configuration
    """
    # If provider not specified, ask user to choose
    if not provider_name:
        print("\nAvailable embedding providers:")
        print("  1. openai - OpenAI API")
        print("  2. azure-openai - Azure OpenAI")
        print()
        
        while True:
            choice = input("Select provider (1 or 2, default: 1): ").strip()
            if not choice:
                provider_name = 'openai'
                break
            elif choice == '1':
                provider_name = 'openai'
                break
            elif choice == '2':
                provider_name = 'azure-openai'
                break
            else:
                print("Invalid choice. Please enter 1 or 2.")
    
    config = {'provider': provider_name}
    
    if provider_name == 'azure-openai':
        print("\nAzure OpenAI Configuration")
        print("=" * 50)
        print("You need the following information from your Azure OpenAI resource:")
        print("  - API Key")
        print("  - Endpoint URL (e.g., https://your-resource.openai.azure.com)")
        print("  - Deployment ID (the name of your embedding model deployment)")
        print()
        
        api_key = getpass.getpass("Azure OpenAI API Key (input is hidden): ").strip()
        if not api_key:
            raise ValueError("API key cannot be empty")
        config['api_key'] = api_key
        
        endpoint = input("Azure OpenAI Endpoint URL: ").strip()
        if not endpoint:
            raise ValueError("Endpoint cannot be empty")
        # Remove trailing slash if present
        endpoint = endpoint.rstrip('/')
        config['endpoint'] = endpoint
        
        deployment_id = input("Deployment ID: ").strip()
        if not deployment_id:
            raise ValueError("Deployment ID cannot be empty")
        config['deployment_id'] = deployment_id
        
        api_version = input("API Version (default: 2024-02-15-preview): ").strip()
        if api_version:
            config['api_version'] = api_version
        else:
            config['api_version'] = '2024-02-15-preview'
    
    else:  # openai
        print("\nOpenAI Configuration")
        print("=" * 50)
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
                config['api_key'] = api_key
                break
            else:
                print("API key cannot be empty. Please try again.")
    
    return config


def prompt_api_key() -> str:
    """
    Interactively prompt user for OpenAI API key (backward compatibility).
    """
    config = prompt_provider_config('openai')
    return config['api_key']


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

