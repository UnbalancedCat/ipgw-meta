package cli

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	ipgw "github.com/UnbalancedCat/ipgw-meta"
)

func parseScanScope(args []string) (string, error) {
	if len(args) == 0 {
		return "all", nil
	}
	var scope string
	switch {
	case len(args) == 2 && args[0] == "--scope":
		scope = args[1]
	case len(args) == 1 && strings.HasPrefix(args[0], "--scope="):
		scope = strings.TrimPrefix(args[0], "--scope=")
	default:
		return "", fmt.Errorf("scan accepts one --scope all|bound")
	}
	if scope != "all" && scope != "bound" {
		return "", fmt.Errorf("invalid scan scope")
	}
	return scope, nil
}

// Selection is local-only. No profile fallback or network probe is permitted.
func selectScanInterfaces(ctx context.Context, scope, bind string, interfaces []ipgw.Interface) ([]ipgw.Interface, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if scope == "all" {
		return interfaces, nil
	}
	ip, err := netip.ParseAddr(bind)
	if scope != "bound" || err != nil || !ip.Is4() || ip.IsUnspecified() || ip.IsLoopback() || ip.IsMulticast() || ip.IsLinkLocalUnicast() {
		return nil, invalidArgument(fmt.Errorf("bound scan requires explicit usable IPv4"))
	}
	var matches []ipgw.Interface
	for _, iface := range interfaces {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if iface.IP == ip {
			matches = append(matches, iface)
		}
	}
	if len(matches) != 1 {
		return nil, invalidArgument(fmt.Errorf("bound scan requires one local address match"))
	}
	return matches, nil
}
