package authobserve

import (
	"bytes"
	"encoding/json"
	"io"
	"unicode/utf8"
)

// Unlike a closed record, the product branch has an intentional data/error
// XOR and transient opaque JSON. Validate it without weakening old decoders.
func strictObject(data []byte, limit int, keys []string) (map[string]json.RawMessage, error) {
	if len(data) > limit || !utf8.Valid(data) || !json.Valid(data) {
		return nil, ErrProtocol
	}
	d := json.NewDecoder(bytes.NewReader(data))
	if uniqueValue(d, 0) != nil {
		return nil, ErrProtocol
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, ErrProtocol
	}
	var m map[string]json.RawMessage
	if json.Unmarshal(data, &m) != nil || len(m) != len(keys) {
		return nil, ErrProtocol
	}
	for _, k := range keys {
		v, ok := m[k]
		if !ok || bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return nil, ErrProtocol
		}
	}
	return m, nil
}
func uniqueValue(d *json.Decoder, depth int) error {
	if depth > 16 {
		return ErrProtocol
	}
	token, err := d.Token()
	if err != nil {
		return ErrProtocol
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for d.More() {
			k, err := d.Token()
			if err != nil {
				return ErrProtocol
			}
			s, ok := k.(string)
			if !ok || seen[s] {
				return ErrProtocol
			}
			seen[s] = true
			if uniqueValue(d, depth+1) != nil {
				return ErrProtocol
			}
		}
	case '[':
		for d.More() {
			if uniqueValue(d, depth+1) != nil {
				return ErrProtocol
			}
		}
	default:
		return ErrProtocol
	}
	end, err := d.Token()
	if err != nil || delim == '{' && end != json.Delim('}') || delim == '[' && end != json.Delim(']') {
		return ErrProtocol
	}
	return nil
}
