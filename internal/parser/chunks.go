package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"clangd-parser/internal/lsp"
)

// SemanticChunk represents a semantic code chunk
type SemanticChunk struct {
	Name      string       `json:"name"`
	Signature string       `json:"signature"`
	CodeType  string       `json:"code_type"`
	Docstring string       `json:"docstring"`
	Line      int          `json:"line"`
	LineFrom  int          `json:"line_from"`
	LineTo    int          `json:"line_to"`
	Context   ChunkContext `json:"context"`
}

// ChunkContext provides context information for a chunk
type ChunkContext struct {
	Module     string `json:"module"`
	FilePath   string `json:"file_path"`
	FileName   string `json:"file_name"`
	StructName string `json:"struct_name,omitempty"`
	Snippet    string `json:"snippet"`
}

// ConvertSymbolsToChunks converts LSP symbols to semantic chunks
func ConvertSymbolsToChunks(symbols []lsp.DocumentSymbol, filePath string) []SemanticChunk {
	fileLines := readFileLines(filePath)
	var chunks []SemanticChunk

	for _, symbol := range symbols {
		processSymbol(symbol, filePath, fileLines, "", &chunks)
	}

	return chunks
}

func processSymbol(symbol lsp.DocumentSymbol, filePath string, fileLines []string, parentStruct string, chunks *[]SemanticChunk) {
	// Extract this symbol if it's a relevant type
	if shouldExtractSymbol(symbol.Kind) {
		chunk := SemanticChunk{
			Name:      symbol.Name,
			Signature: getSignature(symbol),
			CodeType:  symbolKindToString(symbol.Kind),
			Docstring: extractDocstring(symbol, fileLines),
			Line:      symbol.Range.Start.Line + 1, // LSP is 0-indexed
			LineFrom:  symbol.Range.Start.Line + 1,
			LineTo:    symbol.Range.End.Line + 1,
			Context: ChunkContext{
				Module:     extractModule(filePath),
				FilePath:   filePath,
				FileName:   filepath.Base(filePath),
				StructName: parentStruct,
				Snippet:    extractSnippet(symbol.Range, fileLines),
			},
		}

		*chunks = append(*chunks, chunk)

		// Update parent for children if this is a class/struct
		if symbol.Kind == lsp.SymbolKindClass || symbol.Kind == lsp.SymbolKindStruct {
			parentStruct = symbol.Name
		}
	}

	// Process children recursively
	for _, child := range symbol.Children {
		processSymbol(child, filePath, fileLines, parentStruct, chunks)
	}
}

func shouldExtractSymbol(kind int) bool {
	return kind == lsp.SymbolKindFunction ||
		kind == lsp.SymbolKindMethod ||
		kind == lsp.SymbolKindClass ||
		kind == lsp.SymbolKindStruct ||
		kind == lsp.SymbolKindConstructor ||
		kind == lsp.SymbolKindEnum ||
		kind == lsp.SymbolKindInterface ||
		kind == lsp.SymbolKindNamespace
}

func symbolKindToString(kind int) string {
	kinds := map[int]string{
		lsp.SymbolKindFunction:    "Function",
		lsp.SymbolKindMethod:      "Method",
		lsp.SymbolKindClass:       "Class",
		lsp.SymbolKindStruct:      "Struct",
		lsp.SymbolKindConstructor: "Constructor",
		lsp.SymbolKindEnum:        "Enum",
		lsp.SymbolKindInterface:   "Interface",
		lsp.SymbolKindNamespace:   "Namespace",
	}

	if name, ok := kinds[kind]; ok {
		return name
	}
	return "Unknown"
}

func getSignature(symbol lsp.DocumentSymbol) string {
	if symbol.Detail != "" {
		return symbol.Detail
	}
	return symbol.Name
}

func extractDocstring(symbol lsp.DocumentSymbol, fileLines []string) string {
	startLine := symbol.Range.Start.Line
	if startLine == 0 || startLine > len(fileLines) {
		return ""
	}

	var docLines []string
	for i := startLine - 1; i >= 0 && i < len(fileLines); i-- {
		line := strings.TrimSpace(fileLines[i])

		// C++ documentation comments
		if strings.HasPrefix(line, "///") {
			doc := strings.TrimSpace(strings.TrimPrefix(line, "///"))
			docLines = append([]string{doc}, docLines...)
		} else if strings.HasPrefix(line, "//!") {
			doc := strings.TrimSpace(strings.TrimPrefix(line, "//!"))
			docLines = append([]string{doc}, docLines...)
		} else if line == "" {
			continue // Skip empty lines
		} else {
			break // Stop at non-comment line
		}
	}

	return strings.Join(docLines, " ")
}

func extractSnippet(rng lsp.Range, fileLines []string) string {
	start := rng.Start.Line
	end := rng.End.Line

	if start < 0 || end >= len(fileLines) || start > end {
		return ""
	}

	return strings.Join(fileLines[start:end+1], "\n")
}

func extractModule(filePath string) string {
	dir := filepath.Dir(filePath)
	parts := strings.Split(dir, string(filepath.Separator))

	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func readFileLines(filePath string) []string {
	file, err := os.Open(filePath)
	if err != nil {
		return []string{}
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}
