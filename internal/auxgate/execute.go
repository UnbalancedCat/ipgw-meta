// Package auxgate implements maintainer-only auxiliary acceptance. Product
// entry points must not import this package.
package auxgate

import (
	"context"
	"encoding/json"
	"net/netip"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/candidate"
	"github.com/UnbalancedCat/ipgw-meta/internal/labpath"
	"github.com/UnbalancedCat/ipgw-meta/internal/livegate"
)

func Handles(args []string) bool {
	return len(args) > 0 && (args[0] == "check-install" || args[0] == "check-network")
}

// Execute has no login or cleanup actions. Network collection is an explicit
// future live window; tests use the pure collectNetwork dependency seam.
func Execute(ctx context.Context, o livegate.Options) int {
	if ctx == nil || runtime.GOOS != "windows" || runtime.GOARCH != "amd64" || !Handles(o.Args) {
		return 2
	}
	flags, ok := parse(o.Args)
	if !ok {
		return 2
	}
	install := o.Args[0] == "check-install"
	if install && !safeWorkRoot(flags["--work-root"], flags["--repository"]) {
		return 2
	}
	if !install && !o.IsTTY {
		return 2
	}
	started := time.Now().UTC()
	duration := 120 * time.Second
	if install {
		duration = 15 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	identity, err := candidate.Verify(flags["--candidate-root"])
	if err != nil {
		return 2
	}
	helper, err := livegate.VerifyHelperBefore(filepath.Join(identity.Root, "candidate-manifest.json"), livegate.PlatformWindowsAMD64)
	if err != nil || helper.Binding.CandidateSetSHA256 != identity.CandidateSetSHA256 {
		return 2
	}
	kind := "bhk_wifi_network"
	if install {
		kind = "bhk_windows_native_install"
	}
	a := livegate.Auxiliary{SchemaVersion: 1, PlanID: livegate.PlanID, Revision: livegate.Revision, Kind: kind, CandidateID: identity.CandidateID, CandidateSetSHA256: identity.CandidateSetSHA256, SourceCommit: identity.SourceCommit, BuildInputSHA256: identity.BuildInputSHA256, RunnerSHA256: helper.Binding.RunnerSHA256, ExecutionContext: "windows_native", Result: livegate.ResultBlocked, StartedAt: livegate.Timestamp(started.Format(time.RFC3339))}
	a.Checks = append(a.Checks, livegate.AuxiliaryCheck{Name: "candidate_verified", Result: livegate.ResultPass})
	var collectionErr error
	if install {
		if !protectInstall() {
			return 2
		}
		a.Verifier, collectionErr = collectInstall(ctx, flags, identity, &a)
	} else {
		ip, _ := netip.ParseAddr(flags["--bind-ip"])
		product, e := livegate.VerifyCandidateBefore(flags["--candidate"], filepath.Join(identity.Root, "candidate-manifest.json"), helper.Binding.SHA256, livegate.PlatformWindowsAMD64)
		if e != nil {
			return 2
		}
		snapshot, e := labpath.Capture(ctx, "campus_wifi", ip)
		if e != nil {
			fail(&a, "selected_wifi_topology")
		} else {
			guard := livegate.GuardPath(snapshot.Check)
			executor := &livegate.ProcessExecutor{CandidatePath: product.CandidatePath, Profile: flags["--profile"], BindIP: ip, Stderr: o.Stderr, RequireRefuse: true}
			deps := networkDependencies{guard: guard, prepare: func(c context.Context) error { return executor.PrepareProfile(c, livegate.SuiteTerminalQR) }, status: func(c context.Context, action livegate.CommandAction) livegate.Observation {
				var observed livegate.Observation
				if !guard(c, func(child context.Context) { observed = executor.Execute(child, action) }) {
					return livegate.Observation{RunnerFault: true}
				}
				return observed
			}, list: executor.NetworkList, scan: executor.NetworkScan}
			collectNetwork(ctx, ip, deps, &a)
		}
		if product.VerifyAfter() != nil {
			return 2
		}
	}
	if collectionErr != nil && a.Checks[len(a.Checks)-1].Result == livegate.ResultPass {
		fail(&a, livegate.InstallChecks[len(a.Checks)])
	}
	again, e := candidate.Verify(identity.Root)
	if e != nil || again != identity || helper.VerifyAfter() != nil {
		return 2
	}
	if a.Checks[len(a.Checks)-1].Result == livegate.ResultPass {
		pass(&a, "candidate_unchanged")
		a.Result = livegate.ResultPass
	}
	a.FinishedAt = livegate.Timestamp(time.Now().UTC().Format(time.RFC3339))
	if ctx.Err() != nil {
		return 2
	}
	id, err := livegate.PublishAuxiliary(ctx, flags["--out"], a)
	if err != nil {
		return 2
	}
	if o.Stdout != nil {
		_ = json.NewEncoder(o.Stdout).Encode(struct {
			Schema int             `json:"schema_version"`
			ID     string          `json:"summary_id"`
			Result livegate.Result `json:"result"`
		}{1, id, a.Result})
	}
	if a.Result != livegate.ResultPass {
		return 1
	}
	return 0
}

func safeWorkRoot(root, repo string) bool {
	a, err := filepath.EvalSymlinks(root)
	if err != nil || !strings.EqualFold(filepath.Clean(a), filepath.Clean(root)) {
		return false
	}
	b, err := filepath.EvalSymlinks(repo)
	if err != nil {
		return false
	}
	info, err := os.Stat(a)
	if err != nil || !info.IsDir() {
		return false
	}
	within := func(child, parent string) bool {
		rel, e := filepath.Rel(strings.ToLower(parent), strings.ToLower(child))
		return e == nil && (rel == "." || !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && rel != "..")
	}
	home, _ := os.UserHomeDir()
	return !within(a, b) && !within(b, a) && !strings.EqualFold(a, filepath.VolumeName(a)+string(os.PathSeparator)) && !strings.EqualFold(a, home)
}

func parse(args []string) (map[string]string, bool) {
	required := []string{"--candidate-root", "--out"}
	if args[0] == "check-install" {
		required = append(required, "--repository", "--go", "--module-cache", "--work-root")
	} else {
		required = append(required, "--candidate", "--profile", "--bind-ip")
	}
	if len(args) != 1+2*len(required) {
		return nil, false
	}
	allowed := map[string]bool{}
	for _, k := range required {
		allowed[k] = true
	}
	result := map[string]string{}
	for i := 1; i < len(args); i += 2 {
		if !allowed[args[i]] || result[args[i]] != "" || args[i+1] == "" {
			return nil, false
		}
		result[args[i]] = args[i+1]
	}
	for k, v := range result {
		if k != "--profile" && k != "--bind-ip" && (!filepath.IsAbs(v) || filepath.Clean(v) != v) {
			return nil, false
		}
	}
	if args[0] == "check-network" {
		ip, e := netip.ParseAddr(result["--bind-ip"])
		if e != nil || !ip.Is4() || ip.IsLoopback() || ip.IsUnspecified() || ip.IsMulticast() || ip.IsLinkLocalUnicast() || !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`).MatchString(result["--profile"]) {
			return nil, false
		}
	}
	return result, true
}
func pass(a *livegate.Auxiliary, name string) {
	a.Checks = append(a.Checks, livegate.AuxiliaryCheck{Name: name, Result: livegate.ResultPass})
}
func fail(a *livegate.Auxiliary, name string) {
	a.Checks = append(a.Checks, livegate.AuxiliaryCheck{Name: name, Result: livegate.ResultBlocked})
	a.Result = livegate.ResultBlocked
}

type networkDependencies struct {
	guard   func(context.Context, func(context.Context)) bool
	prepare func(context.Context) error
	status  func(context.Context, livegate.CommandAction) livegate.Observation
	list    func(context.Context) ([]netip.Addr, error)
	scan    func(context.Context) error
}

func collectNetwork(ctx context.Context, ip netip.Addr, d networkDependencies, a *livegate.Auxiliary) {
	check := func(name string, work func(context.Context) bool) bool {
		ok := false
		guarded := d.guard(ctx, func(c context.Context) {
			child, cancel := context.WithTimeout(c, 30*time.Second)
			defer cancel()
			ok = work(child) && child.Err() == nil
		})
		if !guarded || !ok {
			fail(a, name)
			return false
		}
		pass(a, name)
		return true
	}
	offline := func(c context.Context) bool {
		v := d.status(c, livegate.ActionInitialStatus)
		return !v.RunnerFault && v.ExitCode == 0 && v.ErrorCode == nil && v.Session == livegate.SessionOffline
	}
	if !check("selected_wifi_topology", func(c context.Context) bool { return d.prepare(c) == nil && offline(c) }) {
		return
	}
	var ips []netip.Addr
	if !check("list_complete", func(c context.Context) bool { var e error; ips, e = d.list(c); return e == nil }) {
		return
	}
	if !check("selected_address_unique", func(context.Context) bool { return len(ips) == 1 && ips[0] == ip }) {
		return
	}
	if !check("bound_status_valid", offline) {
		return
	}
	if !check("scan_scope_valid", func(c context.Context) bool { v, e := d.list(c); return e == nil && len(v) == 1 && v[0] == ip }) {
		return
	}
	if !check("scan_complete", func(c context.Context) bool { return d.scan(c) == nil }) {
		return
	}
	// NetworkScan validates every result, including nested errors and membership.
	if !check("selected_result_valid", func(context.Context) bool { return true }) {
		return
	}
	check("topology_unchanged", func(context.Context) bool { return true })
}
