package labpath

import "context"

// Changes contains categories only, never interface identities or addresses.
type Changes struct {
	Address   bool `json:"address"`
	Interface bool `json:"interface"`
	DNS       bool `json:"dns"`
	Routes    bool `json:"routes"`
	Adapters  bool `json:"adapters"`
	SSID      bool `json:"ssid"`
}

func (c Changes) Any() bool {
	return c.Address || c.Interface || c.DNS || c.Routes || c.Adapters || c.SSID
}

type Inspection struct {
	Reason      string
	QueryReason string
	Changes     Changes
}

func compareSnapshots(base, current Snapshot) Inspection {
	c := Changes{Address: base.components[0] != current.components[0], Interface: base.components[1] != current.components[1], DNS: base.components[2] != current.components[2], Routes: base.components[3] != current.components[3], Adapters: base.components[4] != current.components[4], SSID: base.WiFiSSID != current.WiFiSSID}
	if c.Any() {
		return Inspection{Reason: "snapshot_changed", QueryReason: "none", Changes: c}
	}
	if base != current {
		return Inspection{Reason: "query_unavailable", QueryReason: "structure_rejected"}
	}
	return Inspection{Reason: "none", QueryReason: "none"}
}

// Inspect uses precisely the existing local query and limits. Original Check
// callers retain their boolean contract; no extra requests are introduced.
func (base Snapshot) Inspect(ctx context.Context) Inspection {
	prefix := "$selectedType=6; $otherType=71\n"
	if base.network == "campus_wifi" {
		prefix = "$selectedType=71; $otherType=6\n"
	} else if base.network != "campus_wired" {
		return Inspection{Reason: "query_unavailable", QueryReason: "unsupported"}
	}
	current, reason := queryTopologyDetailed(ctx, prefix+topologyScript)
	if reason != "none" {
		return Inspection{Reason: "query_unavailable", QueryReason: reason}
	}
	current.network = base.network
	if base.network == "campus_wifi" {
		ssid, err := ReadWiFiSSID(ctx, current.Index)
		if err != nil {
			return Inspection{Reason: "query_unavailable", QueryReason: "query_rejected"}
		}
		current.WiFiSSID = ssid
	}
	return compareSnapshots(base, current)
}
