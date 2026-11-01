package authcontext

import (
	"context"
	"encoding/json"
	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
	"net/http"
)

// StaticReport keeps the old classification verbatim inside Context. The new
// Syntax projection never changes a historical unknown into an old known fact.
type StaticReport struct {
	SchemaVersion   int                         `json:"schema_version"`
	ToolKind        string                      `json:"tool_kind"`
	Mode            string                      `json:"mode"`
	State           string                      `json:"state"`
	Reason          string                      `json:"reason"`
	ReleaseEvidence bool                        `json:"release_evidence"`
	Context         Report                      `json:"context"`
	Syntax          *casprovenance.StaticReport `json:"syntax"`
}

func NewStaticReport(mode string) StaticReport {
	return StaticReport{SchemaVersion: 2, ToolKind: "auth-static-context-dev", Mode: mode, State: "blocked", Reason: "unavailable", Context: NewReport(mode)}
}

type StaticAnalyzer func(context.Context, []casprovenance.Input, int) casprovenance.StaticReport

func AnonymousStatic(ctx context.Context, next http.RoundTripper, check func(context.Context) bool, analyze StaticAnalyzer) StaticReport {
	r := NewStaticReport("anonymous")
	if analyze == nil {
		return r
	}
	r.Context = anonymous(ctx, next, check, func(ctx context.Context, inputs []casprovenance.Input, index int) {
		s := analyze(ctx, inputs, index)
		r.Syntax = &s
	})
	r.State, r.Reason = r.Context.State, r.Context.Reason
	if r.Syntax != nil {
		if r.Syntax.Validate() != nil {
			s := casprovenance.EmptyStatic("parser_failure")
			r.Syntax = &s
		}
		if r.Syntax.State != "observed" {
			r.State, r.Reason = "blocked", "static_rejected"
		}
	}
	return r
}
func (r StaticReport) Validate() error {
	if r.SchemaVersion != 2 || r.ToolKind != "auth-static-context-dev" || r.ReleaseEvidence || r.Context.Validate() != nil || r.Context.Mode != r.Mode {
		return ErrReport
	}
	if r.Mode != "anonymous" {
		if r.Syntax != nil || r.State != r.Context.State || r.Reason != r.Context.Reason {
			return ErrReport
		}
		return nil
	}
	if r.Context.State != "observed" {
		if r.Syntax != nil || r.State != r.Context.State || r.Reason != r.Context.Reason {
			return ErrReport
		}
		return nil
	}
	if r.Syntax == nil || r.Syntax.Validate() != nil {
		return ErrReport
	}
	if r.Syntax.State == "rejected" {
		if r.State != "blocked" || r.Reason != "static_rejected" {
			return ErrReport
		}
		return nil
	}
	if r.State != "observed" || r.Reason != "credential_boundary_reached" || len(r.Syntax.Sources) != len(r.Context.Analysis.Sources) {
		return ErrReport
	}
	for i, s := range r.Syntax.Sources {
		old := r.Context.Analysis.Sources[i]
		if s.ID != old.ID || s.Kind != old.Kind || s.Ordinal != old.Ordinal {
			return ErrReport
		}
	}
	if r.Syntax.ParserMatch.SourceID != r.Context.ParserMatch.SourceID {
		return ErrReport
	}
	return nil
}
func EncodeStatic(r StaticReport) ([]byte, error) {
	if r.Validate() != nil {
		return nil, ErrReport
	}
	b, e := json.Marshal(r)
	if len(b) > casprovenance.MaxReportBytes {
		return nil, ErrReport
	}
	return b, e
}
func DecodeStatic(b []byte) (StaticReport, error) {
	var r StaticReport
	if casprovenance.DecodeClosedJSON(b, &r, casprovenance.MaxReportBytes) != nil || r.Validate() != nil {
		return StaticReport{}, ErrReport
	}
	return r, nil
}
