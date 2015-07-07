#!/bin/sh
set -e

# builds for each OS and then uploads to a fresh github release.
# make an access token first here: https://github.com/settings/tokens
# and save it somewhere.

old=$(grep -E "release.*=.*'.*'" install.sh | grep -Eo "'.*'")

# TODO taking ideas for automating this, can we make a bot+token and stick it in CircleCI?
echo -n "GitHub username: "
read name
echo -n "Access Token (https://github.com/settings/tokens): "
read tok
echo -n "New Version (current: $old): "
read version

url='https://api.github.com/repos/iron-io/ironcli/releases'

output=$(curl -s -u $name:$tok -d "{\"tag_name\": \"$version\", \"name\": \"$version\"}" $url)
upload_url=$(echo "$output" | python -c 'import json,sys;obj=json.load(sys.stdin);print obj["upload_url"]' | sed -E "s/\{.*//")
html_url=$(echo "$output" | python -c 'import json,sys;obj=json.load(sys.stdin);print obj["html_url"]')

sed -Ei "s/release.*=.*'.*'/release='$version'/"        install.sh
sed -Ei "s/Version.*=.*\".*\"/Version = \"$version\"/"  main.go

# NOTE: do the builds after the version has been bumped in main.go
build.sh

echo "uploading exe..."
curl --progress-bar --data-binary "@bin/ironcli-windows-amd64"    -H "Content-Type: application/octet-stream" -u $name:$tok $upload_url\?name\=ironcli.exe >/dev/null
echo "uploading elf..."
curl --progress-bar --data-binary "@bin/ironcli-linux-amd64"  -H "Content-Type: application/octet-stream" -u $name:$tok $upload_url\?name\=ironcli_linux >/dev/null
echo "uploading mach-o..."
curl --progress-bar --data-binary "@bin/ironcli-darwin-amd64"    -H "Content-Type: application/octet-stream" -u $name:$tok $upload_url\?name\=ironcli_mac >/dev/null

echo "Done! Go edit the description: $html_url"
