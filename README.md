# SDLC AI - Semantic Code Search System

A powerful semantic code search system that combines natural language queries with vector embeddings and keyword search for comprehensive codebase exploration.

## Features

- **Semantic Search**: Find code by meaning using vector embeddings (Ollama + ChromaDB)
- **Hybrid Search**: Combines semantic search with Zoekt keyword search for best results
- **Tree-sitter Parsing**: AST-based code parsing for accurate function/class extraction
- **Natural Language Queries**: Ask questions in plain English
- **Web Interface**: Beautiful dashboard for interactive code search
- **Free & Open Source**: Uses Ollama for local, free embeddings

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   Web UI    │────▶│  NL Server   │────▶│   Zoekt     │
│  (Port 6071)│     │  (Go)        │     │  (Keyword)  │
└─────────────┘     └──────────────┘     └─────────────┘
                            │
                            ├────▶ ChromaDB (Vector Store)
                            │      Port 8000
                            │
                            └────▶ Ollama (Embeddings)
                                   Port 11434
```

## Quick Start

### Prerequisites

- Go 1.19+
- Python 3.10+
- Ollama installed and running
- ChromaDB (included as Python server)

### Installation

1. **Install Ollama**:
   ```bash
   brew install ollama  # macOS
   ollama serve &
   ollama pull nomic-embed-text
   ```

2. **Start ChromaDB Server**:
   ```bash
   cd zoekt-nl-query
   python3 chromadb_server.py &
   ```

3. **Build and Run NL Server**:
   ```bash
   cd zoekt-nl-query
   go build -o zoekt-nl-server .
   ./zoekt-nl-server -port 6071
   ```

4. **Index Your Codebase**:
   ```bash
   curl -X POST "http://localhost:6071/api/index-codebase?path=/path/to/your/codebase"
   ```

5. **Access Web Interface**:
   Open http://localhost:6071 in your browser

## API Endpoints

### Search
- `GET /api/nl-search?q=<query>&mode=<semantic|hybrid|keyword>`
  - `semantic`: Pure semantic search using embeddings
  - `hybrid`: Semantic + Zoekt keyword search (enhanced)
  - `keyword`: Zoekt keyword search only

### Indexing
- `POST /api/index-codebase?path=<codebase_path>` - Index a codebase for semantic search

### Stats
- `GET /api/semantic-stats` - Get indexing statistics

## Search Modes

### Semantic Mode
Uses vector embeddings to find code by meaning:
```bash
curl "http://localhost:6071/api/nl-search?q=authentication logic&mode=semantic"
```

### Hybrid Mode (Recommended)
Combines semantic search with Zoekt keyword matching:
```bash
curl "http://localhost:6071/api/nl-search?q=find user login&mode=hybrid"
```

### Keyword Mode
Traditional Zoekt keyword search:
```bash
curl "http://localhost:6071/api/nl-search?q=login&mode=keyword"
```

## Components

### zoekt-nl-query/
Main application directory containing:
- **semantic_indexer.go**: Indexes code using Tree-sitter + embeddings
- **semantic_search.go**: Performs semantic search via ChromaDB
- **treesitter_parser.go**: AST-based code parsing
- **embedding_service.go**: Ollama embedding integration
- **server.go**: HTTP server and API endpoints
- **unified-dashboard.html**: Web interface

## Technology Stack

- **Go**: Main server and indexing logic
- **Ollama**: Local embedding generation (nomic-embed-text)
- **ChromaDB**: Vector database for semantic search
- **Tree-sitter**: AST parsing for code chunking
- **Zoekt**: Fast keyword-based code search

## Configuration

### Environment Variables

- `USE_OLLAMA=true` (default): Use Ollama for embeddings
- `OLLAMA_URL=http://localhost:11434`: Ollama server URL
- `CHROMA_URL=http://localhost:8000`: ChromaDB server URL
- `EMBEDDING_MODEL=nomic-embed-text`: Embedding model name
- `OPENROUTER_API_KEY`: Optional, for OpenRouter fallback

## License

MIT

