package livegate

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/netip"
	"reflect"
	"time"
)

// NetworkList and NetworkScan are maintainer-only projections. Raw interface
// identifiers and status data stay in memory; they are never evidence fields.
func (e *ProcessExecutor) NetworkList(ctx context.Context) ([]netip.Addr, error) {
	raw, err := e.auxiliaryJSON(ctx, "list")
	if err != nil {
		return nil, err
	}
	var data struct {
		Interfaces []struct {
			Name  string `json:"name"`
			Index int    `json:"index"`
			IP    string `json:"ip"`
		} `json:"interfaces"`
	}
	if decodeAuxData(raw, &data) != nil || data.Interfaces == nil || len(data.Interfaces) > 256 {
		return nil, ErrProcessRunnerFault
	}
	out := make([]netip.Addr, 0, len(data.Interfaces))
	seen := map[netip.Addr]bool{}
	for _, v := range data.Interfaces {
		ip, err := netip.ParseAddr(v.IP)
		if err != nil || !usableBindIP(ip) || v.Index < 1 || v.Name == "" || seen[ip] {
			return nil, ErrProcessRunnerFault
		}
		seen[ip] = true
		out = append(out, ip)
	}
	return out, nil
}
func (e *ProcessExecutor) NetworkScan(ctx context.Context) error {
	raw, err := e.auxiliaryJSON(ctx, "scan")
	if err != nil {
		return err
	}
	var data struct {
		Results []struct {
			Interface struct {
				Name  string `json:"name"`
				Index int    `json:"index"`
				IP    string `json:"ip"`
			} `json:"interface"`
			Status json.RawMessage `json:"status"`
			Error  json.RawMessage `json:"error"`
		} `json:"results"`
	}
	if decodeAuxData(raw, &data) != nil || len(data.Results) != 1 {
		return ErrProcessRunnerFault
	}
	v := data.Results[0]
	if v.Interface.IP != e.BindIP.String() || v.Interface.Name == "" || v.Interface.Index < 1 || v.Error != nil || v.Status == nil {
		return ErrProcessRunnerFault
	}
	var status struct {
		Network    string          `json:"network"`
		Session    string          `json:"session"`
		Username   *string         `json:"username"`
		OnlineIP   *string         `json:"online_ip"`
		ObservedAt time.Time       `json:"observed_at"`
		Summary    json.RawMessage `json:"summary"`
	}
	if decodeAuxData(v.Status, &status) != nil || status.Session != "offline" || status.Username != nil || status.OnlineIP != nil || status.ObservedAt.IsZero() || !bytes.Equal(status.Summary, []byte("null")) || status.Network != "reachable" {
		return ErrProcessRunnerFault
	}
	return nil
}
func decodeAuxData(raw []byte, dst any) error {
	if CheckJSONStructure(raw, 1<<20) != nil {
		return ErrProcessRunnerFault
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(dst) != nil {
		return ErrProcessRunnerFault
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return ErrProcessRunnerFault
	}
	typ := reflect.TypeOf(dst).Elem()
	if len(fields) != typ.NumField() {
		return ErrProcessRunnerFault
	}
	for i := 0; i < typ.NumField(); i++ {
		if _, ok := fields[typ.Field(i).Tag.Get("json")]; !ok {
			return ErrProcessRunnerFault
		}
	}
	return nil
}
func (e *ProcessExecutor) auxiliaryJSON(ctx context.Context, action string) (json.RawMessage, error) {
	if e == nil || e.expectedUsername == "" || !usableBindIP(e.BindIP) {
		return nil, ErrProcessRunnerFault
	}
	var out boundedAuxBuffer
	args := append(e.jsonPrefix(), "network", action)
	run := e.processRunner()
	exit, err := run(ctx, e.CandidatePath, args, nil, &out, io.Discard)
	if err != nil || exit != 0 || ctx.Err() != nil || out.overflow {
		return nil, ErrProcessRunnerFault
	}
	var envelope struct {
		Schema  int             `json:"schema_version"`
		Command string          `json:"command"`
		OK      bool            `json:"ok"`
		Data    json.RawMessage `json:"data"`
	}
	if decodeAuxData(out.Bytes(), &envelope) != nil || envelope.Schema != 1 || envelope.Command != "network."+action || !envelope.OK || len(envelope.Data) == 0 {
		return nil, ErrProcessRunnerFault
	}
	return envelope.Data, nil
}

type boundedAuxBuffer struct {
	buffer   bytes.Buffer
	overflow bool
}

func (b *boundedAuxBuffer) Write(p []byte) (int, error) {
	if len(p) > 1<<20-b.Len() {
		b.overflow = true
		return 0, io.ErrShortWrite
	}
	return b.buffer.Write(p)
}

// PrivateWorkDirectory applies existing evidence ACL policy only to the new
// auxiliary fixture root. It never adjusts an existing repository or user home.
func PrivateWorkDirectory(path string) error { return ensureBundlePrivateDirectory(path) }

func (b *boundedAuxBuffer) Bytes() []byte { return b.buffer.Bytes() }
func (b *boundedAuxBuffer) Len() int      { return b.buffer.Len() }
