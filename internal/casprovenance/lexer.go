package casprovenance

import (
	"context"
	"strings"
)

type token struct {
	kind       byte
	start, end int
	escaped    bool
}
type span struct {
	start, end int
	context    string
}
type lexed struct {
	tokens  []token
	spans   []span
	partial bool
}

func identifierStart(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' || c == '$'
}
func identifierPart(c byte) bool { return identifierStart(c) || c >= '0' && c <= '9' }
func space(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\v' || c == '\f'
}
func reserved(name string) bool {
	return strings.Contains("|await|break|case|catch|class|const|continue|debugger|default|delete|do|else|enum|export|extends|false|finally|for|function|if|implements|import|in|instanceof|interface|let|new|null|package|private|protected|public|return|static|super|switch|this|throw|true|try|typeof|var|void|while|with|yield|", "|"+name+"|")
}

func lex(ctx context.Context, data []byte, budget *int) (lexed, error) {
	l := lexed{}
	var stack []byte
	for i := 0; i < len(data); {
		if err := ctx.Err(); err != nil {
			return l, err
		}
		if space(data[i]) {
			i++
			continue
		}
		*budget++
		if *budget > MaxTokens {
			return l, ErrLimit
		}
		start := i
		c := data[i]
		unknown := func() (lexed, error) {
			l.spans = append(l.spans, span{start, len(data), "unknown"})
			l.partial = true
			return l, nil
		}
		if c == '/' {
			if i+1 == len(data) || data[i+1] != '/' && data[i+1] != '*' {
				return unknown()
			}
			block := data[i+1] == '*'
			i += 2
			closed := !block
			for i < len(data) {
				// JavaScript also terminates line comments at Unicode line
				// separators. This ASCII subset cannot classify that suffix.
				if !block && i+2 < len(data) && data[i] == 0xe2 && data[i+1] == 0x80 && (data[i+2] == 0xa8 || data[i+2] == 0xa9) {
					return unknown()
				}
				if i&255 == 0 {
					if err := ctx.Err(); err != nil {
						return l, err
					}
				}
				if block && i+1 < len(data) && data[i] == '*' && data[i+1] == '/' {
					i += 2
					closed = true
					break
				}
				if !block && (data[i] == '\r' || data[i] == '\n') {
					break
				}
				i++
			}
			if !closed {
				return l, ErrInvalid
			}
			l.spans = append(l.spans, span{start, i, "comment"})
			continue
		}
		if c == '\'' || c == '"' || c == '`' {
			i++
			escaped, closed := false, false
			for i < len(data) {
				if i&255 == 0 {
					if err := ctx.Err(); err != nil {
						return l, err
					}
				}
				if c == '`' && data[i] == '$' && i+1 < len(data) && data[i+1] == '{' {
					return unknown()
				}
				if data[i] == c {
					i++
					closed = true
					break
				}
				if c != '`' && (data[i] == '\r' || data[i] == '\n') {
					return l, ErrInvalid
				}
				if data[i] == '\\' {
					escaped = true
					i++
					if i == len(data) {
						return l, ErrInvalid
					}
					if data[i] == '\r' && i+1 < len(data) && data[i+1] == '\n' {
						i++
					}
				}
				i++
			}
			if !closed {
				return l, ErrInvalid
			}
			kind, location := byte('s'), "string"
			if c == '`' {
				kind, location = 't', "template"
			}
			l.spans = append(l.spans, span{start, i, location})
			l.tokens = append(l.tokens, token{kind, start, i, escaped})
			continue
		}
		if identifierStart(c) {
			i++
			for i < len(data) && identifierPart(data[i]) {
				i++
			}
			if i < len(data) && (data[i] >= 128 || data[i] == '\\') {
				return unknown()
			}
			l.tokens = append(l.tokens, token{'i', start, i, false})
			continue
		}
		if c >= 128 || c == '\\' || c < 32 || c == 127 {
			return unknown()
		}
		if c == '(' || c == '[' || c == '{' {
			stack = append(stack, c)
			if len(stack) > MaxDepth {
				return l, ErrLimit
			}
		} else if c == ')' || c == ']' || c == '}' {
			if len(stack) == 0 {
				return l, ErrInvalid
			}
			opener := stack[len(stack)-1]
			if !(opener == '(' && c == ')' || opener == '[' && c == ']' || opener == '{' && c == '}') {
				return l, ErrInvalid
			}
			stack = stack[:len(stack)-1]
		}
		l.tokens = append(l.tokens, token{c, i, i + 1, false})
		i++
	}
	if len(stack) != 0 {
		return l, ErrInvalid
	}
	return l, nil
}

type assignment struct {
	start   int
	class   string
	value   string
	escaped bool
}

// A deliberately narrow straight-line grammar, not a JavaScript interpreter.
func assignments(data []byte, tokens []token) ([]assignment, bool) {
	var found []assignment
	text := func(t token) string { return string(data[t.start:t.end]) }
	for i := 0; i < len(tokens); {
		if tokens[i].kind == ';' {
			i++
			continue
		}
		class := "bare"
		if tokens[i].kind != 'i' {
			return found, true
		}
		if name := text(tokens[i]); name == "var" || name == "let" || name == "const" {
			class = "declaration"
			i++
		}
		if i >= len(tokens) || tokens[i].kind != 'i' {
			return found, true
		}
		target := tokens[i]
		i++
		if reserved(text(target)) {
			return found, true
		}
		if i < len(tokens) && tokens[i].kind == '.' {
			if class != "bare" || i+1 >= len(tokens) || tokens[i+1].kind != 'i' {
				return found, true
			}
			class = "property"
			target = tokens[i+1]
			i += 2
		}
		if i+1 >= len(tokens) || tokens[i].kind != '=' || tokens[i+1].kind != 's' {
			return found, true
		}
		value := tokens[i+1]
		i += 2
		if i < len(tokens) && tokens[i].kind != ';' {
			return found, true
		}
		if name := text(target); name == "publicKeyStr" || name == "publicKey" {
			found = append(found, assignment{target.start, class, string(data[value.start+1 : value.end-1]), value.escaped})
		}
	}
	return found, false
}
