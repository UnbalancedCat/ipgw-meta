package casacquire

import (
	"context"
	"crypto/x509"
	"testing"
	"time"
)

var probeFailureCases = []string{
	"tls_chain_probe_canceled", "tls_chain_probe_deadline", "tls_chain_probe_input_rejected", "tls_chain_probe_peer_invalid", "tls_chain_probe_peer_limit", "tls_chain_probe_baseline_mismatch", "tls_chain_probe_peer_subject_limit", "tls_chain_probe_cache_subject_limit", "tls_chain_probe_store_open_failed", "tls_chain_probe_store_read_failed", "tls_chain_probe_cache_invalid", "tls_chain_probe_cache_limit", "tls_chain_probe_loader_failed",
}

func TestProbeLoaderCausesAreClosed(t *testing.T) {
	c := generateChain(t, func(*x509.Certificate, *x509.Certificate) {})
	for _, reason := range append(append([]string{}, probeFailureCases...), "synthetic-private-error", "tls_chain_local_intermediates_available") {
		t.Run(reason, func(t *testing.T) {
			got := compareChain(context.Background(), failureFor(c.leaf), c.pool, map[[32]byte]bool{}, "pass.neu.edu.cn", c.at, func(context.Context) (cacheSnapshot, string) {
				return cacheSnapshot{certs: []*x509.Certificate{c.intermediate}}, reason
			})
			want := "tls_chain_probe_loader_failed"
			switch reason {
			case "tls_chain_probe_canceled", "tls_chain_probe_deadline", "tls_chain_probe_store_open_failed", "tls_chain_probe_store_read_failed", "tls_chain_probe_cache_invalid", "tls_chain_probe_cache_limit", "tls_chain_probe_loader_failed":
				want = reason
			}
			if got != want {
				t.Fatalf("reason=%s want=%s", got, want)
			}
		})
	}
}

func TestProbeDeadlineAndInvalidCacheClassification(t *testing.T) {
	c := generateChain(t, func(*x509.Certificate, *x509.Certificate) {})
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	called := false
	got := compareChain(ctx, failureFor(c.leaf), c.pool, map[[32]byte]bool{}, "pass.neu.edu.cn", c.at, func(context.Context) (cacheSnapshot, string) { called = true; return cacheSnapshot{}, "" })
	if got != "tls_chain_probe_deadline" || called {
		t.Fatal("expired context advanced")
	}
	got = compareChain(context.Background(), failureFor(c.leaf), c.pool, map[[32]byte]bool{}, "pass.neu.edu.cn", c.at, func(context.Context) (cacheSnapshot, string) {
		return cacheSnapshot{certs: []*x509.Certificate{nil}}, ""
	})
	if got != "tls_chain_probe_cache_invalid" {
		t.Fatal("cache structural error lost")
	}
	var cached cachedCertificates
	if cached.add(ctx, c.intermediate.Raw) || cached.failure != "tls_chain_probe_deadline" {
		t.Fatal("cache deadline lost")
	}
	if !cached.add(context.Background(), []byte("synthetic-invalid")) || cached.failure != "" || !cached.partial || len(cached.certs) != 0 {
		t.Fatal("cache parse rejection not marked partial")
	}
	cached.count = maxCachedCerts
	if cached.add(context.Background(), c.intermediate.Raw) || cached.failure != "tls_chain_probe_cache_limit" {
		t.Fatal("cache limit lost")
	}
}
