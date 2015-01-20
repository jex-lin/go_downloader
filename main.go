package main

import (
	"code.google.com/p/go.net/websocket"
	"go_downloader/controller/index"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", index.Home)
	http.Handle("/download/", websocket.Handler(index.Download))
	http.HandleFunc("/playVideo/", index.PlayVideo)
	http.HandleFunc("/static/", index.Static)
	err := http.ListenAndServe(":9111", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
