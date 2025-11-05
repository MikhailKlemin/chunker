package nl

import (
	"fmt"
	"regexp"
	"strings"

	"clangd-parser/internal/model"

	"github.com/fatih/camelcase"
)

type Views struct {
	TextView    string
	CodeView    string
	IdentTokens []string
}

func BuildViews(c model.SemanticChunk) Views {
	nameRaw := c.Name
	sigRaw := c.Signature

	identTokens := Subtokenize(nameRaw)

	nameH := Humanize(nameRaw)
	sigH := Humanize(sigRaw)

	ctx := c.Context
	doc := strings.TrimSpace(c.Docstring)
	docPart := ""
	if doc != "" {
		docPart = "that does " + doc + " "
	}

	summary := strings.TrimSpace(fmt.Sprintf(
		"%s %s %sdefined as %s module %s file %s original_name %s original_signature %s identifiers %s",
		strings.TrimSpace(c.CodeType),
		strings.TrimSpace(nameH),
		docPart,
		strings.TrimSpace(sigH),
		strings.TrimSpace(ctx.Module),
		strings.TrimSpace(ctx.FileName),
		strings.TrimSpace(nameRaw),
		strings.TrimSpace(sigRaw),
		strings.Join(identTokens, " "),
	))

	textView := TokenizeForText(summary)
	codeView := strings.TrimSpace(sigRaw + "\n" + ctx.Snippet)

	return Views{TextView: textView, CodeView: codeView, IdentTokens: identTokens}
}

// Humanize converts identifiers/signatures into space-separated words without destroying acronyms.
func Humanize(s string) string {
	if s == "" {
		return s
	}
	// Split common C++ separators first.
	repl := strings.NewReplacer("::", " ", "->", " ", ".", " ")
	s = repl.Replace(s)
	parts := splitAlphaNum(s)
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return strings.Join(filterEmpty(parts), " ")
}

// Subtokenize returns a list including the original token, its lowercase,
// and camel/snake subtokens (with lowercase variants), deduped in order.
func Subtokenize(s string) []string {
	if s == "" {
		return nil
	}
	// Normalize common separators found in C++ identifiers.
	repl := strings.NewReplacer("::", " ", "->", " ", ".", " ")
	s = repl.Replace(s)

	rawTokens := splitAlphaNum(s) // keeps letters/digits/underscore chunks
	seen := map[string]struct{}{}
	out := make([]string, 0, len(rawTokens)*4)

	add := func(t string) {
		if t == "" {
			return
		}
		if _, ok := seen[t]; ok {
			return
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}

	for _, tok := range rawTokens {
		// 1) Keep the original chunk for exact matches (e.g., QMetaObject).
		add(tok)
		add(strings.ToLower(tok))

		// 2) Split underscores, then camel-case each subtoken.
		for _, u := range strings.FieldsFunc(tok, func(r rune) bool { return r == '_' }) {
			if u == "" {
				continue
			}
			parts := camelcase.Split(u) // e.g., QMetaObject -> [Q, Meta, Object]
			for _, p := range parts {
				add(p)
				add(strings.ToLower(p))
			}
		}
	}
	return out
}

var nonWord = regexp.MustCompile(`\W+`)

func TokenizeForText(s string) string {
	toks := nonWord.Split(s, -1)
	return strings.Join(filterEmpty(toks), " ")
}

func splitAlphaNum(s string) []string {
	// Keep letters, digits, underscores, split others.
	re := regexp.MustCompile(`[^A-Za-z0-9_]+`)
	return filterEmpty(re.Split(s, -1))
}

func filterEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}
