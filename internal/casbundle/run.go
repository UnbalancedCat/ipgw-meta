package casbundle

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
)

type Receipt struct {
	SchemaVersion int                             `json:"schema_version"`
	ToolKind      string                          `json:"tool_kind"`
	Binding       string                          `json:"binding"`
	BundleID      *string                         `json:"bundle_id"`
	Runner        string                          `json:"runner"`
	Reason        string                          `json:"reason"`
	EXEAttempts   int                             `json:"exe_attempts"`
	EXEStarts     int                             `json:"exe_starts"`
	Operations    casprovenance.Operations        `json:"operations"`
	Result        *casprovenance.SelfTestEnvelope `json:"result"`
}

type boundedOutput struct {
	data     []byte
	limit    int
	overflow bool
	cancel   context.CancelFunc
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	if len(p) > b.limit-len(b.data) {
		b.overflow = true
		b.cancel()
		return 0, io.ErrShortWrite
	}
	b.data = append(b.data, p...)
	return len(p), nil
}

// Run starts only a fully verified, locked EXE through the declarative runner.
func Run(ctx context.Context, directory, expected string) Receipt {
	return run(ctx, directory, expected, nil)
}

// afterVerification is test-only instrumentation, never selectable by CLI/env.
func run(ctx context.Context, directory, expected string, afterVerification func(*lockedBundle)) Receipt {
	out := Receipt{SchemaVersion: 1, ToolKind: ToolKind, Binding: "invalid", Runner: "not_started", Reason: "binding_invalid"}
	if ctx == nil || ctx.Err() != nil {
		out.Reason = "canceled"
		return out
	}
	if !digestPattern.MatchString(expected) {
		return out
	}
	b, err := openLocked(directory)
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			out.Reason = "unsupported_platform"
		}
		return out
	}
	defer b.close()
	manifest, err := b.read("manifest.json", MaxMetadata)
	if err != nil {
		return out
	}
	if Digest(manifest) != expected {
		return out
	}
	runner, err := b.read("runner.json", MaxMetadata)
	if err != nil {
		return out
	}
	inputs, err := b.read("build-inputs.json", MaxMetadata)
	if err != nil {
		return out
	}
	m, r, err := ValidateMetadata(manifest, runner, inputs, expected)
	if err != nil {
		return out
	}
	exe, err := b.read("cas-provenance.exe", MaxExecutable)
	if err != nil || int64(len(exe)) != r.Size || Digest(exe) != r.SHA256 {
		return out
	}
	if afterVerification != nil {
		afterVerification(b)
	}
	if ctx.Err() != nil {
		out.Reason = "canceled"
		return out
	}
	out.Binding = "verified"
	out.BundleID = &m.BundleID
	out.Runner = "failed"
	out.Reason = "execution_failed"
	child, cancel := context.WithTimeout(ctx, time.Duration(r.TimeoutSeconds)*time.Second)
	defer cancel()
	stdout := &boundedOutput{limit: r.MaxOutputBytes, cancel: cancel}
	stderr := &boundedOutput{limit: r.MaxOutputBytes, cancel: cancel}
	cmd := exec.CommandContext(child, b.executable, r.Args...)
	cmd.Env = childEnvironment()
	cmd.Dir = directory
	cmd.Stdin = bytes.NewReader(nil)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.WaitDelay = time.Second
	out.EXEAttempts = 1
	if err = cmd.Start(); err != nil {
		return out
	}
	out.EXEStarts = 1
	err = cmd.Wait() // Locks remain held through process termination and pipe completion.
	if stdout.overflow || stderr.overflow {
		out.Reason = "output_limit"
		return out
	}
	if child.Err() != nil {
		out.Reason = "canceled_or_timeout"
		return out
	}
	if err != nil || len(stderr.data) != 0 {
		return out
	}
	result, err := casprovenance.DecodeSelfTest(stdout.data)
	if err != nil {
		out.Reason = "output_invalid"
		return out
	}
	out.Runner = "passed"
	out.Reason = "none"
	out.Result = &result
	return out
}
