package ctlcheck

import (
	"context"
	"testing"
)

func TestCTLParseStageIsClosedAndPreservesRejection(t *testing.T) {
	f := synthetic(t)
	base := [][]byte{der(0x30, encOID(oidRootList)), encTime(f.now), encAlg(oidSHA1)}
	for _, tc := range []struct {
		name string
		data []byte
		want string
	}{
		{"header", []byte{0x30, 0x80, 0, 0}, "ctl_header"},
		{"usage", der(0x30, der(4, []byte{1})), "ctl_usage"},
		{"identifier", der(0x30, base[0], der(4, make([]byte, 129))), "ctl_identifier"},
		{"sequence", der(0x30, base[0], []byte{2, 1, 255}), "ctl_sequence"},
		{"time", der(0x30, base[0], der(23, []byte("bad"))), "ctl_time"},
		{"algorithm-unsupported", der(0x30, base[0], base[1], encAlg("1.2.3.4")), "ctl_algorithm_unsupported"},
		{"algorithm-malformed", der(0x30, base[0], base[1], der(0x30, der(4, []byte{1}))), "ctl_algorithm"},
		{"entry", der(0x30, base[0], base[1], base[2], der(0x30, der(0x30))), "ctl_entry"},
		{"attributes", der(0x30, base[0], base[1], base[2], der(0x30, der(0x30, der(4, make([]byte, 20)), der(4)))), "ctl_attributes"},
		{"attribute", der(0x30, base[0], base[1], base[2], der(0x30, der(0x30, der(4, make([]byte, 20)), der(0x31, der(0x30))))), "ctl_attribute"},
		{"values", der(0x30, base[0], base[1], base[2], der(0x30, der(0x30, der(4, make([]byte, 20)), der(0x31, der(0x30, encOID("1.2.3.4"), der(0x31)))))), "ctl_values"},
		{"extensions-malformed", der(0x30, base[0], base[1], base[2], der(0xa0, der(4))), "ctl_extensions"},
		{"tail", der(0x30, base[0], base[1], base[2], der(4)), "ctl_tail"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, e := parseCTL(context.Background(), tc.data)
			if e == nil || parseReason(e) != "parse_failed_"+tc.want {
				t.Fatalf("unexpected category: %s", parseReason(e))
			}
		})
	}
	for _, reason := range []string{"parse_failed_envelope", "parse_failed_ctl_algorithm", "parse_failed_certificate", "parse_failed_required_attributes"} {
		r := emptyReport("local-ctl-check")
		r.Sources[0].Read = "complete"
		r.Sources[0].Reason = reason
		b, e := Encode(r)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = Decode(b); e != nil {
			t.Fatal(e)
		}
		r.Sources[0].Structure = "valid"
		if Valid(r) {
			t.Fatal("parse rejection claimed valid structure")
		}
	}
	r := emptyReport("local-ctl-check")
	r.Sources[0].Read = "complete"
	r.Sources[0].Reason = "parse_failed_private_content"
	if Valid(r) {
		t.Fatal("unbounded reason")
	}
	legacy := emptyReport("local-ctl-check")
	legacy.Sources[0].Read = "complete"
	legacy.Sources[0].Reason = "parse_failed"
	if !Valid(legacy) {
		t.Fatal("old receipt invalidated")
	}
}
