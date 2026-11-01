//go:build windows

package casbundle

import "path/filepath"

// LockAuthProfile holds the selected configuration read-only while the product
// reopens it. No configuration bytes are read by the development launcher.
func LockAuthProfile(path string) (func(), error) {
	b, err := lockDirectories(filepath.Dir(path))
	if err != nil {
		return nil, ErrInvalid
	}
	f, err := protectedOpen(path, false)
	if err != nil {
		b.close()
		return nil, ErrInvalid
	}
	b.handles = append(b.handles, f)
	return b.close, nil
}
