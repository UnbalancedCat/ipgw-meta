package ctlcheck

import (
	"context"
	"crypto/x509"
	"os"
	"time"
)

// Membership is a diagnostic relation, never a TLS acceptance decision.
type Membership struct {
	Phase          string         `json:"phase"`
	State          string         `json:"state"`
	Reason         string         `json:"reason"`
	Authentication string         `json:"authentication"`
	Trust          string         `json:"trust"`
	Sources        []MemberSource `json:"sources"`
}
type MemberSource struct {
	Source          string `json:"source"`
	State           string `json:"state"`
	Reason          string `json:"reason"`
	Relation        string `json:"relation"`
	EntryAttributes string `json:"entry_attributes"`
	Extensions      string `json:"extensions"`
	Window          string `json:"window"`
}
type memberIndex struct {
	entries            map[string]bool
	extensions, window string
}

// MemberSession can only be constructed after the fixed signed observations.
// It retains identifiers and coverage, not raw CTLs or signing certificates.
type MemberSession struct{ indexes [2]memberIndex }

func MemberUnavailable(phase, state, reason string) *Membership {
	r := &Membership{phase, state, reason, "blocked", "not_established", nil}
	for _, name := range []string{"authroot_ctl", "disallowed_ctl"} {
		r.Sources = append(r.Sources, MemberSource{name, "unavailable", reason, "unknown", "unknown", "unknown", "unknown"})
	}
	return r
}
func ValidMembership(r *Membership) bool {
	if r == nil || !one(r.Phase, "gate", "comparison") || r.Authentication != "blocked" || r.Trust != "not_established" || len(r.Sources) != 2 {
		return false
	}
	if r.State == "complete" {
		if r.Phase != "comparison" || r.Reason != "none" {
			return false
		}
	} else {
		if r.State == "not_run" {
			if r.Phase != "comparison" || r.Reason != "baseline_not_observed" {
				return false
			}
		} else if r.State != "unavailable" || !one(r.Reason, "preparation_failed", "protection_failed", "canceled", "deadline", "peer_unsupported", "peer_limit") {
			return false
		}
		if r.Phase == "gate" && one(r.Reason, "peer_unsupported", "peer_limit") {
			return false
		}
		if r.Phase == "comparison" && one(r.Reason, "preparation_failed", "protection_failed") {
			return false
		}
	}
	for i, s := range r.Sources {
		if s.Source != []string{"authroot_ctl", "disallowed_ctl"}[i] {
			return false
		}
		if r.State != "complete" {
			if s.State != "unavailable" || s.Reason != r.Reason || s.Relation != "unknown" || s.EntryAttributes != "unknown" || s.Extensions != "unknown" || s.Window != "unknown" {
				return false
			}
		} else {
			if s.State != "complete" || s.Reason != "none" || !one(s.Relation, "any_match", "no_match") || !one(s.Extensions, "absent", "present_uninterpreted", "critical_present_uninterpreted") || !one(s.Window, "within", "expired", "future", "unbounded", "invalid") {
				return false
			}
			if s.Relation == "no_match" {
				if s.EntryAttributes != "not_applicable" {
					return false
				}
			} else if !one(s.EntryAttributes, "absent", "present_uninterpreted") {
				return false
			}
		}
	}
	return true
}

func prepareMembers(ctx context.Context, read blobReader, load materialReader, p pins, now time.Time, verify verifier) (session *MemberSession, failure *Membership) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	fail := func() (*MemberSession, *Membership) {
		reason := stopped(ctx)
		if reason == "" {
			reason = "preparation_failed"
		}
		return nil, MemberUnavailable("gate", "unavailable", reason)
	}
	if stopped(ctx) != "" {
		return fail()
	}
	var blobs [2][]byte
	for i := range blobs {
		b, reason := read(ctx, i)
		if reason != "" || len(b) == 0 || len(b) > maxBlob || stopped(ctx) != "" {
			return fail()
		}
		blobs[i] = b
	}
	mat, reason := load(ctx)
	if reason != "" || stopped(ctx) != "" {
		return fail()
	}
	out := &MemberSession{}
	for i, name := range []string{"authroot_ctl", "disallowed_ctl"} {
		observation := observe(ctx, name, blobs[i], mat, p, now, verify)
		if observation.State != "complete" || (i == 0 && (observation.Usage != "root_list" || observation.MemberAlgorithm != "sha1")) || (i == 1 && (observation.Usage != "disallowed_list" || observation.MemberAlgorithm != "signature_hash_property")) {
			return fail()
		}
		m, reason := parseMessage(ctx, blobs[i])
		if reason != "" {
			return fail()
		}
		entries, e := memberEntries(m.content)
		if e != nil {
			return fail()
		}
		index := memberIndex{map[string]bool{}, observation.Extensions, observation.Window}
		for _, entry := range entries {
			if stopped(ctx) != "" {
				return fail()
			}
			pair, e := children(entry, 0x30, 2)
			if e != nil || len(pair) == 0 {
				return fail()
			}
			key := string(pair[0].body)
			if _, exists := index.entries[key]; exists {
				return fail()
			}
			index.entries[key] = len(pair) == 2 && len(pair[1].body) > 0
		}
		out.indexes[i] = index
	}
	if stopped(ctx) != "" {
		return fail()
	}
	return out, nil
}

// PrepareLocalMembers is called only by the confirmed Windows acquisition
// path, after topology subprocesses finish and before any network attempt.
func PrepareLocalMembers(ctx context.Context) (*MemberSession, *Membership) {
	return prepareProtectedMembers(ctx, protectProcess, func() (*MemberSession, *Membership) {
		return prepareMembers(ctx, localBlob, localMaterials, officialPins(), time.Now(), verifyNative)
	}, 15*time.Second)
}

func prepareProtectedMembers(ctx context.Context, protect func() bool, work func() (*MemberSession, *Membership), hardLimit time.Duration) (*MemberSession, *Membership) {
	if stopped(ctx) != "" {
		return nil, MemberUnavailable("gate", "unavailable", stopped(ctx))
	}
	if !protect() {
		return nil, MemberUnavailable("gate", "unavailable", "protection_failed")
	}
	// Native parsing cannot be preempted by a Go context. Kill the dedicated
	// process without a partial receipt if the preparation phase hangs.
	timer := time.AfterFunc(hardLimit, func() { os.Exit(4) })
	defer timer.Stop()
	return work()
}

func (s *MemberSession) Compare(ctx context.Context, peer [][]byte) (out *Membership) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	defer func() {
		if reason := stopped(ctx); reason != "" {
			out = MemberUnavailable("comparison", "unavailable", reason)
		}
	}()
	fail := func(reason string) *Membership { return MemberUnavailable("comparison", "unavailable", reason) }
	if s == nil || s.indexes[0].entries == nil || s.indexes[1].entries == nil {
		return fail("peer_unsupported")
	}
	if len(peer) == 0 {
		return fail("peer_unsupported")
	}
	if len(peer) > 16 {
		return fail("peer_limit")
	}
	remaining := 512 << 10
	for _, b := range peer {
		if len(b) > 64<<10 || len(b) > remaining {
			return fail("peer_limit")
		}
		remaining -= len(b)
	}
	r := &Membership{"comparison", "complete", "none", "blocked", "not_established", nil}
	for i, name := range []string{"authroot_ctl", "disallowed_ctl"} {
		index := s.indexes[i]
		result := MemberSource{name, "complete", "none", "no_match", "not_applicable", index.extensions, index.window}
		for _, b := range peer {
			if stopped(ctx) != "" {
				return fail(stopped(ctx))
			}
			key, e := syntheticMemberKey(name, b)
			if e != nil {
				return fail("peer_unsupported")
			}
			if attributes, found := index.entries[string(key)]; found {
				if result.Relation == "no_match" {
					result.EntryAttributes = "absent"
				}
				result.Relation = "any_match"
				if attributes {
					result.EntryAttributes = "present_uninterpreted"
				}
			}
		}
		r.Sources = append(r.Sources, result)
	}
	return r
}

// MembershipSelfTest never calls local readers or takes external input.
func MembershipSelfTest(ctx context.Context) bool {
	return membershipSelfTest(ctx, func(_ []byte, m *message, c *x509.Certificate) bool { return verifyGo(m, c) })
}
func membershipSelfTest(ctx context.Context, verify verifier) bool {
	f, e := generateFixture()
	if e != nil {
		return false
	}
	var blobs [2][]byte
	for i, name := range []string{"authroot_ctl", "disallowed_ctl"} {
		key, e := syntheticMemberKey(name, f.leaf.Raw)
		if e != nil {
			return false
		}
		blobs[i], e = f.signWithAttributes(memberSyntheticContent(f, name, der(0x30, der(4, key))), i == 1, false, nil)
		if e != nil {
			return false
		}
	}
	session, failure := prepareMembers(ctx, func(_ context.Context, i int) ([]byte, string) { return blobs[i], "" }, func(context.Context) (materials, string) { return f.mat, "" }, f.pins, f.now, verify)
	if failure != nil {
		return false
	}
	r := session.Compare(ctx, [][]byte{f.leaf.Raw})
	return ValidMembership(r) && r.State == "complete" && r.Sources[0].Relation == "any_match" && r.Sources[1].Relation == "any_match"
}
