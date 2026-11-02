package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestWorkerChild(t *testing.T) {
	if len(os.Args) < 3 || os.Args[len(os.Args)-2] != "--synthetic-worker" {
		return
	}
	switch os.Args[len(os.Args)-1] {
	case "valid":
		fmt.Print(`{"passed":true}`)
	case "unknown":
		fmt.Print(`{"passed":true,"extra":true}`)
	case "overflow":
		fmt.Print(strings.Repeat("x", 70<<10))
	case "wait":
		time.Sleep(5 * time.Second)
	}
	os.Exit(0)
}

func TestWorkerClosedOutputAndHardCancellation(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"valid", "unknown", "overflow", "wait"} {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		if mode == "wait" {
			cancel()
			ctx, cancel = context.WithTimeout(context.Background(), 50*time.Millisecond)
		}
		var out struct {
			Passed bool `json:"passed"`
		}
		ok := worker(ctx, exe, []string{"-test.run=^TestWorkerChild$", "--", "--synthetic-worker", mode}, &out)
		cancel()
		if ok != (mode == "valid") {
			t.Fatalf("worker %s: %v", mode, ok)
		}
	}
}

func TestRejectArgumentsBeforeBindingOrExecution(t *testing.T) {
	for _, args := range [][]string{nil, {"--run-once", "neu_mobile", "test"}, {"--run-once", "neu", "../secret"}, {"--preflight", "other"}, {"--run-once", "neu"}, {"--anonymous-once"}} {
		r := run(context.Background(), args, t.TempDir(), nil, nil, nil, false)
		if r.Reason != "arguments_rejected" || r.LiveAttempted || r.ReleaseEvidence {
			t.Fatal(r)
		}
	}
}
func TestConfirmationDoesNotConsumePassword(t *testing.T) {
	in := strings.NewReader("RUN-WIFI-NEU-ONCE\r\nremaining")
	if !confirm(in, "RUN-WIFI-NEU-ONCE") || in.Len() != len("remaining") {
		t.Fatal("confirmation read beyond line")
	}
	if confirm(strings.NewReader("RUN-WIFI-NEU_24G-ONCE\n"), "RUN-WIFI-NEU-ONCE") {
		t.Fatal("wrong SSID confirmed")
	}
}
