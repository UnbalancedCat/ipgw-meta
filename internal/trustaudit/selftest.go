package trustaudit

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"
)

func syntheticDER() ([]byte, error) {
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	t := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "synthetic-audit"}, NotBefore: time.Unix(1, 0), NotAfter: time.Unix(2, 0), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
	return x509.CreateCertificate(rand.Reader, t, t, &k.PublicKey, k)
}
func SelfTest(ctx context.Context) Report {
	der, err := syntheticDER()
	read := func(ctx context.Context, name string, location int, a *accumulator) (string, string) {
		if err != nil {
			return "unavailable", "read_failed"
		}
		if location == 1 {
			return "absent", ""
		}
		if reason := a.add(ctx, der); reason != "" {
			return "unavailable", reason
		}
		if name == "CA" || name == "Disallowed" {
			if reason := a.add(ctx, []byte("synthetic-invalid-DER")); reason != "" {
				return "unavailable", reason
			}
		}
		return "complete", ""
	}
	return audit(ctx, "self-test", read)
}
