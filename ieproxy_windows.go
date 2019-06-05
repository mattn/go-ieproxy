package ieproxy

import (
	"golang.org/x/sys/windows/registry"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

type regeditValues struct {
	ProxyServer   string
	ProxyOverride string
	ProxyEnable   uint64
	AutoConfigURL string
}

var once sync.Once
var windowsProxyConf ProxyConf

// GetConf retrieves the proxy configuration from the Windows Regedit
func getConf() ProxyConf {
	once.Do(writeConf)
	return windowsProxyConf
}

func writeConf() {
	if cfg, err := getConfFromCall(); err == nil {
		windowsProxyConf = ProxyConf{
			Static: StaticProxyConf{
				Active: cfg.lpszProxy != nil,
			},
			Script: ProxyScriptConf{
				Active:       cfg.lpszAutoConfigUrl != nil || cfg.fAutoDetect,
				AutoDiscover: cfg.fAutoDetect,
			},
		}

		if windowsProxyConf.Static.Active {
			protocol := make(map[string]string)
			for _, s := range strings.Split(StringFromUTF16Ptr(cfg.lpszProxy), ";") {
				if s == "" {
					continue
				}
				pair := strings.SplitN(s, "=", 2)
				if len(pair) > 1 {
					protocol[pair[0]] = pair[1]
				} else {
					protocol[""] = pair[0]
				}
			}

			windowsProxyConf.Static.Protocols = protocol
			if cfg.lpszProxyBypass != nil {
				windowsProxyConf.Static.NoProxy = strings.Replace(StringFromUTF16Ptr(cfg.lpszProxyBypass), ";", ",", -1)
			}
		}

		if windowsProxyConf.Script.Active {
			windowsProxyConf.Script.URL = StringFromUTF16Ptr(cfg.lpszAutoConfigUrl)
		}
	} else {
		regedit, _ := readRegedit() //If the syscall fails, backup to manual detection.
		windowsProxyConf = parseRegedit(regedit)
	}
}

func getConfFromCall() (*tWINHTTP_CURRENT_USER_IE_PROXY_CONFIG, error) {
	winHttp := syscall.NewLazyDLL("Winhttp.dll")
	open := winHttp.NewProc("WinHttpOpen")
	handle, _, err := open.Call(0, 0, 0, 0, 0)
	if handle == 0 {
		return &tWINHTTP_CURRENT_USER_IE_PROXY_CONFIG{}, err
	}
	close := winHttp.NewProc("WinHttpCloseHandle")
	defer close.Call(handle)

	getIEProxyConfigForCurrentUser := winHttp.NewProc("WinHttpGetIEProxyConfigForCurrentUser")

	config := new(tWINHTTP_CURRENT_USER_IE_PROXY_CONFIG)

	ret, _, err := getIEProxyConfigForCurrentUser.Call(uintptr(unsafe.Pointer(config)))
	if ret > 0 {
		err = nil
	}

	if config.fAutoDetect == true {
		detectAutoProxyConfigURL := winHttp.NewProc("WinHttpDetectAutoProxyConfigUrl")
		adFlag := uintptr(fWINHTTP_AUTO_DETECT_TYPE_DNS_A | fWINHTTP_AUTO_DETECT_TYPE_DHCP)
		acURL := new(uint16)

		ret, _, err = detectAutoProxyConfigURL.Call(
			adFlag,
			uintptr(unsafe.Pointer(acURL)),
		)
		if ret > 0 {
			err = nil
		}

		//if err.(syscall.Errno) == syscall.Errno(12180) {
		if err != nil {
			config.fAutoDetect = false
			err = nil
		}
	}

	return config, err
}

// OverrideEnvWithStaticProxy writes new values to the
// http_proxy, https_proxy and no_proxy environment variables.
// The values are taken from the Windows Regedit (should be called in init() function)
func overrideEnvWithStaticProxy(conf ProxyConf, setenv envSetter) {
	if conf.Static.Active {
		for _, scheme := range []string{"http", "https"} {
			url := mapFallback(scheme, "", conf.Static.Protocols)
			setenv(scheme+"_proxy", url)
		}
		if conf.Static.NoProxy != "" {
			setenv("no_proxy", conf.Static.NoProxy)
		}
	}
}

func parseRegedit(regedit regeditValues) ProxyConf {
	protocol := make(map[string]string)
	for _, s := range strings.Split(regedit.ProxyServer, ";") {
		if s == "" {
			continue
		}
		pair := strings.SplitN(s, "=", 2)
		if len(pair) > 1 {
			protocol[pair[0]] = pair[1]
		} else {
			protocol[""] = pair[0]
		}
	}

	return ProxyConf{
		Static: StaticProxyConf{
			Active:    regedit.ProxyEnable > 0,
			Protocols: protocol,
			NoProxy:   strings.Replace(regedit.ProxyOverride, ";", ",", -1), // to match linux style
		},
		Script: ProxyScriptConf{
			Active: regedit.AutoConfigURL != "",
			URL:    regedit.AutoConfigURL,
		},
	}
}

func readRegedit() (values regeditValues, err error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
	if err != nil {
		return
	}
	defer k.Close()

	values.ProxyServer, _, err = k.GetStringValue("ProxyServer")
	if err != nil && err != registry.ErrNotExist {
		return
	}
	values.ProxyOverride, _, err = k.GetStringValue("ProxyOverride")
	if err != nil && err != registry.ErrNotExist {
		return
	}

	values.ProxyEnable, _, err = k.GetIntegerValue("ProxyEnable")
	if err != nil && err != registry.ErrNotExist {
		return
	}

	values.AutoConfigURL, _, err = k.GetStringValue("AutoConfigURL")
	if err != nil && err != registry.ErrNotExist {
		return
	}
	err = nil
	return
}
