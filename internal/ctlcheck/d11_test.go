package ctlcheck

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

var memberCases = []struct {
	oid, name string
	size      int
}{
	{oidSHA1, "sha1", 20}, {"2.16.840.1.101.3.4.2.1", "sha256", 32},
	{"2.16.840.1.101.3.4.2.2", "sha384", 48}, {"2.16.840.1.101.3.4.2.3", "sha512", 64},
}

func testExtension(id string, critical bool, value []byte) []byte {
	p := [][]byte{encOID(id)}
	if critical {
		p = append(p, []byte{1, 1, 255})
	}
	return der(0x30, append(p, der(4, value))...)
}
func testCTL(f fixture, algorithm []byte, size int, attributes, tail []byte) []byte {
	p := [][]byte{der(0x30, encOID(oidRootList)), encTime(f.now), algorithm,
		der(0x30, der(0x30, der(4, make([]byte, size)), attributes))}
	if tail != nil {
		p = append(p, tail)
	}
	return der(0x30, p...)
}

// Shared matrix also runs against the independent Windows decoder/verifier.
func checkD11Signatures(t *testing.T, f fixture, verify verifier) {
	t.Helper()
	for _, a := range memberCases {
		for _, cms := range []bool{false, true} {
			for _, omit := range []bool{false, true} {
				for _, mode := range []string{"absent", "present_uninterpreted", "critical_present_uninterpreted"} {
					t.Run(fmt.Sprintf("%s/cms=%v/omit=%v/%s", a.name, cms, omit, mode), func(t *testing.T) {
						var tail, attrs []byte
						if !omit {
							attrs = der(0x31, attribute("1.2.3.4", der(4, []byte("synthetic-attribute"))))
						}
						if mode != "absent" {
							tail = der(0xa0, der(0x30, testExtension("1.2.3.5", mode == "critical_present_uninterpreted", []byte("synthetic-extension"))))
						}
						content := testCTL(f, encAlg(a.oid), a.size, attrs, tail)
						b, e := f.signEncoding(content, cms, false)
						if e != nil {
							t.Fatal(e)
						}
						r := observe(context.Background(), "authroot_ctl", b, f.mat, f.pins, f.now, verify)
						if !validSource(r) || r.State != "complete" || r.MemberAlgorithm != a.name || r.Extensions != mode {
							t.Fatalf("%+v", r)
						}
						expected := "present_uninterpreted"
						if omit {
							expected = "absent"
						}
						if r.Attributes != expected {
							t.Fatal("attribute coverage conflated")
						}
						if mode != "absent" {
							bad := bytes.Replace(b, []byte("synthetic-extension"), []byte("synthetic-extensioN"), 1)
							tampered := bytes.Replace(content, []byte("synthetic-extension"), []byte("synthetic-extensioN"), 1)
							if observe(context.Background(), "authroot_ctl", bad, f.mat, f.pins, f.now, verify).State == "complete" {
								t.Fatal("accepted modified extension")
							}
							// Native verification independently rejects the same-byte-length mutation.
							original, reason := parseMessage(context.Background(), b)
							if reason != "" {
								t.Fatal(reason)
							}
							original.content = tampered
							before, after := content, tampered
							if !cms {
								originalNode, _ := whole(content, 0x30)
								changedNode, _ := whole(tampered, 0x30)
								before, after = originalNode.body, changedNode.body
							}
							previous := append([]byte(nil), original.signed...)
							original.signed = bytes.Replace(original.signed, hashed(original.hash, before), hashed(original.hash, after), 1)
							if bytes.Equal(previous, original.signed) {
								t.Fatal("mutation fixture did not replace signed digest")
							}
							if verify(bad, original, f.leaf) {
								t.Fatal("signature accepted tampering")
							}
						}
					})
				}
			}
		}
	}
}
func TestD11SignatureMatrix(t *testing.T) { checkD11Signatures(t, synthetic(t), pure) }

func TestD11StructuralBoundaries(t *testing.T) {
	f := synthetic(t)
	for _, a := range memberCases {
		for _, n := range []int{0, a.size - 1, a.size + 1} {
			if _, e := parseCTL(context.Background(), testCTL(f, encAlg(a.oid), n, nil, nil)); e == nil {
				t.Fatal("bad identifier length")
			}
		}
		if _, e := parseCTL(context.Background(), testCTL(f, der(0x30, encOID(a.oid)), a.size, nil, nil)); e != nil {
			t.Fatal("absent parameters")
		}
		if _, e := parseCTL(context.Background(), testCTL(f, der(0x30, encOID(a.oid), der(4)), a.size, nil, nil)); e == nil {
			t.Fatal("bad parameters")
		}
	}
	good := testExtension("1.2.3.4", false, []byte{1})
	many := [][]byte{}
	for i := 0; i < 33; i++ {
		many = append(many, testExtension(fmt.Sprintf("1.2.3.%d", i), false, []byte{1}))
	}
	for name, tail := range map[string][]byte{
		"duplicate":        der(0xa0, der(0x30, good, good)),
		"too-many":         der(0xa0, der(0x30, many...)),
		"large-value":      der(0xa0, der(0x30, testExtension("1.2.3.4", false, make([]byte, (64<<10)+1)))),
		"explicit-default": der(0xa0, der(0x30, der(0x30, encOID("1.2.3.4"), []byte{1, 1, 0}, der(4)))),
		"non-DER-boolean":  der(0xa0, der(0x30, der(0x30, encOID("1.2.3.4"), []byte{1, 1, 1}, der(4)))),
		"no-value":         der(0xa0, der(0x30, der(0x30, encOID("1.2.3.4"), []byte{1, 1, 255}))),
		"wrong-value":      der(0xa0, der(0x30, der(0x30, encOID("1.2.3.4"), der(5)))),
		"extra-field":      der(0xa0, der(0x30, der(0x30, encOID("1.2.3.4"), der(4), der(4)))),
		"extra-container":  der(0xa0, der(0x30, good), der(0x30)),
		"truncated":        {0xa0, 2, 0x30},
		"large-OID":        der(0xa0, der(0x30, der(0x30, der(6, bytes.Repeat([]byte{1}, 129)), der(4)))),
	} {
		t.Run(name, func(t *testing.T) {
			if _, e := parseCTL(context.Background(), testCTL(f, encAlg(oidSHA1), 20, nil, tail)); e == nil {
				t.Fatal("accepted malformed extension")
			}
		})
	}
	for _, tail := range [][]byte{der(0xa0, der(0x30)), der(0xa0, der(0x30, many[:32]...)), der(0xa0, der(0x30, testExtension("1.2.3.4", false, make([]byte, 64<<10))))} {
		if _, e := parseCTL(context.Background(), testCTL(f, encAlg(oidSHA1), 20, nil, tail)); e != nil {
			t.Fatal("valid boundary rejected")
		}
	}
	// Direct parser test exceeds the cumulative content budget; production input
	// is also independently limited to 4 MiB before reaching this function.
	entry := der(0x30, der(4, make([]byte, 20)), der(0x31, attribute("1.2.3.4", der(4, make([]byte, (64<<10)-8)))))
	oversized := der(0x30, der(0x30, encOID(oidRootList)), encTime(f.now), encAlg(oidSHA1), der(0x30, bytes.Repeat(entry, 129)))
	if _, e := parseCTL(context.Background(), oversized); e == nil {
		t.Fatal("aggregate budget ignored")
	}
	nearLimit := [][]byte{der(0x30, encOID(oidRootList)), encTime(f.now), encAlg(oidSHA1), der(0x30, bytes.Repeat(entry, 128))}
	if _, e := parseCTL(context.Background(), der(0x30, nearLimit...)); e != nil {
		t.Fatal("attribute-only fixture must remain below budget")
	}
	combined := append(nearLimit, der(0xa0, der(0x30, testExtension("1.2.3.5", false, make([]byte, 1024)))))
	if _, e := parseCTL(context.Background(), der(0x30, combined...)); e == nil || parseReason(e) != "parse_failed_ctl_extension_limit" {
		t.Fatal("extensions did not share attribute budget")
	}
	content := testCTL(f, encAlg(oidSHA1), 20, nil, der(0xa0, der(0x30, good)))
	for at := 1; at < 8; at++ {
		ctx := &cancelAt{Context: context.Background(), at: at}
		b, _ := f.sign(content)
		r := observe(ctx, "authroot_ctl", b, f.mat, f.pins, f.now, pure)
		if !validSource(r) || r.MemberAlgorithm != "unknown" || r.Extensions != "unknown" {
			t.Fatal("partial coverage after cancellation")
		}
	}
}

func TestD11VersionedWire(t *testing.T) {
	r := emptyReport("local-ctl-check")
	r.SchemaVersion = 2
	for i := range r.Sources {
		r.Sources[i].MemberIdentifierCoverage = ""
		r.Sources[i].SignedAttributeCoverage = ""
	}
	b, e := Encode(r)
	if e != nil {
		t.Fatal(e)
	}
	legacy := r
	legacy.SchemaVersion = 1
	legacy.Sources = append([]Source(nil), r.Sources...)
	for i := range legacy.Sources {
		legacy.Sources[i].MemberAlgorithm = ""
		legacy.Sources[i].Extensions = ""
	}
	old, e := Encode(legacy)
	if e != nil {
		t.Fatal(e)
	}
	if bytes.Contains(old, []byte("member_algorithm")) || bytes.Contains(old, []byte("extensions")) {
		t.Fatal("legacy fields invented")
	}
	decoded, e := Decode(old)
	if e != nil || decoded.Sources[0].MemberAlgorithm != "" {
		t.Fatal("legacy not preserved")
	}
	round, e := Encode(decoded)
	if e != nil || !bytes.Equal(round, old) {
		t.Fatal("legacy round trip")
	}
	for name, bad := range map[string][]byte{
		"v1-with-new":          bytes.Replace(b, []byte(`"schema_version":2`), []byte(`"schema_version":1`), 1),
		"v2-missing-new":       bytes.Replace(old, []byte(`"schema_version":1`), []byte(`"schema_version":2`), 1),
		"unknown-algorithm":    bytes.Replace(b, []byte(`"member_algorithm":"unknown"`), []byte(`"member_algorithm":"md5"`), 1),
		"premature-algorithm":  bytes.Replace(b, []byte(`"member_algorithm":"unknown"`), []byte(`"member_algorithm":"sha256"`), 1),
		"premature-extensions": bytes.Replace(b, []byte(`"extensions":"unknown"`), []byte(`"extensions":"absent"`), 1),
		"duplicate-version":    bytes.Replace(b, []byte(`"schema_version":2`), []byte(`"schema_version":1,"schema_version":2`), 1),
		"missing-field":        bytes.Replace(b, []byte(`,"extensions":"unknown"`), nil, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, e := Decode(bad); e == nil {
				t.Fatal("invalid wire accepted")
			}
		})
	}
	if _, e := json.Marshal(Report{}); e == nil {
		t.Fatal("invalid report marshaled")
	}
}

func TestD11CMSStopCategories(t *testing.T) {
	f := synthetic(t)
	bad := bytes.Replace(f.blob, encOID("1.2.840.113549.1.9.3"), encOID("1.2.840.113549.1.9.6"), 1)
	if _, reason := parseMessage(context.Background(), bad); reason != "unsupported_signed_attribute_unknown" {
		t.Fatal(reason)
	}
	for _, stage := range []string{"envelope", "signed_data", "digest_set", "digest_algorithm", "content_encoding", "signer_info", "signer_identifier", "signer_digest", "signed_attributes", "signed_attribute_values", "signed_attribute_unknown", "signature_algorithm", "signature_size"} {
		r := emptyReport("local-ctl-check")
		r.Sources[0].Read = "complete"
		r.Sources[0].Reason = "unsupported_" + stage
		b, e := Encode(r)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = Decode(b); e != nil {
			t.Fatal(e)
		}
		r.Sources[0].Structure = "valid"
		if Valid(r) {
			t.Fatal("unsupported CMS claimed valid structure")
		}
		r.Sources[0].Structure = "unknown"
		r.SchemaVersion = 1
		for i := range r.Sources {
			r.Sources[i].MemberAlgorithm = ""
			r.Sources[i].Extensions = ""
		}
		if Valid(r) {
			t.Fatal("new reason accepted in historical schema")
		}
	}
}
