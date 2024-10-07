#!/bin/bash

# Input args:
#
# OS
# COMMIT 
# BRANCH
# TAG

script_dir="$(
    cd "$(dirname "${BASH_SOURCE[0]}")" || exit 1
    pwd
)"
project_dir="$(
    cd "${script_dir}/.." || exit 1
    pwd
)"

if [[ -z ${APP} ]]; then
    APP="web"
fi

commit=${COMMIT:-"$(git rev-parse --short HEAD)"}
branch=${BRANCH:-"$(git rev-parse --abbrev-ref HEAD)"}
tag=${TAG:-"$(git describe --tags --exact-match 2>/dev/null || true)"}
version=${branch}-${commit} 

ldflags="-s -w"
ldflags+=" -X github.com/prometheus/common/version.Revision=${commit}"
ldflags+=" -X github.com/prometheus/common/version.Branch=${branch}"
ldflags+=" -X github.com/prometheus/common/version.Version=${version}"
ldflags+=" -X 'main.version=${version}'"
ldflags+=" -X 'main.tag=${tag}'"
ldflags+=" -X 'main.buildTime=$(date "+%Y-%m-%d %H:%M:%S")'"

echo "Building ${APP} ${version} ($tag) for ${OS:-'linux'} ${ARCH:-'amd64'}"

GOOS=${OS:-'linux'} GOARCH='amd64' CGO_ENABLED=0 go build \
    -ldflags="${ldflags}" \
    -o "${project_dir}/bin/${APP}" \
    "${project_dir}/cmd/${APP}"
