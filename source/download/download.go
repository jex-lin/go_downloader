package download

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
    "runtime"
    "strings"
)

const (
    httpTimeout time.Duration = 5 * time.Second
    tryCountLimit int = 5
)

type File struct {
	Url        string
	Name       string
	Path       string
	RetryCount int
	ConnStatus bool
	Msg        string
}

var DefaultFile = File{
	RetryCount: 0,
	ConnStatus: false,
	Msg:        "",
}

type ConnReturn struct {
	FileSize  int64
	SpendTime string
	Err       error
}

var DefaultConnReturn = ConnReturn{
	FileSize:  0,
	SpendTime: "",
	Err:       nil,
}

func Download(file File) (ConnReturn ConnReturn) {
	ConnReturn = DefaultConnReturn

	// Set timeout for http.get
	client := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(httpTimeout)
				c, err := net.DialTimeout(netw, addr, time.Second*5)
				if err != nil {
					return nil, errors.New("Timeout")
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

	// Get data
	resp, err := client.Get(file.Url)
	if err != nil {
		ConnReturn.Err = err
		return ConnReturn
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("server return non-200 status: %v", resp.Status)
		ConnReturn.Err = errors.New(errMsg)
		return ConnReturn
	}
	i, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		ConnReturn.Err = err
		return ConnReturn
	}
	fileSize := int64(i)
	var fileData io.Reader = resp.Body

	// Create file
	dest, err := os.Create(file.Path)
	if err != nil {
		errMsg := fmt.Sprintf("Can't create %s : %v", file.Path, err)
		ConnReturn.Err = errors.New(errMsg)
		return ConnReturn
	}
	defer dest.Close()

	// Progress
	startTime := time.Now()
	p := Progress(&file.Name, dest, fileData, fileSize)
	endTime := time.Now()

	// Print result
	if p == 100 {
		err = nil
	} else {
		os.Remove(file.Path)
		err = errors.New("p isn't 100 percent")
	}
	subTime := endTime.Sub(startTime)
	ConnReturn.FileSize = fileSize
	ConnReturn.SpendTime = subTime.String()
	ConnReturn.Err = err
	return ConnReturn
}

func Progress(fileName *string, dest *os.File, fileData io.Reader, fileSize int64) (p float32) {
	var read int64
	buffer := make([]byte, 1448)
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

func HandleDownload(file File, chFile chan File) {
	ConnReturn := Download(file)
	if ConnReturn.Err == nil {
		file.Msg = fmt.Sprintf("%s (%d bytes) has been download! Spend time : %s", file.Name, ConnReturn.FileSize, ConnReturn.SpendTime)
		file.ConnStatus = true
		chFile <- file
	} else {
		file.RetryCount++
		file.Msg = fmt.Sprintf("  **Fail to connect %s %d time(s).", file.Name, file.RetryCount)
		chFile <- file
	}
}

func DownloadFiles(urlList []string, storagePath string) (err error){
    if len(urlList) == 0 {
        err = errors.New("Url doesn't exsit!")
        return err
    }

	// Full CPU Running
	runtime.GOMAXPROCS(runtime.NumCPU())

	var chReturn File
	var files []File
	var file File

	ch := make(chan File, len(urlList))
	for _, url := range urlList {
		urlSplit := strings.Split(url, "/")
		file = DefaultFile
		file.Url = url
		file.Name = urlSplit[len(urlSplit)-1]
		file.Path = storagePath + string(os.PathSeparator) +file.Name
		files = append(files, file)
		go HandleDownload(file, ch)
	}
	chCount := len(urlList)
	for i := 0; i < chCount; i++ {
		chReturn = <-ch
		if chReturn.ConnStatus == false {
			if chReturn.RetryCount < tryCountLimit {
				fmt.Println(chReturn.Msg)
				go HandleDownload(chReturn, ch)
				chCount++
			} else {
				fmt.Println(chReturn.Msg)
				err = errors.New(fmt.Sprintf("  **Give up to connect %s\n", chReturn.Name))
			}
		} else {
			fmt.Println(chReturn.Msg)
		}
	}
    return
}
