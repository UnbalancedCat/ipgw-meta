package authcontext

import (
	"context"
	"encoding/json"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
	"github.com/UnbalancedCat/ipgw-meta/internal/castest"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestStaticObserverPreservesLegacyAndRequests(t *testing.T) {
	key := generatedKey(t)
	page := `<form id="loginForm"><input name="lt" value="synthetic-state"><input name="execution" value="synthetic-execution"></form><script src="/tpass/login-synthetic.js"></script>`
	makeTransport := func(trace *[]string) http.RoundTripper {
		return roundTripFunc(func(req *http.Request) (*http.Response, error) {
			*trace = append(*trace, req.Method+" "+req.URL.String())
			switch req.URL.Host + req.URL.Path {
			case "ipgw.neu.edu.cn/cgi-bin/rad_user_info":
				return syntheticResponse(req, 200, "not_online_error"), nil
			case "ipgw.neu.edu.cn/":
				return syntheticResponse(req, 200, "ac_id=1"), nil
			case "pass.neu.edu.cn/tpass/login":
				return syntheticResponse(req, 200, page), nil
			case "pass.neu.edu.cn/tpass/login-synthetic.js":
				return syntheticResponse(req, 200, `var ratio=6/2;function f(){publicKeyStr='`+key+`';}`+castest.LoginScript(key)), nil
			}
			return nil, ErrReport
		})
	}
	var before, after []string
	old := Anonymous(context.Background(), makeTransport(&before), func(context.Context) bool { return true })
	r := AnonymousStatic(context.Background(), makeTransport(&after), func(context.Context) bool { return true }, casprovenance.AnalyzeStaticParserMatch)
	if r.Validate() != nil || r.State != "observed" || !reflect.DeepEqual(old, r.Context) || !reflect.DeepEqual(before, after) || len(after) != 4 {
		t.Fatalf("SDK or old observation changed: state=%s reason=%s valid=%t old_same=%t trace_same=%t count=%d", r.State, r.Reason, r.Validate() == nil, reflect.DeepEqual(old, r.Context), reflect.DeepEqual(before, after), len(after))
	}
	if old.Analysis.Occurrences[0].Context != "unknown" || r.Syntax.Occurrences[0].Context != "code" || r.Syntax.Occurrences[0].Placement != "function" {
		t.Fatal("static separation missing")
	}
	b, e := EncodeStatic(r)
	if e != nil || strings.Contains(string(b), key) || strings.Contains(string(b), "synthetic-state") {
		t.Fatal("leak")
	}
	if _, e = DecodeStatic(b); e != nil {
		t.Fatal("decode")
	}
	for _, mutate := range []func(*StaticReport){func(r *StaticReport) { r.SchemaVersion = 1 }, func(r *StaticReport) { r.ReleaseEvidence = true }, func(r *StaticReport) { r.Syntax = nil }, func(r *StaticReport) { r.State = "pass" }} {
		bad := r
		mutate(&bad)
		b, _ := json.Marshal(bad)
		if _, e = DecodeStatic(b); e == nil {
			t.Fatal("contradiction accepted")
		}
	}
	for _, b := range []string{string(b) + string(b), strings.Replace(string(b), `"schema_version":2`, `"schema_version":2,"extra":true`, 1), strings.Replace(string(b), `"schema_version":2`, `"schema_version":2,"schema_version":2`, 1)} {
		if _, e = DecodeStatic([]byte(b)); e == nil {
			t.Fatal("framing accepted")
		}
	}
}
func TestStaticProtectionFailureNeverParses(t *testing.T) {
	called := false
	r := protectedStatic(context.Background(), func() bool { return false }, func(context.Context) casprovenance.StaticReport {
		called = true
		return casprovenance.EmptyStatic("parser_failure")
	}, 1)
	if called || r.Reason != "protection_unavailable" || r.Validate() != nil {
		t.Fatal("unprotected parse")
	}
}
func FuzzStaticEnvelope(f *testing.F) {
	b, _ := EncodeStatic(NewStaticReport("preflight"))
	f.Add(b)
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 65536 {
			return
		}
		r, e := DecodeStatic(b)
		if e == nil {
			encoded, e := EncodeStatic(r)
			if e != nil {
				t.Fatal("invalid roundtrip")
			}
			if _, e = DecodeStatic(encoded); e != nil {
				t.Fatal("roundtrip decode")
			}
		}
	})
}
