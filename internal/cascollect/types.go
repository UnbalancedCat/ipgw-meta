// Package cascollect plans bounded acquisition over in-memory response prefixes.
// It has no network, filesystem, credential, SDK or callback transport capability.
package cascollect

import "github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"

const (
	MaxAttempts       = 12
	MaxHeaderBytes    = 16 << 10
	MaxDiscoveryBytes = 256 << 10
	MaxBodyBytes      = 1 << 20
	MaxTotalBytes     = 4 << 20
	MaxReportBytes    = 65536
)

type Scope struct{ Captive, Service, CAS string }
type Header struct{ Name, Value string }

// Exchange is a scripted response, not an executable transport callback.
// Fault may be empty, dns, connect, or request; no raw error is accepted.
type Exchange struct {
	Target  string
	Status  int
	Headers []Header
	Body    []byte
	Fault   string
}
type Scenario struct {
	Scope     Scope
	Exchanges []Exchange
}

// Request has deliberately no headers, body or authentication fields.
type Request struct{ Method, Target string }
type Attempts struct {
	Requests int `json:"requests"`
	DNS      int `json:"dns"`
	Connects int `json:"connects"`
}
type Reference struct {
	Ordinal  int    `json:"ordinal"`
	Resource string `json:"resource"`
}
type Resource struct {
	ID       string `json:"id"`
	SourceID string `json:"source_id"`
}
type Report struct {
	SchemaVersion      int                      `json:"schema_version"`
	ToolKind           string                   `json:"tool_kind"`
	Mode               string                   `json:"mode"`
	Binding            string                   `json:"binding"`
	Scope              string                   `json:"scope"`
	ContextConstructed bool                     `json:"context_constructed"`
	Collection         string                   `json:"collection"`
	Reason             string                   `json:"reason"`
	Simulated          Attempts                 `json:"simulated_attempts"`
	Operations         casprovenance.Operations `json:"operations"`
	References         []Reference              `json:"references"`
	Resources          []Resource               `json:"resources"`
	Analysis           *casprovenance.Report    `json:"analysis"`
}

func emptyReport() Report {
	return Report{SchemaVersion: 1, ToolKind: "cas-collection-dev", Mode: "synthetic-replay", Binding: "not_packaged", Scope: "explicit_page_scripts", Collection: "blocked", Reason: "invalid_scenario", References: []Reference{}, Resources: []Resource{}}
}
