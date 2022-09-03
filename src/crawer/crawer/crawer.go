package crawer

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jackdanger/collectlinks"
	"github.com/ocean233/go-crawer/src/album"
	"github.com/ocean233/go-crawer/src/download"
	"github.com/ocean233/go-crawer/src/logger"
	"github.com/ocean233/go-crawer/src/proxy"
)

var (
	handler *download.DownloadHandler
	manager *album.AlbumManager
)

func init() {
	handler = download.GetHandler()
	manager = album.GetAlbumManager()
	handler.Start()
}

type Crawer struct {
	searchChan, pageChan, albumChan, downloadChan chan string
	transport                                     *http.Transport
}

func NewCrawer() *Crawer {
	return &Crawer{
		searchChan:   make(chan string, 10),
		pageChan:     make(chan string, 10),
		albumChan:    make(chan string, 10),
		downloadChan: make(chan string, 10),
		transport:    proxy.Transport,
	}
}

// Start 启动爬虫服务
func (c *Crawer) Start() {
	go c.searchDealer()
	go c.pageDealer()
	go c.albumDealer()
	go c.downloadDealer()
}

func (c *Crawer) searchDealer() {
	logger.Log.Infof("searchDealer started")
	client := http.Client{
		Transport: c.transport,
		Timeout:   50 * time.Second,
	}
	for words := range c.searchChan {
		logger.Log.Infof("get search words: " + words)
		url := "https://wnacg.net/search/?q=" + words + "&s=create_time_DESC"
		resp, err := client.Get(url)
		if err != nil {
			logger.Log.Errorf("Errorf while get %s, err: %v", url, err)
			continue
		}
		if resp.StatusCode != 200 {
			logger.Log.Errorf("get %s err, return code: %s", url, resp.Status)
			continue
		}
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		defer resp.Body.Close()
		if doc == nil {
			continue
		}
		albumNum := doc.Find("p[class=result]").Find("b").Eq(0).Text()
		if albumNum == "0" {
			continue
		}
		albumNumInt, err := strconv.Atoi(strings.Replace(albumNum, ",", "", -1))
		if err != nil {
			logger.Log.Errorf("albumNum convert err: %v", err)
			continue
		}
		albumPageNum := albumNumInt / 16
		if albumNumInt%16 != 0 {
			albumPageNum++
		}

		for i := 1; i <= albumPageNum; i++ {
			link := "https://wnacg.net/search/?q=" + words + "&s=create_time_DESC" + "&p=" + strconv.Itoa(i)
			c.pageChan <- link
		}

		runtime.Gosched()
	}
	runtime.Goexit()
}

func (c *Crawer) pageDealer() {
	logger.Log.Infof("pageDealer started")
	client := http.Client{
		Transport: c.transport,
		Timeout:   10 * time.Second,
	}

	for pageURL := range c.pageChan {
		resp, err := client.Get(pageURL)
		if err != nil {
			logger.Log.Errorf("get %s err: %v", pageURL, err)
			continue
		}
		links := collectlinks.All(resp.Body)
		for _, link := range links {
			matched, err := regexp.MatchString("photos-index-aid-\\d+\\.html", link)
			if err != nil {
				logger.Log.Errorf("match string Errorf, err: %v", err)
			}
			if matched {
				c.albumChan <- link
			}
		}
		_ = resp.Body.Close()
		runtime.Gosched()
	}
}

func (c *Crawer) albumDealer() {
	logger.Log.Infof("albumDealer started")
	domainName := "https://wnacg.net"
	client := http.Client{
		Transport: c.transport,
		Timeout:   10 * time.Second,
	}
	for albumLink := range c.albumChan {
		resp, err := client.Get(domainName + albumLink)
		if err != nil {
			logger.Log.Errorf("Errorf in get Album: %s", albumLink)
			continue
		}
		if resp.StatusCode != 200 {
			logger.Log.Error("get " + albumLink + " response" + resp.Status)
			continue
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Log.Errorf("read resp body err: %v", err)
		}
		//links := collectlinks.All(bytes.NewReader(b))
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(b))
		if err != nil {
			logger.Log.Error(err)
			continue
		}
		name := doc.Find("div.userwrap > h2").Eq(0).Text()
		img, e := doc.Find("div.asTBcell.uwthumb > img").Eq(0).Attr("src")
		if !e {
			img = "../img/img.png"
		}
		logger.Log.Infof("get albumLink: %s, name: %s", albumLink, name)
		link := strings.Replace(albumLink, "photos", "download", 1)
		album.Add(&album.Album{
			URL:  link,
			Name: name,
			Img:  img,
		})
		c.AddDownloadRequest(link)
		resp.Body.Close()

		runtime.Gosched()
	}
	close(c.downloadChan)
	runtime.Goexit()
}

func (c *Crawer) downloadDealer() {
	logger.Log.Info("downloadDealer started")
	domainName := "https://wnacg.net"
	client := http.Client{
		Transport: c.transport,
		Timeout:   10 * time.Second,
	}
	for downloadLink := range c.downloadChan {
		logger.Log.Info("get " + downloadLink)
		if !strings.Contains(downloadLink, domainName) {
			downloadLink = domainName + downloadLink
		}
		resp, err := client.Get(downloadLink)
		if err != nil {
			logger.Log.Error(err)
			continue
		}
		if resp.StatusCode != 200 {
			logger.Log.Error("get " + downloadLink + " response" + resp.Status)
			continue
		}
		links := collectlinks.All(resp.Body)
		for _, link := range links {
			matched := strings.Contains(link, "d7.wzip.ru")
			if matched {
				handler.AddJob("https:" + link)
				//go c.download("https:" + link)
				break
			}
		}
		_ = resp.Body.Close()
	}
	return
}

// GetDownloadingJob 查看正在下载的任务
func (c *Crawer) GetDownloadingJob() []string {
	res := make([]string, 0)
	for _, job := range handler.DoingJobs() {
		if job != "" {
			res = append(res, job)
		}
	}
	return res
}

// AddSearchRequest 添加搜索需求
func (c *Crawer) AddSearchRequest(keyWords string) {
	logger.Log.Infof("Get " + keyWords)
	c.searchChan <- keyWords
}

// AddDownloadRequest 添加下载需求
func (c *Crawer) AddDownloadRequest(url string) {
	c.downloadChan <- url
	album.Delete(url)
}
