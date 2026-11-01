package ctlcheck

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"testing"
)

func d13Content(f fixture, usage string, algorithm []byte, identifier []byte) []byte {
	return der(0x30, der(0x30, encOID(usage)), encTime(f.now), algorithm, der(0x30, der(0x30, der(4, identifier))))
}

// Shared matrix: same signed bytes must also pass the independent Windows API.
func checkD13Signatures(t *testing.T, f fixture, verify verifier) {
	t.Helper()
	for _, cms := range []bool{false, true} {
		for _, property := range []bool{false, true} {
			for _, opus := range []bool{false, true} {
				t.Run(fmt.Sprintf("cms=%v/property=%v/opus=%v", cms, property, opus), func(t *testing.T) {
					name, usage, algorithm, size := "authroot_ctl", oidRootList, oidSHA1, 20
					if property {
						name, usage, algorithm, size = "disallowed_ctl", oidDisallowedList, oidSignatureHashProperty, 32
					}
					id := bytes.Repeat([]byte{0x7e}, size)
					content := d13Content(f, usage, encAlg(algorithm), id)
					extras := [][]byte{}
					if opus {
						extras = append(extras, attribute(oidSPCOpus, der(0x30, der(0xa0, der(0x80, []byte{0, 0x51, 0, 0x52})))))
					}
					blob, e := f.signWithAttributes(content, cms, false, extras)
					if e != nil {
						t.Fatal(e)
					}
					r := observe(context.Background(), name, blob, f.mat, f.pins, f.now, verify)
					if !validSource(r) || r.State != "complete" || r.Structure != "bounded_checked" {
						t.Fatalf("%+v", r)
					}
					expectedMember, expectedAttr := "exact_length_checked", "known_subset_checked"
					if property {
						expectedMember = "bounded_uninterpreted"
					}
					if opus {
						expectedAttr = "spc_opus_outer_only"
					}
					if r.MemberIdentifierCoverage != expectedMember || r.SignedAttributeCoverage != expectedAttr {
						t.Fatal("coverage conflated")
					}
					m, reason := parseMessage(context.Background(), blob)
					if reason != "" {
						t.Fatal(reason)
					}
					// Content tampering must fail both observation and independently invoked verifier.
					changed := append([]byte(nil), id...)
					changed[0] ^= 1
					tampered := bytes.Replace(blob, id, changed, 1)
					changedContent := bytes.Replace(content, id, changed, 1)
					copyMessage := *m
					copyMessage.content = changedContent
					before, after := content, changedContent
					if !cms {
						a, _ := whole(content, 0x30)
						b, _ := whole(changedContent, 0x30)
						before, after = a.body, b.body
					}
					copyMessage.signed = bytes.Replace(m.signed, hashed(m.hash, before), hashed(m.hash, after), 1)
					if bytes.Equal(copyMessage.signed, m.signed) || verify(tampered, &copyMessage, f.leaf) || observe(context.Background(), name, tampered, f.mat, f.pins, f.now, verify).State == "complete" {
						t.Fatal("member mutation accepted")
					}
					if opus {
						tampered = bytes.Replace(blob, []byte{0, 0x51, 0, 0x52}, []byte{0, 0x51, 0, 0x53}, 1)
						changedMessage, reason := parseMessage(context.Background(), tampered)
						if reason != "" {
							t.Fatal(reason)
						}
						if bytes.Equal(m.signed, changedMessage.signed) || verify(tampered, changedMessage, f.leaf) {
							t.Fatal("SPC mutation excluded from signature")
						}
						rejected := observe(context.Background(), name, tampered, f.mat, f.pins, f.now, verify)
						if rejected.Reason != "signature_failed" || rejected.SignedAttributeCoverage != "unknown" || rejected.MemberIdentifierCoverage != "unknown" || rejected.Structure != "bounded_checked" {
							t.Fatalf("%+v", rejected)
						}
					}
				})
			}
		}
	}
	// Native decoding must independently tolerate each declared bounded identifier size.
	for _, size := range []int{1, 16, 20, 32, 48, 64} {
		for _, cms := range []bool{false, true} {
			for _, algorithm := range [][]byte{encAlg(oidSignatureHashProperty), der(0x30, encOID(oidSignatureHashProperty))} {
				content := d13Content(f, oidDisallowedList, algorithm, make([]byte, size))
				blob, e := f.signWithAttributes(content, cms, false, [][]byte{attribute(oidSPCOpus, der(0x30))})
				if e != nil {
					t.Fatal(e)
				}
				r := observe(context.Background(), "disallowed_ctl", blob, f.mat, f.pins, f.now, verify)
				if r.State != "complete" || !validSource(r) {
					t.Fatalf("boundary %d cms %v: %+v", size, cms, r)
				}
			}
		}
	}
}
func TestD13Signatures(t *testing.T) { checkD13Signatures(t, synthetic(t), pure) }

func TestD13RejectAndOpaqueBoundaries(t *testing.T) {
	f := synthetic(t)
	ctx := context.Background()
	good := d13Content(f, oidDisallowedList, encAlg(oidSignatureHashProperty), make([]byte, 32))
	for _, size := range []int{0, 65} {
		if _, e := parseCTL(ctx, d13Content(f, oidDisallowedList, encAlg(oidSignatureHashProperty), make([]byte, size))); e == nil {
			t.Fatal("member limit")
		}
	}
	for _, usage := range []string{oidRootList, "1.2.3.4"} {
		if _, e := parseCTL(ctx, d13Content(f, usage, encAlg(oidSignatureHashProperty), make([]byte, 32))); e == nil {
			t.Fatal("wrong usage")
		}
	}
	for _, algorithm := range [][]byte{der(0x30, encOID(oidSignatureHashProperty), der(4)), encAlg("1.2.3.4")} {
		if _, e := parseCTL(ctx, d13Content(f, oidDisallowedList, algorithm, make([]byte, 32))); e == nil {
			t.Fatal("algorithm")
		}
	}
	b, _ := f.signWithAttributes(good, false, false, [][]byte{attribute(oidSPCOpus, der(0x30))})
	called := false
	r := observe(ctx, "authroot_ctl", b, f.mat, f.pins, f.now, func([]byte, *message, *x509.Certificate) bool { called = true; return true })
	if called || r.State != "unavailable" || r.Structure != "unknown" {
		t.Fatal("source crossed")
	}
	one := attribute(oidSPCOpus, der(0x30))
	for name, extras := range map[string][][]byte{
		"duplicate": {one, one}, "wrong-outer": {attribute(oidSPCOpus, der(4))},
		"multiple-values": {der(0x30, encOID(oidSPCOpus), der(0x31, der(0x30), der(0x30)))},
		"empty-set":       {der(0x30, encOID(oidSPCOpus), der(0x31))},
		"large":           {attribute(oidSPCOpus, der(0x30, make([]byte, 64<<10)))},
		"unknown":         {attribute("1.2.3.9", der(0x30))},
		"truncated":       {attribute(oidSPCOpus, []byte{0x30, 2, 0})},
	} {
		t.Run(name, func(t *testing.T) {
			b, e := f.signWithAttributes(good, false, false, extras)
			if e != nil {
				t.Fatal(e)
			}
			r := observe(ctx, "disallowed_ctl", b, f.mat, f.pins, f.now, pure)
			if r.State == "complete" || !validSource(r) || r.SignedAttributeCoverage != "unknown" {
				t.Fatal("accepted or partial")
			}
		})
	}
	// These are opaque payloads, not evidence that their internal SPC syntax is valid.
	for _, value := range [][]byte{der(0x30), der(0x30, []byte{0xff, 0}), der(0x30, make([]byte, (64<<10)-4))} {
		b, _ := f.signWithAttributes(good, false, false, [][]byte{attribute(oidSPCOpus, value)})
		m, reason := parseMessage(ctx, b)
		if reason != "" || m.meta.signedCoverage != "spc_opus_outer_only" {
			t.Fatal(reason)
		}
	}
	for at := 1; at < 14; at++ {
		ctx := &cancelAt{Context: context.Background(), at: at}
		r := observe(ctx, "disallowed_ctl", b, f.mat, f.pins, f.now, pure)
		if r.State != "complete" && (r.MemberIdentifierCoverage != "unknown" || r.SignedAttributeCoverage != "unknown") {
			t.Fatal("cancel partial")
		}
	}
}

func TestD13WireVersionsAndCoverage(t *testing.T) {
	f := synthetic(t)
	s := observe(context.Background(), "authroot_ctl", f.blob, f.mat, f.pins, f.now, pure)
	r := emptyReport("self-test")
	r.State = "complete"
	r.Sources[0] = s
	r.Sources[1] = s
	r.Sources[1].Source = "disallowed_ctl"
	b, e := Encode(r)
	if e != nil {
		t.Fatal(e)
	}
	for _, version := range []int{1, 2} {
		old := r
		old.SchemaVersion = version
		old.Sources = append([]Source(nil), r.Sources...)
		for i := range old.Sources {
			old.Sources[i].Structure = "valid"
			old.Sources[i].MemberIdentifierCoverage = ""
			old.Sources[i].SignedAttributeCoverage = ""
			if version == 1 {
				old.Sources[i].MemberAlgorithm = ""
				old.Sources[i].Extensions = ""
			}
		}
		raw, e := Encode(old)
		if e != nil {
			t.Fatal(e)
		}
		decoded, e := Decode(raw)
		if e != nil {
			t.Fatal(e)
		}
		round, e := Encode(decoded)
		if e != nil || !bytes.Equal(raw, round) || bytes.Contains(raw, []byte("coverage")) {
			t.Fatal("legacy rewritten")
		}
		wrong := bytes.Replace(raw, []byte(fmt.Sprintf(`"schema_version":%d`, version)), []byte(`"schema_version":3`), 1)
		if _, e := Decode(wrong); e == nil {
			t.Fatal("missing v3 fields")
		}
	}
	for name, mutate := range map[string]func(*Report){
		"v2-new-fields": func(r *Report) { r.SchemaVersion = 2 }, "invalid-version": func(r *Report) { r.SchemaVersion = 4 },
		"old-structure":      func(r *Report) { r.Sources[0].Structure = "valid" },
		"missing-coverage":   func(r *Report) { r.Sources[0].MemberIdentifierCoverage = "unknown" },
		"wrong-sha-coverage": func(r *Report) { r.Sources[0].MemberIdentifierCoverage = "bounded_uninterpreted" },
		"premature":          func(r *Report) { r.Sources[0].Signature = "unknown" },
		"wrong-property-source": func(r *Report) {
			r.Sources[0].MemberAlgorithm = "signature_hash_property"
			r.Sources[0].MemberIdentifierCoverage = "bounded_uninterpreted"
		},
	} {
		t.Run(name, func(t *testing.T) {
			var copy Report
			json.Unmarshal(b, &copy)
			mutate(&copy)
			type plain Report
			raw, _ := json.Marshal(plain(copy))
			if _, e := Decode(raw); e == nil {
				t.Fatal("inconsistent report accepted")
			}
		})
	}
	for _, bad := range [][]byte{bytes.Replace(b, []byte(`,"member_identifier_coverage":"exact_length_checked"`), nil, 1), bytes.Replace(b, []byte(`"schema_version":3`), []byte(`"schema_version":3,"schema_version":3`), 1)} {
		if _, e := Decode(bad); e == nil {
			t.Fatal("incomplete/duplicate fields")
		}
	}
}
