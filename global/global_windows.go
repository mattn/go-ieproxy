package global

import (
	"github.com/mattn/go-ieproxy"
	"net/http"
)

func init() {
	http.DefaultTransport.(*http.Transport).Proxy = ieproxy.ProxyFromIE
}
