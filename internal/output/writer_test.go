package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"clangd-parser/internal/model"
)

func TestWriteJSON(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "output-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test chunks
	chunks := []model.SemanticChunk{
		{
			Name:      "testFunction",
			Signature: "void testFunction()",
			CodeType:  "Function",
			Docstring: "A test function",
			Line:      10,
			LineFrom:  10,
			LineTo:    15,
			Context: model.ChunkContext{
				Module:   "test",
				FilePath: "/test/file.cpp",
				FileName: "file.cpp",
				Snippet:  "void testFunction() {\n  // code\n}",
			},
		},
		{
			Name:      "TestClass",
			Signature: "class TestClass",
			CodeType:  "Class",
			Docstring: "A test class",
			Line:      20,
			LineFrom:  20,
			LineTo:    30,
			Context: model.ChunkContext{
				Module:   "test",
				FilePath: "/test/file.cpp",
				FileName: "file.cpp",
				Snippet:  "class TestClass {\n  // members\n};",
			},
		},
	}

	// Write to JSON
	outputPath := filepath.Join(tmpDir, "output", "chunks.json")
	err = WriteJSON(chunks, outputPath)
	if err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	// Read and parse the JSON
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var readChunks []model.SemanticChunk
	if err := json.Unmarshal(data, &readChunks); err != nil {
		t.Fatalf("Failed to parse output JSON: %v", err)
	}

	// Verify content
	if len(readChunks) != len(chunks) {
		t.Errorf("Expected %d chunks, got %d", len(chunks), len(readChunks))
	}

	if readChunks[0].Name != "testFunction" {
		t.Errorf("Expected first chunk name 'testFunction', got '%s'", readChunks[0].Name)
	}

	if readChunks[1].CodeType != "Class" {
		t.Errorf("Expected second chunk type 'Class', got '%s'", readChunks[1].CodeType)
	}

	t.Logf("✓ Successfully wrote and verified %d chunks to JSON", len(chunks))
}

func TestWriteJSONCompact(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "output-compact-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	chunks := []model.SemanticChunk{
		{
			Name:     "func1",
			CodeType: "Function",
			Line:     1,
			Context:  model.ChunkContext{FilePath: "/test.cpp"},
		},
	}

	// Write compact JSON
	outputPath := filepath.Join(tmpDir, "compact.json")
	err = WriteJSONCompact(chunks, outputPath)
	if err != nil {
		t.Fatalf("WriteJSONCompact failed: %v", err)
	}

	// Read file
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	// Verify it's valid JSON
	var readChunks []model.SemanticChunk
	if err := json.Unmarshal(data, &readChunks); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Compact JSON should not have newlines (except possibly one at end)
	dataStr := string(data)
	newlineCount := 0
	for _, c := range dataStr {
		if c == '\n' {
			newlineCount++
		}
	}

	if newlineCount > 1 {
		t.Errorf("Compact JSON has too many newlines (%d)", newlineCount)
	}

	t.Logf("✓ Compact JSON written successfully (%d bytes)", len(data))
}

func TestGetOutputStats(t *testing.T) {
	chunks := []model.SemanticChunk{
		{Name: "func1", CodeType: "Function", Docstring: "Has docs"},
		{Name: "func2", CodeType: "Function", Docstring: ""},
		{Name: "class1", CodeType: "Class", Docstring: "Has docs"},
		{Name: "method1", CodeType: "Method", Docstring: ""},
	}

	stats := GetOutputStats(chunks)

	// Verify total
	if stats["total_chunks"].(int) != 4 {
		t.Errorf("Expected total 4, got %v", stats["total_chunks"])
	}

	// Verify by type
	byType := stats["by_type"].(map[string]int)
	if byType["Function"] != 2 {
		t.Errorf("Expected 2 Functions, got %d", byType["Function"])
	}
	if byType["Class"] != 1 {
		t.Errorf("Expected 1 Class, got %d", byType["Class"])
	}
	if byType["Method"] != 1 {
		t.Errorf("Expected 1 Method, got %d", byType["Method"])
	}

	// Verify docstring count
	if stats["with_docstring"].(int) != 2 {
		t.Errorf("Expected 2 chunks with docstrings, got %v", stats["with_docstring"])
	}

	t.Logf("✓ Stats: %v", stats)
}

func TestWriteJSONEmptyChunks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "output-empty-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Empty chunks should produce valid JSON array
	chunks := []model.SemanticChunk{}
	outputPath := filepath.Join(tmpDir, "empty.json")

	err = WriteJSON(chunks, outputPath)
	if err != nil {
		t.Fatalf("WriteJSON failed on empty chunks: %v", err)
	}

	data, _ := os.ReadFile(outputPath)
	if string(data) != "[]" && string(data) != "[\n]" {
		t.Errorf("Expected empty array, got: %s", string(data))
	}

	t.Log("✓ Empty chunks handled correctly")
}
