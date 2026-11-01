package casbundle

import (
	"encoding/json"
	"strings"

	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
)

func validReceipt(r Receipt) bool {
	if r.SchemaVersion != 1 || r.ToolKind != ToolKind || r.Operations != (casprovenance.Operations{}) || r.EXEAttempts < 0 || r.EXEAttempts > 1 || r.EXEStarts < 0 || r.EXEStarts > r.EXEAttempts {
		return false
	}
	if r.Binding == "invalid" {
		return r.BundleID == nil && r.Result == nil && r.EXEAttempts == 0 && r.EXEStarts == 0 && r.Runner == "not_started" && (r.Reason == "binding_invalid" || r.Reason == "unsupported_platform" || r.Reason == "canceled")
	}
	if r.Binding != "verified" || r.BundleID == nil || !strings.HasPrefix(*r.BundleID, "dev-") || !digestPattern.MatchString(strings.TrimPrefix(*r.BundleID, "dev-")) || r.EXEAttempts != 1 {
		return false
	}
	if r.Runner == "failed" {
		return r.Result == nil && (r.Reason == "execution_failed" || r.EXEStarts == 1 && (r.Reason == "output_limit" || r.Reason == "canceled_or_timeout" || r.Reason == "output_invalid"))
	}
	if r.Runner != "passed" || r.Reason != "none" || r.EXEStarts != 1 || r.Result == nil {
		return false
	}
	data, err := json.Marshal(r.Result)
	if err != nil {
		return false
	}
	_, err = casprovenance.DecodeSelfTest(data)
	return err == nil
}

func EncodeReceipt(r Receipt) ([]byte, error) {
	if !validReceipt(r) {
		return nil, ErrInvalid
	}
	data, err := json.Marshal(r)
	if err != nil || len(data)+1 > MaxMetadata {
		return nil, ErrInvalid
	}
	return append(data, '\n'), nil
}

func DecodeReceipt(data []byte) (Receipt, error) {
	var r Receipt
	if casprovenance.DecodeClosedJSON(data, &r, MaxMetadata) != nil || !validReceipt(r) {
		return Receipt{}, ErrInvalid
	}
	return r, nil
}
