// Package caslegacy freezes the historical text matcher for diagnostic comparisons only.
package caslegacy

import "regexp"

var pattern = regexp.MustCompile(`(?i)(?:publicKeyStr|publicKey)\s*=\s*["']([A-Za-z0-9+/=]+)["']`)

func ExtractPublicKey(data []byte) string {
	m := pattern.FindSubmatch(data)
	if len(m) != 2 || len(m[1]) < 128 || len(m[1]) > 8192 {
		return ""
	}
	return string(m[1])
}
