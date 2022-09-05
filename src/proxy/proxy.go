package proxy

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ocean233/go-crawer/src/logger"
)

var Transport *http.Transport

func init() {
	proxyUrl, err := getLocalProxyURL(10909)
	if err != nil {
		panic(err)
	}
	Transport = &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}
}

func testProxy(port int) bool {
	proxy, _ := url.Parse("http://127.0.0.1:" + strconv.Itoa(port))
	client := http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 1000 * time.Second,
	}

	_, err := client.Get("google.com")
	if err != nil {
		logger.Log.Warnf("proxy at port %d not usable", port)
		return false
	}
	return true
}

func getLocalProxyURL(port int) (*url.URL, error) {
	if port < 65535 && port > 0 {
		if testProxy(port) {
			return url.Parse("http://127.0.0.1:" + strconv.Itoa(port))
		}
	}
	return nil, errors.New("wrong port number")
}

func startV2rays() error {
	for port := proxyPortBegin; port < proxyPortEnd; port++ {
		err := execV2ray(port)
		if err != nil {
			return errors.New("failed to exec v2ray")
		}
	}
	return nil
}

func execV2ray(port int) error {
	// do something
	// syscall.Exec("v2ray.exe")
	return nil
}
