#!/bin/bash
# Complete setup and startup script for Dynamic Few-Shot RAG System

set -e  # Exit on error

echo "🚀 Setting up Dynamic Few-Shot RAG System for Zoekt NL Translation"
echo "=================================================================="
echo ""

cd "$(dirname "$0")"

# Step 1: Install Python dependencies
echo "📦 Step 1: Installing Python dependencies..."
if ! command -v pip3 &> /dev/null; then
    echo "❌ pip3 not found. Please install Python 3 with pip."
    exit 1
fi

pip3 install -q -r requirements.txt
echo "✅ Python dependencies installed"
echo ""

# Step 2: Generate examples (if not already generated)
if [ ! -f "zoekt_examples.jsonl" ]; then
    echo "📝 Step 2: Generating 1,000 NL-to-Zoekt query examples..."
    echo "   This will use mistralai/mistral-7b-instruct:free (free model)"
    echo "   This may take 15-20 minutes due to rate limits..."
    echo ""
    python3 generate_examples.py
    
    if [ ! -f "zoekt_examples.jsonl" ]; then
        echo "❌ Failed to generate examples"
        exit 1
    fi
    
    echo "✅ Examples generated"
else
    echo "📝 Step 2: Examples file already exists (zoekt_examples.jsonl)"
    echo "   Skipping generation. Delete the file to regenerate."
fi
echo ""

# Step 3: Build vector index (if not already built)
if [ ! -d "chroma_db" ] || [ -z "$(ls -A chroma_db 2>/dev/null)" ]; then
    echo "🔨 Step 3: Building ChromaDB vector index..."
    python3 build_fewshot_index.py
    
    if [ ! -d "chroma_db" ]; then
        echo "❌ Failed to build index"
        exit 1
    fi
    
    echo "✅ Vector index built"
else
    echo "🔨 Step 3: Vector index already exists (chroma_db/)"
    echo "   Skipping index build. Delete the directory to rebuild."
fi
echo ""

# Step 4: Check if services are running
echo "🔍 Step 4: Checking services..."
FEWSHOT_PID=$(lsof -ti:6072 2>/dev/null || echo "")
GO_PID=$(lsof -ti:6071 2>/dev/null || echo "")

if [ -n "$FEWSHOT_PID" ]; then
    echo "⚠️  Few-shot service already running on port 6072 (PID: $FEWSHOT_PID)"
else
    echo "✅ Few-shot service not running (will start)"
fi

if [ -n "$GO_PID" ]; then
    echo "⚠️  Go service already running on port 6071 (PID: $GO_PID)"
else
    echo "✅ Go service not running (will start)"
fi
echo ""

# Step 5: Start services
echo "🎯 Step 5: Starting services..."
echo ""

# Start few-shot service in background
if [ -z "$FEWSHOT_PID" ]; then
    echo "Starting few-shot retrieval service on port 6072..."
    python3 fewshot_service.py 6072 > fewshot_service.log 2>&1 &
    FEWSHOT_NEW_PID=$!
    sleep 2
    
    # Check if it started
    if ps -p $FEWSHOT_NEW_PID > /dev/null; then
        echo "✅ Few-shot service started (PID: $FEWSHOT_NEW_PID)"
    else
        echo "❌ Failed to start few-shot service. Check fewshot_service.log"
        exit 1
    fi
else
    echo "✅ Few-shot service already running"
fi

# Wait a bit for service to be ready
sleep 2

# Start Go service
if [ -z "$GO_PID" ]; then
    echo ""
    echo "Starting Go NL Query Server on port 6071..."
    echo "   Dashboard: http://localhost:6071/dashboard"
    echo "   API: http://localhost:6071/api/nl-search?q=your%20query"
    echo ""
    go run . &
    GO_NEW_PID=$!
    sleep 3
    
    # Check if it started
    if ps -p $GO_NEW_PID > /dev/null; then
        echo "✅ Go service started (PID: $GO_NEW_PID)"
    else
        echo "❌ Failed to start Go service"
        exit 1
    fi
else
    echo "✅ Go service already running"
fi

echo ""
echo "=================================================================="
echo "✅ Setup complete! Services are running:"
echo ""
echo "   📊 Few-Shot Service: http://localhost:6072"
echo "   🔍 NL Query Server:  http://localhost:6071"
echo "   🎨 Dashboard:        http://localhost:6071/dashboard"
echo ""
echo "   To stop services:"
echo "   pkill -f fewshot_service.py"
echo "   pkill -f 'go run'"
echo ""
echo "   To view logs:"
echo "   tail -f fewshot_service.log"
echo "=================================================================="

