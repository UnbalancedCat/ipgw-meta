package main

import (
	"regexp"
	"slices"

	"github.com/UnbalancedCat/ipgw-meta/internal/authobserve"
)

var digestShape = regexp.MustCompile(`^[a-f0-9]{64}$`)

func decodeReceipt(data []byte) (receipt, bool) {
	var r receipt
	if authobserve.DecodeClosed(data, &r, 64<<10) != nil || (r.SchemaVersion != 2 && r.SchemaVersion != 3) || r.ToolKind != "auth-dev-observed" || r.Observations.ValidateVersion(r.SchemaVersion-1) != nil {
		return receipt{}, false
	}
	if !slices.Contains([]string{"invalid", "verified"}, r.Binding) || !slices.Contains([]string{"arguments_rejected", "binding_rejected", "self_test_failed", "completed", "execution_indeterminate", "confirmation_required", "canceled", "report_rejected"}, r.State) {
		return receipt{}, false
	}
	if r.Binding == "verified" && (!digestShape.MatchString(r.SourceDigest) || !digestShape.MatchString(r.ProductDigest) || !digestShape.MatchString(r.ObservedDigest)) {
		return receipt{}, false
	}
	if r.Binding == "invalid" && r.Result != nil {
		return receipt{}, false
	}
	if r.Result == nil {
		return r, r.Observations.State == "not_started"
	}
	if r.Result.Validate() != nil || r.State != "completed" || r.Binding != "verified" {
		return receipt{}, false
	}
	if r.Result.Mode != "password" && r.Observations.State != "not_started" {
		return receipt{}, false
	}
	if (r.Result.Mode == "anonymous" || r.Result.Mode == "password" && r.Result.StepCoverage != "not_started") != r.LiveAttempted {
		return receipt{}, false
	}
	if r.Result.Mode == "password" {
		if r.Observations.State == "complete" && len(r.Observations.Commands) != r.Result.Commands {
			return receipt{}, false
		}
		if r.Observations.State == "not_started" && r.Result.Commands != 0 {
			return receipt{}, false
		}
		if r.Observations.State == "complete" {
			if r.Result.StepCoverage == "partial" || len(r.Observations.Commands) != len(r.Result.Steps)+1 || r.Observations.Commands[0].Command != "profile_show" {
				return receipt{}, false
			}
			for i, step := range r.Result.Steps {
				command := ""
				switch string(step.Name) {
				case "initial_status_offline", "status_online", "final_status_offline", "cleanup_status_offline":
					command = "status"
				case "login_logged_in", "second_login_already_online":
					command = "login"
				case "logout_logged_out", "second_logout_already_offline", "cleanup_logout":
					command = "logout"
				default:
					return receipt{}, false
				}
				o := r.Observations.Commands[i+1]
				if o.Command != command {
					return receipt{}, false
				}
				if o.SDKCode != "" && (step.ErrorCode == nil || o.SDKCode != string(*step.ErrorCode)) {
					return receipt{}, false
				}
			}
		}
	}
	return r, true
}
