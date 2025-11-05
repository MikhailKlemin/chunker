package nl

import (
	"strings"
	"testing"

	"clangd-parser/internal/model"
)

func TestSubtokenize(t *testing.T) {
	tests := []struct {
		in       string
		contains []string
	}{
		{"QuaGzipFile", []string{"Qua", "Gzip", "File", "qua", "gzip", "file"}},
		{"QMetaObject::Call", []string{"QMetaObject", "Call", "qmetaobject", "call"}},
		{"QObject_ptr", []string{"QObject", "ptr", "qobject", "ptr"}},
	}
	for _, tt := range tests {
		got := Subtokenize(tt.in)
		joined := strings.Join(got, " ")
		for _, want := range tt.contains {
			if !strings.Contains(joined, want) {
				t.Fatalf("Subtokenize(%q) missing %q in %q", tt.in, want, joined)
			}
		}
	}
}

func TestBuildViews(t *testing.T) {
	chunk := model.SemanticChunk{
		Name:      "QuaGzipFile::qt_static_metacall",
		Signature: "void (QObject *, QMetaObject::Call, int, void **)",
		CodeType:  "Method",
		Docstring: "",
		Context: model.ChunkContext{
			Module:   "EWIEGA46WW",
			FileName: "moc_quagzipfile.cpp",
			Snippet:  `void QuaGzipFile::qt_static_metacall(QObject *_o, QMetaObject::Call _c, int _id, void **_a) {}`,
		},
	}
	v := BuildViews(chunk)
	if v.TextView == "" {
		t.Fatal("TextView is empty")
	}
	if !strings.Contains(v.TextView, "original_name") || !strings.Contains(v.TextView, "original_signature") {
		t.Fatalf("TextView missing raw fields: %q", v.TextView)
	}
	wantCode := chunk.Signature + "\n" + chunk.Context.Snippet
	if v.CodeView != wantCode {
		t.Fatalf("CodeView mismatch:\nwant: %q\ngot:  %q", wantCode, v.CodeView)
	}
	if len(v.IdentTokens) == 0 {
		t.Fatal("IdentTokens empty")
	}
}
