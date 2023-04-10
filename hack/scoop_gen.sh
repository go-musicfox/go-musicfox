#!/usr/bin/env bash

# generate windows scoop config 

set -o nounset
set -o pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

TAG="$(git describe --tags --abbrev=0)"

if [[ -z "$TAG" ]]; then
  echo "TAG is empty"
  exit -1
fi

SCOOP_VERSION=${TAG#refs/tags/}
SCOOP_VERSION=${SCOOP_VERSION#v}

SCOOP_HASH="$(sha256sum "dist/go-musicfox_${SCOOP_VERSION}_windows_amd64.zip" | awk '{print $1}')"

eval "cat <<EOF
$(< "$ROOT"/deploy/scoop/go-musicfox.json.tpl)
EOF
"  > "$ROOT"/deploy/scoop/go-musicfox.json

echo $SCOOP_VERSION $SCOOP_HASH
cat "$ROOT"/deploy/scoop/go-musicfox.json
