//go:build !windows

package auxgate

func userPathDigest() ([32]byte, error) { return [32]byte{}, errInstall }
func protectInstall() bool              { return false }
