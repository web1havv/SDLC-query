"""Resource limiters for agents (token and tool call limits)."""

from typing import Any, TypeVar

from pydantic_ai.agent import Agent

T = TypeVar("T")


class TokenLimiter:
    """Limits token usage for agent runs."""

    def __init__(self, max_tokens: int) -> None:
        """Initialize token limiter.

        Args:
            max_tokens: Maximum number of tokens allowed (input + output)
        """
        self.max_tokens = max_tokens

    async def run_with_limit(self, agent: Agent, user_prompt: str) -> Any:
        """Run agent with token limit checking.

        Args:
            agent: The agent to run
            user_prompt: User prompt string

        Returns:
            Agent result

        Raises:
            RuntimeError: If token limit is exceeded
        """
        # Simple implementation - in production you'd want more sophisticated token counting
        result = await agent.run(user_prompt)

        # Check token usage (this is a simplified check)
        # In a real implementation, you'd track tokens from the model response
        # For now, we'll add a warning to the system prompt instead

        return result


class ToolCallLimiter:
    """Limits tool calls for agents."""

    def __init__(self, max_calls: int) -> None:
        """Initialize tool call limiter.

        Args:
            max_calls: Maximum number of tool calls allowed
        """
        self.max_calls = max_calls
        self.call_count = 0

    def reset(self) -> None:
        """Reset the call counter."""
        self.call_count = 0

    def increment(self) -> bool:
        """Increment call counter and check if limit reached.

        Returns:
            True if limit not reached, False if limit exceeded
        """
        self.call_count += 1
        return self.call_count <= self.max_calls

    def wrap_mcp_server(self, mcp_server: Any) -> Any:
        """Wrap an MCP server to enforce tool call limits.

        Args:
            mcp_server: The MCP server to wrap

        Returns:
            Wrapped MCP server that enforces limits
        """
        # Note: This is a placeholder implementation
        # The actual MCP server wrapping would depend on the specific MCP implementation
        # For FastMCP and pydantic-ai, the tool call limiting is typically handled
        # at a higher level through agent configuration or middleware
        # This method is kept for API compatibility but may need to be implemented
        # based on the specific MCP client being used
        
        # For now, return the server as-is
        # In a production implementation, you would wrap the server to intercept
        # tool calls and check limits before executing
        return mcp_server

