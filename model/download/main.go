package download

import (
	"errors"
	"fmt"
	"io"
    "net/url"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
    "go_downloader/model/osmod"
    "code.google.com/p/go.net/websocket"
)

const (
    MulDowAtLeastSize = 1000 * 1024 * 1024
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
			if ew != nil {
				err = ew
			}
			if nr != nw {
				err = errors.New("short write")
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
		}
		if er != nil {
            if er.Error() == "EOF" {
                if written == file.Size {
                    // Sucessfully finish downloading
                    return written, nil
                } else {
                    msg := fmt.Sprintf("%s written %d (unfinished)\n", file.Name, written)
                    return written, errors.New(msg)
                }
            }
			err = er
			break
		}
	}
	return written, err
}

// Get http status
func (file *File) GetHttpResp(url string) (err error) {
	// Get data
	resp, err := http.Get(url)
	if err != nil {
		return
	}
    if resp.Close {
		err = errors.New("Response has closed")
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

// Checking header support Accept-ranges or not.
func (file *File) CheckHttpRange() bool {
    if file.HttpResp.Header.Get("Accept-Ranges") == "bytes" {
        return true
    }
    return false
}

// Check file already has been downloaded or not.
func (file *File) FileHasDownload () bool {
    isExistent, fileInfo := osmod.GetFileInfo(file.Path)
    if isExistent {
        if file.Size == fileInfo.Size() {
            return true
        }
    }
    return false
}

// Not support Accept-ranges
func (file *File) SingleDownload () (err error){
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
	durTime := endTime.Sub(startTime)
	file.SpendTime = durTime.String()
	return
}

// support Accept-ranges
func (file *File) MultiDownload() (err error) {
    // If file is too small, use single download
    if file.Size < MulDowAtLeastSize {
        err = file.SingleDownload()
        return
    }

	// Create file
	dest, err := os.Create(file.Path)
	if err != nil {
		errMsg := fmt.Sprintf("Can't create %s : %v", file.Path, err)
		err = errors.New(errMsg)
		return
	}
	defer dest.Close()

    var start, end int64
    sectionCount := int64(2)
    chMulDow := make(chan int64, sectionCount)
    fmt.Println("total: " + strconv.Itoa(int(file.Size)))
    ReqRangeSize := int64(file.Size / sectionCount)

    startTime := time.Now()
    for i := int64(1); i <= sectionCount; i++ {
        if i == sectionCount {
            end = file.Size
        } else {
            end = start + ReqRangeSize
        }
        //fmt.Println(fmt.Sprintf("%d  ->  %d", start, end-1))
        go file.RangeWrite(dest, start, end, chMulDow, i)
        start = end
    }
    for i := int64(1); i <= sectionCount; i++ {
        written := <-chMulDow
        if written == -1 {
            return errors.New("Multi downloading - range write error")
        }
    }
    endTime := time.Now()
    durTime := endTime.Sub(startTime)
	file.SpendTime = durTime.String()
    return
}

func (file *File) ReqHttpRange (start int64, end int64) (respBody io.Reader,err error) {
    var req http.Request
    header := http.Header{}
    header.Set("Range", "bytes=" + strconv.Itoa(int(start)) + "-" + strconv.Itoa(int(end)))
    req.Header = header
    req.Method = "GET"              // Must, prevent 303
    req.URL, _ = url.Parse(file.Url)
    for {
        resp, err := http.DefaultClient.Do(&req)
        if err != nil {
            return nil, err
        }
        fmt.Printf("multi header : ")
        fmt.Println(resp.Header)
        fmt.Printf("multi body :")
        fmt.Println(resp.Body)
        if resp.Close {
            continue
        }
        if resp.Header.Get("Accept-Ranges") == "bytes" {
            fmt.Println("multi support range")
        } else {
            fmt.Println("multi not support range")
            //return nil, errors.New("multi not support range")
        }
        return resp.Body, nil
    }
    return
}

func (file *File) RangeWrite (dest *os.File, start int64, end int64, chMulDow chan int64, partNum int64) {
    var written int64
    var p float32
    var flag = map[int] interface{}{}
    ioReader, err := file.ReqHttpRange(start, end - 1)
    reqRangeSize := end - start
    if err != nil { return }
    buf := make([]byte, 32 * 1024)
    for {
        nr, er := ioReader.Read(buf)
        if nr > 0 {
            nw, ew := dest.WriteAt(buf[0:nr], start)
            start = int64(nw) + start
            if nw > 0 {
                written += int64(nw)
            }
            if ew != nil {
                err = ew
            }
            if nr != nw {
                err = errors.New("short write")
            }

			p = float32(written) / float32(reqRangeSize) * 100
            pp := int(p)
            if pp >= 25 && pp % 25 == 0 {
                if flag[pp] != true {
                    //file.WsRespData.Progress = pp
                    //websocket.JSON.Send(file.Ws, file.WsRespData)
                    fmt.Printf("%s part%d progress: %v%%\n", file.Name, partNum, int(p))
                }
                flag[pp] = true
            }
        }
        if er != nil {
            if er.Error() == "EOF" {
                //Successfully finish downloading
                if reqRangeSize == written {
                    fmt.Printf("%s part%d written  %d\n", file.Name, partNum, written)
                    chMulDow <- written
                } else {
                    fmt.Printf("%s part%d written  %d (unfinished)\n", file.Name, partNum, written)
                    chMulDow <- -1
                }
                break
            }
            fmt.Printf("part%d downloading error : %s\n", partNum, er.Error())
            chMulDow <- -1
            break
        }
    }
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
        if ! file.FileHasDownload() {
            if file.CheckHttpRange() {
                fmt.Println("Support http range")
                err = file.MultiDownload()
            } else {
                fmt.Println("Not support http range")
                err = file.SingleDownload()
            }
        }
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
