package casprovenance

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
)

var rawKey = regexp.MustCompile(`(?i)(?:publicKeyStr|publicKey)\s*=\s*["']([A-Za-z0-9+/=]+)["']`)

type sourceData struct {
	kind    string
	data    []byte
	partial bool
	input   int
	offset  int
}

func splitPage(ctx context.Context, data []byte, budget *int) ([]sourceData, error) {
	page := sourceData{kind: "page", data: append([]byte(nil), data...)}
	var children []sourceData
	z := html.NewTokenizer(bytes.NewReader(data))
	z.SetMaxBuf(MaxSourceBytes)
	position, depth, scriptStart := 0, 0, -1
	ordinary := false
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		kind := z.Next()
		start := position
		position += len(z.Raw())
		if kind == html.ErrorToken {
			if z.Err() == html.ErrBufferExceeded {
				return nil, ErrLimit
			}
			if z.Err() != io.EOF || scriptStart >= 0 {
				return nil, ErrInvalid
			}
			break
		}
		*budget++
		if *budget > MaxTokens {
			return nil, ErrLimit
		}
		if kind != html.StartTagToken && kind != html.EndTagToken && kind != html.SelfClosingTagToken {
			continue
		}
		t := z.Token()
		if kind == html.StartTagToken && !strings.Contains("|area|base|br|col|embed|hr|img|input|link|meta|param|source|track|wbr|", "|"+t.Data+"|") {
			depth++
			if depth > MaxDepth {
				return nil, ErrLimit
			}
		} else if kind == html.EndTagToken && depth > 0 {
			depth--
		}
		if t.Data != "script" {
			continue
		}
		if kind == html.StartTagToken || kind == html.SelfClosingTagToken {
			scriptStart = position
			ordinary = kind == html.StartTagToken
			for _, a := range t.Attr {
				if a.Key != "type" || a.Val != "" && a.Val != "text/javascript" && a.Val != "application/javascript" {
					ordinary = false
				}
			}
			if !ordinary {
				page.partial = true
			}
		} else if kind == html.EndTagToken && scriptStart >= 0 {
			childKind := "inline_script"
			if !ordinary {
				childKind = "unknown_script"
			}
			children = append(children, sourceData{kind: childKind, data: append([]byte(nil), data[scriptStart:start]...), partial: !ordinary, offset: scriptStart})
			for i := scriptStart; i < start; i++ {
				page.data[i] = ' '
			}
			scriptStart = -1
			if len(children)+1 > MaxSources {
				return nil, ErrLimit
			}
		}
	}
	return append([]sourceData{page}, children...), nil
}

// Analyze reports facts from a limited grammar. It never decides which key to use.
// On rejection the report is empty and the error contains no input values.
func Analyze(ctx context.Context, inputs []Input) (Report, error) {
	return analyzeCore(ctx, inputs, nil)
}

func analyzeCore(ctx context.Context, inputs []Input, match *matchTracker) (Report, error) {
	r := emptyReport()
	reject := func(err error) (Report, error) {
		out := emptyReport()
		out.Analysis = "rejected"
		out.Coverage = Coverage{"unknown", "unknown", "unknown"}
		switch {
		case errors.Is(err, context.Canceled):
			out.Reason = "canceled"
		case errors.Is(err, context.DeadlineExceeded):
			out.Reason = "deadline"
		case errors.Is(err, ErrLimit):
			out.Reason = "resource_limit"
		default:
			out.Reason = "invalid_input"
		}
		return out, err
	}
	if ctx == nil || len(inputs) == 0 {
		return reject(ErrInvalid)
	}
	if err := ctx.Err(); err != nil {
		return reject(err)
	}
	if len(inputs) > MaxSources {
		return reject(ErrLimit)
	}
	total, pages, budget := 0, 0, 0
	var sources []sourceData
	for inputIndex, in := range inputs {
		if len(in.Data) > MaxSourceBytes {
			return reject(ErrLimit)
		}
		total += len(in.Data)
		if total > MaxTotalBytes {
			return reject(ErrLimit)
		}
		if !utf8.Valid(in.Data) {
			return reject(ErrInvalid)
		}
		switch in.Kind {
		case "page":
			pages++
			if pages > 1 {
				return reject(ErrInvalid)
			}
			parts, err := splitPage(ctx, in.Data, &budget)
			if err != nil {
				return reject(err)
			}
			for i := range parts {
				parts[i].input = inputIndex
			}
			sources = append(sources, parts...)
		case "inline_script", "external_script":
			sources = append(sources, sourceData{kind: in.Kind, data: in.Data, input: inputIndex})
		default:
			return reject(ErrInvalid)
		}
		if len(sources) > MaxSources {
			return reject(ErrLimit)
		}
	}
	// Count normalized slices too: HTML and extracted scripts occupy separate memory.
	total = 0
	for _, s := range sources {
		total += len(s.data)
	}
	if total > MaxTotalBytes || len(sources)-pages > MaxSources-1 {
		return reject(ErrLimit)
	}
	var decoded []*rsa.PublicKey
	for n, s := range sources {
		if err := ctx.Err(); err != nil {
			return reject(err)
		}
		id := fmt.Sprintf("S%d", n+1)
		source := Source{id, s.kind, n + 1, "complete"}
		l := lexed{}
		var candidates []assignment
		partial := s.partial
		if s.kind == "page" {
			l.spans = []span{{0, len(s.data), "html"}}
		} else if s.kind == "unknown_script" {
			l.partial = true
			l.spans = []span{{0, len(s.data), "unknown"}}
		} else {
			var err error
			l, err = lex(ctx, s.data, &budget)
			if err != nil {
				return reject(err)
			}
			var assignmentPartial bool
			grammarTokens := l.tokens
			if l.partial {
				// An unknown suffix is not EOF: the preceding literal may be
				// only the start of an expression, not the full assigned value.
				grammarTokens = append(append([]token(nil), l.tokens...), token{kind: '?'})
			}
			candidates, assignmentPartial = assignments(s.data, grammarTokens)
			partial = partial || assignmentPartial || l.partial
		}
		if partial {
			source.Analysis = "partial"
			r.Coverage.Assignment = "partial"
			r.Reason = "unsupported_syntax"
		}
		if l.partial {
			r.Coverage.Lexical = "partial"
		}
		r.Sources = append(r.Sources, source)
		positions := map[int]bool{}
		matches := rawKey.FindAllSubmatchIndex(s.data, MaxOccurrences+1)
		if len(matches) > MaxOccurrences {
			return reject(ErrLimit)
		}
		for _, m := range matches {
			positions[m[0]] = true
		}
		candidateAt := map[int]assignment{}
		for _, a := range candidates {
			positions[a.start] = true
			candidateAt[a.start] = a
		}
		ordered := make([]int, 0, len(positions))
		for pos := range positions {
			ordered = append(ordered, pos)
		}
		sort.Ints(ordered)
		for _, pos := range ordered {
			if len(r.Occurrences) >= MaxOccurrences {
				return reject(ErrLimit)
			}
			out := Occurrence{SourceID: id, Ordinal: len(r.Occurrences) + 1, Context: "code", IdentifierClass: "unknown", AssignmentClass: "none", KeyClass: "not_checked"}
			for _, p := range l.spans {
				if pos >= p.start && pos < p.end {
					out.Context = p.context
					break
				}
			}
			if out.Context == "unknown" {
				out.AssignmentClass = "unknown"
			}
			if out.Context == "code" {
				for _, t := range l.tokens {
					if pos >= t.start && pos < t.end && t.kind == 'i' {
						out.IdentifierClass = "other"
						if string(s.data[t.start:t.end]) == "publicKey" {
							out.IdentifierClass = "public_key"
						}
						if string(s.data[t.start:t.end]) == "publicKeyStr" {
							out.IdentifierClass = "public_key_str"
						}
						break
					}
				}
				if out.IdentifierClass == "public_key" || out.IdentifierClass == "public_key_str" {
					out.AssignmentClass = "unknown"
				}
			}
			if a, ok := candidateAt[pos]; ok && out.Context == "code" {
				if a.escaped {
					out.AssignmentClass = "unknown"
					r.Coverage.Assignment = "partial"
					r.Reason = "unsupported_syntax"
					r.Sources[n].Analysis = "partial"
				} else {
					out.AssignmentClass = a.class
					key, class := parseKey(a.value)
					out.KeyClass = class
					if key != nil {
						index := -1
						for i, other := range decoded {
							if other.E == key.E && other.N.Cmp(key.N) == 0 {
								index = i
								break
							}
						}
						if index < 0 {
							index = len(decoded)
							decoded = append(decoded, key)
							e := "other"
							if key.E == 65537 {
								e = "common"
							}
							r.Keys = append(r.Keys, Key{fmt.Sprintf("K%d", index+1), key.N.BitLen(), e})
						}
						ref := r.Keys[index].ID
						out.KeyRef = &ref
					}
				}
			}
			r.Occurrences = append(r.Occurrences, out)
			if match != nil && s.input == match.input && pos+s.offset == match.position {
				match.count++
				match.result = ParserMatch{State: "mapped", SourceID: id, Occurrence: out.Ordinal}
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return reject(err)
	}
	return r, nil
}

func parseKey(value string) (*rsa.PublicKey, string) {
	if len(value) > 8192 {
		return nil, "out_of_bounds"
	}
	der, err := base64.StdEncoding.Strict().DecodeString(value)
	if err != nil {
		return nil, "invalid_encoding"
	}
	parsed, err := x509.ParsePKIXPublicKey(der)
	var key *rsa.PublicKey
	if err == nil {
		var ok bool
		key, ok = parsed.(*rsa.PublicKey)
		if !ok {
			return nil, "non_rsa"
		}
	} else {
		key, err = x509.ParsePKCS1PublicKey(der)
		if err != nil {
			return nil, "invalid_encoding"
		}
	}
	if key.N == nil || key.N.Sign() <= 0 || key.N.Bit(0) == 0 || key.N.BitLen() < 512 || key.N.BitLen() > 8192 || key.E < 3 || key.E > 2147483647 || key.E%2 == 0 {
		return nil, "out_of_bounds"
	}
	return key, "valid_rsa"
}
