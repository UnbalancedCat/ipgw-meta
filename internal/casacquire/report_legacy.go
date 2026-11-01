package casacquire

import (
	"encoding/json"

	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
)

// The launcher embeds Report directly in its receipt, so apply the same
// historical shape there as in Encode, not only at the payload boundary.
func (r Report) MarshalJSON() ([]byte, error) {
	if r.CTLMembership != nil {
		type extended Report
		return json.Marshal(extended(r))
	}
	if r.SourceDifference != nil {
		return json.Marshal(r.d9())
	}
	if r.ChainRelation == nil {
		return json.Marshal(r.legacy())
	}
	if r.ChainTopology == nil {
		return json.Marshal(r.d6())
	}
	return json.Marshal(r.d7())
}

// Preserve the exact historical closed shape without weakening the shared
// decoder's required-field, duplicate-field, or unknown-field checks.
type legacyReport struct {
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
}

func (r Report) legacy() legacyReport {
	return legacyReport{r.SchemaVersion, r.ToolKind, r.Mode, r.Binding, r.Scope, r.ContextConstructed, r.Collection, r.Reason, r.Attempts, r.Operations, r.References, r.Resources, r.Analysis}
}

func (r legacyReport) current() Report {
	return Report{nil, r.SchemaVersion, r.ToolKind, r.Mode, r.Binding, r.Scope, r.ContextConstructed, r.Collection, r.Reason, r.Attempts, r.Operations, r.References, r.Resources, r.Analysis, nil, nil, nil}
}

type d6Report struct {
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
}

func (r Report) d6() d6Report {
	return d6Report{r.SchemaVersion, r.ToolKind, r.Mode, r.Binding, r.Scope, r.ContextConstructed, r.Collection, r.Reason, r.Attempts, r.Operations, r.References, r.Resources, r.Analysis, r.ChainRelation}
}

func (r d6Report) current() Report {
	return Report{nil, r.SchemaVersion, r.ToolKind, r.Mode, r.Binding, r.Scope, r.ContextConstructed, r.Collection, r.Reason, r.Attempts, r.Operations, r.References, r.Resources, r.Analysis, r.ChainRelation, nil, nil}
}

type d7Report struct {
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
}

func (r Report) d7() d7Report {
	return d7Report{r.SchemaVersion, r.ToolKind, r.Mode, r.Binding, r.Scope, r.ContextConstructed, r.Collection, r.Reason, r.Attempts, r.Operations, r.References, r.Resources, r.Analysis, r.ChainRelation, r.ChainTopology}
}
func (r d7Report) current() Report {
	return Report{nil, r.SchemaVersion, r.ToolKind, r.Mode, r.Binding, r.Scope, r.ContextConstructed, r.Collection, r.Reason, r.Attempts, r.Operations, r.References, r.Resources, r.Analysis, r.ChainRelation, r.ChainTopology, nil}
}

type d9Report struct {
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

func (r Report) d9() d9Report {
	return d9Report{r.SchemaVersion, r.ToolKind, r.Mode, r.Binding, r.Scope, r.ContextConstructed, r.Collection, r.Reason, r.Attempts, r.Operations, r.References, r.Resources, r.Analysis, r.ChainRelation, r.ChainTopology, r.SourceDifference}
}
func (r d9Report) current() Report {
	return Report{nil, r.SchemaVersion, r.ToolKind, r.Mode, r.Binding, r.Scope, r.ContextConstructed, r.Collection, r.Reason, r.Attempts, r.Operations, r.References, r.Resources, r.Analysis, r.ChainRelation, r.ChainTopology, r.SourceDifference}
}
