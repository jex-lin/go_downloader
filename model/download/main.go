package download

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
    "go_downloader/model/osmod"
    "code.google.com/p/go.net/websocket"
)

const (
	tryCountLimit int           = 1
)

type RespData struct {
    Target string
    Url string
    Progress int
    Status string
    Msg string
    FilePath string
}

type File struct {
	Url        string
	Name       string
    Size       int64
    SpendTime  string
	Path       string
	ConnStatus bool
    Ws         *websocket.Conn
    RespData    *RespData
}

var DefaultFile = File{
	ConnStatus: false,
    SpendTime: "0s",
}

func (file *File) Download () (err error){
	// Get data
	resp, err := http.Get(file.Url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("server return non-200 status: %v", resp.Status)
		err = errors.New(errMsg)
		return
	}
	i, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return
	}
    fileSize := int64(i)
	file.Size = fileSize

    // If file already had been downloaded, don't do it again.
    isExistent, fileInfo := osmod.GetFileInfo(file.Path)
    if isExistent {
        if fileSize == fileInfo.Size() {
            file.Size = fileSize
            return
        }
    }

	var fileData io.Reader = resp.Body

	// Create file
	dest, err := os.Create(file.Path)
	if err != nil {
		errMsg := fmt.Sprintf("Can't create %s : %v", file.Path, err)
		err = errors.New(errMsg)
		return
	}
	defer dest.Close()

	// Progress
	startTime := time.Now()
	_, err = file.progress(dest, fileData)
	endTime := time.Now()

	// Print result
	subTime := endTime.Sub(startTime)
	file.Size = fileSize
	file.SpendTime = subTime.String()
	return
}

func DownloadFile(url string, storagePath string, ws *websocket.Conn, rec *RespData, ch chan int) {
    urlSplit := strings.Split(url, "/")
    file := DefaultFile
    file.Url = url
    file.Name = urlSplit[len(urlSplit)-1]
    file.Path = storagePath + string(os.PathSeparator) + file.Name
    file.Ws = ws
    file.RespData = rec
    file.RespData.FilePath = file.Path

	err := file.Download()
	if err == nil {
		file.RespData.Msg = fmt.Sprintf("%s (%d bytes) has been download! Spend time : %s", file.Name, file.Size, file.SpendTime)
		file.ConnStatus = true
        ch <- 1
	} else {
		file.RespData.Msg = fmt.Sprintf("  **Fail to connect %s", file.Name)
        ch <- 0
	}
}

func (file *File) progress(dest *os.File, fileData io.Reader) (written int64, err error) {
	var p float32
	buf := make([]byte, 32*1024)

    file.RespData.Status = "keep"
    var flag = map[int] interface{}{}

	for {
		nr, er := fileData.Read(buf)
		if nr > 0 {
			nw, ew := dest.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			p = float32(written) / float32(file.Size) * 100

            // Response 5% -> 10% -> 15% -> 20% ...... 95% -> 100%
            pp := int(p)
            if pp >= 5 && pp % 5 == 0 {
                if flag[pp] != true {
                    file.RespData.Progress = pp
                    websocket.JSON.Send(file.Ws, file.RespData)
                    fmt.Printf("%s progress: %v%%\n", file.Name, int(p))
                }
                flag[pp] = true
            }

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
