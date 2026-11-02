package qrdev

import (
	"context"
	"encoding/json"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
	"github.com/UnbalancedCat/ipgw-meta/internal/livegate"
	"testing"
	"time"
)

type executorFunc func(context.Context, livegate.CommandAction) livegate.Observation

func (f executorFunc) Execute(c context.Context, a livegate.CommandAction) livegate.Observation {
	return f(c, a)
}
func TestQRSelfTest(t *testing.T) {
	if !SelfTest() {
		t.Fatal("self test failed")
	}
}
func TestQRTraceDistinctFailures(t *testing.T) {
	for _, a := range []livegate.CommandAction{livegate.ActionQRLogin, livegate.ActionStatusOnline, livegate.ActionLogout, livegate.ActionFinalStatus} {
		e := &syntheticExecutor{fail: a}
		r := Run(context.Background(), e, func(context.Context) bool { return true })
		if r.Validate() != nil || r.State == "pass" {
			t.Fatalf("invalid result %s: %#v", a, r)
		}
		found := false
		for _, c := range r.Commands {
			if c.Action == a {
				found = c.Step != nil && c.Step.Result != livegate.ResultPass
			}
		}
		if !found {
			t.Fatal("failure not attributed")
		}
		if a == livegate.ActionQRLogin && len(r.Commands) != 2 || a == livegate.ActionStatusOnline && len(r.Commands) != 3 {
			t.Fatal("failure invoked extra commands")
		}
	}
}
func TestQRCancelAndTopologyStopCleanup(t *testing.T) {
	for _, mode := range []string{"cancel", "topology", "wrong_identity", "fault"} {
		ctx, cancel := context.WithCancel(context.Background())
		valid := true
		e := &syntheticExecutor{}
		wrapped := executorFunc(func(c context.Context, a livegate.CommandAction) livegate.Observation {
			if a == livegate.ActionQRLogin {
				d, ok := c.Deadline()
				if !ok || time.Until(d) > 300*time.Second || time.Until(d) < 290*time.Second {
					t.Fatal("QR deadline")
				}
			}
			o := e.Execute(c, a)
			if a == livegate.ActionStatusOnline && mode == "wrong_identity" {
				o.IdentityMatches = false
			}
			if a == livegate.ActionLogout {
				switch mode {
				case "cancel":
					cancel()
				case "topology":
					valid = false
				case "fault":
					o.RunnerFault = true
				}
			}
			return o
		})
		r := Run(ctx, wrapped, func(context.Context) bool { return valid })
		cancel()
		if r.State == "pass" || r.Validate() != nil {
			t.Fatalf("unexpected %s: %#v", mode, r)
		}
		for _, a := range e.actions {
			if a == livegate.ActionCleanupLogout || a == livegate.ActionCleanupStatus {
				t.Fatal("cleanup after lost permission")
			}
		}
	}
}
func TestQRReportClosedAndContradictions(t *testing.T) {
	r := Run(context.Background(), &syntheticExecutor{}, func(context.Context) bool { return true })
	b, _ := json.Marshal(r)
	var decoded Report
	if casprovenance.DecodeClosedJSON(b, &decoded, 64<<10) != nil || decoded.Validate() != nil {
		t.Fatal("roundtrip")
	}
	for _, input := range [][]byte{append(append([]byte{}, b[:len(b)-1]...), []byte(`,"unknown":1}`)...), b[:len(b)-1]} {
		if casprovenance.DecodeClosedJSON(input, &decoded, 64<<10) == nil {
			t.Fatal("bad envelope")
		}
	}
	r.ReleaseEvidence = true
	if r.Validate() == nil {
		t.Fatal("release evidence")
	}
	r.ReleaseEvidence = false
	r.Commands = r.Commands[:4]
	if r.Validate() == nil {
		t.Fatal("incomplete pass")
	}
}

func TestQRTraceMustSupportSummary(t *testing.T) {
	for _, mutate := range []func(*Report){
		func(r *Report) { r.Commands = []Command{} },
		func(r *Report) {
			r.Commands = Run(context.Background(), &syntheticExecutor{}, func(context.Context) bool { return true }).Commands
		},
		func(r *Report) { r.Coverage = "partial"; r.Steps[0].Name = livegate.StepLogoutLoggedOut },
		func(r *Report) { r.Coverage = "partial"; r.Steps = append(r.Steps, r.Steps[0]) },
	} {
		r := Run(context.Background(), &syntheticExecutor{fail: livegate.ActionQRLogin}, func(context.Context) bool { return true })
		mutate(&r)
		if r.Validate() == nil {
			t.Fatal("contradictory trace accepted")
		}
	}
}

func TestQRShortParentStopsBeforeNextCommand(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	calls := 0
	r := Run(ctx, executorFunc(func(c context.Context, a livegate.CommandAction) livegate.Observation {
		calls++
		<-c.Done()
		return livegate.Observation{RunnerFault: true}
	}), func(context.Context) bool { return true })
	if calls != 1 || r.Reason != "canceled_or_timeout" || r.Validate() != nil {
		t.Fatalf("deadline result: %#v", r)
	}
}
