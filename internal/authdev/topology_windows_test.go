//go:build windows

package authdev

import (
	"context"
	"testing"
)

// All network cmdlets are replaced before the fixed script is evaluated. No
// host topology, registry, certificate store, profile or network is accessed.
const fakeTopologyCmdlets = `
# The fixture needs only inbox Utility cmdlets. Avoid discovering unrelated
# modules on the CI image; every networking cmdlet below is synthetic.
$PSModuleAutoLoadingPreference='None'
Import-Module ($PSHOME+'\Modules\Microsoft.PowerShell.Utility\Microsoft.PowerShell.Utility.psd1') -ErrorAction Stop
function Get-NetAdapter {
 [pscustomobject]@{Status='Up';HardwareInterface=$true;InterfaceType=6;ifIndex=12;InterfaceGuid='synthetic-wired'}
 [pscustomobject]@{Status='Up';HardwareInterface=$false;InterfaceType=6;ifIndex=22;InterfaceGuid='synthetic-virtual'}
 if($scenario -eq 'wifi'){[pscustomobject]@{Status='Up';HardwareInterface=$true;InterfaceType=71;ifIndex=13;InterfaceGuid='synthetic-wifi'}}
}
function Get-NetRoute {
 [pscustomobject]@{DestinationPrefix='0.0.0.0/0';InterfaceIndex=12;NextHop='192.0.2.254';RouteMetric=10}
 if($scenario -eq 'split'){[pscustomobject]@{DestinationPrefix='0.0.0.0/1';InterfaceIndex=22;NextHop='192.0.2.253';RouteMetric=1}}
}
function Get-NetIPAddress { [pscustomobject]@{AddressState='Preferred';SkipAsSource=$false;IPAddress='192.0.2.1'} }
function Get-DnsClientServerAddress { [pscustomobject]@{ServerAddresses=@('192.0.2.2')} }
`

func TestNativeTopologyScriptAgainstSyntheticCmdlets(t *testing.T) {
	for _, scenario := range []string{"valid", "wifi", "split"} {
		t.Run(scenario, func(t *testing.T) {
			topology, err := queryTopology(context.Background(), "$scenario='"+scenario+"'\n"+fakeTopologyCmdlets+topologyScript)
			if scenario == "valid" {
				if err != nil || topology.Local.String() != "192.0.2.1" {
					t.Fatal("synthetic valid wired path rejected")
				}
			} else if err == nil {
				t.Fatal("synthetic invalid path accepted")
			}
		})
	}
}
