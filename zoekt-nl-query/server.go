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
	searcher            zoekt.Searcher
	openRouterTranslator *OpenRouterTranslator
	basicTranslator     *BasicTranslator
	semanticSearch      *SemanticSearchService
	semanticIndexer     *SemanticIndexer
}

func NewNLQueryServer(searcher zoekt.Searcher) *NLQueryServer {
	// Try to initialize semantic search (may fail if ChromaDB not available)
	semanticSearch, err1 := NewSemanticSearchService()
	if err1 != nil {
		log.Printf("⚠️ Semantic search service not available: %v", err1)
	}
	
	semanticIndexer, err2 := NewSemanticIndexer("")
	if err2 != nil {
		log.Printf("⚠️ Semantic indexer not available: %v", err2)
	} else {
		log.Printf("✅ Semantic indexer initialized successfully")
	}
	
	return &NLQueryServer{
		searcher:             searcher,
		openRouterTranslator: NewOpenRouterTranslator(searcher),
		basicTranslator:     NewBasicTranslator(),
		semanticSearch:       semanticSearch,
		semanticIndexer:      semanticIndexer,
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

	// Check search mode: "semantic", "hybrid", or default (keyword)
	searchMode := r.URL.Query().Get("mode")
	if searchMode == "" {
		searchMode = "hybrid" // Default to hybrid
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

	// Try semantic search if enabled and mode is semantic/hybrid
	var semanticResults []SemanticSearchResultV2
	if s.semanticSearch != nil && (searchMode == "semantic" || searchMode == "hybrid") {
		isIndexed, _ := s.semanticSearch.IsIndexed()
		if isIndexed {
			semResults, err := s.semanticSearch.Search(nlQuery, 20)
			if err == nil {
				semanticResults = semResults
				log.Printf("🔍 Semantic search found %d results", len(semanticResults))
			} else {
				log.Printf("⚠️ Semantic search failed: %v", err)
			}
		} else {
			log.Printf("⚠️ Semantic index not available, using keyword search only")
		}
	}

	// If semantic-only mode and we have results, return them
	if searchMode == "semantic" && len(semanticResults) > 0 {
		responseTime := time.Since(startTime)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"originalQuery": nlQuery,
			"queryType":     "semantic",
			"isNL":          true,
			"success":       true,
			"resultCount":   len(semanticResults),
			"responseTime":  responseTime.Milliseconds(),
			"semanticResults": semanticResults,
			"mode":          "semantic",
		})
		return
	}

	// Parse Zoekt query for keyword search
	parsedQuery, err := query.Parse(zoektQuery)
	if err != nil {
		// If semantic search worked, use that instead
		if len(semanticResults) > 0 {
			responseTime := time.Since(startTime)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"originalQuery": nlQuery,
				"queryType":     "semantic",
				"isNL":          true,
				"success":       true,
				"resultCount":   len(semanticResults),
				"responseTime":  responseTime.Milliseconds(),
				"semanticResults": semanticResults,
				"mode":          "hybrid",
				"note":          "Keyword query failed, using semantic results",
			})
			return
		}
		
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

	// Execute keyword search
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
		"mode": searchMode,
		"intent": map[string]interface{}{
			"type":       intent.Type,
			"entity":     intent.Entity,
			"topic":      intent.Topic,
			"confidence": intent.Confidence,
			"variations": intent.Variations,
		},
	}

	// Enhance semantic results with Zoekt in hybrid mode
	if searchMode == "hybrid" && len(semanticResults) > 0 && result != nil {
		enhancedResults := s.enhanceSemanticWithZoekt(semanticResults, result, nlQuery)
		response["semanticResults"] = enhancedResults
		response["semanticResultCount"] = len(enhancedResults)
		response["enhanced"] = true // Flag to show results were enhanced
	} else if len(semanticResults) > 0 {
		// Just return semantic results as-is (semantic-only mode)
		response["semanticResults"] = semanticResults
		response["semanticResultCount"] = len(semanticResults)
	}
	
	if err != nil {
		response["error"] = err.Error()
	}

	json.NewEncoder(w).Encode(response)
}

// extractAnswerSnippets extracts code snippets from search results for answering questions
func (s *NLQueryServer) extractAnswerSnippets(result *zoekt.SearchResult, maxFiles int) []string {
	var snippets []string
	count := 0
	for _, file := range result.Files {
		if count >= maxFiles {
			break
		}
		var snippet strings.Builder
		snippet.WriteString(fmt.Sprintf("File: %s\n", file.FileName))
		
		// Extract from line matches
		if len(file.LineMatches) > 0 {
			for i, match := range file.LineMatches {
				if i >= 5 { // Limit to 5 lines per file
					break
				}
				if len(match.Before) > 0 {
					snippet.Write(match.Before)
					snippet.WriteString("\n")
				}
				snippet.Write(match.Line)
				snippet.WriteString("\n")
				if len(match.After) > 0 {
					snippet.Write(match.After)
					snippet.WriteString("\n")
				}
			}
		}
		snippets = append(snippets, snippet.String())
		count++
	}
	return snippets
}

// enhanceSemanticWithZoekt uses Zoekt to enhance semantic search results
// This makes Zoekt a "weapon" - a layer that validates, boosts, and enriches semantic results
func (s *NLQueryServer) enhanceSemanticWithZoekt(semanticResults []SemanticSearchResultV2, zoektResult *zoekt.SearchResult, originalQuery string) []SemanticSearchResultV2 {
	if s.searcher == nil || zoektResult == nil {
		return semanticResults
	}

	// Create a map of Zoekt-matched files for fast lookup
	zoektFileMap := make(map[string]*zoekt.FileMatch)
	for i := range zoektResult.Files {
		file := &zoektResult.Files[i]
		zoektFileMap[file.FileName] = file
	}

	// Enhance each semantic result
	enhanced := make([]SemanticSearchResultV2, 0, len(semanticResults))
	
	for _, semResult := range semanticResults {
		enhancedResult := semResult
		fileName := semResult.Chunk.FileName
		
		// Check if this file also matches in Zoekt (keyword search)
		if zoektFile, found := zoektFileMap[fileName]; found {
			// Boost: This semantic match also has keyword matches - highly relevant!
			enhancedResult.Score = enhancedResult.Score * 1.3 // Boost by 30%
			if enhancedResult.Score > 1.0 {
				enhancedResult.Score = 1.0
			}
			enhancedResult.Similarity = enhancedResult.Similarity * 1.2 // Boost similarity too
			if enhancedResult.Similarity > 1.0 {
				enhancedResult.Similarity = 1.0
			}
			
			// Add Zoekt metadata: keyword match count
			keywordMatchCount := len(zoektFile.LineMatches)
			
			log.Printf("🎯 Enhanced semantic result: %s (similarity: %.2f → %.2f, keyword matches: %d)", 
				fileName, semResult.Similarity, enhancedResult.Similarity, keywordMatchCount)
		} else {
			// Semantic match but no keyword match - still valid but lower confidence
			log.Printf("💡 Semantic-only match: %s (no keyword matches)", fileName)
		}
		
		enhanced = append(enhanced, enhancedResult)
	}
	
	// Re-sort by enhanced score (highest first)
	for i := 0; i < len(enhanced); i++ {
		for j := i + 1; j < len(enhanced); j++ {
			if enhanced[j].Score > enhanced[i].Score {
				enhanced[i], enhanced[j] = enhanced[j], enhanced[i]
			}
		}
	}
	
	log.Printf("⚡ Enhanced %d semantic results with Zoekt validation", len(enhanced))
	return enhanced
}

func (s *NLQueryServer) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/nl-search", s.HandleSearch)
	mux.HandleFunc("/api/index-codebase", s.HandleIndexCodebase)
	mux.HandleFunc("/api/semantic-stats", s.HandleSemanticStats)
	
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
	log.Println("  GET /api/nl-search?q=<query>&mode=semantic|hybrid|keyword - Natural language search API")
	log.Println("  POST /api/index-codebase?path=<codebase_path> - Index codebase for semantic search")
	log.Println("  GET /api/semantic-stats - Get semantic index statistics")
	log.Println("  GET / - Redirects to /dashboard")
}

// HandleIndexCodebase indexes a codebase for semantic search
func (s *NLQueryServer) HandleIndexCodebase(w http.ResponseWriter, r *http.Request) {
	if s.semanticIndexer == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Semantic indexing not available (embedding service not enabled)",
			"success": false,
		})
		return
	}

	codebasePath := r.URL.Query().Get("path")
	if codebasePath == "" {
		codebasePath = "/Users/web1havv/SDLC_AI" // Default to workspace root
	}

	log.Printf("🚀 Starting codebase indexing: %s", codebasePath)

	// Run indexing in goroutine to avoid blocking
	go func() {
		if err := s.semanticIndexer.IndexCodebase(codebasePath); err != nil {
			log.Printf("❌ Indexing failed: %v", err)
		} else {
			log.Printf("✅ Indexing completed successfully")
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Indexing started in background",
		"path":    codebasePath,
	})
}

// HandleSemanticStats returns statistics about the semantic index
func (s *NLQueryServer) HandleSemanticStats(w http.ResponseWriter, r *http.Request) {
	if s.semanticSearch == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"indexed": false,
			"error":   "Semantic search not available",
		})
		return
	}

	stats, err := s.semanticSearch.GetIndexStats()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"indexed": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

