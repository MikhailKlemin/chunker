package model

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
