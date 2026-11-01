package casacquire

import (
	"bufio"
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
	"github.com/UnbalancedCat/ipgw-meta/internal/ctlcheck"
	"math/big"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
)

type generatedChain struct {
	root, intermediate, leaf *x509.Certificate
	key                      *ecdsa.PrivateKey
	pool                     *x509.CertPool
	at                       time.Time
}

func generateChain(t *testing.T, change func(*x509.Certificate, *x509.Certificate)) generatedChain {
	t.Helper()
	at := time.Now()
	template := func(serial int64, ca bool) *x509.Certificate {
		c := &x509.Certificate{SerialNumber: big.NewInt(serial), Subject: pkix.Name{CommonName: "synthetic-chain"}, NotBefore: at.Add(-time.Hour), NotAfter: at.Add(time.Hour), BasicConstraintsValid: true, IsCA: ca, KeyUsage: x509.KeyUsageDigitalSignature}
		if ca {
			c.KeyUsage |= x509.KeyUsageCertSign
		} else {
			c.DNSNames = []string{"pass.neu.edu.cn"}
			c.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
		}
		return c
	}
	issue := func(c, parent *x509.Certificate, signer *ecdsa.PrivateKey) (*x509.Certificate, *ecdsa.PrivateKey) {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatal("synthetic key generation failed")
		}
		if parent == nil {
			parent = c
			signer = key
		}
		der, err := x509.CreateCertificate(rand.Reader, c, parent, &key.PublicKey, signer)
		if err != nil {
			t.Fatal("synthetic certificate generation failed")
		}
		parsed, err := x509.ParseCertificate(der)
		if err != nil {
			t.Fatal("synthetic certificate parse failed")
		}
		return parsed, key
	}
	r, rk := issue(template(1, true), nil, nil)
	it, lt := template(2, true), template(3, false)
	change(it, lt)
	i, ik := issue(it, r, rk)
	l, lk := issue(lt, i, ik)
	pool := x509.NewCertPool()
	pool.AddCert(r)
	return generatedChain{r, i, l, lk, pool, at}
}

func failureFor(peer ...*x509.Certificate) error {
	return &tls.CertificateVerificationError{UnverifiedCertificates: peer, Err: x509.UnknownAuthorityError{Cert: peer[0]}}
}

func TestChainComparisonTrustAndBaseline(t *testing.T) {
	for _, name := range []string{"available", "empty", "missing-root", "foreign-intermediate", "self-signed-ca", "complete-peer", "wrong-host", "expired-leaf", "expired-intermediate", "wrong-usage", "denied-leaf", "denied-intermediate", "denied-root", "loader-failed", "subject-limit"} {
		t.Run(name, func(t *testing.T) {
			c := generateChain(t, func(i, l *x509.Certificate) {
				switch name {
				case "wrong-host":
					l.DNSNames = []string{"synthetic.invalid"}
				case "expired-leaf":
					l.NotAfter = l.NotBefore.Add(time.Minute)
				case "expired-intermediate":
					i.NotAfter = i.NotBefore.Add(time.Minute)
				case "wrong-usage":
					l.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
				}
			})
			roots := c.pool
			peer := []*x509.Certificate{c.leaf}
			cached := []*x509.Certificate{c.intermediate}
			denied := map[[32]byte]bool{}
			want := "tls_chain_cache_unknown_authority"
			loadFailure := ""
			switch name {
			case "available":
				want = "tls_chain_local_intermediates_available"
			case "empty":
				cached = nil
			case "missing-root":
				roots = x509.NewCertPool()
				cached = append(cached, c.root)
			case "foreign-intermediate":
				cached = []*x509.Certificate{generateChain(t, func(*x509.Certificate, *x509.Certificate) {}).intermediate}
			case "self-signed-ca":
				cached = []*x509.Certificate{generateChain(t, func(*x509.Certificate, *x509.Certificate) {}).root}
			case "complete-peer":
				peer = append(peer, c.intermediate)
				want = "tls_chain_probe_baseline_mismatch"
			case "wrong-host", "expired-leaf":
				want = "tls_chain_probe_baseline_mismatch"
			case "expired-intermediate":
				want = "tls_chain_cache_time_invalid"
			case "wrong-usage":
				want = "tls_chain_cache_usage_invalid"
			case "denied-leaf":
				denied[sha256.Sum256(c.leaf.Raw)] = true
				want = "tls_chain_probe_denied"
			case "denied-intermediate":
				denied[sha256.Sum256(c.intermediate.Raw)] = true
				want = "tls_chain_probe_denied"
			case "denied-root":
				denied[sha256.Sum256(c.root.Raw)] = true
				want = "tls_chain_probe_denied"
			case "loader-failed":
				loadFailure = "tls_chain_probe_store_read_failed"
				want = loadFailure
			case "subject-limit":
				for len(cached) <= maxSubjectCandidates {
					cached = append(cached, c.intermediate)
				}
				want = "tls_chain_probe_cache_subject_limit"
			}
			before := roots.Clone()
			calls := 0
			got := compareChain(context.Background(), failureFor(peer...), roots, denied, "pass.neu.edu.cn", c.at, func(context.Context) (cacheSnapshot, string) {
				calls++
				return cacheSnapshot{certs: cached}, loadFailure
			})
			if got != want {
				t.Fatalf("reason=%s want=%s", got, want)
			}
			if !roots.Equal(before) {
				t.Fatal("root pool mutated")
			}
			if (name == "complete-peer" || name == "wrong-host" || name == "expired-leaf") && calls != 0 {
				t.Fatal("cache loaded without baseline failure")
			}
		})
	}
}

func TestChainComparisonInputAndCancellationBounds(t *testing.T) {
	c := generateChain(t, func(*x509.Certificate, *x509.Certificate) {})
	for _, name := range []string{"canceled-before", "canceled-loader", "no-wrapper", "no-peer", "peer-count", "peer-bytes", "nil-peer", "cached-count", "cached-bytes", "nil-root", "nil-denied", "zero-time"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var failure error = failureFor(c.leaf)
			roots, denied, at := c.pool, map[[32]byte]bool{}, c.at
			cached := []*x509.Certificate{c.intermediate}
			switch name {
			case "canceled-before":
				cancel()
			case "no-wrapper":
				failure = opaqueError{}
			case "no-peer":
				failure = &tls.CertificateVerificationError{Err: x509.UnknownAuthorityError{}}
			case "nil-peer":
				failure = failureFor(nil)
			case "peer-count":
				peer := make([]*x509.Certificate, maxPeerCerts+1)
				for i := range peer {
					peer[i] = c.leaf
				}
				failure = failureFor(peer...)
			case "peer-bytes":
				bad := *c.leaf
				bad.Raw = make([]byte, maxProbeDER+1)
				failure = failureFor(&bad)
			case "cached-count":
				cached = make([]*x509.Certificate, maxCachedCerts+1)
			case "cached-bytes":
				bad := *c.intermediate
				bad.Raw = make([]byte, maxProbeDER+1)
				cached = []*x509.Certificate{&bad}
			case "nil-root":
				roots = nil
			case "nil-denied":
				denied = nil
			case "zero-time":
				at = time.Time{}
			}
			got := compareChain(ctx, failure, roots, denied, "pass.neu.edu.cn", at, func(context.Context) (cacheSnapshot, string) {
				if name == "canceled-loader" {
					cancel()
				}
				return cacheSnapshot{certs: cached}, ""
			})
			want := "tls_chain_probe_input_rejected"
			switch name {
			case "canceled-before", "canceled-loader":
				want = "tls_chain_probe_canceled"
			case "no-peer", "nil-peer":
				want = "tls_chain_probe_peer_invalid"
			case "peer-count", "peer-bytes":
				want = "tls_chain_probe_peer_limit"
			case "cached-count", "cached-bytes":
				want = "tls_chain_probe_cache_limit"
			}
			if got != want {
				t.Fatalf("reason=%s want=%s", got, want)
			}
		})
	}
}

func TestCachedCertificateOwnershipAndLimits(t *testing.T) {
	c := generateChain(t, func(*x509.Certificate, *x509.Certificate) {})
	ctx := context.Background()
	var cached cachedCertificates
	der := append([]byte(nil), c.intermediate.Raw...)
	if !cached.add(ctx, der) {
		t.Fatal("valid DER rejected")
	}
	der[0] ^= 0xff
	if !bytes.Equal(cached.certs[0].Raw, c.intermediate.Raw) {
		t.Fatal("enumeration buffer retained")
	}
	for _, data := range [][]byte{nil, make([]byte, maxProbeDER+1)} {
		var a cachedCertificates
		if a.add(ctx, data) {
			t.Fatal("invalid DER accepted")
		}
	}
	for _, a := range []cachedCertificates{{count: maxCachedCerts}, {bytes: maxCachedBytes}} {
		if a.add(ctx, c.intermediate.Raw) {
			t.Fatal("limit ignored")
		}
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if cached.add(canceled, c.intermediate.Raw) {
		t.Fatal("cancellation ignored")
	}
	for i := 1; i < maxCachedCerts; i++ {
		if !cached.add(ctx, c.intermediate.Raw) {
			t.Fatal("duplicate enumeration limit premature")
		}
	}
	if cached.add(ctx, c.intermediate.Raw) {
		t.Fatal("duplicate entries bypassed count")
	}
	bigCert := &x509.Certificate{Raw: make([]byte, maxProbeDER)}
	list := make([]*x509.Certificate, maxPeerBytes/maxProbeDER+1)
	for i := range list {
		list[i] = bigCert
	}
	if boundedCertificates(list, maxPeerCerts, maxPeerBytes) {
		t.Fatal("peer total ignored")
	}
	list = make([]*x509.Certificate, maxCachedBytes/maxProbeDER+1)
	for i := range list {
		list[i] = bigCert
	}
	if boundedCertificates(list, maxCachedCerts, maxCachedBytes) {
		t.Fatal("cached total ignored")
	}
}

func TestD4ClosesFailedSocketBeforeProbeAndNeverGETs(t *testing.T) {
	t.Run("complete-cache", func(t *testing.T) { checkClosedSocketProbe(t, false) })
	t.Run("partial-cache", func(t *testing.T) { checkClosedSocketProbe(t, true) })
	t.Run("relation-observation", func(t *testing.T) { checkClosedSocketProbe(t, true, true) })
	t.Run("topology-observation", func(t *testing.T) { checkClosedSocketProbe(t, true, true, true) })
	t.Run("source-difference", func(t *testing.T) { checkClosedSocketProbe(t, true, true, true, true) })
	t.Run("ctl-membership", func(t *testing.T) { checkClosedSocketProbe(t, true, true, true, true, true) })
}

func checkClosedSocketProbe(t *testing.T, partial bool, relation ...bool) {
	c := generateChain(t, func(*x509.Certificate, *x509.Certificate) {})
	dns, dnsCalls := dnsFixture(t, nil)
	closed := make(chan struct{})
	getCalls := &atomic.Int32{}
	connections := &atomic.Int32{}
	target, _ := rawFixture(t, false, func(raw net.Conn) {
		connections.Add(1)
		conn := tls.Server(raw, &tls.Config{Certificates: []tls.Certificate{{Certificate: [][]byte{c.leaf.Raw}, PrivateKey: c.key}}, MinVersion: tls.VersionTLS12})
		if conn.Handshake() == nil {
			if req, err := http.ReadRequest(bufio.NewReader(conn)); err == nil {
				getCalls.Add(1)
				req.Body.Close()
			}
		}
		// TLS consumed the alert; the raw connection must reach EOF before
		// the local loader is allowed to proceed.
		var b [1]byte
		if _, err := raw.Read(b[:]); err == nil {
			t.Error("unexpected bytes after failed handshake")
		}
		close(closed)
	})
	loads := 0
	w := &wire{local: netip.MustParseAddr("127.0.0.1"), dns: dns, testTarget: target, roots: c.pool, denied: map[[32]byte]bool{}, loadIntermediates: func(context.Context) (cacheSnapshot, string) {
		loads++
		select {
		case <-closed:
		case <-time.After(time.Second):
			t.Error("socket not closed before loader")
			return cacheSnapshot{}, "tls_chain_probe_loader_failed"
		}
		certs := []*x509.Certificate{c.intermediate}
		if len(relation) > 1 {
			certs = append(certs, c.root)
		}
		return cacheSnapshot{certs: certs, partial: partial}, ""
	}}
	if len(relation) != 0 {
		// The generated leaf/intermediate remain usable relationships, but
		// this empty original root pool prevents a trusted chain.
		w.roots = x509.NewCertPool()
		w.rootCapture = &rootSnapshot{}
	}
	if len(relation) > 2 {
		w.sourceSnapshot = sourceFixture(t, map[int][][]byte{0: {c.root.Raw}, 4: {c.leaf.Raw}})
	}

	compared := 0
	if len(relation) > 3 {
		w.memberSession = &ctlcheck.MemberSession{}
		w.testMemberCompare = func(ctx context.Context, peer [][]byte) *ctlcheck.Membership {
			compared++
			select {
			case <-closed:
			case <-time.After(time.Second):
				t.Error("member comparison before close")
			}
			if len(peer) != 1 || !bytes.Equal(peer[0], c.leaf.Raw) {
				t.Error("not same-round peer")
			}
			return ctlcheck.MemberUnavailable("comparison", "unavailable", "peer_unsupported")
		}
	}
	_, reason := w.fetch(context.Background(), cascollect.Request{Method: "GET", Target: "https://pass.neu.edu.cn/tpass/login"}, 16)
	if len(relation) > 3 && (compared != 1 || w.memberResult == nil) {
		t.Fatal("missing closed-socket membership")
	}
	want := "tls_chain_local_intermediates_available"
	if partial {
		want = "tls_chain_partial_cache_available"
	}
	if len(relation) != 0 {
		want = "tls_chain_relation_observed"
		if len(relation) > 1 {
			want = "tls_chain_topology_observed"
			if !validTopology(want, w.chainRelation, w.chainTopology) || !w.chainTopology.Facts.SelfLoop.Present {
				t.Fatal("D7 socket path lost topology")
			}
		} else if !validRelation(want, w.chainRelation) {
			t.Fatal("D6 socket path did not produce a closed observation")
		}
	}
	if reason != want || loads != 1 || getCalls.Load() != 0 || connections.Load() != 1 || dnsCalls.Load() != 1 || w.counts != (cascollect.Attempts{Requests: 1, DNS: 1, Connects: 1}) {
		t.Fatalf("D4 boundary failed: %s", reason)
	}
	if len(relation) > 2 {
		d := w.sourceDifference
		if d == nil || d.Status != "complete" || !d.Sources[0].Facts.ReachableDEROverlap || !d.Sources[0].Facts.TerminalSameSubjectSPKI || d.Sources[0].Facts.AdditionalSigningParent || !d.Sources[4].Facts.ReachableDEROverlap {
			t.Fatal("same-round D9 observation missing or original root promoted")
		}
	}
}

func TestD4ReportClosed(t *testing.T) {
	for _, reason := range append([]string{"tls_chain_local_intermediates_available", "tls_chain_unresolved", "tls_chain_probe_denied", "tls_chain_partial_cache_available", "tls_chain_partial_cache_unresolved", "tls_chain_partial_cache_empty", probeUnavailable}, probeFailureCases...) {
		for _, constructed := range []bool{true, false} {
			r := blocked("anonymous-once", reason)
			r.ContextConstructed = constructed
			r.Attempts = cascollect.Attempts{Requests: 3, DNS: 3, Connects: 3}
			r.Operations.ExternalRequests = 3
			r.Operations.DNS = 3
			b, err := Encode(r)
			if err != nil {
				t.Fatal("D4 reason rejected")
			}
			if _, err = Decode(b); err != nil {
				t.Fatal("D4 roundtrip failed")
			}
			for _, canary := range []string{"synthetic-chain", "neu.edu.cn", "BEGIN CERTIFICATE"} {
				if strings.Contains(string(b), canary) {
					t.Fatal("certificate data leaked")
				}
			}
			for _, mutate := range []func(*Report){
				func(r *Report) { r.Collection = "complete" }, func(r *Report) { r.Mode = "self-test" }, func(r *Report) { r.Mode = "local-preflight" },
				func(r *Report) { r.Attempts.Connects-- }, func(r *Report) { r.Reason += "_raw-canary" },
				func(r *Report) {
					r.Attempts = cascollect.Attempts{}
					r.Operations.ExternalRequests = 0
					r.Operations.DNS = 0
				},
			} {
				bad := r
				mutate(&bad)
				b, _ := json.Marshal(bad)
				if _, err := Decode(b); err == nil {
					t.Fatal("D4 contradiction decoded")
				}
				if _, err := Encode(bad); err == nil {
					t.Fatal("D4 contradiction encoded")
				}
			}
		}
	}
}
