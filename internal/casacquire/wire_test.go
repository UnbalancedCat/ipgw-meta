package casacquire

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
	"golang.org/x/net/dns/dnsmessage"
)

func dnsFixture(t *testing.T, mutate func(*dnsmessage.Message)) (netip.AddrPort, *atomic.Int32) {
	t.Helper()
	c, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	count := &atomic.Int32{}
	go func() {
		defer close(done)
		for {
			buf := make([]byte, 4096)
			n, peer, err := c.ReadFromUDP(buf)
			if err != nil {
				return
			}
			count.Add(1)
			var q dnsmessage.Message
			if q.Unpack(buf[:n]) != nil || len(q.Questions) != 1 {
				continue
			}
			m := dnsmessage.Message{Header: dnsmessage.Header{ID: q.ID, Response: true}, Questions: q.Questions, Answers: []dnsmessage.Resource{{Header: dnsmessage.ResourceHeader{Name: q.Questions[0].Name, Class: dnsmessage.ClassINET, Type: dnsmessage.TypeA}, Body: &dnsmessage.AResource{A: [4]byte{127, 0, 0, 1}}}}}
			if mutate != nil {
				mutate(&m)
			}
			packet, err := m.Pack()
			if err == nil {
				_, _ = c.WriteToUDP(packet, peer)
			}
		}
	}()
	t.Cleanup(func() { c.Close(); <-done })
	return c.LocalAddr().(*net.UDPAddr).AddrPort(), count
}

func testCertificate(t *testing.T) (tls.Certificate, *x509.CertPool) {
	return testCertificateWith(t, func(*x509.Certificate) {})
}

func testCertificateWith(t *testing.T, change func(*x509.Certificate)) (tls.Certificate, *x509.CertPool) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	cert := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "synthetic"}, DNSNames: []string{"pass.neu.edu.cn", "ipgw.neu.edu.cn"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	change(cert)
	der, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	k, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	pair, err := tls.X509KeyPair(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: k}))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(parsed)
	return pair, roots
}

func rawFixture(t *testing.T, secure bool, handler func(net.Conn)) (string, *tls.Config) {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var config *tls.Config
	if secure {
		pair, roots := testCertificate(t)
		listener = tls.NewListener(listener, &tls.Config{Certificates: []tls.Certificate{pair}, MinVersion: tls.VersionTLS12})
		config = &tls.Config{RootCAs: roots, ServerName: "pass.neu.edu.cn", MinVersion: tls.VersionTLS12}
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			c, err := listener.Accept()
			if err != nil {
				return
			}
			_ = c.SetDeadline(time.Now().Add(2 * time.Second))
			handler(c)
			c.Close()
		}
	}()
	t.Cleanup(func() { listener.Close(); <-done })
	return listener.Addr().String(), config
}

func fixtureWire(t *testing.T, secure bool, response string) (*wire, *atomic.Int32, *atomic.Int32) {
	dns, dnsCount := dnsFixture(t, nil)
	requests := &atomic.Int32{}
	target, config := rawFixture(t, secure, func(c net.Conn) {
		req, err := http.ReadRequest(bufio.NewReader(c))
		if err != nil {
			return
		}
		requests.Add(1)
		req.Body.Close()
		if req.Method != "GET" || req.Header.Get("Cookie") != "" || req.Header.Get("Authorization") != "" || req.Header.Get("Accept-Encoding") != "identity" || !req.Close {
			t.Error("unexpected outgoing request")
		}
		_, _ = io.WriteString(c, response)
	})
	return &wire{local: netip.MustParseAddr("127.0.0.1"), dns: dns, testTarget: target, testTLS: config}, dnsCount, requests
}

func TestWireTLSCookieAndCounts(t *testing.T) {
	w, dns, reqs := fixtureWire(t, true, "HTTP/1.1 200 OK\r\nContent-Type: text/html\r\nSet-Cookie: synthetic-secret\r\nContent-Length: 2\r\n\r\nok")
	for i := 0; i < 2; i++ {
		e, reason := w.fetch(context.Background(), cascollect.Request{Method: "GET", Target: "https://pass.neu.edu.cn/tpass/login"}, 16)
		if reason != "" || string(e.Body) != "ok" {
			t.Fatal(reason)
		}
		for _, h := range e.Headers {
			if strings.Contains(h.Value, "synthetic-secret") {
				t.Fatal("cookie retained")
			}
		}
	}
	if dns.Load() != 2 || reqs.Load() != 2 || w.counts != (cascollect.Attempts{Requests: 2, DNS: 2, Connects: 2}) {
		t.Fatal("wrong attempts")
	}
}
func TestWireRejectsCertificateBeforeGET(t *testing.T) {
	w, _, requests := fixtureWire(t, true, "HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")
	w.testTLS = &tls.Config{RootCAs: x509.NewCertPool(), ServerName: "pass.neu.edu.cn", MinVersion: tls.VersionTLS12}
	_, reason := w.fetch(context.Background(), cascollect.Request{Method: "GET", Target: "https://pass.neu.edu.cn/tpass/login"}, 16)
	if reason != "tls_unknown_authority" || requests.Load() != 0 || w.counts != (cascollect.Attempts{Requests: 1, DNS: 1, Connects: 1}) {
		t.Fatal("TLS did not fail closed")
	}
}

func TestOfflineRootsWithProductionTLSConfiguration(t *testing.T) {
	w, _, requests := fixtureWire(t, true, "HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")
	w.roots = w.testTLS.RootCAs
	w.testTLS = nil
	_, reason := w.fetch(context.Background(), cascollect.Request{Method: "GET", Target: "https://pass.neu.edu.cn/tpass/login"}, 16)
	if reason != "" || requests.Load() != 1 {
		t.Fatal("offline root verification failed")
	}
}
func TestWireNoRedirectOrRetry(t *testing.T) {
	w, dns, requests := fixtureWire(t, false, "HTTP/1.1 302 Found\r\nLocation: https://outside.invalid/private\r\nContent-Length: 0\r\n\r\n")
	r := acquire(context.Background(), w)
	if r.Collection != "blocked" || r.Reason != "url_rejected" || dns.Load() != 1 || requests.Load() != 1 {
		t.Fatal("redirect followed")
	}
}
func TestWireDNSFailureDoesNotConnect(t *testing.T) {
	dns, count := dnsFixture(t, func(m *dnsmessage.Message) { m.Truncated = true })
	w := &wire{local: netip.MustParseAddr("127.0.0.1"), dns: dns, testTarget: "127.0.0.1:1"}
	r := acquire(context.Background(), w)
	if r.Reason != "dns_failed" || count.Load() != 1 || r.Attempts != (cascollect.Attempts{Requests: 1, DNS: 1}) {
		t.Fatal("DNS fallback")
	}
}
func TestWireLimitsAndFailures(t *testing.T) {
	for _, tc := range []struct {
		name, response string
		limit          int
		want           string
	}{
		{"compressed", "HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\n\r\n", 4, "response_rejected"},
		{"chunked", "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n0\r\n\r\n", 4, "response_rejected"},
		{"duplicate", "HTTP/1.1 200 OK\r\nContent-Length: 0\r\nContent-Length: 0\r\n\r\n", 4, "response_rejected"},
		{"length", "HTTP/1.1 200 OK\r\nContent-Length: 3\r\n\r\nx", 4, "response_rejected"},
		{"body", "HTTP/1.1 200 OK\r\n\r\n12345", 4, "resource_limit"},
		{"header", "HTTP/1.1 200 OK\r\nX: " + strings.Repeat("x", cascollect.MaxHeaderBytes) + "\r\n\r\n", 4, "resource_limit"},
		{"folding", "HTTP/1.1 200 OK\r\n x: y\r\n\r\n", 4, "response_rejected"},
		{"interim", "HTTP/1.1 100 Continue\r\n\r\n", 4, "response_rejected"},
		{"exact", "HTTP/1.1 200 OK\r\nContent-Length: 4\r\n\r\n1234", 4, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w, dns, requests := fixtureWire(t, false, tc.response)
			_, reason := w.fetch(context.Background(), cascollect.Request{Method: "GET", Target: "http://ipgw.neu.edu.cn/"}, tc.limit)
			if reason != tc.want || dns.Load() != 1 || requests.Load() != 1 {
				t.Fatalf("reason=%s want=%s", reason, tc.want)
			}
		})
	}
}
func TestWireCancellationAndHardBudget(t *testing.T) {
	dns, _ := dnsFixture(t, nil)
	closed := make(chan struct{}, 1)
	target, _ := rawFixture(t, false, func(c net.Conn) {
		req, err := http.ReadRequest(bufio.NewReader(c))
		if err != nil {
			return
		}
		req.Body.Close()
		_, _ = io.Copy(io.Discard, c)
		closed <- struct{}{}
	})
	w := &wire{local: netip.MustParseAddr("127.0.0.1"), dns: dns, testTarget: target, testTimeout: 100 * time.Millisecond}
	start := time.Now()
	_, reason := w.fetch(context.Background(), cascollect.Request{Method: "GET", Target: "http://ipgw.neu.edu.cn/"}, 4)
	if reason == "" || time.Since(start) > time.Second {
		t.Fatal("deadline ignored")
	}
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("connection not closed")
	}
	w.counts = cascollect.Attempts{Requests: 12, DNS: 12, Connects: 12}
	_, reason = w.fetch(context.Background(), cascollect.Request{Method: "GET", Target: "http://ipgw.neu.edu.cn/"}, 4)
	if reason != "budget_exceeded" || w.counts.Requests != 12 {
		t.Fatal("budget exceeded")
	}
}
func TestDNSValidation(t *testing.T) {
	name, _ := dnsmessage.NewName("example.invalid.")
	alias, _ := dnsmessage.NewName("alias.invalid.")
	base := dnsmessage.Message{Header: dnsmessage.Header{ID: 7, Response: true}, Questions: []dnsmessage.Question{{Name: name, Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET}}, Answers: []dnsmessage.Resource{{Header: dnsmessage.ResourceHeader{Name: name, Type: dnsmessage.TypeCNAME, Class: dnsmessage.ClassINET}, Body: &dnsmessage.CNAMEResource{CNAME: alias}}, {Header: dnsmessage.ResourceHeader{Name: alias, Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET}, Body: &dnsmessage.AResource{A: [4]byte{192, 0, 2, 1}}}}}
	for _, tc := range []struct {
		name   string
		mutate func(*dnsmessage.Message)
		ok     bool
	}{
		{"cname", func(*dnsmessage.Message) {}, true},
		{"id", func(m *dnsmessage.Message) { m.ID++ }, false},
		{"rcode", func(m *dnsmessage.Message) { m.RCode = dnsmessage.RCodeNameError }, false},
		{"truncated", func(m *dnsmessage.Message) { m.Truncated = true }, false},
		{"question", func(m *dnsmessage.Message) { m.Questions = nil }, false},
		{"missing-cname-target", func(m *dnsmessage.Message) { m.Answers = m.Answers[:1] }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := base
			tc.mutate(&m)
			b, err := m.Pack()
			if err != nil {
				t.Fatal(err)
			}
			_, ok := dnsAnswer(b, 7, name)
			if ok != tc.ok {
				t.Fatal("unexpected DNS acceptance")
			}
		})
	}
}

type bufferedTestConn struct {
	net.Conn
	reader *bufio.Reader
}

func (c *bufferedTestConn) Read(p []byte) (int, error) { return c.reader.Read(p) }

func TestAcquisitionThroughLoopbackDNSHTTPAndTLS(t *testing.T) {
	for _, redirect := range []bool{false, true} {
		t.Run(fmt.Sprint(redirect), func(t *testing.T) {
			pair, roots := testCertificate(t)
			dns, dnsCount := dnsFixture(t, nil)
			calls := &atomic.Int32{}
			target, _ := rawFixture(t, false, func(raw net.Conn) {
				reader := bufio.NewReader(raw)
				prefix, err := reader.Peek(1)
				if err != nil {
					return
				}
				var conn net.Conn = &bufferedTestConn{Conn: raw, reader: reader}
				if prefix[0] == 22 {
					secure := tls.Server(conn, &tls.Config{Certificates: []tls.Certificate{pair}, MinVersion: tls.VersionTLS12})
					if secure.Handshake() != nil {
						return
					}
					defer secure.Close()
					conn = secure
				}
				req, err := http.ReadRequest(bufio.NewReader(conn))
				if err != nil {
					return
				}
				defer req.Body.Close()
				calls.Add(1)
				if req.Header.Get("Cookie") != "" || req.Header.Get("Authorization") != "" || req.Method != "GET" {
					t.Error("authenticated request")
				}
				body, media := "", "text/html"
				switch req.URL.Path {
				case "/":
					if req.Host != "ipgw.neu.edu.cn" {
						t.Error("wrong discovery origin")
					}
					body = "ac_id=007"
				case "/tpass/login":
					if req.Host != "pass.neu.edu.cn" || req.URL.Query().Get("service") != "http://ipgw.neu.edu.cn/srun_portal_sso?ac_id=007" {
						t.Error("wrong service context")
					}
					if redirect {
						io.WriteString(conn, "HTTP/1.1 302 Found\r\nLocation: /session\r\nContent-Length: 0\r\n\r\n")
						return
					}
					body = `<script>var publicKeyStr = "synthetic-invalid";</script><script src="/key.js"></script><script src="/key.js"></script>`
				case "/key.js":
					body = `var publicKeyStr = "synthetic-other";`
					media = "application/javascript"
				default:
					t.Error("unexpected request")
				}
				fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: %s\r\nSet-Cookie: private-synthetic-canary\r\nContent-Length: %d\r\n\r\n%s", media, len(body), body)
			})
			w := &wire{local: netip.MustParseAddr("127.0.0.1"), dns: dns, testTarget: target, roots: roots}
			r := acquire(context.Background(), w)
			if redirect {
				if r.Collection != "blocked" || r.Reason != "response_rejected" || calls.Load() != 2 {
					t.Fatal("CAS redirect followed")
				}
			} else {
				if r.Collection != "complete" || len(r.Resources) != 1 || len(r.References) != 2 || calls.Load() != 3 {
					t.Fatalf("collection %s/%s", r.Collection, r.Reason)
				}
			}
			if dnsCount.Load() != calls.Load() || r.Attempts.Requests != int(calls.Load()) || r.Attempts.Connects != int(calls.Load()) {
				t.Fatal("physical attempt accounting")
			}
			data, err := Encode(r)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = Decode(data); err != nil {
				t.Fatal(err)
			}
			for _, secret := range []string{"private-synthetic-canary", "synthetic-invalid", "synthetic-other", "neu.edu.cn", "ac_id"} {
				if strings.Contains(string(data), secret) {
					t.Fatal("raw data in projection")
				}
			}
		})
	}
}
