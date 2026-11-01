package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/UnbalancedCat/ipgw-meta/internal/authcontext"
	"strings"
	"testing"
)

func TestConfirmation(t *testing.T) {
	for _, tc := range []struct {
		input string
		ok    bool
	}{{"RUN-DEV-CONTEXT-ONCE\n", true}, {"RUN-DEV-CONTEXT-ONCE\r\n", true}, {"RUN-DEV-ANONYMOUS-ONCE\n", false}, {"RUN-DEV-CONTEXT-ONCE", false}, {strings.Repeat("a", 65), false}} {
		if confirm(strings.NewReader(tc.input)) != tc.ok {
			t.Fatal("confirmation mismatch")
		}
	}
}
func TestArgumentsNeverStart(t *testing.T) {
	for _, args := range [][]string{nil, {"--password-core", "synthetic"}, {"--anonymous-confirmed"}, {"--anonymous-once", "extra"}} {
		var p bytes.Buffer
		r := run(context.Background(), args, "invalid", strings.NewReader(""), &p, false)
		if r.State != "arguments_rejected" || r.LiveAttempted || p.Len() != 0 {
			t.Fatal("argument boundary")
		}
	}
}
func TestStrictReceipt(t *testing.T) {
	result := authcontext.NewReport("preflight")
	result.State, result.Reason = "ready", "none"
	result.SDK.State, result.SDK.Reason = "ready", "none"
	r := receipt{SchemaVersion: 1, ToolKind: "auth-context-dev", SourceDigest: strings.Repeat("a", 64), ProductDigest: strings.Repeat("b", 64), ObserverDigest: strings.Repeat("c", 64), Binding: "verified", State: "completed", Result: &result}
	encode := func(r receipt) []byte {
		d, e := json.Marshal(r)
		if e != nil {
			t.Fatal(e)
		}
		return d
	}
	if _, ok := decodeReceipt(encode(r)); !ok {
		t.Fatal("valid receipt rejected")
	}
	for _, mutate := range []func(*receipt){func(r *receipt) { r.SchemaVersion = 2 }, func(r *receipt) { r.ToolKind = "auth-dev" }, func(r *receipt) { r.LiveAttempted = true }, func(r *receipt) { r.Binding = "invalid" }, func(r *receipt) { r.ProductDigest = "unknown" }, func(r *receipt) { r.Result = nil }, func(r *receipt) { r.State = "canceled" }} {
		bad := r
		mutate(&bad)
		if _, ok := decodeReceipt(encode(bad)); ok {
			t.Fatal("contradiction accepted")
		}
	}
	d := string(encode(r))
	for _, s := range []string{d + d, strings.Replace(d, `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1), strings.Replace(d, `"schema_version":1`, `"schema_version":1,"unknown":true`, 1)} {
		if _, ok := decodeReceipt([]byte(s)); ok {
			t.Fatal("invalid framing")
		}
	}
}
