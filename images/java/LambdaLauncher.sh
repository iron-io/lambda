#!/bin/sh

contains() {
  string="$1"
  substring="$2"
  if test "${string#*$substring}" != "$string"; then
    return 0
  else
    return 1
  fi
}

if contains "$1" ".jar" || contains "$1" ".zip";then
  mv $1 __UserCodeLambdaFunction__.jar
else
  echo "Please set jar|zip filename in first param"
  exit 1
fi

if [ -z "$HANDLER" ];then
  if [ -z "$2" ];then
    echo "Please set HANDLER env var or specify handler in the second param of CMD command in Dockerfile"
    exit 1
  else
    export HANDLER="$2"
  fi
fi

if [ -z "$PAYLOAD_FILE" ];then
  if [ -z "$3" ];then
    echo "Please set PAYLOAD_FILE env var or specify payload in the third param of CMD command in Dockerfile"
    exit 1
  else
    export PAYLOAD_FILE="$3"
  fi
fi

java -jar lambda.jar

exit 0
