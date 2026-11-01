package casprovenance

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestStrictReportDecoder(t *testing.T) {
	a, _ := generated(t)
	valid, err := EncodeReport(analyze(t, assign(a)))
	if err != nil {
		t.Fatal(err)
	}
	for name, input := range map[string][]byte{
		"unknown":            bytes.Replace(valid, []byte(`"schema_version":1`), []byte(`"schema_version":1,"unexpected":"untrusted"`), 1),
		"duplicate":          bytes.Replace(valid, []byte(`"schema_version":1`), []byte(`"schema_version":1,"schema_version":1`), 1),
		"nested-duplicate":   bytes.Replace(valid, []byte(`"dns":0`), []byte(`"dns":0,"dns":0`), 1),
		"missing-zero-field": bytes.Replace(valid, []byte(`"dns":0,`), nil, 1),
		"missing-null-field": bytes.Replace(valid, []byte(`"bundle_id":null,`), nil, 1),
		"null-array":         bytes.Replace(valid, []byte(`"relations":[]`), []byte(`"relations":null`), 1),
		"null-number":        bytes.Replace(valid, []byte(`"dns":0`), []byte(`"dns":null`), 1),
		"forged-binding":     bytes.Replace(valid, []byte(`"not_packaged"`), []byte(`"verified"`), 1),
		"operation":          bytes.Replace(valid, []byte(`"dns":0`), []byte(`"dns":1`), 1),
		"bad-reference":      bytes.Replace(valid, []byte(`"key_ref":"K1"`), []byte(`"key_ref":"K9"`), 1),
		"bad-source":         bytes.Replace(valid, []byte(`"source_id":"S1"`), []byte(`"source_id":"S9"`), 1),
		"wrong-reason":       bytes.Replace(valid, []byte(`"use_relation_unknown"`), []byte(`"unsupported_syntax"`), 1),
		"trailing":           append(append([]byte{}, valid...), []byte(`{}`)...),
		"utf8":               append(append([]byte{}, valid...), 0xff),
		"oversize":           []byte(strings.Repeat(" ", MaxReportBytes+1)),
	} {
		t.Run(name, func(t *testing.T) {
			if bytes.Equal(input, valid) {
				t.Fatal("fixture did not mutate input")
			}
			if _, err := DecodeReport(input); !errors.Is(err, ErrReport) || err.Error() != "invalid provenance report" {
				t.Fatal("accepted malformed projection or leaked parser error")
			}
		})
	}
}

func TestEncodeRejectsInconsistentFacts(t *testing.T) {
	a, _ := generated(t)
	for _, mutate := range []func(*Report){
		func(r *Report) { r.Keys[0].ID = "arbitrary" },
		func(r *Report) { r.Occurrences[0].Context = "comment" },
		func(r *Report) { r.Occurrences[0].KeyRef = nil },
		func(r *Report) { r.Occurrences[0].AssignmentClass = "none" },
		func(r *Report) { r.Keys = append(r.Keys, Key{"K2", 1024, "common"}) },
		func(r *Report) { r.Relations = append(r.Relations, Relation{1, 1, "direct"}) },
		func(r *Report) { r.Analysis = "complete" },
		func(r *Report) { r.Coverage.UseRelation = "complete" },
	} {
		r := analyze(t, assign(a))
		mutate(&r)
		if _, err := EncodeReport(r); !errors.Is(err, ErrReport) {
			t.Fatal("accepted inconsistent facts")
		}
	}
}

func FuzzReportDecoder(f *testing.F) {
	r := emptyReport()
	r.Sources = append(r.Sources, Source{"S1", "inline_script", 1, "complete"})
	valid, _ := json.Marshal(r)
	f.Add(string(valid))
	f.Add(`{"schema_version":1}`)
	f.Add(`{"x":1,"x":2}`)
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > MaxReportBytes+1 {
			t.Skip()
		}
		r, err := DecodeReport([]byte(input))
		if err != nil {
			if err != ErrReport {
				t.Fatal("nonconstant error")
			}
			return
		}
		if _, err := EncodeReport(r); err != nil {
			t.Fatal("decoded invalid report")
		}
	})
}
