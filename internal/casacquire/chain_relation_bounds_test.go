package casacquire

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"math/big"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestRelationPathDoesNotMeanValidChain(t *testing.T) {
	c := relationChain(t)
	template := relationTemplate("synthetic-intermediate", true)
	template.NotAfter = template.NotBefore.Add(time.Minute)
	expired := relationIssue(t, template, c.root, c.ik, c.rk)
	pool, snapshot := relationRoots(c.root)
	intermediates := x509.NewCertPool()
	intermediates.AddCert(expired)
	if _, err := c.leaf.Verify(x509.VerifyOptions{Roots: pool, Intermediates: intermediates, DNSName: "pass.neu.edu.cn", CurrentTime: time.Now()}); err == nil {
		t.Fatal("expired chain unexpectedly valid")
	}
	reason, r := observeRelations(context.Background(), []*x509.Certificate{c.leaf}, cacheSnapshot{certs: []*x509.Certificate{expired}}, pool, snapshot, map[[32]byte]bool{})
	if reason != "tls_chain_relation_observed" || r.Facts == nil || !r.Facts.RootPathObserved {
		t.Fatal("signature path incorrectly treated as full verification")
	}
}

func TestRelationCrossSignedDenialCannotBeCoveredByCleanPath(t *testing.T) {
	c := relationChain(t)
	key := relationKey(t)
	template := relationTemplate("other-root", true)
	other := relationIssue(t, template, template, key, key)
	cross := relationIssue(t, relationTemplate("synthetic-intermediate", true), other, c.ik, key)
	pool, snapshot := relationRoots(c.root, other)
	denied := map[[32]byte]bool{sha256.Sum256(c.intermediate.Raw): true}
	cached := cacheSnapshot{certs: []*x509.Certificate{c.intermediate, cross}, partial: true}
	var prior *ChainRelation
	for n := 0; n < 2; n++ {
		reason, r := observeRelations(context.Background(), []*x509.Certificate{c.leaf}, cached, pool, snapshot, denied)
		if reason != "tls_chain_relation_denied" || r.Facts == nil || !r.Facts.RootPathObserved || !r.Facts.DeniedObserved {
			t.Fatal("clean branch hid denial")
		}
		if prior != nil && !reflect.DeepEqual(prior, r) {
			t.Fatal("denial depends on order")
		}
		prior = r
		cached.certs[0], cached.certs[1] = cached.certs[1], cached.certs[0]
		snapshot.certs[0], snapshot.certs[1] = snapshot.certs[1], snapshot.certs[0]
	}
}

func TestRelationProductionDepthAndSignatureCaps(t *testing.T) {
	t.Run("depth", func(t *testing.T) {
		key := relationKey(t)
		template := relationTemplate("root", true)
		root := relationIssue(t, template, template, key, key)
		parent, signer := root, key
		var cached []*x509.Certificate
		for n := 0; n < 17; n++ {
			nextKey := relationKey(t)
			cert := relationIssue(t, relationTemplate("synthetic-layer-"+strconv.Itoa(n), true), parent, nextKey, signer)
			cached = append(cached, cert)
			parent, signer = cert, nextKey
		}
		leaf := relationIssue(t, relationTemplate("leaf", false), parent, relationKey(t), signer)
		pool, snapshot := relationRoots(root)
		reason, r := observeRelations(context.Background(), []*x509.Certificate{leaf}, cacheSnapshot{certs: cached}, pool, snapshot, map[[32]byte]bool{})
		if reason != "tls_chain_relation_unavailable" || r.Reason != "depth_limit" || r.Facts != nil {
			t.Fatalf("depth cap failed: %+v", r)
		}
	})
	t.Run("signatures", func(t *testing.T) {
		key := relationKey(t)
		template := relationTemplate("same-name", true)
		var cached []*x509.Certificate
		for n := 0; n < 16; n++ {
			template.SerialNumber = big.NewInt(int64(n + 1))
			cached = append(cached, relationIssue(t, template, template, key, key))
		}
		leaf := relationIssue(t, relationTemplate("leaf", false), cached[0], relationKey(t), key)
		pool, snapshot := relationRoots()
		reason, r := observeRelations(context.Background(), []*x509.Certificate{leaf}, cacheSnapshot{certs: cached}, pool, snapshot, map[[32]byte]bool{})
		if reason != "tls_chain_relation_unavailable" || r.Reason != "signature_limit" || r.Facts != nil {
			t.Fatalf("signature cap failed: %+v", r)
		}
	})
}

func TestRelationKeyWorkBoundsAndTypedFailures(t *testing.T) {
	for _, tc := range []struct {
		key any
		ok  bool
	}{
		{&rsa.PublicKey{N: new(big.Int).Lsh(big.NewInt(1), 8191), E: 65537}, true},
		{&rsa.PublicKey{N: new(big.Int).Lsh(big.NewInt(1), 8192), E: 65537}, false},
		{&rsa.PublicKey{N: big.NewInt(17), E: 0}, false},
		{&rsa.PublicKey{}, false},
		{&ecdsa.PublicKey{Curve: elliptic.P256(), X: big.NewInt(1), Y: big.NewInt(1)}, true},
		{&ecdsa.PublicKey{Curve: elliptic.P224(), X: big.NewInt(1), Y: big.NewInt(1)}, false},
		{ed25519.PublicKey(make([]byte, 32)), true},
		{ed25519.PublicKey(make([]byte, 31)), false},
		{nil, false},
	} {
		if boundedRelationKey(&x509.Certificate{PublicKey: tc.key}) != tc.ok {
			t.Fatal("key work bound mismatch")
		}
	}
	for _, tc := range []struct {
		err  error
		want RelationFacts
	}{
		{x509.ConstraintViolationError{}, RelationFacts{CAConstraintRejected: true}},
		{x509.InsecureAlgorithmError(x509.ECDSAWithSHA1), RelationFacts{AlgorithmRejected: true}},
		{x509.ErrUnsupportedAlgorithm, RelationFacts{AlgorithmRejected: true}},
		{opaqueError{}, RelationFacts{SignatureCheckFailed: true}},
	} {
		var facts RelationFacts
		recordRelationFailure(&facts, wrappedError{tc.err})
		if facts != tc.want {
			t.Fatal("signature failure text or type confusion")
		}
	}
}
