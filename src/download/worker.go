package download

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/ocean233/go-crawer/src/config"
	"github.com/ocean233/go-crawer/src/logger"
	"github.com/ocean233/go-crawer/src/proxy"
)

type downloadWorker struct {
	jobChan   chan string
	sleepTime uint64
	resChan   chan string
}

func (w *downloadWorker) work() {
	for url := range w.jobChan {
		if check(url) {
			logger.Log.Infof("download start: %s", url)
			handler.registJob(url, w)
			var client http.Client
			proxy, err := proxy.GetAProxyUrl()
			if err != nil {
				logger.Log.Warnf("worker get proxy url err, failed to work")
				logger.Log.Warnf("send url %s back", url)

				continue
			}
			client = http.Client{
				Transport: &http.Transport{
					Proxy:           http.ProxyURL(proxy),
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
				Timeout: 1000 * time.Second,
			}

			nameRegExp, _ := regexp.Compile("n=.+")
			fileName := nameRegExp.FindString(url)

			resp, err := client.Get(url)
			if err != nil {
				logger.Log.Errorf("get %s Errorf, err: %v", err)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				logger.Log.Errorf("get %s, response %s", url, resp.Status)
				continue
			}
			content, _ := ioutil.ReadAll(resp.Body)
			err = ioutil.WriteFile(path.Join(config.Get().SavePath, fileName[2:]+".zip"), content, os.ModeAppend)
			if err != nil {
				return
			}
			handler.doneJob(url)
			logger.Log.Infof("download end: %s", url)
			resp.Body.Close()
			time.Sleep(time.Duration(w.sleepTime) * time.Second)
		}
	}
}

// TODO: check the url is zip file
func check(url string) bool {
	return true
}
