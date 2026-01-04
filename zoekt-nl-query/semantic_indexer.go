package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// SemanticIndexer indexes codebase into ChromaDB for semantic search
type SemanticIndexer struct {
	embeddingService *EmbeddingService
	chromaURL        string
	collectionName   string
	indexDir         string
	treeSitterParser *TreeSitterParser
	useTreeSitter    bool
}

// NewSemanticIndexer creates a new semantic indexer
func NewSemanticIndexer(indexDir string) (*SemanticIndexer, error) {
	embeddingService := NewEmbeddingService()
	if !embeddingService.enabled {
		return nil, fmt.Errorf("embedding service not enabled (no API key)")
	}

	// ChromaDB HTTP API URL
	chromaURL := os.Getenv("CHROMA_URL")
	if chromaURL == "" {
		chromaURL = "http://localhost:8000"
	}

	collectionName := "codebase_chunks"

	// Create or get collection
	if err := ensureCollection(chromaURL, collectionName); err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %v", err)
	}

	// Initialize Tree-sitter parser (with fallback to regex)
	treeSitterParser := NewTreeSitterParser()
	useTreeSitter := os.Getenv("USE_TREESITTER") != "false" // Default to true

	return &SemanticIndexer{
		embeddingService: embeddingService,
		chromaURL:        chromaURL,
		collectionName:   collectionName,
		indexDir:         indexDir,
		treeSitterParser: treeSitterParser,
		useTreeSitter:    useTreeSitter,
	}, nil
}

// ensureCollection creates collection if it doesn't exist
func ensureCollection(chromaURL, collectionName string) error {
	// Try to get collection first
	getURL := fmt.Sprintf("%s/api/v1/collections/%s", chromaURL, collectionName)
	resp, err := http.Get(getURL)
	if err == nil && resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		return nil // Collection exists
	}

	// Create collection
	createURL := fmt.Sprintf("%s/api/v1/collections", chromaURL)
	payload := map[string]interface{}{
		"name":         collectionName,
		"metadata":     map[string]interface{}{},
		"get_or_create": true,
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", createURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create collection: status %d", resp.StatusCode)
	}

	return nil
}

// SemanticCodeChunk represents a semantic chunk of code (for indexing)
type SemanticCodeChunk struct {
	ID        string
	Content   string
	FileName  string
	StartLine int
	EndLine   int
	Type      string // "function", "class", "struct", "interface", "block"
	Language  string
	Embedding []float32
}

// IndexCodebase indexes the entire codebase
func (idx *SemanticIndexer) IndexCodebase(codebasePath string) error {
	log.Printf("🔍 Starting codebase indexing: %s", codebasePath)

	var chunks []SemanticCodeChunk
	var totalFiles int

	// Walk through codebase
	err := filepath.WalkDir(codebasePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files and directories
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip common non-code directories
		skipDirs := []string{"node_modules", ".git", "vendor", "dist", "build", "target", "__pycache__", ".zoekt"}
		for _, skipDir := range skipDirs {
			if strings.Contains(path, skipDir) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Only process code files
		if !d.IsDir() && isCodeFile(path) {
			totalFiles++
			fileChunks, err := idx.parseFile(path)
			if err != nil {
				log.Printf("⚠️ Failed to parse %s: %v", path, err)
				return nil // Continue with other files
			}
			chunks = append(chunks, fileChunks...)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking codebase: %v", err)
	}

	log.Printf("📊 Parsed %d files, extracted %d code chunks", totalFiles, len(chunks))

	// Generate embeddings and index in batches
	batchSize := 50
	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		batch := chunks[i:end]

		if err := idx.indexChunks(batch); err != nil {
			log.Printf("⚠️ Failed to index batch %d-%d: %v", i, end, err)
			continue
		}

		log.Printf("✅ Indexed batch %d-%d/%d", i+1, end, len(chunks))
	}

	log.Printf("🎉 Successfully indexed %d code chunks", len(chunks))
	return nil
}

// parseFile parses a file into semantic chunks
func (idx *SemanticIndexer) parseFile(filePath string) ([]SemanticCodeChunk, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	language := detectLanguage(filePath)

	// Try Tree-sitter first if enabled
	if idx.useTreeSitter && idx.treeSitterParser != nil {
		chunks, err := idx.treeSitterParser.ParseFile(content, filePath, language)
		if err == nil && len(chunks) > 0 {
			log.Printf("✅ Parsed %s using Tree-sitter: %d chunks", filePath, len(chunks))
			return chunks, nil
		}
		// Fall through to regex if Tree-sitter fails
		if err != nil {
			log.Printf("⚠️ Tree-sitter parsing failed for %s: %v, using regex fallback", filePath, err)
		}
	}

	// Fallback to regex-based chunking
	chunks := idx.chunkCode(string(content), filePath, language)
	log.Printf("📝 Parsed %s using regex: %d chunks", filePath, len(chunks))
	return chunks, nil
}

// chunkCode splits code into semantic chunks
func (idx *SemanticIndexer) chunkCode(content, filePath, language string) []SemanticCodeChunk {
	var chunks []SemanticCodeChunk
	lines := strings.Split(content, "\n")

	switch language {
	case "go":
		chunks = idx.chunkGoCode(lines, filePath)
	case "javascript", "typescript":
		chunks = idx.chunkJavaScriptCode(lines, filePath)
	case "python":
		chunks = idx.chunkPythonCode(lines, filePath)
	default:
		// Generic chunking for other languages
		chunks = idx.chunkGenericCode(lines, filePath, language)
	}

	return chunks
}

// chunkGoCode chunks Go code by functions, structs, interfaces
func (idx *SemanticIndexer) chunkGoCode(lines []string, filePath string) []SemanticCodeChunk {
	var chunks []SemanticCodeChunk
	var currentChunk strings.Builder
	var startLine int
	var chunkType string
	var inBlock int

	funcPattern := regexp.MustCompile(`^func\s+(\([^)]+\)\s+)?(\w+)\s*\(`)
	structPattern := regexp.MustCompile(`^type\s+(\w+)\s+struct`)
	interfacePattern := regexp.MustCompile(`^type\s+(\w+)\s+interface`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for function
		if funcPattern.MatchString(trimmed) {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, SemanticCodeChunk{
					ID:        fmt.Sprintf("%s:%d-%d", filePath, startLine+1, i),
					Content:   currentChunk.String(),
					FileName:  filePath,
					StartLine: startLine,
					EndLine:   i - 1,
					Type:      chunkType,
					Language:  "go",
				})
			}
			currentChunk.Reset()
			startLine = i
			chunkType = "function"
			inBlock = 0
			currentChunk.WriteString(line + "\n")
			continue
		}

		// Check for struct
		if structPattern.MatchString(trimmed) {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, SemanticCodeChunk{
					ID:        fmt.Sprintf("%s:%d-%d", filePath, startLine+1, i),
					Content:   currentChunk.String(),
					FileName:  filePath,
					StartLine: startLine,
					EndLine:   i - 1,
					Type:      chunkType,
					Language:  "go",
				})
			}
			currentChunk.Reset()
			startLine = i
			chunkType = "struct"
			inBlock = 0
			currentChunk.WriteString(line + "\n")
			continue
		}

		// Check for interface
		if interfacePattern.MatchString(trimmed) {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, SemanticCodeChunk{
					ID:        fmt.Sprintf("%s:%d-%d", filePath, startLine+1, i),
					Content:   currentChunk.String(),
					FileName:  filePath,
					StartLine: startLine,
					EndLine:   i - 1,
					Type:      chunkType,
					Language:  "go",
				})
			}
			currentChunk.Reset()
			startLine = i
			chunkType = "interface"
			inBlock = 0
			currentChunk.WriteString(line + "\n")
			continue
		}

		// Track braces for block detection
		if currentChunk.Len() > 0 {
			inBlock += strings.Count(line, "{") - strings.Count(line, "}")
			currentChunk.WriteString(line + "\n")

			// If we've closed all blocks and have a meaningful chunk, save it
			if inBlock == 0 && currentChunk.Len() > 50 {
				chunks = append(chunks, SemanticCodeChunk{
					ID:        fmt.Sprintf("%s:%d-%d", filePath, startLine+1, i+1),
					Content:   currentChunk.String(),
					FileName:  filePath,
					StartLine: startLine,
					EndLine:   i,
					Type:      chunkType,
					Language:  "go",
				})
				currentChunk.Reset()
			}
		}
	}

	// Add remaining chunk
	if currentChunk.Len() > 0 {
		chunks = append(chunks, SemanticCodeChunk{
			ID:        fmt.Sprintf("%s:%d-%d", filePath, startLine+1, len(lines)),
			Content:   currentChunk.String(),
			FileName:  filePath,
			StartLine: startLine,
			EndLine:   len(lines) - 1,
			Type:      chunkType,
			Language:  "go",
		})
	}

	return chunks
}

// chunkJavaScriptCode chunks JavaScript/TypeScript code
func (idx *SemanticIndexer) chunkJavaScriptCode(lines []string, filePath string) []SemanticCodeChunk {
	var chunks []SemanticCodeChunk
	var currentChunk strings.Builder
	var startLine int
	var chunkType string
	var inBlock int

	funcPattern := regexp.MustCompile(`^(export\s+)?(async\s+)?function\s+(\w+)`)
	classPattern := regexp.MustCompile(`^(export\s+)?class\s+(\w+)`)
	constPattern := regexp.MustCompile(`^(export\s+)?const\s+(\w+)\s*=\s*(async\s+)?\(`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if funcPattern.MatchString(trimmed) || constPattern.MatchString(trimmed) {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, SemanticCodeChunk{
					ID:        fmt.Sprintf("%s:%d-%d", filePath, startLine+1, i),
					Content:   currentChunk.String(),
					FileName:  filePath,
					StartLine: startLine,
					EndLine:   i - 1,
					Type:      chunkType,
					Language:  "javascript",
				})
			}
			currentChunk.Reset()
			startLine = i
			chunkType = "function"
			inBlock = 0
			currentChunk.WriteString(line + "\n")
			continue
		}

		if classPattern.MatchString(trimmed) {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, SemanticCodeChunk{
					ID:        fmt.Sprintf("%s:%d-%d", filePath, startLine+1, i),
					Content:   currentChunk.String(),
					FileName:  filePath,
					StartLine: startLine,
					EndLine:   i - 1,
					Type:      chunkType,
					Language:  "javascript",
				})
			}
			currentChunk.Reset()
			startLine = i
			chunkType = "class"
			inBlock = 0
			currentChunk.WriteString(line + "\n")
			continue
		}

		if currentChunk.Len() > 0 {
			inBlock += strings.Count(line, "{") - strings.Count(line, "}")
			currentChunk.WriteString(line + "\n")

			if inBlock == 0 && currentChunk.Len() > 50 {
				chunks = append(chunks, SemanticCodeChunk{
					ID:        fmt.Sprintf("%s:%d-%d", filePath, startLine+1, i+1),
					Content:   currentChunk.String(),
					FileName:  filePath,
					StartLine: startLine,
					EndLine:   i,
					Type:      chunkType,
					Language:  "javascript",
				})
				currentChunk.Reset()
			}
		}
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, SemanticCodeChunk{
			ID:        fmt.Sprintf("%s:%d-%d", filePath, startLine+1, len(lines)),
			Content:   currentChunk.String(),
			FileName:  filePath,
			StartLine: startLine,
			EndLine:   len(lines) - 1,
			Type:      chunkType,
			Language:  "javascript",
		})
	}

	return chunks
}

// chunkPythonCode chunks Python code
func (idx *SemanticIndexer) chunkPythonCode(lines []string, filePath string) []SemanticCodeChunk {
	var chunks []SemanticCodeChunk
	var currentChunk strings.Builder
	var startLine int
	var chunkType string
	var indentLevel int

	funcPattern := regexp.MustCompile(`^def\s+(\w+)`)
	classPattern := regexp.MustCompile(`^class\s+(\w+)`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		currentIndent := len(line) - len(strings.TrimLeft(line, " \t"))

		if funcPattern.MatchString(trimmed) {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, SemanticCodeChunk{
					ID:        fmt.Sprintf("%s:%d-%d", filePath, startLine+1, i),
					Content:   currentChunk.String(),
					FileName:  filePath,
					StartLine: startLine,
					EndLine:   i - 1,
					Type:      chunkType,
					Language:  "python",
				})
			}
			currentChunk.Reset()
			startLine = i
			chunkType = "function"
			indentLevel = currentIndent
			currentChunk.WriteString(line + "\n")
			continue
		}

		if classPattern.MatchString(trimmed) {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, SemanticCodeChunk{
					ID:        fmt.Sprintf("%s:%d-%d", filePath, startLine+1, i),
					Content:   currentChunk.String(),
					FileName:  filePath,
					StartLine: startLine,
					EndLine:   i - 1,
					Type:      chunkType,
					Language:  "python",
				})
			}
			currentChunk.Reset()
			startLine = i
			chunkType = "class"
			indentLevel = currentIndent
			currentChunk.WriteString(line + "\n")
			continue
		}

		if currentChunk.Len() > 0 {
			if currentIndent > indentLevel {
				currentChunk.WriteString(line + "\n")
			} else if currentIndent == indentLevel && trimmed != "" {
				// Same level, continue
				currentChunk.WriteString(line + "\n")
			} else {
				// Back to outer level, save chunk
				if currentChunk.Len() > 50 {
					chunks = append(chunks, SemanticCodeChunk{
						ID:        fmt.Sprintf("%s:%d-%d", filePath, startLine+1, i),
						Content:   currentChunk.String(),
						FileName:  filePath,
						StartLine: startLine,
						EndLine:   i - 1,
						Type:      chunkType,
						Language:  "python",
					})
				}
				currentChunk.Reset()
			}
		}
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, SemanticCodeChunk{
			ID:        fmt.Sprintf("%s:%d-%d", filePath, startLine+1, len(lines)),
			Content:   currentChunk.String(),
			FileName:  filePath,
			StartLine: startLine,
			EndLine:   len(lines) - 1,
			Type:      chunkType,
			Language:  "python",
		})
	}

	return chunks
}

// chunkGenericCode chunks code generically (fallback)
func (idx *SemanticIndexer) chunkGenericCode(lines []string, filePath, language string) []SemanticCodeChunk {
	// Simple chunking: split by blank lines and size
	var chunks []SemanticCodeChunk
	var currentChunk strings.Builder
	var startLine int
	maxChunkSize := 100 // lines

	for i, line := range lines {
		currentChunk.WriteString(line + "\n")

		// Create chunk if we hit max size or blank line
		if (i-startLine >= maxChunkSize) || (strings.TrimSpace(line) == "" && currentChunk.Len() > 200) {
			if currentChunk.Len() > 50 {
				chunks = append(chunks, SemanticCodeChunk{
					ID:        fmt.Sprintf("%s:%d-%d", filePath, startLine+1, i+1),
					Content:   currentChunk.String(),
					FileName:  filePath,
					StartLine: startLine,
					EndLine:   i,
					Type:      "block",
					Language:  language,
				})
			}
			currentChunk.Reset()
			startLine = i + 1
		}
	}

	if currentChunk.Len() > 50 {
		chunks = append(chunks, SemanticCodeChunk{
			ID:        fmt.Sprintf("%s:%d-%d", filePath, startLine+1, len(lines)),
			Content:   currentChunk.String(),
			FileName:  filePath,
			StartLine: startLine,
			EndLine:   len(lines) - 1,
			Type:      "block",
			Language:  language,
		})
	}

	return chunks
}

// indexChunks generates embeddings and stores chunks in ChromaDB
func (idx *SemanticIndexer) indexChunks(chunks []SemanticCodeChunk) error {
	if len(chunks) == 0 {
		return nil
	}

	// Generate embeddings for all chunks
	var embeddings [][]float64
	var ids []string
	var documents []string
	var metadatas []map[string]interface{}

	for _, chunk := range chunks {
		// Generate embedding
		embedding, err := idx.embeddingService.Embed(chunk.Content)
		if err != nil {
			log.Printf("⚠️ Failed to generate embedding for %s: %v", chunk.ID, err)
			continue
		}

		// Convert float32 to float64 for JSON
		embedding64 := make([]float64, len(embedding))
		for i, v := range embedding {
			embedding64[i] = float64(v)
		}

		embeddings = append(embeddings, embedding64)
		ids = append(ids, chunk.ID)
		documents = append(documents, chunk.Content)

		// Store metadata
		metadata := map[string]interface{}{
			"file_name":  chunk.FileName,
			"start_line": chunk.StartLine,
			"end_line":   chunk.EndLine,
			"type":       chunk.Type,
			"language":   chunk.Language,
		}
		metadatas = append(metadatas, metadata)
	}

	if len(ids) == 0 {
		return nil
	}

	// Add to ChromaDB via HTTP API
	addURL := fmt.Sprintf("%s/api/v1/collections/%s/add", idx.chromaURL, idx.collectionName)
	payload := map[string]interface{}{
		"ids":       ids,
		"embeddings": embeddings,
		"documents": documents,
		"metadatas": metadatas,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", addURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add chunks: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add chunks: status %d", resp.StatusCode)
	}

	return nil
}

// Helper functions
func isCodeFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	codeExts := []string{".go", ".js", ".ts", ".jsx", ".tsx", ".py", ".java", ".cpp", ".c", ".h", ".rs", ".rb", ".php", ".swift", ".kt", ".scala"}
	for _, codeExt := range codeExts {
		if ext == codeExt {
			return true
		}
	}
	return false
}

func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	langMap := map[string]string{
		".go":   "go",
		".js":   "javascript",
		".jsx":  "javascript",
		".ts":   "typescript",
		".tsx":  "typescript",
		".py":   "python",
		".java": "java",
		".cpp":  "cpp",
		".c":    "c",
		".h":    "c",
		".rs":   "rust",
		".rb":   "ruby",
		".php":  "php",
		".swift": "swift",
		".kt":   "kotlin",
		".scala": "scala",
	}
	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return "unknown"
}

