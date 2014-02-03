package osmod

import (
    "os"
    "errors"
)

var storagePath string

func GetFileInfo(path string) (isExistent bool, fileInfo os.FileInfo) {
    fileInfo, err := os.Stat(path)
    if err != nil {
        // no such file or dir
        return false, nil
    }
    if fileInfo.IsDir() {
        // it's a directory
        return false, nil
    }
    // it's a file
    return true, fileInfo
}

func DirExists(path string) (bool) {
    fileInfo, err := os.Stat(path)
    if err != nil {
        // no such file or dir
        return false
    }
    if fileInfo.IsDir() {
        // it's a directory
        return true
    }
    // it's a file
    return false
}

func FileExists(path string) (bool) {
    fileInfo, err := os.Stat(path)
    if err != nil {
        // no such file or dir
        return false
    }
    if fileInfo.IsDir() {
        // it's a directory
        return false
    }
    // it's a file
    return true
}

func GetStoragePath() (string, error) {
    if DirExists(storagePath) {
        return storagePath, nil
    }
    return "", errors.New("Directory doesn't exist")
}

func SetStoragePath(path string) bool {
    if (DirExists(path)) {
        storagePath = path
        return true
    }
    return false
}



