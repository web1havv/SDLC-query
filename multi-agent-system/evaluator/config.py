"""Configuration for evaluator/judge components."""

import os

from dotenv import load_dotenv

load_dotenv()


class JudgeConfig:
    """Configuration for LLM judge."""

    def __init__(self) -> None:
        """Initialize judge configuration."""
        self.langfuse_enabled = os.getenv("LANGFUSE_ENABLED", "false").lower() == "true"
        self.llm_judge_model_name = os.getenv(
            "LLM_JUDGE_MODEL_NAME", "gpt-4o-mini"
        )
        self.code_agent_type_parser_model_name = os.getenv(
            "CODE_AGENT_TYPE_PARSER_MODEL_NAME", "gpt-4o-mini"
        )
        # Default to OpenAI
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
        kwargs = {}

        base_url_key = f"{agent_name.upper()}_BASE_URL"
        base_url = os.getenv(base_url_key)
        if base_url:
            kwargs["base_url"] = base_url

        api_key_key = f"{agent_name.upper()}_API_KEY"
        api_key = os.getenv(api_key_key)
        if api_key:
            kwargs["api_key"] = api_key

        return kwargs

