package labpath

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
	"unicode/utf16"
	"unicode/utf8"
)

// Fixed, read-only local cmdlets. This is not a route lock or a DNS interface
// binding. Changes between snapshots remain a documented limitation.
var ErrTopology = errors.New("labpath: topology unavailable or changed")

const topologyScript = `$ErrorActionPreference='Stop'
$ProgressPreference='SilentlyContinue'
try {
 $up=@(Get-NetAdapter -IncludeHidden | Where-Object {$_.Status -eq 'Up'})
 $wired=@($up | Where-Object {$_.HardwareInterface -and $_.InterfaceType -eq $selectedType})
 if($wired.Count -ne 1 -or @($up | Where-Object {$_.HardwareInterface -and $_.InterfaceType -eq $otherType}).Count -ne 0){throw 'gate'}
 $idx=$wired[0].ifIndex
 $routes=@(Get-NetRoute -AddressFamily IPv4 -PolicyStore ActiveStore)
 if($routes.Count -gt 256){throw 'gate'}
 $def=@($routes | Where-Object {$_.DestinationPrefix -eq '0.0.0.0/0'})
 if($def.Count -ne 1 -or $def[0].InterfaceIndex -ne $idx){throw 'gate'}
 if(@($routes | Where-Object {$_.DestinationPrefix -in @('0.0.0.0/1','128.0.0.0/1')}).Count -ne 0){throw 'gate'}
 $ips=@(Get-NetIPAddress -InterfaceIndex $idx -AddressFamily IPv4 | Where-Object {$_.AddressState -eq 'Preferred' -and !$_.SkipAsSource})
 if($ips.Count -ne 1){throw 'gate'}
 $dns=@((Get-DnsClientServerAddress -InterfaceIndex $idx -AddressFamily IPv4).ServerAddresses | Sort-Object)
 if($dns.Count -lt 1 -or $dns.Count -gt 16){throw 'gate'}
 $routeState=@($routes | ForEach-Object { '{0}|{1}|{2}|{3}' -f $_.DestinationPrefix,$_.InterfaceIndex,$_.NextHop,$_.RouteMetric } | Sort-Object)
 $adapterState=@($up | ForEach-Object { '{0}|{1}' -f $_.ifIndex,$_.InterfaceGuid } | Sort-Object)
 @{ok=$true;local=$ips[0].IPAddress;index=[uint32]$idx;dns=$dns;routes=$routeState;adapters=$adapterState} | ConvertTo-Json -Compress -Depth 4
} catch { @{ok=$false;local='';index=0;dns=@();routes=@();adapters=@()} | ConvertTo-Json -Compress }
`

type topologyData struct {
	OK       bool     `json:"ok"`
	Local    string   `json:"local"`
	Index    uint32   `json:"index"`
	DNS      []string `json:"dns"`
	Routes   []string `json:"routes"`
	Adapters []string `json:"adapters"`
}
type Snapshot struct {
	Local       netip.Addr
	network     string
	fingerprint [32]byte
}

func Capture(ctx context.Context, network string, bind netip.Addr) (Snapshot, error) {
	if !usableIPv4(bind) {
		return Snapshot{}, ErrTopology
	}
	prefix := "$selectedType=6; $otherType=71\n"
	if network == "campus_wifi" {
		prefix = "$selectedType=71; $otherType=6\n"
	} else if network != "campus_wired" {
		return Snapshot{}, ErrTopology
	}
	t, err := queryTopology(ctx, prefix+topologyScript)
	if err != nil || t.Local != bind {
		return Snapshot{}, ErrTopology
	}
	t.network = network
	return t, nil
}

func queryTopology(ctx context.Context, script string) (Snapshot, error) {
	if runtime.GOOS != "windows" || ctx == nil {
		return Snapshot{}, ErrTopology
	}
	root := os.Getenv("SystemRoot")
	if !filepath.IsAbs(root) {
		return Snapshot{}, ErrTopology
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
	out, stderr := &limitedBuffer{limit: 64 << 10}, &limitedBuffer{limit: 4096}
	cmd.Stdout, cmd.Stderr = out, stderr
	cmd.WaitDelay = time.Second
	if cmd.Run() != nil || ctx.Err() != nil || stderr.Len() != 0 {
		return Snapshot{}, ErrTopology
	}
	return parseTopology(bytes.TrimPrefix(out.Bytes(), []byte{239, 187, 191}))
}
func parseTopology(data []byte) (Snapshot, error) {
	var t topologyData
	if decodeClosed(data, &t) != nil || !t.OK || t.Index == 0 || t.Index > 0xffffff || len(t.Routes) == 0 || len(t.Routes) > 256 || len(t.DNS) == 0 || len(t.DNS) > 16 || len(t.Adapters) == 0 || len(t.Adapters) > 256 {
		return Snapshot{}, ErrTopology
	}
	ip, err := netip.ParseAddr(t.Local)
	if err != nil || !usableIPv4(ip) {
		return Snapshot{}, ErrTopology
	}
	for _, value := range t.DNS {
		d, err := netip.ParseAddr(value)
		if err != nil || !usableIPv4(d) {
			return Snapshot{}, ErrTopology
		}
	}
	canonical, _ := json.Marshal(t)
	return Snapshot{Local: ip, fingerprint: sha256.Sum256(canonical)}, nil
}
func usableIPv4(ip netip.Addr) bool {
	return ip.Is4() && !ip.IsUnspecified() && !ip.IsLoopback() && !ip.IsMulticast() && !ip.IsLinkLocalUnicast()
}
func (t Snapshot) Check(ctx context.Context) bool {
	current, err := Capture(ctx, t.network, t.Local)
	return err == nil && current == t
}

type limitedBuffer struct {
	buffer bytes.Buffer
	limit  int
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.Len() {
		return 0, io.ErrShortWrite
	}
	return b.buffer.Write(p)
}

func decodeClosed(data []byte, value any) error {
	if len(data) == 0 || len(data) > 64<<10 || !utf8.Valid(data) {
		return ErrTopology
	}
	// This local projection contains one flat object with scalar arrays. Reject
	// duplicate and case-alias keys before decoding Go's case-insensitive fields.
	check := json.NewDecoder(bytes.NewReader(data))
	tok, err := check.Token()
	if err != nil || tok != json.Delim('{') {
		return ErrTopology
	}
	allowed := map[string]bool{"ok": true, "local": true, "index": true, "dns": true, "routes": true, "adapters": true}
	seen := map[string]bool{}
	for check.More() {
		key, err := check.Token()
		if err != nil {
			return ErrTopology
		}
		name, ok := key.(string)
		if !ok || !allowed[name] || seen[name] {
			return ErrTopology
		}
		seen[name] = true
		var raw json.RawMessage
		if check.Decode(&raw) != nil {
			return ErrTopology
		}
	}
	if len(seen) != len(allowed) {
		return ErrTopology
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if d.Decode(value) != nil {
		return ErrTopology
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return ErrTopology
	}
	return nil
}

func (b *limitedBuffer) Bytes() []byte { return b.buffer.Bytes() }
func (b *limitedBuffer) Len() int      { return b.buffer.Len() }
