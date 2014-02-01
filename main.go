package main

import (
	"go_downloader/source/web"
	"go_downloader/source/os"
	"net/http"
	"log"
)

func main() {
	http.HandleFunc("/", web.Home)
    http.HandleFunc("/os/", os.Os)
	http.HandleFunc("/static/", web.Static)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

