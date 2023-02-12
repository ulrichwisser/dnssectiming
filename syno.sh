#!/bin/sh
GOOS=linux
GOARCH=amd64
CC=/usr/local/musl/bin/musl-gcc

apk add --no-cache gcc musl-dev
go build -v -buildvcs=false -ldflags "-linkmode external -extldflags -static"
