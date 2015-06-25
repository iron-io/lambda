#!/usr/bin/env bash

#godir="$(dirname "$(dirname "$(dirname "$(dirname "$(pwd)")")")")"
#echo "dir $godir"
#exit

docker run --rm -it -v "$GOPATH":/go -w /go/src/github.com/iron-io/ironcli golang:1.3-cross sh -c '
for GOOS in darwin linux windows; do
#  for GOARCH in 386 amd64; do
    go build -o bin/ironcli_$GOOS
#  done
done
'
