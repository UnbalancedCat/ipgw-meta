//go:build !windows

package casbundle

func openDevelopment(string) (*lockedBundle, error) { return nil, ErrUnsupported }

func openDevelopmentNamed(string, string) (*lockedBundle, error) { return nil, ErrUnsupported }
