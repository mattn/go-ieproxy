package ieproxy

import (
	"strings"
	"sync"

	"golang.org/x/sys/windows/registry"
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
	regedit, _ := readRegedit()
	windowsProxyConf = parseRegedit(regedit)
}

// OverrideEnvWithStaticProxy writes new values to the
// http_proxy, https_proxy and no_proxy environment variables.
// The values are taken from the Windows Regedit (should be called in init() function)
func overrideEnvWithStaticProxy(conf ProxyConf, setenv envSetter) {
	if conf.Static.Active {
		for _, scheme := range []string{"http", "https"} {
			url, ok := conf.Static.Protocols[scheme]
			if !ok {
				url, ok = conf.Static.Protocols[""] // fallback conf
			}
			if ok {
				setenv(scheme+"_proxy", url)
			}
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
		Automatic: AutomaticProxyConf{
			Active: regedit.AutoConfigURL != "",
			URL:    regedit.AutoConfigURL,
		},
	}
}

func readRegedit() (values regeditValues, err error) {
	// Check config per user or per machine
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `Software\Policies\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
	if err != nil {
		return
	}
	defer k.Close()
	proxySettingsPerUser, _, err := k.GetIntegerValue("ProxySettingsPerUser")
	if err != nil && err != registry.ErrNotExist {
		return
	}
	var hkey registry.Key
	if err == nil && proxySettingsPerUser == 0 {
		hkey = registry.LOCAL_MACHINE
	} else {
		hkey = registry.CURRENT_USER
	}
	k.Close()

	k, err = registry.OpenKey(hkey, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
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
