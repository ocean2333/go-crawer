package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/ocean233/go-crawer/src/album"
	"github.com/ocean233/go-crawer/src/crawer"
	"github.com/ocean233/go-crawer/src/logger"
)

type ab struct {
	Url, Img, Name string
}

type dt struct {
	Details []string
}

var c = crawer.NewCrawer()

func index(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //解析参数，默认是不会解析的
	logger.Log.Info("path", r.URL.Path)
	for k, v := range r.Form {
		if k == "input" {
			c.AddSearchRequest(v[0])
		}
	}
	if len(r.Form) != 0 {
		http.Redirect(w, r, "/index/", http.StatusFound)
	}
	content, _ := ioutil.ReadFile("../html/index.html")
	w.Write(content)
}

func detail(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //解析参数，默认是不会解析的
	logger.Log.Info("path", r.URL.Path)
	jobs := c.GetDownloadingJob()
	t, err := template.ParseFiles("../html/detail.html")
	if err != nil {
		logger.Log.Error(err)
		t, _ = template.ParseFiles("../html/5xx.html")
		t.Execute(w, nil)
		return
	}
	t.Execute(w, &dt{Details: jobs})
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/login/index", http.StatusFound)
	}

	t, err := template.ParseFiles("../html/404.html")
	if err != nil {
		log.Println(err)
	}
	t.Execute(w, nil)

}

func selector(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //解析参数，默认是不会解析的
	logger.Log.Info("path", r.URL.Path)
	jobName := r.FormValue("job_name")
	if jobName != "" {
		var URL = ""
		// TODO: multi page show
		list := album.GetAPageAlbums(0)
		for _, album := range list {
			if album.Name == jobName {
				URL = album.URL
				continue
			}
		}
		c.AddDownloadRequest(URL)
	}
	t, err := template.ParseFiles("../html/selector.html")
	if err != nil {
		logger.Log.Error(err)
		return
	}
	if len(r.Form) != 0 {
		http.Redirect(w, r, "/selector/", http.StatusFound)
	}
	var abs = make([]ab, 0)
	list := album.GetAPageAlbums(0)
	for _, a := range list {
		abs = append(abs, ab{
			Url:  a.URL,
			Img:  a.Img,
			Name: a.Name,
		})
	}
	t.Execute(w, struct {
		Albums []ab
	}{abs})
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		logger.Log.Infof("path %s", r.URL.Path)
		r.ParseForm()
		url := r.FormValue("album_url")
		if url != "" {
			c.AddDownloadRequest(url)
		}
	}
}

func main() {
	c.Start()
	http.HandleFunc("/index/", index) //设置访问的路由
	http.HandleFunc("/detail.html", detail)
	http.HandleFunc("/selector/", selector)
	http.HandleFunc("/download/", downloadHandler)
	http.HandleFunc("/", index)
	err := http.ListenAndServe(":9090", nil) //设置监听的端口
	if err != nil {
		logger.Log.Fatal("ListenAndServe: ", err)
	}
}
