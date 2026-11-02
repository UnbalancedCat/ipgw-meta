package ipgw

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/srun"
)

// Matching is deliberately stricter than browser URL normalization.
func exactQRTarget(target, expected *url.URL) bool {
	return sameServiceTarget(target, expected) && target.RawPath == expected.RawPath && target.Opaque == "" && !target.ForceQuery
}

func noTicketQRService(target, service *url.URL) bool {
	if !exactQRTarget(target, service) {
		return false
	}
	q, err := url.ParseQuery(target.RawQuery)
	return err == nil && len(q) == 1 && len(q["ac_id"]) == 1 && q.Get("ac_id") == service.Query().Get("ac_id") && q.Get("ac_id") != ""
}

func qrContinuationLogin(raw string, casLogin, service *url.URL) (*url.URL, error) {
	fail := func() (*url.URL, error) {
		return nil, newError(CodeProtocolChanged, "QR gateway continuation is invalid", false, nil)
	}
	u, err := url.Parse(raw)
	if err != nil || strings.Contains(raw, "#") || !exactQRTarget(u, casLogin) || u.Scheme != "https" {
		return fail()
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil || len(q) != 1 || len(q["service"]) != 1 {
		return fail()
	}
	embedded, err := url.Parse(q.Get("service"))
	base := cloneURL(service)
	base.RawQuery = ""
	if err != nil || strings.Contains(q.Get("service"), "#") || !exactQRTarget(embedded, base) || embedded.RawQuery != "" {
		return fail()
	}
	login := cloneURL(casLogin)
	values := url.Values{}
	values.Set("service", service.String())
	login.RawQuery = values.Encode()
	return login, nil
}

func (c *Client) continueQRService(ctx context.Context, client *http.Client, request LoginRequest, service *url.URL, acID string) (LoginResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	target := cloneURL(c.endpoints.activation)
	if err := requireHTTPS(target); err != nil {
		return LoginResult{}, err
	}
	if !sameHostname(target, service) {
		return LoginResult{}, newError(CodeProtocolChanged, "QR gateway host mismatch", false, nil)
	}
	q := url.Values{}
	q.Set("ac_id", acID)
	target.RawQuery = q.Encode()
	gateway := c.newHTTPClient(nil)
	gateway.Jar = client.Jar
	data, _, err := c.request(ctx, gateway, http.MethodGet, target, nil, apiResponseLimit)
	if err != nil {
		return LoginResult{}, err
	}
	raw, err := srun.ParseQRContinuation(data)
	if err != nil {
		return LoginResult{}, newError(CodeProtocolChanged, "QR gateway continuation is not recognized", false, nil)
	}
	login, err := qrContinuationLogin(raw, c.endpoints.casLogin, service)
	if err != nil {
		return LoginResult{}, err
	}
	return c.finishQRConfirmation(ctx, gateway, request, service, acID, login.String())
}
