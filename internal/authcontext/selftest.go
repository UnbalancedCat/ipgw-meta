package authcontext

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"github.com/UnbalancedCat/ipgw-meta/internal/castest"
	"io"
	"net/http"
	"strings"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func syntheticResponse(r *http.Request, status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: r}
}
func SelfTest() Report {
	out := NewReport("self-test")
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return out
	}
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return out
	}
	encoded := base64.StdEncoding.EncodeToString(der)
	calls := 0
	tr := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		if req.Method != "GET" {
			return nil, ErrReport
		}
		switch req.URL.Host + req.URL.Path {
		case "ipgw.neu.edu.cn/cgi-bin/rad_user_info":
			return syntheticResponse(req, 200, "not_online_error"), nil
		case "ipgw.neu.edu.cn/":
			return syntheticResponse(req, 200, "ac_id=1"), nil
		case "pass.neu.edu.cn/tpass/login":
			return syntheticResponse(req, 200, `<form id="loginForm"><input name="lt" value="synthetic-state"><input name="execution" value="synthetic-execution"></form><script>`+castest.LoginScript(encoded)+`</script>`), nil
		default:
			return nil, ErrReport
		}
	})
	r := Anonymous(context.Background(), tr, func(context.Context) bool { return true })
	data, err := Encode(r)
	if err != nil || calls != 3 || r.State != "observed" || r.ParserMatch.State != "mapped" || r.Form.FinalTarget != "same" {
		return out
	}
	if _, err = Decode(data); err != nil || strings.Contains(string(data), encoded) || strings.Contains(string(data), "synthetic-state") {
		return out
	}
	out.State, out.Reason = "pass", "none"
	out.SDK.State, out.SDK.Reason = "pass", "none"
	return out
}
