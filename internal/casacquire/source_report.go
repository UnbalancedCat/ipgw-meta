package casacquire

import "github.com/UnbalancedCat/ipgw-meta/internal/cascollect"

func validSourceDifference(r Report) bool {
	d := r.SourceDifference
	if d == nil || r.Mode != "anonymous-once" || r.Collection != "blocked" || r.Analysis != nil || len(r.References)+len(r.Resources) != 0 || len(d.Sources) != 7 {
		return false
	}
	material, failed, notChecked := false, false, false
	failedReason := ""
	for i, s := range d.Sources {
		if s.Source != sourceNames[i] {
			return false
		}
		switch s.State {
		case "unavailable":
			if failed || notChecked || !sourceReadFailure(s.Reason) || s.Presence != "unknown" || s.ParseCoverage != "unknown" {
				return false
			}
			failed = true
			failedReason = s.Reason
		case "not_checked":
			if !failed || s.Reason != "not_checked" || s.Presence != "unknown" || s.ParseCoverage != "unknown" {
				return false
			}
			notChecked = true
		case "complete", "absent":
			if failed || notChecked || s.Reason != "none" || (s.Presence != "present" && s.Presence != "empty") || s.State == "absent" && s.Presence != "empty" {
				return false
			}
			if denialSource(i) {
				if s.ParseCoverage != "not_applicable" {
					return false
				}
				material = material || s.Presence == "present"
			} else {
				if s.Presence == "empty" {
					if s.ParseCoverage != "empty" {
						return false
					}
				} else if s.ParseCoverage != "complete" && s.ParseCoverage != "partial" && s.ParseCoverage != "none" {
					return false
				}
				material = material || s.ParseCoverage == "complete" || s.ParseCoverage == "partial"
			}
		default:
			return false
		}
		if d.Phase == "comparison" && d.Status == "complete" {
			if s.Facts == nil {
				return false
			}
			f := *s.Facts
			if (s.Presence == "empty" || s.ParseCoverage == "none") && f != (SourceFacts{}) {
				return false
			}
			if denialSource(i) && (f.TerminalSameSubjectSPKI || f.TerminalDifferentSubjectSPKI || f.AdditionalSigningParent) {
				return false
			}
		} else if s.Facts != nil {
			return false
		}
	}
	if d.Phase == "gate" {
		if r.Attempts != (cascollect.Attempts{}) || r.ContextConstructed || r.ChainRelation != nil || r.ChainTopology != nil {
			return false
		}
		if d.Status == "no_material" {
			return !material && !failed && d.Reason == "no_comparable_material" && r.Reason == "source_gate_no_material"
		}
		return d.Status == "unavailable" && failed && d.Reason == failedReason && sourceReadFailure(d.Reason) && r.Reason == "source_gate_unavailable"
	}
	if d.Phase != "comparison" || failed || !material {
		return false
	}
	switch d.Status {
	case "not_run":
		return d.Reason == "baseline_not_observed" && (r.ChainTopology == nil || r.ChainTopology.Status != "complete")
	case "complete", "unavailable":
		if r.ChainTopology == nil || r.ChainTopology.Status != "complete" || !validTopology(r.Reason, r.ChainRelation, r.ChainTopology) {
			return false
		}
		return d.Status == "complete" && d.Reason == "none" || d.Status == "unavailable" && sourceFailure(d.Reason)
	}
	return false
}
