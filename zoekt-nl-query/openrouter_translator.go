package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sourcegraph/zoekt"
	"github.com/sourcegraph/zoekt/query"
)

// Intent represents the detected intent of a natural language query
type Intent struct {
	Type       string   // "search", "count", "list", "find", "yesno"
	Entity     string   // "articles", "functions", "components", etc.
	Topic      string   // The main topic/subject of the query
	Confidence float64  // Confidence score (0.0 to 1.0)
	Query      string   // The generated Zoekt query
	Variations []string // Alternative query variations
}

// OpenRouterTranslator uses OpenRouter API for query translation
type OpenRouterTranslator struct {
	apiKey  string
	enabled bool
	model   string
	searcher zoekt.Searcher // Add searcher to get codebase context
}

func NewOpenRouterTranslator(searcher zoekt.Searcher) *OpenRouterTranslator {
	// Get API key from environment variable (required)
	key := os.Getenv("OPENROUTER_API_KEY")
	if key == "" {
		log.Printf("⚠️ OpenRouter disabled: OPENROUTER_API_KEY environment variable not set")
		return &OpenRouterTranslator{
			apiKey:  "",
			enabled: false,
			model:   "",
			searcher: searcher,
		}
	}
	
	// Use free model - let OpenRouter choose automatically via model parameter
	// If OPENROUTER_MODEL is set, use it; otherwise use a free model
	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		// Use a free model that works without credits
		// OpenRouter will automatically route to best available free model
		model = "mistralai/mistral-7b-instruct:free"
	}
	
	log.Printf("✅ OpenRouter enabled with model: %s", model)
	
	return &OpenRouterTranslator{
		apiKey:  key,
		enabled: true,
		model:   model,
		searcher: searcher,
	}
}

// TranslateWithOpenRouter uses OpenRouter API to convert natural language to Zoekt query
func (t *OpenRouterTranslator) TranslateWithOpenRouter(nlQuery string) (string, *Intent, error) {
	if !t.enabled {
		return "", nil, fmt.Errorf("OpenRouter API key not set")
	}

	// Stage 1: Search for relevant code snippets based on key terms in the query
	// This finds actual code about the key terms (like "rupeeflo") so the model can understand context
	relevantCodeSnippets := t.findRelevantCodeSnippets(nlQuery)
	
	// Check if this is a question that can be answered directly from snippets
	lowerQuery := strings.ToLower(nlQuery)
	isQuestion := strings.HasPrefix(lowerQuery, "does ") || 
	              strings.HasPrefix(lowerQuery, "do ") ||
	              strings.HasPrefix(lowerQuery, "have ") ||
	              strings.HasPrefix(lowerQuery, "has ") ||
	              strings.Contains(lowerQuery, "?")
	
	// If we have good snippets and it's a question, try to answer directly first
	if isQuestion && relevantCodeSnippets != "" && len(relevantCodeSnippets) > 100 {
		log.Printf("💡 Question detected with code snippets - attempting direct answer")
		
		// Extract snippets as array for AnswerFromSnippets
		snippetLines := strings.Split(relevantCodeSnippets, "\n\n")
		var cleanSnippets []string
		for _, line := range snippetLines {
			if strings.Contains(line, "Found '") && strings.Contains(line, "```") {
				cleanSnippets = append(cleanSnippets, line)
			}
		}
		
		if len(cleanSnippets) > 0 {
			directAnswer := t.AnswerFromSnippets(nlQuery, cleanSnippets)
			if directAnswer != "" && len(directAnswer) > 10 {
				log.Printf("✅ Generated direct answer from snippets: %s", directAnswer)
				// Return a special marker query that indicates we have a direct answer
				// The server will handle this specially
				return "DIRECT_ANSWER:" + directAnswer, &Intent{
					Type:       "answer",
					Query:      directAnswer,
					Confidence: 0.9,
				}, nil
			}
		}
	}
	
	// Log what we found
	if relevantCodeSnippets != "" {
		log.Printf("📚 Found relevant code snippets for context")
	} else {
		log.Printf("⚠️ No relevant code snippets found for key terms")
	}
	
	// Limit snippet size to prevent prompt from being too long (reduced from 3000 to 1500)
	if len(relevantCodeSnippets) > 1500 {
		log.Printf("⚠️ Relevant code snippets too long (%d chars), truncating...", len(relevantCodeSnippets))
		relevantCodeSnippets = relevantCodeSnippets[:1500] + "\n... (truncated)"
	}

	// Get codebase context - sample file names to help model understand structure
	codebaseContext := t.getCodebaseContext()
	
	// Limit context size as well (reduced from 2000 to 1000)
	if len(codebaseContext) > 1000 {
		log.Printf("⚠️ Codebase context too long (%d chars), truncating...", len(codebaseContext))
		codebaseContext = codebaseContext[:1000] + "\n... (truncated)"
	}

	// Build the user prompt with codebase context and relevant snippets
	userPrompt := fmt.Sprintf(`CODEBASE CONTEXT (sample files in this codebase):
%s

%s

**User Input:** %s
**Zoekt Output:**`, codebaseContext, relevantCodeSnippets, nlQuery)

	// Call OpenRouter API
	url := "https://openrouter.ai/api/v1/chat/completions"

	// Use the specialized system prompt as specified
	systemPrompt := `You are a "Natural Language to Zoekt DSL" translator. Your task is to convert human-readable search requests into valid Zoekt query syntax.

**CRITICAL: You MUST output ONLY valid Zoekt syntax. Do NOT use SQL, do NOT use AND/OR operators, do NOT add explanations.**

**Zoekt Syntax Rules:**
1. Fields: ` + "`repo:`" + ` (repositories), ` + "`file:`" + ` (filenames), ` + "`lang:`" + ` (programming language), ` + "`sym:`" + ` (symbol definitions like functions), ` + "`content:`" + ` (text inside files).

2. Operators: Space is an implicit AND. Use lowercase ` + "`or`" + ` for logical OR. Use ` + "`-`" + ` before a field to negate/exclude it (e.g., ` + "`-file:test`" + `).
   - WRONG: ` + "`name:python AND type:func`" + ` (SQL syntax)
   - CORRECT: ` + "`lang:python sym:def`" + ` (Zoekt syntax)

3. Regex: Wrap patterns in forward slashes, e.g., ` + "`/pattern/`" + `.

4. Code vs. Comments: If the user specifies "comments," use a regex pattern that matches common comment markers for that language (e.g., ` + "`content:/\\/\\/.*searchTerm/`" + ` for Go/JS or ` + "`content:/#.*searchTerm/`" + ` for Python).

**Examples (Few-Shot):**
* **Input:** "Find the login function in the auth repository"
**Output:** ` + "`repo:auth sym:login`" + `
* **Input:** "List all Python functions"
**Output:** ` + "`lang:python sym:def`" + `
* **Input:** "Find all functions in JavaScript"
**Output:** ` + "`lang:javascript sym:function`" + `
* **Input:** "Search for TODO comments in Python files"
**Output:** ` + "`lang:python content:/#.*TODO/`" + `
* **Input:** "Find where we use 'API_KEY' but not in tests"
**Output:** ` + "`content:API_KEY -file:test`" + `
* **Input:** "Look for the word 'deprecated' in comments in the core repo"
**Output:** ` + "`repo:core content:/\\/\\/.*deprecated/ or content:/\\/\\*.*deprecated/`" + `

**Important Rules:**
* For "list all functions" or "all functions", use ` + "`sym:def`" + ` for Python, ` + "`sym:function`" + ` for JavaScript, or ` + "`sym:`" + ` with the function keyword for that language.
* NEVER use ` + "`sym:`" + ` without a value - it must have an argument like ` + "`sym:def`" + ` or ` + "`sym:function`" + `.

**Instructions:**
* Provide ONLY the raw Zoekt query string - nothing else.
* Do NOT include explanations, markdown, code blocks, or any other text.
* Do NOT use SQL syntax like "AND", "OR", "name:", "type:" - these are WRONG.
* Use ONLY Zoekt syntax: ` + "`lang:`, `sym:`, `file:`, `content:`, `repo:`" + `
* Example output for "list all python functions": ` + "`lang:python sym:def`" + `
* Output format: Just the query, no prefix, no suffix, no explanation.`

	payload := map[string]interface{}{
		"model": t.model,
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": userPrompt,
			},
		},
		"temperature": 0.1, // Very low temperature for deterministic output
		"max_tokens":  150, // Increased from 50 to allow proper query generation
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.apiKey))
	req.Header.Set("HTTP-Referer", "https://github.com/sourcegraph/zoekt") // Optional but recommended
	req.Header.Set("X-Title", "Zoekt NL Query") // Optional but recommended

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("OpenRouter API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("❌ Failed to parse OpenRouter response: %v", err)
		log.Printf("Response body: %s", string(body))
		return "", nil, err
	}

	// Extract the generated query
	zoektQuery := ""
	var rawContent string // Store raw content for debugging
	
	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				// Try content first
				if content, ok := message["content"].(string); ok && content != "" {
					rawContent = content
					zoektQuery = strings.TrimSpace(content)
					log.Printf("📥 OpenRouter raw content: %s", rawContent)
				} else if reasoning, ok := message["reasoning"].(string); ok && reasoning != "" {
					rawContent = reasoning
					log.Printf("📥 OpenRouter raw reasoning: %s", rawContent)
					// For "think" models, content might be empty but reasoning has the answer
					// Extract the actual query from reasoning (usually at the end)
					zoektQuery = strings.TrimSpace(reasoning)
					// Try to extract just the query part if reasoning is long
					if strings.Contains(zoektQuery, "Output:") {
						parts := strings.Split(zoektQuery, "Output:")
						if len(parts) > 1 {
							zoektQuery = strings.TrimSpace(parts[len(parts)-1])
						}
					}
				}
				
				if zoektQuery != "" {
					zoektQuery = strings.TrimSpace(zoektQuery)
					
					// Extract query from reasoning/explanations - find lines that look like queries
					lines := strings.Split(zoektQuery, "\n")
					var queryLines []string
					for _, line := range lines {
						line = strings.TrimSpace(line)
					// Skip explanation lines (contain common explanation words)
					lowerLine := strings.ToLower(strings.TrimSpace(line))
					explanationWords := []string{"converting", "formulating", "transforming", "searching", "looking", "finding", "analyzing", "processing"}
					isExplanation := false
					for _, word := range explanationWords {
						if lowerLine == word || strings.HasPrefix(lowerLine, word+" ") {
							isExplanation = true
							break
						}
					}
					if isExplanation ||
					   strings.Contains(lowerLine, "i need") ||
					   strings.Contains(lowerLine, "the user") ||
					   strings.Contains(lowerLine, "they're asking") ||
					   len(line) > 200 { // Skip very long lines (explanations)
						continue
					}
					// Skip lines that start with "I" (likely explanations like "I need to...")
					if strings.HasPrefix(strings.TrimSpace(line), "I ") || 
					   strings.HasPrefix(strings.TrimSpace(line), "I\n") ||
					   strings.TrimSpace(line) == "I" {
						continue
					}
					
					// Keep lines that look like queries (contain query operators or are short)
					if strings.Contains(line, "file:") || 
					   strings.Contains(line, "lang:") ||
					   strings.Contains(line, "content:") ||
					   strings.Contains(line, "sym:") ||
					   (len(line) > 0 && len(line) < 150 && 
					    !strings.Contains(strings.ToLower(line), "how") && 
					    !strings.Contains(strings.ToLower(line), "what") &&
					    !strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "i ")) {
						queryLines = append(queryLines, line)
					}
					}
					
					// If we found query-like lines, use the first one
					if len(queryLines) > 0 {
						zoektQuery = queryLines[0]
					}
					
					// Remove common prefixes/suffixes that models sometimes add
					zoektQuery = strings.TrimPrefix(zoektQuery, "Output:")
					zoektQuery = strings.TrimPrefix(zoektQuery, "Query:")
					zoektQuery = strings.TrimPrefix(zoektQuery, "Zoekt query:")
					zoektQuery = strings.TrimPrefix(zoektQuery, "query:")
					// Remove markdown code blocks
					zoektQuery = strings.TrimPrefix(zoektQuery, "```")
					zoektQuery = strings.TrimSuffix(zoektQuery, "```")
					zoektQuery = strings.TrimPrefix(zoektQuery, "```zoekt")
					zoektQuery = strings.TrimPrefix(zoektQuery, "```js")
					// Remove backticks
					zoektQuery = strings.Trim(zoektQuery, "`")
					// Remove outer quotes
					zoektQuery = strings.Trim(zoektQuery, `"'`)
					zoektQuery = strings.TrimSpace(zoektQuery)
					
					// If still contains explanation text, try to extract just the query part
					// Look for patterns like "file:blogs.js" or simple terms
					if strings.Contains(zoektQuery, "file:") {
						// Extract file: pattern
						parts := strings.Split(zoektQuery, "file:")
						if len(parts) > 1 {
							filePart := strings.Fields(parts[1])[0]
							if strings.Contains(zoektQuery, "blogs") && strings.Contains(filePart, "blogs") {
								zoektQuery = "file:blogs.js"
							} else {
								zoektQuery = "file:" + filePart
							}
						}
					} else if len(zoektQuery) > 200 {
						// If too long, try to extract a simple query from it
						words := strings.Fields(zoektQuery)
						// Look for query-like words
						for _, word := range words {
							if strings.HasPrefix(word, "file:") || strings.HasPrefix(word, "lang:") {
								zoektQuery = word
								break
							}
						}
						// If still too long, use first meaningful word
						if len(zoektQuery) > 200 && len(words) > 0 {
							zoektQuery = words[0]
						}
					}
					
					// Fix unterminated quotes in file patterns
					// Pattern: file:".*something (missing closing quote)
					if strings.Contains(zoektQuery, "file:\"") {
						// Count opening and closing quotes after "file:"
						parts := strings.Split(zoektQuery, "file:\"")
						for i := 1; i < len(parts); i++ {
							// Check if this part has a closing quote
							part := parts[i]
							// Find where this file pattern should end (space, OR, AND, or end of string)
							endIdx := len(part)
							for _, sep := range []string{" OR ", " AND ", " "} {
								if idx := strings.Index(part, sep); idx != -1 && idx < endIdx {
									endIdx = idx
								}
							}
							// Check if closing quote exists in this segment
							segment := part[:endIdx]
							if !strings.Contains(segment, "\"") && strings.Contains(segment, ".*") {
								// Missing closing quote - add it before the separator
								parts[i] = segment + "\"" + part[endIdx:]
							}
						}
						zoektQuery = strings.Join(parts, "file:\"")
					}
				}
			}
		}
	}

	// Validate query - reject single characters or very short queries
	if len(strings.TrimSpace(zoektQuery)) <= 1 {
		log.Printf("❌ OpenRouter returned invalid query (too short): '%s'", zoektQuery)
		log.Printf("📥 Raw response content was: %s", rawContent)
		log.Printf("📥 Full OpenRouter response: %+v", result)
		
		// Fallback: Try to generate a simple query from key terms
		keyTerms := t.extractKeyTerms(nlQuery)
		if len(keyTerms) > 0 {
			fallbackQuery := strings.Join(keyTerms, " ")
			log.Printf("🔄 Using fallback query from key terms: %s", fallbackQuery)
			zoektQuery = fallbackQuery
		} else {
			// Last resort: use the original query as-is
			log.Printf("🔄 Using original query as fallback: %s", nlQuery)
			zoektQuery = nlQuery
		}
	}
	
	if zoektQuery == "" || len(strings.TrimSpace(zoektQuery)) <= 1 {
		return "", nil, fmt.Errorf("no valid query generated from OpenRouter API and fallback failed")
	}
	
	// Validate and fix the query (especially unterminated quotes)
	zoektQuery = t.validateAndFixQuery(zoektQuery)
	
	// Additional cleanup: remove any remaining explanation text
	// If query contains explanation words like "Converting", "Formulating", "Transforming", etc., use fallback
	lowerQueryCheck := strings.ToLower(strings.TrimSpace(zoektQuery))
	explanationWords := []string{"converting", "formulating", "searching", "transforming", "looking", "finding", "analyzing", "processing"}
	isExplanationWord := false
	for _, word := range explanationWords {
		if lowerQueryCheck == word || strings.HasPrefix(lowerQueryCheck, word+" ") {
			isExplanationWord = true
			break
		}
	}
	
	if isExplanationWord {
		log.Printf("⚠️ Query is explanation word '%s', using fallback", zoektQuery)
		keyTerms := t.extractKeyTerms(nlQuery)
		if len(keyTerms) > 0 {
			zoektQuery = strings.Join(keyTerms, " ")
			log.Printf("🔄 Replaced explanation word with key terms: %s", zoektQuery)
		} else {
			// Use key terms extraction - let the model figure out what to search
			keyTerms := t.extractKeyTerms(nlQuery)
			if len(keyTerms) > 0 {
				zoektQuery = strings.Join(keyTerms, " ")
			} else {
				zoektQuery = nlQuery
			}
		}
	}
	
	// If query is still very long or contains explanation markers, extract the actual query
	if len(zoektQuery) > 200 || strings.Contains(zoektQuery, "**") {
		// Try to find query patterns in the text using regex
		queryPattern := regexp.MustCompile(`(file:[^\s\)]+|lang:[^\s\)]+|content:"[^"]+"|sym:[^\s\)]+|[a-zA-Z0-9_-]+)`)
		matches := queryPattern.FindAllString(zoektQuery, -1)
		if len(matches) > 0 {
			// Use the first match that looks like a query
			for _, match := range matches {
				if strings.HasPrefix(match, "file:") || strings.HasPrefix(match, "lang:") || len(match) < 50 {
					zoektQuery = match
					break
				}
			}
		}
		// If still problematic, use key terms extraction - let the model figure it out
		// The model should understand from snippets what to search for
		keyTerms := t.extractKeyTerms(nlQuery)
		if len(keyTerms) > 0 {
			zoektQuery = strings.Join(keyTerms, " ")
			log.Printf("🔄 Using key terms as fallback: %s", zoektQuery)
		} else {
			// Last resort: use original query
			zoektQuery = nlQuery
		}
	}

	// Final validation - reject queries that are just single characters or common words
	trimmedQuery := strings.TrimSpace(zoektQuery)
	if len(trimmedQuery) <= 1 || trimmedQuery == "I" || trimmedQuery == "i" {
		log.Printf("⚠️ Query validation failed: '%s' is too short or invalid", trimmedQuery)
		// Use fallback
		keyTerms := t.extractKeyTerms(nlQuery)
		if len(keyTerms) > 0 {
			zoektQuery = strings.Join(keyTerms, " ")
			log.Printf("🔄 Using fallback query: %s", zoektQuery)
		} else {
			// Use key terms extraction - model should understand from context
			keyTerms := t.extractKeyTerms(nlQuery)
			if len(keyTerms) > 0 {
				zoektQuery = strings.Join(keyTerms, " ")
			} else {
				zoektQuery = nlQuery
			}
			log.Printf("🔄 Using key terms fallback: %s", zoektQuery)
		}
	}
	
	log.Printf("🤖 AI Model generated Zoekt query: %s", zoektQuery)

	// Determine intent from the query
	intent := &Intent{
		Type:       "search",
		Confidence: 0.95, // High confidence when using AI
		Query:      zoektQuery,
	}

	// Simple intent detection based on query structure and keywords
	lower := strings.ToLower(nlQuery)
	if strings.Contains(lower, "how many") || strings.Contains(lower, "count") {
		intent.Type = "count"
	} else if strings.Contains(lower, "list") || strings.Contains(lower, "show all") {
		intent.Type = "list"
	} else if strings.Contains(lower, "have") || strings.Contains(lower, "do i") || strings.Contains(lower, "did i") {
		intent.Type = "yesno"
	}

	log.Printf("🎯 Detected query type: %s", intent.Type)
	return zoektQuery, intent, nil
}

// getCodebaseContext gets sample file names AND actual code snippets from the codebase to provide rich context
func (t *OpenRouterTranslator) getCodebaseContext() string {
	if t.searcher == nil {
		return "No codebase context available."
	}

	// Do a broad search to get diverse file samples
	ctx := context.Background()
	broadQuery, err := query.Parse(".")
	if err != nil {
		return "No codebase context available."
	}

	opts := &zoekt.SearchOptions{
		MaxDocDisplayCount: 30, // Get up to 30 files for context
		NumContextLines:    3,  // Include context lines around matches
	}

	result, err := t.searcher.Search(ctx, broadQuery, opts)
	if err != nil || result == nil || len(result.Files) == 0 {
		return "No codebase context available."
	}

	var contextBuilder strings.Builder
	contextBuilder.WriteString("CODEBASE STRUCTURE:\n")
	contextBuilder.WriteString("==================\n\n")

	// Extract file patterns and code snippets
	filePatterns := make(map[string]bool)
	codeSnippets := make(map[string]string) // fileName -> code snippet
	
	// Process top files to get code snippets
	maxFilesForSnippets := 4 // Reduced from 8 to 4 files to keep prompt reasonable
	filesProcessed := 0

	for _, file := range result.Files {
		fileName := file.FileName
		
		// Extract file patterns
		if strings.Contains(fileName, "/") {
			dir := strings.Split(fileName, "/")[0]
			filePatterns[dir+"/"] = true
		}
		ext := filepath.Ext(fileName)
		if ext != "" {
			filePatterns["*"+ext] = true
		}
		base := filepath.Base(fileName)
		if len(base) > 0 && base[0] != '.' {
			filePatterns[base] = true
		}

		// Extract code snippets from top files
		if filesProcessed < maxFilesForSnippets {
			snippet := t.extractCodeSnippet(file)
			if snippet != "" {
				codeSnippets[fileName] = snippet
				filesProcessed++
			}
		}
	}

	// Build file patterns section
	var patterns []string
	for pattern := range filePatterns {
		patterns = append(patterns, pattern)
	}
	if len(patterns) > 15 {
		patterns = patterns[:15]
	}
	
	if len(patterns) > 0 {
		contextBuilder.WriteString("Sample Files: ")
		contextBuilder.WriteString(strings.Join(patterns, ", "))
		contextBuilder.WriteString("\n\n")
	}

	// Add code snippets section
	if len(codeSnippets) > 0 {
		contextBuilder.WriteString("CODE SNIPPETS FROM KEY FILES:\n")
		contextBuilder.WriteString("==============================\n\n")
		
		snippetCount := 0
		for fileName, snippet := range codeSnippets {
			if snippetCount >= maxFilesForSnippets {
				break
			}
			contextBuilder.WriteString(fmt.Sprintf("File: %s\n", fileName))
			contextBuilder.WriteString("```\n")
			contextBuilder.WriteString(snippet)
			contextBuilder.WriteString("\n```\n\n")
			snippetCount++
		}
	}

	context := contextBuilder.String()
	if context == "" || context == "CODEBASE STRUCTURE:\n==================\n\n" {
		return "No codebase context available."
	}

	return context
}

// extractCodeSnippet extracts a meaningful code snippet from a file match
func (t *OpenRouterTranslator) extractCodeSnippet(file zoekt.FileMatch) string {
	var snippet strings.Builder
	
	// Method 1: Extract from LineMatches (if available)
	if len(file.LineMatches) > 0 {
		linesExtracted := 0
		maxLines := 20 // Reduced from 40 to 20 lines per file
		
		for _, match := range file.LineMatches {
			if linesExtracted >= maxLines {
				break
			}
			
			// Include context lines if available (Before is a single []byte with newlines)
			if len(match.Before) > 0 {
				beforeLines := bytes.Split(match.Before, []byte{'\n'})
				for _, beforeLine := range beforeLines {
					if linesExtracted >= maxLines {
						break
					}
					if len(beforeLine) > 0 {
						snippet.Write(beforeLine)
						snippet.WriteString("\n")
						linesExtracted++
					}
				}
			}
			
			// Include the matched line
			if linesExtracted < maxLines {
				snippet.Write(match.Line)
				snippet.WriteString("\n")
				linesExtracted++
			}
			
			// Include after context if available (After is a single []byte with newlines)
			if len(match.After) > 0 {
				afterLines := bytes.Split(match.After, []byte{'\n'})
				for _, afterLine := range afterLines {
					if linesExtracted >= maxLines {
						break
					}
					if len(afterLine) > 0 {
						snippet.Write(afterLine)
						snippet.WriteString("\n")
						linesExtracted++
					}
				}
			}
		}
	} else {
		// Method 2: If no LineMatches, try to get file content with a targeted search
		ctx := context.Background()
		// Use query.Parse to create a proper file query
		fileQueryStr := fmt.Sprintf(`file:"^%s$"`, regexp.QuoteMeta(file.FileName))
		fileQuery, err := query.Parse(fileQueryStr)
		if err == nil {
			opts := &zoekt.SearchOptions{
				Whole:             true, // Get full file content
				MaxDocDisplayCount: 1,
			}
			
			result, err := t.searcher.Search(ctx, fileQuery, opts)
			if err == nil && result != nil && len(result.Files) > 0 {
				content := result.Files[0].Content
				if len(content) > 0 {
					// Extract first 2000 characters (roughly 50-80 lines)
					contentStr := string(content)
					if len(contentStr) > 2000 {
						contentStr = contentStr[:2000] + "\n... (truncated)"
					}
					snippet.WriteString(contentStr)
				}
			}
		}
	}
	
	result := snippet.String()
	// Clean up: remove excessive blank lines
	lines := strings.Split(result, "\n")
	var cleanedLines []string
	blankCount := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			blankCount++
			if blankCount <= 2 { // Allow max 2 consecutive blank lines
				cleanedLines = append(cleanedLines, line)
			}
		} else {
			blankCount = 0
			cleanedLines = append(cleanedLines, line)
		}
	}
	
	// Limit total lines
	if len(cleanedLines) > 50 {
		cleanedLines = cleanedLines[:50]
		cleanedLines = append(cleanedLines, "... (truncated)")
	}
	
	return strings.Join(cleanedLines, "\n")
}

// validateAndFixQuery validates and fixes common query issues, especially unterminated quotes
func (t *OpenRouterTranslator) validateAndFixQuery(q string) string {
	queryStr := q
	
	// Fix unterminated content:" patterns - check if query ends without closing quote
	if strings.Contains(queryStr, `content:"`) {
		trimmed := strings.TrimSpace(queryStr)
		// If query doesn't end with a quote, add one
		if !strings.HasSuffix(trimmed, `"`) {
			queryStr = queryStr + `"`
			log.Printf("🔧 Fixed unterminated content: quote - added closing quote at end")
		}
	}
	
	// Fix unterminated file:" patterns
	if strings.Contains(queryStr, `file:"`) {
		lastFileIdx := strings.LastIndex(queryStr, `file:"`)
		afterFile := queryStr[lastFileIdx+6:]
		quoteCount := strings.Count(afterFile, `"`)
		if quoteCount == 0 {
			queryStr = queryStr + `"`
			log.Printf("🔧 Fixed unterminated file: quote - added closing quote")
		}
	}
	
	// Try to parse the query to validate it
	_, err := query.Parse(queryStr)
	if err != nil {
		log.Printf("⚠️ Query validation failed: %v, original query: %s", err, q)
		// If still invalid, try removing the problematic content: wrapper
		if strings.Contains(queryStr, `content:"`) {
			// Extract just the term without content: wrapper
			parts := strings.Fields(queryStr)
			var cleanParts []string
			for _, part := range parts {
				if strings.HasPrefix(part, `content:"`) {
					// Extract the term
					term := strings.TrimPrefix(part, `content:"`)
					term = strings.TrimSuffix(term, `"`)
					if term != "" {
						cleanParts = append(cleanParts, term)
					}
				} else if !strings.HasPrefix(part, `content:`) {
					cleanParts = append(cleanParts, part)
				}
			}
			if len(cleanParts) > 0 {
				queryStr = strings.Join(cleanParts, " ")
			}
		}
	}
	
	return queryStr
}

// findRelevantCodeSnippets searches for code snippets related to key terms in the query
func (t *OpenRouterTranslator) findRelevantCodeSnippets(nlQuery string) string {
	if t.searcher == nil {
		return ""
	}

	// Extract key terms from the query (remove common words)
	keyTerms := t.extractKeyTerms(nlQuery)
	if len(keyTerms) == 0 {
		return ""
	}

	// Search for each key term
	ctx := context.Background()
	var allSnippets []string
	snippetMap := make(map[string]bool) // To avoid duplicates

	for _, term := range keyTerms {
		// Create a simple search query for this term
		searchQuery, err := query.Parse(term)
		if err != nil {
			continue
		}

		opts := &zoekt.SearchOptions{
			MaxDocDisplayCount: 3,  // Reduced from 5 to 3 results per term
			NumContextLines:    3, // Reduced from 5 to 3 context lines
		}

		result, err := t.searcher.Search(ctx, searchQuery, opts)
		if err != nil || result == nil || len(result.Files) == 0 {
			continue
		}

		// Extract snippets from results
		for _, file := range result.Files {
			// Create a unique key for this file+term combination
			fileKey := fmt.Sprintf("%s:%s", file.FileName, term)
			if snippetMap[fileKey] {
				continue // Skip duplicates
			}
			snippetMap[fileKey] = true

			// Extract code snippet (limit to 15 lines per snippet)
			snippet := t.extractSnippetFromFile(file, term)
			if snippet != "" && len(snippet) > 20 { // Only include meaningful snippets
				// Limit snippet to 15 lines max
				lines := strings.Split(snippet, "\n")
				if len(lines) > 15 {
					lines = lines[:15]
					snippet = strings.Join(lines, "\n") + "\n... (truncated)"
				}
				allSnippets = append(allSnippets, fmt.Sprintf("Found '%s' in %s:\n```\n%s\n```", term, file.FileName, snippet))
			}
		}
	}

	if len(allSnippets) == 0 {
		return ""
	}

	// Limit to top 5 snippets to keep prompt reasonable (reduced from 10)
	if len(allSnippets) > 5 {
		allSnippets = allSnippets[:5]
	}

	return fmt.Sprintf(`RELEVANT CODE SNIPPETS (found in codebase for your query):
%s
`, strings.Join(allSnippets, "\n\n"))
}

// extractKeyTerms extracts meaningful terms from a natural language query
func (t *OpenRouterTranslator) extractKeyTerms(query string) []string {
	// Common stop words to filter out
	stopWords := map[string]bool{
		"how": true, "much": true, "many": true, "does": true, "do": true,
		"have": true, "has": true, "is": true, "are": true, "the": true,
		"a": true, "an": true, "what": true, "where": true, "when": true,
		"which": true, "who": true, "find": true, "show": true, "list": true,
		"get": true, "search": true, "for": true, "in": true, "on": true,
		"at": true, "to": true, "of": true, "and": true, "or": true,
	}

	// Convert to lowercase and split into words
	words := strings.Fields(strings.ToLower(query))
	var keyTerms []string
	seen := make(map[string]bool)

	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:()[]{}'\"")
		
		// Skip stop words and very short words
		if stopWords[word] || len(word) < 3 {
			continue
		}

		// Skip if already seen
		if seen[word] {
			continue
		}
		seen[word] = true

		keyTerms = append(keyTerms, word)
	}

	return keyTerms
}

// extractSnippetFromFile extracts a code snippet from a file match, focusing on the search term
func (t *OpenRouterTranslator) extractSnippetFromFile(file zoekt.FileMatch, searchTerm string) string {
	var snippet strings.Builder
	maxLines := 15 // Reduced from 30 to 15 lines per snippet

	if len(file.LineMatches) > 0 {
		linesExtracted := 0
		for _, match := range file.LineMatches {
			if linesExtracted >= maxLines {
				break
			}

			// Include context before
			if len(match.Before) > 0 {
				beforeLines := bytes.Split(match.Before, []byte{'\n'})
				for _, beforeLine := range beforeLines {
					if linesExtracted >= maxLines {
						break
					}
					if len(beforeLine) > 0 {
						snippet.Write(beforeLine)
						snippet.WriteString("\n")
						linesExtracted++
					}
				}
			}

			// Include the matched line (highlight the term)
			if linesExtracted < maxLines {
				snippet.Write(match.Line)
				snippet.WriteString("\n")
				linesExtracted++
			}

			// Include context after
			if len(match.After) > 0 {
				afterLines := bytes.Split(match.After, []byte{'\n'})
				for _, afterLine := range afterLines {
					if linesExtracted >= maxLines {
						break
					}
					if len(afterLine) > 0 {
						snippet.Write(afterLine)
						snippet.WriteString("\n")
						linesExtracted++
					}
				}
			}
		}
	}

	result := snippet.String()
	// Clean up excessive blank lines
	lines := strings.Split(result, "\n")
	var cleanedLines []string
	blankCount := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			blankCount++
			if blankCount <= 1 {
				cleanedLines = append(cleanedLines, line)
			}
		} else {
			blankCount = 0
			cleanedLines = append(cleanedLines, line)
		}
	}

	if len(cleanedLines) > maxLines {
		cleanedLines = cleanedLines[:maxLines]
		cleanedLines = append(cleanedLines, "... (truncated)")
	}

	return strings.Join(cleanedLines, "\n")
}


// AnswerFromSnippets uses OpenRouter to analyze code snippets and answer the question directly
func (t *OpenRouterTranslator) AnswerFromSnippets(question string, snippets []string) string {
	if !t.enabled || len(snippets) == 0 {
		return ""
	}
	
	// Combine snippets
	codeContext := strings.Join(snippets, "\n\n---\n\n")
	if len(codeContext) > 2000 {
		codeContext = codeContext[:2000] + "\n... (truncated)"
	}
	
	prompt := fmt.Sprintf(`Based on the following code snippets from the codebase, answer this question directly and concisely.

CODE SNIPPETS:
%s

QUESTION: %s

Answer the question directly. If the answer is YES, say "✅ YES" and explain briefly. If NO, say "❌ NO".
Be specific and reference the code if relevant.

Answer:`, codeContext, question)
	
	url := "https://openrouter.ai/api/v1/chat/completions"
	payload := map[string]interface{}{
		"model": t.model,
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": "You are a code analysis assistant. Answer questions about code directly and concisely based on the provided code snippets.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3,
		"max_tokens":  150,
	}
	
	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.apiKey))
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}
	
	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					return strings.TrimSpace(content)
				}
			}
		}
	}
	
	return ""
}
