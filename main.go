package main

import (
	"go_downloader/source/download"
	"go_downloader/source/web"
	"log"
	"net/http"
	"html/template"
    "fmt"
)

// Static file (img, js, css)
func staticHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    var tmplPath string = "view/template/"
    var indexPath string = "view/index/"
    t, _ := template.ParseFiles(
        tmplPath + "header.tmpl",
        indexPath + "body.html",
        tmplPath + "footer.tmpl",
    )
    var data = map[string] interface{}{
        "content" : "Do you copy?",
    }
    t.ExecuteTemplate(w, "body", data)
	t.Execute(w, nil)
}
	// Create file
	//dest, err := os.Create("C:\\Go\\mygo\\src\\go_downloader\\qq.txt")
	//if err != nil {
	//	log.Fatal("create file error")
	//}
	//defer dest.Close()

	//r.ParseForm() //解析參數，默認是不會解析的
	//fmt.Println(r.Form) //這些信息是輸出到服務器端的打印信息
	//fmt.Println("path", r.URL.Path)
	//fmt.Println("scheme", r.URL.Scheme)
	//fmt.Println(r.Form["url_long"])
	//for k, v := range r.Form {
	//    fmt.Println("key:", k)
	//    fmt.Println("val:", strings.Join(v, ""))
	//}
	//fmt.Fprintf(w, "Hello astaxie!") //這個寫入到w的是輸出到客戶端的


func main() {
	http.HandleFunc("/", homeHandler)
    http.HandleFunc("/os/", osHandler)
	http.HandleFunc("/static/", staticHandler)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func osHandler(w http.ResponseWriter, r *http.Request) {
    if web.SetStoragePath("/Users/apple/Desktop/ttt") {
        fmt.Fprintf(w, "Exist")
    } else {
        fmt.Fprintf(w, "Doesn't exist")
    }
    path, err := web.GetStoragePath()
    if err == nil {
        fmt.Fprintf(w, "    " + path)
    } else {
        fmt.Fprintf(w, "    " + err.Error())
    }


	// Urls
	urlList := []string{
		//"https://calibre-ebook.googlecode.com/files/eight-demo.flv",
		//"http://www.paulgu.com/w/images/f/f0/Honda_accord.flv",
		//"http://vault.futurama.sk/joomla/media/video/video2.flv",
		//"http://video.disclose.tv/12/69/demo_video_13_FLV_126943.flv",
	}
    if err := download.DownloadFiles(urlList); err != nil {
        //fmt.Fprintf(w, err.Error())
    }
}

