package main

import (
	"context"
	"fmt"
	"github.com/UnbalancedCat/ipgw-meta/internal/authdev"
	"os"
	"os/signal"
	"syscall"
)

func run(ctx context.Context, args []string) authdev.Report {
	if len(args) == 1 && args[0] == "--self-test" {
		return authdev.SelfTest()
	}
	r := authdev.NewReport("preflight")
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
		return r
	}
	return r

}
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	r := run(ctx, os.Args[1:])
	data, err := authdev.Encode(r)
	if err != nil {
		os.Exit(2)
	}
	fmt.Println(string(data))
}
