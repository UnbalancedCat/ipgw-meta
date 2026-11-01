// Package casbundle validates a synthetic-only development bundle.
// An explicit external manifest digest is required; local consistency is not trust.
package casbundle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"path"
	"regexp"
	"strings"

	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
)

const MaxMetadata = 131072
const MaxExecutable = 32 << 20
const ToolKind = "cas-provenance-dev"

var ErrInvalid = errors.New("development bundle rejected")
var ErrUnsupported = errors.New("protected execution unavailable")
var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
var headPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
var sourcePattern = regexp.MustCompile(`^[a-zA-Z0-9_./-]+$`)
var names = []string{"build-inputs.json", "cas-provenance.exe", "manifest.json", "runner.json"}
var flags = []string{"-trimpath", "-buildvcs=false", "-ldflags=-s -w"}

type FileRecord struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}
type Manifest struct {
	SchemaVersion int          `json:"schema_version"`
	ToolKind      string       `json:"tool_kind"`
	BundleID      string       `json:"bundle_id"`
	Platform      string       `json:"platform"`
	Files         []FileRecord `json:"files"`
}
type Runner struct {
	SchemaVersion  int      `json:"schema_version"`
	ToolKind       string   `json:"tool_kind"`
	Mode           string   `json:"mode"`
	Executable     string   `json:"executable"`
	Size           int64    `json:"size"`
	SHA256         string   `json:"sha256"`
	Args           []string `json:"args"`
	TimeoutSeconds int      `json:"timeout_seconds"`
	MaxOutputBytes int      `json:"max_output_bytes"`
}
type BuildInputs struct {
	SchemaVersion int          `json:"schema_version"`
	ToolKind      string       `json:"tool_kind"`
	GoVersion     string       `json:"go_version"`
	Platform      string       `json:"platform"`
	CGOEnabled    string       `json:"cgo_enabled"`
	Head          string       `json:"head"`
	Worktree      string       `json:"worktree"`
	SourceDigest  string       `json:"source_digest"`
	Sources       []FileRecord `json:"sources"`
	BuildFlags    []string     `json:"build_flags"`
}

func Digest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func validRecord(f FileRecord, limit int64) bool {
	return f.Size > 0 && f.Size <= limit && digestPattern.MatchString(f.SHA256)
}

// ValidateMetadata validates bytes already read from protected handles. It does
// not claim the supplied digest was independently reviewed by a human.
func ValidateMetadata(manifest, runner, inputs []byte, expected string) (Manifest, Runner, error) {
	var m Manifest
	var r Runner
	var b BuildInputs
	fail := func() (Manifest, Runner, error) { return Manifest{}, Runner{}, ErrInvalid }
	if !digestPattern.MatchString(expected) || Digest(manifest) != expected {
		return fail()
	}
	if casprovenance.DecodeClosedJSON(manifest, &m, MaxMetadata) != nil || casprovenance.DecodeClosedJSON(runner, &r, MaxMetadata) != nil || casprovenance.DecodeClosedJSON(inputs, &b, MaxMetadata) != nil {
		return fail()
	}
	if m.SchemaVersion != 1 || m.ToolKind != ToolKind || m.Platform != "windows/amd64" || m.BundleID != "dev-"+Digest(inputs) || len(m.Files) != 3 {
		return fail()
	}
	for i, name := range []string{"build-inputs.json", "cas-provenance.exe", "runner.json"} {
		limit := int64(MaxMetadata)
		if name == "cas-provenance.exe" {
			limit = MaxExecutable
		}
		if m.Files[i].Name != name || !validRecord(m.Files[i], limit) {
			return fail()
		}
	}
	for i, data := range [][]byte{inputs, runner} {
		index := 0
		if i == 1 {
			index = 2
		}
		if m.Files[index].Size != int64(len(data)) || m.Files[index].SHA256 != Digest(data) {
			return fail()
		}
	}
	if r.SchemaVersion != 1 || r.ToolKind != ToolKind || r.Mode != "synthetic-only" || r.Executable != "cas-provenance.exe" || r.Size != m.Files[1].Size || r.SHA256 != m.Files[1].SHA256 || len(r.Args) != 1 || r.Args[0] != "--self-test" || r.TimeoutSeconds != 30 || r.MaxOutputBytes != casprovenance.MaxReportBytes {
		return fail()
	}
	if b.SchemaVersion != 1 || b.ToolKind != ToolKind || b.GoVersion != "go1.25.0" || b.Platform != "windows/amd64" || b.CGOEnabled != "0" || !headPattern.MatchString(b.Head) || (b.Worktree != "dirty" && b.Worktree != "clean") || len(b.Sources) < 1 || len(b.Sources) > 256 || len(b.BuildFlags) != len(flags) {
		return fail()
	}
	for i := range flags {
		if b.BuildFlags[i] != flags[i] {
			return fail()
		}
	}
	previous := ""
	for _, f := range b.Sources {
		if !sourcePattern.MatchString(f.Name) || path.Clean(f.Name) != f.Name || strings.HasPrefix(f.Name, "/") || f.Name == ".." || strings.HasPrefix(f.Name, "../") || f.Name <= previous || !validRecord(f, 8<<20) {
			return fail()
		}
		previous = f.Name
	}
	encoded, err := json.Marshal(b.Sources)
	if err != nil || !digestPattern.MatchString(b.SourceDigest) || b.SourceDigest != Digest(encoded) {
		return fail()
	}
	return m, r, nil
}
