package main

import (
	"io"
	"strings"
	"testing"
)

func TestNoArgumentsCannotStart(t *testing.T) {
	for _, args := range [][]string{nil, {"--run-once", "extra"}, {"--anonymous-once-confirmed"}, {"https://outside.invalid/"}} {
		r := runHistorical(args, strings.NewReader("RUN-ANONYMOUS-ONCE\n"), true, io.Discard, t.TempDir())
		if r.State != "arguments_rejected" || r.LiveStarted || r.Result != nil {
			t.Fatal("argument bypass")
		}
	}
}
func TestNoAnchorCannotStart(t *testing.T) {
	for _, mode := range []string{"--run-once", "--preflight", "--cache-check"} {
		r := runHistorical([]string{mode}, strings.NewReader("RUN-ANONYMOUS-ONCE\n"), true, io.Discard, t.TempDir())
		if r.State != "binding_invalid" || r.LiveStarted || r.LiveAttempted || r.Result != nil {
			t.Fatal("binding bypass")
		}
	}
}

func TestRetiredLiveEntryRejectsBeforeBinding(t *testing.T) {
	r := run([]string{"--run-once"}, strings.NewReader("RUN-ANONYMOUS-ONCE\n"), true, io.Discard, t.TempDir())
	if r.State != "arguments_rejected" || r.LiveStarted || r.LiveAttempted || r.Result != nil {
		t.Fatal("retired live mode")
	}
}
