package qrdev

import (
	"encoding/json"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
	"regexp"
)

type Receipt struct {
	Schema          int              `json:"schema_version"`
	Kind            string           `json:"tool_kind"`
	Source          string           `json:"source_digest"`
	Product         string           `json:"product_sha256"`
	Tool            string           `json:"tool_sha256"`
	ReleaseEvidence bool             `json:"release_evidence"`
	Binding         string           `json:"binding"`
	State           string           `json:"state"`
	Reason          string           `json:"reason"`
	LiveAttempted   bool             `json:"live_attempted"`
	Preflight       *PreflightReport `json:"preflight"`
	QR              *Report          `json:"qr"`
	GateFailure     *GateFailure     `json:"gate_failure"`
}
type legacyReceipt struct {
	Schema          int              `json:"schema_version"`
	Kind            string           `json:"tool_kind"`
	Source          string           `json:"source_digest"`
	Product         string           `json:"product_sha256"`
	Tool            string           `json:"tool_sha256"`
	ReleaseEvidence bool             `json:"release_evidence"`
	Binding         string           `json:"binding"`
	State           string           `json:"state"`
	Reason          string           `json:"reason"`
	LiveAttempted   bool             `json:"live_attempted"`
	Preflight       *PreflightReport `json:"preflight"`
	QR              *Report          `json:"qr"`
}

var receiptDigest = regexp.MustCompile(`^[a-f0-9]{64}$`)

func (r Receipt) Validate() error {
	if !oneOf(r.Kind, "qr-dev") || r.Schema != 1 && r.Schema != 2 || r.ReleaseEvidence || !receiptDigest.MatchString(r.Source) || !receiptDigest.MatchString(r.Product) || !receiptDigest.MatchString(r.Tool) || !oneOf(r.Binding, "verified", "invalid") || !oneOf(r.State, "pass", "ready", "canceled", "blocked", "fail") {
		return ErrReport
	}
	if !oneOf(r.Reason, "none", "arguments_rejected", "binding_rejected", "self_test_failed", "private_terminal_required", "preflight_failed", "topology_unavailable", "environment_rejected", "profile_rejected", "topology_changed", "report_rejected", "canceled_or_timeout", "state_machine_failed", "topology_or_binding_changed", "command_invalid", "command_failed") {
		return ErrReport
	}
	if r.Preflight != nil && !r.Preflight.Valid() || r.QR != nil && (r.QR.Validate() != nil || !r.LiveAttempted) {
		return ErrReport
	}
	if r.Binding == "invalid" && (r.State != "blocked" || r.QR != nil) || (r.State == "pass" || r.State == "ready" || r.State == "canceled") && (r.Binding != "verified" || r.Reason != "none") {
		return ErrReport
	}
	if r.LiveAttempted && r.State == "pass" && (r.QR == nil || r.QR.State != "pass") {
		return ErrReport
	}
	if r.State == "fail" && r.QR == nil || r.State == "canceled" && (r.LiveAttempted || r.QR != nil) {
		return ErrReport
	}
	if r.QR != nil {
		override := r.State == "blocked" && oneOf(r.Reason, "topology_changed", "canceled_or_timeout")
		if !override && (r.State != r.QR.State || r.Reason != r.QR.Reason) {
			return ErrReport
		}
	}
	if r.State == "ready" && (r.Preflight == nil || r.Preflight.State != "ready" || r.LiveAttempted) {
		return ErrReport
	}
	if r.Schema == 2 && r.GateFailure == nil && (r.Reason == "topology_changed" || r.Reason == "topology_or_binding_changed" || r.QR != nil && r.QR.Reason == "topology_or_binding_changed") {
		return ErrReport
	}
	if f := r.GateFailure; f != nil {
		if r.Schema != 2 || !f.Valid() || r.State == "pass" || r.State == "ready" || r.State == "canceled" {
			return ErrReport
		}
		if f.Command == "initial_check" || f.Command == "profile_show" {
			if r.LiveAttempted {
				return ErrReport
			}
		} else if f.Command == "final_check" {
			if !r.LiveAttempted {
				return ErrReport
			}
		} else {
			if !r.LiveAttempted {
				return ErrReport
			}
			// Final binding invalidation deliberately discards the authentication trace.
			if r.Binding == "verified" {
				if r.QR == nil || len(r.QR.Commands) == 0 {
					return ErrReport
				}
				last := r.QR.Commands[len(r.QR.Commands)-1]
				if string(last.Action) != f.Command || last.State != "blocked" {
					return ErrReport
				}
			}
		}
	}
	return nil
}

func keysExactly(raw []byte, keys ...string) bool {
	var m map[string]json.RawMessage
	if json.Unmarshal(raw, &m) != nil || len(m) != len(keys) {
		return false
	}
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			return false
		}
	}
	return true
}

// DecodeReceipt reads old receipts without inventing diagnostics. Only the
// new writer emits v2; v1 may not carry even a null gate_failure field.
func DecodeReceipt(raw []byte) (Receipt, error) {
	var r Receipt
	if len(raw) > 64<<10 {
		return r, ErrReport
	}
	var version struct {
		Schema int `json:"schema_version"`
	}
	if json.Unmarshal(raw, &version) != nil {
		return r, ErrReport
	}
	if version.Schema == 1 {
		var old legacyReceipt
		if casprovenance.DecodeClosedJSON(raw, &old, 64<<10) != nil {
			return r, ErrReport
		}
		if json.Unmarshal(raw, &r) != nil {
			return r, ErrReport
		}
	} else if version.Schema == 2 {
		if casprovenance.DecodeClosedJSON(raw, &r, 64<<10) != nil {
			return r, ErrReport
		}
	} else {
		return r, ErrReport
	}

	keys := []string{"schema_version", "tool_kind", "source_digest", "product_sha256", "tool_sha256", "release_evidence", "binding", "state", "reason", "live_attempted", "preflight", "qr"}
	if r.Schema == 2 {
		keys = append(keys, "gate_failure")
	}
	if !keysExactly(raw, keys...) {
		return r, ErrReport
	}
	if r.GateFailure != nil {
		var m map[string]json.RawMessage
		_ = json.Unmarshal(raw, &m)
		if !keysExactly(m["gate_failure"], "command", "phase", "reason", "query_reason", "changes") {
			return r, ErrReport
		}
		var f map[string]json.RawMessage
		_ = json.Unmarshal(m["gate_failure"], &f)
		if !keysExactly(f["changes"], "address", "interface", "dns", "routes", "adapters", "ssid") {
			return r, ErrReport
		}
	}
	return r, r.Validate()
}
