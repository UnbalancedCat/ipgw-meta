package livegate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func validAuxiliary() Auxiliary {
	e := validPasswordEvidence()
	a := Auxiliary{SchemaVersion: 1, PlanID: PlanID, Revision: Revision, SummaryID: strings.Replace(e.EvidenceID, "EVID-", "AUX-", 1), Kind: "bhk_wifi_network", CandidateID: e.CandidateID, CandidateSetSHA256: e.CandidateSetSHA256, SourceCommit: e.SourceCommit, BuildInputSHA256: strings.Repeat("a", 64), RunnerSHA256: e.RunnerSHA256, ExecutionContext: "windows_native", Result: ResultPass, StartedAt: e.StartedAt, FinishedAt: e.FinishedAt}
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
	for _, bad := range [][]byte{append(raw, raw...), bytes.Replace(raw, []byte(`"schema_version":1`), []byte(`"schema_version":1,"schema_version":1`), 1), bytes.Replace(raw, []byte(`"verifier":null`), []byte(`"private":"canary","verifier":null`), 1)} {
		if _, err := DecodeAuxiliary(bad); err == nil {
			t.Fatal("invalid json accepted")
		}
	}
}
func TestAuxiliaryPublicationAndCanonicalSummary(t *testing.T) {
	a := validAuxiliary()
	a.SummaryID = ""
	root := bundleTestOutputDir(t)
	id, err := PublishAuxiliary(context.Background(), root, a)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, a.CandidateID, id)
	got, files, err := ReadAuxiliaryBundle(path)
	if err != nil || got.SummaryID != id {
		t.Fatal("bundle read", err)
	}
	raw, err := PublicSummary(path)
	if err != nil {
		t.Fatal(err)
	}
	var summary map[string]any
	if json.Unmarshal(raw, &summary) != nil || summary["bundle_sha256"] != bundleDigest(files) {
		t.Fatal("summary")
	}
	framed := fmt.Sprintf("SHA256SUMS\x00%d\x00%s\nevidence.json\x00%d\x00%s\nsummary.md\x00%d\x00%s\n", len(files[2]), testHash(files[2]), len(files[0]), testHash(files[0]), len(files[1]), testHash(files[1]))
	if summary["bundle_sha256"] != testHash([]byte(framed)) {
		t.Fatal("bundle framing changed")
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = PublishAuxiliary(canceled, root, a); err == nil {
		t.Fatal("canceled publication")
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	for _, entry := range entries {
		if entry.Name() != id {
			t.Fatal("canceled bundle residue")
		}
	}
	if err = os.WriteFile(filepath.Join(path, bundleSummaryName), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = PublicSummary(path); err == nil {
		t.Fatal("tamper exported")
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
		err := e.NetworkScan(context.Background())
		if (err == nil) != (i == 0) {
			t.Fatalf("projection %d %v", i, err)
		}
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

func TestQRPassesSameExplicitBindingToPrivateChild(t *testing.T) {
	e, script := newProcessTestExecutor(t, processTestStep{})
	e.expectedUsername = "synthetic"
	e.BindIP = netip.MustParseAddr("192.0.2.11")
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
