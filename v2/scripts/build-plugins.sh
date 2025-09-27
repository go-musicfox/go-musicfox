#!/bin/bash
# go-musicfox v2 插件构建脚本
# 支持多种插件类型的构建

set -euo pipefail

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
PLUGINS_DIR="$PROJECT_ROOT/plugins"
EXAMPLES_DIR="$PROJECT_ROOT/examples"
OUTPUT_DIR="$PROJECT_ROOT/build/plugins"
VERSION="2.0.0"
COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# 默认参数
PLUGIN_TYPE="all"
PLATFORM="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
DEBUG=false
VERBOSE=false
PARALLEL=true
VALIDATE=true
CLEAN=false

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
go-musicfox v2 插件构建脚本

用法: $0 [选项] [插件类型]

插件类型:
  all                     构建所有类型插件 (默认)
  dynamic-library         构建动态链接库插件
  rpc                     构建RPC插件
  webassembly            构建WebAssembly插件
  hot-reload             构建热加载插件

选项:
  -h, --help              显示帮助信息
  -p, --platform PLATFORM 目标平台 (linux, darwin, windows)
  -a, --arch ARCH         目标架构 (amd64, arm64)
  -o, --output DIR        输出目录
  -d, --debug             启用调试模式
  -v, --verbose           详细输出
  -j, --parallel          并行构建 (默认)
  -s, --serial            串行构建
  --no-validate           跳过插件验证
  --clean                 构建前清理

示例:
  $0                      # 构建所有插件
  $0 dynamic-library      # 只构建动态链接库插件
  $0 -p linux -a amd64   # 为Linux amd64构建
  $0 --debug --verbose    # 调试模式构建

EOF
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help) show_help; exit 0 ;;
            -p|--platform) PLATFORM="$2"; shift 2 ;;
            -a|--arch) ARCH="$2"; shift 2 ;;
            -o|--output) OUTPUT_DIR="$2"; shift 2 ;;
            -d|--debug) DEBUG=true; shift ;;
            -v|--verbose) VERBOSE=true; shift ;;
            -j|--parallel) PARALLEL=true; shift ;;
            -s|--serial) PARALLEL=false; shift ;;
            --no-validate) VALIDATE=false; shift ;;
            --clean) CLEAN=true; shift ;;
            all|dynamic-library|rpc|webassembly|hot-reload)
                PLUGIN_TYPE="$1"; shift ;;
            -*) error_exit "未知选项: $1" ;;
            *) PLUGIN_TYPE="$1"; shift ;;
        esac
    done
    
    # 架构映射
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
    esac
}

# 准备构建环境
prepare_build() {
    log_info "准备构建环境..."
    
    # 清理输出目录
    if [[ $CLEAN == true ]]; then
        log_info "清理输出目录..."
        rm -rf "$OUTPUT_DIR"
    fi
    
    # 创建输出目录
    mkdir -p "$OUTPUT_DIR"/{dynamic-library,rpc,webassembly,hot-reload}
    
    # 检查Go环境
    if ! command -v go &> /dev/null; then
        error_exit "Go未安装或不在PATH中"
    fi
    
    log_success "构建环境准备完成"
}

# 构建动态链接库插件
build_dynamic_library() {
    log_info "构建动态链接库插件..."
    
    local plugin_dir="$EXAMPLES_DIR/audio-processor"
    if [[ ! -d "$plugin_dir" ]]; then
        log_warning "音频处理插件示例不存在，跳过动态链接库插件构建"
        return 0
    fi
    
    cd "$plugin_dir"
    
    local output_name="audio-processor-${PLATFORM}-${ARCH}"
    case $PLATFORM in
        linux|darwin) output_name="${output_name}.so" ;;
        windows) output_name="${output_name}.dll" ;;
        *) output_name="${output_name}.so" ;;
    esac
    
    local ldflags="-X main.Version=$VERSION -X main.Commit=$COMMIT -X main.BuildTime=$BUILD_TIME"
    if [[ $DEBUG != true ]]; then
        ldflags="$ldflags -s -w"
    fi
    
    local build_flags="-buildmode=plugin -ldflags=\"$ldflags\" -trimpath"
    if [[ $VERBOSE == true ]]; then
        build_flags="$build_flags -v"
    fi
    
    GOOS="$PLATFORM" GOARCH="$ARCH" CGO_ENABLED=1 \
    go build $build_flags -o "$OUTPUT_DIR/dynamic-library/$output_name" .
    
    if [[ $? -eq 0 ]]; then
        log_success "动态链接库插件构建完成: $output_name"
    else
        error_exit "动态链接库插件构建失败"
    fi
}

# 构建RPC插件
build_rpc() {
    log_info "构建RPC插件..."
    
    local plugin_dir="$EXAMPLES_DIR/music-source"
    if [[ ! -d "$plugin_dir" ]]; then
        log_warning "音乐源插件示例不存在，跳过RPC插件构建"
        return 0
    fi
    
    cd "$plugin_dir"
    
    # 构建服务端
    local server_name="music-source-server-${PLATFORM}-${ARCH}"
    if [[ $PLATFORM == "windows" ]]; then
        server_name="${server_name}.exe"
    fi
    
    local ldflags="-X main.Version=$VERSION -X main.Commit=$COMMIT -X main.BuildTime=$BUILD_TIME"
    if [[ $DEBUG != true ]]; then
        ldflags="$ldflags -s -w"
    fi
    
    local build_flags="-ldflags=\"$ldflags\" -trimpath"
    if [[ $VERBOSE == true ]]; then
        build_flags="$build_flags -v"
    fi
    
    GOOS="$PLATFORM" GOARCH="$ARCH" CGO_ENABLED=0 \
    go build $build_flags -o "$OUTPUT_DIR/rpc/$server_name" ./cmd/server
    
    if [[ $? -eq 0 ]]; then
        log_success "RPC插件构建完成: $server_name"
    else
        error_exit "RPC插件构建失败"
    fi
}

# 构建WebAssembly插件
build_webassembly() {
    log_info "构建WebAssembly插件..."
    
    local plugin_dir="$EXAMPLES_DIR/audio-filter"
    if [[ ! -d "$plugin_dir" ]]; then
        log_warning "音频滤波器插件示例不存在，跳过WebAssembly插件构建"
        return 0
    fi
    
    cd "$plugin_dir"
    
    local output_name="audio-filter.wasm"
    local ldflags="-X main.Version=$VERSION -X main.Commit=$COMMIT"
    
    if [[ $DEBUG != true ]]; then
        ldflags="$ldflags -s -w"
    fi
    
    GOOS=wasip1 GOARCH=wasm go build -ldflags="$ldflags" -o "$OUTPUT_DIR/webassembly/$output_name" .
    
    if [[ $? -eq 0 ]]; then
        log_success "WebAssembly插件构建完成: $output_name"
    else
        error_exit "WebAssembly插件构建失败"
    fi
}

# 构建热加载插件
build_hot_reload() {
    log_info "构建热加载插件..."
    
    local plugin_dir="$EXAMPLES_DIR/ui-extension"
    if [[ ! -d "$plugin_dir" ]]; then
        log_warning "UI扩展插件示例不存在，跳过热加载插件构建"
        return 0
    fi
    
    cd "$plugin_dir"
    
    local output_name="ui-extension-${PLATFORM}-${ARCH}"
    if [[ $PLATFORM == "windows" ]]; then
        output_name="${output_name}.exe"
    fi
    
    local ldflags="-X main.Version=$VERSION -X main.Commit=$COMMIT -X main.BuildTime=$BUILD_TIME"
    if [[ $DEBUG != true ]]; then
        ldflags="$ldflags -s -w"
    fi
    
    local build_flags="-ldflags=\"$ldflags\" -trimpath"
    if [[ $VERBOSE == true ]]; then
        build_flags="$build_flags -v"
    fi
    
    GOOS="$PLATFORM" GOARCH="$ARCH" CGO_ENABLED=1 \
    go build $build_flags -o "$OUTPUT_DIR/hot-reload/$output_name" .
    
    if [[ $? -eq 0 ]]; then
        log_success "热加载插件构建完成: $output_name"
    else
        error_exit "热加载插件构建失败"
    fi
}

# 验证插件
validate_plugins() {
    if [[ $VALIDATE != true ]]; then
        return 0
    fi
    
    log_info "验证构建的插件..."
    
    local validator_script="$PROJECT_ROOT/tools/plugin-validator.sh"
    if [[ ! -f "$validator_script" ]]; then
        log_warning "插件验证脚本不存在，跳过验证"
        return 0
    fi
    
    find "$OUTPUT_DIR" -type f \( -name "*.so" -o -name "*.dll" -o -name "*.dylib" -o -name "*.wasm" \) | while read -r plugin_file; do
        log_info "验证插件: $(basename "$plugin_file")"
        "$validator_script" "$plugin_file" || log_warning "插件验证失败: $plugin_file"
    done
}

# 生成插件清单
generate_manifest() {
    local manifest_file="$OUTPUT_DIR/plugins-manifest.json"
    
    log_info "生成插件清单..."
    
    cat > "$manifest_file" << 'EOF'
{
  "version": "VERSION_PLACEHOLDER",
  "build_time": "BUILD_TIME_PLACEHOLDER",
  "commit": "COMMIT_PLACEHOLDER",
  "platform": "PLATFORM_PLACEHOLDER",
  "arch": "ARCH_PLACEHOLDER",
  "plugins": [
EOF
    
    # 替换占位符
    sed -i.bak "s/VERSION_PLACEHOLDER/$VERSION/g; s/BUILD_TIME_PLACEHOLDER/$BUILD_TIME/g; s/COMMIT_PLACEHOLDER/$COMMIT/g; s/PLATFORM_PLACEHOLDER/$PLATFORM/g; s/ARCH_PLACEHOLDER/$ARCH/g" "$manifest_file"
    rm -f "$manifest_file.bak"
    
    local first=true
    find "$OUTPUT_DIR" -type f \( -name "*.so" -o -name "*.dll" -o -name "*.dylib" -o -name "*.wasm" -o -name "*.exe" \) | while read -r plugin_file; do
        if [[ $first == true ]]; then
            first=false
        else
            echo "," >> "$manifest_file"
        fi
        
        local plugin_name=$(basename "$plugin_file")
        local plugin_size=$(stat -f%z "$plugin_file" 2>/dev/null || stat -c%s "$plugin_file" 2>/dev/null || echo "0")
        local plugin_type="unknown"
        
        case "$plugin_name" in
            *.so|*.dll|*.dylib) plugin_type="dynamic-library" ;;
            *.wasm) plugin_type="webassembly" ;;
            *server*) plugin_type="rpc-server" ;;
            *client*) plugin_type="rpc-client" ;;
            *) plugin_type="hot-reload" ;;
        esac
        
        cat >> "$manifest_file" << EOF
    {
      "name": "$plugin_name",
      "type": "$plugin_type",
      "size": $plugin_size,
      "path": "$(basename "$plugin_file")"
    }
EOF
    done
    
    echo "" >> "$manifest_file"
    echo "  ]" >> "$manifest_file"
    echo "}" >> "$manifest_file"
    
    log_success "插件清单已生成: $manifest_file"
}

# 并行构建函数
build_parallel() {
    local pids=()
    
    case $PLUGIN_TYPE in
        all)
            build_dynamic_library & pids+=($!)
            build_rpc & pids+=($!)
            build_webassembly & pids+=($!)
            build_hot_reload & pids+=($!)
            ;;
        dynamic-library) build_dynamic_library & pids+=($!) ;;
        rpc) build_rpc & pids+=($!) ;;
        webassembly) build_webassembly & pids+=($!) ;;
        hot-reload) build_hot_reload & pids+=($!) ;;
    esac
    
    # 等待所有后台任务完成
    local failed=false
    for pid in "${pids[@]}"; do
        if ! wait "$pid"; then
            failed=true
        fi
    done
    
    if [[ $failed == true ]]; then
        error_exit "部分插件构建失败"
    fi
}

# 串行构建函数
build_serial() {
    case $PLUGIN_TYPE in
        all)
            build_dynamic_library
            build_rpc
            build_webassembly
            build_hot_reload
            ;;
        dynamic-library) build_dynamic_library ;;
        rpc) build_rpc ;;
        webassembly) build_webassembly ;;
        hot-reload) build_hot_reload ;;
    esac
}

# 主函数
main() {
    log_info "开始构建插件 ($PLUGIN_TYPE)..."
    
    prepare_build
    
    if [[ $PARALLEL == true ]]; then
        build_parallel
    else
        build_serial
    fi
    
    validate_plugins
    generate_manifest
    
    log_success "插件构建完成！"
}

# 解析参数并执行主函数
parse_args "$@"
main