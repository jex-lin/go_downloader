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
            if osmod.FileExists(ffplayPath) {
                data["checkFFmpegPath"] = true
            } else {
                data["checkFFmpegPath"] = false
            }
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
        tmplPath + "index/multiProgress.tmpl",
        tmplPath + "footer.tmpl",
    )

    // For loop url item
    var urlBuf bytes.Buffer
    var progressBuf bytes.Buffer
    var urlItemNums = map[string] interface{}{}
    var progressNums = map[string] interface{}{}
    var progressBarStatus = []string {1: "progress-bar", 2: "progress-bar-warning", 3: "progress-bar-info", 4: "progress-bar-success", 5: "progress-bar-danger"}
    for urlItemNum := 1; urlItemNum <= 10; urlItemNum++ {
        urlItemNums["num"] = urlItemNum
        // For loop multi progress
        for progressNum := 1; progressNum <= int(download.MulSectionDowCount); progressNum++ {
            progressNums["num"] = urlItemNum
            progressNums["partNum"] = progressNum
            progressNums["progressBarStatus"] = progressBarStatus[progressNum]
            t.ExecuteTemplate(&progressBuf, "multiProgress", progressNums)
        }
        urlItemNums["multiProgress"] = template.HTML(progressBuf.String())
        t.ExecuteTemplate(&urlBuf, "urlItem", urlItemNums)
        progressBuf.Truncate(0)
    }
    data["urlItem"] = template.HTML(urlBuf.String())

    t.ExecuteTemplate(w, "body", data)
	t.Execute(w, nil)
}

func Download(ws *websocket.Conn) {

    var err error
    var rec download.WsRespData
    var file download.File
    ch := make(chan int)

    // Full CPU Running
    runtime.GOMAXPROCS(runtime.NumCPU())

    for {
        err = websocket.JSON.Receive(ws, &rec)
        if err != nil {
            var reply download.WsRespData
            reply.Status = "fail"
            reply.Msg = "Not JSON format"
            websocket.JSON.Send(ws, reply)
            break
        }

        storagePath, err2 := osmod.GetStoragePath()
        if err2 != nil {
            rec.Status = "fail"
            rec.Msg = "Storage path doesn't exist."
            websocket.JSON.Send(ws, rec)
            break
        }

        go download.DownloadFile(rec.Url, storagePath, ws, &rec, ch);
        errNum := <-ch
        if  errNum == 0 {
            rec.Status = "fail"
            os.Remove(file.Path)
        } else {
            // Success
            rec.Status = "ok"
        }
        fmt.Println(rec.Msg)

        if err = websocket.JSON.Send(ws, rec); err == nil {
            // If success then close connection.
            if errNum == 1 {
                fmt.Println("Close websocket connection.")
                ws.Close()
            }
        } else {
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
