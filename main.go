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
    tryCount = 3
)

type File struct {
	url  string
	name string
	path string
    tryCount int
}

func file_default_data(url string) (file File) {
	urlSplit := strings.Split(url, "/")
	name := urlSplit[len(urlSplit)-1]
	path := "/tmp/" + name
    tryCount := tryCount
	return File{url, name, path, tryCount}
}

func download(file File) (fileSize int64, spendTime string, err error) {
	// Get data
	resp, _err := http.Get(file.url)
	if _err != nil {
		log.Fatal(_err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("server return non-200 status: %v", resp.Status)
		err = errors.New(errMsg)
		return 0, "", err
	}
	i, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	fileSize = int64(i)
	var fileData io.Reader = resp.Body

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
	p := progress(dest, fileData, fileSize)
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

func progress(dest *os.File, fileData io.Reader, fileSize int64) (p float32) {
	var read int64
	buffer := make([]byte, 1024)
	for {
		cBytes, _ := fileData.Read(buffer)
		if cBytes == 0 {
			break
		}
		read = read + int64(cBytes)
		p = float32(read) / float32(fileSize) * 100
		fmt.Printf("progress: %v%% \n", int(p))
		if _, err := dest.Write(buffer[:cBytes]); err != nil {
            panic(err)
        }
	}
	return
}

func handleDownload(key int, file File, ch chan int) {
	fileSize, spendTime, err := download(file)

	if err == nil {
		fmt.Printf("%s (%d bytes) has been download! Spend time : %s\n", file.name, fileSize, spendTime)
		ch <- -1
	} else {
		fmt.Println("  **Error :", err)
		ch <- key
	}
}

func main() {
    // Full CPU Running
    runtime.GOMAXPROCS(runtime.NumCPU())

    var urlCount, failKey int
    var files []File
    var file File

	// Urls
	urlList := []string {
        "https://calibre-ebook.googlecode.com/files/eight-demo.flv",
        "http://www.hdflvplayer.net/hdflvplayer/hdplayer.swf",
	}
    urlCount = len(urlList)
	ch := make(chan int, urlCount)
	for key, url := range urlList {
        file = file_default_data(url)
        files = append(files, file)
		go handleDownload(key, file, ch)
	}
	for i := 0; i < urlCount; i++ {
        failKey = <-ch
        if failKey != -1 {
            if files[failKey].tryCount <= 3 {
                files[failKey].tryCount++
                go handleDownload(failKey, files[failKey], ch)
            } else {
                fmt.Println("Fail to connect %s", files[failKey].name)
            }
        }
	}
}
