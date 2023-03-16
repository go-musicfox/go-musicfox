#!/bin/bash

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

GOBINARY=${GOBINARY:-go}
GOBUILDFLAGS=${GOBUILDFLAGS:-""}
GCFLAGS=${GCFLAGS:-}
LD_FLAGS=${LD_FLAGS:-""}
INJECT_PACKAGE=${INJECT_PACKAGE:-"go-musicfox/pkg/constants"}


VERSION_INFO=$(mktemp)
"${ROOT}"/hack/version.sh > "${VERSION_INFO}"

while read -r line; do
    read -r KEY VALUE < <(echo "$line")
    LD_FLAGS=${LD_FLAGS}" -X ${INJECT_PACKAGE}.${KEY}='${VALUE}'"
done < "${VERSION_INFO}"

time CGO_ENABLED=1 ${GOBINARY} build ${GOBUILDFLAGS} ${GCFLAGS:+-gcflags "${GCFLAGS}"} -o ${ROOT}"/bin/musicfox" \
            -ldflags "${LD_FLAGS}" "${ROOT}"/cmd/musicfox.go


