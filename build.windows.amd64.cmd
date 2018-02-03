@echo off
set GOOS=windows
set GOARCH=amd64
go build -o dns-dependency-go-windows-amd64.exe -i -a -v -ldflags="-s -w"
upx -9 dns-dependency-go-windows-amd64.exe
dns-dependency-go-windows-amd64.exe