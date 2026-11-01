package casbundle

import (
	"context"
	"time"
)

type CTLIdentifiers struct{ dev *Development }

func OpenCTLIdentifiers(directory, exeDigest, inputsDigest string) (*CTLIdentifiers, error) {
	d, e := openDevelopmentBound(directory, exeDigest, inputsDigest, "ctl-identifiers.exe")
	if e != nil {
		return nil, e
	}
	return &CTLIdentifiers{d}, nil
}
func (c *CTLIdentifiers) Close() {
	if c != nil && c.dev != nil {
		c.dev.Close()
		c.dev = nil
	}
}
func (c *CTLIdentifiers) Execute(ctx context.Context, mode string) ([]byte, error) {
	if c == nil || c.dev == nil || c.dev.bundle == nil || ctx == nil || (mode != "--self-test" && mode != "--ctl-identifiers") {
		return nil, ErrInvalid
	}
	data, _, e := c.dev.executeBounded(ctx, mode, 15*time.Second)
	return data, e
}
