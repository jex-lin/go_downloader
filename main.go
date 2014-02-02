package main

import (
    "code.google.com/p/go.net/websocket"
	"go_downloader/source/web"
	"net/http"
	"log"
)

func main() {
	http.HandleFunc("/", web.Home)
    http.HandleFunc("/api/", web.Api)
	http.HandleFunc("/static/", web.Static)
    http.Handle("/echo/", websocket.Handler(web.Echo))
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

