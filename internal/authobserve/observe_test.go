package authobserve

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/UnbalancedCat/ipgw-meta/internal/castest"
	"io"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"

	ipgw "github.com/UnbalancedCat/ipgw-meta"
)

type forbiddenBody struct{}

func (forbiddenBody) Read([]byte) (int, error) { panic("observation must not read body") }
func (forbiddenBody) Close() error             { panic("observation must not close body") }
func TestPassiveTransportDoesNotTouchMaterials(t *testing.T) {
	c := New("login")
	req, _ := http.NewRequest("POST", "https://pass.neu.edu.cn/tpass/auth?private=synthetic", forbiddenBody{})
	req.Header.Set("Private-Canary", "synthetic-private-header")
	response := &http.Response{StatusCode: 500, Body: forbiddenBody{}}
	original := errors.New("synthetic-private-error")
	tr := c.Decorate(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r != req {
			t.Fatal("request replaced")
		}
		return response, original
	}))
	got, err := tr.RoundTrip(req)
	if got != response || err != original {
		t.Fatal("result changed")
	}
	c.result(&ipgw.Error{Code: ipgw.CodeNetwork, Message: "synthetic-private-message"})
	report := c.Snapshot()
	data, _ := json.Marshal(report)
	if report.Validate() != nil || strings.Contains(string(data), "synthetic") || len(report.Requests) != 1 || report.Requests[0].StatusCode != 500 || report.Requests[0].Transport != "other" {
		t.Fatal("invalid or leaking observation")
	}
}
func TestOverflowAndConcurrentObservationNeverChangeCalls(t *testing.T) {
	c := New("status")
	var mu sync.Mutex
	calls := 0
	tr := c.Decorate(roundTripFunc(func(*http.Request) (*http.Response, error) {
		mu.Lock()
		calls++
		mu.Unlock()
		return &http.Response{StatusCode: 200}, nil
	}))
	var wg sync.WaitGroup
	for i := 0; i < 80; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("GET", "https://ipgw.neu.edu.cn/cgi-bin/rad_user_info", nil)
			if _, err := tr.RoundTrip(req); err != nil {
				t.Error("changed result")
			}
		}()
	}
	wg.Wait()
	c.result(nil)
	r := c.Snapshot()
	if calls != 80 || r.State != "incomplete" || len(r.Requests) != 0 || r.Validate() != nil {
		t.Fatal("overflow truncated or altered calls")
	}
}
func TestTypedTransportClassification(t *testing.T) {
	for _, tt := range []struct {
		err  error
		want string
	}{
		{nil, "none"}, {context.Canceled, "canceled"}, {context.DeadlineExceeded, "timeout"},
		{&net.DNSError{Err: "synthetic", Name: "synthetic.invalid"}, "dns"},
		{&net.DNSError{Err: "synthetic", IsTimeout: true}, "timeout"},
		{x509.UnknownAuthorityError{}, "tls_certificate"}, {tls.RecordHeaderError{}, "tls_other"},
		{errors.New("tls: synthetic-private-message"), "other"},
	} {
		if got := transportCategory(tt.err); got != tt.want {
			t.Fatalf("got %s want %s", got, tt.want)
		}
	}
}

type exchange struct{ method, path, query, cookie, body, ua string }

// Requests are sent only to this fake. Request material is compared in test
// memory; no actual HTTP client transport, profile or password prompt is used.
func TestSDKTraceEquivalentWithAndWithoutObservation(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	der, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	public := base64.StdEncoding.EncodeToString(der)
	for _, scenario := range []string{"success", "cas500", "cas_timeout", "activation503", "final_status_error", "redirect", "body_error", "cancel"} {
		t.Run(scenario, func(t *testing.T) {
			var tracks [2][]exchange
			var results [2]ipgw.LoginResult
			var codes [2]ipgw.ErrorCode
			var providerCalls [2]int
			observed := New("login")
			for mode := 0; mode < 2; mode++ {
				statusCalls := 0
				fake := roundTripFunc(func(req *http.Request) (*http.Response, error) {
					record := exchange{method: req.Method, path: req.URL.Path, query: req.URL.RawQuery, cookie: req.Header.Get("Cookie"), ua: req.Header.Get("User-Agent")}
					if req.Body != nil && req.Body != http.NoBody {
						data, _ := io.ReadAll(req.Body)
						form, _ := url.ParseQuery(string(data))
						if form.Get("rsa") != "" {
							cipher, _ := base64.StdEncoding.DecodeString(form.Get("rsa"))
							clear, err := rsa.DecryptPKCS1v15(rand.Reader, key, cipher)
							if err != nil {
								t.Fatal("synthetic RSA failure")
							}
							form.Set("rsa", string(clear))
						}
						record.body = form.Encode()
					}
					tracks[mode] = append(tracks[mode], record)
					response := &http.Response{StatusCode: 200, Header: make(http.Header), Request: req}
					body := ""
					switch {
					case req.URL.Path == "/cgi-bin/rad_user_info":
						statusCalls++
						body = "not_online_error"
						if statusCalls > 1 {
							if scenario == "final_status_error" {
								return nil, errors.New("synthetic-private-error")
							}
							body = `{"error":"ok","user_name":"synthetic-account","online_ip":"192.0.2.9"}`
						}
					case req.URL.Scheme == "http" && req.URL.Path == "/":
						body = "ac_id=1"
					case req.Method == "GET" && req.URL.Path == "/tpass/login":
						body = `<form id="loginForm"><input name="lt" value="synthetic-state"><input name="execution" value="synthetic-execution"></form><script>` + castest.LoginScript(public) + `</script>`
						response.Header.Set("Set-Cookie", "synthetic_cookie=private; Path=/tpass; Secure")
					case req.Method == "POST" && req.URL.Hostname() == "pass.neu.edu.cn":
						if scenario == "cas500" {
							response.StatusCode = 500
							break
						}
						if scenario == "cas_timeout" {
							return nil, context.DeadlineExceeded
						}
						if scenario == "cancel" {
							return nil, context.Canceled
						}
						if scenario == "body_error" {
							response.Body = badReader{}
							return response, nil
						}
						response.StatusCode = 302
						location := "http://ipgw.neu.edu.cn/srun_portal_sso?ac_id=1&ticket=synthetic-ticket"
						if scenario == "redirect" {
							location = "https://pass.neu.edu.cn/tpass/redirect"
						}
						response.Header.Set("Location", location)
					case req.URL.Path == "/tpass/redirect":
						response.StatusCode = 500
					case req.URL.Path == "/v1/srun_portal_sso":
						body = `{"code":0,"message":"success"}`
						if scenario == "activation503" {
							response.StatusCode = 503
						}
					default:
						t.Fatal("unexpected synthetic request")
					}
					response.Body = io.NopCloser(strings.NewReader(body))
					return response, nil
				})
				var rt http.RoundTripper = fake
				if mode == 1 {
					rt = observed.Decorate(fake)
				}
				client, e := ipgw.NewClient(ipgw.WithRoundTripper(rt))
				if e != nil {
					t.Fatal(e)
				}
				results[mode], e = client.Login(context.Background(), ipgw.LoginRequest{Method: ipgw.AuthMethodPassword, ExpectedUsername: "synthetic-account", Switch: ipgw.SwitchRefuse, Credentials: ipgw.CredentialProviderFunc(func(context.Context, ipgw.CredentialRequest) (ipgw.Credential, error) {
					providerCalls[mode]++
					return ipgw.Credential{Password: "synthetic-password"}, nil
				})})
				codes[mode] = ipgw.CodeOf(e)
				if mode == 1 {
					observed.result(e)
				}
			}
			results[0].Status.ObservedAt = results[1].Status.ObservedAt
			if !reflect.DeepEqual(tracks[0], tracks[1]) || !reflect.DeepEqual(results[0], results[1]) || codes[0] != codes[1] || providerCalls[0] != 1 || providerCalls[1] != 1 {
				t.Fatal("observer changed SDK semantics")
			}
			r := observed.Snapshot()
			data, _ := json.Marshal(r)
			if r.Validate() != nil || strings.Contains(string(data), "synthetic-") {
				t.Fatal("invalid or leaking report")
			}
			if scenario == "cas500" && r.SDKCategory != "cas_http_status" || scenario == "activation503" && r.SDKCategory != "activation_http_status" || scenario == "success" && results[1].Outcome != ipgw.LoginLoggedIn {
				t.Fatal("unexpected classification")
			}
		})
	}
}

type badReader struct{}

func (badReader) Read([]byte) (int, error) { return 0, errors.New("synthetic-private-body-error") }
func (badReader) Close() error             { return nil }
func TestSelfTest(t *testing.T) {
	if !SelfTest() {
		t.Fatal("selftest")
	}
}
