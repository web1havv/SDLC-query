#!/bin/bash
# Quick test script for natural language query translation

echo "🔍 Testing Natural Language Query Translation"
echo "=============================================="
echo ""
echo "Query: 'search for getAllBlogs() function in JavaScript'"
echo ""

# Test via API if server is running
if curl -s "http://localhost:6071/api/nl-search?q=search%20for%20getAllBlogs%28%29%20function%20in%20JavaScript" > /dev/null 2>&1; then
    echo "✅ Server is running! Testing query..."
    echo ""
    curl -s "http://localhost:6071/api/nl-search?q=search%20for%20getAllBlogs%28%29%20function%20in%20JavaScript" | python3 -m json.tool 2>/dev/null || \
    curl -s "http://localhost:6071/api/nl-search?q=search%20for%20getAllBlogs%28%29%20function%20in%20JavaScript"
    echo ""
else
    echo "⚠️  Server not running. Expected Zoekt query would be:"
    echo ""
    echo "   lang:javascript sym:getAllBlogs"
    echo ""
    echo "   OR"
    echo ""
    echo "   lang:javascript content:getAllBlogs"
    echo ""
    echo "To start the server, run:"
    echo "   cd /Users/web1havv/SDLC_AI/zoekt-nl-query"
    echo "   go run ."
    echo ""
    echo "Then test with:"
    echo "   curl 'http://localhost:6071/api/nl-search?q=search%20for%20getAllBlogs%28%29%20function%20in%20JavaScript'"
    echo ""
    echo "Or use the dashboard:"
    echo "   http://localhost:6071/dashboard"
fi



