package ipgw

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

// These synthetic alternatives explain the same bounded observation; neither
// is asserted to be the real server response from the maintainer's live run.
func TestLogoutResponseFailureThenAlreadyOffline(t *testing.T) {
	for _, tc := range []struct{ name, body, message string }{
		{"unrecognized_business", `{"ecode":"0"}`, "gateway logout response is not recognized"},
		{"response_limit", strings.Repeat(" ", int(apiResponseLimit)+1), "gateway response exceeded the safe limit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			online := true
			statusCalls, logoutCalls := 0, 0
			tr := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				switch req.URL.Path {
				case "/cgi-bin/rad_user_info":
					statusCalls++
					if online {
						return testResponse(req, 200, `{"error":"ok","user_name":"synthetic-user","online_ip":"192.0.2.8"}`, ""), nil
					}
					return testResponse(req, 200, "not_online_error", ""), nil
				case "/cgi-bin/srun_portal":
					logoutCalls++
					online = false
					return testResponse(req, 200, tc.body, ""), nil
				default:
					t.Fatal("unexpected request")
					return nil, nil
				}
			})
			client, err := NewClient(WithRoundTripper(tr))
			if err != nil {
				t.Fatal(err)
			}
			first, err := client.Logout(context.Background())
			if !IsCode(err, CodeProtocolChanged) || first.Outcome != "" || statusCalls != 1 || logoutCalls != 1 {
				t.Fatal("unconfirmed business response promoted to success")
			}
			typed, ok := err.(*Error)
			if !ok || typed.Message != tc.message {
				t.Fatal("synthetic failure path not distinguished internally")
			}
			cleanup, err := client.Logout(context.Background())
			if err != nil || cleanup.Outcome != LogoutAlreadyOffline || cleanup.Status.Session != SessionOffline || statusCalls != 2 || logoutCalls != 1 {
				t.Fatal("cleanup must use offline idempotence without another logout request")
			}
		})
	}
}
