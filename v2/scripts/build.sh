#!/bin/bash
# go-musicfox v2 主构建脚本
# 支持微内核架构的完整构建流程

set -euo pipefail

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"
DIST_DIR="$PROJECT_ROOT/dist"
PLUGIN_DIR="$PROJECT_ROOT/plugins"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 错误处理
error_exit() {
    log_error "$1"
    exit 1
}

# 清理函数
cleanup() {
    log_info "清理临时文件..."
    # 清理临时文件
}

# 注册清理函数
trap cleanup EXIT

# 显示帮助信息
show_help() {
    cat << EOF
go-musicfox v2 构建脚本

用法: $0 [选项] [目标]

选项:
  -h, --help          显示帮助信息
  -v, --verbose       详细输出
  -c, --clean         构建前清理
  -t, --target        指定构建目标 (kernel|plugins|all)
  -p, --platform      指定目标平台 (linux|darwin|windows|all)
  -a, --arch          指定架构 (amd64|arm64|all)
  -o, --output        指定输出目录
  --version           显示版本信息
  --debug             启用调试模式
  --race              启用竞态检测
  --cgo               启用CGO (默认启用)

目标:
  kernel              仅构建微内核
  plugins             仅构建插件
  all                 构建所有组件 (默认)

示例:
  $0                  # 构建所有组件
  $0 -t kernel        # 仅构建微内核
  $0 -p linux -a amd64 # 构建Linux AMD64版本
  $0 -c -v            # 清理后详细构建

EOF
}

# 默认配置
VERBOSE=false
CLEAN=false
TARGET="all"
PLATFORM="$(go env GOOS)"
ARCH="$(go env GOARCH)"
OUTPUT_DIR="$BUILD_DIR"
DEBUG=false
RACE=false
CGO_ENABLED=1

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -c|--clean)
            CLEAN=true
            shift
            ;;
        -t|--target)
            TARGET="$2"
            shift 2
            ;;
        -p|--platform)
            PLATFORM="$2"
            shift 2
            ;;
        -a|--arch)
            ARCH="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --version)
            echo "go-musicfox v2 构建脚本 1.0.0"
            exit 0
            ;;
        --debug)
            DEBUG=true
            shift
            ;;
        --race)
            RACE=true
            shift
            ;;
        --cgo)
            CGO_ENABLED=1
            shift
            ;;
        --no-cgo)
            CGO_ENABLED=0
            shift
            ;;
        -*)
            error_exit "未知选项: $1"
            ;;
        *)
            TARGET="$1"
            shift
            ;;
    esac
done

# 验证参数
case $TARGET in
    kernel|plugins|all) ;;
    *) error_exit "无效的构建目标: $TARGET" ;;
esac

case $PLATFORM in
    linux|darwin|windows|all) ;;
    *) error_exit "无效的平台: $PLATFORM" ;;
esac

case $ARCH in
    amd64|arm64|all) ;;
    *) error_exit "无效的架构: $ARCH" ;;
esac

# 获取版本信息
get_version_info() {
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
    GO_VERSION=$(go version | cut -d' ' -f3)
    
    if [[ $VERBOSE == true ]]; then
        log_info "版本信息:"
        echo "  版本: $VERSION"
        echo "  提交: $COMMIT"
        echo "  构建时间: $BUILD_TIME"
        echo "  Go版本: $GO_VERSION"
    fi
}

# 准备构建环境
prepare_build() {
    log_info "准备构建环境..."
    
    # 创建必要目录
    mkdir -p "$OUTPUT_DIR" "$DIST_DIR" "$PLUGIN_DIR"
    
    # 清理旧文件
    if [[ $CLEAN == true ]]; then
        log_info "清理旧的构建文件..."
        rm -rf "$BUILD_DIR"/* "$DIST_DIR"/*
        go clean -cache -testcache
    fi
    
    # 检查Go环境
    if ! command -v go &> /dev/null; then
        error_exit "Go未安装或不在PATH中"
    fi
    
    # 检查依赖
    log_info "检查依赖..."
    cd "$PROJECT_ROOT"
    go mod download
    go mod verify
}

# 构建微内核
build_kernel() {
    log_info "构建微内核..."
    
    local output_name="go-musicfox"
    if [[ $PLATFORM != "$(go env GOOS)" ]] || [[ $ARCH != "$(go env GOARCH)" ]]; then
        output_name="${output_name}-${PLATFORM}-${ARCH}"
    fi
    
    if [[ $PLATFORM == "windows" ]]; then
        output_name="${output_name}.exe"
    fi
    
    local ldflags="-X main.Version=$VERSION -X main.Commit=$COMMIT -X main.BuildTime=$BUILD_TIME -X main.GoVersion=$GO_VERSION"
    if [[ $DEBUG != true ]]; then
        ldflags="$ldflags -s -w"
    fi
    
    local build_flags="-ldflags=\"$ldflags\" -trimpath"
    if [[ $RACE == true ]]; then
        build_flags="$build_flags -race"
    fi
    
    if [[ $VERBOSE == true ]]; then
        build_flags="$build_flags -v"
    fi
    
    cd "$PROJECT_ROOT"
    
    GOOS="$PLATFORM" GOARCH="$ARCH" CGO_ENABLED="$CGO_ENABLED" \
    go build $build_flags -o "$OUTPUT_DIR/$output_name" ./cmd/musicfox
    
    if [[ $? -eq 0 ]]; then
        log_success "微内核构建完成: $OUTPUT_DIR/$output_name"
    else
        error_exit "微内核构建失败"
    fi
}

# 构建插件
build_plugins() {
    log_info "构建插件..."
    
    if [[ ! -f "$SCRIPT_DIR/build-plugins.sh" ]]; then
        log_warning "插件构建脚本不存在，跳过插件构建"
        return 0
    fi
    
    local plugin_args=""
    if [[ $VERBOSE == true ]]; then
        plugin_args="$plugin_args -v"
    fi
    if [[ $DEBUG == true ]]; then
        plugin_args="$plugin_args --debug"
    fi
    
    "$SCRIPT_DIR/build-plugins.sh" $plugin_args -p "$PLATFORM" -a "$ARCH"
    
    if [[ $? -eq 0 ]]; then
        log_success "插件构建完成"
    else
        error_exit "插件构建失败"
    fi
}

# 运行测试
run_tests() {
    if [[ $TARGET == "all" ]]; then
        log_info "运行测试..."
        cd "$PROJECT_ROOT"
        
        local test_flags="-v"
        if [[ $RACE == true ]]; then
            test_flags="$test_flags -race"
        fi
        
        go test $test_flags ./...
        
        if [[ $? -eq 0 ]]; then
            log_success "测试通过"
        else
            log_warning "测试失败，但继续构建"
        fi
    fi
}

# 生成构建报告
generate_report() {
    local report_file="$OUTPUT_DIR/build-report.txt"
    
    cat > "$report_file" << EOF
go-musicfox v2 构建报告
========================

构建信息:
  版本: $VERSION
  提交: $COMMIT
  构建时间: $BUILD_TIME
  Go版本: $GO_VERSION
  目标: $TARGET
  平台: $PLATFORM
  架构: $ARCH
  CGO: $([ $CGO_ENABLED -eq 1 ] && echo "启用" || echo "禁用")
  竞态检测: $([ $RACE == true ] && echo "启用" || echo "禁用")
  调试模式: $([ $DEBUG == true ] && echo "启用" || echo "禁用")

构建文件:
EOF
    
    find "$OUTPUT_DIR" -type f -name "*" | while read -r file; do
        if [[ -f "$file" ]]; then
            local size=$(du -h "$file" | cut -f1)
            echo "  $(basename "$file") ($size)" >> "$report_file"
        fi
    done
    
    log_info "构建报告已生成: $report_file"
}

# 主函数
main() {
    log_info "开始构建 go-musicfox v2..."
    
    get_version_info
    prepare_build
    
    case $TARGET in
        kernel)
            build_kernel
            ;;
        plugins)
            build_plugins
            ;;
        all)
            run_tests
            build_kernel
            build_plugins
            ;;
    esac
    
    generate_report
    log_success "构建完成！"
}

# 执行主函数
main "$@"