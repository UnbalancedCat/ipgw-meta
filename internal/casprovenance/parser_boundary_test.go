package casprovenance

import (
	"context"
	"reflect"
	"testing"
)

// These are generated synthetic inputs, not reconstructions of a live script.
// Mapping a text match and classifying its lexical context are independent.
func TestParserMatchBoundaryClassification(t *testing.T) {
	key, _ := generated(t)
	a := assign(key)
	cases := []struct {
		name, input, lexical, location, assignment string
		keys                                       int
	}{
		{"division_before", "var ratio=6/2;" + a, "partial", "unknown", "unknown", 0},
		{"regexp_before", "var pattern=/x/;" + a, "partial", "unknown", "unknown", 0},
		{"unicode_before", "var café='x';" + a, "partial", "unknown", "unknown", 0},
		{"escaped_identifier_before", `var ca\u0066e='x';` + a, "partial", "unknown", "unknown", 0},
		{"template_interpolation_before", "var message=`${name}`;" + a, "partial", "unknown", "unknown", 0},
		{"unicode_line_separator_before", "// comment\u2028" + a, "partial", "unknown", "unknown", 0},
		{"function_only", "function f(){" + a + "}", "complete", "code", "unknown", 0},
		{"numeric_statement_before", "var count=1;" + a, "complete", "code", "unknown", 0},
		{"division_after", a + "var ratio=6/2;", "partial", "code", "bare", 1},
		{"comment_after_division", "var ratio=6/2; /* " + a + " */", "partial", "unknown", "unknown", 0},
		{"string_after_division", "var ratio=6/2; var text=\"" + a + "\";", "partial", "unknown", "unknown", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, m, err := AnalyzeParserMatch(context.Background(), []Input{{"external_script", []byte(tc.input)}}, 0)
			if err != nil || ValidateReport(r) != nil || m.Validate(&r) != nil || m.State != "mapped" || len(r.Occurrences) != 1 {
				t.Fatal("synthetic match not mapped or report invalid")
			}
			o := r.Occurrences[m.Occurrence-1]
			if r.Coverage.Lexical != tc.lexical || o.Context != tc.location || o.AssignmentClass != tc.assignment || len(r.Keys) != tc.keys {
				t.Fatal("lexical suffix and assignment boundary conflated")
			}
			if r.Coverage.UseRelation != "unknown" || len(r.Relations) != 0 {
				t.Fatal("mapping inferred runtime use")
			}
		})
	}
}

func TestParserMatchEarlierSourceDoesNotPoisonLaterSource(t *testing.T) {
	key, _ := generated(t)
	inputs := []Input{
		{"page", []byte(`<script src="/synthetic.js"></script>`)},
		{"external_script", []byte("var pattern=/x/;")},
		{"external_script", []byte(assign(key))},
	}
	r, m, err := AnalyzeParserMatch(context.Background(), inputs, 2)
	if err != nil || ValidateReport(r) != nil || m.Validate(&r) != nil || m.State != "mapped" || len(r.Sources) != 4 {
		t.Fatal("source isolation setup failed")
	}
	o := r.Occurrences[m.Occurrence-1]
	if r.Coverage.Lexical != "partial" || r.Sources[1].Kind != "unknown_script" || r.Sources[3].Analysis != "complete" || o.Context != "code" || o.AssignmentClass != "bare" || o.KeyClass != "valid_rsa" {
		t.Fatal("earlier partial source contaminated a later source")
	}
}

func TestParserMatchClosedReportCannotIdentifyUnknownCause(t *testing.T) {
	key, _ := generated(t)
	a := assign(key)
	// Distinct causes and even distinct true lexical locations collapse to the
	// same closed observation. Do not infer a real cause from this projection.
	inputs := []string{
		"var ratio=6/2;" + a,
		"var pattern=/x/;" + a,
		"var café='x';" + a,
		"var ratio=6/2; /* " + a + " */",
		"var ratio=6/2; var text=\"" + a + "\";",
	}
	var first Report
	var firstMatch ParserMatch
	for i, input := range inputs {
		r, m, err := AnalyzeParserMatch(context.Background(), []Input{{"external_script", []byte(input)}}, 0)
		if err != nil || ValidateReport(r) != nil || m.State != "mapped" || m.Validate(&r) != nil {
			t.Fatal("closed projection rejected")
		}
		if i == 0 {
			first, firstMatch = r, m
		} else if !reflect.DeepEqual(first, r) || firstMatch != m {
			t.Fatal("expected indistinguishable closed observations")
		}
	}
}
