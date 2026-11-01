// Package casprovenance classifies bounded, in-memory CAS sources. Its standalone
// command remains synthetic-only; casacquire projects classified facts into a
// separate acquisition envelope. No acquisition, credential, script execution
// or key-selection capability is implemented here.
package casprovenance

import "errors"

const (
	MaxSources     = 17
	MaxSourceBytes = 1 << 20
	MaxTotalBytes  = 4 << 20
	MaxOccurrences = 64
	MaxTokens      = 100000
	MaxDepth       = 64
	MaxReportBytes = 65536
)

var (
	ErrInvalid = errors.New("invalid synthetic input")
	ErrLimit   = errors.New("synthetic analysis resource limit")
	ErrReport  = errors.New("invalid provenance report")
)

// Input is an internal in-memory boundary, never a CLI file/URL input.
type Input struct {
	Kind string
	Data []byte
}
type Coverage struct {
	Lexical     string `json:"lexical"`
	Assignment  string `json:"assignment"`
	UseRelation string `json:"use_relation"`
}
type Source struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Ordinal  int    `json:"ordinal"`
	Analysis string `json:"analysis"`
}
type Occurrence struct {
	SourceID        string  `json:"source_id"`
	Ordinal         int     `json:"ordinal"`
	Context         string  `json:"context"`
	IdentifierClass string  `json:"identifier_class"`
	AssignmentClass string  `json:"assignment_class"`
	KeyClass        string  `json:"key_class"`
	KeyRef          *string `json:"key_ref"`
}
type Key struct {
	ID            string `json:"id"`
	Bits          int    `json:"rsa_bits"`
	ExponentClass string `json:"exponent_class"`
}

// No relations are emitted until independent rules can prove them.
type Relation struct {
	Definition int    `json:"definition"`
	Use        int    `json:"use"`
	Kind       string `json:"kind"`
}
type Operations struct {
	ExternalRequests int `json:"external_requests"`
	DNS              int `json:"dns"`
	ProfileReads     int `json:"profile_reads"`
	CredentialReads  int `json:"credential_reads"`
	CASPosts         int `json:"cas_posts"`
	Activation       int `json:"activation"`
	Logout           int `json:"logout"`
}
type Report struct {
	SchemaVersion int          `json:"schema_version"`
	ToolKind      string       `json:"tool_kind"`
	Mode          string       `json:"mode"`
	BundleID      *string      `json:"bundle_id"`
	Binding       string       `json:"binding"`
	Analysis      string       `json:"analysis"`
	Coverage      Coverage     `json:"coverage"`
	Reason        string       `json:"reason"`
	Sources       []Source     `json:"sources"`
	Occurrences   []Occurrence `json:"occurrences"`
	Keys          []Key        `json:"keys"`
	Relations     []Relation   `json:"relations"`
	Operations    Operations   `json:"operations"`
}

func emptyReport() Report {
	return Report{SchemaVersion: 1, ToolKind: "cas-provenance-dev", Mode: "synthetic-prototype", Binding: "not_packaged", Analysis: "partial", Coverage: Coverage{"complete", "complete", "unknown"}, Reason: "use_relation_unknown", Sources: []Source{}, Occurrences: []Occurrence{}, Keys: []Key{}, Relations: []Relation{}}
}
