package casacquire

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
	"github.com/UnbalancedCat/ipgw-meta/internal/ctlcheck"
)

func acquireWithMembers(ctx context.Context, w *wire, prepare func(context.Context) (*ctlcheck.MemberSession, *ctlcheck.Membership), read sourceReader) Report {
	session, failure := prepare(ctx)
	if failure != nil || session == nil {
		if failure == nil {
			failure = ctlcheck.MemberUnavailable("gate", "unavailable", "preparation_failed")
		}
		r := blocked("anonymous-once", "ctl_membership_unavailable")
		r.CTLMembership = failure
		return r
	}
	w.memberSession = session
	r := acquireWithSources(ctx, w, read)
	attachMembership(&r, w)
	return r
}

func (w *wire) compareMembership(ctx context.Context, failure error) {
	if w.memberSession == nil {
		return
	}
	var verification *tls.CertificateVerificationError
	if !errors.As(failure, &verification) || verification == nil {
		w.memberResult = ctlcheck.MemberUnavailable("comparison", "unavailable", "peer_unsupported")
		return
	}
	peer := verification.UnverifiedCertificates
	if len(peer) > maxPeerCerts || !boundedCertificates(peer, maxPeerCerts, maxPeerBytes) {
		w.memberResult = ctlcheck.MemberUnavailable("comparison", "unavailable", "peer_limit")
		return
	}
	raw := make([][]byte, len(peer))
	for i, c := range peer {
		raw[i] = c.Raw
	}
	if w.testMemberCompare != nil {
		w.memberResult = w.testMemberCompare(ctx, raw)
	} else {
		w.memberResult = w.memberSession.Compare(ctx, raw)
	}
}

func attachMembership(r *Report, w *wire) {
	if w.memberSession == nil {
		return
	}
	r.CTLMembership = w.memberResult
	if r.CTLMembership == nil {
		r.CTLMembership = ctlcheck.MemberUnavailable("comparison", "not_run", "baseline_not_observed")
	}
}

func validMemberExtension(r Report) bool {
	m := r.CTLMembership
	if !ctlcheck.ValidMembership(m) || r.Mode != "anonymous-once" || r.Collection != "blocked" || r.Analysis != nil || len(r.References)+len(r.Resources) != 0 {
		return false
	}
	if m.Phase == "gate" {
		return m.State == "unavailable" && r.Reason == "ctl_membership_unavailable" && r.Attempts == (cascollect.Attempts{}) && !r.ContextConstructed && r.SourceDifference == nil && r.ChainRelation == nil && r.ChainTopology == nil
	}
	if r.SourceDifference == nil {
		return false
	}
	if m.State == "not_run" {
		return !r.ContextConstructed || r.ChainRelation == nil && r.ChainTopology == nil
	}
	return r.ContextConstructed && r.Attempts.Requests > 0 && r.Attempts.Requests == r.Attempts.DNS && r.Attempts.DNS == r.Attempts.Connects && (topologyObservationReason(r.Reason) || relationReason(r.Reason) || probeReason(r.Reason) || r.Reason == "tls_unknown_authority")
}
