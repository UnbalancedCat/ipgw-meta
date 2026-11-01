package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/UnbalancedCat/ipgw-meta/internal/casbundle"
	"github.com/UnbalancedCat/ipgw-meta/internal/ctlcheck"
	"os"
	"path/filepath"
)

var executableDigest, inputsDigest string

type receipt struct {
	SchemaVersion int              `json:"schema_version"`
	ToolKind      string           `json:"tool_kind"`
	SourceDigest  string           `json:"source_digest"`
	Binding       string           `json:"binding"`
	State         string           `json:"state"`
	Result        *ctlcheck.Report `json:"result"`
}

func run(args []string, directory string) receipt {
	r := receipt{SchemaVersion: 1, ToolKind: "ctl-check-dev", SourceDigest: inputsDigest, Binding: "invalid", State: "arguments_rejected"}
	if len(args) != 1 || (args[0] != "--self-test" && args[0] != "--ctl-check") {
		return r
	}
	r.State = "binding_invalid"
	b, e := casbundle.OpenCTLCheck(filepath.Join(directory, "payload"), executableDigest, inputsDigest)
	if e != nil {
		return r
	}
	defer b.Close()
	r.Binding = "verified"
	r.State = "self_test_failed"
	data, e := b.Execute(context.Background(), "--self-test")
	if e != nil {
		return r
	}
	test, e := ctlcheck.Decode(data)
	if e != nil || test.Mode != "self-test" || test.State != "complete" {
		return r
	}
	if args[0] == "--self-test" {
		r.State = "self_test_passed"
		r.Result = &test
		return r
	}
	r.State = "check_failed"
	data, e = b.Execute(context.Background(), "--ctl-check")
	if e != nil {
		return r
	}
	result, e := ctlcheck.Decode(data)
	if e != nil || result.Mode != "local-ctl-check" {
		return r
	}
	r.State = "check_completed"
	r.Result = &result
	return r
}
func main() {
	p, e := os.Executable()
	if e != nil {
		os.Exit(2)
	}
	r := run(os.Args[1:], filepath.Dir(p))
	b, e := json.Marshal(r)
	if e != nil {
		os.Exit(3)
	}
	fmt.Println(string(b))
	if r.State != "self_test_passed" && r.State != "check_completed" {
		os.Exit(2)
	}
}
