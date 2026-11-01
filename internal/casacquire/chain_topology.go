package casacquire

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"errors"
	"time"
)

// Each flag is existential over members of the named terminal category.
// Different true flags need not describe the same certificate or component.
type TerminalFacts struct {
	Present                  bool `json:"present"`
	PeerOnlyMember           bool `json:"peer_only_member"`
	CacheOnlyMember          bool `json:"cache_only_member"`
	SharedMember             bool `json:"shared_member"`
	RootSameSubjectSPKI      bool `json:"root_same_subject_spki"`
	RootDifferentSubjectSPKI bool `json:"root_different_subject_spki"`
}

type TopologyFacts struct {
	NonterminalCycle bool          `json:"nonterminal_cycle"`
	SelfLoop         TerminalFacts `json:"self_loop"`
	MultiNodeCycle   TerminalFacts `json:"multi_node_cycle"`
	AcyclicDeadEnd   TerminalFacts `json:"acyclic_dead_end"`
}

type ChainTopology struct {
	Status string         `json:"status"`
	Reason string         `json:"reason"`
	Facts  *TopologyFacts `json:"facts"`
}

func topologyObservationReason(reason string) bool {
	return reason == "tls_chain_topology_observed" || reason == "tls_chain_topology_denied" || reason == "tls_chain_topology_unavailable"
}

func topologyFailure(reason string) bool {
	switch reason {
	case "input_invalid", "source_invalid", "node_limit", "edge_limit", "root_limit", "byte_limit", "comparison_limit", "canceled", "deadline":
		return true
	}
	return false
}

func validTopology(reason string, relation *ChainRelation, topology *ChainTopology) bool {
	if relation == nil || relation.Status != "complete" || relation.Facts == nil || !relation.Facts.CycleObserved || relation.Facts.RootPathObserved || topology == nil {
		return false
	}
	base := "tls_chain_relation_observed"
	if relation.Facts.DeniedObserved {
		base = "tls_chain_relation_denied"
	}
	if !validRelation(base, relation) {
		return false
	}
	if topology.Status == "unavailable" {
		return reason == "tls_chain_topology_unavailable" && topology.Facts == nil && topologyFailure(topology.Reason)
	}
	if topology.Status != "complete" || topology.Reason != "none" || topology.Facts == nil {
		return false
	}
	if relation.Facts.DeniedObserved != (reason == "tls_chain_topology_denied") || (reason != "tls_chain_topology_denied" && reason != "tls_chain_topology_observed") {
		return false
	}
	f := topology.Facts
	for _, group := range []TerminalFacts{f.SelfLoop, f.MultiNodeCycle, f.AcyclicDeadEnd} {
		if !group.Present && group != (TerminalFacts{}) {
			return false
		}
		if group.Present && !group.PeerOnlyMember && !group.CacheOnlyMember && !group.SharedMember {
			return false
		}
	}
	if !f.SelfLoop.Present && !f.MultiNodeCycle.Present && !f.AcyclicDeadEnd.Present {
		return false
	}
	if !f.SelfLoop.Present && !f.MultiNodeCycle.Present && !f.NonterminalCycle {
		return false
	}
	rf := relation.Facts
	if f.AcyclicDeadEnd.Present && !rf.UpstreamNoParent && !rf.CAConstraintRejected && !rf.AlgorithmRejected && !rf.SignatureCheckFailed {
		return false
	}
	return true
}

type relationGraph struct {
	nodes   [][32]byte
	certs   map[[32]byte]*x509.Certificate
	edges   map[[32]byte][][32]byte
	sources map[[32]byte]uint8 // peer=1, CA=2, ROOT=4; union before DER dedup.
	roots   []*x509.Certificate
}

type topologyLimits struct{ nodes, edges, roots, bytes, comparisons int }

var defaultTopologyLimits = topologyLimits{128, 100, 512, 32 << 20, 65536}

// Only inspects the completed D6 graph. It never verifies signatures or trust.
func observeTopology(ctx context.Context, graph relationGraph, limits topologyLimits) *ChainTopology {
	return observeTopologyCaptured(ctx, graph, limits, nil)
}
func observeTopologyCaptured(ctx context.Context, graph relationGraph, limits topologyLimits, terminals map[[32]byte]bool) *ChainTopology {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	r := &ChainTopology{Status: "unavailable"}
	fail := func(reason string) *ChainTopology {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			reason = "deadline"
		} else if ctx.Err() != nil {
			reason = "canceled"
		}
		r.Reason, r.Facts = reason, nil
		return r
	}
	if ctx.Err() != nil {
		return fail("canceled")
	}
	if limits.nodes < 1 || limits.nodes > 128 || limits.edges < 0 || limits.edges > 100 || limits.roots < 0 || limits.roots > 512 || limits.bytes < 0 || limits.bytes > 32<<20 || limits.comparisons < 0 || limits.comparisons > 65536 {
		return fail("input_invalid")
	}
	if len(graph.nodes) == 0 {
		return fail("input_invalid")
	}
	if len(graph.nodes) > limits.nodes {
		return fail("node_limit")
	}
	if len(graph.roots) > limits.roots {
		return fail("root_limit")
	}
	ids := map[[32]byte]int{}
	for i, id := range graph.nodes {
		if ctx.Err() != nil {
			return fail("canceled")
		}
		if _, ok := ids[id]; ok {
			return fail("input_invalid")
		}
		c := graph.certs[id]
		if c == nil || len(c.Raw) == 0 || len(c.RawSubject) == 0 || len(c.RawSubjectPublicKeyInfo) == 0 {
			return fail("input_invalid")
		}
		// At this gate no reachable node can belong to the original root pool.
		if graph.sources[id] < 1 || graph.sources[id] > 3 {
			return fail("source_invalid")
		}
		ids[id] = i
	}
	n := len(graph.nodes)
	adj := make([][]int, n)
	totalEdges := 0
	for id, targets := range graph.edges {
		if ctx.Err() != nil {
			return fail("canceled")
		}
		i, ok := ids[id]
		if !ok {
			return fail("input_invalid")
		}
		seen := map[int]bool{}
		for _, target := range targets {
			j, ok := ids[target]
			if !ok || seen[j] {
				return fail("input_invalid")
			}
			seen[j] = true
			totalEdges++
			if totalEdges > limits.edges {
				return fail("edge_limit")
			}
			adj[i] = append(adj[i], j)
		}
	}
	// Tarjan SCC: at most 128 stack frames and 100 edges, no path enumeration.
	index, low, component := make([]int, n), make([]int, n), make([]int, n)
	onStack := make([]bool, n)
	for i := range index {
		index[i], component[i] = -1, -1
	}
	stack := []int{}
	groups := [][]int{}
	next := 0
	var visit func(int) bool
	visit = func(v int) bool {
		if ctx.Err() != nil {
			return false
		}
		index[v], low[v] = next, next
		next++
		stack = append(stack, v)
		onStack[v] = true
		for _, w := range adj[v] {
			if index[w] == -1 {
				if !visit(w) {
					return false
				}
				if low[w] < low[v] {
					low[v] = low[w]
				}
			} else if onStack[w] && index[w] < low[v] {
				low[v] = index[w]
			}
		}
		if low[v] == index[v] {
			group := []int{}
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				component[w] = len(groups)
				group = append(group, w)
				if w == v {
					break
				}
			}
			groups = append(groups, group)
		}
		return true
	}
	for i := range index {
		if index[i] == -1 && !visit(i) {
			return fail("canceled")
		}
	}
	usedBytes, comparisons := 0, 0
	charge := func(size int) bool {
		if size < 0 || size > limits.bytes-usedBytes {
			return false
		}
		usedBytes += size
		return true
	}
	rootIDs := make([][32]byte, len(graph.roots))
	for i, c := range graph.roots {
		if ctx.Err() != nil {
			return fail("canceled")
		}
		if c == nil || len(c.Raw) == 0 || len(c.RawSubject) == 0 || len(c.RawSubjectPublicKeyInfo) == 0 {
			return fail("input_invalid")
		}
		if !charge(len(c.Raw)) {
			return fail("byte_limit")
		}
		rootIDs[i] = sha256.Sum256(c.Raw)
	}
	f := &TopologyFacts{}
	for groupID, members := range groups {
		if ctx.Err() != nil {
			return fail("canceled")
		}
		terminal, cyclic := true, len(members) > 1
		for _, v := range members {
			for _, w := range adj[v] {
				if component[w] != groupID {
					terminal = false
				}
				if v == w {
					cyclic = true
				}
			}
		}
		if !terminal {
			f.NonterminalCycle = f.NonterminalCycle || cyclic
			continue
		}
		target := &f.AcyclicDeadEnd
		if len(members) > 1 {
			target = &f.MultiNodeCycle
		} else if cyclic {
			target = &f.SelfLoop
		}
		target.Present = true
		for _, v := range members {
			id := graph.nodes[v]
			if terminals != nil {
				terminals[id] = true
			}
			cert := graph.certs[id]
			switch graph.sources[id] {
			case 1:
				target.PeerOnlyMember = true
			case 2:
				target.CacheOnlyMember = true
			case 3:
				target.SharedMember = true
			}
			for i, root := range graph.roots {
				if ctx.Err() != nil {
					return fail("canceled")
				}
				if comparisons >= limits.comparisons {
					return fail("comparison_limit")
				}
				comparisons++
				if id == rootIDs[i] {
					return fail("input_invalid")
				}
				if !charge(len(cert.RawSubjectPublicKeyInfo) + len(root.RawSubjectPublicKeyInfo)) {
					return fail("byte_limit")
				}
				if !bytes.Equal(cert.RawSubjectPublicKeyInfo, root.RawSubjectPublicKeyInfo) {
					continue
				}
				if !charge(len(cert.RawSubject) + len(root.RawSubject)) {
					return fail("byte_limit")
				}
				if bytes.Equal(cert.RawSubject, root.RawSubject) {
					target.RootSameSubjectSPKI = true
				} else {
					target.RootDifferentSubjectSPKI = true
				}
			}
		}
	}
	if ctx.Err() != nil {
		return fail("canceled")
	}
	r.Status, r.Reason, r.Facts = "complete", "none", f
	return r
}
