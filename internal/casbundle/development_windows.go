//go:build windows

package casbundle

import (
	"errors"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/sys/windows"
)

func openDevelopment(directory string) (*lockedBundle, error) {
	return openDevelopmentNamed(directory, "cas-acquire.exe")
}

func openDevelopmentNamed(directory, executable string) (*lockedBundle, error) {
	if executable != "cas-acquire.exe" && executable != "trust-audit.exe" && executable != "ctl-check.exe" && executable != "ctl-identifiers.exe" {
		return nil, ErrInvalid
	}
	b, err := lockDirectories(directory)
	if err != nil {
		return nil, err
	}
	fail := func() (*lockedBundle, error) { b.close(); return nil, ErrInvalid }
	entries, err := b.handles[len(b.handles)-1].Readdirnames(3)
	if len(entries) != 2 || (err != nil && !errors.Is(err, io.EOF)) {
		return fail()
	}
	sort.Strings(entries)
	if entries[0] != "build-inputs.json" || entries[1] != executable {
		return fail()
	}
	for _, name := range entries {
		f, err := protectedOpen(filepath.Join(directory, name), false)
		if err != nil {
			return fail()
		}
		b.handles = append(b.handles, f)
		b.files[name] = f
	}
	b.executable, err = finalName(windows.Handle(b.files[executable].Fd()), 1)
	if err != nil || !strings.HasPrefix(b.executable, `\\?\Volume{`) {
		return fail()
	}
	return b, nil
}
