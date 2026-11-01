package ctlcheck

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/x509"
	"encoding/binary"
	"math/big"
	"testing"
)

func TestMemberSignedRelationsAndPolicy(t *testing.T) {
	f := synthetic(t)
	for _, source := range []string{"authroot_ctl", "disallowed_ctl"} {
		key, e := syntheticMemberKey(source, f.leaf.Raw)
		if e != nil {
			t.Fatal(e)
		}
		other, e := syntheticMemberKey(source, f.root.Raw)
		if e != nil {
			t.Fatal(e)
		}
		for _, cms := range []bool{false, true} {
			for _, constraints := range []bool{false, true} {
				entry := der(0x30, der(4, key))
				want := "absent"
				if constraints {
					entry = der(0x30, der(4, key), der(0x31, attribute("1.2.3.4", der(4, []byte("synthetic-policy")))))
					want = "present_uninterpreted"
				}
				content := memberSyntheticContent(f, source, entry)
				if constraints {
					n, _ := whole(content, 0x30)
					content = der(0x30, n.body, der(0xa0, der(0x30, testExtension("1.2.3.5", true, []byte{1}))))
				}
				b, e := f.signWithAttributes(content, cms, false, [][]byte{attribute(oidSPCOpus, der(0x30))})
				if e != nil {
					t.Fatal(e)
				}
				r := observeSyntheticMember(context.Background(), source, b, f.leaf.Raw, f, pure)
				if r.state != "complete" || r.relation != "match" || r.entryAttributes != want || r.window != "unbounded" {
					t.Fatalf("%+v", r)
				}
				if constraints && r.extensions != "critical_present_uninterpreted" {
					t.Fatal("critical policy erased")
				}
				r = observeSyntheticMember(context.Background(), source, b, f.root.Raw, f, pure)
				if r.relation != "no_match" || r.entryAttributes != "not_applicable" {
					t.Fatal("false member")
				}
				// A duplicate elsewhere in the signed table is ambiguous even after a hit.
				b, _ = f.signWithAttributes(memberSyntheticContent(f, source, entry, der(0x30, der(4, other)), der(0x30, der(4, other))), cms, false, nil)
				r = observeSyntheticMember(context.Background(), source, b, f.leaf.Raw, f, pure)
				if r.reason != "duplicate_identifier" || r.relation != "unknown" || r.entryAttributes != "unknown" {
					t.Fatal("duplicate partial result")
				}
				b[len(b)-1] ^= 1
				if observeSyntheticMember(context.Background(), source, b, f.leaf.Raw, f, pure).reason != "authentication_unavailable" {
					t.Fatal("unsigned inference")
				}
			}
		}
	}
}

func TestMemberKeysDoNotProveCertificateValidity(t *testing.T) {
	f := synthetic(t)
	changed := append([]byte(nil), f.leaf.Raw...)
	changed[len(changed)-1] ^= 1
	a, _ := syntheticMemberKey("authroot_ctl", f.leaf.Raw)
	b, e := syntheticMemberKey("authroot_ctl", changed)
	if e != nil || bytes.Equal(a, b) {
		t.Fatal("DER identity omitted signature")
	}
	a, _ = syntheticMemberKey("disallowed_ctl", f.leaf.Raw)
	b, e = syntheticMemberKey("disallowed_ctl", changed)
	if e != nil || !bytes.Equal(a, b) {
		t.Fatal("TBS property included signature bytes")
	}
	// Reissue the same SPKI with different content: neither identifier is a key identity.
	template := *f.leaf
	template.SerialNumber = big.NewInt(900)
	reissued, e := x509.CreateCertificate(rand.Reader, &template, &template, &f.key.PublicKey, f.key)
	if e != nil {
		t.Fatal(e)
	}
	for _, source := range []string{"authroot_ctl", "disallowed_ctl"} {
		a, _ := syntheticMemberKey(source, f.leaf.Raw)
		b, e := syntheticMemberKey(source, reissued)
		if e != nil || bytes.Equal(a, b) {
			t.Fatal("SPKI used as member identity")
		}
	}
}

func TestMemberCertificateAlgorithmBoundaries(t *testing.T) {
	f := synthetic(t)
	for _, a := range []x509.SignatureAlgorithm{x509.SHA256WithRSA, x509.SHA384WithRSA, x509.SHA512WithRSA, x509.SHA256WithRSAPSS} {
		c := *f.leaf
		c.SignatureAlgorithm = a
		raw, e := x509.CreateCertificate(rand.Reader, &c, &c, &f.key.PublicKey, f.key)
		if e != nil {
			t.Fatal(e)
		}
		key, e := syntheticMemberKey("disallowed_ctl", raw)
		if a == x509.SHA256WithRSAPSS {
			if e == nil {
				t.Fatal("PSS accepted without independent evidence")
			}
			continue
		}
		expected := map[x509.SignatureAlgorithm]int{x509.SHA256WithRSA: 32, x509.SHA384WithRSA: 48, x509.SHA512WithRSA: 64}[a]
		if e != nil || len(key) != expected {
			t.Fatal("hash width")
		}
	}
	badOuter := identifierMutate(f.leaf.Raw, []int{1}, func(node) []byte { return encAlg("1.2.840.113549.1.1.12") })
	badParams := identifierMutate(f.leaf.Raw, []int{1}, func(node) []byte { return der(0x30, encOID("1.2.840.113549.1.1.11"), der(4)) })
	badParams = identifierMutate(badParams, []int{0, 2}, func(node) []byte { return der(0x30, encOID("1.2.840.113549.1.1.11"), der(4)) })
	for _, raw := range [][]byte{nil, make([]byte, (64<<10)+1), badOuter, badParams, append(append([]byte(nil), f.leaf.Raw...), 0)} {
		if _, e := syntheticMemberKey("disallowed_ctl", raw); e == nil {
			t.Fatal("candidate shape")
		}
	}
}

func TestMemberSourceScopeCancellationAndLimits(t *testing.T) {
	f := synthetic(t)
	key, _ := syntheticMemberKey("disallowed_ctl", f.leaf.Raw)
	b, _ := f.sign(memberSyntheticContent(f, "disallowed_ctl", der(0x30, der(4, key))))
	if observeSyntheticMember(context.Background(), "authroot_ctl", b, f.leaf.Raw, f, pure).relation != "unknown" {
		t.Fatal("wrong source")
	}
	b, _ = f.sign(memberSyntheticContent(f, "authroot_ctl"))
	if observeSyntheticMember(context.Background(), "disallowed_ctl", b, f.leaf.Raw, f, pure).reason != "index_unsupported" {
		t.Fatal("wrong index")
	}
	for _, blob := range [][]byte{nil, make([]byte, maxBlob+1)} {
		if observeSyntheticMember(context.Background(), "authroot_ctl", blob, f.leaf.Raw, f, pure).relation != "unknown" {
			t.Fatal("limit")
		}
	}
	for at := 1; at < 20; at++ {
		ctx := &cancelAt{Context: context.Background(), at: at}
		r := observeSyntheticMember(ctx, "authroot_ctl", b, f.leaf.Raw, f, pure)
		if r.state != "complete" && (r.relation != "unknown" || r.entryAttributes != "unknown") {
			t.Fatal("partial cancellation")
		}
	}
}

func FuzzMemberKey(f *testing.F) {
	generated, e := generateFixture()
	if e != nil {
		f.Fatal(e)
	}
	f.Add(generated.leaf.Raw)
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > (64<<10)+1 {
			return
		}
		for _, source := range []string{"authroot_ctl", "disallowed_ctl"} {
			key, e := syntheticMemberKey(source, b)
			if e == nil && len(key) != 20 && len(key) != 32 && len(key) != 48 && len(key) != 64 {
				t.Fatal("unbounded key")
			}
		}
	})
}

func TestMemberWholeTableBoundary(t *testing.T) {
	f := synthetic(t)
	key, _ := syntheticMemberKey("authroot_ctl", f.leaf.Raw)
	entries := make([][]byte, 16385)
	for i := range entries {
		id := make([]byte, 20)
		binary.BigEndian.PutUint64(id[12:], uint64(i))
		entries[i] = der(0x30, der(4, id))
	}
	// A match at the end must not be lost by truncating a bounded table.
	entries[16383] = der(0x30, der(4, key))
	for _, count := range []int{16384, 16385} {
		blob, e := f.sign(memberSyntheticContent(f, "authroot_ctl", entries[:count]...))
		if e != nil {
			t.Fatal(e)
		}
		r := observeSyntheticMember(context.Background(), "authroot_ctl", blob, f.leaf.Raw, f, pure)
		if count == 16384 && (r.state != "complete" || r.relation != "match") {
			t.Fatal("last bounded member lost")
		}
		if count == 16385 && (r.state != "unavailable" || r.relation != "unknown") {
			t.Fatal("oversize table truncated")
		}
	}
}
