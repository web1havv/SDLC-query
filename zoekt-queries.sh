#!/bin/bash
# Zoekt Query Commands - Reference Guide
# Make sure to add Go bin to PATH: export PATH="$PATH:$(go env GOPATH)/bin"

echo "🔍 ZOEKT QUERY COMMANDS FOR YOUR REPOSITORY"
echo "============================================"
echo ""

# Set PATH
export PATH="$PATH:$(go env GOPATH)/bin"

echo "📚 BLOG QUERIES"
echo "---------------"
echo ""
echo "# Find all blog titles"
echo 'zoekt "Amazon ML Challenge" -r'
echo ""
zoekt "Amazon ML Challenge" -r | head -5

echo ""
echo "# Search for specific blog: Mistakes"
echo 'zoekt "Mistakes I Made" -r'
echo ""
zoekt "Mistakes I Made" -r | head -3

echo ""
echo "# Search for AI Code Agents blog"
echo 'zoekt "AI Code Agents" -r'
echo ""
zoekt "AI Code Agents" -r | head -3

echo ""
echo "# Find methodology blog"
echo 'zoekt "How We Built" -r'
echo ""
zoekt "How We Built" -r | head -3

echo ""
echo "# Find blog dates"
echo 'zoekt "2025-12" -r'
echo ""
zoekt "2025-12" -r | head -3

echo ""
echo "# Search blog topics"
echo 'zoekt "methodology" -r'
echo ""
zoekt "methodology" -r | head -3

echo ""
echo "# Find learned/lessons"
echo 'zoekt "learned" -r'
echo ""
zoekt "learned" -r | head -3

echo ""
echo "# Multiple topics (OR)"
echo 'zoekt "(Amazon OR ML OR Challenge)" -r'
echo ""
zoekt "(Amazon OR ML OR Challenge)" -r | head -5

echo ""
echo "============================================"
echo ""
echo "💻 CODE QUERIES"
echo "---------------"
echo ""
echo "# Find all React components"
echo 'zoekt "export default function" -r'
echo ""
zoekt "export default function" -r | head -5

echo ""
echo "# Search for imports"
echo 'zoekt "import" -r'
echo ""
zoekt "import" -r | head -5

echo ""
echo "# Find specific component"
echo 'zoekt "function Home" -r'
echo ""
zoekt "function Home" -r | head -3

echo ""
echo "# Find in specific file type"
echo 'zoekt "function file:*.jsx" -r'
echo ""
zoekt "function file:*.jsx" -r | head -3

echo ""
echo "# Find React hooks"
echo 'zoekt "(useState OR useEffect)" -r'
echo ""
zoekt "(useState OR useEffect)" -r | head -3

echo ""
echo "# Regex: Find function definitions"
echo 'zoekt "/export \w+/" -r'
echo ""
zoekt "/export \w+/" -r | head -3

echo ""
echo "# List matching files only (no content)"
echo 'zoekt -l "function"'
echo ""
zoekt -l "function" | head -5

echo ""
echo "# Exclude node_modules"
echo 'zoekt "import" -r | grep -v node_modules'
echo ""
zoekt "import" -r | grep -v node_modules | head -3

echo ""
echo "============================================"
echo ""
echo "📊 STATISTICS & COUNTS"
echo "----------------------"
echo ""
echo "# Count total matches for a term"
echo 'zoekt "export" -r | wc -l'
echo "Results: $(zoekt "export" -r | wc -l | xargs) lines"

echo ""
echo "# Count unique files with matches"
echo 'zoekt -l "function" | wc -l'
echo "Results: $(zoekt -l "function" | wc -l | xargs) files"

echo ""
echo "============================================"
echo ""
echo "🎯 ADVANCED QUERIES"
echo "-------------------"
echo ""
echo "# Complex: React components in JSX files"
echo 'zoekt "(export default) file:*.jsx" -r'
echo ""
zoekt "(export default) file:*.jsx" -r | head -3

echo ""
echo "# Find specific file"
echo 'zoekt "content file:App.jsx" -r'
echo ""
zoekt "content file:App.jsx" -r | head -3

echo ""
echo "# Case sensitive search"
echo 'zoekt "case:yes React" -r'
echo ""
zoekt "case:yes React" -r | head -3

echo ""
echo "============================================"
echo ""
echo "✅ All queries executed!"
echo ""
echo "💡 TIP: Use these commands directly in your terminal"
echo "💡 TIP: Add -r flag to show repository name"
echo "💡 TIP: Use -l flag to list only filenames"
echo "💡 TIP: Pipe to head -N to limit results"





