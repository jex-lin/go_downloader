package main

import (
    "code.google.com/p/go.net/websocket"
	"go_downloader/controller/index"
	"go_downloader/controller/api"
	"net/http"
	"log"
)

func main() {
	http.HandleFunc("/", index.Home)
    http.Handle("/api/", websocket.Handler(api.Api))
	http.HandleFunc("/static/", index.Static)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
