#!/bin/bash
# go-musicfox v2 交叉编译脚本
# 支持多平台、多架构的交叉编译

set -euo pipefail

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"
DIST_DIR="$PROJECT_ROOT/dist"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志函数
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

error_exit() {
    log_error "$1"
    exit 1
}

# 显示帮助信息
show_help() {
    cat << EOF
go-musicfox v2 交叉编译脚本

用法: $0 [选项]

选项:
  -h, --help          显示帮助信息
  -v, --verbose       详细输出
  -c, --clean         编译前清理
  -p, --platforms     指定平台列表 (用逗号分隔)
  -a, --archs         指定架构列表 (用逗号分隔)
  -o, --output        指定输出目录
  --kernel-only       仅编译微内核
  --plugins-only      仅编译插件
  --debug             启用调试模式
  --parallel          并行编译
  --compress          压缩二进制文件
  --strip             去除符号表

默认平台: linux,darwin,windows
默认架构: amd64,arm64

示例:
  $0                                    # 编译所有平台和架构
  $0 -p linux,darwin -a amd64          # 仅编译Linux和macOS的AMD64版本
  $0 --kernel-only --parallel           # 并行编译微内核
  $0 --compress --strip                 # 编译并压缩优化

EOF
}

# 默认配置
VERBOSE=false
CLEAN=false
PLATFORMS="linux,darwin,windows"
ARCHS="amd64,arm64"
OUTPUT_DIR="$BUILD_DIR"
KERNEL_ONLY=false
PLUGINS_ONLY=false
DEBUG=false
PARALLEL=false
COMPRESS=false
STRIP=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help) show_help; exit 0 ;;
        -v|--verbose) VERBOSE=true; shift ;;
        -c|--clean) CLEAN=true; shift ;;
        -p|--platforms) PLATFORMS="$2"; shift 2 ;;
        -a|--archs) ARCHS="$2"; shift 2 ;;
        -o|--output) OUTPUT_DIR="$2"; shift 2 ;;
        --kernel-only) KERNEL_ONLY=true; shift ;;
        --plugins-only) PLUGINS_ONLY=true; shift ;;
        --debug) DEBUG=true; shift ;;
        --parallel) PARALLEL=true; shift ;;
        --compress) COMPRESS=true; shift ;;
        --strip) STRIP=true; shift ;;
        -*) error_exit "未知选项: $1" ;;
        *) error_exit "未知参数: $1" ;;
    esac
done

# 获取版本信息
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION=$(go version | cut -d' ' -f3)

# 支持的平台和架构组合
declare -A PLATFORM_ARCH_SUPPORT=(
    ["linux,amd64"]=true
    ["linux,arm64"]=true
    ["darwin,amd64"]=true
    ["darwin,arm64"]=true
    ["windows,amd64"]=true
    ["windows,arm64"]=true
    ["freebsd,amd64"]=true
    ["openbsd,amd64"]=true
    ["netbsd,amd64"]=true
)

# CGO支持的平台
declare -A CGO_SUPPORT=(
    ["linux,amd64"]=true
    ["linux,arm64"]=true
    ["darwin,amd64"]=true
    ["darwin,arm64"]=true
    ["windows,amd64"]=true
)

# 准备编译环境
prepare_build() {
    log_info "准备交叉编译环境..."
    
    mkdir -p "$OUTPUT_DIR"
    
    if [[ $CLEAN == true ]]; then
        log_info "清理旧的构建文件..."
        rm -rf "$BUILD_DIR"/*
        go clean -cache -testcache
    fi
    
    # 检查Go环境
    if ! command -v go &> /dev/null; then
        error_exit "Go未安装或不在PATH中"
    fi
    
    # 检查Go版本
    local go_version=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    local major=$(echo "$go_version" | cut -d. -f1)
    local minor=$(echo "$go_version" | cut -d. -f2)
    
    if [[ $major -lt 1 ]] || [[ $major -eq 1 && $minor -lt 21 ]]; then
        error_exit "需要Go 1.21或更高版本"
    fi
    
    # 检查依赖
    cd "$PROJECT_ROOT"
    go mod download
    go mod verify
    
    # 检查压缩工具
    if [[ $COMPRESS == true ]]; then
        if ! command -v upx &> /dev/null; then
            log_warning "UPX未安装，跳过压缩"
            COMPRESS=false
        fi
    fi
}

# 获取二进制文件扩展名
get_binary_extension() {
    local platform="$1"
    case $platform in
        windows) echo ".exe" ;;
        *) echo "" ;;
    esac
}

# 获取动态库扩展名
get_library_extension() {
    local platform="$1"
    case $platform in
        linux|freebsd|openbsd|netbsd) echo ".so" ;;
        darwin) echo ".dylib" ;;
        windows) echo ".dll" ;;
        *) echo ".so" ;;
    esac
}

# 编译微内核
compile_kernel() {
    local platform="$1"
    local arch="$2"
    
    local combo="$platform,$arch"
    if [[ -z "${PLATFORM_ARCH_SUPPORT[$combo]:-}" ]]; then
        log_warning "不支持的平台组合: $platform/$arch"
        return 0
    fi
    
    log_info "编译微内核: $platform/$arch"
    
    local binary_name="go-musicfox-$platform-$arch$(get_binary_extension "$platform")"
    local output_path="$OUTPUT_DIR/$binary_name"
    
    # 构建标志
    local ldflags="-X main.Version=$VERSION -X main.Commit=$COMMIT -X main.BuildTime=$BUILD_TIME -X main.GoVersion=$GO_VERSION"
    if [[ $DEBUG != true && $STRIP == true ]]; then
        ldflags="$ldflags -s -w"
    fi
    
    local build_flags="-ldflags=\"$ldflags\" -trimpath"
    if [[ $VERBOSE == true ]]; then
        build_flags="$build_flags -v"
    fi
    
    # CGO设置
    local cgo_enabled=0
    if [[ -n "${CGO_SUPPORT[$combo]:-}" ]]; then
        cgo_enabled=1
    fi
    
    # 交叉编译
    cd "$PROJECT_ROOT"
    
    GOOS="$platform" GOARCH="$arch" CGO_ENABLED="$cgo_enabled" \
    go build $build_flags -o "$output_path" ./cmd/musicfox
    
    if [[ $? -ne 0 ]]; then
        log_error "微内核编译失败: $platform/$arch"
        return 1
    fi
    
    # 后处理
    post_process_binary "$output_path" "$platform"
    
    log_success "微内核编译完成: $binary_name"
}

# 编译插件
compile_plugins() {
    local platform="$1"
    local arch="$2"
    
    local combo="$platform,$arch"
    if [[ -z "${PLATFORM_ARCH_SUPPORT[$combo]:-}" ]]; then
        log_warning "不支持的平台组合: $platform/$arch"
        return 0
    fi
    
    log_info "编译插件: $platform/$arch"
    
    # 调用插件构建脚本
    local plugin_args="-p $platform -a $arch -o $OUTPUT_DIR/plugins"
    if [[ $VERBOSE == true ]]; then
        plugin_args="$plugin_args -v"
    fi
    if [[ $DEBUG == true ]]; then
        plugin_args="$plugin_args --debug"
    fi
    
    "$SCRIPT_DIR/build-plugins.sh" $plugin_args
    
    if [[ $? -ne 0 ]]; then
        log_error "插件编译失败: $platform/$arch"
        return 1
    fi
    
    log_success "插件编译完成: $platform/$arch"
}

# 后处理二进制文件
post_process_binary() {
    local binary_path="$1"
    local platform="$2"
    
    if [[ ! -f "$binary_path" ]]; then
        return 0
    fi
    
    # 去除符号表
    if [[ $STRIP == true && $DEBUG != true ]]; then
        case $platform in
            linux|freebsd|openbsd|netbsd)
                if command -v strip &> /dev/null; then
                    strip "$binary_path" 2>/dev/null || true
                fi
                ;;
            darwin)
                if command -v strip &> /dev/null; then
                    strip -x "$binary_path" 2>/dev/null || true
                fi
                ;;
        esac
    fi
    
    # 压缩二进制文件
    if [[ $COMPRESS == true && $platform != "darwin" ]]; then
        log_info "压缩二进制文件: $(basename "$binary_path")"
        upx --best --lzma "$binary_path" 2>/dev/null || {
            log_warning "压缩失败: $(basename "$binary_path")"
        }
    fi
}

# 并行编译
compile_parallel() {
    local platforms=(${PLATFORMS//,/ })
    local archs=(${ARCHS//,/ })
    local pids=()
    
    for platform in "${platforms[@]}"; do
        for arch in "${archs[@]}"; do
            {
                if [[ $KERNEL_ONLY != true ]]; then
                    compile_plugins "$platform" "$arch"
                fi
                
                if [[ $PLUGINS_ONLY != true ]]; then
                    compile_kernel "$platform" "$arch"
                fi
            } &
            pids+=($!)
            
            # 限制并发数
            if [[ ${#pids[@]} -ge 4 ]]; then
                wait "${pids[0]}"
                pids=("${pids[@]:1}")
            fi
        done
    done
    
    # 等待所有任务完成
    for pid in "${pids[@]}"; do
        wait "$pid"
    done
}

# 串行编译
compile_serial() {
    local platforms=(${PLATFORMS//,/ })
    local archs=(${ARCHS//,/ })
    
    for platform in "${platforms[@]}"; do
        for arch in "${archs[@]}"; do
            if [[ $PLUGINS_ONLY != true ]]; then
                compile_kernel "$platform" "$arch"
            fi
            
            if [[ $KERNEL_ONLY != true ]]; then
                compile_plugins "$platform" "$arch"
            fi
        done
    done
}

# 生成编译报告
generate_report() {
    local report_file="$OUTPUT_DIR/cross-compile-report.txt"
    
    log_info "生成编译报告..."
    
    cat > "$report_file" << EOF
go-musicfox v2 交叉编译报告
============================

编译信息:
  版本: $VERSION
  提交: $COMMIT
  编译时间: $BUILD_TIME
  Go版本: $GO_VERSION
  平台: $PLATFORMS
  架构: $ARCHS
  调试模式: $([ $DEBUG == true ] && echo "启用" || echo "禁用")
  并行编译: $([ $PARALLEL == true ] && echo "启用" || echo "禁用")
  压缩: $([ $COMPRESS == true ] && echo "启用" || echo "禁用")
  去除符号: $([ $STRIP == true ] && echo "启用" || echo "禁用")

编译结果:
EOF
    
    # 统计文件
    local total_files=0
    local total_size=0
    
    find "$OUTPUT_DIR" -type f \( -name "go-musicfox-*" -o -name "*.so" -o -name "*.dll" -o -name "*.dylib" -o -name "*.wasm" \) | while read -r file; do
        if [[ -f "$file" ]]; then
            local size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo "0")
            local human_size=$(numfmt --to=iec-i --suffix=B "$size" 2>/dev/null || echo "${size}B")
            echo "  $(basename "$file") ($human_size)" >> "$report_file"
            ((total_files++))
            ((total_size += size))
        fi
    done
    
    local human_total=$(numfmt --to=iec-i --suffix=B "$total_size" 2>/dev/null || echo "${total_size}B")
    
    cat >> "$report_file" << EOF

统计信息:
  总文件数: $total_files
  总大小: $human_total

EOF
    
    log_success "编译报告已生成: $report_file"
}

# 验证编译结果
validate_build() {
    log_info "验证编译结果..."
    
    local platforms=(${PLATFORMS//,/ })
    local archs=(${ARCHS//,/ })
    local failed=false
    
    for platform in "${platforms[@]}"; do
        for arch in "${archs[@]}"; do
            local combo="$platform,$arch"
            if [[ -z "${PLATFORM_ARCH_SUPPORT[$combo]:-}" ]]; then
                continue
            fi
            
            if [[ $PLUGINS_ONLY != true ]]; then
                local binary_name="go-musicfox-$platform-$arch$(get_binary_extension "$platform")"
                if [[ ! -f "$OUTPUT_DIR/$binary_name" ]]; then
                    log_error "缺少二进制文件: $binary_name"
                    failed=true
                fi
            fi
        done
    done
    
    if [[ $failed == true ]]; then
        error_exit "编译验证失败"
    fi
    
    log_success "编译验证通过"
}

# 主函数
main() {
    log_info "开始交叉编译 go-musicfox v2..."
    
    if [[ $VERBOSE == true ]]; then
        log_info "编译配置:"
        echo "  平台: $PLATFORMS"
        echo "  架构: $ARCHS"
        echo "  输出目录: $OUTPUT_DIR"
        echo "  仅微内核: $KERNEL_ONLY"
        echo "  仅插件: $PLUGINS_ONLY"
        echo "  并行编译: $PARALLEL"
    fi
    
    prepare_build
    
    if [[ $PARALLEL == true ]]; then
        compile_parallel
    else
        compile_serial
    fi
    
    validate_build
    generate_report
    
    log_success "交叉编译完成！"
}

# 执行主函数
main "$@"