// Package qrdev implements only command-level development validation, not HTTP observation.
package qrdev

import (
	"context"
	"errors"
	"github.com/UnbalancedCat/ipgw-meta/internal/livegate"
	"time"
)

var ErrReport = errors.New("qr-dev: invalid report")

type Command struct {
	Action livegate.CommandAction `json:"command"`
	State  string                 `json:"state"`
	Step   *livegate.Step         `json:"result"`
}
type Report struct {
	Schema          int             `json:"schema_version"`
	Kind            string          `json:"tool_kind"`
	ReleaseEvidence bool            `json:"release_evidence"`
	State           string          `json:"state"`
	Reason          string          `json:"reason"`
	Coverage        string          `json:"coverage"`
	Commands        []Command       `json:"commands"`
	Steps           []livegate.Step `json:"steps"`
}

type recorder struct {
	root     context.Context
	next     livegate.CommandExecutor
	guard    func(context.Context, func(context.Context)) bool
	commands []Command
	stopped  bool
	reason   string
	action   livegate.CommandAction
}

func (r *recorder) Execute(ctx context.Context, a livegate.CommandAction) livegate.Observation {
	fault := livegate.Observation{RunnerFault: true}
	if r.stopped || r.root.Err() != nil || ctx.Err() != nil {
		r.stopped = true
		return fault
	}
	limit := 120 * time.Second
	if a == livegate.ActionQRLogin {
		limit = 300 * time.Second
	}
	child, cancel := context.WithTimeout(ctx, limit)
	defer cancel()
	// StateMachine cleanup gets an independent context; root cancellation must still stop it.
	stop := context.AfterFunc(r.root, cancel)
	defer stop()
	entry := Command{Action: a, State: "blocked"}
	var obs livegate.Observation
	started := time.Now()
	r.action = a
	guardOK := r.guard(child, func(c context.Context) { obs = r.next.Execute(c, a) })
	if !guardOK {
		r.stopped = true
		r.reason = "topology_or_binding_changed"
		obs = fault
	}
	if child.Err() != nil {
		r.stopped = true
		r.reason = "canceled_or_timeout"
		obs = fault
	}
	if !obs.RunnerFault {
		step, err := livegate.ProjectDevelopmentQRCommand(a, obs, int64(time.Since(started).Seconds()))
		if err == nil {
			entry.State = "complete"
			entry.Step = &step
		} else {
			obs = fault
			r.reason = "command_invalid"
		}
	}
	if obs.RunnerFault {
		r.stopped = true
		if r.reason == "" {
			r.reason = "command_failed"
		}
	}
	r.commands = append(r.commands, entry)
	return obs
}

func Run(ctx context.Context, e livegate.CommandExecutor, check func(context.Context) bool) Report {
	return RunObserved(ctx, e, check, nil)
}

// RunObserved adds guard location reporting without observing authentication
// data or changing the state machine. The callback runs before monitor join.
func RunObserved(ctx context.Context, e livegate.CommandExecutor, check func(context.Context) bool, rejected func(livegate.CommandAction, string)) Report {
	r := Report{Schema: 1, Kind: "qr-dev", State: "blocked", Reason: "arguments_rejected", Coverage: "partial", Commands: []Command{}, Steps: []livegate.Step{}}
	if ctx == nil || e == nil || check == nil {
		return r
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	rec := &recorder{root: ctx, next: e}
	rec.guard = livegate.GuardPathObserved(check, func(phase string) {
		if rejected != nil {
			rejected(rec.action, phase)
		}
	})
	steps, result, canceled, err := (livegate.StateMachine{Executor: rec}).Run(ctx, livegate.SuiteTerminalQR)
	if steps != nil {
		r.Steps = steps
	}
	if rec.commands != nil {
		r.Commands = rec.commands
	}
	r.State = string(result)
	r.Reason = "state_machine_failed"
	if err == nil && !canceled && !rec.stopped {
		r.Coverage = "complete"
	}
	if result == livegate.ResultPass && r.Coverage == "complete" {
		r.Reason = "none"
	}
	if rec.reason != "" {
		r.Reason = rec.reason
	}
	if ctx.Err() != nil {
		r.Reason = "canceled_or_timeout"
	}
	if r.Coverage != "complete" && r.State == "pass" {
		r.State = "blocked"
	}
	return r
}

func (r Report) Validate() error {
	if r.Schema != 1 || r.Kind != "qr-dev" || r.ReleaseEvidence || r.Commands == nil || r.Steps == nil || len(r.Commands) > 7 || len(r.Steps) > 7 {
		return ErrReport
	}
	if r.State != "pass" && r.State != "fail" && r.State != "blocked" {
		return ErrReport
	}
	if r.Coverage != "complete" && r.Coverage != "partial" {
		return ErrReport
	}
	switch r.Reason {
	case "none", "arguments_rejected", "state_machine_failed", "topology_or_binding_changed", "canceled_or_timeout", "command_invalid", "command_failed":
	default:
		return ErrReport
	}
	order := []livegate.CommandAction{livegate.ActionInitialStatus, livegate.ActionQRLogin, livegate.ActionStatusOnline, livegate.ActionLogout, livegate.ActionFinalStatus, livegate.ActionCleanupLogout, livegate.ActionCleanupStatus}
	names := []livegate.StepName{livegate.StepInitialStatusOffline, livegate.StepQRLoginLoggedIn, livegate.StepStatusOnline, livegate.StepLogoutLoggedOut, livegate.StepFinalStatusOffline, livegate.StepCleanupLogout, livegate.StepCleanupStatusOffline}
	last := -1
	owned := false
	offline := false
	blocked := false
	primaryFailed := false
	projected := []livegate.Step{}
	var pendingQR *livegate.Step
	for _, c := range r.Commands {
		idx := -1
		for i, a := range order {
			if a == c.Action {
				idx = i
			}
		}
		if blocked || idx < 0 || idx <= last || idx < 5 && (idx != last+1 || primaryFailed) || idx >= 5 && (!owned || offline || idx == 6 && last != 5) {
			return ErrReport
		}
		last = idx
		if c.State == "blocked" {
			if c.Step != nil || r.Coverage != "partial" {
				return ErrReport
			}
			blocked = true
			continue
		}
		if c.State != "complete" || c.Step == nil || c.Step.Validate() != nil || c.Step.Name != names[idx] {
			return ErrReport
		}
		if idx == 2 && c.Step.Result == livegate.ResultPass {
			owned = true
		}
		if idx == 4 && c.Step.Result == livegate.ResultPass {
			offline = true
		}
		step := *c.Step
		if idx < 5 && step.Result != livegate.ResultPass {
			primaryFailed = true
		}
		switch {
		case idx == 1 && step.Result == livegate.ResultPass:
			pendingQR = &step
		case idx == 2:
			if pendingQR == nil {
				return ErrReport
			}
			if step.Result == livegate.ResultPass {
				projected = append(projected, *pendingQR, step)
			} else {
				merged := *pendingQR
				merged.Result = step.Result
				merged.DurationSeconds += step.DurationSeconds
				merged.ExitCode, merged.ErrorCode = step.ExitCode, step.ErrorCode
				projected = append(projected, merged)
			}
			pendingQR = nil
		default:
			projected = append(projected, step)
		}
	}
	if len(r.Steps) > len(projected) || r.Coverage == "complete" && (len(r.Steps) != len(projected) || pendingQR != nil) {
		return ErrReport
	}
	for i, s := range r.Steps {
		p := projected[i]
		// Both timings measure the same guarded command; the outer state
		// machine can cross one extra second boundary (two for a QR pair).
		slack := int64(1)
		if s.Name == livegate.StepQRLoginLoggedIn && s.Result != livegate.ResultPass {
			slack = 2
		}
		if s.Validate() != nil || s.Name != p.Name || s.Result != p.Result || s.ExitCode != p.ExitCode || !sameError(s.ErrorCode, p.ErrorCode) || s.DurationSeconds < p.DurationSeconds || s.DurationSeconds > p.DurationSeconds+slack {
			return ErrReport
		}
	}
	if r.Coverage == "complete" && r.Reason != "none" && r.Reason != "state_machine_failed" {
		return ErrReport
	}
	if r.State != "pass" {
		expected := "blocked"
		for _, s := range r.Steps {
			if s.Result == livegate.ResultFail {
				expected = "fail"
			}
		}
		if r.State != expected {
			return ErrReport
		}
	}
	if r.Coverage == "complete" && livegate.ValidateDevelopmentQRSteps(r.Steps, livegate.Result(r.State)) != nil {
		return ErrReport
	}
	if r.State == "pass" {
		if r.Reason != "none" || r.Coverage != "complete" || len(r.Commands) != 5 {
			return ErrReport
		}
		for _, c := range r.Commands {
			if c.State != "complete" || c.Step.Result != livegate.ResultPass {
				return ErrReport
			}
		}
	} else if r.Reason == "none" {
		return ErrReport
	}
	return nil
}

func sameError(a, b *livegate.ErrorCode) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}
