package cascollect

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDiscoveryFailureDetailsAndStop(t *testing.T) {
	cases := []struct {
		name, body, location, reason string
	}{
		{"no-match", "synthetic-body-canary", "", "discovery_no_match"},
		{"unsupported-form", "var ac_id = 7;", "", "discovery_no_match"},
		{"empty", "", "", "discovery_no_match"},
		{"boundary", "ac_id=7synthetic-body-canary", "", "discovery_body_boundary"},
		{"value-length", "ac_id=12345678901", "", "discovery_body_value_rejected"},
		{"conflict", "ac_id=7 ac_id=8", "", "discovery_body_conflict"},
		{"query-duplicate", "", "/portal?ac_id=7&ac_id=7", "discovery_query_duplicate"},
		{"query-duplicate-conflict", "", "/portal?ac_id=7&ac_id=8", "discovery_query_duplicate"},
		{"query-encoded-duplicate", "", "/portal?ac_id=7&ac%5fid=7", "discovery_query_duplicate"},
		{"query-duplicate-extra", "", "/portal?ac_id=7&ac_id=7&extra=1", "discovery_query_duplicate"},
		{"query-duplicate-invalid", "", "/portal?ac_id=bad&ac_id=7&extra=1", "discovery_query_duplicate"},
		{"query-extra", "", "/portal?ac_id=7&ticket=synthetic-canary", "discovery_query_auth_marker"},
		{"query-extra-empty-key", "", "/portal?ac_id=7&=synthetic-body-canary", "discovery_query_key_rejected"},
		{"query-empty", "", "/portal?ac_id=", "discovery_query_invalid_value"},
		{"query-length", "", "/portal?ac_id=12345678901", "discovery_query_invalid_value"},
		{"query-text", "", "/portal?ac_id=synthetic-body-canary", "discovery_query_invalid_value"},
		{"query-invalid-extra", "", "/portal?ac_id=bad&extra=1", "discovery_query_invalid_value"},
		{"query-nonascii", "", "/portal?ac_id=%EF%BC%97", "discovery_query_invalid_value"},
	}
	for _, tc := range cases {
		for _, hop := range []bool{false, true} {
			name := tc.name + "/direct"
			if hop {
				name = tc.name + "/after-hop"
			}
			t.Run(name, func(t *testing.T) {
				s := scenario("")
				e := Exchange{Target: s.Scope.Captive, Status: 200, Body: []byte(tc.body)}
				if tc.location != "" {
					e.Status, e.Headers = 302, []Header{{"Location", tc.location}}
				}
				s.Exchanges = []Exchange{e}
				want := 1
				if hop {
					s.Exchanges[0].Target = "http://gateway.invalid/next"
					s.Exchanges = append([]Exchange{{Target: s.Scope.Captive, Status: 302, Headers: []Header{{"Location", "/next"}}}}, s.Exchanges...)
					want = 2
				}
				r := Replay(context.Background(), s)
				assertReport(t, r)
				if r.Reason != tc.reason || r.ContextConstructed || r.Simulated != (Attempts{want, want, want}) {
					t.Fatalf("unexpected reason or stop: %s %+v", r.Reason, r.Simulated)
				}
				// The production planner uses the same response policy. Translate only
				// the fixed synthetic targets; no network or real response is used.
				prefix := []Exchange{}
				for _, e := range s.Exchanges {
					req, _, _ := NextOfficial(context.Background(), prefix)
					if req == nil {
						t.Fatal("stopped before fixture response")
					}
					e.Target = req.Target
					prefix = append(prefix, e)
				}
				next, _, got := NextOfficial(context.Background(), prefix)
				if next != nil || got.Reason != tc.reason || got.ContextConstructed {
					t.Fatal("failure planned another request or changed classification")
				}
			})
		}
	}
}

func TestDiscoveryRedirectLimitAndPrecedence(t *testing.T) {
	for _, tc := range []struct{ body, location, reason string }{
		{"", "/third", "discovery_redirect_limit"},
		{"ac_id=7bad", "/third", "discovery_body_boundary"},
		{"ac_id=12345678901bad", "/third", "discovery_body_boundary"},
		{"ac_id=7 ac_id=8", "/third", "discovery_body_conflict"},
		{"ac_id=7", "/third?ac_id=7&token=", "discovery_query_auth_marker"},
	} {
		s := scenario("")
		s.Exchanges = []Exchange{
			{Target: s.Scope.Captive, Status: 302, Headers: []Header{{"Location", "/next"}}},
			{Target: "http://gateway.invalid/next", Status: 302, Headers: []Header{{"Location", tc.location}}, Body: []byte(tc.body)},
		}
		r := Replay(context.Background(), s)
		assertReport(t, r)
		if r.Reason != tc.reason || r.Simulated != (Attempts{2, 2, 2}) || r.ContextConstructed {
			t.Fatalf("wrong stop precedence: %s", r.Reason)
		}
	}
	// A valid body at the last allowed discovery response still constructs the
	// context, even if that response also offers a further redirect.
	prefix := []Exchange{
		{Target: "http://ipgw.neu.edu.cn/", Status: 302, Headers: []Header{{"Location", "/next"}}},
		{Target: "http://ipgw.neu.edu.cn/next", Status: 302, Headers: []Header{{"Location", "/third"}}, Body: []byte("ac_id=7 ac_id=7")},
	}
	next, _, r := NextOfficial(context.Background(), prefix)
	if next == nil || !r.ContextConstructed {
		t.Fatal("changed valid body acceptance")
	}
	prefix[1].Body = nil
	next, _, r = NextOfficial(context.Background(), prefix)
	if next != nil || r.Reason != "discovery_redirect_limit" || r.ContextConstructed {
		t.Fatal("followed forbidden third discovery request")
	}
}

func TestDiscoveryDetailReceiptValidation(t *testing.T) {
	for _, reason := range []string{"discovery_query_key_rejected", "discovery_query_auth_marker", "discovery_query_rejected", "discovery_query_duplicate", "discovery_query_invalid_value", "discovery_query_extra_fields", "discovery_body_boundary", "discovery_body_value_rejected", "discovery_body_conflict", "discovery_redirect_limit", "discovery_no_match"} {
		r := emptyReport()
		r.Reason, r.Simulated = reason, Attempts{2, 2, 2}
		assertReport(t, r)
		for _, mutate := range []func(*Report){
			func(r *Report) { r.ContextConstructed = true },
			func(r *Report) { r.Simulated = Attempts{} },
			func(r *Report) { r.Simulated = Attempts{3, 3, 3} },
			func(r *Report) { r.Simulated.Connects-- },
			func(r *Report) { r.Reason += ":synthetic-body-canary" },
		} {
			bad := r
			mutate(&bad)
			b, _ := json.Marshal(bad)
			if _, err := DecodeReport(b); err == nil {
				t.Fatal("accepted contradictory or unbounded detail")
			}
		}
	}
	r := emptyReport()
	r.Reason, r.Simulated = "discovery_failed", Attempts{2, 2, 2}
	assertReport(t, r) // Historical receipts retain their original coarse meaning.
}
