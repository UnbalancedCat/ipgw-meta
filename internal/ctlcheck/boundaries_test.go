package ctlcheck

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/json"
	"testing"
	"time"
)

func TestEncodingDigestRules(t *testing.T) {
	f := synthetic(t)
	m, reason := parseMessage(context.Background(), f.blob)
	if reason != "" {
		t.Fatal(reason)
	}
	for _, cms := range []bool{false, true} {
		for _, wrong := range []bool{false, true} {
			b, e := f.signEncoding(m.content, cms, wrong)
			if e != nil {
				t.Fatal(e)
			}
			r := observe(context.Background(), "authroot_ctl", b, f.mat, f.pins, f.now, pure)
			if (r.State == "complete") == wrong || !validSource(r) {
				t.Fatalf("cms=%v wrong=%v %+v", cms, wrong, r)
			}
		}
	}
}
func TestReadExactlyOnceAndProtectedEntry(t *testing.T) {
	for _, tc := range []struct {
		name         string
		n, n2        int
		typ, typ2    uint32
		reason, want string
		calls        int
	}{
		{"ok", 8, 8, 3, 3, "", "", 2}, {"grow", 8, 9, 3, 3, "", "size_changed", 2}, {"shrink", 8, 7, 3, 3, "", "size_changed", 2},
		{"type-before", 8, 8, 1, 3, "", "type_invalid", 1}, {"type-after", 8, 8, 3, 1, "", "type_invalid", 2},
		{"empty", 0, 0, 3, 3, "", "limit", 1}, {"large", maxBlob + 1, 0, 3, 3, "", "limit", 1},
		{"error", 8, 8, 3, 3, "read_failed", "read_failed", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			b, r := readBinary(context.Background(), func(b []byte) (int, uint32, string) {
				calls++
				if calls == 1 {
					if b != nil {
						t.Fatal("not a size query")
					}
					return tc.n, tc.typ, tc.reason
				}
				if len(b) != tc.n {
					t.Fatal("not exact allocation")
				}
				return tc.n2, tc.typ2, ""
			})
			if r != tc.want || calls != tc.calls || (r != "" && b != nil) {
				t.Fatalf("reason=%s calls=%d", r, calls)
			}
		})
	}
	called := false
	r := runProtected(context.Background(), func() bool { return false }, func() Report { called = true; return Report{} })
	if called || !Valid(r) || r.Sources[0].Reason != "protection_failed" {
		t.Fatal("work after protection failure")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r = runProtected(ctx, func() bool { t.Fatal("protection after cancel"); return true }, func() Report { t.Fatal("work after cancel"); return Report{} })
	if !Valid(r) {
		t.Fatal("cancel report")
	}
}
func TestCertificateConditions(t *testing.T) {
	f := synthetic(t)
	for _, tc := range []struct {
		name   string
		change func(*x509.Certificate)
		want   string
	}{
		{"critical", func(c *x509.Certificate) { c.UnhandledCriticalExtensions = []asn1.ObjectIdentifier{{1, 2, 3, 4}} }, "conditions_rejected"},
		{"root-keyusage", func(c *x509.Certificate) { c.KeyUsage = 0 }, "conditions_rejected"},
		{"root-ca", func(c *x509.Certificate) { c.IsCA = false }, "conditions_rejected"},
		{"root-wrong-key", func(c *x509.Certificate) { c.PublicKey = f.leaf.PublicKey }, "signature_failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := *f.root
			tc.change(&c)
			mat := materials{roots: []*x509.Certificate{&c}, denied: f.mat.denied}
			r := observe(context.Background(), "authroot_ctl", f.blob, mat, f.pins, f.now, pure)
			if r.Reason != tc.want || !validSource(r) {
				t.Fatalf("%+v", r)
			}
		})
	}
	// Wrong-name identity is never used to replace a pinned certificate.
	c := *f.root
	c.Raw = append(append([]byte(nil), c.Raw...), 0)
	p := f.pins
	p.root = sha256.Sum256(f.root.Raw)
	r := observe(context.Background(), "authroot_ctl", f.blob, materials{roots: []*x509.Certificate{&c}}, p, f.now, pure)
	if r.Reason != "material_missing" {
		t.Fatal(r.Reason)
	}
}
func TestDERBudgetsAndSignerAmbiguity(t *testing.T) {
	f := synthetic(t)
	n, _ := whole(f.blob, 0x30)
	outer, _ := children(n, 0x30, 2)
	sdnode, _ := whole(outer[1].body, 0x30)
	sd, _ := children(sdnode, 0x30, 6)
	wrap := func(nodes []node) []byte {
		var raw [][]byte
		for _, n := range nodes {
			raw = append(raw, n.raw)
		}
		return der(0x30, outer[0].raw, der(0xa0, der(0x30, raw...)))
	}
	for _, version := range []int64{0, 2, 257, 258, 259} {
		badVersion := append([]node(nil), sd...)
		badVersion[0] = node{raw: encInt(version)}
		if _, reason := parseMessage(context.Background(), wrap(badVersion)); reason != "unsupported_signed_data" {
			t.Fatalf("version %d accepted: %s", version, reason)
		}
	}
	signers, _ := children(sd[4], 0x31, 1)
	copySD := append([]node(nil), sd...)
	copySD[4] = node{raw: der(0x31, signers[0].raw, signers[0].raw)}
	if _, reason := parseMessage(context.Background(), wrap(copySD)); reason != "signer_ambiguous" {
		t.Fatal(reason)
	}
	copySD = append([]node(nil), sd...)
	copySD[3] = node{raw: der(0xa0, f.ca.Raw, f.leaf.Raw, f.leaf.Raw)}
	r := observe(context.Background(), "authroot_ctl", wrap(copySD), f.mat, f.pins, f.now, pure)
	if r.Reason != "signer_ambiguous" {
		t.Fatal(r.Reason)
	}
	entry := der(0x30, der(4, make([]byte, 20)), der(0x31))
	content := func(entries []byte) []byte {
		return der(0x30, der(0x30, encOID(oidRootList)), encTime(f.now), encAlg(oidSHA1), der(0x30, entries))
	}
	if _, e := parseCTL(context.Background(), content(bytes.Repeat(entry, 16385))); e == nil {
		t.Fatal("entry budget ignored")
	}
	attr := attribute("1.2.3.4", der(4, []byte("synthetic")))
	entry = der(0x30, der(4, make([]byte, 20)), der(0x31, bytes.Repeat(attr, 33)))
	if _, e := parseCTL(context.Background(), content(entry)); e == nil {
		t.Fatal("attribute budget ignored")
	}
	entry = der(0x30, der(4, make([]byte, 20)), der(0x31, attribute("1.2.3.4", der(4, make([]byte, 64<<10)))))
	if _, e := parseCTL(context.Background(), content(entry)); e == nil {
		t.Fatal("value budget ignored")
	}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	if _, reason := parseMessage(ctx, f.blob); reason != "deadline" {
		t.Fatal(reason)
	}
}
func TestNestedReportShape(t *testing.T) {
	b, _ := Encode(emptyReport("local-ctl-check"))
	var r map[string]json.RawMessage
	json.Unmarshal(b, &r)
	var sources []map[string]json.RawMessage
	json.Unmarshal(r["sources"], &sources)
	delete(sources[0], "attributes")
	r["sources"], _ = json.Marshal(sources)
	bad, _ := json.Marshal(r)
	if _, e := Decode(bad); e == nil {
		t.Fatal("missing nested field accepted")
	}
	sources = append(sources, sources[1])
	r["sources"], _ = json.Marshal(sources)
	bad, _ = json.Marshal(r)
	if _, e := Decode(bad); e == nil {
		t.Fatal("extra source accepted")
	}
}

type cancelAt struct {
	context.Context
	calls, at int
}

func (c *cancelAt) Err() error {
	c.calls++
	if c.calls >= c.at {
		return context.Canceled
	}
	return nil
}
func TestCancellationNeverLeavesContradictoryFacts(t *testing.T) {
	f := synthetic(t)
	for _, denied := range []bool{false, true} {
		for _, expired := range []bool{false, true} {
			for at := 1; at <= 40; at++ {
				ctx := &cancelAt{Context: context.Background(), at: at}
				mat := materials{roots: f.mat.roots, denied: map[[32]byte]bool{sha256.Sum256(f.leaf.Raw): denied}}
				now := f.now
				if expired {
					now = now.Add(48 * time.Hour)
				}
				r := observe(ctx, "authroot_ctl", f.blob, mat, f.pins, now, pure)
				if !validSource(r) {
					t.Fatalf("cancel at=%d denied=%v expired=%v: %+v", at, denied, expired, r)
				}
			}
		}
	}
}
