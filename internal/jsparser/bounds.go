package parser

import (
	"context"
	_ "embed"
	"github.com/dop251/goja/ast"
	"unicode/utf8"
)

type Span struct {
	Start, End int
	Kind       string
}

//go:embed LICENSE
var license string

//go:embed upstream.json
var provenance string

type stop struct{ reason string }
type Guard struct {
	Context              context.Context
	Tokens, Reads, Bytes int
	TokenLimit           int
}

func (g *Guard) check() {
	if g.Context == nil {
		panic(stop{"invalid_context"})
	}
	if g.Context.Err() != nil {
		panic(stop{"canceled"})
	}
}
func (g *Guard) token() {
	g.check()
	g.Tokens++
	limit := g.TokenLimit
	if limit == 0 {
		limit = 100000
	}
	if g.Tokens > limit {
		panic(stop{"token_limit"})
	}
}
func (g *Guard) read() {
	g.Reads++
	if g.Reads > 16<<20 {
		panic(stop{"work_limit"})
	}
	if g.Reads&255 == 0 {
		g.check()
	}
}

// ParseLimited accepts only supplied memory and never loads
// source maps. Raw errors/ASTs must not be serialized by a diagnostic caller.
func ParseLimited(data []byte, g *Guard) (program *ast.Program, comments []Span, reason string) {
	defer func() {
		if r := recover(); r != nil {
			program = nil
			comments = nil
			if s, ok := r.(stop); ok {
				reason = s.reason
			} else {
				reason = "parser_failure"
			}
		}
	}()
	if g == nil || len(data) > 1<<20 || !utf8.Valid(data) {
		return nil, nil, "invalid_input"
	}
	g.check()
	g.Bytes += len(data)
	if g.Bytes > 4<<20 {
		return nil, nil, "input_limit"
	}
	p := _newParser("synthetic", string(data), 1, WithDisableSourceMaps)
	p.guard = g
	result, err := p.parse()
	if err != nil {
		return nil, nil, "syntax_rejected"
	}
	g.check()
	return result, p.comments, "none"
}
