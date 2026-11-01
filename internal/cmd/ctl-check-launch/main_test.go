package main

import "testing"

func TestArgumentsBeforeBinding(t *testing.T) {
	for _, a := range [][]string{nil, {"--run-once"}, {"--anonymous-once-confirmed"}, {"--ctl-check", "extra"}, {"--trust-audit"}} {
		r := run(a, "unread-synthetic")
		if r.State != "arguments_rejected" || r.Result != nil {
			t.Fatal(r)
		}
	}
}
