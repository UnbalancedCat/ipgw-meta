package main

import "testing"

func TestRejectLiveCommandsBeforeBinding(t *testing.T) {
	for _, args := range [][]string{nil, {"--run-once"}, {"--anonymous-once-confirmed"}, {"--trust-audit", "extra"}} {
		r := run(args, t.TempDir())
		if r.State != "arguments_rejected" || r.Result != nil {
			t.Fatal("unexpected entry")
		}
	}
}
