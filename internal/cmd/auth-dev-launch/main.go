package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/UnbalancedCat/ipgw-meta/internal/authdev"
	"github.com/UnbalancedCat/ipgw-meta/internal/authobserve"
	"github.com/UnbalancedCat/ipgw-meta/internal/casbundle"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
	"github.com/UnbalancedCat/ipgw-meta/internal/livegate"
	"golang.org/x/term"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

// Only the snapshot builder supplies these values. No runtime hash override.
var observerDigest, inputsDigest, productDigest, observedDigest string
var profilePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

type receipt struct {
	SchemaVersion  int                `json:"schema_version"`
	ToolKind       string             `json:"tool_kind"`
	SourceDigest   string             `json:"source_digest"`
	ProductDigest  string             `json:"product_sha256"`
	ObservedDigest string             `json:"observed_cli_sha256"`
	Binding        string             `json:"binding"`
	State          string             `json:"state"`
	LiveAttempted  bool               `json:"live_attempted"`
	Result         *authdev.Report    `json:"result"`
	Observations   authobserve.Window `json:"observations"`
}
type capped struct {
	bytes.Buffer
	limit int
}

func observedSelfTest(ctx context.Context, path string) bool {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, "--self-test")
	cmd.Env = []string{"SystemRoot=" + os.Getenv("SystemRoot"), "WINDIR=" + os.Getenv("SystemRoot")}
	out, stderr := &capped{limit: 4096}, &capped{limit: 4096}
	cmd.Stdout, cmd.Stderr = out, stderr
	if cmd.Run() != nil || ctx.Err() != nil || stderr.Len() != 0 {
		return false
	}
	var r struct {
		SchemaVersion int    `json:"schema_version"`
		ToolKind      string `json:"tool_kind"`
		Passed        bool   `json:"passed"`
	}
	return casprovenance.DecodeClosedJSON(out.Bytes(), &r, 4096) == nil && r.SchemaVersion == 1 && r.ToolKind == "auth-observed-selftest" && r.Passed
}

func (b *capped) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.Len() {
		return 0, io.ErrShortWrite
	}
	return b.Buffer.Write(p)
}

func executeObserver(ctx context.Context, b *casbundle.AuthDevelopment, mode string) (authdev.Report, error) {
	timeout := 145 * time.Second
	if mode == "--self-test" {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, b.ObserverPath, mode)
	cmd.Env = []string{"SystemRoot=" + os.Getenv("SystemRoot"), "WINDIR=" + os.Getenv("SystemRoot")}
	out, stderr := &capped{limit: 64 << 10}, &capped{limit: 4096}
	cmd.Stdout, cmd.Stderr = out, stderr
	cmd.WaitDelay = time.Second
	if cmd.Run() != nil || ctx.Err() != nil || stderr.Len() != 0 {
		return authdev.Report{}, authdev.ErrReport
	}
	return authdev.Decode(out.Bytes())
}

// Read exactly the confirmation line, without buffering future password input.
func confirm(input io.Reader, token string) bool {
	var line []byte
	for len(line) < 64 {
		one := make([]byte, 1)
		n, err := input.Read(one)
		if n == 1 {
			if one[0] == '\n' {
				return strings.TrimSuffix(string(line), "\r") == token
			}
			line = append(line, one[0])
		}
		if err != nil {
			return false
		}
	}
	return false
}

func passwordEnvironment() ([]string, bool) {
	var env []string
	for _, key := range []string{"SystemRoot", "WINDIR", "USERPROFILE", "APPDATA", "LOCALAPPDATA", "TEMP", "TMP"} {
		value := os.Getenv(key)
		if value == "" || !filepath.IsAbs(value) {
			return nil, false
		}
		env = append(env, key+"="+value)
	}
	return env, true
}

func run(ctx context.Context, args []string, directory string, input, output, prompt *os.File, terminal bool) (r receipt) {
	r = receipt{SchemaVersion: 3, ToolKind: "auth-dev-observed", SourceDigest: inputsDigest, ProductDigest: productDigest, ObservedDigest: observedDigest, Binding: "invalid", State: "arguments_rejected", Observations: authobserve.NewWindow()}
	mode := ""
	if len(args) == 1 {
		switch args[0] {
		case "--self-test":
			mode = "self-test"
		case "--preflight":
			mode = "preflight"
		case "--anonymous-once":
			return r // ADR-0021 retires the anonymous live mode.
		}
	}
	if len(args) == 2 && args[0] == "--password-core" && profilePattern.MatchString(args[1]) {
		mode = "password"
	}
	if mode == "" {
		return
	}
	r.State = "binding_rejected"
	b, err := casbundle.OpenAuthObserved(filepath.Join(directory, "payload"), observerDigest, observedDigest, inputsDigest, productDigest)
	if err != nil {
		return
	}
	defer b.Close()
	defer func() {
		if b.Verify() != nil {
			r.Binding = "invalid"
			r.State = "binding_rejected"
			r.Result = nil
			r.Observations = authobserve.NewWindow()
		}
	}()
	r.Binding = "verified"
	test, err := executeObserver(ctx, b, "--self-test")
	if err != nil || test.Mode != "self-test" || test.State != "pass" {
		r.State = "self_test_failed"
		return
	}
	if !observedSelfTest(ctx, b.ObservedPath) {
		r.State = "self_test_failed"
		return
	}
	if mode == "self-test" {
		r.State = "completed"
		r.Result = &test
		return
	}
	if mode == "preflight" {
		result, err := executeObserver(ctx, b, "--local-preflight")
		if err != nil || result.Mode != "preflight" {
			r.State = "execution_indeterminate"
			return
		}
		r.State = "completed"
		r.Result = &result
		return
	}
	r.State = "confirmation_required"
	if !terminal {
		return
	}
	fmt.Fprintln(prompt, "ONE development validation; not Candidate evidence. Confirm BHKDesktop native Windows, TUN OFF, wired exclusive window, account OFFLINE; keep topology unchanged. Restore the original Clash TUN manually if needed.")
	fmt.Fprintln(prompt, "Default product system DNS/PKI may fetch certificates or update caches; application counts do not measure all OS traffic.")
	token := "RUN-DEV-ANONYMOUS-ONCE"
	if mode == "anonymous" {
		fmt.Fprintln(prompt, "SDK GET prelude only: 32 application requests, 120 seconds, same-run in-memory CAS Cookies; no profile, credential, POST, activation or logout. Stop before credentials.")
	} else {
		token = "RUN-DEV-PASSWORD-ONCE"
		fmt.Fprintln(prompt, "Read the selected prompt profile; password is entered only in this private terminal. Product password_core: seven primary commands, up to two owned cleanup commands, 10-minute cooperative window. No QR or account switching. No HTTP/DNS/TCP total is claimed.")
		fmt.Fprintln(prompt, "Independent observed CLI, not ordinary product bytes: passive request roles, HTTP statuses and fixed error categories; bounded in-memory logout response structure only, no raw body/header/query logging. Original SDK, prompt, system TLS and cleanup rules remain in use.")
	}
	fmt.Fprint(prompt, "Type "+token+" to authorize exactly this run: ")
	if !confirm(input, token) {
		r.State = "canceled"
		return
	}
	if mode == "anonymous" {
		r.LiveAttempted = true
		result, err := executeObserver(ctx, b, "--anonymous-confirmed")
		if err != nil || result.Mode != "anonymous" {
			r.State = "execution_indeterminate"
			return
		}
		r.State = "completed"
		r.Result = &result
		return
	}
	result := authdev.NewReport("password")
	r.State = "completed"
	r.Result = &result
	env, ok := passwordEnvironment()
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	topology, err := authdev.ReadTopology(ctx)
	if err != nil {
		return
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return
	}
	unlock, err := casbundle.LockAuthProfile(filepath.Join(configDir, "ipgw-meta", "config.yaml"))
	if err != nil {
		result.Reason = "profile_rejected"
		return
	}
	defer unlock()
	adapter := &authobserve.Adapter{Environment: env, Window: authobserve.NewWindow()}
	defer func() { r.Observations = adapter.Window }()
	executor := &livegate.ProcessExecutor{CandidatePath: b.ObservedPath, Profile: args[1], Stdin: input, Stdout: output, Stderr: prompt, BindIP: topology.Local, Environment: env, RequireRefuse: true, ProcessAdapter: adapter.Run}
	preparation, done := context.WithTimeout(ctx, 30*time.Second)
	err = executor.PrepareProfile(preparation, livegate.SuitePasswordCore)
	done()
	result.Commands = 1
	if err != nil {
		result.Reason = "profile_rejected"
		return
	}
	r.LiveAttempted = true
	result = authdev.Password(ctx, executor, topology.Check)
	result.Commands++ // bounded local profile show; no profile data in report
	// A partial state-machine prefix does not provide a complete alignment
	// between child calls and published steps. Never invent that alignment.
	if result.StepCoverage != "complete" || result.Reason == "topology_changed" {
		if adapter.Window.State != "incomplete" {
			adapter.Window.Incomplete("interrupted")
		}
	}
	return
}
func main() {
	path, err := os.Executable()
	if err != nil {
		os.Exit(2)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	r := run(ctx, os.Args[1:], filepath.Dir(path), os.Stdin, os.Stdout, os.Stderr, term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stderr.Fd())))
	if r.Result != nil && r.Result.Validate() != nil {
		r.Result = nil
		r.State = "report_rejected"
	}
	if r.Observations.Validate() != nil {
		r.Result = nil
		r.State = "report_rejected"
		r.Observations = authobserve.NewWindow()
	}
	data, err := json.Marshal(r)
	if err != nil || len(data) > 64<<10 {
		os.Exit(2)
	}
	if _, ok := decodeReceipt(data); !ok {
		os.Exit(2)
	}
	fmt.Println(string(data))
	if r.State == "completed" && r.Result != nil && (r.Result.State == "pass" || r.Result.State == "ready" || r.Result.State == "observed") && (r.Result.Mode != "password" || r.Observations.State == "complete") || r.State == "canceled" {
		return
	}
	os.Exit(11)
}
