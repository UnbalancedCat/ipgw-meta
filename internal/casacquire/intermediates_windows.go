//go:build windows

package casacquire

import (
	"context"
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

// This registry provider only enumerates existing local public certificates.
// It does not invoke CryptoAPI chain building or automatic URL retrieval.
func localIntermediates(ctx context.Context) (cacheSnapshot, string) {
	var cached cachedCertificates
	name, err := windows.UTF16PtrFromString("CA")
	if err != nil {
		return cacheSnapshot{}, "tls_chain_probe_store_open_failed"
	}
	for _, location := range []uint32{windows.CERT_SYSTEM_STORE_CURRENT_USER, windows.CERT_SYSTEM_STORE_LOCAL_MACHINE} {
		if reason := probeContext(ctx); reason != "" {
			return cacheSnapshot{}, reason
		}
		store, err := windows.CertOpenStore(windows.CERT_STORE_PROV_SYSTEM_REGISTRY_W, 0, 0, location|windows.CERT_STORE_READONLY_FLAG|windows.CERT_STORE_OPEN_EXISTING_FLAG, uintptr(unsafe.Pointer(name)))
		if err != nil {
			if reason := probeContext(ctx); reason != "" {
				return cacheSnapshot{}, reason
			}
			if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) || errors.Is(err, windows.Errno(0x80092004)) {
				continue
			}
			return cacheSnapshot{}, "tls_chain_probe_store_open_failed"
		}
		reason := readIntermediateStore(ctx, store, &cached)
		windows.CertCloseStore(store, 0)
		if canceled := probeContext(ctx); canceled != "" {
			return cacheSnapshot{}, canceled
		}
		if reason != "" {
			return cacheSnapshot{}, reason
		}
	}
	if reason := probeContext(ctx); reason != "" {
		return cacheSnapshot{}, reason
	}
	return cacheSnapshot{certs: cached.certs, partial: cached.partial}, ""
}

// The caller owns the store handle. Every enumeration context is freed, also
// on cancellation and limit failures. Tests use only a synthetic memory store.
func readIntermediateStore(ctx context.Context, store windows.Handle, cached *cachedCertificates) string {
	var previous *windows.CertContext
	for {
		if reason := probeContext(ctx); reason != "" {
			if previous != nil {
				windows.CertFreeCertificateContext(previous)
			}
			return reason
		}
		cert, err := windows.CertEnumCertificatesInStore(store, previous)
		previous = cert
		if reason := probeContext(ctx); reason != "" {
			if cert != nil {
				windows.CertFreeCertificateContext(cert)
			}
			return reason
		}
		if err != nil {
			if errors.Is(err, windows.Errno(0x80092004)) {
				return ""
			}
			return "tls_chain_probe_store_read_failed"
		}
		if cert == nil {
			return "tls_chain_probe_store_read_failed"
		}
		if cert.EncodedCert == nil || cert.Length == 0 {
			windows.CertFreeCertificateContext(cert)
			return "tls_chain_probe_cache_invalid"
		}
		if cert.Length > maxProbeDER || cached.count >= maxCachedCerts || int(cert.Length) > maxCachedBytes-cached.bytes {
			windows.CertFreeCertificateContext(cert)
			return "tls_chain_probe_cache_limit"
		}
		if !cached.add(ctx, unsafe.Slice(cert.EncodedCert, cert.Length)) {
			windows.CertFreeCertificateContext(cert)
			return cached.failure
		}
	}
}
