//go:build windows

package casacquire

import (
	"crypto/sha256"
	"crypto/x509"
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Read the local registry certificate stores only. x509.NewCertPool does NOT
// enable Go's Windows system verifier (which can retrieve missing certificates).
// Verification therefore uses Go's offline chain builder, not CryptoAPI URL
// retrieval. A missing chain fails; there is no alternate trust source.
func localRoots(capture *rootSnapshot) (*x509.CertPool, map[[32]byte]bool, bool) {
	roots := x509.NewCertPool()
	denied := map[[32]byte]bool{}
	count := 0
	for _, name := range []string{"Disallowed", "ROOT"} {
		for _, location := range []uint32{windows.CERT_SYSTEM_STORE_CURRENT_USER, windows.CERT_SYSTEM_STORE_LOCAL_MACHINE} {
			storeName, err := windows.UTF16PtrFromString(name)
			if err != nil {
				return nil, nil, false
			}
			h, err := windows.CertOpenStore(windows.CERT_STORE_PROV_SYSTEM_REGISTRY_W, 0, 0, location|windows.CERT_STORE_READONLY_FLAG|windows.CERT_STORE_OPEN_EXISTING_FLAG, uintptr(unsafe.Pointer(storeName)))
			if err != nil {
				if name == "Disallowed" && (errors.Is(err, windows.ERROR_FILE_NOT_FOUND) || errors.Is(err, windows.Errno(0x80092004))) {
					continue
				}
				return nil, nil, false
			}
			ok := func() bool {
				defer windows.CertCloseStore(h, 0)
				var previous *windows.CertContext
				for n := 0; ; n++ {
					certificate, err := windows.CertEnumCertificatesInStore(h, previous)
					previous = certificate
					if err != nil {
						return errors.Is(err, windows.Errno(0x80092004))
					}
					if certificate == nil {
						return false
					}
					if n >= 4096 || certificate.Length > 1<<20 {
						windows.CertFreeCertificateContext(certificate)
						return false
					}
					data := unsafe.Slice(certificate.EncodedCert, certificate.Length)
					digest := sha256.Sum256(data)
					if name == "Disallowed" {
						denied[digest] = true
						continue
					}
					capture.noteDER(data)
					if denied[digest] {
						continue
					}
					parsed, err := x509.ParseCertificate(data)
					if err != nil {
						continue
					}
					// ParseCertificate may retain DER slices; copy before enumeration
					// frees the Windows context on the next iteration.
					parsed, err = x509.ParseCertificate(append([]byte(nil), parsed.Raw...))
					if err != nil {
						return false
					}
					roots.AddCert(parsed)
					capture.keep(parsed)
					count++
				}
			}()
			if !ok {
				return nil, nil, false
			}
		}
	}
	return roots, denied, count > 0
}
