//go:build darwin
// +build darwin

package ieproxy

import (
	"net/http"
	"reflect"
	"testing"
)

func TestPacfile(t *testing.T) {
	listener, err := listenAndServeWithClose("127.0.0.1:0", http.FileServer(http.Dir("pacfile_examples")))
	serverBase := "http://" + listener.Addr().String() + "/"
	if err != nil {
		t.Fatal(err)
	}

	// test inactive proxy
	proxy := ProxyScriptConf{
		Active:           false,
		PreConfiguredURL: serverBase + "simple.pac",
	}
	out := proxy.FindProxyForURL("http://google.com")
	if out != "" {
		t.Error("Got: ", out, "Expected: ", "")
	}
	proxy.Active = true

	pacSet := []struct {
		pacfile  string
		url      string
		expected string
	}{
		{
			serverBase + "direct.pac",
			"http://google.com",
			"",
		},
		{
			serverBase + "404.pac",
			"http://google.com",
			"",
		},
		{
			serverBase + "simple.pac",
			"http://google.com",
			"127.0.0.1:8",
		},
		{
			serverBase + "simple.pac",
			"https://google.com",
			"127.0.0.1:8",
		},
		{
			serverBase + "multiple.pac",
			"http://google.com",
			"127.0.0.1:8081",
		},
		{
			serverBase + "except.pac",
			"http://imgur.com",
			"localhost:9999",
		},
		{
			serverBase + "except.pac",
			"http://example.com",
			"",
		},
		{
			"",
			"http://example.com",
			"",
		},
		{
			" ",
			"http://example.com",
			"",
		},
		{
			"wrong_format",
			"http://example.com",
			"",
		},
	}
	for _, p := range pacSet {
		proxy.PreConfiguredURL = p.pacfile
		out := proxy.FindProxyForURL(p.url)
		if out != p.expected {
			t.Error("Got: ", out, "Expected: ", p.expected)
		}
	}
	listener.Close()
}

var multipleMap map[string]string

func init() {
	multipleMap = make(map[string]string)
	multipleMap["http"] = "127.0.0.1"
	multipleMap["ftp"] = "128"
}

func TestOverrideEnv(t *testing.T) {
	var callStack []string
	pseudoSetEnv := func(key, value string) error {
		if value != "" {
			callStack = append(callStack, key)
			callStack = append(callStack, value)
		}
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
					Protocols: multipleMap,
				},
			},
			callStack: []string{"http_proxy", "127.0.0.1"},
		},
		{
			in: ProxyConf{
				Static: StaticProxyConf{
					Active:  true,
					NoProxy: "example.com,microsoft.com",
				},
			},
			callStack: []string{"no_proxy", "example.com,microsoft.com"},
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

func TestParseConf(t *testing.T) {
	input := `<dictionary> {
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
  ProxyAutoConfigURLString : https://example.com/foo.pac
}`
	parseConf([]byte(input))
	if darwinProxyConf.Static.Active != true {
		t.Error("darwinProxyConf.Static.Active is not true")
	}
	if darwinProxyConf.Static.Protocols["http"] != "example.jp:8081" {
		t.Error("http proxy does not match")
	}
	if darwinProxyConf.Static.Protocols["https"] != "example.com:8080" {
		t.Error("https proxy does not match")
	}
	if darwinProxyConf.Automatic.Active != true {
		t.Error("darwinProxyConf.Automatic.Active is not true")
	}
	if darwinProxyConf.Automatic.PreConfiguredURL != "https://example.com/foo.pac" {
		t.Error("PreConfiguredURL does not match")
	}
	if darwinProxyConf.Static.NoProxy != "*.local,169.254/16" {
		t.Error("NoProxy does not match")
	}
}
