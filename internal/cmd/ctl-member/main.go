package main

import (
	"context"
	"github.com/UnbalancedCat/ipgw-meta/internal/ctlcheck"
	"os"
)

func run(args []string) ([]byte, error) {
	if len(args) != 1 || args[0] != "--self-test" {
		return nil, ctlcheck.ErrInvalid
	}
	return ctlcheck.MemberSyntheticProcess(context.Background())
}
func main() {
	b, e := run(os.Args[1:])
	if e != nil {
		os.Exit(2)
	}
	os.Stdout.Write(b)
}
