@echo off

SET GOPATH=%cd%

go build -o bin\server.exe src\server.go
bin\server 8088

pause
