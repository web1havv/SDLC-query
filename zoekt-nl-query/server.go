package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegraph/zoekt"
	"github.com/sourcegraph/zoekt/query"
)

type NLQueryServer struct {
	searcher zoekt.Searcher
	openRouterTranslator *OpenRouterTranslator
	basicTranslator *BasicTranslator
}

func NewNLQueryServer(searcher zoekt.Searcher) *NLQueryServer {
	return &NLQueryServer{
		searcher: searcher,
		openRouterTranslator: NewOpenRouterTranslator(searcher),
		basicTranslator: NewBasicTranslator(),
	}
}

func (s *NLQueryServer) HandleSearch(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	
	// Get query from request
	nlQuery := r.URL.Query().Get("q")
	if nlQuery == "" {
		// Return JSON error instead of plain text
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Missing query parameter 'q'",
			"success": false,
		})
		return
	}

	// Check if direct query mode (skip translation)
	directQuery := r.URL.Query().Get("direct") == "true"
	
	var intent *Intent
	var err error
	zoektQuery := ""
	queryType := "search"
	translatorUsed := "direct"
	isNL := false
	
	if directQuery {
		// Use query as-is, no translation
		zoektQuery = nlQuery
		intent = &Intent{
			Type:       "search",
			Query:      zoektQuery,
			Confidence: 1.0,
		}
		log.Printf("🔵 Using direct query mode (no translation): %s", zoektQuery)
	} else {
		// Try OpenRouter AI model first, fallback to basic translator if it fails
		if !s.openRouterTranslator.enabled {
			log.Printf("⚠️ OpenRouter not enabled, using basic translator")
			// Fallback to basic translator
			basicQuery, basicQueryType := s.basicTranslator.Translate(nlQuery)
			zoektQuery = basicQuery
			intent = &Intent{
				Type:       basicQueryType,
				Query:      basicQuery,
				Confidence: 0.7,
			}
			queryType = basicQueryType
			translatorUsed = "basic"
			isNL = true
			log.Printf("📝 Using basic translator: %s", zoektQuery)
		} else {
			// Use OpenRouter AI model
			zoektQuery, intent, err = s.openRouterTranslator.TranslateWithOpenRouter(nlQuery)
			if err != nil {
				log.Printf("❌ OpenRouter translation failed: %v", err)
				log.Printf("🔄 Falling back to basic translator...")
				// Fallback to basic translator
				basicQuery, basicQueryType := s.basicTranslator.Translate(nlQuery)
				zoektQuery = basicQuery
				intent = &Intent{
					Type:       basicQueryType,
					Query:      basicQuery,
					Confidence: 0.7,
				}
				queryType = basicQueryType
				translatorUsed = "basic"
				isNL = true
				log.Printf("📝 Using basic translator fallback: %s", zoektQuery)
			} else if zoektQuery == "" {
				log.Printf("❌ OpenRouter returned empty query, falling back to basic translator")
				// Fallback to basic translator
				basicQuery, basicQueryType := s.basicTranslator.Translate(nlQuery)
				zoektQuery = basicQuery
				intent = &Intent{
					Type:       basicQueryType,
					Query:      basicQuery,
					Confidence: 0.7,
				}
				queryType = basicQueryType
				translatorUsed = "basic"
				isNL = true
				log.Printf("📝 Using basic translator fallback: %s", zoektQuery)
			} else {
				// Check if we got a direct answer instead of a query
				if strings.HasPrefix(zoektQuery, "DIRECT_ANSWER:") {
					directAnswer := strings.TrimPrefix(zoektQuery, "DIRECT_ANSWER:")
					log.Printf("✅ Got direct answer from OpenRouter: %s", directAnswer)
					
					// Return direct answer without executing Zoekt query
					w.Header().Set("Content-Type", "application/json")
					response := map[string]interface{}{
						"originalQuery": nlQuery,
						"translatedQuery": "",
						"generatedQuery": "",
						"queryType": "answer",
						"isNL": true,
						"success": true,
						"resultCount": 0,
						"responseTime": time.Since(startTime).Milliseconds(),
						"results": nil,
						"error": "",
						"yesNoAnswer": directAnswer,
						"countAnswer": "",
						"directAnswer": directAnswer,
						"translatorUsed": "openrouter",
						"intent": map[string]interface{}{
							"type":       intent.Type,
							"entity":     intent.Entity,
							"topic":      intent.Topic,
							"confidence": intent.Confidence,
							"variations": intent.Variations,
						},
					}
					json.NewEncoder(w).Encode(response)
					return
				}
				
				queryType = intent.Type
				translatorUsed = "openrouter"
				isNL = true
				log.Printf("✅ Using OpenRouter AI model")
				log.Printf("📝 Original query: %s", nlQuery)
				log.Printf("📝 Generated Zoekt query: %s", zoektQuery)
				log.Printf("📝 Query type: %s", queryType)
			}
		}
	}

	// Parse Zoekt query
	parsedQuery, err := query.Parse(zoektQuery)
	if err != nil {
		// Return JSON error instead of plain text
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("Query parse error: %v", err),
			"originalQuery": nlQuery,
			"translatedQuery": zoektQuery,
			"success": false,
		})
		return
	}

	// Execute search
	opts := &zoekt.SearchOptions{
		MaxDocDisplayCount: 100,
	}
	
	result, err := s.searcher.Search(context.Background(), parsedQuery, opts)
	responseTime := time.Since(startTime)

	success := err == nil
	resultCount := 0
	lineMatchCount := 0 // Count total line matches for "count" queries
	var directAnswer string // For direct answers from code snippets
	
	if result != nil {
		resultCount = len(result.Files)
		// For count queries, count the total number of line matches across all files
		if queryType == "count" {
			for _, file := range result.Files {
				if file.LineMatches != nil {
					lineMatchCount += len(file.LineMatches)
				}
			}
			// Use line match count as the count result
			if lineMatchCount > 0 {
				resultCount = lineMatchCount
			}
		}
		
		// For yes/no questions, try to generate a direct answer from the results
		if queryType == "yesno" && resultCount > 0 {
			// Extract relevant code snippets from results
			snippets := s.extractAnswerSnippets(result, 3) // Top 3 files
			if len(snippets) > 0 {
				// Use OpenRouter to analyze snippets and answer the question
				answer := s.openRouterTranslator.AnswerFromSnippets(nlQuery, snippets)
				if answer != "" {
					directAnswer = answer
				}
			}
		}
	}

	// Handle yes/no questions - provide clear answer
	yesNoAnswer := ""
	if queryType == "yesno" {
		if directAnswer != "" {
			yesNoAnswer = directAnswer
		} else if resultCount > 0 {
			yesNoAnswer = fmt.Sprintf("✅ YES, found %d result(s).", resultCount)
		} else {
			yesNoAnswer = "❌ NO, no results found."
		}
	}
	
	// Handle count queries - provide clear count answer
	countAnswer := ""
	if queryType == "count" {
		if lineMatchCount > 0 {
			countAnswer = fmt.Sprintf("Found %d item(s).", lineMatchCount)
		} else if resultCount > 0 {
			countAnswer = fmt.Sprintf("Found %d file(s).", resultCount)
		} else {
			countAnswer = "Found 0 items."
		}
	}

	// Return results
	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]interface{}{
		"originalQuery": nlQuery,
		"translatedQuery": zoektQuery,
		"generatedQuery": zoektQuery, // Always show the generated Zoekt query (primary field)
		"queryType": queryType,
		"isNL": isNL,
		"success": success,
		"resultCount": resultCount,
		"responseTime": responseTime.Milliseconds(),
		"results": result,
		"error": "",
		"yesNoAnswer": yesNoAnswer,
		"countAnswer": countAnswer, // Answer for count queries
		"directAnswer": directAnswer, // Direct answer from code analysis
		"translatorUsed": translatorUsed, // Show which translator was used
		"intent": map[string]interface{}{
			"type":       intent.Type,
			"entity":     intent.Entity,
			"topic":      intent.Topic,
			"confidence": intent.Confidence,
			"variations": intent.Variations,
		},
	}
	
	if err != nil {
		response["error"] = err.Error()
	}

	json.NewEncoder(w).Encode(response)
}

// extractAnswerSnippets extracts code snippets from search results for answering questions
// Extracts at least 100 lines total across multiple files
func (s *NLQueryServer) extractAnswerSnippets(result *zoekt.SearchResult, maxFiles int) []string {
	var snippets []string
	count := 0
	targetLinesTotal := 100 // Target at least 100 lines total
	linesPerFile := targetLinesTotal / maxFiles
	if linesPerFile < 15 {
		linesPerFile = 15 // Minimum 15 lines per file
	}
	
	totalLines := 0
	for _, file := range result.Files {
		if count >= maxFiles || totalLines >= targetLinesTotal {
			break
		}
		var snippet strings.Builder
		snippet.WriteString(fmt.Sprintf("File: %s\n", file.FileName))
		snippet.WriteString(fmt.Sprintf("Repository: %s\n\n", file.Repository))
		
		fileLines := 0
		
		// Method 1: If we have full file content, extract from it
		if len(file.Content) > 0 {
			contentStr := string(file.Content)
			allLines := strings.Split(contentStr, "\n")
			
			// If we have line matches, extract around those lines
			if len(file.LineMatches) > 0 {
				for _, match := range file.LineMatches {
					if fileLines >= linesPerFile || totalLines >= targetLinesTotal {
						break
					}
					
					lineNum := match.LineNumber
					startLine := int(lineNum) - 10 // 10 lines before
					if startLine < 0 {
						startLine = 0
					}
					endLine := int(lineNum) + 10 // 10 lines after
					if endLine > len(allLines) {
						endLine = len(allLines)
					}
					
					// Extract lines around the match
					for i := startLine; i < endLine && fileLines < linesPerFile && totalLines < targetLinesTotal; i++ {
						line := strings.TrimRight(allLines[i], "\r\n")
						if line != "" || i == int(lineNum)-1 { // Include empty lines if it's the match line
							snippet.WriteString(line)
							snippet.WriteString("\n")
							fileLines++
							totalLines++
						}
					}
				}
			} else {
				// No line matches, extract first N lines of file
				extractCount := linesPerFile
				if extractCount > len(allLines) {
					extractCount = len(allLines)
				}
				for i := 0; i < extractCount && totalLines < targetLinesTotal; i++ {
					line := strings.TrimRight(allLines[i], "\r\n")
					snippet.WriteString(line)
					snippet.WriteString("\n")
					fileLines++
					totalLines++
				}
			}
		} else if len(file.LineMatches) > 0 {
			// Method 2: Extract from line matches with context (when content not available)
			for _, match := range file.LineMatches {
				if fileLines >= linesPerFile || totalLines >= targetLinesTotal {
					break
				}
				
				// Include before context (all available)
				if len(match.Before) > 0 {
					beforeLines := strings.Split(string(match.Before), "\n")
					for j := 0; j < len(beforeLines) && fileLines < linesPerFile && totalLines < targetLinesTotal; j++ {
						line := strings.TrimSpace(beforeLines[j])
						if line != "" {
							snippet.WriteString(line)
							snippet.WriteString("\n")
							fileLines++
							totalLines++
						}
					}
				}
				
				// Include the matched line
				if fileLines < linesPerFile && totalLines < targetLinesTotal {
					lineStr := strings.TrimSpace(string(match.Line))
					if lineStr != "" {
						snippet.WriteString(lineStr)
						snippet.WriteString("\n")
						fileLines++
						totalLines++
					}
				}
				
				// Include after context (all available)
				if len(match.After) > 0 && fileLines < linesPerFile && totalLines < targetLinesTotal {
					afterLines := strings.Split(string(match.After), "\n")
					for j := 0; j < len(afterLines) && fileLines < linesPerFile && totalLines < targetLinesTotal; j++ {
						line := strings.TrimSpace(afterLines[j])
						if line != "" {
							snippet.WriteString(line)
							snippet.WriteString("\n")
							fileLines++
							totalLines++
						}
					}
				}
			}
		}
		
		if snippet.Len() > 0 {
			snippets = append(snippets, snippet.String())
			count++
		}
	}
	
	log.Printf("📝 Extracted %d total lines of code across %d files", totalLines, len(snippets))
	return snippets
}

func (s *NLQueryServer) HandleClearCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Method not allowed. Use POST.",
			"success": false,
		})
		return
	}
	
	// Cache functionality was removed, so this is a no-op
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Cache clear endpoint (cache functionality removed)",
	})
	log.Printf("🗑️  Cache clear endpoint called (no-op)")
}

func (s *NLQueryServer) HandleAsk(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	
	// Get question from request
	question := r.URL.Query().Get("q")
	if question == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Missing question parameter 'q'",
			"success": false,
		})
		return
	}

	log.Printf("💬 Question received: %s", question)

	// Step 1: Generate a Zoekt query to find relevant code
	var zoektQuery string
	var searchResult *zoekt.SearchResult
	var searchErr error
	
	if s.openRouterTranslator != nil && s.openRouterTranslator.enabled {
		// Generate a query to find relevant code
		zoektQuery, _, searchErr = s.openRouterTranslator.TranslateWithOpenRouter(question)
		if searchErr == nil && zoektQuery != "" {
			// Skip if we got a DIRECT_ANSWER (not a real query)
			if !strings.HasPrefix(zoektQuery, "DIRECT_ANSWER:") {
				// Parse and execute the Zoekt query
				parsedQuery, parseErr := query.Parse(zoektQuery)
				if parseErr == nil {
				opts := &zoekt.SearchOptions{
					MaxDocDisplayCount: 20, // Get top 20 files for more context
					NumContextLines:    30, // Get 30 lines before/after each match
					Whole:              true, // Request full file content for better extraction
				}
					ctx := context.Background()
					searchResult, searchErr = s.searcher.Search(ctx, parsedQuery, opts)
					if searchErr == nil {
						log.Printf("📚 Found %d files for context", len(searchResult.Files))
					}
				}
			} else {
				// If we got DIRECT_ANSWER, clear the query
				zoektQuery = ""
			}
		}
	}

	// Step 2: Extract code snippets from search results
	var codeSnippets []string
	if searchResult != nil && len(searchResult.Files) > 0 {
		codeSnippets = s.extractAnswerSnippets(searchResult, 10) // Top 10 files for more context
		log.Printf("📝 Extracted %d code snippets", len(codeSnippets))
	}

	// Step 3: Use OpenRouter to generate text answer based on code snippets
	var answer string
	var err error
	
	if s.openRouterTranslator != nil && s.openRouterTranslator.enabled {
		if len(codeSnippets) > 0 {
			// Answer based on actual code snippets from Zoekt search
			answer = s.openRouterTranslator.AnswerFromSnippets(question, codeSnippets)
			if answer == "" {
				err = fmt.Errorf("failed to generate answer from snippets")
			}
		} else {
			// Fallback: use basic answer
			err = fmt.Errorf("no code snippets available for answering")
		}
		
		if err != nil {
			log.Printf("❌ Failed to generate answer: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": fmt.Sprintf("Failed to generate answer: %v", err),
				"success": false,
				"question": question,
			})
			return
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "OpenRouter translator not available",
			"success": false,
			"question": question,
		})
		return
	}

	responseTime := time.Since(startTime)

	// Return answer
	filesFound := 0
	if searchResult != nil {
		filesFound = len(searchResult.Files)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"question": question,
		"answer": answer,
		"success": true,
		"responseTime": responseTime.Milliseconds(),
		"type": "text_answer",
		"queryUsed": zoektQuery,
		"filesFound": filesFound,
	})
	
	log.Printf("✅ Answer generated in %dms", responseTime.Milliseconds())
}

func (s *NLQueryServer) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/nl-search", s.HandleSearch)
	mux.HandleFunc("/api/ask", s.HandleAsk)
	mux.HandleFunc("/api/clear-cache", s.HandleClearCache)
	
	// Serve our unified dashboard at /dashboard (main route)
	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/Users/web1havv/SDLC_AI/zoekt-nl-query/unified-dashboard.html")
	})
	
	// Redirect root to dashboard for convenience
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			// Redirect to dashboard
			http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
		} else {
			http.NotFound(w, r)
		}
	})
	
	log.Println("NL Query Server routes registered:")
	log.Println("  GET /dashboard - Natural Language Search Dashboard (main)")
	log.Println("  GET /api/nl-search?q=<query> - Natural language search API")
	log.Println("  GET /api/ask?q=<question> - Ask questions, get text answers")
	log.Println("  POST /api/clear-cache - Clear query translation cache")
	log.Println("  GET / - Redirects to /dashboard")
}

