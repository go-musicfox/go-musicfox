#!/bin/bash
# go-musicfox v2 插件验证器
# 验证插件的完整性、兼容性和安全性

set -euo pipefail

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
PLUGINS_DIR="$PROJECT_ROOT/plugins"
CORE_BINARY="$PROJECT_ROOT/build/go-musicfox-linux-amd64"

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
go-musicfox v2 插件验证器

用法: $0 [选项] [插件路径...]

选项:
  -h, --help              显示帮助信息
  -t, --type TYPE         验证指定类型的插件 (shared|rpc|wasm|hotload|all)
  -l, --level LEVEL       验证级别 (basic|standard|strict)
  -o, --output FORMAT     输出格式 (text|json|xml)
  -r, --report FILE       生成验证报告文件
  -f, --fix               尝试自动修复问题
  -v, --verbose           详细输出
  -q, --quiet             静默模式
  --no-security           跳过安全检查
  --no-compatibility      跳过兼容性检查
  --no-performance        跳过性能检查
  --timeout SECONDS       验证超时时间 (默认: 30)

验证级别:
  basic                   基础验证（文件存在、格式正确）
  standard                标准验证（包含接口检查、依赖验证）
  strict                  严格验证（包含安全扫描、性能测试）

输出格式:
  text                    文本格式 (默认)
  json                    JSON格式
  xml                     XML格式

示例:
  $0 plugins/shared/my-plugin.so                    # 验证单个插件
  $0 -t shared plugins/shared/                      # 验证所有动态链接库插件
  $0 -l strict -o json -r report.json plugins/     # 严格验证并生成JSON报告
  $0 --fix plugins/wasm/broken-plugin.wasm         # 验证并尝试修复插件

EOF
}

# 解析命令行参数
parse_args() {
    PLUGIN_TYPE="all"
    VALIDATION_LEVEL="standard"
    OUTPUT_FORMAT="text"
    REPORT_FILE=""
    AUTO_FIX=false
    VERBOSE=false
    QUIET=false
    SKIP_SECURITY=false
    SKIP_COMPATIBILITY=false
    SKIP_PERFORMANCE=false
    TIMEOUT=30
    PLUGIN_PATHS=()
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help) show_help; exit 0 ;;
            -t|--type) PLUGIN_TYPE="$2"; shift 2 ;;
            -l|--level) VALIDATION_LEVEL="$2"; shift 2 ;;
            -o|--output) OUTPUT_FORMAT="$2"; shift 2 ;;
            -r|--report) REPORT_FILE="$2"; shift 2 ;;
            -f|--fix) AUTO_FIX=true; shift ;;
            -v|--verbose) VERBOSE=true; shift ;;
            -q|--quiet) QUIET=true; shift ;;
            --no-security) SKIP_SECURITY=true; shift ;;
            --no-compatibility) SKIP_COMPATIBILITY=true; shift ;;
            --no-performance) SKIP_PERFORMANCE=true; shift ;;
            --timeout) TIMEOUT="$2"; shift 2 ;;
            -*) error_exit "未知选项: $1" ;;
            *) PLUGIN_PATHS+=("$1"); shift ;;
        esac
    done
    
    # 验证参数
    case $PLUGIN_TYPE in
        shared|rpc|wasm|hotload|all) ;;
        *) error_exit "无效的插件类型: $PLUGIN_TYPE" ;;
    esac
    
    case $VALIDATION_LEVEL in
        basic|standard|strict) ;;
        *) error_exit "无效的验证级别: $VALIDATION_LEVEL" ;;
    esac
    
    case $OUTPUT_FORMAT in
        text|json|xml) ;;
        *) error_exit "无效的输出格式: $OUTPUT_FORMAT" ;;
    esac
    
    # 如果没有指定插件路径，使用默认插件目录
    if [[ ${#PLUGIN_PATHS[@]} -eq 0 ]]; then
        if [[ $PLUGIN_TYPE == "all" ]]; then
            PLUGIN_PATHS=("$PLUGINS_DIR")
        else
            PLUGIN_PATHS=("$PLUGINS_DIR/$PLUGIN_TYPE")
        fi
    fi
}

# 初始化验证环境
init_validation() {
    log_info "初始化验证环境..."
    
    # 检查必要工具
    local missing_tools=()
    
    if ! command -v file &> /dev/null; then
        missing_tools+=("file")
    fi
    
    if ! command -v objdump &> /dev/null && [[ $PLUGIN_TYPE == "shared" || $PLUGIN_TYPE == "all" ]]; then
        missing_tools+=("objdump")
    fi
    
    if ! command -v wasm-validate &> /dev/null && [[ $PLUGIN_TYPE == "wasm" || $PLUGIN_TYPE == "all" ]]; then
        log_warning "wasm-validate未安装，将跳过WASM验证"
    fi
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        error_exit "缺少必要工具: ${missing_tools[*]}"
    fi
    
    # 创建临时目录
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT
    
    # 初始化验证结果
    VALIDATION_RESULTS=()
    TOTAL_PLUGINS=0
    PASSED_PLUGINS=0
    FAILED_PLUGINS=0
    WARNING_PLUGINS=0
}

# 查找插件文件
find_plugins() {
    local search_paths=("${PLUGIN_PATHS[@]}")
    local found_plugins=()
    
    for path in "${search_paths[@]}"; do
        if [[ -f "$path" ]]; then
            # 单个文件
            found_plugins+=("$path")
        elif [[ -d "$path" ]]; then
            # 目录，查找插件文件
            case $PLUGIN_TYPE in
                shared|all)
                    while IFS= read -r -d '' file; do
                        found_plugins+=("$file")
                    done < <(find "$path" -name "*.so" -o -name "*.dll" -o -name "*.dylib" -print0 2>/dev/null)
                    ;;
                wasm|all)
                    while IFS= read -r -d '' file; do
                        found_plugins+=("$file")
                    done < <(find "$path" -name "*.wasm" -print0 2>/dev/null)
                    ;;
                rpc|hotload|all)
                    # RPC和热加载插件通常是可执行文件
                    while IFS= read -r -d '' file; do
                        if [[ -x "$file" && ! -d "$file" ]]; then
                            found_plugins+=("$file")
                        fi
                    done < <(find "$path" -type f -executable -print0 2>/dev/null)
                    ;;
            esac
        else
            log_warning "路径不存在: $path"
        fi
    done
    
    PLUGIN_FILES=("${found_plugins[@]}")
    TOTAL_PLUGINS=${#PLUGIN_FILES[@]}
    
    if [[ $TOTAL_PLUGINS -eq 0 ]]; then
        error_exit "未找到任何插件文件"
    fi
    
    log_info "找到 $TOTAL_PLUGINS 个插件文件"
}

# 验证单个插件
validate_plugin() {
    local plugin_file="$1"
    local plugin_name=$(basename "$plugin_file")
    local plugin_type=$(detect_plugin_type "$plugin_file")
    
    if [[ $VERBOSE == true ]]; then
        log_info "验证插件: $plugin_name ($plugin_type)"
    fi
    
    local result={
        "file": "$plugin_file",
        "name": "$plugin_name",
        "type": "$plugin_type",
        "status": "unknown",
        "issues": [],
        "warnings": [],
        "info": {}
    }
    
    local issues=()
    local warnings=()
    local status="passed"
    
    # 基础验证
    validate_basic "$plugin_file" "$plugin_type" issues warnings
    
    # 标准验证
    if [[ $VALIDATION_LEVEL == "standard" || $VALIDATION_LEVEL == "strict" ]]; then
        validate_standard "$plugin_file" "$plugin_type" issues warnings
    fi
    
    # 严格验证
    if [[ $VALIDATION_LEVEL == "strict" ]]; then
        validate_strict "$plugin_file" "$plugin_type" issues warnings
    fi
    
    # 确定状态
    if [[ ${#issues[@]} -gt 0 ]]; then
        status="failed"
        ((FAILED_PLUGINS++))
    elif [[ ${#warnings[@]} -gt 0 ]]; then
        status="warning"
        ((WARNING_PLUGINS++))
    else
        status="passed"
        ((PASSED_PLUGINS++))
    fi
    
    # 保存结果
    local result_json=$(jq -n \
        --arg file "$plugin_file" \
        --arg name "$plugin_name" \
        --arg type "$plugin_type" \
        --arg status "$status" \
        --argjson issues "$(printf '%s\n' "${issues[@]}" | jq -R . | jq -s .)" \
        --argjson warnings "$(printf '%s\n' "${warnings[@]}" | jq -R . | jq -s .)" \
        '{
            file: $file,
            name: $name,
            type: $type,
            status: $status,
            issues: $issues,
            warnings: $warnings,
            info: {}
        }')
    
    VALIDATION_RESULTS+=("$result_json")
    
    # 输出结果
    if [[ $QUIET == false ]]; then
        case $status in
            passed)
                log_success "✓ $plugin_name"
                ;;
            warning)
                log_warning "⚠ $plugin_name (${#warnings[@]} warnings)"
                if [[ $VERBOSE == true ]]; then
                    for warning in "${warnings[@]}"; do
                        log_warning "  - $warning"
                    done
                fi
                ;;
            failed)
                log_error "✗ $plugin_name (${#issues[@]} issues)"
                if [[ $VERBOSE == true ]]; then
                    for issue in "${issues[@]}"; do
                        log_error "  - $issue"
                    done
                fi
                ;;
        esac
    fi
    
    # 自动修复
    if [[ $AUTO_FIX == true && $status == "failed" ]]; then
        attempt_fix "$plugin_file" "$plugin_type" "${issues[@]}"
    fi
}

# 检测插件类型
detect_plugin_type() {
    local plugin_file="$1"
    local file_info=$(file "$plugin_file")
    
    if [[ $file_info == *"WebAssembly"* ]]; then
        echo "wasm"
    elif [[ $file_info == *"shared object"* || $file_info == *"dynamic library"* ]]; then
        echo "shared"
    elif [[ -x "$plugin_file" ]]; then
        # 检查是否是RPC或热加载插件
        if strings "$plugin_file" 2>/dev/null | grep -q "rpc\|RPC"; then
            echo "rpc"
        elif strings "$plugin_file" 2>/dev/null | grep -q "hotload\|hot.reload"; then
            echo "hotload"
        else
            echo "executable"
        fi
    else
        echo "unknown"
    fi
}

# 基础验证
validate_basic() {
    local plugin_file="$1"
    local plugin_type="$2"
    local -n issues_ref=$3
    local -n warnings_ref=$4
    
    # 文件存在性检查
    if [[ ! -f "$plugin_file" ]]; then
        issues_ref+=("文件不存在")
        return
    fi
    
    # 文件可读性检查
    if [[ ! -r "$plugin_file" ]]; then
        issues_ref+=("文件不可读")
        return
    fi
    
    # 文件大小检查
    local file_size=$(stat -c%s "$plugin_file" 2>/dev/null || stat -f%z "$plugin_file" 2>/dev/null)
    if [[ $file_size -eq 0 ]]; then
        issues_ref+=("文件为空")
        return
    fi
    
    if [[ $file_size -gt 104857600 ]]; then  # 100MB
        warnings_ref+=("文件过大 ($(numfmt --to=iec $file_size))")
    fi
    
    # 文件格式验证
    case $plugin_type in
        shared)
            validate_shared_format "$plugin_file" issues_ref warnings_ref
            ;;
        wasm)
            validate_wasm_format "$plugin_file" issues_ref warnings_ref
            ;;
        rpc|hotload)
            validate_executable_format "$plugin_file" issues_ref warnings_ref
            ;;
    esac
}

# 验证动态链接库格式
validate_shared_format() {
    local plugin_file="$1"
    local -n issues_ref=$2
    local -n warnings_ref=$3
    
    local file_info=$(file "$plugin_file")
    
    if [[ $file_info != *"shared object"* && $file_info != *"dynamic library"* ]]; then
        issues_ref+=("不是有效的动态链接库")
        return
    fi
    
    # 检查必要的符号
    if command -v objdump &> /dev/null; then
        local symbols=$(objdump -T "$plugin_file" 2>/dev/null | grep -E "(Initialize|Execute|Cleanup|GetInfo)" || true)
        if [[ -z "$symbols" ]]; then
            warnings_ref+=("未找到标准插件接口符号")
        fi
    fi
    
    # 检查依赖
    if command -v ldd &> /dev/null; then
        local missing_deps=$(ldd "$plugin_file" 2>/dev/null | grep "not found" || true)
        if [[ -n "$missing_deps" ]]; then
            issues_ref+=("存在缺失的依赖库")
        fi
    fi
}

# 验证WebAssembly格式
validate_wasm_format() {
    local plugin_file="$1"
    local -n issues_ref=$2
    local -n warnings_ref=$3
    
    local file_info=$(file "$plugin_file")
    
    if [[ $file_info != *"WebAssembly"* ]]; then
        issues_ref+=("不是有效的WebAssembly文件")
        return
    fi
    
    # 使用wasm-validate验证
    if command -v wasm-validate &> /dev/null; then
        if ! wasm-validate "$plugin_file" &>/dev/null; then
            issues_ref+=("WebAssembly格式验证失败")
        fi
    fi
    
    # 检查文件头
    local magic=$(hexdump -C "$plugin_file" | head -1 | cut -d' ' -f2-5)
    if [[ "$magic" != "00 61 73 6d" ]]; then
        issues_ref+=("WebAssembly魔数不正确")
    fi
}

# 验证可执行文件格式
validate_executable_format() {
    local plugin_file="$1"
    local -n issues_ref=$2
    local -n warnings_ref=$3
    
    if [[ ! -x "$plugin_file" ]]; then
        issues_ref+=("文件不可执行")
        return
    fi
    
    local file_info=$(file "$plugin_file")
    
    if [[ $file_info != *"executable"* ]]; then
        warnings_ref+=("可能不是可执行文件")
    fi
}

# 标准验证
validate_standard() {
    local plugin_file="$1"
    local plugin_type="$2"
    local -n issues_ref=$3
    local -n warnings_ref=$4
    
    # 接口兼容性检查
    if [[ $SKIP_COMPATIBILITY == false ]]; then
        validate_compatibility "$plugin_file" "$plugin_type" issues_ref warnings_ref
    fi
    
    # 依赖检查
    validate_dependencies "$plugin_file" "$plugin_type" issues_ref warnings_ref
    
    # 元数据检查
    validate_metadata "$plugin_file" "$plugin_type" issues_ref warnings_ref
}

# 兼容性验证
validate_compatibility() {
    local plugin_file="$1"
    local plugin_type="$2"
    local -n issues_ref=$3
    local -n warnings_ref=$4
    
    # 检查核心二进制文件
    if [[ ! -f "$CORE_BINARY" ]]; then
        warnings_ref+=("核心二进制文件不存在，跳过兼容性检查")
        return
    fi
    
    case $plugin_type in
        shared)
            # 检查ABI兼容性
            local plugin_arch=$(file "$plugin_file" | grep -oE "(x86-64|i386|ARM|aarch64)")
            local core_arch=$(file "$CORE_BINARY" | grep -oE "(x86-64|i386|ARM|aarch64)")
            
            if [[ "$plugin_arch" != "$core_arch" ]]; then
                issues_ref+=("架构不兼容: 插件($plugin_arch) vs 核心($core_arch)")
            fi
            ;;
        rpc|hotload)
            # 尝试启动插件进行兼容性测试
            timeout $TIMEOUT "$plugin_file" --version &>/dev/null || \
                warnings_ref+=("插件启动测试失败")
            ;;
    esac
}

# 依赖验证
validate_dependencies() {
    local plugin_file="$1"
    local plugin_type="$2"
    local -n issues_ref=$3
    local -n warnings_ref=$4
    
    case $plugin_type in
        shared)
            if command -v ldd &> /dev/null; then
                local deps=$(ldd "$plugin_file" 2>/dev/null || true)
                local missing=$(echo "$deps" | grep "not found" || true)
                
                if [[ -n "$missing" ]]; then
                    issues_ref+=("缺失依赖: $(echo "$missing" | cut -d' ' -f1 | tr '\n' ' ')")
                fi
                
                # 检查危险依赖
                local dangerous=$(echo "$deps" | grep -E "(libssl|libcrypto|libcurl)" || true)
                if [[ -n "$dangerous" ]]; then
                    warnings_ref+=("包含潜在安全风险的依赖")
                fi
            fi
            ;;
    esac
}

# 元数据验证
validate_metadata() {
    local plugin_file="$1"
    local plugin_type="$2"
    local -n issues_ref=$3
    local -n warnings_ref=$4
    
    # 检查插件信息
    case $plugin_type in
        shared)
            if command -v strings &> /dev/null; then
                local strings_output=$(strings "$plugin_file")
                
                # 检查版本信息
                if ! echo "$strings_output" | grep -qE "v?[0-9]+\.[0-9]+\.[0-9]+"; then
                    warnings_ref+=("未找到版本信息")
                fi
                
                # 检查插件名称
                if ! echo "$strings_output" | grep -qE "(plugin|Plugin|PLUGIN)"; then
                    warnings_ref+=("未找到插件标识")
                fi
            fi
            ;;
    esac
}

# 严格验证
validate_strict() {
    local plugin_file="$1"
    local plugin_type="$2"
    local -n issues_ref=$3
    local -n warnings_ref=$4
    
    # 安全检查
    if [[ $SKIP_SECURITY == false ]]; then
        validate_security "$plugin_file" "$plugin_type" issues_ref warnings_ref
    fi
    
    # 性能检查
    if [[ $SKIP_PERFORMANCE == false ]]; then
        validate_performance "$plugin_file" "$plugin_type" issues_ref warnings_ref
    fi
    
    # 代码质量检查
    validate_code_quality "$plugin_file" "$plugin_type" issues_ref warnings_ref
}

# 安全验证
validate_security() {
    local plugin_file="$1"
    local plugin_type="$2"
    local -n issues_ref=$3
    local -n warnings_ref=$4
    
    # 检查危险函数
    if command -v strings &> /dev/null; then
        local strings_output=$(strings "$plugin_file")
        
        # 危险函数列表
        local dangerous_funcs=("system" "exec" "popen" "fork" "chmod" "chown")
        
        for func in "${dangerous_funcs[@]}"; do
            if echo "$strings_output" | grep -q "$func"; then
                warnings_ref+=("包含潜在危险函数: $func")
            fi
        done
        
        # 检查硬编码密码或密钥
        if echo "$strings_output" | grep -qE "(password|secret|key|token).*=.*[a-zA-Z0-9]{8,}"; then
            issues_ref+=("可能包含硬编码的敏感信息")
        fi
    fi
    
    # 检查文件权限
    local perms=$(stat -c "%a" "$plugin_file" 2>/dev/null || stat -f "%Lp" "$plugin_file" 2>/dev/null)
    if [[ "$perms" == *"777"* ]]; then
        warnings_ref+=("文件权限过于宽松")
    fi
}

# 性能验证
validate_performance() {
    local plugin_file="$1"
    local plugin_type="$2"
    local -n issues_ref=$3
    local -n warnings_ref=$4
    
    # 文件大小检查
    local file_size=$(stat -c%s "$plugin_file" 2>/dev/null || stat -f%z "$plugin_file" 2>/dev/null)
    
    case $plugin_type in
        shared)
            if [[ $file_size -gt 52428800 ]]; then  # 50MB
                warnings_ref+=("动态库文件过大，可能影响加载性能")
            fi
            ;;
        wasm)
            if [[ $file_size -gt 10485760 ]]; then  # 10MB
                warnings_ref+=("WebAssembly文件过大，可能影响加载性能")
            fi
            ;;
    esac
    
    # 简单的加载时间测试
    case $plugin_type in
        shared)
            local start_time=$(date +%s%N)
            dlopen_test "$plugin_file" &>/dev/null || true
            local end_time=$(date +%s%N)
            local load_time=$(( (end_time - start_time) / 1000000 ))  # 毫秒
            
            if [[ $load_time -gt 1000 ]]; then  # 1秒
                warnings_ref+=("插件加载时间过长: ${load_time}ms")
            fi
            ;;
    esac
}

# dlopen测试函数
dlopen_test() {
    local plugin_file="$1"
    
    # 创建简单的测试程序
    cat > "$TEMP_DIR/dlopen_test.c" << 'EOF'
#include <dlfcn.h>
#include <stdio.h>

int main(int argc, char* argv[]) {
    if (argc != 2) return 1;
    
    void* handle = dlopen(argv[1], RTLD_LAZY);
    if (!handle) {
        fprintf(stderr, "dlopen failed: %s\n", dlerror());
        return 1;
    }
    
    dlclose(handle);
    return 0;
}
EOF
    
    gcc -o "$TEMP_DIR/dlopen_test" "$TEMP_DIR/dlopen_test.c" -ldl 2>/dev/null && \
    "$TEMP_DIR/dlopen_test" "$plugin_file"
}

# 代码质量验证
validate_code_quality() {
    local plugin_file="$1"
    local plugin_type="$2"
    local -n issues_ref=$3
    local -n warnings_ref=$4
    
    # 检查调试信息
    if command -v objdump &> /dev/null && [[ $plugin_type == "shared" ]]; then
        if objdump -h "$plugin_file" 2>/dev/null | grep -q ".debug"; then
            warnings_ref+=("包含调试信息，建议strip优化")
        fi
    fi
    
    # 检查符号表
    if command -v nm &> /dev/null && [[ $plugin_type == "shared" ]]; then
        local symbol_count=$(nm -D "$plugin_file" 2>/dev/null | wc -l)
        if [[ $symbol_count -gt 1000 ]]; then
            warnings_ref+=("导出符号过多，可能影响性能")
        fi
    fi
}

# 尝试自动修复
attempt_fix() {
    local plugin_file="$1"
    local plugin_type="$2"
    shift 2
    local issues=("$@")
    
    log_info "尝试修复插件: $(basename "$plugin_file")..."
    
    for issue in "${issues[@]}"; do
        case $issue in
            *"文件权限"*)
                chmod 755 "$plugin_file"
                log_info "已修复文件权限"
                ;;
            *"调试信息"*)
                if command -v strip &> /dev/null; then
                    strip "$plugin_file"
                    log_info "已移除调试信息"
                fi
                ;;
        esac
    done
}

# 生成验证报告
generate_report() {
    local report_data=$(jq -n \
        --arg timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        --arg level "$VALIDATION_LEVEL" \
        --arg total "$TOTAL_PLUGINS" \
        --arg passed "$PASSED_PLUGINS" \
        --arg failed "$FAILED_PLUGINS" \
        --arg warnings "$WARNING_PLUGINS" \
        --argjson results "$(printf '%s\n' "${VALIDATION_RESULTS[@]}" | jq -s .)" \
        '{
            timestamp: $timestamp,
            validation_level: $level,
            summary: {
                total: ($total | tonumber),
                passed: ($passed | tonumber),
                failed: ($failed | tonumber),
                warnings: ($warnings | tonumber)
            },
            results: $results
        }')
    
    case $OUTPUT_FORMAT in
        json)
            echo "$report_data"
            ;;
        xml)
            echo "$report_data" | jq -r 'to_entries | map("<\(.key)>\(.value)</\(.key)>") | join("\n")' | \
            sed 's/^/<report>/' | sed '$s/$/<\/report>/'
            ;;
        text)
            generate_text_report "$report_data"
            ;;
    esac
}

# 生成文本报告
generate_text_report() {
    local report_data="$1"
    
    echo "go-musicfox v2 插件验证报告"
    echo "=============================="
    echo
    echo "验证时间: $(echo "$report_data" | jq -r '.timestamp')"
    echo "验证级别: $(echo "$report_data" | jq -r '.validation_level')"
    echo
    echo "验证摘要:"
    echo "  总计: $(echo "$report_data" | jq -r '.summary.total')"
    echo "  通过: $(echo "$report_data" | jq -r '.summary.passed')"
    echo "  失败: $(echo "$report_data" | jq -r '.summary.failed')"
    echo "  警告: $(echo "$report_data" | jq -r '.summary.warnings')"
    echo
    
    if [[ $(echo "$report_data" | jq -r '.summary.failed') -gt 0 ]]; then
        echo "失败的插件:"
        echo "$report_data" | jq -r '.results[] | select(.status == "failed") | "  ✗ \(.name): \(.issues | join(", "))"'
        echo
    fi
    
    if [[ $(echo "$report_data" | jq -r '.summary.warnings') -gt 0 ]]; then
        echo "有警告的插件:"
        echo "$report_data" | jq -r '.results[] | select(.status == "warning") | "  ⚠ \(.name): \(.warnings | join(", "))"'
        echo
    fi
    
    echo "通过的插件:"
    echo "$report_data" | jq -r '.results[] | select(.status == "passed") | "  ✓ \(.name)"'
}

# 主函数
main() {
    log_info "go-musicfox v2 插件验证器"
    
    parse_args "$@"
    init_validation
    find_plugins
    
    # 验证所有插件
    for plugin_file in "${PLUGIN_FILES[@]}"; do
        validate_plugin "$plugin_file"
    done
    
    # 生成报告
    local report=$(generate_report)
    
    if [[ -n "$REPORT_FILE" ]]; then
        echo "$report" > "$REPORT_FILE"
        log_info "验证报告已保存: $REPORT_FILE"
    elif [[ $OUTPUT_FORMAT != "text" || $QUIET == true ]]; then
        echo "$report"
    fi
    
    # 显示摘要
    if [[ $QUIET == false ]]; then
        echo
        log_info "验证完成: $PASSED_PLUGINS 通过, $WARNING_PLUGINS 警告, $FAILED_PLUGINS 失败"
    fi
    
    # 设置退出码
    if [[ $FAILED_PLUGINS -gt 0 ]]; then
        exit 1
    elif [[ $WARNING_PLUGINS -gt 0 ]]; then
        exit 2
    else
        exit 0
    fi
}

# 执行主函数
main "$@"