package casacquire

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
	"github.com/UnbalancedCat/ipgw-meta/internal/ctlcheck"
)

func acquire(ctx context.Context, w *wire) Report {
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	prefix := []cascollect.Exchange{}
	total := 0
	for {
		req, limit, r := cascollect.NextOfficial(ctx, prefix)
		if req == nil {
			o := project("anonymous-once", r, w.counts)
			if w.sourceSnapshot != nil {
				o.SourceDifference = sourceReport(w.sourceSnapshot, "comparison", "not_run", "baseline_not_observed")
			}
			attachMembership(&o, w)
			return o
		}
		if remaining := cascollect.MaxTotalBytes - total; limit > remaining {
			limit = remaining
		}
		e, reason := w.fetch(ctx, *req, limit)
		if reason != "" {
			o := blocked("anonymous-once", reason)
			o.ContextConstructed, o.Attempts = r.ContextConstructed, w.counts
			o.ChainRelation = w.chainRelation
			o.ChainTopology = w.chainTopology
			if w.sourceSnapshot != nil {
				o.SourceDifference = w.sourceDifference
				if o.SourceDifference == nil {
					o.SourceDifference = sourceReport(w.sourceSnapshot, "comparison", "not_run", "baseline_not_observed")
				}
			}
			o.Operations.ExternalRequests, o.Operations.DNS = w.counts.Requests, w.counts.DNS
			attachMembership(&o, w)
			return o
		}
		total += len(e.Body)
		prefix = append(prefix, e)
	}
}

// RunOnce performs a local topology gate before any network activity.
// Only the separate, bound native launcher offers the interactive live entry.
func RunOnce(ctx context.Context) Report {
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	w, reason := topology(ctx)
	if reason != "" {
		return blocked("anonymous-once", reason)
	}
	var ok bool
	w.rootCapture = &rootSnapshot{}
	w.roots, w.denied, ok = localRoots(w.rootCapture)
	if !ok {
		return blocked("anonymous-once", "transport_failed")
	}
	w.loadIntermediates = localIntermediates
	return acquireWithMembers(ctx, w, ctlcheck.PrepareLocalMembers, localSource)
}

func acquireWithSources(ctx context.Context, w *wire, read sourceReader) Report {
	s, gate := readSources(ctx, read)
	if gate != nil {
		reason := "source_gate_unavailable"
		if gate.Status == "no_material" {
			reason = "source_gate_no_material"
		}
		r := blocked("anonymous-once", reason)
		r.SourceDifference = gate
		return r
	}
	w.sourceSnapshot = s
	return acquire(ctx, w)
}

// Preflight is local-only: no resolver, dialer or packet send is called.
func Preflight(ctx context.Context) Report {
	_, reason := topology(ctx)
	if reason != "" {
		return blocked("local-preflight", reason)
	}
	if _, _, ok := localRoots(nil); !ok {
		return blocked("local-preflight", "transport_failed")
	}
	r := blocked("local-preflight", "none")
	r.Collection = "ready"
	return r
}

// SelfTest uses only generated keys and in-memory responses: zero sockets.
func SelfTest(ctx context.Context) Report {
	fail := func() Report { return blocked("self-test", "self_test_failed") }
	if !ctlcheck.MembershipSelfTest(ctx) || !sourceSelfTest(ctx) {
		return fail()
	}
	keys := []string{}
	for i := 0; i < 2; i++ {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return fail()
		}
		b, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if err != nil {
			return fail()
		}
		keys = append(keys, base64.StdEncoding.EncodeToString(b))
	}
	prefix := []cascollect.Exchange{}
	for i, body := range []string{"ac_id=123", `<script>var publicKeyStr = "` + keys[0] + `";</script><script src="/key.js"></script>`, `var publicKeyStr = "` + keys[1] + `";`} {
		req, _, _ := cascollect.NextOfficial(ctx, prefix)
		if req == nil {
			return fail()
		}
		media := "text/html"
		if i == 2 {
			media = "application/javascript"
		}
		prefix = append(prefix, cascollect.Exchange{Target: req.Target, Status: 200, Headers: []cascollect.Header{{Name: "Content-Type", Value: media}, {Name: "Set-Cookie", Value: "synthetic-canary"}}, Body: []byte(body)})
		if i == 0 {
			prefix[0].Status = 302
			prefix[0].Headers = append(prefix[0].Headers, cascollect.Header{Name: "Location", Value: "/never-follow?ac_id=123&hint=synthetic-canary"})
			prefix[0].Body = nil
		}
	}
	next, _, r := cascollect.NextOfficial(ctx, prefix)
	if next != nil || r.Collection != "complete" || r.Analysis == nil || len(r.Analysis.Keys) != 2 {
		return fail()
	}
	negative := append([]cascollect.Exchange{}, prefix[:2]...)
	negative[1].Body = []byte(`<script src="/key.js?forbidden=1"></script>`)
	next, _, bad := cascollect.NextOfficial(ctx, negative)
	if next != nil || bad.Collection != "blocked" {
		return fail()
	}
	for _, tc := range []struct{ query, reason string }{
		{"ac_id=123&token=", "discovery_query_auth_marker"},
		{"ac_id=123&AC_ID=123", "discovery_query_key_rejected"},
	} {
		guarded := cascollect.Exchange{Target: prefix[0].Target, Status: 302, Headers: []cascollect.Header{{Name: "Location", Value: "/never-follow?" + tc.query}}}
		next, _, rejected := cascollect.NextOfficial(ctx, []cascollect.Exchange{guarded})
		if next != nil || rejected.ContextConstructed || rejected.Reason != tc.reason {
			return fail()
		}
	}
	return project("self-test", r, cascollect.Attempts{})
}
