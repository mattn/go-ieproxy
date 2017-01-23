package ieproxy

import (
	"strings"
	"syscall"
	"unsafe"
)

type WINHTTP_AUTOPROXY_OPTIONS struct {
	dwFlags                AutoProxyFlag
	dwAutoDetectFlags      uint32
	lpszAutoConfigUrl      *uint16
	lpvReserved            *uint16
	dwReserved             uint32
	fAutoLogonIfChallenged bool
}
type AutoProxyFlag uint32

const (
	WINHTTP_AUTOPROXY_AUTO_DETECT         = AutoProxyFlag(0x00000001)
	WINHTTP_AUTOPROXY_CONFIG_URL          = AutoProxyFlag(0x00000002)
	WINHTTP_AUTOPROXY_NO_CACHE_CLIENT     = AutoProxyFlag(0x00080000)
	WINHTTP_AUTOPROXY_NO_CACHE_SVC        = AutoProxyFlag(0x00100000)
	WINHTTP_AUTOPROXY_NO_DIRECTACCESS     = AutoProxyFlag(0x00040000)
	WINHTTP_AUTOPROXY_RUN_INPROCESS       = AutoProxyFlag(0x00010000)
	WINHTTP_AUTOPROXY_RUN_OUTPROCESS_ONLY = AutoProxyFlag(0x00020000)
	WINHTTP_AUTOPROXY_SORT_RESULTS        = AutoProxyFlag(0x00400000)
)

type WINHTTP_PROXY_INFO struct {
	dwAccessType    uint32
	lpszProxy       *uint16
	lpszProxyBypass *uint16
}

func (apc *AutomaticProxyConf) findProxyForURL(URL string) string {
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

	options := WINHTTP_AUTOPROXY_OPTIONS{
		dwFlags:                WINHTTP_AUTOPROXY_CONFIG_URL + WINHTTP_AUTOPROXY_NO_CACHE_CLIENT + WINHTTP_AUTOPROXY_NO_CACHE_SVC,
		dwAutoDetectFlags:      0,
		lpszAutoConfigUrl:      pacfileURLPtr,
		lpvReserved:            nil,
		dwReserved:             0,
		fAutoLogonIfChallenged: true, // may not be optimal https://msdn.microsoft.com/en-us/library/windows/desktop/aa383153(v=vs.85).aspx
	}

	info := new(WINHTTP_PROXY_INFO)

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
