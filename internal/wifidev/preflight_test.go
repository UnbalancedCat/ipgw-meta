package wifidev

import (
	"context"
	ipgw "github.com/UnbalancedCat/ipgw-meta"
	"github.com/UnbalancedCat/ipgw-meta/internal/labpath"
	"net/netip"
	"testing"
)

func TestPreflightScopeAndClosedFailure(t *testing.T) {
	ip := netip.MustParseAddr("192.0.2.4")
	s := labpath.Snapshot{Local: ip, Index: 7, WiFiSSID: labpath.WiFiNEU}
	good := []ipgw.Interface{{Name: "test", Index: 7, IP: ip}, {Name: "virtual", Index: 8, IP: netip.MustParseAddr("198.51.100.3")}}
	for _, tc := range []struct {
		name  string
		items []ipgw.Interface
		want  string
	}{
		{"unrelated allowed", good, "ready"}, {"missing", good[1:], "blocked"}, {"duplicate", append(append([]ipgw.Interface{}, good...), good[0]), "blocked"},
		{"wrong index", []ipgw.Interface{{Index: 9, IP: ip}}, "blocked"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, r := preflight(context.Background(), labpath.WiFiNEU, func(context.Context) (labpath.Snapshot, error) { return s, nil }, func(context.Context) ([]ipgw.Interface, error) { return tc.items, nil })
			if r.State != tc.want || r.NetworkRequests != 0 || r.ProfileReads != 0 {
				t.Fatal(r)
			}
		})
	}
	for _, err := range []error{labpath.ErrWiFiPermission, labpath.ErrWiFiDisconnected, labpath.ErrWiFiUnsupported, labpath.ErrWiFiUnknown, labpath.ErrWiFiUnavailable} {
		_, r := preflight(context.Background(), labpath.WiFiNEU, func(context.Context) (labpath.Snapshot, error) { return labpath.Snapshot{}, err }, func(context.Context) ([]ipgw.Interface, error) { t.Fatal("enumerated after failure"); return nil, nil })
		if r.State != "blocked" || r.Reason == "topology_unavailable" {
			t.Fatal(r)
		}
	}
	if !SelfTest() {
		t.Fatal("self test")
	}
}
