package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
)

func TestRejectInputArguments(t *testing.T) {
	for _, args := range [][]string{nil, {"--live"}, {"--self-test", "input"}, {"--self-test=true"}, {"https://example.invalid"}} {
		var out, diagnostic bytes.Buffer
		if run(args, &out, &diagnostic) != 2 || out.Len() != 0 || diagnostic.String() != "only --self-test is supported\n" {
			t.Fatal("argument rejection contract")
		}
	}
}

func TestSyntheticSelfTest(t *testing.T) {
	var out, diagnostic bytes.Buffer
	if run([]string{"--self-test"}, &out, &diagnostic) != 0 || diagnostic.Len() != 0 {
		t.Fatal("self-test failed")
	}
	var result envelope
	decoder := json.NewDecoder(&out)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		t.Fatal("invalid envelope")
	}
	if result.Checks != 11 || result.SelfTestResult != "passed" || result.Binding != "not_packaged" || result.Mode != "synthetic-prototype" || result.ToolKind != "cas-provenance-dev" {
		t.Fatal("invalid self-test identity")
	}
	if casprovenance.ValidateReport(result.Report) != nil {
		t.Fatal("invalid report")
	}
}
