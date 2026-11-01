//go:build !windows

package casacquire

import "crypto/x509"

// Production acquisition is Windows-only. Linux tests supply loopback roots.
func localRoots(*rootSnapshot) (*x509.CertPool, map[[32]byte]bool, bool) { return nil, nil, false }
