PACKAGE_NAME          := go-musicfox
GOLANG_CROSS_VERSION  ?= v1.20.6
INJECT_PACKAGE        ?= github.com/go-musicfox/go-musicfox/pkg/constants
LDFLAGS               := -s -w
LASTFM_KEY            ?=
LASTFM_SECRET         ?=
REGISTRY              ?=
GORELEASER_IMAGE      ?= alanalbert/goreleaser-musicfox:$(GOLANG_CROSS_VERSION)

SYSROOT_DIR     ?= sysroots
SYSROOT_ARCHIVE ?= sysroots.tar.bz2

ifneq ($(REGISTRY),)
	GORELEASER_IMAGE := $(REGISTRY)/go-musicfox/goreleaser-musicfox:$(GOLANG_CROSS_VERSION)
endif

.PHONY: build
build:
	./hack/build.sh build

.PHONY: install
install:
	./hack/build.sh install

.PHONY: scoop-config-gen
scoop-config-gen:
	./hack/scoop_gen.sh

.PHONY: lint
lint:
	golangci-lint run -v

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix -v

.PHONY: sysroot-pack
sysroot-pack:
	@tar cf - $(SYSROOT_DIR) -P | pv -s $[$(du -sk $(SYSROOT_DIR) | awk '{print $1}') * 1024] | pbzip2 > $(SYSROOT_ARCHIVE)

.PHONY: sysroot-unpack
sysroot-unpack:
	@pv $(SYSROOT_ARCHIVE) | pbzip2 -cd | tar -xf -

.PHONY: release-dry-run
release-dry-run:
	@docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		--env-file .release-env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-v `pwd`/sysroot:/sysroot \
		-w /go/src/$(PACKAGE_NAME) \
		$(GORELEASER_IMAGE) \
		--clean --skip-validate --skip-publish

.PHONY: release
release:
	@if [ ! -f ".release-env" ]; then \
		echo "\033[91m.release-env is required for release\033[0m";\
		exit 1;\
	fi
	docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		--env-file .release-env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-v `pwd`/sysroot:/sysroot \
		-w /go/src/$(PACKAGE_NAME) \
		$(GORELEASER_IMAGE) \
		release --clean

.PHONY: release-debug-shell
release-debug-shell:
	docker run \
    	-it \
    	--rm \
    	--privileged \
    	-e CGO_ENABLED=1 \
    	-v /var/run/docker.sock:/var/run/docker.sock \
    	-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-v `pwd`/sysroot:/sysroot \
		-w /go/src/$(PACKAGE_NAME) \
    	-w /go/src/go-musicfox \
    	--entrypoint="/bin/bash" \
    	$(GORELEASER_IMAGE)