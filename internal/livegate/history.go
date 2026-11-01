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
	e, err := DecodeEvidence(bytes.NewReader(raw))
	if err != nil || e.SchemaVersion != HistoricalSchemaVersion {
		return ErrInvalidEvidence
	}
	summary, err := readPrivateBundleFile(filepath.Join(path, bundleSummaryName))
	if err != nil || !bytes.Equal(summary, buildBundleSummary(e)) {
		return ErrInvalidEvidence
	}
	sums, err := readPrivateBundleFile(filepath.Join(path, bundleChecksumsName))
	if err != nil || !bytes.Equal(sums, buildBundleChecksums(raw, summary)) {
		return ErrInvalidEvidence
	}
	return nil
}

func executeHistory(ctx context.Context, options Options) GateExitCode {
	if len(options.Args) != 3 || options.Args[0] != "audit-v1" || options.Args[1] != "--bundle" || options.Stdout == nil {
		return renderGateExit(options, GateExitSecurityReject, "", "")
	}
	if ctx.Err() != nil {
		return renderGateExit(options, GateExitCanceled, "", "")
	}
	state, code := "valid", GateExitPass
	if AuditHistoricalBundle(options.Args[2]) != nil {
		state, code = "invalid", GateExitSecurityReject
	}
	if ctx.Err() != nil {
		return renderGateExit(options, GateExitCanceled, "", "")
	}
	if _, err := fmt.Fprintf(options.Stdout, "{\"schema_version\":1,\"mode\":\"historical-v1\",\"state\":%q,\"scope\":\"local_bundle\",\"provenance\":\"unverified\"}\n", state); err != nil {
		return GateExitInternal
	}
	return code
}
