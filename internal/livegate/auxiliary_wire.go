package livegate

import "encoding/json"

func (a Auxiliary) MarshalJSON() ([]byte, error) {
	type wire Auxiliary
	raw, err := json.Marshal(wire(a))
	if err == nil && a.SchemaVersion == 2 && a.WiFiSSID == nil {
		raw = append(raw[:len(raw)-1], []byte(`,"wifi_ssid":null}`)...)
	}
	return raw, err
}
