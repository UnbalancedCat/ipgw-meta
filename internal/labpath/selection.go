package labpath

import (
	"context"
	ipgw "github.com/UnbalancedCat/ipgw-meta"
	"net/netip"
)

// ValidateLocalSelection is shared by zero-network preflight and the formal
// list projection. Extra addresses are retained, but malformed/duplicate rows
// or a selected address on another interface cannot pass.
func ValidateLocalSelection(ctx context.Context, items []ipgw.Interface, selected netip.Addr, index uint32) (int, error) {
	if ctx == nil || ctx.Err() != nil || !usableIPv4(selected) || index == 0 || len(items) > 256 {
		return 0, ErrTopology
	}
	seen := map[netip.Addr]bool{}
	count := 0
	for _, item := range items {
		if ctx.Err() != nil || !usableIPv4(item.IP) || item.Index < 1 || item.Name == "" || seen[item.IP] {
			return count, ErrTopology
		}
		seen[item.IP] = true
		if item.IP == selected {
			count++
			if item.Index != int(index) {
				return count, ErrTopology
			}
		}
	}
	if count != 1 {
		return count, ErrTopology
	}
	return count, nil
}
