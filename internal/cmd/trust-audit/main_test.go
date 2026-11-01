package main

import (
	"github.com/UnbalancedCat/ipgw-meta/internal/trustaudit"
	"testing"
)

func TestOnlyLocalCommands(t *testing.T) {
	for _, args := range [][]string{nil, {"--run-once"}, {"--anonymous-once-confirmed"}, {"--trust-audit", "extra"}} {
		if _, err := run(args); err == nil {
			t.Fatal("unknown operation")
		}
	}
	b, err := run([]string{"--self-test"})
	if err != nil {
		t.Fatal(err)
	}
	r, err := trustaudit.Decode(b)
	if err != nil || r.Mode != "self-test" || r.State != "complete" {
		t.Fatal("self-test")
	}
}
