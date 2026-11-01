package main

import (
	"context"
	"github.com/UnbalancedCat/ipgw-meta/internal/trustaudit"
	"os"
)

func run(args []string) ([]byte, error) {
	if len(args) != 1 || (args[0] != "--self-test" && args[0] != "--trust-audit") {
		return nil, trustaudit.ErrReport
	}
	var r trustaudit.Report
	if args[0] == "--self-test" {
		r = trustaudit.SelfTest(context.Background())
	} else {
		r = trustaudit.Run(context.Background())
	}
	return trustaudit.Encode(r)
}
func main() {
	b, err := run(os.Args[1:])
	if err != nil {
		os.Exit(2)
	}
	os.Stdout.Write(b)
}
