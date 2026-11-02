package srun

import (
	"bytes"
	"github.com/UnbalancedCat/ipgw-meta/internal/wirejson"
)

// ParseQRContinuation returns a transient CAS instruction, never login proof.
func ParseQRContinuation(data []byte) (string, error) {
	if len(data) > 256<<10 {
		return "", ErrUnrecognized
	}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return "", ErrUnrecognized
	}
	o, err := wirejson.DecodeObjectOrJSONP(trimmed)
	if err != nil || len(o.Names()) != 4 {
		return "", ErrUnrecognized
	}
	for _, name := range o.Names() {
		switch name {
		case "code", "message", "Redirect", "ID":
		default:
			return "", ErrUnrecognized
		}
	}
	code, ok := requiredInteger(o, "code")
	if !ok || code != 0 {
		return "", ErrUnrecognized
	}
	for _, name := range []string{"message", "ID"} {
		raw, _ := o.Raw(name)
		if b := bytes.TrimSpace(raw); len(b) == 0 || b[0] != '"' {
			return "", ErrUnrecognized
		}
		if _, ok := requiredString(o, name); !ok {
			return "", ErrUnrecognized
		}
	}
	redirect, ok := requiredString(o, "Redirect")
	if !ok || redirect == "" || len(redirect) > 4096 {
		return "", ErrUnrecognized
	}
	return redirect, nil
}
