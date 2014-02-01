package web

import (
    "html/template"
    "net/http"
    "strings"
    "fmt"
    "path/filepath"
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
        fmt.Println(storagePath)
        data["storagePath"] = storagePath
        if SetStoragePath(storagePath) {
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
