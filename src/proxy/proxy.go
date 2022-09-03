package proxy

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

var Transport *http.Transport

func init() {
	proxyURL, err := getLocalProxyURL(10909)
	if err != nil {
		panic(err)
	}
	Transport = &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
}

func getLocalProxyURL(port int) (*url.URL, error) {
	if port < 65535 && port > 0 {
		return url.Parse("http://127.0.0.1:" + strconv.Itoa(port))
	}
	return nil, errors.New("wrong port number")
}
