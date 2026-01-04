# Zoekt Natural Language Query & Metrics Dashboard

This extension adds natural language query support and comprehensive metrics tracking to Zoekt.

## Features

### 1. Natural Language Query Translation
Converts natural language queries to Zoekt queries:

- **Count queries**: "how many articles are there" → `title: file:blogs.js`
- **List queries**: "list all blogs" → `title: file:blogs.js`
- **Find queries**: "find articles about Amazon ML" → `Amazon ML file:blogs.js`
- **Blog queries**: "blog titles" → `title: file:blogs.js`

### 2. Metrics Tracking
Tracks comprehensive query metrics:
- Total queries, success rate
- Response times (min, max, average)
- Result counts
- Natural language vs direct queries
- Query type distribution
- Recent query history

### 3. Real-time Dashboard
Beautiful web dashboard showing:
- Key metrics cards
- Query type distribution chart
- Response time trends
- Recent queries table

## Usage

### Integration with Zoekt Web Server

```go
package main

import (
    "github.com/sourcegraph/zoekt"
    "github.com/sourcegraph/zoekt/search"
    nlquery "path/to/zoekt-nl-query"
)

func main() {
    // Create searcher
    searcher, _ := search.NewDirectorySearcher("~/.zoekt")
    
    // Create NL query server
    nlServer := nlquery.NewNLQueryServer(searcher)
    
    // Setup routes
    mux := http.NewServeMux()
    nlServer.SetupRoutes(mux)
    
    // Start server
    http.ListenAndServe(":6070", mux)
}
```

### API Endpoints

#### Natural Language Search
```
GET /api/nl-search?q=how many articles are there

Response:
{
  "originalQuery": "how many articles are there",
  "translatedQuery": "title: file:blogs.js",
  "queryType": "count",
  "isNL": true,
  "success": true,
  "resultCount": 3,
  "responseTime": 45,
  "results": { ... }
}
```

#### Metrics
```
GET /api/metrics

Response:
{
  "totalQueries": 150,
  "successRate": 98.5,
  "avgResponseTime": 45,
  "nlQueries": 50,
  "nlQuerySuccessRate": 96.0,
  ...
}
```

#### Dashboard
```
GET /dashboard
```

## Example Natural Language Queries

| Natural Language | Translated Zoekt Query |
|-----------------|------------------------|
| "how many articles are there" | `title: file:blogs.js` |
| "list all blogs" | `title: file:blogs.js` |
| "find articles about Amazon ML" | `Amazon ML file:blogs.js` |
| "how many functions" | `function` |
| "show all components" | `export default function file:*.jsx` |
| "blog titles" | `title: file:blogs.js` |
| "search for useState" | `useState` |

## Metrics Dashboard

The dashboard shows:
- **Key Metrics**: Total queries, success rate, response times, result counts
- **Query Distribution**: Pie chart of query types
- **Response Time Trends**: Line chart of recent query performance
- **Recent Queries**: Table of last 20 queries with details

## Architecture

```
User Query (Natural Language)
    ↓
NaturalLanguageTranslator
    ↓
Zoekt Query
    ↓
Zoekt Searcher
    ↓
Results + Metrics Recording
    ↓
Response + Dashboard Update
```

## Metrics Tracked

1. **Query Statistics**
   - Total queries
   - Successful/failed queries
   - Success rate

2. **Performance**
   - Response time (min, max, average)
   - Result count per query
   - Average results

3. **Query Types**
   - Natural language queries
   - Direct Zoekt queries
   - Count/List/Find/Blog queries

4. **Recent History**
   - Last 100 queries
   - Query, translation, results, timing

## Future Enhancements

- [ ] Machine learning for better NL translation
- [ ] Query suggestions/autocomplete
- [ ] Export metrics to Prometheus
- [ ] Query performance optimization hints
- [ ] Multi-language support
- [ ] Query templates







