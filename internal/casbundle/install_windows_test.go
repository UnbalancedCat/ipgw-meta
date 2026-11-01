//go:build windows

package casbundle

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func installFixture(t *testing.T) (string, string, string, string, []byte) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	members := archiveFixture(data)
	z := zipBytes(t, members, zip.Deflate)
	root := canonicalTempDir(t)
	source := filepath.Join(root, "source.zip")
	if err = os.WriteFile(source, z, 0600); err != nil {
		t.Fatal(err)
	}
	return source, filepath.Join(root, "installed"), Digest(z), Digest(members[2].data), z
}
func checkInstallReceipt(t *testing.T, r InstallReceipt) {
	t.Helper()
	b, err := EncodeInstallReceipt(r)
	if err != nil {
		t.Fatalf("invalid receipt: %+v", r)
	}
	if _, err = DecodeInstallReceipt(b); err != nil {
		t.Fatal(err)
	}
}
func TestFreshInstallation(t *testing.T) {
	source, dest, za, ma, original := installFixture(t)
	var done <-chan bool
	r := install(context.Background(), source, dest, za, ma, &installHooks{afterCreate: func() {
		if os.Rename(dest, dest+".moved") == nil {
			t.Fatal("destination unlocked")
		}
		if os.WriteFile(source, []byte("replace"), 0600) == nil {
			t.Fatal("source writable")
		}
	}, beforeSelfTest: func() { done = releaseAfterMarker(t, dest, nil) }})
	checkInstallReceipt(t, r)
	if r.State != "installed" || r.FilesWritten != 4 || r.SelfTest.EXEStarts != 1 {
		t.Fatalf("installation: %s/%s", r.State, r.Reason)
	}
	if done == nil || !<-done {
		t.Fatal("fixture release failed")
	}
	after, err := os.ReadFile(source)
	if err != nil || !bytes.Equal(after, original) {
		t.Fatal("source changed")
	}
	a, err := verifyArchive(context.Background(), original, za, ma)
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range a.files {
		got, err := os.ReadFile(filepath.Join(dest, name))
		if err != nil || !bytes.Equal(got, want) {
			t.Fatal("installed bytes differ")
		}
	}
	again := Install(context.Background(), source, dest, za, ma)
	checkInstallReceipt(t, again)
	if again.State != "rejected" || again.Reason != "destination_invalid" {
		t.Fatal("existing installation replaced")
	}
	for name, want := range a.files {
		got, _ := os.ReadFile(filepath.Join(dest, name))
		if !bytes.Equal(got, want) {
			t.Fatal("repeat modified files")
		}
	}
}
func TestInstallRejectsBeforeCreation(t *testing.T) {
	for _, name := range []string{"archive-anchor", "manifest-anchor", "tampered", "existing-empty", "existing-file", "existing-sentinel", "hardlink", "junction-parent", "junction-source"} {
		t.Run(name, func(t *testing.T) {
			source, dest, za, ma, _ := installFixture(t)
			var sentinel string
			switch name {
			case "archive-anchor":
				za = strings.Repeat("0", 64)
			case "manifest-anchor":
				ma = strings.Repeat("0", 64)
			case "tampered":
				if err := os.WriteFile(source, []byte("replacement"), 0600); err != nil {
					t.Fatal(err)
				}
			case "existing-empty":
				if err := os.Mkdir(dest, 0700); err != nil {
					t.Fatal(err)
				}
			case "existing-file":
				if err := os.WriteFile(dest, []byte("keep"), 0600); err != nil {
					t.Fatal(err)
				}
				sentinel = dest
			case "existing-sentinel":
				if err := os.Mkdir(dest, 0700); err != nil {
					t.Fatal(err)
				}
				sentinel = filepath.Join(dest, "keep")
				if err := os.WriteFile(sentinel, []byte("keep"), 0600); err != nil {
					t.Fatal(err)
				}
			case "hardlink":
				if err := os.Link(source, source+".alias"); err != nil {
					t.Fatal(err)
				}
			case "junction-parent", "junction-source":
				alias := filepath.Join(filepath.Dir(dest), "alias")
				if err := exec.Command("cmd.exe", "/d", "/c", "mklink", "/J", alias, filepath.Dir(dest)).Run(); err != nil {
					t.Fatal("junction fixture failed")
				}
				defer os.Remove(alias)
				if name == "junction-parent" {
					dest = filepath.Join(alias, "installed")
				} else {
					source = filepath.Join(alias, "source.zip")
				}
			}
			r := Install(context.Background(), source, dest, za, ma)
			checkInstallReceipt(t, r)
			if r.State != "rejected" || r.DirectoryCreated || r.FilesWritten != 0 || r.SelfTest != nil {
				t.Fatalf("unexpected install: %+v", r)
			}
			if sentinel != "" {
				b, err := os.ReadFile(sentinel)
				if err != nil || string(b) != "keep" {
					t.Fatal("existing file changed")
				}
			}
			if !strings.HasPrefix(name, "existing-") {
				if _, err := os.Stat(dest); !os.IsNotExist(err) {
					t.Fatal("rejection created destination")
				}
			}
		})
	}
}
func TestInstallPartialFailure(t *testing.T) {
	for _, name := range []string{"collision", "cancel-after-write", "tamper-before-run", "extra-before-run"} {
		t.Run(name, func(t *testing.T) {
			source, dest, za, ma, original := installFixture(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			h := &installHooks{}
			switch name {
			case "collision":
				h.afterCreate = func() {
					if err := os.WriteFile(filepath.Join(dest, names[0]), []byte("keep"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			case "cancel-after-write":
				h.afterWrite = func(n int) {
					if n == 1 {
						cancel()
					}
				}
			case "tamper-before-run":
				h.beforeSelfTest = func() {
					if err := os.WriteFile(filepath.Join(dest, "cas-provenance.exe"), []byte("replacement"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			case "extra-before-run":
				h.beforeSelfTest = func() {
					if err := os.WriteFile(filepath.Join(dest, "unexpected"), []byte("x"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			r := install(ctx, source, dest, za, ma, h)
			checkInstallReceipt(t, r)
			if r.State != "incomplete" || !r.DirectoryCreated || r.SelfTest != nil && (r.SelfTest.EXEAttempts != 0 || r.SelfTest.EXEStarts != 0) {
				t.Fatalf("fault crossed boundary: %+v", r)
			}
			if _, err := os.Stat(filepath.Join(dest, "executed.marker")); !os.IsNotExist(err) {
				t.Fatal("fault executed child")
			}
			if _, err := os.Stat(dest); err != nil {
				t.Fatal("partial directory removed")
			}
			got, _ := os.ReadFile(source)
			if !bytes.Equal(got, original) {
				t.Fatal("source changed")
			}
			if name == "collision" {
				got, _ := os.ReadFile(filepath.Join(dest, names[0]))
				if string(got) != "keep" || r.FilesWritten != 0 {
					t.Fatal("collision overwritten")
				}
			}
			if name == "cancel-after-write" && r.FilesWritten != 1 {
				t.Fatal("inaccurate write count")
			}
		})
	}
}
