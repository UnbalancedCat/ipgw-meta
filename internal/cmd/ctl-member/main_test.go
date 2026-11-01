package main

import "testing"

func TestNoExternalInputModes(t *testing.T) {
	for _, a := range [][]string{nil, {"--ctl-check"}, {"--ctl-identifiers"}, {"--run-once"}, {"--file", "synthetic"}, {"--self-test", "extra"}} {
		if _, e := run(a); e == nil {
			t.Fatal("external input accepted")
		}
	}
}
