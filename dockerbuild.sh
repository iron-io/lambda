#!/bin/sh

img="treeder/golang-ubuntu:1.4.2on14.04"

go build -o ironcli_mac

# note: Could use GOPATH instead to map volumes (can you do more than one -v?)
cdir=$(pwd)
godir="$(dirname "$cdir")"
echo "dir $godir"
godir="$(dirname "$godir")"
echo "dir $godir"
godir="$(dirname "$godir")"
echo "dir $godir"
#godir="$(dirname "$godir")"
#echo "dir $godir"


# To remove a bad container and start fresh:
# docker rm ironcli-build

# todo: should check if ironmq-build container exists and reuse it so we don't start with a fresh build each time
# Using go install here so it installs gorocksdb the first time. It also does go build, so all good.
docker run -i --name ironcli-build -v "$godir":/go/src -w /go/src/github.com/iron-io/ironcli -p 0.0.0.0:8080:8080 $img sh -c 'go install && go build -o ironcli_linux' || docker start -i ironcli-build

# to bash in
#docker run -it --name ironcli-build -v "$godir":/go/src -w /go/src/github.com/iron-io/ironcli -p 0.0.0.0:8080:8080 $img /bin/bash
