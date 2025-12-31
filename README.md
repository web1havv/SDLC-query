# SDLC Query - Natural Language to Zoekt Query Translation

A natural language query translation layer built on top of Zoekt code search engine. This system converts human-readable questions into Zoekt DSL queries and provides conversational answers about codebases.

## Overview

This project extends Zoekt with AI-powered natural language processing capabilities, allowing developers to search codebases using plain English instead of learning Zoekt's query syntax.

## Key Features

### 1. Natural Language to Zoekt DSL Translation
- Converts natural language queries to valid Zoekt syntax
- Uses OpenRouter API with free Mistral model
- Dynamic few-shot retrieval from ChromaDB vector database
- Intelligent query generation with context awareness

### 2. Conversational Q&A System
- Ask questions about codebase in natural language
- Generates Zoekt queries to find relevant code
- Extracts 100+ lines of code snippets for context
- Provides text-based answers using LLM

### 3. Dynamic Few-Shot RAG
- Vector database with 400 NL-to-Zoekt query examples
- Semantic search for similar examples
- Dynamic injection into prompts for better accuracy
- ChromaDB-based retrieval service

### 4. Code Snippet Extraction
- Extracts comprehensive code context (100+ lines)
- Uses full file content when available
- 30 context lines per match
- Smart extraction from multiple files

## Architecture

```
User Query (Natural Language)
    ↓
OpenRouterTranslator
    ├─→ FewShotClient → ChromaDB (retrieve similar examples)
    ├─→ Generate Zoekt Query
    ↓
Zoekt Search Engine
    ↓
Extract Code Snippets (100+ lines)
    ↓
LLM Answer Generation
    ↓
Response (Text Answer + Search Results)
```

## Components

### Core Go Files
- `main.go` - Entry point, initializes server
- `server.go` - HTTP handlers for `/api/nl-search` and `/api/ask`
- `openrouter_translator.go` - NL-to-Zoekt translation with few-shot RAG
- `basic_translator.go` - Fallback pattern-based translator
- `fewshot_client.go` - ChromaDB retrieval client

### Python Services
- `fewshot_service.py` - Flask service for example retrieval (port 6072)
- `build_fewshot_index.py` - Builds ChromaDB vector index
- `generate_examples_local.py` - Generates 400 training examples

### Frontend
- `unified-dashboard.html` - Web interface with Search and Ask tabs

## Setup

1. **Install Dependencies**
   ```bash
   cd zoekt-nl-query
   pip install -r requirements.txt
   go mod download
   ```

2. **Set Environment Variables**
   ```bash
   export OPENROUTER_API_KEY="your-api-key"
   export OPENROUTER_MODEL="mistralai/mistral-7b-instruct:free"  # Optional
   ```

3. **Build ChromaDB Index**
   ```bash
   python3 build_fewshot_index.py
   ```

4. **Start Services**
   ```bash
   ./setup_and_start.sh
   ```

   Or manually:
   ```bash
   # Start few-shot service
   python3 fewshot_service.py 6072 &
   
   # Start Go server
   go run . -port 6071
   ```

## Usage

### API Endpoints

**Natural Language Search**
```
GET /api/nl-search?q=list all python functions
```

**Ask Question**
```
GET /api/ask?q=have I used React in this codebase?
```

**Dashboard**
```
http://localhost:6071/dashboard
```

### Example Queries

- "list all python functions" → `lang:python content:def`
- "find login function" → `sym:login`
- "search for TODO comments in Python" → `lang:python content:/#.*TODO/`
- "find where API_KEY is used but not in tests" → `content:API_KEY -file:test`

## Query Translation Rules

- Use `content:` for searching text in files (e.g., `content:def` for function definitions)
- Use `sym:` only for specific named functions (e.g., `sym:login`)
- Use `lang:` for language filtering (e.g., `lang:python`)
- Use `file:` for file patterns (e.g., `file:*.py`)
- Use `-file:` for exclusions (e.g., `-file:test`)
- Use `or` for logical OR operations
- Wrap regex patterns in forward slashes (e.g., `/pattern/`)

## Important Files

- `zoekt-nl-query/` - Main application directory
- `zoekt/` - Original Zoekt repository (unchanged)
- `chroma_db/` - Vector database for few-shot examples
- `zoekt_examples.jsonl` - Training data (400 examples)

## Configuration

The system requires:
- `OPENROUTER_API_KEY` environment variable (no hardcoded keys)
- ChromaDB service running on port 6072
- Zoekt index directory (default: `~/.zoekt`)

## Notes

- Original Zoekt repository remains completely unchanged
- All modifications are in the `zoekt-nl-query/` directory
- System falls back to basic translator if OpenRouter is unavailable
- Few-shot retrieval gracefully degrades to static examples if ChromaDB is unavailable

