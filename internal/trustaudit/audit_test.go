package trustaudit

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestCoverageAndRejection(t *testing.T) {
	der, err := syntheticDER()
	if err != nil {
		t.Fatal(err)
	}
	other, _ := syntheticDER()
	for _, name := range []string{"empty", "absent", "all-good", "partial", "all-bad", "denied", "mixed-denial", "read-error", "denied-error", "canceled", "deadline"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if name == "canceled" {
				cancel()
			}
			if name == "deadline" {
				var end context.CancelFunc
				ctx, end = context.WithDeadline(ctx, time.Now().Add(-time.Second))
				defer end()
			}
			calls := 0
			read := func(ctx context.Context, kind string, location int, a *accumulator) (string, string) {
				calls++
				if name == "absent" {
					return "absent", ""
				}
				if name == "empty" {
					return "complete", ""
				}
				if name == "denied-error" && kind == "Disallowed" {
					return "unavailable", "read_failed"
				}
				if kind == "Disallowed" {
					if name == "denied" || name == "mixed-denial" {
						return "complete", a.add(ctx, der)
					}
					return "complete", a.add(ctx, []byte("unparseable-denial"))
				}
				if location == 1 && name == "read-error" {
					return "unavailable", "read_failed"
				}
				if name != "all-bad" {
					if reason := a.add(ctx, der); reason != "" {
						return "unavailable", reason
					}
				}
				if name == "partial" || name == "all-bad" {
					if reason := a.add(ctx, []byte("bad-DER")); reason != "" {
						return "unavailable", reason
					}
				}
				if name == "mixed-denial" {
					if reason := a.add(ctx, other); reason != "" {
						return "unavailable", reason
					}
				}
				return "complete", ""
			}
			r := audit(ctx, "local-trust-audit", read)
			if !Valid(r) {
				t.Fatalf("invalid %s: %+v", name, r)
			}
			switch name {
			case "empty", "absent":
				if r.Root.Presence != "empty" || r.Root.ParseCoverage != "empty" {
					t.Fatal("empty")
				}
			case "partial":
				if r.Root.ParseCoverage != "partial" {
					t.Fatal("partial")
				}
			case "all-bad":
				if r.Root.ParseCoverage != "none" || r.Comparison.RemainingParsedRoots != "empty" {
					t.Fatal("all-bad")
				}
			case "denied":
				if r.Comparison.ParsedRootDeniedOverlap != "present" || r.Comparison.RemainingParsedRoots != "empty" {
					t.Fatal("denial lost")
				}
			case "mixed-denial":
				if r.Comparison.ParsedRootDeniedOverlap != "present" || r.Comparison.RemainingParsedRoots != "present" {
					t.Fatal("mixed")
				}
			case "read-error", "denied-error":
				if r.Comparison.State != "unavailable" || r.State != "unavailable" {
					t.Fatal("partial facts escaped")
				}
			case "canceled", "deadline":
				if calls != 0 || r.State != "unavailable" {
					t.Fatal("read after cancellation")
				}
			default:
				if r.Root.ParseCoverage != "complete" || r.Comparison.RemainingParsedRoots != "present" {
					t.Fatal("good")
				}
			}
			b, err := Encode(r)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = Decode(b); err != nil {
				t.Fatal("roundtrip")
			}
			for _, canary := range []string{"synthetic-audit", "bad-DER", "unparseable-denial", "BEGIN CERTIFICATE"} {
				if strings.Contains(string(b), canary) {
					t.Fatal("data leak")
				}
			}
		})
	}
}
func TestBudgetsAndRawDenial(t *testing.T) {
	ctx := context.Background()
	bad := []byte("unparseable")
	a := &accumulator{raw: true}
	if a.add(ctx, bad) != "" || !a.hashes[sha256.Sum256(bad)] {
		t.Fatal("raw denial discarded")
	}
	for _, name := range []string{"count", "bytes", "single", "empty"} {
		t.Run(name, func(t *testing.T) {
			a := &accumulator{}
			data := bad
			want := "limit"
			switch name {
			case "count":
				for i := 0; i < maxEntries; i++ {
					if a.add(ctx, bad) != "" {
						t.Fatal("premature limit")
					}
				}
			case "bytes":
				a.bytes = maxBytes
			case "single":
				data = make([]byte, maxDER+1)
			case "empty":
				data = nil
				want = "invalid_entry"
			}
			if a.add(ctx, data) != want {
				t.Fatal("budget bypass")
			}
		})
	}
	s, hashes := readKind(ctx, "ROOT", func(ctx context.Context, _ string, location int, a *accumulator) (string, string) {
		for i := 0; i < 300; i++ {
			if reason := a.add(ctx, bad); reason != "" {
				return "unavailable", reason
			}
		}
		return "complete", ""
	})
	if s.Reason != "limit" || hashes != nil || s.Presence != "unknown" {
		t.Fatal("location reset combined budget")
	}
}
func TestReportClosed(t *testing.T) {
	r := SelfTest(context.Background())
	b, err := Encode(r)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*Report){
		func(r *Report) { r.Operations.DNS = 1 }, func(r *Report) { r.Mode = "anonymous-once" }, func(r *Report) { r.Root.ParseCoverage = "empty" }, func(r *Report) { r.Disallowed.ParseCoverage = "complete" }, func(r *Report) { r.Comparison.RemainingParsedRoots = "unknown" }, func(r *Report) { r.Root.CurrentUser = "not_checked" }, func(r *Report) { r.State = "unavailable" },
	} {
		var x Report
		json.Unmarshal(b, &x)
		change(&x)
		if _, err := Encode(x); err == nil {
			t.Fatal("bad report encoded")
		}
		data, _ := json.Marshal(x)
		if _, err := Decode(data); err == nil {
			t.Fatal("bad report decoded")
		}
	}
	for _, data := range []string{strings.Replace(string(b), `"network_requests":0`, `"network_requests":0,"network_requests":0`, 1), strings.Replace(string(b), `"dns":0,`, "", 1), strings.Replace(string(b), `"dns":0`, `"dns":null`, 1), strings.Replace(string(b), `"dns":0`, `"private":"canary","dns":0`, 1)} {
		if _, err := Decode([]byte(data)); err == nil {
			t.Fatal("open shape")
		}
	}
}
func FuzzTrustAuditReport(f *testing.F) {
	b, err := Encode(SelfTest(context.Background()))
	if err != nil {
		f.Fatal(err)
	}
	f.Add(b)
	f.Fuzz(func(t *testing.T, b []byte) {
		if r, err := Decode(b); err == nil {
			encoded, err := Encode(r)
			if err != nil {
				t.Fatal("encode")
			}
			if _, err = Decode(encoded); err != nil {
				t.Fatal("roundtrip")
			}
		}
	})
}
