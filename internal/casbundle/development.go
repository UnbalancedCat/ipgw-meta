package casbundle

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

// Development holds the two-file acquisition bundle until Close. It is a new
// identity: the historical four-file provenance runner is not reused or widened.
type Development struct {
	bundle    *lockedBundle
	directory string
}

func OpenDevelopment(directory, exeDigest, inputsDigest string) (*Development, error) {
	return openDevelopmentBound(directory, exeDigest, inputsDigest, "cas-acquire.exe")
}

func openDevelopmentBound(directory, exeDigest, inputsDigest, executable string) (*Development, error) {
	if !digestPattern.MatchString(exeDigest) || !digestPattern.MatchString(inputsDigest) {
		return nil, ErrInvalid
	}
	b, err := openDevelopmentNamed(directory, executable)
	if err != nil {
		return nil, err
	}
	fail := func() (*Development, error) { b.close(); return nil, ErrInvalid }
	exe, err := b.read(executable, MaxExecutable)
	if err != nil || Digest(exe) != exeDigest {
		return fail()
	}
	inputs, err := b.read("build-inputs.json", MaxMetadata)
	if err != nil || Digest(inputs) != inputsDigest {
		return fail()
	}
	return &Development{bundle: b, directory: directory}, nil
}
func (d *Development) Close() {
	if d != nil && d.bundle != nil {
		d.bundle.close()
		d.bundle = nil
	}
}

// Execute accepts only the acquisition command's reviewed operations.
// Error output is never returned; a failed live child has indeterminate counts.
func (d *Development) Execute(ctx context.Context, mode string) ([]byte, error) {
	data, _, err := d.ExecuteObserved(ctx, mode)
	return data, err
}

// ExecuteObserved distinguishes a launch attempt from successful process start.
func (d *Development) ExecuteObserved(ctx context.Context, mode string) ([]byte, bool, error) {
	if d == nil || d.bundle == nil || ctx == nil || (mode != "--self-test" && mode != "--anonymous-once-confirmed" && mode != "--local-preflight" && mode != "--local-cache-check") {
		return nil, false, ErrInvalid
	}
	timeout := 145 * time.Second
	if mode == "--local-cache-check" {
		timeout = 15 * time.Second
	}
	return d.executeBounded(ctx, mode, timeout)
}

func (d *Development) executeBounded(ctx context.Context, mode string, timeout time.Duration) ([]byte, bool, error) {
	child, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	out := &boundedOutput{limit: 65536, cancel: cancel}
	errOut := &boundedOutput{limit: 4096, cancel: cancel}
	cmd := exec.CommandContext(child, d.bundle.executable, mode)
	cmd.Env = childEnvironment()
	cmd.Dir = d.directory
	cmd.Stdin = bytes.NewReader(nil)
	cmd.Stdout, cmd.Stderr = out, errOut
	cmd.WaitDelay = time.Second
	if cmd.Start() != nil {
		return nil, false, ErrInvalid
	}
	if cmd.Wait() != nil || child.Err() != nil || out.overflow || errOut.overflow || len(errOut.data) != 0 {
		return nil, true, ErrInvalid
	}
	return out.data, true, nil
}
