//go:build windows

package casacquire

import (
	"bytes"
	"context"
	"crypto/x509"
	"testing"

	"golang.org/x/sys/windows"
)

func TestIntermediateEnumerationMemoryStore(t *testing.T) {
	c := generateChain(t, func(*x509.Certificate, *x509.Certificate) {})
	for _, name := range []string{"copy", "empty", "count-limit", "byte-limit", "canceled"} {
		t.Run(name, func(t *testing.T) {
			store, err := windows.CertOpenStore(windows.CERT_STORE_PROV_MEMORY, 0, 0, 0, 0)
			if err != nil {
				t.Fatal("memory store creation failed")
			}
			t.Cleanup(func() {
				// CertCloseStore always closes; CHECK reports outstanding contexts.
				if windows.CertCloseStore(store, windows.CERT_CLOSE_STORE_CHECK_FLAG) != nil {
					t.Error("enumeration context leaked")
				}
			})
			if name != "empty" {
				cert, err := windows.CertCreateCertificateContext(windows.X509_ASN_ENCODING, &c.intermediate.Raw[0], uint32(len(c.intermediate.Raw)))
				if err != nil {
					t.Fatal("synthetic context creation failed")
				}
				err = windows.CertAddCertificateContextToStore(store, cert, windows.CERT_STORE_ADD_ALWAYS, nil)
				windows.CertFreeCertificateContext(cert)
				if err != nil {
					t.Fatal("memory store fixture insertion failed")
				}
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var cached cachedCertificates
			switch name {
			case "count-limit":
				cached.count = maxCachedCerts
			case "byte-limit":
				cached.bytes = maxCachedBytes
			case "canceled":
				cancel()
			}
			reason := readIntermediateStore(ctx, store, &cached)
			want := ""
			if name == "count-limit" || name == "byte-limit" {
				want = "tls_chain_probe_cache_limit"
			}
			if name == "canceled" {
				want = "tls_chain_probe_canceled"
			}
			if reason != want {
				t.Fatalf("enumeration reason=%s want=%s", reason, want)
			}
			if name == "copy" && (len(cached.certs) != 1 || !bytes.Equal(cached.certs[0].Raw, c.intermediate.Raw)) {
				t.Fatal("DER copy mismatch")
			}
			if name != "copy" && len(cached.certs) != 0 {
				t.Fatal("partial data on failure")
			}
		})
	}
}

func TestLocalIntermediateCanceledBeforeStore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	certs, reason := localIntermediates(ctx)
	if reason != "tls_chain_probe_canceled" || certs.certs != nil || certs.partial {
		t.Fatal("canceled local enumeration accepted")
	}
}
