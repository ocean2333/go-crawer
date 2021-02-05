package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/crawer"
)

type ab struct {
	Url, Img, Name string
}

type dt struct {
	Details []string
}

var c = crawer.NewCrawer()

func sayhelloName(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()       //解析参数，默认是不会解析的
	fmt.Println(r.Form) //这些信息是输出到服务器端的打印信息
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	// if r.Form["input"] != nil {
	// 	AddSearchRequest(r.Form["input"])
	// }
}

func index(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //解析参数，默认是不会解析的
	fmt.Println("path", r.URL.Path)
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
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
	fmt.Println("path", r.URL.Path)
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	jobs := c.GetDownloadingJob()
	t, err := template.ParseFiles("../html/detail.html")
	if err != nil {
		fmt.Println(err)
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
	fmt.Println("path", r.URL.Path)
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	jobName := r.FormValue("job_name")
	if jobName != "" {
		var URL = ""
		list := c.GetList()
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
		fmt.Println(err)
		return
	}
	if len(r.Form) != 0 {
		http.Redirect(w, r, "/selector/", http.StatusFound)
	}
	var abs = make([]ab, 0)
	list := c.GetList()
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

func main() {
	c.Start()
	http.HandleFunc("/index/", index) //设置访问的路由
	http.HandleFunc("/detail.html", detail)
	http.HandleFunc("/selector/", selector)
	http.HandleFunc("/", NotFoundHandler)
	err := http.ListenAndServe(":9090", nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
