package livegate

import (
	"bytes"
	"context"
	"net/netip"
	"os"
	"strings"
	"testing"
)

func TestDevelopmentContextIsOptInAndKeepsSecondLoginInputClosed(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		executor, script := newProcessTestExecutor(t, processTestStep{output: processProfileEnvelope(processTestUsername, `{"provider":"prompt"}`)})
		if enabled {
			executor.BindIP = netip.MustParseAddr("192.0.2.10")
			executor.RequireRefuse = true
		}
		if err := executor.PrepareProfile(context.Background(), SuitePasswordCore); err != nil {
			t.Fatal(err)
		}
		prefix := []string{"--json", "--profile", "campus"}
		if enabled {
			prefix = append(prefix, "--bind-ip", "192.0.2.10")
		}
		requireProcessCall(t, executor, script, 0, append(prefix, "profile", "show"), nil, true)
		args, _, stdin, ok := executor.jsonAction(ActionSecondPasswordLogin)
		if !ok || stdin != nil || strings.Join(args, " ") != strings.Join(append(prefix, "login", "--method", "password"), " ") {
			t.Fatal("second login changed")
		}
		_, _, stdin, ok = executor.jsonAction(ActionPasswordLogin)
		if !ok || stdin != executor.Stdin {
			t.Fatal("first login lost private stdin")
		}
	}
}

func TestDevelopmentProcessEnvironmentIsExplicit(t *testing.T) {
	path, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("AUTHDEV_UNWANTED", "synthetic-parent-only")
	var out, stderr bytes.Buffer
	env := []string{"AUTHDEV_ENV_HELPER=1", "SystemRoot=" + os.Getenv("SystemRoot"), "WINDIR=" + os.Getenv("WINDIR")}
	exit, err := runCandidateProcessEnvironment(context.Background(), path, []string{"-test.run=^TestDevelopmentEnvironmentHelper$"}, nil, &out, &stderr, env)
	if err != nil || exit != 0 || strings.TrimSpace(out.String()) != "environment-isolated" || stderr.Len() != 0 {
		t.Fatal("explicit process environment not isolated")
	}
}
func TestDevelopmentEnvironmentHelper(t *testing.T) {
	if os.Getenv("AUTHDEV_ENV_HELPER") != "1" {
		return
	}
	if os.Getenv("AUTHDEV_UNWANTED") != "" {
		os.Exit(2)
	}
	_, _ = os.Stdout.WriteString("environment-isolated\n")
	os.Exit(0)
}
func TestDevelopmentRefusesProfileSwitchAndIPv6(t *testing.T) {
	for _, policy := range []string{`"logout-current"`, `true`, `{}`} {
		output := processSuccessEnvelope("profile.show", `{"profile":{"username":"synthetic","switch":`+policy+`,"credential":{"provider":"prompt"}}}`)
		executor, _ := newProcessTestExecutor(t, processTestStep{output: output})
		executor.RequireRefuse = true
		if executor.PrepareProfile(context.Background(), SuitePasswordCore) == nil {
			t.Fatal("switch accepted")
		}
	}
	executor, script := newProcessTestExecutor(t)
	executor.BindIP = netip.MustParseAddr("::1")
	if executor.PrepareProfile(context.Background(), SuitePasswordCore) == nil || len(script.calls) != 0 {
		t.Fatal("IPv6 accepted")
	}
}
