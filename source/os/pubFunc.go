package os

import (
    "os"
    "errors"
)

var storagePath string

func IsDir(path string) (bool) {
    _, err := os.Stat(path)
    if err == nil { return true }
    if os.IsNotExist(err) { return false }
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
