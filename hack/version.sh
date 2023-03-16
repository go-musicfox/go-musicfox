#!/bin/bash

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

GIT_REVISION=$(git rev-parse HEAD 2> /dev/null)
if [[  $? == 0 ]]; then
    git diff-index --quiet HEAD --
    if [[  $? != 0 ]]; then
        GIT_REVISION=${GIT_REVISION}"-dev"
    fi
else
    GIT_REVISION=unknown
fi

RELEASE_TAG=$(git describe --tags --always --dirty=-dev || echo "")

VERSION="${GIT_REVISION}"
if [[ -n "${RELEASE_TAG}" ]]; then
  VERSION="${RELEASE_TAG}"
fi

echo AppVersion    "${VERSION}"
echo GitRevision   "${GIT_REVISION}"
echo User          "$(whoami)"
echo Host          "$(hostname -f)"
echo Time          "$(date '+%Y-%m-%d %T')"
