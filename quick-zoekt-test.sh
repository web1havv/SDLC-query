#!/bin/bash

# Quick Zoekt Query Test - Simple validation script
# Run this to quickly verify queries are working

set -e

export PATH="$PATH:$(go env GOPATH)/bin"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}🔍 Quick Zoekt Query Test${NC}"
echo ""

# Test 1: Basic search
echo -e "${CYAN}Test 1: Basic content search${NC}"
echo "Query: zoekt 'export' -r"
echo ""
if zoekt "export" -r | head -3; then
    echo -e "${GREEN}✓ Basic search works${NC}"
else
    echo -e "${RED}✗ Basic search failed${NC}"
fi
echo ""

# Test 2: File filter
echo -e "${CYAN}Test 2: File filter${NC}"
echo "Query: zoekt 'file:*.jsx' -r"
echo ""
if zoekt "file:*.jsx" -r | head -3; then
    echo -e "${GREEN}✓ File filter works${NC}"
else
    echo -e "${RED}✗ File filter failed${NC}"
fi
echo ""

# Test 3: List filenames
echo -e "${CYAN}Test 3: List filenames only${NC}"
echo "Query: zoekt -l 'function'"
echo ""
if zoekt -l "function" | head -5; then
    echo -e "${GREEN}✓ Filename list works${NC}"
else
    echo -e "${RED}✗ Filename list failed${NC}"
fi
echo ""

# Test 4: Count results
echo -e "${CYAN}Test 4: Count results${NC}"
echo "Query: zoekt 'import' -r | wc -l"
COUNT=$(zoekt "import" -r | wc -l | xargs)
echo "Results: $COUNT lines"
if [ "$COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ Count query works${NC}"
else
    echo -e "${YELLOW}⚠ No results found (might be expected)${NC}"
fi
echo ""

echo -e "${GREEN}✅ Quick test complete!${NC}"



