// Package authcontext observes only the SDK's consumed anonymous prelude.
// Product entry points must not import this development-only package.
package authcontext

import (
	"encoding/json"
	"errors"
	"github.com/UnbalancedCat/ipgw-meta/internal/authdev"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
	"slices"
)

var ErrReport = errors.New("authcontext: invalid report")

type Form struct {
	Selection       string `json:"selection"`
	LT              string `json:"lt_shape"`
	Execution       string `json:"execution_shape"`
	SameOriginHTTPS string `json:"same_origin_https"`
	InitialTarget   string `json:"initial_target"`
	FinalTarget     string `json:"final_target"`
	Service         string `json:"service"`
}

func unknownForm() Form {
	return Form{"unknown", "unknown", "unknown", "unknown", "unknown", "unknown", "unknown"}
}

type Report struct {
	SchemaVersion       int                       `json:"schema_version"`
	ToolKind            string                    `json:"tool_kind"`
	Mode                string                    `json:"mode"`
	State               string                    `json:"state"`
	Reason              string                    `json:"reason"`
	Scope               string                    `json:"scope"`
	ReleaseEvidence     bool                      `json:"release_evidence"`
	SDK                 authdev.Report            `json:"sdk"`
	Form                Form                      `json:"form"`
	SuccessfulSources   int                       `json:"successful_sources"`
	NonSuccessResponses int                       `json:"non_success_responses"`
	Analysis            *Analysis                 `json:"analysis"`
	ParserMatch         casprovenance.ParserMatch `json:"parser_match"`
}

// Only analysis facts are projected; the legacy prototype's synthetic mode and
// zero-operation envelope must not be misrepresented as live acquisition facts.
type Analysis struct {
	State       string                     `json:"state"`
	Reason      string                     `json:"reason"`
	Coverage    casprovenance.Coverage     `json:"coverage"`
	Sources     []casprovenance.Source     `json:"sources"`
	Occurrences []casprovenance.Occurrence `json:"occurrences"`
	Keys        []casprovenance.Key        `json:"keys"`
	Relations   []casprovenance.Relation   `json:"relations"`
}

func project(a casprovenance.Report) *Analysis {
	return &Analysis{a.Analysis, a.Reason, a.Coverage, a.Sources, a.Occurrences, a.Keys, a.Relations}
}
func (a *Analysis) prototype() *casprovenance.Report {
	if a == nil {
		return nil
	}
	return &casprovenance.Report{SchemaVersion: 1, ToolKind: "cas-provenance-dev", Mode: "synthetic-prototype", Binding: "not_packaged", Analysis: a.State, Reason: a.Reason, Coverage: a.Coverage, Sources: a.Sources, Occurrences: a.Occurrences, Keys: a.Keys, Relations: a.Relations}
}

func NewReport(mode string) Report {
	return Report{SchemaVersion: 1, ToolKind: "auth-context-dev", Mode: mode, State: "blocked", Reason: "unavailable", Scope: "sdk_consumed_only", SDK: authdev.NewReport(mode), Form: unknownForm(), ParserMatch: casprovenance.UnknownParserMatch()}
}
func one(v string, list ...string) bool { return slices.Contains(list, v) }
func (r Report) Validate() error {
	if r.SchemaVersion != 1 || r.ToolKind != "auth-context-dev" || r.Scope != "sdk_consumed_only" || r.ReleaseEvidence || !one(r.Mode, "anonymous", "preflight", "self-test") || r.SDK.Mode != r.Mode || r.SDK.Validate() != nil || r.SDK.Commands != 0 || r.ParserMatch.Validate(r.Analysis.prototype()) != nil {
		return ErrReport
	}
	if !one(r.State, "blocked", "observed", "ready", "pass") || !one(r.Reason, "unavailable", "none", "credential_boundary_reached", "sdk_stopped", "resource_limit", "incomplete_body", "role_uncertain", "analysis_rejected", "parser_unmapped", "canceled") {
		return ErrReport
	}
	if r.SuccessfulSources < 0 || r.SuccessfulSources > casprovenance.MaxSources || r.NonSuccessResponses < 0 || r.NonSuccessResponses > authdev.MaxRequests {
		return ErrReport
	}
	f := r.Form
	if !one(f.Selection, "unknown", "missing", "unique", "ambiguous") || !one(f.LT, "unknown", "missing", "empty", "unique", "ambiguous") || !one(f.Execution, "unknown", "missing", "empty", "unique", "ambiguous") || !one(f.SameOriginHTTPS, "unknown", "same", "different") || !one(f.InitialTarget, "unknown", "same", "different") || !one(f.FinalTarget, "unknown", "same", "different") || !one(f.Service, "unknown", "same", "absent", "different") {
		return ErrReport
	}
	if r.Mode != "anonymous" {
		if r.Analysis != nil || f != unknownForm() || r.ParserMatch.State != "unknown" || r.SuccessfulSources != 0 || r.NonSuccessResponses != 0 {
			return ErrReport
		}
		if r.Mode == "self-test" && r.State == "pass" && r.Reason == "none" && r.SDK.State == "pass" || r.Mode == "preflight" && r.State == "ready" && r.Reason == "none" && r.SDK.State == "ready" || r.State == "blocked" && r.Reason == "unavailable" && r.SDK.State == "blocked" {
			return nil
		}
		return ErrReport
	}
	if r.State == "observed" {
		if r.Reason != "credential_boundary_reached" || r.SDK.State != "observed" || r.Analysis == nil || casprovenance.ValidateReport(*r.Analysis.prototype()) != nil || r.Analysis.State != "partial" || r.ParserMatch.State != "mapped" || r.SuccessfulSources < 1 || f.Selection == "missing" || f.Selection == "unknown" || f.LT != "unique" || f.Execution != "unique" || f.SameOriginHTTPS != "same" || f.InitialTarget == "unknown" || f.FinalTarget == "unknown" {
			return ErrReport
		}
		if r.SuccessfulSources+r.NonSuccessResponses > r.SDK.CASRequests || len(r.Analysis.Sources) < r.SuccessfulSources || r.Analysis.Sources[0].Kind != "page" {
			return ErrReport
		}
		responses := 1
		lastExternal := ""
		for _, s := range r.Analysis.Sources {
			if s.Kind == "external_script" {
				responses++
				lastExternal = s.ID
			}
		}
		if responses != r.SuccessfulSources || lastExternal != "" && r.ParserMatch.SourceID != lastExternal {
			return ErrReport
		}
		return nil
	}
	if r.State != "blocked" || one(r.Reason, "none", "credential_boundary_reached") || r.Analysis != nil || f != unknownForm() || r.ParserMatch.State != "unknown" || r.SuccessfulSources != 0 || r.NonSuccessResponses > r.SDK.CASRequests {
		return ErrReport
	}
	if r.Reason == "sdk_stopped" && r.SDK.State == "observed" || r.Reason == "unavailable" && r.SDK.Requests != 0 {
		return ErrReport
	}
	return nil
}
func Encode(r Report) ([]byte, error) {
	if r.Validate() != nil {
		return nil, ErrReport
	}
	b, e := json.Marshal(r)
	if len(b) > casprovenance.MaxReportBytes {
		return nil, ErrReport
	}
	return b, e
}
func Decode(data []byte) (Report, error) {
	var r Report
	if casprovenance.DecodeClosedJSON(data, &r, casprovenance.MaxReportBytes) != nil || r.Validate() != nil {
		return Report{}, ErrReport
	}
	return r, nil
}
