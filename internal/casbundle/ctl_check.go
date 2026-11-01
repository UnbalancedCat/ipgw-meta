package casbundle

import (
	"context"
	"time"
)

type CTLCheck struct{ dev *Development }

func OpenCTLCheck(directory, exeDigest, inputsDigest string) (*CTLCheck, error) {
	d, e := openDevelopmentBound(directory, exeDigest, inputsDigest, "ctl-check.exe")
	if e != nil {
		return nil, e
	}
	return &CTLCheck{d}, nil
}
func (c *CTLCheck) Close() {
	if c != nil && c.dev != nil {
		c.dev.Close()
		c.dev = nil
	}
}
func (c *CTLCheck) Execute(ctx context.Context, mode string) ([]byte, error) {
	if c == nil || c.dev == nil || c.dev.bundle == nil || ctx == nil || (mode != "--self-test" && mode != "--ctl-check") {
		return nil, ErrInvalid
	}
	data, _, e := c.dev.executeBounded(ctx, mode, 15*time.Second)
	return data, e
}
