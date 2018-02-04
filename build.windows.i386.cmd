@echo off
set GOOS=windows
set GOARCH=386
go build -o dns-dependency-go-windows-i386.exe -v -ldflags="-s -w"
upx -9 dns-dependency-go-windows-i386.exe
dns-dependency-go-windows-i386.exe