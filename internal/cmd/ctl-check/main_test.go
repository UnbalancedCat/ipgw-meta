package main

import "testing"

func TestUnsupportedArgsNeverRead(t *testing.T) {
	for _, a := range [][]string{nil, {"--run-once"}, {"--ctl-check", "extra"}, {"--file"}, {"--trust-audit"}} {
		if _, e := run(a); e == nil {
			t.Fatal("accepted unsafe argument")
		}
	}
}
