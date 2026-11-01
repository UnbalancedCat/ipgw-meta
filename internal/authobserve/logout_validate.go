package authobserve

import "slices"

func (o Observation) ValidateVersion(version int) error {
	if version != 1 && version != 2 || o.Validate() != nil {
		return ErrProtocol
	}
	if version == 1 {
		if o.Logout != nil || slices.Contains([]string{"logout_response_limit", "logout_response_parse", "logout_business_rejected"}, o.SDKCategory) {
			return ErrProtocol
		}
	} else if (o.Command == "logout") != (o.Logout != nil) {
		return ErrProtocol
	}
	return nil
}
func (w Window) ValidateVersion(version int) error {
	if version != 1 && version != 2 || w.Validate() != nil {
		return ErrProtocol
	}
	for _, o := range w.Commands {
		if o.ValidateVersion(version) != nil {
			return ErrProtocol
		}
	}
	return nil
}
func (o Observation) validateLogout() error {
	if o.Logout == nil {
		return nil
	} // Legacy: version layer checks presence.
	if o.Command != "logout" || o.Logout.Responses == nil {
		return ErrProtocol
	}
	l := o.Logout
	if l.State == "incomplete" {
		if o.State != "incomplete" || len(l.Responses) != 0 {
			return ErrProtocol
		}
		return nil
	}
	if o.State != "complete" {
		return ErrProtocol
	}
	eligible := 0
	for _, r := range o.Requests {
		if r.Role == "logout" && r.ResponseReceived && r.Transport == "none" {
			eligible++
		}
	}
	switch l.State {
	case "not_observed":
		if len(l.Responses) != 0 {
			return ErrProtocol
		}
	case "budget_exceeded":
		if len(l.Responses) != 0 || eligible < 3 {
			return ErrProtocol
		}
	case "observed":
		if len(l.Responses) < 1 || len(l.Responses) > 2 || len(l.Responses) > eligible {
			return ErrProtocol
		}
	default:
		return ErrProtocol
	}
	previous := -1
	for _, r := range l.Responses {
		if r.RequestIndex <= previous || r.RequestIndex >= len(o.Requests) {
			return ErrProtocol
		}
		previous = r.RequestIndex
		req := o.Requests[r.RequestIndex]
		if req.Role != "logout" || !req.ResponseReceived || req.Transport != "none" || r.validate() != nil {
			return ErrProtocol
		}
	}
	return nil
}
func (r LogoutResponse) validate() error {
	if !slices.Contains([]string{"complete", "too_large", "read_error", "incomplete", "budget_exceeded", "canceled"}, r.Read) || !slices.Contains([]string{"unknown", "empty", "json_object", "jsonp_object", "other", "malformed"}, r.Envelope) || !slices.Contains([]string{"none", "duplicate_key", "malformed", "depth_limit", "token_limit", "work_limit", "canceled"}, r.Structure) || !slices.Contains([]string{"unknown", "absent", "present"}, r.BOM) {
		return ErrProtocol
	}
	known := r.Read == "complete" && (r.Envelope == "json_object" || r.Envelope == "jsonp_object") && r.Structure == "none"
	if !known {
		if r.Fields != fields("unknown") || r.Ecode != "unknown" || r.Success != "unknown" || r.OtherFields != "unknown" || r.Parser != "unknown" {
			return ErrProtocol
		}
		if r.Read != "complete" {
			if r.Envelope != "unknown" || r.BOM != "unknown" {
				return ErrProtocol
			}
			if r.Read == "budget_exceeded" {
				if !slices.Contains([]string{"depth_limit", "token_limit", "work_limit", "none"}, r.Structure) {
					return ErrProtocol
				}
			} else if r.Read == "canceled" {
				if r.Structure != "none" && r.Structure != "canceled" {
					return ErrProtocol
				}
			} else if r.Structure != "none" {
				return ErrProtocol
			}
		} else if r.BOM == "unknown" || r.Envelope == "unknown" || r.Envelope == "malformed" && !slices.Contains([]string{"malformed", "duplicate_key"}, r.Structure) || r.Envelope != "malformed" && r.Structure != "none" {
			return ErrProtocol
		}
		return nil
	}
	if r.BOM != "absent" || !slices.Contains([]string{"absent", "present"}, r.OtherFields) {
		return ErrProtocol
	}
	for _, t := range []string{r.Fields.Ecode, r.Fields.Success, r.Fields.Code, r.Fields.Message, r.Fields.Error, r.Fields.ErrorMsg, r.Fields.Res} {
		if !slices.Contains([]string{"missing", "null", "boolean", "number", "string", "object", "array"}, t) {
			return ErrProtocol
		}
	}
	switch r.Fields.Ecode {
	case "missing":
		if r.Ecode != "missing" {
			return ErrProtocol
		}
	case "number":
		if !slices.Contains([]string{"integer_zero", "integer_nonzero", "noninteger_number", "integer_out_of_range"}, r.Ecode) {
			return ErrProtocol
		}
	case "string":
		if r.Ecode != "string_zero" && r.Ecode != "string_other" {
			return ErrProtocol
		}
	default:
		if r.Ecode != "other_type" {
			return ErrProtocol
		}
	}
	switch r.Fields.Success {
	case "missing":
		if r.Success != "missing" {
			return ErrProtocol
		}
	case "boolean":
		if r.Success != "true" && r.Success != "false" {
			return ErrProtocol
		}
	default:
		if r.Success != "other_type" {
			return ErrProtocol
		}
	}
	want := "rejected"
	if r.Ecode == "integer_zero" && (r.Success == "missing" || r.Success == "true") {
		want = "accepted_success"
	}
	if r.Ecode == "integer_nonzero" && (r.Success == "missing" || r.Success == "false") {
		want = "accepted_rejection"
	}
	if r.Parser != want {
		return ErrProtocol
	}
	return nil
}
