// +build !windows

package ieproxy

func getConf() WindowsProxyConf {
	return WindowsProxyConf{}
}

func overrideEnvWithStaticProxy() {
}
