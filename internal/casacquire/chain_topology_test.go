package casacquire

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/UnbalancedCat/ipgw-meta/internal/cascollect"
)

func TestTopologyGeneratedOriginsAndGate(t *testing.T) {
	c := relationChain(t)
	for _, name := range []string{"peer", "cache", "shared", "denied", "multi-cycle", "root-reached", "no-cycle", "root-other-name", "root-same-name", "unavailable"} {
		t.Run(name, func(t *testing.T) {
			peer := []*x509.Certificate{c.leaf, c.intermediate}
			cached := cacheSnapshot{partial: true}
			roots, snapshot := relationRoots()
			denied := map[[32]byte]bool{}
			want := TerminalFacts{Present: true, CacheOnlyMember: true}
			cached.certs = []*x509.Certificate{c.root}
			switch name {
			case "peer":
				peer = append(peer, c.root)
				cached.certs = nil
				want.CacheOnlyMember = false
				want.PeerOnlyMember = true
			case "shared":
				peer = append(peer, c.root)
				cached.certs = append(cached.certs, c.root)
				want.CacheOnlyMember = false
				want.SharedMember = true
			case "denied":
				denied[sha256.Sum256(c.root.Raw)] = true
			case "multi-cycle":
				a, b := relationTemplate("topology-cycle-a", true), relationTemplate("topology-cycle-b", true)
				ac := relationIssue(t, a, b, c.ik, c.rk)
				bc := relationIssue(t, b, a, c.rk, c.ik)
				peer = []*x509.Certificate{relationIssue(t, relationTemplate("synthetic-leaf", false), ac, c.lk, c.ik), ac}
				cached.certs = []*x509.Certificate{bc}
				want = TerminalFacts{Present: true, PeerOnlyMember: true, CacheOnlyMember: true}
			case "root-reached":
				roots, snapshot = relationRoots(c.root)
			case "no-cycle":
				cached.certs = nil
			case "root-other-name", "root-same-name", "unavailable":
				template := relationTemplate("different-root-name", true)
				if name == "root-same-name" {
					template = relationTemplate("synthetic-root", false)
					want.RootSameSubjectSPKI = true
				} else {
					want.RootDifferentSubjectSPKI = true
				}
				template.SerialNumber = big.NewInt(44)
				alternative := relationIssue(t, template, template, c.rk, c.rk)
				roots, snapshot = relationRoots(alternative)
			}
			if name == "unavailable" {
				old := defaultTopologyLimits
				defaultTopologyLimits.bytes = 0
				defer func() { defaultTopologyLimits = old }()
			}
			var prior *ChainTopology
			for n := 0; n < 2; n++ {
				var topology *ChainTopology
				reason, relation := observeRelations(context.Background(), peer, cached, roots, snapshot, denied, &topology)
				if name == "root-reached" || name == "no-cycle" {
					if topology != nil || !validRelation(reason, relation) {
						t.Fatal("D7 ran outside gate")
					}
					return
				}
				if !validTopology(reason, relation, topology) {
					t.Fatalf("invalid topology result: %s", reason)
				}
				if name == "unavailable" {
					if reason != "tls_chain_topology_unavailable" || topology.Reason != "byte_limit" || topology.Facts != nil || !relation.Facts.CycleObserved {
						t.Fatal("D7 limit lost D6 or leaked partial facts")
					}
					return
				}
				terminal, other := topology.Facts.SelfLoop, topology.Facts.MultiNodeCycle
				if name == "multi-cycle" {
					terminal, other = other, terminal
				}
				if terminal != want || other.Present || topology.Facts.AcyclicDeadEnd.Present || topology.Facts.NonterminalCycle {
					t.Fatalf("wrong terminal facts: %+v", topology.Facts)
				}
				if name == "denied" && reason != "tls_chain_topology_denied" {
					t.Fatal("denial lost")
				}
				if prior != nil && !reflect.DeepEqual(prior, topology) {
					t.Fatal("source order affected result")
				}
				prior = topology
				for i, j := 1, len(peer)-1; i < j; i, j = i+1, j-1 {
					peer[i], peer[j] = peer[j], peer[i]
				}
				for i, j := 0, len(cached.certs)-1; i < j; i, j = i+1, j-1 {
					cached.certs[i], cached.certs[j] = cached.certs[j], cached.certs[i]
				}
			}
		})
	}
}

// A synthetic graph tests SCC logic independently of signature construction.
func topologyGraphFixture(t *testing.T) relationGraph {
	t.Helper()
	g := relationGraph{certs: map[[32]byte]*x509.Certificate{}, edges: map[[32]byte][][32]byte{}, sources: map[[32]byte]uint8{}}
	for i := 0; i < 6; i++ {
		k := relationKey(t)
		template := relationTemplate("synthetic-graph", true)
		template.SerialNumber = big.NewInt(int64(i + 1))
		c := relationIssue(t, template, template, k, k)
		id := sha256.Sum256(c.Raw)
		g.nodes = append(g.nodes, id)
		g.certs[id] = c
		g.sources[id] = uint8(i%3 + 1)
	}
	for i, targets := range [][]int{{1, 3, 5}, {1, 2}, {2}, {4}, {3}, {}} {
		for _, j := range targets {
			g.edges[g.nodes[i]] = append(g.edges[g.nodes[i]], g.nodes[j])
		}
	}
	return g
}

func TestTopologySCCMixedBranchesAndRootRepresentations(t *testing.T) {
	g := topologyGraphFixture(t)
	want := &TopologyFacts{NonterminalCycle: true, SelfLoop: TerminalFacts{Present: true, SharedMember: true}, MultiNodeCycle: TerminalFacts{Present: true, PeerOnlyMember: true, CacheOnlyMember: true}, AcyclicDeadEnd: TerminalFacts{Present: true, SharedMember: true}}
	for n := 0; n < 2; n++ {
		r := observeTopology(context.Background(), g, defaultTopologyLimits)
		if r.Status != "complete" || !reflect.DeepEqual(r.Facts, want) {
			t.Fatalf("mixed SCC: %+v", r.Facts)
		}
		for i, j := 0, len(g.nodes)-1; i < j; i, j = i+1, j-1 {
			g.nodes[i], g.nodes[j] = g.nodes[j], g.nodes[i]
		}
		for id, targets := range g.edges {
			for i, j := 0, len(targets)-1; i < j; i, j = i+1, j-1 {
				targets[i], targets[j] = targets[j], targets[i]
			}
			g.edges[id] = targets
		}
	}
	// Same subject with a different public key is never a root representation match.
	k := relationKey(t)
	template := relationTemplate("synthetic-graph", true)
	g.roots = []*x509.Certificate{relationIssue(t, template, template, k, k)}
	r := observeTopology(context.Background(), g, defaultTopologyLimits)
	if r.Status != "complete" || !reflect.DeepEqual(r.Facts, want) {
		t.Fatal("same name promoted different key")
	}
}

func TestTopologyBudgetsDiscardFacts(t *testing.T) {
	for _, name := range []string{"nodes", "edges", "roots", "bytes", "comparisons", "source", "nil-cert", "duplicate", "outside-edge", "cancel", "deadline"} {
		t.Run(name, func(t *testing.T) {
			g := topologyGraphFixture(t)
			limits := defaultTopologyLimits
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			root := relationChain(t).root
			g.roots = []*x509.Certificate{root}
			want := ""
			switch name {
			case "nodes":
				limits.nodes = 5
				want = "node_limit"
			case "edges":
				limits.edges = 1
				want = "edge_limit"
			case "roots":
				limits.roots = 0
				want = "root_limit"
			case "bytes":
				limits.bytes = 1
				want = "byte_limit"
			case "comparisons":
				limits.comparisons = 1
				want = "comparison_limit"
			case "source":
				g.sources[g.nodes[0]] = 0
				want = "source_invalid"
			case "nil-cert":
				g.certs[g.nodes[0]] = nil
				want = "input_invalid"
			case "duplicate":
				g.nodes = append(g.nodes, g.nodes[0])
				want = "input_invalid"
			case "outside-edge":
				g.edges[g.nodes[0]] = append(g.edges[g.nodes[0]], [32]byte{})
				want = "input_invalid"
			case "cancel":
				cancel()
				want = "canceled"
			case "deadline":
				var end context.CancelFunc
				ctx, end = context.WithDeadline(ctx, time.Now().Add(-time.Second))
				defer end()
				want = "deadline"
			}
			r := observeTopology(ctx, g, limits)
			if r.Status != "unavailable" || r.Reason != want || r.Facts != nil {
				t.Fatalf("budget %s: %+v", name, r)
			}
		})
	}
}

func topologyReportFixture() Report {
	r := blocked("anonymous-once", "tls_chain_topology_observed")
	r.Attempts = cascollect.Attempts{Requests: 3, DNS: 3, Connects: 3}
	r.Operations.ExternalRequests = 3
	r.Operations.DNS = 3
	r.ChainRelation = &ChainRelation{Status: "complete", CacheCoverage: "partial", Reason: "none", Facts: &RelationFacts{CycleObserved: true}}
	r.ChainTopology = &ChainTopology{Status: "complete", Reason: "none", Facts: &TopologyFacts{SelfLoop: TerminalFacts{Present: true, CacheOnlyMember: true}}}
	return r
}

func TestTopologyReportClosedAndLegacyShapes(t *testing.T) {
	r := topologyReportFixture()
	b, err := Encode(r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Decode(b); err != nil {
		t.Fatal("D7 roundtrip")
	}
	for _, mutate := range []func(*Report){
		func(r *Report) { r.Mode = "local-cache-check" },
		func(r *Report) { r.ChainRelation.Facts.RootPathObserved = true },
		func(r *Report) { r.ChainRelation.Facts.CycleObserved = false },
		func(r *Report) { r.ChainRelation.Facts.DeniedObserved = true },
		func(r *Report) { r.ChainTopology.Facts.SelfLoop.Present = false },
		func(r *Report) { r.ChainTopology.Facts.SelfLoop.CacheOnlyMember = false },
		func(r *Report) {
			r.ChainTopology.Facts.AcyclicDeadEnd = TerminalFacts{Present: true, PeerOnlyMember: true}
		},
		func(r *Report) { r.ChainTopology.Status = "unavailable" },
		func(r *Report) { r.Reason = "tls_chain_relation_observed" },
	} {
		var bad Report
		if json.Unmarshal(b, &bad) != nil {
			t.Fatal("fixture")
		}
		mutate(&bad)
		if _, err := Encode(bad); err == nil {
			t.Fatal("contradictory D7 encoded")
		}
		data, _ := json.Marshal(bad)
		if _, err := Decode(data); err == nil {
			t.Fatal("contradictory D7 decoded")
		}
	}
	for _, data := range []string{
		strings.Replace(string(b), `"nonterminal_cycle":false`, `"nonterminal_cycle":false,"nonterminal_cycle":false`, 1),
		strings.Replace(string(b), `"nonterminal_cycle":false,`, "", 1),
		strings.Replace(string(b), `"nonterminal_cycle":false`, `"nonterminal_cycle":null`, 1),
		strings.Replace(string(b), `"nonterminal_cycle":false`, `"private_name":"canary","nonterminal_cycle":false`, 1),
	} {
		if _, err := Decode([]byte(data)); err == nil {
			t.Fatal("open D7 schema")
		}
	}
	r.ChainRelation.Facts.DeniedObserved = true
	r.Reason = "tls_chain_topology_denied"
	if _, err = Encode(r); err != nil {
		t.Fatal("denial rejected")
	}
	r.Reason = "tls_chain_topology_unavailable"
	r.ChainTopology = &ChainTopology{Status: "unavailable", Reason: "byte_limit"}
	b, err = Encode(r)
	if err != nil {
		t.Fatal("unavailable rejected")
	}
	if _, err = Decode(b); err != nil {
		t.Fatal("unavailable decode")
	}
	r = topologyReportFixture()
	r.Reason = "tls_chain_relation_observed"
	r.ChainTopology = nil
	b, err = Encode(r)
	if err != nil || strings.Contains(string(b), "chain_topology") {
		t.Fatal("D6 shape changed")
	}
	if _, err = Decode(b); err != nil {
		t.Fatal("D6 compatibility")
	}
	data := strings.Replace(string(b), `"analysis":null`, `"analysis":null,"chain_topology":null`, 1)
	if _, err = Decode([]byte(data)); err == nil {
		t.Fatal("null D7 extension accepted")
	}
}
