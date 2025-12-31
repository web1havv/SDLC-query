#!/bin/bash
# Start the zoekt-nl-query server with visible logs

cd /Users/web1havv/SDLC_AI/zoekt-nl-query

echo "🛑 Stopping any existing server..."
pkill -f "zoekt-nl-query" 2>/dev/null
sleep 1

echo "🚀 Starting server with logs..."
echo "📝 Logs will be saved to: server.log"
echo "📺 View logs in real-time with: tail -f server.log"
echo ""
echo "Press Ctrl+C to stop the server"
echo ""

# Start server and save logs to file
go run . 2>&1 | tee server.log

