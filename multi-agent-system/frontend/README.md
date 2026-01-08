# Frontend Web Interface

Simple web frontend for the multi-agent code search system.

## Usage

### Start the frontend server:

```bash
# From the multi-agent-system directory
python -m frontend.app

# Or with custom port
FRONTEND_PORT=3000 python -m frontend.app
```

### Access the web interface:

Open your browser and navigate to:
```
http://localhost:3000
```

## Features

- **Query Reformulation**: Reformulate your natural language questions into better search queries
- **Code Search**: Search the codebase using natural language questions

## API Endpoints

- `POST /api/query-reformulate` - Reformulate a query
- `POST /api/code-search` - Search for code snippets
- `GET /` - Main web interface

