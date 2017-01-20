// +build !windows

package ieproxy

func GetConf() WindowsProxyConf {
	return nil
}

func OverrideEnvWithStaticProxy() {
}
