#!/bin/sh

if [ -z "$JAR_ZIP_FILENAME" ];then
  echo "Please set JAR_ZIP_FILENAME env var"
  exit 1
fi

cp $JAR_ZIP_FILENAME /LambdaLauncher/

cd /LambdaLauncher
javac -cp ".:./*:/LambdaLauncher/*" LambdaLauncher.java
java  -cp ".:./*:/LambdaLauncher/*" LambdaLauncher

exit 0
