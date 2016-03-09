#!/bin/sh

( cd ../../python && make )
( make )
docker run --rm -it -e PAYLOAD_FILE=/mnt/example-payload.json -v `pwd`:/mnt iron/lambda-python-example

