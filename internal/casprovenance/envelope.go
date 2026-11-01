package casprovenance

type SelfTestEnvelope struct {
	ToolKind       string `json:"tool_kind"`
	Mode           string `json:"mode"`
	Binding        string `json:"binding"`
	SelfTestResult string `json:"self_test_result"`
	Checks         int    `json:"checks"`
	Report         Report `json:"report"`
}

func DecodeSelfTest(data []byte) (SelfTestEnvelope, error) {
	var e SelfTestEnvelope
	if DecodeClosedJSON(data, &e, MaxReportBytes) != nil || e.ToolKind != "cas-provenance-dev" || e.Mode != "synthetic-prototype" || e.Binding != "not_packaged" || e.SelfTestResult != "passed" || e.Checks != 11 || ValidateReport(e.Report) != nil || e.Report.Analysis != "partial" {
		return SelfTestEnvelope{}, ErrReport
	}
	return e, nil
}
