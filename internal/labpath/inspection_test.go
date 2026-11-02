package labpath

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestInspectionChangesMatchOriginalEquality(t *testing.T) {
	base := topologyData{OK: true, Local: "192.0.2.11", Index: 7, DNS: []string{"192.0.2.53"}, Routes: []string{"synthetic-route"}, Adapters: []string{"synthetic-adapter"}}
	encode := func(v topologyData) Snapshot {
		raw, _ := json.Marshal(v)
		s, e := parseTopology(raw)
		if e != nil {
			t.Fatal(e)
		}
		s.network = "campus_wired"
		return s
	}
	b := encode(base)
	if compareSnapshots(b, b).Reason != "none" {
		t.Fatal("identical")
	}
	for _, field := range []string{"address", "interface", "dns", "routes", "adapters", "ssid"} {
		t.Run(field, func(t *testing.T) {
			v := base
			switch field {
			case "address":
				v.Local = "192.0.2.12"
			case "interface":
				v.Index++
			case "dns":
				v.DNS = []string{"192.0.2.54"}
			case "routes":
				v.Routes = []string{"private-route-canary"}
			case "adapters":
				v.Adapters = []string{"private-interface-canary"}
			}
			c := encode(v)
			if field == "ssid" {
				c.WiFiSSID = WiFiSSID("neu")
			}
			r := compareSnapshots(b, c)
			raw, _ := json.Marshal(r.Changes)
			var fields map[string]bool
			_ = json.Unmarshal(raw, &fields)
			count := 0
			for _, yes := range fields {
				if yes {
					count++
				}
			}
			if r.Reason != "snapshot_changed" || count != 1 || !fields[field] || b == c {
				t.Fatalf("wrong projection: %s", raw)
			}
			if strings.Contains(string(raw), "private") || strings.Contains(string(raw), "192.0.2") {
				t.Fatal("leak")
			}
		})
	}
}
