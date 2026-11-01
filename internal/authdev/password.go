package authdev

import (
	"context"
	"github.com/UnbalancedCat/ipgw-meta/internal/livegate"
	"sync/atomic"
	"time"
)

// Password shares the existing password state machine, including cleanup
// ownership. A failed topology check is sticky and blocks cleanup too.
func Password(ctx context.Context, executor livegate.CommandExecutor, check func(context.Context) bool) Report {
	r := NewReport("password")
	if ctx == nil || executor == nil || check == nil {
		return r
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	g := &checkedExecutor{next: executor, check: check}
	steps, result, canceled, err := (livegate.StateMachine{Executor: g}).Run(ctx, livegate.SuitePasswordCore)
	if steps != nil {
		r.Steps = steps
	}
	r.Commands = g.calls
	r.State = string(result)
	r.StepCoverage = "complete"
	if err != nil || canceled {
		r.StepCoverage = "partial"
	}
	switch {
	case g.changed.Load():
		r.Reason = "topology_changed"
	case canceled:
		r.Reason = "canceled"
	case err != nil || result != livegate.ResultPass:
		r.Reason = "state_machine_failed"
	default:
		r.Reason = "none"
	}
	return r
}

type checkedExecutor struct {
	next    livegate.CommandExecutor
	check   func(context.Context) bool
	changed atomic.Bool
	calls   int
}

func (g *checkedExecutor) Execute(ctx context.Context, action livegate.CommandAction) livegate.Observation {
	if g.changed.Load() || !g.check(ctx) {
		g.changed.Store(true)
		return livegate.Observation{RunnerFault: true}
	}
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	done := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		tick := time.NewTicker(time.Second)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			case <-tick.C:
				if !g.check(ctx) {
					if ctx.Err() == nil {
						g.changed.Store(true)
					}
					cancel()
					return
				}
			}
		}
	}()
	g.calls++
	observation := g.next.Execute(ctx, action)
	close(done)
	cancel()
	<-stopped
	if g.changed.Load() {
		return livegate.Observation{RunnerFault: true}
	}
	return observation
}
