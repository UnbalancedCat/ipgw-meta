package authcontext

import (
	"bytes"
	"context"
	"github.com/PuerkitoBio/goquery"
	"github.com/UnbalancedCat/ipgw-meta/internal/authdev"
	"github.com/UnbalancedCat/ipgw-meta/internal/cas"
	"github.com/UnbalancedCat/ipgw-meta/internal/caslegacy"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func Anonymous(ctx context.Context, next http.RoundTripper, check func(context.Context) bool) Report {
	return anonymous(ctx, next, check, nil)
}

func anonymous(ctx context.Context, next http.RoundTripper, check func(context.Context) bool, observe func(context.Context, []casprovenance.Input, int)) Report {
	r := NewReport("anonymous")
	if ctx == nil || next == nil || check == nil {
		return r
	}
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	c := &collector{ctx: ctx, next: next, form: unknownForm(), matchInput: -1}
	defer c.clear()
	r.SDK = authdev.Anonymous(ctx, c, check)
	r.NonSuccessResponses = c.nonSuccess
	switch {
	case ctx.Err() != nil:
		r.Reason = "canceled"
	case c.failure != "":
		r.Reason = c.failure
	case r.SDK.State != "observed":
		r.Reason = "sdk_stopped"
	case c.matchInput < 0 || len(c.inputs) == 0:
		r.Reason = "role_uncertain"
	default:
		a, m, err := casprovenance.AnalyzeParserMatch(ctx, c.inputs, c.matchInput)
		if err != nil {
			r.Reason = "analysis_rejected"
			if a.Reason == "resource_limit" {
				r.Reason = "resource_limit"
			}
			if ctx.Err() != nil {
				r.Reason = "canceled"
			}
			return r
		}
		if m.State != "mapped" {
			r.Reason = "parser_unmapped"
			return r
		}
		r.State, r.Reason = "observed", "credential_boundary_reached"
		r.Form, r.Analysis, r.ParserMatch, r.SuccessfulSources = c.form, project(a), m, len(c.inputs)
		if observe != nil {
			observe(ctx, c.inputs, c.matchInput)
		}
	}
	return r
}

type collector struct {
	ctx                           context.Context
	next                          http.RoundTripper
	inputs                        []casprovenance.Input
	initial                       string
	form                          Form
	matchInput, total, nonSuccess int
	failure                       string
}

func (c *collector) clear() {
	for i := range c.inputs {
		clear(c.inputs[i].Data)
		c.inputs[i].Data = nil
	}
	c.inputs = nil
	c.initial = ""
	c.total = 0
}
func (c *collector) fail(reason string) {
	if c.failure == "" {
		c.failure = reason
		c.clear()
		c.form = unknownForm()
		c.matchInput = -1
	}
}
func (c *collector) RoundTrip(req *http.Request) (*http.Response, error) {
	// The unchanged authdev guard already validated this request. Never issue
	// a second request or request a source the SDK did not consume.
	res, err := c.next.RoundTrip(req)
	if !strings.EqualFold(req.URL.Hostname(), "pass.neu.edu.cn") {
		return res, err
	}
	if c.initial == "" {
		c.initial = req.URL.String()
	}
	if err != nil || res == nil {
		return res, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		c.nonSuccess++
		return res, err
	}
	if c.failure != "" {
		return res, err
	}
	if res.Body == nil {
		c.fail("incomplete_body")
		return res, err
	}
	if len(c.inputs) >= casprovenance.MaxSources {
		c.fail("resource_limit")
		return res, err
	}
	base := *req.URL
	res.Body = &bodyTee{ReadCloser: res.Body, c: c, base: &base}
	return res, err
}

// No pre-read, drain, changed n/error, or replacement of original bytes.
type bodyTee struct {
	io.ReadCloser
	c    *collector
	base *url.URL
	data []byte
	done bool
}

func (b *bodyTee) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if b.done {
		return n, err
	}
	if b.c.failure != "" {
		clear(b.data)
		b.data = nil
		b.done = true
		return n, err
	}
	if b.c.ctx.Err() != nil {
		b.c.fail("canceled")
		clear(b.data)
		b.data = nil
		b.done = true
		return n, err
	}
	weight := 1
	if len(b.c.inputs) == 0 {
		weight = 3
	} // page plus masked page and inline copies
	if n > casprovenance.MaxSourceBytes-len(b.data) || n > (casprovenance.MaxTotalBytes-b.c.total)/weight-len(b.data) {
		b.c.fail("resource_limit")
		clear(b.data)
		b.data = nil
		b.done = true
		return n, err
	}
	b.data = append(b.data, p[:n]...)
	if err != nil {
		if err == io.EOF {
			b.c.consume(b.base, b.data)
			b.data = nil
		} else {
			b.c.fail("incomplete_body")
			clear(b.data)
			b.data = nil
		}
		b.done = true
	}
	return n, err
}
func (b *bodyTee) Close() error {
	if !b.done && b.c.failure == "" {
		b.c.fail("incomplete_body")
	}
	clear(b.data)
	b.data = nil
	b.done = true
	return b.ReadCloser.Close()
}
func (c *collector) consume(base *url.URL, data []byte) {
	if c.ctx.Err() != nil {
		clear(data)
		c.fail("canceled")
		return
	}
	kind := "external_script"
	if len(c.inputs) == 0 {
		kind = "page"
		p, err := cas.ParsePage(base, data)
		if err != nil || p.Action == nil || p.Challenge != cas.ChallengeNone {
			clear(data)
			c.fail("role_uncertain")
			return
		}
		c.form = formShape(base, c.initial, data, p)
		if c.form.SameOriginHTTPS != "same" {
			clear(data)
			c.fail("role_uncertain")
			return
		}
	}
	if c.matchInput >= 0 {
		clear(data)
		c.fail("role_uncertain")
		return
	} // SDK should short-circuit.
	if caslegacy.ExtractPublicKey(data) != "" {
		c.matchInput = len(c.inputs)
	}
	weight := 1
	if kind == "page" {
		weight = 3
	}
	c.total += weight * len(data)
	c.inputs = append(c.inputs, casprovenance.Input{Kind: kind, Data: data})
}
func shape(values []string) string {
	if len(values) == 0 {
		return "missing"
	}
	if len(values) != 1 {
		return "ambiguous"
	}
	if values[0] == "" {
		return "empty"
	}
	return "unique"
}
func equal(a, b string) string {
	if a == b {
		return "same"
	}
	return "different"
}
func origin(u *url.URL) string {
	if u == nil {
		return ""
	}
	p := u.Port()
	if p == "" {
		p = "443"
	}
	return strings.ToLower(u.Hostname()) + ":" + p
}
func formShape(base *url.URL, initial string, data []byte, p cas.Page) Form {
	f := unknownForm()
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		return f
	}
	forms := doc.Find("form#loginForm")
	if forms.Length() == 0 {
		forms = doc.Find("form[action]")
	}
	f.Selection = "unique"
	if forms.Length() == 0 {
		f.Selection = "missing"
	} else if forms.Length() != 1 {
		f.Selection = "ambiguous"
	}
	f.LT, f.Execution = shape(p.Hidden["lt"]), shape(p.Hidden["execution"])
	f.SameOriginHTTPS = "different"
	start, err := url.Parse(initial)
	if err != nil || p.Action == nil {
		return f
	}
	if p.Action.Scheme == "https" && p.Action.User == nil && origin(p.Action) == origin(start) {
		f.SameOriginHTTPS = "same"
	}
	f.InitialTarget = equal(p.Action.String(), initial)
	f.FinalTarget = equal(p.Action.String(), base.String())
	f.Service = "different"
	a, ea := url.ParseQuery(p.Action.RawQuery)
	s, es := url.ParseQuery(start.RawQuery)
	if ea != nil || es != nil {
		f.Service = "unknown"
	} else if len(a["service"]) == 0 {
		f.Service = "absent"
	} else if len(a["service"]) == 1 && len(s["service"]) == 1 && a.Get("service") == s.Get("service") && s.Get("service") != "" {
		f.Service = "same"
	}
	return f
}
