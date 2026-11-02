package main

import (
	"context"
	"encoding/json"
	"github.com/UnbalancedCat/ipgw-meta/internal/labpath"
	"github.com/UnbalancedCat/ipgw-meta/internal/wifidev"
	"os"
	"time"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--self-test" {
		_ = json.NewEncoder(os.Stdout).Encode(struct {
			Passed bool `json:"passed"`
		}{wifidev.SelfTest()})
		return
	}
	if len(os.Args) != 3 || os.Args[1] != "--preflight" {
		os.Exit(2)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	_, r := wifidev.Preflight(ctx, labpath.WiFiSSID(os.Args[2]))
	_ = json.NewEncoder(os.Stdout).Encode(r)
}
