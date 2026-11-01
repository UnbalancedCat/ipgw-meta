//go:build windows

package casbundle

import (
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

type installTarget struct {
	locks     *lockedBundle
	directory *os.File
	path      string
	created   bool
}

func (t *installTarget) close() { t.locks.close() }

func openArchive(path string) (*lockedBundle, error) {
	if _, err := pathPrefixes(path); err != nil {
		return nil, err
	}
	b, err := lockDirectories(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	f, err := protectedOpen(path, false)
	if err != nil {
		b.close()
		return nil, err
	}
	b.handles = append(b.handles, f)
	b.files["archive"] = f
	return b, nil
}

// Relative FILE_CREATE never follows or replaces a preexisting final member.
func createRelative(parent *os.File, name string, directory bool) (windows.Handle, error) {
	objectName, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return 0, ErrInvalid
	}
	attributes := &windows.OBJECT_ATTRIBUTES{RootDirectory: windows.Handle(parent.Fd()), ObjectName: objectName, Attributes: windows.OBJ_CASE_INSENSITIVE}
	attributes.Length = uint32(unsafe.Sizeof(*attributes))
	access := uint32(windows.FILE_GENERIC_READ)
	options := uint32(windows.FILE_SYNCHRONOUS_IO_NONALERT | windows.FILE_OPEN_REPARSE_POINT)
	if directory {
		options |= windows.FILE_DIRECTORY_FILE
	} else {
		options |= windows.FILE_NON_DIRECTORY_FILE
		access |= windows.FILE_GENERIC_WRITE
	}
	var handle windows.Handle
	var status windows.IO_STATUS_BLOCK
	var allocation int64
	err = windows.NtCreateFile(&handle, access, attributes, &status, &allocation, windows.FILE_ATTRIBUTE_NORMAL, windows.FILE_SHARE_READ, windows.FILE_CREATE, options, 0, 0)
	if err != nil {
		return 0, ErrInvalid
	}
	return handle, nil
}
func validateCreated(h windows.Handle, path string, directory bool) bool {
	var info windows.ByHandleFileInformation
	if windows.GetFileInformationByHandle(h, &info) != nil || info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || (info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0) != directory || !directory && info.NumberOfLinks != 1 {
		return false
	}
	typ, err := windows.GetFileType(h)
	if err != nil || typ != windows.FILE_TYPE_DISK {
		return false
	}
	actual, err := finalName(h, 0)
	return err == nil && strings.EqualFold(actual, `\\?\`+path)
}
func createTarget(path string) (*installTarget, error) {
	if _, err := pathPrefixes(path); err != nil {
		return nil, err
	}
	b, err := lockDirectories(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	t := &installTarget{locks: b, path: path}
	h, err := createRelative(b.handles[len(b.handles)-1], filepath.Base(path), true)
	if err != nil {
		b.close()
		return nil, err
	}
	t.created = true
	t.directory = os.NewFile(uintptr(h), path)
	b.handles = append(b.handles, t.directory)
	if !validateCreated(h, path, true) {
		return t, ErrInvalid
	}
	return t, nil
}
func (t *installTarget) write(name string, data []byte) error {
	allowed := false
	for _, n := range names {
		if n == name {
			allowed = true
		}
	}
	if !allowed {
		return ErrInvalid
	}
	h, err := createRelative(t.directory, name, false)
	if err != nil {
		return err
	}
	f := os.NewFile(uintptr(h), filepath.Join(t.path, name))
	if !validateCreated(h, filepath.Join(t.path, name), false) {
		f.Close()
		return ErrInvalid
	}
	n, writeErr := f.Write(data)
	syncErr := f.Sync()
	closeErr := f.Close()
	if writeErr != nil || syncErr != nil || closeErr != nil || n != len(data) {
		return ErrInvalid
	}
	return nil
}
