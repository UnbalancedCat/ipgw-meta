package ipgw

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestQRContinuationTargets(t *testing.T) {
	service, _ := url.Parse("http://ipgw.neu.edu.cn/srun_portal_sso?ac_id=15")
	casURL, _ := url.Parse("https://pass.neu.edu.cn/tpass/login")
	for _, tail := range []string{"?ac_id=15", "?ac_id=1", "?ac_id=15&ticket=", "?ac_id=15&ticket=&ticket=x", "?ac_id=15&extra=x", "?ac_id=15&ac_id=15", "?ac_id=15#x", "?ac_id=15;bad=x", "?ac_id=15&service=x"} {
		u, _ := url.Parse("http://ipgw.neu.edu.cn/srun_portal_sso" + tail)
		if noTicketQRService(u, service) != (tail == "?ac_id=15") {
			t.Fatalf("classification %q", tail)
		}
	}
	base := "http://ipgw.neu.edu.cn/srun_portal_sso"
	good := casURL.String() + "?service=" + url.QueryEscape(base)
	login, err := qrContinuationLogin(good, casURL, service)
	if err != nil || login.Query().Get("service") != service.String() {
		t.Fatal("valid continuation rejected")
	}
	for _, raw := range []string{
		strings.Replace(good, "https:", "http:", 1), good + "&x=1", good + "&service=x", good + "#x", good + "#",
		strings.Replace(good, "pass.neu.edu.cn", "evil.example", 1), strings.Replace(good, "/tpass/login", "/tpass/%6cogin", 1),
		casURL.String() + "?service=" + url.QueryEscape(base+"?ac_id=15"), casURL.String() + "?service=" + url.QueryEscape(base+"?"),
		casURL.String() + "?service=" + url.QueryEscape(base+"#"),
		casURL.String() + "?service=" + url.QueryEscape("http://ipgw.neu.edu.cn:81/srun_portal_sso"),
		casURL.String() + "?service=" + url.QueryEscape("http://user@ipgw.neu.edu.cn/srun_portal_sso"),
	} {
		if _, err := qrContinuationLogin(raw, casURL, service); err == nil {
			t.Fatal("unsafe redirect accepted")
		}
	}
}

func TestQRInvalidServiceNeverQueriesGateway(t *testing.T) {
	service, _ := url.Parse("http://ipgw.neu.edu.cn/srun_portal_sso?ac_id=15")
	for _, raw := range []string{
		service.String() + "#", service.String() + "#x", service.String() + "&ticket=",
		service.String() + "&ticket=&ticket=synthetic", service.String() + "&extra=x",
		service.String() + "&ac_id=15", "http://ipgw.neu.edu.cn/srun_portal_sso?ac_id=1",
		"http://user@ipgw.neu.edu.cn/srun_portal_sso?ac_id=15", "http://ipgw.neu.edu.cn:81/srun_portal_sso?ac_id=15",
		"http://ipgw.neu.edu.cn/srun_portal_%73so?ac_id=15", "http://other.invalid/srun_portal_sso?ac_id=15",
	} {
		calls := 0
		c, err := NewClient(WithRoundTripper(roundTripFunc(func(*http.Request) (*http.Response, error) {
			calls++
			return nil, errors.New("synthetic unexpected request")
		})))
		if err != nil {
			t.Fatal(err)
		}
		_, err = c.finishQRConfirmation(context.Background(), c.newHTTPClient(nil), LoginRequest{Method: AuthMethodTerminalQR, ExpectedUsername: "alice"}, service, "15", raw)
		if err == nil || calls != 0 {
			t.Fatal("invalid service emitted a request")
		}
	}
}

func TestQRNoTicketServiceFullFlow(t *testing.T) {
	for _, mode := range []string{"pass", "gateway_invalid", "gateway_redirect", "cas_challenge", "cas_no_ticket", "activation_failure", "wrong_identity", "status_invalid", "cancel_gateway", "cancel_cas", "cancel_activation", "cancel_status", "oversize", "short_parent"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var parentDeadline time.Time
			if mode == "short_parent" {
				var stop context.CancelFunc
				ctx, stop = context.WithTimeout(ctx, 2*time.Second)
				defer stop()
				parentDeadline, _ = ctx.Deadline()
			}
			gatewayCalls, casCalls, activationCalls, statusCalls := 0, 0, 0, 0
			confirmationUpperBound := time.Now().Add(60 * time.Second)
			var events []Event
			rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if gatewayCalls > 0 {
					d, ok := r.Context().Deadline()
					if !ok || d.After(confirmationUpperBound) {
						t.Fatal("confirmation deadline missing or extended")
					}
					if mode == "short_parent" && !d.Equal(parentDeadline) {
						t.Fatal("short parent deadline extended")
					}
				}
				if r.Method != http.MethodGet {
					t.Fatal("non-GET")
				}
				if r.URL.Scheme == "http" && r.URL.Path != "/" {
					t.Fatal("HTTP service sent")
				}
				if r.URL.Path == "/" {
					return testResponse(r, 302, "", "http://ipgw.neu.edu.cn/?ac_id=15"), nil
				}
				if r.URL.Path == "/cgi-bin/rad_user_info" {
					statusCalls++
					if statusCalls == 1 {
						return testResponse(r, 200, "not_online_error", ""), nil
					}
					if deadline, ok := r.Context().Deadline(); !ok || deadline.After(confirmationUpperBound) {
						t.Fatal("status outside confirmation deadline")
					}
					if mode == "cancel_status" {
						cancel()
						return nil, r.Context().Err()
					}
					if mode == "status_invalid" {
						return testResponse(r, 200, "unknown", ""), nil
					}
					name := "alice"
					if mode == "wrong_identity" {
						name = "other"
					}
					return testResponse(r, 200, `{"error":"ok","user_name":"`+name+`","online_ip":"10.0.0.8"}`, ""), nil
				}
				if r.URL.Host == "pass.neu.edu.cn" {
					if _, err := r.Cookie("gateway_synthetic"); err == nil {
						t.Fatal("gateway cookie leaked to CAS")
					}
					if r.URL.Path == "/tpass/checkQRCodeScan" {
						return testResponse(r, 200, `{"redirect_url":"http://ipgw.neu.edu.cn/srun_portal_sso?ac_id=15"}`, ""), nil
					}
					casCalls++
					if casCalls == 1 {
						resp := testResponse(r, 200, `<script src="/tpass/comm/js/login-qrcode.js"></script>`, "")
						resp.Header.Add("Set-Cookie", "cas_synthetic=canary; Path=/; Secure")
						return resp, nil
					}
					if c, err := r.Cookie("cas_synthetic"); err != nil || c.Value != "canary" {
						t.Fatal("CAS session lost")
					}
					if mode == "cancel_cas" {
						cancel()
						return nil, r.Context().Err()
					}
					if mode == "cas_challenge" {
						return testResponse(r, 200, "<html>需要手机验证码</html>", ""), nil
					}
					if mode == "cas_no_ticket" {
						return testResponse(r, 200, "<html>login</html>", ""), nil
					}
					return testResponse(r, 302, "", "http://ipgw.neu.edu.cn/srun_portal_sso?ac_id=15&ticket=SYNTHETIC-QR-CANARY"), nil
				}
				if r.URL.Path != "/v1/srun_portal_sso" {
					t.Fatal("unexpected path")
				}
				if _, err := r.Cookie("cas_synthetic"); err == nil {
					t.Fatal("CAS cookie leaked to gateway")
				}
				if r.URL.Query().Get("ticket") == "" {
					gatewayCalls++
					if gatewayCalls != 1 || len(r.URL.Query()) != 1 {
						t.Fatal("gateway retry or extra query")
					}
					d, ok := r.Context().Deadline()
					if !ok || d.After(confirmationUpperBound) || mode == "short_parent" && !d.Equal(parentDeadline) {
						t.Fatal("confirmation budget missing")
					}
					if mode == "cancel_gateway" {
						cancel()
						return nil, r.Context().Err()
					}
					if mode == "gateway_redirect" {
						return testResponse(r, 302, "", "https://pass.neu.edu.cn/tpass/login"), nil
					}
					if mode == "gateway_invalid" {
						return testResponse(r, 200, `{"code":0,"message":"success"}`, ""), nil
					}
					if mode == "oversize" {
						return testResponse(r, 200, strings.Repeat("x", int(apiResponseLimit)+1), ""), nil
					}
					body, _ := json.Marshal(map[string]any{"code": 0, "message": "continue", "Redirect": "https://pass.neu.edu.cn/tpass/login?service=" + url.QueryEscape("http://ipgw.neu.edu.cn/srun_portal_sso"), "ID": ""})
					resp := testResponse(r, 200, string(body), "")
					resp.Header.Add("Set-Cookie", "gateway_synthetic=canary; Path=/; Secure")
					return resp, nil
				}
				activationCalls++
				if c, err := r.Cookie("gateway_synthetic"); err != nil || c.Value != "canary" {
					t.Fatal("gateway session lost")
				}
				if mode == "cancel_activation" {
					cancel()
					return nil, r.Context().Err()
				}
				if mode == "activation_failure" {
					return testResponse(r, 200, `{"code":1,"message":"denied"}`, ""), nil
				}
				return testResponse(r, 200, `{"code":0,"message":"success"}`, ""), nil
			})
			c, err := NewClient(WithRoundTripper(rt), WithObserver(ObserverFunc(func(_ context.Context, e Event) { events = append(events, e) })))
			if err != nil {
				t.Fatal(err)
			}
			c.pollInterval = 0
			c.verifyDelay = 0
			result, err := c.Login(ctx, LoginRequest{Method: AuthMethodTerminalQR, ExpectedUsername: "alice", Interactions: InteractionHandlerFunc(func(context.Context, QRCodePrompt) error { return nil })})
			if mode == "pass" || mode == "short_parent" {
				if err != nil || result.Outcome != LoginLoggedIn || activationCalls != 1 || casCalls != 2 {
					t.Fatalf("flow failed: %v", err)
				}
			} else if err == nil {
				t.Fatal("negative flow passed")
			}
			if gatewayCalls != 1 {
				t.Fatal("gateway count")
			}
			switch mode {
			case "gateway_invalid", "gateway_redirect", "oversize", "cancel_gateway":
				if casCalls != 1 || activationCalls != 0 || statusCalls != 1 {
					t.Fatal("continued after gateway failure")
				}
			case "cas_challenge", "cas_no_ticket", "cancel_cas":
				if casCalls != 2 || activationCalls != 0 || statusCalls != 1 {
					t.Fatal("continued after CAS failure")
				}
			case "activation_failure", "cancel_activation":
				if activationCalls != 1 || statusCalls != 1 {
					t.Fatal("continued after activation failure")
				}
			}
			if strings.HasPrefix(mode, "cancel_") && !errors.Is(err, context.Canceled) {
				t.Fatal("cancellation lost")
			}
			b, _ := json.Marshal(events)
			if strings.Contains(string(b), "SYNTHETIC-QR-CANARY") || strings.Contains(string(b), "cas_synthetic") {
				t.Fatal("sensitive event")
			}
		})
	}
}
