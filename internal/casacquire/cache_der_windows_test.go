//go:build windows

package casacquire

import (
	"context"
	"crypto/x509"
	"encoding/asn1"
	"testing"

	"golang.org/x/sys/windows"
)

// Generated, deliberately invalid DER only. This proves that enumeration from
// a Windows store does not guarantee acceptance by the pinned Go parser. It
// does not identify any actual user's certificate or justify relaxing parsing.
func TestWindowsMemoryStoreCanContainGoRejectedDER(t *testing.T) {
	c := generateChain(t, func(*x509.Certificate, *x509.Certificate) {})
	der := append([]byte(nil), c.intermediate.Raw...)
	var outer, tbs, field asn1.RawValue
	if _, err := asn1.Unmarshal(der, &outer); err != nil {
		t.Fatal("synthetic outer parse failed")
	}
	if _, err := asn1.Unmarshal(outer.Bytes, &tbs); err != nil {
		t.Fatal("synthetic TBS parse failed")
	}
	rest, err := asn1.Unmarshal(tbs.Bytes, &field)
	if err != nil {
		t.Fatal("synthetic version parse failed")
	}
	if field.Class == 2 && field.Tag == 0 {
		if _, err = asn1.Unmarshal(rest, &field); err != nil {
			t.Fatal("synthetic serial parse failed")
		}
	}
	if field.Class != 0 || field.Tag != 2 || len(field.Bytes) != 1 {
		t.Fatal("unexpected synthetic serial shape")
	}
	field.Bytes[0] |= 0x80 // Negative serial; preserve DER framing, invalidate signature.
	if _, err = x509.ParseCertificate(der); err == nil {
		t.Fatal("Go unexpectedly accepted synthetic negative serial")
	}
	store, err := windows.CertOpenStore(windows.CERT_STORE_PROV_MEMORY, 0, 0, 0, 0)
	if err != nil {
		t.Fatal("memory store creation failed")
	}
	t.Cleanup(func() {
		if windows.CertCloseStore(store, windows.CERT_CLOSE_STORE_CHECK_FLAG) != nil {
			t.Error("context leaked")
		}
	})
	cert, err := windows.CertCreateCertificateContext(windows.X509_ASN_ENCODING, &der[0], uint32(len(der)))
	if err != nil {
		t.Fatal("Windows did not accept synthetic DER context")
	}
	err = windows.CertAddCertificateContextToStore(store, cert, windows.CERT_STORE_ADD_ALWAYS, nil)
	windows.CertFreeCertificateContext(cert)
	if err != nil {
		t.Fatal("memory store insertion failed")
	}
	var cached cachedCertificates
	if reason := readIntermediateStore(context.Background(), store, &cached); reason != "" || len(cached.certs) != 0 || !cached.partial || cached.count != 1 || cached.bytes != len(der) {
		t.Fatalf("rejected cache entry not accounted as partial: %s", reason)
	}
	for _, valid := range [][]byte{c.intermediate.Raw, c.root.Raw} {
		cert, err := windows.CertCreateCertificateContext(windows.X509_ASN_ENCODING, &valid[0], uint32(len(valid)))
		if err != nil {
			t.Fatal("valid synthetic context failed")
		}
		err = windows.CertAddCertificateContextToStore(store, cert, windows.CERT_STORE_ADD_ALWAYS, nil)
		windows.CertFreeCertificateContext(cert)
		if err != nil {
			t.Fatal("valid synthetic insertion failed")
		}
	}
	cached = cachedCertificates{}
	if reason := readIntermediateStore(context.Background(), store, &cached); reason != "" || len(cached.certs) != 2 || !cached.partial || cached.count != 3 || cached.bytes != len(der)+len(c.intermediate.Raw)+len(c.root.Raw) {
		t.Fatalf("mixed memory store did not preserve valid entries: %s", reason)
	}
}
