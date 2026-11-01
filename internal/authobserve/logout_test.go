package authobserve

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	ipgw "github.com/UnbalancedCat/ipgw-meta"
	"github.com/UnbalancedCat/ipgw-meta/internal/srun"
	transport "github.com/UnbalancedCat/ipgw-meta/internal/transport"
)

func TestLogoutStructureMatrix(t *testing.T) {
	for _, tc := range []struct{ input, envelope, ecode, success, parser string }{
		{`{"ecode":0}`, "json_object", "integer_zero", "missing", "accepted_success"},
		{`synthetic_cb({"ecode":-0,"success":true});`, "jsonp_object", "integer_zero", "true", "accepted_success"},
		{`{"ecode":2,"success":false}`, "json_object", "integer_nonzero", "false", "accepted_rejection"},
		{`{"ecode":"0"}`, "json_object", "string_zero", "missing", "rejected"},
		{`{"ecode":"private-canary"}`, "json_object", "string_other", "missing", "rejected"},
		{`{"ecode":0.0}`, "json_object", "noninteger_number", "missing", "rejected"},
		{`{"ecode":0e0}`, "json_object", "noninteger_number", "missing", "rejected"},
		{`{"ecode":9223372036854775808}`, "json_object", "integer_out_of_range", "missing", "rejected"},
		{`{"ecode":0,"success":false}`, "json_object", "integer_zero", "false", "rejected"},
		{`{"ecode":0,"success":"true"}`, "json_object", "integer_zero", "other_type", "rejected"},
		{`{"code":0,"message":"private-canary","error":null,"error_msg":[],"res":{},"another-private-field":true}`, "json_object", "missing", "missing", "rejected"},
		{`{"ecode":null}`, "json_object", "other_type", "missing", "rejected"},
		{`{"ecode":{}}`, "json_object", "other_type", "missing", "rejected"},
		{`{"ecode":[]}`, "json_object", "other_type", "missing", "rejected"},
		{`{"ecode":true}`, "json_object", "other_type", "missing", "rejected"},
	} {
		r := analyzeLogout(context.Background(), []byte(tc.input), 0)
		if r.Envelope != tc.envelope || r.Ecode != tc.ecode || r.Success != tc.success || r.Parser != tc.parser || r.validate() != nil {
			t.Fatalf("projection mismatch: %+v", r)
		}
		encoded, _ := json.Marshal(r)
		if strings.Contains(string(encoded), "private") || strings.Contains(string(encoded), "synthetic_cb") {
			t.Fatal("material leaked")
		}
	}
}
func TestLogoutMalformedAndBudgets(t *testing.T) {
	for _, input := range []string{`{"ecode":0,"\u0065code":0}`, `{"ecode":0,"x":{"a":1,"a":2}}`} {
		r := analyzeLogout(context.Background(), []byte(input), 0)
		if r.Structure != "duplicate_key" || r.Fields != fields("unknown") || r.validate() != nil {
			t.Fatal("duplicate projection")
		}
	}
	for _, input := range []string{`{"ecode":0};`, `{"ecode":0} {}`, `cb({"ecode":0},1)`, `cb({"ecode":0});;`, `{"ecode":`, `{"ecode":01}`} {
		r := analyzeLogout(context.Background(), []byte(input), 0)
		if r.Envelope != "malformed" || r.Structure != "malformed" || r.validate() != nil {
			t.Fatalf("malformed classified as %+v", r)
		}
	}
	for _, input := range []string{"", " ", "<synthetic>", "[]", "\xef\xbb\xbf{\"ecode\":0}"} {
		r := analyzeLogout(context.Background(), []byte(input), 0)
		if r.Parser != "unknown" || r.validate() != nil {
			t.Fatal("non-object accepted")
		}
	}
	if r := analyzeLogout(context.Background(), []byte("\xef\xbb\xbf{}"), 0); r.BOM != "present" {
		t.Fatal("BOM lost")
	}
	for _, tc := range []struct{ input, reason string }{
		{`{"x":` + strings.Repeat("[", 33) + "0" + strings.Repeat("]", 33) + "}", "depth_limit"},
		{`{"x":[` + strings.Repeat("0,", 32768) + "0]}", "token_limit"},
		{`{"ecode":0,"x":"` + strings.Repeat("x", 250000) + `"}`, "work_limit"},
	} {
		r := analyzeLogout(context.Background(), []byte(tc.input), 0)
		if r.Read != "budget_exceeded" || r.Structure != tc.reason || r.validate() != nil {
			t.Fatalf("budget: %+v", r)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if r := analyzeLogout(ctx, []byte(`{"ecode":0}`), 0); r.Read != "canceled" || r.validate() != nil {
		t.Fatal("cancellation")
	}
}

type probeBody struct {
	data          []byte
	chunk         int
	final         error
	closeErr      error
	reads, closes int
}

func (p *probeBody) Read(b []byte) (int, error) {
	p.reads++
	if len(p.data) == 0 {
		if p.final != nil {
			return 0, p.final
		}
		return 0, io.EOF
	}
	n := min(len(b), len(p.data))
	if p.chunk > 0 {
		n = min(n, p.chunk)
	}
	copy(b, p.data[:n])
	p.data = p.data[n:]
	if len(p.data) == 0 && p.final != nil {
		return n, p.final
	}
	return n, nil
}
func (p *probeBody) Close() error { p.closes++; return p.closeErr }
func wrapSynthetic(t *testing.T, c *Collector, ctx context.Context, p io.ReadCloser) io.ReadCloser {
	t.Helper()
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://ipgw.neu.edu.cn/cgi-bin/srun_portal", nil)
	r, e := c.Decorate(roundTripFunc(func(*http.Request) (*http.Response, error) { return &http.Response{StatusCode: 200, Body: p}, nil })).RoundTrip(req)
	if e != nil {
		t.Fatal(e)
	}
	return r.Body
}
func TestLogoutBodyPreservesIO(t *testing.T) {
	privateErr := errors.New("private-read-canary")
	for _, tc := range []struct {
		name, input, read string
		final, close      error
		chunk             int
	}{
		{"split", `{"ecode":0}`, "complete", nil, nil, 1},
		{"data-eof", `{"ecode":0}`, "complete", io.EOF, nil, 64},
		{"data-error", `{"ecode":0}`, "read_error", privateErr, nil, 64},
		{"close-error", `{"ecode":0}`, "read_error", nil, privateErr, 64},
		{"limit", strings.Repeat(" ", logoutLimit), "complete", nil, nil, 0},
		{"overflow", strings.Repeat(" ", logoutLimit+1), "too_large", nil, nil, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plain := &probeBody{data: []byte(tc.input), final: tc.final, closeErr: tc.close, chunk: tc.chunk}
			body := &probeBody{data: []byte(tc.input), final: tc.final, closeErr: tc.close, chunk: tc.chunk}
			c := New("logout")
			wrapped := wrapSynthetic(t, c, context.Background(), body)
			a, ae := transport.ReadAll(plain, logoutLimit)
			ac := plain.Close()
			b, be := transport.ReadAll(wrapped, logoutLimit)
			bc := wrapped.Close()
			if string(a) != string(b) || ae != be || ac != bc || plain.reads != body.reads || plain.closes != body.closes {
				t.Fatal("I/O changed")
			}
			c.result(nil)
			o := c.Snapshot()
			if o.ValidateVersion(2) != nil || len(o.Logout.Responses) != 1 || o.Logout.Responses[0].Read != tc.read {
				t.Fatalf("report: %+v", o.Logout)
			}
			encoded, _ := json.Marshal(o)
			if strings.Contains(string(encoded), "private-") {
				t.Fatal("error leaked")
			}
		})
	}
}
func TestLogoutEarlyCloseCancellationAndCommandLimit(t *testing.T) {
	c := New("logout")
	b := wrapSynthetic(t, c, context.Background(), io.NopCloser(strings.NewReader(`{"ecode":0}`)))
	if c.Snapshot().State != "incomplete" {
		t.Fatal("pending body published")
	}
	b.Close()
	c.result(nil)
	if c.Snapshot().Logout.Responses[0].Read != "incomplete" {
		t.Fatal("early close")
	}
	ctx, cancel := context.WithCancel(context.Background())
	b = wrapSynthetic(t, c, ctx, io.NopCloser(strings.NewReader(`{"ecode":0}`)))
	cancel()
	io.ReadAll(b)
	b.Close()
	if c.Snapshot().Logout.Responses[1].Read != "canceled" {
		t.Fatal("cancel")
	}
	b = wrapSynthetic(t, c, context.Background(), io.NopCloser(strings.NewReader(`{"ecode":0}`)))
	io.ReadAll(b)
	b.Close()
	o := c.Snapshot()
	if o.Logout.State != "budget_exceeded" || len(o.Logout.Responses) != 0 || o.ValidateVersion(2) != nil {
		t.Fatal("third body accepted")
	}
}
func TestLogoutNotCollectedOutsideExactCommandAndGET(t *testing.T) {
	for _, tc := range []struct{ command, method, url string }{
		{"login", "GET", "https://ipgw.neu.edu.cn/cgi-bin/srun_portal"},
		{"logout", "POST", "https://ipgw.neu.edu.cn/cgi-bin/srun_portal"},
		{"logout", "GET", "https://ipgw.neu.edu.cn/cgi-bin/rad_user_info"},
		{"logout", "GET", "http://ipgw.neu.edu.cn/cgi-bin/srun_portal"},
		{"logout", "GET", "https://other.invalid/cgi-bin/srun_portal"},
	} {
		c := New(tc.command)
		req, _ := http.NewRequest(tc.method, tc.url, nil)
		body := forbiddenBody{}
		resp := &http.Response{StatusCode: 200, Body: body}
		got, e := c.Decorate(roundTripFunc(func(*http.Request) (*http.Response, error) { return resp, nil })).RoundTrip(req)
		if e != nil || got.Body != body {
			t.Fatal("unrelated body wrapped")
		}
	}
}
func TestLogoutVersionedRecords(t *testing.T) {
	c := New("logout")
	b := wrapSynthetic(t, c, context.Background(), io.NopCloser(strings.NewReader(`{"ecode":"0"}`)))
	io.ReadAll(b)
	b.Close()
	c.result(&ipgw.Error{Code: ipgw.CodeProtocolChanged, Message: "gateway logout response is not recognized"})
	o := c.Snapshot()
	child := Child{SchemaVersion: 2, ToolKind: "auth-observed-cli", Command: "logout", ExitCode: 6, Product: json.RawMessage(`{"schema_version":1,"command":"logout","ok":false,"error":{"code":"protocol_changed","message":"fixed","retryable":false,"details":{}}}`), Observation: o}
	data, _ := json.Marshal(child)
	if _, e := DecodeChild(data, "logout", 6); e != nil {
		t.Fatal("new child rejected", e)
	}
	for _, mutate := range []func([]byte) []byte{
		func(b []byte) []byte {
			return []byte(strings.Replace(string(b), `"schema_version":2`, `"schema_version":1`, 1))
		},
		func(b []byte) []byte { return []byte(strings.Replace(string(b), `"string_zero"`, `"integer_zero"`, 1)) },
		func(b []byte) []byte {
			return []byte(strings.Replace(string(b), `"request_index":0`, `"request_index":1`, 1))
		},
		func(b []byte) []byte {
			return []byte(strings.Replace(string(b), `"read":"complete"`, `"read":"too_large"`, 1))
		},
		func(b []byte) []byte {
			return []byte(strings.Replace(string(b), `"ecode":"string"`, `"ecode":"string","extra":0`, 1))
		},
	} {
		if _, e := DecodeChild(mutate(data), "logout", 6); e == nil {
			t.Fatal("mixed or contradictory record")
		}
	}
	child.Observation.Logout = nil
	child.Observation.SDKCategory = "unknown"
	data, _ = json.Marshal(child)
	if _, e := DecodeChild(data, "logout", 6); e == nil {
		t.Fatal("new child missing projection")
	}
	child.SchemaVersion = 1
	data, _ = json.Marshal(child)
	if _, e := DecodeChild(data, "logout", 6); e != nil {
		t.Fatal("historical child rejected")
	}
}

func TestLogoutCumulativeDiscardAndStatusLimitAttribution(t *testing.T) {
	c := New("logout")
	b1 := wrapSynthetic(t, c, context.Background(), io.NopCloser(strings.NewReader(`{"ecode":0}`)))
	b2 := wrapSynthetic(t, c, context.Background(), io.NopCloser(strings.NewReader(`{"ecode":0}`)))
	buf := make([]byte, 2)
	b1.Read(buf)
	b2.Read(buf)
	b3 := wrapSynthetic(t, c, context.Background(), io.NopCloser(strings.NewReader(`{"ecode":0}`)))
	for _, b := range []*logoutBody{b1.(*logoutBody), b2.(*logoutBody)} {
		if len(b.data) != 0 || b.read != "budget_exceeded" {
			t.Fatal("discard delayed")
		}
	}
	b1.Close()
	b2.Close()
	b3.Close()
	c.result(nil)
	if c.Snapshot().ValidateVersion(2) != nil {
		t.Fatal("invalid discarded window")
	}
	c = New("logout")
	req, _ := http.NewRequest("GET", "https://ipgw.neu.edu.cn/cgi-bin/rad_user_info", nil)
	c.Decorate(roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: forbiddenBody{}}, nil
	})).RoundTrip(req)
	c.result(&ipgw.Error{Code: ipgw.CodeProtocolChanged, Message: "gateway response exceeded the safe limit"})
	if c.Snapshot().SDKCategory != "unknown" {
		t.Fatal("status failure attributed to logout response")
	}
}

func TestLogoutConcurrentResponseOrder(t *testing.T) {
	c := New("logout")
	firstEntered, releaseFirst, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	r1, _ := http.NewRequest("GET", "https://ipgw.neu.edu.cn/cgi-bin/srun_portal", nil)
	r2 := r1.Clone(context.Background())
	rt := c.Decorate(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r == r1 {
			close(firstEntered)
			<-releaseFirst
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"ecode":0}`))}, nil
	}))
	go func() { defer close(done); r, _ := rt.RoundTrip(r1); io.ReadAll(r.Body); r.Body.Close() }()
	<-firstEntered
	r, _ := rt.RoundTrip(r2)
	io.ReadAll(r.Body)
	r.Body.Close()
	close(releaseFirst)
	<-done
	c.result(nil)
	o := c.Snapshot()
	if o.ValidateVersion(2) != nil || len(o.Logout.Responses) != 2 || o.Logout.Responses[0].RequestIndex != 0 || o.Logout.Responses[1].RequestIndex != 1 {
		t.Fatal("response completion reordered request identity")
	}
}
func FuzzLogoutStructure(f *testing.F) {
	for _, s := range []string{`{"ecode":0}`, `cb({"ecode":"0"});`, `{"ecode":0,"ecode":1}`, `{"success":true}`, "\xef\xbb\xbf{}"} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > logoutLimit+1 {
			return
		}
		r := analyzeLogout(context.Background(), data, 0)
		if r.validate() != nil {
			t.Fatalf("invalid projection: %+v", r)
		}
		if r.Parser != "unknown" {
			ok, e := srun.ParseLogoutSuccess(data)
			if (e != nil) != (r.Parser == "rejected") || e == nil && ok != (r.Parser == "accepted_success") {
				t.Fatal("parser disagreement")
			}
		}
	})
}

func TestLogoutSDKAndCleanupEquivalence(t *testing.T) {
	for _, tc := range []struct {
		name, body, category string
		offline              bool
	}{
		{"success", `{"ecode":0}`, "success", true},
		{"jsonp_success", `ipgw_logout({"ecode":0,"error":"ok"})`, "success", true},
		{"jsonp_ok_still_online", `ipgw_logout({"ecode":0,"error":"ok"})`, "unknown", false},
		{"rejected", `{"ecode":1}`, "logout_business_rejected", true},
		{"parse", `{"ecode":"0"}`, "logout_response_parse", true},
		{"large", strings.Repeat(" ", logoutLimit+1), "logout_response_limit", true},
		{"still_online", `{"ecode":1}`, "logout_business_rejected", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var traces [2][]string
			var codes [2][]ipgw.ErrorCode
			var outcomes [2][]ipgw.LogoutOutcome
			var first Observation
			for mode := 0; mode < 2; mode++ {
				online := true
				var active *Collector
				rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
					traces[mode] = append(traces[mode], req.Method+" "+req.URL.String())
					body := "not_online_error"
					switch req.URL.Path {
					case "/cgi-bin/rad_user_info":
						if online {
							body = `{"error":"ok","user_name":"synthetic-account","online_ip":"192.0.2.9"}`
						}
					case "/cgi-bin/srun_portal":
						body = tc.body
						online = !tc.offline
					default:
						t.Fatal("unexpected fake request")
					}
					return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: req}, nil
				})
				// A new collector per CLI command, while the fake server state stays.
				for step := 0; step < 2; step++ {
					var transport http.RoundTripper = rt
					if mode == 1 {
						active = New("logout")
						transport = active.Decorate(rt)
					}
					client, e := ipgw.NewClient(ipgw.WithRoundTripper(transport))
					if e != nil {
						t.Fatal(e)
					}
					result, e := client.Logout(context.Background())
					codes[mode] = append(codes[mode], ipgw.CodeOf(e))
					outcomes[mode] = append(outcomes[mode], result.Outcome)
					if mode == 1 {
						active.result(e)
						if active.Snapshot().ValidateVersion(2) != nil {
							t.Fatal("invalid SDK projection")
						}
						if step == 0 {
							first = active.Snapshot()
						}
					}
				}
			}
			if !reflect.DeepEqual(traces[0], traces[1]) || !reflect.DeepEqual(codes[0], codes[1]) || !reflect.DeepEqual(outcomes[0], outcomes[1]) {
				t.Fatal("SDK/cleanup changed")
			}
			if first.SDKCategory != tc.category {
				t.Fatal("wrong SDK classification")
			}
			if tc.offline && outcomes[1][1] != ipgw.LogoutAlreadyOffline {
				t.Fatal("cleanup idempotence changed")
			}
		})
	}
}
