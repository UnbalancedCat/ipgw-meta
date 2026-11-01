package main

import (
	"context"
	"testing"
)

func TestRetiredAnonymousEntry(t *testing.T) {
	r := run(context.Background(), []string{"--anonymous-confirmed"})
	if r.Mode != "preflight" || r.State != "blocked" {
		t.Fatal("retired live entry accepted")
	}
}
