package cas

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"net/url"
	"strings"
)

// KeySources separates executable classic scripts from HTML and inert data.
// External origin/path policy remains the caller's responsibility.
func KeySources(base *url.URL, data []byte) (inline [][]byte, external []*url.URL, err error) {
	if base == nil {
		return nil, nil, ErrUnrecognized
	}
	doc, e := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if e != nil {
		return nil, nil, ErrUnrecognized
	}
	// A base element changes browser URL resolution; this finite protocol does
	// not guess whether it is effective or reproduce a different source set.
	if doc.Find("base[href]").Length() != 0 {
		return nil, nil, ErrUnrecognized
	}
	doc.Find("script").Each(func(_ int, s *goquery.Selection) {
		if err != nil || s.ParentsFiltered("template,noscript").Length() > 0 {
			return
		}
		typ, _ := s.Attr("type")
		typ = strings.ToLower(strings.TrimSpace(typ))
		if typ == "module" {
			err = ErrUnrecognized
			return
		}
		if typ != "" && typ != "text/javascript" && typ != "application/javascript" && typ != "text/ecmascript" && typ != "application/ecmascript" {
			return
		}
		if _, ok := s.Attr("nomodule"); ok {
			err = ErrUnrecognized
			return
		}
		if src, ok := s.Attr("src"); ok {
			u, e := base.Parse(src)
			if e != nil || strings.TrimSpace(src) == "" {
				err = ErrUnrecognized
				return
			}
			external = append(external, u)
		} else {
			inline = append(inline, []byte(s.Text()))
		}
	})
	return
}
