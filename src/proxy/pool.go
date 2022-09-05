package proxy

import (
	"errors"
	"net/url"
	"sync"

	"github.com/ocean233/go-crawer/src/logger"
)

var (
	proxyPool      map[*url.URL]bool
	proxyPoolLock  sync.Mutex
	proxyPortBegin int = 11000
	proxyPortEnd   int = 11010
)

func init() {
	startV2rays()
	proxyPoolLock.Lock()
	defer proxyPoolLock.Unlock()
	for port := proxyPortBegin; port < proxyPortEnd; port++ {
		proxyUrl, err := getLocalProxyURL(port)
		if err != nil {
			logger.Log.Warnf("get proxy url error at port %d, err: %v", port, err)
			continue
		}
		proxyPool[proxyUrl] = true
	}
	logger.Log.Infof("init proxy pool done, get %d proxy url", len(proxyPool))
}

func GetAProxyUrl() (*url.URL, error) {
	proxyPoolLock.Lock()
	defer proxyPoolLock.Unlock()
	if len(proxyPool) > 0 {
		for url, useable := range proxyPool {
			if useable {
				proxyPool[url] = false
				return url, nil
			}
		}
	}
	return nil, errors.New("no enough proxy url")
}

func ReturnAProxyUrl(url *url.URL) {
	proxyPoolLock.Lock()
	defer proxyPoolLock.Unlock()
	proxyPool[url] = true
}
