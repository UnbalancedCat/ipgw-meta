//go:build windows

package ctlcheck

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"errors"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"runtime"
	"unsafe"
)

const ctlKey = `SOFTWARE\Microsoft\SystemCertificates\AuthRoot\AutoUpdate`

var crypt = windows.NewLazySystemDLL("crypt32.dll")
var createCTL = crypt.NewProc("CertCreateCTLContext")
var freeCTL = crypt.NewProc("CertFreeCTLContext")
var msgControl = crypt.NewProc("CryptMsgControl")

type ctlContext struct {
	encoding   uint32
	encoded    *byte
	encodedLen uint32
	info       uintptr
	store, msg windows.Handle
	content    *byte
	contentLen uint32
}
type verifyPara struct {
	size        uint32
	provider    uintptr
	index, kind uint32
	signer      *windows.CertContext
}

func verifyNative(b []byte, m *message, c *x509.Certificate) bool {
	if len(b) == 0 || len(b) > maxBlob || len(c.Raw) == 0 || len(c.Raw) > 64<<10 {
		return false
	}
	ptr, _, _ := createCTL.Call(0x10001, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)))
	runtime.KeepAlive(b)
	if ptr == 0 {
		return false
	}
	defer freeCTL.Call(ptr)
	// Copy only our own API-allocated header through the checked Win32 memory
	// reader. No uintptr-to-Go-pointer conversion and no other process handle.
	var cc ctlContext
	var copied uintptr
	if windows.ReadProcessMemory(windows.CurrentProcess(), ptr, (*byte)(unsafe.Pointer(&cc)), unsafe.Sizeof(cc), &copied) != nil || copied != unsafe.Sizeof(cc) {
		return false
	}
	if cc.msg == 0 || cc.content == nil || cc.contentLen == 0 || cc.contentLen > maxBlob || !bytes.Equal(unsafe.Slice(cc.content, cc.contentLen), m.content) {
		return false
	}
	signer, e := windows.CertCreateCertificateContext(windows.X509_ASN_ENCODING, &c.Raw[0], uint32(len(c.Raw)))
	runtime.KeepAlive(c)
	if e != nil {
		return false
	}
	defer windows.CertFreeCertificateContext(signer)
	p := verifyPara{size: uint32(unsafe.Sizeof(verifyPara{})), kind: 2, signer: signer}
	ok, _, _ := msgControl.Call(uintptr(cc.msg), 0, 19, uintptr(unsafe.Pointer(&p)))
	runtime.KeepAlive(p)
	return ok != 0
}

// Called by the dedicated command before any source read or native parsing.
// The handle intentionally lives until process exit: closing it kills this process.
func protectProcess() bool {
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
func localBlob(ctx context.Context, index int) ([]byte, string) {
	if r := stopped(ctx); r != "" {
		return nil, r
	}
	if index < 0 || index > 1 {
		return nil, "unsupported"
	}
	name := [2]string{"EncodedCtl", "DisallowedCertEncodedCtl"}[index]
	k, e := registry.OpenKey(registry.LOCAL_MACHINE, ctlKey, registry.QUERY_VALUE|registry.WOW64_64KEY)
	if errors.Is(e, windows.ERROR_FILE_NOT_FOUND) {
		return nil, "absent"
	}
	if e != nil {
		return nil, "read_failed"
	}
	defer k.Close()
	return readBinary(ctx, func(b []byte) (int, uint32, string) {
		n, typ, e := k.GetValue(name, b)
		if errors.Is(e, windows.ERROR_FILE_NOT_FOUND) {
			return 0, 0, "absent"
		}
		if errors.Is(e, windows.ERROR_MORE_DATA) {
			return 0, 0, "size_changed"
		}
		if e != nil {
			return 0, 0, "read_failed"
		}
		return n, typ, ""
	})
}
func localMaterials(ctx context.Context) (materials, string) {
	m := materials{denied: map[[32]byte]bool{}}
	for _, name := range []string{"Disallowed", "ROOT"} {
		count, total := 0, 0
		for _, location := range []uint32{windows.CERT_SYSTEM_STORE_CURRENT_USER, windows.CERT_SYSTEM_STORE_LOCAL_MACHINE} {
			if r := stopped(ctx); r != "" {
				return materials{}, r
			}
			p, _ := windows.UTF16PtrFromString(name)
			h, e := windows.CertOpenStore(windows.CERT_STORE_PROV_SYSTEM_REGISTRY_W, 0, 0, location|windows.CERT_STORE_READONLY_FLAG|windows.CERT_STORE_OPEN_EXISTING_FLAG, uintptr(unsafe.Pointer(p)))
			if errors.Is(e, windows.ERROR_FILE_NOT_FOUND) || errors.Is(e, windows.Errno(0x80092004)) {
				continue
			}
			if e != nil {
				return materials{}, "store_unavailable"
			}
			reason := collectStore(ctx, h, name, &m, &count, &total)
			closeErr := windows.CertCloseStore(h, 0)
			if reason != "" {
				return materials{}, reason
			}
			if closeErr != nil {
				return materials{}, "store_unavailable"
			}
		}
	}
	return m, ""
}
func collectStore(ctx context.Context, h windows.Handle, name string, m *materials, count, total *int) string {
	var prev *windows.CertContext
	defer func() {
		if prev != nil {
			windows.CertFreeCertificateContext(prev)
		}
	}()
	for {
		if r := stopped(ctx); r != "" {
			return r
		}
		c, e := windows.CertEnumCertificatesInStore(h, prev)
		prev = c
		if e != nil {
			if errors.Is(e, windows.Errno(0x80092004)) {
				return ""
			}
			return "store_unavailable"
		}
		if c == nil || c.EncodedCert == nil || c.Length == 0 {
			return "store_unavailable"
		}
		if *count >= 512 || c.Length > 64<<10 || int(c.Length) > (8<<20)-*total {
			return "limit"
		}
		*count++
		*total += int(c.Length)
		b := append([]byte(nil), unsafe.Slice(c.EncodedCert, c.Length)...)
		if name == "Disallowed" {
			m.denied[sha256.Sum256(b)] = true
		} else if cert, e := x509.ParseCertificate(b); e == nil {
			m.roots = append(m.roots, cert)
		}
	}
}
