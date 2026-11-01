package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/UnbalancedCat/ipgw-meta/internal/casbundle"
	"github.com/UnbalancedCat/ipgw-meta/internal/trustaudit"
	"os"
	"path/filepath"
)

var executableDigest, inputsDigest string

type receipt struct {
	SchemaVersion int                `json:"schema_version"`
	ToolKind      string             `json:"tool_kind"`
	SourceDigest  string             `json:"source_digest"`
	Binding       string             `json:"binding"`
	State         string             `json:"state"`
	Result        *trustaudit.Report `json:"result"`
}

func run(args []string, directory string) receipt {
	r := receipt{SchemaVersion: 1, ToolKind: "trust-audit-dev", SourceDigest: inputsDigest, Binding: "invalid", State: "arguments_rejected"}
	if len(args) != 1 || (args[0] != "--self-test" && args[0] != "--trust-audit") {
		return r
	}
	r.State = "binding_invalid"
	b, err := casbundle.OpenTrustAudit(filepath.Join(directory, "payload"), executableDigest, inputsDigest)
	if err != nil {
		return r
	}
	defer b.Close()
	r.Binding = "verified"
	r.State = "self_test_failed"
	data, err := b.Execute(context.Background(), "--self-test")
	if err != nil {
		return r
	}
	test, err := trustaudit.Decode(data)
	if err != nil || test.Mode != "self-test" || test.State != "complete" || test.Root.ParseCoverage != "complete" || test.CA.ParseCoverage != "partial" || test.Comparison.ParsedRootDeniedOverlap != "present" || test.Comparison.RemainingParsedRoots != "empty" {
		return r
	}
	if args[0] == "--self-test" {
		r.State = "self_test_passed"
		r.Result = &test
		return r
	}
	r.State = "audit_failed"
	data, err = b.Execute(context.Background(), "--trust-audit")
	if err != nil {
		return r
	}
	result, err := trustaudit.Decode(data)
	if err != nil || result.Mode != "local-trust-audit" {
		return r
	}
	r.State = "audit_completed"
	r.Result = &result
	return r
}
func main() {
	p, err := os.Executable()
	if err != nil {
		os.Exit(2)
	}
	r := run(os.Args[1:], filepath.Dir(p))
	b, err := json.Marshal(r)
	if err != nil {
		os.Exit(3)
	}
	fmt.Println(string(b))
	if r.State != "self_test_passed" && r.State != "audit_completed" {
		os.Exit(2)
	}
}
