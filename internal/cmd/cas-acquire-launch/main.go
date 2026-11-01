package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/UnbalancedCat/ipgw-meta/internal/casacquire"
	"github.com/UnbalancedCat/ipgw-meta/internal/casbundle"
	"golang.org/x/term"
)

// Supplied by the reviewed local snapshot builder; no runtime hash overrides.
var executableDigest, inputsDigest string

type receipt struct {
	SchemaVersion int                `json:"schema_version"`
	ToolKind      string             `json:"tool_kind"`
	SourceDigest  string             `json:"source_digest"`
	Binding       string             `json:"binding"`
	State         string             `json:"state"`
	LiveStarted   bool               `json:"live_started"`
	LiveAttempted bool               `json:"live_attempted"`
	Result        *casacquire.Report `json:"result"`
}

// run is the only CLI entry. The original control flow remains testable with
// synthetic children so historical confirmation and envelope tests retain meaning.
func run(args []string, input io.Reader, terminal bool, prompt io.Writer, directory string) receipt {
	if len(args) == 1 && args[0] == "--run-once" {
		return receipt{SchemaVersion: 1, ToolKind: "cas-acquire-dev", SourceDigest: inputsDigest, Binding: "invalid", State: "arguments_rejected"}
	}
	return runHistorical(args, input, terminal, prompt, directory)
}

func runHistorical(args []string, input io.Reader, terminal bool, prompt io.Writer, directory string) receipt {
	r := receipt{SchemaVersion: 1, ToolKind: "cas-acquire-dev", SourceDigest: inputsDigest, Binding: "invalid", State: "arguments_rejected"}
	if len(args) != 1 || (args[0] != "--self-test" && args[0] != "--run-once" && args[0] != "--preflight" && args[0] != "--cache-check") {
		return r
	}
	r.State = "binding_invalid"
	b, err := casbundle.OpenDevelopment(filepath.Join(directory, "payload"), executableDigest, inputsDigest)
	if err != nil {
		return r
	}
	defer b.Close()
	r.Binding = "verified"
	data, err := b.Execute(context.Background(), "--self-test")
	if err != nil {
		r.State = "self_test_failed"
		return r
	}
	test, err := casacquire.Decode(data)
	if err != nil || test.Mode != "self-test" || test.Collection != "complete" {
		r.State = "self_test_failed"
		return r
	}
	if args[0] == "--self-test" {
		r.State = "self_test_passed"
		r.Result = &test
		return r
	}
	if args[0] == "--preflight" || args[0] == "--cache-check" {
		mode, prefix := "local-preflight", "preflight"
		if args[0] == "--cache-check" {
			mode, prefix = "local-cache-check", "cache_check"
		}
		r.State = prefix + "_failed"
		data, err = b.Execute(context.Background(), "--"+mode)
		if err != nil {
			return r
		}
		result, err := casacquire.Decode(data)
		if err != nil || result.Mode != mode {
			return r
		}
		r.State = prefix + "_completed"
		r.Result = &result
		return r
	}
	r.State = "confirmation_required"
	if !terminal {
		return r
	}
	fmt.Fprintln(prompt, "Single anonymous CAS source diagnostic. Up to 12 GET attempts, 12 DNS UDP sends and 12 TCP connections; 120 seconds maximum network window.")
	fmt.Fprintln(prompt, "No credentials, profile reads, Cookie replay, CAS POST, activation or logout. Output is sanitized classification only.")
	fmt.Fprintln(prompt, "D15: authenticate the two fixed local CTLs before networking; stop on unavailable/ambiguous material. Compare only the same failed peer after socket close. Membership never adds trust or resumes TLS/GET.")
	fmt.Fprintln(prompt, "D9: read seven fixed local policy/AuthRoot certificate sources first; stop without network if unavailable or without comparable material. After a failed connection closes, observe source differences only; no added trust, Windows chain engine or resumed GET.")
	fmt.Fprintln(prompt, "Confirm: BHKDesktop native Windows; TUN is OFF; wired exclusive window; account already OFFLINE; restore TUN manually if needed. Do not change topology during the run.")
	fmt.Fprint(prompt, "To authorize ONE run, type RUN-ANONYMOUS-ONCE (anything else cancels): ")
	line, err := bufio.NewReader(io.LimitReader(input, 64)).ReadString('\n')
	if err != nil || strings.TrimRight(line, "\r\n") != "RUN-ANONYMOUS-ONCE" {
		r.State = "canceled"
		return r
	}
	r.LiveAttempted = true
	r.State = "execution_indeterminate"
	data, r.LiveStarted, err = b.ExecuteObserved(context.Background(), "--anonymous-once-confirmed")
	if err != nil {
		return r
	}
	result, err := casacquire.Decode(data)
	if err != nil || result.Mode != "anonymous-once" {
		return r
	}
	r.State = "completed"
	r.Result = &result
	return r
}
func main() {
	path, err := os.Executable()
	if err != nil {
		os.Exit(2)
	}
	r := run(os.Args[1:], os.Stdin, term.IsTerminal(int(os.Stdin.Fd())), os.Stderr, filepath.Dir(path))
	b, err := json.Marshal(r)
	if err != nil {
		os.Exit(3)
	}
	fmt.Println(string(b))
	if r.State != "completed" && r.State != "self_test_passed" && r.State != "canceled" && r.State != "preflight_completed" && r.State != "cache_check_completed" {
		os.Exit(2)
	}
}
