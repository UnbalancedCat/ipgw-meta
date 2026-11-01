//go:build windows

package casbundle

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func developmentFixture(t *testing.T) (string, string, string) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(canonicalTempDir(t), "payload")
	if os.Mkdir(dir, 0700) != nil {
		t.Fatal("fixture")
	}
	if os.WriteFile(filepath.Join(dir, "cas-acquire.exe"), data, 0600) != nil || os.WriteFile(filepath.Join(dir, "build-inputs.json"), []byte("synthetic-inputs"), 0600) != nil {
		t.Fatal("fixture write")
	}
	return dir, Digest(data), Digest([]byte("synthetic-inputs"))
}
func TestDevelopmentExactSetAndAnchors(t *testing.T) {
	for _, name := range []string{"exe-hash", "inputs-hash", "missing", "extra", "mixed", "hardlink"} {
		t.Run(name, func(t *testing.T) {
			dir, ed, id := developmentFixture(t)
			switch name {
			case "exe-hash":
				ed = strings.Repeat("0", 64)
			case "inputs-hash":
				id = strings.Repeat("0", 64)
			case "missing":
				if os.Remove(filepath.Join(dir, "build-inputs.json")) != nil {
					t.Fatal("fixture")
				}
			case "extra":
				if os.WriteFile(filepath.Join(dir, "runner.json"), []byte("old"), 0600) != nil {
					t.Fatal("fixture")
				}
			case "mixed":
				if os.WriteFile(filepath.Join(dir, "build-inputs.json"), []byte("other-snapshot"), 0600) != nil {
					t.Fatal("fixture")
				}
			case "hardlink":
				if os.Link(filepath.Join(dir, "cas-acquire.exe"), filepath.Join(filepath.Dir(dir), "alias.exe")) != nil {
					t.Fatal("fixture link")
				}
			}
			b, err := OpenDevelopment(dir, ed, id)
			if err == nil {
				b.Close()
				t.Fatal("unverified bundle accepted")
			}
			if _, err = os.Stat(filepath.Join(dir, "executed.marker")); !os.IsNotExist(err) {
				t.Fatal("unexpected child start")
			}
		})
	}
}
func TestDevelopmentLocksAndProcess(t *testing.T) {
	dir, ed, id := developmentFixture(t)
	b, err := OpenDevelopment(dir, ed, id)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	check := func() bool {
		for _, n := range []string{"cas-acquire.exe", "build-inputs.json"} {
			p := filepath.Join(dir, n)
			if os.WriteFile(p, []byte("bad"), 0600) == nil || os.Rename(p, p+".moved") == nil {
				return false
			}
		}
		return os.Rename(dir, dir+".moved") != nil
	}
	if !check() {
		t.Fatal("unlocked before execute")
	}
	done := releaseAfterMarker(t, dir, func() {
		if !check() {
			t.Error("unlocked during child")
		}
	})
	data, err := b.Execute(context.Background(), "--self-test")
	if err != nil || len(data) == 0 || !<-done {
		t.Fatal("synthetic process failed")
	}
	b.Close()
	if os.WriteFile(filepath.Join(dir, "build-inputs.json"), []byte("released"), 0600) != nil {
		t.Fatal("locks not released")
	}
}
func TestDevelopmentCancelAndMode(t *testing.T) {
	dir, ed, id := developmentFixture(t)
	b, err := OpenDevelopment(dir, ed, id)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if _, err = b.Execute(context.Background(), "--arbitrary"); err == nil {
		t.Fatal("unknown operation")
	}
	if _, err = os.Stat(filepath.Join(dir, "executed.marker")); !os.IsNotExist(err) {
		t.Fatal("unknown operation started")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if _, err = b.Execute(ctx, "--self-test"); err == nil {
		t.Fatal("cancel ignored")
	}
	b.Close()
	if os.WriteFile(filepath.Join(dir, "cas-acquire.exe"), []byte("released"), 0600) != nil {
		t.Fatal("child or lock remains")
	}
}

func TestDevelopmentReportsActualStart(t *testing.T) {
	dir, ed, id := developmentFixture(t)
	b, err := OpenDevelopment(dir, ed, id)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, started, err := b.ExecuteObserved(ctx, "--self-test"); err == nil || started {
		t.Fatal("canceled launch claimed started")
	}
	b.bundle.executable = filepath.Join(dir, "absent.exe") // Inert test-only fault.
	if _, started, err := b.ExecuteObserved(context.Background(), "--self-test"); err == nil || started {
		t.Fatal("failed CreateProcess claimed started")
	}
	if _, err := os.Stat(filepath.Join(dir, "executed.marker")); !os.IsNotExist(err) {
		t.Fatal("unexpected fixture execution")
	}
}
