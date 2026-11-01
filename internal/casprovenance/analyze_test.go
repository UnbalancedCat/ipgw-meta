package casprovenance

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"
)

func generated(t *testing.T) (string, string) {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal("synthetic key generation failed")
	}
	der, err := x509.MarshalPKIXPublicKey(&k.PublicKey)
	if err != nil {
		t.Fatal("synthetic key encoding failed")
	}
	return base64.StdEncoding.EncodeToString(der), base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PublicKey(&k.PublicKey))
}
func assign(key string) string { return "publicKeyStr='" + key + "';" }
func analyze(t *testing.T, source string) Report {
	t.Helper()
	r, err := Analyze(context.Background(), []Input{{"inline_script", []byte(source)}})
	if err != nil {
		t.Fatal("analysis failed:", err)
	}
	if err = ValidateReport(r); err != nil {
		t.Fatal("invalid produced report")
	}
	return r
}

func TestEquivalenceOrderAndRoundLabels(t *testing.T) {
	a, a1 := generated(t)
	b, _ := generated(t)
	for _, source := range []string{assign(a) + assign(a1) + assign(b), assign(b) + assign(b) + assign(a)} {
		r := analyze(t, source)
		if len(r.Keys) != 2 || len(r.Occurrences) != 3 {
			t.Fatal("equivalence count")
		}
		for i, want := range []string{"K1", "K1", "K2"} {
			if *r.Occurrences[i].KeyRef != want {
				t.Fatal("round-local equivalence")
			}
		}
		if r.Analysis != "partial" || r.Coverage.UseRelation != "unknown" || len(r.Relations) != 0 {
			t.Fatal("must not infer key use")
		}
	}
	if *analyze(t, assign(b)).Occurrences[0].KeyRef != "K1" {
		t.Fatal("labels must reset")
	}
}

func TestLexicalAndAssignmentBoundaries(t *testing.T) {
	a, _ := generated(t)
	cases := []struct {
		name, source, context, identifier, assignment, reason string
		keys                                                  int
	}{
		{"bare", assign(a), "code", "public_key_str", "bare", "use_relation_unknown", 1},
		{"declaration", "const " + assign(a), "code", "public_key_str", "declaration", "use_relation_unknown", 1},
		{"property", "holder." + assign(a), "code", "public_key_str", "property", "use_relation_unknown", 1},
		{"generic", strings.Replace(assign(a), "publicKeyStr", "publicKey", 1), "code", "public_key", "bare", "use_relation_unknown", 1},
		{"prefix", "not" + assign(a), "code", "other", "none", "use_relation_unknown", 0},
		{"case", strings.Replace(assign(a), "publicKeyStr", "PUBLICKEYSTR", 1), "code", "other", "none", "use_relation_unknown", 0},
		{"line-comment", "// " + assign(a), "comment", "unknown", "none", "use_relation_unknown", 0},
		{"unicode-line-end", "// note\u2028" + assign(a), "unknown", "unknown", "unknown", "unsupported_syntax", 0},
		{"block-comment", "/* " + assign(a) + " */", "comment", "unknown", "none", "use_relation_unknown", 0},
		{"string", "var text=\"" + assign(a) + "\";", "string", "unknown", "none", "use_relation_unknown", 0},
		{"template", "`" + assign(a) + "`", "template", "unknown", "none", "unsupported_syntax", 0},
		{"function", "function f(){" + assign(a) + "}", "code", "public_key_str", "unknown", "unsupported_syntax", 0},
		{"shadowed", "{let " + assign(a) + "}", "code", "public_key_str", "unknown", "unsupported_syntax", 0},
		{"regexp", "/" + assign(a) + "/", "unknown", "unknown", "unknown", "unsupported_syntax", 0},
		{"interpolation", "`${" + assign(a) + "}`", "unknown", "unknown", "unknown", "unsupported_syntax", 0},
		{"unicode", "é" + assign(a), "unknown", "unknown", "unknown", "unsupported_syntax", 0},
		{"expression-suffix", strings.TrimSuffix(assign(a), ";") + "/2", "code", "public_key_str", "unknown", "unsupported_syntax", 0},
		{"escaped-value", "publicKeyStr='\\x41';", "code", "public_key_str", "unknown", "unsupported_syntax", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := analyze(t, tc.source)
			if len(r.Keys) != tc.keys || len(r.Occurrences) != 1 || r.Reason != tc.reason {
				t.Fatal("wrong counts or coverage")
			}
			o := r.Occurrences[0]
			if o.Context != tc.context || o.IdentifierClass != tc.identifier || o.AssignmentClass != tc.assignment {
				t.Fatal("incorrect context classification")
			}
		})
	}
	for _, source := range []string{"publicKeyStrExtra='AA==';", "public\\u004beyStr='AA==';"} {
		r := analyze(t, source)
		if len(r.Keys) != 0 || len(r.Occurrences) != 0 {
			t.Fatal("invented exact identifier")
		}
	}
	r := analyze(t, assign(a)+"function f(){"+assign(a)+"}")
	if len(r.Keys) != 1 || r.Occurrences[1].AssignmentClass != "unknown" || r.Reason != "unsupported_syntax" {
		t.Fatal("prefix fact or unsupported suffix lost")
	}
}

func TestHTMLSources(t *testing.T) {
	a, _ := generated(t)
	b, _ := generated(t)
	r, err := Analyze(context.Background(), []Input{{"page", []byte("<div data-x=\"" + assign(a) + "\"></div><script>" + assign(a) + "</script><script type='module'>" + assign(b) + "</script>")}, {"external_script", []byte(assign(b))}})
	if err != nil || ValidateReport(r) != nil {
		t.Fatal("page analysis failed")
	}
	if len(r.Sources) != 4 || len(r.Occurrences) != 4 || len(r.Keys) != 2 {
		t.Fatal("source separation failed")
	}
	for i, want := range []string{"html", "code", "unknown", "code"} {
		if r.Occurrences[i].Context != want {
			t.Fatal("source context mismatch")
		}
	}
	for _, attrs := range []string{"src='https://example.invalid/private'", "async", "defer", "type='unknown'", "language='vbscript'", "nomodule", "nonce='synthetic'"} {
		r, err = Analyze(context.Background(), []Input{{"page", []byte("<script " + attrs + ">" + assign(a) + "</script>")}})
		if err != nil || ValidateReport(r) != nil || len(r.Keys) != 0 || r.Sources[1].Kind != "unknown_script" {
			t.Fatal("unsupported script promoted")
		}
	}
}

func TestRejectedInputs(t *testing.T) {
	for _, in := range [][]Input{nil, {{"inline_script", []byte{0xff}}}, {{"other", nil}}, {{"page", nil}, {"page", nil}}, {{"page", []byte("<script>unfinished")}}, {{"page", []byte("<script/>")}}} {
		r, err := Analyze(context.Background(), in)
		checkRejected(t, r, err, ErrInvalid)
	}
	for _, s := range []string{"'", "/*", "(", "[)", "'line\nbreak'", "'escape\\"} {
		r, err := Analyze(context.Background(), []Input{{"inline_script", []byte(s)}})
		checkRejected(t, r, err, ErrInvalid)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r, err := Analyze(ctx, []Input{{"inline_script", nil}})
	checkRejected(t, r, err, context.Canceled)
	ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	r, err = Analyze(ctx, []Input{{"inline_script", nil}})
	checkRejected(t, r, err, context.DeadlineExceeded)
}

// Cancellation is triggered from inside scanning, deterministically, rather
// than relying on scheduler timing or only testing an already-canceled input.
type cancelDuringScan struct {
	context.Context
	calls  int
	cancel context.CancelFunc
}

func (c *cancelDuringScan) Err() error {
	c.calls++
	if c.calls == 20 {
		c.cancel()
	}
	return c.Context.Err()
}
func TestCancellationDuringScan(t *testing.T) {
	base, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx := &cancelDuringScan{Context: base, cancel: cancel}
	r, err := Analyze(ctx, []Input{{"inline_script", []byte(strings.Repeat(";", 10000))}})
	checkRejected(t, r, err, context.Canceled)
	if ctx.calls < 20 {
		t.Fatal("did not enter scanner")
	}
}

func checkRejected(t *testing.T, r Report, err, want error) {
	t.Helper()
	if !errors.Is(err, want) || r.Analysis != "rejected" || len(r.Keys)+len(r.Sources)+len(r.Occurrences) != 0 || ValidateReport(r) != nil {
		t.Fatal("rejection must clear partial results:", err)
	}
}

func TestResourceBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name     string
		at, over []Input
	}{
		{"bytes", []Input{{"inline_script", bytes.Repeat([]byte(" "), MaxSourceBytes)}}, []Input{{"inline_script", bytes.Repeat([]byte(" "), MaxSourceBytes+1)}}},
		{"occurrences", []Input{{"inline_script", []byte(strings.Repeat("// publicKey='AA=='\n", MaxOccurrences))}}, []Input{{"inline_script", []byte(strings.Repeat("// publicKey='AA=='\n", MaxOccurrences+1))}}},
		{"tokens", []Input{{"inline_script", []byte(strings.Repeat(";", MaxTokens))}}, []Input{{"inline_script", []byte(strings.Repeat(";", MaxTokens+1))}}},
		{"depth", []Input{{"inline_script", []byte(strings.Repeat("(", MaxDepth) + strings.Repeat(")", MaxDepth))}}, []Input{{"inline_script", []byte(strings.Repeat("(", MaxDepth+1) + strings.Repeat(")", MaxDepth+1))}}},
		{"html-depth", []Input{{"page", []byte(strings.Repeat("<div>", MaxDepth) + strings.Repeat("</div>", MaxDepth))}}, []Input{{"page", []byte(strings.Repeat("<div>", MaxDepth+1) + strings.Repeat("</div>", MaxDepth+1))}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := Analyze(context.Background(), tc.at)
			if err != nil || ValidateReport(r) != nil {
				t.Fatal("boundary rejected:", err)
			}
			r, err = Analyze(context.Background(), tc.over)
			checkRejected(t, r, err, ErrLimit)
		})
	}
	inputs := make([]Input, MaxSources+1)
	for i := range inputs {
		inputs[i].Kind = "external_script"
	}
	r, err := Analyze(context.Background(), inputs[:MaxSources])
	checkRejected(t, r, err, ErrLimit)
	inputs[0].Kind = "page"
	r, err = Analyze(context.Background(), inputs[:MaxSources])
	if err != nil || ValidateReport(r) != nil {
		t.Fatal("source boundary")
	}
	r, err = Analyze(context.Background(), inputs)
	checkRejected(t, r, err, ErrLimit)
	for _, n := range []int{MaxSources - 1, MaxSources} {
		r, err = Analyze(context.Background(), []Input{{"page", []byte(strings.Repeat("<script></script>", n))}})
		if n == MaxSources {
			checkRejected(t, r, err, ErrLimit)
		} else if err != nil || ValidateReport(r) != nil {
			t.Fatal("extracted source boundary")
		}
	}
	r, err = Analyze(context.Background(), []Input{{"inline_script", []byte(strings.Repeat("// publicKey='AA=='\n", 32))}, {"external_script", []byte(strings.Repeat("// publicKey='AA=='\n", 33))}})
	checkRejected(t, r, err, ErrLimit)
	r, err = Analyze(context.Background(), []Input{{"inline_script", []byte(strings.Repeat(";", MaxTokens/2))}, {"external_script", []byte(strings.Repeat(";", MaxTokens/2+1))}})
	checkRejected(t, r, err, ErrLimit)
	inputs = make([]Input, 4)
	for i := range inputs {
		inputs[i] = Input{"external_script", bytes.Repeat([]byte(" "), MaxSourceBytes)}
	}
	r, err = Analyze(context.Background(), inputs)
	if err != nil || ValidateReport(r) != nil {
		t.Fatal("total boundary")
	}
	r, err = Analyze(context.Background(), append(inputs, Input{"external_script", []byte(" ")}))
	checkRejected(t, r, err, ErrLimit)
	// Extracted script slices also count against the normalized memory limit.
	inputs[0] = Input{"page", []byte("<script>" + strings.Repeat(" ", MaxSourceBytes-17) + "</script>")}
	r, err = Analyze(context.Background(), inputs)
	checkRejected(t, r, err, ErrLimit)
}

func TestRSAEnvelope(t *testing.T) {
	for _, tc := range []struct {
		bits, exponent int
		valid          bool
	}{{511, 65537, false}, {512, 3, true}, {8192, 65537, true}, {8193, 65537, false}, {1024, 2, false}, {1024, 2147483647, true}} {
		n := new(big.Int).Lsh(big.NewInt(1), uint(tc.bits-1))
		n.SetBit(n, 0, 1)
		value := base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PublicKey(&rsa.PublicKey{N: n, E: tc.exponent}))
		k, class := parseKey(value)
		if (k != nil) != tc.valid || (class == "valid_rsa") != tc.valid {
			t.Fatal("RSA boundary mismatch")
		}
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal("synthetic generation")
	}
	der, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	for _, tc := range []struct{ value, class string }{{"!", "invalid_encoding"}, {"AA==", "invalid_encoding"}, {strings.Repeat("A", 8193), "out_of_bounds"}, {base64.StdEncoding.EncodeToString(der), "non_rsa"}} {
		k, class := parseKey(tc.value)
		if k != nil || class != tc.class {
			t.Fatal("invalid key classification")
		}
	}
}

func TestClosedProjectionDoesNotEchoInput(t *testing.T) {
	a, _ := generated(t)
	canary := "synthetic-" + "untrusted-source-canary"
	r := analyze(t, "var "+strings.ReplaceAll(canary, "-", "_")+"='https://example.invalid/"+canary+"';"+assign(a))
	encoded, err := EncodeReport(r)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256([]byte(a))
	for _, forbidden := range []string{a, canary, "https://", base64.StdEncoding.EncodeToString([]byte(a)), hex.EncodeToString(hash[:])} {
		if bytes.Contains(encoded, []byte(forbidden)) {
			t.Fatal("input leaked into report")
		}
	}
	decoded, err := DecodeReport(encoded)
	if err != nil || !reflect.DeepEqual(r, decoded) {
		t.Fatal("roundtrip failed")
	}
}

func FuzzBoundedAnalysis(f *testing.F) {
	for _, seed := range []string{"publicKeyStr='AA==';", "/*", "`${x}`", "<script>publicKey='AA=='</script>", "'\\\"'"} {
		f.Add(seed, false)
		f.Add(seed, true)
	}
	f.Fuzz(func(t *testing.T, source string, page bool) {
		if len(source) > 65536 {
			t.Skip()
		}
		kind := "inline_script"
		if page {
			kind = "page"
		}
		r, _ := Analyze(context.Background(), []Input{{kind, []byte(source)}})
		data, err := EncodeReport(r)
		if err != nil {
			t.Fatal("invalid generated report")
		}
		if _, err = DecodeReport(data); err != nil {
			t.Fatal("report cannot roundtrip")
		}
	})
}
