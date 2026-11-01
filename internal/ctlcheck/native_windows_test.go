//go:build windows

package ctlcheck

import (
	"bytes"
	"context"
	"fmt"
	"golang.org/x/sys/windows"
	"os"
	"os/exec"
	"testing"
	"time"
	"unsafe"
)

func TestNativeLayout(t *testing.T) {
	if unsafe.Sizeof(ctlContext{}) != 64 || unsafe.Offsetof(ctlContext{}.msg) != 40 || unsafe.Sizeof(verifyPara{}) != 32 || unsafe.Offsetof(verifyPara{}.signer) != 24 {
		t.Fatal("ABI mismatch")
	}
}
func TestNativeSyntheticWorker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestNativeWorkerHelper$", "-test.v")
	cmd.Env = append(os.Environ(), "IPGW_CTL_SYNTHETIC_HELPER=1")
	b, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("synthetic worker: %v %s", e, b)
	}
}
func TestNativeWorkerHelper(t *testing.T) {
	if os.Getenv("IPGW_CTL_SYNTHETIC_HELPER") != "1" {
		return
	}
	if !protectProcess() {
		t.Fatal("protection setup failed")
	}
	f := synthetic(t)
	m, reason := parseMessage(context.Background(), f.blob)
	if reason != "" {
		t.Fatal(reason)
	}
	ptr, _, code := createCTL.Call(0x10001, uintptr(unsafe.Pointer(&f.blob[0])), uintptr(len(f.blob)))
	if ptr == 0 {
		t.Fatalf("synthetic create context: %v", code)
	}
	defer freeCTL.Call(ptr)
	var cc ctlContext
	var copied uintptr
	if e := windows.ReadProcessMemory(windows.CurrentProcess(), ptr, (*byte)(unsafe.Pointer(&cc)), unsafe.Sizeof(cc), &copied); e != nil {
		t.Fatal(e)
	}
	if cc.content == nil {
		t.Fatal("missing content")
	}
	actual := unsafe.Slice(cc.content, cc.contentLen)
	if !bytes.Equal(actual, m.content) {
		t.Fatalf("synthetic content shape: got %d expected %d equal-body %v", len(actual), len(m.content), bytes.Equal(actual, m.content[2:]))
	}
	signer, e := windows.CertCreateCertificateContext(windows.X509_ASN_ENCODING, &f.leaf.Raw[0], uint32(len(f.leaf.Raw)))
	if e != nil {
		t.Fatal(e)
	}
	defer windows.CertFreeCertificateContext(signer)
	p := verifyPara{size: uint32(unsafe.Sizeof(verifyPara{})), kind: 2, signer: signer}
	ok, _, code := msgControl.Call(uintptr(cc.msg), 0, 19, uintptr(unsafe.Pointer(&p)))
	if ok == 0 {
		t.Fatalf("synthetic signature: %v", code)
	}
	if !verifyNative(f.blob, m, f.leaf) {
		t.Fatal("native disagreement")
	}
	for _, cms := range []bool{false, true} {
		blob, e := f.signEncoding(m.content, cms, false)
		if e != nil {
			t.Fatal(e)
		}
		parsed, reason := parseMessage(context.Background(), blob)
		if reason != "" || !verifyNative(blob, parsed, f.leaf) {
			t.Fatalf("synthetic encapsulation failed: cms=%v reason=%s", cms, reason)
		}
	}
	// Independent Windows decoder/signature evidence for omitted entry attributes.
	// Deliberately bypass our CTL parser; only the native decoder sees this shape.
	for _, cms := range []bool{false, true} {
		content := der(0x30, der(0x30, encOID(oidRootList)), encTime(f.now), encAlg(oidSHA1),
			der(0x30, der(0x30, der(4, make([]byte, 20)))))
		blob, err := f.signEncoding(content, cms, false)
		if err != nil || !verifyNative(blob, &message{content: content}, f.leaf) {
			t.Fatalf("native omitted attributes not established: cms=%v", cms)
		}
	}
	checkD11Signatures(t, f, verifyNative)
	checkD13Signatures(t, f, verifyNative)
	cmd := exec.Command("cmd.exe", "/d", "/c", "exit", "0")
	if e := cmd.Run(); e == nil {
		t.Fatal("job allowed descendant")
	}
	fmt.Println("synthetic native signature and job passed")
}
