package index

import (
    "html/template"
    "net/http"
    "strings"
    "path/filepath"
    "go_downloader/model/osmod"
    "go_downloader/model/download"
    "code.google.com/p/go.net/websocket"
    "fmt"
    "os"
    "strconv"
    "os/exec"
    "runtime"
    "encoding/json"
    "bytes"
)

// Static file (img, js, css)
func Static(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

func Home(w http.ResponseWriter, r *http.Request) {

    var data = map[string] interface{}{}

    // Receive post
    r.ParseForm()
    if r.Method == "POST" {
        // shutdown
        shutdown := strings.TrimSpace(r.FormValue("shutdown"))
        if shutdown != "" {
            shutdownValue, _ := strconv.Atoi(shutdown)
            if shutdownValue == 1 {
                os.Exit(0)
            }
        }

        storagePath := strings.TrimSpace(r.FormValue("storagePath"))
        ffmpegPath := strings.TrimSpace(r.FormValue("ffmpegPath"))
        if storagePath != "" {
            storagePath = filepath.Clean(storagePath)
            data["storagePath"] = storagePath
            if osmod.SetStoragePath(storagePath) {
                data["checkStoragePath"] = true
            } else {
                data["checkStoragePath"] = false
            }
        }
        if ffmpegPath != "" {
            ffmpegPath = filepath.Clean(ffmpegPath)
            data["ffmpegPath"] = ffmpegPath
            if osmod.FileExists(ffmpegPath) {
                data["checkFFmpegPath"] = true
            } else {
                data["checkFFmpegPath"] = false
            }
        }
    } else {
        // Default data
        currentPath, err := os.Getwd()
        if err != nil {
            fmt.Println(err)
            currentPath = ""
        }
        if currentPath != "" {
            // Storage path
            osmod.SetStoragePath(currentPath)
            data["checkStoragePath"] = true
            data["storagePath"] = currentPath
            // ffplay path
            ffplayPath := currentPath + string(os.PathSeparator) + "ffplay.exe"
            data["checkFFmpegPath"] = true
            data["ffmpegPath"] = ffplayPath
        }
    }

    if runtime.GOOS == "windows" {
        data["isWindows"] = true
    }

    // Show view
    var tmplPath string = "view/template/"
    var indexPath string = "view/index/"
    t, _ := template.ParseFiles(
        tmplPath + "header.tmpl",
        indexPath + "body.html",
        tmplPath + "index/urlItem.tmpl",
        tmplPath + "footer.tmpl",
    )

    // For loop url item
    var tmplBuf bytes.Buffer
    var nums = map[string] interface{}{}
    for num := 1; num <= 10; num++ {
        nums["num"] = num
        t.ExecuteTemplate(&tmplBuf, "urlItem", nums)
    }
    data["urlItem"] = template.HTML(tmplBuf.String())

    t.ExecuteTemplate(w, "body", data)
	t.Execute(w, nil)
}

func Download(ws *websocket.Conn) {

    var err error
    var rec download.UrlData
    var file download.File
    ch := make(chan download.File)

    // Full CPU Running
    runtime.GOMAXPROCS(runtime.NumCPU())

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

        go download.DownloadFile(rec.Url, storagePath, ws, &rec, ch);
        file = <-ch
        if  file.Err != nil {
            rec.Status = "fail"
            rec.ErrMsg = file.Err.Error()
            os.Remove(file.Path)
            fmt.Println(file.Msg)
        } else {
            // Success
            rec.Status = "ok"
            rec.FilePath = file.UrlData.FilePath
            fmt.Println(file.Msg)
        }

        if err = websocket.JSON.Send(ws, rec); err != nil {
            fmt.Println("Can't send")
            break
        }
    }
}

func PlayVideo(w http.ResponseWriter, r *http.Request) {
    output := map[string] interface{} {}

    // Receive post
    r.ParseForm()
    if r.Method == "POST" {
        ffmpegPath := filepath.Clean(strings.TrimSpace(r.FormValue("FFmpegPath")))
        filePath := filepath.Clean(strings.TrimSpace(r.FormValue("FilePath")))

        if ! osmod.FileExists(ffmpegPath) || ! osmod.FileExists(filePath) {
            output["Status"] = "fail"
            output["ErrMsg"] = "FFmpeg path or file path doesn't exist."
            outputJSON, _ := json.Marshal(output)
            fmt.Fprintf(w, string(outputJSON))
            return
        }

        cmd := exec.Command(ffmpegPath, filePath)
        err := cmd.Run()
        if err == nil {
            output["Status"] = "ok"
        } else {
            output["Status"] = "fail"
            output["ErrMsg"] = err.Error()
        }
        outputJSON, _ := json.Marshal(output);
        fmt.Fprintf(w, string(outputJSON))
    }
}
