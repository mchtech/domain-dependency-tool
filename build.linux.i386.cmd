@echo off
set GOOS=linux
set GOARCH=386
go build -o dns-dependency-go-linux-i386.bin -i -a -v -ldflags="-s -w"
upx -9 dns-dependency-go-linux-i386.bin