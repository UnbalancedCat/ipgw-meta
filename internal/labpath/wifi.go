package labpath

import (
	"context"
	"encoding/binary"
	"errors"
)

// WiFiSSID contains only public campus scenario labels, never arbitrary SSIDs.
type WiFiSSID string

const (
	WiFiNEU    WiFiSSID = "neu"
	WiFiNEU24G WiFiSSID = "neu_24g"
	WiFiMobile WiFiSSID = "neu_mobile"
)

var (
	ErrWiFiUnavailable  = errors.New("labpath: wifi unavailable")
	ErrWiFiPermission   = errors.New("labpath: wifi permission denied")
	ErrWiFiDisconnected = errors.New("labpath: wifi disconnected")
	ErrWiFiUnknown      = errors.New("labpath: wifi outside allowlist")
	ErrWiFiUnsupported  = errors.New("labpath: wifi scenario unsupported")
)

func (s WiFiSSID) Supported() bool { return s == WiFiNEU || s == WiFiNEU24G }

// Parse the fixed Windows ABI prefix, without interpreting profile or BSSID.
// WLAN_CONNECTION_ATTRIBUTES is 604 bytes on both supported Windows targets.
func parseWiFiConnection(raw []byte) (WiFiSSID, error) {
	if len(raw) != 604 {
		return "", ErrWiFiUnavailable
	}
	if binary.LittleEndian.Uint32(raw[:4]) != 1 {
		return "", ErrWiFiDisconnected
	}
	n := binary.LittleEndian.Uint32(raw[520:524])
	if n == 0 || n > 32 {
		return "", ErrWiFiUnavailable
	}
	switch string(raw[524 : 524+int(n)]) {
	case "NEU":
		return WiFiNEU, nil
	case "NEU-2.4G":
		return WiFiNEU24G, nil
	case "NEU-Mobile":
		return WiFiMobile, ErrWiFiUnsupported
	default:
		return "", ErrWiFiUnknown
	}
}

func ReadWiFiSSID(ctx context.Context, index uint32) (WiFiSSID, error) {
	if ctx == nil || index == 0 {
		return "", ErrWiFiUnavailable
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	ssid, err := readWiFiSSID(index)
	if canceled := ctx.Err(); canceled != nil {
		return "", canceled
	}
	return ssid, err
}
