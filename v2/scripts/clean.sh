#!/bin/bash
# go-musicfox v2 清理脚本
# 清理构建产物、缓存和临时文件

set -euo pipefail

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"
DIST_DIR="$PROJECT_ROOT/dist"
CACHE_DIR="$PROJECT_ROOT/.cache"
TMP_DIR="$PROJECT_ROOT/tmp"
COVERAGE_DIR="$PROJECT_ROOT/coverage"
LOG_DIR="$PROJECT_ROOT/logs"

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
go-musicfox v2 清理脚本

用法: $0 [选项]

选项:
  -h, --help          显示帮助信息
  -v, --verbose       详细输出
  -a, --all           清理所有文件（包括依赖缓存）
  -b, --build         仅清理构建产物
  -c, --cache         仅清理缓存文件
  -d, --deps          清理依赖缓存
  -l, --logs          清理日志文件
  -t, --temp          清理临时文件
  -p, --plugins       清理插件构建产物
  -f, --force         强制清理（不询问确认）
  --dry-run           预览清理操作（不实际删除）

清理范围:
  构建产物:    $BUILD_DIR
  发布包:      $DIST_DIR
  缓存文件:    $CACHE_DIR
  临时文件:    $TMP_DIR
  覆盖率报告:  $COVERAGE_DIR
  日志文件:    $LOG_DIR
  Go模块缓存:  \$(go env GOMODCACHE)
  Go构建缓存:  \$(go env GOCACHE)

示例:
  $0                    # 交互式清理
  $0 -a                 # 清理所有文件
  $0 -b -p              # 仅清理构建产物和插件
  $0 --dry-run          # 预览清理操作

EOF
}

# 默认配置
VERBOSE=false
CLEAN_ALL=false
CLEAN_BUILD=false
CLEAN_CACHE=false
CLEAN_DEPS=false
CLEAN_LOGS=false
CLEAN_TEMP=false
CLEAN_PLUGINS=false
FORCE=false
DRY_RUN=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help) show_help; exit 0 ;;
        -v|--verbose) VERBOSE=true; shift ;;
        -a|--all) CLEAN_ALL=true; shift ;;
        -b|--build) CLEAN_BUILD=true; shift ;;
        -c|--cache) CLEAN_CACHE=true; shift ;;
        -d|--deps) CLEAN_DEPS=true; shift ;;
        -l|--logs) CLEAN_LOGS=true; shift ;;
        -t|--temp) CLEAN_TEMP=true; shift ;;
        -p|--plugins) CLEAN_PLUGINS=true; shift ;;
        -f|--force) FORCE=true; shift ;;
        --dry-run) DRY_RUN=true; shift ;;
        -*) error_exit "未知选项: $1" ;;
        *) error_exit "无效参数: $1" ;;
    esac
done

# 如果没有指定具体选项，启用交互模式
if [[ $CLEAN_ALL == false && $CLEAN_BUILD == false && $CLEAN_CACHE == false && \
      $CLEAN_DEPS == false && $CLEAN_LOGS == false && $CLEAN_TEMP == false && \
      $CLEAN_PLUGINS == false ]]; then
    INTERACTIVE=true
else
    INTERACTIVE=false
fi

# 获取目录大小
get_dir_size() {
    local dir="$1"
    if [[ -d "$dir" ]]; then
        if command -v du &> /dev/null; then
            du -sh "$dir" 2>/dev/null | cut -f1 || echo "0B"
        else
            echo "unknown"
        fi
    else
        echo "0B"
    fi
}

# 安全删除目录
safe_remove() {
    local path="$1"
    local description="$2"
    
    if [[ ! -e "$path" ]]; then
        if [[ $VERBOSE == true ]]; then
            log_info "跳过不存在的路径: $path"
        fi
        return 0
    fi
    
    local size=$(get_dir_size "$path")
    
    if [[ $DRY_RUN == true ]]; then
        log_info "[DRY-RUN] 将删除 $description: $path ($size)"
        return 0
    fi
    
    if [[ $FORCE == false && $INTERACTIVE == true ]]; then
        echo -n "删除 $description ($size)? [y/N] "
        read -r response
        case $response in
            [yY]|[yY][eE][sS]) ;;
            *) log_info "跳过删除: $description"; return 0 ;;
        esac
    fi
    
    log_info "删除 $description: $path ($size)"
    
    if [[ -d "$path" ]]; then
        rm -rf "$path"
    elif [[ -f "$path" ]]; then
        rm -f "$path"
    fi
    
    if [[ $? -eq 0 ]]; then
        log_success "已删除: $description"
    else
        log_error "删除失败: $description"
    fi
}

# 清理构建产物
clean_build() {
    log_info "清理构建产物..."
    
    # 清理主构建目录
    safe_remove "$BUILD_DIR" "构建产物目录"
    
    # 清理Go构建产物
    find "$PROJECT_ROOT" -name "*.exe" -o -name "go-musicfox" -o -name "go-musicfox-*" | while read -r file; do
        if [[ -f "$file" && "$file" != *"/vendor/"* ]]; then
            safe_remove "$file" "二进制文件"
        fi
    done
    
    # 清理对象文件
    find "$PROJECT_ROOT" -name "*.o" -o -name "*.a" -o -name "*.so" -o -name "*.dylib" -o -name "*.dll" | while read -r file; do
        if [[ -f "$file" && "$file" != *"/vendor/"* ]]; then
            safe_remove "$file" "对象文件"
        fi
    done
}

# 清理插件构建产物
clean_plugins() {
    log_info "清理插件构建产物..."
    
    # 清理插件构建目录
    if [[ -d "$BUILD_DIR/plugins" ]]; then
        safe_remove "$BUILD_DIR/plugins" "插件构建目录"
    fi
    
    # 清理WebAssembly文件
    find "$PROJECT_ROOT" -name "*.wasm" | while read -r file; do
        if [[ -f "$file" && "$file" != *"/vendor/"* ]]; then
            safe_remove "$file" "WebAssembly文件"
        fi
    done
    
    # 清理插件临时文件
    find "$PROJECT_ROOT" -name "plugin_*.tmp" -o -name "*.plugin" | while read -r file; do
        if [[ -f "$file" ]]; then
            safe_remove "$file" "插件临时文件"
        fi
    done
}

# 清理发布包
clean_dist() {
    log_info "清理发布包..."
    safe_remove "$DIST_DIR" "发布包目录"
}

# 清理缓存文件
clean_cache() {
    log_info "清理缓存文件..."
    
    safe_remove "$CACHE_DIR" "缓存目录"
    
    # 清理测试缓存
    if command -v go &> /dev/null; then
        if [[ $DRY_RUN == false ]]; then
            log_info "清理Go测试缓存..."
            go clean -testcache
        else
            log_info "[DRY-RUN] 将清理Go测试缓存"
        fi
    fi
}

# 清理依赖缓存
clean_deps() {
    log_info "清理依赖缓存..."
    
    if command -v go &> /dev/null; then
        local mod_cache=$(go env GOMODCACHE 2>/dev/null || echo "")
        local build_cache=$(go env GOCACHE 2>/dev/null || echo "")
        
        if [[ -n "$mod_cache" && -d "$mod_cache" ]]; then
            if [[ $DRY_RUN == false ]]; then
                if [[ $FORCE == false && $INTERACTIVE == true ]]; then
                    local size=$(get_dir_size "$mod_cache")
                    echo -n "清理Go模块缓存 ($size)? [y/N] "
                    read -r response
                    case $response in
                        [yY]|[yY][eE][sS]) go clean -modcache ;;
                        *) log_info "跳过Go模块缓存清理" ;;
                    esac
                else
                    log_info "清理Go模块缓存..."
                    go clean -modcache
                fi
            else
                local size=$(get_dir_size "$mod_cache")
                log_info "[DRY-RUN] 将清理Go模块缓存: $mod_cache ($size)"
            fi
        fi
        
        if [[ -n "$build_cache" && -d "$build_cache" ]]; then
            if [[ $DRY_RUN == false ]]; then
                log_info "清理Go构建缓存..."
                go clean -cache
            else
                local size=$(get_dir_size "$build_cache")
                log_info "[DRY-RUN] 将清理Go构建缓存: $build_cache ($size)"
            fi
        fi
    fi
    
    # 清理vendor目录（如果存在且不是必需的）
    if [[ -d "$PROJECT_ROOT/vendor" ]]; then
        if [[ $FORCE == false && $INTERACTIVE == true ]]; then
            echo -n "删除vendor目录? [y/N] "
            read -r response
            case $response in
                [yY]|[yY][eE][sS]) safe_remove "$PROJECT_ROOT/vendor" "vendor目录" ;;
                *) log_info "保留vendor目录" ;;
            esac
        else
            safe_remove "$PROJECT_ROOT/vendor" "vendor目录"
        fi
    fi
}

# 清理临时文件
clean_temp() {
    log_info "清理临时文件..."
    
    safe_remove "$TMP_DIR" "临时文件目录"
    
    # 清理系统临时文件
    find "$PROJECT_ROOT" -name "*.tmp" -o -name "*.temp" -o -name "*~" -o -name ".DS_Store" | while read -r file; do
        if [[ -f "$file" ]]; then
            safe_remove "$file" "临时文件"
        fi
    done
    
    # 清理编辑器临时文件
    find "$PROJECT_ROOT" -name "*.swp" -o -name "*.swo" -o -name "*~" -o -name ".#*" | while read -r file; do
        if [[ -f "$file" ]]; then
            safe_remove "$file" "编辑器临时文件"
        fi
    done
}

# 清理日志文件
clean_logs() {
    log_info "清理日志文件..."
    
    safe_remove "$LOG_DIR" "日志目录"
    
    # 清理覆盖率报告
    safe_remove "$COVERAGE_DIR" "覆盖率报告目录"
    
    # 清理其他日志文件
    find "$PROJECT_ROOT" -name "*.log" -o -name "coverage.out" -o -name "profile.out" | while read -r file; do
        if [[ -f "$file" && "$file" != *"/vendor/"* ]]; then
            safe_remove "$file" "日志文件"
        fi
    done
}

# 显示清理摘要
show_summary() {
    log_info "清理操作摘要:"
    
    local total_cleaned=0
    
    if [[ $CLEAN_ALL == true || $CLEAN_BUILD == true ]]; then
        echo "  ✓ 构建产物"
        ((total_cleaned++))
    fi
    
    if [[ $CLEAN_ALL == true || $CLEAN_PLUGINS == true ]]; then
        echo "  ✓ 插件构建产物"
        ((total_cleaned++))
    fi
    
    if [[ $CLEAN_ALL == true || $CLEAN_CACHE == true ]]; then
        echo "  ✓ 缓存文件"
        ((total_cleaned++))
    fi
    
    if [[ $CLEAN_ALL == true || $CLEAN_DEPS == true ]]; then
        echo "  ✓ 依赖缓存"
        ((total_cleaned++))
    fi
    
    if [[ $CLEAN_ALL == true || $CLEAN_TEMP == true ]]; then
        echo "  ✓ 临时文件"
        ((total_cleaned++))
    fi
    
    if [[ $CLEAN_ALL == true || $CLEAN_LOGS == true ]]; then
        echo "  ✓ 日志文件"
        ((total_cleaned++))
    fi
    
    if [[ $total_cleaned -eq 0 ]]; then
        log_warning "没有执行任何清理操作"
    else
        log_success "清理完成！共执行了 $total_cleaned 项清理操作"
    fi
}

# 交互式清理
interactive_clean() {
    log_info "交互式清理模式"
    echo
    
    # 显示当前状态
    log_info "当前磁盘使用情况:"
    [[ -d "$BUILD_DIR" ]] && echo "  构建产物: $(get_dir_size "$BUILD_DIR")"
    [[ -d "$DIST_DIR" ]] && echo "  发布包: $(get_dir_size "$DIST_DIR")"
    [[ -d "$CACHE_DIR" ]] && echo "  缓存文件: $(get_dir_size "$CACHE_DIR")"
    [[ -d "$TMP_DIR" ]] && echo "  临时文件: $(get_dir_size "$TMP_DIR")"
    [[ -d "$LOG_DIR" ]] && echo "  日志文件: $(get_dir_size "$LOG_DIR")"
    [[ -d "$COVERAGE_DIR" ]] && echo "  覆盖率报告: $(get_dir_size "$COVERAGE_DIR")"
    
    if command -v go &> /dev/null; then
        local mod_cache=$(go env GOMODCACHE 2>/dev/null || echo "")
        local build_cache=$(go env GOCACHE 2>/dev/null || echo "")
        [[ -n "$mod_cache" && -d "$mod_cache" ]] && echo "  Go模块缓存: $(get_dir_size "$mod_cache")"
        [[ -n "$build_cache" && -d "$build_cache" ]] && echo "  Go构建缓存: $(get_dir_size "$build_cache")"
    fi
    
    echo
    echo "请选择要清理的内容:"
    echo "  1) 构建产物"
    echo "  2) 插件构建产物"
    echo "  3) 发布包"
    echo "  4) 缓存文件"
    echo "  5) 依赖缓存"
    echo "  6) 临时文件"
    echo "  7) 日志文件"
    echo "  8) 全部清理"
    echo "  0) 退出"
    echo
    
    while true; do
        echo -n "请输入选项 (可多选，用空格分隔): "
        read -r choices
        
        if [[ -z "$choices" ]]; then
            continue
        fi
        
        for choice in $choices; do
            case $choice in
                1) clean_build ;;
                2) clean_plugins ;;
                3) clean_dist ;;
                4) clean_cache ;;
                5) clean_deps ;;
                6) clean_temp ;;
                7) clean_logs ;;
                8) 
                    clean_build
                    clean_plugins
                    clean_dist
                    clean_cache
                    clean_deps
                    clean_temp
                    clean_logs
                    ;;
                0) log_info "退出清理"; exit 0 ;;
                *) log_warning "无效选项: $choice" ;;
            esac
        done
        
        break
    done
}

# 主函数
main() {
    log_info "go-musicfox v2 清理工具"
    
    # 检查项目根目录
    if [[ ! -f "$PROJECT_ROOT/go.mod" ]]; then
        error_exit "未找到go.mod文件，请在项目根目录运行此脚本"
    fi
    
    if [[ $DRY_RUN == true ]]; then
        log_warning "预览模式 - 不会实际删除文件"
    fi
    
    if [[ $INTERACTIVE == true ]]; then
        interactive_clean
    else
        # 执行指定的清理操作
        if [[ $CLEAN_ALL == true ]]; then
            clean_build
            clean_plugins
            clean_dist
            clean_cache
            clean_deps
            clean_temp
            clean_logs
        else
            [[ $CLEAN_BUILD == true ]] && clean_build
            [[ $CLEAN_PLUGINS == true ]] && clean_plugins
            [[ $CLEAN_CACHE == true ]] && clean_cache
            [[ $CLEAN_DEPS == true ]] && clean_deps
            [[ $CLEAN_TEMP == true ]] && clean_temp
            [[ $CLEAN_LOGS == true ]] && clean_logs
        fi
        
        # 总是清理发布包（除非明确指定不清理）
        clean_dist
    fi
    
    show_summary
}

# 执行主函数
main "$@"