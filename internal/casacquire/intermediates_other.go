//go:build !windows

package casacquire

import (
	"context"
)

// Real acquisition remains Windows-only. Tests inject generated certificates.
func localIntermediates(context.Context) (cacheSnapshot, string) {
	return cacheSnapshot{}, "tls_chain_probe_loader_failed"
}
