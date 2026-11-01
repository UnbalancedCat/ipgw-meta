//go:build windows

package casacquire

import (
	"bytes"
	"context"
	"crypto/x509"
	"testing"

	"golang.org/x/sys/windows"
)

func TestExtraSourceAllowlistWithoutOpeningStores(t *testing.T) {
	names := []string{"ROOT", "ROOT", "CA", "CA", "Disallowed", "Disallowed", "AuthRoot"}
	flags := []uint32{windows.CERT_SYSTEM_STORE_CURRENT_USER_GROUP_POLICY, windows.CERT_SYSTEM_STORE_LOCAL_MACHINE_GROUP_POLICY, windows.CERT_SYSTEM_STORE_CURRENT_USER_GROUP_POLICY, windows.CERT_SYSTEM_STORE_LOCAL_MACHINE_GROUP_POLICY, windows.CERT_SYSTEM_STORE_CURRENT_USER_GROUP_POLICY, windows.CERT_SYSTEM_STORE_LOCAL_MACHINE_GROUP_POLICY, windows.CERT_SYSTEM_STORE_LOCAL_MACHINE}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for i := range names {
		n, f, ok := sourceStoreSpec(i)
		if !ok || n != names[i] || f != flags[i] {
			t.Fatal("source allowlist differs")
		}
		state, reason := localSource(ctx, i, func([]byte) string { t.Fatal("canceled read"); return "" })
		if state != "unavailable" || reason != "canceled" {
			t.Fatal("canceled source accepted")
		}
	}
	for _, i := range []int{-1, 7, 100} {
		if _, _, ok := sourceStoreSpec(i); ok {
			t.Fatal("unknown source accepted")
		}
		if _, reason := localSource(context.Background(), i, func([]byte) string { return "" }); reason != "open_failed" {
			t.Fatal("invalid source reached provider")
		}
	}
}

func TestExtraSourceMemoryStoreEnumeration(t *testing.T) {
	c := generateChain(t, func(*x509.Certificate, *x509.Certificate) {})
	for _, name := range []string{"copy", "empty", "limit", "cancel-during", "cancel-before"} {
		t.Run(name, func(t *testing.T) {
			h, err := windows.CertOpenStore(windows.CERT_STORE_PROV_MEMORY, 0, 0, 0, 0)
			if err != nil {
				t.Fatal("memory store creation")
			}
			t.Cleanup(func() {
				if windows.CertCloseStore(h, windows.CERT_CLOSE_STORE_CHECK_FLAG) != nil {
					t.Error("enumeration context leak")
				}
			})
			if name != "empty" {
				cert, err := windows.CertCreateCertificateContext(windows.X509_ASN_ENCODING, &c.root.Raw[0], uint32(len(c.root.Raw)))
				if err != nil {
					t.Fatal("generated context")
				}
				err = windows.CertAddCertificateContextToStore(h, cert, windows.CERT_STORE_ADD_ALWAYS, nil)
				windows.CertFreeCertificateContext(cert)
				if err != nil {
					t.Fatal("memory insertion")
				}
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if name == "cancel-before" {
				cancel()
			}
			calls := 0
			var copied []byte
			reason := readSourceStore(ctx, h, func(der []byte) string {
				calls++
				if name == "limit" {
					return "limit"
				}
				if name == "cancel-during" {
					cancel()
				}
				copied = bytes.Clone(der)
				return ""
			})
			want := ""
			if name == "limit" {
				want = "limit"
			}
			if name == "cancel-before" || name == "cancel-during" {
				want = "canceled"
			}
			if reason != want {
				t.Fatalf("got %s", reason)
			}
			if name == "copy" && (calls != 1 || !bytes.Equal(copied, c.root.Raw)) {
				t.Fatal("owned DER mismatch")
			}
			if (name == "empty" || name == "cancel-before") && calls != 0 {
				t.Fatal("unexpected callback")
			}
		})
	}
}
