// Package casacquire implements the separate, maintainer-only anonymous probe.
// No product SDK, profiles, credentials or session state are used.
package casacquire

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
	"github.com/UnbalancedCat/ipgw-meta/internal/ctlcheck"
	"golang.org/x/net/dns/dnsmessage"
)

type wire struct {
	memberSession     *ctlcheck.MemberSession
	memberResult      *ctlcheck.Membership
	local             netip.Addr
	dns               netip.AddrPort
	counts            cascollect.Attempts
	roots             *x509.CertPool
	denied            map[[32]byte]bool
	binding           *pathBinding
	loadIntermediates intermediateLoader
	rootCapture       *rootSnapshot
	chainRelation     *ChainRelation
	chainTopology     *ChainTopology
	sourceSnapshot    *sourceSnapshot
	sourceDifference  *SourceDifference
	// Only loopback tests set these unexported fields. No CLI/env configuration.
	testTarget        string
	testTLS           *tls.Config
	testTimeout       time.Duration
	testMemberCompare func(context.Context, [][]byte) *ctlcheck.Membership
}

func usableIP(a netip.Addr) bool {
	return a.Is4() && !a.IsLoopback() && !a.IsUnspecified() && !a.IsMulticast() && a != netip.MustParseAddr("255.255.255.255")
}

func dnsAnswer(packet []byte, id uint16, name dnsmessage.Name) (netip.Addr, bool) {
	if len(packet) < 12 || len(packet) > 4096 || binary.BigEndian.Uint16(packet[4:6]) != 1 || binary.BigEndian.Uint16(packet[6:8]) > 32 {
		return netip.Addr{}, false
	}
	var parser dnsmessage.Parser
	h, err := parser.Start(packet)
	if err != nil || !h.Response || h.ID != id || h.OpCode != 0 || h.Truncated || h.RCode != dnsmessage.RCodeSuccess {
		return netip.Addr{}, false
	}
	q, err := parser.Question()
	if err != nil || q.Type != dnsmessage.TypeA || q.Class != dnsmessage.ClassINET || !strings.EqualFold(q.Name.String(), name.String()) {
		return netip.Addr{}, false
	}
	if _, err = parser.Question(); err != dnsmessage.ErrSectionDone {
		return netip.Addr{}, false
	}
	// Do not dispatch to parsers for unrelated RDATA (including SVCB/HTTPS).
	// Check every record's framing, but interpret only the A/CNAME answer subset.
	answers := []dnsmessage.Resource{}
	for {
		header, err := parser.AnswerHeader()
		if err == dnsmessage.ErrSectionDone {
			break
		}
		if err != nil {
			return netip.Addr{}, false
		}
		next := parser
		if next.SkipAnswer() != nil {
			return netip.Addr{}, false
		}
		switch header.Type {
		case dnsmessage.TypeA:
			if header.Length != 4 {
				return netip.Addr{}, false
			}
			body, err := parser.AResource()
			if err != nil {
				return netip.Addr{}, false
			}
			answers = append(answers, dnsmessage.Resource{Header: header, Body: &body})
		case dnsmessage.TypeCNAME:
			if header.Length == 0 {
				return netip.Addr{}, false
			}
			body, err := parser.CNAMEResource()
			if err != nil {
				return netip.Addr{}, false
			}
			answers = append(answers, dnsmessage.Resource{Header: header, Body: &body})
		}
		parser = next
	}
	for _, section := range []struct {
		header func() (dnsmessage.ResourceHeader, error)
		skip   func() error
	}{{parser.AuthorityHeader, parser.SkipAuthority}, {parser.AdditionalHeader, parser.SkipAdditional}} {
		for {
			_, err := section.header()
			if err == dnsmessage.ErrSectionDone {
				break
			}
			if err != nil || section.skip() != nil {
				return netip.Addr{}, false
			}
		}
	}
	if len(answers) == 0 {
		return netip.Addr{}, false
	}
	current := strings.ToLower(name.String())
	seen := map[string]bool{}
	for depth := 0; depth < 8; depth++ {
		if seen[current] {
			return netip.Addr{}, false
		}
		seen[current] = true
		var address netip.Addr
		next := ""
		for _, answer := range answers {
			if answer.Header.Class != dnsmessage.ClassINET || strings.ToLower(answer.Header.Name.String()) != current {
				continue
			}
			switch b := answer.Body.(type) {
			case *dnsmessage.AResource:
				if !address.IsValid() {
					address = netip.AddrFrom4(b.A)
				}
			case *dnsmessage.CNAMEResource:
				n := strings.ToLower(b.CNAME.String())
				if next != "" && next != n {
					return netip.Addr{}, false
				}
				next = n
			}
		}
		if address.IsValid() {
			return address, next == ""
		}
		if next == "" {
			break
		}
		current = next
	}
	return netip.Addr{}, false
}

func (w *wire) resolve(ctx context.Context, host string) (netip.Addr, string) {
	if w.counts.DNS >= cascollect.MaxAttempts {
		return netip.Addr{}, "budget_exceeded"
	}
	name, err := dnsmessage.NewName(host + ".")
	if err != nil {
		return netip.Addr{}, "dns_failed"
	}
	var random [2]byte
	if _, err = rand.Read(random[:]); err != nil {
		return netip.Addr{}, "dns_failed"
	}
	id := binary.BigEndian.Uint16(random[:])
	m := dnsmessage.Message{Header: dnsmessage.Header{ID: id, RecursionDesired: true}, Questions: []dnsmessage.Question{{Name: name, Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET}}}
	packet, err := m.Pack()
	if err != nil {
		return netip.Addr{}, "dns_failed"
	}
	control, reason := w.control(w.dns.Addr())
	if reason != "" {
		return netip.Addr{}, reason
	}
	dialer := net.Dialer{LocalAddr: &net.UDPAddr{IP: net.IP(w.local.AsSlice())}, Control: control}
	conn, err := dialer.DialContext(ctx, "udp4", w.dns.String())
	if err != nil {
		if errors.Is(err, errInterfaceBinding) {
			return netip.Addr{}, "interface_binding_failed"
		}
		return netip.Addr{}, "dns_failed"
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()
	deadline, _ := ctx.Deadline()
	if conn.SetDeadline(deadline) != nil {
		return netip.Addr{}, "dns_failed"
	}
	if !w.pathStillValid(w.dns.Addr()) {
		return netip.Addr{}, "route_rejected"
	}
	w.counts.DNS++ // Exactly one application UDP send attempt, including failures.
	if _, err = conn.Write(packet); err != nil {
		return netip.Addr{}, "dns_failed"
	}
	buf := make([]byte, 4097)
	n, err := conn.Read(buf)
	if err != nil || n > 4096 {
		return netip.Addr{}, "dns_failed"
	}
	a, ok := dnsAnswer(buf[:n], id, name)
	if !ok || (!usableIP(a) && !(w.testTarget != "" && a.IsLoopback())) {
		return netip.Addr{}, "dns_failed"
	}
	return a, ""
}

func (w *wire) fetch(parent context.Context, req cascollect.Request, limit int) (cascollect.Exchange, string) {
	w.chainRelation = nil
	w.chainTopology = nil
	w.sourceDifference = nil
	w.memberResult = nil
	e := cascollect.Exchange{Target: req.Target}
	u, err := url.Parse(req.Target)
	if err != nil || req.Method != "GET" || u.User != nil || u.Fragment != "" || u.Port() != "" || !(u.Host == "ipgw.neu.edu.cn" && (u.Scheme == "http" || u.Scheme == "https") || u.Host == "pass.neu.edu.cn" && u.Scheme == "https") {
		return e, "url_rejected"
	}
	if w.counts.Requests >= cascollect.MaxAttempts || limit < 0 {
		return e, "budget_exceeded"
	}
	if u.Scheme == "https" && w.testTLS == nil && w.roots == nil {
		return e, "transport_failed"
	}
	timeout := 10 * time.Second
	if w.testTimeout != 0 {
		timeout = w.testTimeout
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	if ctx.Err() != nil {
		return e, "canceled"
	}
	w.counts.Requests++
	a, reason := w.resolve(ctx, u.Hostname())
	if reason != "" {
		return e, reason
	}
	if w.counts.Connects >= cascollect.MaxAttempts {
		return e, "budget_exceeded"
	}
	port := "80"
	if u.Scheme == "https" {
		port = "443"
	}
	target := net.JoinHostPort(a.String(), port)
	if w.testTarget != "" {
		host, _, err := net.SplitHostPort(w.testTarget)
		ip, parseErr := netip.ParseAddr(host)
		if err != nil || parseErr != nil || !ip.IsLoopback() {
			return e, "request_failed"
		}
		target = w.testTarget
	}
	control, reason := w.control(a)
	if reason != "" {
		return e, reason
	}
	dialer := net.Dialer{LocalAddr: &net.TCPAddr{IP: net.IP(w.local.AsSlice())}, Control: control}
	w.counts.Connects++
	conn, err := dialer.DialContext(ctx, "tcp4", target)
	if err != nil {
		if errors.Is(err, errInterfaceBinding) {
			return e, "interface_binding_failed"
		}
		return e, "connect_failed"
	}
	defer conn.Close()
	if !w.pathStillValid(a) {
		return e, "route_rejected"
	}
	rawConn := conn
	stop := context.AfterFunc(ctx, func() { rawConn.Close() })
	defer stop()
	deadline, _ := ctx.Deadline()
	if conn.SetDeadline(deadline) != nil {
		return e, "request_deadline_setup_failed"
	}
	if u.Scheme == "https" {
		config := &tls.Config{RootCAs: w.roots, ServerName: u.Hostname(), MinVersion: tls.VersionTLS12, NextProtos: []string{"http/1.1"}, VerifyConnection: func(s tls.ConnectionState) error {
			for _, chain := range s.VerifiedChains {
				for _, c := range chain {
					if w.denied[sha256.Sum256(c.Raw)] {
						return errCertificateDenied
					}
				}
			}
			return nil
		}}
		if w.testTLS != nil {
			config = w.testTLS.Clone()
		}
		// In pinned Go 1.25 the last clock call before a certificate verification
		// failure is VerifyOptions.CurrentTime. Reuse that exact instant, not a
		// later wall clock value, for the diagnostic-only comparison.
		clock := config.Time
		if clock == nil {
			clock = time.Now
		}
		var verifiedAt time.Time
		config.Time = func() time.Time { verifiedAt = clock(); return verifiedAt }
		secure := tls.Client(conn, config)
		if err := secure.HandshakeContext(ctx); err != nil {
			reason := tlsFailure(ctx, err)
			if reason == "tls_unknown_authority" && w.loadIntermediates != nil {
				// Close the raw socket before reading CA or comparing chains.
				// No handshake resumption or HTTP write exists on this path.
				rawConn.Close()
				if u.Host == "pass.neu.edu.cn" {
					w.compareMembership(ctx, err)
				}
				if w.rootCapture != nil {
					observation := &relationObservation{roots: w.rootCapture}
					reason := compareChain(ctx, err, config.RootCAs, w.denied, config.ServerName, verifiedAt, w.loadIntermediates, observation)
					w.chainRelation = observation.result
					w.chainTopology = observation.topology
					if w.sourceSnapshot != nil && observation.graph != nil && ctx.Err() == nil {
						w.sourceDifference = compareSources(ctx, w.sourceSnapshot, observation.graph, defaultDifferenceLimits)
					}
					if ctx.Err() != nil {
						if w.memberResult != nil {
							reason := "canceled"
							if errors.Is(ctx.Err(), context.DeadlineExceeded) {
								reason = "deadline"
							}
							w.memberResult = ctlcheck.MemberUnavailable("comparison", "unavailable", reason)
						}
						w.chainRelation = nil
						w.chainTopology = nil
						w.sourceDifference = nil
						return e, probeContext(ctx)
					}
					return e, reason
				}
				return e, compareChain(ctx, err, config.RootCAs, w.denied, config.ServerName, verifiedAt, w.loadIntermediates)
			}
			return e, reason
		}
		if w.sourceSnapshot != nil {
			return e, "source_baseline_tls_verified"
		}
		conn = secure
	}
	httpReq, err := http.NewRequestWithContext(ctx, "GET", req.Target, nil)
	if err != nil {
		return e, "request_build_failed"
	}
	httpReq.Close = true
	httpReq.Header.Set("Accept-Encoding", "identity")
	httpReq.Header.Set("User-Agent", "ipgw-meta-cas-acquire-dev/1")
	if err := httpReq.Write(conn); err != nil {
		return e, ioFailure("request_write", ctx, err)
	}
	return readResponse(ctx, bufio.NewReader(conn), e, limit)
}

func readResponse(ctx context.Context, reader *bufio.Reader, e cascollect.Exchange, limit int) (cascollect.Exchange, string) {
	remaining, count := cascollect.MaxHeaderBytes, 0
	line := func(stage, rejection string) (string, string) {
		var b []byte
		for {
			part, err := reader.ReadSlice('\n')
			if len(part) > remaining {
				return "", rejection
			}
			remaining -= len(part)
			b = append(b, part...)
			if err == bufio.ErrBufferFull {
				continue
			}
			if err != nil {
				return "", ioFailure(stage, ctx, err)
			}
			if !bytes.HasSuffix(b, []byte("\r\n")) {
				return "", rejection
			}
			return string(b[:len(b)-2]), ""
		}
	}
	status, reason := line("response_status", "response_rejected")
	if reason != "" {
		return e, reason
	}
	if len(status) < 13 || !strings.HasPrefix(status, "HTTP/1.1 ") || status[12] != ' ' {
		return e, "response_rejected"
	}
	n, err := strconv.Atoi(status[9:12])
	if err != nil || n < 200 || n > 599 {
		return e, "response_rejected"
	}
	e.Status = n
	seen := map[string]bool{}
	for {
		s, reason := line("response_headers", "resource_limit")
		if reason != "" {
			return e, reason
		}
		if s == "" {
			break
		}
		count++
		if count > 64 {
			return e, "resource_limit"
		}
		name, value, ok := strings.Cut(s, ":")
		if !ok || name == "" {
			return e, "response_rejected"
		}
		for _, c := range name {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune("!#$%&'*+-.^_`|~", c)) {
				return e, "response_rejected"
			}
		}
		for _, c := range value {
			if c < 32 && c != '\t' || c == 127 {
				return e, "response_rejected"
			}
		}
		name = strings.ToLower(name)
		value = strings.Trim(value, " \t")
		if name == "transfer-encoding" {
			return e, "response_rejected"
		}
		switch name {
		case "content-type", "content-length", "content-encoding", "location":
			if seen[name] {
				return e, "response_rejected"
			}
			seen[name] = true
			if name == "content-encoding" && value != "identity" {
				return e, "response_rejected"
			}
			if name == "content-length" {
				n, err := strconv.Atoi(value)
				if err != nil || n < 0 || strconv.Itoa(n) != value || n > limit {
					return e, "resource_limit"
				}
			}
			e.Headers = append(e.Headers, cascollect.Header{Name: name, Value: value})
		}
	}
	body, err := io.ReadAll(io.LimitReader(reader, int64(limit)+1))
	if len(body) > limit {
		return e, "resource_limit"
	}
	if err != nil {
		return e, ioFailure("response_body", ctx, err)
	}
	for _, h := range e.Headers {
		if h.Name == "content-length" && h.Value != strconv.Itoa(len(body)) {
			return e, "response_rejected"
		}
	}
	e.Body = body
	return e, ""
}
