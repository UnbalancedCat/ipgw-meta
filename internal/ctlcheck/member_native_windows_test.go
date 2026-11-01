//go:build windows

package ctlcheck

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"
	"unsafe"
)

// This API exists only in test code, and receives freshly generated DER only.
func syntheticSignatureProperty(der []byte) ([]byte, bool) {
	if len(der) == 0 || len(der) > 64<<10 {
		return nil, false
	}
	proc := crypt.NewProc("CryptHashToBeSigned")
	out := make([]byte, 64)
	size := uint32(len(out))
	ok, _, _ := proc.Call(0, 1, uintptr(unsafe.Pointer(&der[0])), uintptr(len(der)), uintptr(unsafe.Pointer(&out[0])), uintptr(unsafe.Pointer(&size)))
	runtime.KeepAlive(der)
	runtime.KeepAlive(out)
	if ok == 0 || size == 0 || size > 64 {
		return nil, false
	}
	return out[:size], true
}
func TestMemberNativeSynthetic(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestMemberNativeHelper$", "-test.v")
	cmd.Env = append(os.Environ(), "IPGW_CTL_MEMBER_SYNTHETIC_HELPER=1")
	b, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("synthetic native worker: %v %s", e, b)
	}
	t.Log(string(b))
}

func TestMemberPreparationHardStop(t *testing.T) {
	if os.Getenv("IPGW_MEMBER_PREPARATION_HANG") == "1" {
		prepareProtectedMembers(context.Background(), protectProcess, func() (*MemberSession, *Membership) { select {} }, 50*time.Millisecond)
		t.Fatal("hanging preparation returned")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestMemberPreparationHardStop$")
	cmd.Env = append(os.Environ(), "IPGW_MEMBER_PREPARATION_HANG=1")
	b, e := cmd.CombinedOutput()
	if e == nil || cmd.ProcessState.ExitCode() != 4 || len(b) != 0 || ctx.Err() != nil {
		t.Fatal("hard stop missing or partial output")
	}
}
func TestMemberNativeHelper(t *testing.T) {
	if os.Getenv("IPGW_CTL_MEMBER_SYNTHETIC_HELPER") != "1" {
		return
	}
	if !protectProcess() {
		t.Fatal("job protection")
	}
	if !membershipSelfTest(context.Background(), verifyNative) {
		t.Fatal("prepared native comparison")
	}
	f := synthetic(t)
	for i, a := range []x509.SignatureAlgorithm{x509.SHA256WithRSA, x509.SHA384WithRSA, x509.SHA512WithRSA} {
		c := &x509.Certificate{SerialNumber: big.NewInt(int64(i + 10)), Subject: pkix.Name{CommonName: "synthetic-member-only"}, NotBefore: f.now.Add(-time.Hour), NotAfter: f.now.Add(time.Hour), SignatureAlgorithm: a}
		der, e := x509.CreateCertificate(rand.Reader, c, c, &f.key.PublicKey, f.key)
		if e != nil {
			t.Fatal(e)
		}
		cert, e := x509.ParseCertificate(der)
		if e != nil {
			t.Fatal(e)
		}
		observed, ok := syntheticSignatureProperty(der)
		if !ok {
			t.Fatal("native property unavailable")
		}
		portable, err := syntheticMemberKey("disallowed_ctl", der)
		if err != nil || !bytes.Equal(portable, observed) {
			t.Fatal("portable/native mismatch")
		}
		mutated := append([]byte(nil), der...)
		mutated[len(mutated)-1] ^= 1
		changed, ok := syntheticSignatureProperty(mutated)
		if !ok || !bytes.Equal(changed, observed) {
			t.Fatal("signature bytes changed TBS property")
		}

		joined := append(append([]byte(nil), cert.RawTBSCertificate...), cert.Signature...)
		s1 := sha1.Sum(cert.RawTBSCertificate)
		s2 := sha256.Sum256(cert.RawTBSCertificate)
		s3 := sha512.Sum384(cert.RawTBSCertificate)
		s4 := sha512.Sum512(cert.RawTBSCertificate)
		s5 := sha1.Sum(cert.Signature)
		s6 := sha1.Sum(joined)
		s7 := sha1.Sum(der)
		s8 := sha256.Sum256(joined)
		candidates := []struct {
			name  string
			value []byte
		}{{"sha1_tbs", s1[:]}, {"sha256_tbs", s2[:]}, {"sha384_tbs", s3[:]}, {"sha512_tbs", s4[:]}, {"sha1_signature", s5[:]}, {"sha1_tbs_signature", s6[:]}, {"sha1_full_der", s7[:]}, {"sha256_tbs_signature", s8[:]}}
		matched := 0
		for _, candidate := range candidates {
			if bytes.Equal(observed, candidate.value) {
				matched++
				fmt.Printf("synthetic_signature_property algorithm_case=%d matches=%s\n", i, candidate.name)
			}
		}
		if matched != 1 {
			t.Fatal("synthetic hash rule not uniquely established")
		}
	}
}
