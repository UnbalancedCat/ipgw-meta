package main

import (
	"context"
	"github.com/UnbalancedCat/ipgw-meta/internal/ctlcheck"
	"os"
)

func run(args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, ctlcheck.ErrInvalid
	}
	switch args[0] {
	case "--self-test":
		r := ctlcheck.SyntheticProcess(context.Background())
		if r.State != "complete" {
			return nil, ctlcheck.ErrInvalid
		}
		return ctlcheck.Encode(r)
	case "--ctl-check":
		return ctlcheck.Encode(ctlcheck.Run(context.Background()))
	}
	return nil, ctlcheck.ErrInvalid
}
func main() {
	b, e := run(os.Args[1:])
	if e != nil {
		os.Exit(2)
	}
	os.Stdout.Write(b)
}
