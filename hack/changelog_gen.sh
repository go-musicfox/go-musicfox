#!/usr/bin/env bash

# generate changelog

set -o nounset
set -o pipefail
set -x

if ! command -v chglog >/dev/null 2>&1; then
	go install github.com/goreleaser/chglog/cmd/chglog@latest
fi

"$GOPATH/bin/chglog" init
"$GOPATH/bin/chglog" format --template repo >CHANGELOG.md
