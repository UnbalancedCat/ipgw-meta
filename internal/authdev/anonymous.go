// Package authdev implements non-release development validation. Product
// commands must never import this package. See REL-DEV-ANON-001.
package authdev

import (
	"context"
	"errors"
	ipgw "github.com/UnbalancedCat/ipgw-meta"
	"github.com/UnbalancedCat/ipgw-meta/internal/cas"
	"github.com/UnbalancedCat/ipgw-meta/internal/caslegacy"
	"github.com/UnbalancedCat/ipgw-meta/internal/srun"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const MaxRequests = 32
const bodyLimit = 2 << 20
const anonymousIdentity = "ipgw-dev-anonymous-sentinel"

var errStop = errors.New("authdev: stopped")

// Anonymous uses the SDK and a provider which can never return a credential.
// next is fake in tests; native execution supplies the source-bound product
// transport. check is mandatory before every application GET.
func Anonymous(ctx context.Context, next http.RoundTripper, check func(context.Context) bool) Report {
	r := NewReport("anonymous")
	if ctx == nil || next == nil || check == nil {
		return r
	}
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	g := &guard{next: next, check: check, report: &r}
	client, err := ipgw.NewClient(ipgw.WithRoundTripper(g))
	if err != nil {
		return r
	}
	reached := false
	_, err = client.Login(ctx, ipgw.LoginRequest{
		Method: ipgw.AuthMethodPassword, ExpectedUsername: anonymousIdentity, Switch: ipgw.SwitchRefuse,
		Credentials: ipgw.CredentialProviderFunc(func(ctx context.Context, _ ipgw.CredentialRequest) (ipgw.Credential, error) {
			if ctx.Err() == nil && g.reason == "" {
				if g.check(ctx) {
					reached = true
				} else {
					g.reason = "topology_changed"
				}
			}
			return ipgw.Credential{}, errStop
		}),
	})
	switch {
	case ctx.Err() != nil:
		r.Reason = "canceled"
	case g.reason != "":
		r.Reason = g.reason
	case reached:
		r.State, r.Reason = "observed", "credential_boundary_reached"
	default:
		r.Reason = "sdk_rejected"
		if err != nil {
			r.SDKCode = string(ipgw.CodeOf(err))
		}
	}
	return r
}

type guard struct {
	next     http.RoundTripper
	check    func(context.Context) bool
	report   *Report
	reason   string
	service  string // ephemeral, never serialized
	pageSeen bool
}

func (g *guard) stop(reason string) (*http.Response, error) {
	if g.reason == "" {
		g.reason = reason
	}
	return nil, errStop
}
func (g *guard) RoundTrip(req *http.Request) (*http.Response, error) {
	if g.reason != "" {
		return nil, errStop
	}
	if req == nil || req.URL == nil || req.Context().Err() != nil {
		return g.stop("canceled")
	}
	u := req.URL
	if req.Method != http.MethodGet || req.Body != nil && req.Body != http.NoBody {
		return g.stop("request_rejected")
	}
	if u.User != nil || u.Fragment != "" || u.Opaque != "" || u.Port() != "" && u.Port() != map[string]string{"http": "80", "https": "443"}[u.Scheme] {
		return g.stop("request_rejected")
	}
	query, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return g.stop("request_rejected")
	}
	for key := range query {
		if strings.EqualFold(key, "ticket") {
			return g.stop("early_ticket")
		}
	}
	role := ""
	switch strings.ToLower(u.Hostname()) {
	case "ipgw.neu.edu.cn":
		if u.Path == "/v1/srun_portal_sso" {
			return g.stop("early_ticket")
		}
		if u.Scheme == "https" && u.Path == "/cgi-bin/rad_user_info" && u.RawQuery == "" && g.report.Requests == 0 {
			role = "status"
		} else if (u.Scheme == "http" || u.Scheme == "https") && g.report.StatusRequests == 1 && g.report.CASRequests == 0 && g.report.DiscoveryRequests < 2 && !strings.HasPrefix(path.Clean(u.Path), "/cgi-bin") && !strings.HasPrefix(path.Clean(u.Path), "/v1") {
			role = "discovery"
		}
		if req.Header.Get("Cookie") != "" {
			return g.stop("request_rejected")
		}
	case "pass.neu.edu.cn":
		if strings.Contains(strings.ToLower(path.Clean(u.Path)), "logout") {
			return g.stop("request_rejected")
		}
		if u.Scheme == "https" && g.report.DiscoveryRequests > 0 {
			role = "cas"
		}
	}
	if role == "" || req.Header.Get("Authorization") != "" {
		return g.stop("request_rejected")
	}
	if g.report.Requests >= MaxRequests {
		return g.stop("request_budget")
	}
	if !g.check(req.Context()) {
		return g.stop("topology_changed")
	}
	g.report.Requests++
	switch role {
	case "status":
		g.report.StatusRequests++
	case "discovery":
		g.report.DiscoveryRequests++
	case "cas":
		g.report.CASRequests++
		if g.report.CASRequests == 1 {
			g.service = query.Get("service")
		}
		if req.Header.Get("Cookie") != "" {
			g.report.CookieSent = true
		}
	}
	response, err := g.next.RoundTrip(req)
	if err != nil || response == nil {
		return response, err
	}
	if role == "cas" && len(response.Cookies()) > 0 {
		g.report.CookieReceived = true
	}
	if response.Body != nil {
		base := *u
		response.Body = &observedBody{ReadCloser: response.Body, complete: func(data []byte) { g.observe(role, &base, data) }}
	}
	return response, nil
}
func (g *guard) observe(role string, base *url.URL, data []byte) {
	if role == "status" {
		if status, err := srun.ParseStatus(data); err == nil && status.Online {
			g.reason = "initial_online"
		}
		return
	}
	if role != "cas" {
		return
	}
	if !g.pageSeen {
		page, err := cas.ParsePage(base, data)
		if err != nil || page.Action == nil {
			return
		}
		g.pageSeen = true
		g.report.PageObserved = true
		if caslegacy.ExtractPublicKey(data) != "" {
			g.report.KeySource = "page"
		}
		service, err := url.ParseQuery(page.Action.RawQuery)
		g.report.ActionService = "different"
		if err == nil {
			values := service["service"]
			if len(values) == 0 {
				g.report.ActionService = "absent"
			} else if len(values) == 1 && values[0] == g.service && g.service != "" {
				g.report.ActionService = "same"
			}
		}
	} else if g.report.KeySource == "none" && caslegacy.ExtractPublicKey(data) != "" {
		g.report.KeySource = "script"
	}
}

// Passive tee: preserve n/error, do not drain on Close, only observe complete
// bounded bodies. The SDK consumes bodies synchronously within this call.
type observedBody struct {
	io.ReadCloser
	data     []byte
	done     bool
	complete func([]byte)
}

func (b *observedBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if !b.done {
		if n > bodyLimit-len(b.data) {
			b.done = true
			b.data = nil
		} else {
			b.data = append(b.data, p[:n]...)
		}
		if !b.done && err != nil {
			if err == io.EOF {
				b.complete(b.data)
			}
			b.done, b.data = true, nil
		}
	}
	return n, err
}
func (b *observedBody) Close() error { b.data = nil; b.done = true; return b.ReadCloser.Close() }
