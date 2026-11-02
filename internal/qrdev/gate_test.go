package qrdev

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/UnbalancedCat/ipgw-meta/internal/labpath"
	"github.com/UnbalancedCat/ipgw-meta/internal/livegate"
	"strings"
	"testing"
)

func TestGateFirstFailureAndReasonMatrix(t *testing.T) {
	for _, reason := range []string{"unsupported", "launch_failed", "canceled", "deadline", "exit_nonzero", "stderr", "output_limit", "structure_rejected", "local_rule_rejected", "local_query_failed", "query_rejected"} {
		t.Run(reason, func(t *testing.T) {
			g := NewGateTracker(func() error { return nil }, func(context.Context) labpath.Inspection {
				return labpath.Inspection{Reason: "query_unavailable", QueryReason: reason}
			})
			if g.At(context.Background(), "logout", "before") {
				t.Fatal("accepted")
			}
			first := g.Failure()
			g.verify = func() error { return errors.New("PRIVATE_CANARY") }
			g.At(context.Background(), "final_check", "final")
			if first == nil || !first.Valid() || *first != *g.Failure() {
				t.Fatal("first lost")
			}
			raw, _ := json.Marshal(first)
			if strings.Contains(string(raw), "PRIVATE_CANARY") {
				t.Fatal("leak")
			}
		})
	}
	g := NewGateTracker(func() error { return errors.New("PRIVATE_CANARY") }, nil)
	g.At(context.Background(), "initial_check", "before")
	if g.Failure().Reason != "binding_failed" {
		t.Fatal("binding")
	}
}

func TestGateObservedRunPreservesActionAndPhase(t *testing.T) {
	for _, phase := range []string{"before", "after"} {
		checks := 0
		g := NewGateTracker(func() error { return nil }, func(context.Context) labpath.Inspection {
			checks++
			if checks == 7 && phase == "before" || checks == 8 && phase == "after" {
				return labpath.Inspection{Reason: "snapshot_changed", QueryReason: "none", Changes: labpath.Changes{Routes: true}}
			}
			return labpath.Inspection{Reason: "none", QueryReason: "none"}
		})
		e := &syntheticExecutor{}
		r := RunObserved(context.Background(), e, g.Check, func(a livegate.CommandAction, p string) { g.Reject(string(a), p) })
		f := g.Failure()
		if r.Validate() != nil || f == nil || f.Command != "logout" || f.Phase != phase || !f.Valid() {
			t.Fatalf("wrong failure %#v", f)
		}
		g.At(context.Background(), "final_check", "final")
		if *f != *g.Failure() {
			t.Fatal("overwritten")
		}
	}
}

func TestGateFailureRejectsContradictions(t *testing.T) {
	f := GateFailure{Command: "logout", Phase: "before", Reason: "snapshot_changed", QueryReason: "none", Changes: labpath.Changes{DNS: true}}
	for _, mutate := range []func(*GateFailure){func(f *GateFailure) { f.Command = "secret" }, func(f *GateFailure) { f.Phase = "final" }, func(f *GateFailure) { f.Reason = "secret" }, func(f *GateFailure) { f.QueryReason = "secret" }, func(f *GateFailure) { f.Changes = labpath.Changes{} }, func(f *GateFailure) { f.Reason = "binding_failed" }} {
		v := f
		mutate(&v)
		if v.Valid() {
			t.Fatal("contradiction accepted")
		}
	}
}
