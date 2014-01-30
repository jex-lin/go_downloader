package web

import(
    "os"
    "errors"
)

var storagePath string

func IsDir(path string) bool {
    if res, err := os.Stat(path); err != nil {
        return os.IsExist(err)
    } else {
        return res.IsDir()
    }
    return false
}

func GetStoragePath() (string, error) {
    if IsDir(storagePath) {
        return storagePath, nil
    }
    return "", errors.New("Directory doesn't exist")
}

func SetStoragePath(path string) bool {
    if (IsDir(path)) {
        storagePath = path
        return true
    }
    return false
}


