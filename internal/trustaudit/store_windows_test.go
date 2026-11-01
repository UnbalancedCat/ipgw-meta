//go:build windows

package trustaudit

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"golang.org/x/sys/windows"
	"testing"
)

// Only a generated memory-store fixture: no system store access.
func TestMemoryStoreGoRejectedDERStillParticipatesInDenial(t *testing.T) {
	der, err := syntheticDER()
	if err != nil {
		t.Fatal(err)
	}
	var outer, tbs, serial asn1.RawValue
	if _, err = asn1.Unmarshal(der, &outer); err != nil {
		t.Fatal("outer")
	}
	if _, err = asn1.Unmarshal(outer.Bytes, &tbs); err != nil {
		t.Fatal("tbs")
	}
	rest, err := asn1.Unmarshal(tbs.Bytes, &serial)
	if err != nil {
		t.Fatal("version")
	}
	if serial.Class == 2 && serial.Tag == 0 {
		if _, err = asn1.Unmarshal(rest, &serial); err != nil {
			t.Fatal("serial")
		}
	}
	if serial.Class != 0 || serial.Tag != 2 || len(serial.Bytes) != 1 {
		t.Fatal("serial shape")
	}
	serial.Bytes[0] |= 0x80
	if _, err = x509.ParseCertificate(der); err == nil {
		t.Fatal("Go accepted fixture")
	}
	h, err := windows.CertOpenStore(windows.CERT_STORE_PROV_MEMORY, 0, 0, 0, 0)
	if err != nil {
		t.Fatal("memory store")
	}
	defer func() {
		if windows.CertCloseStore(h, windows.CERT_CLOSE_STORE_CHECK_FLAG) != nil {
			t.Error("context leaked")
		}
	}()
	c, err := windows.CertCreateCertificateContext(windows.X509_ASN_ENCODING, &der[0], uint32(len(der)))
	if err != nil {
		t.Fatal("fixture context")
	}
	err = windows.CertAddCertificateContextToStore(h, c, windows.CERT_STORE_ADD_ALWAYS, nil)
	windows.CertFreeCertificateContext(c)
	if err != nil {
		t.Fatal("insert")
	}
	parsed := &accumulator{}
	if reason := readStore(context.Background(), h, parsed); reason != "" || parsed.count != 1 || parsed.bad != 1 || len(parsed.hashes) != 0 {
		t.Fatal("parse coverage")
	}
	raw := &accumulator{raw: true}
	if reason := readStore(context.Background(), h, raw); reason != "" || raw.count != 1 || !raw.hashes[sha256.Sum256(der)] {
		t.Fatal("raw denial lost")
	}
}

func TestMemoryStoreReadAndRelease(t *testing.T) {
	der, err := syntheticDER()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"valid", "empty", "count", "bytes", "canceled"} {
		t.Run(name, func(t *testing.T) {
			h, err := windows.CertOpenStore(windows.CERT_STORE_PROV_MEMORY, 0, 0, 0, 0)
			if err != nil {
				t.Fatal("memory store")
			}
			defer func() {
				if windows.CertCloseStore(h, windows.CERT_CLOSE_STORE_CHECK_FLAG) != nil {
					t.Error("context leaked")
				}
			}()
			if name != "empty" {
				c, err := windows.CertCreateCertificateContext(windows.X509_ASN_ENCODING, &der[0], uint32(len(der)))
				if err != nil {
					t.Fatal("fixture")
				}
				err = windows.CertAddCertificateContextToStore(h, c, windows.CERT_STORE_ADD_ALWAYS, nil)
				windows.CertFreeCertificateContext(c)
				if err != nil {
					t.Fatal("insert")
				}
			}
			a := &accumulator{}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			want := ""
			switch name {
			case "count":
				a.count = maxEntries
				want = "limit"
			case "bytes":
				a.bytes = maxBytes
				want = "limit"
			case "canceled":
				cancel()
				want = "canceled"
			}
			if reason := readStore(ctx, h, a); reason != want {
				t.Fatalf("reason %s", reason)
			}
			if name == "valid" && a.parsed != 1 {
				t.Fatal("valid missing")
			}
		})
	}
}
func TestLocalReadGuardsBeforeStore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, reason := localRead(ctx, "ROOT", 0, &accumulator{}); reason != "canceled" {
		t.Fatal("cancellation")
	}
	for _, name := range []string{"MY", "AuthRoot", "ROOT/raw"} {
		if _, reason := localRead(context.Background(), name, 0, &accumulator{}); reason != "open_failed" {
			t.Fatal("source widened")
		}
	}
}
