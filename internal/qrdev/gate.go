package qrdev

import (
	"context"
	"github.com/UnbalancedCat/ipgw-meta/internal/labpath"
	"sync"
)

type GateFailure struct {
	Command     string          `json:"command"`
	Phase       string          `json:"phase"`
	Reason      string          `json:"reason"`
	QueryReason string          `json:"query_reason"`
	Changes     labpath.Changes `json:"changes"`
}

func oneOf(s string, values ...string) bool {
	for _, v := range values {
		if s == v {
			return true
		}
	}
	return false
}

func (f GateFailure) Valid() bool {
	if !oneOf(f.Command, "initial_check", "profile_show", "initial_status", "qr_login", "status_online", "logout", "final_status", "cleanup_logout", "cleanup_status", "final_check") || !oneOf(f.Phase, "before", "during", "after", "final") {
		return false
	}
	if (f.Command == "final_check") != (f.Phase == "final") || f.Command == "initial_check" && f.Phase != "before" {
		return false
	}
	switch f.Reason {
	case "binding_failed":
		return f.QueryReason == "none" && !f.Changes.Any()
	case "snapshot_changed":
		return f.QueryReason == "none" && f.Changes.Any()
	case "query_unavailable":
		return !f.Changes.Any() && oneOf(f.QueryReason, "unsupported", "launch_failed", "canceled", "deadline", "exit_nonzero", "stderr", "output_limit", "structure_rejected", "local_rule_rejected", "local_query_failed", "query_rejected")
	}
	return false
}

// GateTracker keeps no raw topology. Check and Reject are called serially by
// the joined guard; the mutex also protects readers while the monitor runs.
type GateTracker struct {
	mu      sync.Mutex
	verify  func() error
	inspect func(context.Context) labpath.Inspection
	pending labpath.Inspection
	first   *GateFailure
}

func NewGateTracker(verify func() error, inspect func(context.Context) labpath.Inspection) *GateTracker {
	return &GateTracker{verify: verify, inspect: inspect}
}

func (g *GateTracker) Check(ctx context.Context) bool {
	r := labpath.Inspection{Reason: "binding_failed", QueryReason: "none"}
	if g.verify != nil && g.verify() == nil {
		r = labpath.Inspection{Reason: "query_unavailable", QueryReason: "unsupported"}
		if g.inspect != nil {
			r = g.inspect(ctx)
		}
	}
	g.mu.Lock()
	g.pending = r
	g.mu.Unlock()
	return r.Reason == "none"
}

// Reject is not called for a monitor canceled by ordinary command completion.
func (g *GateTracker) Reject(command, phase string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.first != nil || g.pending.Reason == "none" {
		return
	}
	f := GateFailure{Command: command, Phase: phase, Reason: g.pending.Reason, QueryReason: g.pending.QueryReason, Changes: g.pending.Changes}
	if f.Valid() {
		g.first = &f
	}
}

func (g *GateTracker) At(ctx context.Context, command, phase string) bool {
	ok := g.Check(ctx)
	if !ok {
		g.Reject(command, phase)
	}
	return ok
}

func (g *GateTracker) Failure() *GateFailure {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.first == nil {
		return nil
	}
	f := *g.first
	return &f
}

// BindingRejected records a failure already obtained by the mandatory final
// file verification, without opening or hashing any file a second time.
func (g *GateTracker) BindingRejected(command, phase string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.first == nil {
		f := GateFailure{Command: command, Phase: phase, Reason: "binding_failed", QueryReason: "none"}
		if f.Valid() {
			g.first = &f
		}
	}
}
