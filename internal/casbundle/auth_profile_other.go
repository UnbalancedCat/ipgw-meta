//go:build !windows

package casbundle

func LockAuthProfile(string) (func(), error) { return nil, ErrInvalid }
