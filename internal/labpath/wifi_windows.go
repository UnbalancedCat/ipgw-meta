//go:build windows

package labpath

import (
	"golang.org/x/sys/windows"
	"unsafe"
)

// Only current-connection queries are imported: no scan, connect, or profile API.
func readWiFiSSID(index uint32) (WiFiSSID, error) {
	iphlp := windows.NewLazySystemDLL("iphlpapi.dll")
	wlan := windows.NewLazySystemDLL("wlanapi.dll")
	toLuid := iphlp.NewProc("ConvertInterfaceIndexToLuid")
	toGuid := iphlp.NewProc("ConvertInterfaceLuidToGuid")
	open := wlan.NewProc("WlanOpenHandle")
	closeHandle := wlan.NewProc("WlanCloseHandle")
	query := wlan.NewProc("WlanQueryInterface")
	free := wlan.NewProc("WlanFreeMemory")
	for _, p := range []*windows.LazyProc{toLuid, toGuid, open, closeHandle, query, free} {
		if p.Find() != nil {
			return "", ErrWiFiUnavailable
		}
	}
	var luid uint64
	var guid windows.GUID
	if code, _, _ := toLuid.Call(uintptr(index), uintptr(unsafe.Pointer(&luid))); code != 0 {
		return "", ErrWiFiUnavailable
	}
	if code, _, _ := toGuid.Call(uintptr(unsafe.Pointer(&luid)), uintptr(unsafe.Pointer(&guid))); code != 0 {
		return "", ErrWiFiUnavailable
	}
	var version uint32
	var handle windows.Handle
	code, _, _ := open.Call(2, 0, uintptr(unsafe.Pointer(&version)), uintptr(unsafe.Pointer(&handle)))
	if code != 0 {
		return "", wifiNativeError(code)
	}
	defer closeHandle.Call(uintptr(handle), 0)
	var size, kind uint32
	var data unsafe.Pointer
	// wlan_intf_opcode_current_connection = 7.
	code, _, _ = query.Call(uintptr(handle), uintptr(unsafe.Pointer(&guid)), 7, 0, uintptr(unsafe.Pointer(&size)), uintptr(unsafe.Pointer(&data)), uintptr(unsafe.Pointer(&kind)))
	if data != nil {
		defer free.Call(uintptr(data))
	}
	if code != 0 {
		return "", wifiNativeError(code)
	}
	if data == nil || size != 604 {
		return "", ErrWiFiUnavailable
	}
	return parseWiFiConnection(unsafe.Slice((*byte)(data), int(size)))
}

func wifiNativeError(code uintptr) error {
	if code == 5 {
		return ErrWiFiPermission
	}
	if code == 5023 {
		return ErrWiFiDisconnected
	}
	return ErrWiFiUnavailable
}
