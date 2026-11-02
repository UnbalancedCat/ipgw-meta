package qrdev

import (
	"context"
	"github.com/UnbalancedCat/ipgw-meta/internal/labpath"
)

type PreflightReport struct {
	State           string `json:"state"`
	Reason          string `json:"reason"`
	NetworkRequests int    `json:"network_requests"`
	ProfileReads    int    `json:"profile_reads"`
}

func Preflight(ctx context.Context) (labpath.Snapshot, PreflightReport) {
	s, err := labpath.DiscoverWired(ctx)
	if err != nil {
		return s, PreflightReport{State: "blocked", Reason: "topology_unavailable"}
	}
	return s, PreflightReport{State: "ready", Reason: "none"}
}
func (r PreflightReport) Valid() bool {
	return r.NetworkRequests == 0 && r.ProfileReads == 0 && (r.State == "ready" && r.Reason == "none" || r.State == "blocked" && r.Reason == "topology_unavailable")
}
