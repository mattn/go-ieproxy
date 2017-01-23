// +build !windows

package ieproxy

func (apc *AutomaticProxyConf) findProxyForURL(URL string) string {
	return ""
}
