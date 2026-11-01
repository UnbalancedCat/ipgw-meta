package casacquire

import (
	"encoding/json"
	"errors"

	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
	"github.com/UnbalancedCat/ipgw-meta/internal/ctlcheck"
)

type Facts struct {
	Coverage    casprovenance.Coverage     `json:"coverage"`
	Reason      string                     `json:"reason"`
	Sources     []casprovenance.Source     `json:"sources"`
	Occurrences []casprovenance.Occurrence `json:"occurrences"`
	Keys        []casprovenance.Key        `json:"keys"`
	Relations   []casprovenance.Relation   `json:"relations"`
}
type Report struct {
	CTLMembership      *ctlcheck.Membership     `json:"ctl_membership"`
	SchemaVersion      int                      `json:"schema_version"`
	ToolKind           string                   `json:"tool_kind"`
	Mode               string                   `json:"mode"`
	Binding            string                   `json:"binding"`
	Scope              string                   `json:"scope"`
	ContextConstructed bool                     `json:"context_constructed"`
	Collection         string                   `json:"collection"`
	Reason             string                   `json:"reason"`
	Attempts           cascollect.Attempts      `json:"attempts"`
	Operations         casprovenance.Operations `json:"operations"`
	References         []cascollect.Reference   `json:"references"`
	Resources          []cascollect.Resource    `json:"resources"`
	Analysis           *Facts                   `json:"analysis"`
	ChainRelation      *ChainRelation           `json:"chain_relation"`
	ChainTopology      *ChainTopology           `json:"chain_topology"`
	SourceDifference   *SourceDifference        `json:"source_difference"`
}

var ErrReport = errors.New("invalid acquisition report")

func topologyReason(s string) bool {
	switch s {
	case "topology_query_failed", "topology_timeout", "topology_output_rejected", "wired_interface_ambiguous", "wireless_active", "default_route_rejected", "split_route_rejected", "source_address_rejected", "dns_configuration_rejected":
		return true
	}
	return false
}

func blocked(mode, reason string) Report {
	return Report{SchemaVersion: 1, ToolKind: "cas-acquire-dev", Mode: mode, Binding: "not_packaged", Scope: "explicit_page_scripts", Collection: "blocked", Reason: reason, References: []cascollect.Reference{}, Resources: []cascollect.Resource{}}
}
func project(mode string, r cascollect.Report, counts cascollect.Attempts) Report {
	o := blocked(mode, r.Reason)
	o.ContextConstructed, o.Collection, o.Attempts = r.ContextConstructed, r.Collection, counts
	if mode == "anonymous-once" {
		o.Operations.ExternalRequests, o.Operations.DNS = counts.Requests, counts.DNS
	}
	if r.Collection == "complete" && r.Analysis != nil {
		a := r.Analysis
		o.Analysis = &Facts{Coverage: a.Coverage, Reason: a.Reason, Sources: a.Sources, Occurrences: a.Occurrences, Keys: a.Keys, Relations: a.Relations}
		o.References, o.Resources = r.References, r.Resources
	}
	return o
}
func valid(r Report) bool {
	if r.SchemaVersion != 1 || r.ToolKind != "cas-acquire-dev" || r.Binding != "not_packaged" || r.Scope != "explicit_page_scripts" || r.References == nil || r.Resources == nil {
		return false
	}
	a := r.Attempts
	if a.Requests < 0 || a.Requests > 12 || a.DNS < 0 || a.DNS > a.Requests || a.Connects < 0 || a.Connects > a.DNS {
		return false
	}
	ops := casprovenance.Operations{}
	switch r.Mode {
	case "self-test", "local-preflight", "local-cache-check":
		if a != (cascollect.Attempts{}) {
			return false
		}
	case "anonymous-once":
		ops.ExternalRequests, ops.DNS = a.Requests, a.DNS
	default:
		return false
	}
	if r.Operations != ops {
		return false
	}
	if r.CTLMembership != nil {
		if !validMemberExtension(r) {
			return false
		}
		if r.CTLMembership.Phase == "gate" {
			return true
		}
	}
	if r.SourceDifference != nil {
		if !validSourceDifference(r) {
			return false
		}
		if r.SourceDifference.Phase == "gate" {
			return true
		}
	}
	if topologyObservationReason(r.Reason) {
		return r.Mode == "anonymous-once" && r.Collection == "blocked" && a.Requests > 0 && a.Requests == a.DNS && a.DNS == a.Connects && r.Analysis == nil && len(r.Resources)+len(r.References) == 0 && validTopology(r.Reason, r.ChainRelation, r.ChainTopology)
	}
	if r.ChainTopology != nil {
		return false
	}
	if relationReason(r.Reason) {
		return r.Mode == "anonymous-once" && r.Collection == "blocked" && a.Requests > 0 && a.Requests == a.DNS && a.DNS == a.Connects && r.Analysis == nil && len(r.Resources)+len(r.References) == 0 && validRelation(r.Reason, r.ChainRelation)
	}
	if r.ChainRelation != nil {
		return false
	}
	if r.Mode == "local-cache-check" {
		if r.ContextConstructed || r.Analysis != nil || len(r.Resources)+len(r.References) != 0 {
			return false
		}
		if r.Collection == "ready" {
			return r.Reason == "none"
		}
		return r.Collection == "blocked" && (r.Reason == "cache_partial" || r.Reason == "cache_partial_empty" || loaderFailureReason(r.Reason) || r.Reason == "tls_chain_probe_cache_subject_limit")
	}
	if r.Mode == "local-preflight" {
		if r.ContextConstructed || r.Analysis != nil || len(r.Resources)+len(r.References) != 0 {
			return false
		}
		if r.Collection == "ready" {
			return r.Reason == "none"
		}
		if r.Collection != "blocked" {
			return false
		}
	}
	if r.Collection == "blocked" {
		if r.Analysis != nil || len(r.References)+len(r.Resources) != 0 {
			return false
		}
		if topologyReason(r.Reason) {
			return a == (cascollect.Attempts{}) && !r.ContextConstructed
		}
		if cascollect.DiscoveryReason(r.Reason) {
			return r.Mode == "anonymous-once" && cascollect.ValidDiscoveryFailure(r.Reason, r.ContextConstructed, a)
		}
		if transportReason(r.Reason) {
			return r.Mode == "anonymous-once" && a.Requests > 0 && a.Requests == a.DNS && a.DNS == a.Connects
		}
		if r.Reason == "source_baseline_tls_verified" {
			return r.SourceDifference != nil && r.SourceDifference.Status == "not_run" && r.Mode == "anonymous-once" && a.Requests > 0 && a.Requests == a.DNS && a.DNS == a.Connects
		}
		switch r.Reason {
		case "invalid_scenario", "url_rejected", "discovery_failed", "response_rejected", "unsupported_source", "budget_exceeded", "resource_limit", "dns_failed", "connect_failed", "request_failed", "analysis_rejected", "canceled", "deadline", "topology_rejected", "transport_failed", "self_test_failed", "route_rejected", "interface_binding_failed":
			return r.Reason != "topology_rejected" || a == (cascollect.Attempts{}) && !r.ContextConstructed
		}
		return false
	}
	if r.Collection != "complete" || r.Reason != "none" || !r.ContextConstructed || r.Analysis == nil {
		return false
	}
	f := r.Analysis
	p := casprovenance.Report{SchemaVersion: 1, ToolKind: "cas-provenance-dev", Mode: "synthetic-prototype", Binding: "not_packaged", Analysis: "partial", Coverage: f.Coverage, Reason: f.Reason, Sources: f.Sources, Occurrences: f.Occurrences, Keys: f.Keys, Relations: f.Relations}
	// Reuse fact/reference validation only; never emit this internal envelope.
	n := 2 + len(r.Resources)
	if r.Mode == "anonymous-once" {
		if a.DNS != a.Requests || a.Connects != a.Requests {
			return false
		}
		n = a.Requests
	}
	c := cascollect.Report{SchemaVersion: 1, ToolKind: "cas-collection-dev", Mode: "synthetic-replay", Binding: "not_packaged", Scope: r.Scope, ContextConstructed: true, Collection: "complete", Reason: "none", Simulated: cascollect.Attempts{Requests: n, DNS: n, Connects: n}, References: r.References, Resources: r.Resources, Analysis: &p}
	_, err := cascollect.EncodeReport(c)
	return err == nil
}
func Encode(r Report) ([]byte, error) {
	if !valid(r) {
		return nil, ErrReport
	}
	b, err := json.Marshal(r)
	if err != nil || len(b)+1 > cascollect.MaxReportBytes {
		return nil, ErrReport
	}
	return append(b, '\n'), nil
}
func decodeHistorical(b []byte) (Report, error) {
	var r Report
	var d7 d7Report
	if casprovenance.DecodeClosedJSON(b, &d7, cascollect.MaxReportBytes) != nil {
		var d6 d6Report
		if casprovenance.DecodeClosedJSON(b, &d6, cascollect.MaxReportBytes) == nil {
			if d6.ChainRelation == nil {
				return Report{}, ErrReport
			}
			r = d6.current()
		} else {
			var old legacyReport
			if casprovenance.DecodeClosedJSON(b, &old, cascollect.MaxReportBytes) != nil {
				return Report{}, ErrReport
			}
			r = old.current()
		}
	} else if d7.ChainRelation == nil || d7.ChainTopology == nil {
		// An explicit null extension is not an old record or a D6 observation.
		return Report{}, ErrReport
	} else {
		r = d7.current()
	}
	if !valid(r) {
		return Report{}, ErrReport
	}
	return r, nil
}

func Decode(b []byte) (Report, error) {
	var r Report
	if casprovenance.DecodeClosedJSON(b, &r, cascollect.MaxReportBytes) == nil {
		if r.CTLMembership == nil || !valid(r) {
			return Report{}, ErrReport
		}
		return r, nil
	}
	var d9 d9Report
	if casprovenance.DecodeClosedJSON(b, &d9, cascollect.MaxReportBytes) == nil {
		r = d9.current()
		if r.SourceDifference == nil || !valid(r) {
			return Report{}, ErrReport
		}
		return r, nil
	}
	return decodeHistorical(b)
}
