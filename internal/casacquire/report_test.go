package casacquire

import (
	"context"
	"strings"
	"testing"

	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
	"golang.org/x/net/dns/dnsmessage"
)

func TestSelfTestClosedOutput(t *testing.T) {
	r := SelfTest(context.Background())
	if r.Collection != "complete" || len(r.Analysis.Keys) != 2 {
		t.Fatal("self test failed")
	}
	b, err := Encode(r)
	if err != nil {
		t.Fatal(err)
	}
	for _, canary := range []string{"synthetic-canary", "neu.edu.cn", "ac_id", "publicKeyStr", "MIIB", "synthetic-prototype", "simulated_attempts"} {
		if strings.Contains(string(b), canary) {
			t.Fatal("unexpected output data")
		}
	}
	if _, err = Decode(b); err != nil {
		t.Fatal(err)
	}
	for _, mutated := range []string{strings.Replace(string(b), `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1), strings.Replace(string(b), `"scope":`, `"unexpected":0,"scope":`, 1), string(b) + "{}"} {
		if _, err = Decode([]byte(mutated)); err == nil {
			t.Fatal("open schema")
		}
	}
	r.Attempts.Requests = 1
	if _, err = Encode(r); err == nil {
		t.Fatal("selftest counted network")
	}
}
func TestOfficialPlannerBoundaries(t *testing.T) {
	req, limit, _ := cascollect.NextOfficial(context.Background(), nil)
	if req == nil || req.Target != "http://ipgw.neu.edu.cn/" || limit != cascollect.MaxDiscoveryBytes {
		t.Fatal("wrong first request")
	}
	p := []cascollect.Exchange{{Target: req.Target, Status: 200, Body: []byte("ac_id=123")}}
	req, _, _ = cascollect.NextOfficial(context.Background(), p)
	if req == nil || !strings.Contains(req.Target, "service=http%3A%2F%2Fipgw.neu.edu.cn%2Fsrun_portal_sso%3Fac_id%3D123") {
		t.Fatal("context lost")
	}
	p = append(p, cascollect.Exchange{Target: req.Target, Status: 200, Headers: []cascollect.Header{{Name: "content-type", Value: "text/html"}}, Body: []byte(`<script src="https://other.invalid/rsa.js"></script>`)})
	req, _, r := cascollect.NextOfficial(context.Background(), p)
	if req != nil || r.Collection != "blocked" {
		t.Fatal("external source planned")
	}
	r = cascollect.Replay(context.Background(), cascollect.Scenario{Scope: cascollect.Scope{Captive: "http://ipgw.neu.edu.cn/", Service: "http://ipgw.neu.edu.cn/srun_portal_sso", CAS: "https://pass.neu.edu.cn/tpass/login"}})
	if r.Reason != "invalid_scenario" || r.Simulated.Requests != 0 {
		t.Fatal("synthetic entry widened")
	}
}

func FuzzDNSPacket(f *testing.F) {
	f.Add([]byte{0})
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 4096 {
			return
		}
		name, _ := dnsmessage.NewName("example.invalid.")
		_, _ = dnsAnswer(b, 7, name)
	})
}

func FuzzAcquisitionReport(f *testing.F) {
	b, _ := Encode(blocked("anonymous-once", "topology_rejected"))
	f.Add(b)
	for _, unavailable := range []bool{false, true} {
		r := topologyReportFixture()
		if unavailable {
			r.Reason = "tls_chain_topology_unavailable"
			r.ChainTopology = &ChainTopology{Status: "unavailable", Reason: "byte_limit"}
		}
		seed, err := Encode(r)
		if err != nil {
			f.Fatal("D7 seed rejected")
		}
		f.Add(seed)
	}
	for _, reason := range []string{"tls_chain_relation_observed", "tls_chain_relation_unavailable"} {
		r := blocked("anonymous-once", reason)
		r.Attempts = cascollect.Attempts{Requests: 3, DNS: 3, Connects: 3}
		r.Operations.ExternalRequests, r.Operations.DNS = 3, 3
		r.ChainRelation = &ChainRelation{Status: "complete", CacheCoverage: "partial", Reason: "none", Facts: &RelationFacts{LeafNoParent: true}}
		if reason == "tls_chain_relation_unavailable" {
			r.ChainRelation.Status = "unavailable"
			r.ChainRelation.Reason = "signature_limit"
			r.ChainRelation.Facts = nil
		}
		seed, err := Encode(r)
		if err != nil {
			f.Fatal("D6 seed rejected")
		}
		f.Add(seed)
	}
	for _, reason := range []string{"cache_partial", "cache_partial_empty", "tls_chain_probe_cache_invalid"} {
		seed, err := Encode(blocked("local-cache-check", reason))
		if err != nil {
			f.Fatal("local cache seed rejected")
		}
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > cascollect.MaxReportBytes {
			return
		}
		if r, err := Decode(b); err == nil {
			out, err := Encode(r)
			if err != nil {
				t.Fatal("decode/encode mismatch")
			}
			if _, err = Decode(out); err != nil {
				t.Fatal("roundtrip")
			}
		}
	})
}
