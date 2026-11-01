package main

import (
	"io"
	"strings"
	"testing"
)

func TestConfirmationDoesNotConsumeFutureCredentialInput(t *testing.T) {
	r := strings.NewReader("RUN-DEV-PASSWORD-ONCE\r\nprivate-synthetic-input")
	if !confirm(r, "RUN-DEV-PASSWORD-ONCE") {
		t.Fatal("confirmation failed")
	}
	left, _ := io.ReadAll(r)
	if string(left) != "private-synthetic-input" {
		t.Fatal("confirmation consumed future input")
	}
	for _, s := range []string{"RUN-DEV-PASSWORD-ONCE", "no\n", strings.Repeat("x", 65) + "\n"} {
		if confirm(strings.NewReader(s), "RUN-DEV-PASSWORD-ONCE") {
			t.Fatal("bad confirmation accepted")
		}
	}
}
func TestCappedOutput(t *testing.T) {
	b := &capped{limit: 3}
	if _, err := b.Write([]byte("abcd")); err == nil || b.Len() != 0 {
		t.Fatal("overflow retained")
	}
}
