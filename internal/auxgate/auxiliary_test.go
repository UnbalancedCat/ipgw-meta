package auxgate

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/UnbalancedCat/ipgw-meta/internal/candidate"
	"github.com/UnbalancedCat/ipgw-meta/internal/livegate"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWiFiRefusesAmbiguousScopeAndNeverAuthenticates(t *testing.T) {
	ip := netip.MustParseAddr("192.0.2.11")
	for _, tc := range []struct {
		name                            string
		multi, drift, scanError, online bool
		scans                           int
		last                            string
	}{
		{name: "complete", scans: 1, last: "topology_unchanged"}, {name: "multiple", multi: true, last: "selected_address_unique"}, {name: "changed_list", drift: true, last: "scan_scope_valid"}, {name: "nested_error", scanError: true, scans: 1, last: "scan_complete"}, {name: "already_online", online: true, last: "selected_wifi_topology"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lists, scans := 0, 0
			a := livegate.Auxiliary{Result: livegate.ResultBlocked}
			pass(&a, "candidate_verified")
			d := networkDependencies{guard: func(c context.Context, f func(context.Context)) bool { f(c); return true }, prepare: func(context.Context) error { return nil }, status: func(_ context.Context, action livegate.CommandAction) livegate.Observation {
				if action != livegate.ActionInitialStatus {
					t.Fatal("authentication action")
				}
				session := livegate.SessionOffline
				if tc.online {
					session = livegate.SessionOnline
				}
				return livegate.Observation{Session: session}
			}, list: func(context.Context) ([]netip.Addr, error) {
				lists++
				if tc.multi || tc.drift && lists == 2 {
					return []netip.Addr{ip, netip.MustParseAddr("192.0.2.12")}, nil
				}
				return []netip.Addr{ip}, nil
			}, scan: func(context.Context) error {
				scans++
				if tc.scanError {
					return errors.New("synthetic")
				}
				return nil
			}}
			collectNetwork(context.Background(), ip, d, &a)
			if scans != tc.scans || a.Checks[len(a.Checks)-1].Name != tc.last {
				t.Fatalf("scans %d checks %#v", scans, a.Checks)
			}
		})
	}
}

func TestWiFiPathLossAndCancellationStopFollowingCommands(t *testing.T) {
	ip := netip.MustParseAddr("192.0.2.11")
	for stop := 0; stop < 8; stop++ {
		t.Run(string(rune('a'+stop)), func(t *testing.T) {
			count := 0
			a := livegate.Auxiliary{}
			pass(&a, "candidate_verified")
			d := networkDependencies{guard: func(c context.Context, f func(context.Context)) bool {
				count++
				if count > stop {
					return false
				}
				f(c)
				return true
			}, prepare: func(context.Context) error { return nil }, status: func(context.Context, livegate.CommandAction) livegate.Observation {
				return livegate.Observation{Session: livegate.SessionOffline}
			}, list: func(context.Context) ([]netip.Addr, error) { return []netip.Addr{ip}, nil }, scan: func(context.Context) error { return nil }}
			collectNetwork(context.Background(), ip, d, &a)
			if count != stop+1 || a.Checks[len(a.Checks)-1].Result != livegate.ResultBlocked {
				t.Fatal("continued after stop")
			}
		})
	}
}

func syntheticInstallEvents() []byte {
	var out []byte
	appendEvent := func(action, name string) {
		raw, _ := json.Marshal(map[string]any{"Action": action, "Package": "github.com/UnbalancedCat/ipgw-meta/internal/installtest", "Test": name})
		out = append(out, raw...)
		out = append(out, '\n')
	}
	appendEvent("start", "")
	for _, name := range []string{"TestNativeReleaseAssetSmoke", "TestNativeReleaseAssetSmoke/fresh_install_and_launcher_default", "TestNativeReleaseAssetSmoke/upgrade_preserves_launcher", "TestNativeReleaseAssetSmoke/basic_rollback_restores_baseline"} {
		appendEvent("run", name)
		appendEvent("pass", name)
	}
	appendEvent("pass", "")
	return out
}
func TestInstallRequiresAllNamedTestsWithoutSkip(t *testing.T) {
	valid := syntheticInstallEvents()
	if validateInstallEvents(valid) != nil {
		t.Fatal("valid transcript")
	}
	for _, raw := range [][]byte{nil, valid[:len(valid)/2], []byte(strings.Replace(string(valid), `"pass"`, `"skip"`, 1)), []byte(strings.Replace(string(valid), "basic_rollback_restores_baseline", "unexpected", -1)), append(valid, valid...), []byte(strings.Replace(string(valid), `"run"`, `"output"`, 1))} {
		if validateInstallEvents(raw) == nil {
			t.Fatal("invalid transcript accepted")
		}
	}
}
func TestInstallEnvironmentCannotInheritInjection(t *testing.T) {
	t.Setenv("GOFLAGS", "-overlay=private")
	t.Setenv("GOWORK", "private")
	t.Setenv("GOTOOLCHAIN", "auto")
	t.Setenv("IPGW_NATIVE_INSTALL_REQUIRED", "0")
	t.Setenv("HTTP_PROXY", "synthetic-private")
	env := installEnvironment(map[string]string{}, "go.exe", "cache", candidate.Result{})
	joined := strings.Join(env, "\n")
	for _, bad := range []string{"-overlay=private", "GOWORK=private", "GOTOOLCHAIN=auto", "IPGW_NATIVE_INSTALL_REQUIRED=0", "synthetic-private"} {
		if strings.Contains(joined, bad) {
			t.Fatal("inherited injection")
		}
	}
	for _, required := range []string{"GOPROXY=off", "GOWORK=off", "GOENV=off", "IPGW_NATIVE_INSTALL_CANDIDATE_ONLY=1", "IPGW_NATIVE_INSTALL_REQUIRED=1"} {
		if !strings.Contains(joined, required) {
			t.Fatal("missing isolation")
		}
	}
}

func TestWorkRootCannotOverlapSourceOrVolume(t *testing.T) {
	base := t.TempDir()
	repo := filepath.Join(base, "repo")
	work := filepath.Join(base, "work")
	inside := filepath.Join(repo, "temp")
	for _, dir := range []string{repo, work, inside} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	if !safeWorkRoot(work, repo) {
		t.Fatal("independent fixture rejected")
	}
	for _, dir := range []string{repo, base, inside, filepath.VolumeName(base) + string(os.PathSeparator)} {
		if safeWorkRoot(dir, repo) {
			t.Fatal("unsafe work root")
		}
	}
}
