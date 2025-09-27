# 开发工具文档索引

本目录包含 go-musicfox v2 微内核插件架构的开发工具使用指南，帮助开发者提高开发效率和代码质量。

## 工具分类

### [开发工具使用说明](development.md)
日常开发中使用的各种工具和命令。

**包含内容：**
- 项目构建工具
- 代码生成工具
- 依赖管理工具
- 版本控制工具

### [调试和诊断工具](debugging.md)
用于问题排查和系统诊断的工具。

**包含内容：**
- 日志分析工具
- 调试器使用
- 问题诊断方法
- 故障排除指南

### [性能分析工具](performance.md)
用于性能测试、分析和优化的工具。

**包含内容：**
- 性能测试工具
- 性能分析方法
- 优化建议
- 基准测试

## 工具概览

### 构建工具

#### Make
项目使用 Makefile 管理构建任务：

```bash
# 构建所有组件
make build

# 运行测试
make test

# 代码检查
make lint

# 清理构建产物
make clean

# 生成文档
make docs
```

#### Go 工具链

```bash
# 构建项目
go build ./...

# 运行测试
go test ./...

# 代码格式化
go fmt ./...

# 静态分析
go vet ./...

# 依赖管理
go mod tidy
go mod download
```

### 代码质量工具

#### golangci-lint
代码静态分析工具：

```bash
# 安装
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 运行检查
golangci-lint run

# 修复可自动修复的问题
golangci-lint run --fix
```

配置文件 `.golangci.yml`：
```yaml
linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck
    - ineffassign
    - deadcode

linters-settings:
  gofmt:
    simplify: true
  goimports:
    local-prefixes: github.com/go-musicfox/go-musicfox
```

#### gofumpt
更严格的代码格式化工具：

```bash
# 安装
go install mvdan.cc/gofumpt@latest

# 格式化代码
gofumpt -w .
```

### 测试工具

#### testify
测试断言库：

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/mock"
)

func TestExample(t *testing.T) {
    // 基本断言
    assert.Equal(t, expected, actual)
    assert.NoError(t, err)
    
    // 必需断言（失败时停止测试）
    require.NotNil(t, obj)
    require.NoError(t, err)
}
```

#### go-mock
模拟对象生成工具：

```bash
# 安装
go install github.com/golang/mock/mockgen@latest

# 生成模拟对象
mockgen -source=interface.go -destination=mock.go
```

### 调试工具

#### Delve
Go 语言调试器：

```bash
# 安装
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试程序
dlv debug ./cmd/musicfox

# 调试测试
dlv test ./pkg/plugin

# 附加到运行中的进程
dlv attach <pid>
```

#### pprof
性能分析工具：

```bash
# CPU 性能分析
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# 内存分析
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# 在线分析
go tool pprof http://localhost:6060/debug/pprof/profile
```

### 文档工具

#### godoc
API 文档生成工具：

```bash
# 安装
go install golang.org/x/tools/cmd/godoc@latest

# 启动文档服务器
godoc -http=:6060

# 访问文档
open http://localhost:6060/pkg/github.com/go-musicfox/go-musicfox/v2/
```

#### swagger
API 文档生成（用于 RPC 插件）：

```bash
# 安装
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
swag init -g main.go
```

### 容器工具

#### Docker
容器化部署：

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o musicfox ./cmd/musicfox

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/musicfox .
CMD ["./musicfox"]
```

```bash
# 构建镜像
docker build -t go-musicfox:v2 .

# 运行容器
docker run -it go-musicfox:v2
```

#### docker-compose
多容器编排：

```yaml
# docker-compose.yml
version: '3.8'

services:
  musicfox:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./config:/app/config
      - ./plugins:/app/plugins
    environment:
      - LOG_LEVEL=debug
  
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
```

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

## 开发工作流

### 1. 项目初始化

```bash
# 克隆项目
git clone https://github.com/go-musicfox/go-musicfox.git
cd go-musicfox/v2

# 安装依赖
go mod download

# 安装开发工具
make install-tools

# 验证环境
make verify
```

### 2. 日常开发

```bash
# 创建功能分支
git checkout -b feature/new-plugin

# 编写代码
vim pkg/plugin/new_plugin.go

# 运行测试
make test

# 代码检查
make lint

# 提交代码
git add .
git commit -m "feat: add new plugin support"
```

### 3. 插件开发

```bash
# 创建插件项目
mkdir my-plugin
cd my-plugin

# 初始化项目
go mod init my-plugin

# 添加依赖
go get github.com/go-musicfox/go-musicfox/v2/pkg/plugin

# 生成插件模板
musicfox-cli plugin init --name=my-plugin --type=dynamic

# 构建插件
make build

# 测试插件
make test
```

### 4. 调试流程

```bash
# 启用调试模式
export LOG_LEVEL=debug

# 使用调试器
dlv debug ./cmd/musicfox -- --config=debug.yaml

# 设置断点
(dlv) break main.main
(dlv) continue

# 查看变量
(dlv) print variable_name
(dlv) locals
```

### 5. 性能分析

```bash
# 运行基准测试
go test -bench=. -benchmem ./...

# 生成性能报告
go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=.

# 分析性能报告
go tool pprof cpu.prof
(pprof) top10
(pprof) web
```

## 自动化工具

### GitHub Actions
持续集成配置：

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Cache dependencies
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run tests
      run: make test
    
    - name: Run linter
      run: make lint
    
    - name: Build
      run: make build
```

### Pre-commit Hooks
提交前检查：

```bash
#!/bin/sh
# .git/hooks/pre-commit

set -e

# 运行代码格式化
echo "Running gofmt..."
gofmt -w .

# 运行静态分析
echo "Running golangci-lint..."
golangci-lint run

# 运行测试
echo "Running tests..."
go test ./...

echo "Pre-commit checks passed!"
```

### Makefile 示例

```makefile
.PHONY: build test lint clean install-tools

# 变量定义
GO_VERSION := 1.21
BINARY_NAME := musicfox
BUILD_DIR := build
COVERAGE_FILE := coverage.out

# 默认目标
all: build

# 构建
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/musicfox

# 测试
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=$(COVERAGE_FILE) ./...
	go tool cover -html=$(COVERAGE_FILE) -o coverage.html

# 代码检查
lint:
	@echo "Running linter..."
	golangci-lint run

# 格式化代码
fmt:
	@echo "Formatting code..."
	gofumpt -w .
	goimports -w .

# 安装开发工具
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install mvdan.cc/gofumpt@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/go-delve/delve/cmd/dlv@latest

# 清理
clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_FILE) coverage.html

# 验证环境
verify:
	@echo "Verifying environment..."
	@go version
	@which golangci-lint || (echo "golangci-lint not found" && exit 1)
	@which gofumpt || (echo "gofumpt not found" && exit 1)
	@echo "Environment verification passed!"
```

## 工具配置

### VS Code 配置

```json
// .vscode/settings.json
{
    "go.useLanguageServer": true,
    "go.formatTool": "gofumpt",
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "package",
    "go.testFlags": ["-v", "-race"],
    "go.coverOnSave": true,
    "go.coverageDecorator": {
        "type": "gutter",
        "coveredHighlightColor": "rgba(64,128,128,0.5)",
        "uncoveredHighlightColor": "rgba(128,64,64,0.25)"
    }
}
```

### GoLand 配置

1. **代码风格**：Settings → Editor → Code Style → Go
2. **静态分析**：Settings → Tools → Go Linter
3. **测试配置**：Run/Debug Configurations → Go Test
4. **调试配置**：Run/Debug Configurations → Go Build

## 常见问题

### Q: 如何设置开发环境？

A: 按照以下步骤：
1. 安装 Go 1.21+
2. 克隆项目代码
3. 运行 `make install-tools`
4. 运行 `make verify` 验证环境

### Q: 如何调试插件？

A: 使用以下方法：
1. 启用调试日志
2. 使用 delve 调试器
3. 添加调试输出
4. 使用单元测试

### Q: 如何进行性能优化？

A: 参考以下步骤：
1. 运行基准测试
2. 使用 pprof 分析
3. 识别性能瓶颈
4. 优化关键路径
5. 验证优化效果

### Q: 如何贡献代码？

A: 遵循以下流程：
1. Fork 项目仓库
2. 创建功能分支
3. 编写代码和测试
4. 运行代码检查
5. 提交 Pull Request

---

更多详细信息请查看具体的工具使用文档。