//go:build windows

package casacquire

import (
	"context"
	"errors"
	"golang.org/x/sys/windows"
	"unsafe"
)

func sourceStoreSpec(index int) (string, uint32, bool) {
	switch index {
	case 0:
		return "ROOT", windows.CERT_SYSTEM_STORE_CURRENT_USER_GROUP_POLICY, true
	case 1:
		return "ROOT", windows.CERT_SYSTEM_STORE_LOCAL_MACHINE_GROUP_POLICY, true
	case 2:
		return "CA", windows.CERT_SYSTEM_STORE_CURRENT_USER_GROUP_POLICY, true
	case 3:
		return "CA", windows.CERT_SYSTEM_STORE_LOCAL_MACHINE_GROUP_POLICY, true
	case 4:
		return "Disallowed", windows.CERT_SYSTEM_STORE_CURRENT_USER_GROUP_POLICY, true
	case 5:
		return "Disallowed", windows.CERT_SYSTEM_STORE_LOCAL_MACHINE_GROUP_POLICY, true
	case 6:
		return "AuthRoot", windows.CERT_SYSTEM_STORE_LOCAL_MACHINE, true
	}
	return "", 0, false
}
func localSource(ctx context.Context, index int, add func([]byte) string) (string, string) {
	if stop := sourceContext(ctx); stop != "" {
		return "unavailable", stop
	}
	name, location, ok := sourceStoreSpec(index)
	if !ok || add == nil {
		return "unavailable", "open_failed"
	}
	p, _ := windows.UTF16PtrFromString(name)
	h, err := windows.CertOpenStore(windows.CERT_STORE_PROV_SYSTEM_REGISTRY_W, 0, 0, location|windows.CERT_STORE_READONLY_FLAG|windows.CERT_STORE_OPEN_EXISTING_FLAG, uintptr(unsafe.Pointer(p)))
	if err != nil {
		if stop := sourceContext(ctx); stop != "" {
			return "unavailable", stop
		}
		if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) || errors.Is(err, windows.Errno(0x80092004)) {
			return "absent", ""
		}
		return "unavailable", "open_failed"
	}
	reason := readSourceStore(ctx, h, add)
	closeErr := windows.CertCloseStore(h, 0)
	if stop := sourceContext(ctx); stop != "" {
		reason = stop
	} else if closeErr != nil && reason == "" {
		reason = "read_failed"
	}
	if reason != "" {
		return "unavailable", reason
	}
	return "complete", ""
}
func readSourceStore(ctx context.Context, h windows.Handle, add func([]byte) string) string {
	var prior *windows.CertContext
	defer func() {
		if prior != nil {
			windows.CertFreeCertificateContext(prior)
		}
	}()
	for {
		if stop := sourceContext(ctx); stop != "" {
			return stop
		}
		c, err := windows.CertEnumCertificatesInStore(h, prior)
		prior = c
		if stop := sourceContext(ctx); stop != "" {
			return stop
		}
		if err != nil {
			if errors.Is(err, windows.Errno(0x80092004)) {
				return ""
			}
			return "read_failed"
		}
		if c == nil || c.EncodedCert == nil || c.Length == 0 {
			return "invalid_entry"
		}
		if c.Length > 64<<10 {
			return "limit"
		}
		if reason := add(unsafe.Slice(c.EncodedCert, c.Length)); reason != "" {
			return reason
		}
	}
}
