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
$script:gateRejected=$false
try {
 $adapters=@(Get-NetAdapter -IncludeHidden)
 $ipInterfaces=@(Get-NetIPInterface -AddressFamily IPv4)
 if($adapters.Count -gt 256 -or $ipInterfaces.Count -gt 256){$script:gateRejected=$true;throw 'gate'}
 $up=@($adapters | Where-Object {$_.Status -eq 'Up'})
 $wired=@($up | Where-Object {$_.HardwareInterface -and $_.InterfaceType -eq $selectedType})
 if($wired.Count -ne 1 -or @($up | Where-Object {$_.HardwareInterface -and $_.InterfaceType -eq $otherType}).Count -ne 0){$script:gateRejected=$true;throw 'gate'}
 $idx=$wired[0].ifIndex
 $selectedIP=@($ipInterfaces | Where-Object {$_.InterfaceIndex -eq $idx})
 if($selectedIP.Count -ne 1 -or $selectedIP[0].ConnectionState -ne 'Connected'){$script:gateRejected=$true;throw 'gate'}
 $routes=@(Get-NetRoute -AddressFamily IPv4 -PolicyStore ActiveStore)
 if($routes.Count -gt 256){$script:gateRejected=$true;throw 'gate'}
 $def=@($routes | Where-Object {$_.DestinationPrefix -eq '0.0.0.0/0'} | Where-Object {
  $routeIndex=$_.InterfaceIndex
  $adapter=@($adapters | Where-Object {$_.ifIndex -eq $routeIndex})
  $ipInterface=@($ipInterfaces | Where-Object {$_.InterfaceIndex -eq $routeIndex})
  -not ($adapter.Count -eq 1 -and $ipInterface.Count -eq 1 -and $adapter[0].Status -eq 'Disconnected' -and $ipInterface[0].ConnectionState -eq 'Disconnected')
 })
 if($def.Count -ne 1 -or $def[0].InterfaceIndex -ne $idx){$script:gateRejected=$true;throw 'gate'}
 if(@($routes | Where-Object {$_.DestinationPrefix -in @('0.0.0.0/1','128.0.0.0/1')}).Count -ne 0){$script:gateRejected=$true;throw 'gate'}
 $ips=@(Get-NetIPAddress -InterfaceIndex $idx -AddressFamily IPv4 | Where-Object {$_.AddressState -eq 'Preferred' -and !$_.SkipAsSource})
 if($ips.Count -ne 1){$script:gateRejected=$true;throw 'gate'}
 $dns=@((Get-DnsClientServerAddress -InterfaceIndex $idx -AddressFamily IPv4).ServerAddresses | Sort-Object)
 if($dns.Count -lt 1 -or $dns.Count -gt 16){$script:gateRejected=$true;throw 'gate'}
 $routeState=@($routes | ForEach-Object { '{0}|{1}|{2}|{3}' -f $_.DestinationPrefix,$_.InterfaceIndex,$_.NextHop,$_.RouteMetric } | Sort-Object)
 $adapterState=@($adapters | ForEach-Object {
  $adapterIndex=$_.ifIndex
  $states=@($ipInterfaces | Where-Object {$_.InterfaceIndex -eq $adapterIndex} | ForEach-Object { '{0}|{1}' -f $_.ConnectionState,$_.InterfaceMetric } | Sort-Object)
  '{0}|{1}|{2}|{3}' -f $_.ifIndex,$_.InterfaceGuid,$_.Status,($states -join ',')
 } | Sort-Object)
 @{ok=$true;local=$ips[0].IPAddress;index=[uint32]$idx;dns=$dns;routes=$routeState;adapters=$adapterState} | ConvertTo-Json -Compress -Depth 4
} catch { $failure=if($script:gateRejected){'local_rule_rejected'}else{'local_query_failed'}; @{ok=$false;local='';index=0;dns=@();routes=@();adapters=@();failure=$failure} | ConvertTo-Json -Compress }
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
	Index       uint32
	WiFiSSID    WiFiSSID
	network     string
	fingerprint [32]byte
	components  [5][32]byte
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
	if network == "campus_wifi" {
		t.WiFiSSID, err = ReadWiFiSSID(ctx, t.Index)
		if err != nil {
			return Snapshot{}, err
		}
	}
	return t, nil
}

// DiscoverWired selects the unique active physical wired path without requests.
func DiscoverWired(ctx context.Context) (Snapshot, error) {
	t, err := queryTopology(ctx, "$selectedType=6; $otherType=71\n"+topologyScript)
	if err != nil {
		return Snapshot{}, err
	}
	t.network = "campus_wired"
	return t, nil
}

// DiscoverWiFi selects the unique active physical Wi-Fi path without scanning.
func DiscoverWiFi(ctx context.Context) (Snapshot, error) {
	t, err := queryTopology(ctx, "$selectedType=71; $otherType=6\n"+topologyScript)
	if err != nil {
		return Snapshot{}, err
	}
	t.network = "campus_wifi"
	t.WiFiSSID, err = ReadWiFiSSID(ctx, t.Index)
	if err != nil {
		return Snapshot{}, err
	}
	return t, nil
}

func queryTopology(ctx context.Context, script string) (Snapshot, error) {
	s, reason := queryTopologyDetailed(ctx, script)
	if reason != "none" {
		return Snapshot{}, ErrTopology
	}
	return s, nil
}

func queryTopologyDetailed(ctx context.Context, script string) (Snapshot, string) {
	if runtime.GOOS != "windows" || ctx == nil {
		return Snapshot{}, "unsupported"
	}
	root := os.Getenv("SystemRoot")
	if !filepath.IsAbs(root) {
		return Snapshot{}, "unsupported"
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
	err := cmd.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return Snapshot{}, "deadline"
	}
	if ctx.Err() != nil {
		return Snapshot{}, "canceled"
	}
	if out.exceeded || stderr.exceeded {
		return Snapshot{}, "output_limit"
	}
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return Snapshot{}, "exit_nonzero"
		}
		return Snapshot{}, "launch_failed"
	}
	if stderr.Len() != 0 {
		return Snapshot{}, "stderr"
	}
	data := bytes.TrimPrefix(out.Bytes(), []byte{239, 187, 191})
	var failure struct {
		OK       bool     `json:"ok"`
		Local    string   `json:"local"`
		Index    uint32   `json:"index"`
		DNS      []string `json:"dns"`
		Routes   []string `json:"routes"`
		Adapters []string `json:"adapters"`
		Failure  string   `json:"failure"`
	}
	// Failure projections are never accepted as snapshots. Only fixed reasons
	// from the local script may leave this boundary; errors and values stay local.
	if decodeProjection(data, &failure, true) == nil && !failure.OK && failure.Local == "" && failure.Index == 0 && len(failure.DNS) == 0 && len(failure.Routes) == 0 && len(failure.Adapters) == 0 {
		if failure.Failure == "local_rule_rejected" || failure.Failure == "local_query_failed" {
			return Snapshot{}, failure.Failure
		}
		return Snapshot{}, "query_rejected"
	}
	snapshot, parseErr := parseTopology(data)
	if parseErr != nil {
		return Snapshot{}, "structure_rejected"
	}
	return snapshot, "none"
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
	snapshot := Snapshot{Local: ip, Index: t.Index, fingerprint: sha256.Sum256(canonical)}
	for i, value := range []any{t.Local, t.Index, t.DNS, t.Routes, t.Adapters} {
		raw, _ := json.Marshal(value)
		snapshot.components[i] = sha256.Sum256(raw)
	}
	return snapshot, nil
}
func usableIPv4(ip netip.Addr) bool {
	return ip.Is4() && !ip.IsUnspecified() && !ip.IsLoopback() && !ip.IsMulticast() && !ip.IsLinkLocalUnicast()
}
func (t Snapshot) Check(ctx context.Context) bool {
	current, err := Capture(ctx, t.network, t.Local)
	return err == nil && current == t
}

type limitedBuffer struct {
	buffer   bytes.Buffer
	limit    int
	exceeded bool
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.Len() {
		b.exceeded = true
		return 0, io.ErrShortWrite
	}
	return b.buffer.Write(p)
}

func decodeClosed(data []byte, value any) error { return decodeProjection(data, value, false) }

func decodeProjection(data []byte, value any, failure bool) error {
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
	if failure {
		allowed["failure"] = true
	}
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
