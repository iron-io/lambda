#!/bin/sh

# This will build both the linux and mac binaries (only if you're on a Mac)

img="treeder/golang-ubuntu:1.4.2on14.04"

go build -o ironcli_mac

# note: Could use GOPATH instead to map volumes
cdir=$(pwd)
godir="$(dirname "$cdir")"
echo "dir $godir"
godir="$(dirname "$godir")"
echo "dir $godir"
godir="$(dirname "$godir")"
echo "dir $godir"

# To remove a bad container and start fresh:
# docker rm ironcli-build

docker run -i --rm --name ironcli-build -v "$godir":/go/src -w /go/src/github.com/iron-io/ironcli -p 0.0.0.0:8080:8080 $img sh -c 'go install && go build -o ironcli_linux' # || docker start -i ironcli-build

# to bash in
#docker run -it --name ironcli-bash -v "$godir":/go/src -w /go/src/github.com/iron-io/ironcli -p 0.0.0.0:8080:8080 $img /bin/bash
