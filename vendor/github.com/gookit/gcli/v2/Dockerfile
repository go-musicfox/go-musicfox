#
# @build-example build . -f Dockerfile -t gcli:test
#

################################################################################
###  builder image
################################################################################
FROM golang:1.14-alpine as Builder

# Recompile the standard library without CGO
#RUN CGO_ENABLED=0 go install -a std

ENV APP_DIR $GOPATH/src/github.com/gookit/gcli
RUN mkdir -p $APP_DIR

ADD . $APP_DIR

# Compile the binary and statically link
# -ldflags '-w -s'
#   -s: 去掉符号表
#   -w: 去掉调试信息，不能gdb调试了
# RUN cd $APP_DIR && CGO_ENABLED=0 go build -ldflags '-d -w -s' -o /tmp/app-server
RUN go version && cd $APP_DIR && go build -ldflags '-w -s' -o /tmp/app-server
# RUN cd $APP_DIR && go build -o /tmp/app-server

################################################################################
###  target image
################################################################################
FROM alpine:3.10
LABEL maintainer="inhere <in.798@qq.com>" version="1.0"

##
# ---------- env settings ----------
##

ARG timezone
# prod audit test dev. --build-arg app_env=dev
ARG app_env=dev
ARG app_port

ENV APP_ENV=${app_env:-"dev"} \
    APP_PORT=${app_port:-59430} \
    TIMEZONE=${timezone:-"Asia/Shanghai"}

##
# ---------- some config, clear work ----------
##
RUN set -ex \
        && sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/' /etc/apk/repositories \
        # install some tools
        && apk update && apk add --no-cache tzdata ca-certificates \

        # clear caches
        && rm -rf /var/cache/apk/* \

        # - config timezone
        && ln -sf /usr/share/zoneinfo/${TIMEZONE} /etc/localtime \
        && echo "${TIMEZONE}" > /etc/timezone \
        && date -R \

        # - create logs, caches dir
        && mkdir -p /data/logs /var/www \
        # && chown -R www:www /data/logs \

        && echo -e "\033[42;97m Build Completed :).\033[0m\n"

EXPOSE ${APP_PORT}
WORKDIR "/var/www"

COPY --from=Builder /tmp/app-server app-server
#COPY conf conf
#COPY static static
#COPY resources resources

ENTRYPOINT ["./app-server"]
