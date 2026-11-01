//go:build linux || windows

package livegate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Private bundle and Candidate filesystem contracts are supported on Linux and
// Windows. Schema, process projection and topology tests remain cross-platform.
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
