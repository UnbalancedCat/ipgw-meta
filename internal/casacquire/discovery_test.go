package casacquire

import (
	"bufio"
	"context"
	"encoding/json"
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

func TestDiscoveryStopsAfterTwoLoopbackResponses(t *testing.T) {
	for _, tc := range []struct{ reason, body, location string }{
		{"discovery_no_match", "synthetic-canary", ""},
		{"discovery_body_boundary", "ac_id=7synthetic-canary", ""},
		{"discovery_body_value_rejected", "ac_id=12345678901", ""},
		{"discovery_body_conflict", "ac_id=7 ac_id=8", ""},
		{"discovery_query_auth_marker", "", "/portal?ac_id=7&token="},
		{"discovery_query_key_rejected", "", "/portal?ac_id=7&AC_ID=7"},
		{"discovery_query_duplicate", "", "/portal?ac_id=7&ac_id=7&extra=synthetic-canary"},
		{"discovery_query_invalid_value", "", "/portal?ac_id=bad&extra=synthetic-canary"},
		{"discovery_redirect_limit", "", "/third"},
	} {
		t.Run(tc.reason, func(t *testing.T) {
			dns, dnsCount := dnsFixture(t, nil)
			calls := &atomic.Int32{}
			target, _ := rawFixture(t, false, func(conn net.Conn) {
				req, err := http.ReadRequest(bufio.NewReader(conn))
				if err != nil {
					return
				}
				defer req.Body.Close()
				n := calls.Add(1)
				if req.Method != "GET" || req.Header.Get("Cookie") != "" || req.Header.Get("Authorization") != "" {
					t.Error("unexpected authenticated request")
				}
				if n == 1 && req.URL.Path == "/" {
					io.WriteString(conn, "HTTP/1.1 302 Found\r\nLocation: /next\r\nSet-Cookie: synthetic-canary\r\nContent-Length: 0\r\n\r\n")
					return
				}
				if n != 2 || req.URL.Path != "/next" {
					t.Error("unexpected request beyond discovery boundary")
				}
				if tc.location != "" {
					fmt.Fprintf(conn, "HTTP/1.1 302 Found\r\nLocation: %s\r\nContent-Length: %d\r\n\r\n%s", tc.location, len(tc.body), tc.body)
				} else {
					fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Length: %d\r\n\r\n%s", len(tc.body), tc.body)
				}
			})
			w := &wire{local: netip.MustParseAddr("127.0.0.1"), dns: dns, testTarget: target}
			r := acquire(context.Background(), w)
			if r.Reason != tc.reason || r.ContextConstructed || r.Collection != "blocked" || calls.Load() != 2 || dnsCount.Load() != 2 || r.Attempts != (cascollect.Attempts{Requests: 2, DNS: 2, Connects: 2}) {
				t.Fatalf("wrong classification or request boundary: %s", r.Reason)
			}
			data, err := Encode(r)
			if err != nil {
				t.Fatal(err)
			}
			for _, canary := range []string{"synthetic-canary", "ac_id=", "neu.edu.cn", "/next", "/portal", "/third"} {
				if strings.Contains(string(data), canary) {
					t.Fatal("raw discovery data in report")
				}
			}
		})
	}
}

func TestDiscoveryFailureProjection(t *testing.T) {
	for _, reason := range []string{"discovery_query_key_rejected", "discovery_query_auth_marker", "discovery_query_rejected", "discovery_query_duplicate", "discovery_query_invalid_value", "discovery_query_extra_fields", "discovery_body_boundary", "discovery_body_value_rejected", "discovery_body_conflict", "discovery_redirect_limit", "discovery_no_match"} {
		r := project("anonymous-once", cascollect.Report{Collection: "blocked", Reason: reason}, cascollect.Attempts{Requests: 2, DNS: 2, Connects: 2})
		b, err := Encode(r)
		if err != nil {
			t.Fatal(err)
		}
		if got, err := Decode(b); err != nil || got.Reason != reason || got.ContextConstructed || got.Analysis != nil {
			t.Fatal("lost discovery classification")
		}
		for _, mutate := range []func(*Report){
			func(r *Report) { r.ContextConstructed = true },
			func(r *Report) {
				r.Attempts = cascollect.Attempts{}
				r.Operations.ExternalRequests = 0
				r.Operations.DNS = 0
			},
			func(r *Report) {
				r.Attempts = cascollect.Attempts{Requests: 3, DNS: 3, Connects: 3}
				r.Operations.ExternalRequests = 3
				r.Operations.DNS = 3
			},
			func(r *Report) { r.Attempts.Connects-- },
			func(r *Report) { r.Mode = "local-preflight" },
			func(r *Report) { r.Mode = "self-test" },
			func(r *Report) { r.Reason += ":synthetic-canary" },
		} {
			bad := r
			mutate(&bad)
			data, _ := json.Marshal(bad)
			if _, err := Decode(data); err == nil {
				t.Fatal("accepted impossible or nonclosed report")
			}
		}
	}
	// Verify the real planner's fact projection, with entirely synthetic bytes.
	prefix := []cascollect.Exchange{
		{Target: "http://ipgw.neu.edu.cn/", Status: 302, Headers: []cascollect.Header{{Name: "Location", Value: "/next"}}},
		{Target: "http://ipgw.neu.edu.cn/next", Status: 200, Body: []byte("ac_id=7 ac_id=8 synthetic-canary")},
	}
	next, _, planned := cascollect.NextOfficial(context.Background(), prefix)
	r := project("anonymous-once", planned, cascollect.Attempts{Requests: 2, DNS: 2, Connects: 2})
	b, err := Encode(r)
	if next != nil || err != nil || r.Reason != "discovery_body_conflict" {
		t.Fatal("planner projection changed")
	}
	for _, canary := range []string{"neu.edu.cn", "ac_id=", "synthetic-canary", "/next"} {
		if strings.Contains(string(b), canary) {
			t.Fatal("discovery data leaked")
		}
	}
	r.Reason = "discovery_failed"
	if _, err := Encode(r); err != nil {
		t.Fatal("historical discovery receipt rejected")
	}
}
