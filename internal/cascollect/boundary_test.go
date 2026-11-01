package cascollect

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestExactByteAndCountBoundaries(t *testing.T) {
	for _, limit := range []int{MaxDiscoveryBytes, MaxBodyBytes} {
		for _, n := range []int{limit, limit + 1} {
			rr := replay{ctx: context.Background(), report: emptyReport(), exchanges: []Exchange{{Target: "synthetic", Body: make([]byte, n)}}}
			_, _, reason := rr.take(Request{"GET", "synthetic"}, limit)
			if (reason == "") != (n == limit) {
				t.Fatal("body bound")
			}
		}
	}
	for _, n := range []int{MaxHeaderBytes, MaxHeaderBytes + 1} {
		rr := replay{ctx: context.Background(), report: emptyReport(), exchanges: []Exchange{{Target: "synthetic", Headers: []Header{{"X", strings.Repeat("s", n-1)}}}}}
		_, _, reason := rr.take(Request{"GET", "synthetic"}, MaxBodyBytes)
		if (reason == "") != (n == MaxHeaderBytes) {
			t.Fatal("header bound")
		}
	}
	for _, n := range []int{64, 65} {
		headers := make([]Header, n)
		for i := range headers {
			headers[i] = Header{"Set-Cookie", "opaque"}
		}
		rr := replay{ctx: context.Background(), report: emptyReport(), exchanges: []Exchange{{Target: "synthetic", Headers: headers}}}
		_, _, reason := rr.take(Request{"GET", "synthetic"}, MaxBodyBytes)
		if (reason == "") != (n == 64) {
			t.Fatal("header count bound")
		}
	}
	for _, n := range []int{MaxTotalBytes - 1, MaxTotalBytes} {
		rr := replay{ctx: context.Background(), report: emptyReport(), total: n, exchanges: []Exchange{{Target: "synthetic", Body: []byte("x")}}}
		_, _, reason := rr.take(Request{"GET", "synthetic"}, MaxBodyBytes)
		if (reason == "") != (n == MaxTotalBytes-1) {
			t.Fatal("total bound")
		}
	}
	rr := replay{ctx: context.Background(), report: emptyReport(), exchanges: make([]Exchange, MaxAttempts)}
	for i := range rr.exchanges {
		rr.exchanges[i] = Exchange{Target: "synthetic"}
	}
	for i := 0; i < MaxAttempts; i++ {
		if _, _, why := rr.take(Request{"GET", "synthetic"}, MaxBodyBytes); why != "" {
			t.Fatal("early budget rejection")
		}
	}
	if _, _, why := rr.take(Request{"GET", "synthetic"}, MaxBodyBytes); why != "budget_exceeded" {
		t.Fatal("attempt limit")
	}
	for _, n := range []int{32, 33} {
		s := scenario(`<p>synthetic</p>`)
		s.Exchanges[0].Body = []byte(strings.Repeat("ac_id=7 ", n))
		r := Replay(context.Background(), s)
		assertReport(t, r)
		if (r.Collection == "complete") != (n == 32) {
			t.Fatal("discovery candidate bound")
		}
	}
	for _, n := range []int{16, 17} {
		s := scenario(strings.Repeat(`<script>;</script>`, n))
		r := Replay(context.Background(), s)
		assertReport(t, r)
		if (r.Collection == "complete") != (n == 16) {
			t.Fatal("source count bound")
		}
	}
	for _, n := range []int{64, 65} {
		s := scenario(strings.Repeat("<div>", n) + strings.Repeat("</div>", n))
		r := Replay(context.Background(), s)
		assertReport(t, r)
		if (r.Collection == "complete") != (n == 64) {
			t.Fatal("depth bound")
		}
	}
}
func TestEightDistinctSourcesAndNoTruncation(t *testing.T) {
	for _, n := range []int{8, 9} {
		s := scenario("")
		s.Exchanges = s.Exchanges[:2]
		var page strings.Builder
		for i := 0; i < n; i++ {
			path := fmt.Sprintf("/asset%d.js", i)
			fmt.Fprintf(&page, `<script src="%s"></script>`, path)
			s.Exchanges = append(s.Exchanges, Exchange{Target: "https://cas.invalid" + path, Status: 200, Headers: []Header{{"Content-Type", "text/javascript"}}, Body: []byte(";")})
		}
		s.Exchanges[1].Body = []byte(page.String())
		r := Replay(context.Background(), s)
		assertReport(t, r)
		if n == 8 {
			if r.Collection != "complete" || len(r.Resources) != 8 || r.Simulated.Requests != 10 {
				t.Fatal("eight source boundary")
			}
		} else if r.Collection != "blocked" || r.Simulated.Requests != 2 || r.Reason != "resource_limit" {
			t.Fatal("partial external acquisition")
		}
	}
}
func TestScopeAndURLBoundaries(t *testing.T) {
	for _, raw := range []string{"https://real.example/login", "http://cas.invalid/login", "https://user@cas.invalid/login", "https://cas.invalid/login?x=1", "https://cas.invalid/login#x", "https://cas.invalid:0443/login"} {
		s := scenario("<p>x</p>")
		s.Scope.CAS = raw
		r := Replay(context.Background(), s)
		assertReport(t, r)
		if r.Simulated.Requests != 0 || r.Collection != "blocked" {
			t.Fatal("unsafe scope consumed fixture")
		}
	}
	for _, n := range []int{256, 257} {
		_, ok := parseURL("https://cas.invalid/"+strings.Repeat("x", n-1), true, false)
		if ok != (n == 256) {
			t.Fatal("path byte limit")
		}
	}
	for _, n := range []int{2048, 2049} {
		raw := "https://gateway.invalid/portal?x="
		raw += strings.Repeat("a", n-len(raw))
		_, ok := parseURL(raw, true, true)
		if ok != (n == 2048) {
			t.Fatal("URL byte limit")
		}
	}
}
func TestCompleteAndBlockedReceiptConsistency(t *testing.T) {
	for _, mutate := range []func(*Report){
		func(r *Report) { r.References[0].Ordinal = 2 },
		func(r *Report) { r.Resources[0].SourceID = "S1" },
		func(r *Report) { r.References = []Reference{} },
		func(r *Report) { r.Analysis.Sources[1].Kind = "external_script" },
		func(r *Report) { r.Collection = "blocked"; r.Reason = "dns_failed" },
		func(r *Report) { r.Operations.ExternalRequests = 1 },
		func(r *Report) { r.ContextConstructed = false },
	} {
		r := Replay(context.Background(), scenario(`<script src="/one.js"></script>`))
		mutate(&r)
		b, _ := json.Marshal(r)
		if _, err := DecodeReport(b); err != ErrReport {
			t.Fatal("contradictory complete receipt")
		}
	}
	for _, reason := range []string{"dns_failed", "connect_failed", "request_failed", "analysis_rejected", "unsupported_source", "discovery_failed"} {
		r := emptyReport()
		r.Reason = reason
		b, _ := json.Marshal(r)
		if _, err := DecodeReport(b); err != ErrReport {
			t.Fatal("impossible blocked receipt")
		}
	}
}
func TestConcurrentReplayHasNoSharedSession(t *testing.T) {
	s := scenario(`<script src="/one.js"></script>`)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := Replay(context.Background(), s)
			if r.Collection != "complete" || r.Simulated.Requests != 3 {
				t.Error("shared state drift")
			}
		}()
	}
	wg.Wait()
}
