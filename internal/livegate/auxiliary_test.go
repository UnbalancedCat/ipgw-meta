package livegate

import (
	"bytes"
	"context"
	"encoding/json"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func validAuxiliary() Auxiliary {
	e := validPasswordEvidence()
	a := Auxiliary{SchemaVersion: 2, PlanID: PlanID, Revision: Revision, SummaryID: strings.Replace(e.EvidenceID, "EVID-", "AUX-", 1), Kind: "bhk_wifi_network", CandidateID: e.CandidateID, CandidateSetSHA256: e.CandidateSetSHA256, SourceCommit: e.SourceCommit, BuildInputSHA256: strings.Repeat("a", 64), RunnerSHA256: e.RunnerSHA256, ExecutionContext: "windows_native", Result: ResultPass, StartedAt: e.StartedAt, FinishedAt: e.FinishedAt}
	ssid := "neu"
	a.WiFiSSID = &ssid
	for _, name := range WiFiChecks {
		a.Checks = append(a.Checks, AuxiliaryCheck{Name: name, Result: ResultPass})
	}
	return a
}
func TestAuxiliaryClosedSchemaAndCompleteChecks(t *testing.T) {
	a := validAuxiliary()
	if err := a.Validate(); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(a)
	raw = append(raw, '\n')
	if _, err := DecodeAuxiliary(raw); err != nil {
		t.Fatal(err)
	}
	mutations := []func(*Auxiliary){func(v *Auxiliary) { v.Checks = v.Checks[:2] }, func(v *Auxiliary) { v.Checks[1].Name = v.Checks[0].Name }, func(v *Auxiliary) { v.Checks[0].Result = Result("skip") }, func(v *Auxiliary) { v.RunnerSHA256 = "" }, func(v *Auxiliary) { v.Kind = "bhk_windows_native_install" }, func(v *Auxiliary) { v.ExecutionContext = "linux" }, func(v *Auxiliary) { v.SummaryID = "../AUX" }, func(v *Auxiliary) { v.Verifier = &InstallVerifier{} }}
	for i, mutation := range mutations {
		v := validAuxiliary()
		mutation(&v)
		if v.Validate() == nil {
			t.Fatalf("mutation %d", i)
		}
	}
	for _, bad := range [][]byte{append(raw, raw...), bytes.Replace(raw, []byte(`"schema_version":2`), []byte(`"schema_version":2,"schema_version":2`), 1), bytes.Replace(raw, []byte(`"verifier":null`), []byte(`"private":"canary","verifier":null`), 1)} {
		if _, err := DecodeAuxiliary(bad); err == nil {
			t.Fatal("invalid json accepted")
		}
	}
}
func TestNetworkProjectionRejectsNestedErrorsExtraAddressesAndTruncation(t *testing.T) {
	status := `{"network":"reachable","session":"offline","username":null,"online_ip":null,"observed_at":"2026-10-31T00:00:00Z","summary":null}`
	item := `{"interface":{"name":"synthetic","index":7,"ip":"192.0.2.11"},"status":` + status + `}`
	valid := `{"schema_version":1,"command":"network.scan","ok":true,"data":{"results":[` + item + `]}}`
	for i, raw := range []string{valid, strings.Replace(valid, `"results":[`+item+`]`, `"results":[]`, 1), strings.Replace(valid, `"results":[`+item+`]`, `"results":[`+item+`,`+item+`]`, 1), strings.Replace(valid, `"status":`+status, `"error":{"code":"network"}`, 1), strings.Replace(valid, "192.0.2.11", "192.0.2.12", 1), valid[:len(valid)-1], strings.Replace(valid, `"ok":true`, `"ok":true,"ok":true`, 1), valid + strings.Repeat(" ", 1<<20)} {
		e, _ := newProcessTestExecutor(t, processTestStep{output: raw})
		e.expectedUsername = "synthetic"
		e.BindIP = netip.MustParseAddr("192.0.2.11")
		e.NetworkIndex = 7
		err := e.NetworkScan(context.Background())
		if (err == nil) != (i == 0) {
			t.Fatalf("projection %d %v", i, err)
		}
	}
}

func TestNetworkListBindsSelectedIndexAndScanIsExplicitBound(t *testing.T) {
	list := `{"schema_version":1,"command":"network.list","ok":true,"data":{"interfaces":[{"name":"selected","index":7,"ip":"192.0.2.11"},{"name":"other","index":8,"ip":"198.51.100.4"}]}}`
	for _, index := range []int{7, 8, 0} {
		e, _ := newProcessTestExecutor(t, processTestStep{output: list})
		e.expectedUsername, e.BindIP, e.NetworkIndex = "synthetic", netip.MustParseAddr("192.0.2.11"), index
		ips, err := e.NetworkList(context.Background())
		if (err == nil) != (index == 7) || err == nil && len(ips) != 2 {
			t.Fatal("selected index projection")
		}
	}
	e, script := newProcessTestExecutor(t, processTestStep{output: `{}`})
	e.expectedUsername, e.BindIP, e.NetworkIndex = "synthetic", netip.MustParseAddr("192.0.2.11"), 7
	_ = e.NetworkScan(context.Background())
	if len(script.calls) != 1 || !strings.HasSuffix(strings.Join(script.calls[0].args, " "), "network scan --scope bound") {
		t.Fatal("unbound scan")
	}
}

func TestAuxiliaryVersionsDoNotInventSSID(t *testing.T) {
	for _, ssid := range []string{"neu", "neu_24g", "neu_mobile", "unknown", ""} {
		a := validAuxiliary()
		a.WiFiSSID = &ssid
		if (a.Validate() == nil) != (ssid == "neu" || ssid == "neu_24g") {
			t.Fatal("SSID validation")
		}
	}
	a := validAuxiliary()
	a.SchemaVersion, a.WiFiSSID = 1, nil
	raw, _ := json.Marshal(a)
	raw = append(raw, '\n')
	if _, err := DecodeAuxiliary(raw); err != nil || bytes.Contains(raw, []byte("wifi_ssid")) {
		t.Fatal("old auxiliary changed", err)
	}
	if _, err := PublishAuxiliary(context.Background(), t.TempDir(), a); err == nil {
		t.Fatal("old auxiliary published")
	}
	a = validAuxiliary()
	a.WiFiSSID = nil
	if a.Validate() == nil {
		t.Fatal("missing current SSID")
	}
}
func TestPathMonitorCancelsActiveCommandAndStaysStopped(t *testing.T) {
	var good atomic.Bool
	good.Store(true)
	e := &pathExecutor{check: func(context.Context) bool { return good.Load() }, interval: time.Millisecond}
	if e.guarded(context.Background(), func(c context.Context) {
		good.Store(false)
		select {
		case <-c.Done():
		case <-time.After(time.Second):
			t.Fatal("monitor did not cancel")
		}
	}) {
		t.Fatal("lost path accepted")
	}
	good.Store(true)
	if e.guarded(context.Background(), func(context.Context) { t.Fatal("cleanup after path loss") }) {
		t.Fatal("sticky lost")
	}
}

// Finishing a command cancels an in-flight local query. That cancellation is
// not evidence of path loss and must not poison subsequent commands.
func TestPathMonitorIgnoresQueryCanceledByWorkCompletion(t *testing.T) {
	entered := make(chan struct{})
	var calls atomic.Int32
	var failures atomic.Int32
	e := &pathExecutor{interval: time.Millisecond, check: func(ctx context.Context) bool {
		if calls.Add(1) == 2 {
			close(entered)
			<-ctx.Done()
			return false
		}
		return true
	}, rejected: func(string) { failures.Add(1) }}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if !e.guarded(ctx, func(c context.Context) {
		select {
		case <-entered:
		case <-c.Done():
			t.Error("monitor did not enter query")
		}
	}) || e.stopped.Load() || ctx.Err() != nil || failures.Load() != 0 {
		t.Fatal("normal query cancellation classified as path loss")
	}
	called := false
	if !e.guarded(ctx, func(context.Context) { called = true }) || !called {
		t.Fatal("next command blocked after normal completion")
	}
}

func TestPathMonitorObservesDuringFailureOnce(t *testing.T) {
	var calls atomic.Int32
	var phases []string
	e := &pathExecutor{interval: time.Millisecond, check: func(context.Context) bool { return calls.Add(1) == 1 }, rejected: func(p string) { phases = append(phases, p) }}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if e.guarded(ctx, func(c context.Context) { <-c.Done() }) || ctx.Err() != nil || len(phases) != 1 || phases[0] != "during" {
		t.Fatal("wrong monitor observation")
	}
	e.guarded(ctx, func(context.Context) { t.Fatal("sticky failure bypassed") })
	if len(phases) != 1 {
		t.Fatal("failure replaced")
	}
}

func TestPathMonitorFailedPrecheckDoesNotExecuteOrCleanup(t *testing.T) {
	good := false
	e := &pathExecutor{check: func(context.Context) bool { return good }}
	if e.guarded(context.Background(), func(context.Context) { t.Fatal("command started") }) {
		t.Fatal("failed precheck accepted")
	}
	good = true
	if e.guarded(context.Background(), func(context.Context) { t.Fatal("cleanup started") }) {
		t.Fatal("stop was not sticky")
	}
}

func TestQRPassesSameExplicitBindingToPrivateChild(t *testing.T) {
	e, script := newProcessTestExecutor(t, processTestStep{})
	e.expectedUsername = "synthetic"
	e.BindIP = netip.MustParseAddr("192.0.2.11")
	e.NetworkIndex = 7
	v := e.Execute(context.Background(), ActionQRLogin)
	if v.RunnerFault || len(script.calls) != 1 {
		t.Fatal("QR invocation")
	}
	if got := strings.Join(script.calls[0].args, " "); got != "--profile campus --bind-ip 192.0.2.11 login --method qr" {
		t.Fatal("QR binding omitted")
	}
	if script.calls[0].stdin != e.Stdin || script.calls[0].stdout != e.Stdout || script.calls[0].stderr != e.Stderr {
		t.Fatal("private TTY changed")
	}
}
func FuzzAuxiliaryDecoder(f *testing.F) {
	a := validAuxiliary()
	raw, _ := json.Marshal(a)
	f.Add(append(raw, '\n'))
	f.Add([]byte(`{}`))
	f.Fuzz(func(t *testing.T, b []byte) {
		v, e := DecodeAuxiliary(b)
		if e == nil && v.Validate() != nil {
			t.Fatal("invalid decode")
		}
	})
}
