package authdev

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	ipgw "github.com/UnbalancedCat/ipgw-meta"
	"github.com/UnbalancedCat/ipgw-meta/internal/castest"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func syntheticKey(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(der)
}
func response(req *http.Request, status int, body string, header http.Header) *http.Response {
	if header == nil {
		header = make(http.Header)
	}
	return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(strings.NewReader(body)), Request: req}
}

type trace struct {
	Method, URL string
	Header      http.Header
}

func TestAnonymousMatchesUnobservedSDKSession(t *testing.T) {
	key := syntheticKey(t)
	for _, inline := range []bool{false, true} {
		t.Run(map[bool]string{false: "external", true: "inline"}[inline], func(t *testing.T) {
			makeTransport := func(log *[]trace) http.RoundTripper {
				return roundTripFunc(func(req *http.Request) (*http.Response, error) {
					*log = append(*log, trace{req.Method, req.URL.String(), req.Header.Clone()})
					switch req.URL.Hostname() + req.URL.Path {
					case "ipgw.neu.edu.cn/cgi-bin/rad_user_info":
						return response(req, 200, "not_online_error", nil), nil
					case "ipgw.neu.edu.cn/":
						if req.Header.Get("Cookie") != "" {
							t.Fatal("discovery cookie")
						}
						return response(req, 302, "", http.Header{"Location": []string{"/srun_portal_pc"}, "Set-Cookie": []string{"discovery_canary=never_replay"}}), nil
					case "ipgw.neu.edu.cn/srun_portal_pc":
						if req.Header.Get("Cookie") != "" {
							t.Fatal("discovery replayed cookie")
						}
						return response(req, 200, "ac_id=1", nil), nil
					case "pass.neu.edu.cn/tpass/login":
						body := `<html><form id="loginForm"><input name="lt" value="ephemeral-lt-canary"><input name="execution" value="ephemeral-execution-canary"></form><script src="/tpass/login.js"></script></html>`
						if inline {
							body += `<script>` + castest.LoginScript(key) + `</script>`
						}
						return response(req, 200, body, http.Header{"Set-Cookie": []string{"session_canary=memory_only; Path=/tpass; Secure"}}), nil
					case "pass.neu.edu.cn/tpass/login.js":
						if req.Header.Get("Cookie") != "session_canary=memory_only" {
							t.Fatal("CAS session not preserved")
						}
						return response(req, 200, castest.LoginScript(key), nil), nil
					default:
						t.Fatal("unexpected request")
						return nil, errStop
					}
				})
			}
			var plain, observed []trace
			client, err := ipgw.NewClient(ipgw.WithRoundTripper(makeTransport(&plain)))
			if err != nil {
				t.Fatal(err)
			}
			providerCalls := 0
			_, err = client.Login(context.Background(), ipgw.LoginRequest{ExpectedUsername: anonymousIdentity, Method: ipgw.AuthMethodPassword, Switch: ipgw.SwitchRefuse, Credentials: ipgw.CredentialProviderFunc(func(context.Context, ipgw.CredentialRequest) (ipgw.Credential, error) {
				providerCalls++
				return ipgw.Credential{}, errStop
			})})
			if err == nil || providerCalls != 1 {
				t.Fatal("reference did not reach provider")
			}
			r := Anonymous(context.Background(), makeTransport(&observed), func(context.Context) bool { return true })
			if !reflect.DeepEqual(plain, observed) {
				t.Fatal("observer changed SDK request trace")
			}
			// ADR-0021 always reads the advertised script, including with an inline key.
			if r.State != "observed" || r.Reason != "credential_boundary_reached" || r.ActionService != "same" || !r.CookieReceived || !r.CookieSent {
				t.Fatalf("bad projection: %+v", r)
			}
			data, err := Encode(r)
			if err != nil {
				t.Fatal(err)
			}
			for _, secret := range []string{key, "ephemeral-lt-canary", "ephemeral-execution-canary", "session_canary", "memory_only", "service="} {
				if strings.Contains(string(data), secret) {
					t.Fatal("projection leaked material")
				}
			}
			if _, err = Decode(data); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAnonymousInitialOnlineAndEarlySSOStop(t *testing.T) {
	for _, online := range []bool{false, true} {
		calls := 0
		tr := roundTripFunc(func(req *http.Request) (*http.Response, error) {
			calls++
			if calls == 1 {
				body := "not_online_error"
				if online {
					body = `{"error":"ok","user_name":"synthetic-account","online_ip":"192.0.2.1"}`
				}
				return response(req, 200, body, nil), nil
			}
			if calls == 2 {
				return response(req, 200, "ac_id=1", nil), nil
			}
			if calls == 3 {
				u, err := url.Parse(req.URL.Query().Get("service"))
				if err != nil {
					t.Fatal(err)
				}
				q := u.Query()
				q.Set("ticket", "synthetic-ticket-canary")
				u.RawQuery = q.Encode()
				return response(req, 302, "", http.Header{"Location": []string{u.String()}}), nil
			}
			t.Fatal("SSO sent activation")
			return nil, errStop
		})
		r := Anonymous(context.Background(), tr, func(context.Context) bool { return true })
		want, count := "early_ticket", 3
		if online {
			want, count = "initial_online", 1
		}
		if r.Reason != want || r.State != "blocked" || calls != count {
			t.Fatalf("unexpected result: %+v", r)
		}
	}
}

func TestGuardRefusesBeforeTransport(t *testing.T) {
	for _, tc := range []struct {
		name, method, target, cookie string
		budget                       bool
		check                        bool
		reason                       string
	}{
		{"post", "POST", "https://pass.neu.edu.cn/tpass/login", "", false, true, "request_rejected"},
		{"outside", "GET", "https://outside.invalid/", "", false, true, "request_rejected"},
		{"http-cookie", "GET", "http://ipgw.neu.edu.cn/", "secret", false, true, "request_rejected"},
		{"activation", "GET", "https://ipgw.neu.edu.cn/v1/srun_portal_sso", "", false, true, "early_ticket"},
		{"logout", "GET", "https://ipgw.neu.edu.cn/cgi-bin/srun_portal", "", false, true, "request_rejected"},
		{"cas-logout", "GET", "https://pass.neu.edu.cn/tpass/logout", "", false, true, "request_rejected"},
		{"budget", "GET", "https://pass.neu.edu.cn/tpass/login", "", true, true, "request_budget"},
		{"topology", "GET", "https://pass.neu.edu.cn/tpass/login", "", false, false, "topology_changed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := NewReport("anonymous")
			r.StatusRequests = 1
			r.DiscoveryRequests = 1
			r.Requests = 2
			if tc.budget {
				r.Requests = MaxRequests
			}
			g := guard{report: &r, check: func(context.Context) bool { return tc.check }, next: roundTripFunc(func(*http.Request) (*http.Response, error) { t.Fatal("transport called"); return nil, errStop })}
			req, _ := http.NewRequest(tc.method, tc.target, nil)
			if tc.cookie != "" {
				req.Header.Set("Cookie", tc.cookie)
			}
			if _, err := g.RoundTrip(req); err == nil || g.reason != tc.reason {
				t.Fatal("guard failed")
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := Anonymous(ctx, roundTripFunc(func(*http.Request) (*http.Response, error) { t.Fatal("canceled transport"); return nil, errStop }), func(context.Context) bool { return true })
	if r.Reason != "canceled" {
		t.Fatal(r.Reason)
	}
}

type terminalErrorReader struct{ err error }

func (r terminalErrorReader) Read(p []byte) (int, error) { copy(p, "x"); return 1, r.err }
func (terminalErrorReader) Close() error                 { return nil }
func TestPassiveBodyPreservesErrorsAndBounds(t *testing.T) {
	sentinel := errors.New("synthetic read error")
	for _, ending := range []error{sentinel, io.EOF} {
		called := false
		b := &observedBody{ReadCloser: terminalErrorReader{ending}, complete: func(data []byte) {
			called = true
			if string(data) != "x" {
				t.Fatal("changed bytes")
			}
		}}
		p := make([]byte, 10)
		n, err := b.Read(p)
		if n != 1 || err != ending || called != (ending == io.EOF) {
			t.Fatal("changed read contract")
		}
		_ = b.Close()
	}
	called := false
	b := &observedBody{ReadCloser: io.NopCloser(strings.NewReader(strings.Repeat("x", bodyLimit+1))), complete: func([]byte) { called = true }}
	if n, err := io.Copy(io.Discard, b); err != nil || n != bodyLimit+1 {
		t.Fatal("changed large body")
	}
	_ = b.Close()
	if called || b.data != nil {
		t.Fatal("observed truncated body")
	}
	b = &observedBody{ReadCloser: io.NopCloser(strings.NewReader("not-consumed")), complete: func([]byte) { t.Fatal("drained on Close") }}
	_ = b.Close()
}

func TestSelfTestIsSynthetic(t *testing.T) {
	r := SelfTest()
	if r.State != "pass" || r.Validate() != nil {
		t.Fatalf("self test: %+v", r)
	}
}
