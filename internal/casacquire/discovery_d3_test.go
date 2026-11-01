package casacquire

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
)

func TestD3OnlyRebuiltCASRequestThroughLoopback(t *testing.T) {
	for _, hop := range []bool{false, true} {
		t.Run(fmt.Sprint(hop), func(t *testing.T) {
			pair, roots := testCertificate(t)
			dns, dnsCount := dnsFixture(t, nil)
			calls := &atomic.Int32{}
			target, _ := rawFixture(t, false, func(raw net.Conn) {
				reader := bufio.NewReader(raw)
				first, err := reader.Peek(1)
				if err != nil {
					return
				}
				var conn net.Conn = &bufferedTestConn{Conn: raw, reader: reader}
				secure := first[0] == 22
				if secure {
					tlsConn := tls.Server(conn, &tls.Config{Certificates: []tls.Certificate{pair}, MinVersion: tls.VersionTLS12})
					if tlsConn.Handshake() != nil {
						return
					}
					defer tlsConn.Close()
					conn = tlsConn
				}
				req, err := http.ReadRequest(bufio.NewReader(conn))
				if err != nil {
					return
				}
				defer req.Body.Close()
				calls.Add(1)
				if req.Method != "GET" || req.Header.Get("Cookie") != "" || req.Header.Get("Authorization") != "" || req.ContentLength > 0 {
					t.Error("unexpected request state")
				}
				if strings.Contains(req.RequestURI, "hint") || strings.Contains(req.RequestURI, "synthetic") || strings.Contains(req.RequestURI, "never-follow") {
					t.Error("discarded Location data sent")
				}
				switch req.URL.Path {
				case "/", "/next":
					if req.Host != "ipgw.neu.edu.cn" || secure || req.URL.RawQuery != "" || !hop && req.URL.Path == "/next" {
						t.Error("discovery scope changed")
					}
					location := "/never-follow?hint=https%3A%2F%2Foutside.invalid%2Fsecret%3Ftoken%3Dsynthetic-canary&ac_id=007"
					if hop && req.URL.Path == "/" {
						location = "/next"
					}
					fmt.Fprintf(conn, "HTTP/1.1 302 Found\r\nLocation: %s\r\nSet-Cookie: synthetic-canary\r\nContent-Length: 0\r\n\r\n", location)
				case "/tpass/login":
					if !secure || req.Host != "pass.neu.edu.cn" || req.RequestURI != "/tpass/login?service=http%3A%2F%2Fipgw.neu.edu.cn%2Fsrun_portal_sso%3Fac_id%3D007" {
						t.Error("CAS request was not rebuilt exclusively from ac_id")
					}
					io.WriteString(conn, "HTTP/1.1 200 OK\r\nContent-Type: text/html\r\nContent-Length: 13\r\n\r\n<html></html>")
				default:
					t.Error("followed unapproved path")
				}
			})
			w := &wire{local: netip.MustParseAddr("127.0.0.1"), dns: dns, testTarget: target, roots: roots}
			r := acquire(context.Background(), w)
			want := 2
			if hop {
				want = 3
			}
			if r.Collection != "complete" || !r.ContextConstructed || int(calls.Load()) != want || int(dnsCount.Load()) != want || r.Attempts != (cascollect.Attempts{Requests: want, DNS: want, Connects: want}) {
				t.Fatalf("unexpected result: %s/%s", r.Collection, r.Reason)
			}
			b, err := Encode(r)
			if err != nil {
				t.Fatal(err)
			}
			for _, value := range []string{"synthetic-canary", "outside.invalid", "never-follow", "ac_id=", "hint", "neu.edu.cn"} {
				if strings.Contains(string(b), value) {
					t.Fatal("discovery data leaked")
				}
			}
		})
	}
}
