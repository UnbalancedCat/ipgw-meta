package qrdev

import (
	"context"
	"encoding/json"
	"github.com/UnbalancedCat/ipgw-meta/internal/livegate"
	"strings"
)

type syntheticExecutor struct {
	actions []livegate.CommandAction
	fail    livegate.CommandAction
}

func (e *syntheticExecutor) Execute(_ context.Context, a livegate.CommandAction) livegate.Observation {
	e.actions = append(e.actions, a)
	if a == e.fail {
		code := livegate.ErrorProtocolChanged
		return livegate.Observation{ExitCode: 6, ErrorCode: &code}
	}
	switch a {
	case livegate.ActionInitialStatus, livegate.ActionFinalStatus, livegate.ActionCleanupStatus:
		return livegate.Observation{Session: livegate.SessionOffline}
	case livegate.ActionQRLogin:
		return livegate.Observation{}
	case livegate.ActionStatusOnline:
		return livegate.Observation{Session: livegate.SessionOnline, IdentityMatches: true}
	case livegate.ActionLogout, livegate.ActionCleanupLogout:
		return livegate.Observation{Session: livegate.SessionOffline, Outcome: livegate.OutcomeLoggedOut}
	default:
		return livegate.Observation{RunnerFault: true}
	}
}
func SelfTest() bool {
	for _, fail := range []livegate.CommandAction{"", livegate.ActionQRLogin, livegate.ActionStatusOnline, livegate.ActionLogout} {
		e := &syntheticExecutor{fail: fail}
		r := Run(context.Background(), e, func(context.Context) bool { return true })
		if r.Validate() != nil || (r.State == "pass") != (fail == "") {
			return false
		}
		if fail == livegate.ActionQRLogin && len(r.Commands) != 2 {
			return false
		}
		if fail == livegate.ActionStatusOnline && len(r.Commands) != 3 {
			return false
		}
	}
	inner := Run(context.Background(), &syntheticExecutor{}, func(context.Context) bool { return true })
	r := Receipt{Schema: 2, Kind: "qr-dev", Source: strings.Repeat("a", 64), Product: strings.Repeat("b", 64), Tool: strings.Repeat("c", 64), Binding: "verified", State: "blocked", Reason: "topology_changed", LiveAttempted: true, QR: &inner, GateFailure: &GateFailure{Command: "final_check", Phase: "final", Reason: "query_unavailable", QueryReason: "deadline"}}
	raw, err := json.Marshal(r)
	if err != nil {
		return false
	}
	if _, err = DecodeReceipt(raw); err != nil {
		return false
	}
	r.Schema = 1
	if r.Validate() == nil {
		return false
	}
	return true
}
