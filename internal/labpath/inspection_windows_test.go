//go:build windows

package labpath

import (
	"context"
	"testing"
	"time"
)

func TestDetailedTopologyQueryFailureCategories(t *testing.T) {
	for _, tc := range []struct{ name, script, want string }{
		{"rule", syntheticTopology + "\n$fixtureInterfaces=@()\n" + topologyScript, "local_rule_rejected"},
		{"cmdlet", syntheticTopology + "\nfunction Get-NetRoute {throw 'PRIVATE_CANARY'}\n" + topologyScript, "local_query_failed"},
		{"exit", "exit 7", "exit_nonzero"},
		{"stderr", "[Console]::Error.Write('PRIVATE_CANARY')", "stderr"},
		{"structure", "[Console]::Out.Write('PRIVATE_CANARY')", "structure_rejected"},
		{"limit", "[Console]::Out.Write(('x'*70000))", "output_limit"},
		{"duplicate", "[Console]::Out.Write('{\"ok\":false,\"ok\":false,\"local\":\"\",\"index\":0,\"dns\":[],\"routes\":[],\"adapters\":[],\"failure\":\"local_rule_rejected\"}')", "structure_rejected"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, got := queryTopologyDetailed(context.Background(), tc.script)
			if got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, reason := queryTopologyDetailed(ctx, "exit 0"); reason != "canceled" {
		t.Fatal(reason)
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if _, reason := queryTopologyDetailed(ctx, "Start-Sleep -Seconds 3"); reason != "deadline" {
		t.Fatal(reason)
	}
	t.Setenv("SystemRoot", t.TempDir())
	if _, reason := queryTopologyDetailed(context.Background(), "exit 0"); reason != "launch_failed" {
		t.Fatal(reason)
	}
	if _, reason := queryTopologyDetailed(nil, ""); reason != "unsupported" {
		t.Fatal(reason)
	}
}
