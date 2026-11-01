package main

import (
	"context"
	"os"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/casacquire"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--anonymous-once-confirmed" {
		os.Exit(2)
	} // ADR-0021: no live entry in new builds.
	if len(os.Args) != 2 || (os.Args[1] != "--self-test" && os.Args[1] != "--anonymous-once-confirmed" && os.Args[1] != "--local-preflight" && os.Args[1] != "--local-cache-check") {
		os.Exit(2)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 140*time.Second)
	defer cancel()
	var r casacquire.Report
	if os.Args[1] == "--self-test" {
		r = casacquire.SelfTest(ctx)
	} else if os.Args[1] == "--local-cache-check" {
		r = casacquire.CacheCheck(ctx)
	} else if os.Args[1] == "--local-preflight" {
		r = casacquire.Preflight(ctx)
	} else {
		os.Exit(2)
	}
	b, err := casacquire.Encode(r)
	if err != nil {
		os.Exit(3)
	}
	if _, err = os.Stdout.Write(b); err != nil {
		os.Exit(3)
	}
}
