// +build windows

package ieproxy

import (
	"reflect"
	"testing"
)

func TestParseRegedit(t *testing.T) {

	emptyMap := make(map[string]string)
	catchAllMap := make(map[string]string)
	catchAllMap[""] = "127.0.0.1"
	multipleMap := make(map[string]string)
	multipleMap["http"] = "127.0.0.1"
	multipleMap["ftp"] = "128"
	multipleMapWithCatchAll := make(map[string]string)
	multipleMapWithCatchAll["http"] = "127.0.0.1"
	multipleMapWithCatchAll["ftp"] = "128"
	multipleMapWithCatchAll[""] = "129"

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
