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

	"github.com/UnbalancedCat/ipgw-meta/internal/authdev"
	"github.com/UnbalancedCat/ipgw-meta/internal/auxgate"
	"github.com/UnbalancedCat/ipgw-meta/internal/casbundle"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
	"github.com/UnbalancedCat/ipgw-meta/internal/labpath"
	"github.com/UnbalancedCat/ipgw-meta/internal/livegate"
	"github.com/UnbalancedCat/ipgw-meta/internal/wifidev"
	"golang.org/x/term"
)

var workerDigest, inputsDigest, productDigest string
var profilePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

type receipt struct {
	Schema          int                       `json:"schema_version"`
	Kind            string                    `json:"tool_kind"`
	Source          string                    `json:"source_digest"`
	Product         string                    `json:"product_sha256"`
	ReleaseEvidence bool                      `json:"release_evidence"`
	Binding         string                    `json:"binding"`
	State           string                    `json:"state"`
	Reason          string                    `json:"reason"`
	SSID            labpath.WiFiSSID          `json:"wifi_ssid"`
	LiveAttempted   bool                      `json:"live_attempted"`
	Preflight       *wifidev.PreflightReport  `json:"preflight"`
	Auxiliary       []livegate.AuxiliaryCheck `json:"auxiliary_checks"`
	Password        *authdev.Report           `json:"password"`
}

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
	r = receipt{Schema: 1, Kind: "wifi-dev", Source: inputsDigest, Product: productDigest, Binding: "invalid", State: "blocked", Reason: "arguments_rejected", Auxiliary: []livegate.AuxiliaryCheck{}}
	mode := ""
	if len(args) == 1 && args[0] == "--self-test" {
		mode = "self-test"
	}
	if len(args) == 2 && args[0] == "--preflight" && labpath.WiFiSSID(args[1]).Supported() {
		mode = "preflight"
		r.SSID = labpath.WiFiSSID(args[1])
	}
	if len(args) == 3 && args[0] == "--run-once" && labpath.WiFiSSID(args[1]).Supported() && profilePattern.MatchString(args[2]) {
		mode = "live"
		r.SSID = labpath.WiFiSSID(args[1])
	}
	if mode == "" {
		return
	}
	r.Reason = "binding_rejected"
	b, err := casbundle.OpenWiFiDevelopment(filepath.Join(directory, "payload"), workerDigest, inputsDigest, productDigest)
	if err != nil {
		return
	}
	defer b.Close()
	defer func() {
		if b.Verify() != nil {
			r.Binding, r.State, r.Reason = "invalid", "blocked", "binding_rejected"
			r.Password = nil
			r.Auxiliary = nil
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
	r.Reason = "preflight_failed"
	var pre wifidev.PreflightReport
	if !worker(ctx, b.ObserverPath, []string{"--preflight", string(r.SSID)}, &pre) || !pre.Valid(r.SSID) {
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
	r.Reason = "private_terminal_required"
	if !terminal {
		return
	}
	fmt.Fprintln(prompt, "ONE Wi-Fi DEVELOPMENT window, not Candidate evidence. Ethernet disconnected, TUN OFF, selected campus SSID exclusive, target account OFFLINE; keep topology unchanged.")
	fmt.Fprintln(prompt, "First bound status/list/scan; only complete auxiliary pass permits ONE password_core. Password stays in this private terminal. No QR or automatic retry. Default system DNS/PKI may access network or update caches.")
	fmt.Fprintln(prompt, "After exit restore Ethernet for reporting. Hard interruption means account state unknown; never infer logout from restored connectivity.")
	token := "RUN-WIFI-" + strings.ToUpper(string(r.SSID)) + "-ONCE"
	fmt.Fprint(prompt, "Type "+token+" to confirm this exact window: ")
	if !confirm(input, token) {
		r.State, r.Reason = "canceled", "none"
		return
	}
	env, ok := passwordEnvironment()
	if !ok {
		r.Reason = "environment_rejected"
		return
	}
	snapshot, pre := wifidev.Preflight(ctx, r.SSID)
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
	e := &livegate.ProcessExecutor{CandidatePath: b.ProductPath, Profile: args[2], Stdin: input, Stdout: output, Stderr: prompt, BindIP: snapshot.Local, NetworkIndex: int(snapshot.Index), Environment: env, RequireRefuse: true}
	check := func(c context.Context) bool { return b.Verify() == nil && snapshot.Check(c) }
	if !check(ctx) {
		r.Reason = "topology_changed"
		return
	}
	r.LiveAttempted = true
	var result livegate.Result
	r.Auxiliary, result = auxgate.DevelopmentNetwork(ctx, snapshot.Local, e, check)
	if result != livegate.ResultPass {
		r.Reason = "auxiliary_failed"
		return
	}
	guard := livegate.GuardPath(check)
	var prepare error
	if !guard(ctx, func(c context.Context) {
		child, cancel := context.WithTimeout(c, 30*time.Second)
		defer cancel()
		prepare = e.PrepareProfile(child, livegate.SuitePasswordCore)
	}) || prepare != nil {
		r.Reason = "profile_rejected"
		return
	}
	password := authdev.Password(ctx, e, check)
	if !check(ctx) {
		r.Reason = "topology_changed"
		return
	}
	if password.Validate() != nil {
		r.Reason = "report_rejected"
		return
	}
	r.Password = &password
	r.State = password.State
	r.Reason = password.Reason
	return
}

func main() {
	path, err := os.Executable()
	if err != nil {
		os.Exit(2)
	}
	if !protect() {
		fmt.Println(`{"schema_version":1,"tool_kind":"wifi-dev","state":"blocked","reason":"process_protection_unavailable","release_evidence":false}`)
		os.Exit(2)
	}
	timer := time.AfterFunc(14*time.Minute, func() { os.Exit(124) })
	defer timer.Stop()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	r := run(ctx, os.Args[1:], filepath.Dir(path), os.Stdin, os.Stdout, os.Stderr, term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stderr.Fd())))
	data, err := json.Marshal(r)
	if err != nil || len(data) > 64<<10 {
		os.Exit(2)
	}
	fmt.Println(string(data))
	if r.State != "pass" && r.State != "ready" && r.State != "canceled" {
		os.Exit(11)
	}
}
