package main

import (
	"go_downloader/source/web"
	"net/http"
	"log"
)

func main() {
	http.HandleFunc("/", web.Home)
    http.HandleFunc("/api/", web.Api)
	http.HandleFunc("/static/", web.Static)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

