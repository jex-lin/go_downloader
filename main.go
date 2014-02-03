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
    http.HandleFunc("/api/", api.Api)
	http.HandleFunc("/static/", index.Static)
    http.Handle("/progress/", websocket.Handler(index.RespondProgress))
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

