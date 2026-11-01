//go:build !windows

package casacquire

import "context"

func localSource(context.Context, int, func([]byte) string) (string, string) {
	return "unavailable", "unsupported"
}
