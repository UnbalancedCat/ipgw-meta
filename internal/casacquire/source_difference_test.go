package casacquire

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
)

func sourceFixture(t *testing.T, entries map[int][][]byte) *sourceSnapshot {
	t.Helper()
	s, gate := readSources(context.Background(), func(_ context.Context, i int, add func([]byte) string) (string, string) {
		for _, der := range entries[i] {
			if reason := add(der); reason != "" {
				return "unavailable", reason
			}
		}
		return "complete", ""
	})
	if gate != nil {
		t.Fatalf("unexpected source gate: %s", gate.Reason)
	}
	return s
}
func graphFixture(certs ...*x509.Certificate) *differenceGraph {
	g := &differenceGraph{graph: relationGraph{certs: map[[32]byte]*x509.Certificate{}}, terminals: map[[32]byte]bool{}}
	for _, c := range certs {
		id := sha256.Sum256(c.Raw)
		g.graph.nodes = append(g.graph.nodes, id)
		g.graph.certs[id] = c
		g.terminals[id] = true
	}
	return g
}

func TestSourceReaderCoverageAndOwnership(t *testing.T) {
	c := generateChain(t, func(*x509.Certificate, *x509.Certificate) {})
	der := bytes.Clone(c.root.Raw)
	s := sourceFixture(t, map[int][][]byte{0: {der, []byte{1}}, 2: {[]byte{1}}, 4: {[]byte{1}}})
	if s.states[0].ParseCoverage != "partial" || s.states[2].ParseCoverage != "none" || s.states[4].ParseCoverage != "not_applicable" || len(s.denied[4]) != 1 {
		t.Fatal("coverage or raw denial lost")
	}
	id := sha256.Sum256(der)
	der[0] ^= 255
	if !bytes.Equal(s.certs[0][id].Raw, c.root.Raw) {
		t.Fatal("provider memory retained")
	}
}

func TestSourceGateFailuresAreZeroNetwork(t *testing.T) {
	for _, name := range []string{"empty", "absent", "bad-only", "read", "zero", "oversize", "per-store", "aggregate", "bytes", "ignored-callback", "invalid-state", "absent-with-data", "canceled", "deadline"} {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			if name == "canceled" {
				c, cancel := context.WithCancel(ctx)
				cancel()
				ctx = c
			}
			if name == "deadline" {
				c, cancel := context.WithDeadline(ctx, time.Now().Add(-time.Second))
				defer cancel()
				ctx = c
			}
			calls := 0
			r := acquireWithSources(ctx, nil, func(_ context.Context, i int, add func([]byte) string) (string, string) {
				calls++
				state, n, der := "complete", 0, []byte{1}
				switch name {
				case "absent":
					state = "absent"
				case "bad-only":
					if !denialSource(i) {
						n = 1
					}
				case "read":
					return "unavailable", "read_failed"
				case "zero":
					n, der = 1, nil
				case "oversize":
					n, der = 1, make([]byte, (64<<10)+1)
				case "per-store":
					n = 513
				case "aggregate":
					n = 512
				case "bytes":
					n, der = 257, make([]byte, 64<<10)
				case "ignored-callback":
					add(nil)
					return "complete", ""
				case "invalid-state":
					state = "invented"
				case "absent-with-data":
					state, n = "absent", 1
				}
				for j := 0; j < n; j++ {
					if reason := add(der); reason != "" {
						return "unavailable", reason
					}
				}
				return state, ""
			})
			if r.Attempts != (cascollect.Attempts{}) || r.ContextConstructed {
				t.Fatal("gate reached transport")
			}
			if _, err := Encode(r); err != nil {
				t.Fatal("gate report invalid")
			}
			want := "source_gate_unavailable"
			if name == "empty" || name == "absent" || name == "bad-only" {
				want = "source_gate_no_material"
			}
			if r.Reason != want {
				t.Fatalf("got %s", r.Reason)
			}
			if (name == "canceled" || name == "deadline") && calls != 0 {
				t.Fatal("provider called after cancellation")
			}
			if name == "aggregate" && calls != 5 {
				t.Fatal("aggregate duplicate budget not charged")
			}
		})
	}
}

func TestSourceComparisonFactsAndNoTrustMutation(t *testing.T) {
	c := generateChain(t, func(*x509.Certificate, *x509.Certificate) {})
	s := sourceFixture(t, map[int][][]byte{0: {c.root.Raw}, 1: {c.leaf.Raw}, 2: {c.intermediate.Raw}, 4: {c.leaf.Raw}})
	g := graphFixture(c.leaf, c.intermediate)
	before := c.pool.Clone()
	r := compareSources(context.Background(), s, g, defaultDifferenceLimits)
	if r.Status != "complete" || !r.Sources[0].Facts.AdditionalSigningParent || !r.Sources[1].Facts.ReachableDEROverlap || !r.Sources[1].Facts.TerminalSameSubjectSPKI || !r.Sources[4].Facts.ReachableDEROverlap || r.Sources[2].Facts.AdditionalSigningParent {
		t.Fatal("incorrect one-hop or identity facts")
	}
	if !c.pool.Equal(before) || len(g.graph.nodes) != 2 || len(g.graph.certs) != 2 {
		t.Fatal("baseline expanded")
	}
	// A certificate already anywhere in the original graph is not additional.
	g.graph.certs[sha256.Sum256(c.root.Raw)] = c.root
	r = compareSources(context.Background(), s, g, defaultDifferenceLimits)
	if r.Sources[0].Facts.AdditionalSigningParent {
		t.Fatal("unreachable baseline member called additional")
	}
	// Input order cannot change existential facts.
	g.graph.nodes[0], g.graph.nodes[1] = g.graph.nodes[1], g.graph.nodes[0]
	r2 := compareSources(context.Background(), s, g, defaultDifferenceLimits)
	if !reflect.DeepEqual(r, r2) {
		t.Fatal("order changed observation")
	}
}

func TestSourceReaderCancellationDiscardsEarlierMaterial(t *testing.T) {
	c := generateChain(t, func(*x509.Certificate, *x509.Certificate) {})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	s, gate := readSources(ctx, func(_ context.Context, i int, add func([]byte) string) (string, string) {
		calls++
		if i == 0 {
			return "complete", add(c.root.Raw)
		}
		cancel()
		return "complete", ""
	})
	if s != nil || gate.Status != "unavailable" || gate.Reason != "canceled" || calls != 2 || gate.Sources[2].State != "not_checked" {
		t.Fatal("partial snapshot survived cancellation")
	}
	r := blocked("anonymous-once", "source_gate_unavailable")
	r.SourceDifference = gate
	if _, err := Encode(r); err != nil {
		t.Fatal("cancellation report rejected")
	}
	gate.Reason = "deadline"
	if _, err := Encode(r); err == nil {
		t.Fatal("inconsistent aggregate failure accepted")
	}
}

func TestSourceOnlyRawDenialIsComparableMaterial(t *testing.T) {
	s := sourceFixture(t, map[int][][]byte{4: {[]byte{1}}})
	if s.states[4].Presence != "present" || len(s.denied[4]) != 1 {
		t.Fatal("unparseable denial lost")
	}
}

func TestD9AcquisitionPreservesGateSnapshotOnDiscoveryStop(t *testing.T) {
	w, dns, requests := fixtureWire(t, false, "HTTP/1.1 200 OK\r\nContent-Type: text/html\r\nContent-Length: 0\r\n\r\n")
	loaded := 0
	r := acquireWithSources(context.Background(), w, func(_ context.Context, i int, add func([]byte) string) (string, string) {
		if dns.Load() != 0 || requests.Load() != 0 {
			t.Error("network before source gate completion")
		}
		loaded++
		if i == 4 {
			return "complete", add([]byte{1})
		}
		return "absent", ""
	})
	// An empty 200 has no redirect to follow: discovery must stop after one GET.
	if loaded != 7 || r.SourceDifference == nil || r.SourceDifference.Status != "not_run" || r.Collection != "blocked" || r.ContextConstructed || dns.Load() != 1 || requests.Load() != 1 {
		t.Fatalf("discovery stop lost source coverage: %s", r.Reason)
	}
	if _, err := Encode(r); err != nil {
		t.Fatal("discovery report rejected")
	}
}

func TestD9BaselineTLSVerifiedReport(t *testing.T) {
	s := sourceFixture(t, map[int][][]byte{4: {[]byte{1}}})
	r := blocked("anonymous-once", "source_baseline_tls_verified")
	r.Attempts = cascollect.Attempts{Requests: 1, DNS: 1, Connects: 1}
	r.Operations.ExternalRequests, r.Operations.DNS = 1, 1
	r.SourceDifference = sourceReport(s, "comparison", "not_run", "baseline_not_observed")
	b, err := Encode(r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Decode(b); err != nil {
		t.Fatal(err)
	}
	r.SourceDifference = nil
	if _, err = Encode(r); err == nil {
		t.Fatal("D9-only reason accepted without gate")
	}
}

func TestSourceSPKIRepresentation(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	issue := func(name string, serial int64, k *ecdsa.PrivateKey) *x509.Certificate {
		x := &x509.Certificate{SerialNumber: big.NewInt(serial), Subject: pkix.Name{CommonName: name}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
		der, err := x509.CreateCertificate(rand.Reader, x, x, &k.PublicKey, k)
		if err != nil {
			t.Fatal(err)
		}
		c, err := x509.ParseCertificate(der)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	base, same, different := issue("synthetic-a", 1, key), issue("synthetic-a", 2, key), issue("synthetic-b", 3, key)
	otherKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	other := issue("synthetic-a", 4, otherKey)
	s := sourceFixture(t, map[int][][]byte{0: {same.Raw}, 1: {different.Raw}, 2: {other.Raw}})
	r := compareSources(context.Background(), s, graphFixture(base), defaultDifferenceLimits)
	if r.Status != "complete" || !r.Sources[0].Facts.TerminalSameSubjectSPKI || !r.Sources[0].Facts.AdditionalSigningParent || !r.Sources[1].Facts.TerminalDifferentSubjectSPKI || r.Sources[1].Facts.AdditionalSigningParent || *r.Sources[2].Facts != (SourceFacts{}) {
		t.Fatal("SPKI, subject or signature conflated")
	}
}

func TestSourceComparisonFailClosed(t *testing.T) {
	c := generateChain(t, func(*x509.Certificate, *x509.Certificate) {})
	for _, name := range []string{"comparison", "bytes", "signature", "nodes", "invalid", "canceled", "deadline"} {
		t.Run(name, func(t *testing.T) {
			s := sourceFixture(t, map[int][][]byte{0: {c.leaf.Raw}, 1: {c.intermediate.Raw}})
			g, limit, ctx, want := graphFixture(c.leaf), defaultDifferenceLimits, context.Background(), "input_invalid"
			switch name {
			case "comparison":
				limit.comparisons, want = 1, "comparison_limit"
			case "bytes":
				limit.bytes, want = 1, "byte_limit"
			case "signature":
				limit.signatures, want = 0, "signature_limit"
			case "nodes":
				g.graph.nodes = make([][32]byte, 129)
				want = "node_limit"
			case "invalid":
				g.terminals[[32]byte{}] = true
			case "canceled":
				cc, cancel := context.WithCancel(ctx)
				cancel()
				ctx, want = cc, "canceled"
			case "deadline":
				cc, cancel := context.WithDeadline(ctx, time.Now().Add(-time.Second))
				defer cancel()
				ctx, want = cc, "deadline"
			}
			r := compareSources(ctx, s, g, limit)
			if r.Status != "unavailable" || r.Reason != want {
				t.Fatalf("got %s/%s", r.Status, r.Reason)
			}
			for _, source := range r.Sources {
				if source.Facts != nil {
					t.Fatal("partial facts escaped")
				}
			}
		})
	}
}

func TestD9VerifiedTLSStopsBeforeHTTP(t *testing.T) {
	w, dns, requests := fixtureWire(t, true, "HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")
	c := generateChain(t, func(*x509.Certificate, *x509.Certificate) {})
	w.sourceSnapshot = sourceFixture(t, map[int][][]byte{6: {c.root.Raw}})
	_, reason := w.fetch(context.Background(), cascollect.Request{Method: "GET", Target: "https://pass.neu.edu.cn/tpass/login"}, 16)
	if reason != "source_baseline_tls_verified" || dns.Load() != 1 || requests.Load() != 0 || w.sourceDifference != nil {
		t.Fatal("D9 continued after baseline TLS succeeded")
	}
}

func TestD9DoesNotTrustExtraRoot(t *testing.T) {
	w, _, requests := fixtureWire(t, true, "HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")
	w.testTLS = &tls.Config{RootCAs: x509.NewCertPool(), ServerName: "pass.neu.edu.cn", MinVersion: tls.VersionTLS12}
	c := generateChain(t, func(*x509.Certificate, *x509.Certificate) {})
	w.sourceSnapshot = sourceFixture(t, map[int][][]byte{0: {c.root.Raw}})
	_, reason := w.fetch(context.Background(), cascollect.Request{Method: "GET", Target: "https://pass.neu.edu.cn/tpass/login"}, 16)
	if reason != "tls_unknown_authority" || requests.Load() != 0 {
		t.Fatal("extra source bypassed original TLS failure")
	}
}

func sourceReportFixture() Report {
	s, _ := readSources(context.Background(), func(_ context.Context, i int, add func([]byte) string) (string, string) {
		if i == 4 {
			return "complete", add([]byte{1})
		}
		return "absent", ""
	})
	r := topologyReportFixture()
	r.SourceDifference = sourceReport(s, "comparison", "complete", "none")
	for i := range r.SourceDifference.Sources {
		r.SourceDifference.Sources[i].Facts = &SourceFacts{}
	}
	r.SourceDifference.Sources[4].Facts.ReachableDEROverlap = true
	return r
}
func TestD9ReportClosedSchema(t *testing.T) {
	r := sourceReportFixture()
	b, err := Encode(r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Decode(b); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"missing", "duplicate", "unknown", "null", "source-order", "facts-null", "denial-spki", "gate-count", "failure-reason"} {
		t.Run(name, func(t *testing.T) {
			var m map[string]any
			json.Unmarshal(b, &m)
			d := m["source_difference"].(map[string]any)
			rows := d["sources"].([]any)
			switch name {
			case "missing":
				delete(d, "phase")
			case "unknown":
				d["raw_certificate"] = "synthetic-canary"
			case "null":
				m["source_difference"] = nil
			case "source-order":
				rows[0], rows[1] = rows[1], rows[0]
			case "facts-null":
				rows[4].(map[string]any)["facts"] = nil
			case "denial-spki":
				rows[4].(map[string]any)["facts"].(map[string]any)["terminal_same_subject_spki"] = true
			case "gate-count":
				d["phase"] = "gate"
			case "failure-reason":
				d["reason"] = "read_failed"
			}
			x, _ := json.Marshal(m)
			if name == "duplicate" {
				x = bytes.Replace(x, []byte(`"phase":"comparison"`), []byte(`"phase":"comparison","phase":"comparison"`), 1)
			}
			if _, err := Decode(x); err == nil {
				t.Fatal("invalid report accepted")
			}
		})
	}
}
func FuzzSourceDifferenceReport(f *testing.F) {
	b, _ := Encode(sourceReportFixture())
	f.Add(b)
	r := acquireWithSources(context.Background(), nil, func(context.Context, int, func([]byte) string) (string, string) { return "absent", "" })
	b, _ = Encode(r)
	f.Add(b)
	f.Fuzz(func(t *testing.T, b []byte) {
		r, err := Decode(b)
		if err != nil {
			return
		}
		encoded, err := Encode(r)
		if err != nil {
			t.Fatal("accepted report cannot encode")
		}
		if _, err = Decode(encoded); err != nil {
			t.Fatal("roundtrip failed")
		}
	})
}
