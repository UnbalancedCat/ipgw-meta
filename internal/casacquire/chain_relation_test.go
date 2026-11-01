package casacquire

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
)

func relationKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal("generated key failed")
	}
	return k
}

func relationTemplate(name string, ca bool) *x509.Certificate {
	return &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: name}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), BasicConstraintsValid: true, IsCA: ca, KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign, DNSNames: []string{"pass.neu.edu.cn"}, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
}

func relationIssue(t *testing.T, template, parent *x509.Certificate, key, signer *ecdsa.PrivateKey) *x509.Certificate {
	t.Helper()
	der, err := x509.CreateCertificate(rand.Reader, template, parent, &key.PublicKey, signer)
	if err != nil {
		t.Fatal("generated certificate failed")
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal("generated certificate parse failed")
	}
	return cert
}

func relationRoots(certs ...*x509.Certificate) (*x509.CertPool, *rootSnapshot) {
	p := x509.NewCertPool()
	s := &rootSnapshot{}
	for _, c := range certs {
		p.AddCert(c)
		s.add(c)
	}
	return p, s
}

type relationFixture struct {
	root, intermediate, leaf *x509.Certificate
	rk, ik, lk               *ecdsa.PrivateKey
}

func relationChain(t *testing.T) relationFixture {
	t.Helper()
	rk, ik, lk := relationKey(t), relationKey(t), relationKey(t)
	rt := relationTemplate("synthetic-root", true)
	r := relationIssue(t, rt, rt, rk, rk)
	i := relationIssue(t, relationTemplate("synthetic-intermediate", true), r, ik, rk)
	l := relationIssue(t, relationTemplate("synthetic-leaf", false), i, lk, ik)
	return relationFixture{r, i, l, rk, ik, lk}
}

func TestRelationObservedFactsAndOrder(t *testing.T) {
	c := relationChain(t)
	for _, name := range []string{"root-path", "no-leaf-parent", "no-upstream-parent", "bad-key", "non-ca", "algorithm", "akid-mismatch", "denied", "self-cycle-outside-roots", "cycle", "cross-sign-mixed"} {
		t.Run(name, func(t *testing.T) {
			peer := []*x509.Certificate{c.leaf}
			cached := cacheSnapshot{certs: []*x509.Certificate{c.intermediate}, partial: true}
			roots, snapshot := relationRoots(c.root)
			denied := map[[32]byte]bool{}
			want := RelationFacts{RootPathObserved: true}
			switch name {
			case "no-leaf-parent":
				cached.certs = nil
				want = RelationFacts{LeafNoParent: true}
			case "no-upstream-parent":
				roots, snapshot = relationRoots()
				want = RelationFacts{UpstreamNoParent: true}
			case "bad-key":
				cached.certs = []*x509.Certificate{relationIssue(t, relationTemplate("synthetic-intermediate", true), c.root, relationKey(t), c.rk)}
				want = RelationFacts{SignatureCheckFailed: true}
			case "non-ca":
				cached.certs = []*x509.Certificate{relationIssue(t, relationTemplate("synthetic-intermediate", false), c.root, c.ik, c.rk)}
				want = RelationFacts{CAConstraintRejected: true}
			case "algorithm":
				template := relationTemplate("synthetic-leaf", false)
				template.SignatureAlgorithm = x509.ECDSAWithSHA1
				peer = []*x509.Certificate{relationIssue(t, template, c.intermediate, c.lk, c.ik)}
				want = RelationFacts{AlgorithmRejected: true}
			case "akid-mismatch":
				parent := *c.intermediate
				parent.SubjectKeyId = []byte("synthetic-mismatched-id")
				peer = []*x509.Certificate{relationIssue(t, relationTemplate("synthetic-leaf", false), &parent, c.lk, c.ik)}
			case "denied":
				denied[sha256.Sum256(c.intermediate.Raw)] = true
				want.DeniedObserved = true
			case "self-cycle-outside-roots":
				// A normal self-signed chain tail supplied only as an intermediate
				// has the same D6 facts as the two-certificate cycle below.
				// Neither shape establishes membership in the original root pool.
				roots, snapshot = relationRoots()
				cached.certs = append(cached.certs, c.root)
				want = RelationFacts{CycleObserved: true}
			case "cycle":
				a, b := relationTemplate("cycle-a", true), relationTemplate("cycle-b", true)
				ak, bk := relationKey(t), relationKey(t)
				ac := relationIssue(t, a, b, ak, bk)
				bc := relationIssue(t, b, a, bk, ak)
				peer = []*x509.Certificate{relationIssue(t, relationTemplate("synthetic-leaf", false), ac, c.lk, ak)}
				cached.certs = []*x509.Certificate{ac, bc}
				want = RelationFacts{CycleObserved: true}
			case "cross-sign-mixed":
				otherKey := relationKey(t)
				other := relationTemplate("other-root", true)
				otherRoot := relationIssue(t, other, other, otherKey, otherKey)
				cross := relationIssue(t, relationTemplate("synthetic-intermediate", true), otherRoot, c.ik, otherKey)
				bad := relationIssue(t, relationTemplate("synthetic-intermediate", true), c.root, relationKey(t), c.rk)
				cached.certs = append(cached.certs, cross, bad)
				want.UpstreamNoParent = true
				want.SignatureCheckFailed = true
			}
			before := roots.Clone()
			for _, partial := range []bool{false, true} {
				cached.partial = partial
				var previous *ChainRelation
				for n := 0; n < 2; n++ {
					reason, observation := observeRelations(context.Background(), peer, cached, roots, snapshot, denied)
					if !validRelation(reason, observation) || observation.Facts == nil || *observation.Facts != want || !roots.Equal(before) {
						t.Fatalf("unexpected relation: %s %+v", reason, observation)
					}
					if previous != nil && !reflect.DeepEqual(previous, observation) {
						t.Fatal("input order changed relations")
					}
					previous = observation
					for i, j := 0, len(cached.certs)-1; i < j; i, j = i+1, j-1 {
						cached.certs[i], cached.certs[j] = cached.certs[j], cached.certs[i]
					}
				}
			}
		})
	}
}

func TestRelationLimitsDiscardPartialFacts(t *testing.T) {
	c := relationChain(t)
	for _, name := range []string{"root-mismatch", "root-limit", "nil-peer", "key", "nodes", "depth", "candidates", "signatures", "cancel", "deadline"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			roots, snapshot := relationRoots(c.root)
			peer := []*x509.Certificate{c.leaf}
			cached := cacheSnapshot{certs: []*x509.Certificate{c.intermediate}, partial: true}
			limits := defaultRelationLimits
			want := ""
			switch name {
			case "root-mismatch":
				snapshot = &rootSnapshot{}
				want = "root_snapshot_invalid"
			case "root-limit":
				snapshot.failure = "root_snapshot_limit"
				snapshot.certs = nil
				want = "root_snapshot_limit"
			case "nil-peer":
				peer = nil
				want = "input_invalid"
			case "key":
				bad := *c.intermediate
				bad.PublicKey = nil
				cached.certs = []*x509.Certificate{&bad}
				want = "key_unsupported"
			case "nodes":
				limits.nodes = 1
				want = "node_limit"
			case "depth":
				limits.depth = 0
				want = "depth_limit"
			case "candidates":
				limits.candidates = 1
				cached.certs = append(cached.certs, relationIssue(t, relationTemplate("synthetic-intermediate", true), c.root, relationKey(t), c.rk))
				want = "candidate_limit"
			case "signatures":
				limits.signatures = 1
				want = "signature_limit"
			case "cancel":
				cancel()
				want = "canceled"
			case "deadline":
				var end context.CancelFunc
				ctx, end = context.WithDeadline(ctx, time.Now().Add(-time.Second))
				defer end()
				want = "deadline"
			}
			reason, r := observeRelationsBounded(ctx, peer, cached, roots, snapshot, map[[32]byte]bool{}, limits)
			if reason != "tls_chain_relation_unavailable" || r.Status != "unavailable" || r.Reason != want || r.Facts != nil || r.CacheCoverage != "partial" || !validRelation(reason, r) {
				t.Fatalf("limit %s: %+v", name, r)
			}
		})
	}
}

func TestRootSnapshotBudgetDoesNotChangeTrust(t *testing.T) {
	c := relationChain(t)
	for _, name := range []string{"count", "bytes", "single", "empty"} {
		t.Run(name, func(t *testing.T) {
			pool, s := relationRoots(c.root)
			before := pool.Clone()
			switch name {
			case "count":
				for s.count < maxCachedCerts {
					s.add(c.root)
				}
				s.add(c.root)
			case "bytes":
				s.bytes = maxCachedBytes
				s.add(c.root)
			case "single":
				s.noteDER(make([]byte, maxProbeDER+1))
			case "empty":
				s.noteDER(nil)
			}
			s.add(c.root)
			if s.failure == "" || s.certs != nil || !pool.Equal(before) {
				t.Fatal("snapshot limit affected root trust or retained partial roots")
			}
		})
	}
}

func TestRelationOnlyAfterSecondUnknownAuthority(t *testing.T) {
	c := relationChain(t)
	for _, name := range []string{"unknown", "success", "usage", "baseline", "loader-failed"} {
		t.Run(name, func(t *testing.T) {
			roots, snapshot := relationRoots(c.root)
			peer := []*x509.Certificate{c.leaf}
			cached := cacheSnapshot{certs: []*x509.Certificate{c.intermediate}, partial: true}
			failure := ""
			switch name {
			case "unknown":
				roots, snapshot = relationRoots()
			case "usage":
				tmp := relationTemplate("synthetic-leaf", false)
				tmp.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
				peer = []*x509.Certificate{relationIssue(t, tmp, c.intermediate, c.lk, c.ik)}
			case "baseline":
				peer = append(peer, c.intermediate)
			case "loader-failed":
				failure = "tls_chain_probe_store_read_failed"
			}
			observer := &relationObservation{roots: snapshot}
			reason := compareChain(context.Background(), failureFor(peer...), roots, map[[32]byte]bool{}, "pass.neu.edu.cn", time.Now(), func(context.Context) (cacheSnapshot, string) { return cached, failure }, observer)
			if name == "unknown" {
				if !validRelation(reason, observer.result) || !observer.result.Facts.UpstreamNoParent {
					t.Fatal("unknown authority did not reach same-source observation")
				}
			} else if observer.result != nil || relationReason(reason) {
				t.Fatal("relation observer invoked outside its failure gate")
			}
		})
	}
}

func TestRelationReportClosedAndLegacyPreserved(t *testing.T) {
	r := blocked("anonymous-once", "tls_chain_relation_observed")
	r.ContextConstructed = true
	r.Attempts = cascollect.Attempts{Requests: 3, DNS: 3, Connects: 3}
	r.Operations.ExternalRequests = 3
	r.Operations.DNS = 3
	r.ChainRelation = &ChainRelation{Status: "complete", CacheCoverage: "partial", Reason: "none", Facts: &RelationFacts{LeafNoParent: true}}
	b, err := Encode(r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Decode(b); err != nil {
		t.Fatal("new report rejected")
	}
	for _, mutate := range []func(*Report){
		func(r *Report) {
			r.Mode = "local-cache-check"
			r.Attempts = cascollect.Attempts{}
			r.Operations.ExternalRequests = 0
			r.Operations.DNS = 0
			r.ContextConstructed = false
		},
		func(r *Report) { r.Reason = "tls_chain_partial_cache_unknown_authority" },
		func(r *Report) { r.ChainRelation.Facts.DeniedObserved = true },
		func(r *Report) { r.ChainRelation.Facts.RootPathObserved = true },
		func(r *Report) { r.ChainRelation.Status = "unavailable" },
		func(r *Report) { r.ChainRelation.CacheCoverage = "other" },
		func(r *Report) { r.ChainRelation.Facts = nil },
		func(r *Report) { r.ChainRelation = nil },
	} {
		var bad Report
		if json.Unmarshal(b, &bad) != nil {
			t.Fatal("fixture decode")
		}
		mutate(&bad)
		if _, err := Encode(bad); err == nil {
			t.Fatal("contradictory D6 encoded")
		}
		data, _ := json.Marshal(bad)
		if _, err := Decode(data); err == nil {
			t.Fatal("contradictory D6 decoded")
		}
	}
	for _, data := range []string{
		strings.Replace(string(b), `"cycle_observed":false`, `"cycle_observed":false,"cycle_observed":false`, 1),
		strings.Replace(string(b), `"cycle_observed":false,`, ``, 1),
		strings.Replace(string(b), `"cycle_observed":false`, `"private_name":"canary","cycle_observed":false`, 1),
		strings.Replace(string(b), `"cycle_observed":false`, `"cycle_observed":null`, 1),
	} {
		if _, err := Decode([]byte(data)); err == nil {
			t.Fatal("open D6 schema")
		}
	}
	old := blocked("anonymous-once", "tls_chain_partial_cache_unknown_authority")
	old.Attempts = r.Attempts
	old.Operations = r.Operations
	data, err := Encode(old)
	if err != nil || strings.Contains(string(data), "chain_relation") {
		t.Fatal("legacy shape changed")
	}
	if _, err := Decode(data); err != nil {
		t.Fatal("legacy decode lost")
	}
	null := strings.Replace(string(data), `"analysis":null`, `"analysis":null,"chain_relation":null`, 1)
	if _, err := Decode([]byte(null)); err == nil {
		t.Fatal("explicit null extension accepted")
	}
}
