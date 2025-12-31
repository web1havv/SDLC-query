#!/bin/bash

# Script to demonstrate expected Zoekt query outputs
# This shows what different query types should produce

export PATH="$PATH:$(go env GOPATH)/bin"

# Colors
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}   ZOEKT QUERY OUTPUT EXAMPLES${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""

echo -e "${YELLOW}This script demonstrates what output you should expect from different query types.${NC}"
echo ""

# Example 1: Basic query
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Example 1: Basic Content Search${NC}"
echo -e "${CYAN}Command:${NC} zoekt \"export\" -r"
echo ""
echo -e "${CYAN}Expected Output Format:${NC}"
echo "  repository:file/path.jsx:line_number:content snippet"
echo ""
echo -e "${CYAN}Example Output:${NC}"
echo "  github.com/user/repo:src/App.jsx:1:export default function App() {"
echo "  github.com/user/repo:src/components/Home.jsx:1:export default function Home() {"
echo "  github.com/user/repo:src/utils.js:5:export const helper = () => {"
echo ""

# Example 2: Filename list
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Example 2: Filename-Only Output${NC}"
echo -e "${CYAN}Command:${NC} zoekt -l \"function\""
echo ""
echo -e "${CYAN}Expected Output Format:${NC}"
echo "  One filename per line (no line numbers, no content)"
echo ""
echo -e "${CYAN}Example Output:${NC}"
echo "  src/components/App.jsx"
echo "  src/components/Home.jsx"
echo "  src/utils/helpers.js"
echo "  src/index.js"
echo ""

# Example 3: File filter
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Example 3: File Type Filter${NC}"
echo -e "${CYAN}Command:${NC} zoekt \"file:*.jsx\" -r"
echo ""
echo -e "${CYAN}Expected Output Format:${NC}"
echo "  repository:file.jsx:line_number:content"
echo ""
echo -e "${CYAN}Example Output:${NC}"
echo "  github.com/user/repo:src/App.jsx:1:import React from 'react';"
echo "  github.com/user/repo:src/components/Home.jsx:1:export default function Home() {"
echo "  github.com/user/repo:src/components/Nav.jsx:1:export default function Nav() {"
echo ""

# Example 4: Boolean OR
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Example 4: Boolean OR Query${NC}"
echo -e "${CYAN}Command:${NC} zoekt \"(useState OR useEffect)\" -r"
echo ""
echo -e "${CYAN}Expected Output Format:${NC}"
echo "  Results matching either useState OR useEffect"
echo ""
echo -e "${CYAN}Example Output:${NC}"
echo "  github.com/user/repo:src/App.jsx:5:const [state, setState] = useState();"
echo "  github.com/user/repo:src/App.jsx:10:useEffect(() => {"
echo "  github.com/user/repo:src/components/Home.jsx:3:const [count, setCount] = useState(0);"
echo ""

# Example 5: Regex
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Example 5: Regex Pattern${NC}"
echo -e "${CYAN}Command:${NC} zoekt \"/export \\\\w+/\" -r"
echo ""
echo -e "${CYAN}Expected Output Format:${NC}"
echo "  Matches regex pattern: export followed by word characters"
echo ""
echo -e "${CYAN}Example Output:${NC}"
echo "  github.com/user/repo:src/App.jsx:1:export default"
echo "  github.com/user/repo:src/utils.js:5:export const"
echo "  github.com/user/repo:src/components/Home.jsx:1:export function"
echo ""

# Example 6: JSON output
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Example 6: JSON Output${NC}"
echo -e "${CYAN}Command:${NC} zoekt \"export\" -jsonl | jq ."
echo ""
echo -e "${CYAN}Expected Output Format:${NC}"
echo "  One JSON object per line (JSON Lines format)"
echo ""
echo -e "${CYAN}Example Output:${NC}"
cat << 'EOF'
  {
    "Repository": "github.com/user/repo",
    "FileName": "src/App.jsx",
    "LineMatches": [
      {
        "LineNumber": 1,
        "LineFragments": [
          {
            "Offset": 0,
            "Line": "export default function App() {"
          }
        ]
      }
    ],
    "Score": 1.5
  }
EOF
echo ""

# Example 7: No results
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Example 7: No Results Found${NC}"
echo -e "${CYAN}Command:${NC} zoekt \"nonexistent_term_xyz123\" -r"
echo ""
echo -e "${CYAN}Expected Output Format:${NC}"
echo "  (empty output - no lines)"
echo ""
echo -e "${CYAN}Example Output:${NC}"
echo "  (nothing printed)"
echo ""
echo -e "${YELLOW}Note:${NC} This is normal if the search term doesn't exist in your repository."
echo ""

# Example 8: Count results
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Example 8: Counting Results${NC}"
echo -e "${CYAN}Command:${NC} zoekt \"export\" -r | wc -l"
echo ""
echo -e "${CYAN}Expected Output Format:${NC}"
echo "  A single number (count of result lines)"
echo ""
echo -e "${CYAN}Example Output:${NC}"
echo "  23"
echo ""
echo -e "${YELLOW}Note:${NC} This counts total matches, not unique files."
echo ""

# Example 9: Complex query
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}Example 9: Complex Query${NC}"
echo -e "${CYAN}Command:${NC} zoekt \"(export default) file:*.jsx\" -r"
echo ""
echo -e "${CYAN}Expected Output Format:${NC}"
echo "  Results matching 'export default' in .jsx files only"
echo ""
echo -e "${CYAN}Example Output:${NC}"
echo "  github.com/user/repo:src/App.jsx:1:export default function App() {"
echo "  github.com/user/repo:src/components/Home.jsx:1:export default function Home() {"
echo "  github.com/user/repo:src/components/Nav.jsx:1:export default function Nav() {"
echo ""

echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Summary${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${YELLOW}Key Points:${NC}"
echo "  • Default format: file:line:content"
echo "  • With -r flag: repo:file:line:content"
echo "  • With -l flag: filename only (one per line)"
echo "  • With -jsonl flag: JSON objects (one per line)"
echo "  • Empty output = no matches (this is normal)"
echo "  • Each line = one match"
echo ""
echo -e "${GREEN}To test your queries:${NC}"
echo "  ./quick-zoekt-test.sh      # Quick validation"
echo "  ./test-zoekt-queries.sh     # Comprehensive tests"
echo ""
echo -e "${CYAN}For more details, see:${NC}"
echo "  • TESTING_ZOEKT_QUERIES.md - Complete testing guide"
echo "  • ZOEKT_EXPECTED_OUTPUT.md - Detailed output reference"
echo ""



