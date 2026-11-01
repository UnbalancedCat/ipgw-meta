package ipgw

import (
	"context"
	"net/http"
	"testing"
)

func TestLogoutExplicitCallbackAndStateVerification(t *testing.T) {
	const username = "synthetic.user@realm"
	const successBody = `ipgw_logout({"ecode":0,"error":"ok","res":"ok"})`
	for _, tc := range []struct {
		name         string
		body         string
		httpStatus   int
		offlineAfter bool
		wantCode     ErrorCode
		wantStatus   int
	}{
		{"jsonp_offline", successBody, 200, true, "", 2},
		{"bare_json_compatibility", `{"ecode":0}`, 200, true, "", 2},
		{"business_rejected", `ipgw_logout({"ecode":1})`, 200, true, CodeAuthentication, 1},
		{"truncated", `ipgw_logout({"ecode":0}`, 200, true, CodeProtocolChanged, 1},
		{"trailing", successBody + "extra", 200, true, CodeProtocolChanged, 1},
		{"duplicate", `ipgw_logout({"ecode":0,"ecode":0})`, 200, true, CodeProtocolChanged, 1},
		{"conflict", `ipgw_logout({"ecode":0,"success":false})`, 200, true, CodeProtocolChanged, 1},
		{"http_failure", successBody, 500, true, CodeNetwork, 1},
		{"business_ok_still_online", successBody, 200, false, CodeAuthentication, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			online := true
			statusCalls, logoutCalls := 0, 0
			client, err := NewClient(WithRoundTripper(roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodGet || req.URL.Scheme != "https" || (req.Body != nil && req.Body != http.NoBody) {
					t.Fatal("logout flow changed method, scheme or body")
				}
				if req.Header.Get("Cookie") != "" || req.Header.Get("Referer") != "" {
					t.Fatal("browser session material was added")
				}
				switch req.URL.Path {
				case "/cgi-bin/rad_user_info":
					statusCalls++
					if req.URL.Query().Has("callback") {
						t.Fatal("logout callback leaked into status request")
					}
					if online {
						return testResponse(req, 200, `{"error":"ok","user_name":"`+username+`","online_ip":"192.0.2.20"}`, ""), nil
					}
					return testResponse(req, 200, "not_online_error", ""), nil
				case "/cgi-bin/srun_portal":
					logoutCalls++
					query := req.URL.Query()
					if len(query) != 3 || query.Get("action") != "logout" || query.Get("username") != username || query.Get("callback") != "ipgw_logout" {
						t.Fatal("logout must preserve encoded username and add only the fixed callback")
					}
					for _, values := range query {
						if len(values) != 1 {
							t.Fatal("duplicate query values")
						}
					}
					online = !tc.offlineAfter
					return testResponse(req, tc.httpStatus, tc.body, ""), nil
				default:
					t.Fatal("unexpected discovery, CAS or additional request")
					return nil, nil
				}
			})))
			if err != nil {
				t.Fatal(err)
			}
			client.verifyDelay = 0
			before := client.endpoints.logout.String()
			result, err := client.Logout(context.Background())
			if CodeOf(err) != tc.wantCode || statusCalls != tc.wantStatus || logoutCalls != 1 {
				t.Fatalf("code=%q, status calls=%d, logout calls=%d", CodeOf(err), statusCalls, logoutCalls)
			}
			if client.endpoints.logout.String() != before {
				t.Fatal("shared endpoint was mutated")
			}
			if tc.wantCode == "" {
				if result.Outcome != LogoutLoggedOut || result.Status.Session != SessionOffline {
					t.Fatal("business and final offline must both be established")
				}
				second, err := client.Logout(context.Background())
				if err != nil || second.Outcome != LogoutAlreadyOffline || logoutCalls != 1 || statusCalls != tc.wantStatus+1 {
					t.Fatal("offline logout must not send another logout request")
				}
			} else if result.Outcome != "" {
				t.Fatal("unconfirmed logout returned a successful outcome")
			}
		})
	}
}

func TestLogoutInitiallyOfflineDoesNotRequestCallback(t *testing.T) {
	calls := 0
	client, err := NewClient(WithRoundTripper(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		if req.URL.Path != "/cgi-bin/rad_user_info" || req.URL.Query().Has("callback") {
			t.Fatal("initially offline session must only query status")
		}
		return testResponse(req, 200, "not_online_error", ""), nil
	})))
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Logout(context.Background())
	if err != nil || result.Outcome != LogoutAlreadyOffline || calls != 1 {
		t.Fatal("initial offline idempotence changed")
	}
}

// Exercise encoding below status validation; this does not extend the public
// status username grammar to accept '+' or '&'.
func TestLogoutCallbackPreservesQueryEncoding(t *testing.T) {
	const username = "synthetic+user&scope=one"
	calls := 0
	client, err := NewClient(WithRoundTripper(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		switch req.URL.Path {
		case "/cgi-bin/srun_portal":
			q := req.URL.Query()
			if len(q) != 3 || len(q["username"]) != 1 || q.Get("username") != username || q.Get("callback") != "ipgw_logout" || q.Get("action") != "logout" {
				t.Fatal("synthetic username changed the query shape")
			}
			return testResponse(req, 200, `ipgw_logout({"ecode":0})`, ""), nil
		case "/cgi-bin/rad_user_info":
			return testResponse(req, 200, "not_online_error", ""), nil
		default:
			t.Fatal("unexpected request")
			return nil, nil
		}
	})))
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.logoutOnline(context.Background(), client.newHTTPClient(nil), username)
	if err != nil || result.Outcome != LogoutLoggedOut || calls != 2 {
		t.Fatal("encoded callback flow did not verify offline")
	}
}
