package livegate

import (
	"context"
	"github.com/UnbalancedCat/ipgw-meta/internal/labpath"
	"net/netip"
	"sync/atomic"
	"time"
)

func usableBindIP(ip netip.Addr) bool {
	return ip.Is4() && !ip.IsUnspecified() && !ip.IsLoopback() && !ip.IsMulticast() && !ip.IsLinkLocalUnicast()
}

type pathExecutor struct {
	inner    preparedCommandExecutor
	capture  func(context.Context) (func(context.Context) bool, error)
	check    func(context.Context) bool
	stopped  atomic.Bool
	interval time.Duration
	rejected func(string)
}

// GuardPath shares the formal monitor with auxiliary commands. The returned
// function remains stopped after a path loss, including on a fresh context.
func GuardPath(check func(context.Context) bool) func(context.Context, func(context.Context)) bool {
	e := &pathExecutor{check: check, interval: time.Second}
	return e.guarded
}

// GuardPathObserved reports only confirmed guard failures. A query canceled by
// normal work completion is not a failure. The monitor is joined as before.
func GuardPathObserved(check func(context.Context) bool, rejected func(string)) func(context.Context, func(context.Context)) bool {
	e := &pathExecutor{check: check, interval: time.Second, rejected: rejected}
	return e.guarded
}

func (e *pathExecutor) reject(phase string) {
	e.stopped.Store(true)
	if e.rejected != nil {
		e.rejected(phase)
	}
}

func newPathExecutor(inner preparedCommandExecutor, invocation liveGateInvocation) *pathExecutor {
	return &pathExecutor{inner: inner, interval: time.Second, capture: func(ctx context.Context) (func(context.Context) bool, error) {
		snapshot, err := labpath.Capture(ctx, string(invocation.network), invocation.bindIP)
		if err == nil && invocation.network == NetworkCampusWiFi && (invocation.wifiSSID == nil || string(snapshot.WiFiSSID) != *invocation.wifiSSID) {
			return nil, labpath.ErrTopology
		}
		return snapshot.Check, err
	}}
}

func (e *pathExecutor) PrepareProfile(ctx context.Context, suite Suite) error {
	check, err := e.capture(ctx)
	if err != nil {
		e.stopped.Store(true)
		return ErrProcessProfile
	}
	e.check = check
	var preparation error
	if !e.guarded(ctx, func(c context.Context) { preparation = e.inner.PrepareProfile(c, suite) }) {
		return ErrProcessProfile
	}
	return preparation
}

func (e *pathExecutor) Execute(ctx context.Context, action CommandAction) Observation {
	var observation Observation
	if !e.guarded(ctx, func(c context.Context) { observation = e.inner.Execute(c, action) }) {
		return runnerFaultObservation()
	}
	return observation
}

// Each monitor is joined before returning. A lost path remains stopped even
// when the state machine supplies an independent cleanup context.
func (e *pathExecutor) guarded(ctx context.Context, work func(context.Context)) bool {
	if e.stopped.Load() {
		return false
	}
	if e.check == nil || !e.check(ctx) {
		e.reject("before")
		return false
	}
	child, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	interval := e.interval
	if interval <= 0 {
		interval = time.Second
	}
	go func() {
		defer close(done)
		timer := time.NewTicker(interval)
		defer timer.Stop()
		for {
			select {
			case <-child.Done():
				return
			case <-timer.C:
				if !e.check(child) {
					if child.Err() == nil {
						e.reject("during")
					}
					cancel()
					return
				}
			}
		}
	}()
	work(child)
	cancel()
	<-done
	if !e.stopped.Load() && ctx.Err() == nil && !e.check(ctx) {
		e.reject("after")
	}
	return !e.stopped.Load()
}
