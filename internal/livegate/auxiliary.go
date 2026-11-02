package livegate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type AuxiliaryCheck struct {
	Name   string `json:"name"`
	Result Result `json:"result"`
}
type InstallVerifier struct {
	SourceTree       string `json:"source_tree"`
	SnapshotSHA256   string `json:"snapshot_sha256"`
	GoVersion        string `json:"go_version"`
	TestBinarySHA256 string `json:"test_binary_sha256"`
}
type Auxiliary struct {
	SchemaVersion      int              `json:"schema_version"`
	PlanID             string           `json:"plan_id"`
	Revision           string           `json:"revision"`
	SummaryID          string           `json:"summary_id"`
	Kind               string           `json:"kind"`
	CandidateID        string           `json:"candidate_id"`
	CandidateSetSHA256 string           `json:"candidate_set_sha256"`
	SourceCommit       string           `json:"source_commit"`
	BuildInputSHA256   string           `json:"build_input_sha256"`
	RunnerSHA256       string           `json:"runner_sha256"`
	ExecutionContext   string           `json:"execution_context"`
	Result             Result           `json:"result"`
	StartedAt          Timestamp        `json:"started_at"`
	FinishedAt         Timestamp        `json:"finished_at"`
	Checks             []AuxiliaryCheck `json:"checks"`
	Verifier           *InstallVerifier `json:"verifier"`
	WiFiSSID           *string          `json:"wifi_ssid,omitempty"`
}

var InstallChecks = []string{"candidate_verified", "fresh_install", "three_entry_versions", "active_junction_execution", "fresh_default_meta", "upgrade_preserves_launcher", "basic_rollback", "private_layout", "real_user_path_unchanged", "candidate_unchanged"}
var WiFiChecks = []string{"candidate_verified", "selected_wifi_topology", "list_complete", "selected_address_unique", "bound_status_valid", "scan_scope_valid", "scan_complete", "selected_result_valid", "topology_unchanged", "candidate_unchanged"}

func (a Auxiliary) Validate() error {
	if (a.SchemaVersion != 1 && a.SchemaVersion != 2) || a.PlanID != PlanID || a.Revision != Revision || a.ExecutionContext != "windows_native" || !a.Result.valid() {
		return ErrInvalidEvidence
	}
	if a.SchemaVersion == 1 && a.WiFiSSID != nil {
		return ErrInvalidEvidence
	}
	if a.SchemaVersion == 2 && !validWiFiIdentity(a.Kind == "bhk_wifi_network", a.WiFiSSID) {
		return ErrInvalidEvidence
	}
	start, err := parseTimestamp(a.StartedAt)
	if err != nil {
		return ErrInvalidEvidence
	}
	end, err := parseTimestamp(a.FinishedAt)
	if err != nil || end.Before(start) {
		return ErrInvalidEvidence
	}
	if !strings.HasPrefix(a.SummaryID, "AUX-") || validateEvidenceID("EVID-"+strings.TrimPrefix(a.SummaryID, "AUX-"), start) != nil {
		return ErrInvalidEvidence
	}
	if !lowerHex40Pattern.MatchString(a.SourceCommit) || validateCandidateID(a.CandidateID, a.SourceCommit) != nil {
		return ErrInvalidEvidence
	}
	for _, hash := range []string{a.CandidateSetSHA256, a.BuildInputSHA256, a.RunnerSHA256} {
		if !lowerHex64Pattern.MatchString(hash) {
			return ErrInvalidEvidence
		}
	}
	var names []string
	switch a.Kind {
	case "bhk_windows_native_install":
		names = InstallChecks
		if a.Verifier != nil && (!lowerHex40Pattern.MatchString(a.Verifier.SourceTree) || !lowerHex64Pattern.MatchString(a.Verifier.SnapshotSHA256) || !lowerHex64Pattern.MatchString(a.Verifier.TestBinarySHA256) || a.Verifier.GoVersion != "go1.25.0 windows/amd64") {
			return ErrInvalidEvidence
		}
		if a.Result == ResultPass && a.Verifier == nil {
			return ErrInvalidEvidence
		}
	case "bhk_wifi_network":
		names = WiFiChecks
		if a.Verifier != nil {
			return ErrInvalidEvidence
		}
	default:
		return ErrInvalidEvidence
	}
	if len(a.Checks) == 0 || len(a.Checks) > len(names) {
		return ErrInvalidEvidence
	}
	for i, check := range a.Checks {
		if check.Name != names[i] || !check.Result.valid() || (i < len(a.Checks)-1 && check.Result != ResultPass) {
			return ErrInvalidEvidence
		}
	}
	if a.Result == ResultPass {
		if len(a.Checks) != len(names) || a.Checks[len(a.Checks)-1].Result != ResultPass {
			return ErrInvalidEvidence
		}
	} else if a.Checks[len(a.Checks)-1].Result != a.Result {
		return ErrInvalidEvidence
	}
	return nil
}

func DecodeAuxiliary(raw []byte) (Auxiliary, error) {
	var a Auxiliary
	if len(raw) > MaxEvidenceJSONBytes || scanSingleJSONValue(raw) != nil {
		return a, ErrInvalidEvidenceJSON
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&a) != nil {
		return Auxiliary{}, ErrInvalidEvidenceJSON
	}
	// Exact canonical encoding also rejects missing/null/duplicate/case aliases.
	encoded, err := json.Marshal(a)
	if err != nil || !bytes.Equal(append(encoded, '\n'), raw) || a.Validate() != nil {
		return Auxiliary{}, ErrInvalidEvidenceJSON
	}
	return a, nil
}

func auxiliarySummary(a Auxiliary) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "# IPGW-Meta auxiliary summary\n\n- schema_version: %d\n- summary_id: %s\n- kind: %s\n- candidate_id: %s\n- result: %s\n\n", a.SchemaVersion, a.SummaryID, a.Kind, a.CandidateID, a.Result)
	for _, c := range a.Checks {
		fmt.Fprintf(&b, "- %s | %s\n", c.Name, c.Result)
	}
	return []byte(b.String())
}

// PublishAuxiliary uses the same private, durable, no-clobber publication
// primitives as authentication bundles; no raw command output is accepted.
func PublishAuxiliary(ctx context.Context, output string, a Auxiliary) (string, error) {
	if ctx == nil || a.SchemaVersion != 2 || a.SummaryID != "" || validateBundleOutputDir(output) != nil {
		return "", ErrEvidenceDurability
	}
	start, err := parseTimestamp(a.StartedAt)
	if err != nil {
		return "", err
	}
	if err = ensureBundlePrivateDirectory(output); err != nil {
		return "", err
	}
	parent := filepath.Join(output, a.CandidateID)
	// Validate identity before using it as a path component.
	a.SummaryID = fmt.Sprintf("AUX-%s-001", start.Format("20060102"))
	if a.Validate() != nil {
		return "", ErrInvalidEvidence
	}
	if err = ensureBundlePrivateDirectory(parent); err != nil {
		return "", err
	}
	for sequence := 1; sequence <= maxBundleSequence; sequence++ {
		a.SummaryID = fmt.Sprintf("AUX-%s-%03d", start.Format("20060102"), sequence)
		path := filepath.Join(parent, a.SummaryID)
		exists, e := bundlePathExists(path)
		if e != nil {
			return "", e
		}
		if exists {
			continue
		}
		reservation := filepath.Join(parent, "."+a.SummaryID+".reserve")
		if e = reserveBundleID(reservation); e == errBundleCollision {
			continue
		} else if e != nil {
			return "", e
		}
		raw, e := json.Marshal(a)
		if e != nil {
			return "", ErrInvalidEvidence
		}
		raw = append(raw, '\n')
		if _, e = DecodeAuxiliary(raw); e != nil {
			return "", e
		}
		if e = publishAuxiliaryFiles(ctx, parent, path, reservation, raw, auxiliarySummary(a)); e != nil {
			return "", e
		}
		return a.SummaryID, nil
	}
	return "", ErrEvidenceDurability
}

func publishAuxiliaryFiles(ctx context.Context, parent, final, reservation string, raw, summary []byte) (result error) {
	ops := defaultBundleOps()
	stage := ""
	published, success := false, false
	defer func() {
		if success {
			return
		}
		if published {
			_ = invalidatePublishedBundleWithOps(ops, final, parent)
		}
		if stage != "" {
			_ = ops.removeAll(stage)
		}
		_ = ops.removeFile(reservation)
		_ = ops.syncDirectory(parent)
		result = ErrEvidenceDurability
	}()
	var err error
	stage, err = createBundleStagingDirectory(parent)
	if err != nil {
		return err
	}
	files := map[string][]byte{bundleEvidenceName: raw, bundleSummaryName: summary, bundleChecksumsName: buildBundleChecksums(raw, summary)}
	for _, name := range []string{bundleEvidenceName, bundleSummaryName, bundleChecksumsName} {
		if err = ops.writeFile(filepath.Join(stage, name), files[name]); err != nil {
			return err
		}
	}
	if err = verifyCompleteStagingBundle(ops, stage, files); err != nil {
		return err
	}
	if err = ops.syncDirectory(stage); err != nil {
		return err
	}
	if err = bundleContextError(ctx); err != nil {
		return err
	}
	if err = ops.publishDirectory(stage, final); err != nil {
		return err
	}
	published = true
	stage = ""
	if err = verifyCompleteStagingBundle(ops, final, files); err != nil {
		return err
	}
	if err = ops.removeFile(reservation); err != nil {
		return err
	}
	if err = ops.syncDirectory(parent); err != nil {
		return err
	}
	success = true
	return nil
}

// ReadAuxiliaryBundle never mutates the input and verifies the exact trio.
func ReadAuxiliaryBundle(path string) (Auxiliary, [][]byte, error) {
	var a Auxiliary
	if !validCandidatePath(path) || verifyBundlePrivateDirectory(path) != nil {
		return a, nil, ErrInvalidEvidence
	}
	shape, err := bundleHasExactThreeFileShape(path)
	if err != nil || !shape {
		return a, nil, ErrInvalidEvidence
	}
	var files [][]byte
	for _, name := range []string{bundleEvidenceName, bundleSummaryName, bundleChecksumsName} {
		raw, e := readPrivateBundleFile(filepath.Join(path, name))
		if e != nil {
			return a, nil, e
		}
		files = append(files, raw)
	}
	a, err = DecodeAuxiliary(files[0])
	if err != nil {
		return a, nil, err
	}
	if !bytes.Equal(files[1], auxiliarySummary(a)) || !bytes.Equal(files[2], buildBundleChecksums(files[0], files[1])) {
		return a, nil, ErrInvalidEvidence
	}
	return a, files, nil
}
