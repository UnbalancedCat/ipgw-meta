package casprovenance

import "fmt"

func (r StaticReport) Validate() error {
	if r.UseRelation != "unknown" || r.Sources == nil || r.Occurrences == nil || r.Keys == nil {
		return ErrReport
	}
	if r.State == "rejected" {
		if !oneOf(r.Reason, "invalid_input", "syntax_rejected", "token_limit", "depth_limit", "work_limit", "input_limit", "canceled", "parser_failure", "mapping_failed", "protection_unavailable") || len(r.Sources)+len(r.Occurrences)+len(r.Keys) != 0 || r.ParserMatch != UnknownParserMatch() {
			return ErrReport
		}
		return nil
	}
	if r.State != "observed" || r.Reason != "none" || len(r.Sources) < 1 || len(r.Sources) > MaxSources || len(r.Occurrences) < 1 || len(r.Occurrences) > MaxOccurrences || len(r.Keys) > MaxOccurrences {
		return ErrReport
	}
	sources := map[string]int{}
	pages := 0
	for i, s := range r.Sources {
		if s.ID != fmt.Sprintf("S%d", i+1) || s.Ordinal != i+1 || !oneOf(s.Kind, "page", "inline_script", "external_script", "unknown_script") {
			return ErrReport
		}
		if s.Kind == "page" {
			pages++
		}
		sources[s.ID] = i + 1
	}
	if pages > 1 || len(r.Sources)-pages > MaxSources-1 {
		return ErrReport
	}
	keys := map[string]bool{}
	for i, k := range r.Keys {
		if k.ID != fmt.Sprintf("K%d", i+1) || k.Bits < 512 || k.Bits > 8192 || !oneOf(k.ExponentClass, "common", "other") {
			return ErrReport
		}
		keys[k.ID] = false
	}
	prev := 0
	for i, o := range r.Occurrences {
		s, ok := sources[o.SourceID]
		if !ok || s < prev || o.Ordinal != i+1 {
			return ErrReport
		}
		prev = s
		if !oneOf(o.Context, "code", "comment", "string", "template", "regexp", "html", "unknown") || !oneOf(o.IdentifierClass, "public_key", "public_key_str", "other", "unknown") || !oneOf(o.AssignmentClass, "declaration", "bare", "property", "none", "unknown") || !oneOf(o.Placement, "top_level", "function", "control", "other", "unknown") || !oneOf(o.KeyClass, "not_checked", "valid_rsa", "invalid_encoding", "non_rsa", "out_of_bounds") {
			return ErrReport
		}
		candidate := oneOf(o.AssignmentClass, "declaration", "bare", "property")
		kind := r.Sources[s-1].Kind
		if kind == "page" && o.Context != "html" || kind == "unknown_script" && o.Context != "unknown" {
			return ErrReport
		}
		if candidate && (o.Context != "code" || !oneOf(o.IdentifierClass, "public_key", "public_key_str") || o.Placement == "unknown" || o.KeyClass == "not_checked") {
			return ErrReport
		}
		if !candidate && (o.KeyClass != "not_checked" || o.KeyRef != nil) {
			return ErrReport
		}
		if o.Context != "code" && o.IdentifierClass != "unknown" {
			return ErrReport
		}
		if o.KeyClass == "valid_rsa" {
			if o.KeyRef == nil {
				return ErrReport
			}
			if _, ok := keys[*o.KeyRef]; !ok {
				return ErrReport
			}
			keys[*o.KeyRef] = true
		} else if o.KeyRef != nil {
			return ErrReport
		}
	}
	for _, used := range keys {
		if !used {
			return ErrReport
		}
	}
	m := r.ParserMatch
	if m.State != "mapped" || m.Occurrence < 1 || m.Occurrence > len(r.Occurrences) || r.Occurrences[m.Occurrence-1].SourceID != m.SourceID {
		return ErrReport
	}
	return nil
}
