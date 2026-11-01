// Package ctlcheck observes fixed local CTL material, never TLS acceptance.
package ctlcheck

import (
	"encoding/json"
	"errors"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
)

var ErrInvalid = errors.New("invalid ctl observation")

type Operations struct {
	NetworkRequests   int `json:"network_requests"`
	DNS               int `json:"dns"`
	Connects          int `json:"connects"`
	ProfileReads      int `json:"profile_reads"`
	CredentialReads   int `json:"credential_reads"`
	CertificateWrites int `json:"certificate_writes"`
	RegistryWrites    int `json:"registry_writes"`
}
type Source struct {
	Source                   string `json:"source"`
	State                    string `json:"state"`
	Reason                   string `json:"reason"`
	Read                     string `json:"read"`
	Structure                string `json:"structure"`
	Hierarchy                string `json:"hierarchy"`
	Signature                string `json:"signature"`
	CertificateTime          string `json:"certificate_time"`
	Conditions               string `json:"conditions"`
	Rejection                string `json:"rejection"`
	Window                   string `json:"window"`
	Usage                    string `json:"usage"`
	Attributes               string `json:"attributes"`
	MemberAlgorithm          string `json:"member_algorithm"`
	Extensions               string `json:"extensions"`
	MemberIdentifierCoverage string `json:"member_identifier_coverage"`
	SignedAttributeCoverage  string `json:"signed_attribute_coverage"`
}
type Report struct {
	SchemaVersion  int        `json:"schema_version"`
	ToolKind       string     `json:"tool_kind"`
	Mode           string     `json:"mode"`
	State          string     `json:"state"`
	Authentication string     `json:"authentication"`
	PurposePolicy  string     `json:"purpose_policy"`
	Revocation     string     `json:"revocation"`
	Latest         string     `json:"latest"`
	Sources        []Source   `json:"sources"`
	Operations     Operations `json:"operations"`
}

func emptySource(name string) Source {
	return Source{name, "unavailable", "read_failed", "not_checked", "unknown", "unknown", "unknown", "unknown", "unknown", "unknown", "unknown", "unknown", "unknown", "unknown", "unknown", "unknown", "unknown"}
}
func emptyReport(mode string) Report {
	return Report{SchemaVersion: 3, ToolKind: "ctl-check-dev", Mode: mode, State: "unavailable", Authentication: "blocked", PurposePolicy: "not_established", Revocation: "not_checked", Latest: "not_established", Sources: []Source{emptySource("authroot_ctl"), emptySource("disallowed_ctl")}}
}
func one(s string, values ...string) bool {
	for _, v := range values {
		if s == v {
			return true
		}
	}
	return false
}
func isParseFailure(s string) bool {
	if s == "parse_failed_ctl_algorithm_unsupported" || s == "parse_failed_ctl_extensions_unsupported" {
		return true
	}
	return one(s, "parse_failed_ctl_extensions", "parse_failed_ctl_extension", "parse_failed_ctl_extension_duplicate", "parse_failed_ctl_extension_limit", "parse_failed_envelope", "parse_failed_signed_data", "parse_failed_content_info", "parse_failed_content", "parse_failed_ctl_header", "parse_failed_ctl_usage", "parse_failed_ctl_identifier", "parse_failed_ctl_sequence", "parse_failed_ctl_time", "parse_failed_ctl_algorithm", "parse_failed_ctl_entries", "parse_failed_ctl_entry", "parse_failed_ctl_attributes", "parse_failed_ctl_attribute", "parse_failed_ctl_values", "parse_failed_ctl_tail", "parse_failed_certificate", "parse_failed_signed_attribute_order", "parse_failed_signed_attribute", "parse_failed_signed_attribute_identifier", "parse_failed_content_type", "parse_failed_signing_time", "parse_failed_required_attributes")
}
func isCMSUnsupported(s string) bool {
	return one(s, "unsupported_envelope", "unsupported_signed_data", "unsupported_digest_set", "unsupported_digest_algorithm", "unsupported_content_encoding", "unsupported_signer_info", "unsupported_signer_identifier", "unsupported_signer_digest", "unsupported_signed_attributes", "unsupported_signed_attribute_values", "unsupported_signed_attribute_unknown", "unsupported_signature_algorithm", "unsupported_signature_size")
}
func validSourceV2(s Source) bool {
	if !validSourceBase(s) || !one(s.MemberAlgorithm, "unknown", "sha1", "sha256", "sha384", "sha512") || !one(s.Extensions, "unknown", "absent", "present_uninterpreted", "critical_present_uninterpreted") {
		return false
	}
	if s.Signature != "verified" {
		return s.MemberAlgorithm == "unknown" && s.Extensions == "unknown"
	}
	return s.MemberAlgorithm != "unknown" && s.Extensions != "unknown"
}

// V3 reports only the explicitly bounded subset, never full structure validity.
func validSource(s Source) bool {
	if !one(s.Structure, "unknown", "bounded_checked") || !one(s.MemberIdentifierCoverage, "unknown", "exact_length_checked", "bounded_uninterpreted") || !one(s.SignedAttributeCoverage, "unknown", "known_subset_checked", "spc_opus_outer_only") {
		return false
	}
	check := s
	if check.Structure == "bounded_checked" {
		check.Structure = "valid"
	}
	if s.MemberAlgorithm == "signature_hash_property" {
		if s.Source != "disallowed_ctl" || s.Usage != "disallowed_list" || s.MemberIdentifierCoverage != "bounded_uninterpreted" {
			return false
		}
		check.MemberAlgorithm = "sha1" // Reuse historical stage validation, not algorithm semantics.
	}
	if !validSourceV2(check) {
		return false
	}
	if s.Signature != "verified" {
		return s.MemberIdentifierCoverage == "unknown" && s.SignedAttributeCoverage == "unknown"
	}
	if s.SignedAttributeCoverage == "unknown" {
		return false
	}
	if s.MemberAlgorithm == "signature_hash_property" {
		return true
	}
	return s.MemberIdentifierCoverage == "exact_length_checked"
}
func validSourceBase(s Source) bool {
	if !one(s.State, "complete", "unavailable") || (!isParseFailure(s.Reason) && !isCMSUnsupported(s.Reason) && !one(s.Reason, "none", "absent", "read_failed", "type_invalid", "size_changed", "limit", "parse_failed", "unsupported", "material_missing", "signer_ambiguous", "signature_failed", "certificate_time_invalid", "conditions_rejected", "denied", "store_unavailable", "canceled", "deadline", "protection_failed")) {
		return false
	}
	if (isParseFailure(s.Reason) || isCMSUnsupported(s.Reason)) && (s.Read != "complete" || s.Structure != "unknown" || s.State != "unavailable") {
		return false
	}
	if !one(s.Read, "complete", "absent", "unavailable", "not_checked") || !one(s.Structure, "valid", "unknown") || !one(s.Hierarchy, "observed", "unknown") || !one(s.Signature, "verified", "unknown") || !one(s.CertificateTime, "valid", "invalid", "unknown") || !one(s.Conditions, "satisfied", "unknown") || !one(s.Rejection, "none", "present", "unknown") || !one(s.Window, "within", "expired", "future", "unbounded", "invalid", "unknown") || !one(s.Usage, "root_list", "disallowed_list", "other", "unknown") || !one(s.Attributes, "absent", "present_uninterpreted", "unknown") {
		return false
	}
	if s.Read != "complete" && (s.Structure != "unknown" || s.Hierarchy != "unknown" || s.Signature != "unknown" || s.CertificateTime != "unknown" || s.Conditions != "unknown" || s.Rejection != "unknown" || s.Window != "unknown" || s.Usage != "unknown" || s.Attributes != "unknown") {
		return false
	}
	if s.Structure == "unknown" && (s.Hierarchy != "unknown" || s.Signature != "unknown" || s.CertificateTime != "unknown" || s.Conditions != "unknown" || s.Rejection != "unknown" || s.Window != "unknown" || s.Usage != "unknown" || s.Attributes != "unknown") {
		return false
	}
	if s.Signature != "verified" && (s.Window != "unknown" || s.Usage != "unknown" || s.Attributes != "unknown") {
		return false
	}
	if s.Signature == "verified" && (s.Hierarchy != "observed" || s.CertificateTime != "valid" || s.Conditions != "satisfied" || s.Rejection != "none") {
		return false
	}
	if s.Hierarchy == "observed" && (s.Conditions != "satisfied" || s.CertificateTime != "valid" || s.Rejection != "none") {
		return false
	}
	if s.State == "complete" {
		return s.Reason == "none" && s.Read == "complete" && s.Structure == "valid" && s.Signature == "verified" && s.Window != "unknown" && s.Usage != "unknown" && s.Attributes != "unknown"
	}
	if s.Reason == "none" || s.Signature == "verified" {
		return false
	}
	if (s.Reason == "denied") != (s.Rejection == "present") {
		return false
	}
	if (s.Reason == "certificate_time_invalid") != (s.CertificateTime == "invalid") {
		return false
	}
	if (s.Read == "absent") != (s.Reason == "absent") {
		return false
	}
	return true
}
func oneVersion(v int) bool { return v == 1 || v == 2 || v == 3 }
func validVersionSource(v int, s Source) bool {
	if v == 3 {
		return validSource(s)
	}
	if s.MemberIdentifierCoverage != "" || s.SignedAttributeCoverage != "" {
		return false
	}
	if v == 2 {
		return validSourceV2(s)
	}
	if s.MemberAlgorithm != "" || s.Extensions != "" || isCMSUnsupported(s.Reason) || one(s.Reason, "parse_failed_ctl_extensions", "parse_failed_ctl_extension", "parse_failed_ctl_extension_duplicate", "parse_failed_ctl_extension_limit") {
		return false
	}
	return validSourceBase(s)
}
func Valid(r Report) bool {
	if !oneVersion(r.SchemaVersion) || r.ToolKind != "ctl-check-dev" || !one(r.Mode, "self-test", "local-ctl-check") || r.Authentication != "blocked" || r.PurposePolicy != "not_established" || r.Revocation != "not_checked" || r.Latest != "not_established" || r.Operations != (Operations{}) {
		return false
	}
	if len(r.Sources) != 2 || r.Sources[0].Source != "authroot_ctl" || r.Sources[1].Source != "disallowed_ctl" || !validVersionSource(r.SchemaVersion, r.Sources[0]) || !validVersionSource(r.SchemaVersion, r.Sources[1]) {
		return false
	}
	complete := r.Sources[0].State == "complete" && r.Sources[1].State == "complete"
	return one(r.State, "complete", "unavailable") && (r.State == "complete") == complete
}
func Encode(r Report) ([]byte, error) {
	if !Valid(r) {
		return nil, ErrInvalid
	}
	b, e := json.Marshal(r)
	return append(b, '\n'), e
}

// Preserve the exact historical field set, including in launcher envelopes.
func (r Report) MarshalJSON() ([]byte, error) {
	type plain Report
	if !Valid(r) {
		return nil, ErrInvalid
	}
	b, e := json.Marshal(plain(r))
	if e != nil || r.SchemaVersion == 3 {
		return b, e
	}
	if r.SchemaVersion == 2 {
		var old v2Report
		if json.Unmarshal(b, &old) != nil {
			return nil, ErrInvalid
		}
		return json.Marshal(old)
	}
	var legacy legacyReport
	if json.Unmarshal(b, &legacy) != nil {
		return nil, ErrInvalid
	}
	return json.Marshal(legacy)
}
func Decode(b []byte) (Report, error) {
	var r Report
	var header struct {
		Version int `json:"schema_version"`
	}
	if len(b) > 16384 || json.Unmarshal(b, &header) != nil {
		return Report{}, ErrInvalid
	}
	var err error
	switch header.Version {
	case 1:
		var legacy legacyReport
		err = casprovenance.DecodeClosedJSON(b, &legacy, 16384)
		if err == nil {
			err = json.Unmarshal(b, &r)
		}
	case 2:
		var old v2Report
		err = casprovenance.DecodeClosedJSON(b, &old, 16384)
		if err == nil {
			err = json.Unmarshal(b, &r)
		}
	case 3:
		err = casprovenance.DecodeClosedJSON(b, &r, 16384)
	default:
		err = ErrInvalid
	}
	if err != nil || !Valid(r) {
		return Report{}, ErrInvalid
	}
	return r, nil
}
