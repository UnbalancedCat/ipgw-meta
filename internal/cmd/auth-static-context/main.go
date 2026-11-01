package main

import (
	"context"
	"fmt"
	"github.com/UnbalancedCat/ipgw-meta/internal/authcontext"
	"github.com/UnbalancedCat/ipgw-meta/internal/authdev"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func run(ctx context.Context, args []string) authcontext.StaticReport {
	if len(args) == 1 && args[0] == "--self-test" {
		return authcontext.SelfTestStatic()
	}
	r := authcontext.NewStaticReport("preflight")
	if len(args) != 1 || args[0] != "--local-preflight" && args[0] != "--anonymous-confirmed" {
		return r
	}
	if args[0] == "--anonymous-confirmed" {
		return r // Retired by ADR-0021: no topology or network operation.
	}
	_, err := authdev.ReadTopology(ctx)
	if err != nil {
		return r
	}
	if r.Mode == "preflight" {
		r.State, r.Reason = "ready", "none"
		r.Context.State, r.Context.Reason = "ready", "none"
		r.Context.SDK.State, r.Context.SDK.Reason = "ready", "none"
		return r
	}
	return r

}
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if len(os.Args) == 2 && os.Args[1] == "--self-test" && runtime.GOOS == "windows" && !authcontext.ProtectStaticProcess() {
		os.Exit(4)
	}
	r := run(ctx, os.Args[1:])
	b, e := authcontext.EncodeStatic(r)
	if e != nil {
		os.Exit(2)
	}
	fmt.Println(string(b))
}
