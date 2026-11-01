package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestNoImplicitTrustAnchor(t *testing.T) {
	for _, args := range [][]string{nil, {"--bundle", "synthetic"}, {"--bundle", "synthetic", "--manifest-sha256", "", "extra"}} {
		var out, err bytes.Buffer
		if run(args, &out, &err) != 2 || out.Len() != 0 || strings.Contains(err.String(), "synthetic") {
			t.Fatal("argument contract")
		}
	}
}
func TestRejectionDoesNotEchoPath(t *testing.T) {
	var out, err bytes.Buffer
	if run([]string{"--bundle", "untrusted-canary", "--manifest-sha256", ""}, &out, &err) != 1 || bytes.Contains(out.Bytes(), []byte("untrusted-canary")) || err.Len() != 0 {
		t.Fatal("rejection output")
	}
}

func TestInstallRequiresExplicitAnchors(t *testing.T) {
	for _, args := range [][]string{{"--install-archive", "untrusted-canary"}, {"--install-archive", "untrusted-canary", "--archive-sha256", "", "--manifest-sha256", "", "--destination", "private-canary"}} {
		var out, diagnostic bytes.Buffer
		code := run(args, &out, &diagnostic)
		if code == 0 || bytes.Contains(out.Bytes(), []byte("canary")) || strings.Contains(diagnostic.String(), "canary") {
			t.Fatal("install argument contract")
		}
	}
}
