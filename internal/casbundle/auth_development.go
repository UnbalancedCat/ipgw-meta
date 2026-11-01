package casbundle

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
)

// AuthDevelopment holds a separately identified product-equivalent dev bundle.
// It does not accept Candidate manifests or widen historical diagnostic modes.
type AuthDevelopment struct {
	b            *lockedBundle
	hashes       []string
	names        []string
	ObserverPath string
	ProductPath  string
	ObservedPath string
}

var authNames = [3]string{"auth-dev.exe", "build-inputs.json", "ipgw-meta.exe"}

var authContextNames = []string{"auth-context.exe", "build-inputs.json", "ipgw-meta.exe"}

// OpenAuthContext binds the separate anonymous-context identity. No historical
// tool name or allowlist is reinterpreted.
func OpenAuthContext(directory, observerHash, inputsHash, productHash string) (*AuthDevelopment, error) {
	hashes := []string{observerHash, inputsHash, productHash}
	for _, h := range hashes {
		if !digestPattern.MatchString(h) {
			return nil, ErrInvalid
		}
	}
	b, paths, err := openAuthFiles(directory, authContextNames)
	if err != nil {
		return nil, ErrInvalid
	}
	d := &AuthDevelopment{b: b, names: authContextNames, hashes: hashes, ObserverPath: paths[0], ProductPath: paths[2]}
	if d.Verify() != nil {
		d.Close()
		return nil, ErrInvalid
	}
	return d, nil
}

var authStaticContextNames = []string{"auth-static-context.exe", "build-inputs.json", "ipgw-meta.exe"}

// OpenAuthStaticContext binds the separate anonymous-context identity. No historical
// tool name or allowlist is reinterpreted.
func OpenAuthStaticContext(directory, observerHash, inputsHash, productHash string) (*AuthDevelopment, error) {
	hashes := []string{observerHash, inputsHash, productHash}
	for _, h := range hashes {
		if !digestPattern.MatchString(h) {
			return nil, ErrInvalid
		}
	}
	b, paths, err := openAuthFiles(directory, authStaticContextNames)
	if err != nil {
		return nil, ErrInvalid
	}
	d := &AuthDevelopment{b: b, names: authStaticContextNames, hashes: hashes, ObserverPath: paths[0], ProductPath: paths[2]}
	if d.Verify() != nil {
		d.Close()
		return nil, ErrInvalid
	}
	return d, nil
}

func OpenAuthDevelopment(directory, observerHash, inputsHash, productHash string) (*AuthDevelopment, error) {
	hashes := [3]string{observerHash, inputsHash, productHash}
	for _, hash := range hashes {
		if !digestPattern.MatchString(hash) {
			return nil, ErrInvalid
		}
	}
	b, observer, product, err := openAuthDevelopment(directory)
	if err != nil {
		return nil, ErrInvalid
	}
	d := &AuthDevelopment{b: b, hashes: hashes[:], names: authNames[:], ObserverPath: observer, ProductPath: product}
	if d.Verify() != nil {
		d.Close()
		return nil, ErrInvalid
	}
	return d, nil
}

var authObservedNames = []string{"auth-dev.exe", "auth-observed.exe", "build-inputs.json", "ipgw-meta.exe"}

// OpenAuthObserved is a new identity, never a relaxed interpretation of an
// existing three-file P2 bundle.
func OpenAuthObserved(directory, observerHash, observedHash, inputsHash, productHash string) (*AuthDevelopment, error) {
	hashes := []string{observerHash, observedHash, inputsHash, productHash}
	for _, h := range hashes {
		if !digestPattern.MatchString(h) {
			return nil, ErrInvalid
		}
	}
	b, paths, err := openAuthFiles(directory, authObservedNames)
	if err != nil {
		return nil, ErrInvalid
	}
	d := &AuthDevelopment{b: b, names: authObservedNames, hashes: hashes, ObserverPath: paths[0], ObservedPath: paths[1], ProductPath: paths[3]}
	if d.Verify() != nil {
		d.Close()
		return nil, ErrInvalid
	}
	return d, nil
}
func (d *AuthDevelopment) Close() {
	if d != nil && d.b != nil {
		d.b.close()
		d.b = nil
	}
}
func (d *AuthDevelopment) Verify() error {
	if d == nil || d.b == nil {
		return ErrInvalid
	}
	for i, name := range d.names {
		f := d.b.files[name]
		limit := int64(MaxExecutable)
		if name == "build-inputs.json" {
			limit = MaxMetadata
		}
		if f == nil {
			return ErrInvalid
		}
		info, err := f.Stat()
		if err != nil || info.Size() < 1 || info.Size() > limit {
			return ErrInvalid
		}
		if _, err = f.Seek(0, io.SeekStart); err != nil {
			return ErrInvalid
		}
		h := sha256.New()
		n, err := io.Copy(h, io.LimitReader(f, limit+1))
		if err != nil || n != info.Size() || hex.EncodeToString(h.Sum(nil)) != d.hashes[i] {
			return ErrInvalid
		}
	}
	return nil
}
