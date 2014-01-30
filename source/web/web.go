package web

import(
    "os"
)

var HomePath string = os.Getenv("HOME")
var DesktopPath string = HomePath + "/Desktop"
var StoragePath string

func IsDir(path string) bool {
    if res, err := os.Stat(path); err != nil {
        return os.IsExist(err)
    } else {
        return res.IsDir()
    }
    return false
}

func SetDesktopPath(path string) bool {
    if (IsDir(path)) {
        DesktopPath = path
        return true
    }
    return false
}


