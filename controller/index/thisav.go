package index

import(
    "html/template"
    "net/http"
    "strings"
    "path/filepath"
    "go_downloader/model/osmod"
    "go_downloader/model/download"
    "fmt"
    "os"
    "strconv"
    "runtime"
    "bytes"
)

func Thisav(w http.ResponseWriter, r *http.Request) {
    var data = map[string] interface{}{}
    data["evil"] = true

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

    // parse thisav

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

