package authobserve

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	ipgw "github.com/UnbalancedCat/ipgw-meta"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// SelfTest uses an in-memory response only. There is no native transport,
// profile, certificate source, topology read or credential provider.
func SelfTest() bool {
	c := New("login")
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "https://pass.neu.edu.cn/tpass/login", strings.NewReader("synthetic-private-body"))
	response := &http.Response{StatusCode: 500, Body: io.NopCloser(strings.NewReader("synthetic-private-response"))}
	calls := 0
	next := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r != req {
			return nil, ErrProtocol
		}
		return response, nil
	})
	got, err := c.Decorate(next).RoundTrip(req)
	if got != response || err != nil || calls != 1 {
		return false
	}
	c.result(&ipgw.Error{Code: ipgw.CodeNetwork, Message: "CAS returned an unexpected HTTP status"})
	report := c.Snapshot()
	data, err := json.Marshal(report)
	if err != nil || report.Validate() != nil || len(report.Requests) != 1 || report.Requests[0].Role != "cas_post" || report.Requests[0].StatusCode != 500 || report.SDKCategory != "cas_http_status" || strings.Contains(string(data), "synthetic-private") {
		return false
	}
	child := Child{SchemaVersion: 1, ToolKind: "auth-observed-cli", Command: "login", ExitCode: 3, Observation: report, Product: json.RawMessage(`{"schema_version":1,"command":"login","ok":false,"error":{"code":"network","message":"network request failed","retryable":true,"details":{}}}`)}
	encoded, err := json.Marshal(child)
	if err != nil {
		return false
	}
	decoded, err := DecodeChild(encoded, "login", 3)
	return err == nil && decoded.Observation.SDKCategory == "cas_http_status" && logoutSelfTest()
}

func logoutSelfTest() bool {
	c := New("logout")
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://ipgw.neu.edu.cn/cgi-bin/srun_portal", nil)
	response, err := c.Decorate(roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"ecode":"0","success":true}`))}, nil
	})).RoundTrip(req)
	if err != nil {
		return false
	}
	_, err = io.ReadAll(response.Body)
	closed := response.Body.Close()
	c.result(&ipgw.Error{Code: ipgw.CodeProtocolChanged, Message: "gateway logout response is not recognized"})
	o := c.Snapshot()
	if err != nil || closed != nil || o.ValidateVersion(2) != nil || len(o.Logout.Responses) != 1 || o.Logout.Responses[0].Parser != "rejected" || o.Logout.Responses[0].Ecode != "string_zero" {
		return false
	}
	child := Child{SchemaVersion: 2, ToolKind: "auth-observed-cli", Command: "logout", ExitCode: 6, Observation: o, Product: json.RawMessage(`{"schema_version":1,"command":"logout","ok":false,"error":{"code":"protocol_changed","message":"fixed","retryable":false,"details":{}}}`)}
	data, err := json.Marshal(child)
	if err != nil {
		return false
	}
	_, err = DecodeChild(data, "logout", 6)
	return err == nil
}
