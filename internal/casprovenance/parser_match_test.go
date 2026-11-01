package casprovenance

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestParserMatchDoesNotChangeHistoricalAnalysis(t *testing.T) {
	key, _ := generated(t)
	for _, data := range []string{`<div data-test='publicKeyStr="` + key + `"'></div><script>` + assign(key) + `</script>`, `<script>// ` + assign(key) + "\n" + assign(key) + `</script>`, `<script type="module">` + assign(key) + `</script>`} {
		inputs := []Input{{"page", []byte(data)}}
		old, e := Analyze(context.Background(), inputs)
		if e != nil {
			t.Fatal(e)
		}
		next, m, e := AnalyzeParserMatch(context.Background(), inputs, 0)
		if e != nil || !reflect.DeepEqual(old, next) || m.State != "mapped" || m.Validate(&next) != nil {
			t.Fatal("mapping changed old report or lost source position")
		}
		if next.Occurrences[m.Occurrence-1].SourceID != m.SourceID {
			t.Fatal("reference mismatch")
		}
	}
}
func TestParserMatchRejectsIncompleteOrUnmappable(t *testing.T) {
	for _, data := range []string{`var publicKeyStr="short";`, `var publicKeyStr="` + strings.Repeat("A", 8193) + `";`, `/* ` + assign(strings.Repeat("A", 160))} {
		_, m, _ := AnalyzeParserMatch(context.Background(), []Input{{"external_script", []byte(data)}}, 0)
		if m.State != "unknown" {
			t.Fatal("bad parser match published")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, m, e := AnalyzeParserMatch(ctx, []Input{{"external_script", []byte(assign(strings.Repeat("A", 160)))}}, 0)
	if e == nil || m.State != "unknown" {
		t.Fatal("canceled match published")
	}
}
func FuzzParserMatch(f *testing.F) {
	f.Add([]byte(`var publicKeyStr="short";`), false)
	f.Add([]byte(`<script>var publicKeyStr="short";</script>`), true)
	f.Fuzz(func(t *testing.T, b []byte, page bool) {
		if len(b) > 32768 {
			return
		}
		kind := "external_script"
		if page {
			kind = "page"
		}
		inputs := []Input{{kind, b}}
		r, m, e := AnalyzeParserMatch(context.Background(), inputs, 0)
		if e == nil {
			if ValidateReport(r) != nil || m.Validate(&r) != nil {
				t.Fatal("invalid analysis mapping")
			}
			old, e := Analyze(context.Background(), inputs)
			if e != nil || !reflect.DeepEqual(r, old) {
				t.Fatal("mapping changed analysis")
			}
		} else if m.State != "unknown" {
			t.Fatal("failed analysis mapped")
		}
	})
}
