package main

import (
	"context"
	"github.com/UnbalancedCat/ipgw-meta/internal/qrdev"
	"os"
	"strings"
	"testing"
)

func TestQRArgumentsNeverStartLive(t *testing.T) {
	for _, args := range [][]string{nil, {"--run-once"}, {"--run-once", "../x"}, {"--password-core", "x"}, {"--anonymous-once"}, {"--preflight", "x"}} {
		r := run(context.Background(), args, t.TempDir(), os.Stdin, os.Stdout, os.Stderr, false)
		if r.LiveAttempted || r.Reason != "arguments_rejected" {
			t.Fatal("invalid args started")
		}
	}
}

func TestQRFinishPreservesInterruptedTrace(t *testing.T) {
	result := qrdev.Run(nil, nil, nil)
	for _, canceled := range []bool{false, true} {
		r := receipt{State: "blocked"}
		var err error
		if canceled {
			err = context.Canceled
		}
		checks := 0
		finish(&r, result, err, func() bool { checks++; return false })
		if r.QR == nil || r.State != "blocked" {
			t.Fatal("lost trace")
		}
		if canceled && (r.Reason != "canceled_or_timeout" || checks != 0) {
			t.Fatal("cancellation masked")
		}
		if !canceled && (r.Reason != "topology_changed" || checks != 1) {
			t.Fatal("topology not checked")
		}
	}
}
func TestQRConfirmationExact(t *testing.T) {
	for _, s := range []string{"RUN-DEV-QR-ONCE\n", "RUN-DEV-QR-ONCE\r\n"} {
		if !confirm(strings.NewReader(s), "RUN-DEV-QR-ONCE") {
			t.Fatal("confirmation rejected")
		}
	}
	for _, s := range []string{"yes\n", "RUN-DEV-QR-ONCE x\n", strings.Repeat("x", 70)} {
		if confirm(strings.NewReader(s), "RUN-DEV-QR-ONCE") {
			t.Fatal("confirmation widened")
		}
	}
}
