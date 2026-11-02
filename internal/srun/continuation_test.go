package srun

import (
	"strings"
	"testing"
)

func TestQRContinuationClosedEnvelope(t *testing.T) {
	good := `{"code":0,"message":"continue","Redirect":"https://cas.example/login","ID":""}`
	if _, err := ParseQRContinuation([]byte(good)); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{
		`{}`, `null`, `callback(` + good + `)`, good + `{}`, strings.Repeat(" ", 256<<10) + good,
		strings.Replace(good, `"code":0`, `"code":0,"code":0`, 1),
		strings.Replace(good, `"code":0`, `"code":"0"`, 1),
		strings.Replace(good, `"code":0`, `"code":1`, 1),
		strings.Replace(good, `"code":0`, `"code":0.0`, 1),
		strings.Replace(good, `"message":"continue"`, `"message":null`, 1),
		strings.Replace(good, `"ID":""`, `"ID":4`, 1),
		strings.Replace(good, `"ID":""`, `"success":true`, 1),
		strings.Replace(good, `https://cas.example/login`, ``, 1),
		good[:len(good)-1],
	} {
		if _, err := ParseQRContinuation([]byte(input)); err == nil {
			t.Fatal("invalid envelope accepted")
		}
	}
}

func FuzzQRContinuation(f *testing.F) {
	f.Add([]byte(`{"code":0,"message":"continue","Redirect":"https://cas.example/login","ID":""}`))
	f.Add([]byte(`{"code":0,"code":0}`))
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 300<<10 {
			return
		}
		_, _ = ParseQRContinuation(b)
	})
}
