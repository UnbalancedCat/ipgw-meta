package ctlcheck

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"time"
)

type pins struct{ root, ca [32]byte }

func officialPins() pins {
	p := pins{}
	r, _ := hex.DecodeString("df545bf919a2439c36983b54cdfc903dfa4f37d3996d8d84b4c31eec6f3c163e")
	c, _ := hex.DecodeString("56625ec7886445706fe74ccd2abf69b7c9c9ccde3c7354903af1c8f0e30e9d1c")
	copy(p.root[:], r)
	copy(p.ca[:], c)
	return p
}

type materials struct {
	roots  []*x509.Certificate
	denied map[[32]byte]bool
}
type verifier func([]byte, *message, *x509.Certificate) bool

func observe(ctx context.Context, name string, b []byte, mat materials, p pins, now time.Time, verify verifier) Source {
	s := emptySource(name)
	s.Read = "complete"
	fail := func(reason string) Source {
		if r := stopped(ctx); r != "" {
			// Cancellation invalidates the unfinished observation atomically.
			out := emptySource(name)
			out.Read = "complete"
			out.Reason = r
			return out
		}
		s.Reason = reason
		return s
	}
	if r := stopped(ctx); r != "" {
		return fail(r)
	}
	m, reason := parseMessage(ctx, b)
	if reason != "" {
		return fail(reason)
	}
	if m.meta.memberAlgorithm == "signature_hash_property" && name != "disallowed_ctl" {
		return fail("parse_failed_ctl_algorithm_unsupported")
	}
	s.Structure = "bounded_checked"
	var root, ca, signer *x509.Certificate
	count := 0
	for _, c := range append(append([]*x509.Certificate(nil), mat.roots...), m.certs...) {
		id := sha256.Sum256(c.Raw)
		if id == p.root {
			root = c
		}
		if id == p.ca {
			ca = c
		}
	}
	for _, c := range m.certs {
		if bytes.Equal(c.RawIssuer, m.issuer) && c.SerialNumber.Cmp(m.serial) == 0 {
			signer = c
			count++
		}
	}
	if count > 1 {
		return fail("signer_ambiguous")
	}
	if root == nil || ca == nil || signer == nil {
		return fail("material_missing")
	}
	for _, c := range []*x509.Certificate{root, ca, signer} {
		if mat.denied[sha256.Sum256(c.Raw)] {
			s.Rejection = "present"
			return fail("denied")
		}
	}
	s.Rejection = "none"
	for _, c := range []*x509.Certificate{root, ca, signer} {
		if now.Before(c.NotBefore) || now.After(c.NotAfter) {
			s.CertificateTime = "invalid"
			return fail("certificate_time_invalid")
		}
	}
	s.CertificateTime = "valid"
	for _, c := range []*x509.Certificate{root, ca, signer} {
		k, ok := c.PublicKey.(*rsa.PublicKey)
		if !ok || k.N.BitLen() < 2048 || k.N.BitLen() > 8192 || k.E < 3 || k.E > 2147483647 || k.E%2 == 0 {
			return fail("unsupported")
		}
		if len(c.UnhandledCriticalExtensions) > 0 {
			return fail("conditions_rejected")
		}
	}
	if !root.BasicConstraintsValid || !root.IsCA || !ca.BasicConstraintsValid || !ca.IsCA || signer.IsCA || signer.KeyUsage&x509.KeyUsageDigitalSignature == 0 || root.KeyUsage&x509.KeyUsageCertSign == 0 || ca.KeyUsage&x509.KeyUsageCertSign == 0 {
		return fail("conditions_rejected")
	}
	for _, c := range []*x509.Certificate{ca, signer} {
		if c.SignatureAlgorithm != x509.SHA256WithRSA && c.SignatureAlgorithm != x509.SHA384WithRSA && c.SignatureAlgorithm != x509.SHA512WithRSA {
			return fail("unsupported")
		}
	}
	if !bytes.Equal(ca.RawIssuer, root.RawSubject) || !bytes.Equal(signer.RawIssuer, ca.RawSubject) {
		return fail("conditions_rejected")
	}
	// Two fixed edges, not a general-purpose chain builder or policy verifier.
	if stopped(ctx) != "" {
		return fail(stopped(ctx))
	}
	if ca.CheckSignatureFrom(root) != nil {
		return fail("signature_failed")
	}
	if stopped(ctx) != "" {
		return fail(stopped(ctx))
	}
	if signer.CheckSignatureFrom(ca) != nil {
		return fail("signature_failed")
	}
	s.Conditions = "satisfied"
	s.Hierarchy = "observed"
	if stopped(ctx) != "" {
		return fail(stopped(ctx))
	}
	if !verify(b, m, signer) {
		return fail("signature_failed")
	}
	if stopped(ctx) != "" {
		return fail(stopped(ctx))
	}
	s.Signature = "verified"
	s.Window = "within"
	switch {
	case !m.meta.next.IsZero() && !m.meta.next.After(m.meta.this):
		s.Window = "invalid"
	case m.meta.this.After(now):
		s.Window = "future"
	case m.meta.next.IsZero():
		s.Window = "unbounded"
	case now.After(m.meta.next):
		s.Window = "expired"
	}
	s.Usage = m.meta.usage
	s.Attributes = m.meta.attributes
	s.MemberAlgorithm = m.meta.memberAlgorithm
	s.Extensions = m.meta.extensions
	s.MemberIdentifierCoverage = m.meta.memberCoverage
	s.SignedAttributeCoverage = m.meta.signedCoverage
	s.State = "complete"
	s.Reason = "none"
	return s
}

type blobReader func(context.Context, int) ([]byte, string)
type materialReader func(context.Context) (materials, string)

func audit(ctx context.Context, mode string, read blobReader, load materialReader, p pins, verify verifier) Report {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	r := emptyReport(mode)
	var blobs [2][]byte
	any := false
	for i := range blobs {
		if reason := stopped(ctx); reason != "" {
			r.Sources[i].Reason = reason
			continue
		}
		b, reason := read(ctx, i)
		if reason == "" && (len(b) == 0 || len(b) > maxBlob) {
			reason = "limit"
		}
		if st := stopped(ctx); st != "" {
			reason = st
		}
		if reason != "" {
			r.Sources[i].Reason = reason
			r.Sources[i].Read = "unavailable"
			if reason == "absent" {
				r.Sources[i].Read = "absent"
			}
			continue
		}
		blobs[i] = b
		r.Sources[i].Read = "complete"
		any = true
	}
	if !any {
		return r
	}
	mat, reason := load(ctx)
	if st := stopped(ctx); st != "" {
		reason = st
	}
	for i, b := range blobs {
		if b == nil {
			continue
		}
		if reason != "" {
			r.Sources[i].Reason = reason
			continue
		}
		r.Sources[i] = observe(ctx, r.Sources[i].Source, b, mat, p, time.Now(), verify)
	}
	if r.Sources[0].State == "complete" && r.Sources[1].State == "complete" {
		r.State = "complete"
	}
	return r
}
func Run(ctx context.Context) Report {
	return runProtected(ctx, protectProcess, func() Report {
		return audit(ctx, "local-ctl-check", localBlob, localMaterials, officialPins(), verifyNative)
	})
}
func runProtected(ctx context.Context, protect func() bool, work func() Report) Report {
	if stopped(ctx) != "" || !protect() {
		r := emptyReport("local-ctl-check")
		for i := range r.Sources {
			r.Sources[i].Reason = "protection_failed"
			if reason := stopped(ctx); reason != "" {
				r.Sources[i].Reason = reason
			}
		}
		return r
	}
	return work()
}

// A size query and one exact read; no retry/growth loop even on changing values.
func readBinary(ctx context.Context, query func([]byte) (int, uint32, string)) ([]byte, string) {
	if reason := stopped(ctx); reason != "" {
		return nil, reason
	}
	n, typ, reason := query(nil)
	if reason != "" {
		return nil, reason
	}
	if typ != 3 {
		return nil, "type_invalid"
	}
	if n <= 0 || n > maxBlob {
		return nil, "limit"
	}
	if reason := stopped(ctx); reason != "" {
		return nil, reason
	}
	b := make([]byte, n)
	n2, typ2, reason := query(b)
	if st := stopped(ctx); st != "" {
		return nil, st
	}
	if reason != "" {
		return nil, reason
	}
	if n2 != n {
		return nil, "size_changed"
	}
	if typ2 != typ {
		return nil, "type_invalid"
	}
	return b, ""
}
