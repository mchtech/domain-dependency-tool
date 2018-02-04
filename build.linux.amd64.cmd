@echo off
set GOOS=linux
set GOARCH=amd64
go build -o dns-dependency-go-linux-amd64.bin -v -ldflags="-s -w"
upx -9 dns-dependency-go-linux-amd64.bin