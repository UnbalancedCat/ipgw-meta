//go:build windows

package casacquire

import (
	"math/bits"
	"net/netip"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const ipUnicastIF = 31

func newPathBinding(local netip.Addr, index uint32) (*pathBinding, bool) {
	if !local.Is4() || index == 0 || index > 0xffffff {
		return nil, false
	}
	row := windows.MibIfRow2{InterfaceIndex: index}
	if windows.GetIfEntry2Ex(windows.MibIfEntryNormalWithoutStatistics, &row) != nil {
		return nil, false
	}
	b := &pathBinding{local: local, index: index, luid: row.InterfaceLuid}
	return b, b.interfaceValid()
}
func (b *pathBinding) interfaceValid() bool {
	row := windows.MibIfRow2{InterfaceIndex: b.index}
	if windows.GetIfEntry2Ex(windows.MibIfEntryNormalWithoutStatistics, &row) != nil || row.InterfaceLuid != b.luid || row.OperStatus != windows.IfOperStatusUp {
		return false
	}
	if b.local.IsLoopback() {
		if row.Type != windows.IF_TYPE_SOFTWARE_LOOPBACK {
			return false
		}
	} else if row.Type != windows.IF_TYPE_ETHERNET_CSMACD || row.InterfaceAndOperStatusFlags&1 == 0 {
		return false
	}
	address := windows.MibUnicastIpAddressRow{InterfaceIndex: b.index, InterfaceLuid: b.luid}
	sock := (*windows.RawSockaddrInet4)(unsafe.Pointer(&address.Address))
	sock.Family = windows.AF_INET
	sock.Addr = b.local.As4()
	return windows.GetUnicastIpAddressEntry(&address) == nil && address.DadState == windows.IpDadStatePreferred && address.SkipAsSource == 0
}
func (b *pathBinding) check(destination netip.Addr) bool {
	if b == nil || !destination.Is4() || !b.interfaceValid() {
		return false
	}
	var table *windows.MibIpForwardTable2
	if windows.GetIpForwardTable2(windows.AF_INET, &table) != nil || table == nil {
		return false
	}
	defer windows.FreeMibTable(unsafe.Pointer(table))
	if table.NumEntries > 4096 {
		return false
	}
	routes := make([]routeEntry, 0, table.NumEntries)
	for _, r := range table.Rows() {
		if r.DestinationPrefix.Prefix.Family != windows.AF_INET || r.DestinationPrefix.PrefixLength > 32 {
			return false
		}
		address := (*windows.RawSockaddrInet4)(unsafe.Pointer(&r.DestinationPrefix.Prefix)).Addr
		p := netip.PrefixFrom(netip.AddrFrom4(address), int(r.DestinationPrefix.PrefixLength)).Masked()
		routes = append(routes, routeEntry{prefix: p, index: r.InterfaceIndex})
	}
	return routeExclusive(routes, b.index, destination)
}
func (b *pathBinding) control(destination netip.Addr) func(string, string, syscall.RawConn) error {
	return func(network, address string, raw syscall.RawConn) error {
		if (network != "udp4" && network != "tcp4") || !b.check(destination) {
			return errInterfaceBinding
		}
		var optionErr error
		err := raw.Control(func(fd uintptr) {
			// Windows sets in network byte order but gets in host byte order.
			optionErr = windows.SetsockoptInt(windows.Handle(fd), windows.IPPROTO_IP, ipUnicastIF, int(bits.ReverseBytes32(b.index)))
			if optionErr != nil {
				return
			}
			var got int
			got, optionErr = windows.GetsockoptInt(windows.Handle(fd), windows.IPPROTO_IP, ipUnicastIF)
			if optionErr == nil && got != int(b.index) {
				optionErr = ErrReport
			}
		})
		if err != nil || optionErr != nil {
			return errInterfaceBinding
		}
		return nil
	}
}
