//go:build windows

package casacquire

import (
	"context"
	"net"
	"net/netip"
	"testing"

	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
)

func loopbackBinding(t *testing.T) *pathBinding {
	t.Helper()
	interfaces, err := net.Interfaces()
	if err != nil {
		t.Fatal("local interfaces unavailable")
	}
	for _, i := range interfaces {
		if i.Flags&net.FlagLoopback != 0 && i.Flags&net.FlagUp != 0 {
			b, ok := newPathBinding(netip.MustParseAddr("127.0.0.1"), uint32(i.Index))
			if ok {
				return b
			}
		}
	}
	t.Fatal("loopback interface binding unavailable")
	return nil
}
func TestWindowsBoundUDPAndTCP(t *testing.T) {
	w, dns, requests := fixtureWire(t, false, "HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")
	w.binding = loopbackBinding(t)
	_, reason := w.fetch(context.Background(), cascollect.Request{Method: "GET", Target: "http://ipgw.neu.edu.cn/"}, 4)
	if reason != "" || dns.Load() != 1 || requests.Load() != 1 {
		t.Fatalf("loopback socket binding: %s", reason)
	}
}
func TestChangedInterfaceStopsBeforeDNS(t *testing.T) {
	w, dns, requests := fixtureWire(t, false, "HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")
	w.binding = loopbackBinding(t)
	w.binding.luid++
	_, reason := w.fetch(context.Background(), cascollect.Request{Method: "GET", Target: "http://ipgw.neu.edu.cn/"}, 4)
	if reason != "route_rejected" || dns.Load() != 0 || requests.Load() != 0 || w.counts.DNS != 0 || w.counts.Connects != 0 {
		t.Fatal("changed interface allowed a packet")
	}
}
