package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/UnbalancedCat/ipgw-meta/internal/authcontext"
	"github.com/UnbalancedCat/ipgw-meta/internal/casbundle"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
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

var observerDigest, inputsDigest, productDigest string

type receipt struct {
	SchemaVersion  int                 `json:"schema_version"`
	ToolKind       string              `json:"tool_kind"`
	SourceDigest   string              `json:"source_digest"`
	ProductDigest  string              `json:"product_sha256"`
	ObserverDigest string              `json:"observer_sha256"`
	Binding        string              `json:"binding"`
	State          string              `json:"state"`
	LiveAttempted  bool                `json:"live_attempted"`
	Result         *authcontext.Report `json:"result"`
}

var digestShape = regexp.MustCompile(`^[a-f0-9]{64}$`)

func decodeReceipt(data []byte) (receipt, bool) {
	var r receipt
	if casprovenance.DecodeClosedJSON(data, &r, 64<<10) != nil || r.SchemaVersion != 1 || r.ToolKind != "auth-context-dev" {
		return receipt{}, false
	}
	switch r.State {
	case "arguments_rejected", "binding_rejected", "self_test_failed", "confirmation_required", "canceled", "execution_indeterminate", "completed":
	default:
		return receipt{}, false
	}
	if r.Binding != "invalid" && r.Binding != "verified" {
		return receipt{}, false
	}
	if (r.State == "arguments_rejected" || r.State == "binding_rejected") != (r.Binding == "invalid") {
		return receipt{}, false
	}
	if r.Binding == "verified" && (!digestShape.MatchString(r.SourceDigest) || !digestShape.MatchString(r.ProductDigest) || !digestShape.MatchString(r.ObserverDigest)) {
		return receipt{}, false
	}
	if r.Result == nil {
		return r, r.State != "completed" && (!r.LiveAttempted || r.State == "execution_indeterminate" || r.State == "binding_rejected")
	}
	if r.Binding != "verified" || r.State != "completed" || r.Result.Validate() != nil || (r.Result.Mode == "anonymous") != r.LiveAttempted {
		return receipt{}, false
	}
	return r, true
}

type capped struct {
	bytes.Buffer
	limit int
}

func (b *capped) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.Len() {
		return 0, io.ErrShortWrite
	}
	return b.Buffer.Write(p)
}
func execute(ctx context.Context, b *casbundle.AuthDevelopment, mode string) (authcontext.Report, error) {
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
		return authcontext.Report{}, authcontext.ErrReport
	}
	return authcontext.Decode(out.Bytes())
}
func confirm(input io.Reader) bool {
	var line []byte
	for len(line) < 64 {
		one := []byte{0}
		n, e := input.Read(one)
		if n == 1 {
			if one[0] == '\n' {
				return strings.TrimSuffix(string(line), "\r") == "RUN-DEV-CONTEXT-ONCE"
			}
			line = append(line, one[0])
		}
		if e != nil {
			return false
		}
	}
	return false
}
func run(ctx context.Context, args []string, directory string, input io.Reader, prompt io.Writer, terminal bool) (r receipt) {
	r = receipt{SchemaVersion: 1, ToolKind: "auth-context-dev", SourceDigest: inputsDigest, ProductDigest: productDigest, ObserverDigest: observerDigest, Binding: "invalid", State: "arguments_rejected"}
	if len(args) != 1 || args[0] != "--self-test" && args[0] != "--preflight" && args[0] != "--anonymous-once" {
		return
	}
	if args[0] == "--anonymous-once" {
		return r
	} // ADR-0021: retired before any execution.
	r.State = "binding_rejected"
	b, e := casbundle.OpenAuthContext(filepath.Join(directory, "payload"), observerDigest, inputsDigest, productDigest)
	if e != nil {
		return
	}
	defer b.Close()
	defer func() {
		if b.Verify() != nil {
			r.Binding, r.State = "invalid", "binding_rejected"
			r.Result = nil
		}
	}()
	r.Binding = "verified"
	test, e := execute(ctx, b, "--self-test")
	if e != nil || test.Mode != "self-test" || test.State != "pass" {
		r.State = "self_test_failed"
		return
	}
	if args[0] == "--self-test" {
		r.State = "completed"
		r.Result = &test
		return
	}
	if args[0] == "--preflight" {
		result, e := execute(ctx, b, "--local-preflight")
		if e != nil || result.Mode != "preflight" {
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
	fmt.Fprintln(prompt, "ONE anonymous development context observation; not Candidate evidence. Confirm BHKDesktop native Windows, TUN OFF, wired exclusive window, account OFFLINE; topology unchanged. Restore the original Clash TUN manually if needed.")
	fmt.Fprintln(prompt, "32 application GETs / 120 seconds; product system DNS/PKI may fetch certificates or update caches. Same-run CAS Cookies in memory. No extra scripts, profile, credentials, POST, activation or logout. Parser text match is not browser key usage; use_relation stays unknown.")
	fmt.Fprint(prompt, "Type RUN-DEV-CONTEXT-ONCE to authorize exactly this run: ")
	if !confirm(input) {
		r.State = "canceled"
		return
	}
	r.LiveAttempted = true
	result, e := execute(ctx, b, "--anonymous-confirmed")
	if e != nil || result.Mode != "anonymous" {
		r.State = "execution_indeterminate"
		return
	}
	r.State = "completed"
	r.Result = &result
	return
}
func main() {
	path, e := os.Executable()
	if e != nil {
		os.Exit(2)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	r := run(ctx, os.Args[1:], filepath.Dir(path), os.Stdin, os.Stderr, term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stderr.Fd())))
	b, e := json.Marshal(r)
	if e != nil {
		os.Exit(2)
	}
	if _, ok := decodeReceipt(b); !ok {
		os.Exit(2)
	}
	fmt.Println(string(b))
	if r.State == "canceled" || r.State == "completed" && r.Result != nil && (r.Result.State == "pass" || r.Result.State == "ready" || r.Result.State == "observed") {
		return
	}
	os.Exit(11)
}
