// Package ieproxy is a utility to retrieve the proxy parameters (especially of Internet Explorer on windows)
//
// On windows, it gathers the parameters from the registry (regedit), while it uses env variable on other platforms
package ieproxy

import "os"

// ProxyConf gathers the configuration for proxy
type ProxyConf struct {
	Static    StaticProxyConf    // static configuration
	Script    ProxyScriptConf    // script configuration
	Automatic AutomaticProxyConf // automatic configuration
}

// StaticProxyConf contains the configuration for static proxy
type StaticProxyConf struct {
	// Is the proxy active?
	Active bool
	// Proxy address for each scheme (http, https)
	// "" (empty string) is the fallback proxy
	Protocols map[string]string
	// Addresses not to be browsed via the proxy (comma-separated, linux-like)
	NoProxy string
}

// ProxyScriptConf contains the configuration for automatic proxy
type ProxyScriptConf struct {
	// Is the proxy active?
	Active bool
	// URL of the .pac file
	URL string
}

type AutomaticProxyConf struct {
	// Is the proxy active?
	Active bool

	/*
		Note we have no proper way to detect if this *is* used.
		The best thing to do is just to make a syscall and see what it returns.
	*/
}

// GetConf retrieves the proxy configuration from the Windows Regedit
func GetConf() ProxyConf {
	return getConf()
}

// OverrideEnvWithStaticProxy writes new values to the
// `http_proxy`, `https_proxy` and `no_proxy` environment variables.
// The values are taken from the Windows Regedit (should be called in `init()` function - see example)
func OverrideEnvWithStaticProxy() {
	overrideEnvWithStaticProxy(GetConf(), os.Setenv)
}

// FindProxyForURL computes the proxy for a given URL according to the pac file
func (psc *ProxyScriptConf) FindProxyForURL(URL string) string {
	return psc.findProxyForURL(URL)
}

func (apc *AutomaticProxyConf) FindProxyForURL(URL string) string {
	return apc.findProxyForURL(URL)
}

type envSetter func(string, string) error
