package livegate

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testHash(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

func TestHistoricalV1CannotBeEmittedOrReinterpreted(t *testing.T) {
	e := validPasswordEvidence()
	e.SchemaVersion = 1
	e.RunnerSHA256 = ""
	raw, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeEvidence(bytes.NewReader(raw))
	if err != nil || got.SchemaVersion != 1 {
		t.Fatal("historical read", err)
	}
	if EncodeEvidence(&bytes.Buffer{}, got) == nil {
		t.Fatal("historical emission allowed")
	}
	if _, err := (BundlePublisher{OutputDir: bundleTestOutputDir(t)}).Publish(got); err == nil {
		t.Fatal("historical publication allowed")
	}
	qr := validTerminalQREvidence()
	qr.Platform = PlatformWindowsAMD64
	qr.Testbed = TestbedBHKWindows
	if qr.Validate() != nil {
		t.Fatal("v2 Windows QR rejected")
	}
	qr.SchemaVersion = 1
	qr.RunnerSHA256 = ""
	if qr.Validate() == nil {
		t.Fatal("v1 Windows QR reinterpreted")
	}
	e.RunnerSHA256 = strings.Repeat("a", 64)
	if e.Validate() == nil {
		t.Fatal("v1 accepted v2 field")
	}
	e.SchemaVersion = 2
	e.RunnerSHA256 = ""
	if e.Validate() == nil {
		t.Fatal("v2 missing helper binding")
	}
}

func TestHistoricalBundleAuditIsReadOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history")
	if err := ensureBundlePrivateDirectory(path); err != nil {
		t.Fatal(err)
	}
	e := validPasswordEvidence()
	e.SchemaVersion = 1
	e.RunnerSHA256 = ""
	raw, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	raw = append(raw, '\n')
	summary := buildBundleSummary(e)
	files := map[string][]byte{bundleEvidenceName: raw, bundleSummaryName: summary, bundleChecksumsName: buildBundleChecksums(raw, summary)}
	for name, data := range files {
		if err := writePrivateBundleFile(filepath.Join(path, name), data); err != nil {
			t.Fatal(err)
		}
	}
	if err := AuditHistoricalBundle(path); err != nil {
		t.Fatal(err)
	}
	for name, data := range files {
		after, err := os.ReadFile(filepath.Join(path, name))
		if err != nil || !bytes.Equal(data, after) {
			t.Fatal("audit mutated input")
		}
	}
	if err := os.WriteFile(filepath.Join(path, bundleSummaryName), []byte("altered"), 0600); err != nil {
		t.Fatal(err)
	}
	if AuditHistoricalBundle(path) == nil {
		t.Fatal("altered historical summary accepted")
	}
}

func TestPathLossStopsCleanupWithoutAnotherCommand(t *testing.T) {
	inner := &scriptedPreparedExecutor{observations: successfulObservations(SuitePasswordCore)}
	valid := true
	guard := &pathExecutor{inner: inner, check: func(context.Context) bool { return valid }, interval: time.Hour}
	if guard.Execute(context.Background(), ActionInitialStatus).RunnerFault {
		t.Fatal("valid path rejected")
	}
	valid = false
	for _, a := range []CommandAction{ActionPasswordLogin, ActionCleanupLogout, ActionCleanupStatus} {
		if !guard.Execute(context.Background(), a).RunnerFault {
			t.Fatal("lost path accepted")
		}
	}
	if len(inner.actions) != 1 {
		t.Fatal("commands ran after path loss")
	}
}

func TestV2RunnerBindingDetectsMismatchAndDrift(t *testing.T) {
	fixture := newCandidateTestFixture(t)
	v, err := VerifyCandidateBefore(fixture.candidate, fixture.manifest, fixture.explicitSHA, fixture.platform)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(fixture.directory, filepath.Base(v.Binding.RunnerName))
	data := []byte("helper-bytes")
	writePrivateTestExecutableFile(t, path, data)
	if _, err := bindRunner(v, path); err == nil {
		t.Fatal("wrong helper accepted")
	}
	// Rebind the synthetic manifest to the independent helper bytes.
	object := mustJSONObject(t, fixture.manifestRaw)
	items := mustJSONArrayField(t, object, "test_tools")
	for _, item := range items {
		m := item.(map[string]any)
		m["sha256"] = testHash(data)
		m["size"] = len(data)
	}
	writePrivateTestReadableFile(t, fixture.manifest, mustMarshalJSON(t, object))
	v, err = VerifyCandidateBefore(fixture.candidate, fixture.manifest, fixture.explicitSHA, fixture.platform)
	if err != nil {
		t.Fatal(err)
	}
	v, err = bindRunner(v, path)
	if err != nil {
		t.Fatal(err)
	}
	if err = v.VerifyAfter(); err != nil {
		t.Fatal(err)
	}
	writePrivateTestExecutableFile(t, path, []byte("helper-other"))
	if v.VerifyAfter() == nil {
		t.Fatal("helper drift accepted")
	}
}
