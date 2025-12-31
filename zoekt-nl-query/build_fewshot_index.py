#!/usr/bin/env python3
"""
Build ChromaDB vector database for few-shot example retrieval.
Uses sentence-transformers for embeddings (CPU-friendly).
"""

import json
import os
import sys
from typing import List, Dict, Tuple

try:
    import chromadb
    from chromadb.config import Settings
    from sentence_transformers import SentenceTransformer
except ImportError:
    print("Installing required packages...")
    os.system("pip install chromadb sentence-transformers -q")
    import chromadb
    from chromadb.config import Settings
    from sentence_transformers import SentenceTransformer

# Initialize embedding model (CPU-friendly)
print("Loading embedding model...")
embedding_model = SentenceTransformer('sentence-transformers/all-MiniLM-L6-v2')

# Initialize ChromaDB (new API)
chroma_client = chromadb.PersistentClient(path="./chroma_db")

def load_examples(jsonl_file: str = "zoekt_examples.jsonl") -> List[Dict]:
    """Load examples from JSONL file"""
    examples = []
    if not os.path.exists(jsonl_file):
        print(f"❌ File not found: {jsonl_file}")
        print("   Run generate_examples.py first to create the dataset")
        return []
    
    with open(jsonl_file, 'r') as f:
        for line in f:
            if line.strip():
                try:
                    example = json.loads(line)
                    # Ensure it's a dict, not a list
                    if isinstance(example, dict):
                        examples.append(example)
                except json.JSONDecodeError:
                    continue
    
    return examples

def build_index(examples: List[Dict], collection_name: str = "zoekt_examples"):
    """Build ChromaDB index from examples"""
    # Delete existing collection if it exists
    try:
        chroma_client.delete_collection(collection_name)
    except:
        pass
    
    # Create new collection
    collection = chroma_client.create_collection(
        name=collection_name,
        metadata={"description": "NL-to-Zoekt query examples for few-shot retrieval"}
    )
    
    print(f"📦 Indexing {len(examples)} examples...")
    
    # Prepare data
    ids = []
    documents = []
    metadatas = []
    
    for i, example in enumerate(examples):
        instruction = example.get("instruction", "")
        output = example.get("output", "")
        
        # Use instruction as the document (what we'll search against)
        documents.append(instruction)
        
        # Store full example in metadata
        metadatas.append({
            "instruction": instruction,
            "output": output,
            "index": i
        })
        
        ids.append(f"example_{i}")
    
    # Add to collection
    collection.add(
        ids=ids,
        documents=documents,
        metadatas=metadatas
    )
    
    # PersistentClient auto-persists, no need to call persist()
    print(f"✅ Indexed {len(examples)} examples in collection '{collection_name}'")
    return collection

def get_examples(query: str, n: int = 5, collection_name: str = "zoekt_examples") -> List[Dict]:
    """
    Retrieve top N similar examples for a given query.
    This is the function that will be called by the translation service.
    """
    try:
        collection = chroma_client.get_collection(collection_name)
    except:
        print(f"❌ Collection '{collection_name}' not found. Run build_fewshot_index.py first.")
        return []
    
    # Query the collection
    results = collection.query(
        query_texts=[query],
        n_results=n
    )
    
    # Format results
    examples = []
    if results['metadatas'] and len(results['metadatas']) > 0:
        for metadata in results['metadatas'][0]:
            examples.append({
                "instruction": metadata["instruction"],
                "output": metadata["output"]
            })
    
    return examples

if __name__ == "__main__":
    print("🔨 Building Few-Shot Example Index\n")
    
    # Load examples
    examples = load_examples()
    
    if not examples:
        sys.exit(1)
    
    # Build index
    collection = build_index(examples)
    
    # Test retrieval
    print("\n🧪 Testing retrieval...")
    test_queries = [
        "find all Python functions",
        "search for TODO comments",
        "find configure_gemini in services.py"
    ]
    
    for test_query in test_queries:
        results = get_examples(test_query, n=3)
        print(f"\nQuery: '{test_query}'")
        print(f"Retrieved {len(results)} examples:")
        for i, ex in enumerate(results, 1):
            print(f"  {i}. {ex['instruction']} → {ex['output']}")
    
    print("\n✅ Index built successfully!")
    print("📁 Database saved to: ./chroma_db/")

