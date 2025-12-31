#!/usr/bin/env python3
"""
HTTP service for few-shot example retrieval.
This service runs as a separate process and the Go code calls it via HTTP.
"""

import json
import sys
import os
from flask import Flask, request, jsonify

# Change to script directory to ensure relative paths work
script_dir = os.path.dirname(os.path.abspath(__file__))
os.chdir(script_dir)

# Import after changing directory
from build_fewshot_index import get_examples

app = Flask(__name__)

@app.route('/get_examples', methods=['POST'])
def retrieve_examples():
    """Retrieve similar examples for a given query"""
    try:
        data = request.json
        query = data.get('query', '')
        n = data.get('n', 5)
        
        if not query:
            return jsonify({"error": "Missing 'query' parameter"}), 400
        
        examples = get_examples(query, n=n)
        
        return jsonify({
            "examples": examples,
            "count": len(examples)
        })
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({"status": "ok"})

if __name__ == "__main__":
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 6072
    print(f"🚀 Starting Few-Shot Retrieval Service on port {port}")
    print(f"   Endpoint: http://localhost:{port}/get_examples")
    app.run(host='0.0.0.0', port=port, debug=False)

