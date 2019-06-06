package ieproxy

import (
	"strings"
	"syscall"
	"unsafe"
)

func (psc *ProxyScriptConf) findProxyForURL(URL string) string {
	if !psc.Active {
		return ""
	}
	proxy, _ := getProxyForURL(psc.PreConfiguredURL, URL, psc.PreConfiguredURL == "")
	i := strings.Index(proxy, ";")
	if i >= 0 {
		return proxy[:i]
	}
	return proxy
}

func getProxyForURL(pacfileURL, URL string, autoDetect bool) (string, error) {
	pacfileURLPtr, err := syscall.UTF16PtrFromString(pacfileURL)
	if err != nil {
		return "", err
	}
	URLPtr, err := syscall.UTF16PtrFromString(URL)
	if err != nil {
		return "", err
	}

	winHttp := syscall.NewLazyDLL("Winhttp.dll")
	open := winHttp.NewProc("WinHttpOpen")
	handle, _, err := open.Call(0, 0, 0, 0, 0)
	if handle == 0 {
		return "", err
	}
	close := winHttp.NewProc("WinHttpCloseHandle")
	defer close.Call(handle)

	getProxyForUrl := winHttp.NewProc("WinHttpGetProxyForUrl")

	dwFlags := fWINHTTP_AUTOPROXY_CONFIG_URL
	dwAutoDetectFlags := autoDetectFlag(0)
	pfURLptr := pacfileURLPtr

	if autoDetect {
		dwFlags = fWINHTTP_AUTOPROXY_AUTO_DETECT
		dwAutoDetectFlags = fWINHTTP_AUTO_DETECT_TYPE_DNS_A | fWINHTTP_AUTO_DETECT_TYPE_DHCP
		pfURLptr = nil
	}

	options := tWINHTTP_AUTOPROXY_OPTIONS{
		dwFlags:                dwFlags, // adding cache might cause issues: https://github.com/mattn/go-ieproxy/issues/6
		dwAutoDetectFlags:      dwAutoDetectFlags,
		lpszAutoConfigUrl:      pfURLptr,
		lpvReserved:            nil,
		dwReserved:             0,
		fAutoLogonIfChallenged: true, // may not be optimal https://msdn.microsoft.com/en-us/library/windows/desktop/aa383153(v=vs.85).aspx
	}

	info := new(tWINHTTP_PROXY_INFO)

	ret, _, err := getProxyForUrl.Call(
		handle,
		uintptr(unsafe.Pointer(URLPtr)),
		uintptr(unsafe.Pointer(&options)),
		uintptr(unsafe.Pointer(info)),
	)
	if ret > 0 {
		err = nil
	}

	return StringFromUTF16Ptr(info.lpszProxy), err
}
