#!/bin/bash
# go-musicfox v2 开发服务器
# 提供插件开发和调试环境

set -euo pipefail

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
PLUGINS_DIR="$PROJECT_ROOT/plugins"
CORE_BINARY="$PROJECT_ROOT/build/go-musicfox-linux-amd64"
DEV_CONFIG="$PROJECT_ROOT/configs/dev.yaml"
DEV_PORT=8080
DEV_RPC_PORT=8081
DEV_PLUGIN_PORT=8082
DEV_DEBUG_PORT=2345

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
go-musicfox v2 开发服务器

用法: $0 [选项] [命令]

命令:
  start                   启动开发服务器 (默认)
  stop                    停止开发服务器
  restart                 重启开发服务器
  status                  查看服务器状态
  logs                    查看服务器日志
  reload                  重新加载插件
  debug                   启动调试模式
  test                    运行插件测试
  build                   构建插件
  clean                   清理开发环境

选项:
  -h, --help              显示帮助信息
  -p, --port PORT         HTTP端口 (默认: 8080)
  -r, --rpc-port PORT     RPC端口 (默认: 8081)
  -g, --plugin-port PORT  插件端口 (默认: 8082)
  -d, --debug-port PORT   调试端口 (默认: 2345)
  -c, --config FILE       配置文件路径
  -w, --watch             监视文件变化并自动重载
  -v, --verbose           详细输出
  -q, --quiet             静默模式
  --hot-reload            启用热重载
  --no-build              跳过构建步骤
  --profile               启用性能分析
  --race                  启用竞态检测

示例:
  $0                      # 启动开发服务器
  $0 -w                   # 启动并监视文件变化
  $0 debug                # 启动调试模式
  $0 test plugins/shared/ # 测试指定插件
  $0 build --race         # 构建并启用竞态检测

EOF
}

# 解析命令行参数
parse_args() {
    COMMAND="start"
    CONFIG_FILE="$DEV_CONFIG"
    WATCH_MODE=false
    VERBOSE=false
    QUIET=false
    HOT_RELOAD=true
    SKIP_BUILD=false
    ENABLE_PROFILE=false
    ENABLE_RACE=false
    PLUGIN_PATHS=()
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help) show_help; exit 0 ;;
            -p|--port) DEV_PORT="$2"; shift 2 ;;
            -r|--rpc-port) DEV_RPC_PORT="$2"; shift 2 ;;
            -g|--plugin-port) DEV_PLUGIN_PORT="$2"; shift 2 ;;
            -d|--debug-port) DEV_DEBUG_PORT="$2"; shift 2 ;;
            -c|--config) CONFIG_FILE="$2"; shift 2 ;;
            -w|--watch) WATCH_MODE=true; shift ;;
            -v|--verbose) VERBOSE=true; shift ;;
            -q|--quiet) QUIET=true; shift ;;
            --hot-reload) HOT_RELOAD=true; shift ;;
            --no-build) SKIP_BUILD=true; shift ;;
            --profile) ENABLE_PROFILE=true; shift ;;
            --race) ENABLE_RACE=true; shift ;;
            start|stop|restart|status|logs|reload|debug|test|build|clean)
                COMMAND="$1"; shift ;;
            -*) error_exit "未知选项: $1" ;;
            *) PLUGIN_PATHS+=("$1"); shift ;;
        esac
    done
    
    # 验证端口
    for port in "$DEV_PORT" "$DEV_RPC_PORT" "$DEV_PLUGIN_PORT" "$DEV_DEBUG_PORT"; do
        if [[ ! "$port" =~ ^[0-9]+$ ]] || [[ $port -lt 1024 || $port -gt 65535 ]]; then
            error_exit "无效的端口号: $port"
        fi
    done
}

# 初始化开发环境
init_dev_environment() {
    log_info "初始化开发环境..."
    
    # 创建必要目录
    mkdir -p "$PROJECT_ROOT"/{build,logs,tmp,data}
    
    # 检查必要工具
    local missing_tools=()
    
    if ! command -v go &> /dev/null; then
        missing_tools+=("go")
    fi
    
    if [[ $WATCH_MODE == true ]] && ! command -v inotifywait &> /dev/null && ! command -v fswatch &> /dev/null; then
        log_warning "未安装文件监视工具，将禁用监视模式"
        WATCH_MODE=false
    fi
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        error_exit "缺少必要工具: ${missing_tools[*]}"
    fi
    
    # 设置环境变量
    export MUSICFOX_ENV=development
    export MUSICFOX_LOG_LEVEL=debug
    export MUSICFOX_HOT_RELOAD=$HOT_RELOAD
    export MUSICFOX_CONFIG_DIR="$PROJECT_ROOT/configs"
    export MUSICFOX_PLUGIN_DIR="$PLUGINS_DIR"
    export MUSICFOX_DATA_DIR="$PROJECT_ROOT/data"
    export MUSICFOX_LOG_DIR="$PROJECT_ROOT/logs"
    
    # 创建PID文件目录
    PID_DIR="$PROJECT_ROOT/tmp"
    PID_FILE="$PID_DIR/dev-server.pid"
    LOG_FILE="$PROJECT_ROOT/logs/dev-server.log"
    
    log_success "开发环境初始化完成"
}

# 生成开发配置
generate_dev_config() {
    if [[ -f "$CONFIG_FILE" ]]; then
        return 0
    fi
    
    log_info "生成开发配置文件..."
    
    mkdir -p "$(dirname "$CONFIG_FILE")"
    
    cat > "$CONFIG_FILE" << EOF
# go-musicfox v2 开发配置

server:
  host: "127.0.0.1"
  port: $DEV_PORT
  rpc_port: $DEV_RPC_PORT
  plugin_port: $DEV_PLUGIN_PORT
  debug_port: $DEV_DEBUG_PORT

log:
  level: "debug"
  format: "text"
  output: "stdout"
  file: "$LOG_FILE"

development:
  enabled: true
  hot_reload: $HOT_RELOAD
  auto_build: true
  watch_paths:
    - "$PLUGINS_DIR"
    - "$PROJECT_ROOT/pkg"
    - "$PROJECT_ROOT/internal"
  build_flags:
    - "-race=$ENABLE_RACE"
    - "-v=$VERBOSE"

plugins:
  enabled: true
  directory: "$PLUGINS_DIR"
  hot_reload: $HOT_RELOAD
  development_mode: true
  types:
    - "shared"
    - "rpc"
    - "wasm"
    - "hotload"
  security:
    verify_signatures: false
    sandbox_mode: false

api:
  enabled: true
  cors_enabled: true
  cors_origins:
    - "http://localhost:*"
    - "http://127.0.0.1:*"
  rate_limit:
    enabled: false

monitoring:
  enabled: true
  metrics_enabled: true
  pprof_enabled: $ENABLE_PROFILE
  health_check_enabled: true

database:
  type: "sqlite"
  connection: "$PROJECT_ROOT/data/dev.db"

redis:
  enabled: false

security:
  development_mode: true
  disable_auth: true
EOF
    
    log_success "开发配置文件已生成: $CONFIG_FILE"
}

# 构建核心和插件
build_project() {
    if [[ $SKIP_BUILD == true ]]; then
        log_info "跳过构建步骤"
        return 0
    fi
    
    log_info "构建项目..."
    
    local build_flags=()
    
    if [[ $ENABLE_RACE == true ]]; then
        build_flags+=("-race")
    fi
    
    if [[ $VERBOSE == true ]]; then
        build_flags+=("-v")
    fi
    
    # 构建核心
    log_info "构建核心二进制..."
    cd "$PROJECT_ROOT"
    
    go build "${build_flags[@]}" -o "$CORE_BINARY" ./cmd/go-musicfox
    
    if [[ ! -f "$CORE_BINARY" ]]; then
        error_exit "核心二进制构建失败"
    fi
    
    # 构建插件
    if [[ -d "$PLUGINS_DIR" ]]; then
        log_info "构建插件..."
        "$SCRIPT_DIR/build-plugins.sh" --dev
    fi
    
    log_success "项目构建完成"
}

# 启动开发服务器
start_server() {
    if is_server_running; then
        log_warning "开发服务器已在运行 (PID: $(cat "$PID_FILE"))"
        return 0
    fi
    
    log_info "启动开发服务器..."
    
    # 构建项目
    build_project
    
    # 生成配置
    generate_dev_config
    
    # 启动服务器
    local server_args=(
        "--config=$CONFIG_FILE"
        "--dev"
        "--port=$DEV_PORT"
        "--rpc-port=$DEV_RPC_PORT"
        "--plugin-port=$DEV_PLUGIN_PORT"
    )
    
    if [[ $ENABLE_PROFILE == true ]]; then
        server_args+=("--pprof")
    fi
    
    if [[ $HOT_RELOAD == true ]]; then
        server_args+=("--hot-reload")
    fi
    
    # 启动服务器进程
    if [[ $QUIET == false ]]; then
        "$CORE_BINARY" "${server_args[@]}" &
    else
        "$CORE_BINARY" "${server_args[@]}" > "$LOG_FILE" 2>&1 &
    fi
    
    local server_pid=$!
    echo $server_pid > "$PID_FILE"
    
    # 等待服务器启动
    sleep 2
    
    if ! kill -0 $server_pid 2>/dev/null; then
        error_exit "服务器启动失败"
    fi
    
    log_success "开发服务器已启动 (PID: $server_pid)"
    log_info "HTTP服务: http://127.0.0.1:$DEV_PORT"
    log_info "RPC服务: 127.0.0.1:$DEV_RPC_PORT"
    log_info "插件服务: 127.0.0.1:$DEV_PLUGIN_PORT"
    
    if [[ $ENABLE_PROFILE == true ]]; then
        log_info "性能分析: http://127.0.0.1:$DEV_PORT/debug/pprof/"
    fi
    
    # 启动文件监视
    if [[ $WATCH_MODE == true ]]; then
        start_file_watcher
    fi
}

# 停止开发服务器
stop_server() {
    if ! is_server_running; then
        log_warning "开发服务器未运行"
        return 0
    fi
    
    local server_pid=$(cat "$PID_FILE")
    log_info "停止开发服务器 (PID: $server_pid)..."
    
    # 优雅关闭
    kill -TERM $server_pid 2>/dev/null || true
    
    # 等待进程结束
    local count=0
    while kill -0 $server_pid 2>/dev/null && [[ $count -lt 10 ]]; do
        sleep 1
        ((count++))
    done
    
    # 强制关闭
    if kill -0 $server_pid 2>/dev/null; then
        log_warning "强制关闭服务器"
        kill -KILL $server_pid 2>/dev/null || true
    fi
    
    rm -f "$PID_FILE"
    
    # 停止文件监视
    stop_file_watcher
    
    log_success "开发服务器已停止"
}

# 重启开发服务器
restart_server() {
    log_info "重启开发服务器..."
    stop_server
    sleep 1
    start_server
}

# 检查服务器状态
check_server_status() {
    if is_server_running; then
        local server_pid=$(cat "$PID_FILE")
        log_success "开发服务器正在运行 (PID: $server_pid)"
        
        # 检查端口
        if command -v netstat &> /dev/null; then
            local listening_ports=$(netstat -tlnp 2>/dev/null | grep ":$DEV_PORT\|:$DEV_RPC_PORT\|:$DEV_PLUGIN_PORT" | wc -l)
            log_info "监听端口数: $listening_ports"
        fi
        
        # 检查内存使用
        if command -v ps &> /dev/null; then
            local memory_usage=$(ps -p $server_pid -o rss= 2>/dev/null | tr -d ' ')
            if [[ -n "$memory_usage" ]]; then
                local memory_mb=$((memory_usage / 1024))
                log_info "内存使用: ${memory_mb}MB"
            fi
        fi
        
        return 0
    else
        log_error "开发服务器未运行"
        return 1
    fi
}

# 查看服务器日志
show_server_logs() {
    if [[ ! -f "$LOG_FILE" ]]; then
        log_warning "日志文件不存在: $LOG_FILE"
        return 0
    fi
    
    log_info "显示服务器日志 (按Ctrl+C退出):"
    echo
    
    if command -v tail &> /dev/null; then
        tail -f "$LOG_FILE"
    else
        cat "$LOG_FILE"
    fi
}

# 重新加载插件
reload_plugins() {
    if ! is_server_running; then
        log_error "开发服务器未运行"
        return 1
    fi
    
    log_info "重新加载插件..."
    
    # 发送重载信号
    local server_pid=$(cat "$PID_FILE")
    kill -HUP $server_pid 2>/dev/null || {
        log_error "发送重载信号失败"
        return 1
    }
    
    log_success "插件重载信号已发送"
}

# 启动调试模式
start_debug_mode() {
    if is_server_running; then
        log_error "请先停止开发服务器"
        return 1
    fi
    
    log_info "启动调试模式..."
    
    # 检查delve
    if ! command -v dlv &> /dev/null; then
        log_info "安装delve调试器..."
        go install github.com/go-delve/delve/cmd/dlv@latest
    fi
    
    # 构建调试版本
    log_info "构建调试版本..."
    cd "$PROJECT_ROOT"
    go build -gcflags="-N -l" -o "$CORE_BINARY-debug" ./cmd/go-musicfox
    
    # 生成配置
    generate_dev_config
    
    # 启动调试器
    log_info "启动调试器 (端口: $DEV_DEBUG_PORT)..."
    log_info "连接方式: dlv connect 127.0.0.1:$DEV_DEBUG_PORT"
    
    dlv --listen=127.0.0.1:$DEV_DEBUG_PORT --headless=true --api-version=2 --accept-multiclient \
        exec "$CORE_BINARY-debug" -- --config="$CONFIG_FILE" --dev
}

# 运行插件测试
run_plugin_tests() {
    log_info "运行插件测试..."
    
    local test_paths=("${PLUGIN_PATHS[@]}")
    
    if [[ ${#test_paths[@]} -eq 0 ]]; then
        test_paths=("$PLUGINS_DIR")
    fi
    
    local test_flags=()
    
    if [[ $ENABLE_RACE == true ]]; then
        test_flags+=("-race")
    fi
    
    if [[ $VERBOSE == true ]]; then
        test_flags+=("-v")
    fi
    
    for test_path in "${test_paths[@]}"; do
        if [[ -d "$test_path" ]]; then
            log_info "测试路径: $test_path"
            
            cd "$test_path"
            
            # 查找测试文件
            if find . -name "*_test.go" | grep -q .; then
                go test "${test_flags[@]}" ./...
            else
                log_warning "未找到测试文件: $test_path"
            fi
        else
            log_warning "测试路径不存在: $test_path"
        fi
    done
    
    log_success "插件测试完成"
}

# 构建插件
build_plugins() {
    log_info "构建插件..."
    
    local build_paths=("${PLUGIN_PATHS[@]}")
    
    if [[ ${#build_paths[@]} -eq 0 ]]; then
        build_paths=("$PLUGINS_DIR")
    fi
    
    local build_flags=()
    
    if [[ $ENABLE_RACE == true ]]; then
        build_flags+=("--race")
    fi
    
    if [[ $VERBOSE == true ]]; then
        build_flags+=("--verbose")
    fi
    
    for build_path in "${build_paths[@]}"; do
        if [[ -d "$build_path" ]]; then
            log_info "构建路径: $build_path"
            "$SCRIPT_DIR/build-plugins.sh" "${build_flags[@]}" "$build_path"
        else
            log_warning "构建路径不存在: $build_path"
        fi
    done
    
    log_success "插件构建完成"
}

# 清理开发环境
clean_dev_environment() {
    log_info "清理开发环境..."
    
    # 停止服务器
    if is_server_running; then
        stop_server
    fi
    
    # 清理构建产物
    rm -rf "$PROJECT_ROOT/build"
    rm -rf "$PROJECT_ROOT/tmp"
    rm -f "$CORE_BINARY-debug"
    
    # 清理日志
    rm -f "$LOG_FILE"
    
    # 清理插件构建产物
    find "$PLUGINS_DIR" -name "*.so" -o -name "*.dll" -o -name "*.dylib" -o -name "*.wasm" | xargs rm -f
    
    log_success "开发环境清理完成"
}

# 启动文件监视
start_file_watcher() {
    if [[ $WATCH_MODE == false ]]; then
        return 0
    fi
    
    log_info "启动文件监视..."
    
    local watch_paths=(
        "$PLUGINS_DIR"
        "$PROJECT_ROOT/pkg"
        "$PROJECT_ROOT/internal"
        "$PROJECT_ROOT/cmd"
    )
    
    # 创建监视脚本
    local watcher_script="$PID_DIR/file-watcher.sh"
    cat > "$watcher_script" << 'EOF'
#!/bin/bash

WATCH_PATHS=("$@")
RELOAD_DELAY=2
LAST_RELOAD=0

handle_change() {
    local current_time=$(date +%s)
    
    if [[ $((current_time - LAST_RELOAD)) -lt $RELOAD_DELAY ]]; then
        return 0
    fi
    
    echo "[$(date)] 检测到文件变化，重新加载..."
    
    # 重新构建
    if ! make build-plugins >/dev/null 2>&1; then
        echo "[$(date)] 构建失败"
        return 1
    fi
    
    # 重新加载插件
    if [[ -f "__PID_FILE__" ]]; then
        local server_pid=$(cat "__PID_FILE__")
        kill -HUP $server_pid 2>/dev/null || true
    fi
    
    LAST_RELOAD=$current_time
    echo "[$(date)] 重新加载完成"
}

if command -v inotifywait >/dev/null 2>&1; then
    # Linux
    inotifywait -m -r -e modify,create,delete,move "${WATCH_PATHS[@]}" --format '%w%f %e' |
    while read file event; do
        if [[ "$file" =~ \.(go|yaml|json)$ ]]; then
            handle_change
        fi
    done
elif command -v fswatch >/dev/null 2>&1; then
    # macOS
    fswatch -o "${WATCH_PATHS[@]}" | while read num; do
        handle_change
    done
else
    echo "未找到文件监视工具"
    exit 1
fi
EOF
    
    # 替换占位符
    sed -i "s|__PID_FILE__|$PID_FILE|g" "$watcher_script"
    chmod +x "$watcher_script"
    
    # 启动监视进程
    "$watcher_script" "${watch_paths[@]}" > "$PROJECT_ROOT/logs/file-watcher.log" 2>&1 &
    local watcher_pid=$!
    echo $watcher_pid > "$PID_DIR/file-watcher.pid"
    
    log_success "文件监视已启动 (PID: $watcher_pid)"
}

# 停止文件监视
stop_file_watcher() {
    local watcher_pid_file="$PID_DIR/file-watcher.pid"
    
    if [[ -f "$watcher_pid_file" ]]; then
        local watcher_pid=$(cat "$watcher_pid_file")
        
        if kill -0 $watcher_pid 2>/dev/null; then
            log_info "停止文件监视 (PID: $watcher_pid)..."
            kill -TERM $watcher_pid 2>/dev/null || true
            
            # 等待进程结束
            local count=0
            while kill -0 $watcher_pid 2>/dev/null && [[ $count -lt 5 ]]; do
                sleep 1
                ((count++))
            done
            
            # 强制关闭
            if kill -0 $watcher_pid 2>/dev/null; then
                kill -KILL $watcher_pid 2>/dev/null || true
            fi
        fi
        
        rm -f "$watcher_pid_file"
    fi
}

# 检查服务器是否运行
is_server_running() {
    if [[ ! -f "$PID_FILE" ]]; then
        return 1
    fi
    
    local server_pid=$(cat "$PID_FILE")
    
    if ! kill -0 $server_pid 2>/dev/null; then
        rm -f "$PID_FILE"
        return 1
    fi
    
    return 0
}

# 信号处理
cleanup() {
    log_info "清理资源..."
    stop_server
    stop_file_watcher
    exit 0
}

trap cleanup SIGINT SIGTERM

# 主函数
main() {
    log_info "go-musicfox v2 开发服务器"
    
    parse_args "$@"
    init_dev_environment
    
    case $COMMAND in
        start)
            start_server
            if [[ $WATCH_MODE == true ]]; then
                log_info "开发服务器运行中，按Ctrl+C停止"
                # 保持脚本运行
                while is_server_running; do
                    sleep 5
                done
            fi
            ;;
        stop)
            stop_server
            ;;
        restart)
            restart_server
            ;;
        status)
            check_server_status
            ;;
        logs)
            show_server_logs
            ;;
        reload)
            reload_plugins
            ;;
        debug)
            start_debug_mode
            ;;
        test)
            run_plugin_tests
            ;;
        build)
            build_plugins
            ;;
        clean)
            clean_dev_environment
            ;;
        *)
            error_exit "未知命令: $COMMAND"
            ;;
    esac
}

# 执行主函数
main "$@"