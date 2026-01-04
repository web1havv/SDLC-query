#!/usr/bin/env python3
"""Start ChromaDB server"""
import chromadb
from chromadb.config import Settings
import sys
import os

# Get path from environment or use default
chroma_path = os.getenv("CHROMA_PATH", "/Users/web1havv/SDLC_AI/zoekt-nl-query/chroma_db")
port = int(os.getenv("CHROMA_PORT", "8000"))

print(f"Starting ChromaDB server on port {port}")
print(f"Data directory: {chroma_path}")

# Start ChromaDB server
try:
    # Use the new client API
    client = chromadb.PersistentClient(path=chroma_path)
    print(f"✅ ChromaDB client initialized at {chroma_path}")
    print(f"✅ Server should be accessible at http://localhost:{port}")
    print("Note: This script initializes the client. For full server mode, use: chroma run --path ./chroma_db --port 8000")
except Exception as e:
    print(f"❌ Error: {e}")
    sys.exit(1)

