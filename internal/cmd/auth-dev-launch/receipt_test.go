package main

import (
	"encoding/json"
	"github.com/UnbalancedCat/ipgw-meta/internal/authdev"
	"github.com/UnbalancedCat/ipgw-meta/internal/authobserve"
	"github.com/UnbalancedCat/ipgw-meta/internal/livegate"
	"strings"
	"testing"
)

func TestReceiptDoesNotAcceptV1OrMixedObservation(t *testing.T) {
	result := authdev.NewReport("preflight")
	result.State = "ready"
	result.Reason = "none"
	r := receipt{SchemaVersion: 2, ToolKind: "auth-dev-observed", SourceDigest: strings.Repeat("a", 64), ProductDigest: strings.Repeat("b", 64), ObservedDigest: strings.Repeat("c", 64), Binding: "verified", State: "completed", Result: &result, Observations: authobserve.NewWindow()}
	data, _ := json.Marshal(r)
	if _, ok := decodeReceipt(data); !ok {
		t.Fatal("valid receipt")
	}
	for _, bad := range []string{strings.Replace(string(data), `"schema_version":2`, `"schema_version":1`, 1), strings.Replace(string(data), `"tool_kind":"auth-dev-observed"`, `"tool_kind":"auth-dev"`, 1), strings.Replace(string(data), `"observations":`, `"unexpected":`, 1), strings.Replace(string(data), `"live_attempted":false`, `"live_attempted":false,"live_attempted":false`, 1)} {
		if _, ok := decodeReceipt([]byte(bad)); ok {
			t.Fatal("mixed/duplicate/unknown receipt accepted")
		}
	}
	r.Observations = authobserve.Window{State: "complete", Reason: "none", Commands: []authobserve.Observation{authobserve.New("profile_show").Snapshot()}}
	data, _ = json.Marshal(r)
	if _, ok := decodeReceipt(data); ok {
		t.Fatal("preflight with password observations")
	}
}

func TestPasswordReceiptRejectsFalseAttemptAndWrongCommandSequence(t *testing.T) {
	code := livegate.ErrorNetwork
	result := authdev.NewReport("password")
	result.Reason = "state_machine_failed"
	result.StepCoverage = "complete"
	result.Commands = 3
	result.Steps = []livegate.Step{
		{Name: livegate.StepInitialStatusOffline, Result: livegate.ResultPass},
		{Name: livegate.StepLoginLoggedIn, Result: livegate.ResultBlocked, ExitCode: 3, ErrorCode: &code},
	}
	status := authobserve.New("status").Snapshot()
	status.SDKCategory = "success"
	login := authobserve.New("login").Snapshot()
	login.SDKCode = "network"
	login.SDKCategory = "cas_http_status"
	r := receipt{SchemaVersion: 2, ToolKind: "auth-dev-observed", SourceDigest: strings.Repeat("a", 64), ProductDigest: strings.Repeat("b", 64), ObservedDigest: strings.Repeat("c", 64), Binding: "verified", State: "completed", LiveAttempted: true, Result: &result, Observations: authobserve.Window{State: "complete", Reason: "none", Commands: []authobserve.Observation{authobserve.New("profile_show").Snapshot(), status, login}}}
	data, _ := json.Marshal(r)
	if _, ok := decodeReceipt(data); !ok {
		t.Fatal("valid password failed receipt rejected")
	}
	r.LiveAttempted = false
	data, _ = json.Marshal(r)
	if _, ok := decodeReceipt(data); ok {
		t.Fatal("false attempt accepted")
	}
	r.LiveAttempted = true
	r.Observations.Commands[1] = authobserve.New("profile_show").Snapshot()
	data, _ = json.Marshal(r)
	if _, ok := decodeReceipt(data); ok {
		t.Fatal("wrong command sequence accepted")
	}
}
