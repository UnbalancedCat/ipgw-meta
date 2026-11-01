package ipgw

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/UnbalancedCat/ipgw-meta/internal/castest"
	"html"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// All traffic terminates in the fake RoundTripper below. Generated RSA and
// synthetic state are used only in memory; no profile or real server is read.
// A valid synthetic submission deliberately receives 500: a matching payload
// and an HTTP failure can coexist, so HTTP 500 alone cannot identify a key bug.
func TestCASSubmitContextSessionAndForm(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	encoded := base64.StdEncoding.EncodeToString(der)
	for _, tc := range []struct {
		name, action, mutation string
		redirect               bool
	}{
		{"relative_action", "relative", "rotate", false},
		{"absolute_action", "absolute", "rotate", false},
		{"missing_action", "missing", "rotate", false},
		{"query_action", "query", "rotate", false},
		{"redirected_page_relative_action", "relative", "rotate", true},
		{"cookie_deletion", "relative", "delete", false},
		{"cookie_path_restriction", "relative", "path", false},
		{"cookie_domain_rejected", "relative", "domain", false},
		{"cookie_default_path", "relative", "default_path", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const user, password = "synthetic-user", "synthetic-密碼-🙂"
			const initialCookie, rotatedCookie = "synthetic_session=initial", "synthetic_session=rotated"
			pagePath := "/tpass/login"
			if tc.redirect {
				pagePath = "/tpass/view/login"
			}
			directory := pagePath[:strings.LastIndex(pagePath, "/")+1]
			round, providerCalls, posts, scripts, requests := 0, 0, 0, 0, 0
			var expectedAction, expectedQuery, expectedLT, expectedExecution string
			expectCookie := func(req *http.Request, want string) {
				t.Helper()
				if req.Header.Get("Cookie") != want {
					t.Fatal("synthetic session scope/continuity mismatch")
				}
			}
			postCookie := rotatedCookie
			if tc.mutation == "delete" {
				postCookie = ""
			} else if tc.mutation == "path" || tc.mutation == "domain" {
				postCookie = initialCookie
			}
			fake := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				requests++
				if req.URL.Scheme == "http" {
					if req.URL.Host != "ipgw.neu.edu.cn" || req.URL.Path != "/" || req.Method != "GET" {
						t.Fatal("unexpected HTTP path; no fake ticket may be sent")
					}
					expectCookie(req, "")
					res := testResponse(req, 200, "ac_id=1", "")
					res.Header.Add("Set-Cookie", "synthetic_discovery=discard; Path=/")
					return res, nil
				}
				if req.URL.Host == "ipgw.neu.edu.cn" && req.URL.Path == "/cgi-bin/rad_user_info" && req.Method == "GET" {
					round++
					expectCookie(req, "")
					return testResponse(req, 200, "not_online_error", ""), nil
				}
				if req.URL.Host != "pass.neu.edu.cn" || req.URL.Scheme != "https" {
					t.Fatal("unexpected host/scheme or activation after 500")
				}
				if req.Method == "GET" && req.URL.Path == "/tpass/login" {
					// Each Login must create a fresh CAS session, including after 500.
					expectCookie(req, "")
					if tc.redirect {
						res := testResponse(req, 302, "", pagePath+"?"+req.URL.RawQuery)
						res.Header.Add("Set-Cookie", "synthetic_bootstrap=seen; Path=/tpass; Secure")
						return res, nil
					}
				}
				if req.Method == "GET" && req.URL.Path == pagePath {
					if tc.redirect {
						expectCookie(req, "synthetic_bootstrap=seen")
					}
					expectedQuery = req.URL.RawQuery
					expectedAction = directory + "submit"
					action := "submit?" + expectedQuery
					switch tc.action {
					case "absolute":
						action = "https://pass.neu.edu.cn" + expectedAction + "?" + expectedQuery
					case "query":
						action = "?" + expectedQuery
						expectedAction = pagePath
					case "missing":
						expectedAction = pagePath
					}
					attr := ` action="` + html.EscapeString(action) + `"`
					if tc.action == "missing" {
						attr = ""
					}
					expectedLT = fmt.Sprintf("synthetic-state-%d", round)
					expectedExecution = fmt.Sprintf("synthetic-execution-%d", round)
					body := `<form action="/decoy"><input name="lt" value="decoy"><input name="execution" value="decoy"></form>` +
						`<form id="loginForm"` + attr + `><input name="lt" value="` + expectedLT + `">` +
						`<input name="execution" value="` + expectedExecution + `"><input name="rsa" value="decoy">` +
						`<input name="unknown" value="decoy"></form>` +
						`<script src="login-state.js"></script><script src="login-key.js"></script>`
					res := testResponse(req, 200, body, "")
					res.Header.Add("Set-Cookie", initialCookie+"; Path=/tpass; Secure; HttpOnly")
					res.Header.Add("Set-Cookie", "synthetic_bootstrap=; Path=/tpass; Max-Age=0; Secure")
					return res, nil
				}
				if req.Method == "GET" && req.URL.Path == directory+"login-state.js" {
					scripts++
					expectCookie(req, initialCookie)
					res := testResponse(req, 200, "/* synthetic state rotation; no key */", "")
					value := rotatedCookie + "; Path=/tpass; Secure"
					switch tc.mutation {
					case "delete":
						value = "synthetic_session=; Path=/tpass; Max-Age=0; Secure"
					case "path":
						value = "synthetic_path=excluded; Path=/unrelated; Secure"
					case "domain":
						value = "synthetic_domain=excluded; Domain=outside.invalid; Path=/tpass; Secure"
					case "default_path":
						value = rotatedCookie + "; Secure"
					}
					res.Header.Add("Set-Cookie", value)
					return res, nil
				}
				if req.Method == "GET" && req.URL.Path == directory+"login-key.js" {
					scripts++
					expectCookie(req, postCookie)
					return testResponse(req, 200, castest.LoginScript(encoded), ""), nil
				}
				if req.Method == "POST" && req.URL.Path == expectedAction {
					posts++
					expectCookie(req, postCookie)
					if req.URL.RawQuery != expectedQuery || req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
						t.Fatal("action/service or form encoding changed")
					}
					if err := req.ParseForm(); err != nil {
						t.Fatal("synthetic form parse failed")
					}
					want := url.Values{"lt": {expectedLT}, "execution": {expectedExecution}, "_eventId": {"submit"}, "ul": {"14"}, "pl": {"15"}}
					if len(req.PostForm) != 6 {
						t.Fatal("form is not the six-field closed set")
					}
					for name, values := range want {
						if got := req.PostForm[name]; len(got) != 1 || got[0] != values[0] {
							t.Fatal("synthetic form scope/state/length mismatch")
						}
					}
					if len(req.PostForm["rsa"]) != 1 {
						t.Fatal("RSA field cardinality mismatch")
					}
					cipher, err := base64.StdEncoding.DecodeString(req.PostForm.Get("rsa"))
					if err != nil {
						t.Fatal("synthetic ciphertext encoding failed")
					}
					clear, err := rsa.DecryptPKCS1v15(rand.Reader, key, cipher)
					if err != nil || string(clear) != user+password {
						t.Fatal("synthetic RSA input mismatch")
					}
					return testResponse(req, 500, "synthetic-server-error", ""), nil
				}
				t.Fatal("unexpected synthetic request or retry")
				return nil, fmt.Errorf("unreachable synthetic request")
			})
			client, err := NewClient(WithRoundTripper(fake))
			if err != nil {
				t.Fatal(err)
			}
			for attempt := 1; attempt <= 2; attempt++ {
				_, err := client.Login(context.Background(), LoginRequest{Method: AuthMethodPassword, ExpectedUsername: user,
					Credentials: CredentialProviderFunc(func(context.Context, CredentialRequest) (Credential, error) {
						providerCalls++
						if scripts != 2*providerCalls {
							t.Fatal("provider reached before script discovery")
						}
						return Credential{Password: password}, nil
					})})
				if !IsCode(err, CodeNetwork) || providerCalls != attempt || posts != attempt || round != attempt {
					t.Fatal("unexpected result, retry or credential count")
				}
				for _, canary := range []string{user, password, expectedLT, expectedExecution, encoded, "synthetic-server-error"} {
					assertErrorDoesNotContain(t, err, canary)
				}
			}
			wantRequests := 12
			if tc.redirect {
				wantRequests = 14
			}
			if requests != wantRequests {
				t.Fatal("unexpected request total; no activation or retry is allowed")
			}
		})
	}
}

func TestCASSubmitContextRejectsBeforeProvider(t *testing.T) {
	for _, tc := range []struct{ name, action, inputs string }{
		{"missing_execution", "/tpass/submit", `<input name="lt" value="synthetic-state">`},
		{"empty_lt", "/tpass/submit", `<input name="lt" value=""><input name="execution" value="synthetic-execution">`},
		{"duplicate_execution", "/tpass/submit", `<input name="lt" value="synthetic-state"><input name="execution" value="one"><input name="execution" value="two">`},
		{"http_action", "http://pass.neu.edu.cn/tpass/submit", `<input name="lt" value="synthetic-state"><input name="execution" value="synthetic-execution">`},
		{"cross_origin_action", "https://outside.invalid/tpass/submit", `<input name="lt" value="synthetic-state"><input name="execution" value="synthetic-execution">`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			providerCalls, calls := 0, 0
			fake := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				calls++
				if req.Method != "GET" {
					t.Fatal("invalid form reached POST")
				}
				switch req.URL.Host + req.URL.Path {
				case "ipgw.neu.edu.cn/cgi-bin/rad_user_info":
					return testResponse(req, 200, "not_online_error", ""), nil
				case "ipgw.neu.edu.cn/":
					return testResponse(req, 200, "ac_id=1", ""), nil
				case "pass.neu.edu.cn/tpass/login":
					// Lexical candidate only: these guards precede key decoding.
					body := `<form id="loginForm" action="` + html.EscapeString(tc.action) + `">` + tc.inputs + `</form><script>var publicKeyStr="` + strings.Repeat("A", 160) + `";</script>`
					return testResponse(req, 200, body, ""), nil
				default:
					t.Fatal("unexpected synthetic request")
					return nil, fmt.Errorf("unreachable synthetic request")
				}
			})
			client, err := NewClient(WithRoundTripper(fake))
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.Login(context.Background(), LoginRequest{Method: AuthMethodPassword, ExpectedUsername: "synthetic-user", Credentials: CredentialProviderFunc(func(context.Context, CredentialRequest) (Credential, error) {
				providerCalls++
				return Credential{Password: "synthetic-password"}, nil
			})})
			if !IsCode(err, CodeProtocolChanged) || providerCalls != 0 || calls != 3 {
				t.Fatal("invalid dynamic form did not stop before credentials")
			}
		})
	}
}
