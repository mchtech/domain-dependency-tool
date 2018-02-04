@echo off
set GOOS=darwin
set GOARCH=amd64
go build -o dns-dependency-go-darwin-amd64.bin -v -ldflags="-s -w"
upx -9 dns-dependency-go-darwin-amd64.bin