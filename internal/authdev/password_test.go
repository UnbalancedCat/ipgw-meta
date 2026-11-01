package authdev

import (
	"context"
	"github.com/UnbalancedCat/ipgw-meta/internal/livegate"
	"testing"
)

type commandFunc func(context.Context, livegate.CommandAction) livegate.Observation

func (f commandFunc) Execute(ctx context.Context, a livegate.CommandAction) livegate.Observation {
	return f(ctx, a)
}
func passing(a livegate.CommandAction) livegate.Observation {
	o := livegate.Observation{}
	switch a {
	case livegate.ActionInitialStatus, livegate.ActionFinalStatus, livegate.ActionCleanupStatus:
		o.Session = livegate.SessionOffline
	case livegate.ActionStatusOnline:
		o.Session = livegate.SessionOnline
		o.IdentityMatches = true
	case livegate.ActionPasswordLogin:
		o.Session = livegate.SessionOnline
		o.IdentityMatches = true
		o.Outcome = livegate.OutcomeLoggedIn
	case livegate.ActionSecondPasswordLogin:
		o.Session = livegate.SessionOnline
		o.IdentityMatches = true
		o.Outcome = livegate.OutcomeAlreadyOnline
	case livegate.ActionLogout, livegate.ActionCleanupLogout:
		o.Outcome = livegate.OutcomeLoggedOut
		o.Session = livegate.SessionOffline
	case livegate.ActionSecondLogout:
		o.Outcome = livegate.OutcomeAlreadyOffline
		o.Session = livegate.SessionOffline
	}
	return o
}
func TestPasswordStateMachinePassAndCleanup(t *testing.T) {
	for _, fail := range []bool{false, true} {
		var actions []livegate.CommandAction
		r := Password(context.Background(), commandFunc(func(_ context.Context, a livegate.CommandAction) livegate.Observation {
			actions = append(actions, a)
			if fail && a == livegate.ActionSecondPasswordLogin {
				return livegate.Observation{RunnerFault: true}
			}
			return passing(a)
		}), func(context.Context) bool { return true })
		if !fail && (r.State != "pass" || len(actions) != 7) {
			t.Fatalf("pass failed: %+v", r)
		}
		if fail && (r.State == "pass" || len(actions) != 6 || actions[4] != livegate.ActionCleanupLogout || actions[5] != livegate.ActionCleanupStatus) {
			t.Fatal("cleanup ownership changed")
		}
		if r.Validate() != nil {
			t.Fatalf("invalid result: %+v", r)
		}
	}
}
func TestPasswordTopologyChangeDoesNotSendCleanup(t *testing.T) {
	checks := 0
	calls := 0
	r := Password(context.Background(), commandFunc(func(_ context.Context, a livegate.CommandAction) livegate.Observation { calls++; return passing(a) }), func(context.Context) bool { checks++; return checks <= 3 })
	if calls != 3 || r.Reason != "topology_changed" || r.State == "pass" || len(r.Steps) != 3 || r.StepCoverage != "partial" || r.Validate() != nil {
		t.Fatalf("bad changed path: %+v", r)
	}
}
