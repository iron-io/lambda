#!/bin/bash

# This script builds the iron/cli Docker image, tags and pushes it to Docker Hub.

#git clone https://github.com/iron-io/ironcli.git
#cd ironcli

img=iron/go
wd=/go/src/github.com/iron-io/ironcli
docker run --rm -v "$PWD":$wd -w $wd $img go build -o iron

docker build -t iron/cli:latest .
# Tag it with the version too
x="$(docker run --rm iron/cli --version)"
echo "$x"
y=${x:1}
echo "version tag: $y"
docker tag -f iron/cli:latest iron/cli:$y
docker push iron/cli
