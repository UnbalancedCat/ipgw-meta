package main

import (
	"bytes"
	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
	"testing"
)

func TestOnlySyntheticSelfTest(t *testing.T) {
	for _, args := range [][]string{nil, {"--url", "https://private-canary.invalid/"}, {"--self-test", "--live"}} {
		var out, diagnostic bytes.Buffer
		if run(args, &out, &diagnostic) != 2 || out.Len() != 0 || bytes.Contains(diagnostic.Bytes(), []byte("private-canary")) {
			t.Fatal("argument boundary")
		}
	}
	var out, diagnostic bytes.Buffer
	if run([]string{"--self-test"}, &out, &diagnostic) != 0 || diagnostic.Len() != 0 {
		t.Fatal("self-test failed")
	}
	r, err := cascollect.DecodeReport(out.Bytes())
	if err != nil || r.Collection != "complete" {
		t.Fatal("invalid output")
	}
	for _, canary := range []string{"gateway.invalid", "cas.invalid", "synthetic-cookie-canary", "publicKeyStr=", "BEGIN PUBLIC"} {
		if bytes.Contains(out.Bytes(), []byte(canary)) {
			t.Fatal("sensitive input echoed")
		}
	}
}
