package authdev

import (
	"encoding/json"
	"github.com/UnbalancedCat/ipgw-meta/internal/livegate"
	"strings"
	"testing"
)

func TestClosedReportAndFormalSeparation(t *testing.T) {
	r := NewReport("self-test")
	r.State = "pass"
	r.Reason = "none"
	data, err := Encode(r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Decode(data); err != nil {
		t.Fatal(err)
	}
	for _, changed := range []string{
		strings.Replace(string(data), `"schema_version":1`, `"schema_version":2`, 1),
		strings.Replace(string(data), `"release_evidence":false`, `"release_evidence":true`, 1),
		strings.Replace(string(data), `"release_evidence":false`, `"release_evidence":null`, 1),
		strings.Replace(string(data), `"state":"pass"`, `"state":"observed"`, 1),
		strings.Replace(string(data), `"application_requests":0`, `"application_requests":33`, 1),
		strings.Replace(string(data), `"reason":"none"`, `"reason":"sensitive-canary"`, 1),
		strings.Replace(string(data), `"schema_version":1,`, ``, 1),
		strings.Replace(string(data), `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1),
		string(data[:len(data)-1]) + `,"raw":"canary"}`,
	} {
		if _, err := Decode([]byte(changed)); err == nil {
			t.Fatal("accepted invalid report")
		}
	}
	var evidence livegate.Evidence
	_ = json.Unmarshal(data, &evidence)
	if evidence.Validate() == nil {
		t.Fatal("development accepted as formal evidence")
	}
}
func FuzzReport(f *testing.F) {
	data, _ := Encode(NewReport("anonymous"))
	f.Add(data)
	f.Fuzz(func(t *testing.T, b []byte) {
		r, err := Decode(b)
		if err == nil {
			out, err := Encode(r)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := Decode(out); err != nil {
				t.Fatal(err)
			}
		}
	})
}
