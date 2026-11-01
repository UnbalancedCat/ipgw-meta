package ctlcheck

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"testing"
	"time"
)

func synthetic(t *testing.T) fixture {
	t.Helper()
	f, e := generateFixture()
	if e != nil {
		t.Fatal(e)
	}
	return f
}
func pure(_ []byte, m *message, c *x509.Certificate) bool { return verifyGo(m, c) }
func TestFixedHierarchyAndRejections(t *testing.T) {
	f := synthetic(t)
	good := observe(context.Background(), "authroot_ctl", f.blob, f.mat, f.pins, f.now, pure)
	if good.State != "complete" || good.Signature != "verified" || good.Usage != "root_list" || good.Window != "within" || !validSource(good) {
		t.Fatalf("unexpected classification: %+v", good)
	}
	cases := []struct {
		name   string
		data   []byte
		mat    materials
		p      pins
		now    time.Time
		reason string
	}{
		{"missing-root", f.blob, materials{}, f.pins, f.now, "material_missing"},
		{"wrong-pin", f.blob, f.mat, pins{}, f.now, "material_missing"},
		{"expired", f.blob, f.mat, f.pins, f.now.Add(48 * time.Hour), "certificate_time_invalid"},
		{"future", f.blob, f.mat, f.pins, f.now.Add(-48 * time.Hour), "certificate_time_invalid"},
		{"denied-leaf", f.blob, materials{f.mat.roots, map[[32]byte]bool{sha256.Sum256(f.leaf.Raw): true}}, f.pins, f.now, "denied"},
		{"denied-ca", f.blob, materials{f.mat.roots, map[[32]byte]bool{sha256.Sum256(f.ca.Raw): true}}, f.pins, f.now, "denied"},
		{"denied-root", f.blob, materials{f.mat.roots, map[[32]byte]bool{sha256.Sum256(f.root.Raw): true}}, f.pins, f.now, "denied"},
	}
	bad := append([]byte(nil), f.blob...)
	bad[len(bad)-1] ^= 1
	cases = append(cases, struct {
		name   string
		data   []byte
		mat    materials
		p      pins
		now    time.Time
		reason string
	}{"tampered-signature", bad, f.mat, f.pins, f.now, "signature_failed"})
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := observe(context.Background(), "authroot_ctl", c.data, c.mat, c.p, c.now, pure)
			if r.Reason != c.reason || r.State != "unavailable" || !validSource(r) {
				t.Fatalf("unexpected classification %+v", r)
			}
		})
	}
}
func TestContentAndAlgorithmRejection(t *testing.T) {
	f := synthetic(t)
	for name, transform := range map[string]func([]byte) []byte{
		"trailing":        func(b []byte) []byte { return append(b, 0) },
		"indefinite":      func(b []byte) []byte { return []byte{0x30, 0x80, 0, 0} },
		"content-corrupt": func(b []byte) []byte { return bytes.Replace(b, encOID(oidRootList), encOID(oidDisallowedList), 1) },
		"digest-weak": func(b []byte) []byte {
			return bytes.ReplaceAll(b, encOID("2.16.840.1.101.3.4.2.1"), encOID("2.16.840.1.101.3.4.2.4"))
		},
		"signer-id": func(b []byte) []byte {
			return bytes.Replace(b, der(0x30, f.leaf.RawIssuer, encInt(3)), der(0x30, f.leaf.RawIssuer, encInt(4)), 1)
		},
	} {
		t.Run(name, func(t *testing.T) {
			b := transform(append([]byte(nil), f.blob...))
			r := observe(context.Background(), "authroot_ctl", b, f.mat, f.pins, f.now, pure)
			if r.State == "complete" || !validSource(r) {
				t.Fatalf("accepted corrupt input %+v", r)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := observe(ctx, "authroot_ctl", f.blob, f.mat, f.pins, f.now, pure)
	if r.Reason != "canceled" {
		t.Fatal(r.Reason)
	}
}
func TestCTLWindowsAndAttributes(t *testing.T) {
	f := synthetic(t)
	for _, tc := range []struct {
		name       string
		this, next time.Time
		usage      string
		attr       bool
		want       string
	}{
		{"within", f.now.Add(-time.Minute), f.now.Add(time.Hour), oidRootList, false, "within"},
		{"expired", f.now.Add(-time.Hour), f.now.Add(-time.Minute), oidRootList, false, "expired"},
		{"future", f.now.Add(time.Minute), f.now.Add(time.Hour), oidRootList, false, "future"},
		{"unbounded", f.now.Add(-time.Minute), time.Time{}, oidRootList, false, "unbounded"},
		{"invalid", f.now.Add(-time.Minute), f.now.Add(-time.Hour), oidRootList, false, "invalid"},
		{"unknown-attrs", f.now.Add(-time.Minute), f.now.Add(time.Hour), oidDisallowedList, true, "within"},
		{"other-usage", f.now.Add(-time.Minute), f.now.Add(time.Hour), "1.2.3.4", false, "within"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parts := [][]byte{der(0x30, encOID(tc.usage)), encTime(tc.this)}
			if !tc.next.IsZero() {
				parts = append(parts, encTime(tc.next))
			}
			parts = append(parts, encAlg(oidSHA1))
			attrs := der(0x31)
			if tc.attr {
				attrs = der(0x31, attribute("1.2.3.4", der(4, []byte("synthetic-policy"))))
			}
			parts = append(parts, der(0x30, der(0x30, der(4, make([]byte, 20)), attrs)))
			blob, e := f.sign(der(0x30, parts...))
			if e != nil {
				t.Fatal(e)
			}
			r := observe(context.Background(), "authroot_ctl", blob, f.mat, f.pins, f.now, pure)
			if r.State != "complete" || r.Window != tc.want || !validSource(r) {
				t.Fatalf("%+v", r)
			}
			if tc.attr && r.Attributes != "present_uninterpreted" {
				t.Fatal(r.Attributes)
			}
		})
	}
}
func TestSourceIsolationAndLimits(t *testing.T) {
	calls := 0
	for _, reason := range []string{"absent", "type_invalid", "size_changed", "limit", "read_failed"} {
		r := audit(context.Background(), "local-ctl-check", func(context.Context, int) ([]byte, string) { return nil, reason }, func(context.Context) (materials, string) { calls++; return materials{}, "" }, pins{}, pure)
		if calls != 0 || r.State != "unavailable" || !Valid(r) {
			t.Fatalf("%s %+v", reason, r)
		}
	}
	f := synthetic(t)
	r := audit(context.Background(), "local-ctl-check", func(context.Context, int) ([]byte, string) { return f.blob, "" }, func(context.Context) (materials, string) { return f.mat, "" }, f.pins, pure)
	if r.State != "complete" || !Valid(r) {
		t.Fatalf("%+v", r)
	}
	for _, n := range []int{0, maxBlob + 1} {
		r := audit(context.Background(), "local-ctl-check", func(context.Context, int) ([]byte, string) { return make([]byte, n), "" }, func(context.Context) (materials, string) { t.Fatal("loader invoked"); return materials{}, "" }, pins{}, pure)
		if !Valid(r) || r.Sources[0].Reason != "limit" {
			t.Fatal(r)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r = audit(ctx, "local-ctl-check", func(context.Context, int) ([]byte, string) { t.Fatal("read after cancel"); return nil, "" }, func(context.Context) (materials, string) { t.Fatal("load after cancel"); return materials{}, "" }, pins{}, pure)
	if !Valid(r) || r.Sources[0].Reason != "canceled" {
		t.Fatal(r)
	}
}
func TestReportClosedAndNoLeak(t *testing.T) {
	f := synthetic(t)
	s := observe(context.Background(), "authroot_ctl", f.blob, f.mat, f.pins, f.now, pure)
	r := emptyReport("self-test")
	r.Sources[0] = s
	r.Sources[1] = s
	r.Sources[1].Source = "disallowed_ctl"
	r.State = "complete"
	b, e := Encode(r)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Decode(b); e != nil {
		t.Fatal(e)
	}
	for _, bad := range [][]byte{bytes.Replace(b, []byte(`"schema_version":3`), []byte(`"schema_version":3,"schema_version":3`), 1), bytes.Replace(b, []byte(`"schema_version":3`), []byte(`"extra":0,"schema_version":3`), 1), bytes.Replace(b, []byte(`"signature":"verified"`), []byte(`"signature":"unknown"`), 1), bytes.Replace(b, []byte(`"dns":0`), []byte(`"dns":1`), 1), bytes.Replace(b, []byte(`"latest":"not_established"`), []byte(`"latest":"current"`), 1)} {
		if _, e := Decode(bad); e == nil {
			t.Fatal("accepted inconsistent/unknown report")
		}
	}
	if bytes.Contains(b, []byte("synthetic-ctl-only")) || bytes.Contains(b, f.leaf.Raw) {
		t.Fatal("source material leaked")
	}
	var obj map[string]json.RawMessage
	json.Unmarshal(b, &obj)
	delete(obj, "revocation")
	bad, _ := json.Marshal(obj)
	if _, e := Decode(bad); e == nil {
		t.Fatal("missing field")
	}
}
func FuzzCTLReport(f *testing.F) {
	b, _ := Encode(emptyReport("local-ctl-check"))
	f.Add(b)
	f.Add([]byte(`{}`))
	f.Fuzz(func(t *testing.T, b []byte) {
		r, e := Decode(b)
		if e == nil {
			if !Valid(r) {
				t.Fatal("invalid accepted")
			}
			out, e := Encode(r)
			if e != nil {
				t.Fatal(e)
			}
			if _, e := Decode(out); e != nil {
				t.Fatal(e)
			}
		}
	})
}
func FuzzCTLDER(f *testing.F) {
	f.Add([]byte{0x30, 0})
	f.Add([]byte{0x30, 0x80, 0, 0})
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > maxBlob {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		parseMessage(ctx, b)
		parseCTL(ctx, b)
	})
}
