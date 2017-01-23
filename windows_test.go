// +build windows

package ieproxy

import (
	"reflect"
	"testing"
)

var emptyMap, catchAllMap, multipleMap, multipleMapWithCatchAll map[string]string

func init() {
	emptyMap = make(map[string]string)
	catchAllMap = make(map[string]string)
	catchAllMap[""] = "127.0.0.1"
	multipleMap = make(map[string]string)
	multipleMap["http"] = "127.0.0.1"
	multipleMap["ftp"] = "128"
	multipleMapWithCatchAll = make(map[string]string)
	multipleMapWithCatchAll["http"] = "127.0.0.1"
	multipleMapWithCatchAll["ftp"] = "128"
	multipleMapWithCatchAll[""] = "129"
}

func TestParseRegedit(t *testing.T) {

	parsingSet := []struct {
		in  regeditValues
		out ProxyConf
	}{
		{
			in: regeditValues{},
			out: ProxyConf{
				Static: StaticProxyConf{
					Protocols: emptyMap, // to prevent it being <nil>
				},
			},
		},
		{
			in: regeditValues{
				ProxyServer: "127.0.0.1",
			},
			out: ProxyConf{
				Static: StaticProxyConf{
					Protocols: catchAllMap,
				},
			},
		},
		{
			in: regeditValues{
				ProxyServer: "http=127.0.0.1;ftp=128",
			},
			out: ProxyConf{
				Static: StaticProxyConf{
					Protocols: multipleMap,
				},
			},
		},
		{
			in: regeditValues{
				ProxyServer: "http=127.0.0.1;ftp=128;129",
			},
			out: ProxyConf{
				Static: StaticProxyConf{
					Protocols: multipleMapWithCatchAll,
				},
			},
		},
		{
			in: regeditValues{
				ProxyOverride: "example.com;microsoft.com",
			},
			out: ProxyConf{
				Static: StaticProxyConf{
					Protocols: emptyMap,
					NoProxy:   "example.com,microsoft.com",
				},
			},
		},
		{
			in: regeditValues{
				ProxyEnable: 1,
			},
			out: ProxyConf{
				Static: StaticProxyConf{
					Active:    true,
					Protocols: emptyMap,
				},
			},
		},
		{
			in: regeditValues{
				AutoConfigURL: "localhost/proxy.pac",
			},
			out: ProxyConf{
				Static: StaticProxyConf{
					Protocols: emptyMap,
				},
				Automatic: AutomaticProxyConf{
					Active: true,
					URL:    "localhost/proxy.pac",
				},
			},
		},
	}

	for _, p := range parsingSet {
		out := parseRegedit(p.in)
		if !reflect.DeepEqual(p.out, out) {
			t.Error("Got: ", out, "Expected: ", p.out)
		}
	}
}

func TestOverrideEnv(t *testing.T) {
	var callStack []string
	pseudoSetEnv := func(key, value string) error {
		callStack = append(callStack, key)
		callStack = append(callStack, value)
		return nil
	}
	overrideSet := []struct {
		in        ProxyConf
		callStack []string
	}{
		{
			callStack: []string{},
		},
		{
			in: ProxyConf{
				Static: StaticProxyConf{
					Active:    true,
					Protocols: catchAllMap,
				},
			},
			callStack: []string{"http_proxy", "127.0.0.1", "https_proxy", "127.0.0.1"},
		},
		{
			in: ProxyConf{
				Static: StaticProxyConf{
					Active:    false,
					Protocols: catchAllMap,
				},
			},
			callStack: []string{},
		},
		{
			in: ProxyConf{
				Static: StaticProxyConf{
					Active:    true,
					Protocols: multipleMap,
				},
			},
			callStack: []string{"http_proxy", "127.0.0.1"},
		},
		{
			in: ProxyConf{
				Static: StaticProxyConf{
					Active:    true,
					Protocols: multipleMapWithCatchAll,
				},
			},
			callStack: []string{"http_proxy", "127.0.0.1", "https_proxy", "129"},
		},
	}
	for _, o := range overrideSet {
		callStack = []string{}
		overrideEnvWithStaticProxy(o.in, pseudoSetEnv)
		if !reflect.DeepEqual(o.callStack, callStack) {
			t.Error("Got: ", callStack, "Expected: ", o.callStack)
		}
	}
}
