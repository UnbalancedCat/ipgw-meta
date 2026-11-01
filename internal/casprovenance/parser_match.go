package casprovenance

import (
	"context"
	"github.com/UnbalancedCat/ipgw-meta/internal/caslegacy"
)

// ParserMatch describes the frozen pre-ADR-0021 text parser, never JavaScript execution or
// a correct/server-accepted key. It is separate from the historical Report.
type ParserMatch struct {
	State      string `json:"state"`
	SourceID   string `json:"source_id"`
	Occurrence int    `json:"occurrence"`
}
type matchTracker struct {
	input, position, count int
	result                 ParserMatch
}

func UnknownParserMatch() ParserMatch { return ParserMatch{State: "unknown"} }

func (m ParserMatch) Validate(r *Report) error {
	if m.State == "unknown" && m.SourceID == "" && m.Occurrence == 0 {
		return nil
	}
	if m.State != "mapped" || r == nil || ValidateReport(*r) != nil || r.Analysis == "rejected" || m.Occurrence < 1 || m.Occurrence > len(r.Occurrences) {
		return ErrReport
	}
	if r.Occurrences[m.Occurrence-1].SourceID != m.SourceID {
		return ErrReport
	}
	return nil
}

// inputIndex is the zero-based original response whose parser first matched.
// The caller must establish its successful-response/SDK-phase eligibility.
func AnalyzeParserMatch(ctx context.Context, inputs []Input, inputIndex int) (Report, ParserMatch, error) {
	m := matchTracker{input: inputIndex, position: -1}
	// Bound before regex or parser work; Analyze performs cumulative checks too.
	if ctx != nil && ctx.Err() == nil && inputIndex >= 0 && inputIndex < len(inputs) && len(inputs[inputIndex].Data) <= MaxSourceBytes {
		data := inputs[inputIndex].Data
		p := rawKey.FindSubmatchIndex(data)
		key := caslegacy.ExtractPublicKey(data)
		if key != "" && len(p) == 4 && string(data[p[2]:p[3]]) == key {
			m.position = p[0]
		}
	}
	r, err := analyzeCore(ctx, inputs, &m)
	if err != nil || m.count != 1 || m.result.Validate(&r) != nil {
		return r, UnknownParserMatch(), err
	}
	return r, m.result, nil
}
