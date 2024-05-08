//go:build !windows && (!darwin || !cgo) && !ios
// +build !windows
// +build !darwin !cgo
// +build !ios

package ieproxy

func getConf() ProxyConf {
	return ProxyConf{}
}

func reloadConf() ProxyConf {
	return getConf()
}

func overrideEnvWithStaticProxy(pc ProxyConf, setenv envSetter) {
}
