package casacquire

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"testing"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
)

func TestPartialCachePreservesValidOrderAndBudgets(t *testing.T) {
	c := generateChain(t, func(*x509.Certificate, *x509.Certificate) {})
	var cached cachedCertificates
	bad := []byte("synthetic-unparseable-DER")
	for _, der := range [][]byte{c.intermediate.Raw, bad, c.root.Raw} {
		if !cached.add(context.Background(), der) {
			t.Fatal("nonfatal entry stopped enumeration")
		}
	}
	if !cached.partial || cached.failure != "" || cached.count != 3 || cached.bytes != len(c.intermediate.Raw)+len(bad)+len(c.root.Raw) || len(cached.certs) != 2 || !bytes.Equal(cached.certs[0].Raw, c.intermediate.Raw) || !bytes.Equal(cached.certs[1].Raw, c.root.Raw) {
		t.Fatal("partial cache order or accounting lost")
	}
	for _, name := range []string{"count", "bytes", "zero-length", "oversized", "canceled"} {
		t.Run(name, func(t *testing.T) {
			var a cachedCertificates
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			der := bad
			want := "tls_chain_probe_cache_limit"
			switch name {
			case "count":
				for i := 0; i < maxCachedCerts; i++ {
					if !a.add(ctx, bad) {
						t.Fatal("count limit premature")
					}
				}
			case "bytes":
				for i := 0; i < maxCachedBytes/maxProbeDER; i++ {
					if !a.add(ctx, make([]byte, maxProbeDER)) {
						t.Fatal("byte limit premature")
					}
				}
			case "zero-length":
				a.add(ctx, bad)
				der, want = nil, "tls_chain_probe_cache_invalid"
			case "oversized":
				a.add(ctx, bad)
				der = make([]byte, maxProbeDER+1)
			case "canceled":
				a.add(ctx, bad)
				cancel()
				want = "tls_chain_probe_canceled"
			}
			if a.add(ctx, der) || a.failure != want || !a.partial || len(a.certs) != 0 {
				t.Fatal("fatal failure softened by partial cache")
			}
		})
	}
}

func TestPartialComparisonNeverChangesTrustOrAcceptance(t *testing.T) {
	c := generateChain(t, func(*x509.Certificate, *x509.Certificate) {})
	for _, name := range []string{"available", "unresolved", "empty", "denied", "read-failed", "combined-subject-limit"} {
		t.Run(name, func(t *testing.T) {
			cached := cacheSnapshot{certs: []*x509.Certificate{c.intermediate}, partial: true}
			roots := c.pool.Clone()
			denied := map[[32]byte]bool{}
			peer := []*x509.Certificate{c.leaf}
			failure := ""
			want := "tls_chain_partial_cache_available"
			switch name {
			case "unresolved":
				roots = x509.NewCertPool()
				cached.certs = append(cached.certs, c.root)
				want = "tls_chain_partial_cache_unknown_authority"
			case "empty":
				cached.certs = nil
				want = "tls_chain_partial_cache_empty"
			case "denied":
				denied[sha256.Sum256(c.intermediate.Raw)] = true
				want = "tls_chain_probe_denied"
			case "read-failed":
				failure, want = "tls_chain_probe_store_read_failed", "tls_chain_probe_store_read_failed"
			case "combined-subject-limit":
				foreign := generateChain(t, func(*x509.Certificate, *x509.Certificate) {}).intermediate
				peer = append(peer, foreign)
				for len(cached.certs) < maxSubjectCandidates {
					cached.certs = append(cached.certs, c.intermediate)
				}
				want = "tls_chain_probe_cache_subject_limit"
			}
			before := roots.Clone()
			got := compareChain(context.Background(), failureFor(peer...), roots, denied, "pass.neu.edu.cn", c.at, func(context.Context) (cacheSnapshot, string) { return cached, failure })
			if got != want || !roots.Equal(before) {
				t.Fatalf("reason=%s want=%s or roots mutated", got, want)
			}
		})
	}
}

func TestLocalCacheCheckClosedZeroOperations(t *testing.T) {
	c := generateChain(t, func(*x509.Certificate, *x509.Certificate) {})
	for _, name := range []string{"full", "empty", "partial", "partial-empty", "structural", "read-failed", "hostile", "subject-limit", "count-limit", "canceled-before", "canceled-loader", "deadline"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			cached := cacheSnapshot{certs: []*x509.Certificate{c.intermediate}}
			failure, want, collection := "", "none", "ready"
			switch name {
			case "empty":
				cached.certs = nil
			case "partial":
				cached.partial, want, collection = true, "cache_partial", "blocked"
			case "partial-empty":
				cached = cacheSnapshot{partial: true}
				want, collection = "cache_partial_empty", "blocked"
			case "structural":
				cached = cacheSnapshot{certs: []*x509.Certificate{nil}, partial: true}
				want, collection = "tls_chain_probe_cache_invalid", "blocked"
			case "read-failed":
				cached.partial = true
				failure, want, collection = "tls_chain_probe_store_read_failed", "tls_chain_probe_store_read_failed", "blocked"
			case "hostile":
				failure, want, collection = "private-error-canary", "tls_chain_probe_loader_failed", "blocked"
			case "subject-limit":
				for len(cached.certs) <= maxSubjectCandidates {
					cached.certs = append(cached.certs, c.intermediate)
				}
				want, collection = "tls_chain_probe_cache_subject_limit", "blocked"
			case "count-limit":
				cached.certs = make([]*x509.Certificate, maxCachedCerts+1)
				want, collection = "tls_chain_probe_cache_limit", "blocked"
			case "canceled-before":
				cancel()
				want, collection = "tls_chain_probe_canceled", "blocked"
			case "canceled-loader":
				want, collection = "tls_chain_probe_canceled", "blocked"
			case "deadline":
				var end context.CancelFunc
				ctx, end = context.WithDeadline(ctx, time.Now().Add(-time.Second))
				defer end()
				want, collection = "tls_chain_probe_deadline", "blocked"
			}
			calls := 0
			r := checkCache(ctx, func(local context.Context) (cacheSnapshot, string) {
				calls++
				deadline, ok := local.Deadline()
				if !ok || time.Until(deadline) > 10*time.Second {
					t.Fatal("local deadline absent")
				}
				if name == "canceled-loader" {
					cancel()
				}
				return cached, failure
			})
			if (name == "canceled-before" || name == "deadline") && calls != 0 {
				t.Fatal("canceled loader invoked")
			}
			if r.Reason != want || r.Collection != collection || r.ContextConstructed || r.Attempts != (cascollect.Attempts{}) || r.Operations != (casprovenance.Operations{}) || r.Analysis != nil || len(r.Resources)+len(r.References) != 0 {
				t.Fatalf("local report mismatch: %s", r.Reason)
			}
			data, err := Encode(r)
			if err != nil {
				t.Fatal("local report rejected")
			}
			if _, err := Decode(data); err != nil {
				t.Fatal("local roundtrip failed")
			}
			for _, canary := range [][]byte{[]byte("synthetic-chain"), []byte("private-error-canary"), []byte("CERTIFICATE")} {
				if bytes.Contains(data, canary) {
					t.Fatal("local certificate data leaked")
				}
			}
			for _, mutate := range []func(*Report){
				func(r *Report) { r.Attempts.Requests = 1; r.Operations.ExternalRequests = 1 },
				func(r *Report) { r.Operations.ProfileReads = 1 },
				func(r *Report) { r.ContextConstructed = true },
				func(r *Report) { r.Collection = "complete" },
				func(r *Report) { r.Collection = "blocked"; r.Reason = "tls_chain_partial_cache_available" },
				func(r *Report) { r.Collection = "blocked"; r.Reason = "topology_rejected" },
				func(r *Report) { r.Collection = "blocked"; r.Reason = "tls_chain_probe_peer_invalid" },
				func(r *Report) { r.Analysis = &Facts{} },
				func(r *Report) { r.Reason += "_private-error" },
			} {
				bad := r
				mutate(&bad)
				if _, err := Encode(bad); err == nil {
					t.Fatal("invalid local report encoded")
				}
				data, _ := json.Marshal(bad)
				if _, err := Decode(data); err == nil {
					t.Fatal("invalid local report decoded")
				}
			}
		})
	}
}
