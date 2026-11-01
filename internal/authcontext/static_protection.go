package authcontext

import (
	"context"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
	"os"
	"time"
)

// The dedicated child invokes this only after SDK collection has stopped.
// The launcher separately enforces its overall hard deadline and closed output.
func ProtectedStaticAnalysis(ctx context.Context, inputs []casprovenance.Input, index int) casprovenance.StaticReport {
	return protectedStatic(ctx, ProtectStaticProcess, func(ctx context.Context) casprovenance.StaticReport {
		return casprovenance.AnalyzeStaticParserMatch(ctx, inputs, index)
	}, 10*time.Second)
}
func protectedStatic(ctx context.Context, protect func() bool, work func(context.Context) casprovenance.StaticReport, limit time.Duration) casprovenance.StaticReport {
	if ctx == nil || ctx.Err() != nil {
		return casprovenance.EmptyStatic("canceled")
	}
	if !protect() {
		return casprovenance.EmptyStatic("protection_unavailable")
	}
	timer := time.AfterFunc(limit, func() { os.Exit(4) })
	defer timer.Stop()
	ctx, cancel := context.WithTimeout(ctx, limit)
	defer cancel()
	return work(ctx)
}
