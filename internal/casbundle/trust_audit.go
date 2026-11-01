package casbundle

import (
	"context"
	"time"
)

// TrustAudit is a separate bound identity with no live execution method.
type TrustAudit struct{ dev *Development }

func OpenTrustAudit(directory, exeDigest, inputsDigest string) (*TrustAudit, error) {
	d, err := openDevelopmentBound(directory, exeDigest, inputsDigest, "trust-audit.exe")
	if err != nil {
		return nil, err
	}
	return &TrustAudit{d}, nil
}
func (t *TrustAudit) Close() {
	if t != nil && t.dev != nil {
		t.dev.Close()
		t.dev = nil
	}
}
func (t *TrustAudit) Execute(ctx context.Context, mode string) ([]byte, error) {
	if t == nil || t.dev == nil || t.dev.bundle == nil || ctx == nil || (mode != "--self-test" && mode != "--trust-audit") {
		return nil, ErrInvalid
	}
	data, _, err := t.dev.executeBounded(ctx, mode, 15*time.Second)
	return data, err
}
