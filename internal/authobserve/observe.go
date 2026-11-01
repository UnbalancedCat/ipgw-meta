// Package authobserve is only linked by independent development entrypoints.
// It never serializes request/response bodies, headers, queries or error text.
package authobserve

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"

	ipgw "github.com/UnbalancedCat/ipgw-meta"
	"github.com/UnbalancedCat/ipgw-meta/internal/app"
	"github.com/UnbalancedCat/ipgw-meta/internal/cli"
)

const MaxRequests = 32

type Request struct {
	Role             string `json:"role"`
	StatusCode       int    `json:"status_code"`
	ResponseReceived bool   `json:"response_received"`
	Transport        string `json:"transport"`
}
type Observation struct {
	Command     string             `json:"command"`
	State       string             `json:"state"`
	SDKCode     string             `json:"sdk_code"`
	SDKCategory string             `json:"sdk_category"`
	Requests    []Request          `json:"requests"`
	Logout      *LogoutObservation `json:"logout,omitempty"`
}
type Collector struct {
	mu           sync.Mutex
	data         Observation
	pending      int
	logoutBytes  int
	logoutBodies []*logoutBody
}

func New(command string) *Collector {
	c := &Collector{data: Observation{Command: command, State: "complete", SDKCategory: "not_called", Requests: []Request{}}}
	if command == "logout" {
		c.data.Logout = &LogoutObservation{State: "not_observed", Responses: []LogoutResponse{}}
	}
	return c
}
func (c *Collector) Snapshot() Observation {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := c.data
	out.Requests = append([]Request{}, c.data.Requests...)
	if c.data.Logout != nil {
		v := *c.data.Logout
		v.Responses = append([]LogoutResponse{}, v.Responses...)
		sort.Slice(v.Responses, func(i, j int) bool { return v.Responses[i].RequestIndex < v.Responses[j].RequestIndex })
		out.Logout = &v
	}
	if c.pending != 0 {
		out.State = "incomplete"
		out.Requests = []Request{}
	}
	if out.State == "incomplete" && out.Logout != nil {
		out.Logout = &LogoutObservation{State: "incomplete", Responses: []LogoutResponse{}}
	}
	return out
}

type observedTransport struct {
	next http.RoundTripper
	c    *Collector
}

func (c *Collector) Decorate(next http.RoundTripper) http.RoundTripper {
	return &observedTransport{next: next, c: c}
}
func (t *observedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	c := t.c
	requestRole := role(req)
	// Reserve before the call, so order is call order even with concurrent use.
	c.mu.Lock()
	index := -1
	if c.data.State == "complete" {
		if len(c.data.Requests) == MaxRequests {
			c.data.State = "incomplete"
			c.data.Requests = []Request{}
		} else {
			index = len(c.data.Requests)
			c.data.Requests = append(c.data.Requests, Request{Role: requestRole, Transport: "none"})
		}
	}
	c.pending++
	c.mu.Unlock()
	response, err := t.next.RoundTrip(req)
	entry := Request{Role: requestRole, Transport: transportCategory(err)}
	if response != nil {
		entry.ResponseReceived = true
		if response.StatusCode >= 100 && response.StatusCode <= 599 {
			entry.StatusCode = response.StatusCode
		}
	}
	c.mu.Lock()
	c.pending--
	if index >= 0 && c.data.State == "complete" {
		c.data.Requests[index] = entry
	}
	if err == nil && response != nil && response.Body != nil && req != nil && req.Method == http.MethodGet && requestRole == "logout" && c.data.Command == "logout" {
		response.Body = c.wrapLogout(req.Context(), response.Body, index)
	}
	var discard []*logoutBody
	if c.data.Logout != nil && (c.data.State != "complete" || c.data.Logout.State == "budget_exceeded") {
		discard = append(discard, c.logoutBodies...)
	}
	c.mu.Unlock()
	for _, b := range discard {
		b.discard()
	}
	return response, err
}
func role(r *http.Request) string {
	if r == nil || r.URL == nil {
		return "other"
	}
	u := r.URL
	if strings.EqualFold(u.Hostname(), "pass.neu.edu.cn") && u.Scheme == "https" && (u.Port() == "" || u.Port() == "443") {
		if r.Method == http.MethodPost {
			return "cas_post"
		}
		if r.Method == http.MethodGet {
			return "cas_get"
		}
	}
	if !strings.EqualFold(u.Hostname(), "ipgw.neu.edu.cn") {
		return "other"
	}
	if u.Scheme == "http" && (u.Port() == "" || u.Port() == "80") && r.Method == http.MethodGet {
		return "discovery"
	}
	if u.Scheme != "https" || (u.Port() != "" && u.Port() != "443") {
		return "other"
	}
	switch u.Path {
	case "/cgi-bin/rad_user_info":
		if r.Method == http.MethodGet {
			return "status"
		}
	case "/v1/srun_portal_sso":
		if r.Method == http.MethodGet {
			return "activation"
		}
	case "/cgi-bin/srun_portal":
		return "logout"
	}
	return "other"
}
func transportCategory(err error) string {
	if err == nil {
		return "none"
	}
	if errors.Is(err, context.Canceled) {
		return "canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	var n net.Error
	if errors.As(err, &n) && n.Timeout() {
		return "timeout"
	}
	var dns *net.DNSError
	if errors.As(err, &dns) {
		return "dns"
	}
	var cert *tls.CertificateVerificationError
	var unknown x509.UnknownAuthorityError
	var invalid x509.CertificateInvalidError
	var host x509.HostnameError
	if errors.As(err, &cert) || errors.As(err, &unknown) || errors.As(err, &invalid) || errors.As(err, &host) {
		return "tls_certificate"
	}
	var record tls.RecordHeaderError
	var alert tls.AlertError
	if errors.As(err, &record) || errors.As(err, &alert) {
		return "tls_other"
	}
	return "other"
}
func (c *Collector) result(err error) {
	c.mu.Lock()
	logoutResponse := c.data.Command == "logout" && c.data.State == "complete" && len(c.data.Requests) > 0 && c.data.Requests[len(c.data.Requests)-1].Role == "logout"
	c.mu.Unlock()
	code, category := "", "success"
	if err != nil {
		code = string(ipgw.CodeOf(err))
		category = "unknown"
		var e *ipgw.Error
		if errors.As(err, &e) {
			switch {
			case logoutResponse && e.Code == ipgw.CodeProtocolChanged && e.Message == "gateway response exceeded the safe limit":
				category = "logout_response_limit"
			case c.data.Command == "logout" && e.Code == ipgw.CodeProtocolChanged && e.Message == "gateway logout response is not recognized":
				category = "logout_response_parse"
			case c.data.Command == "logout" && e.Code == ipgw.CodeAuthentication && e.Message == "gateway rejected logout":
				category = "logout_business_rejected"
			case e.Code == ipgw.CodeNetwork && e.Message == "CAS request failed":
				category = "cas_request"
			case e.Code == ipgw.CodeNetwork && e.Message == "CAS returned an unexpected HTTP status":
				category = "cas_http_status"
			case e.Code == ipgw.CodeNetwork && e.Message == "gateway activation request failed":
				category = "activation_request"
			case e.Code == ipgw.CodeNetwork && e.Message == "gateway activation returned an unexpected HTTP status":
				category = "activation_http_status"
			case e.Code == ipgw.CodeConfig && e.Message == "credential provider failed":
				category = "credential_provider"
			case e.Code == ipgw.CodeProtocolChanged && e.Message == "CAS public key or encryption contract changed":
				category = "encryption"
			case e.Code == ipgw.CodeSessionConflict && e.Message == "gateway activated a different account":
				category = "status_identity"
			}
		}
		if errors.Is(err, context.Canceled) {
			category = "canceled"
		} else if errors.Is(err, context.DeadlineExceeded) {
			category = "timeout"
		}
	}
	if !validCode(code) {
		code = "internal"
		category = "unknown"
	}
	c.mu.Lock()
	c.data.SDKCode, c.data.SDKCategory = code, category
	c.mu.Unlock()
}

type gateway struct {
	cli.Gateway
	c *Collector
}

func (c *Collector) Gateway(next cli.Gateway) cli.Gateway { return &gateway{Gateway: next, c: c} }
func (g *gateway) Login(ctx context.Context, o app.LoginOptions) (ipgw.LoginResult, error) {
	r, e := g.Gateway.Login(ctx, o)
	g.c.result(e)
	return r, e
}
func (g *gateway) Status(ctx context.Context, o app.RequestOptions) (ipgw.Status, error) {
	r, e := g.Gateway.Status(ctx, o)
	g.c.result(e)
	return r, e
}
func (g *gateway) Logout(ctx context.Context, o app.RequestOptions) (ipgw.LogoutResult, error) {
	r, e := g.Gateway.Logout(ctx, o)
	g.c.result(e)
	return r, e
}
