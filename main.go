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

type File struct {
	url  string
	name string
	path string
}

func file_default_parameter(url string) (file File) {
	urlSplit := strings.Split(url, "/")
	name := urlSplit[len(urlSplit)-1]
	path := "/tmp/" + name
	return File{url, name, path}
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

func handleDownload(key int, url string, ch chan int) {
	file := file_default_parameter(url)
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

	// Urls
	urls := []string{
		"https://calibre-ebook.googlecode.com/files/eight-demo.flv",
        "http://www.hdflvplayer.net/hdflvplayer/hdplayer.swf",
	}
	ch := make(chan int, len(urls))
	for key, url := range urls {
		go handleDownload(key, url, ch)
	}
	for i := 0; i < len(urls); i++ {
		fmt.Println(<-ch)
	}
}
