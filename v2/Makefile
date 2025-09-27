# go-musicfox v2 微内核插件架构 Makefile
# 支持多平台构建、插件开发、容器化部署

.PHONY: help build test lint clean deps fmt vet \
        build-kernel build-plugins build-all \
        cross-compile package docker \
        install uninstall dev-server \
        plugin-gen plugin-validate plugin-package \
        security-scan benchmark docs

# 版本信息
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION ?= $(shell go version | cut -d' ' -f3)

# 构建配置
APP_NAME = go-musicfox
MAIN_PACKAGE = ./cmd/musicfox
BUILD_DIR = build
DIST_DIR = dist
PLUGIN_DIR = plugins
DOCS_DIR = docs

# Go构建标志
LDFLAGS = -X main.Version=$(VERSION) \
          -X main.Commit=$(COMMIT) \
          -X main.BuildTime=$(BUILD_TIME) \
          -X main.GoVersion=$(GO_VERSION) \
          -s -w

BUILD_FLAGS = -ldflags="$(LDFLAGS)" -trimpath
TEST_FLAGS = -v -race -coverprofile=coverage.out
LINT_FLAGS = --timeout=5m

# 平台配置
PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
PLUGIN_PLATFORMS = linux/amd64 darwin/amd64 windows/amd64

# 插件类型
PLUGIN_TYPES = dynamic-library rpc webassembly hot-reload

# 默认目标
help:
	@echo "go-musicfox v2 微内核插件架构构建系统"
	@echo ""
	@echo "核心构建目标:"
	@echo "  build         - 构建微内核应用"
	@echo "  build-kernel  - 仅构建微内核"
	@echo "  build-plugins - 构建所有插件"
	@echo "  build-all     - 构建应用和插件"
	@echo "  cross-compile - 交叉编译所有平台"
	@echo "  package       - 打包发布版本"
	@echo ""
	@echo "开发工具:"
	@echo "  test          - 运行测试"
	@echo "  lint          - 代码检查"
	@echo "  fmt           - 格式化代码"
	@echo "  vet           - 静态分析"
	@echo "  security-scan - 安全扫描"
	@echo "  benchmark     - 性能测试"
	@echo ""
	@echo "插件开发:"
	@echo "  plugin-gen    - 生成插件模板"
	@echo "  plugin-validate - 验证插件"
	@echo "  plugin-package  - 打包插件"
	@echo "  dev-server    - 启动开发服务器"
	@echo ""
	@echo "部署相关:"
	@echo "  docker        - 构建Docker镜像"
	@echo "  install       - 安装到系统"
	@echo "  uninstall     - 从系统卸载"
	@echo ""
	@echo "其他:"
	@echo "  deps          - 下载依赖"
	@echo "  clean         - 清理构建文件"
	@echo "  docs          - 生成文档"
	@echo "  tools         - 安装开发工具"

# 准备构建目录
prepare:
	@mkdir -p $(BUILD_DIR) $(DIST_DIR) $(PLUGIN_DIR)

# 构建微内核应用
build: prepare
	@echo "构建微内核应用 $(VERSION)..."
	CGO_ENABLED=1 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PACKAGE)
	@echo "构建完成: $(BUILD_DIR)/$(APP_NAME)"

# 仅构建微内核
build-kernel: prepare
	@echo "构建微内核..."
	CGO_ENABLED=1 go build $(BUILD_FLAGS) -tags kernel -o $(BUILD_DIR)/$(APP_NAME)-kernel ./pkg/kernel
	@echo "微内核构建完成: $(BUILD_DIR)/$(APP_NAME)-kernel"

# 构建插件
build-plugins: prepare
	@echo "构建所有插件..."
	@for plugin_type in $(PLUGIN_TYPES); do \
		echo "构建 $$plugin_type 插件..."; \
		./scripts/build-plugins.sh $$plugin_type || exit 1; \
	done
	@echo "所有插件构建完成"

# 构建应用和插件
build-all: build build-plugins
	@echo "完整构建完成"

# 交叉编译
cross-compile: prepare
	@echo "开始交叉编译..."
	@for platform in $(PLATFORMS); do \
		echo "编译平台: $$platform"; \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		CGO_ENABLED=1 go build $(BUILD_FLAGS) \
		-o $(BUILD_DIR)/$(APP_NAME)-$${platform%/*}-$${platform#*/} \
		$(MAIN_PACKAGE) || exit 1; \
	done
	@echo "交叉编译完成"

# 打包发布版本
package: cross-compile build-plugins
	@echo "打包发布版本..."
	./scripts/package.sh $(VERSION)
	@echo "打包完成: $(DIST_DIR)/"

# 运行测试
test:
	@echo "运行测试套件..."
	go test $(TEST_FLAGS) ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "测试完成，覆盖率报告: coverage.html"

# 运行基准测试
benchmark:
	@echo "运行性能测试..."
	go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./...
	@echo "性能测试完成"

# 代码检查
lint:
	@echo "运行代码检查..."
	golangci-lint run $(LINT_FLAGS)
	@echo "代码检查完成"

# 安全扫描
security-scan:
	@echo "运行安全扫描..."
	gosec ./...
	nancy sleuth
	@echo "安全扫描完成"

# 格式化代码
fmt:
	@echo "格式化代码..."
	go fmt ./...
	goimports -w .
	gofumpt -w .

# 静态分析
vet:
	@echo "运行静态分析..."
	go vet ./...
	staticcheck ./...

# 下载依赖
deps:
	@echo "下载依赖..."
	go mod download
	go mod tidy
	go mod verify

# 生成文档
docs:
	@echo "生成文档..."
	godoc -http=:6060 &
	@echo "文档服务器启动: http://localhost:6060"

# 插件生成器
plugin-gen:
	@echo "生成插件模板..."
	./tools/plugin-generator.sh $(PLUGIN_TYPE) $(PLUGIN_NAME)

# 插件验证
plugin-validate:
	@echo "验证插件..."
	./tools/plugin-validator.sh $(PLUGIN_PATH)

# 插件打包
plugin-package:
	@echo "打包插件..."
	./tools/plugin-packager.sh $(PLUGIN_PATH)

# 开发服务器
dev-server:
	@echo "启动开发服务器..."
	./tools/dev-server.sh

# Docker构建
docker:
	@echo "构建Docker镜像..."
	docker build -f docker/Dockerfile -t $(APP_NAME):$(VERSION) .
	docker build -f docker/Dockerfile.dev -t $(APP_NAME):dev .
	@echo "Docker镜像构建完成"

# 系统安装
install: build
	@echo "安装到系统..."
	./deploy/install.sh

# 系统卸载
uninstall:
	@echo "从系统卸载..."
	./deploy/uninstall.sh

# 清理构建文件
clean:
	@echo "清理构建文件..."
	go clean -cache -testcache -modcache
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	rm -f coverage.out coverage.html cpu.prof mem.prof
	find . -name "*.so" -delete
	find . -name "*.wasm" -delete
	@echo "清理完成"

# 安装开发工具
tools:
	@echo "安装开发工具..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install mvdan.cc/gofumpt@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install github.com/sonatypecommunity/nancy@latest
	@echo "开发工具安装完成"

# 显示构建信息
info:
	@echo "构建信息:"
	@echo "  版本: $(VERSION)"
	@echo "  提交: $(COMMIT)"
	@echo "  构建时间: $(BUILD_TIME)"
	@echo "  Go版本: $(GO_VERSION)"
	@echo "  目标平台: $(PLATFORMS)"
	@echo "  插件类型: $(PLUGIN_TYPES)"

# 快速开发构建
dev: fmt vet test build
	@echo "开发构建完成"

# 生产构建
prod: clean security-scan lint test cross-compile package
	@echo "生产构建完成"

# CI构建
ci: deps tools lint test security-scan build
	@echo "CI构建完成"