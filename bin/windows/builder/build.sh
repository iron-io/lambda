#!/bin/bash

cd /go/src
go get github.com/iron-io/ironcli
export CLI_REPO=/go/src/github.com/iron-io/ironcli
cd $CLI_REPO
GOOS=windows GOARCH=amd64 go build -o iron.exe
export IRONCLI_VERSION=$(cat main.go | grep -E 'Version[ ]{1,9}=' | grep -Eo '[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}')

cd /home/wix
cp $CLI_REPO/iron.exe ./iron.exe
cp $CLI_REPO/bin/windows/iron.ico ./iron.ico
cp $CLI_REPO/bin/windows/iron.wxs ./iron.wxs
curl 'https://gitprint.com/iron-io/ironcli/blob/master/README.md?download' > IronCLI_README.pdf

export IRONCLI_GUID_ID=$(cat /proc/sys/kernel/random/uuid)
echo Building version $IRONCLI_VERSION with GUID $IRONCLI_GUID_ID

wine candle.exe -dCliVersion="$IRONCLI_VERSION" -dCliGuid="$IRONCLI_GUID_ID" iron.wxs
wine light.exe iron.wixobj -sval -out IronCLI-$IRONCLI_VERSION.msi

cp /home/wix/IronCLI-$IRONCLI_VERSION.msi /home/out/
