//go:build windows

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/UnbalancedCat/ipgw-meta/internal/casbundle"
	"github.com/UnbalancedCat/ipgw-meta/internal/trustaudit"
)

func TestMain(m *testing.M) {
	if len(os.Args) == 2 && (os.Args[1] == "--self-test" || os.Args[1] == "--trust-audit") {
		// Inert child: both modes use generated data, never the real store reader.
		r := trustaudit.SelfTest(context.Background())
		if os.Args[1] == "--trust-audit" {
			r.Mode = "local-trust-audit"
			if os.WriteFile(filepath.Join("..", "synthetic-audit.marker"), []byte("fixture-only"), 0600) != nil {
				os.Exit(3)
			}
		}
		b, err := trustaudit.Encode(r)
		if err != nil {
			os.Exit(3)
		}
		os.Stdout.Write(b)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestBoundSyntheticAuditChild(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	payload := filepath.Join(dir, "payload")
	if os.Mkdir(payload, 0700) != nil {
		t.Fatal("directory")
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	inputs := []byte("synthetic-inputs")
	if os.WriteFile(filepath.Join(payload, "trust-audit.exe"), data, 0600) != nil || os.WriteFile(filepath.Join(payload, "build-inputs.json"), inputs, 0600) != nil {
		t.Fatal("fixture write")
	}
	oldExe, oldInputs := executableDigest, inputsDigest
	executableDigest, inputsDigest = casbundle.Digest(data), casbundle.Digest(inputs)
	defer func() { executableDigest, inputsDigest = oldExe, oldInputs }()
	for _, mode := range []string{"--run-once", "--anonymous-once-confirmed", "--cache-check"} {
		if r := run([]string{mode}, dir); r.State != "arguments_rejected" {
			t.Fatal("live mode accepted")
		}
	}
	if r := run([]string{"--self-test"}, dir); r.State != "self_test_passed" {
		t.Fatal("self test")
	}
	marker := filepath.Join(dir, "synthetic-audit.marker")
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("audit started early")
	}
	r := run([]string{"--trust-audit"}, dir)
	if r.State != "audit_completed" || r.Binding != "verified" || r.Result == nil || r.Result.Mode != "local-trust-audit" || r.Result.CA.ParseCoverage != "partial" || r.Result.Operations != (trustaudit.Operations{}) {
		t.Fatal("synthetic audit result lost")
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("synthetic audit did not start")
	}
}
