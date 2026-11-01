//go:build windows

package casbundle

import (
	"path/filepath"
	"testing"
)

// Hosted Windows runners can expose TEMP through a short-name alias. Positive
// fixtures use the long canonical spelling, as the bundle verifier requires.
// Production checks still reject aliases, reparse points and hard links.
func canonicalTempDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return dir
}
