package casacquire

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
)

func TestChainFailureTypesNeverInspectText(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"authority", x509.UnknownAuthorityError{}, "unknown_authority"},
		{"hostname", x509.HostnameError{}, "hostname_mismatch"},
		{"validity", x509.CertificateInvalidError{Reason: x509.Expired, Detail: "private-canary"}, "time_invalid"},
		{"usage", x509.CertificateInvalidError{Reason: x509.IncompatibleUsage}, "usage_invalid"},
		{"ca-usage", x509.CertificateInvalidError{Reason: x509.CANotAuthorizedForExtKeyUsage}, "usage_invalid"},
		{"name-constraint", x509.CertificateInvalidError{Reason: x509.CANotAuthorizedForThisName}, "constraints_invalid"},
		{"path-length", x509.CertificateInvalidError{Reason: x509.TooManyIntermediates}, "constraints_invalid"},
		{"legacy-san", x509.CertificateInvalidError{Reason: x509.NameConstraintsWithoutSANs}, "constraints_invalid"},
		{"name-type", x509.CertificateInvalidError{Reason: x509.UnconstrainedName}, "constraints_invalid"},
		{"constraint-budget", x509.CertificateInvalidError{Reason: x509.TooManyConstraints}, "constraints_invalid"},
		{"non-ca", x509.CertificateInvalidError{Reason: x509.NotAuthorizedToSign}, "certificate_invalid"},
		{"issuer-name", x509.CertificateInvalidError{Reason: x509.NameMismatch}, "certificate_invalid"},
		{"policy", x509.CertificateInvalidError{Reason: x509.NoValidChains, Detail: "private-canary"}, "certificate_invalid"},
		{"unknown-enum", x509.CertificateInvalidError{Reason: x509.InvalidReason(-1)}, "certificate_invalid"},
		{"critical", x509.UnhandledCriticalExtension{}, "certificate_invalid"},
		{"algorithm", x509.InsecureAlgorithmError(x509.MD5WithRSA), "certificate_invalid"},
		{"constraint-error", x509.ConstraintViolationError{}, "certificate_invalid"},
		{"opaque", opaqueError{}, "verify_failed"},
		{"nil-defensive", nil, "verify_failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, partial := range []bool{false, true} {
				prefix := "tls_chain_cache_"
				if partial {
					prefix = "tls_chain_partial_cache_"
				}
				if got := chainFailure(wrappedError{tc.err}, partial); got != prefix+tc.want {
					t.Fatalf("got %s want %s", got, prefix+tc.want)
				}
			}
		})
	}
}

func TestSecondVerifyFailureFromGeneratedChains(t *testing.T) {
	for _, tc := range []struct{ name, suffix string }{
		{"missing-root", "unknown_authority"},
		{"expired-intermediate", "time_invalid"},
		{"future-intermediate", "time_invalid"},
		{"leaf-usage", "usage_invalid"},
		{"ca-name-constraint", "constraints_invalid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := generateChain(t, func(i, l *x509.Certificate) {
				switch tc.name {
				case "expired-intermediate":
					i.NotAfter = i.NotBefore.Add(time.Minute)
				case "future-intermediate":
					i.NotBefore = time.Now().Add(time.Minute)
				case "leaf-usage":
					l.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
				case "ca-name-constraint":
					i.PermittedDNSDomains = []string{"synthetic.invalid"}
					i.PermittedDNSDomainsCritical = true
				}
			})
			roots := c.pool
			if tc.name == "missing-root" {
				roots = x509.NewCertPool()
			}
			before := roots.Clone()
			for _, partial := range []bool{false, true} {
				loads := 0
				got := compareChain(context.Background(), failureFor(c.leaf), roots, map[[32]byte]bool{}, "pass.neu.edu.cn", c.at, func(context.Context) (cacheSnapshot, string) {
					loads++
					return cacheSnapshot{certs: []*x509.Certificate{c.intermediate}, partial: partial}, ""
				})
				prefix := "tls_chain_cache_"
				if partial {
					prefix = "tls_chain_partial_cache_"
				}
				if got != prefix+tc.suffix || loads != 1 || !roots.Equal(before) {
					t.Fatalf("second Verify classification: %s", got)
				}
			}
		})
	}
}

func TestChainFailureReportClosed(t *testing.T) {
	for _, prefix := range []string{"tls_chain_cache_", "tls_chain_partial_cache_"} {
		for _, suffix := range []string{"unknown_authority", "hostname_mismatch", "time_invalid", "usage_invalid", "constraints_invalid", "certificate_invalid", "verify_failed"} {
			r := blocked("anonymous-once", prefix+suffix)
			r.ContextConstructed = true
			r.Attempts = cascollect.Attempts{Requests: 3, DNS: 3, Connects: 3}
			r.Operations.ExternalRequests, r.Operations.DNS = 3, 3
			b, err := Encode(r)
			if err != nil {
				t.Fatal("new fixed category rejected")
			}
			if _, err := Decode(b); err != nil {
				t.Fatal("roundtrip failed")
			}
			if strings.Contains(string(b), "private-canary") {
				t.Fatal("error text leaked")
			}
			for _, mutate := range []func(*Report){
				func(r *Report) { r.Reason += "_private-canary" },
				func(r *Report) { r.Collection = "complete" },
				func(r *Report) { r.Attempts.Connects-- },
				func(r *Report) {
					r.Attempts.Requests = 13
					r.Attempts.DNS = 13
					r.Attempts.Connects = 13
					r.Operations.ExternalRequests = 13
					r.Operations.DNS = 13
				},
				func(r *Report) { r.Analysis = &Facts{} },
				func(r *Report) {
					r.Mode = "local-cache-check"
					r.Attempts = cascollect.Attempts{}
					r.Operations.ExternalRequests = 0
					r.Operations.DNS = 0
					r.ContextConstructed = false
				},
				func(r *Report) { r.Mode = "local-preflight" },
				func(r *Report) { r.Mode = "self-test" },
			} {
				bad := r
				mutate(&bad)
				if _, err := Encode(bad); err == nil {
					t.Fatal("contradictory report encoded")
				}
				b, _ := json.Marshal(bad)
				if _, err := Decode(b); err == nil {
					t.Fatal("contradictory report decoded")
				}
			}
		}
	}
}
