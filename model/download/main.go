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
    MulDowAtLeastSize = 30 * 1024 * 1024
    MulSectionDowCount = int64(5)    // max = 5
)

type WsRespData struct {
    Target string           // #url-1
    Url string
    Progress int            // 22  (%)
    Status string           // ok  fail
    SingleOrMulti string
    PartNum int             // multi part num.    #url-1-4
    Msg string              // message
    FilePath string         // /tmp/video.flv
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
    file.WsRespData.SingleOrMulti = "single"
    file.WsRespData.Status = "UpdateUI"
    websocket.JSON.Send(file.Ws, file.WsRespData)

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
    file.WsRespData.SingleOrMulti = "multi"
    file.WsRespData.Status = "UpdateUI"
    websocket.JSON.Send(file.Ws, file.WsRespData)

	// Create file
	dest, err := os.Create(file.Path)
	if err != nil {
		errMsg := fmt.Sprintf("Can't create %s : %v", file.Path, err)
		err = errors.New(errMsg)
		return
	}
	defer dest.Close()

    var start, end int64
    chMulDow := make(chan int64, MulSectionDowCount)
    fmt.Println("total: " + strconv.Itoa(int(file.Size)))
    ReqRangeSize := int64(file.Size / MulSectionDowCount)

    startTime := time.Now()
    for partNum := int64(1); partNum <= MulSectionDowCount; partNum++ {
        if partNum == MulSectionDowCount {
            end = file.Size
        } else {
            end = start + ReqRangeSize
        }
        //fmt.Println(fmt.Sprintf("%d  ->  %d", start, end-1))
        go file.RangeWrite(dest, start, end, chMulDow, partNum)
        start = end
    }
    for i := int64(1); i <= MulSectionDowCount; i++ {
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
        if resp.Close {
            continue
        }
        if resp.StatusCode == 206 {
            fmt.Println("Support http range")
        } else {
            return nil, errors.New("Not support http range")
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
    file.WsRespData.Status = "keep"
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
            if pp >= 20 && pp % 20 == 0 {
                if flag[pp] != true {
                    file.WsRespData.Progress = pp / int(MulSectionDowCount)
                    file.WsRespData.PartNum = int(partNum)
                    websocket.JSON.Send(file.Ws, file.WsRespData)
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
                // If file is too small, use single download
                if file.Size < MulDowAtLeastSize {
                    fmt.Println("Support http range, but file size is too small, choose single download")
                    err = file.SingleDownload()
                } else {
                    fmt.Println("Support http range")
                    err = file.MultiDownload()
                }
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
