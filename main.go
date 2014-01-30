package main

import(
    "go_downloader/source/download"
    "net/http"
    "html/template"
    "log"
	"runtime"
    "strings"
    "fmt"
    "os"
)

var tryCountLimit int = 5

func DownloadFiles() {
	// Full CPU Running
	runtime.GOMAXPROCS(runtime.NumCPU())

	var chReturn download.File
	var files []download.File
	var file download.File

	// Urls
	urlList := []string{
		//"https://calibre-ebook.googlecode.com/files/eight-demo.flv",
        //"http://www.paulgu.com/w/images/f/f0/Honda_accord.flv",
        //"http://vault.futurama.sk/joomla/media/video/video2.flv",
        "http://video.disclose.tv/12/69/demo_video_13_FLV_126943.flv",
	}
	ch := make(chan download.File, len(urlList))
	for _, url := range urlList {
        urlSplit := strings.Split(url, "/")
        file = download.DefaultFile
        file.Url = url
        file.Name = urlSplit[len(urlSplit)-1]
        file.Path = "/tmp/" + file.Name
		files = append(files, file)
		go download.HandleDownload(file, ch)
	}
	chCount := len(urlList)
	for i := 0; i < chCount; i++ {
		chReturn = <-ch
		if chReturn.ConnStatus == false {
			if chReturn.RetryCount < tryCountLimit {
				fmt.Println(chReturn.Msg)
				go download.HandleDownload(chReturn, ch)
				chCount++
			} else {
				fmt.Println(chReturn.Msg)
				fmt.Printf("  **Give up to connect %s\n", chReturn.Name)
			}
		} else {
			fmt.Println(chReturn.Msg)
		}
	}
}

func sayhelloName(w http.ResponseWriter, r *http.Request) {

    t1, err := template.ParseFiles("view/template/header.tmpl", "view/html/body.html", "view/template/footer.tmpl")
    if  err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    t1.ExecuteTemplate(os.Stdout, "header", nil)
    t1.ExecuteTemplate(os.Stdout, "body", nil)
    t1.ExecuteTemplate(os.Stdout, "footer", nil)
    err = t1.Execute(os.Stdout, "dd")
    if err != nil {
        fmt.Fprintf(w, "fff")
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
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "<html><head></head><body><h1>Welcome Home!</h1><a href=\"/static/img/go.png\">Show Image!</a></body></html>")
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, r.URL.Path[1:])
}


func main() {
    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/static/", staticHandler)
    http.HandleFunc("/admin/", adminHandler)
    err := http.ListenAndServe(":9090", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
