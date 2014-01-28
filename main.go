package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var tryCountLimit int = 5
var httpTimeout time.Duration = 5 * time.Second

type File struct {
	url        string
	name       string
	path       string
	retryCount int
	connStatus bool
	msg        string
}

var DefaultFile = File {
	retryCount : 0,
	connStatus : false,
	msg : "",
}

type ConnReturn struct {
    fileSize int64
    spendTime string
    err error
}

var DefaultConnReturn = ConnReturn {
fileSize : 0,
spendTime : "",
err : nil,
}

func download(file File) (ConnReturn ConnReturn) {
    ConnReturn = DefaultConnReturn

    // Set timeout for http.get
	client := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(httpTimeout)
				c, err := net.DialTimeout(netw, addr, time.Second * 5)
				if err != nil {
					return nil, errors.New("Timeout")
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

	// Get data
	resp, err := client.Get(file.url)
	if err != nil {
        ConnReturn.err = err
		return ConnReturn
	}
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("server return non-200 status: %v", resp.Status)
		ConnReturn.err = errors.New(errMsg)
		return ConnReturn
	}
	i, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
        ConnReturn.err = err
		return ConnReturn
	}
	defer resp.Body.Close()
    fileSize := int64(i)
	var fileData io.Reader = resp.Body

	// Create file
	dest, err := os.Create(file.path)
	if err != nil {
		errMsg := fmt.Sprintf("Can't create %s : %v", file.path, err)
		ConnReturn.err = errors.New(errMsg)
		return ConnReturn
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
        os.Remove(file.path)
        err = errors.New("p isn't 100 percent")
	}
	subTime := endTime.Sub(startTime)
    ConnReturn.fileSize = fileSize
    ConnReturn.spendTime = subTime.String()
    ConnReturn.err = err
	return ConnReturn
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
	ConnReturn := download(file)
	if ConnReturn.err == nil {
		file.msg = fmt.Sprintf("%s (%d bytes) has been download! Spend time : %s", file.name, ConnReturn.fileSize, ConnReturn.spendTime)
		file.connStatus = true
		chFile <- file
	} else {
		file.retryCount++
		file.msg = fmt.Sprintf("  **Fail to connect %s %d time(s).", file.name, file.retryCount)
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
	urlList := []string{
		"https://calibre-ebook.googlecode.com/files/eight-demo.flv",
        "http://www.paulgu.com/w/images/f/f0/Honda_accord.flv",
        "http://vault.futurama.sk/joomla/media/video/video2.flv",
        "http://video.disclose.tv/12/69/demo_video_13_FLV_126943.flv",
	}
	ch := make(chan File, len(urlList))
	for _, url := range urlList {
        urlSplit := strings.Split(url, "/")
        file = DefaultFile
        file.url = url
        file.name = urlSplit[len(urlSplit)-1]
        file.path = "/tmp/" + file.name
		files = append(files, file)
		go handleDownload(file, ch)
	}
	chCount := len(urlList)
	for i := 0; i < chCount; i++ {
		chReturn = <-ch
		if chReturn.connStatus == false {
			if chReturn.retryCount < tryCountLimit {
				fmt.Println(chReturn.msg)
				go handleDownload(chReturn, ch)
				chCount++
			} else {
				fmt.Println(chReturn.msg)
				fmt.Printf("  **Give up to connect %s\n", chReturn.name)
			}
		} else {
			fmt.Println(chReturn.msg)
		}
	}
}
