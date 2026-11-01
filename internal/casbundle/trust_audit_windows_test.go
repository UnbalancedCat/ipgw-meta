//go:build windows

package casbundle

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTrustAuditIdentityAndModeSeparation(t *testing.T) {
	dir, ed, id := developmentFixture(t)
	if b, err := OpenTrustAudit(dir, ed, id); err == nil {
		b.Close()
		t.Fatal("acquisition identity accepted")
	}
	if err := os.Rename(filepath.Join(dir, "cas-acquire.exe"), filepath.Join(dir, "trust-audit.exe")); err != nil {
		t.Fatal(err)
	}
	if b, err := OpenDevelopment(dir, ed, id); err == nil {
		b.Close()
		t.Fatal("trust identity accepted as acquisition")
	}
	b, err := OpenTrustAudit(dir, ed, id)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	for _, mode := range []string{"--anonymous-once-confirmed", "--run-once", "--local-cache-check"} {
		if _, err := b.Execute(context.Background(), mode); err == nil {
			t.Fatal("mode widened")
		}
	}
	if os.WriteFile(filepath.Join(dir, "trust-audit.exe"), []byte("mutation"), 0600) == nil {
		t.Fatal("payload unlocked")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if _, err := b.Execute(ctx, "--self-test"); err == nil {
		t.Fatal("cancel ignored")
	}
	b.Close()
	if os.WriteFile(filepath.Join(dir, "trust-audit.exe"), []byte("released"), 0600) != nil {
		t.Fatal("lock remains")
	}
}
