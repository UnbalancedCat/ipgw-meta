package ctlcheck

import (
	"bytes"
	"context"
	"crypto"
	"crypto/sha1"
	"crypto/x509"
	"encoding/json"
	"runtime"
	"time"
)

// This prototype is reachable only from generated self-tests. It is not a
// trust evaluator, and is not wired to local stores, a peer, or any product.
type memberObservation struct {
	state, reason, relation, entryAttributes, extensions, window string
}

func unknownMember(reason string) memberObservation {
	return memberObservation{"unavailable", reason, "unknown", "unknown", "unknown", "unknown"}
}

func syntheticMemberKey(source string, der []byte) ([]byte, error) {
	if len(der) == 0 || len(der) > 64<<10 {
		return nil, ErrInvalid
	}
	cert, e := x509.ParseCertificate(der)
	if e != nil {
		return nil, ErrInvalid
	}
	if source == "authroot_ctl" {
		h := sha1.Sum(der)
		return h[:], nil
	} // Identifier only; not a signature acceptance algorithm.
	if source != "disallowed_ctl" {
		return nil, ErrInvalid
	}
	outer, e := whole(der, 0x30)
	if e != nil {
		return nil, ErrInvalid
	}
	fields, e := children(outer, 0x30, 3)
	if e != nil || len(fields) != 3 {
		return nil, ErrInvalid
	}
	tbs, e := children(fields[0], 0x30, 20)
	if e != nil || len(tbs) < 3 {
		return nil, ErrInvalid
	}
	at := 1
	if tbs[0].tag == 0xa0 {
		at++
	}
	if !bytes.Equal(tbs[at].raw, fields[1].raw) {
		return nil, ErrInvalid
	}
	id, e := alg(fields[1])
	if e != nil {
		return nil, ErrInvalid
	}
	var hash crypto.Hash
	switch id {
	case "1.2.840.113549.1.1.11":
		hash = crypto.SHA256
	case "1.2.840.113549.1.1.12":
		hash = crypto.SHA384
	case "1.2.840.113549.1.1.13":
		hash = crypto.SHA512
	default:
		return nil, ErrInvalid
	}
	return hashed(hash, cert.RawTBSCertificate), nil
}

// The content was already checked by parseMessage; no new semantic acceptance.
func memberEntries(content []byte) ([]node, error) {
	n, e := whole(content, 0x30)
	if e != nil {
		return nil, e
	}
	c, e := children(n, 0x30, 10)
	if e != nil || len(c) < 3 {
		return nil, ErrInvalid
	}
	i := 0
	if c[i].tag == 2 {
		i++
	}
	i++
	if i < len(c) && c[i].tag == 4 {
		i++
	}
	if i < len(c) && c[i].tag == 2 {
		i++
	}
	if i >= len(c) || !oneTimeTag(c[i].tag) {
		return nil, ErrInvalid
	}
	i++
	if i < len(c) && oneTimeTag(c[i].tag) {
		i++
	}
	if i >= len(c) {
		return nil, ErrInvalid
	}
	i++
	if i < len(c) && c[i].tag == 0x30 {
		return children(c[i], 0x30, 16384)
	}
	return nil, nil
}

func observeSyntheticMember(ctx context.Context, source string, blob, target []byte, f fixture, verify verifier) (out memberObservation) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	defer func() {
		if reason := stopped(ctx); reason != "" {
			out = unknownMember(reason)
		}
	}()
	if stopped(ctx) != "" {
		return unknownMember(stopped(ctx))
	}
	key, e := syntheticMemberKey(source, target)
	if e != nil {
		return unknownMember("candidate_unsupported")
	}
	// Same bytes and same generated pins; a caller cannot supply a trusted=true flag.
	verified := observe(ctx, source, blob, f.mat, f.pins, f.now, verify)
	if verified.State != "complete" {
		return unknownMember("authentication_unavailable")
	}
	if source == "authroot_ctl" && (verified.Usage != "root_list" || verified.MemberAlgorithm != "sha1") || source == "disallowed_ctl" && (verified.Usage != "disallowed_list" || verified.MemberAlgorithm != "signature_hash_property") {
		return unknownMember("index_unsupported")
	}
	m, reason := parseMessage(ctx, blob)
	if reason != "" {
		return unknownMember("input_invalid")
	}
	entries, e := memberEntries(m.content)
	if e != nil {
		return unknownMember("input_invalid")
	}
	seen := map[string]bool{}
	relation, attributes := "no_match", "not_applicable"
	for _, entry := range entries {
		if stopped(ctx) != "" {
			return unknownMember(stopped(ctx))
		}
		pair, e := children(entry, 0x30, 2)
		if e != nil || len(pair) == 0 || pair[0].tag != 4 {
			return unknownMember("input_invalid")
		}
		id := string(pair[0].body)
		if seen[id] {
			return unknownMember("duplicate_identifier")
		}
		seen[id] = true
		if bytes.Equal(key, pair[0].body) {
			relation, attributes = "match", "absent"
			if len(pair) == 2 && len(pair[1].body) != 0 {
				attributes = "present_uninterpreted"
			}
		}
	}
	return memberObservation{"complete", "none", relation, attributes, verified.Extensions, verified.Window}
}

func memberSyntheticContent(f fixture, source string, entries ...[]byte) []byte {
	usage, algorithm := oidRootList, oidSHA1
	if source == "disallowed_ctl" {
		usage, algorithm = oidDisallowedList, oidSignatureHashProperty
	}
	return der(0x30, der(0x30, encOID(usage)), encTime(f.now), encAlg(algorithm), der(0x30, entries...))
}

// MemberSyntheticProcess has no input or local-reader capability. Its only
// returned JSON is a fixed summary after all generated checks succeed.
func MemberSyntheticProcess(ctx context.Context) ([]byte, error) {
	if runtime.GOOS == "windows" && !protectProcess() {
		return nil, ErrInvalid
	}
	f, e := generateFixture()
	if e != nil {
		return nil, ErrInvalid
	}
	for _, source := range []string{"authroot_ctl", "disallowed_ctl"} {
		key, e := syntheticMemberKey(source, f.leaf.Raw)
		if e != nil {
			return nil, ErrInvalid
		}
		other, e := syntheticMemberKey(source, f.root.Raw)
		if e != nil {
			return nil, ErrInvalid
		}
		member := der(0x30, der(4, key), der(0x31, attribute("1.2.3.4", der(4, []byte("synthetic-policy-only")))))
		for _, cms := range []bool{false, true} {
			b, e := f.signWithAttributes(memberSyntheticContent(f, source, member), cms, false, [][]byte{attribute(oidSPCOpus, der(0x30))})
			if e != nil {
				return nil, ErrInvalid
			}
			result := observeSyntheticMember(ctx, source, b, f.leaf.Raw, f, verifyNative)
			if result.state != "complete" || result.relation != "match" || result.entryAttributes != "present_uninterpreted" || result.window != "unbounded" {
				return nil, ErrInvalid
			}
			b, e = f.signWithAttributes(memberSyntheticContent(f, source, der(0x30, der(4, other))), cms, false, nil)
			if e != nil {
				return nil, ErrInvalid
			}
			if observeSyntheticMember(ctx, source, b, f.leaf.Raw, f, verifyNative).relation != "no_match" {
				return nil, ErrInvalid
			}
			b, e = f.signWithAttributes(memberSyntheticContent(f, source, member, member), cms, false, nil)
			if e != nil {
				return nil, ErrInvalid
			}
			if observeSyntheticMember(ctx, source, b, f.leaf.Raw, f, verifyNative).reason != "duplicate_identifier" {
				return nil, ErrInvalid
			}
			b[len(b)-1] ^= 1
			if observeSyntheticMember(ctx, source, b, f.leaf.Raw, f, verifyNative).relation != "unknown" {
				return nil, ErrInvalid
			}
		}
	}
	if stopped(ctx) != "" {
		return nil, ErrInvalid
	}
	r := struct {
		SchemaVersion  int        `json:"schema_version"`
		ToolKind       string     `json:"tool_kind"`
		Mode           string     `json:"mode"`
		State          string     `json:"state"`
		Authentication string     `json:"authentication"`
		Trust          string     `json:"trust"`
		Checks         []string   `json:"checks"`
		Operations     Operations `json:"operations"`
	}{1, "ctl-member-synthetic", "self-test", "complete", "blocked", "not_established", []string{"signed_member_match", "signed_member_absent", "opaque_policy_preserved", "duplicate_rejected", "tamper_rejected", "both_encapsulations"}, Operations{}}
	b, e := json.Marshal(r)
	if e != nil {
		return nil, ErrInvalid
	}
	return append(b, '\n'), nil
}
