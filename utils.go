package ieproxy

import (
	"log"
	"net"
	"net/http"
	"unicode/utf16"
	"unsafe"
)

// StringFromUTF16Ptr converts a *uint16 C string to a Go String
func StringFromUTF16Ptr(s *uint16) string {
	if s == nil {
		return ""
	}

	p := (*[1<<30 - 1]uint16)(unsafe.Pointer(s))

	// find the string length
	sz := 0
	for p[sz] != 0 {
		sz++
	}

	return string(utf16.Decode(p[:sz:sz]))
}

// for testing purposes
func listenAndServeWithClose(addr string, handler http.Handler) (net.Listener, error) {

	var (
		listener net.Listener
		err      error
	)

	srv := &http.Server{Addr: addr, Handler: handler}

	if addr == "" {
		addr = ":http"
	}

	listener, err = net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	go func() {
		err := srv.Serve(listener.(*net.TCPListener))
		if err != nil {
			log.Println("HTTP Server Error - ", err)
		}
	}()

	return listener, nil
}
