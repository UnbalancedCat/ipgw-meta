package authcontext

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
	"github.com/UnbalancedCat/ipgw-meta/internal/castest"
	"net/http"
	"strings"
)

func SelfTestStatic() StaticReport {
	out := NewStaticReport("self-test")
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
	r := AnonymousStatic(context.Background(), tr, func(context.Context) bool { return true }, casprovenance.AnalyzeStaticParserMatch)
	data, err := EncodeStatic(r)
	if err != nil || calls != 3 || r.State != "observed" || r.Syntax.ParserMatch.State != "mapped" || r.Context.Form.FinalTarget != "same" {
		return out
	}
	if _, err = DecodeStatic(data); err != nil || strings.Contains(string(data), encoded) || strings.Contains(string(data), "synthetic-state") {
		return out
	}
	out.State, out.Reason = "pass", "none"
	out.Context.State, out.Context.Reason = "pass", "none"
	out.Context.SDK.State, out.Context.SDK.Reason = "pass", "none"
	return out
}
