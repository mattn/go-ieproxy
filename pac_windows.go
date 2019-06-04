package ieproxy

import (
	"strings"
	"syscall"
	"unsafe"
)

func (apc *ProxyScriptConf) findProxyForURL(URL string) string {
	if !apc.Active {
		return ""
	}
	proxy, _ := getProxyForURL(apc.URL, URL)
	i := strings.Index(proxy, ";")
	if i >= 0 {
		return proxy[:i]
	}
	return proxy
}

func getProxyForURL(pacfileURL, URL string) (string, error) {
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

	options := tWINHTTP_AUTOPROXY_OPTIONS{
		dwFlags:                fWINHTTP_AUTOPROXY_CONFIG_URL, // adding cache might cause issues: https://github.com/mattn/go-ieproxy/issues/6
		dwAutoDetectFlags:      0,
		lpszAutoConfigUrl:      pacfileURLPtr,
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
