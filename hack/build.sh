#!/usr/bin/env bash

set -o nounset
set -o pipefail

export GO111MODULE=auto

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

ACTION=${1:-build}
BUILD_GOOS=${GOOS:-$(go env GOOS)}
BUILD_GOARCH=${GOARCH:-$(go env GOARCH)}
GOBINARY=${GOBINARY:-go}
LDFLAGS=${LDFLAGS:-""}
INJECT_PACKAGE=${INJECT_PACKAGE:-"github.com/go-musicfox/go-musicfox/pkg/constants"}
LASTFM_KEY=${LASTFM_KEY:-""}
LASTFM_SECRET=${LASTFM_SECRET:-""}
BUILD_TARGET=${BUILD_TARGET:-""}

if [[ "${ACTION}" = "build" ]]; then
    BUILD_TARGET="-o ${ROOT}/bin/musicfox"
fi

VERSION_INFO=$(mktemp)
"${ROOT}"/hack/version.sh > "${VERSION_INFO}"

while read -r line; do
    read -r KEY VALUE < <(echo "$line")
    LDFLAGS=${LDFLAGS}" -X ${INJECT_PACKAGE}.${KEY}=${VALUE}"
done < "${VERSION_INFO}"

LDFLAGS=${LDFLAGS}" -X ${INJECT_PACKAGE}.LastfmKey=${LASTFM_KEY}"
LDFLAGS=${LDFLAGS}" -X ${INJECT_PACKAGE}.LastfmSecret=${LASTFM_SECRET}"

CGO_ENABLED=1 GOOS=${BUILD_GOOS} GOARCH=${BUILD_GOARCH} ${GOBINARY} ${ACTION} \
         -ldflags "${LDFLAGS}" ${BUILD_TARGET} \
         ${ROOT}/cmd/musicfox.go
