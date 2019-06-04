package ieproxy

import (
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"unsafe"
)

func usingWPAD() bool {
	req := http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "http",
			Host:   "wpad",
			Path:   "/wpad.dat", //Windows maps wpad to the auto discovered URL, which should contain wpad.dat
		},
	}
	resp, err := http.DefaultClient.Do(&req)

	if err != nil {
		return false
	}

	return resp.StatusCode == 200 //If we can't obtain it in the first place, there's no point in activating WPAD.
}

func (apc *AutomaticProxyConf) findProxyForURL(URL string) string {
	if !apc.Active {
		return ""
	}
	proxy, _ := getAutoProxyForURL(URL)
	i := strings.Index(proxy, ";")
	if i >= 0 {
		return proxy[:i]
	}
	return proxy
}

func getAutoProxyForURL(URL string) (string, error) {
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
		dwFlags:                fWINHTTP_AUTOPROXY_AUTO_DETECT, // adding cache might cause issues: https://github.com/mattn/go-ieproxy/issues/6
		dwAutoDetectFlags:      fWINHTTP_AUTO_DETECT_TYPE_DHCP & fWINHTTP_AUTO_DETECT_TYPE_DNS_A,
		lpszAutoConfigUrl:      nil,
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
