#!/usr/bin/env bash

set -o nounset
set -o pipefail
set -x

export GO111MODULE=auto

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

ACTION=${1:-build}
GOBINARY=${GOBINARY:-go}
BUILD_GOOS=${GOOS:-$(${GOBINARY} env GOOS)}
BUILD_GOARCH=${GOARCH:-$(${GOBINARY} env GOARCH)}
LDFLAGS=${LDFLAGS:-""}
INJECT_PACKAGE=${INJECT_PACKAGE:-"github.com/go-musicfox/go-musicfox/internal/types"}
LASTFM_KEY=${LASTFM_KEY:-""}
LASTFM_SECRET=${LASTFM_SECRET:-""}
BUILD_TARGET=${BUILD_TARGET:-""}
BUILD_TAGS=${BUILD_TAGS:-""}

if [[ "${ACTION}" = "build" ]]; then
	BUILD_TARGET="${ROOT}/bin/musicfox"
	TARGET_FLAG="-o ${BUILD_TARGET}"
fi

VERSION_INFO=$(mktemp)
"${ROOT}"/hack/version.sh >"${VERSION_INFO}"

while read -r line; do
	read -r KEY VALUE < <(echo "$line")
	LDFLAGS=${LDFLAGS}" -X ${INJECT_PACKAGE}.${KEY}=${VALUE}"
done <"${VERSION_INFO}"

LDFLAGS=${LDFLAGS}" -X ${INJECT_PACKAGE}.LastfmKey=${LASTFM_KEY}"
LDFLAGS=${LDFLAGS}" -X ${INJECT_PACKAGE}.LastfmSecret=${LASTFM_SECRET}"
LDFLAGS=${LDFLAGS}" -X ${INJECT_PACKAGE}.BuildTags=${BUILD_TAGS}"

CGO_ENABLED=1 GOOS=${BUILD_GOOS} GOARCH=${BUILD_GOARCH} ${GOBINARY} "${ACTION}" \
	-tags "${BUILD_TAGS}" \
	-ldflags "${LDFLAGS}"  ${TARGET_FLAG} \
	"${ROOT}/cmd/musicfox.go"
