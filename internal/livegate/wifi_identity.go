package livegate

// Arbitrary SSIDs are never accepted into evidence. Mobile remains deferred.
func validWiFiIdentity(wifi bool, ssid *string) bool {
	if !wifi {
		return ssid == nil
	}
	return ssid != nil && (*ssid == "neu" || *ssid == "neu_24g")
}
