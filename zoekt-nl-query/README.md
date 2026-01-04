# Zoekt Natural Language Query with Semantic Search

A semantic code search system that extends Zoekt with natural language queries, vector embeddings, and hybrid search capabilities.

## Overview

This system provides:
1. **Natural Language to Zoekt Query Translation** - Converts plain English to Zoekt queries
2. **Semantic Code Search** - Finds code by meaning using vector embeddings
3. **Hybrid Search** - Combines semantic and keyword search for best results
4. **Tree-sitter Parsing** - AST-based code chunking for accurate indexing
5. **Web Dashboard** - Interactive interface for code search

## Features

### Semantic Search
- Uses Ollama (nomic-embed-text) for free, local embeddings
- ChromaDB vector store for fast similarity search
- Finds code by semantic meaning, not just keywords

### Hybrid Search
- Combines semantic search with Zoekt keyword matching
- Zoekt results enhance and boost semantic results
- Best of both worlds: meaning + exact matches

### Code Parsing
- Tree-sitter for AST-based parsing (Go, JavaScript, Python)
- Regex fallback for other languages
- Extracts functions, classes, structs, interfaces

## Installation

### 1. Install Dependencies

```bash
# Install Ollama
brew install ollama  # macOS
# or visit https://ollama.com

# Start Ollama
ollama serve &
ollama pull nomic-embed-text

# Install Python dependencies (for ChromaDB)
pip install chromadb
```

### 2. Build the Server

```bash
go mod download
go build -o zoekt-nl-server .
```

### 3. Start Services

```bash
# Terminal 1: Start ChromaDB
python3 chromadb_server.py

# Terminal 2: Start NL Server
./zoekt-nl-server -port 6071
```

## Usage

### Index a Codebase

```bash
curl -X POST "http://localhost:6071/api/index-codebase?path=/path/to/your/codebase"
```

### Search

**Web Interface:**
```
http://localhost:6071/dashboard
```

**API:**
```bash
# Semantic search
curl "http://localhost:6071/api/nl-search?q=authentication logic&mode=semantic"

# Hybrid search (recommended)
curl "http://localhost:6071/api/nl-search?q=find user login&mode=hybrid"

# Keyword search
curl "http://localhost:6071/api/nl-search?q=login&mode=keyword"
```

## API Reference

### GET /api/nl-search
Natural language search endpoint.

**Query Parameters:**
- `q` (required): Search query
- `mode` (optional): `semantic`, `hybrid`, or `keyword` (default: `hybrid`)
- `direct` (optional): `true` to skip NL translation

**Response:**
```json
{
  "originalQuery": "authentication logic",
  "semanticResults": [
    {
      "Chunk": {
        "FileName": "auth.py",
        "Content": "def authenticate(...)",
        "StartLine": 10,
        "EndLine": 25
      },
      "Score": 0.85,
      "Similarity": 0.85
    }
  ],
  "resultCount": 1,
  "mode": "semantic"
}
```

### POST /api/index-codebase
Index a codebase for semantic search.

**Query Parameters:**
- `path` (required): Path to codebase directory

**Response:**
```json
{
  "success": true,
  "filesIndexed": 150,
  "chunksIndexed": 2000
}
```

### GET /api/semantic-stats
Get semantic indexing statistics.

**Response:**
```json
{
  "indexed": true,
  "total_chunks": 2000
}
```

## Architecture

### Components

1. **semantic_indexer.go**
   - Parses code files using Tree-sitter
   - Generates embeddings via Ollama
   - Stores chunks in ChromaDB

2. **semantic_search.go**
   - Queries ChromaDB with embeddings
   - Returns semantically similar code chunks

3. **treesitter_parser.go**
   - AST-based code parsing
   - Extracts functions, classes, etc.
   - Falls back to regex for unsupported languages

4. **embedding_service.go**
   - Ollama integration for embeddings
   - OpenRouter fallback support

5. **server.go**
   - HTTP server and routing
   - Handles search requests
   - Orchestrates hybrid search

### Data Flow

```
User Query
    ↓
Generate Embedding (Ollama)
    ↓
Query ChromaDB (Vector Search)
    ↓
Get Semantic Results
    ↓
[Hybrid Mode] Query Zoekt (Keyword Search)
    ↓
Enhance & Merge Results
    ↓
Return to User
```

## Configuration

### Environment Variables

- `USE_OLLAMA=true`: Enable Ollama (default: true)
- `OLLAMA_URL=http://localhost:11434`: Ollama server URL
- `CHROMA_URL=http://localhost:8000`: ChromaDB server URL
- `EMBEDDING_MODEL=nomic-embed-text`: Embedding model
- `OPENROUTER_API_KEY`: Optional API key for OpenRouter fallback

### Ports

- **6071**: NL Query Server (main API)
- **8000**: ChromaDB Server
- **11434**: Ollama Server

## Development

### Project Structure

```
zoekt-nl-query/
├── main.go                 # Entry point
├── server.go              # HTTP server and routes
├── semantic_indexer.go    # Code indexing logic
├── semantic_search.go     # Semantic search service
├── treesitter_parser.go   # AST parsing
├── embedding_service.go   # Embedding generation
├── openrouter_translator.go # NL to Zoekt translation
├── basic_translator.go    # Fallback translator
├── chromadb_server.py     # ChromaDB HTTP wrapper
├── unified-dashboard.html # Web interface
└── README.md             # This file
```

### Building

```bash
go build -o zoekt-nl-server .
```

### Running Tests

```bash
go test ./...
```

## Troubleshooting

### ChromaDB Connection Issues
- Ensure `chromadb_server.py` is running on port 8000
- Check logs: `tail -f chromadb_server.log`

### Ollama Connection Issues
- Verify Ollama is running: `curl http://localhost:11434/api/tags`
- Ensure model is pulled: `ollama pull nomic-embed-text`

### No Search Results
- Check if codebase is indexed: `curl http://localhost:6071/api/semantic-stats`
- Re-index if needed: `curl -X POST "http://localhost:6071/api/index-codebase?path=..."`

## License

MIT
