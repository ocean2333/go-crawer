package download

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/ocean233/go-crawer/src/config"
	"github.com/ocean233/go-crawer/src/logger"
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
			client := http.Client{
				// seems zip file can be downloaded without proxy
				// Transport: proxy.Transport,
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
