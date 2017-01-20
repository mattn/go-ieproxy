package ieproxy

import (
	"net/http"
	"net/url"
)

func ProxyFromIE(req *http.Request) (*url.URL, error) {
	return http.ProxyFromEnvironment(req)
}
