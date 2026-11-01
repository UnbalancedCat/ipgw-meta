package cascollect

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func d3Prefix(location string, hop bool) []Exchange {
	first := Exchange{Target: "http://ipgw.neu.edu.cn/", Status: 302, Headers: []Header{{"Location", location}}}
	if !hop {
		return []Exchange{first}
	}
	first.Target = "http://ipgw.neu.edu.cn/next"
	return []Exchange{{Target: "http://ipgw.neu.edu.cn/", Status: 302, Headers: []Header{{"Location", "/next"}}}, first}
}

func TestD3ExtractOnlyACID(t *testing.T) {
	for i, location := range []string{
		"/never-follow?ac_id=7&extra=1",
		"/never-follow?extra=synthetic-body-canary&ac_id=7",
		"/never-follow?ac_id=7&extra=&extra=synthetic-body-canary",
		"/never-follow?ac_id=7&hint=https%3A%2F%2Foutside.invalid%2Fsecret%3Ftoken%3Dsynthetic-canary",
		"/never-follow?ac_id=7&hint=%2525%00%0A",
		"/never-follow?ac_id=7&" + strings.Repeat("a", 64) + "=" + strings.Repeat("z", 1500),
		"https://ipgw.neu.edu.cn/never-follow?ac_id=%37&ex%74ra=1",
	} {
		for _, hop := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/%t", i, hop), func(t *testing.T) {
				prefix := d3Prefix(location, hop)
				next, _, r := NextOfficial(context.Background(), prefix)
				if next == nil || next.Method != "GET" || next.Target != "https://pass.neu.edu.cn/tpass/login?service=http%3A%2F%2Fipgw.neu.edu.cn%2Fsrun_portal_sso%3Fac_id%3D7" || !r.ContextConstructed {
					t.Fatal("did not construct fixed ac_id-only request")
				}
				prefix = append(prefix, Exchange{Target: next.Target, Status: 200, Headers: []Header{{"Content-Type", "text/html"}}, Body: []byte("<html></html>")})
				next, _, r = NextOfficial(context.Background(), prefix)
				assertReport(t, r)
				b, _ := EncodeReport(r)
				if next != nil || r.Collection != "complete" || r.Simulated.Requests != len(prefix) {
					t.Fatal("unexpected extra request")
				}
				for _, forbidden := range []string{"never-follow", "extra", "hint", "outside.invalid", "synthetic-canary"} {
					if strings.Contains(string(b), forbidden) {
						t.Fatal("discarded data in output")
					}
				}
			})
		}
	}
}

func TestD3AuthenticationMarkersStop(t *testing.T) {
	// Independent list derived from the accepted contract, not the guard helper.
	for _, key := range []string{"ticket", "tgt", "lt", "execution", "eventid", "code", "state", "session", "sessionid", "sid", "jsessionid", "token", "accesstoken", "idtoken", "refreshtoken", "samlrequest", "samlresponse", "relaystate", "assertion", "cookie", "authorization", "password", "passwd", "pwd", "credential", "credentials", "otp", "smscode"} {
		for variant, name := range []string{key, strings.ToUpper(key), fmt.Sprintf("%%%02x%s", key[0], key[1:]), key[:1] + "_" + key[1:]} {
			for _, hop := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%d/%t", key, variant, hop), func(t *testing.T) {
					prefix := d3Prefix("/never-follow?ac_id=7&"+name+"=&"+name+"=", hop)
					next, _, r := NextOfficial(context.Background(), prefix)
					assertReport(t, r)
					if next != nil || r.ContextConstructed || r.Reason != "discovery_query_auth_marker" || r.Simulated.Requests != len(prefix) {
						t.Fatal("authentication marker did not stop before CAS")
					}
				})
			}
		}
	}
}

func TestD3KeyShapesAndFailurePrecedence(t *testing.T) {
	for _, key := range []string{"", "AC_ID", "Ac_Id", "acid", "a_c_i_d", "_eventId", "ticket%5B%5D", "ticket+", "%00ticket", "%0Aticket", "%EF%BC%94", "%2574icket", "session-id", "1hint", strings.Repeat("a", 65)} {
		for _, hop := range []bool{false, true} {
			// A marker is present too: invalid shape must win deterministically.
			prefix := d3Prefix("/never-follow?ac_id=7&token=&"+key+"=", hop)
			for repeat := 0; repeat < 8; repeat++ {
				next, _, r := NextOfficial(context.Background(), prefix)
				assertReport(t, r)
				if next != nil || r.ContextConstructed || r.Reason != "discovery_query_key_rejected" {
					t.Fatal("key shape precedence drift")
				}
			}
		}
	}
	for _, tc := range []struct{ query, reason string }{
		{"ac_id=7&ac%5fid=7&token=", "discovery_query_duplicate"},
		{"ac_id=bad&token=", "discovery_query_invalid_value"},
		{"ac_id=7&extra=%XX", "url_rejected"},
		{"ac_id=7&extra=a;b", "url_rejected"},
		{"extra=1", "url_rejected"},
	} {
		next, _, r := NextOfficial(context.Background(), d3Prefix("/never-follow?"+tc.query, true))
		assertReport(t, r)
		if next != nil || r.ContextConstructed || r.Reason != tc.reason {
			t.Fatal("earlier failure changed")
		}
	}
}
