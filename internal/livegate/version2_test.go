package livegate

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func testHash(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

func TestHistoricalV1CannotBeEmittedOrReinterpreted(t *testing.T) {
	e := validPasswordEvidence()
	e.SchemaVersion = 1
	e.RunnerSHA256 = ""
	raw, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeEvidence(bytes.NewReader(raw))
	if err != nil || got.SchemaVersion != 1 {
		t.Fatal("historical read", err)
	}
	if EncodeEvidence(&bytes.Buffer{}, got) == nil {
		t.Fatal("historical emission allowed")
	}
	// Version rejection precedes every platform-specific filesystem operation.
	if _, err := (BundlePublisher{OutputDir: t.TempDir()}).Publish(got); err == nil {
		t.Fatal("historical publication allowed")
	}
	qr := validTerminalQREvidence()
	qr.Platform = PlatformWindowsAMD64
	qr.Testbed = TestbedBHKWindows
	if qr.Validate() != nil {
		t.Fatal("v2 Windows QR rejected")
	}
	qr.SchemaVersion = 1
	qr.RunnerSHA256 = ""
	if qr.Validate() == nil {
		t.Fatal("v1 Windows QR reinterpreted")
	}
	e.RunnerSHA256 = strings.Repeat("a", 64)
	if e.Validate() == nil {
		t.Fatal("v1 accepted v2 field")
	}
	e.SchemaVersion = 2
	e.RunnerSHA256 = ""
	if e.Validate() == nil {
		t.Fatal("v2 missing helper binding")
	}
}

func TestPathLossStopsCleanupWithoutAnotherCommand(t *testing.T) {
	inner := &scriptedPreparedExecutor{observations: successfulObservations(SuitePasswordCore)}
	valid := true
	guard := &pathExecutor{inner: inner, check: func(context.Context) bool { return valid }, interval: time.Hour}
	if guard.Execute(context.Background(), ActionInitialStatus).RunnerFault {
		t.Fatal("valid path rejected")
	}
	valid = false
	for _, a := range []CommandAction{ActionPasswordLogin, ActionCleanupLogout, ActionCleanupStatus} {
		if !guard.Execute(context.Background(), a).RunnerFault {
			t.Fatal("lost path accepted")
		}
	}
	if len(inner.actions) != 1 {
		t.Fatal("commands ran after path loss")
	}
}
