package casacquire

import (
	"crypto/x509"
	"errors"
)

// This only describes the second, offline Verify failure. UnknownAuthorityError
// can hide rejected candidates; its private hints and Error text are not read.
func chainFailure(err error, partial bool) string {
	prefix := "tls_chain_cache_"
	if partial {
		prefix = "tls_chain_partial_cache_"
	}
	var authority x509.UnknownAuthorityError
	var hostname x509.HostnameError
	var invalid x509.CertificateInvalidError
	var critical x509.UnhandledCriticalExtension
	var algorithm x509.InsecureAlgorithmError
	var constraint x509.ConstraintViolationError
	switch {
	case errors.As(err, &authority):
		return prefix + "unknown_authority"
	case errors.As(err, &hostname):
		return prefix + "hostname_mismatch"
	case errors.As(err, &invalid):
		switch invalid.Reason {
		case x509.Expired:
			return prefix + "time_invalid"
		case x509.IncompatibleUsage, x509.CANotAuthorizedForExtKeyUsage:
			return prefix + "usage_invalid"
		case x509.CANotAuthorizedForThisName, x509.TooManyIntermediates, x509.NameConstraintsWithoutSANs, x509.UnconstrainedName, x509.TooManyConstraints:
			return prefix + "constraints_invalid"
		default:
			return prefix + "certificate_invalid"
		}
	case errors.As(err, &critical), errors.As(err, &algorithm), errors.As(err, &constraint):
		return prefix + "certificate_invalid"
	default:
		return prefix + "verify_failed"
	}
}

func chainFailureReason(reason string) bool {
	for _, prefix := range []string{"tls_chain_cache_", "tls_chain_partial_cache_"} {
		for _, suffix := range []string{"unknown_authority", "hostname_mismatch", "time_invalid", "usage_invalid", "constraints_invalid", "certificate_invalid", "verify_failed"} {
			if reason == prefix+suffix {
				return true
			}
		}
	}
	return false
}
