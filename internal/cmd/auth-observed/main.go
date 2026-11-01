package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/UnbalancedCat/ipgw-meta/internal/app"
	"github.com/UnbalancedCat/ipgw-meta/internal/authobserve"
	"github.com/UnbalancedCat/ipgw-meta/internal/cli"
	"github.com/UnbalancedCat/ipgw-meta/internal/config"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--self-test" {
		ok := authobserve.SelfTest()
		fmt.Printf("{\"schema_version\":1,\"tool_kind\":\"auth-observed-selftest\",\"passed\":%t}\n", ok)
		if !ok {
			os.Exit(2)
		}
		return
	}
	if len(os.Args) < 3 || os.Args[1] != "--internal-observed" {
		os.Exit(2)
	}
	args := os.Args[2:]
	command := authobserve.Command(args)
	if command == "" {
		os.Exit(2)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	collector := authobserve.New(command)
	product := &authobserve.LimitBuffer{Limit: authobserve.MaxProduct}
	code := cli.Execute(ctx, cli.Options{
		Mode: cli.ModeMeta, Args: args, ResolvePaths: config.DefaultPaths,
		NewGateway: func(paths config.Paths) cli.Gateway {
			return collector.Gateway(app.NewObserved(paths, os.Stdin, os.Stderr, collector.Decorate))
		},
		Input: os.Stdin, Out: product, Err: os.Stderr, IsTTY: cli.IsSafeTerminal(os.Stdin, os.Stderr), Version: "dev-auth-observed",
	})
	result := authobserve.Child{SchemaVersion: 2, ToolKind: "auth-observed-cli", Command: command, ExitCode: code, Product: product.Bytes(), Observation: collector.Snapshot()}
	data, err := json.Marshal(result)
	if err != nil {
		os.Exit(11)
	}
	if _, err = authobserve.DecodeChild(data, command, code); err != nil {
		os.Exit(11)
	}
	_, err = os.Stdout.Write(append(data, '\n'))
	clear(product.Bytes())
	clear(data)
	if err != nil {
		os.Exit(11)
	}
	os.Exit(code)
}
