package casprovenance

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestStaticContextMatrix(t *testing.T) {
	key, _ := generated(t)
	a := assign(key)
	cases := []struct {
		name, input, location, class, place string
		keys                                int
	}{
		{"division", "var ratio=6/2;" + a, "code", "bare", "top_level", 1},
		{"regexp", "var pattern=/x/;" + a, "code", "bare", "top_level", 1},
		{"unicode", "var café='x';" + a, "code", "bare", "top_level", 1},
		{"escape", `var ca\u0066e='x';` + a, "code", "bare", "top_level", 1},
		{"template", "var x=`${name}`;" + a, "code", "bare", "top_level", 1},
		{"unicode_comment", "// comment\u2028" + a, "code", "bare", "top_level", 1},
		{"function", "function f(){" + a + "}", "code", "bare", "function", 1},
		{"control", "if(false){" + a + "}", "code", "bare", "control", 1},
		{"short_circuit", "false && (" + strings.TrimSuffix(a, ";") + ");", "code", "bare", "control", 1},
		{"nullish", "null ?? (" + strings.TrimSuffix(a, ";") + ");", "code", "bare", "control", 1},
		{"class_field", "class F { value=(" + strings.TrimSuffix(a, ";") + ");}", "code", "bare", "other", 1},
		{"nested_block", "{" + a + "}", "code", "bare", "other", 1},
		{"regexp_flag_then_comment", "var r=/x/g /* " + a + " */;", "comment", "none", "unknown", 0},
		{"regexp_flag_then_assignment", "var r=/x/g; " + a, "code", "bare", "top_level", 1},
		{"regexp_flag_then_newline", "var r=/x/g\n" + a, "code", "bare", "top_level", 1},
		{"comment", "var ratio=6/2; /* " + a + " */", "comment", "none", "unknown", 0},
		{"string", "var ratio=6/2; var x=\"" + a + "\";", "string", "none", "top_level", 0},
		{"template_text", "var x=`" + a + "${name}`;", "template", "none", "top_level", 0},
		{"template_crlf", "var x=`first\r\n" + a + "${name}`;", "template", "none", "top_level", 0},
		{"unicode_substring", "é" + a, "code", "none", "top_level", 0},
		{"escaped_substring", `ca\u0066e` + a, "code", "none", "top_level", 0},
		{"arrow_comment", "var fn=(x /* " + a + " */)=>x/2;", "comment", "none", "unknown", 0},
		{"property", "object." + a, "code", "property", "top_level", 1},
		{"substring", "not" + a, "code", "none", "top_level", 0},
		{"escaped_identifier", `public\u004beyStr='` + key + `';` + a, "code", "bare", "top_level", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := AnalyzeStaticParserMatch(context.Background(), []Input{{"external_script", []byte(tc.input)}}, 0)
			if r.Validate() != nil || r.State != "observed" {
				t.Fatalf("static rejected: %s", r.Reason)
			}
			o := r.Occurrences[r.ParserMatch.Occurrence-1]
			if o.Context != tc.location || o.AssignmentClass != tc.class || o.Placement != tc.place || len(r.Keys) != tc.keys {
				t.Fatalf("classification mismatch: %s %s %s keys %d", o.Context, o.AssignmentClass, o.Placement, len(r.Keys))
			}
			b, e := json.Marshal(r)
			if e != nil || strings.Contains(string(b), key) || strings.Contains(string(b), "café") {
				t.Fatal("material leaked")
			}
		})
	}
}

func FuzzStaticAnalysis(f *testing.F) {
	f.Add([]byte("var ratio=6/2;publicKeyStr='" + strings.Repeat("A", 160) + "';"))
	f.Add([]byte("var f=(x)=>`${x}`;/*publicKeyStr='" + strings.Repeat("A", 160) + "';*/"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 32768 {
			return
		}
		r := AnalyzeStaticParserMatch(context.Background(), []Input{{"external_script", data}}, 0)
		if r.Validate() != nil {
			t.Fatal("invalid projection")
		}
	})
}
func TestStaticKeyRelationsAndSources(t *testing.T) {
	a, a2 := generated(t)
	b, _ := generated(t)
	for _, values := range [][]string{{a, a2, b}, {b, b, a}} {
		r := AnalyzeStaticParserMatch(context.Background(), []Input{{"page", []byte(`<script src="/synthetic"></script>`)}, {"external_script", []byte("function f(){" + assign(values[0]) + assign(values[1]) + assign(values[2]) + "}")}}, 1)
		if r.State != "observed" || r.Validate() != nil || len(r.Keys) != 2 || len(r.Occurrences) != 3 {
			t.Fatalf("relation missing: %s", r.Reason)
		}
		for i, want := range []string{"K1", "K1", "K2"} {
			if r.Occurrences[i].KeyRef == nil || *r.Occurrences[i].KeyRef != want {
				t.Fatal("wrong equality")
			}
		}
	}
}
func TestStaticRejectsAndClears(t *testing.T) {
	key, _ := generated(t)
	for _, input := range []string{assign(key) + "/*", strings.Repeat("(", 65) + "1" + strings.Repeat(")", 65) + ";" + assign(key), strings.Repeat(";", 100001) + assign(key)} {
		r := AnalyzeStaticParserMatch(context.Background(), []Input{{"external_script", []byte(input)}}, 0)
		if r.Validate() != nil || r.State != "rejected" || len(r.Occurrences) != 0 {
			t.Fatal("partial rejection")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := AnalyzeStaticParserMatch(ctx, []Input{{"external_script", []byte(assign(key))}}, 0)
	if r.Reason != "canceled" || r.Validate() != nil {
		t.Fatal("cancellation")
	}
}
