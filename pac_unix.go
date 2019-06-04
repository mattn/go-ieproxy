// +build !windows

package ieproxy

func (apc *ProxyScriptConf) findProxyForURL(URL string) string {
	return ""
}
