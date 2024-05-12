#!/usr/bin/env bash

# get version info from git

set -o nounset
set -o pipefail
set -x

if GIT_REVISION=$(git rev-parse HEAD 2>/dev/null); then
	if ! git diff-index --quiet HEAD --; then
		GIT_REVISION=${GIT_REVISION}"-dev"
	fi
else
	GIT_REVISION=unknown
fi

RELEASE_TAG=$(git describe --tags --always --dirty=-dev || echo "")

VERSION="0.0.0-${GIT_REVISION}"
if [[ -n "${RELEASE_TAG}" ]]; then
	VERSION="${RELEASE_TAG}"
fi

echo AppVersion "${VERSION}"
echo GitRevision "${GIT_REVISION}"
echo User "$(whoami)"
if [[ ! $(hostname -f 2>/dev/null) ]]; then
	echo Host "$(cat /etc/hostname)"
else
	echo Host "$(hostname -f)"
fi
echo Time "$(date '+%Y-%m-%dT%T')"$()
