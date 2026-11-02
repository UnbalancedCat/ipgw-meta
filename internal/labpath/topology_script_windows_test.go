//go:build windows

package labpath

import (
	"context"
	"testing"
)

// These functions replace every queried cmdlet before the actual production
// script runs. No host adapters, routes, DNS, or campus requests are queried.
const syntheticTopology = `
# Load the sole inbox module required by this fixture through its fixed path.
# Every network cmdlet is replaced below; image-wide module discovery must not
# turn a synthetic policy rejection into a query timeout.
$PSModuleAutoLoadingPreference='None'
Import-Module ($PSHOME+'\Modules\Microsoft.PowerShell.Utility\Microsoft.PowerShell.Utility.psd1') -ErrorAction Stop
$selectedType=6; $otherType=71
$script:fixtureAdapters=@(
 [pscustomobject]@{ifIndex=7;Status='Up';HardwareInterface=$true;InterfaceType=6;InterfaceGuid='synthetic-selected'},
 [pscustomobject]@{ifIndex=8;Status='Disconnected';HardwareInterface=$true;InterfaceType=71;InterfaceGuid='synthetic-disconnected'}
)
$script:fixtureInterfaces=@(
 [pscustomobject]@{InterfaceIndex=7;ConnectionState='Connected';InterfaceMetric=10},
 [pscustomobject]@{InterfaceIndex=8;ConnectionState='Disconnected';InterfaceMetric=20}
)
$script:fixtureRoutes=@(
 [pscustomobject]@{DestinationPrefix='0.0.0.0/0';InterfaceIndex=7;NextHop='192.0.2.1';RouteMetric=0},
 [pscustomobject]@{DestinationPrefix='0.0.0.0/0';InterfaceIndex=8;NextHop='198.51.100.1';RouteMetric=0}
)
function Get-NetAdapter { param([switch]$IncludeHidden) $script:fixtureAdapters }
function Get-NetIPInterface { param($AddressFamily) $script:fixtureInterfaces }
function Get-NetRoute { param($AddressFamily,$PolicyStore) $script:fixtureRoutes }
function Get-NetIPAddress { param($InterfaceIndex,$AddressFamily) [pscustomobject]@{IPAddress='192.0.2.11';AddressState='Preferred';SkipAsSource=$false} }
function Get-DnsClientServerAddress { param($InterfaceIndex,$AddressFamily) [pscustomobject]@{ServerAddresses=@('192.0.2.53')} }
`

func TestTopologyScriptDisconnectedDefaultMatrix(t *testing.T) {
	for _, tc := range []struct {
		name, change string
		pass         bool
	}{
		{"disconnected_remnant", "", true},
		{"wifi_selected", "$selectedType=71;$otherType=6;$fixtureAdapters[0].InterfaceType=71;$fixtureAdapters[1].InterfaceType=6", true},
		{"no_remnant", "$fixtureRoutes=@($fixtureRoutes[0])", true},
		{"physical_competitor", "$fixtureAdapters[1].Status='Up';$fixtureInterfaces[1].ConnectionState='Connected'", false},
		{"virtual_competitor", "$fixtureAdapters[1].Status='Up';$fixtureAdapters[1].HardwareInterface=$false;$fixtureInterfaces[1].ConnectionState='Connected'", false},
		{"contradictory_adapter", "$fixtureAdapters[1].Status='Up'", false},
		{"contradictory_ip", "$fixtureInterfaces[1].ConnectionState='Connected'", false},
		{"unknown_ip", "$fixtureInterfaces[1].ConnectionState='Unknown'", false},
		{"unknown_adapter", "$fixtureAdapters[1].Status='Unknown'", false},
		{"missing_ip", "$fixtureInterfaces=@($fixtureInterfaces[0])", false},
		{"duplicate_ip", "$fixtureInterfaces+= $fixtureInterfaces[1]", false},
		{"missing_adapter", "$fixtureAdapters=@($fixtureAdapters[0])", false},
		{"duplicate_adapter", "$fixtureAdapters+= $fixtureAdapters[1]", false},
		{"selected_disconnected", "$fixtureInterfaces[0].ConnectionState='Disconnected'", false},
		{"selected_duplicate", "$fixtureInterfaces+= $fixtureInterfaces[0]", false},
		{"no_active_default", "$fixtureRoutes=@($fixtureRoutes[1])", false},
		{"duplicate_active_default", "$fixtureRoutes+= $fixtureRoutes[0]", false},
		{"split_default_still_rejected", "$fixtureRoutes[1].DestinationPrefix='0.0.0.0/1'", false},
		{"query_error", "function Get-NetIPInterface {throw 'synthetic'}", false},
		{"interface_overflow", "$fixtureInterfaces=@(1..257 | ForEach-Object {$fixtureInterfaces[0]})", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, reason := queryTopologyDetailed(context.Background(), syntheticTopology+tc.change+"\n"+topologyScript)
			want := "local_rule_rejected"
			if tc.pass {
				want = "none"
			} else if tc.name == "query_error" {
				want = "local_query_failed"
			}
			if reason != want {
				t.Fatalf("got %s want %s", reason, want)
			}
		})
	}
}

func TestTopologyScriptRetainsDisconnectedStateInFingerprint(t *testing.T) {
	base, err := queryTopology(context.Background(), syntheticTopology+topologyScript)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []string{
		"$fixtureRoutes[1].NextHop='198.51.100.2'",
		"$fixtureInterfaces[1].InterfaceMetric=21",
		// Remove the stale route so reconnecting a virtual interface still
		// satisfies default uniqueness; the fingerprint must still change.
		"$fixtureRoutes=@($fixtureRoutes[0]);$fixtureAdapters[1].HardwareInterface=$false;$fixtureAdapters[1].Status='Up';$fixtureInterfaces[1].ConnectionState='Connected'",
	} {
		next, err := queryTopology(context.Background(), syntheticTopology+change+"\n"+topologyScript)
		if err != nil || next == base {
			t.Fatal("state change omitted from fingerprint")
		}
	}
}
