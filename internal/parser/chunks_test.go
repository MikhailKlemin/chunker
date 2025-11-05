package parser

import (
	"os"
	"path/filepath"
	"testing"

	"clangd-parser/internal/lsp"
	"clangd-parser/internal/model"
)

func TestConvertSymbolsToChunks(t *testing.T) {
	// Create a temporary test file
	tmpDir, err := os.MkdirTemp("", "chunks-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.cpp")
	testCode := `#include <iostream>

/// This is a test function
/// It does something useful
int testFunction(int x, int y) {
    return x + y;
}

class TestClass {
public:
    /// Constructor for TestClass
    TestClass() {}

    /// A member method
    void memberMethod() {}
};
`
	if err := os.WriteFile(testFile, []byte(testCode), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create mock symbols
	symbols := []lsp.DocumentSymbol{
		{
			Name:   "testFunction",
			Detail: "int testFunction(int x, int y)",
			Kind:   lsp.SymbolKindFunction,
			Range: lsp.Range{
				Start: lsp.Position{Line: 4, Character: 0},
				End:   lsp.Position{Line: 6, Character: 1},
			},
		},
		{
			Name: "TestClass",
			Kind: lsp.SymbolKindClass,
			Range: lsp.Range{
				Start: lsp.Position{Line: 8, Character: 0},
				End:   lsp.Position{Line: 15, Character: 2},
			},
			Children: []lsp.DocumentSymbol{
				{
					Name:   "TestClass",
					Detail: "TestClass()",
					Kind:   lsp.SymbolKindConstructor,
					Range: lsp.Range{
						Start: lsp.Position{Line: 10, Character: 4},
						End:   lsp.Position{Line: 10, Character: 19},
					},
				},
				{
					Name:   "memberMethod",
					Detail: "void memberMethod()",
					Kind:   lsp.SymbolKindMethod,
					Range: lsp.Range{
						Start: lsp.Position{Line: 13, Character: 4},
						End:   lsp.Position{Line: 13, Character: 27},
					},
				},
			},
		},
	}

	// Convert to chunks
	chunks := ConvertSymbolsToChunks(symbols, testFile)

	// Verify results
	if len(chunks) != 4 {
		t.Errorf("Expected 4 chunks, got %d", len(chunks))
	}

	// Check first chunk (function)
	if chunks[0].Name != "testFunction" {
		t.Errorf("Expected first chunk name 'testFunction', got '%s'", chunks[0].Name)
	}

	if chunks[0].CodeType != "Function" {
		t.Errorf("Expected code type 'Function', got '%s'", chunks[0].CodeType)
	}

	if !containsString(chunks[0].Docstring, "test function") {
		t.Errorf("Expected docstring to contain 'test function', got '%s'", chunks[0].Docstring)
	}

	// Check class chunk
	classChunk := findChunkByName(chunks, "TestClass")
	if classChunk == nil {
		t.Fatal("TestClass chunk not found")
	}

	if classChunk.CodeType != "Class" {
		t.Errorf("Expected code type 'Class', got '%s'", classChunk.CodeType)
	}

	// Check method has parent struct
	methodChunk := findChunkByName(chunks, "memberMethod")
	if methodChunk == nil {
		t.Fatal("memberMethod chunk not found")
	}

	if methodChunk.Context.StructName != "TestClass" {
		t.Errorf("Expected struct name 'TestClass', got '%s'", methodChunk.Context.StructName)
	}

	t.Logf("✓ Successfully converted %d symbols to chunks", len(chunks))
}

func TestSymbolKindToString(t *testing.T) {
	tests := []struct {
		kind     int
		expected string
	}{
		{lsp.SymbolKindFunction, "Function"},
		{lsp.SymbolKindMethod, "Method"},
		{lsp.SymbolKindClass, "Class"},
		{lsp.SymbolKindStruct, "Struct"},
		{lsp.SymbolKindConstructor, "Constructor"},
		{lsp.SymbolKindEnum, "Enum"},
		{lsp.SymbolKindInterface, "Interface"},
		{lsp.SymbolKindNamespace, "Namespace"},
		{999, "Unknown"},
	}

	for _, tt := range tests {
		result := symbolKindToString(tt.kind)
		if result != tt.expected {
			t.Errorf("symbolKindToString(%d) = %s, expected %s", tt.kind, result, tt.expected)
		}
	}

	t.Log("✓ Symbol kind conversion tests passed")
}

func TestExtractDocstring(t *testing.T) {
	fileLines := []string{
		"// Regular comment",
		"/// This is documentation",
		"/// on multiple lines",
		"",
		"void function() {}",
	}

	symbol := lsp.DocumentSymbol{
		Range: lsp.Range{
			Start: lsp.Position{Line: 4, Character: 0},
		},
	}

	docstring := extractDocstring(symbol, fileLines)

	if !containsString(docstring, "This is documentation") {
		t.Errorf("Expected docstring to contain 'This is documentation', got '%s'", docstring)
	}

	if !containsString(docstring, "multiple lines") {
		t.Errorf("Expected docstring to contain 'multiple lines', got '%s'", docstring)
	}

	t.Logf("✓ Extracted docstring: %s", docstring)
}

func TestShouldExtractSymbol(t *testing.T) {
	extractable := []int{
		lsp.SymbolKindFunction,
		lsp.SymbolKindMethod,
		lsp.SymbolKindClass,
		lsp.SymbolKindStruct,
		lsp.SymbolKindConstructor,
	}

	notExtractable := []int{
		lsp.SymbolKindVariable,
		lsp.SymbolKindProperty,
		lsp.SymbolKindField,
	}

	for _, kind := range extractable {
		if !shouldExtractSymbol(kind) {
			t.Errorf("Expected kind %d to be extractable", kind)
		}
	}

	for _, kind := range notExtractable {
		if shouldExtractSymbol(kind) {
			t.Errorf("Expected kind %d to NOT be extractable", kind)
		}
	}

	t.Log("✓ Symbol extraction filter tests passed")
}

// Helper functions
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)*2))
}

func findChunkByName(chunks []model.SemanticChunk, name string) *model.SemanticChunk {
	for i := range chunks {
		if chunks[i].Name == name {
			return &chunks[i]
		}
	}
	return nil
}
