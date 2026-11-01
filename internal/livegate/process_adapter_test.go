package livegate

import (
	"context"
	"fmt"
	"net/netip"
	"testing"
)

func TestDevelopmentProcessAdapterPreservesProjectionAndTTY(t *testing.T) {
	executor, script := newProcessTestExecutor(t,
		processTestStep{output: processProfileEnvelope(processTestUsername, `{"provider":"prompt"}`)},
		processTestStep{output: processSuccessEnvelope("login", fmt.Sprintf(`{"outcome":"logged_in","status":{"session":"online","username":%q}}`, processTestUsername))},
		processTestStep{output: processSuccessEnvelope("login", fmt.Sprintf(`{"outcome":"already_online","status":{"session":"online","username":%q}}`, processTestUsername))},
	)
	executor.run = nil
	executor.ProcessAdapter = script.run
	executor.BindIP = netip.MustParseAddr("192.0.2.9")
	executor.RequireRefuse = true
	if err := executor.PrepareProfile(context.Background(), SuitePasswordCore); err != nil {
		t.Fatal(err)
	}
	first := executor.Execute(context.Background(), ActionPasswordLogin)
	second := executor.Execute(context.Background(), ActionSecondPasswordLogin)
	if first.RunnerFault || second.RunnerFault || !first.IdentityMatches || !second.IdentityMatches {
		t.Fatal("projection changed")
	}
	if script.calls[1].stdin != executor.Stdin || script.calls[2].stdin != nil || script.calls[1].stderr != executor.Stderr || script.calls[1].stdout == executor.Stdout {
		t.Fatal("private IO contract changed")
	}
	for _, call := range script.calls {
		found := false
		for i, v := range call.args {
			if v == "--bind-ip" && i+1 < len(call.args) && call.args[i+1] == "192.0.2.9" {
				found = true
			}
		}
		if !found {
			t.Fatal("lost bind")
		}
	}
}
