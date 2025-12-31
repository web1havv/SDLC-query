#!/bin/bash
# View live logs from the zoekt-nl-query server

echo "🔍 Finding zoekt-nl-query server process..."
echo ""

# Find the process
PID=$(pgrep -f "zoekt-nl-query" | head -1)

if [ -z "$PID" ]; then
    echo "❌ Server not running!"
    echo ""
    echo "To start the server with logs visible:"
    echo "  cd /Users/web1havv/SDLC_AI/zoekt-nl-query"
    echo "  go run . 2>&1 | tee server.log"
    echo ""
    echo "Then in another terminal, view logs with:"
    echo "  tail -f /Users/web1havv/SDLC_AI/zoekt-nl-query/server.log"
    exit 1
fi

echo "✅ Found server process: PID $PID"
echo ""
echo "📋 Options to view logs:"
echo ""
echo "1. If server was started with logging to file:"
echo "   tail -f /Users/web1havv/SDLC_AI/zoekt-nl-query/server.log"
echo ""
echo "2. View process info:"
echo "   ps -p $PID -o pid,command"
echo ""
echo "3. Restart server with visible logs:"
echo "   pkill -f zoekt-nl-query"
echo "   cd /Users/web1havv/SDLC_AI/zoekt-nl-query"
echo "   go run . 2>&1 | tee server.log"
echo ""
echo "4. Check if log file exists:"
if [ -f "/Users/web1havv/SDLC_AI/zoekt-nl-query/server.log" ]; then
    echo "   ✅ Log file exists! Showing last 50 lines:"
    echo ""
    tail -50 /Users/web1havv/SDLC_AI/zoekt-nl-query/server.log
else
    echo "   ❌ No log file found. Server is running in background."
    echo "   Restart with: cd /Users/web1havv/SDLC_AI/zoekt-nl-query && go run . 2>&1 | tee server.log"
fi

