package casbundle

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
)

type InstallReceipt struct {
	SchemaVersion    int                      `json:"schema_version"`
	ToolKind         string                   `json:"tool_kind"`
	State            string                   `json:"state"`
	Reason           string                   `json:"reason"`
	ArchiveVerified  bool                     `json:"archive_verified"`
	BundleID         *string                  `json:"bundle_id"`
	DirectoryCreated bool                     `json:"directory_created"`
	FilesWritten     int                      `json:"files_written"`
	Operations       casprovenance.Operations `json:"operations"`
	SelfTest         *Receipt                 `json:"self_test"`
}

// Install creates a fresh development directory only; it never overwrites,
// removes, activates, upgrades, or installs over a failed prior attempt.
func Install(ctx context.Context, archivePath, destination, archiveDigest, manifestDigest string) InstallReceipt {
	return install(ctx, archivePath, destination, archiveDigest, manifestDigest, nil)
}

// hooks exist only for deterministic synthetic faults, not CLI/env switches.
type installHooks struct {
	afterCreate    func()
	afterWrite     func(int)
	beforeSelfTest func()
}

func install(ctx context.Context, archivePath, destination, archiveDigest, manifestDigest string, hooks *installHooks) InstallReceipt {
	out := InstallReceipt{SchemaVersion: 1, ToolKind: ToolKind, State: "rejected", Reason: "archive_invalid"}
	if ctx == nil || ctx.Err() != nil {
		out.Reason = "canceled"
		return out
	}
	if !digestPattern.MatchString(archiveDigest) || !digestPattern.MatchString(manifestDigest) {
		return out
	}
	source, err := openArchive(archivePath)
	if err != nil {
		if errors.Is(err, ErrUnsupported) {
			out.Reason = "unsupported_platform"
		}
		return out
	}
	defer source.close()
	data, err := source.read("archive", MaxArchive)
	if err != nil {
		return out
	}
	archive, err := verifyArchive(ctx, data, archiveDigest, manifestDigest)
	if err != nil {
		if ctx.Err() != nil {
			out.Reason = "canceled"
		}
		return out
	}
	out.ArchiveVerified = true
	out.BundleID = &archive.manifest.BundleID
	if ctx.Err() != nil {
		out.Reason = "canceled"
		return out
	}
	target, err := createTarget(destination)
	if target != nil {
		defer target.close()
		out.DirectoryCreated = target.created
	}
	if err != nil {
		out.Reason = "destination_invalid"
		if out.DirectoryCreated {
			out.State = "incomplete"
		}
		return out
	}
	out.DirectoryCreated = true
	out.State = "incomplete"
	out.Reason = "copy_failed"
	if hooks != nil && hooks.afterCreate != nil {
		hooks.afterCreate()
	}
	for _, name := range names {
		if ctx.Err() != nil {
			out.Reason = "canceled"
			return out
		}
		if err := target.write(name, archive.files[name]); err != nil {
			return out
		}
		out.FilesWritten++
		if hooks != nil && hooks.afterWrite != nil {
			hooks.afterWrite(out.FilesWritten)
		}
	}
	if hooks != nil && hooks.beforeSelfTest != nil {
		hooks.beforeSelfTest()
	}
	if ctx.Err() != nil {
		out.Reason = "canceled"
		return out
	}
	result := Run(ctx, destination, manifestDigest)
	out.SelfTest = &result
	if result.Runner != "passed" {
		out.Reason = "self_test_failed"
		return out
	}
	out.State = "installed"
	out.Reason = "none"
	return out
}

func validInstallReceipt(r InstallReceipt) bool {
	if r.SchemaVersion != 1 || r.ToolKind != ToolKind || r.Operations != (casprovenance.Operations{}) || r.FilesWritten < 0 || r.FilesWritten > 4 {
		return false
	}
	if r.ArchiveVerified != (r.BundleID != nil) {
		return false
	}
	if r.BundleID != nil && (len(*r.BundleID) != 68 || (*r.BundleID)[:4] != "dev-" || !digestPattern.MatchString((*r.BundleID)[4:])) {
		return false
	}
	if !r.DirectoryCreated && (r.FilesWritten != 0 || r.SelfTest != nil) {
		return false
	}
	if r.DirectoryCreated && !r.ArchiveVerified {
		return false
	}
	if r.SelfTest != nil && (!validReceipt(*r.SelfTest) || !r.DirectoryCreated || r.FilesWritten != 4) {
		return false
	}
	if r.SelfTest != nil && r.SelfTest.BundleID != nil && *r.SelfTest.BundleID != *r.BundleID {
		return false
	}
	switch r.State {
	case "rejected":
		if r.DirectoryCreated {
			return false
		}
		switch r.Reason {
		case "archive_invalid", "unsupported_platform":
			return !r.ArchiveVerified
		case "destination_invalid":
			return r.ArchiveVerified
		case "canceled":
			return true
		}
	case "incomplete":
		if !r.DirectoryCreated {
			return false
		}
		switch r.Reason {
		case "destination_invalid":
			return r.FilesWritten == 0 && r.SelfTest == nil
		case "copy_failed":
			return r.FilesWritten < 4 && r.SelfTest == nil
		case "canceled":
			return r.SelfTest == nil
		case "self_test_failed":
			return r.SelfTest != nil && r.SelfTest.Runner != "passed"
		}
	case "installed":
		return r.Reason == "none" && r.DirectoryCreated && r.FilesWritten == 4 && r.SelfTest != nil && r.SelfTest.Runner == "passed" && r.SelfTest.BundleID != nil && *r.SelfTest.BundleID == *r.BundleID
	}
	return false
}

func EncodeInstallReceipt(r InstallReceipt) ([]byte, error) {
	if !validInstallReceipt(r) {
		return nil, ErrInvalid
	}
	data, err := json.Marshal(r)
	if err != nil || len(data)+1 > MaxMetadata {
		return nil, ErrInvalid
	}
	return append(data, '\n'), nil
}
func DecodeInstallReceipt(data []byte) (InstallReceipt, error) {
	var r InstallReceipt
	if casprovenance.DecodeClosedJSON(data, &r, MaxMetadata) != nil || !validInstallReceipt(r) {
		return InstallReceipt{}, ErrInvalid
	}
	return r, nil
}
