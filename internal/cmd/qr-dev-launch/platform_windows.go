//go:build windows

package main

import (
	"golang.org/x/sys/windows"
	"unsafe"
)

// The job handle lives until process exit; no child may outlive the window.
func protect() bool {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return false
	}
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	_, err = windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits)))
	if err != nil || windows.AssignProcessToJobObject(job, windows.CurrentProcess()) != nil {
		windows.CloseHandle(job)
		return false
	}
	return true
}
