package authcontext

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
	"github.com/UnbalancedCat/ipgw-meta/internal/castest"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func generatedKey(t *testing.T) string {
	t.Helper()
	k, e := rsa.GenerateKey(rand.Reader, 1024)
	if e != nil {
		t.Fatal(e)
	}
	d, e := x509.MarshalPKIXPublicKey(&k.PublicKey)
	if e != nil {
		t.Fatal(e)
	}
	return base64.StdEncoding.EncodeToString(d)
}

type readTrace struct {
	N   int
	End string
}
type tracedBody struct {
	r      io.Reader
	reads  *[]readTrace
	closes *int
	fail   bool
}

func (b *tracedBody) Read(p []byte) (int, error) {
	n, e := b.r.Read(p)
	if b.fail {
		e = errors.New("synthetic-private-body-error")
	}
	label := "none"
	if e == io.EOF {
		label = "eof"
	} else if e != nil {
		label = "error"
	}
	*b.reads = append(*b.reads, readTrace{n, label})
	return n, e
}
func (b *tracedBody) Close() error { *b.closes++; return nil }

// ADR-0021 retires the live legacy collector. Keep its historical lexical
// classifications as direct material-analysis regressions, not new-SDK claims.
func TestHistoricalContextClassifications(t *testing.T) {
	a, b := generatedKey(t), generatedKey(t)
	for _, tc := range []struct {
		name, text, location, context, identifier, assignment string
		keys                                                  int
	}{
		{"inline", `var publicKeyStr="` + a + `";`, "inline", "code", "public_key_str", "declaration", 1},
		{"external", `var publicKeyStr="` + a + `";`, "external", "code", "public_key_str", "declaration", 1},
		{"comment", `/* publicKeyStr="` + a + `"; */ var publicKeyStr="` + b + `";`, "external", "comment", "unknown", "none", 1},
		{"string", `var decoy='publicKeyStr="` + a + `";'; var publicKeyStr="` + b + `";`, "external", "string", "unknown", "none", 1},
		{"substring", `var unusedpublicKeyStr="` + a + `"; var publicKeyStr="` + b + `";`, "external", "code", "other", "none", 1},
		{"duplicate", `var publicKeyStr="` + a + `"; publicKeyStr="` + a + `";`, "external", "code", "public_key_str", "declaration", 1},
		{"different", `var publicKeyStr="` + a + `"; publicKeyStr="` + b + `";`, "external", "code", "public_key_str", "declaration", 2},
		{"swapped", `var publicKeyStr="` + b + `"; publicKeyStr="` + a + `";`, "external", "code", "public_key_str", "declaration", 2},
		{"unsupported", `function unknown(){ } var publicKeyStr="` + a + `";`, "external", "code", "public_key_str", "unknown", 0},
	} {

		t.Run(tc.name, func(t *testing.T) {
			page := `<form id="loginForm"><input name="lt" value="synthetic-state"><input name="execution" value="synthetic-execution"></form>`
			inputs := []casprovenance.Input{{Kind: "page", Data: []byte(page)}}
			match := 0
			if tc.location == "inline" {
				inputs[0].Data = []byte(page + `<script>` + tc.text + `</script>`)
			} else {
				inputs = append(inputs, casprovenance.Input{Kind: "external_script", Data: []byte(tc.text)})
				match = 1
			}
			report, m, err := casprovenance.AnalyzeParserMatch(context.Background(), inputs, match)
			if err != nil || m.State != "mapped" || len(report.Keys) != tc.keys {
				t.Fatal("historical classification failed")
			}
			o := report.Occurrences[m.Occurrence-1]
			if o.Context != tc.context || o.IdentifierClass != tc.identifier || o.AssignmentClass != tc.assignment {
				t.Fatal("historical lexical meaning changed")
			}
			data, err := json.Marshal(report)
			if err != nil {
				t.Fatal(err)
			}
			for _, secret := range []string{a, b, "synthetic-state"} {
				if strings.Contains(string(data), secret) {
					t.Fatal("historical analysis leaked material")
				}
			}
		})
	}
}

func TestContextFormTargetsAndFailureCoverage(t *testing.T) {
	key := generatedKey(t)
	for _, scenario := range []string{"redirect", "uppercase-redirect", "whitespace-state", "relative", "absolute", "query", "absent-service", "different-service", "ambiguous-form", "invalid-state", "body-error", "oversize", "source-limit", "unknown-role", "early-ticket", "online", "cancel"} {
		t.Run(scenario, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls := 0
			tr := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				calls++
				res := syntheticResponse(req, 200, "")
				body := ""
				switch strings.ToLower(req.URL.Host) + req.URL.Path {
				case "ipgw.neu.edu.cn/cgi-bin/rad_user_info":
					body = "not_online_error"
					if scenario == "online" {
						body = `{"error":"ok","user_name":"synthetic-user","online_ip":"192.0.2.1"}`
					}
				case "ipgw.neu.edu.cn/":
					body = "ac_id=1"
				case "pass.neu.edu.cn/tpass/login", "pass.neu.edu.cn/tpass/redirected":
					if scenario == "early-ticket" {
						res.StatusCode = 302
						res.Header.Set("Location", req.URL.Query().Get("service")+"&ticket=synthetic-ticket")
						return res, nil
					}
					if (scenario == "redirect" || scenario == "uppercase-redirect") && req.URL.Path == "/tpass/login" {
						res.StatusCode = 302
						res.Header.Set("Location", "/tpass/redirected?"+req.URL.RawQuery)
						if scenario == "uppercase-redirect" {
							res.Header.Set("Location", "https://PASS.NEU.EDU.CN/tpass/redirected?"+req.URL.RawQuery)
						}
						res.Body = io.NopCloser(strings.NewReader(`var publicKeyStr="` + key + `";`))
						return res, nil
					}
					attr := ""
					switch scenario {
					case "relative":
						attr = ` action="submit?` + req.URL.RawQuery + `"`
					case "absolute":
						attr = ` action="https://pass.neu.edu.cn/tpass/submit?` + req.URL.RawQuery + `"`
					case "query":
						attr = ` action="?` + req.URL.RawQuery + `"`
					case "absent-service":
						attr = ` action="/tpass/submit"`
					case "different-service":
						attr = ` action="?service=synthetic-other"`
					}
					body = `<form id="loginForm"` + attr + `><input name="lt" value="synthetic-state"><input name="execution" value="synthetic-execution"></form><script>` + castest.LoginScript(key) + `</script>`
					if scenario == "ambiguous-form" {
						body += `<form id="loginForm" action="/decoy"></form>`
					}
					if scenario == "invalid-state" {
						body = strings.Replace(body, `name="lt"`, `name="missing"`, 1)
					}
					if scenario == "whitespace-state" {
						body = strings.Replace(body, `value="synthetic-state"`, `value=" "`, 1)
					}
					if scenario == "oversize" {
						body += strings.Repeat(" ", casprovenance.MaxSourceBytes)
					}
					if scenario == "source-limit" {
						body += strings.Repeat("<script></script>", 17)
					}
					if scenario == "unknown-role" {
						body = "no form"
					}
					if scenario == "cancel" {
						cancel()
					}
				default:
					t.Fatal("unexpected request")
				}
				if scenario == "body-error" && req.URL.Hostname() == "pass.neu.edu.cn" {
					dummy := []readTrace{}
					count := 0
					res.Body = &tracedBody{r: strings.NewReader(body), reads: &dummy, closes: &count, fail: true}
				} else {
					res.Body = io.NopCloser(strings.NewReader(body))
				}
				return res, nil
			})
			out := Anonymous(ctx, tr, func(context.Context) bool { return true })
			if out.Validate() != nil {
				t.Fatalf("report validation: %s/%s", out.State, out.Reason)
			}
			switch scenario {
			case "body-error", "oversize", "source-limit", "unknown-role", "early-ticket", "online", "cancel", "invalid-state":
				if out.State != "blocked" || out.Analysis != nil || out.ParserMatch.State != "unknown" {
					t.Fatal("ineligible source published")
				}
			default:
				if out.State != "observed" {
					t.Fatalf("unexpected blocked: %s", out.Reason)
				}
				initial, final, service, selection := "same", "same", "same", "unique"
				switch scenario {
				case "redirect", "uppercase-redirect":
					initial = "different"
				case "relative", "absolute":
					initial, final = "different", "different"
				case "absent-service":
					initial, final, service = "different", "different", "absent"
				case "different-service":
					initial, final, service = "different", "different", "different"
				case "ambiguous-form":
					selection = "ambiguous"
				}
				if out.Form.InitialTarget != initial || out.Form.FinalTarget != final || out.Form.Service != service || out.Form.Selection != selection {
					t.Fatal("incorrect target relations")
				}
			}
			if calls > 4 {
				t.Fatal("unexpected extra request")
			}
		})
	}
}

func TestContextTeePreservesReadsAndClose(t *testing.T) {
	for _, early := range []bool{false, true} {
		reads := []readTrace{}
		closes := 0
		c := &collector{ctx: context.Background(), matchInput: -1}
		base, _ := url.Parse("https://pass.neu.edu.cn/tpass/login")
		b := &bodyTee{ReadCloser: &tracedBody{r: strings.NewReader("not a page"), reads: &reads, closes: &closes}, c: c, base: base}
		if !early {
			_, _ = io.ReadAll(b)
		}
		_ = b.Close()
		if closes != 1 || early && len(reads) != 0 {
			t.Fatal("tee drained or changed close")
		}
		if c.failure == "" || len(c.inputs) != 0 {
			t.Fatal("incomplete/uncertain body retained")
		}
	}
}
func TestContextClosedReport(t *testing.T) {
	if SelfTest().State != "pass" {
		t.Fatal("self-test")
	}
	r := NewReport("preflight")
	r.State, r.Reason = "ready", "none"
	r.SDK.State, r.SDK.Reason = "ready", "none"
	good, e := Encode(r)
	if e != nil {
		t.Fatal(e)
	}
	for _, bad := range [][]byte{append(append([]byte(nil), good...), good...), []byte(strings.Replace(string(good), `"schema_version":1`, `"schema_version":2`, 1)), []byte(strings.Replace(string(good), `"schema_version":1`, `"schema_version":1,"unknown":true`, 1)), []byte(strings.Replace(string(good), `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1)), []byte(strings.Replace(string(good), `"analysis":null`, `"analysis":{}`, 1))} {
		if _, e := Decode(bad); e == nil {
			t.Fatal("invalid envelope accepted")
		}
	}
	r.ParserMatch = casprovenance.ParserMatch{State: "mapped", SourceID: "S1", Occurrence: 1}
	if _, e := Encode(r); e == nil {
		t.Fatal("mapping without analysis accepted")
	}
}

type oneRead struct {
	data   []byte
	err    error
	used   bool
	closes int
}

func (r *oneRead) Read(p []byte) (int, error) {
	if r.used {
		return 0, io.EOF
	}
	r.used = true
	return copy(p, r.data), r.err
}
func (r *oneRead) Close() error { r.closes++; return nil }
func TestTeeReturnsOriginalReadEvenOnObservationFailure(t *testing.T) {
	for _, scenario := range []string{"eof", "error", "cancel", "overflow", "cumulative"} {
		t.Run(scenario, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			data := []byte("synthetic-body")
			end := error(io.EOF)
			if scenario == "error" {
				end = errors.New("synthetic-private-read-error")
			}
			c := &collector{ctx: ctx, matchInput: -1}
			if scenario == "cancel" {
				cancel()
			}
			if scenario == "overflow" {
				data = bytes.Repeat([]byte(" "), casprovenance.MaxSourceBytes+1)
			}
			if scenario == "cumulative" {
				c.total = casprovenance.MaxTotalBytes - 1
				c.inputs = []casprovenance.Input{{Kind: "page", Data: []byte("retained")}}
			}
			under := &oneRead{data: data, err: end}
			base, _ := url.Parse("https://pass.neu.edu.cn/tpass/login")
			tee := &bodyTee{ReadCloser: under, c: c, base: base}
			p := make([]byte, len(data))
			n, e := tee.Read(p)
			if n != len(data) || e != end || !bytes.Equal(p, data) {
				t.Fatal("tee changed n/error or original bytes")
			}
			if tee.Close() != nil || under.closes != 1 || c.failure == "" || len(c.inputs) != 0 || len(tee.data) != 0 {
				t.Fatal("failed observation retained input or changed close")
			}
		})
	}
}
func FuzzContextReport(f *testing.F) {
	r := NewReport("preflight")
	d, _ := json.Marshal(r)
	f.Add(d)
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 65536 {
			return
		}
		r, e := Decode(b)
		if e == nil {
			d, e := Encode(r)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = Decode(d); e != nil {
				t.Fatal(e)
			}
		}
	})
}
