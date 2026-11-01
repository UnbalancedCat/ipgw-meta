package casbundle

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/UnbalancedCat/ipgw-meta/internal/casprovenance"
)

func TestOutputLimitCancelsAndDoesNotAppend(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b := boundedOutput{limit: 65536, cancel: cancel}
	if n, err := b.Write(bytes.Repeat([]byte("x"), 65536)); n != 65536 || err != nil {
		t.Fatal("exact boundary rejected")
	}
	if _, err := b.Write([]byte("x")); !errors.Is(err, io.ErrShortWrite) || !b.overflow || len(b.data) != 65536 || ctx.Err() != context.Canceled {
		t.Fatal("overflow did not cancel")
	}
}
func TestSelfTestEnvelopeRejectsExtraAndClaims(t *testing.T) {
	r, err := casprovenance.Analyze(context.Background(), []casprovenance.Input{{Kind: "inline_script", Data: []byte(";")}})
	if err != nil {
		t.Fatal(err)
	}
	e := casprovenance.SelfTestEnvelope{ToolKind: ToolKind, Mode: "synthetic-prototype", Binding: "not_packaged", SelfTestResult: "passed", Checks: 11, Report: r}
	data := encode(e)
	if _, err = casprovenance.DecodeSelfTest(data); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range [][]byte{bytes.Replace(data, []byte(`"checks":11`), []byte(`"checks":11,"arbitrary":"synthetic-canary"`), 1), bytes.Replace(data, []byte(`"not_packaged"`), []byte(`"verified"`), 1), append(data, []byte(`{}`)...)} {
		if _, err = casprovenance.DecodeSelfTest(invalid); err != casprovenance.ErrReport {
			t.Fatal("invalid child output accepted")
		}
	}
}

func TestReceiptClosedProjection(t *testing.T) {
	r := Run(context.Background(), "untrusted", "")
	data, err := EncodeReceipt(r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = DecodeReceipt(data); err != nil {
		t.Fatal(err)
	}
	for _, bad := range [][]byte{bytes.Replace(data, []byte(`"exe_attempts":0`), []byte(`"exe_attempts":1`), 1), bytes.Replace(data, []byte(`"exe_attempts":0,`), nil, 1), bytes.Replace(data, []byte(`"dns":0`), []byte(`"dns":0,"dns":0`), 1), append(data, []byte(`{}`)...)} {
		if _, err = DecodeReceipt(bad); err != ErrInvalid {
			t.Fatal("inconsistent receipt accepted")
		}
	}
}
