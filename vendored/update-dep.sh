if [ -z "$1" ]; then
  echo "You must provide an import path"
  exit 1
fi

git subtree add --squash -P "vendored/$1" "https://$1"
