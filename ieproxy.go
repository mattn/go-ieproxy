// Package ieproxy is a utility to retrieve the proxy parameters (especially of Internet Explorer on windows)
//
// On windows, it gathers the parameters from the registry (regedit), while it uses env variable on other platforms
package ieproxy

// ProxyConf gathers the configuration for proxy
type ProxyConf struct {
	Static    StaticProxyConf    // static configuration
	Automatic AutomaticProxyConf // automatic configuration
}

// StaticProxyConf containes the configuration for static proxy
type StaticProxyConf struct {
	// Is the proxy active?
	Active bool
	// Proxy address for each scheme (http, https)
	// "" (empty string) is the fallback proxy
	Protocols map[string]string
	// Addresses not to be browsed via the proxy (comma-separated, linux-like)
	NoProxy string
}

// AutomaticProxyConf contains the configuration for automatic proxy
type AutomaticProxyConf struct {
	// Is the proxy active?
	Active bool
	// URL of the .pac file
	URL string
}

// GetConf retrieves the proxy configuration from the Windows Regedit
func GetConf() ProxyConf {
	return getConf()
}

// OverrideEnvWithStaticProxy writes new values to the
// `http_proxy`, `https_proxy` and `no_proxy` environment variables.
// The values are taken from the Windows Regedit (should be called in `init()` function - see example)
func OverrideEnvWithStaticProxy() {
	overrideEnvWithStaticProxy()
}
