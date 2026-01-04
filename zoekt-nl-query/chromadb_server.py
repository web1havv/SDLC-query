#!/usr/bin/env python3
"""
Simple ChromaDB HTTP server wrapper
Starts a minimal HTTP server that proxies to ChromaDB's internal API
"""
import chromadb
from chromadb.config import Settings
from http.server import HTTPServer, BaseHTTPRequestHandler
import json
import sys
import os
from urllib.parse import urlparse, parse_qs

chroma_path = "/Users/web1havv/SDLC_AI/zoekt-nl-query/chroma_db"
port = 8000

# Initialize ChromaDB client
print(f"Initializing ChromaDB at {chroma_path}...")
try:
    client = chromadb.PersistentClient(path=chroma_path)
    print(f"✅ ChromaDB client initialized")
except Exception as e:
    print(f"❌ Failed to initialize ChromaDB: {e}")
    sys.exit(1)

collection_name = "codebase_chunks"

# Get or create collection
try:
    collection = client.get_or_create_collection(name=collection_name)
    print(f"✅ Collection '{collection_name}' ready")
except Exception as e:
    print(f"❌ Failed to get/create collection: {e}")
    sys.exit(1)

class ChromaDBHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/api/v1/heartbeat":
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(b'{"nanosecond heartbeat": 1}')
        elif "/api/v1/collections/" in self.path and "/count" in self.path:
            # Get collection count
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            count = collection.count()
            self.wfile.write(json.dumps({"count": count}).encode())
        else:
            self.send_response(404)
            self.end_headers()
    
    def do_POST(self):
        if "/api/v1/collections" in self.path and "/query" in self.path:
            # Handle query
            content_length = int(self.headers.get('Content-Length', 0))
            body = self.rfile.read(content_length)
            data = json.loads(body)
            
            query_embeddings = data.get("query_embeddings", [])
            n_results = data.get("n_results", 10)
            
            if query_embeddings and len(query_embeddings) > 0:
                results = collection.query(
                    query_embeddings=query_embeddings[0],
                    n_results=n_results
                )
                
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps(results).encode())
            else:
                self.send_response(400)
                self.end_headers()
        elif "/api/v1/collections" in self.path and "/add" in self.path:
            # Handle add
            content_length = int(self.headers.get('Content-Length', 0))
            body = self.rfile.read(content_length)
            data = json.loads(body)
            
            ids = data.get("ids", [])
            embeddings = data.get("embeddings", [])
            documents = data.get("documents", [])
            metadatas = data.get("metadatas", [])
            
            collection.add(
                ids=ids,
                embeddings=embeddings,
                documents=documents,
                metadatas=metadatas
            )
            
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(b'{"status": "ok"}')
        elif "/api/v1/collections" in self.path:
            # Create collection
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps({"name": collection_name}).encode())
        else:
            self.send_response(404)
            self.end_headers()
    
    def log_message(self, format, *args):
        # Suppress default logging
        pass

if __name__ == "__main__":
    server = HTTPServer(("localhost", port), ChromaDBHandler)
    print(f"🚀 ChromaDB HTTP server starting on http://localhost:{port}")
    print(f"📊 Collection: {collection_name}")
    print(f"💾 Data path: {chroma_path}")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\n🛑 Shutting down server...")
        server.shutdown()

