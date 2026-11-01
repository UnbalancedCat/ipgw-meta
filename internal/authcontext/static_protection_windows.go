//go:build windows

package authcontext

import (
	"golang.org/x/sys/windows"
	"unsafe"
)

// Job handle intentionally remains live until this dedicated process exits.
// Closing a KILL_ON_JOB_CLOSE handle here would kill the process before output.
func ProtectStaticProcess() bool {
	job, e := windows.CreateJobObject(nil, nil)
	if e != nil {
		return false
	}
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_ACTIVE_PROCESS | windows.JOB_OBJECT_LIMIT_PROCESS_MEMORY | windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	limits.BasicLimitInformation.ActiveProcessLimit = 1
	limits.ProcessMemoryLimit = 256 << 20
	_, e = windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits)))
	if e != nil {
		windows.CloseHandle(job)
		return false
	}
	if windows.AssignProcessToJobObject(job, windows.CurrentProcess()) != nil {
		windows.CloseHandle(job)
		return false
	}
	return true
}
