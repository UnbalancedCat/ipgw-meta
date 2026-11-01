package labpath

import (
	"strings"
	"testing"
)

func TestTopologyProjectionFailsClosed(t *testing.T) {
	good := `{"ok":true,"local":"192.0.2.11","index":7,"dns":["192.0.2.53"],"routes":["0.0.0.0/0|7|192.0.2.1|1"],"adapters":["7|synthetic"]}`
	if _, err := parseTopology([]byte(good)); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{good + good, strings.Replace(good, `"ok":true`, `"ok":true,"ok":true`, 1), strings.Replace(good, `"ok"`, `"OK"`, 1), strings.Replace(good, `"index":7,`, "", 1), strings.Replace(good, "192.0.2.11", "127.0.0.1", 1), strings.Replace(good, "192.0.2.53", "bad", 1)} {
		if _, err := parseTopology([]byte(raw)); err == nil {
			t.Fatal("invalid projection")
		}
	}
}
func FuzzTopology(f *testing.F) {
	f.Add([]byte(`{}`))
	f.Fuzz(func(t *testing.T, raw []byte) { _, _ = parseTopology(raw) })
}
