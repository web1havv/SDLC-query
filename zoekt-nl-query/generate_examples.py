#!/usr/bin/env python3
"""
Generate 1,000 synthetic NL-to-Zoekt query pairs using GPT-4o-mini via OpenRouter.
This creates the dataset for the few-shot RAG system.
"""

import json
import os
import requests
from typing import List, Dict

OPENROUTER_API_KEY = os.getenv("OPENROUTER_API_KEY")
if not OPENROUTER_API_KEY:
    raise ValueError("OPENROUTER_API_KEY environment variable is required")
OPENROUTER_URL = "https://openrouter.ai/api/v1/chat/completions"

# Read Zoekt query syntax documentation
def load_zoekt_syntax():
    syntax_path = "../zoekt/doc/query_syntax.md"
    try:
        with open(syntax_path, 'r') as f:
            return f.read()
    except:
        return """
Zoekt Query Syntax:
- Fields: repo:, file:, lang:, sym:, content:
- Operators: Space = AND, 'or' = OR, '-' = NOT
- Regex: Wrap in forward slashes: /pattern/
- Examples: lang:python sym:def, file:test.py content:error
"""

def generate_examples_batch(prompt: str, num_examples: int = 50) -> List[Dict]:
    """Generate a batch of examples using a free model"""
    payload = {
        "model": "mistralai/mistral-7b-instruct:free",
        "messages": [
            {
                "role": "system",
                "content": "You are a Zoekt query syntax expert. Generate diverse natural language questions and their exact Zoekt query equivalents. Always return valid JSON arrays only, no markdown formatting."
            },
            {
                "role": "user",
                "content": prompt
            }
        ],
        "temperature": 0.7,  # Slightly lower for more consistent JSON output
        "max_tokens": 2000
    }
    
    headers = {
        "Authorization": f"Bearer {OPENROUTER_API_KEY}",
        "Content-Type": "application/json"
    }
    
    try:
        response = requests.post(OPENROUTER_URL, json=payload, headers=headers, timeout=30)
        response.raise_for_status()
        result = response.json()
        
        content = result["choices"][0]["message"]["content"]
        # Parse the JSON array from the response
        # Clean up markdown code blocks if present
        content = content.strip()
        if "```json" in content:
            content = content.split("```json")[1].split("```")[0].strip()
        elif "```" in content:
            # Try to extract JSON from code block
            parts = content.split("```")
            for part in parts:
                part = part.strip()
                if part.startswith("[") and part.endswith("]"):
                    content = part
                    break
            else:
                content = parts[1] if len(parts) > 1 else content
        
        # Remove any leading/trailing non-JSON text
        start_idx = content.find("[")
        end_idx = content.rfind("]")
        if start_idx != -1 and end_idx != -1 and end_idx > start_idx:
            content = content[start_idx:end_idx+1]
        
        try:
            examples = json.loads(content)
            return examples if isinstance(examples, list) else []
        except json.JSONDecodeError as e:
            print(f"  ⚠️ JSON parse error: {e}")
            print(f"  Content preview: {content[:200]}...")
            return []
    except Exception as e:
        print(f"Error generating batch: {e}")
        return []

def generate_all_examples():
    """Generate 1,000 examples in batches"""
    zoekt_syntax = load_zoekt_syntax()
    
    all_examples = []
    batch_size = 50
    num_batches = 20  # 20 batches * 50 = 1000 examples
    
    categories = [
        "Python function searches",
        "JavaScript/TypeScript searches",
        "Go code searches",
        "File-specific searches",
        "Repository filters",
        "Symbol searches",
        "Comment searches",
        "Regex patterns",
        "Complex boolean queries",
        "Language-specific patterns",
        "File exclusion patterns",
        "Multi-repository searches",
        "Branch-specific searches",
        "Case-sensitive searches",
        "Content searches with context"
    ]
    
    for batch_num in range(num_batches):
        category = categories[batch_num % len(categories)]
        
        prompt = f"""Based on this Zoekt query syntax documentation:

{zoekt_syntax[:2000]}

Generate exactly {batch_size} diverse examples of natural language questions and their Zoekt query equivalents. Focus on: {category}

IMPORTANT: Return ONLY a valid JSON array, no markdown, no explanations, no code blocks.

Format (JSON array only):
[
  {{"instruction": "Find all Python functions", "output": "lang:python content:def "}},
  {{"instruction": "Search for TODO comments in Go files", "output": "lang:go content:/\\/\\/.*TODO/"}},
  {{"instruction": "find configure_gemini in services.py", "output": "file:services.py configure_gemini"}},
  ...
]

Make each example unique and cover different Zoekt syntax patterns. Ensure all regex patterns use forward slashes."""
        
        print(f"Generating batch {batch_num + 1}/{num_batches} ({category})...")
        examples = generate_examples_batch(prompt, batch_size)
        
        if examples:
            all_examples.extend(examples)
            print(f"  ✓ Generated {len(examples)} examples (total: {len(all_examples)})")
        else:
            print(f"  ✗ Failed to generate batch {batch_num + 1}")
        
        # Rate limiting - be nice to the API (free models may need more time)
        import time
        time.sleep(2)  # Increased delay for free tier
    
    return all_examples

def save_examples(examples: List[Dict], filename: str = "zoekt_examples.jsonl"):
    """Save examples to JSONL format"""
    with open(filename, 'w') as f:
        for example in examples:
            f.write(json.dumps(example) + '\n')
    print(f"\n✓ Saved {len(examples)} examples to {filename}")

if __name__ == "__main__":
    print("🚀 Generating 1,000 NL-to-Zoekt query examples...")
    print("This will use mistralai/mistral-7b-instruct:free via OpenRouter API\n")
    print("⚠️  Note: Free models may take longer and have rate limits\n")
    
    examples = generate_all_examples()
    
    if examples:
        save_examples(examples)
        print(f"\n✅ Successfully generated {len(examples)} examples!")
        print(f"📁 Saved to: zoekt_examples.jsonl")
    else:
        print("\n❌ Failed to generate examples")

