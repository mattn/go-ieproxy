package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	_ "github.com/mattn/go-ieproxy/global"
)

func main() {
	os.Setenv("HTTP_PROXY", "")
	os.Setenv("HTTPS_PROXY", "")
	res, err := http.Get("http://www.google.com")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer res.Body.Close()
	io.Copy(os.Stdout, res.Body)
}
