// +build windows

package ieproxy

import (
	"reflect"
	"testing"
)

func TestParseRegedit(t *testing.T) {

	parsingSet := []struct {
		in  regeditValues
		out ProxyConf
	}{
		{
			in: regeditValues{
				ProxyServer:   "",
				ProxyOverride: "",
				ProxyEnable:   0,
				AutoConfigURL: "",
			},
			out: ProxyConf{},
		},
	}

	for _, p := range parsingSet {
		out := parseRegedit(p.in)
		if !reflect.DeepEqual(p.out, out) {
			t.Error("Got: ", out, "Expected: ", p.out)
		}
	}
}
