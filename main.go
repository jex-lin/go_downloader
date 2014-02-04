package main

import (
    "code.google.com/p/go.net/websocket"
	"go_downloader/controller/index"
	"net/http"
	"log"
)

func main() {
	http.HandleFunc("/", index.Home)
    http.Handle("/download/", websocket.Handler(index.Download))
	http.HandleFunc("/static/", index.Static)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
