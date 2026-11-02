package main

import (
	"context"
	"encoding/json"
	"github.com/UnbalancedCat/ipgw-meta/internal/qrdev"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		os.Exit(2)
	}
	switch os.Args[1] {
	case "--self-test":
		_ = json.NewEncoder(os.Stdout).Encode(struct {
			Passed bool `json:"passed"`
		}{qrdev.SelfTest()})
	case "--preflight":
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
		defer cancel()
		_, r := qrdev.Preflight(ctx)
		_ = json.NewEncoder(os.Stdout).Encode(r)
	default:
		os.Exit(2)
	}
}
