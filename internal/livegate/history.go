package livegate

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
)

// AuditHistoricalBundle validates only existing local v1 material. It never
// emits, upgrades, downloads, publishes or asserts provenance for that material.
func AuditHistoricalBundle(path string) error {
	return auditHistoricalBundleVersion(path, 1)
}

func auditHistoricalBundleVersion(path string, version int) error {
	if version != 1 && version != 2 && version != -1 {
		return ErrInvalidEvidence
	}
	if !validCandidatePath(path) || verifyBundlePrivateDirectory(path) != nil {
		return ErrInvalidEvidence
	}
	shape, err := bundleHasExactThreeFileShape(path)
	if err != nil || !shape {
		return ErrInvalidEvidence
	}
	raw, err := readPrivateBundleFile(filepath.Join(path, bundleEvidenceName))
	if err != nil {
		return ErrInvalidEvidence
	}
	var expected []byte
	if version == -1 {
		a, e := DecodeAuxiliary(raw)
		if e != nil || a.SchemaVersion != 1 {
			return ErrInvalidEvidence
		}
		expected = auxiliarySummary(a)
	} else {
		e, err := DecodeEvidence(bytes.NewReader(raw))
		if err != nil || e.SchemaVersion != version {
			return ErrInvalidEvidence
		}
		expected = buildBundleSummary(e)
	}
	summary, err := readPrivateBundleFile(filepath.Join(path, bundleSummaryName))
	if err != nil || !bytes.Equal(summary, expected) {
		return ErrInvalidEvidence
	}
	sums, err := readPrivateBundleFile(filepath.Join(path, bundleChecksumsName))
	if err != nil || !bytes.Equal(sums, buildBundleChecksums(raw, summary)) {
		return ErrInvalidEvidence
	}
	return nil
}

func executeHistory(ctx context.Context, options Options) GateExitCode {
	if len(options.Args) != 3 || (options.Args[0] != "audit-v1" && options.Args[0] != "audit-v2" && options.Args[0] != "audit-aux-v1") || options.Args[1] != "--bundle" || options.Stdout == nil {
		return renderGateExit(options, GateExitSecurityReject, "", "")
	}
	if ctx.Err() != nil {
		return renderGateExit(options, GateExitCanceled, "", "")
	}
	state, code := "valid", GateExitPass
	version := 1
	if options.Args[0] == "audit-v2" {
		version = 2
	}
	mode := fmt.Sprintf("historical-v%d", version)
	if options.Args[0] == "audit-aux-v1" {
		version, mode = -1, "historical-aux-v1"
	}
	if auditHistoricalBundleVersion(options.Args[2], version) != nil {
		state, code = "invalid", GateExitSecurityReject
	}
	if ctx.Err() != nil {
		return renderGateExit(options, GateExitCanceled, "", "")
	}
	if _, err := fmt.Fprintf(options.Stdout, "{\"schema_version\":1,\"mode\":%q,\"state\":%q,\"scope\":\"local_bundle\",\"provenance\":\"unverified\"}\n", mode, state); err != nil {
		return GateExitInternal
	}
	return code
}
