package authobserve

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"unicode/utf8"
)

// DecodeClosed keeps old records exact. The sole optional member is the new
// Observation.logout; version-specific validation MUST follow this framing.
// Explicit null is not equivalent to omission for that member.
func DecodeClosed(data []byte, dst any, limit int) error {
	t := reflect.TypeOf(dst)
	if t == nil || t.Kind() != reflect.Pointer || t.Elem().Kind() != reflect.Struct || reflect.ValueOf(dst).IsNil() || len(data) > limit || !utf8.Valid(data) || !json.Valid(data) {
		return ErrProtocol
	}
	d := json.NewDecoder(bytes.NewReader(data))
	if uniqueValue(d, 0) != nil {
		return ErrProtocol
	}
	if _, e := d.Token(); e != io.EOF {
		return ErrProtocol
	}
	if !recordShape(data, t.Elem()) {
		return ErrProtocol
	}
	d = json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if d.Decode(dst) != nil {
		return ErrProtocol
	}
	return nil
}
func recordShape(data json.RawMessage, t reflect.Type) bool {
	if t.Kind() == reflect.Pointer {
		if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
			return true
		}
		return recordShape(data, t.Elem())
	}
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return false
	}
	switch t.Kind() {
	case reflect.Struct:
		var m map[string]json.RawMessage
		if json.Unmarshal(data, &m) != nil {
			return false
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			v, ok := m[name]
			optional := t == reflect.TypeOf(Observation{}) && name == "logout"
			if !ok {
				if optional {
					continue
				}
				return false
			}
			if optional && bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
				return false
			}
			if !recordShape(v, f.Type) {
				return false
			}
			delete(m, name)
		}
		return len(m) == 0
	case reflect.Slice:
		var values []json.RawMessage
		if json.Unmarshal(data, &values) != nil {
			return false
		}
		for _, v := range values {
			if !recordShape(v, t.Elem()) {
				return false
			}
		}
	}
	return true
}
