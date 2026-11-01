package main

import "testing"

func TestUnsupportedArgsNeverRead(t *testing.T) {
	for _, a := range [][]string{nil, {"--run-once"}, {"--ctl-identifiers", "extra"}, {"--file"}, {"--trust-audit"}, {"--ctl-check"}} {
		if _, e := run(a); e == nil {
			t.Fatal("accepted unsafe argument")
		}
	}
}
