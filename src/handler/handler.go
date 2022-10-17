package handler

import (
	"encoding/json"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/ocean2333/go-crawer/src/logger"
	"github.com/ocean2333/go-crawer/src/model"
	"github.com/ocean2333/go-crawer/src/net"
	"github.com/ocean2333/go-crawer/src/parser"
)

// download handler receive two args: rid and aid (rule id and album id), use rid to find rule, use aid to find album, it will automatically download all the images in the album
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		logger.Log.Infof("path %s", r.URL.Path)
		r.ParseForm()
		rid := r.FormValue("rid")
		aid := r.FormValue("aid")
		titles := model.AdminTitlesResponse{
			Titles: make([]string, 0),
		}
		recorded := make(map[string]struct{})
		for page := 1; ; page++ {
			// should ignore albums here
			_, pics, err := parser.ParseAlbumPage(rid, aid, page)
			if err != nil {
				logger.Log.Errorf("parse album page error: %v", err)
				break
			}

			hasNewMetaData := false
			// for _, album := range albums {
			// 	if _, ok := recorded[album.Aid]; !ok {
			// 		titles.Titles = append(titles.Titles, album.Title)
			// 		recorded[album.Aid] = struct{}{}
			// 		hasNewMetaData = true
			// 	}
			// }
			for _, pic := range pics {
				if _, ok := recorded[pic.Pid]; !ok {
					titles.Titles = append(titles.Titles, pic.Pid)
					recorded[pic.Pid] = struct{}{}
					hasNewMetaData = true
				}
			}
			if !hasNewMetaData {
				break
			}
		}
		data, err := json.Marshal(titles)
		if err != nil {
			logger.Log.Errorf("json marshal titles err: %v", err)
			return
		}
		w.Write(data)
	}
}

// home page handler receive one arg: rid (rule id), it will automatically load all the albums' metadata and cover thumbnail in the home page
func HomepageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		logger.Log.Infof("path %s", r.URL.Path)
		r.ParseForm()
		rid := r.FormValue("rid")
		if rid != "" {
			rule := parser.GetRule(rid)
			patten := map[string]string{
				"pid": rid,
			}
			url := parser.RepalcePatten(patten, rule.HomepageUrl)
			resp, err := net.GetWithDefaultProxy(url)
			if err != nil {
				logger.Log.Errorf("get %s err: %v", url, err)
				return
			}
			if resp.StatusCode != 200 {
				logger.Log.Errorf("get %s err: %v", url, resp.StatusCode)
				return
			}
			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				logger.Log.Errorf("new document err: %v", err)
				return
			}
			// should ignore pictures here
			albums, _ := parser.ParseHtml(rid, "", rule.HomepageRules, doc)
			titles := make([]string, 0)
			for _, album := range albums {
				titles = append(titles, album.Title)
				go net.GetWorkerPoolManager().GetWorkerPool(rid).GetThumbnailWithProxyAndSave(album.Cover, rid, album.Aid, "cover")
			}
			respStruct := model.AdminTitlesResponse{
				Titles: titles,
			}
			data, err := json.Marshal(respStruct)
			if err != nil {
				logger.Log.Errorf("json marshal titles err: %v", err)
				return
			}
			w.Write(data)
		}
	}
}

func SearchPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		logger.Log.Infof("path %s", r.URL.Path)
		r.ParseForm()
		rid := r.FormValue("rid")
		keywords := r.FormValue("keywords")
		page := r.FormValue("page")
		if rid != "" {
			rule := parser.GetRule(rid)
			patten := map[string]string{
				"rid":      rid,
				"keywords": keywords,
				"page":     page,
			}
			url := parser.RepalcePatten(patten, rule.SearchUrl)
			resp, err := net.GetWithDefaultProxy(url)
			if err != nil {
				logger.Log.Errorf("get %s err: %v", url, err)
				return
			}
			if resp.StatusCode != 200 {
				logger.Log.Errorf("get %s err: %v", url, resp.StatusCode)
				return
			}
			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				logger.Log.Errorf("new document err: %v", err)
				return
			}
			// should ignore pictures here
			albums, _ := parser.ParseHtml(rid, "", rule.SearchRules, doc)
			titles := make([]string, 0)
			for _, album := range albums {
				titles = append(titles, album.Title)
				go net.GetWorkerPoolManager().GetWorkerPool(rid).GetThumbnailWithProxyAndSave(album.Cover, rid, album.Aid, "cover")
			}
			respStruct := model.AdminTitlesResponse{
				Titles: titles,
			}
			data, err := json.Marshal(respStruct)
			if err != nil {
				logger.Log.Errorf("json marshal titles err: %v", err)
				return
			}
			w.Write(data)
		}
	}
}

func CategoriesHandler(w http.ResponseWriter, r *http.Request) {}
