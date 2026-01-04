package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// SemanticSearchService provides semantic search using ChromaDB HTTP API
type SemanticSearchService struct {
	embeddingService *EmbeddingService
	chromaURL        string
	collectionName   string
}

// NewSemanticSearchService creates a new semantic search service
func NewSemanticSearchService() (*SemanticSearchService, error) {
	embeddingService := NewEmbeddingService()
	if !embeddingService.enabled {
		return nil, fmt.Errorf("embedding service not enabled (no API key)")
	}

	// ChromaDB HTTP API URL (default: localhost:8000)
	chromaURL := os.Getenv("CHROMA_URL")
	if chromaURL == "" {
		chromaURL = "http://localhost:8000"
	}

	return &SemanticSearchService{
		embeddingService: embeddingService,
		chromaURL:        chromaURL,
		collectionName:   "codebase_chunks",
	}, nil
}

// SemanticSearchResultV2 represents a semantically similar code chunk (for search results)
// Note: This is different from embedding_service.go's SemanticSearchResult which uses CodeChunk
type SemanticSearchResultV2 struct {
	Chunk      SemanticCodeChunk
	Similarity float32
	Score      float32 // Normalized score 0-1
}

// Search performs semantic search and returns top K results
func (s *SemanticSearchService) Search(query string, topK int) ([]SemanticSearchResultV2, error) {
	if topK <= 0 {
		topK = 10
	}

	// Generate embedding for query
	queryEmbedding, err := s.embeddingService.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %v", err)
	}

	// Convert float32 to float64 for JSON
	queryEmbedding64 := make([]float64, len(queryEmbedding))
	for i, v := range queryEmbedding {
		queryEmbedding64[i] = float64(v)
	}

	// Query ChromaDB via HTTP API
	queryURL := fmt.Sprintf("%s/api/v1/collections/%s/query", s.chromaURL, s.collectionName)
	
	payload := map[string]interface{}{
		"query_embeddings": [][]float64{queryEmbedding64},
		"n_results":        topK,
		"include":          []string{"documents", "metadatas", "distances"},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %v", err)
	}

	req, err := http.NewRequest("POST", queryURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query ChromaDB: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ChromaDB query failed with status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Parse results
	var searchResults []SemanticSearchResultV2

	ids, _ := result["ids"].([]interface{})
	documents, _ := result["documents"].([]interface{})
	metadatas, _ := result["metadatas"].([]interface{})
	distances, _ := result["distances"].([]interface{})

	if len(ids) > 0 && len(documents) > 0 {
		idsList := ids[0].([]interface{})
		documentsList := documents[0].([]interface{})
		metadatasList := metadatas[0].([]interface{})
		distancesList := distances[0].([]interface{})

		for i := 0; i < len(idsList) && i < topK; i++ {
			// Convert distance to similarity
			distance := 0.0
			if distVal, ok := distancesList[i].(float64); ok {
				distance = distVal
			}
			// For cosine similarity: similarity = 1 - distance (if normalized)
			// For L2 distance, we approximate: similarity = 1 / (1 + distance)
			similarity := float32(1.0 / (1.0 + distance))
			if similarity > 1.0 {
				similarity = 1.0
			}

			// Extract data
			chunkID := ""
			if idVal, ok := idsList[i].(string); ok {
				chunkID = idVal
			}

			docContent := ""
			if docVal, ok := documentsList[i].(string); ok {
				docContent = docVal
			}

			metadata := make(map[string]interface{})
			if metaVal, ok := metadatasList[i].(map[string]interface{}); ok {
				metadata = metaVal
			}

			fileName := ""
			startLine := 0
			endLine := 0
			chunkType := ""
			language := ""

			if fileNameVal, ok := metadata["file_name"].(string); ok {
				fileName = fileNameVal
			}
			if startLineVal, ok := metadata["start_line"].(float64); ok {
				startLine = int(startLineVal)
			}
			if endLineVal, ok := metadata["end_line"].(float64); ok {
				endLine = int(endLineVal)
			}
			if typeVal, ok := metadata["type"].(string); ok {
				chunkType = typeVal
			}
			if langVal, ok := metadata["language"].(string); ok {
				language = langVal
			}

			searchResults = append(searchResults, SemanticSearchResultV2{
				Chunk: SemanticCodeChunk{
					ID:        chunkID,
					Content:   docContent,
					FileName:  fileName,
					StartLine: startLine,
					EndLine:   endLine,
					Type:      chunkType,
					Language:  language,
				},
				Similarity: similarity,
				Score:      similarity,
			})
		}
	}

	log.Printf("🔍 Semantic search found %d results for query: %s", len(searchResults), query)
	return searchResults, nil
}

// IsIndexed checks if the codebase has been indexed
func (s *SemanticSearchService) IsIndexed() (bool, error) {
	// Check collection count via HTTP API
	countURL := fmt.Sprintf("%s/api/v1/collections/%s/count", s.chromaURL, s.collectionName)
	
	req, err := http.NewRequest("GET", countURL, nil)
	if err != nil {
		return false, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	count := 0.0
	if countVal, ok := result["count"].(float64); ok {
		count = countVal
	}

	return count > 0, nil
}

// GetIndexStats returns statistics about the index
func (s *SemanticSearchService) GetIndexStats() (map[string]interface{}, error) {
	isIndexed, err := s.IsIndexed()
	if err != nil {
		return nil, err
	}

	countURL := fmt.Sprintf("%s/api/v1/collections/%s/count", s.chromaURL, s.collectionName)
	req, _ := http.NewRequest("GET", countURL, nil)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	
	count := 0
	if err == nil && resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		if json.NewDecoder(resp.Body).Decode(&result) == nil {
			if countVal, ok := result["count"].(float64); ok {
				count = int(countVal)
			}
		}
		resp.Body.Close()
	}

	return map[string]interface{}{
		"total_chunks": count,
		"indexed":      isIndexed,
	}, nil
}

