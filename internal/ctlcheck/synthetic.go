package ctlcheck

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

func der(tag byte, parts ...[]byte) []byte {
	b := bytes.Join(parts, nil)
	out := []byte{tag}
	n := len(b)
	if n < 128 {
		out = append(out, byte(n))
	} else {
		var tmp [8]byte
		i := len(tmp)
		for n > 0 {
			i--
			tmp[i] = byte(n)
			n >>= 8
		}
		out = append(out, 0x80|byte(len(tmp)-i))
		out = append(out, tmp[i:]...)
	}
	return append(out, b...)
}
func encOID(s string) []byte {
	var v asn1.ObjectIdentifier
	for _, p := range strings.Split(s, ".") {
		n, _ := strconv.Atoi(p)
		v = append(v, n)
	}
	b, _ := asn1.Marshal(v)
	return b
}
func encInt(n int64) []byte                   { b, _ := asn1.Marshal(n); return b }
func encTime(t time.Time) []byte              { b, _ := asn1.Marshal(t.UTC()); return b }
func encAlg(s string) []byte                  { return der(0x30, encOID(s), []byte{5, 0}) }
func attribute(s string, value []byte) []byte { return der(0x30, encOID(s), der(0x31, value)) }

type fixture struct {
	blob           []byte
	mat            materials
	pins           pins
	now            time.Time
	root, ca, leaf *x509.Certificate
	key            *rsa.PrivateKey
}

func generateFixture() (fixture, error) {
	f := fixture{now: time.Now().UTC().Truncate(time.Second)}
	keys := make([]*rsa.PrivateKey, 3)
	certs := make([]*x509.Certificate, 3)
	for i := range keys {
		k, e := rsa.GenerateKey(rand.Reader, 2048)
		if e != nil {
			return f, e
		}
		keys[i] = k
		t := &x509.Certificate{SerialNumber: big.NewInt(int64(i + 1)), Subject: pkix.Name{CommonName: "synthetic-ctl-only-" + strconv.Itoa(i)}, NotBefore: f.now.Add(-time.Hour), NotAfter: f.now.Add(24 * time.Hour), BasicConstraintsValid: true, SignatureAlgorithm: x509.SHA256WithRSA, KeyUsage: x509.KeyUsageDigitalSignature}
		if i < 2 {
			t.IsCA = true
			t.KeyUsage |= x509.KeyUsageCertSign
			t.MaxPathLen = 2 - i
		} else {
			t.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning}
		}
		parent, signing := t, k
		if i > 0 {
			parent, signing = certs[i-1], keys[i-1]
		}
		b, e := x509.CreateCertificate(rand.Reader, t, parent, &k.PublicKey, signing)
		if e != nil {
			return f, e
		}
		certs[i], e = x509.ParseCertificate(b)
		if e != nil {
			return f, e
		}
	}
	f.root, f.ca, f.leaf, f.key = certs[0], certs[1], certs[2], keys[2]
	f.pins = pins{sha256.Sum256(f.root.Raw), sha256.Sum256(f.ca.Raw)}
	f.mat = materials{roots: []*x509.Certificate{f.root}, denied: map[[32]byte]bool{}}
	content := der(0x30, der(0x30, encOID(oidRootList)), encInt(1), encTime(f.now.Add(-time.Minute)), encTime(f.now.Add(time.Hour)), encAlg(oidSHA1), der(0x30))
	var e error
	f.blob, e = f.sign(content)
	return f, e
}
func (f fixture) sign(content []byte) ([]byte, error) {
	return f.signEncoding(content, false, false)
}
func (f fixture) signEncoding(content []byte, cms, wrongDigest bool) ([]byte, error) {
	return f.signWithAttributes(content, cms, wrongDigest, nil)
}
func (f fixture) signWithAttributes(content []byte, cms, wrongDigest bool, extra [][]byte) ([]byte, error) {
	value, err := whole(content, 0x30)
	if err != nil {
		return nil, err
	}
	digestContent := value.body
	if cms != wrongDigest {
		digestContent = content
	}
	attrs := [][]byte{attribute("1.2.840.113549.1.9.3", encOID(oidCTL)), attribute("1.2.840.113549.1.9.4", der(4, hashed(crypto.SHA256, digestContent)))}
	attrs = append(attrs, extra...)
	sort.Slice(attrs, func(i, j int) bool { return bytes.Compare(attrs[i], attrs[j]) < 0 })
	signed := der(0x31, attrs...)
	sig, e := rsa.SignPKCS1v15(rand.Reader, f.key, crypto.SHA256, hashed(crypto.SHA256, signed))
	if e != nil {
		return nil, e
	}
	digest := encAlg("2.16.840.1.101.3.4.2.1")
	si := der(0x30, encInt(1), der(0x30, f.leaf.RawIssuer, encInt(f.leaf.SerialNumber.Int64())), digest, der(0xa0, attrs...), encAlg("1.2.840.113549.1.1.1"), der(4, sig))
	wrapped := content
	version := int64(1)
	if cms {
		wrapped = der(4, content)
		version = 3
	}
	sd := der(0x30, encInt(version), der(0x31, digest), der(0x30, encOID(oidCTL), der(0xa0, wrapped)), der(0xa0, f.ca.Raw, f.leaf.Raw), der(0x31, si))
	return der(0x30, encOID(oidSigned), der(0xa0, sd)), nil
}
func SelfTest(ctx context.Context) Report {
	r := emptyReport("self-test")
	f, e := generateFixture()
	if e != nil {
		return r
	}
	good := observe(ctx, "authroot_ctl", f.blob, f.mat, f.pins, f.now, verifyNative)
	// Exercise D11 through the packaged self-test before any local source read.
	for _, a := range []struct {
		oid  string
		size int
	}{
		{oidSHA1, 20}, {"2.16.840.1.101.3.4.2.1", 32}, {"2.16.840.1.101.3.4.2.2", 48}, {"2.16.840.1.101.3.4.2.3", 64},
	} {
		content := der(0x30, der(0x30, encOID(oidRootList)), encTime(f.now), encAlg(a.oid),
			der(0x30, der(0x30, der(4, make([]byte, a.size)))),
			der(0xa0, der(0x30, der(0x30, encOID("1.2.3.4"), []byte{1, 1, 255}, der(4, []byte("synthetic-only"))))))
		blob, err := f.sign(content)
		if err != nil {
			return r
		}
		observed := observe(ctx, "authroot_ctl", blob, f.mat, f.pins, f.now, verifyNative)
		if observed.State != "complete" || observed.Extensions != "critical_present_uninterpreted" {
			return r
		}
	}
	// D13 synthetic-only path, before any real source reader.
	content := der(0x30, der(0x30, encOID(oidDisallowedList)), encTime(f.now), encAlg(oidSignatureHashProperty), der(0x30, der(0x30, der(4, make([]byte, 32)))))
	blob, err := f.signWithAttributes(content, true, false, [][]byte{attribute(oidSPCOpus, der(0x30))})
	if err != nil {
		return r
	}
	observed := observe(ctx, "disallowed_ctl", blob, f.mat, f.pins, f.now, verifyNative)
	if observed.State != "complete" || observed.MemberIdentifierCoverage != "bounded_uninterpreted" || observed.SignedAttributeCoverage != "spc_opus_outer_only" {
		return r
	}
	bad := append([]byte(nil), f.blob...)
	bad[len(bad)-1] ^= 1
	if observe(ctx, "authroot_ctl", bad, f.mat, f.pins, f.now, verifyNative).State == "complete" {
		return r
	}
	f.mat.denied[sha256.Sum256(f.leaf.Raw)] = true
	if observe(ctx, "authroot_ctl", f.blob, f.mat, f.pins, f.now, verifyNative).Reason != "denied" {
		return r
	}
	r.Sources[0] = good
	r.Sources[1] = good
	r.Sources[1].Source = "disallowed_ctl"
	if good.State == "complete" {
		r.State = "complete"
	}
	return r
}

// Only dedicated subprocess entrypoints call this, never the test process.
func SyntheticProcess(ctx context.Context) Report {
	if runtime.GOOS == "windows" && !protectProcess() {
		r := emptyReport("self-test")
		for i := range r.Sources {
			r.Sources[i].Reason = "protection_failed"
		}
		return r
	}
	return SelfTest(ctx)
}
