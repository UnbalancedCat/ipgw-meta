//go:build !windows

package trustaudit

import "context"

func localRead(context.Context, string, int, *accumulator) (string, string) {
	return "unavailable", "unsupported"
}
