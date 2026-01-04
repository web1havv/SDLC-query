package main

import (
	"fmt"
	"log"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	// Language grammars - uncomment and install as needed:
	// gotree "github.com/tree-sitter/tree-sitter-go"
	// javascript "github.com/tree-sitter/tree-sitter-javascript"
	// typescript "github.com/tree-sitter/tree-sitter-typescript"
	// python "github.com/tree-sitter/tree-sitter-python"
)

// TreeSitterParser uses Tree-sitter for accurate AST-based code parsing
type TreeSitterParser struct {
	parsers map[string]*sitter.Parser
}

// NewTreeSitterParser creates a new Tree-sitter parser
func NewTreeSitterParser() *TreeSitterParser {
	return &TreeSitterParser{
		parsers: make(map[string]*sitter.Parser),
	}
}

// ParseFile parses a file using Tree-sitter and extracts semantic chunks
func (p *TreeSitterParser) ParseFile(content []byte, filePath, language string) ([]SemanticCodeChunk, error) {
	// Try Tree-sitter parsing first
	chunks, err := p.parseWithTreeSitter(content, filePath, language)
	if err != nil {
		log.Printf("⚠️ Tree-sitter parsing failed for %s (%s): %v, falling back to regex", filePath, language, err)
		// Fallback to regex-based parsing
		return p.parseWithRegex(content, filePath, language)
	}

	if len(chunks) > 0 {
		return chunks, nil
	}

	// If Tree-sitter returned no chunks, fallback to regex
	log.Printf("⚠️ Tree-sitter returned no chunks for %s, falling back to regex", filePath)
	return p.parseWithRegex(content, filePath, language)
}

// parseWithTreeSitter parses code using Tree-sitter AST
func (p *TreeSitterParser) parseWithTreeSitter(content []byte, filePath, language string) ([]SemanticCodeChunk, error) {
	// Get language parser
	lang, err := p.getLanguage(language)
	if err != nil {
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	// Create or get parser for this language
	parser := p.getParser(language)
	if parser == nil {
		parser = sitter.NewParser()
		parser.SetLanguage(lang)
		p.parsers[language] = parser
	}

	// Parse the code
	tree := parser.Parse(nil, content)
	if tree == nil {
		return nil, fmt.Errorf("failed to parse code")
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		return nil, fmt.Errorf("no root node")
	}

	// Extract semantic chunks based on language
	var chunks []SemanticCodeChunk

	switch language {
	case "go":
		chunks = p.extractGoChunks(content, rootNode, filePath)
	case "javascript", "typescript":
		chunks = p.extractJavaScriptChunks(content, rootNode, filePath)
	case "python":
		chunks = p.extractPythonChunks(content, rootNode, filePath)
	default:
		// Generic extraction for other languages
		chunks = p.extractGenericChunks(content, rootNode, filePath, language)
	}

	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks extracted")
	}

	return chunks, nil
}

// getLanguage returns the Tree-sitter language for the given language name
func (p *TreeSitterParser) getLanguage(lang string) (*sitter.Language, error) {
	// Note: To use Tree-sitter, you need to install language grammars:
	// go get github.com/tree-sitter/tree-sitter-go
	// go get github.com/tree-sitter/tree-sitter-javascript
	// go get github.com/tree-sitter/tree-sitter-python
	// etc.
	// Then uncomment the imports and return statements below

	switch lang {
	case "go":
		// Uncomment when tree-sitter-go is installed:
		// return gotree.GetLanguage(), nil
		return nil, fmt.Errorf("Tree-sitter Go grammar not loaded. Install: go get github.com/tree-sitter/tree-sitter-go")
	case "javascript":
		// Uncomment when tree-sitter-javascript is installed:
		// return javascript.GetLanguage(), nil
		return nil, fmt.Errorf("Tree-sitter JavaScript grammar not loaded. Install: go get github.com/tree-sitter/tree-sitter-javascript")
	case "typescript":
		// Uncomment when tree-sitter-typescript is installed:
		// return typescript.GetLanguage(), nil
		return nil, fmt.Errorf("Tree-sitter TypeScript grammar not loaded. Install: go get github.com/tree-sitter/tree-sitter-typescript")
	case "python":
		// Uncomment when tree-sitter-python is installed:
		// return python.GetLanguage(), nil
		return nil, fmt.Errorf("Tree-sitter Python grammar not loaded. Install: go get github.com/tree-sitter/tree-sitter-python")
	default:
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}
}

// getParser returns a parser for the given language
func (p *TreeSitterParser) getParser(language string) *sitter.Parser {
	return p.parsers[language]
}

// extractGoChunks extracts functions, structs, interfaces from Go AST
func (p *TreeSitterParser) extractGoChunks(content []byte, rootNode *sitter.Node, filePath string) []SemanticCodeChunk {
	var chunks []SemanticCodeChunk

	// Walk the AST and extract function declarations, type declarations, etc.
	p.walkAST(rootNode, func(node *sitter.Node) {
		nodeType := node.Type()

		// Extract function declarations
		if nodeType == "function_declaration" || nodeType == "method_declaration" {
			chunk := p.createChunk(content, node, filePath, "function", "go")
			if chunk != nil {
				chunks = append(chunks, *chunk)
			}
		}

		// Extract type declarations (structs, interfaces)
		if nodeType == "type_declaration" {
			// Check if it's a struct or interface
			child := node.Child(0)
			if child != nil {
				if child.Type() == "struct_type" {
					chunk := p.createChunk(content, node, filePath, "struct", "go")
					if chunk != nil {
						chunks = append(chunks, *chunk)
					}
				} else if child.Type() == "interface_type" {
					chunk := p.createChunk(content, node, filePath, "interface", "go")
					if chunk != nil {
						chunks = append(chunks, *chunk)
					}
				}
			}
		}

		// Also extract const and var declarations (grouped)
		if nodeType == "const_declaration" || nodeType == "var_declaration" {
			chunk := p.createChunk(content, node, filePath, "declaration", "go")
			if chunk != nil && len(chunk.Content) > 50 {
				chunks = append(chunks, *chunk)
			}
		}
	}, 0, uint32(len(content)))

	return chunks
}

// extractJavaScriptChunks extracts functions, classes from JavaScript/TypeScript AST
func (p *TreeSitterParser) extractJavaScriptChunks(content []byte, rootNode *sitter.Node, filePath string) []SemanticCodeChunk {
	var chunks []SemanticCodeChunk

	p.walkAST(rootNode, func(node *sitter.Node) {
		nodeType := node.Type()

		// Extract function declarations
		if nodeType == "function_declaration" || nodeType == "function" {
			chunk := p.createChunk(content, node, filePath, "function", "javascript")
			if chunk != nil {
				chunks = append(chunks, *chunk)
			}
		}

		// Extract arrow functions (assigned to variables)
		if nodeType == "variable_declarator" {
			child := node.Child(1) // The value/initializer
			if child != nil && (child.Type() == "arrow_function" || child.Type() == "function") {
				chunk := p.createChunk(content, node, filePath, "function", "javascript")
				if chunk != nil {
					chunks = append(chunks, *chunk)
				}
			}
		}

		// Extract class declarations
		if nodeType == "class_declaration" {
			chunk := p.createChunk(content, node, filePath, "class", "javascript")
			if chunk != nil {
				chunks = append(chunks, *chunk)
			}
		}

		// Extract method definitions in classes
		if nodeType == "method_definition" {
			chunk := p.createChunk(content, node, filePath, "method", "javascript")
			if chunk != nil {
				chunks = append(chunks, *chunk)
			}
		}
	}, 0, uint32(len(content)))

	return chunks
}

// extractPythonChunks extracts functions, classes from Python AST
func (p *TreeSitterParser) extractPythonChunks(content []byte, rootNode *sitter.Node, filePath string) []SemanticCodeChunk {
	var chunks []SemanticCodeChunk

	p.walkAST(rootNode, func(node *sitter.Node) {
		nodeType := node.Type()

		// Extract function definitions
		if nodeType == "function_definition" {
			chunk := p.createChunk(content, node, filePath, "function", "python")
			if chunk != nil {
				chunks = append(chunks, *chunk)
			}
		}

		// Extract class definitions
		if nodeType == "class_definition" {
			chunk := p.createChunk(content, node, filePath, "class", "python")
			if chunk != nil {
				chunks = append(chunks, *chunk)
			}
		}
	}, 0, uint32(len(content)))

	return chunks
}

// extractGenericChunks extracts chunks generically for unsupported languages
func (p *TreeSitterParser) extractGenericChunks(content []byte, rootNode *sitter.Node, filePath, language string) []SemanticCodeChunk {
	var chunks []SemanticCodeChunk

	// Try to find common patterns: functions, classes, etc.
	p.walkAST(rootNode, func(node *sitter.Node) {
		nodeType := node.Type()
		typeLower := strings.ToLower(nodeType)

		// Look for function-like nodes
		if strings.Contains(typeLower, "function") || strings.Contains(typeLower, "method") {
			chunk := p.createChunk(content, node, filePath, "function", language)
			if chunk != nil {
				chunks = append(chunks, *chunk)
			}
		}

		// Look for class-like nodes
		if strings.Contains(typeLower, "class") || strings.Contains(typeLower, "struct") {
			chunk := p.createChunk(content, node, filePath, "class", language)
			if chunk != nil {
				chunks = append(chunks, *chunk)
			}
		}
	}, 0, uint32(len(content)))

	return chunks
}

// walkAST walks the AST tree and calls the callback for each node
func (p *TreeSitterParser) walkAST(node *sitter.Node, callback func(*sitter.Node), startByte, endByte uint32) {
	if node == nil {
		return
	}

	nodeStart := node.StartByte()
	nodeEnd := node.EndByte()

	// Check if node overlaps with range
	if nodeEnd < startByte || nodeStart > endByte {
		return
	}

	// Call callback for this node
	callback(node)

	// Recursively walk children
	childCount := int(node.ChildCount())
	for i := 0; i < childCount; i++ {
		child := node.Child(i)
		if child != nil {
			p.walkAST(child, callback, startByte, endByte)
		}
	}
}

// createChunk creates a SemanticCodeChunk from an AST node
func (p *TreeSitterParser) createChunk(content []byte, node *sitter.Node, filePath, chunkType, language string) *SemanticCodeChunk {
	if node == nil {
		return nil
	}

	startByte := node.StartByte()
	endByte := node.EndByte()

	if startByte >= endByte || endByte > uint32(len(content)) {
		return nil
	}

	// Extract code content
	codeContent := string(content[startByte:endByte])
	if len(strings.TrimSpace(codeContent)) < 20 {
		return nil // Skip very small chunks
	}

	// Calculate line numbers
	startLine := p.byteToLine(content, startByte)
	endLine := p.byteToLine(content, endByte)

	chunkID := fmt.Sprintf("%s:%d-%d", filePath, startLine+1, endLine+1)

	return &SemanticCodeChunk{
		ID:        chunkID,
		Content:   codeContent,
		FileName:  filePath,
		StartLine: startLine,
		EndLine:   endLine,
		Type:      chunkType,
		Language:  language,
	}
}

// byteToLine converts a byte offset to a line number
func (p *TreeSitterParser) byteToLine(content []byte, offset uint32) int {
	if offset > uint32(len(content)) {
		offset = uint32(len(content))
	}

	line := 0
	for i := uint32(0); i < offset; i++ {
		if content[i] == '\n' {
			line++
		}
	}
	return line
}

// parseWithRegex is the fallback regex-based parser
func (p *TreeSitterParser) parseWithRegex(content []byte, filePath, language string) ([]SemanticCodeChunk, error) {
	// This will be called by the semantic_indexer.go's existing regex-based methods
	// We'll keep the existing regex parsing as fallback
	// Return empty - let the existing regex parser handle it
	// This is just a placeholder to satisfy the interface
	return []SemanticCodeChunk{}, nil
}

