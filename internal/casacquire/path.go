package casacquire

import (
	"errors"
	"net/netip"
	"syscall"
)

var errInterfaceBinding = errors.New("interface binding rejected")

type pathBinding struct {
	local netip.Addr
	index uint32
	luid  uint64
}
type routeEntry struct {
	prefix netip.Prefix
	index  uint32
}

// Any competing equal/more-specific route is rejected even if its metric is
// worse. Unrelated virtual subnets do not affect the selected destination.
func routeExclusive(routes []routeEntry, index uint32, destination netip.Addr) bool {
	best, other := -1, -1
	for _, r := range routes {
		if !r.prefix.IsValid() || !r.prefix.Contains(destination) {
			continue
		}
		n := r.prefix.Bits()
		if r.index == index {
			if n > best {
				best = n
			}
		} else if n > other {
			other = n
		}
	}
	return best >= 0 && best > other
}

func (w *wire) control(destination netip.Addr) (func(string, string, syscall.RawConn) error, string) {
	if w.binding != nil {
		if !w.binding.check(destination) {
			return nil, "route_rejected"
		}
		return w.binding.control(destination), ""
	}
	// The only unbound route is an unexported, fully loopback test fixture.
	if !w.local.IsLoopback() || !w.dns.Addr().IsLoopback() || !destination.IsLoopback() || w.testTarget == "" {
		return nil, "interface_binding_failed"
	}
	return nil, ""
}
func (w *wire) pathStillValid(destination netip.Addr) bool {
	if w.binding != nil {
		return w.binding.check(destination)
	}
	_, reason := w.control(destination)
	return reason == ""
}
