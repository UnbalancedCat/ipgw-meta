package casacquire

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
)

var errCertificateDenied = errors.New("local certificate denial")

// Classify types only. Error text can contain peer data or authentication URLs.
func tlsFailure(ctx context.Context, err error) string {
	if errors.Is(err, errCertificateDenied) {
		return "tls_certificate_denied"
	}
	var authority x509.UnknownAuthorityError
	var hostname x509.HostnameError
	var invalid x509.CertificateInvalidError
	var verification *tls.CertificateVerificationError
	switch {
	case errors.As(err, &authority):
		return "tls_unknown_authority"
	case errors.As(err, &hostname):
		return "tls_hostname_mismatch"
	case errors.As(err, &invalid):
		if invalid.Reason == x509.Expired {
			return "tls_certificate_time_invalid"
		}
		return "tls_certificate_invalid"
	case errors.As(err, &verification):
		return "tls_certificate_invalid"
	default:
		return ioFailure("tls_handshake", ctx, err)
	}
}

func ioFailure(stage string, ctx context.Context, err error) string {
	if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
		return stage + "_canceled"
	}
	var network net.Error
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.As(err, &network) && network.Timeout() {
		return stage + "_timeout"
	}
	return stage + "_failed"
}

func transportReason(reason string) bool {
	if probeReason(reason) {
		return true
	}
	switch reason {
	case "tls_unknown_authority", "tls_hostname_mismatch", "tls_certificate_time_invalid", "tls_certificate_invalid", "tls_certificate_denied", "request_deadline_setup_failed", "request_build_failed":
		return true
	}
	for _, stage := range []string{"tls_handshake", "request_write", "response_status", "response_headers", "response_body"} {
		for _, suffix := range []string{"_timeout", "_canceled", "_failed"} {
			if reason == stage+suffix {
				return true
			}
		}
	}
	return false
}
