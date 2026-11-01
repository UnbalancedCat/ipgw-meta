package authobserve

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func failureChild() Child {
	c := New("login")
	c.result(errors.New("untrusted-private-canary"))
	return Child{SchemaVersion: 1, ToolKind: "auth-observed-cli", Command: "login", ExitCode: 1,
		Product: json.RawMessage(`{"schema_version":1,"command":"login","ok":false,"error":{"code":"internal","message":"internal error","retryable":false,"details":{}}}`), Observation: c.Snapshot()}
}
func TestChildProtocolStrictFramingAndConsistency(t *testing.T) {
	good := failureChild()
	data, _ := json.Marshal(good)
	if _, err := DecodeChild(data, "login", 1); err != nil {
		t.Fatal("valid failure rejected")
	}
	success := good
	success.ExitCode = 0
	success.Observation.SDKCode = ""
	success.Observation.SDKCategory = "success"
	success.Product = json.RawMessage(`{"schema_version":1,"command":"login","ok":true,"data":{"outcome":"logged_in"}}`)
	successData, _ := json.Marshal(success)
	if _, err := DecodeChild(successData, "login", 0); err != nil {
		t.Fatal("valid success rejected")
	}
	mutations := []func([]byte) []byte{
		func(b []byte) []byte {
			return []byte(strings.Replace(string(b), `"schema_version":1`, `"schema_version":3`, 1))
		},
		func(b []byte) []byte {
			return []byte(strings.Replace(string(b), `"exit_code":1`, `"exit_code":null`, 1))
		},
		func(b []byte) []byte { return append([]byte(`{"extra":0,`), b[1:]...) },
		func(b []byte) []byte { return append([]byte(`{"schema_version":1,`), b[1:]...) },
		func(b []byte) []byte { return append(b, []byte(`{}`)...) },
		func(b []byte) []byte {
			return []byte(strings.Replace(string(b), `"code":"internal"`, `"code":"network"`, 1))
		},
		func(b []byte) []byte {
			return []byte(strings.Replace(string(b), `"requests":[]`, `"requests":null`, 1))
		},
		func(b []byte) []byte {
			return []byte(strings.Replace(string(b), `"ok":false`, `"ok":false,"data":{}`, 1))
		},
		func(b []byte) []byte {
			return []byte(strings.Replace(string(b), `"details":{}`, `"details":{"x":0,"x":1}`, 1))
		},
	}
	for i, mutate := range mutations {
		if _, err := DecodeChild(mutate(append([]byte{}, data...)), "login", 1); err == nil {
			t.Fatalf("mutation %d accepted", i)
		}
	}
	if _, err := DecodeChild(data, "status", 1); err == nil {
		t.Fatal("command mismatch")
	}
	if _, err := DecodeChild(data, "login", 3); err == nil {
		t.Fatal("process exit mismatch")
	}
}
func TestWindowOverflowClearsWholeSet(t *testing.T) {
	w := NewWindow()
	for i := 0; i < 5; i++ {
		o := New("login").Snapshot()
		o.SDKCategory = "success"
		for j := 0; j < 32; j++ {
			o.Requests = append(o.Requests, Request{Role: "cas_get", ResponseReceived: true, StatusCode: 200, Transport: "none"})
		}
		w.add(o)
	}
	if w.State != "incomplete" || w.Reason != "overflow" || len(w.Commands) != 0 || w.Validate() != nil {
		t.Fatal("window published truncated observations")
	}
	w.add(New("profile_show").Snapshot())
	if len(w.Commands) != 0 {
		t.Fatal("overflow not sticky")
	}
}
func TestCommandRejectsOverridesAndQR(t *testing.T) {
	prefix := []string{"--json", "--profile", "synthetic", "--bind-ip", "192.0.2.9"}
	for _, tail := range [][]string{{"login", "--method", "qr"}, {"profile", "add"}, {"login", "--method", "password", "--switch"}, {"--config", "anything", "status"}} {
		if Command(append(append([]string{}, prefix...), tail...)) != "" {
			t.Fatal("unsupported command")
		}
	}
	if Command(append(prefix, "profile", "show")) != "profile_show" {
		t.Fatal("profile command")
	}
}

// Only the test executable handles this marker. The child is synthetic and
// never calls app.New, resolves config, reads certificates or opens a socket.
func TestMain(m *testing.M) {
	if mode := os.Getenv("AUTHOBSERVE_SYNTHETIC_CHILD"); mode != "" {
		if len(os.Args) < 2 || os.Args[1] != "--internal-observed" {
			os.Exit(9)
		}
		c := failureChild()
		c.SchemaVersion = 2
		switch mode {
		case "valid":
			input, _ := io.ReadAll(os.Stdin)
			if string(input) != "synthetic-input" {
				os.Exit(9)
			}
			fmt.Fprint(os.Stderr, "synthetic-private-stderr")
		case "invalid":
			fmt.Print(`{"invalid":true}`)
			os.Exit(1)
		case "exit_mismatch":
			c.ExitCode = 3
		}
		b, _ := json.Marshal(c)
		os.Stdout.Write(b)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
func TestAdapterKeepsProductPrivateAndRejectsBadChild(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"--json", "--profile", "synthetic", "--bind-ip", "192.0.2.9", "login", "--method", "password"}
	for _, mode := range []string{"valid", "invalid", "exit_mismatch"} {
		t.Run(mode, func(t *testing.T) {
			adapter := &Adapter{Environment: append(os.Environ(), "AUTHOBSERVE_SYNTHETIC_CHILD="+mode), Window: NewWindow()}
			var out, stderr bytes.Buffer
			code, err := adapter.Run(context.Background(), exe, args, strings.NewReader("synthetic-input"), &out, &stderr)
			if mode == "valid" {
				if code != 1 || err == nil || adapter.Window.State != "complete" || stderr.String() != "synthetic-private-stderr" || !bytes.Equal(out.Bytes(), failureChild().Product) {
					t.Fatal("adapter altered streams or outcome")
				}
				data, _ := json.Marshal(adapter.Window)
				if bytes.Contains(data, []byte("private")) || bytes.Contains(data, []byte("product")) {
					t.Fatal("raw stream leaked")
				}
			} else if code != -1 || err == nil || out.Len() != 0 || adapter.Window.Reason != "child_invalid" {
				t.Fatal("bad child accepted")
			}
		})
	}
}
func FuzzChildProtocol(f *testing.F) {
	data, _ := json.Marshal(failureChild())
	f.Add(data)
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > MaxChild {
			return
		}
		c, err := DecodeChild(b, "login", 1)
		if err == nil && c.Observation.Validate() != nil {
			t.Fatal("invalid accepted")
		}
	})
}
