package labpath

import (
	"context"
	"encoding/binary"
	"errors"
	"testing"
)

func TestWiFiConnectionClosedProjection(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want WiFiSSID
		err  error
	}{
		{"NEU", WiFiNEU, nil}, {"NEU-2.4G", WiFiNEU24G, nil},
		{"NEU-Mobile", WiFiMobile, ErrWiFiUnsupported}, {"neu", "", ErrWiFiUnknown},
		{"prefix-NEU", "", ErrWiFiUnknown}, {"NEU\x00", "", ErrWiFiUnknown},
	} {
		b := make([]byte, 604)
		binary.LittleEndian.PutUint32(b, 1)
		binary.LittleEndian.PutUint32(b[520:], uint32(len(tc.raw)))
		copy(b[524:], tc.raw)
		got, err := parseWiFiConnection(b)
		if got != tc.want || !errors.Is(err, tc.err) {
			t.Fatal("incorrect SSID classification")
		}
		b[0] = 0
		if _, err = parseWiFiConnection(b); !errors.Is(err, ErrWiFiDisconnected) {
			t.Fatal("disconnected accepted")
		}
	}
	for _, n := range []int{0, 556, 603, 605, 1000} {
		if _, err := parseWiFiConnection(make([]byte, n)); err == nil {
			t.Fatal("bad size")
		}
	}
	for _, n := range []uint32{0, 33, 0xffffffff} {
		b := make([]byte, 604)
		b[0] = 1
		binary.LittleEndian.PutUint32(b[520:], n)
		if _, err := parseWiFiConnection(b); !errors.Is(err, ErrWiFiUnavailable) {
			t.Fatal("bad length")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ReadWiFiSSID(ctx, 1); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation lost")
	}
}

func FuzzWiFiConnection(f *testing.F) {
	f.Add(make([]byte, 604))
	f.Fuzz(func(t *testing.T, b []byte) { _, _ = parseWiFiConnection(b) })
}
