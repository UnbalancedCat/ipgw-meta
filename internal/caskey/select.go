// Package caskey recognizes finite CAS login syntax, without executing JS.
package caskey

import (
	"context"
	"errors"
	parser "github.com/UnbalancedCat/ipgw-meta/internal/jsparser"
	"github.com/dop251/goja/ast"
	"github.com/dop251/goja/token"
	"reflect"
	"time"
)

var ErrUnsupported = errors.New("CAS key use relation unavailable")

// Collector retains candidate strings, not ASTs, between complete sources.
type Collector struct {
	guard parser.Guard
	nodes int
	spent time.Duration
	Keys  []string
}

func (c *Collector) Add(ctx context.Context, source []byte) (err error) {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	remaining := 10*time.Second - c.spent
	if remaining <= 0 {
		return ErrUnsupported
	}
	limited, cancel := context.WithTimeout(ctx, remaining)
	defer cancel()
	start := time.Now()
	defer func() {
		c.spent += time.Since(start)
		if recover() != nil {
			err = ErrUnsupported
		}
		if ctx.Err() != nil {
			err = ctx.Err()
		}
	}()
	c.guard.Context = limited
	program, _, reason := parser.ParseLimited(source, &c.guard)
	if reason != "none" {
		return ErrUnsupported
	}
	var funcs []*ast.FunctionLiteral
	var sinks []*ast.CallExpression
	var selectors []*ast.CallExpression
	// Only top-level ordinary declarations have an unambiguous outer binding
	// environment in this finite recognizer. Nested/conditional sinks still fail.
	for _, s := range program.Body {
		var declarations []*ast.Binding
		switch d := s.(type) {
		case *ast.VariableStatement:
			declarations = d.List
		case *ast.LexicalDeclaration:
			declarations = d.List
		}
		for _, b := range declarations {
			if name := id(b.Target); name == "$" || name == "RSAEncryptor" {
				return ErrUnsupported
			}
		}
		if d, ok := s.(*ast.FunctionDeclaration); ok {
			funcs = append(funcs, d.Function)
		}
	}
	w := walker{ctx: limited, nodes: &c.nodes}
	w.walk(reflect.ValueOf(program), 0, func(n any) bool {
		// A later source must not silently replace the protocol bindings used by
		// an earlier candidate. Ordinary JSEncrypt library definitions are distinct.
		switch x := n.(type) {
		case *ast.AssignExpression:
			if protocolTarget(x.Left) {
				panic(ErrUnsupported)
			}
		case *ast.UnaryExpression:
			if (x.Operator == token.DELETE || x.Operator == token.INCREMENT || x.Operator == token.DECREMENT) && protocolTarget(x.Operand) {
				panic(ErrUnsupported)
			}
		}
		if call, ok := n.(*ast.CallExpression); ok && selector(call, "#rsa") {
			selectors = append(selectors, call)
		}
		if call, ok := n.(*ast.CallExpression); ok && rsaWrite(call) {
			sinks = append(sinks, call)
		}
		return true
	})
	accepted := map[*ast.CallExpression]bool{}
	for _, f := range funcs {
		c.function(limited, f, accepted)
	}
	for _, sink := range sinks {
		if !accepted[sink] {
			return ErrUnsupported
		}
	}
	for _, sel := range selectors {
		matched := false
		for sink := range accepted {
			if dot(sink.Callee, "val") == sel {
				matched = true
			}
		}
		if !matched {
			return ErrUnsupported
		}
	}
	if limited.Err() != nil {
		return ErrUnsupported
	}
	return nil
}

// Only AST structs are visited. File/source-map data is never traversed.
type walker struct {
	ctx   context.Context
	nodes *int
	seen  map[uintptr]bool
}

func (w *walker) walk(v reflect.Value, depth int, visit func(any) bool) {
	if w.ctx.Err() != nil {
		panic(ErrUnsupported)
	}
	if !v.IsValid() {
		return
	}
	switch v.Kind() {
	case reflect.Interface:
		if !v.IsNil() {
			w.walk(v.Elem(), depth, visit)
		}
		return
	case reflect.Pointer:
		if v.IsNil() {
			return
		}
		if w.seen == nil {
			w.seen = map[uintptr]bool{}
		}
		p := v.Pointer()
		if w.seen[p] {
			return
		}
		w.seen[p] = true
		w.walk(v.Elem(), depth, visit)
		return
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			w.walk(v.Index(i), depth, visit)
		}
		return
	}
	if v.Kind() != reflect.Struct || v.Type().PkgPath() != "github.com/dop251/goja/ast" {
		return
	}
	*w.nodes++
	if *w.nodes > 100000 || depth >= 64 || !v.CanAddr() {
		panic(ErrUnsupported)
	}
	if !visit(v.Addr().Interface()) {
		return
	}
	for i := 0; i < v.NumField(); i++ {
		if v.Type().Field(i).Name != "DeclarationList" {
			w.walk(v.Field(i), depth+1, visit)
		}
	}
}
func id(e ast.Expression) string {
	if n, ok := e.(*ast.Identifier); ok {
		return n.Name.String()
	}
	return ""
}
func str(e ast.Expression) string {
	if n, ok := e.(*ast.StringLiteral); ok {
		return n.Value.String()
	}
	return ""
}
func dot(e ast.Expression, name string) ast.Expression {
	if d, ok := e.(*ast.DotExpression); ok && d.Identifier.Name.String() == name {
		return d.Left
	}
	return nil
}
func selector(e ast.Expression, value string) bool {
	c, ok := e.(*ast.CallExpression)
	return ok && id(c.Callee) == "$" && len(c.ArgumentList) == 1 && str(c.ArgumentList[0]) == value
}
func rsaWrite(c *ast.CallExpression) bool {
	receiver := dot(c.Callee, "val")
	if b, ok := c.Callee.(*ast.BracketExpression); ok && str(b.Member) == "val" {
		receiver = b.Left
	}
	return len(c.ArgumentList) > 0 && selector(receiver, "#rsa")
}

func member(e ast.Expression) (ast.Expression, string) {
	if d, ok := e.(*ast.DotExpression); ok {
		return d.Left, d.Identifier.Name.String()
	}
	if b, ok := e.(*ast.BracketExpression); ok {
		return b.Left, str(b.Member)
	}
	return nil, ""
}
func protocolTarget(e ast.Expression) bool {
	if name := id(e); name == "$" || name == "RSAEncryptor" {
		return true
	}
	receiver, name := member(e)
	if id(receiver) == "RSAEncryptor" {
		return true
	}
	if (id(receiver) == "window" || id(receiver) == "globalThis") && (name == "$" || name == "RSAEncryptor") {
		return true
	}
	root, part := member(receiver)
	return id(root) == "RSAEncryptor" && part == "prototype" || id(root) == "$" && part == "fn"
}
func methodTarget(e ast.Expression) bool {
	_, name := member(e)
	return name == "encrypt" || name == "val" || name == "submit"
}
func declaration(s ast.Statement, constant bool) (string, ast.Expression) {
	var list []*ast.Binding
	switch d := s.(type) {
	case *ast.LexicalDeclaration:
		if constant && d.Token != token.CONST {
			return "", nil
		}
		list = d.List
	case *ast.VariableStatement:
		if constant {
			return "", nil
		}
		list = d.List
	}
	if len(list) != 1 {
		return "", nil
	}
	return id(list[0].Target), list[0].Initializer
}
func expression(s ast.Statement) ast.Expression {
	if e, ok := s.(*ast.ExpressionStatement); ok {
		return e.Expression
	}
	return nil
}
func submit(e ast.Expression) bool {
	c, ok := e.(*ast.CallExpression)
	if !ok || len(c.ArgumentList) != 0 {
		return false
	}
	b, ok := dot(c.Callee, "submit").(*ast.BracketExpression)
	if !ok || !selector(b.Left, "#loginForm") {
		return false
	}
	n, ok := b.Member.(*ast.NumberLiteral)
	return ok && n.Literal == "0"
}
func (c *Collector) function(ctx context.Context, f *ast.FunctionLiteral, accepted map[*ast.CallExpression]bool) {
	if f.Async || f.Generator || f.Body == nil {
		return
	}
	body := f.Body.List
	possible := false
	for i := 0; i+4 < len(body); i++ {
		name, e := declaration(body[i], true)
		if name != "" && str(e) != "" {
			_, e = declaration(body[i+1], true)
			if ctor, ok := e.(*ast.NewExpression); ok && id(ctor.Callee) == "RSAEncryptor" {
				possible = true
				break
			}
		}
	}
	if !possible {
		return
	}
	counts := map[string]int{}
	changed := map[string]bool{}
	bindings := map[string]*ast.Binding{}
	uses := map[string]int{}
	hazardous := false
	w := walker{ctx: ctx, nodes: &c.nodes}
	w.walk(reflect.ValueOf(f), 0, func(n any) bool {
		switch x := n.(type) {
		case *ast.Binding:
			name := id(x.Target)
			if name == "" {
				hazardous = true
			} else {
				counts[name]++
				bindings[name] = x
			}
		case *ast.Identifier:
			uses[x.Name.String()]++
		case *ast.AssignExpression:
			if name := id(x.Left); name != "" {
				changed[name] = true
			} else { // Member mutation of protocol operations cannot be proved harmless.
				if methodTarget(x.Left) {
					hazardous = true
				}
			}
		case *ast.UnaryExpression:
			if (x.Operator == token.DELETE || x.Operator == token.INCREMENT || x.Operator == token.DECREMENT) && methodTarget(x.Operand) {
				hazardous = true
			}
			if x.Operator == token.INCREMENT || x.Operator == token.DECREMENT {
				changed[id(x.Operand)] = true
			}
		case *ast.FunctionLiteral:
			if x.Name != nil {
				counts[x.Name.Name.String()]++
			}
		case *ast.ClassLiteral:
			if x.Name != nil {
				counts[x.Name.Name.String()]++
			}
		case *ast.CatchStatement:
			if name := id(x.Parameter); name != "" {
				counts[name]++
			} else if x.Parameter != nil {
				hazardous = true
			}
		case *ast.ParameterList:
			if x.Rest != nil {
				if name := id(x.Rest); name != "" {
					counts[name]++
				} else {
					hazardous = true
				}
			}
		case *ast.WithStatement:
			hazardous = true
		case *ast.CallExpression:
			if id(x.Callee) == "eval" {
				hazardous = true
			}
		}
		return true
	})
	if hazardous || counts["$"] != 0 || counts["RSAEncryptor"] != 0 || changed["$"] || changed["RSAEncryptor"] {
		return
	}
	for i := 0; i+4 < len(body); i++ {
		key, kexpr := declaration(body[i], true)
		value := str(kexpr)
		if key == "" || value == "" || len(value) > 8192 {
			continue
		}
		instance, iexpr := declaration(body[i+1], true)
		ctor, ok := iexpr.(*ast.NewExpression)
		if !ok || instance == "" || id(ctor.Callee) != "RSAEncryptor" || len(ctor.ArgumentList) != 1 || id(ctor.ArgumentList[0]) != key {
			continue
		}
		result, rexpr := declaration(body[i+2], false)
		encrypt, ok := rexpr.(*ast.CallExpression)
		if !ok || result == "" || id(dot(encrypt.Callee, "encrypt")) != instance || len(encrypt.ArgumentList) != 1 {
			continue
		}
		sum, ok := encrypt.ArgumentList[0].(*ast.BinaryExpression)
		if !ok || sum.Operator != token.PLUS {
			continue
		}
		names := []string{}
		if !input(sum.Left, "#un", bindings, body[:i], &names) || !input(sum.Right, "#pd", bindings, body[:i], &names) {
			continue
		}
		sink, ok := expression(body[i+3]).(*ast.CallExpression)
		if !ok || !selector(dot(sink.Callee, "val"), "#rsa") || len(sink.ArgumentList) != 1 || id(sink.ArgumentList[0]) != result || !submit(expression(body[i+4])) {
			continue
		}
		valid := true
		names = append(names, key, instance, result)
		for _, name := range names {
			if counts[name] != 1 || changed[name] {
				valid = false
			}
		}
		for _, name := range []string{key, instance, result} {
			if uses[name] != 2 {
				valid = false
			}
		}
		if valid {
			accepted[sink] = true
			c.Keys = append(c.Keys, value)
		}
	}
}
func input(e ast.Expression, sel string, bindings map[string]*ast.Binding, prefix []ast.Statement, names *[]string) bool {
	name := id(e)
	b := bindings[name]
	if name == "" || b == nil || !directBinding(b, prefix) {
		return false
	}
	read, ok := b.Initializer.(*ast.CallExpression)
	if !ok || len(read.ArgumentList) != 0 {
		return false
	}
	receiver := dot(read.Callee, "val")
	if receiver == nil {
		return false
	}
	*names = append(*names, name)
	if selector(receiver, sel) {
		return true
	}
	ref := id(receiver)
	sb := bindings[ref]
	if ref == "" || sb == nil || !directBinding(sb, prefix) || sb.Idx0() >= b.Idx0() {
		return false
	}
	*names = append(*names, ref)
	return selector(sb.Initializer, sel)
}
func directBinding(b *ast.Binding, statements []ast.Statement) bool {
	for _, s := range statements {
		var list []*ast.Binding
		switch d := s.(type) {
		case *ast.VariableStatement:
			list = d.List
		case *ast.LexicalDeclaration:
			list = d.List
		}
		for _, x := range list {
			if x == b {
				return true
			}
		}
	}
	return false
}
