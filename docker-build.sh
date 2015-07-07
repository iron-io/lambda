#!/usr/bin/env bash
docker run --rm -it -v "$GOPATH":/go -w /go/src/github.com/iron-io/ironcli golang:1.4.2-cross sh -c '
for GOOS in darwin linux windows; do
  for GOARCH in 386 amd64; do
    echo "Building $GOOS-$GOARCH"
    export GOOS=$GOOS
    export GOARCH=$GOARCH
    env
    go build -o bin/ironcli-$GOOS-$GOARCH
  done
done
'

sudo chmod -R 777 bin/
