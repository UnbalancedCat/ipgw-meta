//go:build !windows

package casbundle

type installTarget struct{ created bool }

func (t *installTarget) close()                     {}
func (t *installTarget) write(string, []byte) error { return ErrUnsupported }
func openArchive(string) (*lockedBundle, error)     { return nil, ErrUnsupported }
func createTarget(string) (*installTarget, error)   { return nil, ErrUnsupported }
