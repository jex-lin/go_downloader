set GOARCH=386
set GOOS=windows
rsrc -manifest gui.manifest -ico gui.ico -o gui.syso
go build -ldflags="-H windowsgui"