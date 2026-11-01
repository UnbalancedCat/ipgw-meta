// cas-collect has only a built-in synthetic replay self-test.
package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
	"io"
	"os"
	"time"
)

func selfTest() (cascollect.Report, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return cascollect.Report{}, cascollect.ErrReport
	}
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return cascollect.Report{}, cascollect.ErrReport
	}
	body := []byte("publicKeyStr='" + base64.StdEncoding.EncodeToString(der) + "';")
	inlineKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return cascollect.Report{}, cascollect.ErrReport
	}
	inlineDER, err := x509.MarshalPKIXPublicKey(&inlineKey.PublicKey)
	if err != nil {
		return cascollect.Report{}, cascollect.ErrReport
	}
	page := "<script>publicKeyStr='" + base64.StdEncoding.EncodeToString(inlineDER) + "';</script>" + `<script src="/first.js"></script><script src="/first.js"></script>`
	const login = "https://cas.invalid/login?service=http%3A%2F%2Fgateway.invalid%2Fservice%3Fac_id%3D7"
	s := cascollect.Scenario{Scope: cascollect.Scope{Captive: "http://gateway.invalid/", Service: "http://gateway.invalid/service", CAS: "https://cas.invalid/login"}, Exchanges: []cascollect.Exchange{
		{Target: "http://gateway.invalid/", Status: 200, Body: []byte("ac_id=7")},
		{Target: login, Status: 200, Headers: []cascollect.Header{{Name: "Content-Type", Value: "text/html"}, {Name: "Set-Cookie", Value: "synthetic-cookie-canary"}}, Body: []byte(page)},
		{Target: "https://cas.invalid/first.js", Status: 200, Headers: []cascollect.Header{{Name: "Content-Type", Value: "application/javascript"}}, Body: body},
	}}
	r := cascollect.Replay(ctx, s)
	if r.Collection != "complete" || r.Simulated.Requests != 3 || len(r.References) != 2 || len(r.Resources) != 1 || r.Analysis == nil || len(r.Analysis.Keys) != 2 {
		return cascollect.Report{}, cascollect.ErrReport
	}
	s.Exchanges[1].Body = []byte(`<script src="/first.js?synthetic=7"></script>`)
	rejected := cascollect.Replay(ctx, s)
	if rejected.Collection != "blocked" || rejected.Reason != "url_rejected" || rejected.Simulated.Requests != 2 || rejected.Analysis != nil {
		return cascollect.Report{}, cascollect.ErrReport
	}
	s.Exchanges[1].Body = []byte(`<script src="/first.js"></script>`)
	s.Exchanges[2].Fault = "dns"
	rejected = cascollect.Replay(ctx, s)
	if rejected.Reason != "dns_failed" || rejected.Simulated.Connects != 2 || rejected.Simulated.Requests != 3 {
		return cascollect.Report{}, cascollect.ErrReport
	}
	if _, err := cascollect.EncodeReport(rejected); err != nil {
		return cascollect.Report{}, err
	}
	return r, nil
}
func run(args []string, out, diagnostic io.Writer) int {
	if len(args) != 1 || args[0] != "--self-test" {
		fmt.Fprintln(diagnostic, "only --self-test is supported")
		return 2
	}
	r, err := selfTest()
	if err != nil {
		fmt.Fprintln(diagnostic, "synthetic collection self-test failed")
		return 1
	}
	b, err := cascollect.EncodeReport(r)
	if err != nil {
		fmt.Fprintln(diagnostic, "synthetic collection report rejected")
		return 1
	}
	if _, err = out.Write(b); err != nil {
		return 1
	}
	return 0
}
func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
