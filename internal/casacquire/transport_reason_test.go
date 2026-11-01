package casacquire

import (
	"bufio"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
)

// Error text must never be inspected, even for unknown failures.
type opaqueError struct{ timeout bool }

func (opaqueError) Error() string   { panic("raw error text inspected") }
func (e opaqueError) Timeout() bool { return e.timeout }
func (opaqueError) Temporary() bool { return false }

type wrappedError struct{ err error }

func (wrappedError) Error() string   { panic("wrapped error text inspected") }
func (e wrappedError) Unwrap() error { return e.err }

func TestTransportFailureUsesTypesOnly(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"authority", x509.UnknownAuthorityError{}, "tls_unknown_authority"},
		{"hostname", x509.HostnameError{}, "tls_hostname_mismatch"},
		{"time", x509.CertificateInvalidError{Reason: x509.Expired}, "tls_certificate_time_invalid"},
		{"usage", x509.CertificateInvalidError{Reason: x509.IncompatibleUsage}, "tls_certificate_invalid"},
		{"other-verification", &tls.CertificateVerificationError{Err: opaqueError{}}, "tls_certificate_invalid"},
		{"denied", errCertificateDenied, "tls_certificate_denied"},
		{"timeout", opaqueError{timeout: true}, "tls_handshake_timeout"},
		{"deadline", context.DeadlineExceeded, "tls_handshake_timeout"},
		{"cancel", context.Canceled, "tls_handshake_canceled"},
		{"opaque", opaqueError{}, "tls_handshake_failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tlsFailure(context.Background(), wrappedError{tc.err}); got != tc.want {
				t.Fatalf("reason=%s want=%s", got, tc.want)
			}
		})
	}
	for _, stage := range []string{"request_write", "response_status", "response_headers", "response_body"} {
		for _, suffix := range []string{"failed", "timeout", "canceled"} {
			ctx, cancel := context.WithCancel(context.Background())
			if suffix == "canceled" {
				cancel()
			}
			got := ioFailure(stage, ctx, wrappedError{opaqueError{timeout: suffix == "timeout"}})
			cancel()
			if got != stage+"_"+suffix {
				t.Fatal("I/O classification mismatch")
			}
		}
	}
}

func TestCertificateClassesStopBeforeGET(t *testing.T) {
	for _, name := range []string{"trusted", "authority", "hostname", "expired", "not-yet-valid", "usage", "denied"} {
		t.Run(name, func(t *testing.T) {
			pair, roots := testCertificateWith(t, func(c *x509.Certificate) {
				switch name {
				case "hostname":
					c.DNSNames = []string{"synthetic.invalid"}
				case "expired":
					c.NotBefore = time.Now().Add(-2 * time.Hour)
					c.NotAfter = time.Now().Add(-time.Hour)
				case "not-yet-valid":
					c.NotBefore = time.Now().Add(time.Hour)
					c.NotAfter = time.Now().Add(2 * time.Hour)
				case "usage":
					c.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
				}
			})
			calls := &atomic.Int32{}
			target, _ := rawFixture(t, false, func(raw net.Conn) {
				conn := tls.Server(raw, &tls.Config{Certificates: []tls.Certificate{pair}, MinVersion: tls.VersionTLS12})
				defer conn.Close()
				req, err := http.ReadRequest(bufio.NewReader(conn))
				if err != nil {
					return
				}
				calls.Add(1)
				req.Body.Close()
				io.WriteString(conn, "HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")
			})
			dns, _ := dnsFixture(t, nil)
			w := &wire{local: netip.MustParseAddr("127.0.0.1"), dns: dns, testTarget: target, roots: roots}
			want := ""
			switch name {
			case "authority":
				w.roots = x509.NewCertPool()
				want = "tls_unknown_authority"
			case "hostname":
				want = "tls_hostname_mismatch"
			case "expired", "not-yet-valid":
				want = "tls_certificate_time_invalid"
			case "usage":
				want = "tls_certificate_invalid"
			case "denied":
				w.denied = map[[32]byte]bool{sha256.Sum256(pair.Certificate[0]): true}
				want = "tls_certificate_denied"
			}
			_, got := w.fetch(context.Background(), cascollect.Request{Method: "GET", Target: "https://pass.neu.edu.cn/tpass/login"}, 16)
			wantCalls := int32(0)
			if name == "trusted" {
				wantCalls = 1
			}
			if got != want || calls.Load() != wantCalls || w.counts != (cascollect.Attempts{Requests: 1, DNS: 1, Connects: 1}) {
				t.Fatalf("reason=%s want=%s; GET count=%d", got, want, calls.Load())
			}
		})
	}
}

func TestHandshakeIOFailuresThroughLoopback(t *testing.T) {
	for _, name := range []string{"failed", "timeout", "canceled"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			target, _ := rawFixture(t, false, func(c net.Conn) {
				if name == "failed" {
					io.WriteString(c, "HTTP/1.1 400 Synthetic\r\n\r\n")
					return
				}
				if name == "canceled" {
					var b [1]byte
					c.Read(b[:]) // Wait until the client has actually entered TLS.
					cancel()
				}
				io.Copy(io.Discard, c)
			})
			dns, _ := dnsFixture(t, nil)
			w := &wire{local: netip.MustParseAddr("127.0.0.1"), dns: dns, testTarget: target, roots: x509.NewCertPool()}
			if name == "timeout" {
				w.testTimeout = 250 * time.Millisecond
			}
			_, got := w.fetch(ctx, cascollect.Request{Method: "GET", Target: "https://pass.neu.edu.cn/tpass/login"}, 16)
			if got != "tls_handshake_"+name {
				t.Fatalf("reason=%s", got)
			}
		})
	}
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

func TestResponseIOStageBoundaries(t *testing.T) {
	for _, stage := range []struct{ name, prefix string }{
		{"response_status", "HTTP/1.1 20"},
		{"response_headers", "HTTP/1.1 200 OK\r\nX-Synthetic: "},
		{"response_body", "HTTP/1.1 200 OK\r\n\r\nsynthetic-canary"},
	} {
		for _, suffix := range []string{"failed", "timeout", "canceled"} {
			t.Run(stage.name+"_"+suffix, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				if suffix == "canceled" {
					cancel()
				}
				defer cancel()
				reader := io.MultiReader(strings.NewReader(stage.prefix), failingReader{opaqueError{timeout: suffix == "timeout"}})
				e, got := readResponse(ctx, bufio.NewReader(reader), cascollect.Exchange{}, 128)
				if got != stage.name+"_"+suffix || len(e.Body) != 0 {
					t.Fatalf("reason=%s", got)
				}
			})
		}
	}
	for _, prefix := range []string{"", "HTTP/1.1 200 OK\r\n"} {
		_, reason := readResponse(context.Background(), bufio.NewReader(strings.NewReader(prefix)), cascollect.Exchange{}, 16)
		if reason != "response_status_failed" && reason != "response_headers_failed" {
			t.Fatal("EOF stage lost")
		}
	}
}

func TestTransportReportClosedAndCompatible(t *testing.T) {
	reasons := []string{"tls_unknown_authority", "tls_hostname_mismatch", "tls_certificate_time_invalid", "tls_certificate_invalid", "tls_certificate_denied", "request_deadline_setup_failed", "request_build_failed"}
	for _, stage := range []string{"tls_handshake", "request_write", "response_status", "response_headers", "response_body"} {
		for _, suffix := range []string{"timeout", "canceled", "failed"} {
			reasons = append(reasons, stage+"_"+suffix)
		}
	}
	for _, reason := range reasons {
		for _, constructed := range []bool{false, true} {
			r := blocked("anonymous-once", reason)
			r.ContextConstructed = constructed
			r.Attempts = cascollect.Attempts{Requests: 3, DNS: 3, Connects: 3}
			r.Operations.ExternalRequests, r.Operations.DNS = 3, 3
			b, err := Encode(r)
			if err != nil {
				t.Fatalf("reason rejected: %s", reason)
			}
			if _, err := Decode(b); err != nil {
				t.Fatal("roundtrip failed")
			}
			for _, mutation := range []func(*Report){
				func(r *Report) {
					r.Mode = "self-test"
					r.Attempts = cascollect.Attempts{}
					r.Operations.ExternalRequests = 0
					r.Operations.DNS = 0
				},
				func(r *Report) {
					r.Mode = "local-preflight"
					r.ContextConstructed = false
					r.Attempts = cascollect.Attempts{}
					r.Operations.ExternalRequests = 0
					r.Operations.DNS = 0
				},
				func(r *Report) {
					r.Attempts = cascollect.Attempts{}
					r.Operations.ExternalRequests = 0
					r.Operations.DNS = 0
				},
				func(r *Report) { r.Attempts.Connects-- },
				func(r *Report) { r.Reason += "_synthetic-canary" },
				func(r *Report) { r.Collection = "complete" },
			} {
				bad := r
				mutation(&bad)
				if _, err := Encode(bad); !errors.Is(err, ErrReport) {
					t.Fatal("contradictory report encoded")
				}
				b, _ := json.Marshal(bad)
				if _, err := Decode(b); !errors.Is(err, ErrReport) {
					t.Fatal("contradictory report decoded")
				}
			}
		}
	}
	r := blocked("anonymous-once", "request_failed")
	r.ContextConstructed = true
	r.Attempts = cascollect.Attempts{Requests: 3, DNS: 3, Connects: 3}
	r.Operations.ExternalRequests, r.Operations.DNS = 3, 3
	b, err := Encode(r)
	if err != nil {
		t.Fatal("historical reason rejected")
	}
	if _, err = Decode(b); err != nil {
		t.Fatal("historical report rejected")
	}
}
