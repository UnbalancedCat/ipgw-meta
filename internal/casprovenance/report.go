package casprovenance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"unicode/utf8"
)

func oneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

// ValidateReport enforces the prototype's closed projection, including references.
func ValidateReport(r Report) error {
	if r.SchemaVersion != 1 || r.ToolKind != "cas-provenance-dev" || r.Mode != "synthetic-prototype" || r.Binding != "not_packaged" || r.BundleID != nil || r.Operations != (Operations{}) || r.Sources == nil || r.Occurrences == nil || r.Keys == nil || r.Relations == nil || len(r.Relations) != 0 {
		return ErrReport
	}
	if r.Coverage.UseRelation != "unknown" {
		return ErrReport
	}
	if r.Analysis == "rejected" {
		if !oneOf(r.Reason, "resource_limit", "invalid_input", "canceled", "deadline", "internal_failure") || len(r.Sources)+len(r.Occurrences)+len(r.Keys) != 0 || r.Coverage != (Coverage{"unknown", "unknown", "unknown"}) {
			return ErrReport
		}
		return nil
	}
	if r.Analysis != "partial" || !oneOf(r.Reason, "use_relation_unknown", "unsupported_syntax") || !oneOf(r.Coverage.Lexical, "complete", "partial") || !oneOf(r.Coverage.Assignment, "complete", "partial") || len(r.Sources) < 1 || len(r.Sources) > MaxSources || len(r.Occurrences) > MaxOccurrences || len(r.Keys) > MaxOccurrences {
		return ErrReport
	}
	sourceNumbers := map[string]int{}
	partial := false
	pages := 0
	for i, s := range r.Sources {
		if s.ID != fmt.Sprintf("S%d", i+1) || s.Ordinal != i+1 || !oneOf(s.Kind, "page", "inline_script", "external_script", "unknown_script") || !oneOf(s.Analysis, "complete", "partial") || s.Kind == "unknown_script" && s.Analysis != "partial" {
			return ErrReport
		}
		sourceNumbers[s.ID] = i + 1
		if s.Kind == "page" {
			pages++
		}
		partial = partial || s.Analysis == "partial"
	}
	if pages > 1 || len(r.Sources)-pages > MaxSources-1 {
		return ErrReport
	}
	if partial != (r.Reason == "unsupported_syntax") || partial != (r.Coverage.Assignment == "partial") || r.Coverage.Lexical == "partial" && !partial {
		return ErrReport
	}
	keyIDs := map[string]bool{}
	for i, k := range r.Keys {
		if k.ID != fmt.Sprintf("K%d", i+1) || k.Bits < 512 || k.Bits > 8192 || !oneOf(k.ExponentClass, "common", "other") {
			return ErrReport
		}
		keyIDs[k.ID] = false
	}
	previousSource := 0
	for i, o := range r.Occurrences {
		sourceNumber := sourceNumbers[o.SourceID]
		if o.Ordinal != i+1 || sourceNumber == 0 || sourceNumber < previousSource || !oneOf(o.Context, "code", "comment", "string", "template", "html", "unknown") || !oneOf(o.IdentifierClass, "public_key", "public_key_str", "other", "unknown") || !oneOf(o.AssignmentClass, "bare", "declaration", "property", "none", "unknown") || !oneOf(o.KeyClass, "not_checked", "valid_rsa", "invalid_encoding", "non_rsa", "out_of_bounds") {
			return ErrReport
		}
		previousSource = sourceNumber
		source := r.Sources[sourceNumber-1]
		if (source.Kind == "page") != (o.Context == "html") || source.Kind == "unknown_script" && o.Context != "unknown" {
			return ErrReport
		}
		candidate := oneOf(o.AssignmentClass, "bare", "declaration", "property")
		if candidate && (o.Context != "code" || !oneOf(o.IdentifierClass, "public_key", "public_key_str")) {
			return ErrReport
		}
		if candidate == (o.KeyClass == "not_checked") {
			return ErrReport
		}
		if o.Context != "code" && o.IdentifierClass != "unknown" {
			return ErrReport
		}
		if (o.KeyRef != nil) != (o.KeyClass == "valid_rsa") {
			return ErrReport
		}
		if o.KeyRef != nil {
			if _, ok := keyIDs[*o.KeyRef]; !ok {
				return ErrReport
			}
			keyIDs[*o.KeyRef] = true
		}
	}
	for _, referenced := range keyIDs {
		if !referenced {
			return ErrReport
		}
	}
	return nil
}

func EncodeReport(r Report) ([]byte, error) {
	if err := ValidateReport(r); err != nil {
		return nil, ErrReport
	}
	data, err := json.Marshal(r)
	if err != nil || len(data)+1 > MaxReportBytes {
		return nil, ErrReport
	}
	return append(data, '\n'), nil
}

// DecodeReport rejects duplicate, missing and unknown fields at every level.
// All failures use a constant error and never expose the input or parser error.
func DecodeReport(data []byte) (Report, error) {
	var r Report
	if DecodeClosedJSON(data, &r, MaxReportBytes) != nil {
		return Report{}, ErrReport
	}
	if ValidateReport(r) != nil {
		return Report{}, ErrReport
	}
	return r, nil
}

// DecodeClosedJSON decodes internal fixed-schema records, never arbitrary maps.
func DecodeClosedJSON(data []byte, destination any, limit int) error {
	typ := reflect.TypeOf(destination)
	if typ == nil || typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct || reflect.ValueOf(destination).IsNil() {
		return ErrReport
	}
	if len(data) > limit || !utf8.Valid(data) || !json.Valid(data) {
		return ErrReport
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := uniqueJSON(decoder, 0); err != nil {
		return ErrReport
	}
	if _, err := decoder.Token(); err != io.EOF {
		return ErrReport
	}
	if !requiredShape(data, typ.Elem()) {
		return ErrReport
	}
	decoder = json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return ErrReport
	}
	return nil
}

func uniqueJSON(d *json.Decoder, depth int) error {
	if depth > 16 {
		return ErrReport
	}
	t, err := d.Token()
	if err != nil {
		return ErrReport
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for d.More() {
			t, err := d.Token()
			if err != nil {
				return ErrReport
			}
			name, ok := t.(string)
			if !ok || seen[name] {
				return ErrReport
			}
			seen[name] = true
			if err := uniqueJSON(d, depth+1); err != nil {
				return err
			}
		}
	case '[':
		for d.More() {
			if err := uniqueJSON(d, depth+1); err != nil {
				return err
			}
		}
	default:
		return ErrReport
	}
	end, err := d.Token()
	if err != nil || delim == '{' && end != json.Delim('}') || delim == '[' && end != json.Delim(']') {
		return ErrReport
	}
	return nil
}

func requiredShape(data json.RawMessage, typ reflect.Type) bool {
	if typ.Kind() == reflect.Pointer {
		if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
			return true
		}
		return requiredShape(data, typ.Elem())
	}
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return false
	}
	switch typ.Kind() {
	case reflect.Struct:
		var fields map[string]json.RawMessage
		if json.Unmarshal(data, &fields) != nil || len(fields) != typ.NumField() {
			return false
		}
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			raw, ok := fields[f.Tag.Get("json")]
			if !ok || !requiredShape(raw, f.Type) {
				return false
			}
		}
	case reflect.Slice:
		var items []json.RawMessage
		if json.Unmarshal(data, &items) != nil {
			return false
		}
		for _, item := range items {
			if !requiredShape(item, typ.Elem()) {
				return false
			}
		}
	}
	return true
}
