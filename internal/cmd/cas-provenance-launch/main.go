// cas-provenance-launch is reviewed separately from the four-file development bundle.
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/UnbalancedCat/ipgw-meta/internal/casbundle"
)

func run(args []string, out, diagnostic io.Writer) int {
	if len(args) == 8 && args[0] == "--install-archive" && args[2] == "--archive-sha256" && args[4] == "--manifest-sha256" && args[6] == "--destination" {
		receipt := casbundle.Install(context.Background(), args[1], args[7], args[3], args[5])
		data, err := casbundle.EncodeInstallReceipt(receipt)
		if err != nil {
			fmt.Fprintln(diagnostic, "development receipt rejected")
			return 1
		}
		if _, err = out.Write(data); err != nil {
			return 1
		}
		if receipt.State != "installed" {
			return 1
		}
		return 0
	}
	if len(args) != 4 || args[0] != "--bundle" || args[2] != "--manifest-sha256" {
		fmt.Fprintln(diagnostic, "explicit bundle and external manifest digest required")
		return 2
	}
	receipt := casbundle.Run(context.Background(), args[1], args[3])
	data, err := casbundle.EncodeReceipt(receipt)
	if err != nil {
		fmt.Fprintln(diagnostic, "development receipt rejected")
		return 1
	}
	if _, err = out.Write(data); err != nil {
		return 1
	}
	if receipt.Runner != "passed" {
		return 1
	}
	return 0
}
func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
