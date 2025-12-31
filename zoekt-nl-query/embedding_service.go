package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"
)

// EmbeddingService handles embedding generation and semantic search
type EmbeddingService struct {
	apiKey string
	model  string // e.g., "text-embedding-ada-002" or "text-embedding-3-small"
	enabled bool
}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService() *EmbeddingService {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// Try OpenRouter for embeddings
		apiKey = os.Getenv("OPENROUTER_API_KEY")
	}
	
	model := os.Getenv("EMBEDDING_MODEL")
	if model == "" {
		// Use OpenAI's embedding model via OpenRouter
		model = "openai/text-embedding-ada-002"
	}
	
	return &EmbeddingService{
		apiKey:  apiKey,
		model:   model,
		enabled: apiKey != "",
	}
}

// Embed generates an embedding vector for text
func (e *EmbeddingService) Embed(text string) ([]float32, error) {
	if !e.enabled {
		return nil, fmt.Errorf("embedding service not enabled (no API key)")
	}
	
	// Use OpenRouter for embeddings (supports multiple providers)
	url := "https://openrouter.ai/api/v1/embeddings"
	
	payload := map[string]interface{}{
		"model": e.model,
		"input": text,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", e.apiKey))
	req.Header.Set("HTTP-Referer", "https://github.com/sourcegraph/zoekt")
	req.Header.Set("X-Title", "Zoekt Embeddings")
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API error: status %d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	// Extract embedding vector
	if data, ok := result["data"].([]interface{}); ok && len(data) > 0 {
		if item, ok := data[0].(map[string]interface{}); ok {
			if embedding, ok := item["embedding"].([]interface{}); ok {
				vector := make([]float32, len(embedding))
				for i, v := range embedding {
					if f, ok := v.(float64); ok {
						vector[i] = float32(f)
					}
				}
				return vector, nil
			}
		}
	}
	
	return nil, fmt.Errorf("failed to extract embedding from response")
}

// CosineSimilarity calculates cosine similarity between two vectors
func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}
	
	var dotProduct float32
	var normA, normB float32
	
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	
	if normA == 0 || normB == 0 {
		return 0
	}
	
	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// CodeChunk represents a chunk of code with its embedding
type CodeChunk struct {
	FileName string
	Content  string
	Embedding []float32
	StartLine int
	EndLine   int
}

// SemanticSearchResult represents a semantically similar code chunk
type SemanticSearchResult struct {
	Chunk     CodeChunk
	Similarity float32
}

// FindSemanticallySimilar finds code chunks semantically similar to the query
func (e *EmbeddingService) FindSemanticallySimilar(query string, chunks []CodeChunk, topK int) ([]SemanticSearchResult, error) {
	if !e.enabled {
		return nil, fmt.Errorf("embedding service not enabled")
	}
	
	// Generate embedding for query
	queryEmbedding, err := e.Embed(query)
	if err != nil {
		return nil, err
	}
	
	// Calculate similarity for each chunk
	var results []SemanticSearchResult
	for _, chunk := range chunks {
		if len(chunk.Embedding) == 0 {
			continue // Skip chunks without embeddings
		}
		
		similarity := CosineSimilarity(queryEmbedding, chunk.Embedding)
		results = append(results, SemanticSearchResult{
			Chunk:      chunk,
			Similarity: similarity,
		})
	}
	
	// Sort by similarity (highest first)
	// Simple bubble sort for small datasets
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Similarity > results[i].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	
	// Return top K
	if topK > len(results) {
		topK = len(results)
	}
	
	return results[:topK], nil
}

