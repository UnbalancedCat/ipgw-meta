//go:build !windows

package labpath

func readWiFiSSID(uint32) (WiFiSSID, error) { return "", ErrWiFiUnavailable }
