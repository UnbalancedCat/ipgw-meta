package casacquire

import (
	"context"
	"net/netip"
	"testing"
)

func TestExclusiveDestinationRoute(t *testing.T) {
	route := func(s string, i uint32) routeEntry { return routeEntry{prefix: netip.MustParsePrefix(s), index: i} }
	destination := netip.MustParseAddr("192.0.2.42")
	for _, tc := range []struct {
		name   string
		routes []routeEntry
		ok     bool
	}{
		{"wired-default", []routeEntry{route("0.0.0.0/0", 5)}, true},
		{"unrelated-virtual", []routeEntry{route("0.0.0.0/0", 5), route("198.51.100.0/24", 8)}, true},
		{"foreign-more-specific", []routeEntry{route("0.0.0.0/0", 5), route("192.0.2.0/24", 8)}, false},
		{"foreign-equal", []routeEntry{route("192.0.2.0/24", 5), route("192.0.2.0/24", 8)}, false},
		{"selected-more-specific", []routeEntry{route("192.0.2.0/24", 5), route("0.0.0.0/0", 8)}, true},
		{"missing", []routeEntry{route("0.0.0.0/0", 8)}, false},
		{"empty", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if routeExclusive(tc.routes, 5, destination) != tc.ok {
				t.Fatal("route ambiguity accepted")
			}
		})
	}
}
func TestNoUnboundProductionRoute(t *testing.T) {
	w := &wire{local: netip.MustParseAddr("192.0.2.10"), dns: netip.MustParseAddrPort("192.0.2.53:53")}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if _, reason := w.resolve(ctx, "example.invalid"); reason != "interface_binding_failed" || w.counts.DNS != 0 {
		t.Fatal("unbound resolver attempted network")
	}
}
func TestPreflightClosedProjection(t *testing.T) {
	r := blocked("local-preflight", "none")
	r.Collection = "ready"
	b, err := Encode(r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Decode(b); err != nil {
		t.Fatal(err)
	}
	r.Attempts.DNS = 1
	if _, err = Encode(r); err == nil {
		t.Fatal("preflight network claim accepted")
	}
	r = blocked("local-preflight", "topology_query_failed")
	r.ContextConstructed = true
	if _, err = Encode(r); err == nil {
		t.Fatal("preflight context claim accepted")
	}
}
