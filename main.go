package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
    "runtime"
)

const (
    tryCountLimit = 5
)

type File struct {
	url  string
	name string
	path string
    retryCount int
    connStatus bool
    msg string
}

func file_default_data(url string) (file File) {
	urlSplit := strings.Split(url, "/")
	name := urlSplit[len(urlSplit)-1]
	path := "/tmp/" + name
    retryCount := 0
    connStatus := false
    msg := ""
	return File{url, name, path, retryCount, connStatus, msg}
}

func download(file File) (fileSize int64, spendTime string, err error) {
	// Get data
	resp, err := http.Get(file.url)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("server return non-200 status: %v", resp.Status)
		err = errors.New(errMsg)
		return 0, "", err
	}
	i, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	fileSize = int64(i)
	var fileData io.Reader = resp.Body
	defer resp.Body.Close()

	// Create file
	dest, err := os.Create(file.path)
	if err != nil {
		errMsg := fmt.Sprintf("Can't create %s : %v", file.path, err)
		err = errors.New(errMsg)
		return 0, "", err
	}
	defer dest.Close()

	// Progress
	startTime := time.Now()
	p := progress(&file.name, dest, fileData, fileSize)
	endTime := time.Now()

	// Print result
	if p == 100 {
		err = nil
	} else {
		err = errors.New("fail")
	}
	subTime := endTime.Sub(startTime)
	spendTime = subTime.String()
	return fileSize, spendTime, err
}

func progress(fileName *string, dest *os.File, fileData io.Reader, fileSize int64) (p float32) {
	var read int64
	buffer := make([]byte, 1024)
	for {
		cBytes, _ := fileData.Read(buffer)
		if cBytes == 0 {
			break
		}
		read = read + int64(cBytes)
		p = float32(read) / float32(fileSize) * 100
        //fmt.Printf("%s progress: %v%%\n", *fileName, int(p))
		dest.Write(buffer[:cBytes])
	}
	return
}

func handleDownload(file File, chFile chan File) {
	fileSize, spendTime, err := download(file)
	if err == nil {
        file.msg = fmt.Sprintf("%s (%d bytes) has been download! Spend time : %s", file.name, fileSize, spendTime)
        file.connStatus = true
		chFile <- file
	} else {
        file.retryCount++
        file.msg = fmt.Sprintf("  **Fail to connect %s %d time(s)", file.name, file.retryCount)
		chFile <- file
	}
}

func main() {
    // Full CPU Running
    runtime.GOMAXPROCS(runtime.NumCPU())

    var chReturn File
    var files []File
    var file File

	// Urls
	urlList := []string {
        //"https://calibre-ebook.googlecode.com/files/eight-demo.flv",
        "http://www.hdflvplayer.net/hdflvplayer/hdplayer.swf",
	}
	ch := make(chan File, len(urlList))
	for _, url := range urlList {
        file = file_default_data(url)
        files = append(files, file)
		go handleDownload(file, ch)
	}
    chCount := len(urlList)
	for i := 0; i < chCount; i++ {
        chReturn = <-ch
        if chReturn.connStatus == false {
            if chReturn.retryCount <= tryCountLimit {
                fmt.Println(chReturn.msg)
                go handleDownload(chReturn, ch)
                chCount++
            } else {
                fmt.Printf("  **Give up to connect %s\n", chReturn.name)
            }
        } else {
            fmt.Println(chReturn.msg)
        }
	}
}
