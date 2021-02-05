package crawer

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jackdanger/collectlinks"
)

// Album url和名字
type Album struct {
	URL, Name, Img string
}

type Crawer struct {
	searchChan, pageChan, albumChan, downloadChan chan string
	tr                                            *http.Transport
	albumList                                     []Album
	jobSet                                        map[string]bool
}

func NewCrawer() *Crawer {
	proxyURL, err := getLocalProxyURL(10809)
	if err != nil {
		panic(err)
	}
	return &Crawer{make(chan string, 10), make(chan string, 10),
		make(chan string, 10), make(chan string, 10), &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}, make([]Album, 0), make(map[string]bool)}
}

func getLocalProxyURL(port int) (*url.URL, error) {
	if port < 65535 && port > 0 {
		return url.Parse("http://127.0.0.1:" + strconv.Itoa(port))
	}
	return nil, errors.New("wrong port number")
}

func (c *Crawer) searchDealer() {
	fmt.Println("searchDealer started")
	client := http.Client{
		Transport: c.tr,
		Timeout:   50 * time.Second,
	}
	for words := range c.searchChan {
		fmt.Println("get " + words)
		resp, err := client.Get("https://wnacg.net/search/?q=" + words + "&s=create_time_DESC")
		if err != nil {
			fmt.Println(err)
			continue
		}
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if doc == nil {
			continue
		}
		albumNum := doc.Find("p[class=result]").Find("b").Eq(0).Text()
		if albumNum == "0" {
			continue
		}
		albumNumInt, err := strconv.Atoi(albumNum)
		if err != nil {
			fmt.Println(err)
			continue
		}
		albumPageNum := albumNumInt / 16
		if albumNumInt%16 != 0 {
			albumPageNum++
		}
		if resp.StatusCode != 200 {
			fmt.Println("get " + words + " response" + resp.Status)
		}
		for i := 1; i <= albumPageNum; i++ {
			link := "https://wnacg.net/search/?q=" + words + "&s=create_time_DESC" + "&p=" + strconv.Itoa(i)
			c.pageChan <- link
		}
		_ = resp.Body.Close()
		runtime.Gosched()
	}
	runtime.Goexit()

}

func (c *Crawer) pageDealer() {
	fmt.Println("pageDealer started")
	client := http.Client{
		Transport: c.tr,
		Timeout:   10 * time.Second,
	}

	for pageURL := range c.pageChan {
		resp, err := client.Get(pageURL)
		if err != nil {
			fmt.Println(err)
			continue
		}
		//u, err := url.Parse(pageURL)
		//if err != nil {
		//	fmt.Println(err)
		//	continue
		//}
		//m, _ := url.ParseQuery(u.RawQuery)
		//word := m["q"][0]
		links := collectlinks.All(resp.Body)
		for _, link := range links {
			matched, err := regexp.MatchString("photos-index-aid-\\d+\\.html", link)
			if err != nil {
				fmt.Println("match string error")
			}
			if matched {
				//c.albumList[word] = append(c.albumList[word], Album{URL: link, Name: "?"})
				c.albumChan <- link
			}
		}
		_ = resp.Body.Close()
		runtime.Gosched()
	}

}

func (c *Crawer) albumDealer() {
	fmt.Println("albumDealer started")
	domainName := "https://wnacg.net"
	client := http.Client{
		Transport: c.tr,
		Timeout:   10 * time.Second,
	}
	for albumLink := range c.albumChan {
		fmt.Println("get " + albumLink)
		resp, err := client.Get(domainName + albumLink)
		if err != nil {
			fmt.Println("error in get Album")
			continue
		}
		if resp.StatusCode != 200 {
			fmt.Println("get " + albumLink + " response" + resp.Status)
			continue
		}
		b, err := ioutil.ReadAll(resp.Body)
		links := collectlinks.All(bytes.NewReader(b))
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(b))
		if err != nil {
			fmt.Println(err)
			continue
		}
		name := doc.Find("div.userwrap > h2").Eq(0).Text()
		img, e := doc.Find("div.asTBcell.uwthumb > img").Eq(0).Attr("src")
		if !e {
			img = "../img/img.png"
		}
		fmt.Println("name : " + name)
		for _, link := range links {
			matched, err := regexp.MatchString("download-index-aid-\\d+\\.html", link)
			if err != nil {
				fmt.Println("match string error2")
			}
			if matched {
				c.albumList = append(c.albumList, Album{
					URL:  link,
					Name: name,
					Img:  img,
				})
				//c.downloadChan <- link
			}
		}
		_ = resp.Body.Close()
		runtime.Gosched()
	}
	close(c.downloadChan)
	runtime.Goexit()
}

func (c *Crawer) downloadDealer() {
	fmt.Println("downloadDealer started")
	domainName := "https://wnacg.net"
	client := http.Client{
		Transport: c.tr,
		Timeout:   10 * time.Second,
	}
	for downloadLink := range c.downloadChan {
		fmt.Println("get " + downloadLink)
		resp, err := client.Get(domainName + downloadLink)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if resp.StatusCode != 200 {
			fmt.Println("get " + downloadLink + " response" + resp.Status)
			continue
		}
		links := collectlinks.All(resp.Body)
		for _, link := range links {
			matched, _ := regexp.MatchString("//d\\d+.wnacg.download/down/\\d+/.+\\.zip\\?n=.+", link)
			if matched {
				go c.download("https:" + link)
			}
		}
		_ = resp.Body.Close()
	}
	return
}

func (c *Crawer) download(url string) {
	fmt.Println("start " + url)
	client := http.Client{
		Transport: c.tr,
		Timeout:   1000 * time.Second,
	}
	var fileName string
	nameRegExp, _ := regexp.Compile("n=.+")
	fileName = nameRegExp.FindString(url)
	c.jobSet[fileName[2:]] = true
	resp, err := client.Get(url)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	if resp.StatusCode != 200 {
		fmt.Println("get " + url + " response" + resp.Status)
		return
	}
	content, _ := ioutil.ReadAll(resp.Body)
	err = ioutil.WriteFile(fileName[2:]+".zip", content, os.ModeAppend)
	if err != nil {
		return
	}
	fmt.Println("end " + url)
	delete(c.jobSet, fileName[2:])
	return
}

// Start 启动爬虫服务
func (c *Crawer) Start() {
	go c.searchDealer()
	go c.pageDealer()
	go c.albumDealer()
	go c.downloadDealer()
}

// GetDownloadingJob 查看正在下载的任务
func (c *Crawer) GetDownloadingJob() []string {
	res := make([]string, 0)
	for k := range c.jobSet {
		if k != "" {
			res = append(res, k)
		}
	}
	return res
}

// GetList 获得关键词对应的列表
func (c *Crawer) GetList() []Album {
	return c.albumList
}

// AddSearchRequest 添加搜索需求
func (c *Crawer) AddSearchRequest(keyWords string) {
	fmt.Println("Get " + keyWords)
	c.searchChan <- keyWords
}

// AddDownloadRequest 添加下载需求
func (c *Crawer) AddDownloadRequest(url string) {
	c.downloadChan <- url
	for i, v := range c.albumList {
		if v.URL == url {
			copy(c.albumList[i:], c.albumList[i+1:])
			c.albumList = c.albumList[:len(c.albumList)-1]
		}
	}
}

//func main() {
//	proxyURL, err := getLocalProxyURL(10809)
//	if err != nil {
//		panic(err)
//	}
//	tr = &http.Transport{
//		Proxy:           http.ProxyURL(proxyURL),
//		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
//	}
//
//	go searchDealer()
//	go pageDealer()
//	go albumDealer()
//	go downloadDealer()
//
//	for {
//		fmt.Println("请输入关键词")
//		var words string
//		fmt.Scanf("%s\n", &words)
//		searchChan <- words
//	}
//
//}
