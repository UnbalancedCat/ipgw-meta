package casprovenance

import (
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/UnbalancedCat/ipgw-meta/internal/caslegacy"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"

	parser "github.com/UnbalancedCat/ipgw-meta/internal/jsparser"
	"github.com/dop251/goja/ast"
	jasttoken "github.com/dop251/goja/token"
)

type StaticSource struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Ordinal int    `json:"ordinal"`
}
type StaticOccurrence struct {
	SourceID        string  `json:"source_id"`
	Ordinal         int     `json:"ordinal"`
	Context         string  `json:"context"`
	IdentifierClass string  `json:"identifier_class"`
	AssignmentClass string  `json:"assignment_class"`
	Placement       string  `json:"placement"`
	KeyClass        string  `json:"key_class"`
	KeyRef          *string `json:"key_ref"`
}
type StaticReport struct {
	State       string             `json:"state"`
	Reason      string             `json:"reason"`
	UseRelation string             `json:"use_relation"`
	Sources     []StaticSource     `json:"sources"`
	Occurrences []StaticOccurrence `json:"occurrences"`
	Keys        []Key              `json:"keys"`
	ParserMatch ParserMatch        `json:"parser_match"`
}

func EmptyStatic(reason string) StaticReport {
	return StaticReport{State: "rejected", Reason: reason, UseRelation: "unknown", Sources: []StaticSource{}, Occurrences: []StaticOccurrence{}, Keys: []Key{}, ParserMatch: UnknownParserMatch()}
}

type syntaxSpan struct {
	start, end                  int
	kind, identifier, placement string
}
type syntaxAssignment struct {
	start                   int
	class, placement, value string
}
type syntaxTree struct {
	identifierEnds map[int]int
	spans          []syntaxSpan
	assignments    []syntaxAssignment
	count          *int
	ctx            context.Context
	seen           map[uintptr]bool
}
type syntaxStop string

// walk traverses only AST structs, never File/source-map data. Positions and
// literal values remain private and must not enter the diagnostic projection.
func (t *syntaxTree) walk(v reflect.Value, depth int, placement string) {
	if !v.IsValid() {
		return
	}
	if t.ctx.Err() != nil {
		panic(syntaxStop("canceled"))
	}
	if v.Kind() == reflect.Interface {
		if !v.IsNil() {
			t.walk(v.Elem(), depth, placement)
		}
		return
	}
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return
		}
		p := v.Pointer()
		if t.seen[p] {
			return
		}
		t.seen[p] = true
		t.walk(v.Elem(), depth, placement)
		return
	}
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			t.walk(v.Index(i), depth, placement)
		}
		return
	}
	if v.Kind() != reflect.Struct || v.Type().PkgPath() != "github.com/dop251/goja/ast" {
		return
	}
	depth++
	*t.count++
	if depth > 64 {
		panic(syntaxStop("depth_limit"))
	}
	if *t.count > MaxTokens {
		panic(syntaxStop("token_limit"))
	}
	if !v.CanAddr() {
		panic(syntaxStop("parser_failure"))
	}
	n := v.Addr().Interface()
	switch node := n.(type) {
	case *ast.FunctionLiteral, *ast.ArrowFunctionLiteral:
		placement = "function"
	case *ast.ClassLiteral, *ast.FieldDefinition, *ast.ClassStaticBlock, *ast.WithStatement:
		placement = "other"
	case *ast.BlockStatement:
		if placement == "top_level" {
			placement = "other"
		}
	case *ast.BinaryExpression:
		if placement != "function" && (node.Operator == jasttoken.LOGICAL_AND || node.Operator == jasttoken.LOGICAL_OR || node.Operator == jasttoken.COALESCE) {
			placement = "control"
		}
	case *ast.OptionalChain, *ast.Optional:
		if placement != "function" {
			placement = "control"
		}
	case *ast.IfStatement, *ast.ForStatement, *ast.ForInStatement, *ast.ForOfStatement, *ast.WhileStatement, *ast.DoWhileStatement, *ast.SwitchStatement, *ast.TryStatement, *ast.ConditionalExpression:
		if placement != "function" {
			placement = "control"
		}
	}
	add := func(start, end int, kind, id string) {
		if start < 0 || end < start {
			panic(syntaxStop("parser_failure"))
		}
		t.spans = append(t.spans, syntaxSpan{start, end, kind, id, placement})
	}
	switch n := n.(type) {
	case *ast.StringLiteral:
		add(int(n.Idx0())-1, int(n.Idx1())-1, "string", "unknown")
	case *ast.RegExpLiteral:
		for i := range t.spans {
			s := &t.spans[i]
			if s.kind == "regexp" && s.start == int(n.Idx0())-1 {
				s.placement = placement
			}
		}
	case *ast.TemplateLiteral:
		for i := range t.spans {
			s := &t.spans[i]
			if s.kind == "template" && s.start >= int(n.OpenQuote)-1 && s.end <= int(n.CloseQuote) {
				s.placement = placement
			}
		}
	case *ast.Identifier:
		id := "other"
		switch n.Name.String() {
		case "publicKey":
			id = "public_key"
		case "publicKeyStr":
			id = "public_key_str"
		}
		start := int(n.Idx0()) - 1
		end, ok := t.identifierEnds[start]
		if !ok {
			panic(syntaxStop("parser_failure"))
		}
		add(start, end, "code", id)
	case *ast.Binding:
		t.candidate(n.Target, n.Initializer, "declaration", placement)
	case *ast.AssignExpression:
		if n.Operator == jasttoken.ASSIGN {
			t.candidate(n.Left, n.Right, "bare", placement)
		}
	}
	for i := 0; i < v.NumField(); i++ {
		// Upstream declaration indexes alias already traversed syntax and do not
		// carry an independent occurrence or execution order.
		if v.Type().Field(i).Name == "DeclarationList" {
			continue
		}
		t.walk(v.Field(i), depth, placement)
	}
}
func (t *syntaxTree) candidate(left, right ast.Expression, class, placement string) {
	var id *ast.Identifier
	switch n := left.(type) {
	case *ast.Identifier:
		id = n
	case *ast.DotExpression:
		id = &n.Identifier
		class = "property"
	default:
		return
	}
	if id == nil || id.Name.String() != "publicKey" && id.Name.String() != "publicKeyStr" {
		return
	}
	value, ok := right.(*ast.StringLiteral)
	if !ok {
		return
	}
	// Escapes/concatenation are not silently promoted to original SDK values.
	if len(value.Literal) < 2 || strings.Contains(value.Literal, "\\") {
		return
	}
	t.assignments = append(t.assignments, syntaxAssignment{int(id.Idx0()) - 1, class, placement, value.Value.String()})
}

func AnalyzeStaticParserMatch(ctx context.Context, inputs []Input, inputIndex int) (out StaticReport) {
	out = EmptyStatic("invalid_input")
	defer func() {
		if p := recover(); p != nil {
			reason := "parser_failure"
			if s, ok := p.(syntaxStop); ok {
				reason = string(s)
			}
			out = EmptyStatic(reason)
		}
	}()
	if ctx == nil || len(inputs) == 0 || len(inputs) > MaxSources || inputIndex < 0 || inputIndex >= len(inputs) {
		return
	}
	if ctx.Err() != nil {
		return EmptyStatic("canceled")
	}
	var sources []sourceData
	total, pages, budget := 0, 0, 0
	for i, in := range inputs {
		if len(in.Data) > MaxSourceBytes || !utf8.Valid(in.Data) {
			return
		}
		total += len(in.Data)
		if total > MaxTotalBytes {
			return EmptyStatic("input_limit")
		}
		switch in.Kind {
		case "page":
			pages++
			if pages > 1 {
				return
			}
			parts, e := splitPage(ctx, in.Data, &budget)
			if e != nil {
				return EmptyStatic("invalid_input")
			}
			for j := range parts {
				parts[j].input = i
			}
			sources = append(sources, parts...)
		case "inline_script", "external_script":
			sources = append(sources, sourceData{kind: in.Kind, data: in.Data, input: i})
		default:
			return
		}
		if len(sources) > MaxSources || len(sources)-pages > MaxSources-1 {
			return EmptyStatic("input_limit")
		}
	}
	total = 0
	for _, s := range sources {
		total += len(s.data)
	}
	if total > MaxTotalBytes {
		return EmptyStatic("input_limit")
	}
	data := inputs[inputIndex].Data
	match := rawKey.FindSubmatchIndex(data)
	if len(match) != 4 || string(data[match[2]:match[3]]) != caslegacy.ExtractPublicKey(data) {
		return EmptyStatic("mapping_failed")
	}
	out = EmptyStatic("none")
	out.State = "observed"
	guard := &parser.Guard{Context: ctx, Tokens: budget}
	nodes := 0
	matched := 0
	var keys []*rsa.PublicKey
	for i, s := range sources {
		id := fmt.Sprintf("S%d", i+1)
		out.Sources = append(out.Sources, StaticSource{id, s.kind, i + 1})
		t := syntaxTree{count: &nodes, ctx: ctx, seen: map[uintptr]bool{}, identifierEnds: map[int]int{}}
		if s.kind != "page" && s.kind != "unknown_script" {
			program, comments, reason := parser.ParseLimited(s.data, guard)
			if reason != "none" {
				return EmptyStatic(reason)
			}
			for _, span := range comments {
				if span.Start < 0 || span.End > len(s.data) || span.End < span.Start {
					return EmptyStatic("parser_failure")
				}
				if span.Kind == "identifier" {
					t.identifierEnds[span.Start] = span.End
				} else {
					t.spans = append(t.spans, syntaxSpan{span.Start, span.End, span.Kind, "unknown", "unknown"})
				}
			}
			t.walk(reflect.ValueOf(program), 0, "top_level")
		}
		positions := map[int]bool{}
		candidateAt := map[int]syntaxAssignment{}
		for _, m := range rawKey.FindAllSubmatchIndex(s.data, MaxOccurrences+1) {
			positions[m[0]] = true
		}
		for _, a := range t.assignments {
			positions[a.start] = true
			candidateAt[a.start] = a
		}
		ordered := make([]int, 0, len(positions))
		for p := range positions {
			ordered = append(ordered, p)
		}
		sort.Ints(ordered)
		for _, p := range ordered {
			if len(out.Occurrences) >= MaxOccurrences {
				return EmptyStatic("input_limit")
			}
			o := StaticOccurrence{SourceID: id, Ordinal: len(out.Occurrences) + 1, Context: "unknown", IdentifierClass: "unknown", AssignmentClass: "unknown", Placement: "unknown", KeyClass: "not_checked"}
			if s.kind == "page" {
				o.Context = "html"
				o.AssignmentClass = "none"
			}
			best := len(s.data) + 1
			for _, span := range t.spans {
				if p >= span.start && p < span.end && span.end-span.start < best {
					best = span.end - span.start
					o.Context = span.kind
					o.IdentifierClass = span.identifier
					o.Placement = span.placement
				}
			}
			if o.Context != "code" && o.Context != "unknown" {
				o.AssignmentClass = "none"
			}
			if o.Context == "code" && o.IdentifierClass == "other" {
				o.AssignmentClass = "none"
			}
			if a, ok := candidateAt[p]; ok && o.Context == "code" {
				o.AssignmentClass, o.Placement = a.class, a.placement
				key, class := parseKey(a.value)
				o.KeyClass = class
				if key != nil {
					idx := -1
					for j, k := range keys {
						if k.E == key.E && k.N.Cmp(key.N) == 0 {
							idx = j
							break
						}
					}
					if idx < 0 {
						idx = len(keys)
						keys = append(keys, key)
						e := "other"
						if key.E == 65537 {
							e = "common"
						}
						out.Keys = append(out.Keys, Key{fmt.Sprintf("K%d", idx+1), key.N.BitLen(), e})
					}
					ref := out.Keys[idx].ID
					o.KeyRef = &ref
				}
			}
			out.Occurrences = append(out.Occurrences, o)
			if s.input == inputIndex && p+s.offset == match[0] {
				matched++
				out.ParserMatch = ParserMatch{State: "mapped", SourceID: id, Occurrence: o.Ordinal}
			}
		}
	}
	if ctx.Err() != nil {
		return EmptyStatic("canceled")
	}
	if matched != 1 {
		return EmptyStatic("mapping_failed")
	}
	if out.Validate() != nil {
		return EmptyStatic("parser_failure")
	}
	return
}
