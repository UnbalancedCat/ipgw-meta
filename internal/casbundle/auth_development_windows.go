//go:build windows

package casbundle

import (
	"errors"
	"golang.org/x/sys/windows"
	"io"
	"path/filepath"
	"sort"
	"strings"
)

func openAuthDevelopment(directory string) (*lockedBundle, string, string, error) {
	b, paths, err := openAuthFiles(directory, authNames[:])
	if err != nil {
		return nil, "", "", err
	}
	return b, paths[0], paths[2], nil
}

func openAuthFiles(directory string, names []string) (*lockedBundle, []string, error) {
	b, err := lockDirectories(directory)
	if err != nil {
		return nil, nil, ErrInvalid
	}
	fail := func() (*lockedBundle, []string, error) { b.close(); return nil, nil, ErrInvalid }
	entries, err := b.handles[len(b.handles)-1].Readdirnames(len(names) + 1)
	if len(entries) != len(names) || err != nil && !errors.Is(err, io.EOF) {
		return fail()
	}
	sort.Strings(entries)
	for i, name := range entries {
		if name != names[i] {
			return fail()
		}
		f, err := protectedOpen(filepath.Join(directory, name), false)
		if err != nil {
			return fail()
		}
		b.handles = append(b.handles, f)
		b.files[name] = f
	}
	paths := make([]string, len(names))
	for i, name := range names {
		paths[i], err = finalName(windows.Handle(b.files[name].Fd()), 1)
		if err != nil || !strings.HasPrefix(paths[i], `\\?\Volume{`) {
			return fail()
		}
	}
	return b, paths, nil
}
