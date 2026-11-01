//go:build windows

package casacquire

import (
	"context"
	"strings"
	"testing"
	"time"
)

// Fixed stub functions shadow every topology cmdlet. These tests cannot query
// actual interfaces or routes and do not create any network socket.
func TestNativeTopologyPolicyWithFakeCmdlets(t *testing.T) {
	fixture := `
# Load only the inbox module used for JSON/progress in this synthetic fixture.
# No image-wide module discovery or real network cmdlet loading is required.
$PSModuleAutoLoadingPreference='None'
Import-Module ($PSHOME+'\Modules\Microsoft.PowerShell.Utility\Microsoft.PowerShell.Utility.psd1') -ErrorAction Stop
function Get-NetAdapter { Write-Progress -Activity 'synthetic module progress' -Status 'loading'; [pscustomobject]@{Status='Up';HardwareInterface=$true;InterfaceType=6;ifIndex=5} }
function Get-NetRoute { [pscustomobject]@{DestinationPrefix='0.0.0.0/0';InterfaceIndex=5} }
function Get-NetIPAddress { [pscustomobject]@{AddressState='Preferred';SkipAsSource=$false;IPAddress='192.0.2.2'} }
function Get-DnsClientServerAddress { [pscustomobject]@{ServerAddresses=@('192.0.2.53')} }
`
	for _, tc := range []struct {
		name, from, to string
		pass           bool
	}{
		{"wired", "", "", true},
		{"wireless", "InterfaceType=6", "InterfaceType=71", false},
		{"virtual", "HardwareInterface=$true", "HardwareInterface=$false", false},
		{"wrong-default", "DestinationPrefix='0.0.0.0/0';InterfaceIndex=5", "DestinationPrefix='0.0.0.0/0';InterfaceIndex=8", false},
		{"half-route", "DestinationPrefix='0.0.0.0/0'", "DestinationPrefix='0.0.0.0/1'", false},
		{"unpreferred", "AddressState='Preferred'", "AddressState='Tentative'", false},
		{"no-dns", "@('192.0.2.53')", "@()", false},
		{"virtual-retained", "InterfaceType=6;ifIndex=5} }", "InterfaceType=6;ifIndex=5}; [pscustomobject]@{Status='Up';HardwareInterface=$false;InterfaceType=6;ifIndex=8} }", true},
		{"private-route-retained", "DestinationPrefix='0.0.0.0/0';InterfaceIndex=5} }", "DestinationPrefix='0.0.0.0/0';InterfaceIndex=5}; [pscustomobject]@{DestinationPrefix='192.0.2.0/24';InterfaceIndex=8} }", true},
		{"two-wired", "InterfaceType=6;ifIndex=5} }", "InterfaceType=6;ifIndex=5}; [pscustomobject]@{Status='Up';HardwareInterface=$true;InterfaceType=6;ifIndex=8} }", false},
		{"extra-wireless", "InterfaceType=6;ifIndex=5} }", "InterfaceType=6;ifIndex=5}; [pscustomobject]@{Status='Up';HardwareInterface=$true;InterfaceType=71;ifIndex=8} }", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			script := fixture
			if tc.from != "" {
				script = strings.ReplaceAll(script, tc.from, tc.to)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			result, reason := queryTopology(ctx, script+topologyScript)
			if (reason == "" && result.OK) != tc.pass {
				t.Fatalf("topology policy mismatch: %s", reason)
			}
		})
	}
	// Reproduce the original bug using the identical EncodedCommand channel.
	withoutFix := strings.Replace(topologyScript, "$ProgressPreference='SilentlyContinue'", "", 1)
	if _, reason := queryTopology(context.Background(), fixture+withoutFix); reason != "topology_output_rejected" {
		t.Fatal("progress regression fixture did not reproduce old rejection")
	}
	withError := strings.Replace(fixture, "Write-Progress -Activity 'synthetic module progress' -Status 'loading'", "Write-Error -Message 'synthetic failure' -ErrorAction Continue", 1)
	if _, reason := queryTopology(context.Background(), withError+topologyScript); reason != "topology_output_rejected" {
		t.Fatal("actual stderr was ignored")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, reason := queryTopology(ctx, fixture+topologyScript); reason != "topology_timeout" {
		t.Fatal("cancellation ignored")
	}
}
