#!/bin/sh
set -ex

# builds for each OS and then uploads to a fresh github release.
# make an access token first here: https://github.com/settings/tokens
# and save it somewhere.
#
# must have go compiler boot strapped for all OS for go <= 1.4 -- try this:
#   % git clone git://github.com/davecheney/golang-crosscompile.git
#   % source golang-crosscompile/crosscompile.bash
#   % go-crosscompile-build-all
#
# this is not the world's greatest script but it gets the job done, you
# must have installed: {git, curl, python}

if [ -z "${GH_DEPLOY_KEY}" ]; then
  echo "GH_DEPLOY_KEY must be set"
  exit 1
fi

if [ -z "${GH_DEPLOY_USER}" ]; then
  echo "GH_DEPLOY_USER must be set"
  exit 1
fi

git checkout -b master origin/master

# CircleCI has these set in the project
name=${GH_DEPLOY_USER}
tok=${GH_DEPLOY_KEY}

# bump version
perl -i -pe 's/\d+\.\d+\.\K(\d+)/$1+1/e' main.go
perl -i -pe 's/\d+\.\d+\.\K(\d+)/$1+1/e' install.sh
version=$(grep -Eo "[0-9]+\.[0-9]+\.[0-9]+" install.sh)

# add to git
git add -u
git commit -m "$version release"
git push origin master

url='https://api.github.com/repos/iron-io/ironcli/releases'

# create release
output=$(curl -s -u $name:$tok -d "{\"tag_name\": \"$version\", \"name\": \"$version\"}" $url)
upload_url=$(echo "$output" | python -c 'import json,sys;obj=json.load(sys.stdin);print obj["upload_url"]' | sed -E "s/\{.*//")
html_url=$(echo "$output" | python -c 'import json,sys;obj=json.load(sys.stdin);print obj["html_url"]')

# NOTE: do the builds after the version has been bumped in main.go
echo "uploading exe..."
GOOS=windows  GOARCH=amd64 go build -o bin/ironcli.exe
curl --progress-bar --data-binary "@bin/ironcli.exe"    -H "Content-Type: application/octet-stream" -u $name:$tok $upload_url\?name\=ironcli.exe >/dev/null
echo "uploading elf..."
GOOS=linux    GOARCH=amd64 go build -o bin/ironcli_linux
curl --progress-bar --data-binary "@bin/ironcli_linux"  -H "Content-Type: application/octet-stream" -u $name:$tok $upload_url\?name\=ironcli_linux >/dev/null
echo "uploading mach-o..."
GOOS=darwin   GOARCH=amd64 go build -o bin/ironcli_mac
curl --progress-bar --data-binary "@bin/ironcli_mac"    -H "Content-Type: application/octet-stream" -u $name:$tok $upload_url\?name\=ironcli_mac >/dev/null

echo "Done! Go edit the description: $html_url"
exit 0
