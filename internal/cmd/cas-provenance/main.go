// cas-provenance is an internal, synthetic-only prototype, not a live diagnostic.
package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
)

type envelope = casprovenance.SelfTestEnvelope

func run(args []string, out, diagnostic io.Writer) int {
	if len(args) != 1 || args[0] != "--self-test" {
		fmt.Fprintln(diagnostic, "only --self-test is supported")
		return 2
	}
	result, err := selfTest()
	if err != nil {
		fmt.Fprintln(diagnostic, "synthetic self-test failed")
		return 1
	}
	data, err := json.Marshal(result)
	if err != nil || len(data)+1 > casprovenance.MaxReportBytes {
		fmt.Fprintln(diagnostic, "synthetic report rejected")
		return 1
	}
	if _, err = out.Write(append(data, '\n')); err != nil {
		return 1
	}
	return 0
}

func selfTest() (envelope, error) {
	fail := func() (envelope, error) { return envelope{}, casprovenance.ErrReport }
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	keys := make([]string, 2)
	var pkcs1 string
	for i := range keys {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return fail()
		}
		der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if err != nil {
			return fail()
		}
		keys[i] = base64.StdEncoding.EncodeToString(der)
		if i == 0 {
			pkcs1 = base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PublicKey(&key.PublicKey))
		}
	}
	assignment := func(key string) string { return "publicKeyStr='" + key + "';" }
	cases := []struct {
		source            string
		keys, occurrences int
		class             string
	}{
		{assignment(keys[0]), 1, 1, "bare"},
		{assignment(keys[0]) + assignment(pkcs1), 1, 2, "bare"},
		{assignment(keys[0]) + assignment(keys[1]), 2, 2, "bare"},
		{assignment(keys[1]) + assignment(keys[0]), 2, 2, "bare"},
		{"// " + assignment(keys[0]), 0, 1, "none"},
		{"var text=\"" + assignment(keys[0]) + "\";", 0, 1, "none"},
		{"`" + assignment(keys[0]) + "`", 0, 1, "none"},
		{"not" + assignment(keys[0]), 0, 1, "none"},
		{"function f(){" + assignment(keys[0]) + "}", 0, 1, "unknown"},
		{"publicKeyStr='" + keys[0] + "'/2", 0, 1, "unknown"},
	}
	for _, tc := range cases {
		r, err := casprovenance.Analyze(ctx, []casprovenance.Input{{Kind: "inline_script", Data: []byte(tc.source)}})
		if err != nil || len(r.Keys) != tc.keys || len(r.Occurrences) != tc.occurrences || r.Occurrences[0].AssignmentClass != tc.class {
			return fail()
		}
		encoded, err := casprovenance.EncodeReport(r)
		if err != nil {
			return fail()
		}
		if _, err = casprovenance.DecodeReport(encoded); err != nil {
			return fail()
		}
	}
	r, err := casprovenance.Analyze(ctx, []casprovenance.Input{
		{Kind: "page", Data: []byte("<script>" + assignment(keys[0]) + "</script>")},
		{Kind: "external_script", Data: []byte(assignment(keys[1]) + assignment(pkcs1))},
	})
	if err != nil || len(r.Sources) != 3 || len(r.Keys) != 2 || len(r.Occurrences) != 3 || *r.Occurrences[0].KeyRef != *r.Occurrences[2].KeyRef {
		return fail()
	}
	if err = casprovenance.ValidateReport(r); err != nil {
		return fail()
	}
	return envelope{ToolKind: "cas-provenance-dev", Mode: "synthetic-prototype", Binding: "not_packaged", SelfTestResult: "passed", Checks: len(cases) + 1, Report: r}, nil
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
