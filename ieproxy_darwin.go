//go:build !ios && !iossimulator
// +build !ios,!iossimulator

package ieproxy

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

var once sync.Once
var darwinProxyConf ProxyConf

// GetConf retrieves the proxy configuration from the Windows Regedit
func getConf() ProxyConf {
	once.Do(writeConf)
	return darwinProxyConf
}

// reloadConf forces a reload of the proxy configuration.
func reloadConf() ProxyConf {
	writeConf()
	return getConf()
}

func parseConf(b []byte) {
	/*
		% scutil --proxy
		<dictionary> {
		  ExceptionsList : <array> {
		    0 : *.local
		    1 : 169.254/16
		  }
		  FTPPassive : 1
		  HTTPEnable : 1
		  HTTPPort : 8081
		  HTTPProxy : example.jp
		  HTTPSEnable : 1
		  HTTPSPort : 8080
		  HTTPSProxy : example.com
		  HTTPSUser : foo
		  ProxyAutoConfigEnable : 1
		  ProxyAutoConfigURLString : example.com/foo.pac
		}
	*/
	inExceptionsList := false
	exceptionsList := make([]string, 0)
	proxyMap := map[string]string{}
	scanner := bufio.NewScanner(bytes.NewReader(b))
	for scanner.Scan() {
		t := scanner.Text()
		t = strings.TrimSpace(t)
		if inExceptionsList {
			if strings.Index(t, ":") > 0 {
				s := strings.SplitN(t, ":", 2)
				exceptionsList = append(exceptionsList, strings.TrimSpace(s[1]))
			}
			if t == "}" {
				inExceptionsList = false
			}
			continue
		}
		if strings.Index(t, ":") > 0 {
			s := strings.SplitN(t, ":", 2)
			k := strings.TrimSpace(s[0])
			if k == "ExceptionsList" {
				inExceptionsList = true
				continue
			}
			v := strings.TrimSpace(s[1])
			proxyMap[k] = v
		}
	}

	darwinProxyConf = ProxyConf{}

	// http
	if v, ok := proxyMap["HTTPEnable"]; ok && v == "1" {
		darwinProxyConf.Static.Active = true
		if darwinProxyConf.Static.Protocols == nil {
			darwinProxyConf.Static.Protocols = make(map[string]string)
		}
		httpProxy := fmt.Sprintf("%s:%s", proxyMap["HTTPProxy"], proxyMap["HTTPPort"])
		darwinProxyConf.Static.Protocols["http"] = httpProxy
	}

	// https
	if v, ok := proxyMap["HTTPSEnable"]; ok && v == "1" {
		darwinProxyConf.Static.Active = true
		if darwinProxyConf.Static.Protocols == nil {
			darwinProxyConf.Static.Protocols = make(map[string]string)
		}
		httpProxy := fmt.Sprintf("%s:%s", proxyMap["HTTPSProxy"], proxyMap["HTTPSPort"])
		darwinProxyConf.Static.Protocols["https"] = httpProxy
	}

	// noproxy
	if darwinProxyConf.Static.Active {
		if len(exceptionsList) > 0 {
			darwinProxyConf.Static.NoProxy = strings.Join(exceptionsList, ",")
		}
	}

	// pac
	if v, ok := proxyMap["ProxyAutoConfigEnable"]; ok && v == "1" {
		darwinProxyConf.Automatic.PreConfiguredURL = proxyMap["ProxyAutoConfigURLString"]
		darwinProxyConf.Automatic.Active = true
	}

}

func writeConf() {
	cmd := exec.Command("scutil", "--proxy")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return
	}
	parseConf(out.Bytes())
}

// OverrideEnvWithStaticProxy writes new values to the
// http_proxy, https_proxy and no_proxy environment variables.
// The values are taken from the MacOS System Preferences.
func overrideEnvWithStaticProxy(conf ProxyConf, setenv envSetter) {
	if conf.Static.Active {
		for _, scheme := range []string{"http", "https"} {
			url := conf.Static.Protocols[scheme]
			if url != "" {
				setenv(scheme+"_proxy", url)
			}
		}
		if conf.Static.NoProxy != "" {
			setenv("no_proxy", conf.Static.NoProxy)
		}
	}
}
