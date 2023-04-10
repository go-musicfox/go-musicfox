#!/usr/bin/env bash

# install linux dependency in docker

set -o nounset
set -o pipefail

BUILD_HOST=${BUILD_HOST:-"x86_64-linux-gnu"}
BUILD_ARCH="x86_64"

# install mingw
case $(uname -m) in
    x86_64)  BUILD_ARCH=x86_64;;
    aarch64) BUILD_ARCH=aarch64;;
    arm64) BUILD_ARCH=aarch64;
esac

mingw="llvm-mingw-20230320-ucrt-ubuntu-18.04-${BUILD_ARCH}"

cd /tmp
rm -rf "${mingw}.tar.xz" "${mingw}" /usr/local/mingw
wget "https://github.com/mstorsjo/llvm-mingw/releases/download/20230320/${mingw}.tar.xz"
tar -xf "${mingw}.tar.xz"
mv "${mingw}" /usr/local/mingw