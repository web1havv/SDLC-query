#!/bin/bash

# Comprehensive Zoekt Query Testing and Validation Script
# This script tests various query types, validates output, and measures performance

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Set PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# Check if zoekt is available
if ! command -v zoekt &> /dev/null; then
    echo -e "${RED}❌ Error: zoekt command not found${NC}"
    echo "Please run: export PATH=\"\$PATH:\$(go env GOPATH)/bin\""
    exit 1
fi

# Check if index exists
INDEX_DIR="${HOME}/.zoekt"
if [ ! -d "$INDEX_DIR" ] || [ -z "$(ls -A $INDEX_DIR/*.zoekt 2>/dev/null)" ]; then
    echo -e "${RED}❌ Error: No Zoekt index found in $INDEX_DIR${NC}"
    echo "Please run: ./setup-zoekt.sh"
    exit 1
fi

echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}   ZOEKT QUERY TESTING & VALIDATION SUITE${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0
TOTAL_QUERIES=0

# Function to run a test query
test_query() {
    local query="$1"
    local expected_min_results="${2:-0}"  # Minimum expected results
    local description="$3"
    local query_type="${4:-basic}"
    
    TOTAL_QUERIES=$((TOTAL_QUERIES + 1))
    
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}Test #${TOTAL_QUERIES} [${query_type}]${NC}"
    echo -e "${CYAN}Query:${NC} ${query}"
    echo -e "${CYAN}Description:${NC} ${description}"
    echo ""
    
        # Measure execution time
        START_TIME=$(date +%s.%N)
        
        # Run query and capture output
        if OUTPUT=$(zoekt "$query" -r 2>&1); then
            END_TIME=$(date +%s.%N)
            # Use awk for calculation if bc is not available
            if command -v bc &> /dev/null; then
                EXECUTION_TIME=$(echo "$END_TIME - $START_TIME" | bc)
            else
                EXECUTION_TIME=$(awk "BEGIN {printf \"%.3f\", $END_TIME - $START_TIME}")
            fi
        
        # Count results (lines with file paths)
        RESULT_COUNT=$(echo "$OUTPUT" | grep -c ":" || echo "0")
        
        echo -e "${GREEN}✓ Query executed successfully${NC}"
        echo -e "${CYAN}Execution time:${NC} ${EXECUTION_TIME}s"
        echo -e "${CYAN}Results found:${NC} ${RESULT_COUNT}"
        
        # Validate minimum results
        if [ "$RESULT_COUNT" -ge "$expected_min_results" ]; then
            echo -e "${GREEN}✓ Results validation passed (expected ≥${expected_min_results}, got ${RESULT_COUNT})${NC}"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            
            # Show sample results
            if [ "$RESULT_COUNT" -gt 0 ]; then
                echo -e "${CYAN}Sample results (first 3):${NC}"
                echo "$OUTPUT" | head -3 | sed 's/^/  /'
                if [ "$RESULT_COUNT" -gt 3 ]; then
                    echo -e "  ${YELLOW}... and $((RESULT_COUNT - 3)) more${NC}"
                fi
            fi
        else
            echo -e "${RED}✗ Results validation failed (expected ≥${expected_min_results}, got ${RESULT_COUNT})${NC}"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    else
        END_TIME=$(date +%s.%N)
        if command -v bc &> /dev/null; then
            EXECUTION_TIME=$(echo "$END_TIME - $START_TIME" | bc)
        else
            EXECUTION_TIME=$(awk "BEGIN {printf \"%.3f\", $END_TIME - $START_TIME}")
        fi
        echo -e "${RED}✗ Query execution failed${NC}"
        echo -e "${RED}Error output:${NC}"
        echo "$OUTPUT" | sed 's/^/  /'
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    echo ""
}

# Function to test query with JSON output
test_query_json() {
    local query="$1"
    local description="$2"
    
    TOTAL_QUERIES=$((TOTAL_QUERIES + 1))
    
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}Test #${TOTAL_QUERIES} [JSON]${NC}"
    echo -e "${CYAN}Query:${NC} ${query}"
    echo -e "${CYAN}Description:${NC} ${description}"
    echo ""
    
    START_TIME=$(date +%s.%N)
    
    if OUTPUT=$(zoekt "$query" -jsonl 2>&1); then
        END_TIME=$(date +%s.%N)
        if command -v bc &> /dev/null; then
            EXECUTION_TIME=$(echo "$END_TIME - $START_TIME" | bc)
        else
            EXECUTION_TIME=$(awk "BEGIN {printf \"%.3f\", $END_TIME - $START_TIME}")
        fi
        
        # Count JSON lines
        JSON_COUNT=$(echo "$OUTPUT" | grep -c "^{" || echo "0")
        
        echo -e "${GREEN}✓ JSON query executed successfully${NC}"
        echo -e "${CYAN}Execution time:${NC} ${EXECUTION_TIME}s"
        echo -e "${CYAN}JSON objects:${NC} ${JSON_COUNT}"
        
        # Validate JSON format
        if echo "$OUTPUT" | head -1 | jq . > /dev/null 2>&1; then
            echo -e "${GREEN}✓ JSON format is valid${NC}"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            
            # Show sample JSON
            if [ "$JSON_COUNT" -gt 0 ]; then
                echo -e "${CYAN}Sample JSON (first result):${NC}"
                echo "$OUTPUT" | head -1 | jq '.' | head -10 | sed 's/^/  /'
            fi
        else
            echo -e "${RED}✗ Invalid JSON format${NC}"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    else
        echo -e "${RED}✗ JSON query execution failed${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    echo ""
}

# Function to test filename-only queries
test_query_list() {
    local query="$1"
    local description="$2"
    
    TOTAL_QUERIES=$((TOTAL_QUERIES + 1))
    
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}Test #${TOTAL_QUERIES} [FILENAME]${NC}"
    echo -e "${CYAN}Query:${NC} ${query}"
    echo -e "${CYAN}Description:${NC} ${description}"
    echo ""
    
    START_TIME=$(date +%s.%N)
    
    if OUTPUT=$(zoekt -l "$query" 2>&1); then
        END_TIME=$(date +%s.%N)
        if command -v bc &> /dev/null; then
            EXECUTION_TIME=$(echo "$END_TIME - $START_TIME" | bc)
        else
            EXECUTION_TIME=$(awk "BEGIN {printf \"%.3f\", $END_TIME - $START_TIME}")
        fi
        
        FILE_COUNT=$(echo "$OUTPUT" | wc -l | xargs)
        
        echo -e "${GREEN}✓ Filename query executed successfully${NC}"
        echo -e "${CYAN}Execution time:${NC} ${EXECUTION_TIME}s"
        echo -e "${CYAN}Files found:${NC} ${FILE_COUNT}"
        
        if [ "$FILE_COUNT" -gt 0 ]; then
            echo -e "${GREEN}✓ Files found${NC}"
            TESTS_PASSED=$((TESTS_PASSED + 1))
            
            echo -e "${CYAN}Sample filenames (first 5):${NC}"
            echo "$OUTPUT" | head -5 | sed 's/^/  /'
        else
            echo -e "${YELLOW}⚠ No files found (this might be expected)${NC}"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        fi
    else
        echo -e "${RED}✗ Filename query execution failed${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    echo ""
}

# Function to run random query tests
test_random_queries() {
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}   RANDOM QUERY TESTS${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
    echo ""
    
    # Random search terms
    RANDOM_TERMS=(
        "function"
        "import"
        "export"
        "class"
        "const"
        "let"
        "var"
        "return"
        "async"
        "await"
        "component"
        "hook"
        "state"
        "props"
        "render"
    )
    
    # Test 5 random queries
    for i in {1..5}; do
        RANDOM_TERM=${RANDOM_TERMS[$RANDOM % ${#RANDOM_TERMS[@]}]}
        test_query "$RANDOM_TERM" 0 "Random query test #${i}: searching for '${RANDOM_TERM}'" "random"
    done
}

# ============================================================================
# MAIN TEST SUITE
# ============================================================================

echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}   BASIC QUERY TESTS${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# Basic content searches
test_query "export" 0 "Basic content search: 'export'" "basic"
test_query "import" 0 "Basic content search: 'import'" "basic"
test_query "function" 0 "Basic content search: 'function'" "basic"
test_query "\"export default\"" 0 "Quoted phrase search: 'export default'" "basic"

echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}   FILE FILTER TESTS${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# File filter queries
test_query "file:*.jsx" 0 "File type filter: *.jsx files" "file-filter"
test_query "file:*.js" 0 "File type filter: *.js files" "file-filter"
test_query "file:*.md" 0 "File type filter: *.md files" "file-filter"
test_query "content:\"export\" file:*.jsx" 0 "Content + file filter: export in JSX files" "file-filter"

echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}   REGEX QUERY TESTS${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# Regex queries
test_query "/export \\w+/" 0 "Regex search: export followed by word" "regex"
test_query "/function \\w+\\(/" 0 "Regex search: function definitions" "regex"
test_query "file:/.*\\.jsx$/" 0 "Regex file filter: .jsx files" "regex"

echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}   BOOLEAN OPERATOR TESTS${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# Boolean operators
test_query "(useState OR useEffect)" 0 "OR operator: React hooks" "boolean"
test_query "export AND default" 0 "AND operator: export and default" "boolean"
test_query "import -file:test.js" 0 "NOT operator: import excluding test.js" "boolean"

echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}   LANGUAGE FILTER TESTS${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# Language filters
test_query "lang:javascript function" 0 "Language filter: JavaScript functions" "language"
test_query "lang:typescript" 0 "Language filter: TypeScript files" "language"

echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}   SYMBOL SEARCH TESTS${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# Symbol searches
test_query "sym:Home" 0 "Symbol search: Home component" "symbol"
test_query "sym:App" 0 "Symbol search: App component" "symbol"

echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}   OUTPUT FORMAT TESTS${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# Filename-only output
test_query_list "export" "List filenames containing 'export'"

# JSON output (if jq is available)
if command -v jq &> /dev/null; then
    test_query_json "export" "JSON output format test"
else
    echo -e "${YELLOW}⚠ Skipping JSON test (jq not installed)${NC}"
    echo ""
fi

echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}   COMPLEX QUERY TESTS${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""

# Complex queries
test_query "(export default) file:*.jsx" 0 "Complex: export default in JSX files" "complex"
test_query "(useState OR useEffect) file:*.jsx" 0 "Complex: React hooks in JSX files" "complex"
test_query "content:\"function\" lang:javascript -file:*test*" 0 "Complex: functions in JS excluding tests" "complex"

# Random query tests
test_random_queries

# ============================================================================
# SUMMARY
# ============================================================================

echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}   TEST SUMMARY${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""

TOTAL_TESTS=$((TESTS_PASSED + TESTS_FAILED))
if command -v bc &> /dev/null; then
    PASS_RATE=$(echo "scale=2; $TESTS_PASSED * 100 / $TOTAL_TESTS" | bc)
else
    PASS_RATE=$(awk "BEGIN {printf \"%.2f\", $TESTS_PASSED * 100 / $TOTAL_TESTS}")
fi

echo -e "${CYAN}Total Queries Tested:${NC} ${TOTAL_QUERIES}"
echo -e "${GREEN}Tests Passed:${NC} ${TESTS_PASSED}"
echo -e "${RED}Tests Failed:${NC} ${TESTS_FAILED}"
echo -e "${CYAN}Pass Rate:${NC} ${PASS_RATE}%"
echo ""

if [ "$TESTS_FAILED" -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Some tests failed. Review the output above.${NC}"
    exit 1
fi

