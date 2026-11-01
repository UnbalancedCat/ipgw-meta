package ctlcheck

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

func identifierSample(cms bool) []byte {
	return identifierFixture("1.2.3.40", []string{"1.2.3.20", "1.2.3.3"}, cms, []byte("private-synthetic-value"))
}

// Reframe one synthetic TLV along a path, preserving lengths at every level.
func identifierMutate(b []byte, path []int, change func(node) []byte) []byte {
	n, _, e := take(b)
	if e != nil {
		panic(e)
	}
	if len(path) == 0 {
		return change(n)
	}
	parts, e := children(n, n.tag, 32)
	if e != nil {
		panic(e)
	}
	raw := make([][]byte, len(parts))
	for i, p := range parts {
		raw[i] = p.raw
	}
	raw[path[0]] = identifierMutate(raw[path[0]], path[1:], change)
	return der(n.tag, raw...)
}

func TestIdentifiersBothEncapsulationsAndNoValues(t *testing.T) {
	for _, cms := range []bool{false, true} {
		member, ids, reason := locateIdentifiers(context.Background(), identifierSample(cms))
		if member != "1.2.3.40" || !reflect.DeepEqual(ids, []string{"1.2.3.3", "1.2.3.20"}) || reason != "" {
			t.Fatal(member, ids, reason)
		}
	}
	r := IdentifiersSelfTest(context.Background())
	b, e := EncodeIdentifiers(r)
	if e != nil || bytes.Contains(b, []byte("private")) || r.Signature != "not_checked" || r.InputAuthenticity != "unverified" {
		t.Fatal("unsafe report")
	}
	decoded, e := DecodeIdentifiers(b)
	if e != nil || !reflect.DeepEqual(decoded, r) {
		t.Fatal("round trip")
	}
}

func TestIdentifierOpaqueRoles(t *testing.T) {
	// Opaque certificate content and attribute values are deliberately not decoded.
	b := identifierMutate(identifierSample(false), []int{1, 0, 3}, func(node) []byte { return der(0xa0, []byte{0xff, 0}) })
	b = identifierMutate(b, []int{1, 0, 4, 0, 3, 0, 1, 0}, func(node) []byte { return der(0x30, []byte{0xff, 0}) })
	if _, _, reason := locateIdentifiers(context.Background(), b); reason != "" {
		t.Fatal(reason)
	}
}

func TestIdentifierStructureRejectionsClearAll(t *testing.T) {
	signerAttrs := []int{1, 0, 4, 0, 3}
	cases := map[string][]byte{
		"truncated":  identifierSample(false)[:30],
		"trailing":   append(identifierSample(false), 0),
		"indefinite": {0x30, 0x80, 0, 0},
		"over-blob":  make([]byte, maxBlob+1),
	}
	mutate := func(name string, path []int, fn func(node) []byte) {
		cases[name] = identifierMutate(identifierSample(false), path, fn)
	}
	mutate("multi-signer", []int{1, 0, 4}, func(n node) []byte { return der(0x31, n.body, n.body) })
	mutate("new-version", []int{1, 0, 0}, func(node) []byte { return encInt(4) })
	mutate("wrong-content-kind", []int{1, 0, 2, 0}, func(node) []byte { return encOID("1.2.3.4") })
	mutate("duplicate", signerAttrs, func(node) []byte { a := attribute("1.2.3.4", der(4, []byte{1})); return der(0xa0, a, a) })
	mutate("order", signerAttrs, func(n node) []byte { p, _ := children(n, 0xa0, 16); return der(0xa0, p[1].raw, p[0].raw) })
	mutate("empty", signerAttrs, func(node) []byte { return der(0xa0) })
	mutate("empty-set", signerAttrs, func(node) []byte { return der(0xa0, der(0x30, encOID("1.2.3.4"), der(0x31))) })
	mutate("too-many-values", signerAttrs, func(node) []byte {
		return der(0xa0, der(0x30, encOID("1.2.3.4"), der(0x31, encInt(1), encInt(2), encInt(3), encInt(4), encInt(5))))
	})
	mutate("large-set", signerAttrs, func(node) []byte { return der(0xa0, attribute("1.2.3.4", der(4, make([]byte, 64<<10)))) })
	mutate("malformed-oid", []int{1, 0, 4, 0, 3, 0, 0}, func(node) []byte { return der(6, []byte{0x80, 0x2a}) })
	// Use bounded valid arcs for the count limit.
	ids := []string{"1.2.3.1", "1.2.3.2", "1.2.3.3", "1.2.3.4", "1.2.3.5", "1.2.3.6", "1.2.3.7", "1.2.3.8", "1.2.3.9", "1.2.3.10", "1.2.3.11", "1.2.3.12", "1.2.3.13", "1.2.3.14", "1.2.3.15", "1.2.3.16", "1.2.3.17"}
	cases["too-many-attributes"] = identifierFixture("1.2.3.40", ids, false, nil)
	for name, b := range cases {
		t.Run(name, func(t *testing.T) {
			m, a, r := locateIdentifiers(context.Background(), b)
			if r == "" || m != "" || len(a) != 0 {
				t.Fatal("partial or accepted")
			}
		})
	}
}

func TestIdentifierTextBounds(t *testing.T) {
	for _, s := range []string{"", "1", "3.1", "1.40", "1.02.3", "+1.2", "1.2.", "1..2", "1.2.3\n", "１.2", "1.2." + strings.Repeat("9", 100), "1.2" + strings.Repeat(".1", 130)} {
		if _, ok := textOID(s); ok {
			t.Fatalf("accepted invalid OID text")
		}
	}
	for _, s := range []string{"0.0", "1.2.3", "2.999.1"} {
		if _, ok := textOID(s); !ok {
			t.Fatal(s)
		}
	}
	n, _ := whole(der(6, bytes.Repeat([]byte{1}, 129)), 6)
	if _, e := identifierOID(n); e == nil {
		t.Fatal("DER bound")
	}
}

func TestIdentifierCancellationReadAndProtection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	read := func(context.Context, int) ([]byte, string) { calls++; return identifierSample(false), "" }
	r := auditIdentifiers(ctx, "local-ctl-identifiers", read)
	if calls != 0 || r.Sources[0].Reason != "canceled" || !ValidIdentifiers(r) {
		t.Fatal("cancel read")
	}
	work := false
	r = protectedIdentifiers(context.Background(), func() bool { return false }, func() IdentifierReport { work = true; return IdentifiersSelfTest(context.Background()) })
	if work || r.Sources[0].Reason != "protection_failed" {
		t.Fatal("protection bypass")
	}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	r = auditIdentifiers(ctx, "local-ctl-identifiers", func(_ context.Context, i int) ([]byte, string) {
		if i == 1 {
			cancel()
		}
		return identifierSample(false), ""
	})
	for _, s := range r.Sources {
		if s.Reason != "canceled" || s.MemberAlgorithmOID != "" || len(s.SignedAttributeOIDs) != 0 {
			t.Fatal("partial cancel")
		}
	}
	ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	if m, a, r := locateIdentifiers(ctx, identifierSample(false)); m != "" || len(a) != 0 || r != "deadline" {
		t.Fatal("deadline")
	}
	for _, reason := range []string{"absent", "read_failed", "type_invalid", "size_changed", "limit"} {
		r = auditIdentifiers(context.Background(), "local-ctl-identifiers", func(context.Context, int) ([]byte, string) { return nil, reason })
		if !ValidIdentifiers(r) || r.Sources[0].Reason != reason {
			t.Fatal(reason)
		}
	}
}

func TestIdentifierClosedReport(t *testing.T) {
	r := IdentifiersSelfTest(context.Background())
	b, _ := EncodeIdentifiers(r)
	for _, mutate := range []func(*IdentifierReport){
		func(r *IdentifierReport) { r.SchemaVersion = 2 }, func(r *IdentifierReport) { r.Authentication = "passed" }, func(r *IdentifierReport) { r.Signature = "verified" },
		func(r *IdentifierReport) { r.Sources[0].SignedAttributeOIDs = nil }, func(r *IdentifierReport) { r.Sources[0].SignedAttributeOIDs = []string{} },
		func(r *IdentifierReport) { r.Sources[0].SignedAttributeOIDs = []string{"1.2.3.20", "1.2.3.3"} },
		func(r *IdentifierReport) {
			r.Sources[0].State = "unavailable"
			r.Sources[0].Reason = "parse_failed"
			r.State = "unavailable"
		},
		func(r *IdentifierReport) { r.Operations.NetworkRequests = 1 },
	} {
		var copy IdentifierReport
		json.Unmarshal(b, &copy)
		mutate(&copy)
		raw, _ := json.Marshal(copy)
		if _, e := DecodeIdentifiers(raw); e == nil {
			t.Fatal("invalid report accepted")
		}
	}
	for _, raw := range [][]byte{append([]byte(`{"extra":0,`), b[1:]...), append([]byte(`{"schema_version":1,`), b[1:]...), bytes.Replace(b, []byte(`"signature":"not_checked",`), nil, 1), append(b, []byte("{}")...)} {
		if _, e := DecodeIdentifiers(raw); e == nil {
			t.Fatal("open JSON")
		}
	}
}

func FuzzIdentifierReport(f *testing.F) {
	b, _ := EncodeIdentifiers(IdentifiersSelfTest(context.Background()))
	f.Add(b)
	f.Fuzz(func(t *testing.T, b []byte) {
		r, e := DecodeIdentifiers(b)
		if e == nil {
			encoded, e := EncodeIdentifiers(r)
			if e != nil || len(encoded) > 16384 {
				t.Fatal("bound")
			}
		}
	})
}
func FuzzIdentifierLocator(f *testing.F) {
	f.Add(identifierSample(false))
	f.Add(identifierSample(true))
	f.Add([]byte{0x30, 0x80})
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > maxBlob+1 {
			return
		}
		m, a, r := locateIdentifiers(context.Background(), b)
		if r != "" {
			if m != "" || len(a) != 0 {
				t.Fatal("partial")
			}
			return
		}
		if _, ok := textOID(m); !ok {
			t.Fatal("member")
		}
		if len(a) > 16 {
			t.Fatal("count")
		}
		for _, id := range a {
			if _, ok := textOID(id); !ok {
				t.Fatal("attribute")
			}
		}
	})
}
