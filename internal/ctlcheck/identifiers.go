package ctlcheck

import (
	"bytes"
	"context"
	"encoding/asn1"
	"encoding/json"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
)

type IdentifierSource struct {
	Source              string   `json:"source"`
	State               string   `json:"state"`
	Reason              string   `json:"reason"`
	MemberAlgorithmOID  string   `json:"member_algorithm_oid"`
	SignedAttributeOIDs []string `json:"signed_attribute_oids"`
}
type IdentifierReport struct {
	SchemaVersion     int                `json:"schema_version"`
	ToolKind          string             `json:"tool_kind"`
	Mode              string             `json:"mode"`
	State             string             `json:"state"`
	Authentication    string             `json:"authentication"`
	Signature         string             `json:"signature"`
	InputAuthenticity string             `json:"input_authenticity"`
	Sources           []IdentifierSource `json:"sources"`
	Operations        Operations         `json:"operations"`
}

func identifierSource(name, reason string) IdentifierSource {
	return IdentifierSource{Source: name, State: "unavailable", Reason: reason, SignedAttributeOIDs: []string{}}
}
func identifierReport(mode string) IdentifierReport {
	return IdentifierReport{1, "ctl-identifiers-dev", mode, "unavailable", "blocked", "not_checked", "unverified", []IdentifierSource{identifierSource("authroot_ctl", "read_failed"), identifierSource("disallowed_ctl", "read_failed")}, Operations{}}
}
func identifierOID(n node) (string, error) {
	s, e := oid(n)
	if e != nil || len(s) > 256 {
		return "", ErrInvalid
	}
	var arcs asn1.ObjectIdentifier
	rest, e := asn1.Unmarshal(n.raw, &arcs)
	if e != nil || len(rest) != 0 {
		return "", ErrInvalid
	}
	encoded, e := asn1.Marshal(arcs)
	if e != nil || !bytes.Equal(encoded, n.raw) {
		return "", ErrInvalid
	}
	return s, nil
}
func textOID(s string) (asn1.ObjectIdentifier, bool) {
	if len(s) == 0 || len(s) > 256 {
		return nil, false
	}
	parts := strings.Split(s, ".")
	arcs := make(asn1.ObjectIdentifier, len(parts))
	for i, p := range parts {
		if p == "" || len(p) > 1 && p[0] == '0' {
			return nil, false
		}
		for _, c := range p {
			if c < '0' || c > '9' {
				return nil, false
			}
		}
		n, e := strconv.Atoi(p)
		if e != nil {
			return nil, false
		}
		arcs[i] = n
	}
	b, e := asn1.Marshal(arcs)
	if e != nil {
		return nil, false
	}
	n, e := whole(b, 6)
	if e != nil {
		return nil, false
	}
	canonical, e := identifierOID(n)
	return arcs, e == nil && canonical == s
}
func oidLess(a, b string) bool {
	x, _ := textOID(a)
	y, _ := textOID(b)
	for i := 0; i < len(x) && i < len(y); i++ {
		if x[i] != y[i] {
			return x[i] < y[i]
		}
	}
	return len(x) < len(y)
}
func ValidIdentifiers(r IdentifierReport) bool {
	if r.SchemaVersion != 1 || r.ToolKind != "ctl-identifiers-dev" || !one(r.Mode, "self-test", "local-ctl-identifiers") || !one(r.State, "complete", "unavailable") || r.Authentication != "blocked" || r.Signature != "not_checked" || r.InputAuthenticity != "unverified" || r.Operations != (Operations{}) || len(r.Sources) != 2 {
		return false
	}
	complete := true
	for i, s := range r.Sources {
		if s.Source != []string{"authroot_ctl", "disallowed_ctl"}[i] || !one(s.State, "complete", "unavailable") || s.SignedAttributeOIDs == nil || len(s.SignedAttributeOIDs) > 16 {
			return false
		}
		if s.State == "unavailable" {
			complete = false
			if !one(s.Reason, "absent", "read_failed", "type_invalid", "size_changed", "limit", "parse_failed", "unsupported", "canceled", "deadline", "protection_failed") || s.MemberAlgorithmOID != "" || len(s.SignedAttributeOIDs) != 0 {
				return false
			}
			continue
		}
		if s.Reason != "none" || len(s.SignedAttributeOIDs) == 0 {
			return false
		}
		if _, ok := textOID(s.MemberAlgorithmOID); !ok {
			return false
		}
		for j, id := range s.SignedAttributeOIDs {
			if _, ok := textOID(id); !ok || j > 0 && !oidLess(s.SignedAttributeOIDs[j-1], id) {
				return false
			}
		}
	}
	return (r.State == "complete") == complete
}
func EncodeIdentifiers(r IdentifierReport) ([]byte, error) {
	if !ValidIdentifiers(r) {
		return nil, ErrInvalid
	}
	b, e := json.Marshal(r)
	if e != nil || len(b)+1 > 16384 {
		return nil, ErrInvalid
	}
	return append(b, '\n'), nil
}
func DecodeIdentifiers(b []byte) (IdentifierReport, error) {
	var r IdentifierReport
	if casprovenance.DecodeClosedJSON(b, &r, 16384) != nil || !ValidIdentifiers(r) {
		return IdentifierReport{}, ErrInvalid
	}
	return r, nil
}

// This locates two roles in bounded TLVs, not a validation/acceptance path.
// In particular it never calls parseCTL, parseMessage, x509, or a verifier.
func locateIdentifiers(ctx context.Context, b []byte) (member string, attributes []string, reason string) {
	defer func() {
		if stop := stopped(ctx); stop != "" {
			reason = stop
		}
		if reason != "" {
			member = ""
			attributes = nil
		}
	}()
	if stopped(ctx) != "" {
		return "", nil, stopped(ctx)
	}
	if len(b) == 0 || len(b) > maxBlob {
		return "", nil, "limit"
	}
	n, e := whole(b, 0x30)
	if e != nil {
		return "", nil, "parse_failed"
	}
	outer, e := children(n, 0x30, 2)
	if e != nil || len(outer) != 2 || outer[1].tag != 0xa0 {
		return "", nil, "parse_failed"
	}
	id, e := identifierOID(outer[0])
	if e != nil || id != oidSigned {
		return "", nil, "unsupported"
	}
	n, e = whole(outer[1].body, 0x30)
	if e != nil {
		return "", nil, "parse_failed"
	}
	sd, e := children(n, 0x30, 6)
	if e != nil || len(sd) != 5 || sd[1].tag != 0x31 || sd[3].tag != 0xa0 {
		return "", nil, "unsupported"
	}
	version := 0
	if bytes.Equal(sd[0].raw, []byte{2, 1, 1}) {
		version = 1
	}
	if bytes.Equal(sd[0].raw, []byte{2, 1, 3}) {
		version = 3
	}
	if version == 0 {
		return "", nil, "unsupported"
	}
	enc, e := children(sd[2], 0x30, 2)
	if e != nil || len(enc) != 2 || enc[1].tag != 0xa0 {
		return "", nil, "parse_failed"
	}
	id, e = identifierOID(enc[0])
	if e != nil || id != oidCTL {
		return "", nil, "unsupported"
	}
	content, rest, e := take(enc[1].body)
	if e != nil || len(rest) != 0 {
		return "", nil, "parse_failed"
	}
	if version == 3 {
		if content.tag != 4 {
			return "", nil, "unsupported"
		}
		content, e = whole(content.body, 0x30)
		if e != nil {
			return "", nil, "parse_failed"
		}
	} else if content.tag != 0x30 {
		return "", nil, "unsupported"
	}
	c, e := children(content, 0x30, 10)
	if e != nil {
		return "", nil, "parse_failed"
	}
	i := 0
	if len(c) > 0 && c[0].tag == 2 {
		if !bytes.Equal(c[0].raw, []byte{2, 1, 0}) {
			return "", nil, "unsupported"
		}
		i++
	}
	if i >= len(c) || c[i].tag != 0x30 {
		return "", nil, "parse_failed"
	}
	i++ // SubjectUsage, not interpreted.
	if i < len(c) && c[i].tag == 4 {
		if len(c[i].body) > 128 {
			return "", nil, "limit"
		}
		i++
	}
	if i < len(c) && c[i].tag == 2 {
		if !positive(c[i]) {
			return "", nil, "parse_failed"
		}
		i++
	}
	if i >= len(c) || !oneTimeTag(c[i].tag) {
		return "", nil, "parse_failed"
	}
	i++
	if i < len(c) && oneTimeTag(c[i].tag) {
		i++
	}
	if i >= len(c) {
		return "", nil, "parse_failed"
	}
	algorithm, e := children(c[i], 0x30, 2)
	if e != nil || len(algorithm) == 0 {
		return "", nil, "parse_failed"
	}
	member, e = identifierOID(algorithm[0])
	if e != nil {
		return "", nil, "parse_failed"
	}
	i++
	// Algorithm parameters and subsequent entries/extensions are not interpreted.
	if i < len(c) && c[i].tag == 0x30 {
		i++
	}
	if i < len(c) && c[i].tag == 0xa0 {
		i++
	}
	if i != len(c) {
		return "", nil, "parse_failed"
	}
	if stopped(ctx) != "" {
		return "", nil, stopped(ctx)
	}
	signers, e := children(sd[4], 0x31, 2)
	if e != nil || len(signers) != 1 {
		return "", nil, "unsupported"
	}
	si, e := children(signers[0], 0x30, 7)
	if e != nil || len(si) != 6 || !bytes.Equal(si[0].raw, []byte{2, 1, 1}) || si[2].tag != 0x30 || si[4].tag != 0x30 || si[5].tag != 4 {
		return "", nil, "unsupported"
	}
	identity, e := children(si[1], 0x30, 2)
	if e != nil || len(identity) != 2 || identity[0].tag != 0x30 || !positive(identity[1]) {
		return "", nil, "unsupported"
	}
	attrs, e := children(si[3], 0xa0, 16)
	if e != nil || len(attrs) == 0 {
		return "", nil, "parse_failed"
	}
	seen := map[string]bool{}
	for j, attr := range attrs {
		if stopped(ctx) != "" {
			return "", nil, stopped(ctx)
		}
		if j > 0 && bytes.Compare(attrs[j-1].raw, attr.raw) > 0 {
			return "", nil, "parse_failed"
		}
		pair, e := children(attr, 0x30, 2)
		if e != nil || len(pair) != 2 || pair[1].tag != 0x31 {
			return "", nil, "parse_failed"
		}
		id, e := identifierOID(pair[0])
		if e != nil || seen[id] {
			return "", nil, "parse_failed"
		}
		seen[id] = true
		if len(pair[1].raw) > 64<<10 {
			return "", nil, "limit"
		}
		values, e := children(pair[1], 0x31, 4)
		if e != nil || len(values) == 0 {
			return "", nil, "parse_failed"
		}
		for _, v := range values {
			if len(v.raw) > 64<<10 {
				return "", nil, "limit"
			}
		}
		attributes = append(attributes, id)
	}
	sort.Slice(attributes, func(i, j int) bool { return oidLess(attributes[i], attributes[j]) })
	return member, attributes, ""
}
func oneTimeTag(tag byte) bool { return tag == 23 || tag == 24 }
func auditIdentifiers(ctx context.Context, mode string, read blobReader) IdentifierReport {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	r := identifierReport(mode)
	for i := range r.Sources {
		reason := stopped(ctx)
		var b []byte
		if reason == "" {
			b, reason = read(ctx, i)
		}
		if stop := stopped(ctx); stop != "" {
			reason = stop
		}
		var member string
		var ids []string
		if reason == "" {
			member, ids, reason = locateIdentifiers(ctx, b)
		}
		if reason != "" {
			r.Sources[i] = identifierSource(r.Sources[i].Source, reason)
			continue
		}
		r.Sources[i] = IdentifierSource{r.Sources[i].Source, "complete", "none", member, ids}
	}
	if stop := stopped(ctx); stop != "" {
		for i := range r.Sources {
			r.Sources[i] = identifierSource(r.Sources[i].Source, stop)
		}
	}
	if r.Sources[0].State == "complete" && r.Sources[1].State == "complete" {
		r.State = "complete"
	}
	return r
}
func protectedIdentifiers(ctx context.Context, protect func() bool, work func() IdentifierReport) IdentifierReport {
	if stopped(ctx) != "" || !protect() {
		r := identifierReport("local-ctl-identifiers")
		reason := "protection_failed"
		if stop := stopped(ctx); stop != "" {
			reason = stop
		}
		for i := range r.Sources {
			r.Sources[i] = identifierSource(r.Sources[i].Source, reason)
		}
		return r
	}
	return work()
}
func RunIdentifiers(ctx context.Context) IdentifierReport {
	return protectedIdentifiers(ctx, protectProcess, func() IdentifierReport { return auditIdentifiers(ctx, "local-ctl-identifiers", localBlob) })
}

// Unsigned, generated TLV fixtures: no keys, certificate stores, or verifier.
func identifierFixture(member string, ids []string, cms bool, value []byte) []byte {
	attrs := [][]byte{}
	for _, id := range ids {
		attrs = append(attrs, attribute(id, der(4, value)))
	}
	sort.Slice(attrs, func(i, j int) bool { return bytes.Compare(attrs[i], attrs[j]) < 0 })
	content := der(0x30, der(0x30, encOID(oidRootList)), encTime(time.Unix(1700000000, 0)), encAlg(member), der(0x30))
	version := int64(1)
	if cms {
		version = 3
		content = der(4, content)
	}
	digest := encAlg("2.16.840.1.101.3.4.2.1")
	signer := der(0x30, encInt(1), der(0x30, der(0x30), encInt(1)), digest, der(0xa0, attrs...), encAlg("1.2.840.113549.1.1.1"), der(4, []byte{0}))
	return der(0x30, encOID(oidSigned), der(0xa0, der(0x30, encInt(version), der(0x31, digest), der(0x30, encOID(oidCTL), der(0xa0, content)), der(0xa0), der(0x31, signer))))
}
func IdentifiersSelfTest(ctx context.Context) IdentifierReport {
	return auditIdentifiers(ctx, "self-test", func(_ context.Context, i int) ([]byte, string) {
		return identifierFixture("1.2.3.40", []string{"1.2.3.20", "1.2.3.3"}, i == 1, []byte("synthetic-private-value")), ""
	})
}
func IdentifiersSyntheticProcess(ctx context.Context) IdentifierReport {
	if runtime.GOOS == "windows" && !protectProcess() {
		return identifierReport("self-test")
	}
	return IdentifiersSelfTest(ctx)
}
