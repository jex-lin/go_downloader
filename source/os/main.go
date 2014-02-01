package os

import(
    "go_downloader/source/download"
    "net/http"
)

func Os(w http.ResponseWriter, r *http.Request) {


	// Urls
	urlList := []string{
		//"https://calibre-ebook.googlecode.com/files/eight-demo.flv",
		//"http://www.paulgu.com/w/images/f/f0/Honda_accord.flv",
		//"http://vault.futurama.sk/joomla/media/video/video2.flv",
		//"http://video.disclose.tv/12/69/demo_video_13_FLV_126943.flv",
	}
    if err := download.DownloadFiles(urlList); err != nil {
        //fmt.Fprintf(w, err.Error())
    }
}
