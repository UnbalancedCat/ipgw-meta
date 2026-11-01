package casacquire

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"sort"
	"time"
)

var sourceNames = [7]string{"user_policy_root", "machine_policy_root", "user_policy_ca", "machine_policy_ca", "user_policy_disallowed", "machine_policy_disallowed", "machine_authroot_cache"}

func denialSource(i int) bool { return i == 4 || i == 5 }

type SourceFacts struct {
	ReachableDEROverlap          bool `json:"reachable_der_overlap"`
	TerminalSameSubjectSPKI      bool `json:"terminal_same_subject_spki"`
	TerminalDifferentSubjectSPKI bool `json:"terminal_different_subject_spki"`
	AdditionalSigningParent      bool `json:"additional_signing_parent"`
}
type SourceState struct {
	Source        string       `json:"source"`
	State         string       `json:"state"`
	Reason        string       `json:"reason"`
	Presence      string       `json:"presence"`
	ParseCoverage string       `json:"parse_coverage"`
	Facts         *SourceFacts `json:"facts"`
}
type SourceDifference struct {
	Phase   string        `json:"phase"`
	Status  string        `json:"status"`
	Reason  string        `json:"reason"`
	Sources []SourceState `json:"sources"`
}
type sourceSnapshot struct {
	states []SourceState
	certs  [7]map[[32]byte]*x509.Certificate
	denied [7]map[[32]byte]bool
}
type sourceReader func(context.Context, int, func([]byte) string) (string, string)

func sourceContext(ctx context.Context) string {
	if ctx.Err() == context.DeadlineExceeded {
		return "deadline"
	}
	if ctx.Err() != nil {
		return "canceled"
	}
	return ""
}
func sourceReadFailure(s string) bool {
	switch s {
	case "open_failed", "read_failed", "invalid_entry", "limit", "canceled", "deadline", "unsupported":
		return true
	}
	return false
}
func sourceFailure(s string) bool {
	if sourceReadFailure(s) {
		return true
	}
	switch s {
	case "input_invalid", "key_unsupported", "node_limit", "comparison_limit", "byte_limit", "signature_limit":
		return true
	}
	return false
}
func sourceReport(s *sourceSnapshot, phase, status, reason string) *SourceDifference {
	if s == nil {
		return &SourceDifference{Phase: phase, Status: status, Reason: reason}
	}
	r := &SourceDifference{Phase: phase, Status: status, Reason: reason, Sources: append([]SourceState(nil), s.states...)}
	for i := range r.Sources {
		r.Sources[i].Facts = nil
	}
	return r
}

// A reader receives only an index in the fixed allowlist. Callback consumes DER
// before the provider frees it; only copied certificates or denial hashes remain.
func readSources(ctx context.Context, read sourceReader) (*sourceSnapshot, *SourceDifference) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	s := &sourceSnapshot{states: make([]SourceState, 7)}
	for i, n := range sourceNames {
		s.states[i] = SourceState{n, "not_checked", "not_checked", "unknown", "unknown", nil}
	}
	total, totalBytes, material := 0, 0, 0
	fail := func(i int, reason string) (*sourceSnapshot, *SourceDifference) {
		if stop := sourceContext(ctx); stop != "" {
			reason = stop
		}
		if !sourceReadFailure(reason) {
			reason = "read_failed"
		}
		s.states[i].State, s.states[i].Reason = "unavailable", reason
		s.states[i].Presence, s.states[i].ParseCoverage = "unknown", "unknown"
		r := sourceReport(s, "gate", "unavailable", reason)
		s.certs = [7]map[[32]byte]*x509.Certificate{}
		s.denied = [7]map[[32]byte]bool{}
		return nil, r
	}
	for i := range sourceNames {
		if stop := sourceContext(ctx); stop != "" {
			return fail(i, stop)
		}
		count, parsed, bad := 0, 0, 0
		s.certs[i] = map[[32]byte]*x509.Certificate{}
		s.denied[i] = map[[32]byte]bool{}
		callbackFailure := ""
		add := func(der []byte) string {
			if callbackFailure != "" {
				return callbackFailure
			}
			stop := func(reason string) string { callbackFailure = reason; return reason }
			if reason := sourceContext(ctx); reason != "" {
				return stop(reason)
			}
			if len(der) == 0 {
				return stop("invalid_entry")
			}
			if len(der) > 64<<10 || count >= 512 || total >= 2048 || len(der) > (16<<20)-totalBytes {
				return stop("limit")
			}
			count++
			total++
			totalBytes += len(der)
			id := sha256.Sum256(der)
			if denialSource(i) {
				s.denied[i][id] = true
				return ""
			}
			cert, err := x509.ParseCertificate(bytes.Clone(der))
			if err != nil {
				bad++
				return ""
			}
			parsed++
			s.certs[i][id] = cert
			return sourceContext(ctx)
		}
		if read == nil {
			return fail(i, "read_failed")
		}
		state, reason := read(ctx, i, add)
		if callbackFailure != "" {
			reason = callbackFailure
		}
		if stop := sourceContext(ctx); stop != "" {
			reason = stop
		}
		if reason != "" {
			return fail(i, reason)
		}
		if (state != "complete" && state != "absent") || state == "absent" && count != 0 {
			return fail(i, "read_failed")
		}
		presence, coverage := "empty", "empty"
		if count > 0 {
			presence = "present"
			switch {
			case parsed == 0:
				coverage = "none"
			case bad > 0:
				coverage = "partial"
			default:
				coverage = "complete"
			}
		}
		if denialSource(i) {
			coverage = "not_applicable"
			material += len(s.denied[i])
		} else {
			material += len(s.certs[i])
		}
		s.states[i] = SourceState{sourceNames[i], state, "none", presence, coverage, nil}
	}
	if stop := sourceContext(ctx); stop != "" {
		return fail(6, stop)
	}
	if material == 0 {
		return nil, sourceReport(s, "gate", "no_material", "no_comparable_material")
	}
	return s, nil
}

type differenceGraph struct {
	graph     relationGraph
	terminals map[[32]byte]bool
}
type differenceLimits struct{ comparisons, bytes, signatures int }

var defaultDifferenceLimits = differenceLimits{65536, 32 << 20, 100}

func compareSources(ctx context.Context, s *sourceSnapshot, g *differenceGraph, limits differenceLimits) *SourceDifference {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	r := sourceReport(s, "comparison", "unavailable", "input_invalid")
	fail := func(reason string) *SourceDifference {
		if stop := sourceContext(ctx); stop != "" {
			reason = stop
		}
		r.Status, r.Reason = "unavailable", reason
		for i := range r.Sources {
			r.Sources[i].Facts = nil
		}
		return r
	}
	if stop := sourceContext(ctx); stop != "" {
		return fail(stop)
	}
	if s == nil || len(s.states) != 7 || g == nil || len(g.graph.nodes) == 0 || len(g.terminals) == 0 || limits.comparisons < 1 || limits.comparisons > 65536 || limits.bytes < 1 || limits.bytes > 32<<20 || limits.signatures < 0 || limits.signatures > 100 {
		return fail("input_invalid")
	}
	if len(g.graph.nodes) > 128 {
		return fail("node_limit")
	}
	comparisons, used, signatures := 0, 0, 0
	compare := func() bool {
		if comparisons >= limits.comparisons {
			return false
		}
		comparisons++
		return true
	}
	charge := func(n int) bool {
		if n < 0 || n > limits.bytes-used {
			return false
		}
		used += n
		return true
	}
	nodes := map[[32]byte]bool{}
	for _, id := range g.graph.nodes {
		c := g.graph.certs[id]
		if c == nil || len(c.Raw) == 0 || len(c.Raw) > 64<<10 || nodes[id] {
			return fail("input_invalid")
		}
		if !charge(len(c.Raw)) {
			return fail("byte_limit")
		}
		if sha256.Sum256(c.Raw) != id {
			return fail("input_invalid")
		}
		nodes[id] = true
	}
	for id, terminal := range g.terminals {
		if !nodes[id] || !terminal {
			return fail("input_invalid")
		}
	}
	for i := range sourceNames {
		state := r.Sources[i]
		if state.State != "complete" && state.State != "absent" {
			return fail("input_invalid")
		}
		facts := &SourceFacts{}
		r.Sources[i].Facts = facts
		if denialSource(i) {
			for _, id := range g.graph.nodes {
				if sourceContext(ctx) != "" {
					return fail("canceled")
				}
				if !compare() {
					return fail("comparison_limit")
				}
				facts.ReachableDEROverlap = facts.ReachableDEROverlap || s.denied[i][id]
			}
			continue
		}
		ids := make([][32]byte, 0, len(s.certs[i]))
		for id := range s.certs[i] {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(a, b int) bool { return bytes.Compare(ids[a][:], ids[b][:]) < 0 })
		for _, id := range g.graph.nodes {
			node := g.graph.certs[id]
			for _, candidateID := range ids {
				if sourceContext(ctx) != "" {
					return fail("canceled")
				}
				if !compare() {
					return fail("comparison_limit")
				}
				c := s.certs[i][candidateID]
				if c == nil {
					return fail("input_invalid")
				}
				if len(c.Raw) == 0 || len(c.Raw) > 64<<10 {
					return fail("input_invalid")
				}
				if !charge(len(c.Raw) + len(node.RawSubject) + len(node.RawIssuer) + len(node.RawSubjectPublicKeyInfo) + len(c.RawSubject) + len(c.RawSubjectPublicKeyInfo)) {
					return fail("byte_limit")
				}
				if sha256.Sum256(c.Raw) != candidateID {
					return fail("input_invalid")
				}
				if id == candidateID {
					facts.ReachableDEROverlap = true
				}
				if g.terminals[id] {
					if !compare() {
						return fail("comparison_limit")
					}
					if bytes.Equal(node.RawSubjectPublicKeyInfo, c.RawSubjectPublicKeyInfo) {
						if !compare() {
							return fail("comparison_limit")
						}
						if bytes.Equal(node.RawSubject, c.RawSubject) {
							facts.TerminalSameSubjectSPKI = true
						} else {
							facts.TerminalDifferentSubjectSPKI = true
						}
					}
				}
				if g.graph.certs[candidateID] != nil {
					continue
				}
				if !compare() {
					return fail("comparison_limit")
				}
				if !bytes.Equal(node.RawIssuer, c.RawSubject) {
					continue
				}
				if !boundedRelationKey(c) {
					return fail("key_unsupported")
				}
				if signatures >= limits.signatures {
					return fail("signature_limit")
				}
				signatures++
				if !charge(len(node.RawTBSCertificate) + len(node.Signature) + len(c.RawSubjectPublicKeyInfo)) {
					return fail("byte_limit")
				}
				if node.CheckSignatureFrom(c) == nil {
					facts.AdditionalSigningParent = true
				}
				if sourceContext(ctx) != "" {
					return fail("canceled")
				}
			}
		}
	}
	if stop := sourceContext(ctx); stop != "" {
		return fail(stop)
	}
	r.Status, r.Reason = "complete", "none"
	return r
}
