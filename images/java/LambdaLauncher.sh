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

java -jar lambda.jar $2

exit 0
