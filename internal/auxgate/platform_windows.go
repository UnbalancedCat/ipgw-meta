//go:build windows

package auxgate

import (
	"crypto/sha256"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"unsafe"
)

func userPathDigest() ([32]byte, error) {
	k, e := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE)
	if e != nil {
		return [32]byte{}, e
	}
	defer k.Close()
	s, kind, e := k.GetStringValue("Path")
	if e == registry.ErrNotExist {
		return sha256.Sum256([]byte("absent")), nil
	}
	if e != nil {
		return [32]byte{}, e
	}
	return sha256.Sum256(append([]byte{byte(kind)}, []byte(s)...)), nil
}

// The handle intentionally lives until helper exit. Kernel handle cleanup
// terminates any remaining descendant, including after timeout/cancellation.
func protectInstall() bool {
	job, e := windows.CreateJobObject(nil, nil)
	if e != nil {
		return false
	}
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	_, e = windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits)))
	if e != nil || windows.AssignProcessToJobObject(job, windows.CurrentProcess()) != nil {
		windows.CloseHandle(job)
		return false
	}
	return true
}
