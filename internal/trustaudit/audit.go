// Package trustaudit observes fixed local registry sources, never trust or TLS.
package trustaudit

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"errors"
	"time"
)

const maxEntries = 512
const maxDER = 64 << 10
const maxBytes = 8 << 20

type Store struct {
	State         string `json:"state"`
	Reason        string `json:"reason"`
	CurrentUser   string `json:"current_user"`
	LocalMachine  string `json:"local_machine"`
	Presence      string `json:"presence"`
	ParseCoverage string `json:"parse_coverage"`
}
type Comparison struct {
	State                   string `json:"state"`
	ParsedRootDeniedOverlap string `json:"parsed_root_denied_overlap"`
	RemainingParsedRoots    string `json:"remaining_parsed_roots"`
}
type Operations struct {
	NetworkRequests   int `json:"network_requests"`
	DNS               int `json:"dns"`
	Connects          int `json:"connects"`
	ProfileReads      int `json:"profile_reads"`
	CredentialReads   int `json:"credential_reads"`
	CertificateWrites int `json:"certificate_writes"`
}
type Report struct {
	SchemaVersion int        `json:"schema_version"`
	ToolKind      string     `json:"tool_kind"`
	Mode          string     `json:"mode"`
	State         string     `json:"state"`
	Root          Store      `json:"root"`
	Disallowed    Store      `json:"disallowed"`
	CA            Store      `json:"ca"`
	Comparison    Comparison `json:"comparison"`
	Operations    Operations `json:"operations"`
}
type accumulator struct {
	count, bytes, parsed, bad int
	raw                       bool
	hashes                    map[[32]byte]bool
}

func stopped(ctx context.Context) string {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "deadline"
	}
	if ctx.Err() != nil {
		return "canceled"
	}
	return ""
}
func (a *accumulator) add(ctx context.Context, der []byte) string {
	if s := stopped(ctx); s != "" {
		return s
	}
	if len(der) == 0 {
		return "invalid_entry"
	}
	if len(der) > maxDER || a.count >= maxEntries || len(der) > maxBytes-a.bytes {
		return "limit"
	}
	a.count++
	a.bytes += len(der)
	// Disallowed retains bounded raw rejection identities without parsing.
	if !a.raw {
		if _, err := x509.ParseCertificate(der); err != nil {
			a.bad++
			return stopped(ctx)
		}
	}
	a.parsed++
	if a.hashes == nil {
		a.hashes = map[[32]byte]bool{}
	}
	a.hashes[sha256.Sum256(der)] = true
	return stopped(ctx)
}

type reader func(context.Context, string, int, *accumulator) (string, string)

func unavailableStore(reason string) Store {
	return Store{"unavailable", reason, "not_checked", "not_checked", "unknown", "unknown"}
}
func readKind(ctx context.Context, name string, read reader) (Store, map[[32]byte]bool) {
	s := unavailableStore("none")
	a := &accumulator{raw: name == "Disallowed"}
	for location := 0; location < 2; location++ {
		if reason := stopped(ctx); reason != "" {
			s.Reason = reason
			return s, nil
		}
		before := a.count
		status, reason := read(ctx, name, location, a)
		if canceled := stopped(ctx); canceled != "" {
			reason = canceled
		}
		if reason == "" && ((status != "complete" && status != "absent") || status == "absent" && a.count != before) {
			reason = "read_failed"
		}
		if reason != "" {
			if !failureReason(reason) {
				reason = "read_failed"
			}
			status = "unavailable"
		}
		if location == 0 {
			s.CurrentUser = status
		} else {
			s.LocalMachine = status
		}
		if reason != "" {
			s.Reason = reason
			return s, nil
		}
	}
	s.State, s.Reason, s.Presence = "complete", "none", "empty"
	if a.count > 0 {
		s.Presence = "present"
	}
	s.ParseCoverage = "not_applicable"
	if !a.raw {
		switch {
		case a.count == 0:
			s.ParseCoverage = "empty"
		case a.parsed == 0:
			s.ParseCoverage = "none"
		case a.bad > 0:
			s.ParseCoverage = "partial"
		default:
			s.ParseCoverage = "complete"
		}
	}
	return s, a.hashes
}
func audit(ctx context.Context, mode string, read reader) Report {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	r := Report{SchemaVersion: 1, ToolKind: "trust-audit-dev", Mode: mode, State: "unavailable", Comparison: Comparison{"unavailable", "unknown", "unknown"}}
	var roots, denied map[[32]byte]bool
	r.Disallowed, denied = readKind(ctx, "Disallowed", read)
	r.Root, roots = readKind(ctx, "ROOT", read)
	r.CA, _ = readKind(ctx, "CA", read)
	if r.Root.State == "complete" && r.Disallowed.State == "complete" && stopped(ctx) == "" {
		r.Comparison = Comparison{"complete", "none", "empty"}
		for id := range roots {
			if stopped(ctx) != "" {
				r.Comparison = Comparison{"unavailable", "unknown", "unknown"}
				break
			}
			if denied[id] {
				r.Comparison.ParsedRootDeniedOverlap = "present"
			} else {
				r.Comparison.RemainingParsedRoots = "present"
			}
		}
	}
	if reason := stopped(ctx); reason != "" {
		r.CA = unavailableStore(reason)
		r.Comparison = Comparison{"unavailable", "unknown", "unknown"}
	}
	if r.Root.State == "complete" && r.Disallowed.State == "complete" && r.CA.State == "complete" {
		r.State = "complete"
	}
	return r
}
func Run(ctx context.Context) Report { return audit(ctx, "local-trust-audit", localRead) }
