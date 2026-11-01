package ctlcheck

import (
	"bytes"
	"context"
	"encoding/asn1"
	"time"
)

const maxBlob = 4 << 20
const oidSigned = "1.2.840.113549.1.7.2"
const oidCTL = "1.3.6.1.4.1.311.10.1"
const oidRootList = "1.3.6.1.4.1.311.10.3.9"
const oidDisallowedList = "1.3.6.1.4.1.311.10.3.30"
const oidSignatureHashProperty = "1.3.6.1.4.1.311.10.11.15"
const oidSPCOpus = "1.3.6.1.4.1.311.2.1.12"
const oidSHA1 = "1.3.14.3.2.26"

type node struct {
	tag       byte
	raw, body []byte
}

// Nonrecursive DER cursor; only bounded, expected containers are walked.
func take(b []byte) (node, []byte, error) {
	if len(b) < 2 || b[0]&31 == 31 {
		return node{}, nil, ErrInvalid
	}
	n, h := int(b[1]), 2
	if n&128 != 0 {
		count := n & 127
		if count == 0 || count > 4 || len(b) < 2+count || b[2] == 0 {
			return node{}, nil, ErrInvalid
		}
		n = 0
		for _, v := range b[2 : 2+count] {
			n = n*256 + int(v)
		}
		h += count
		if n < 128 {
			return node{}, nil, ErrInvalid
		}
	}
	if n > len(b)-h {
		return node{}, nil, ErrInvalid
	}
	return node{b[0], b[:h+n], b[h : h+n]}, b[h+n:], nil
}
func children(n node, tag byte, max int) ([]node, error) {
	if n.tag != tag {
		return nil, ErrInvalid
	}
	var out []node
	b := n.body
	for len(b) > 0 {
		if len(out) >= max {
			return nil, ErrInvalid
		}
		v, rest, e := take(b)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
		b = rest
	}
	return out, nil
}
func whole(b []byte, tag byte) (node, error) {
	v, rest, e := take(b)
	if e != nil || len(rest) != 0 || v.tag != tag {
		return node{}, ErrInvalid
	}
	return v, nil
}
func oid(n node) (string, error) {
	if n.tag != 6 || len(n.body) == 0 || len(n.body) > 128 {
		return "", ErrInvalid
	}
	var v asn1.ObjectIdentifier
	rest, e := asn1.Unmarshal(n.raw, &v)
	if e != nil || len(rest) != 0 {
		return "", ErrInvalid
	}
	return v.String(), nil
}
func alg(n node) (string, error) {
	c, e := children(n, 0x30, 2)
	if e != nil || len(c) == 0 {
		return "", ErrInvalid
	}
	if len(c) == 2 && !bytes.Equal(c[1].raw, []byte{5, 0}) {
		return "", ErrInvalid
	}
	return oid(c[0])
}
func positive(n node) bool {
	return n.tag == 2 && len(n.body) > 0 && len(n.body) <= 32 && n.body[0]&128 == 0 && (len(n.body) == 1 || n.body[0] != 0 || n.body[1]&128 != 0)
}
func date(n node) (time.Time, error) {
	var t time.Time
	if n.tag != 23 && n.tag != 24 {
		return t, ErrInvalid
	}
	rest, e := asn1.Unmarshal(n.raw, &t)
	if e != nil || len(rest) != 0 {
		return t, ErrInvalid
	}
	return t, nil
}
func stopped(ctx context.Context) string {
	if ctx.Err() == context.DeadlineExceeded {
		return "deadline"
	}
	if ctx.Err() != nil {
		return "canceled"
	}
	return ""
}

type ctlMeta struct {
	this, next                     time.Time
	usage, attributes              string
	memberAlgorithm, extensions    string
	memberCoverage, signedCoverage string
	visitedBytes                   int
}

type parseFailure struct{ stage string }

func (parseFailure) Error() string { return "invalid CTL structure" }
func parseReason(err error) string {
	if failure, ok := err.(parseFailure); ok && isParseFailure("parse_failed_"+failure.stage) {
		return "parse_failed_" + failure.stage
	}
	return "parse_failed"
}

func parseCTL(ctx context.Context, b []byte) (out ctlMeta, err error) {
	stage := "ctl_header"
	defer func() {
		if err != nil {
			err = parseFailure{stage}
		}
	}()
	n, e := whole(b, 0x30)
	if e != nil {
		return out, e
	}
	c, e := children(n, 0x30, 10)
	if e != nil {
		return out, e
	}
	i := 0
	if len(c) > 0 && c[0].tag == 2 {
		if !bytes.Equal(c[0].body, []byte{0}) {
			return out, ErrInvalid
		}
		i++
	}
	stage = "ctl_usage"
	if i >= len(c) {
		return out, ErrInvalid
	}
	us, e := children(c[i], 0x30, 16)
	if e != nil {
		return out, e
	}
	i++
	out.usage = "other"
	for _, u := range us {
		s, e := oid(u)
		if e != nil {
			return out, e
		}
		if len(us) == 1 {
			if s == oidRootList {
				out.usage = "root_list"
			}
			if s == oidDisallowedList {
				out.usage = "disallowed_list"
			}
		}
	}
	stage = "ctl_identifier"
	if i < len(c) && c[i].tag == 4 {
		if len(c[i].body) > 128 {
			return out, ErrInvalid
		}
		i++
	}
	stage = "ctl_sequence"
	if i < len(c) && c[i].tag == 2 {
		if !positive(c[i]) {
			return out, ErrInvalid
		}
		i++
	}
	stage = "ctl_time"
	if i >= len(c) {
		return out, ErrInvalid
	}
	out.this, e = date(c[i])
	if e != nil {
		return out, e
	}
	i++
	if i < len(c) && (c[i].tag == 23 || c[i].tag == 24) {
		out.next, e = date(c[i])
		if e != nil {
			return out, e
		}
		i++
	}
	if i >= len(c) {
		return out, ErrInvalid
	}
	stage = "ctl_algorithm"
	a, e := alg(c[i])
	if e != nil {
		return out, ErrInvalid
	}
	identifierLength := 0
	out.memberCoverage = "exact_length_checked"
	switch a {
	case oidSignatureHashProperty:
		if out.usage != "disallowed_list" {
			stage = "ctl_algorithm_unsupported"
			return out, ErrInvalid
		}
		identifierLength = -1
		out.memberAlgorithm = "signature_hash_property"
		out.memberCoverage = "bounded_uninterpreted"
	case oidSHA1:
		identifierLength = 20
		out.memberAlgorithm = "sha1"
	case "2.16.840.1.101.3.4.2.1":
		identifierLength = 32
		out.memberAlgorithm = "sha256"
	case "2.16.840.1.101.3.4.2.2":
		identifierLength = 48
		out.memberAlgorithm = "sha384"
	case "2.16.840.1.101.3.4.2.3":
		identifierLength = 64
		out.memberAlgorithm = "sha512"
	}
	if identifierLength == 0 {
		stage = "ctl_algorithm_unsupported"
		return out, ErrInvalid
	}
	i++
	out.attributes = "absent"
	total := 0
	if i < len(c) && c[i].tag == 0x30 {
		stage = "ctl_entries"
		entries, e := children(c[i], 0x30, 16384)
		if e != nil {
			return out, e
		}
		i++
		for _, entry := range entries {
			stage = "ctl_entry"
			if stopped(ctx) != "" {
				return out, ErrInvalid
			}
			pair, e := children(entry, 0x30, 2)
			if e != nil || len(pair) == 0 || pair[0].tag != 4 || (identifierLength > 0 && len(pair[0].body) != identifierLength) || (identifierLength < 0 && (len(pair[0].body) == 0 || len(pair[0].body) > 64)) {
				return out, ErrInvalid
			}
			if identifierLength < 0 {
				if len(pair[0].body) > (8<<20)-total {
					return out, ErrInvalid
				}
				total += len(pair[0].body)
			}
			if len(pair) == 1 {
				continue
			} // Windows native synthetic decoding verified this optional form.
			stage = "ctl_attributes"
			attrs, e := children(pair[1], 0x31, 32)
			if e != nil {
				return out, e
			}
			if len(attrs) > 0 {
				out.attributes = "present_uninterpreted"
			}
			for _, attribute := range attrs {
				stage = "ctl_attribute"
				p, e := children(attribute, 0x30, 2)
				if e != nil || len(p) != 2 {
					return out, ErrInvalid
				}
				if _, e = oid(p[0]); e != nil {
					return out, e
				}
				stage = "ctl_values"
				vals, e := children(p[1], 0x31, 4)
				if e != nil || len(vals) == 0 {
					return out, ErrInvalid
				}
				for _, v := range vals {
					if len(v.raw) > 64<<10 || len(v.raw) > (8<<20)-total {
						return out, ErrInvalid
					}
					total += len(v.raw)
				}
			}
		}
	}
	out.extensions = "absent"
	if i < len(c) && c[i].tag == 0xa0 {
		stage = "ctl_extension_limit"
		if len(c[i].raw) > (8<<20)-total {
			return out, ErrInvalid
		}
		total += len(c[i].raw)
		stage = "ctl_extensions"
		container, e := whole(c[i].body, 0x30)
		if e != nil {
			return out, e
		}
		exts, e := children(container, 0x30, 32)
		if e != nil {
			return out, e
		}
		i++
		seen := map[string]bool{}
		for _, extension := range exts {
			stage = "ctl_extension"
			if stopped(ctx) != "" {
				return out, ErrInvalid
			}
			p, e := children(extension, 0x30, 3)
			if e != nil || len(p) < 2 {
				return out, ErrInvalid
			}
			id, e := oid(p[0])
			if e != nil {
				return out, e
			}
			if seen[id] {
				stage = "ctl_extension_duplicate"
				return out, ErrInvalid
			}
			seen[id] = true
			j := 1
			critical := false
			if p[j].tag == 1 {
				if !bytes.Equal(p[j].raw, []byte{1, 1, 255}) {
					return out, ErrInvalid
				}
				critical = true
				j++
			}
			if j != len(p)-1 || p[j].tag != 4 {
				return out, ErrInvalid
			}
			stage = "ctl_extension_limit"
			if len(p[j].body) > 64<<10 {
				return out, ErrInvalid
			}
			if out.extensions == "absent" {
				out.extensions = "present_uninterpreted"
			}
			if critical {
				out.extensions = "critical_present_uninterpreted"
			}
		}
	}
	stage = "ctl_tail"
	if i != len(c) {
		return out, ErrInvalid
	}
	if stopped(ctx) != "" {
		return out, ErrInvalid
	}
	out.visitedBytes = total
	return out, nil
}
