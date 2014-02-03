package download

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
    "go_downloader/model/osmod"
)

const (
	tryCountLimit int           = 3
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
	SpendTime: "0s",
	Err:       nil,
}

func Download(file File) (ConnReturn ConnReturn) {
	ConnReturn = DefaultConnReturn

	// Get data
	resp, err := http.Get(file.Url)
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

    // If file already had been downloaded, don't do it again.
    isExistent, fileInfo := osmod.GetFileInfo(file.Path)
    if isExistent {
        if fileSize == fileInfo.Size() {
            ConnReturn.FileSize = fileSize
            return ConnReturn
        }
    }

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
	_, err = Progress(&file.Name, dest, fileData, fileSize)
	endTime := time.Now()

	// Print result
	subTime := endTime.Sub(startTime)
	ConnReturn.FileSize = fileSize
	ConnReturn.SpendTime = subTime.String()
	ConnReturn.Err = err
	return ConnReturn
}

func Progress(fileName *string, dest *os.File, fileData io.Reader, fileSize int64) (written int64, err error) {
	var p float32
	buf := make([]byte, 32*1024)

	for {
		nr, er := fileData.Read(buf)
		if nr > 0 {
			nw, ew := dest.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			p = float32(written) / float32(fileSize) * 100
			fmt.Printf("%s progress: %v%%\n", *fileName, int(p))
			if ew != nil {
				err = ew
			}
			if nr != nw {
				err = errors.New("short write")
			}
		}
		if er != nil {
            if er.Error() == "EOF" {
                // Sucessfully finish downloading
                return written, nil
            }
			err = er
			break
		}
	}
	return written, err
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

func DownloadFile(url string, storagePath string) (err error) {
    if len(url) == 0 {
		err = errors.New("Url doesn't exsit!")
        return err
    }

	// Full CPU Running
	runtime.GOMAXPROCS(runtime.NumCPU())

	var chReturn File
	var file File
	ch := make(chan File)

    urlSplit := strings.Split(url, "/")
    file = DefaultFile
    file.Url = url
    file.Name = urlSplit[len(urlSplit)-1]
    file.Path = storagePath + string(os.PathSeparator) + file.Name
    go HandleDownload(file, ch)
    chReturn = <-ch
    if chReturn.ConnStatus == false {
        if chReturn.RetryCount < tryCountLimit {
            fmt.Println(chReturn.Msg)
            go HandleDownload(chReturn, ch)
        } else {
            fmt.Println(chReturn.Msg)
            os.Remove(file.Path)
            err = errors.New(fmt.Sprintf("  **Give up to connect %s\n", chReturn.Name))
        }
    } else {
        fmt.Println(chReturn.Msg)
    }
    return
}

func DownloadFiles(urlList []string, storagePath string) (err error) {
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
		file.Path = storagePath + string(os.PathSeparator) + file.Name
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
				os.Remove(file.Path)
				err = errors.New(fmt.Sprintf("  **Give up to connect %s\n", chReturn.Name))
			}
		} else {
			fmt.Println(chReturn.Msg)
		}
	}
	return
}
