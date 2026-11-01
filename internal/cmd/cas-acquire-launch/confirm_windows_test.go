//go:build windows

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/UnbalancedCat/ipgw-meta/internal/casacquire"
	"github.com/UnbalancedCat/ipgw-meta/internal/casbundle"
)

func TestMain(m *testing.M) {
	if len(os.Args) == 2 && os.Args[1] == "--local-cache-check" {
		// Synthetic partial cache report; no real certificate store access.
		fmt.Println(`{"schema_version":1,"tool_kind":"cas-acquire-dev","mode":"local-cache-check","binding":"not_packaged","scope":"explicit_page_scripts","context_constructed":false,"collection":"blocked","reason":"cache_partial","attempts":{"requests":0,"dns":0,"connects":0},"operations":{"external_requests":0,"dns":0,"profile_reads":0,"credential_reads":0,"cas_posts":0,"activation":0,"logout":0},"references":[],"resources":[],"analysis":null}`)
		os.Exit(0)
	}
	if len(os.Args) == 2 && os.Args[1] == "--local-preflight" {
		// Inert local-mode fixture. No topology query, DNS, sockets or prompt.
		fmt.Println(`{"schema_version":1,"tool_kind":"cas-acquire-dev","mode":"local-preflight","binding":"not_packaged","scope":"explicit_page_scripts","context_constructed":false,"collection":"ready","reason":"none","attempts":{"requests":0,"dns":0,"connects":0},"operations":{"external_requests":0,"dns":0,"profile_reads":0,"credential_reads":0,"cas_posts":0,"activation":0,"logout":0},"references":[],"resources":[],"analysis":null}`)
		os.Exit(0)
	}
	if len(os.Args) == 2 && os.Args[1] == "--self-test" {
		b, err := casacquire.Encode(casacquire.SelfTest(context.Background()))
		if err != nil {
			os.Exit(3)
		}
		os.Stdout.Write(b)
		os.Exit(0)
	}
	if len(os.Args) == 2 && os.Args[1] == "--anonymous-once-confirmed" {
		// Inert fixture: no real topology, HTTP or DNS calls.
		if os.WriteFile(filepath.Join("..", "synthetic-live.marker"), []byte("fixture-only"), 0600) != nil {
			os.Exit(3)
		}
		// Counts describe the fixed report fixture, not actual operations.
		synthetic, err := casacquire.Decode([]byte(`{"schema_version":1,"tool_kind":"cas-acquire-dev","mode":"anonymous-once","binding":"not_packaged","scope":"explicit_page_scripts","context_constructed":true,"collection":"blocked","reason":"tls_chain_relation_observed","attempts":{"requests":3,"dns":3,"connects":3},"operations":{"external_requests":3,"dns":3,"profile_reads":0,"credential_reads":0,"cas_posts":0,"activation":0,"logout":0},"references":[],"resources":[],"analysis":null,"chain_relation":{"status":"complete","cache_coverage":"partial","reason":"none","facts":{"root_path_observed":false,"leaf_no_parent":false,"upstream_no_parent":false,"ca_constraint_rejected":false,"algorithm_rejected":false,"signature_check_failed":false,"cycle_observed":true,"denied_observed":false}}}`))
		if err != nil {
			os.Exit(3)
		}
		synthetic.Reason = "tls_chain_topology_observed"
		synthetic.ChainTopology = &casacquire.ChainTopology{Status: "complete", Reason: "none", Facts: &casacquire.TopologyFacts{SelfLoop: casacquire.TerminalFacts{Present: true, CacheOnlyMember: true}}}
		synthetic.SourceDifference = &casacquire.SourceDifference{Phase: "comparison", Status: "complete", Reason: "none"}
		for i, name := range []string{"user_policy_root", "machine_policy_root", "user_policy_ca", "machine_policy_ca", "user_policy_disallowed", "machine_policy_disallowed", "machine_authroot_cache"} {
			s := casacquire.SourceState{Source: name, State: "absent", Reason: "none", Presence: "empty", ParseCoverage: "empty", Facts: &casacquire.SourceFacts{}}
			if i == 4 || i == 5 {
				s.ParseCoverage = "not_applicable"
			}
			if i == 6 {
				s.State, s.Presence, s.ParseCoverage = "complete", "present", "partial"
				s.Facts.TerminalSameSubjectSPKI = true
			}
			synthetic.SourceDifference.Sources = append(synthetic.SourceDifference.Sources, s)
		}
		encoded, err := casacquire.Encode(synthetic)
		if err != nil {
			os.Exit(3)
		}
		os.Stdout.Write(encoded)
		os.Exit(0)
	}
	os.Exit(m.Run())
}
func TestConfirmationControlsSyntheticChild(t *testing.T) {
	for _, tc := range []struct {
		name, input, state string
		terminal, started  bool
	}{
		{"pipe", "RUN-ANONYMOUS-ONCE\n", "confirmation_required", false, false},
		{"eof", "", "canceled", true, false},
		{"wrong", "yes\n", "canceled", true, false},
		{"explicit", "RUN-ANONYMOUS-ONCE\r\n", "completed", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			payload := filepath.Join(dir, "payload")
			if os.Mkdir(payload, 0700) != nil {
				t.Fatal("fixture directory")
			}
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(exe)
			if err != nil {
				t.Fatal(err)
			}
			if os.WriteFile(filepath.Join(payload, "cas-acquire.exe"), data, 0600) != nil || os.WriteFile(filepath.Join(payload, "build-inputs.json"), []byte("fixture-inputs"), 0600) != nil {
				t.Fatal("fixture write")
			}
			oldExe, oldInputs := executableDigest, inputsDigest
			executableDigest, inputsDigest = casbundle.Digest(data), casbundle.Digest([]byte("fixture-inputs"))
			defer func() { executableDigest, inputsDigest = oldExe, oldInputs }()
			if tc.name == "pipe" {
				var prompt strings.Builder
				local := runHistorical([]string{"--preflight"}, strings.NewReader(""), false, &prompt, dir)
				if local.State != "preflight_completed" || local.LiveAttempted || local.LiveStarted || local.Result == nil || local.Result.Mode != "local-preflight" || local.Result.Collection != "ready" || prompt.Len() != 0 {
					t.Fatal("local preflight invoked live behavior")
				}
				local = runHistorical([]string{"--cache-check"}, strings.NewReader("RUN-ANONYMOUS-ONCE\n"), false, &prompt, dir)
				if local.State != "cache_check_completed" || local.LiveAttempted || local.LiveStarted || local.Result == nil || local.Result.Mode != "local-cache-check" || local.Result.Collection != "blocked" || local.Result.Reason != "cache_partial" || prompt.Len() != 0 {
					t.Fatal("local cache check invoked live behavior or lost partial result")
				}
			}
			r := runHistorical([]string{"--run-once"}, strings.NewReader(tc.input), tc.terminal, io.Discard, dir)
			if r.State != tc.state || r.LiveStarted != tc.started || r.LiveAttempted != tc.started {
				t.Fatalf("state %s", r.State)
			}
			if tc.started {
				if r.Result == nil || r.Result.ChainRelation == nil || r.Result.ChainRelation.Facts == nil || !r.Result.ChainRelation.Facts.CycleObserved || r.Result.ChainTopology == nil || r.Result.ChainTopology.Facts == nil || !r.Result.ChainTopology.Facts.SelfLoop.CacheOnlyMember {
					t.Fatal("launcher lost synthetic relation observation")
				}
				encoded, err := json.Marshal(r)
				if err != nil {
					t.Fatal("launcher receipt encoding")
				}
				var envelope struct {
					Result json.RawMessage `json:"result"`
				}
				if json.Unmarshal(encoded, &envelope) != nil {
					t.Fatal("launcher receipt shape")
				}
				decoded, err := casacquire.Decode(envelope.Result)
				if err != nil || decoded.SourceDifference == nil || decoded.SourceDifference.Status != "complete" || !decoded.SourceDifference.Sources[6].Facts.TerminalSameSubjectSPKI {
					t.Fatal("launcher lost D9 source comparison")
				}
				if err != nil || decoded.ChainRelation == nil || !decoded.ChainRelation.Facts.CycleObserved || decoded.ChainTopology == nil || decoded.ChainTopology.Facts == nil || !decoded.ChainTopology.Facts.SelfLoop.CacheOnlyMember {
					t.Fatal("launcher receipt lost closed relation schema")
				}
			}
			_, err = os.Stat(filepath.Join(dir, "synthetic-live.marker"))
			if (err == nil) != tc.started {
				t.Fatal("unexpected synthetic live invocation")
			}
		})
	}
}
