package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/casbundle"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
	"github.com/UnbalancedCat/ipgw-meta/internal/livegate"
	"github.com/UnbalancedCat/ipgw-meta/internal/qrdev"
	"golang.org/x/term"
)

var workerDigest, inputsDigest, productDigest string
var profilePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

type receipt = qrdev.Receipt

type capped struct{ bytes.Buffer }

func (b *capped) Write(p []byte) (int, error) {
	if len(p) > 64<<10-b.Len() {
		return 0, io.ErrShortWrite
	}
	return b.Buffer.Write(p)
}

func worker(ctx context.Context, path string, args []string, output any) bool {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = []string{"SystemRoot=" + os.Getenv("SystemRoot"), "WINDIR=" + os.Getenv("SystemRoot")}
	out, errout := &capped{}, &capped{}
	cmd.Stdout, cmd.Stderr = out, errout
	cmd.WaitDelay = time.Second
	if cmd.Run() != nil || ctx.Err() != nil || errout.Len() != 0 {
		return false
	}
	return casprovenance.DecodeClosedJSON(out.Bytes(), output, 64<<10) == nil
}

func confirm(input io.Reader, token string) bool {
	var line []byte
	for len(line) < 64 {
		var b [1]byte
		n, err := input.Read(b[:])
		if n == 1 {
			if b[0] == '\n' {
				return strings.TrimSuffix(string(line), "\r") == token
			}
			line = append(line, b[0])
		}
		if err != nil {
			return false
		}
	}
	return false
}

func passwordEnvironment() ([]string, bool) {
	var env []string
	for _, k := range []string{"SystemRoot", "WINDIR", "USERPROFILE", "APPDATA", "LOCALAPPDATA", "TEMP", "TMP"} {
		v := os.Getenv(k)
		if !filepath.IsAbs(v) {
			return nil, false
		}
		env = append(env, k+"="+v)
	}
	return env, true
}

func run(ctx context.Context, args []string, directory string, input, output, prompt *os.File, terminal bool) (r receipt) {
	r = receipt{Schema: 2, Kind: "qr-dev", Source: inputsDigest, Product: productDigest, Tool: workerDigest, Binding: "invalid", State: "blocked", Reason: "arguments_rejected"}
	mode := ""
	if len(args) == 1 && args[0] == "--self-test" {
		mode = "self-test"
	}
	if len(args) == 1 && args[0] == "--preflight" {
		mode = "preflight"
	}
	if len(args) == 2 && args[0] == "--run-once" && profilePattern.MatchString(args[1]) {
		mode = "live"
	}
	if mode == "" {
		return
	}
	r.Reason = "binding_rejected"
	b, err := casbundle.OpenQRDevelopment(filepath.Join(directory, "payload"), workerDigest, inputsDigest, productDigest)
	if err != nil {
		return
	}
	defer b.Close()
	var gate *qrdev.GateTracker
	defer func() {
		if b.Verify() != nil {
			if gate != nil {
				if r.LiveAttempted {
					gate.BindingRejected("final_check", "final")
				} else {
					gate.BindingRejected("initial_check", "before")
				}
			}
			r.Binding, r.State, r.Reason = "invalid", "blocked", "binding_rejected"
			r.QR = nil
		}
		if gate != nil {
			r.GateFailure = gate.Failure()
		}
	}()
	r.Binding = "verified"
	var test struct {
		Passed bool `json:"passed"`
	}
	if !worker(ctx, b.ObserverPath, []string{"--self-test"}, &test) || !test.Passed {
		r.Reason = "self_test_failed"
		return
	}
	if mode == "self-test" {
		r.State, r.Reason = "pass", "none"
		return
	}
	if mode == "live" && !terminal {
		r.Reason = "private_terminal_required"
		return
	}
	r.Reason = "preflight_failed"
	var pre qrdev.PreflightReport
	if !worker(ctx, b.ObserverPath, []string{"--preflight"}, &pre) || !pre.Valid() {
		return
	}
	r.Preflight = &pre
	if pre.State != "ready" {
		r.Reason = pre.Reason
		return
	}
	if mode == "preflight" {
		r.State, r.Reason = "ready", "none"
		return
	}
	fmt.Fprintln(prompt, "ONE wired QR DEVELOPMENT window, not Candidate evidence. TUN OFF, Wi-Fi disconnected, target account OFFLINE; keep topology unchanged.")
	fmt.Fprintln(prompt, "One QR login, target status, logout and final offline. QR stays in this private terminal. No retry. Default system DNS/PKI may access network or update caches.")
	fmt.Fprintln(prompt, "Ten minutes total; QR command up to five minutes. Cleanup only with established ownership and unchanged topology. After exit restore original Clash configuration if needed; hard interruption means session unknown.")
	token := "RUN-DEV-QR-ONCE"
	fmt.Fprint(prompt, "Type "+token+" to confirm this exact window: ")
	if !confirm(input, token) {
		r.State, r.Reason = "canceled", "none"
		return
	}
	ctx, cancelWindow := context.WithTimeout(ctx, 10*time.Minute)
	defer cancelWindow()
	env, ok := passwordEnvironment()
	if !ok {
		r.Reason = "environment_rejected"
		return
	}
	snapshot, pre := qrdev.Preflight(ctx)
	r.Preflight = &pre
	if pre.State != "ready" {
		r.Reason = pre.Reason
		return
	}
	config, err := os.UserConfigDir()
	if err != nil {
		r.Reason = "profile_rejected"
		return
	}
	unlock, err := casbundle.LockAuthProfile(filepath.Join(config, "ipgw-meta", "config.yaml"))
	if err != nil {
		r.Reason = "profile_rejected"
		return
	}
	defer unlock()
	e := &livegate.ProcessExecutor{CandidatePath: b.ProductPath, Profile: args[1], Stdin: input, Stdout: output, Stderr: prompt, BindIP: snapshot.Local, NetworkIndex: int(snapshot.Index), Environment: env, RequireRefuse: true}
	gate = qrdev.NewGateTracker(b.Verify, snapshot.Inspect)
	check := gate.Check
	if !gate.At(ctx, "initial_check", "before") {
		r.Reason = "topology_changed"
		return
	}
	guard := livegate.GuardPathObserved(check, func(phase string) { gate.Reject("profile_show", phase) })
	var prepare error
	if !guard(ctx, func(c context.Context) {
		child, cancel := context.WithTimeout(c, 30*time.Second)
		defer cancel()
		prepare = e.PrepareProfile(child, livegate.SuiteTerminalQR)
	}) || prepare != nil {
		r.Reason = "profile_rejected"
		return
	}
	r.LiveAttempted = true
	result := qrdev.RunObserved(ctx, e, check, func(action livegate.CommandAction, phase string) { gate.Reject(string(action), phase) })
	finish(&r, result, ctx.Err(), func() bool { return gate.At(ctx, "final_check", "final") })
	return
}

// Preserve valid command facts on cancellation or topology loss. File drift
// is independently rejected by the enclosing bundle verification defer.
func finish(r *receipt, result qrdev.Report, contextErr error, check func() bool) {
	if result.Validate() != nil {
		r.Reason = "report_rejected"
		return
	}
	r.QR = &result
	r.State = result.State
	r.Reason = result.Reason
	if contextErr != nil {
		r.State, r.Reason = "blocked", "canceled_or_timeout"
		return
	}
	if !check() {
		r.State, r.Reason = "blocked", "topology_changed"
	}
}

func main() {
	path, err := os.Executable()
	if err != nil {
		os.Exit(2)
	}
	if !protect() {
		fmt.Println(`{"schema_version":1,"tool_kind":"qr-dev","state":"blocked","reason":"process_protection_unavailable","release_evidence":false}`)
		os.Exit(2)
	}
	timer := time.AfterFunc(14*time.Minute, func() { os.Exit(124) })
	defer timer.Stop()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	r := run(ctx, os.Args[1:], filepath.Dir(path), os.Stdin, os.Stdout, os.Stderr, term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stderr.Fd())))
	if r.Validate() != nil {
		r.State, r.Reason = "blocked", "report_rejected"
		r.QR = nil
		r.GateFailure = nil
	}
	data, err := json.Marshal(r)
	if err != nil || len(data) > 64<<10 {
		os.Exit(2)
	}
	fmt.Println(string(data))
	if r.State != "pass" && r.State != "ready" && r.State != "canceled" {
		os.Exit(11)
	}
}
