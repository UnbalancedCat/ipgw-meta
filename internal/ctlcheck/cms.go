package ctlcheck

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"math/big"
)

type message struct {
	content, signed, signature, issuer []byte
	serial                             *big.Int
	certs                              []*x509.Certificate
	hash                               crypto.Hash
	meta                               ctlMeta
}

func hashAlgorithm(s string) crypto.Hash {
	switch s {
	case "2.16.840.1.101.3.4.2.1":
		return crypto.SHA256
	case "2.16.840.1.101.3.4.2.2":
		return crypto.SHA384
	case "2.16.840.1.101.3.4.2.3":
		return crypto.SHA512
	}
	return 0
}
func hashed(h crypto.Hash, b []byte) []byte {
	switch h {
	case crypto.SHA256:
		s := sha256.Sum256(b)
		return s[:]
	case crypto.SHA384:
		s := sha512.Sum384(b)
		return s[:]
	case crypto.SHA512:
		s := sha512.Sum512(b)
		return s[:]
	}
	return nil
}
func parseMessage(ctx context.Context, b []byte) (result *message, reason string) {
	stage := "envelope"
	defer func() {
		if reason == "unsupported" {
			reason = "unsupported_" + stage
		}
	}()
	if len(b) == 0 || len(b) > maxBlob {
		return nil, "limit"
	}
	n, e := whole(b, 0x30)
	if e != nil {
		return nil, "parse_failed_envelope"
	}
	outer, e := children(n, 0x30, 2)
	if e != nil || len(outer) != 2 {
		return nil, "parse_failed_envelope"
	}
	s, e := oid(outer[0])
	if e != nil || s != oidSigned || outer[1].tag != 0xa0 {
		return nil, "unsupported"
	}
	n, e = whole(outer[1].body, 0x30)
	if e != nil {
		return nil, "parse_failed_signed_data"
	}
	stage = "signed_data"
	sd, e := children(n, 0x30, 6)
	if e != nil || len(sd) != 5 || (!bytes.Equal(sd[0].raw, []byte{2, 1, 1}) && !bytes.Equal(sd[0].raw, []byte{2, 1, 3})) {
		return nil, "unsupported"
	}
	stage = "digest_set"
	ds, e := children(sd[1], 0x31, 1)
	if e != nil || len(ds) != 1 {
		return nil, "unsupported"
	}
	stage = "digest_algorithm"
	ha, e := alg(ds[0])
	h := hashAlgorithm(ha)
	if e != nil || h == 0 {
		return nil, "unsupported"
	}
	enc, e := children(sd[2], 0x30, 2)
	if e != nil || len(enc) != 2 || enc[1].tag != 0xa0 {
		return nil, "parse_failed_content_info"
	}
	stage = "content_encoding"
	s, e = oid(enc[0])
	if e != nil || s != oidCTL {
		return nil, "unsupported"
	}
	content, rest, e := take(enc[1].body)
	if e != nil || len(rest) != 0 {
		return nil, "parse_failed_content"
	}
	m := &message{hash: h}
	// PKCS#7 carries the CTL SEQUENCE directly; CMS wraps its DER in OCTET STRING.
	// RFC 2315 hashes the value octets of the directly encoded content; RFC 5652
	// hashes the OCTET STRING payload. Select by encoding, never try both hashes.
	digestContent := content.body
	if content.tag == 4 {
		if !bytes.Equal(sd[0].raw, []byte{2, 1, 3}) {
			return nil, "unsupported"
		}
		m.content = content.body
	} else if content.tag == 0x30 {
		if !bytes.Equal(sd[0].raw, []byte{2, 1, 1}) {
			return nil, "unsupported"
		}
		m.content = content.raw
	} else {
		return nil, "unsupported"
	}
	m.meta, e = parseCTL(ctx, m.content)
	if e != nil {
		if reason := stopped(ctx); reason != "" {
			return nil, reason
		}
		return nil, parseReason(e)
	}
	certs, e := children(sd[3], 0xa0, 16)
	if e != nil || len(certs) == 0 {
		return nil, "material_missing"
	}
	total := 0
	for _, v := range certs {
		if stopped(ctx) != "" {
			return nil, stopped(ctx)
		}
		if v.tag != 0x30 || len(v.raw) > 64<<10 || len(v.raw) > (1<<20)-total {
			return nil, "limit"
		}
		total += len(v.raw)
		c, e := x509.ParseCertificate(v.raw)
		if e != nil {
			return nil, "parse_failed_certificate"
		}
		m.certs = append(m.certs, c)
	}
	signers, e := children(sd[4], 0x31, 2)
	if e != nil || len(signers) != 1 {
		return nil, "signer_ambiguous"
	}
	stage = "signer_info"
	si, e := children(signers[0], 0x30, 7)
	if e != nil || len(si) != 6 || !bytes.Equal(si[0].raw, []byte{2, 1, 1}) {
		return nil, "unsupported"
	}
	stage = "signer_identifier"
	id, e := children(si[1], 0x30, 2)
	if e != nil || len(id) != 2 || id[0].tag != 0x30 || !positive(id[1]) {
		return nil, "unsupported"
	}
	m.issuer = id[0].raw
	m.serial = new(big.Int).SetBytes(id[1].body)
	stage = "signer_digest"
	a, e := alg(si[2])
	if e != nil || a != ha {
		return nil, "unsupported"
	}
	stage = "signed_attributes"
	attrs, e := children(si[3], 0xa0, 16)
	if e != nil {
		return nil, "unsupported"
	}
	m.meta.signedCoverage = "known_subset_checked"
	seen := map[string]bool{}
	for i, attr := range attrs {
		if stopped(ctx) != "" {
			return nil, stopped(ctx)
		}
		if i > 0 && bytes.Compare(attrs[i-1].raw, attr.raw) > 0 {
			return nil, "parse_failed_signed_attribute_order"
		}
		pair, e := children(attr, 0x30, 2)
		if e != nil || len(pair) != 2 {
			return nil, "parse_failed_signed_attribute"
		}
		o, e := oid(pair[0])
		if e != nil || seen[o] {
			return nil, "parse_failed_signed_attribute_identifier"
		}
		seen[o] = true
		stage = "signed_attribute_values"
		vals, e := children(pair[1], 0x31, 1)
		if e != nil || len(vals) != 1 {
			return nil, "unsupported"
		}
		stage = "signed_attribute_unknown"
		switch o {
		case "1.2.840.113549.1.9.3":
			v, e := oid(vals[0])
			if e != nil || v != oidCTL {
				return nil, "parse_failed_content_type"
			}
		case "1.2.840.113549.1.9.4":
			if vals[0].tag != 4 || !bytes.Equal(vals[0].body, hashed(h, digestContent)) {
				return nil, "signature_failed"
			}
		case oidSPCOpus:
			if vals[0].tag != 0x30 {
				return nil, "parse_failed_signed_attribute"
			}
			if len(vals[0].raw) > 64<<10 || len(vals[0].raw) > (8<<20)-m.meta.visitedBytes {
				return nil, "limit"
			}
			m.meta.visitedBytes += len(vals[0].raw)
			m.meta.signedCoverage = "spc_opus_outer_only"
			// The complete outer TLV was checked by children; its contents stay opaque.
		case "1.2.840.113549.1.9.5":
			if _, e := date(vals[0]); e != nil {
				return nil, "parse_failed_signing_time"
			}
		default:
			return nil, "unsupported"
		}
	}
	if !seen["1.2.840.113549.1.9.3"] || !seen["1.2.840.113549.1.9.4"] {
		return nil, "parse_failed_required_attributes"
	}
	m.signed = append([]byte(nil), si[3].raw...)
	m.signed[0] = 0x31
	stage = "signature_algorithm"
	sa, e := alg(si[4])
	if e != nil {
		return nil, "unsupported"
	}
	expected := map[crypto.Hash]string{crypto.SHA256: "1.2.840.113549.1.1.11", crypto.SHA384: "1.2.840.113549.1.1.12", crypto.SHA512: "1.2.840.113549.1.1.13"}
	if sa != "1.2.840.113549.1.1.1" && sa != expected[h] {
		return nil, "unsupported"
	}
	stage = "signature_size"
	if si[5].tag != 4 || len(si[5].body) < 256 || len(si[5].body) > 1024 {
		return nil, "unsupported"
	}
	m.signature = si[5].body
	return m, ""
}
func verifyGo(m *message, c *x509.Certificate) bool {
	key, ok := c.PublicKey.(*rsa.PublicKey)
	return ok && key.N.BitLen() >= 2048 && key.N.BitLen() <= 8192 && rsa.VerifyPKCS1v15(key, m.hash, hashed(m.hash, m.signed), m.signature) == nil
}
