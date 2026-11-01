//go:build !windows

package ctlcheck

import (
	"context"
	"crypto/x509"
)

func protectProcess() bool                                        { return false }
func localBlob(context.Context, int) ([]byte, string)             { return nil, "unsupported" }
func localMaterials(context.Context) (materials, string)          { return materials{}, "unsupported" }
func verifyNative(_ []byte, m *message, c *x509.Certificate) bool { return verifyGo(m, c) }
