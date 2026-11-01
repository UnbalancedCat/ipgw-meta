package trustaudit

import (
	"encoding/json"
	"errors"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
)

var ErrReport = errors.New("invalid trust audit report")

func failureReason(s string) bool {
	switch s {
	case "open_failed", "read_failed", "invalid_entry", "limit", "canceled", "deadline", "unsupported":
		return true
	}
	return false
}
func validStore(s Store, raw bool) bool {
	location := func(v string) bool { return v == "complete" || v == "absent" }
	if s.State == "unavailable" {
		if !failureReason(s.Reason) || s.Presence != "unknown" || s.ParseCoverage != "unknown" {
			return false
		}
		if s.CurrentUser == "not_checked" {
			return s.LocalMachine == "not_checked" && (s.Reason == "canceled" || s.Reason == "deadline")
		}
		if s.CurrentUser == "unavailable" {
			return s.LocalMachine == "not_checked"
		}
		return location(s.CurrentUser) && (s.LocalMachine == "unavailable" || s.LocalMachine == "not_checked" && (s.Reason == "canceled" || s.Reason == "deadline"))
	}
	if s.State != "complete" || s.Reason != "none" || !location(s.CurrentUser) || !location(s.LocalMachine) || (s.Presence != "present" && s.Presence != "empty") {
		return false
	}
	if s.CurrentUser == "absent" && s.LocalMachine == "absent" && s.Presence != "empty" {
		return false
	}
	if raw {
		return s.ParseCoverage == "not_applicable"
	}
	if s.Presence == "empty" {
		return s.ParseCoverage == "empty"
	}
	return s.ParseCoverage == "none" || s.ParseCoverage == "partial" || s.ParseCoverage == "complete"
}
func Valid(r Report) bool {
	if r.SchemaVersion != 1 || r.ToolKind != "trust-audit-dev" || (r.Mode != "self-test" && r.Mode != "local-trust-audit") || r.Operations != (Operations{}) {
		return false
	}
	if !validStore(r.Root, false) || !validStore(r.CA, false) || !validStore(r.Disallowed, true) {
		return false
	}
	all := r.Root.State == "complete" && r.CA.State == "complete" && r.Disallowed.State == "complete"
	if (r.State == "complete") != all || r.State != "complete" && r.State != "unavailable" {
		return false
	}
	c := r.Comparison
	canceled := false
	for _, s := range []Store{r.Root, r.Disallowed, r.CA} {
		canceled = canceled || s.Reason == "canceled" || s.Reason == "deadline"
	}
	if r.Root.State != "complete" || r.Disallowed.State != "complete" || canceled {
		return c == (Comparison{"unavailable", "unknown", "unknown"})
	}
	if c.State != "complete" || (c.ParsedRootDeniedOverlap != "present" && c.ParsedRootDeniedOverlap != "none") || (c.RemainingParsedRoots != "present" && c.RemainingParsedRoots != "empty") {
		return false
	}
	if r.Root.ParseCoverage == "none" || r.Root.ParseCoverage == "empty" {
		return c.ParsedRootDeniedOverlap == "none" && c.RemainingParsedRoots == "empty"
	}
	if r.Disallowed.Presence == "empty" && c.ParsedRootDeniedOverlap != "none" {
		return false
	}
	return c.ParsedRootDeniedOverlap == "present" || c.RemainingParsedRoots == "present"
}
func Encode(r Report) ([]byte, error) {
	if !Valid(r) {
		return nil, ErrReport
	}
	b, err := json.Marshal(r)
	if err != nil || len(b) > 16383 {
		return nil, ErrReport
	}
	return append(b, '\n'), nil
}
func Decode(b []byte) (Report, error) {
	var r Report
	if casprovenance.DecodeClosedJSON(b, &r, 16384) != nil || !Valid(r) {
		return Report{}, ErrReport
	}
	return r, nil
}
