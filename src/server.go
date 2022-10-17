package main

import (
	"net/http"

	"github.com/ocean2333/go-crawer/src/handler"
	"github.com/ocean2333/go-crawer/src/logger"
)

func main() {

	// new backend api router
	http.HandleFunc("/album/download/", handler.DownloadHandler)
	http.HandleFunc("/album/homepage/", handler.HomepageHandler)
	http.HandleFunc("/album/search_page/", handler.SearchPageHandler)
	err := http.ListenAndServe(":10320", nil)
	if err != nil {
		logger.Log.Error("ListenAndServe: ", err)
	}
}
