package qrdev

import (
	"context"
	"encoding/json"
	"github.com/UnbalancedCat/ipgw-meta/internal/labpath"
	"github.com/UnbalancedCat/ipgw-meta/internal/livegate"
	"strings"
	"testing"
)

func validGateReceipt() Receipt {
	checks := 0
	g := NewGateTracker(func() error { return nil }, func(context.Context) labpath.Inspection {
		checks++
		if checks == 7 {
			return labpath.Inspection{Reason: "snapshot_changed", QueryReason: "none", Changes: labpath.Changes{Adapters: true}}
		}
		return labpath.Inspection{Reason: "none", QueryReason: "none"}
	})
	r := RunObserved(context.Background(), &syntheticExecutor{}, g.Check, func(a livegate.CommandAction, p string) { g.Reject(string(a), p) })
	return Receipt{Schema: 2, Kind: "qr-dev", Source: strings.Repeat("a", 64), Product: strings.Repeat("b", 64), Tool: strings.Repeat("c", 64), Binding: "verified", State: "blocked", Reason: r.Reason, LiveAttempted: true, Preflight: &PreflightReport{State: "ready", Reason: "none"}, QR: &r, GateFailure: g.Failure()}
}

func TestReceiptVersionsAndStrictFailureShape(t *testing.T) {
	r := validGateReceipt()
	raw, _ := json.Marshal(r)
	if _, e := DecodeReceipt(raw); e != nil {
		t.Fatal("valid v2", e)
	}
	var original map[string]any
	_ = json.Unmarshal(raw, &original)
	for _, mutate := range []func(map[string]any){
		func(m map[string]any) { m["schema_version"] = 1 },
		func(m map[string]any) { delete(m, "gate_failure") },
		func(m map[string]any) { m["gate_failure"] = nil },
		func(m map[string]any) { m["unexpected"] = "PRIVATE_CANARY" },
		func(m map[string]any) { m["state"] = "pass" },
		func(m map[string]any) { m["state"] = "fail" },
		func(m map[string]any) { m["reason"] = "state_machine_failed" },
		func(m map[string]any) { m["gate_failure"].(map[string]any)["command"] = "qr_login" },
		func(m map[string]any) { delete(m["gate_failure"].(map[string]any), "query_reason") },
		func(m map[string]any) { delete(m["gate_failure"].(map[string]any)["changes"].(map[string]any), "dns") },
		func(m map[string]any) { m["gate_failure"].(map[string]any)["changes"].(map[string]any)["dns"] = nil },
	} {
		var m map[string]any
		_ = json.Unmarshal(raw, &m)
		mutate(m)
		b, _ := json.Marshal(m)
		if _, e := DecodeReceipt(b); e == nil {
			t.Fatal("invalid receipt accepted")
		}
	}
	for _, b := range [][]byte{raw[:len(raw)-1], append(append([]byte{}, raw[:len(raw)-1]...), []byte(`,"schema_version":2}`)...)} {
		if _, e := DecodeReceipt(b); e == nil {
			t.Fatal("bad encoding")
		}
	}
	original["schema_version"] = 1
	delete(original, "gate_failure")
	old, _ := json.Marshal(original)
	v, e := DecodeReceipt(old)
	if e != nil || v.GateFailure != nil || v.Schema != 1 {
		t.Fatal("historical v1 changed")
	}
}

func TestReceiptFinalFailureDoesNotRewriteInnerPass(t *testing.T) {
	r := validGateReceipt()
	inner := Run(context.Background(), &syntheticExecutor{}, func(context.Context) bool { return true })
	r.QR = &inner
	r.GateFailure = &GateFailure{Command: "final_check", Phase: "final", Reason: "query_unavailable", QueryReason: "deadline"}
	r.Reason = "topology_changed"
	if r.Validate() != nil || r.QR.State != "pass" {
		t.Fatal("final check semantics")
	}
}

func FuzzQRReceipt(f *testing.F) {
	raw, _ := json.Marshal(validGateReceipt())
	f.Add(raw)
	f.Add([]byte(`{}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		r, e := DecodeReceipt(data)
		if e == nil && r.Validate() != nil {
			t.Fatal("invalid accepted")
		}
	})
}
