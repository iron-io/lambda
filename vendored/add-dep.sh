set -e

if [ -z "$1" ]; then
  echo "You must provide an import path"
  exit 1
fi

# Get the top-level git directory.
GIT_ROOT=$(git rev-parse --show-toplevel)
cd "$GIT_ROOT"

git subtree add --squash -P "vendored/$1" "https://$1" master

# Prefix import paths with "github.com/iron-io/vendored/". This is a pretty dumb
# way to re-write import paths, but it should only go wrong if there's a string
# literal somewhere that happens to start with the exact import path we're looking
# for, which seems pretty unlikely.
find -name '*.go' -exec sed -i s,\""$1\(/.*\)\?"\",\""github.com/iron-io/ironcli/vendored/$1\1"\", {} +
