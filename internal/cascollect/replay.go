package cascollect

import (
	"context"
	"fmt"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
	"net/url"
	"strconv"
	"strings"
)

func contextReason(ctx context.Context) string {
	if ctx == nil || ctx.Err() == context.Canceled {
		return "canceled"
	}
	return "deadline"
}
func resourceID(n int) string { return fmt.Sprintf("R%d", n) }

type replay struct {
	ctx          context.Context
	exchanges    []Exchange
	index, total int
	report       Report
	next         *Request
	nextLimit    int
}

// take compares a fixed GET description with the next in-memory fixture only.
func (r *replay) take(req Request, limit int) (Exchange, map[string]string, string) {
	if r.ctx.Err() != nil {
		return Exchange{}, nil, contextReason(r.ctx)
	}
	if req.Method != "GET" {
		return Exchange{}, nil, "invalid_scenario"
	}
	if r.report.Simulated.Requests >= MaxAttempts {
		return Exchange{}, nil, "budget_exceeded"
	}
	r.report.Simulated.Requests++
	r.report.Simulated.DNS++
	if r.index >= len(r.exchanges) {
		copy := req
		r.next, r.nextLimit = &copy, limit
		return Exchange{}, nil, "invalid_scenario"
	}
	e := r.exchanges[r.index]
	r.index++
	if e.Target != req.Target {
		return Exchange{}, nil, "invalid_scenario"
	}
	if e.Fault == "dns" {
		return Exchange{}, nil, "dns_failed"
	}
	r.report.Simulated.Connects++
	switch e.Fault {
	case "":
	case "connect":
		return Exchange{}, nil, "connect_failed"
	case "request":
		return Exchange{}, nil, "request_failed"
	default:
		return Exchange{}, nil, "invalid_scenario"
	}
	h := map[string]string{}
	n := 0
	if len(e.Headers) > 64 {
		return Exchange{}, nil, "resource_limit"
	}
	for _, header := range e.Headers {
		n += len(header.Name) + len(header.Value)
		if n > MaxHeaderBytes {
			return Exchange{}, nil, "resource_limit"
		}
		name := strings.ToLower(header.Name)
		if name == "" || strings.ContainsAny(header.Value, "\r\n\x00") {
			return Exchange{}, nil, "response_rejected"
		}
		for _, c := range name {
			if !(c >= 'a' && c <= 'z') && c != '-' {
				return Exchange{}, nil, "response_rejected"
			}
		}
		switch name {
		case "content-type", "content-length", "content-encoding", "location":
			if _, ok := h[name]; ok {
				return Exchange{}, nil, "response_rejected"
			}
			h[name] = header.Value
		}
	}
	if v, ok := h["content-encoding"]; ok && v != "identity" {
		return Exchange{}, nil, "response_rejected"
	}
	if len(e.Body) > limit || r.total+len(e.Body) > MaxTotalBytes {
		return Exchange{}, nil, "resource_limit"
	}
	r.total += len(e.Body)
	if v, ok := h["content-length"]; ok {
		size, err := strconv.Atoi(v)
		if err != nil || size < 0 || strconv.Itoa(size) != v || size != len(e.Body) {
			return Exchange{}, nil, "response_rejected"
		}
	}
	return e, h, ""
}
func redirectStatus(n int) bool { return n == 301 || n == 302 || n == 303 || n == 307 || n == 308 }
func inspectDiscovery(e Exchange, h map[string]string, current, origin *url.URL, follow bool) (string, *url.URL, string) {
	var next *url.URL
	if raw, has := h["location"]; has {
		if !redirectStatus(e.Status) {
			return "", nil, "response_rejected"
		}
		u, ok := parseURL(raw, false, true)
		if !ok {
			return "", nil, "url_rejected"
		}
		u = current.ResolveReference(u)
		if u.Host != origin.Host || u.Scheme != "https" && u.Scheme != "http" {
			return "", nil, "url_rejected"
		}
		q, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			return "", nil, "url_rejected"
		}
		if a, ok := q["ac_id"]; ok {
			if len(a) != 1 {
				return "", nil, "discovery_query_duplicate"
			}
			if !acidPattern.MatchString(a[0]) {
				return "", nil, "discovery_query_invalid_value"
			}
			if reason := discoveryQueryGuard(q); reason != "" {
				return "", nil, reason
			}
			return a[0], nil, ""
		}
		if u.RawQuery != "" || u.ForceQuery {
			return "", nil, "url_rejected"
		}
		if follow {
			next = u
		}
	} else if e.Status != 200 {
		return "", nil, "response_rejected"
	}
	matches := acidBodyPattern.FindAllSubmatchIndex(e.Body, 33)
	if len(matches) > 32 {
		return "", nil, "resource_limit"
	}
	acid := ""
	for _, m := range matches {
		v := string(e.Body[m[2]:m[3]])
		if m[3] < len(e.Body) && !strings.ContainsRune("&\"' \t\r\n<>;)]}", rune(e.Body[m[3]])) {
			return "", nil, "discovery_body_boundary"
		}
		if !acidPattern.MatchString(v) {
			return "", nil, "discovery_body_value_rejected"
		}
		if acid != "" && acid != v {
			return "", nil, "discovery_body_conflict"
		}
		acid = v
	}
	if acid != "" {
		return acid, nil, ""
	}
	if _, has := h["location"]; has && !follow {
		return "", nil, "discovery_redirect_limit"
	}
	return "", next, ""
}
func Replay(ctx context.Context, s Scenario) Report {
	c, v, l, ok := parseScope(s.Scope)
	if !ok || !hostPattern.MatchString(c.Host) || !hostPattern.MatchString(v.Host) || !hostPattern.MatchString(l.Host) {
		return emptyReport()
	}
	r := replay{ctx: ctx, exchanges: s.Exchanges, report: emptyReport()}
	return r.run(s.Scope)
}

// NextOfficial is a pure planner over an in-memory response prefix. It performs
// no I/O. The intermediate Report is never a real evidence envelope; callers
// must project only its classified facts and supply measured operation counts.
func NextOfficial(ctx context.Context, prefix []Exchange) (*Request, int, Report) {
	r := replay{ctx: ctx, exchanges: prefix, report: emptyReport()}
	result := r.run(Scope{Captive: "http://ipgw.neu.edu.cn/", Service: "http://ipgw.neu.edu.cn/srun_portal_sso", CAS: "https://pass.neu.edu.cn/tpass/login"})
	return r.next, r.nextLimit, result
}

func (r *replay) run(scope Scope) Report {
	ctx := r.ctx
	stop := func(reason string) Report {
		r.report.Reason = reason
		r.report.References = []Reference{}
		r.report.Resources = []Resource{}
		r.report.Analysis = nil
		return r.report
	}
	if ctx == nil || ctx.Err() != nil {
		return stop(contextReason(ctx))
	}
	captive, service, cas, ok := parseScope(scope)
	if !ok || len(r.exchanges) > MaxAttempts {
		return stop("invalid_scenario")
	}
	current := captive
	acid := ""
	for i := 0; i < 2; i++ {
		e, h, reason := r.take(Request{"GET", current.String()}, MaxDiscoveryBytes)
		if reason != "" {
			return stop(reason)
		}
		var next *url.URL
		acid, next, reason = inspectDiscovery(e, h, current, captive, i == 0)
		if reason != "" {
			return stop(reason)
		}
		if acid != "" {
			break
		}
		if next == nil {
			return stop("discovery_no_match")
		}
		current = next
	}
	if acid == "" {
		return stop("discovery_failed")
	}
	target := loginTarget(service, cas, acid)
	if len(target) > 2048 {
		return stop("resource_limit")
	}
	r.report.ContextConstructed = true
	page, h, reason := r.take(Request{"GET", target}, MaxBodyBytes)
	if reason != "" {
		return stop(reason)
	}
	if _, ok := h["location"]; ok || page.Status != 200 || !mediaOK(h["content-type"], "page") {
		return stop("response_rejected")
	}
	p, reason := planPage(ctx, page.Body, cas)
	if reason != "" {
		return stop(reason)
	}
	if r.report.Simulated.Requests+len(p.urls) > MaxAttempts {
		return stop("budget_exceeded")
	}
	inputs := []casprovenance.Input{{Kind: "page", Data: page.Body}}
	for _, u := range p.urls {
		e, h, reason := r.take(Request{"GET", u}, MaxBodyBytes)
		if reason != "" {
			return stop(reason)
		}
		if _, ok := h["location"]; ok || e.Status != 200 || !mediaOK(h["content-type"], "script") {
			return stop("response_rejected")
		}
		inputs = append(inputs, casprovenance.Input{Kind: "external_script", Data: e.Body})
	}
	if ctx.Err() != nil {
		return stop(contextReason(ctx))
	}
	analysis, err := casprovenance.Analyze(ctx, inputs)
	if err != nil {
		if ctx.Err() != nil {
			return stop(contextReason(ctx))
		}
		return stop("analysis_rejected")
	}
	if len(analysis.Sources) != 1+p.scripts+len(p.urls) {
		return stop("analysis_rejected")
	}
	for i := range p.urls {
		n := 1 + p.scripts + i
		if analysis.Sources[n].Kind != "external_script" {
			return stop("analysis_rejected")
		}
		r.report.Resources = append(r.report.Resources, Resource{resourceID(i + 1), analysis.Sources[n].ID})
	}
	r.report.References = p.refs
	r.report.Analysis = &analysis
	r.report.Collection = "complete"
	r.report.Reason = "none"
	return r.report
}
