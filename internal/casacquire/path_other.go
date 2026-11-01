//go:build !windows

package casacquire

import (
	"net/netip"
	"syscall"
)

func newPathBinding(netip.Addr, uint32) (*pathBinding, bool) { return nil, false }
func (b *pathBinding) check(netip.Addr) bool                 { return false }
func (b *pathBinding) control(netip.Addr) func(string, string, syscall.RawConn) error {
	return func(string, string, syscall.RawConn) error { return ErrReport }
}
