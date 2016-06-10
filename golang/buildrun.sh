#!/bin/bash

export GOPATH=$(pwd)

fuser 8088/tcp -k

go build -o bin/server src/server.go
bin/server 8088
