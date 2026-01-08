# Multi-Agent System with PromptManager

A comprehensive multi-agent system for code search and analysis with centralized prompt management using YAML and Jinja2 templating.

## Overview

This system provides:

1. **PromptManager**: Centralized YAML-based prompt management with Jinja2 templating
2. **Context Provider MCP Server**: Agents for query reformulation and code snippet finding
3. **Code Search MCP Server**: Backend-agnostic code search (Zoekt/Sourcegraph)
4. **Evaluator Components**: LLM-based judge for evaluating agent responses

## Features

- **YAML-based prompts**: All prompts stored in a single YAML file
- **Jinja2 templating**: Dynamic variable substitution in prompts
- **Backend-agnostic**: Supports Zoekt and Sourcegraph search backends
- **Resource limiting**: Token and tool call limiters for agents
- **Telemetry**: Langfuse integration for observability
- **MCP servers**: FastMCP-based servers for agent communication

## Installation

```bash
# Install dependencies
pip install -r requirements.txt

# Copy environment variables
cp .env.example .env
# Edit .env with your configuration
```

## Configuration

### Environment Variables

See `.env.example` for all configuration options. Key settings:

- `OPENAI_API_KEY`: Required for LLM agents
- `SEARCH_BACKEND`: Either `zoekt` or `sourcegraph`
- `MCP_SERVER_URL`: URL for MCP server communication
- `LANGFUSE_ENABLED`: Enable/disable Langfuse telemetry

### Prompts Configuration

Edit `prompts/prompts.yaml` to customize all prompts:

- **tools**: Tool descriptions for MCP tools
- **guides**: Code search guides for Zoekt/Sourcegraph
- **agents**: System and user prompts for agents
- **context_provider**: Tool descriptions for context provider

## Usage

### Start Context Provider Server

```bash
python -m servers.context.server
```

This starts the MCP server on:
- SSE: `http://localhost:8000`
- HTTP: `http://localhost:8080`

### Start Code Search Server

```bash
# Set SEARCH_BACKEND=zoekt or sourcegraph in .env
python -m servers.codesearch.server
```

### Use Agents Programmatically

```python
from servers.context.agent import CodeSnippetFinder, QueryReformater

# Query reformulation
async with QueryReformater() as agent:
    result = await agent.run("find authentication logic")
    print(result.suggested_queries)

# Code snippet finding
async with CodeSnippetFinder() as agent:
    result = await agent.run("How does user authentication work?")
    print(result)
```

## Project Structure

```
multi-agent-system/
├── core/
│   ├── prompt_manager.py    # PromptManager class
│   ├── limiters.py          # Token/Tool call limiters
│   └── __init__.py
├── prompts/
│   └── prompts.yaml         # All prompts in YAML
├── servers/
│   ├── context/
│   │   ├── server.py        # Context provider MCP server
│   │   ├── agent.py         # QueryReformater, CodeSnippetFinder
│   │   └── config.py        # Agent configuration
│   └── codesearch/
│       └── server.py        # Code search MCP server
├── backends/
│   ├── search.py            # Search client implementations
│   ├── content_fetcher.py   # Content fetcher implementations
│   └── models.py            # Data models
├── evaluator/
│   ├── agent.py             # LLMJudge, CodeAgentTypeParser
│   ├── config.py            # Judge configuration
│   └── models.py            # Evaluation models
├── requirements.txt
├── .env.example
└── README.md
```

## PromptManager Usage

```python
from core import PromptManager
from pathlib import Path

# Load entire prompts file
pm = PromptManager(file_path=Path("prompts/prompts.yaml"))

# Load specific section
agent_prompts = PromptManager(
    file_path=Path("prompts/prompts.yaml"),
    section_path="agents.code_snippet_finder"
)

# Render prompt with variables
prompt = agent_prompts.render_prompt(
    "user_prompt",
    question="Find authentication handlers"
)
```

## Development

### Adding New Prompts

1. Edit `prompts/prompts.yaml`
2. Add prompts in the appropriate section
3. Use Jinja2 syntax for variables: `{{ variable_name }}`

### Adding New Agents

1. Create agent class in appropriate module
2. Initialize PromptManager with agent section
3. Render prompts using `render_prompt()`

### Adding New Backends

1. Implement `AbstractSearchClient` or `AbstractContentFetcher`
2. Add factory method in respective Factory class
3. Update `prompts.yaml` with backend-specific prompts

## License

MIT

