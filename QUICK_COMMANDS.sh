#!/bin/bash
# Quick Reference: Copy-paste these commands into your terminal

export PATH="$PATH:$(go env GOPATH)/bin"

# ===== BASIC QUERIES =====
zoekt "export" -r                                    # Basic search
zoekt "import" -r                                    # Search imports
zoekt "function" -r                                 # Search functions
zoekt "export default" -r                            # Phrase search

# ===== FILE FILTERS =====
zoekt "file:*.jsx" -r                               # JSX files only
zoekt "export file:*.jsx" -r                        # Export in JSX files
zoekt "(useState OR useEffect) file:*.jsx" -r       # React hooks in JSX

# ===== LIST & COUNT =====
zoekt -l "function"                                 # List filenames only
zoekt "export" -r | wc -l                           # Count results
zoekt -l "export" | wc -l                           # Count unique files

# ===== REGEX =====
zoekt "/export \\w+/" -r                            # Regex pattern
zoekt "/function \\w+\\(/" -r                       # Function definitions

# ===== BOOLEAN =====
zoekt "(useState OR useEffect)" -r                 # OR operator
zoekt "import -file:test.js" -r                     # NOT operator

# ===== PERFORMANCE =====
time zoekt "export" -r > /dev/null                  # Measure time

# ===== JSON OUTPUT =====
zoekt "export" -jsonl | jq .                        # JSON format
zoekt "export" -jsonl | jq -r '.FileName'           # Extract filenames

# ===== QUICK TEST =====
echo "Testing basic queries..."
zoekt "export" -r | head -3
zoekt "import" -r | head -3
zoekt "function" -r | head -3



