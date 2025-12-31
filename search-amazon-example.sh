#!/bin/bash
# Example: Search for "amazon" and list all files containing it
# This shows how to use the NL query server to find all files with "amazon"

echo "🔍 Searching for 'amazon' in all files..."
echo ""
echo "Option 1: Using direct query mode (Zoekt syntax)"
echo "  curl 'http://localhost:6071/api/nl-search?q=amazon&direct=true' | jq '.results.Files[] | .FileName'"
echo ""

echo "Option 2: Using natural language with AI model"
echo "  curl 'http://localhost:6071/api/nl-search?q=find all files containing amazon' | jq '.generatedQuery, .results.Files[] | .FileName'"
echo ""

echo "Option 3: Using zoekt CLI directly"
echo "  export PATH=\"\$PATH:\$(go env GOPATH)/bin\""
echo "  zoekt 'amazon' -r"
echo "  zoekt -l 'amazon'  # List only filenames"
echo ""

echo "To get just the list of files:"
echo "  curl -s 'http://localhost:6071/api/nl-search?q=amazon&direct=true' | jq -r '.results.Files[] | .FileName'"
echo ""




