package cascollect

import (
	"bytes"
	"context"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
	"golang.org/x/net/html"
	"io"
	"net/url"
	"strings"
	"unicode/utf8"
)

type pagePlan struct {
	urls    []string
	refs    []Reference
	scripts int
}

func planPage(ctx context.Context, data []byte, base *url.URL) (pagePlan, string) {
	p := pagePlan{urls: []string{}, refs: []Reference{}}
	seen := map[string]int{}
	if !utf8.Valid(data) {
		return p, "response_rejected"
	}
	z := html.NewTokenizer(bytes.NewReader(data))
	z.SetMaxBuf(MaxBodyBytes)
	inside, external := false, false
	tokens, depth := 0, 0
	for {
		if ctx.Err() != nil {
			return p, contextReason(ctx)
		}
		k := z.Next()
		tokens++
		if tokens > casprovenance.MaxTokens {
			return p, "resource_limit"
		}
		if k == html.ErrorToken {
			if z.Err() != io.EOF || inside {
				return p, "unsupported_source"
			}
			break
		}
		if inside && k == html.TextToken && external && len(bytes.TrimSpace(z.Text())) != 0 {
			return p, "unsupported_source"
		}
		if k != html.StartTagToken && k != html.EndTagToken && k != html.SelfClosingTagToken {
			continue
		}
		t := z.Token()
		if k == html.StartTagToken && !strings.Contains("|area|base|br|col|embed|hr|img|input|link|meta|param|source|track|wbr|", "|"+t.Data+"|") {
			depth++
			if depth > casprovenance.MaxDepth {
				return p, "resource_limit"
			}
		} else if k == html.EndTagToken && depth > 0 {
			depth--
		}
		if k != html.EndTagToken {
			switch t.Data {
			case "base", "template", "noscript", "iframe", "svg", "math":
				return p, "unsupported_source"
			}
		}
		if t.Data != "script" {
			continue
		}
		if k == html.EndTagToken {
			if !inside {
				return p, "unsupported_source"
			}
			inside = false
			continue
		}
		if inside || k == html.SelfClosingTagToken {
			return p, "unsupported_source"
		}
		inside = true
		external = false
		p.scripts++
		if p.scripts > 16 {
			return p, "resource_limit"
		}
		attrs := map[string]bool{}
		src := ""
		for _, a := range t.Attr {
			if attrs[a.Key] {
				return p, "unsupported_source"
			}
			attrs[a.Key] = true
			switch a.Key {
			case "src":
				external = true
				src = a.Val
			case "type":
				if a.Val != "" && a.Val != "text/javascript" && a.Val != "application/javascript" {
					return p, "unsupported_source"
				}
			default:
				return p, "unsupported_source"
			}
		}
		if !external {
			continue
		}
		target, ok := scriptURL(src, base)
		if !ok {
			return p, "url_rejected"
		}
		n, found := seen[target]
		if !found {
			n = len(p.urls) + 1
			seen[target] = n
			p.urls = append(p.urls, target)
		}
		p.refs = append(p.refs, Reference{Ordinal: len(p.refs) + 1, Resource: resourceID(n)})
	}
	if 1+p.scripts+len(p.urls) > casprovenance.MaxSources {
		return p, "resource_limit"
	}
	return p, ""
}
func mediaOK(value, kind string) bool {
	parts := strings.Split(value, ";")
	if len(parts) > 2 {
		return false
	}
	typ := strings.ToLower(strings.TrimSpace(parts[0]))
	if len(parts) == 2 && strings.ToLower(strings.TrimSpace(parts[1])) != "charset=utf-8" {
		return false
	}
	if kind == "page" {
		return typ == "text/html"
	}
	return typ == "text/javascript" || typ == "application/javascript"
}
