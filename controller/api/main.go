package api

import(
    "fmt"
    "net/http"
    "strings"
    "go_downloader/library/download"
    "encoding/json"
    "go_downloader/model/osmod"
)

func Api(w http.ResponseWriter, r *http.Request) {

    output := map[string] interface{} {}

    storagePath, err := osmod.GetStoragePath()
    if err != nil {
        output["status"] = "fail"
        output["errMsg"] = err.Error()
        outputJson, _ := json.Marshal(output);
        fmt.Fprintf(w, string(outputJson))
        return
    }
    // Receive post
    r.ParseForm()
    if r.Method == "POST" {
        url := strings.TrimSpace(r.FormValue("url"))
        err := download.DownloadFile(url, storagePath);
        if  err != nil {
            output["status"] = "fail"
            output["errMsg"] = err.Error()
            outputJson, _ := json.Marshal(output);
            fmt.Fprintf(w, string(outputJson))
            return
        }
        output["status"] = "ok"
        outputJson, _ := json.Marshal(output);
        fmt.Fprintf(w, string(outputJson))
    }
}

//dec := json.NewDecoder(strings.NewReader(jsonStream)) 
//dec.Decode(&m);

//add custom http header
//
//client := &amp;http.Client{]
//req, err := http.NewRequest("POST", "http://example.com", bytes.NewReader(postData))
//req.Header.Add("User-Agent", "myClient")
//resp, err := client.Do(req)
//defer resp.Body.Close()
