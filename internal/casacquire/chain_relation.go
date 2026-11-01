package casacquire

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"errors"
	"sort"
	"time"
)

// These are relationship observations, not certificate or TLS acceptance.
type RelationFacts struct {
	RootPathObserved     bool `json:"root_path_observed"`
	LeafNoParent         bool `json:"leaf_no_parent"`
	UpstreamNoParent     bool `json:"upstream_no_parent"`
	CAConstraintRejected bool `json:"ca_constraint_rejected"`
	AlgorithmRejected    bool `json:"algorithm_rejected"`
	SignatureCheckFailed bool `json:"signature_check_failed"`
	CycleObserved        bool `json:"cycle_observed"`
	DeniedObserved       bool `json:"denied_observed"`
}

type ChainRelation struct {
	Status        string         `json:"status"`
	CacheCoverage string         `json:"cache_coverage"`
	Reason        string         `json:"reason"`
	Facts         *RelationFacts `json:"facts"`
}

// Capture only the very same objects added to the original pool. Failure here
// invalidates this optional diagnostic snapshot, never the original root pool.
type rootSnapshot struct {
	certs        []*x509.Certificate
	count, bytes int
	failure      string
}

func (s *rootSnapshot) add(cert *x509.Certificate) {
	if cert == nil {
		s.noteDER(nil)
		return
	}
	s.noteDER(cert.Raw)
	s.keep(cert)
}

func (s *rootSnapshot) noteDER(der []byte) {
	if s == nil || s.failure != "" {
		return
	}
	if len(der) == 0 {
		s.failure = "root_snapshot_invalid"
	} else if s.count >= maxCachedCerts || len(der) > maxProbeDER || len(der) > maxCachedBytes-s.bytes {
		s.failure = "root_snapshot_limit"
	}
	if s.failure != "" {
		s.certs = nil
		return
	}
	s.count++
	s.bytes += len(der)
}

func (s *rootSnapshot) keep(cert *x509.Certificate) {
	if s == nil || s.failure != "" {
		return
	}
	s.certs = append(s.certs, cert)
}

type relationObservation struct {
	roots    *rootSnapshot
	result   *ChainRelation
	topology *ChainTopology
	graph    *differenceGraph
}

func relationReason(reason string) bool {
	return reason == "tls_chain_relation_observed" || reason == "tls_chain_relation_denied" || reason == "tls_chain_relation_unavailable"
}

func relationFailure(reason string) bool {
	switch reason {
	case "root_snapshot_invalid", "root_snapshot_limit", "input_invalid", "key_unsupported", "node_limit", "depth_limit", "candidate_limit", "signature_limit", "canceled", "deadline":
		return true
	}
	return false
}

func validRelation(reason string, r *ChainRelation) bool {
	if r == nil || (r.CacheCoverage != "complete" && r.CacheCoverage != "partial") {
		return false
	}
	if r.Status == "unavailable" {
		return reason == "tls_chain_relation_unavailable" && r.Facts == nil && relationFailure(r.Reason)
	}
	if r.Status != "complete" || r.Reason != "none" || r.Facts == nil {
		return false
	}
	f := r.Facts
	if f.LeafNoParent && (f.RootPathObserved || f.UpstreamNoParent || f.CAConstraintRejected || f.AlgorithmRejected || f.SignatureCheckFailed || f.CycleObserved) {
		return false
	}
	if !f.RootPathObserved && !f.LeafNoParent && !f.UpstreamNoParent && !f.CAConstraintRejected && !f.AlgorithmRejected && !f.SignatureCheckFailed && !f.CycleObserved {
		return false
	}
	return f.DeniedObserved && reason == "tls_chain_relation_denied" || !f.DeniedObserved && reason == "tls_chain_relation_observed"
}

func boundedRelationKey(cert *x509.Certificate) bool {
	switch key := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		return key != nil && key.N != nil && key.N.Sign() > 0 && key.N.BitLen() <= 8192 && key.E >= 2 && key.E <= 1<<31-1
	case *ecdsa.PublicKey:
		return key != nil && key.X != nil && key.Y != nil && (key.Curve == elliptic.P256() || key.Curve == elliptic.P384() || key.Curve == elliptic.P521())
	case ed25519.PublicKey:
		return len(key) == ed25519.PublicKeySize
	}
	return false
}

func recordRelationFailure(f *RelationFacts, err error) {
	var constraint x509.ConstraintViolationError
	var algorithm x509.InsecureAlgorithmError
	switch {
	case errors.As(err, &constraint):
		f.CAConstraintRejected = true
	case errors.As(err, &algorithm), errors.Is(err, x509.ErrUnsupportedAlgorithm):
		f.AlgorithmRejected = true
	default:
		f.SignatureCheckFailed = true
	}
}

type relationLimits struct{ nodes, depth, candidates, signatures int }

var defaultRelationLimits = relationLimits{128, 16, 16, 100}

func observeRelations(ctx context.Context, peer []*x509.Certificate, cached cacheSnapshot, roots *x509.CertPool, snapshot *rootSnapshot, denied map[[32]byte]bool, topology ...**ChainTopology) (string, *ChainRelation) {
	return observeRelationsBounded(ctx, peer, cached, roots, snapshot, denied, defaultRelationLimits, topology...)
}

func observeRelationsBounded(ctx context.Context, peer []*x509.Certificate, cached cacheSnapshot, roots *x509.CertPool, snapshot *rootSnapshot, denied map[[32]byte]bool, limits relationLimits, topology ...**ChainTopology) (string, *ChainRelation) {
	return observeRelationsCaptured(ctx, peer, cached, roots, snapshot, denied, limits, nil, topology...)
}

func observeRelationsCaptured(ctx context.Context, peer []*x509.Certificate, cached cacheSnapshot, roots *x509.CertPool, snapshot *rootSnapshot, denied map[[32]byte]bool, limits relationLimits, capture **differenceGraph, topology ...**ChainTopology) (string, *ChainRelation) {
	if capture != nil {
		*capture = nil
	}
	for _, output := range topology {
		if output != nil {
			*output = nil
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	r := &ChainRelation{Status: "unavailable", CacheCoverage: "complete"}
	if cached.partial {
		r.CacheCoverage = "partial"
	}
	fail := func(reason string) (string, *ChainRelation) {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			reason = "deadline"
		} else if ctx.Err() != nil {
			reason = "canceled"
		}
		r.Reason, r.Facts = reason, nil
		return "tls_chain_relation_unavailable", r
	}
	if ctx.Err() != nil {
		return fail("canceled")
	}
	if snapshot == nil || roots == nil {
		return fail("root_snapshot_invalid")
	}
	if snapshot.failure != "" {
		if snapshot.failure == "root_snapshot_limit" {
			return fail(snapshot.failure)
		}
		return fail("root_snapshot_invalid")
	}
	if certificateBounds(snapshot.certs, maxCachedCerts, maxCachedBytes, "cache") != "" {
		return fail("root_snapshot_invalid")
	}
	if len(peer) == 0 || denied == nil || certificateBounds(peer, maxPeerCerts, maxPeerBytes, "peer") != "" || certificateBounds(cached.certs, maxCachedCerts, maxCachedBytes, "cache") != "" {
		return fail("input_invalid")
	}
	if limits.nodes < 1 || limits.nodes > 128 || limits.depth < 0 || limits.depth > 16 || limits.candidates < 1 || limits.candidates > 16 || limits.signatures < 1 || limits.signatures > 100 {
		return fail("input_invalid")
	}
	rootSet := map[[32]byte]bool{}
	reconstructed := x509.NewCertPool()
	for _, cert := range snapshot.certs {
		if ctx.Err() != nil {
			return fail("canceled")
		}
		reconstructed.AddCert(cert)
		rootSet[sha256.Sum256(cert.Raw)] = true
	}
	if !roots.Equal(reconstructed) {
		return fail("root_snapshot_invalid")
	}
	// Stable sorting and DER de-duplication avoid input order deciding which
	// branch is observed first. No digest, subject, or node identifier is output.
	certs := map[[32]byte]*x509.Certificate{}
	sources := map[[32]byte]uint8{}
	for groupIndex, group := range [][]*x509.Certificate{peer, cached.certs, snapshot.certs} {
		for _, cert := range group {
			if ctx.Err() != nil {
				return fail("canceled")
			}
			id := sha256.Sum256(cert.Raw)
			if prior := certs[id]; prior != nil && !bytes.Equal(prior.Raw, cert.Raw) {
				return fail("input_invalid")
			}
			certs[id] = cert
			sources[id] |= 1 << groupIndex
		}
	}
	bySubject := map[string][][32]byte{}
	for id, cert := range certs {
		bySubject[string(cert.RawSubject)] = append(bySubject[string(cert.RawSubject)], id)
	}
	for subject, ids := range bySubject {
		sort.Slice(ids, func(i, j int) bool { return bytes.Compare(ids[i][:], ids[j][:]) < 0 })
		bySubject[subject] = ids
	}
	leaf := sha256.Sum256(peer[0].Raw)
	queue := [][32]byte{leaf}
	distance := map[[32]byte]int{leaf: 0}
	edges := map[[32]byte][][32]byte{}
	facts := &RelationFacts{}
	checks := 0
	for head := 0; head < len(queue); head++ {
		if ctx.Err() != nil {
			return fail("canceled")
		}
		id := queue[head]
		cert := certs[id]
		if denied[id] {
			facts.DeniedObserved = true
		}
		if rootSet[id] {
			facts.RootPathObserved = true
			continue
		}
		parents := bySubject[string(cert.RawIssuer)]
		if len(parents) > limits.candidates {
			return fail("candidate_limit")
		}
		if len(parents) == 0 {
			if id == leaf {
				facts.LeafNoParent = true
			} else {
				facts.UpstreamNoParent = true
			}
		}
		for _, parentID := range parents {
			if ctx.Err() != nil {
				return fail("canceled")
			}
			parent := certs[parentID]
			if !boundedRelationKey(parent) {
				return fail("key_unsupported")
			}
			if checks >= limits.signatures {
				return fail("signature_limit")
			}
			checks++
			err := cert.CheckSignatureFrom(parent)
			if ctx.Err() != nil {
				return fail("canceled")
			}
			if err != nil {
				recordRelationFailure(facts, err)
				continue
			}
			edges[id] = append(edges[id], parentID)
			if _, seen := distance[parentID]; !seen {
				if len(queue) >= limits.nodes {
					return fail("node_limit")
				}
				depth := distance[id] + 1
				if depth > limits.depth {
					return fail("depth_limit")
				}
				distance[parentID] = depth
				queue = append(queue, parentID)
			}
		}
	}
	// Kahn's algorithm detects any cycle in the complete reachable graph,
	// without recursive/exponential path enumeration or infinite cycle depth.
	indegree := map[[32]byte]int{}
	for _, id := range queue {
		for _, parent := range edges[id] {
			indegree[parent]++
		}
	}
	acyclic := [][32]byte{}
	for _, id := range queue {
		if indegree[id] == 0 {
			acyclic = append(acyclic, id)
		}
	}
	for i := 0; i < len(acyclic); i++ {
		if ctx.Err() != nil {
			return fail("canceled")
		}
		for _, parent := range edges[acyclic[i]] {
			indegree[parent]--
			if indegree[parent] == 0 {
				acyclic = append(acyclic, parent)
			}
		}
	}
	facts.CycleObserved = len(acyclic) != len(queue)
	if ctx.Err() != nil {
		return fail("canceled")
	}
	r.Status, r.Reason, r.Facts = "complete", "none", facts
	if facts.CycleObserved && !facts.RootPathObserved && len(topology) == 1 && topology[0] != nil {
		graph := relationGraph{queue, certs, edges, sources, snapshot.certs}
		terminals := map[[32]byte]bool{}
		observation := observeTopologyCaptured(ctx, graph, defaultTopologyLimits, terminals)
		*topology[0] = observation
		if observation.Status == "unavailable" {
			return "tls_chain_topology_unavailable", r
		}
		if capture != nil {
			*capture = &differenceGraph{graph, terminals}
		}
		if facts.DeniedObserved {
			return "tls_chain_topology_denied", r
		}
		return "tls_chain_topology_observed", r
	}
	if facts.DeniedObserved {
		return "tls_chain_relation_denied", r
	}
	return "tls_chain_relation_observed", r
}
