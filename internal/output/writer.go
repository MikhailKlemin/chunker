package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"clangd-parser/internal/model"
)

// WriteJSON writes semantic chunks to a JSON file with pretty formatting
func WriteJSON(chunks []model.SemanticChunk, outputPath string) error {
	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(chunks, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// WriteJSONCompact writes chunks without indentation (smaller file size)
func WriteJSONCompact(chunks []model.SemanticChunk, outputPath string) error {
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	data, err := json.Marshal(chunks)
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// GetOutputStats returns statistics about the output
func GetOutputStats(chunks []model.SemanticChunk) map[string]any {
	stats := make(map[string]any)

	// Count by type
	typeCount := make(map[string]int)
	for _, chunk := range chunks {
		typeCount[chunk.CodeType]++
	}

	stats["total_chunks"] = len(chunks)
	stats["by_type"] = typeCount

	// Count chunks with docstrings
	withDocs := 0
	for _, chunk := range chunks {
		if chunk.Docstring != "" {
			withDocs++
		}
	}
	stats["with_docstring"] = withDocs

	return stats
}
