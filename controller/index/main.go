package index

import (
    "html/template"
    "net/http"
    "strings"
    "path/filepath"
    "go_downloader/model/osmod"
    "go_downloader/library/download"
    "code.google.com/p/go.net/websocket"
    "fmt"
)

// Static file (img, js, css)
func Static(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

func Home(w http.ResponseWriter, r *http.Request) {
    // Prepare data
    var data = map[string] interface{}{}

    // Receive post
    r.ParseForm()
    if r.Method == "POST" {
        storagePath := strings.TrimSpace(r.FormValue("storagePath"))
        storagePath = filepath.Clean(storagePath)
        data["storagePath"] = storagePath
        if osmod.SetStoragePath(storagePath) {
            data["checkPathMsg"] = true
        } else {
            data["checkPathMsg"] = false
        }
    }

    // Show view
    var tmplPath string = "view/template/"
    var indexPath string = "view/index/"
    t, _ := template.ParseFiles(
        tmplPath + "header.tmpl",
        indexPath + "body.html",
        tmplPath + "footer.tmpl",
    )
    t.ExecuteTemplate(w, "body", data)
	t.Execute(w, nil)
}

func Download(ws *websocket.Conn) {

    var err error
    var rec download.UrlData

    for {
        err = websocket.JSON.Receive(ws, &rec)
        if err != nil {
            var reply download.UrlData
            reply.Status = "fail"
            reply.ErrMsg = "Not JSON format"
            websocket.JSON.Send(ws, reply)
            break
        }

        storagePath, err2 := osmod.GetStoragePath()
        if err2 != nil {
            rec.Status = "fail"
            rec.ErrMsg = "Storage path doesn't exist."
            websocket.JSON.Send(ws, rec)
            break
        }

        err = download.DownloadFile(rec.Url, storagePath, ws, &rec);
        if  err != nil {
            rec.Status = "fail"
            rec.ErrMsg = err.Error()
        } else {
            // Success
            rec.Status = "ok"
        }

        if err = websocket.JSON.Send(ws, rec); err != nil {
            fmt.Println("Can't send")
            break
        }
    }

    //output := map[string] interface{} {}

    //storagePath, err := osmod.GetStoragePath()
    //if err != nil {
    //    output["status"] = "fail"
    //    output["errMsg"] = err.Error()
    //    outputJson, _ := json.Marshal(output);
    //    fmt.Fprintf(w, string(outputJson))
    //    return
    //}
    //// Receive post
    //r.ParseForm()
    //if r.Method == "POST" {
    //    url := strings.TrimSpace(r.FormValue("url"))
    //    err := download.DownloadFile(url, storagePath);
    //    if  err != nil {
    //        output["status"] = "fail"
    //        output["errMsg"] = err.Error()
    //        outputJson, _ := json.Marshal(output);
    //        fmt.Fprintf(w, string(outputJson))
    //        return
    //    }
    //    output["status"] = "ok"
    //    outputJson, _ := json.Marshal(output);
    //    fmt.Fprintf(w, string(outputJson))
    //}

}
//func Api(w http.ResponseWriter, r *http.Request) {
//
//    output := map[string] interface{} {}
//
//    storagePath, err := osmod.GetStoragePath()
//    if err != nil {
//        output["status"] = "fail"
//        output["errMsg"] = err.Error()
//        outputJson, _ := json.Marshal(output);
//        fmt.Fprintf(w, string(outputJson))
//        return
//    }
//    // Receive post
//    r.ParseForm()
//    if r.Method == "POST" {
//        url := strings.TrimSpace(r.FormValue("url"))
//        err := download.DownloadFile(url, storagePath);
//        if  err != nil {
//            output["status"] = "fail"
//            output["errMsg"] = err.Error()
//            outputJson, _ := json.Marshal(output);
//            fmt.Fprintf(w, string(outputJson))
//            return
//        }
//        output["status"] = "ok"
//        outputJson, _ := json.Marshal(output);
//        fmt.Fprintf(w, string(outputJson))
//    }
//}

//dec := json.NewDecoder(strings.NewReader(jsonStream)) 
//dec.Decode(&m);

//add custom http header
//
//client := &amp;http.Client{]
//req, err := http.NewRequest("POST", "http://example.com", bytes.NewReader(postData))
//req.Header.Add("User-Agent", "myClient")
//resp, err := client.Do(req)
//defer resp.Body.Close()