package cascollect

import (
	"encoding/json"
	"errors"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
)

var ErrReport = errors.New("invalid collection report")

// DiscoveryReason identifies the finite detail classes, not arbitrary upstream data.
func DiscoveryReason(reason string) bool {
	switch reason {
	case "discovery_query_key_rejected", "discovery_query_auth_marker":
		return true
	case "discovery_query_rejected", "discovery_query_duplicate", "discovery_query_invalid_value", "discovery_query_extra_fields", "discovery_body_boundary", "discovery_body_value_rejected", "discovery_body_conflict", "discovery_redirect_limit", "discovery_no_match":
		return true
	}
	return false
}

// ValidDiscoveryFailure requires a completed discovery response and no CAS context.
func ValidDiscoveryFailure(reason string, constructed bool, a Attempts) bool {
	return DiscoveryReason(reason) && !constructed && a.Requests >= 1 && a.Requests <= 2 && a.DNS == a.Requests && a.Connects == a.Requests && (reason != "discovery_redirect_limit" || a.Requests == 2)
}

func validReport(r Report) bool {
	if r.SchemaVersion != 1 || r.ToolKind != "cas-collection-dev" || r.Mode != "synthetic-replay" || r.Binding != "not_packaged" || r.Scope != "explicit_page_scripts" || r.Operations != (casprovenance.Operations{}) || r.References == nil || r.Resources == nil {
		return false
	}
	a := r.Simulated
	if a.Requests < 0 || a.Requests > MaxAttempts || a.DNS != a.Requests || a.Connects < 0 || a.Connects > a.Requests || a.Requests-a.Connects > 1 {
		return false
	}
	if r.ContextConstructed && a.Requests < 1 {
		return false
	}
	if r.Collection == "blocked" {
		if len(r.Resources)+len(r.References) != 0 || r.Analysis != nil {
			return false
		}
		if !r.ContextConstructed && a.Requests > 2 {
			return false
		}
		if DiscoveryReason(r.Reason) {
			return ValidDiscoveryFailure(r.Reason, r.ContextConstructed, a)
		}
		switch r.Reason {
		case "invalid_scenario", "canceled", "deadline":
			return true
		case "dns_failed":
			return a.Requests > 0 && a.Connects == a.Requests-1
		case "connect_failed", "request_failed":
			return a.Requests > 0 && a.Connects == a.Requests
		case "discovery_failed":
			return !r.ContextConstructed && a.Requests > 0 && a.Connects == a.Requests
		case "unsupported_source", "analysis_rejected":
			return r.ContextConstructed && a.Requests >= 2 && a.Connects == a.Requests
		case "url_rejected", "response_rejected", "resource_limit":
			return a.Requests > 0 && a.Connects == a.Requests
		case "budget_exceeded":
			return r.ContextConstructed && a.Requests > 0 && a.Connects == a.Requests
		}
		return false
	}
	if r.Collection != "complete" || r.Reason != "none" || !r.ContextConstructed || r.Analysis == nil || casprovenance.ValidateReport(*r.Analysis) != nil || r.Analysis.Analysis != "partial" || a.Requests < 2 || a.Connects != a.Requests || len(r.Resources) > 8 || len(r.References) > 16 {
		return false
	}
	n := len(r.Resources)
	if a.Requests != 2+n && a.Requests != 3+n {
		return false
	}
	sources := r.Analysis.Sources
	if len(sources) < 1+n || sources[0].Kind != "page" {
		return false
	}
	ids := map[string]bool{}
	unknown := 0
	for _, source := range sources[1 : len(sources)-n] {
		switch source.Kind {
		case "inline_script":
		case "unknown_script":
			unknown++
		default:
			return false
		}
	}
	if unknown != len(r.References) {
		return false
	}
	for i, res := range r.Resources {
		source := sources[len(sources)-n+i]
		if res.ID != resourceID(i+1) || res.SourceID != source.ID || source.Kind != "external_script" {
			return false
		}
		ids[res.ID] = false
	}
	if len(r.References) < n || len(r.References) > len(sources)-1-n {
		return false
	}
	nextResource := 1
	for i, ref := range r.References {
		if ref.Ordinal != i+1 {
			return false
		}
		if _, ok := ids[ref.Resource]; !ok {
			return false
		}
		if !ids[ref.Resource] {
			if ref.Resource != resourceID(nextResource) {
				return false
			}
			nextResource++
		}
		ids[ref.Resource] = true
	}
	for _, used := range ids {
		if !used {
			return false
		}
	}
	return true
}
func EncodeReport(r Report) ([]byte, error) {
	if !validReport(r) {
		return nil, ErrReport
	}
	b, err := json.Marshal(r)
	if err != nil || len(b)+1 > MaxReportBytes {
		return nil, ErrReport
	}
	return append(b, '\n'), nil
}
func DecodeReport(b []byte) (Report, error) {
	var r Report
	if casprovenance.DecodeClosedJSON(b, &r, MaxReportBytes) != nil || !validReport(r) {
		return Report{}, ErrReport
	}
	return r, nil
}
