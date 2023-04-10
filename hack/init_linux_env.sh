#!/usr/bin/env bash

# install windows dependency in docker

set -o nounset
set -o pipefail

BUILD_HOST=${BUILD_HOST:-"x86_64-linux-gnu"}
LIBFLAC_VER=1.3.3
LIBALSA_VER=1.2.2

# install libflac
cd /tmp
rm -rf "flac-${LIBFLAC_VER}.tar.xz" "flac-${LIBFLAC_VER}"
wget "https://github.com/xiph/flac/releases/download/${LIBFLAC_VER}/flac-${LIBFLAC_VER}.tar.xz"
tar -xf "flac-${LIBFLAC_VER}.tar.xz"
cd "flac-${LIBFLAC_VER}"
./autogen.sh
CFLAGS='' CPPFLAGS='' LDFLAGS='' LIBS='' ./configure --host=${BUILD_HOST} --prefix=/usr/${BUILD_HOST} --disable-ogg
make -j4 && make install

# install alsa-lib
cd /tmp
rm -rf "alsa-lib-${LIBALSA_VER}.tar.bz2" "alsa-lib-${LIBALSA_VER}"
wget "https://www.alsa-project.org/files/pub/lib/alsa-lib-${LIBALSA_VER}.tar.bz2"
tar -jxf "alsa-lib-${LIBALSA_VER}.tar.bz2"
cd "alsa-lib-${LIBALSA_VER}"
CFLAGS='' CPPFLAGS='' LDFLAGS='' LIBS='' ./configure --host=${BUILD_HOST} --prefix=/usr/${BUILD_HOST} --enable-shared=yes --enable-static=no --with-pic
make -j4 && make install