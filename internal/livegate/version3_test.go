package livegate

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestVersion3WiFiIdentityAndHistory(t *testing.T) {
	for _, ssid := range []string{"neu", "neu_24g"} {
		e := validPasswordEvidence()
		e.Platform, e.Testbed, e.NetworkType = PlatformWindowsAMD64, TestbedBHKWindows, NetworkCampusWiFi
		e.WiFiSSID = &ssid
		raw, err := json.Marshal(e)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = DecodeEvidence(bytes.NewReader(raw)); err != nil {
			t.Fatal(err)
		}
		for _, bad := range []string{"NEU", "neu_mobile", "", "private-canary"} {
			e.WiFiSSID = &bad
			if e.Validate() == nil {
				t.Fatal("invalid wifi label")
			}
		}
		e.WiFiSSID = nil
		if e.Validate() == nil {
			t.Fatal("missing wifi label")
		}
	}
	for _, version := range []int{1, 2} {
		e := validPasswordEvidence()
		e.SchemaVersion = version
		if version == 1 {
			e.RunnerSHA256 = ""
		}
		raw, err := json.Marshal(e)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(raw, []byte("wifi_ssid")) {
			t.Fatal("history rewritten")
		}
		got, err := DecodeEvidence(bytes.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		if EncodeEvidence(&bytes.Buffer{}, got) == nil {
			t.Fatal("old emission")
		}
		bad := strings.Replace(string(raw), `"steps":`, `"wifi_ssid":null,"steps":`, 1)
		if _, err = DecodeEvidence(strings.NewReader(bad)); err == nil {
			t.Fatal("mixed versions")
		}
	}
	e := validPasswordEvidence()
	raw, err := json.Marshal(e)
	if err != nil || !bytes.Contains(raw, []byte(`"wifi_ssid":null`)) {
		t.Fatal("wired null required")
	}
	bad := bytes.Replace(raw, []byte(`,"wifi_ssid":null`), nil, 1)
	if _, err = DecodeEvidence(bytes.NewReader(bad)); err == nil {
		t.Fatal("missing field")
	}
}
