package app

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/UnbalancedCat/ipgw-meta/internal/config"
)

type observedFake func(*http.Request) (*http.Response, error)

func (f observedFake) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestObservedConstructionUsesDefaultBoundTransport(t *testing.T) {
	for _, bind := range []string{"127.0.0.1", "192.0.2.10"} {
		t.Run(bind, func(t *testing.T) {
			calls := 0
			s := NewObserved(config.Paths{}, nil, io.Discard, func(next http.RoundTripper) http.RoundTripper {
				tr, ok := next.(*http.Transport)
				if !ok || tr.Proxy != nil || tr.TLSClientConfig != nil || tr.DialContext == nil || tr.TLSHandshakeTimeout == 0 {
					t.Fatal("default transport replaced")
				}
				// A canceled context exercises the dial entry without sending packets.
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				conn, err := tr.DialContext(ctx, "tcp", "127.0.0.1:1")
				if conn != nil {
					conn.Close()
					t.Fatal("canceled dial connected")
				}
				if err == nil {
					t.Fatal("expected cancellation")
				}
				calls++
				return observedFake(func(r *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("not_online_error")), Header: make(http.Header), Request: r}, nil
				})
			})
			client, err := s.client(bind, "")
			if err != nil {
				t.Fatal(err)
			}
			if _, err = client.Status(context.Background()); err != nil || calls != 1 {
				t.Fatal("decorator unavailable")
			}
			if New(config.Paths{}, nil, io.Discard).decorate != nil {
				t.Fatal("ordinary constructor enabled observation")
			}
		})
	}
}

func TestObservedTransportHonorsSelectedSourceOverProfile(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	checked := false
	s := NewObserved(config.Paths{}, nil, io.Discard, func(next http.RoundTripper) http.RoundTripper {
		tr, ok := next.(*http.Transport)
		if !ok {
			t.Fatal("not default transport")
		}
		conn, err := tr.DialContext(context.Background(), "tcp", listener.Addr().String())
		if err != nil {
			t.Fatal("loopback dial failed")
		}
		defer conn.Close()
		local, ok := conn.LocalAddr().(*net.TCPAddr)
		if !ok || !local.IP.Equal(net.ParseIP("127.0.0.1")) {
			t.Fatal("source override was lost")
		}
		checked = true
		return next
	})
	if _, err = s.client("127.0.0.1", "192.0.2.17"); err != nil || !checked {
		t.Fatal("source binding not checked")
	}
}
