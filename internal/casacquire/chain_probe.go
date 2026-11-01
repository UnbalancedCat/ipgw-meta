package casacquire

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"time"
)

const (
	maxProbeDER          = 64 << 10
	maxPeerCerts         = 16
	maxPeerBytes         = 512 << 10
	maxCachedCerts       = 512
	maxCachedBytes       = 8 << 20
	maxSubjectCandidates = 16
	probeUnavailable     = "tls_chain_probe_unavailable"
)

// Sources are local-only. The production loader reads public DER, never keys.
type cacheSnapshot struct {
	certs   []*x509.Certificate
	partial bool
}

type intermediateLoader func(context.Context) (cacheSnapshot, string)

type cachedCertificates struct {
	count, bytes int
	certs        []*x509.Certificate
	failure      string
	partial      bool
}

func (c *cachedCertificates) add(ctx context.Context, der []byte) bool {
	c.failure = probeContext(ctx)
	if c.failure != "" {
		return false
	}
	if len(der) == 0 {
		c.failure = "tls_chain_probe_cache_invalid"
		return false
	}
	if len(der) > maxProbeDER || c.count >= maxCachedCerts || len(der) > maxCachedBytes-c.bytes {
		c.failure = "tls_chain_probe_cache_limit"
		return false
	}
	c.count++
	c.bytes += len(der)
	// Enumeration owns the source buffer; retain only an independent copy.
	cert, err := x509.ParseCertificate(append([]byte(nil), der...))
	if err != nil {
		// The rejected item still consumed the original enumeration budgets.
		c.partial = true
		return true
	}
	c.certs = append(c.certs, cert)
	return true
}

func boundedCertificates(certs []*x509.Certificate, maxCount, maxBytes int) bool {
	return certificateBounds(certs, maxCount, maxBytes, "peer") == ""
}

func certificateBounds(certs []*x509.Certificate, maxCount, maxBytes int, source string) string {
	if len(certs) > maxCount {
		return "tls_chain_probe_" + source + "_limit"
	}
	for _, cert := range certs {
		if cert == nil || len(cert.Raw) == 0 {
			return "tls_chain_probe_" + source + "_invalid"
		}
		if len(cert.Raw) > maxProbeDER || len(cert.Raw) > maxBytes {
			return "tls_chain_probe_" + source + "_limit"
		}
		maxBytes -= len(cert.Raw)
	}
	return ""
}

func probeContext(ctx context.Context) string {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "tls_chain_probe_deadline"
	}
	if ctx.Err() != nil {
		return "tls_chain_probe_canceled"
	}
	return ""
}

func probeFailureReason(reason string) bool {
	switch reason {
	case "tls_chain_probe_canceled", "tls_chain_probe_deadline", "tls_chain_probe_input_rejected", "tls_chain_probe_peer_invalid", "tls_chain_probe_peer_limit", "tls_chain_probe_baseline_mismatch", "tls_chain_probe_peer_subject_limit", "tls_chain_probe_cache_subject_limit", "tls_chain_probe_store_open_failed", "tls_chain_probe_store_read_failed", "tls_chain_probe_cache_invalid", "tls_chain_probe_cache_limit", "tls_chain_probe_loader_failed":
		return true
	}
	return false
}

// Accept only loader-origin causes, not arbitrary text or unrelated statuses.
func loaderFailureReason(reason string) bool {
	switch reason {
	case "tls_chain_probe_canceled", "tls_chain_probe_deadline", "tls_chain_probe_store_open_failed", "tls_chain_probe_store_read_failed", "tls_chain_probe_cache_invalid", "tls_chain_probe_cache_limit", "tls_chain_probe_loader_failed":
		return true
	}
	return false
}

func probeReason(reason string) bool {
	if probeFailureReason(reason) || chainFailureReason(reason) {
		return true
	}
	switch reason {
	case "tls_chain_local_intermediates_available", "tls_chain_unresolved", "tls_chain_probe_denied", "tls_chain_partial_cache_available", "tls_chain_partial_cache_unresolved", "tls_chain_partial_cache_empty", probeUnavailable:
		return true
	}
	return false
}

// Called only after the failed socket is closed. A result is a diagnosis, never
// permission to resume TLS or HTTP. Roots remain the original offline pool.
func compareChain(ctx context.Context, failure error, roots *x509.CertPool, denied map[[32]byte]bool, hostname string, at time.Time, load intermediateLoader, observation ...*relationObservation) string {
	for _, observer := range observation {
		if observer != nil {
			observer.result = nil
			observer.topology = nil
		}
	}
	var verification *tls.CertificateVerificationError
	var authority x509.UnknownAuthorityError
	if reason := probeContext(ctx); reason != "" {
		return reason
	}
	if roots == nil || denied == nil || hostname == "" || at.IsZero() || load == nil || !errors.As(failure, &verification) || verification == nil || !errors.As(verification.Err, &authority) {
		return "tls_chain_probe_input_rejected"
	}
	peer := verification.UnverifiedCertificates
	if len(peer) == 0 {
		return "tls_chain_probe_peer_invalid"
	}
	if reason := certificateBounds(peer, maxPeerCerts, maxPeerBytes, "peer"); reason != "" {
		return reason
	}
	intermediates := x509.NewCertPool()
	subjects := map[string]int{}
	add := func(cert *x509.Certificate) bool {
		subjects[string(cert.RawSubject)]++
		if subjects[string(cert.RawSubject)] > maxSubjectCandidates {
			return false
		}
		intermediates.AddCert(cert)
		return true
	}
	for _, cert := range peer[1:] {
		if !add(cert) {
			return "tls_chain_probe_peer_subject_limit"
		}
	}
	opts := x509.VerifyOptions{Roots: roots, Intermediates: intermediates, DNSName: hostname, CurrentTime: at, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	if reason := probeContext(ctx); reason != "" {
		return reason
	}
	_, err := peer[0].Verify(opts)
	if reason := probeContext(ctx); reason != "" {
		return reason
	}
	if !errors.As(err, &authority) {
		return "tls_chain_probe_baseline_mismatch"
	}
	// Do not read the local CA stores unless the original failure reproduces.
	cached, reason := loadCache(ctx, load)
	if reason != "" {
		return reason
	}
	for _, cert := range cached.certs {
		if !add(cert) {
			return "tls_chain_probe_cache_subject_limit"
		}
	}
	if reason := probeContext(ctx); reason != "" {
		return reason
	}
	if cached.partial && len(cached.certs) == 0 {
		return "tls_chain_partial_cache_empty"
	}
	chains, err := peer[0].Verify(opts)
	if reason := probeContext(ctx); reason != "" {
		return reason
	}
	if err != nil {
		if errors.As(err, &authority) && len(observation) == 1 && observation[0] != nil {
			reason, result := observeRelationsCaptured(ctx, peer, cached, roots, observation[0].roots, denied, defaultRelationLimits, &observation[0].graph, &observation[0].topology)
			if stopped := probeContext(ctx); stopped != "" {
				observation[0].topology = nil
				return stopped
			}
			observation[0].result = result
			return reason
		}
		return chainFailure(err, cached.partial)
	}
	for _, chain := range chains {
		for _, cert := range chain {
			if denied[sha256.Sum256(cert.Raw)] {
				return "tls_chain_probe_denied"
			}
		}
	}
	if cached.partial {
		return "tls_chain_partial_cache_available"
	}
	return "tls_chain_local_intermediates_available"
}
