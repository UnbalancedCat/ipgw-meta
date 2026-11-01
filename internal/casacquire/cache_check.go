package casacquire

import (
	"context"
	"time"
)

// loadCache is shared by the local check and failed-TLS comparison. On any
// fatal error it discards the snapshot; no caller may use a partial read.
func loadCache(ctx context.Context, load intermediateLoader) (cacheSnapshot, string) {
	if reason := probeContext(ctx); reason != "" {
		return cacheSnapshot{}, reason
	}
	if load == nil {
		return cacheSnapshot{}, "tls_chain_probe_loader_failed"
	}
	cached, reason := load(ctx)
	if canceled := probeContext(ctx); canceled != "" {
		return cacheSnapshot{}, canceled
	}
	if reason != "" {
		if !loaderFailureReason(reason) {
			reason = "tls_chain_probe_loader_failed"
		}
		return cacheSnapshot{}, reason
	}
	if reason := certificateBounds(cached.certs, maxCachedCerts, maxCachedBytes, "cache"); reason != "" {
		return cacheSnapshot{}, reason
	}
	subjects := map[string]int{}
	for _, cert := range cached.certs {
		if reason := probeContext(ctx); reason != "" {
			return cacheSnapshot{}, reason
		}
		subjects[string(cert.RawSubject)]++
		if subjects[string(cert.RawSubject)] > maxSubjectCandidates {
			return cacheSnapshot{}, "tls_chain_probe_cache_subject_limit"
		}
	}
	if reason := probeContext(ctx); reason != "" {
		return cacheSnapshot{}, reason
	}
	return cached, ""
}

// CacheCheck reads only the existing local CA registry stores. It has no
// topology, roots, network, peer certificates, profile or persistence path.
func CacheCheck(ctx context.Context) Report {
	return checkCache(ctx, localIntermediates)
}

func checkCache(ctx context.Context, load intermediateLoader) Report {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cached, reason := loadCache(ctx, load)
	r := blocked("local-cache-check", reason)
	if reason != "" {
		return r
	}
	if cached.partial {
		r.Reason = "cache_partial"
		if len(cached.certs) == 0 {
			r.Reason = "cache_partial_empty"
		}
		return r
	}
	r.Collection, r.Reason = "ready", "none"
	return r
}
