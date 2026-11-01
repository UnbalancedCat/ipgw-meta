package casacquire

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
)

// Entirely generated, in-memory sources. Never calls a Windows provider.
func sourceSelfTest(ctx context.Context) bool {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return false
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "synthetic-source"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), BasicConstraintsValid: true, IsCA: true, KeyUsage: x509.KeyUsageCertSign}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return false
	}
	c, err := x509.ParseCertificate(der)
	if err != nil {
		return false
	}
	s, gate := readSources(ctx, func(_ context.Context, i int, add func([]byte) string) (string, string) {
		if i == 2 {
			return "complete", add([]byte{1})
		}
		if i == 4 || i == 6 {
			return "complete", add(der)
		}
		return "absent", ""
	})
	if gate != nil {
		return false
	}
	var graph *differenceGraph
	var topology *ChainTopology
	reason, relation := observeRelationsCaptured(ctx, []*x509.Certificate{c}, cacheSnapshot{}, x509.NewCertPool(), &rootSnapshot{}, map[[32]byte]bool{}, defaultRelationLimits, &graph, &topology)
	if graph == nil {
		return false
	}
	d := compareSources(ctx, s, graph, defaultDifferenceLimits)
	if d.Status != "complete" || !d.Sources[6].Facts.ReachableDEROverlap || !d.Sources[6].Facts.TerminalSameSubjectSPKI || !d.Sources[4].Facts.ReachableDEROverlap || d.Sources[2].ParseCoverage != "none" {
		return false
	}
	r := blocked("anonymous-once", reason)
	r.Attempts = cascollect.Attempts{Requests: 1, DNS: 1, Connects: 1}
	r.Operations.ExternalRequests, r.Operations.DNS = 1, 1
	r.ChainRelation, r.ChainTopology, r.SourceDifference = relation, topology, d
	b, err := Encode(r)
	if err != nil {
		return false
	}
	if _, err = Decode(b); err != nil {
		return false
	}
	// Both local gate failures must return without ever touching a wire.
	for _, failure := range []string{"", "read_failed"} {
		r = acquireWithSources(ctx, nil, func(context.Context, int, func([]byte) string) (string, string) { return "absent", failure })
		if r.Attempts != (cascollect.Attempts{}) || r.Collection != "blocked" || r.SourceDifference == nil {
			return false
		}
		if _, err = Encode(r); err != nil {
			return false
		}
	}
	return true
}
