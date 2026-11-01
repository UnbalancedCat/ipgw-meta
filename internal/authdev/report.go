package authdev

import (
	"encoding/json"
	"errors"
	"slices"

	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
	"github.com/UnbalancedCat/ipgw-meta/internal/livegate"
)

var ErrReport = errors.New("authdev: invalid report")

type Report struct {
	SchemaVersion     int             `json:"schema_version"`
	ToolKind          string          `json:"tool_kind"`
	Mode              string          `json:"mode"`
	ReleaseEvidence   bool            `json:"release_evidence"`
	State             string          `json:"state"`
	Reason            string          `json:"reason"`
	SDKCode           string          `json:"sdk_code"`
	Requests          int             `json:"application_requests"`
	StatusRequests    int             `json:"status_requests"`
	DiscoveryRequests int             `json:"discovery_requests"`
	CASRequests       int             `json:"cas_requests"`
	CookieReceived    bool            `json:"cas_cookie_received"`
	CookieSent        bool            `json:"cas_cookie_sent"`
	PageObserved      bool            `json:"page_observed"`
	KeySource         string          `json:"key_source"`
	ActionService     string          `json:"action_service"`
	Commands          int             `json:"product_commands"`
	StepCoverage      string          `json:"step_coverage"`
	Steps             []livegate.Step `json:"steps"`
}

func NewReport(mode string) Report {
	return Report{SchemaVersion: 1, ToolKind: "auth-dev", Mode: mode, State: "blocked", Reason: "unavailable", KeySource: "none", ActionService: "unknown", StepCoverage: "not_started", Steps: []livegate.Step{}}
}
func (r Report) Validate() error {
	if r.SchemaVersion != 1 || r.ToolKind != "auth-dev" || r.ReleaseEvidence || !slices.Contains([]string{"anonymous", "password", "preflight", "self-test"}, r.Mode) ||
		!slices.Contains([]string{"blocked", "observed", "pass", "fail", "ready"}, r.State) ||
		!slices.Contains([]string{"unavailable", "none", "credential_boundary_reached", "canceled", "sdk_rejected", "initial_online", "early_ticket", "request_rejected", "request_budget", "topology_changed", "profile_rejected", "state_machine_failed", "binding_rejected", "confirmation_required", "private_tty_required"}, r.Reason) ||
		!slices.Contains([]string{"", "invalid_argument", "config", "network", "authentication", "session_conflict", "protocol_changed", "interaction_required", "unsupported", "internal"}, r.SDKCode) ||
		!slices.Contains([]string{"none", "page", "script"}, r.KeySource) || !slices.Contains([]string{"unknown", "absent", "same", "different"}, r.ActionService) ||
		r.Requests < 0 || r.Requests > MaxRequests || r.StatusRequests < 0 || r.StatusRequests > 1 || r.DiscoveryRequests < 0 || r.DiscoveryRequests > 2 || r.CASRequests < 0 || r.Requests != r.StatusRequests+r.DiscoveryRequests+r.CASRequests ||
		r.Commands < 0 || r.Commands > 10 || r.Steps == nil || len(r.Steps) > 9 {
		return ErrReport
	}
	if r.Mode != "anonymous" && (r.Requests != 0 || r.CookieSent || r.CookieReceived || r.PageObserved || r.KeySource != "none" || r.ActionService != "unknown" || r.SDKCode != "") {
		return ErrReport
	}
	if r.Mode != "password" && (r.Commands != 0 || len(r.Steps) != 0 || r.StepCoverage != "not_started") {
		return ErrReport
	}
	if r.State == "observed" && (r.Mode != "anonymous" || r.Reason != "credential_boundary_reached" || !r.PageObserved || r.KeySource == "none") || r.Reason == "credential_boundary_reached" && r.State != "observed" {
		return ErrReport
	}
	if r.State == "ready" && (r.Mode != "preflight" || r.Reason != "none") {
		return ErrReport
	}
	if r.State == "pass" && (r.Mode != "self-test" && r.Mode != "password" || r.Reason != "none") {
		return ErrReport
	}
	if r.State == "fail" && r.Mode != "password" {
		return ErrReport
	}
	if r.Reason == "none" && r.State != "pass" && r.State != "ready" {
		return ErrReport
	}
	if r.SDKCode != "" && r.Reason != "sdk_rejected" {
		return ErrReport
	}
	if r.PageObserved != (r.ActionService != "unknown") || r.KeySource != "none" && !r.PageObserved || (r.CookieReceived || r.CookieSent || r.PageObserved) && r.CASRequests == 0 {
		return ErrReport
	}
	for _, s := range r.Steps {
		if s.Validate() != nil {
			return ErrReport
		}
	}
	if r.Mode == "password" {
		switch r.StepCoverage {
		case "not_started":
			if len(r.Steps) != 0 || r.State != "blocked" || r.Commands > 1 {
				return ErrReport
			}
		case "complete":
			if livegate.ValidateDevelopmentPasswordSteps(r.Steps, livegate.Result(r.State)) != nil || r.Commands < len(r.Steps) || r.Commands > len(r.Steps)+1 {
				return ErrReport
			}
		case "partial":
			if r.State == "pass" || r.Commands < len(r.Steps) || r.Commands > len(r.Steps)+4 || !slices.Contains([]string{"state_machine_failed", "topology_changed", "canceled"}, r.Reason) || livegate.ValidateDevelopmentPasswordPrefix(r.Steps) != nil {
				return ErrReport
			}
		default:
			return ErrReport
		}
	}
	return nil
}
func Encode(r Report) ([]byte, error) {
	if r.Validate() != nil {
		return nil, ErrReport
	}
	return json.Marshal(r)
}
func Decode(data []byte) (Report, error) {
	var r Report
	if casprovenance.DecodeClosedJSON(data, &r, 64<<10) != nil || r.Validate() != nil {
		return Report{}, ErrReport
	}
	return r, nil
}
