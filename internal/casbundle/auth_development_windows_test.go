//go:build windows

package casbundle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func authFixture(t *testing.T) (string, [3]string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "payload")
	if os.Mkdir(dir, 0700) != nil {
		t.Fatal("fixture")
	}
	var hashes [3]string
	for i, name := range authNames {
		data := []byte("synthetic-" + name)
		if os.WriteFile(filepath.Join(dir, name), data, 0600) != nil {
			t.Fatal("write fixture")
		}
		hashes[i] = Digest(data)
	}
	return dir, hashes
}
func TestAuthDevelopmentRejectsMixedInputs(t *testing.T) {
	for _, kind := range []string{"observer", "inputs", "product", "extra", "missing", "hardlink"} {
		t.Run(kind, func(t *testing.T) {
			dir, h := authFixture(t)
			switch kind {
			case "observer":
				h[0] = strings.Repeat("0", 64)
			case "inputs":
				h[1] = strings.Repeat("0", 64)
			case "product":
				h[2] = strings.Repeat("0", 64)
			case "extra":
				if os.WriteFile(filepath.Join(dir, "extra"), []byte("x"), 0600) != nil {
					t.Fatal("fixture")
				}
			case "missing":
				if os.Remove(filepath.Join(dir, authNames[0])) != nil {
					t.Fatal("fixture")
				}
			case "hardlink":
				if os.Link(filepath.Join(dir, authNames[2]), filepath.Join(filepath.Dir(dir), "alias")) != nil {
					t.Fatal("fixture")
				}
			}
			b, err := OpenAuthDevelopment(dir, h[0], h[1], h[2])
			if err == nil {
				b.Close()
				t.Fatal("unbound inputs accepted")
			}
		})
	}
}
func TestAuthDevelopmentHoldsFilesAndRechecks(t *testing.T) {
	dir, h := authFixture(t)
	b, err := OpenAuthDevelopment(dir, h[0], h[1], h[2])
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if !strings.HasPrefix(b.ProductPath, `\\?\Volume{`) || b.Verify() != nil || b.Verify() != nil {
		t.Fatal("anchor or rewind failed")
	}
	if err := os.WriteFile(filepath.Join(dir, authNames[2]), []byte("tamper"), 0600); err == nil {
		t.Fatal("locked product writable")
	}
	if err := os.Rename(dir, dir+"-renamed"); err == nil {
		t.Fatal("locked directory movable")
	}
}
func TestAuthProfileLockDoesNotReadOrWriteProfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if os.WriteFile(path, []byte("synthetic-config"), 0600) != nil {
		t.Fatal("fixture")
	}
	unlock, err := LockAuthProfile(path)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "synthetic-config" {
		t.Fatal("product cannot read locked config")
	}
	if os.WriteFile(path, []byte("other"), 0600) == nil {
		t.Fatal("config changed under lock")
	}
}
