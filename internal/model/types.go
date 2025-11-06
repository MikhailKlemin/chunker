package model

// SemanticChunk represents a semantic code chunk with NL views
type SemanticChunk struct {
	Name      string       `json:"name"`
	Signature string       `json:"signature"`
	CodeType  string       `json:"code_type"`
	Docstring string       `json:"docstring"`
	Line      int          `json:"line"`
	LineFrom  int          `json:"line_from"`
	LineTo    int          `json:"line_to"`
	Context   ChunkContext `json:"context"`

	// NL-enhanced fields for vectorization
	TextView    string   `json:"text_view"`    // Natural language representation (384 dims with all-MiniLM-L6-v2)
	CodeView    string   `json:"code_view"`    // Code representation (768 dims with jina-embeddings-v2-base-code)
	IdentTokens []string `json:"ident_tokens"` // Subtokenized identifiers for search
}

// ChunkContext provides context information for a chunk
type ChunkContext struct {
	Module     string `json:"module"`
	FilePath   string `json:"file_path"`
	FileName   string `json:"file_name"`
	StructName string `json:"struct_name,omitempty"`
	Snippet    string `json:"snippet"`
}
