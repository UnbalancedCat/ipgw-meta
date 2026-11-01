package authobserve

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/UnbalancedCat/ipgw-meta/internal/srun"
)

const logoutLimit = 256 << 10

// All output is a closed projection. No input names, values or lengths escape.
type LogoutFields struct {
	Ecode    string `json:"ecode"`
	Success  string `json:"success"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Error    string `json:"error"`
	ErrorMsg string `json:"error_msg"`
	Res      string `json:"res"`
}
type LogoutResponse struct {
	RequestIndex int          `json:"request_index"`
	Read         string       `json:"read"`
	Envelope     string       `json:"envelope"`
	Structure    string       `json:"structure"`
	BOM          string       `json:"bom"`
	Fields       LogoutFields `json:"fields"`
	Ecode        string       `json:"ecode"`
	Success      string       `json:"success"`
	OtherFields  string       `json:"other_fields"`
	Parser       string       `json:"parser"`
}
type LogoutObservation struct {
	State     string           `json:"state"`
	Responses []LogoutResponse `json:"responses"`
}

func fields(value string) LogoutFields {
	return LogoutFields{value, value, value, value, value, value, value}
}
func (f *LogoutFields) slot(name string) *string {
	switch name {
	case "ecode":
		return &f.Ecode
	case "success":
		return &f.Success
	case "code":
		return &f.Code
	case "message":
		return &f.Message
	case "error":
		return &f.Error
	case "error_msg":
		return &f.ErrorMsg
	case "res":
		return &f.Res
	}
	return nil
}
func unknownLogout(index int, read string) LogoutResponse {
	return LogoutResponse{index, read, "unknown", "none", "unknown", fields("unknown"), "unknown", "unknown", "unknown", "unknown"}
}

// A reader bounds decoder scanning, including long strings/whitespace. Token
// and depth checks also cover More/Token recursion. No goroutine is abandoned.
type structureScan struct {
	ctx               context.Context
	data              []byte
	pos, work, tokens int
	reason            string
}

func (s *structureScan) check(work int) bool {
	s.work += work
	if s.ctx.Err() != nil {
		s.reason = "canceled"
	} else if s.work > 1000000 {
		s.reason = "work_limit"
	}
	return s.reason == ""
}
func (s *structureScan) Read(p []byte) (int, error) {
	if !s.check(0) {
		return 0, ErrProtocol
	}
	if s.pos == len(s.data) {
		return 0, io.EOF
	}
	n := min(len(p), 256, len(s.data)-s.pos)
	if !s.check(n) {
		return 0, ErrProtocol
	}
	copy(p, s.data[s.pos:s.pos+n])
	s.pos += n
	return n, nil
}
func (s *structureScan) token(d *json.Decoder) (json.Token, error) {
	s.tokens++
	if s.tokens > 32768 {
		s.reason = "token_limit"
		return nil, ErrProtocol
	}
	if !s.check(1) {
		return nil, ErrProtocol
	}
	v, e := d.Token()
	return v, e
}
func tokenType(t json.Token) string {
	switch v := t.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case string:
		return "string"
	case json.Number:
		return "number"
	case json.Delim:
		if v == '{' {
			return "object"
		}
		if v == '[' {
			return "array"
		}
	}
	return "unknown"
}
func integerSyntax(s string) bool {
	if len(s) > 0 && s[0] == '-' {
		s = s[1:]
	}
	if s == "" || len(s) > 1 && s[0] == '0' {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
func ecodeClass(t json.Token) string {
	switch v := t.(type) {
	case json.Number:
		if !integerSyntax(string(v)) {
			return "noninteger_number"
		}
		n, e := strconv.ParseInt(string(v), 10, 64)
		if e != nil {
			return "integer_out_of_range"
		}
		if n == 0 {
			return "integer_zero"
		}
		return "integer_nonzero"
	case string:
		if v == "0" {
			return "string_zero"
		}
		return "string_other"
	}
	return "other_type"
}
func (s *structureScan) value(d *json.Decoder, first json.Token, depth int, out *LogoutResponse) error {
	if !s.check(0) {
		return ErrProtocol
	}
	if depth > 32 {
		s.reason = "depth_limit"
		return ErrProtocol
	}
	delim, ok := first.(json.Delim)
	if !ok {
		return nil
	}
	if delim != '{' && delim != '[' {
		return ErrProtocol
	}
	seen := map[string]bool{}
	for d.More() {
		name := ""
		if delim == '{' {
			v, e := s.token(d)
			if e != nil {
				return e
			}
			var ok bool
			name, ok = v.(string)
			if !ok {
				return ErrProtocol
			}
			if seen[name] {
				s.reason = "duplicate_key"
				return ErrProtocol
			}
			seen[name] = true
		}
		v, e := s.token(d)
		if e != nil {
			return e
		}
		if depth == 1 && delim == '{' {
			if slot := out.Fields.slot(name); slot != nil {
				*slot = tokenType(v)
			} else {
				out.OtherFields = "present"
			}
			if name == "ecode" {
				out.Ecode = ecodeClass(v)
			}
			if name == "success" {
				out.Success = "other_type"
				if b, ok := v.(bool); ok {
					out.Success = strconv.FormatBool(b)
				}
			}
		}
		if e = s.value(d, v, depth+1, out); e != nil {
			return e
		}
	}
	v, e := s.token(d)
	if e != nil || delim == '{' && v != json.Delim('}') || delim == '[' && v != json.Delim(']') {
		return ErrProtocol
	}
	return nil
}
func callbackByte(b byte, first bool) bool {
	return b == '_' || b == '$' || b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z' || !first && b >= '0' && b <= '9'
}
func analyzeLogout(parent context.Context, data []byte, index int) LogoutResponse {
	out := unknownLogout(index, "complete")
	ctx, cancel := context.WithTimeout(parent, time.Second)
	defer cancel()
	s := &structureScan{ctx: ctx}
	reject := func(reason string) LogoutResponse {
		r := unknownLogout(index, "complete")
		r.BOM = out.BOM
		r.Envelope = "malformed"
		r.Structure = reason
		if reason == "canceled" {
			r = unknownLogout(index, "canceled")
			r.Structure = reason
		}
		if slices.Contains([]string{"depth_limit", "token_limit", "work_limit"}, reason) {
			r = unknownLogout(index, "budget_exceeded")
			r.Structure = reason
		}
		return r
	}
	if len(data) > logoutLimit {
		return unknownLogout(index, "too_large")
	}
	// Reserve bounded passes for envelope/UTF-8 and existing pure business
	// parser. The decoder's consumed bytes and tokens are charged separately.
	if !s.check(3 * len(data)) {
		return reject(s.reason)
	}
	out.BOM = "absent"
	if bytes.HasPrefix(data, []byte{0xef, 0xbb, 0xbf}) {
		out.BOM = "present"
	}
	if !utf8.Valid(data) {
		return reject("malformed")
	}
	payload := bytes.TrimSpace(data)
	if len(payload) == 0 {
		out.Envelope = "empty"
		return out
	}
	out.Envelope = "json_object"
	if payload[0] != '{' {
		if !callbackByte(payload[0], true) {
			out.Envelope = "other"
			return out
		}
		end := 0
		for end < len(payload) && callbackByte(payload[end], end == 0) {
			end++
		}
		rest := bytes.TrimSpace(payload[end:])
		if len(rest) == 0 || rest[0] != '(' {
			out.Envelope = "other"
			return out
		}
		rest = bytes.TrimSpace(rest[1:])
		if len(rest) > 0 && rest[len(rest)-1] == ';' {
			rest = bytes.TrimSpace(rest[:len(rest)-1])
		}
		if len(rest) == 0 || rest[len(rest)-1] != ')' {
			return reject("malformed")
		}
		payload = bytes.TrimSpace(rest[:len(rest)-1])
		if len(payload) == 0 || payload[0] != '{' {
			return reject("malformed")
		}
		out.Envelope = "jsonp_object"
	}
	out.Fields = fields("missing")
	out.Ecode = "missing"
	out.Success = "missing"
	out.OtherFields = "absent"
	s.data = payload
	d := json.NewDecoder(s)
	d.UseNumber()
	v, e := s.token(d)
	if e == nil && v == json.Delim('{') {
		e = s.value(d, v, 1, &out)
	} else {
		e = ErrProtocol
	}
	if e == nil {
		_, e = s.token(d)
		if errors.Is(e, io.EOF) {
			e = nil
		} else if e == nil {
			e = ErrProtocol
		}
	}
	if e != nil {
		if s.reason == "" {
			s.reason = "malformed"
		}
		return reject(s.reason)
	}
	if !s.check(0) {
		return reject(s.reason)
	}
	ok, e := srun.ParseLogoutSuccess(data)
	if !s.check(0) {
		return reject(s.reason)
	}
	out.Parser = "rejected"
	if e == nil {
		out.Parser = "accepted_rejection"
		if ok {
			out.Parser = "accepted_success"
		}
	}
	return out
}
