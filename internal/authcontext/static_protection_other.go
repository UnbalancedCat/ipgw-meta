//go:build !windows

package authcontext

func ProtectStaticProcess() bool { return false }
