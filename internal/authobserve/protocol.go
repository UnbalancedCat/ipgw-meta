package authobserve

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"slices"
)

const MaxChild = 64 << 10
const MaxProduct = 48 << 10

var ErrProtocol = errors.New("authobserve: invalid child protocol")

type Child struct {
	SchemaVersion int             `json:"schema_version"`
	ToolKind      string          `json:"tool_kind"`
	Command       string          `json:"command"`
	ExitCode      int             `json:"exit_code"`
	Product       json.RawMessage `json:"product"`
	Observation   Observation     `json:"observation"`
}

func validCommand(s string) bool {
	return slices.Contains([]string{"profile_show", "status", "login", "logout"}, s)
}
func validCode(s string) bool {
	return slices.Contains([]string{"", "invalid_argument", "config", "network", "authentication", "session_conflict", "protocol_changed", "interaction_required", "unsupported", "internal"}, s)
}
func (o Observation) Validate() error {
	if !validCommand(o.Command) || !slices.Contains([]string{"complete", "incomplete"}, o.State) || !validCode(o.SDKCode) || o.Requests == nil || len(o.Requests) > MaxRequests {
		return ErrProtocol
	}
	if !slices.Contains([]string{"not_called", "success", "unknown", "cas_request", "cas_http_status", "activation_request", "activation_http_status", "credential_provider", "encryption", "status_identity", "canceled", "timeout", "logout_response_limit", "logout_response_parse", "logout_business_rejected"}, o.SDKCategory) {
		return ErrProtocol
	}
	if o.State == "incomplete" && len(o.Requests) != 0 {
		return ErrProtocol
	}
	if (o.SDKCategory == "success" || o.SDKCategory == "not_called") != (o.SDKCode == "") {
		return ErrProtocol
	}
	if o.SDKCategory == "not_called" && len(o.Requests) != 0 {
		return ErrProtocol
	}
	if o.Command == "profile_show" && (o.SDKCategory != "not_called" || len(o.Requests) != 0) {
		return ErrProtocol
	}
	switch o.SDKCategory {
	case "cas_request", "cas_http_status", "activation_request", "activation_http_status", "canceled", "timeout":
		if o.SDKCode != "network" {
			return ErrProtocol
		}
	case "credential_provider":
		if o.SDKCode != "config" {
			return ErrProtocol
		}
	case "logout_business_rejected":
		if o.Command != "logout" || o.SDKCode != "authentication" {
			return ErrProtocol
		}
	case "logout_response_limit", "logout_response_parse":
		if o.Command != "logout" || o.SDKCode != "protocol_changed" {
			return ErrProtocol
		}
	case "encryption":
		if o.SDKCode != "protocol_changed" {
			return ErrProtocol
		}
	case "status_identity":
		if o.SDKCode != "session_conflict" {
			return ErrProtocol
		}
	}
	for _, r := range o.Requests {
		if !slices.Contains([]string{"status", "discovery", "cas_get", "cas_post", "activation", "logout", "other"}, r.Role) || !slices.Contains([]string{"none", "canceled", "timeout", "dns", "tls_certificate", "tls_other", "other"}, r.Transport) {
			return ErrProtocol
		}
		if r.StatusCode != 0 && (r.StatusCode < 100 || r.StatusCode > 599) || !r.ResponseReceived && r.StatusCode != 0 {
			return ErrProtocol
		}
	}
	return o.validateLogout()
}
func DecodeChild(data []byte, command string, exit int) (Child, error) {
	var c Child
	fields, err := strictObject(data, MaxChild, []string{"schema_version", "tool_kind", "command", "exit_code", "product", "observation"})
	if err != nil || json.Unmarshal(data, &c) != nil || DecodeClosed(fields["observation"], &c.Observation, MaxChild) != nil || (c.SchemaVersion != 1 && c.SchemaVersion != 2) || c.ToolKind != "auth-observed-cli" || !validCommand(command) || c.Command != command || c.ExitCode != exit || c.Observation.Command != command || c.Observation.ValidateVersion(c.SchemaVersion) != nil || len(c.Product) > MaxProduct {
		return Child{}, ErrProtocol
	}
	if !slices.Contains([]int{0, 1, 2, 3, 4, 5, 6, 7, 130}, exit) {
		return Child{}, ErrProtocol
	}
	// The existing bounded CLI projector validates the contents of data/error.
	// Here enforce the independent envelope, command and exit/result relation.
	var p struct {
		SchemaVersion int             `json:"schema_version"`
		Command       string          `json:"command"`
		OK            bool            `json:"ok"`
		Data          json.RawMessage `json:"data"`
		Error         json.RawMessage `json:"error"`
	}
	branch := "error"
	if exit == 0 {
		branch = "data"
	}
	if _, err := strictObject(c.Product, MaxProduct, []string{"schema_version", "command", "ok", branch}); err != nil || json.Unmarshal(c.Product, &p) != nil || p.SchemaVersion != 1 {
		return Child{}, ErrProtocol
	}
	want := command
	if want == "profile_show" {
		want = "profile.show"
	}
	present := func(v json.RawMessage) bool { return len(v) > 0 && !bytes.Equal(bytes.TrimSpace(v), []byte("null")) }
	if p.Command != want || p.OK != (exit == 0) || present(p.Data) == present(p.Error) || p.OK != present(p.Data) {
		return Child{}, ErrProtocol
	}
	if c.Observation.SDKCategory == "success" && exit != 0 || c.Observation.SDKCode != "" && exit == 0 {
		return Child{}, ErrProtocol
	}
	if exit != 0 {
		parts, err := strictObject(p.Error, MaxProduct, []string{"code", "message", "retryable", "details"})
		var code string
		if err != nil || json.Unmarshal(parts["code"], &code) != nil || !validCode(code) || code == "" || c.Observation.SDKCode != "" && c.Observation.SDKCode != code {
			return Child{}, ErrProtocol
		}
	}
	return c, nil
}

type Window struct {
	State    string        `json:"state"`
	Reason   string        `json:"reason"`
	Commands []Observation `json:"commands"`
}

func NewWindow() Window {
	return Window{State: "not_started", Reason: "none", Commands: []Observation{}}
}
func (w Window) Validate() error {
	if w.Commands == nil || len(w.Commands) > 10 {
		return ErrProtocol
	}
	switch w.State {
	case "not_started":
		if w.Reason != "none" || len(w.Commands) != 0 {
			return ErrProtocol
		}
	case "complete":
		if w.Reason != "none" || len(w.Commands) == 0 {
			return ErrProtocol
		}
	case "incomplete":
		if !slices.Contains([]string{"overflow", "child_invalid", "interrupted"}, w.Reason) || len(w.Commands) != 0 {
			return ErrProtocol
		}
	default:
		return ErrProtocol
	}
	count := 0
	for _, o := range w.Commands {
		if o.Validate() != nil || o.State != "complete" {
			return ErrProtocol
		}
		count += len(o.Requests)
	}
	if count > 128 {
		return ErrProtocol
	}
	return nil
}
func (w *Window) Incomplete(reason string) {
	w.State = "incomplete"
	w.Reason = reason
	w.Commands = []Observation{}
}
func (w *Window) add(o Observation) {
	if w.State == "incomplete" {
		return
	}
	count := len(o.Requests)
	for _, old := range w.Commands {
		count += len(old.Requests)
	}
	if o.State != "complete" || len(w.Commands) >= 10 || count > 128 {
		w.Incomplete("overflow")
		return
	}
	w.State = "complete"
	w.Commands = append(w.Commands, o)
}

type LimitBuffer struct {
	bytes.Buffer
	Limit int
}

func (b *LimitBuffer) Write(p []byte) (int, error) {
	if len(p) > b.Limit-b.Len() {
		return 0, io.ErrShortWrite
	}
	return b.Buffer.Write(p)
}

// Adapter is owned by one sequential state machine. Product output stays in
// this transient buffer and the existing CLI projector; only Observation can
// reach the final development receipt.
type Adapter struct {
	Environment []string
	Window      Window
}

func (a *Adapter) Run(ctx context.Context, path string, args []string, stdin io.Reader, stdout, stderr io.Writer) (int, error) {
	command := Command(args)
	if command == "" {
		a.Window.Incomplete("child_invalid")
		return -1, ErrProtocol
	}
	cmd := exec.CommandContext(ctx, path, append([]string{"--internal-observed"}, args...)...)
	cmd.Env = append([]string{}, a.Environment...)
	cmd.Stdin = stdin
	cmd.Stderr = stderr
	out := &LimitBuffer{Limit: MaxChild}
	cmd.Stdout = out
	runErr := cmd.Run()
	defer clear(out.Bytes())
	if ctx.Err() != nil {
		a.Window.Incomplete("interrupted")
		return 130, ctx.Err()
	}
	code := 0
	if runErr != nil {
		var e *exec.ExitError
		if !errors.As(runErr, &e) {
			a.Window.Incomplete("child_invalid")
			return -1, ErrProtocol
		}
		code = e.ExitCode()
	}
	child, err := DecodeChild(out.Bytes(), command, code)
	if err != nil || child.SchemaVersion != 2 {
		a.Window.Incomplete("child_invalid")
		return -1, ErrProtocol
	}
	defer clear(child.Product)
	a.Window.add(child.Observation)
	if _, err = stdout.Write(child.Product); err != nil {
		return -1, err
	}
	return code, runErr
}
