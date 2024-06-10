#!/usr/bin/env bash

# get version info from git

set -o nounset
set -o pipefail
set -x

GIT_TAG=${GIT_TAG:-""}
GIT_REVISION=${GIT_REVISION:-""}
GIT_REMOTE_INFO=""

# get latest git info from remote repo
fetch_remote_git_info() {
	if [[ -z "${GIT_REMOTE_INFO}" ]]; then
		GIT_REMOTE_INFO=$(git ls-remote --tags https://github.com/go-musicfox/go-musicfox)
	fi
}

if [[ -z "${GIT_TAG}" ]]; then
	if ! GIT_TAG=$(git describe --tags --always --dirty=-dev 2>/dev/null); then
		fetch_remote_git_info
		tag=$(echo "$GIT_REMOTE_INFO" | tail -n 1 | awk '{print $2}')
		GIT_TAG="${tag#refs/tags/}"
	fi
fi

if [[ -z "${GIT_REVISION}" ]]; then
	if GIT_REVISION=$(git rev-parse HEAD 2>/dev/null); then
		if ! git diff-index --quiet HEAD --; then
			GIT_REVISION=${GIT_REVISION}"-dev"
		fi
	else
		fetch_remote_git_info
		GIT_REVISION=$(echo "$GIT_REMOTE_INFO" | grep -w "refs/tags/${GIT_TAG}" | tail -n 1 | awk '{print $1}')
	fi

fi

VERSION="0.0.0-${GIT_REVISION}"
if [[ -n "${GIT_TAG}" ]]; then
	VERSION="${GIT_TAG}"
fi

echo AppVersion "${VERSION}"
echo GitRevision "${GIT_REVISION}"
echo User "$(whoami)"
if [[ ! $(hostname -f 2>/dev/null) ]]; then
	echo Host "$(cat /etc/hostname)"
else
	echo Host "$(hostname -f)"
fi
echo Time "$(date '+%Y-%m-%dT%T')"
