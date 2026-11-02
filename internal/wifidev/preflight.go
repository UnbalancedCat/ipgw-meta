// Package wifidev contains non-release, local-only Wi-Fi preparation.
package wifidev

import (
	"context"
	"errors"
	ipgw "github.com/UnbalancedCat/ipgw-meta"
	"github.com/UnbalancedCat/ipgw-meta/internal/labpath"
)

type PreflightReport struct {
	State           string           `json:"state"`
	Reason          string           `json:"reason"`
	SSID            labpath.WiFiSSID `json:"wifi_ssid"`
	LocalItems      int              `json:"local_items"`
	SelectedItems   int              `json:"selected_items"`
	NetworkRequests int              `json:"network_requests"`
	ProfileReads    int              `json:"profile_reads"`
}

func (r PreflightReport) Valid(expected labpath.WiFiSSID) bool {
	if !expected.Supported() || r.SSID != expected || r.NetworkRequests != 0 || r.ProfileReads != 0 || r.LocalItems < 0 || r.LocalItems > 256 || r.SelectedItems < 0 || r.SelectedItems > r.LocalItems {
		return false
	}
	if r.State == "ready" {
		return r.Reason == "none" && r.SelectedItems == 1
	}
	if r.State != "blocked" {
		return false
	}
	switch r.Reason {
	case "arguments_rejected", "canceled", "topology_unavailable", "ssid_permission_denied", "wifi_disconnected", "ssid_unknown", "ssid_unsupported", "ssid_unavailable", "ssid_mismatch", "local_enumeration_failed", "local_selection_invalid":
		return true
	}
	return false
}

func Preflight(ctx context.Context, expected labpath.WiFiSSID) (labpath.Snapshot, PreflightReport) {
	c, err := ipgw.NewClient()
	if err != nil {
		return labpath.Snapshot{}, PreflightReport{State: "blocked", Reason: "local_enumeration_failed", SSID: expected}
	}
	return preflight(ctx, expected, labpath.DiscoverWiFi, c.ListInterfaces)
}

func preflight(ctx context.Context, expected labpath.WiFiSSID, capture func(context.Context) (labpath.Snapshot, error), list func(context.Context) ([]ipgw.Interface, error)) (labpath.Snapshot, PreflightReport) {
	r := PreflightReport{State: "blocked", Reason: "arguments_rejected", SSID: expected}
	if !expected.Supported() {
		r.SSID = ""
		return labpath.Snapshot{}, r
	}
	if ctx == nil || ctx.Err() != nil {
		r.Reason = "canceled"
		return labpath.Snapshot{}, r
	}
	s, err := capture(ctx)
	if err != nil {
		r.Reason = "topology_unavailable"
		switch {
		case errors.Is(err, labpath.ErrWiFiPermission):
			r.Reason = "ssid_permission_denied"
		case errors.Is(err, labpath.ErrWiFiDisconnected):
			r.Reason = "wifi_disconnected"
		case errors.Is(err, labpath.ErrWiFiUnknown):
			r.Reason = "ssid_unknown"
		case errors.Is(err, labpath.ErrWiFiUnsupported):
			r.Reason = "ssid_unsupported"
		case errors.Is(err, labpath.ErrWiFiUnavailable):
			r.Reason = "ssid_unavailable"
		}
		return labpath.Snapshot{}, r
	}
	if s.WiFiSSID != expected {
		r.Reason = "ssid_mismatch"
		return labpath.Snapshot{}, r
	}
	items, err := list(ctx)
	if err != nil || len(items) > 256 {
		r.Reason = "local_enumeration_failed"
		return labpath.Snapshot{}, r
	}
	r.LocalItems = len(items)
	r.SelectedItems, err = labpath.ValidateLocalSelection(ctx, items, s.Local, s.Index)
	if err != nil {
		r.Reason = "local_selection_invalid"
		return labpath.Snapshot{}, r
	}
	if ctx.Err() != nil {
		r.Reason = "canceled"
		return labpath.Snapshot{}, r
	}
	r.State, r.Reason = "ready", "none"
	return s, r
}

func SelfTest() bool {
	c, cancel := context.WithCancel(context.Background())
	cancel()
	_, r := preflight(c, labpath.WiFiNEU, func(context.Context) (labpath.Snapshot, error) { panic("canceled capture") }, nil)
	return r.State == "blocked" && r.Reason == "canceled" && r.NetworkRequests == 0 && r.ProfileReads == 0
}
