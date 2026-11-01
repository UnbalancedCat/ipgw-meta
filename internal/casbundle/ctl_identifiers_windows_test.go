//go:build windows

package casbundle

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCTLIdentifiersIdentityDeadlineAndModes(t *testing.T) {
	dir, ed, id := developmentFixture(t)
	if b, e := OpenCTLIdentifiers(dir, ed, id); e == nil {
		b.Close()
		t.Fatal("wrong identity accepted")
	}
	if e := os.Rename(filepath.Join(dir, "cas-acquire.exe"), filepath.Join(dir, "ctl-identifiers.exe")); e != nil {
		t.Fatal(e)
	}
	if b, e := OpenTrustAudit(dir, ed, id); e == nil {
		b.Close()
		t.Fatal("CTL accepted as trust audit")
	}
	b, e := OpenCTLIdentifiers(dir, ed, id)
	if e != nil {
		t.Fatal(e)
	}
	defer b.Close()
	for _, mode := range []string{"--run-once", "--anonymous-once-confirmed", "--trust-audit", "--ctl-check", "--local-cache-check"} {
		if _, e := b.Execute(context.Background(), mode); e == nil {
			t.Fatal("forbidden mode accepted")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, e := b.Execute(ctx, "--self-test"); e == nil {
		t.Fatal("timeout ignored")
	}
	if time.Since(start) > 3*time.Second {
		t.Fatal("child was not bounded")
	}
	b.Close()
	if e := os.WriteFile(filepath.Join(dir, "ctl-identifiers.exe"), []byte("released"), 0600); e != nil {
		t.Fatal("payload remains locked")
	}
}

func TestCTLIdentifiersHardDeadline(t *testing.T) {
	dir, ed, id := developmentFixture(t)
	if err := os.Rename(filepath.Join(dir, "cas-acquire.exe"), filepath.Join(dir, "ctl-identifiers.exe")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(filepath.Dir(dir), "ctl-deadline-fixture"), []byte("synthetic"), 0600); err != nil {
		t.Fatal(err)
	}
	b, err := OpenCTLIdentifiers(dir, ed, id)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	start := time.Now()
	if _, err = b.Execute(context.Background(), "--self-test"); err == nil {
		t.Fatal("hung child accepted")
	}
	elapsed := time.Since(start)
	if elapsed < 14*time.Second || elapsed > 25*time.Second {
		t.Fatalf("hard deadline not exercised: %s", elapsed)
	}
	if _, err := os.Stat(filepath.Join(dir, "executed.marker")); err != nil {
		t.Fatal("child never started")
	}
	b.Close()
	if err := os.Rename(filepath.Join(dir, "ctl-identifiers.exe"), filepath.Join(dir, "terminated-fixture.exe")); err != nil {
		t.Fatal("child still running")
	}
}
