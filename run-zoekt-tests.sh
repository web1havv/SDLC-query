#!/bin/bash

# Terminal Commands for Testing Zoekt Queries
# Copy and paste these commands directly into your terminal

export PATH="$PATH:$(go env GOPATH)/bin"

echo "🔍 ZOEKT QUERY TESTING - TERMINAL COMMANDS"
echo "=========================================="
echo ""

# Test 1: Basic search
echo "Test 1: Basic content search"
echo "Command: zoekt \"export\" -r"
echo "---"
zoekt "export" -r | head -5
echo ""

# Test 2: File filter
echo "Test 2: File type filter"
echo "Command: zoekt \"file:*.jsx\" -r"
echo "---"
zoekt "file:*.jsx" -r | head -5
echo ""

# Test 3: List filenames
echo "Test 3: List filenames only"
echo "Command: zoekt -l \"function\""
echo "---"
zoekt -l "function" | head -5
echo ""

# Test 4: Count results
echo "Test 4: Count results"
echo "Command: zoekt \"import\" -r | wc -l"
echo "---"
COUNT=$(zoekt "import" -r | wc -l | xargs)
echo "Total results: $COUNT"
echo ""

# Test 5: Boolean OR
echo "Test 5: Boolean OR query"
echo "Command: zoekt \"(useState OR useEffect)\" -r"
echo "---"
zoekt "(useState OR useEffect)" -r | head -5
echo ""

# Test 6: Regex pattern
echo "Test 6: Regex pattern"
echo "Command: zoekt \"/export \\\\w+/\" -r"
echo "---"
zoekt "/export \\w+/" -r | head -5
echo ""

# Test 7: Complex query
echo "Test 7: Complex query"
echo "Command: zoekt \"(export default) file:*.jsx\" -r"
echo "---"
zoekt "(export default) file:*.jsx" -r | head -5
echo ""

# Test 8: Performance test
echo "Test 8: Performance test"
echo "Command: time zoekt \"export\" -r > /dev/null"
echo "---"
time zoekt "export" -r > /dev/null
echo ""

# Test 9: JSON output (if jq available)
if command -v jq &> /dev/null; then
    echo "Test 9: JSON output"
    echo "Command: zoekt \"export\" -jsonl | jq . | head -20"
    echo "---"
    zoekt "export" -jsonl | jq . | head -20
    echo ""
else
    echo "Test 9: JSON output (skipped - jq not installed)"
    echo "Install jq: brew install jq"
    echo ""
fi

# Test 10: Random queries
echo "Test 10: Random query tests"
echo "---"
RANDOM_TERMS=("function" "import" "export" "class" "const")
for term in "${RANDOM_TERMS[@]}"; do
    echo "Testing: $term"
    zoekt "$term" -r | head -2
    echo ""
done

echo "✅ All tests complete!"
echo ""
echo "💡 Tip: Run individual commands from above to test specific queries"
echo "💡 Tip: Use | head -N to limit results"
echo "💡 Tip: Use | wc -l to count results"



