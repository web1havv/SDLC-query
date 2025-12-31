package main

import (
	"strings"
)

// BasicTranslator provides simple pattern-based translation as fallback
type BasicTranslator struct{}

func NewBasicTranslator() *BasicTranslator {
	return &BasicTranslator{}
}

// Translate provides basic translation using simple patterns
func (t *BasicTranslator) Translate(nlQuery string) (string, string) {
	query := strings.TrimSpace(nlQuery)
	lower := strings.ToLower(query)
	queryType := "search"

	// Very simple patterns - just the basics
	// Note: blogs.js may not be indexed if it's gitignored
	// So we search for "title:" pattern which appears in blog data structures
	// This is more flexible and will work even if the file path changes
	
	if strings.Contains(lower, "how many") && (strings.Contains(lower, "article") || strings.Contains(lower, "blog")) {
		// Search for "title:" which appears in blog/article data structures
		// This is more flexible than requiring a specific file
		return `title:`, "count"
	}
	
	if strings.Contains(lower, "list") && (strings.Contains(lower, "article") || strings.Contains(lower, "blog")) {
		return `title:`, "list"
	}
	
	if strings.Contains(lower, "find") && strings.Contains(lower, "article") {
		// Extract topic if possible
		if strings.Contains(lower, "about") {
			parts := strings.Split(lower, "about")
			if len(parts) > 1 {
				topic := strings.TrimSpace(parts[1])
				// Search for the topic in files that might contain blog data
				return topic, "find"
			}
		}
		// Just search for common blog-related terms
		return `title:`, "find"
	}
	
	if (strings.Contains(lower, "have") || strings.Contains(lower, "written about")) && strings.Contains(lower, "about") {
		parts := strings.Split(lower, "about")
		if len(parts) > 1 {
			topic := strings.TrimSpace(parts[1])
			// Remove common suffixes
			topic = strings.TrimSuffix(topic, " in articles")
			topic = strings.TrimSuffix(topic, " or not")
			topic = strings.TrimSpace(topic)
			return topic, "yesno"
		}
		return `title:`, "yesno"
	}

	// Default: return as-is (might be valid Zoekt query already)
	return query, queryType
}

