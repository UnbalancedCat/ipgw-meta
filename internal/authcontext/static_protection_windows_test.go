//go:build windows

package authcontext

import (
	"context"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
	"golang.org/x/sys/windows"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestStaticNativeProtection(t *testing.T) {
	if os.Getenv("IPGW_STATIC_PROTECTION_HELPER") == "1" {
		if !ProtectStaticProcess() {
			t.Fatal("job not established")
		}
		// A commit larger than the per-process Job budget must be rejected.
		p, e := windows.VirtualAlloc(0, 384<<20, windows.MEM_COMMIT|windows.MEM_RESERVE, windows.PAGE_READWRITE)
		if p != 0 {
			windows.VirtualFree(p, 0, windows.MEM_RELEASE)
			t.Fatal("job commit cap not enforced")
		}
		if e == nil {
			t.Fatal("missing allocation rejection")
		}
		if r := SelfTestStatic(); r.State != "pass" || r.Validate() != nil {
			t.Fatal("bounded self-test failed")
		}
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestStaticNativeProtection$")
	cmd.Env = append(os.Environ(), "IPGW_STATIC_PROTECTION_HELPER=1")
	b, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("native synthetic protection failed: %v %s", e, b)
	}
}
func TestStaticNativeHardStop(t *testing.T) {
	if os.Getenv("IPGW_STATIC_HARDSTOP_HELPER") == "1" {
		protectedStatic(context.Background(), ProtectStaticProcess, func(context.Context) casprovenance.StaticReport { select {} }, 50*time.Millisecond)
		t.Fatal("hard stop returned")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestStaticNativeHardStop$")
	cmd.Env = append(os.Environ(), "IPGW_STATIC_HARDSTOP_HELPER=1")
	b, e := cmd.CombinedOutput()
	if e == nil || ctx.Err() != nil || cmd.ProcessState.ExitCode() != 4 || len(b) != 0 {
		t.Fatal("hard termination failed")
	}
}
