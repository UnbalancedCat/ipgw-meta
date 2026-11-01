package casacquire

import (
	"context"
	"encoding/json"
	"github.com/UnbalancedCat/ipgw-meta/internal/ctlcheck"
	"strings"
	"testing"
)

func TestMemberGateStopsBeforeSourcesAndNetwork(t *testing.T) {
	w, dns, requests := fixtureWire(t, false, "")
	r := acquireWithMembers(context.Background(), w, func(context.Context) (*ctlcheck.MemberSession, *ctlcheck.Membership) {
		return nil, ctlcheck.MemberUnavailable("gate", "unavailable", "preparation_failed")
	}, func(context.Context, int, func([]byte) string) (string, string) {
		t.Fatal("source read after gate rejection")
		return "", ""
	})
	if dns.Load() != 0 || requests.Load() != 0 || r.CTLMembership == nil {
		t.Fatal("gate order")
	}
	b, e := Encode(r)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Decode(b); e != nil {
		t.Fatal(e)
	}
}

func memberReportFixture() Report {
	r := sourceReportFixture()
	r.ContextConstructed = true
	r.CTLMembership = &ctlcheck.Membership{Phase: "comparison", State: "complete", Reason: "none", Authentication: "blocked", Trust: "not_established"}
	for _, name := range []string{"authroot_ctl", "disallowed_ctl"} {
		r.CTLMembership.Sources = append(r.CTLMembership.Sources, ctlcheck.MemberSource{Source: name, State: "complete", Reason: "none", Relation: "no_match", EntryAttributes: "not_applicable", Extensions: "absent", Window: "unbounded"})
	}
	return r
}
func TestMembershipReportClosedAndHistorical(t *testing.T) {
	r := memberReportFixture()
	b, e := Encode(r)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Decode(b); e != nil {
		t.Fatal(e)
	}
	for _, mutate := range []func(map[string]any){
		func(m map[string]any) { m["source_difference"] = nil },
		func(m map[string]any) { m["ctl_membership"] = nil },
		func(m map[string]any) { m["ctl_membership"].(map[string]any)["trust"] = "trusted" },
		func(m map[string]any) { delete(m["ctl_membership"].(map[string]any), "reason") },
		func(m map[string]any) { m["ctl_membership"].(map[string]any)["raw"] = "synthetic-canary" },
		func(m map[string]any) { m["context_constructed"] = false },
		func(m map[string]any) { m["collection"] = "complete" },
	} {
		var m map[string]any
		json.Unmarshal(b, &m)
		mutate(m)
		changed, _ := json.Marshal(m)
		if _, e = Decode(changed); e == nil {
			t.Fatal("open/contradictory membership")
		}
	}
	if _, e = Decode([]byte(strings.Replace(string(b), `"phase":`, `"phase":"comparison","phase":`, 1))); e == nil {
		t.Fatal("duplicate extension")
	}
	r.CTLMembership = nil
	old, e := Encode(r)
	if e != nil {
		t.Fatal(e)
	}
	decoded, e := Decode(old)
	if e != nil || decoded.CTLMembership != nil {
		t.Fatal("old meaning altered")
	}
	if strings.Contains(string(old), "ctl_membership") {
		t.Fatal("old report expanded")
	}
	w := &wire{memberSession: &ctlcheck.MemberSession{}}
	attachMembership(&r, w)
	if r.CTLMembership.State != "not_run" || r.CTLMembership.Sources[0].Relation != "unknown" {
		t.Fatal("not-run became no-match")
	}
}

func FuzzMembershipReport(f *testing.F) {
	b, e := Encode(memberReportFixture())
	if e != nil {
		f.Fatal(e)
	}
	f.Add(b)
	r := blocked("anonymous-once", "ctl_membership_unavailable")
	r.CTLMembership = ctlcheck.MemberUnavailable("gate", "unavailable", "preparation_failed")
	b, _ = Encode(r)
	f.Add(b)
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 1<<20 {
			return
		}
		r, e := Decode(b)
		if e == nil {
			encoded, e := Encode(r)
			if e != nil {
				t.Fatal("valid decode not encodable")
			}
			if _, e = Decode(encoded); e != nil {
				t.Fatal("roundtrip")
			}
		}
	})
}
