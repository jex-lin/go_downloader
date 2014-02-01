package web

import(
    "fmt"
    "net/http"
    "strings"
    "go_downloader/source/download"
    "encoding/json"
)

func Api(w http.ResponseWriter, r *http.Request) {

output := map[string] interface{} {}

    // Receive post
    r.ParseForm()
    if r.Method == "POST" {
        url1 := strings.TrimSpace(r.FormValue("url1"))
        urlList := []string {
            url1,
        }
        err := download.DownloadFiles(urlList);
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
