package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// FewShotClient handles retrieval of similar examples from the Python service
type FewShotClient struct {
	baseURL string
	client  *http.Client
	enabled bool
}

type FewShotRequest struct {
	Query string `json:"query"`
	N     int    `json:"n"`
}

type FewShotResponse struct {
	Examples []Example `json:"examples"`
	Count    int      `json:"count"`
}

type Example struct {
	Instruction string `json:"instruction"`
	Output      string `json:"output"`
}

func NewFewShotClient(baseURL string) *FewShotClient {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	
	// Test if service is available
	enabled := false
	resp, err := client.Get(baseURL + "/health")
	if err == nil && resp.StatusCode == http.StatusOK {
		enabled = true
		log.Printf("✅ Few-shot retrieval service enabled at %s", baseURL)
	} else {
		log.Printf("⚠️ Few-shot retrieval service not available at %s (will use static examples)", baseURL)
	}
	
	return &FewShotClient{
		baseURL: baseURL,
		client:  client,
		enabled: enabled,
	}
}

func (f *FewShotClient) GetExamples(query string, n int) []Example {
	if !f.enabled {
		return []Example{} // Return empty, will use static examples
	}
	
	reqBody := FewShotRequest{
		Query: query,
		N:     n,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("❌ Error marshaling few-shot request: %v", err)
		return []Example{}
	}
	
	req, err := http.NewRequest("POST", f.baseURL+"/get_examples", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Error creating few-shot request: %v", err)
		return []Example{}
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := f.client.Do(req)
	if err != nil {
		log.Printf("⚠️ Few-shot service error: %v", err)
		return []Example{}
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("⚠️ Few-shot service returned error: %s", string(body))
		return []Example{}
	}
	
	var result FewShotResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("❌ Error decoding few-shot response: %v", err)
		return []Example{}
	}
	
	log.Printf("📚 Retrieved %d few-shot examples for query: %s", len(result.Examples), query)
	return result.Examples
}

// FormatExamplesForPrompt formats examples for injection into system prompt
func FormatExamplesForPrompt(examples []Example) string {
	if len(examples) == 0 {
		return "No examples available."
	}
	
	var buf bytes.Buffer
	buf.WriteString("RETRIEVED EXAMPLES:\n")
	for i, ex := range examples {
		buf.WriteString(fmt.Sprintf("%d. NL: \"%s\" → Zoekt: `%s`\n", i+1, ex.Instruction, ex.Output))
	}
	return buf.String()
}

