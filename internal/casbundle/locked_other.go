//go:build !windows

package casbundle

func openLocked(string) (*lockedBundle, error) { return nil, ErrUnsupported }
func childEnvironment() []string               { return []string{} }
