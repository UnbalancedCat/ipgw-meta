//go:build windows

package casbundle

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/sys/windows"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
)

// The copied test executable is an inert process fixture, not the shipped
// analyzer. Its marker independently detects any unexpected execution.
func TestMain(m *testing.M) {
	if len(os.Args) == 2 && os.Args[1] == "--self-test" {
		if os.WriteFile("executed.marker", []byte("synthetic"), 0600) != nil {
			os.Exit(3)
		}
		if _, err := os.Stat(filepath.Join("..", "ctl-deadline-fixture")); err == nil {
			// Inert test-only child intentionally outlives the CTL wrapper deadline.
			time.Sleep(time.Minute)
			os.Exit(9)
		}
		deadline := time.Now().Add(10 * time.Second)
		for {
			if _, err := os.Stat(filepath.Join("..", "release.marker")); err == nil {
				break
			}
			if time.Now().After(deadline) {
				os.Exit(4)
			}
			time.Sleep(5 * time.Millisecond)
		}
		r, err := casprovenance.Analyze(context.Background(), []casprovenance.Input{{Kind: "inline_script", Data: []byte(";")}})
		if err != nil {
			os.Exit(5)
		}
		e := casprovenance.SelfTestEnvelope{ToolKind: ToolKind, Mode: "synthetic-prototype", Binding: "not_packaged", SelfTestResult: "passed", Checks: 11, Report: r}
		data, _ := json.Marshal(e)
		fmt.Println(string(data))
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func diskFixture(t *testing.T) (string, string) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "package")
	if err = os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	m, r, b := fixture(data)
	for name, content := range map[string][]byte{"manifest.json": encode(m), "runner.json": encode(r), "build-inputs.json": encode(b), "cas-provenance.exe": data} {
		if err = os.WriteFile(filepath.Join(dir, name), content, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return dir, Digest(encode(m))
}

func releaseAfterMarker(t *testing.T, dir string, check func()) <-chan bool {
	t.Helper()
	done := make(chan bool, 1)
	go func() {
		deadline := time.Now().Add(10 * time.Second)
		for {
			if _, err := os.Stat(filepath.Join(dir, "executed.marker")); err == nil {
				break
			}
			if time.Now().After(deadline) {
				done <- false
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
		if check != nil {
			check()
		}
		err := os.WriteFile(filepath.Join(filepath.Dir(dir), "release.marker"), []byte("synthetic"), 0600)
		done <- err == nil
	}()
	return done
}

func TestProtectedExecutionAndLocks(t *testing.T) {
	dir, anchor := diskFixture(t)
	if probe, err := openLocked(dir); err != nil {
		prefixes, _ := pathPrefixes(dir)
		for i, p := range prefixes {
			f, e := protectedOpen(p, true)
			if e != nil {
				u, _ := windows.UTF16PtrFromString(p)
				h, raw := windows.CreateFile(u, windows.GENERIC_READ, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
				if raw != nil {
					t.Fatalf("directory lock prefix %d: %v", i, raw)
				}
				actual, raw := finalName(h, 0)
				windows.CloseHandle(h)
				t.Fatalf("directory identity prefix %d: canonical_equal=%v error=%v", i, strings.EqualFold(strings.TrimSuffix(actual, `\`), strings.TrimSuffix(`\\?\`+p, `\`)), raw)
			}
			f.Close()
		}
		t.Fatal("bundle member lock or volume identity rejected")
	} else {
		probe.close()
	}
	var done <-chan bool
	attempts := make(chan bool, 1)
	r := run(context.Background(), dir, anchor, func(b *lockedBundle) {
		// Before CreateProcess, the PE loader cannot be responsible for these
		// failures: the verifier's own handles must prevent replacement.
		for _, name := range names {
			p := filepath.Join(dir, name)
			f, err := os.OpenFile(p, os.O_WRONLY, 0)
			if err == nil {
				f.Close()
				t.Fatal("file writable after verification before process start")
			}
			if os.Rename(p, p+".prelaunch") == nil {
				t.Fatal("file replaceable before process start")
			}
		}
		done = releaseAfterMarker(t, dir, func() {
			blocked := true
			for _, name := range names {
				p := filepath.Join(dir, name)
				f, err := os.OpenFile(p, os.O_WRONLY, 0)
				if err == nil {
					blocked = false
					f.Close()
				}
				if os.Rename(p, p+".moved") == nil {
					blocked = false
				}
			}
			if os.Rename(dir, dir+".moved") == nil {
				blocked = false
			}
			if os.Rename(filepath.Dir(dir), filepath.Dir(dir)+".moved") == nil {
				blocked = false
			}
			attempts <- blocked
		})
	})
	if r.Runner != "passed" || r.EXEAttempts != 1 || r.EXEStarts != 1 || r.Binding != "verified" || r.Result == nil {
		t.Fatalf("protected run failed: %s / %s", r.Binding, r.Reason)
	}
	if done == nil || !<-done || !<-attempts {
		t.Fatal("replacement/write was not blocked during child execution")
	}
	// Locks must be released after the child is fully reaped.
	if err := os.Rename(filepath.Join(dir, "runner.json"), filepath.Join(dir, "runner.moved")); err != nil {
		t.Fatal("lock not released")
	}
}

func TestRejectBeforeProcessStart(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(string, string) string
	}{
		{"missing-anchor", func(d, a string) string { return "" }},
		{"wrong-anchor", func(d, a string) string { return strings.Repeat("0", 64) }},
		{"missing-exe", func(d, a string) string { os.Remove(filepath.Join(d, "cas-provenance.exe")); return a }},
		{"tampered-exe", func(d, a string) string {
			os.WriteFile(filepath.Join(d, "cas-provenance.exe"), []byte("synthetic replacement"), 0600)
			return a
		}},
		{"tampered-runner", func(d, a string) string { os.WriteFile(filepath.Join(d, "runner.json"), []byte("{}"), 0600); return a }},
		{"tampered-inputs", func(d, a string) string {
			os.WriteFile(filepath.Join(d, "build-inputs.json"), []byte("{}"), 0600)
			return a
		}},
		{"extra-file", func(d, a string) string { os.WriteFile(filepath.Join(d, "extra"), []byte("synthetic"), 0600); return a }},
		{"directory-member", func(d, a string) string {
			os.Remove(filepath.Join(d, "runner.json"))
			os.Mkdir(filepath.Join(d, "runner.json"), 0700)
			return a
		}},
		{"hard-link", func(d, a string) string {
			if err := os.Link(filepath.Join(d, "cas-provenance.exe"), filepath.Join(filepath.Dir(d), "alias.exe")); err != nil {
				t.Fatal(err)
			}
			return a
		}},
		{"self-consistent-mismatch", func(d, a string) string {
			var r Runner
			var m Manifest
			rb, _ := os.ReadFile(filepath.Join(d, "runner.json"))
			mb, _ := os.ReadFile(filepath.Join(d, "manifest.json"))
			json.Unmarshal(rb, &r)
			json.Unmarshal(mb, &m)
			r.SHA256 = Digest([]byte("older synthetic exe"))
			rb = encode(r)
			m.Files[2] = rec("runner.json", rb)
			mb = encode(m)
			os.WriteFile(filepath.Join(d, "runner.json"), rb, 0600)
			os.WriteFile(filepath.Join(d, "manifest.json"), mb, 0600)
			return Digest(mb)
		}},
		{"rebound-runner-old-anchor", func(d, a string) string {
			mb, _ := os.ReadFile(filepath.Join(d, "manifest.json"))
			os.WriteFile(filepath.Join(d, "manifest.json"), append(mb, ' '), 0600)
			return a
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir, anchor := diskFixture(t)
			anchor = tc.mutate(dir, anchor)
			called := false
			r := run(context.Background(), dir, anchor, func(*lockedBundle) { called = true })
			if r.EXEAttempts != 0 || r.EXEStarts != 0 || r.Binding != "invalid" || r.Result != nil || called {
				t.Fatal("rejection passed execution boundary")
			}
			if _, err := os.Stat(filepath.Join(dir, "executed.marker")); !os.IsNotExist(err) {
				t.Fatal("rejected EXE executed")
			}
		})
	}
}

func TestPathGrammar(t *testing.T) {
	for _, name := range []string{`relative`, `C:relative`, `\\server\share\package`, `\\?\C:\package`, `C:\package\..\other`, `C:\package.`, `C:\package `, `C:\package:stream`, `C:/package`, `C:\package\\child`} {
		if _, err := pathPrefixes(name); err == nil {
			t.Fatal("unsafe path accepted")
		}
	}
}

func TestExistingWriterBlocksValidation(t *testing.T) {
	dir, anchor := diskFixture(t)
	f, err := os.OpenFile(filepath.Join(dir, "runner.json"), os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	r := Run(context.Background(), dir, anchor)
	if r.Binding != "invalid" || r.EXEStarts != 0 {
		t.Fatal("concurrent writer accepted")
	}
}

func TestJunctionParentRejected(t *testing.T) {
	dir, anchor := diskFixture(t)
	alias := filepath.Join(filepath.Dir(dir), "junction")
	output, err := exec.Command("cmd.exe", "/d", "/c", "mklink", "/J", alias, dir).CombinedOutput()
	if err != nil {
		t.Fatalf("synthetic junction creation failed (output bytes %d)", len(output))
	}
	defer os.Remove(alias)
	r := Run(context.Background(), alias, anchor)
	if r.EXEStarts != 0 || r.Binding != "invalid" {
		t.Fatal("reparse parent accepted")
	}
	if _, err = os.Stat(filepath.Join(dir, "executed.marker")); !os.IsNotExist(err) {
		t.Fatal("reparse rejection executed child")
	}
}

func TestCancelChildReapsBeforeUnlock(t *testing.T) {
	dir, anchor := diskFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	seen := make(chan bool, 1)
	r := run(ctx, dir, anchor, func(*lockedBundle) {
		go func() {
			deadline := time.Now().Add(5 * time.Second)
			for {
				if _, err := os.Stat(filepath.Join(dir, "executed.marker")); err == nil {
					seen <- true
					cancel()
					return
				}
				if time.Now().After(deadline) {
					seen <- false
					cancel()
					return
				}
				time.Sleep(5 * time.Millisecond)
			}
		}()
	})
	if r.EXEStarts != 1 || r.Runner != "failed" || r.Reason != "canceled_or_timeout" || !<-seen {
		t.Fatal("child cancellation failed")
	}
	if err := os.Rename(filepath.Join(dir, "cas-provenance.exe"), filepath.Join(dir, "finished.exe")); err != nil {
		t.Fatal("process or lock survived cancellation")
	}
}
