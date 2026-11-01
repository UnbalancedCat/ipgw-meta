package ctlcheck

import (
	"context"
	"crypto/x509"
	"testing"
)

func memberFixture(t *testing.T, change func(int, []byte) []byte, verify verifier) (*MemberSession, *Membership, fixture) {
	t.Helper()
	f := synthetic(t)
	var blobs [2][]byte
	for i, name := range []string{"authroot_ctl", "disallowed_ctl"} {
		key, e := syntheticMemberKey(name, f.leaf.Raw)
		if e != nil {
			t.Fatal(e)
		}
		content := memberSyntheticContent(f, name, der(0x30, der(4, key), der(0x31, attribute("1.2.3.4", der(4, []byte("synthetic-policy"))))))
		if change != nil {
			content = change(i, content)
		}
		blobs[i], e = f.signWithAttributes(content, i == 1, false, nil)
		if e != nil {
			t.Fatal(e)
		}
	}
	s, r := prepareMembers(context.Background(), func(_ context.Context, i int) ([]byte, string) { return blobs[i], "" }, func(context.Context) (materials, string) { return f.mat, "" }, f.pins, f.now, verify)
	return s, r, f
}
func TestPreparedMembership(t *testing.T) {
	s, r, f := memberFixture(t, nil, pure)
	if s == nil || r != nil {
		t.Fatal("preparation")
	}
	for _, c := range []struct {
		peer [][]byte
		want string
	}{{[][]byte{f.leaf.Raw}, "any_match"}, {[][]byte{f.root.Raw}, "no_match"}, {[][]byte{f.root.Raw, f.leaf.Raw}, "any_match"}} {
		r = s.Compare(context.Background(), c.peer)
		if !ValidMembership(r) || r.State != "complete" {
			t.Fatal("comparison")
		}
		for _, source := range r.Sources {
			if source.Relation != c.want || source.Window != "unbounded" {
				t.Fatal("relation/time")
			}
			if c.want == "any_match" && source.EntryAttributes != "present_uninterpreted" {
				t.Fatal("policy erased")
			}
		}
	}
	for _, peer := range [][][]byte{nil, {nil}, {f.leaf.Raw, {1}}, {make([]byte, (64<<10)+1)}, make([][]byte, 17)} {
		r = s.Compare(context.Background(), peer)
		if !ValidMembership(r) || r.State != "unavailable" {
			t.Fatal("partial or oversized result")
		}
	}
	for at := 1; at < 20; at++ {
		r = s.Compare(&cancelAt{Context: context.Background(), at: at}, [][]byte{f.leaf.Raw, f.root.Raw})
		if !ValidMembership(r) {
			t.Fatal("cancel shape")
		}
		if r.State != "complete" {
			for _, source := range r.Sources {
				if source.Relation != "unknown" {
					t.Fatal("partial cancellation")
				}
			}
		}
	}
}
func TestPreparationMustAuthenticateWholeTable(t *testing.T) {
	mutations := []func(int, []byte) []byte{
		func(i int, b []byte) []byte {
			if i != 1 {
				return b
			}
			return identifierMutate(b, []int{3}, func(n node) []byte { return der(0x30, n.body, n.body) })
		},
		func(i int, b []byte) []byte {
			if i != 0 {
				return b
			}
			return identifierMutate(b, []int{0}, func(node) []byte { return der(0x30, encOID(oidDisallowedList)) })
		},
	}
	for _, change := range mutations {
		s, r, _ := memberFixture(t, change, pure)
		if s != nil || !ValidMembership(r) || r.Phase != "gate" {
			t.Fatal("invalid signed table accepted")
		}
	}
	s, r, _ := memberFixture(t, nil, func([]byte, *message, *x509.Certificate) bool { return false })
	if s != nil || !ValidMembership(r) {
		t.Fatal("native rejection bypass")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	s, r = prepareMembers(ctx, func(context.Context, int) ([]byte, string) { calls++; return nil, "" }, nil, pins{}, synthetic(t).now, pure)
	if s != nil || calls != 0 || !ValidMembership(r) || r.Reason != "canceled" {
		t.Fatal("canceled read")
	}
	for _, size := range []int{0, maxBlob + 1} {
		s, r = prepareMembers(context.Background(), func(context.Context, int) ([]byte, string) { return make([]byte, size), "" }, nil, pins{}, synthetic(t).now, pure)
		if s != nil || !ValidMembership(r) {
			t.Fatal("size gate")
		}
	}
}
func TestMembershipClosedStates(t *testing.T) {
	s, _, f := memberFixture(t, nil, pure)
	valid := s.Compare(context.Background(), [][]byte{f.leaf.Raw})
	for _, mutate := range []func(*Membership){
		func(r *Membership) { r.Authentication = "accepted" }, func(r *Membership) { r.Trust = "trusted" }, func(r *Membership) { r.Phase = "gate" }, func(r *Membership) { r.Sources[0].Relation = "unknown" }, func(r *Membership) { r.Sources[0].Window = "current" }, func(r *Membership) { r.Sources[0].Source = "other" },
	} {
		r := *valid
		r.Sources = append([]MemberSource(nil), valid.Sources...)
		mutate(&r)
		if ValidMembership(&r) {
			t.Fatal("contradiction accepted")
		}
	}
}
