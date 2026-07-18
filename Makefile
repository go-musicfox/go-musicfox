PACKAGE_NAME          := go-musicfox
PACKAGE_ROOT          := $(CURDIR)
GOLANG_CROSS_VERSION  ?= v1.26.3
INJECT_PACKAGE        ?= github.com/go-musicfox/go-musicfox/internal/types
LDFLAGS               := -s -w
LASTFM_KEY            ?=
LASTFM_SECRET         ?=
REGISTRY              ?=
GORELEASER_IMAGE      ?= alanalbert/goreleaser-musicfox:$(GOLANG_CROSS_VERSION)

SYSROOT_DIR     ?= sysroots
SYSROOT_ARCHIVE ?= sysroots.tar.bz2

# modvendor 用于在 go mod vendor 后复制 C/C++ 头文件等非 Go 文件
MODVENDOR_BIN := $(shell go env GOPATH)/bin/modvendor

# ── OS 检测 ──────────────────────────────────────────────────────────────────
# Windows_NT → 使用 PowerShell 构建脚本；其他系统 → 使用 Bash 构建脚本
ifeq ($(OS),Windows_NT)
    BUILD_SCRIPT := powershell -NoProfile -ExecutionPolicy Bypass -File hack\build.ps1
    NULL_DEV     := nul
    WHICH_CMD    := where
else
    BUILD_SCRIPT := $(PACKAGE_ROOT)/hack/build.sh
    NULL_DEV     := /dev/null
    WHICH_CMD    := which
endif

ifneq ($(REGISTRY),)
	GORELEASER_IMAGE := $(REGISTRY)/go-musicfox/goreleaser-musicfox:$(GOLANG_CROSS_VERSION)
endif

.PHONY: build
build:
	$(BUILD_SCRIPT) build

# build-macapp 仅在 macOS 下有效
.PHONY: build-macapp
build-macapp:
ifneq ($(OS),Windows_NT)
	@mkdir -p $(PACKAGE_ROOT)/bin
	BUILD_TAGS="enable_global_hotkey" $(PACKAGE_ROOT)/hack/build.sh build
	@mkdir -p $(PACKAGE_ROOT)/bin/musicfox.app/Contents/MacOS
	@mkdir -p $(PACKAGE_ROOT)/bin/musicfox.app/Contents/Resources
	@cp $(PACKAGE_ROOT)/bin/musicfox $(PACKAGE_ROOT)/bin/musicfox.app/Contents/MacOS/go-musicfox
	@cp $(PACKAGE_ROOT)/deploy/musicfox-app.wrapper.sh $(PACKAGE_ROOT)/bin/musicfox.app/Contents/MacOS/musicfox
	@cp $(PACKAGE_ROOT)/deploy/musicfox.app/Contents/Info.plist $(PACKAGE_ROOT)/bin/musicfox.app/Contents/Info.plist
	@cp $(PACKAGE_ROOT)/deploy/musicfox.app/Contents/Resources/Musicfox.icns $(PACKAGE_ROOT)/bin/musicfox.app/Contents/Resources/Musicfox.icns
	@chmod +x $(PACKAGE_ROOT)/bin/musicfox.app/Contents/MacOS/go-musicfox
	@chmod +x $(PACKAGE_ROOT)/bin/musicfox.app/Contents/MacOS/musicfox
else
	@echo "build-macapp is not supported on Windows"
endif

.PHONY: init
init:
	git config --local core.hooksPath githooks

.PHONY: install
install:
	$(BUILD_SCRIPT) install

.PHONY: scoop-config-gen
scoop-config-gen:
ifneq ($(OS),Windows_NT)
	$(PACKAGE_ROOT)/hack/scoop_gen.sh
else
	@echo "Please run 'hack/scoop_gen.sh' in Git Bash or WSL on Windows"
endif

.PHONY: changelog-gen
changelog-gen:
ifneq ($(OS),Windows_NT)
	$(PACKAGE_ROOT)/hack/changelog_gen.sh
else
	@echo "Please run 'hack/changelog_gen.sh' in Git Bash or WSL on Windows"
endif

.PHONY: lint
lint:
	golangci-lint run -v

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix -v

.PHONY: test
test:
	go test ./internal/... ./utils/... \
		-coverpkg=./internal/...,./utils/... \
		-covermode=atomic -coverprofile=coverage.txt

.PHONY: clean
clean:
ifeq ($(OS),Windows_NT)
	@powershell -NoProfile -Command "Remove-Item -Recurse -Force -ErrorAction SilentlyContinue '$(PACKAGE_ROOT)/bin', '$(PACKAGE_ROOT)/vendor', '$(PACKAGE_ROOT)/coverage.txt'"
else
	@rm -rf $(PACKAGE_ROOT)/bin $(PACKAGE_ROOT)/vendor $(PACKAGE_ROOT)/coverage.txt
endif

.PHONY: modvendor-tool
modvendor-tool:
	@$(WHICH_CMD) modvendor >$(NULL_DEV) 2>&1 || ( \
		echo "Installing modvendor..."; \
		go install github.com/goware/modvendor@latest; \
	)

.PHONY: vendor
vendor: modvendor-tool
	go mod tidy
	go mod vendor
	@echo "Copying C/C++ header files for CGo dependencies..."
	modvendor -copy="**/*.h **/*.c" -v

.PHONY: sysroot-pack
sysroot-pack:
ifneq ($(OS),Windows_NT)
	@tar cf - $(SYSROOT_DIR) -P | pv -s $[$(du -sk $(SYSROOT_DIR) | awk '{print $1}') * 1024] | pbzip2 > $(SYSROOT_ARCHIVE)
else
	@echo "sysroot-pack is not supported on Windows"
endif

.PHONY: sysroot-unpack
sysroot-unpack:
ifneq ($(OS),Windows_NT)
	@pv $(SYSROOT_ARCHIVE) | pbzip2 -cd | tar -xf -
else
	@echo "sysroot-unpack is not supported on Windows"
endif

.PHONY: release-dry-run
release-dry-run:
	@docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		--env-file .release-env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(PACKAGE_ROOT):/go/src/$(PACKAGE_NAME) \
		-v $(PACKAGE_ROOT)/sysroot:/sysroot \
		-w /go/src/$(PACKAGE_NAME) \
		$(GORELEASER_IMAGE) \
		--clean --skip validate --skip publish

.PHONY: release
release:
ifneq ($(OS),Windows_NT)
	@if [ ! -f ".release-env" ]; then \
		echo "\033[91m.release-env is required for release\033[0m"; \
		exit 1; \
	fi
endif
	docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		--env-file .release-env \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(PACKAGE_ROOT):/go/src/$(PACKAGE_NAME) \
		-v $(PACKAGE_ROOT)/sysroot:/sysroot \
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
		-v $(PACKAGE_ROOT):/go/src/$(PACKAGE_NAME) \
		-v $(PACKAGE_ROOT)/sysroot:/sysroot \
		-w /go/src/$(PACKAGE_NAME) \
		--entrypoint="/bin/bash" \
		$(GORELEASER_IMAGE)
