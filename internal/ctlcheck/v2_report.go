// Frozen D11 wire shape.
package ctlcheck

type v2Source struct {
	Source          string `json:"source"`
	State           string `json:"state"`
	Reason          string `json:"reason"`
	Read            string `json:"read"`
	Structure       string `json:"structure"`
	Hierarchy       string `json:"hierarchy"`
	Signature       string `json:"signature"`
	CertificateTime string `json:"certificate_time"`
	Conditions      string `json:"conditions"`
	Rejection       string `json:"rejection"`
	Window          string `json:"window"`
	Usage           string `json:"usage"`
	Attributes      string `json:"attributes"`
	MemberAlgorithm string `json:"member_algorithm"`
	Extensions      string `json:"extensions"`
}
type v2Report struct {
	SchemaVersion  int        `json:"schema_version"`
	ToolKind       string     `json:"tool_kind"`
	Mode           string     `json:"mode"`
	State          string     `json:"state"`
	Authentication string     `json:"authentication"`
	PurposePolicy  string     `json:"purpose_policy"`
	Revocation     string     `json:"revocation"`
	Latest         string     `json:"latest"`
	Sources        []v2Source `json:"sources"`
	Operations     Operations `json:"operations"`
}
