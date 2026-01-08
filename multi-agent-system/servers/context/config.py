"""Configuration for context provider agents."""

import os
from pathlib import Path

from dotenv import load_dotenv

load_dotenv()


class AgentConfig:
    """Configuration for agent settings."""

    def __init__(self) -> None:
        """Initialize agent configuration from environment variables."""
        self.mcp_server_url = os.getenv("MCP_SERVER_URL", "http://localhost:8080/codesearch/mcp")
        self.langfuse_enabled = os.getenv("LANGFUSE_ENABLED", "false").lower() == "true"
        
        # Default limits
        self.default_max_tool_calls = int(os.getenv("DEFAULT_MAX_TOOL_CALLS", "20"))
        self.default_max_tokens = int(os.getenv("DEFAULT_MAX_TOKENS", "16000"))
        
        # Model configurations (default to OpenAI models)
        self.query_reformater_model_name = os.getenv(
            "QUERY_REFORMATER_MODEL_NAME", "gpt-4o-mini"
        )
        self.code_snippet_finder_model_name = os.getenv(
            "CODE_SNIPPET_FINDER_MODEL_NAME", "gpt-4o-mini"
        )
        
        # Default to OpenAI (can be overridden with DEFAULT_BASE_URL)
        self.default_base_url = os.getenv(
            "DEFAULT_BASE_URL", "https://api.openai.com/v1"
        )

    def get_model_kwargs(self, agent_name: str) -> dict:
        """Get model-specific kwargs for an agent.

        Args:
            agent_name: Name of the agent

        Returns:
            Dictionary of model kwargs
        """
        # Common kwargs
        kwargs = {}
        
        # Check for custom base URL
        base_url_key = f"{agent_name.upper()}_BASE_URL"
        base_url = os.getenv(base_url_key)
        if base_url:
            kwargs["base_url"] = base_url
        
        # Check for custom API key
        api_key_key = f"{agent_name.upper()}_API_KEY"
        api_key = os.getenv(api_key_key)
        if api_key:
            kwargs["api_key"] = api_key
        
        return kwargs

