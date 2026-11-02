package livegate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
)

// Bundle digest framing is fixed by EVID-REDACT-001, not a hash of paths.
func bundleDigest(files [][]byte) string {
	h := sha256.New()
	for _, i := range []int{2, 0, 1} {
		name := []string{bundleEvidenceName, bundleSummaryName, bundleChecksumsName}[i]
		s := sha256.Sum256(files[i])
		fmt.Fprintf(h, "%s\x00%d\x00%x\n", name, len(files[i]), s)
	}
	return hex.EncodeToString(h.Sum(nil))
}

type publicEvidence struct {
	SchemaVersion      int                `json:"schema_version"`
	PlanID             string             `json:"plan_id"`
	Revision           string             `json:"revision"`
	EvidenceID         string             `json:"evidence_id"`
	CandidateID        string             `json:"candidate_id"`
	CandidateSetSHA256 string             `json:"candidate_set_sha256"`
	SourceCommit       string             `json:"source_commit"`
	RunnerSHA256       string             `json:"runner_sha256"`
	Platform           Platform           `json:"platform"`
	Testbed            Testbed            `json:"testbed"`
	NetworkType        NetworkType        `json:"network_type"`
	AuthMethod         AuthMethod         `json:"auth_method"`
	Suite              Suite              `json:"suite"`
	Result             Result             `json:"result"`
	CapabilityBefore   []CapabilityStatus `json:"capability_before"`
	CapabilityAfter    []CapabilityStatus `json:"capability_after"`
	StartedAt          Timestamp          `json:"started_at"`
	FinishedAt         Timestamp          `json:"finished_at"`
	BundleSHA256       string             `json:"bundle_sha256"`
	WiFiSSID           *string            `json:"wifi_ssid"`
}

// PublicSummary only reads verified local bundles; callers review this closed
// projection before saving it in the release evidence directory.
func PublicSummary(path string) ([]byte, error) {
	if !validCandidatePath(path) || verifyBundlePrivateDirectory(path) != nil {
		return nil, ErrInvalidEvidence
	}
	shape, err := bundleHasExactThreeFileShape(path)
	if err != nil || !shape {
		return nil, ErrInvalidEvidence
	}
	var files [][]byte
	for _, name := range []string{bundleEvidenceName, bundleSummaryName, bundleChecksumsName} {
		data, e := readPrivateBundleFile(filepath.Join(path, name))
		if e != nil {
			return nil, e
		}
		files = append(files, data)
	}
	if !bytes.Equal(files[2], buildBundleChecksums(files[0], files[1])) {
		return nil, ErrInvalidEvidence
	}
	if a, err := DecodeAuxiliary(files[0]); err == nil {
		if a.SchemaVersion != 2 {
			return nil, ErrInvalidEvidence
		}
		if filepath.Base(path) != a.SummaryID || filepath.Base(filepath.Dir(path)) != a.CandidateID {
			return nil, ErrInvalidEvidence
		}
		if !bytes.Equal(files[1], auxiliarySummary(a)) {
			return nil, ErrInvalidEvidence
		}
		raw, e := json.Marshal(a)
		if e != nil {
			return nil, e
		}
		raw = append(raw[:len(raw)-1], []byte(fmt.Sprintf(",\"bundle_sha256\":%q}\n", bundleDigest(files)))...)
		return raw, nil
	}
	e, err := DecodeEvidence(bytes.NewReader(files[0]))
	if err != nil || e.SchemaVersion != SchemaVersion || !bytes.Equal(files[1], buildBundleSummary(e)) {
		return nil, ErrInvalidEvidence
	}
	if filepath.Base(path) != e.EvidenceID || filepath.Base(filepath.Dir(path)) != e.CandidateID {
		return nil, ErrInvalidEvidence
	}
	out := publicEvidence{e.SchemaVersion, e.PlanID, e.Revision, e.EvidenceID, e.CandidateID, e.CandidateSetSHA256, e.SourceCommit, e.RunnerSHA256, e.Platform, e.Testbed, e.NetworkType, e.AuthMethod, e.Suite, e.Result, e.CapabilityBefore, e.CapabilityAfter, e.StartedAt, e.FinishedAt, bundleDigest(files), e.WiFiSSID}
	raw, err := json.Marshal(out)
	return append(raw, '\n'), err
}
