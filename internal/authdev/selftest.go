package authdev

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

// SelfTest uses newly generated synthetic RSA material and an in-memory
// transport. It never queries topology, profiles, certificate stores or DNS.
func SelfTest() Report {
	r := NewReport("self-test")
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return r
	}
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return r
	}
	public := base64.StdEncoding.EncodeToString(der)
	count := 0
	fake := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		count++
		body := ""
		headers := make(http.Header)
		switch count {
		case 1:
			if req.URL.Path != "/cgi-bin/rad_user_info" {
				return nil, errStop
			}
			body = "not_online_error"
		case 2:
			if req.URL.Hostname() != "ipgw.neu.edu.cn" || req.Header.Get("Cookie") != "" {
				return nil, errStop
			}
			body = "ac_id=1"
		case 3:
			if req.URL.Query().Get("service") == "" {
				return nil, errStop
			}
			body = `<html><form id="loginForm"><input name="lt" value="synthetic-lt"><input name="execution" value="synthetic-execution"></form><script src="/tpass/login.js"></script></html>`
			headers.Set("Set-Cookie", "synthetic_session=memory_only; Secure; Path=/tpass")
		case 4:
			if req.Header.Get("Cookie") != "synthetic_session=memory_only" {
				return nil, errStop
			}
			body = castest.LoginScript(public)
		default:
			return nil, errStop
		}
		return &http.Response{StatusCode: 200, Header: headers, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
	})
	observed := Anonymous(context.Background(), fake, func(context.Context) bool { return true })
	if observed.State == "observed" && observed.Requests == 4 && count == 4 && observed.KeySource == "script" && observed.CookieReceived && observed.CookieSent && observed.ActionService == "same" && observed.Validate() == nil {
		r.State, r.Reason = "pass", "none"
	}
	return r
}
