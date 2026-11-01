//go:build windows

package auxgate

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestInstallJobNativeChild(t *testing.T) {
	if os.Getenv("IPGW_AUX_JOB_SYNTHETIC_CHILD") != "1" {
		return
	}
	if !protectInstall() {
		os.Exit(23)
	}
	os.Exit(0)
}
func TestInstallJobCanProtectIsolatedNativeVerifier(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestInstallJobNativeChild$")
	cmd.Env = append(os.Environ(), "IPGW_AUX_JOB_SYNTHETIC_CHILD=1")
	if err := cmd.Run(); err != nil {
		t.Fatal("native auxiliary Job protection unavailable")
	}
}
