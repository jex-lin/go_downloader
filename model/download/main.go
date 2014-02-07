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

type WsRespData struct {
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
    HttpResp   *http.Response
    Ws         *websocket.Conn
    WsRespData    *WsRespData
}

var DefaultFile = File{
	ConnStatus: false,
    SpendTime: "0s",
}

func (file *File) GetHttpResp(url string) (err error) {
	// Get data
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("server return non-200 status: %v", resp.Status)
		err = errors.New(errMsg)
        return
	}

    // Save length
	i, err := strconv.Atoi(resp.Header.Get("Content-Length"))
    file.Size = int64(i)
    file.HttpResp = resp
    return
}

func CheckHttpRange(url string) (has bool, err error){
    resp, err := http.Get(url)
    if err == nil {
        if resp.Header.Get("Accept-Ranges") == "bytes" {
            return true, nil
        }
    }
    defer resp.Body.Close()
    return false, err
}

func (file *File) Download () (err error){
    // If file already had been downloaded, don't do it again.
    isExistent, fileInfo := osmod.GetFileInfo(file.Path)
    if isExistent {
        if file.Size == fileInfo.Size() {
            return
        }
    }

	// Create file
	dest, err := os.Create(file.Path)
	if err != nil {
		errMsg := fmt.Sprintf("Can't create %s : %v", file.Path, err)
		err = errors.New(errMsg)
		return
	}
	defer dest.Close()

	// Progress
    var ioReader io.Reader = file.HttpResp.Body
	defer file.HttpResp.Body.Close()
	startTime := time.Now()
	_, err = file.progress(dest, ioReader)
	endTime := time.Now()

	// Output result
	subTime := endTime.Sub(startTime)
	file.SpendTime = subTime.String()
	return
}

func DownloadFile(url string, storagePath string, ws *websocket.Conn, rec *WsRespData, ch chan int) {
    urlSplit := strings.Split(url, "/")
    file := DefaultFile
    file.Url = url
    file.Name = urlSplit[len(urlSplit)-1]
    file.Path = storagePath + string(os.PathSeparator) + file.Name
    file.Ws = ws
    file.WsRespData = rec
    file.WsRespData.FilePath = file.Path

    // Check connection OK
    err := file.GetHttpResp(url)
    if err != nil {
        file.WsRespData.Msg = err.Error()
        ch <- 0
    } else {
        err = file.Download()
        if err == nil {
            file.WsRespData.Msg = fmt.Sprintf("%s (%d bytes) has been download! Spend time : %s", file.Name, file.Size, file.SpendTime)
            file.ConnStatus = true
            ch <- 1
        } else {
            file.WsRespData.Msg = err.Error()
            ch <- 0
        }
    }
}

func (file *File) progress(dest *os.File, ioReader io.Reader) (written int64, err error) {
	var p float32
	buf := make([]byte, 32*1024)

    file.WsRespData.Status = "keep"
    var flag = map[int] interface{}{}

	for {
		nr, er := ioReader.Read(buf)
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
                    file.WsRespData.Progress = pp
                    websocket.JSON.Send(file.Ws, file.WsRespData)
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
