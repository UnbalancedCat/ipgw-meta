//go:build !windows

package casbundle

func openAuthDevelopment(string) (*lockedBundle, string, string, error) {
	return nil, "", "", ErrInvalid
}

func openAuthFiles(string, []string) (*lockedBundle, []string, error) { return nil, nil, ErrInvalid }
