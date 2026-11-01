//go:build windows

package casbundle

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/sys/windows"
)

func pathPrefixes(name string) ([]string, error) {
	if len(name) < 4 || name[1] != ':' || name[2] != '\\' || !((name[0] >= 'A' && name[0] <= 'Z') || (name[0] >= 'a' && name[0] <= 'z')) || filepath.Clean(name) != name {
		return nil, ErrInvalid
	}
	parts := strings.Split(name[3:], `\`)
	result := []string{name[:3]}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") || strings.ContainsAny(part, `:/<>"|?*`) {
			return nil, ErrInvalid
		}
		for _, c := range part {
			if c < 32 {
				return nil, ErrInvalid
			}
		}
		result = append(result, filepath.Join(result[len(result)-1], part))
	}
	return result, nil
}

func finalName(handle windows.Handle, flags uint32) (string, error) {
	buffer := make([]uint16, 32768)
	n, err := windows.GetFinalPathNameByHandle(handle, &buffer[0], uint32(len(buffer)), flags)
	if err != nil || n == 0 || n >= uint32(len(buffer)) {
		return "", ErrInvalid
	}
	return windows.UTF16ToString(buffer[:n]), nil
}

func protectedOpen(name string, directory bool) (*os.File, error) {
	p, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, ErrInvalid
	}
	attributes := uint32(windows.FILE_FLAG_OPEN_REPARSE_POINT)
	if directory {
		attributes |= windows.FILE_FLAG_BACKUP_SEMANTICS
	}
	h, err := windows.CreateFile(p, windows.GENERIC_READ, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, attributes, 0)
	if err != nil {
		return nil, ErrInvalid
	}
	fail := func() (*os.File, error) { windows.CloseHandle(h); return nil, ErrInvalid }
	var info windows.ByHandleFileInformation
	if windows.GetFileInformationByHandle(h, &info) != nil || info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || (info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0) != directory || !directory && info.NumberOfLinks != 1 {
		return fail()
	}
	typ, err := windows.GetFileType(h)
	if err != nil || typ != windows.FILE_TYPE_DISK {
		return fail()
	}
	actual, err := finalName(h, 0)
	if err != nil || !strings.EqualFold(strings.TrimSuffix(actual, `\`), strings.TrimSuffix(`\\?\`+name, `\`)) {
		return fail()
	}
	return os.NewFile(uintptr(h), name), nil
}

func lockDirectories(directory string) (*lockedBundle, error) {
	prefixes, err := pathPrefixes(directory)
	if err != nil {
		return nil, ErrInvalid
	}
	root, _ := windows.UTF16PtrFromString(prefixes[0])
	if windows.GetDriveType(root) != windows.DRIVE_FIXED {
		return nil, ErrInvalid
	}
	b := &lockedBundle{files: map[string]*os.File{}}
	fail := func() (*lockedBundle, error) { b.close(); return nil, ErrInvalid }
	for _, prefix := range prefixes {
		f, err := protectedOpen(prefix, true)
		if err != nil {
			return fail()
		}
		b.handles = append(b.handles, f)
	}
	return b, nil
}

func openLocked(directory string) (*lockedBundle, error) {
	b, err := lockDirectories(directory)
	if err != nil {
		return nil, err
	}
	fail := func() (*lockedBundle, error) { b.close(); return nil, ErrInvalid }
	entries, err := b.handles[len(b.handles)-1].Readdirnames(5)
	// Request at most five: four exact names, no extra entry and bounded allocation.
	if len(entries) != 4 {
		return fail()
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return fail()
	}
	sort.Strings(entries)
	for i, name := range names {
		if entries[i] != name {
			return fail()
		}
	}
	for _, name := range names {
		f, err := protectedOpen(filepath.Join(directory, name), false)
		if err != nil {
			return fail()
		}
		b.handles = append(b.handles, f)
		b.files[name] = f
	}
	// Use a volume-GUID path for CreateProcess: DOS drive remapping cannot redirect it.
	b.executable, err = finalName(windows.Handle(b.files["cas-provenance.exe"].Fd()), 1)
	if err != nil || !strings.HasPrefix(b.executable, `\\?\Volume{`) {
		return fail()
	}
	return b, nil
}

func childEnvironment() []string {
	directory, err := windows.GetWindowsDirectory()
	if err != nil {
		return []string{}
	}
	return []string{"SystemRoot=" + directory, "WINDIR=" + directory}
}
