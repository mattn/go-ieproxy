// +build windows

package ieproxy

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

var (
	once       sync.Once
	httpProxy  = safeParseURL(os.Getenv("HTTP_PROXY"))
	httpsProxy = safeParseURL(os.Getenv("HTTPS_PROXY"))
	noProxy    = os.Getenv("NO_PROXY")
)

func safeParseURL(s string) *url.URL {
	if !strings.HasPrefix(s, "http") {
		s = "http://" + s
	}
	u, _ := url.Parse(s)
	return u
}

func getProxy() {
	once.Do(func() {
		var h syscall.Handle
		pathp, _ := syscall.UTF16PtrFromString(`Software\Microsoft\Windows\CurrentVersion\Internet Settings`)
		err := syscall.RegOpenKeyEx(syscall.HKEY_CURRENT_USER, pathp, 0, syscall.KEY_READ, &h)
		if err == nil {
			defer syscall.RegCloseKey(h)
			var typ uint32
			var buf [1 << 10]uint16
			n := uint32(len(buf) * 2)
			err = syscall.RegQueryValueEx(h, syscall.StringToUTF16Ptr("ProxyServer"), nil, &typ, (*byte)(unsafe.Pointer(&buf[0])), &n)
			if err == nil {
				var u *url.URL
				for _, setting := range strings.Split(syscall.UTF16ToString(buf[:]), ";") {
					setting = strings.TrimSpace(setting)
					if strings.HasPrefix(setting, "http=") {
						u = safeParseURL(setting[5:])
						if u != nil {
							httpProxy = u
						}
					} else if strings.HasPrefix(setting, "https=") {
						u = safeParseURL(setting[6:])
						if u != nil {
							httpsProxy = u
						}
					}
				}
			}
			err = syscall.RegQueryValueEx(h, syscall.StringToUTF16Ptr("ProxyOverride"), nil, &typ, (*byte)(unsafe.Pointer(&buf[0])), &n)
			if err == nil {
				noProxy = syscall.UTF16ToString(buf[:])
			}
		}
	})
}

func useProxy(addr string) bool {
	if len(addr) == 0 {
		return true
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() {
			return false
		}
	}

	if noProxy == "*" {
		return false
	}

	addr = strings.ToLower(strings.TrimSpace(addr))
	if hasPort(addr) {
		addr = addr[:strings.LastIndex(addr, ":")]
	}

	for _, p := range strings.Split(noProxy, ",") {
		p = strings.ToLower(strings.TrimSpace(p))
		if len(p) == 0 {
			continue
		}
		if hasPort(p) {
			p = p[:strings.LastIndex(p, ":")]
		}
		if addr == p {
			return false
		}
		if p[0] == '.' && (strings.HasSuffix(addr, p) || addr == p[1:]) {
			// noProxy ".foo.com" matches "bar.foo.com" or "foo.com"
			return false
		}
		if p[0] != '.' && strings.HasSuffix(addr, p) && addr[len(addr)-len(p)-1] == '.' {
			// noProxy "foo.com" matches "bar.foo.com"
			return false
		}
	}
	return true
}

func hasPort(s string) bool {
	return strings.LastIndex(s, ":") > strings.LastIndex(s, "]")
}

var portMap = map[string]string{
	"http":  "80",
	"https": "443",
}

func canonicalAddr(url *url.URL) string {
	addr := url.Host
	if !hasPort(addr) {
		return addr + ":" + portMap[url.Scheme]
	}
	return addr
}

func ProxyFromIE(req *http.Request) (*url.URL, error) {
	getProxy()

	if !useProxy(canonicalAddr(req.URL)) {
		return nil, nil
	}
	if req.URL.Scheme == "http" && httpProxy != nil {
		return httpProxy, nil
	} else if req.URL.Scheme == "https" && httpsProxy != nil {
		return httpsProxy, nil
	}
	return nil, nil
}
