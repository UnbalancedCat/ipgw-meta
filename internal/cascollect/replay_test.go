package cascollect

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

const syntheticLogin = "https://cas.invalid/login?service=http%3A%2F%2Fgateway.invalid%2Fservice%3Fac_id%3D7"

func scenario(page string) Scenario {
	return Scenario{Scope: Scope{"http://gateway.invalid/", "http://gateway.invalid/service", "https://cas.invalid/login"}, Exchanges: []Exchange{
		{Target: "http://gateway.invalid/", Status: 200, Body: []byte("ac_id=7")},
		{Target: syntheticLogin, Status: 200, Headers: []Header{{"Content-Type", "text/html"}, {"Set-Cookie", "opaque-synthetic-cookie"}}, Body: []byte(page)},
		{Target: "https://cas.invalid/one.js", Status: 200, Headers: []Header{{"Content-Type", "text/javascript"}}, Body: []byte("var notAKey='synthetic-body-canary';")},
	}}
}
func assertReport(t *testing.T, r Report) {
	t.Helper()
	b, err := EncodeReport(r)
	if err != nil {
		t.Fatalf("invalid emitted state: %s/%s %+v", r.Collection, r.Reason, r.Simulated)
	}
	if _, err = DecodeReport(b); err != nil {
		t.Fatal(err)
	}
	for _, v := range []string{"gateway.invalid", "cas.invalid", "opaque-synthetic-cookie", "synthetic-body-canary", "ac_id=", "service="} {
		if bytes.Contains(b, []byte(v)) {
			t.Fatal("canary echoed")
		}
	}
}
func TestContextSourcesAndDedup(t *testing.T) {
	s := scenario(`<script>publicKeyStr='invalid';</script><script src="/one.js"></script><script src="https://CAS.invalid:443/one.js"></script>`)
	r := Replay(context.Background(), s)
	assertReport(t, r)
	if r.Collection != "complete" || !r.ContextConstructed || r.Simulated != (Attempts{3, 3, 3}) || len(r.References) != 2 || len(r.Resources) != 1 || r.References[0].Resource != r.References[1].Resource || r.Resources[0].SourceID != "S5" {
		t.Fatalf("source contract: %s/%s", r.Collection, r.Reason)
	}
	if r.Analysis.Coverage.UseRelation != "unknown" || r.Analysis.Analysis != "partial" {
		t.Fatal("overclaimed analysis")
	}
	// Unknown path names are still acquired, and an inline key does not short circuit.
	s.Exchanges[2].Target = "https://cas.invalid/unrelated.js"
	s.Exchanges[1].Body = []byte(`<script>publicKeyStr='invalid';</script><script src="/unrelated.js"></script>`)
	if got := Replay(context.Background(), s); got.Collection != "complete" || got.Simulated.Requests != 3 {
		t.Fatal("legacy short circuit reused")
	}
}
func TestDiscoveryBranches(t *testing.T) {
	for _, name := range []string{"one-hop", "location-acid", "repeat-body", "conflict", "long-number", "suffix", "extra-query", "duplicate-acid", "external-location", "third-hop", "empty"} {
		t.Run(name, func(t *testing.T) {
			s := scenario(`<script src="/one.js"></script>`)
			want := "complete"
			requests := 3
			switch name {
			case "one-hop":
				s.Exchanges = append([]Exchange{{Target: s.Scope.Captive, Status: 302, Headers: []Header{{"Location", "/next"}}}, {Target: "http://gateway.invalid/next", Status: 302, Headers: []Header{{"Location", "/portal?ac_id=7"}}}}, s.Exchanges[1:]...)
				requests = 4
			case "location-acid":
				s.Exchanges[0].Status = 302
				s.Exchanges[0].Headers = []Header{{"Location", "/portal?ac_id=7"}}
				s.Exchanges[0].Body = nil
			case "repeat-body":
				s.Exchanges[0].Body = []byte("ac_id=7 ac_id=7")
			case "conflict":
				s.Exchanges[0].Body = []byte("ac_id=7 ac_id=8")
				want = "blocked"
				requests = 1
			case "long-number":
				s.Exchanges[0].Body = []byte("ac_id=12345678901")
				want = "blocked"
				requests = 1
			case "suffix":
				s.Exchanges[0].Body = []byte("ac_id=7secret")
				want = "blocked"
				requests = 1
			case "extra-query", "duplicate-acid", "external-location":
				location := "/portal?ac_id=7&ticket=synthetic-canary"
				if name == "duplicate-acid" {
					location = "/portal?ac_id=7&ac_id=7"
				}
				if name == "external-location" {
					location = "https://outside.invalid/portal?ac_id=7"
				}
				s.Exchanges[0].Status = 302
				s.Exchanges[0].Headers = []Header{{"Location", location}}
				want = "blocked"
				requests = 1
			case "third-hop":
				s.Exchanges = append([]Exchange{{Target: s.Scope.Captive, Status: 302, Headers: []Header{{"Location", "/next"}}}, {Target: "http://gateway.invalid/next", Status: 302, Headers: []Header{{"Location", "/third"}}}}, s.Exchanges[1:]...)
				want = "blocked"
				requests = 2
			case "empty":
				s.Exchanges[0].Body = nil
				want = "blocked"
				requests = 1
			}
			r := Replay(context.Background(), s)
			assertReport(t, r)
			if r.Collection != want || r.Simulated.Requests != requests {
				t.Fatalf("%s / %s / %d", r.Collection, r.Reason, r.Simulated.Requests)
			}
		})
	}
}
func TestRejectSourcesBeforeAnyScript(t *testing.T) {
	cases := []string{`/one.js?x=1`, `/one.js?`, `/one.js#x`, `https://outside.invalid/one.js`, `http://cas.invalid/one.js`, `https://u@cas.invalid/one.js`, `https://cas.invalid:444/one.js`, `https://cas.invalid:/one.js`, `../one.js`, `/%2e%2e/one.js`, `/x//one.js`, `/x\one.js`, `//cas.invalid/one.js`}
	for i, src := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			s := scenario(`<script src="/one.js"></script><script src="` + src + `"></script>`)
			r := Replay(context.Background(), s)
			assertReport(t, r)
			if r.Collection != "blocked" || r.Simulated.Requests != 2 {
				t.Fatal("unsafe URL consumed script")
			}
		})
	}
	for _, page := range []string{`<base href="/x">`, `<script type="module" src="/one.js"></script>`, `<script async src="/one.js"></script>`, `<script src="/one.js" src="/two.js"></script>`, `<script src="/one.js">notEmpty</script>`, `<script src="/one.js"/>`, `<script src="/one.js">`, `<template><script src="/one.js"></script></template>`, `<iframe></iframe>`, `<svg></svg>`, `<noscript></noscript>`} {
		s := scenario(page)
		r := Replay(context.Background(), s)
		assertReport(t, r)
		if r.Collection != "blocked" || r.Simulated.Requests != 2 {
			t.Fatal("unsupported page acquired script")
		}
	}
}
func TestResponseAndFaultStops(t *testing.T) {
	for _, name := range []string{"redirect", "location-200", "unauthorized", "forbidden", "error", "mime", "gzip", "length", "duplicate-header", "header-limit", "body-limit", "dns", "connect", "request", "unknown-fault"} {
		t.Run(name, func(t *testing.T) {
			s := scenario(`<script src="/one.js"></script>`)
			e := &s.Exchanges[2]
			switch name {
			case "redirect":
				e.Status = 302
				e.Headers = append(e.Headers, Header{"Location", "/two.js"})
			case "location-200":
				e.Headers = append(e.Headers, Header{"Location", "/two.js"})
			case "unauthorized":
				e.Status = 401
			case "forbidden":
				e.Status = 403
			case "error":
				e.Status = 500
			case "mime":
				e.Headers[0].Value = "text/html"
			case "gzip":
				e.Headers = append(e.Headers, Header{"Content-Encoding", "gzip"})
			case "length":
				e.Headers = append(e.Headers, Header{"Content-Length", "999"})
			case "duplicate-header":
				e.Headers = append(e.Headers, Header{"content-type", "text/javascript"})
			case "header-limit":
				e.Headers = append(e.Headers, Header{"X-Canary", strings.Repeat("x", MaxHeaderBytes)})
			case "body-limit":
				e.Body = make([]byte, MaxBodyBytes+1)
			case "dns", "connect", "request":
				e.Fault = name
			case "unknown-fault":
				e.Fault = "private-error-canary"
			}
			r := Replay(context.Background(), s)
			assertReport(t, r)
			if r.Collection != "blocked" || r.Simulated.Requests != 3 || r.Analysis != nil || len(r.Resources) != 0 {
				t.Fatal("failure continued or retained prefix")
			}
			if name == "dns" && r.Simulated.Connects != 2 {
				t.Fatal("DNS failure dialed")
			}
		})
	}
}
func TestCancellationAndLimits(t *testing.T) {
	s := scenario(`<script src="/one.js"></script>`)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := Replay(ctx, s)
	assertReport(t, r)
	if r.Reason != "canceled" || r.Simulated.Requests != 0 {
		t.Fatal("cancellation")
	}
	ctx, cancel = context.WithDeadline(context.Background(), time.Unix(1, 0))
	defer cancel()
	r = Replay(ctx, s)
	if r.Reason != "deadline" {
		t.Fatal("deadline")
	}
	s.Exchanges[1].Body = []byte(strings.Repeat(`<script>;</script>`, 17))
	r = Replay(context.Background(), s)
	if r.Reason != "resource_limit" || r.Simulated.Requests != 2 {
		t.Fatal("source bound")
	}
	s = scenario(`<script src="/one.js"></script>`)
	s.Exchanges[0].Body = []byte(strings.Repeat("ac_id=7 ", 33))
	r = Replay(context.Background(), s)
	if r.Reason != "resource_limit" {
		t.Fatal("candidate bound")
	}
	// Exercise the attempt boundary independently of today's tighter source count.
	rr := replay{ctx: context.Background(), report: emptyReport()}
	rr.report.Simulated = Attempts{12, 12, 12}
	_, _, reason := rr.take(Request{"GET", "http://gateway.invalid/"}, MaxBodyBytes)
	if reason != "budget_exceeded" || rr.report.Simulated.Requests != 12 {
		t.Fatal("budget exceeded")
	}
	rr = replay{ctx: context.Background(), report: emptyReport(), exchanges: []Exchange{{Target: "synthetic", Status: 200, Body: []byte("xx")}}, total: MaxTotalBytes - 1}
	_, _, reason = rr.take(Request{"GET", "synthetic"}, MaxBodyBytes)
	if reason != "resource_limit" {
		t.Fatal("total limit")
	}
}
func TestClosedReport(t *testing.T) {
	r := Replay(context.Background(), scenario(`<script src="/one.js"></script>`))
	b, err := EncodeReport(r)
	if err != nil {
		t.Fatal(err)
	}
	for _, mutated := range [][]byte{bytes.Replace(b, []byte(`"schema_version":1,`), nil, 1), bytes.Replace(b, []byte(`"schema_version":1`), []byte(`"schema_version":1,"schema_version":1`), 1), bytes.Replace(b, []byte(`"mode":"synthetic-replay"`), []byte(`"mode":"live"`), 1), append(b, []byte(`{}`)...), bytes.Replace(b, []byte(`"source_id":"S3"`), []byte(`"source_id":"S1"`), 1)} {
		if _, err = DecodeReport(mutated); err != ErrReport {
			t.Fatal("invalid report accepted")
		}
	}
	r.Simulated.DNS = 0
	b, _ = json.Marshal(r)
	if _, err = DecodeReport(b); err != ErrReport {
		t.Fatal("inconsistent counters")
	}
}

type cancelAfterChecks struct {
	context.Context
	cancel    context.CancelFunc
	remaining int
}

func (c *cancelAfterChecks) Err() error {
	c.remaining--
	if c.remaining <= 0 {
		c.cancel()
	}
	return c.Context.Err()
}
func TestCancelBetweenReplayTransitions(t *testing.T) {
	sawBeforeScript := false
	for n := 1; n <= 25; n++ {
		base, cancel := context.WithCancel(context.Background())
		ctx := &cancelAfterChecks{Context: base, cancel: cancel, remaining: n}
		r := Replay(ctx, scenario(`<script src="/one.js"></script>`))
		cancel()
		assertReport(t, r)
		if r.Reason == "canceled" {
			if r.Analysis != nil || len(r.Resources) != 0 || r.Simulated.Requests > 3 {
				t.Fatal("cancellation retained prefix or retried")
			}
			if r.Simulated.Requests == 2 {
				sawBeforeScript = true
			}
		}
	}
	if !sawBeforeScript {
		t.Fatal("missing cancellation between page and script")
	}
}
func FuzzReplayPage(f *testing.F) {
	f.Add(`<script src="/one.js"></script>`)
	f.Add(`<script>;</script>`)
	f.Fuzz(func(t *testing.T, page string) {
		if len(page) > 8192 {
			return
		}
		r := Replay(context.Background(), scenario(page))
		assertReport(t, r)
	})
}
func FuzzCollectionReport(f *testing.F) {
	b, _ := EncodeReport(Replay(context.Background(), scenario(`<script src="/one.js"></script>`)))
	f.Add(b)
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > MaxReportBytes+1 {
			return
		}
		_, err := DecodeReport(b)
		if err != nil && err != ErrReport {
			t.Fatal("unclosed error")
		}
	})
}
