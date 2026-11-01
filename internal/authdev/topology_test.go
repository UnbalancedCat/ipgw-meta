package authdev

import "testing"

func TestTopologyProjectionIgnoresJSONMemberOrdering(t *testing.T) {
	a, err := parseTopology([]byte(`{"ok":true,"local":"192.0.2.1","index":1,"dns":["192.0.2.2"],"routes":["synthetic-route"],"adapters":["synthetic-adapter"]}`))
	if err != nil {
		t.Fatal(err)
	}
	b, err := parseTopology([]byte(`{"index":1,"dns":["192.0.2.2"],"routes":["synthetic-route"],"adapters":["synthetic-adapter"],"ok":true,"local":"192.0.2.1"}`))
	if err != nil || a != b {
		t.Fatal("unstable topology fingerprint")
	}
	for _, data := range []string{`{}`, `{"ok":false,"local":"","index":0,"dns":[],"routes":[],"adapters":[]}`} {
		if _, err := parseTopology([]byte(data)); err == nil {
			t.Fatal("bad topology accepted")
		}
	}
}
