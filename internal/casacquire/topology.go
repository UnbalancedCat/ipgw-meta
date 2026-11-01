package casacquire

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"io"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
	"unicode/utf16"

	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
)

// Fixed read-only command, no remote invocation and no profile/config reads.
const topologyScript = `$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue'
$reason='topology_query_failed'
try {
 $up=@(Get-NetAdapter -IncludeHidden | Where-Object {$_.Status -eq 'Up'})
 $wired=@($up | Where-Object {$_.HardwareInterface -and $_.InterfaceType -eq 6})
 if($wired.Count -ne 1){$reason='wired_interface_ambiguous';throw 'gate'}
 if(@($up | Where-Object {$_.HardwareInterface -and $_.InterfaceType -eq 71}).Count -ne 0){$reason='wireless_active';throw 'gate'}
 $idx=$wired[0].ifIndex
 $routes=@(Get-NetRoute -AddressFamily IPv4 -PolicyStore ActiveStore)
 $def=@($routes | Where-Object {$_.DestinationPrefix -eq '0.0.0.0/0'})
 if($def.Count -ne 1 -or $def[0].InterfaceIndex -ne $idx){$reason='default_route_rejected';throw 'gate'}
 foreach($r in $routes){
  if($r.DestinationPrefix -in @('0.0.0.0/1','128.0.0.0/1')){$reason='split_route_rejected';throw 'gate'}
 }
 $ips=@(Get-NetIPAddress -InterfaceIndex $idx -AddressFamily IPv4 | Where-Object {$_.AddressState -eq 'Preferred' -and !$_.SkipAsSource})
 if($ips.Count -ne 1){$reason='source_address_rejected';throw 'gate'}
 $dns=@((Get-DnsClientServerAddress -InterfaceIndex $idx -AddressFamily IPv4).ServerAddresses)
 if($dns.Count -lt 1){$reason='dns_configuration_rejected';throw 'gate'}
 @{ok=$true;reason='none';local=$ips[0].IPAddress;dns=$dns[0];index=[uint32]$idx} | ConvertTo-Json -Compress
} catch { @{ok=$false;reason=$reason;local='';dns='';index=0} | ConvertTo-Json -Compress }
`

type cappedBuffer struct {
	bytes.Buffer
	limit int
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.Len() {
		return 0, io.ErrShortWrite
	}
	return b.Buffer.Write(p)
}

type topologyData struct {
	OK     bool   `json:"ok"`
	Reason string `json:"reason"`
	Local  string `json:"local"`
	DNS    string `json:"dns"`
	Index  uint32 `json:"index"`
}

func topology(ctx context.Context) (*wire, string) {
	t, reason := queryTopology(ctx, topologyScript)
	if reason != "" {
		return nil, reason
	}
	a, err := netip.ParseAddr(t.Local)
	if err != nil || !usableIP(a) {
		return nil, "source_address_rejected"
	}
	d, err := netip.ParseAddr(t.DNS)
	if err != nil || !usableIP(d) {
		return nil, "dns_configuration_rejected"
	}
	b, ok := newPathBinding(a, t.Index)
	if !ok {
		return nil, "interface_binding_failed"
	}
	if !b.check(d) {
		return nil, "route_rejected"
	}
	return &wire{local: a, dns: netip.AddrPortFrom(d, 53), binding: b}, ""
}

// queryTopology is shared by production and fake-cmdlet tests, including the
// same minimal environment, EncodedCommand and strict stdout/stderr handling.
func queryTopology(ctx context.Context, script string) (topologyData, string) {
	if runtime.GOOS != "windows" {
		return topologyData{}, "topology_query_failed"
	}
	root := os.Getenv("SystemRoot")
	if !filepath.IsAbs(root) {
		return topologyData{}, "topology_query_failed"
	}
	units := utf16.Encode([]rune(script))
	encoded := make([]byte, 2*len(units))
	for i, c := range units {
		binary.LittleEndian.PutUint16(encoded[2*i:], c)
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, filepath.Join(root, "System32", "WindowsPowerShell", "v1.0", "powershell.exe"), "-NoLogo", "-NoProfile", "-NonInteractive", "-EncodedCommand", base64.StdEncoding.EncodeToString(encoded))
	cmd.Env = []string{"SystemRoot=" + root, "WINDIR=" + root}
	stdout := &cappedBuffer{limit: 4096}
	stderr := &cappedBuffer{limit: 4096}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	cmd.Stdin = bytes.NewReader(nil)
	cmd.WaitDelay = time.Second
	err := cmd.Run()
	if ctx.Err() != nil {
		return topologyData{}, "topology_timeout"
	}
	if err != nil {
		return topologyData{}, "topology_query_failed"
	}
	if stderr.Len() != 0 {
		return topologyData{}, "topology_output_rejected"
	}
	var t topologyData
	if casprovenance.DecodeClosedJSON(bytes.TrimPrefix(stdout.Bytes(), []byte{239, 187, 191}), &t, 4096) != nil {
		return topologyData{}, "topology_output_rejected"
	}
	if !t.OK {
		if t.Local != "" || t.DNS != "" || t.Index != 0 || !topologyReason(t.Reason) {
			return topologyData{}, "topology_output_rejected"
		}
		return topologyData{}, t.Reason
	}
	if t.Reason != "none" || t.Index == 0 || t.Index > 0xffffff {
		return topologyData{}, "topology_output_rejected"
	}
	return t, ""
}
