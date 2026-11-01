//go:build windows

package casbundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestObservedBundleBindsAllFourFiles(t *testing.T) {
	for _, mutation := range []string{"valid", "observed_hash", "missing", "extra", "hardlink", "old_identity"} {
		t.Run(mutation, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "payload")
			if os.Mkdir(dir, 0700) != nil {
				t.Fatal("mkdir")
			}
			hashes := make([]string, 4)
			for i, name := range authObservedNames {
				data := []byte("synthetic-" + name)
				if os.WriteFile(filepath.Join(dir, name), data, 0600) != nil {
					t.Fatal("write")
				}
				hashes[i] = Digest(data)
			}
			switch mutation {
			case "observed_hash":
				hashes[1] = hashes[0]
			case "missing":
				if os.Remove(filepath.Join(dir, "auth-observed.exe")) != nil {
					t.Fatal("remove")
				}
			case "extra":
				if os.WriteFile(filepath.Join(dir, "extra"), []byte("synthetic"), 0600) != nil {
					t.Fatal("write")
				}
			case "hardlink":
				if os.Link(filepath.Join(dir, "auth-observed.exe"), filepath.Join(filepath.Dir(dir), "alias")) != nil {
					t.Fatal("link")
				}
			}
			if mutation == "old_identity" {
				b, err := OpenAuthDevelopment(dir, hashes[0], hashes[2], hashes[3])
				if err == nil {
					b.Close()
					t.Fatal("old schema accepted new bundle")
				}
				return
			}
			b, err := OpenAuthObserved(dir, hashes[0], hashes[1], hashes[2], hashes[3])
			if mutation != "valid" {
				if err == nil {
					b.Close()
					t.Fatal("invalid observed bundle accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer b.Close()
			if b.ObservedPath == "" || b.ObservedPath == b.ProductPath || b.Verify() != nil {
				t.Fatal("observed identity")
			}
			if os.WriteFile(filepath.Join(dir, "auth-observed.exe"), []byte("tamper"), 0600) == nil {
				t.Fatal("observed unlocked")
			}
		})
	}
}
