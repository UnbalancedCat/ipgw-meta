package cli

import (
	"context"
	"net/netip"
	"testing"

	ipgw "github.com/UnbalancedCat/ipgw-meta"
	"github.com/UnbalancedCat/ipgw-meta/internal/app"
)

func TestBoundScanOnlySelectedAddress(t *testing.T) {
	for _, scope := range [][]string{{"--scope", "bound"}, {"--scope=bound"}} {
		g := &recordingGateway{interfaces: syntheticInterfaces()}
		g.statusFunc = func(_ context.Context, o app.RequestOptions) (ipgw.Status, error) {
			if o.BindIP != "198.51.100.20" {
				t.Fatal("unselected interface probed")
			}
			return testOfflineStatus(), nil
		}
		args := append([]string{"--json", "--bind-ip", "198.51.100.20", "network", "scan"}, scope...)
		code, out, _ := executeWithRecordingGateway(context.Background(), ModeMeta, args, false, g)
		if code != 0 || g.statusCalls != 1 || g.loginCalls != 0 || g.logoutCalls != 0 {
			t.Fatalf("code=%d calls=%d", code, g.statusCalls)
		}
		data := objectField(t, decodeSingleEnvelope(t, []byte(out)), "data")
		if len(data["results"].([]any)) != 1 {
			t.Fatal("extra results")
		}
	}
}

func TestBoundScanRejectsBeforeRequests(t *testing.T) {
	for _, tc := range []struct {
		name, bind string
		duplicate  bool
	}{
		{"missing", "", false}, {"absent", "192.0.2.99", false}, {"ipv6", "::1", false},
		{"loopback", "127.0.0.1", false}, {"zero", "0.0.0.0", false}, {"multicast", "224.0.0.1", false},
		{"linklocal", "169.254.1.1", false}, {"duplicate", "198.51.100.20", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := &recordingGateway{interfaces: syntheticInterfaces()}
			if tc.duplicate {
				g.interfaces = append(g.interfaces, ipgw.Interface{Name: "other", Index: 9, IP: netip.MustParseAddr(tc.bind)})
			}
			args := []string{"--json", "--profile", "synthetic", "network", "scan", "--scope", "bound"}
			if tc.bind != "" {
				args = append(args, "--bind-ip", tc.bind)
			}
			code, _, _ := executeWithRecordingGateway(context.Background(), ModeMeta, args, false, g)
			if code != 2 || g.statusCalls+g.loginCalls+g.logoutCalls != 0 {
				t.Fatalf("code=%d", code)
			}
		})
	}
}

func TestScanScopeCompatibilityAndCancellation(t *testing.T) {
	for _, extra := range [][]string{nil, {"--scope", "all"}} {
		g := &recordingGateway{interfaces: syntheticInterfaces()}
		args := append([]string{"--bind-ip", "198.51.100.20", "network", "scan"}, extra...)
		code, _, _ := executeWithRecordingGateway(context.Background(), ModeMeta, args, false, g)
		if code != 0 || g.statusCalls != 3 {
			t.Fatal("default full scan changed")
		}
	}
	for _, extra := range [][]string{{"--scope"}, {"--scope", "unknown"}, {"--scope", "bound", "--scope", "all"}} {
		g := &recordingGateway{}
		code, _, _ := executeWithRecordingGateway(context.Background(), ModeMeta, append([]string{"network", "scan"}, extra...), false, g)
		if code != 2 || g.totalCalls() != 0 {
			t.Fatal("invalid scope reached gateway")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g := &recordingGateway{interfaces: syntheticInterfaces()}
	code, _, _ := executeWithRecordingGateway(ctx, ModeMeta, []string{"--bind-ip", "198.51.100.20", "network", "scan", "--scope", "bound"}, false, g)
	if code != 130 || g.totalCalls() != 0 {
		t.Fatal("canceled scan reached gateway")
	}
}
