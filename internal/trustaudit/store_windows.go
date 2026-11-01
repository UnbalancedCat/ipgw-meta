//go:build windows

package trustaudit

import (
	"context"
	"errors"
	"golang.org/x/sys/windows"
	"unsafe"
)

func localRead(ctx context.Context, name string, location int, a *accumulator) (string, string) {
	if s := stopped(ctx); s != "" {
		return "unavailable", s
	}
	if (name != "ROOT" && name != "CA" && name != "Disallowed") || location < 0 || location > 1 {
		return "unavailable", "open_failed"
	}
	loc := uint32(windows.CERT_SYSTEM_STORE_CURRENT_USER)
	if location == 1 {
		loc = windows.CERT_SYSTEM_STORE_LOCAL_MACHINE
	}
	p, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return "unavailable", "open_failed"
	}
	h, err := windows.CertOpenStore(windows.CERT_STORE_PROV_SYSTEM_REGISTRY_W, 0, 0, loc|windows.CERT_STORE_READONLY_FLAG|windows.CERT_STORE_OPEN_EXISTING_FLAG, uintptr(unsafe.Pointer(p)))
	if err != nil {
		if s := stopped(ctx); s != "" {
			return "unavailable", s
		}
		if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) || errors.Is(err, windows.Errno(0x80092004)) {
			return "absent", ""
		}
		return "unavailable", "open_failed"
	}
	reason := readStore(ctx, h, a)
	closeErr := windows.CertCloseStore(h, 0)
	if s := stopped(ctx); s != "" {
		reason = s
	} else if closeErr != nil && reason == "" {
		reason = "read_failed"
	}
	if reason != "" {
		return "unavailable", reason
	}
	return "complete", ""
}
func readStore(ctx context.Context, h windows.Handle, a *accumulator) string {
	var previous *windows.CertContext
	defer func() {
		if previous != nil {
			windows.CertFreeCertificateContext(previous)
		}
	}()
	for {
		if s := stopped(ctx); s != "" {
			return s
		}
		c, err := windows.CertEnumCertificatesInStore(h, previous)
		previous = c
		if s := stopped(ctx); s != "" {
			return s
		}
		if err != nil {
			if errors.Is(err, windows.Errno(0x80092004)) {
				return ""
			}
			return "read_failed"
		}
		if c == nil {
			return "read_failed"
		}
		if c.EncodedCert == nil || c.Length == 0 {
			return "invalid_entry"
		}
		if c.Length > maxDER || a.count >= maxEntries || int(c.Length) > maxBytes-a.bytes {
			return "limit"
		}
		// Only facts and hashes survive; Windows owns this temporary DER buffer.
		if reason := a.add(ctx, unsafe.Slice(c.EncodedCert, c.Length)); reason != "" {
			return reason
		}
	}
}
