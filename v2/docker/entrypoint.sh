#!/bin/sh
# go-musicfox v2 容器启动脚本
# 支持微内核插件架构的容器化部署

set -e

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

# 错误处理
error_exit() {
    log_error "$1"
    exit 1
}

# 信号处理
trap 'log_info "接收到终止信号，正在优雅关闭..."; kill -TERM $PID; wait $PID' TERM INT

# 环境变量默认值
MUSICFOX_ENV=${MUSICFOX_ENV:-production}
MUSICFOX_LOG_LEVEL=${MUSICFOX_LOG_LEVEL:-info}
MUSICFOX_CONFIG_DIR=${MUSICFOX_CONFIG_DIR:-/app/config}
MUSICFOX_PLUGIN_DIR=${MUSICFOX_PLUGIN_DIR:-/app/plugins}
MUSICFOX_DATA_DIR=${MUSICFOX_DATA_DIR:-/app/data}
MUSICFOX_LOG_DIR=${MUSICFOX_LOG_DIR:-/app/logs}

# 显示启动信息
show_startup_info() {
    log_info "=== go-musicfox v2 容器启动 ==="
    log_info "环境: $MUSICFOX_ENV"
    log_info "日志级别: $MUSICFOX_LOG_LEVEL"
    log_info "配置目录: $MUSICFOX_CONFIG_DIR"
    log_info "插件目录: $MUSICFOX_PLUGIN_DIR"
    log_info "数据目录: $MUSICFOX_DATA_DIR"
    log_info "日志目录: $MUSICFOX_LOG_DIR"
    log_info "用户: $(whoami)"
    log_info "工作目录: $(pwd)"
    echo
}

# 检查必要目录
check_directories() {
    log_info "检查目录结构..."
    
    for dir in "$MUSICFOX_CONFIG_DIR" "$MUSICFOX_PLUGIN_DIR" "$MUSICFOX_DATA_DIR" "$MUSICFOX_LOG_DIR"; do
        if [ ! -d "$dir" ]; then
            log_warning "目录不存在，正在创建: $dir"
            mkdir -p "$dir" || error_exit "无法创建目录: $dir"
        fi
        
        if [ ! -w "$dir" ]; then
            error_exit "目录不可写: $dir"
        fi
    done
    
    log_success "目录检查完成"
}

# 检查二进制文件
check_binary() {
    log_info "检查二进制文件..."
    
    if [ ! -f "/app/bin/go-musicfox" ]; then
        error_exit "二进制文件不存在: /app/bin/go-musicfox"
    fi
    
    if [ ! -x "/app/bin/go-musicfox" ]; then
        error_exit "二进制文件不可执行: /app/bin/go-musicfox"
    fi
    
    log_success "二进制文件检查完成"
}

# 初始化配置
init_config() {
    log_info "初始化配置..."
    
    # 检查主配置文件
    local main_config="$MUSICFOX_CONFIG_DIR/config.yaml"
    if [ ! -f "$main_config" ]; then
        log_warning "主配置文件不存在，使用默认配置"
        if [ -f "/app/config/config.yaml.example" ]; then
            cp "/app/config/config.yaml.example" "$main_config"
            log_info "已复制默认配置文件"
        fi
    fi
    
    # 检查插件配置文件
    local plugin_config="$MUSICFOX_CONFIG_DIR/plugins.yaml"
    if [ ! -f "$plugin_config" ]; then
        log_warning "插件配置文件不存在，使用默认配置"
        if [ -f "/app/config/plugins.yaml.example" ]; then
            cp "/app/config/plugins.yaml.example" "$plugin_config"
            log_info "已复制默认插件配置文件"
        fi
    fi
    
    log_success "配置初始化完成"
}

# 检查插件
check_plugins() {
    log_info "检查插件..."
    
    local plugin_count=0
    
    # 检查各种类型的插件
    for plugin_type in "shared" "rpc" "wasm" "hotload"; do
        local plugin_dir="$MUSICFOX_PLUGIN_DIR/$plugin_type"
        if [ -d "$plugin_dir" ]; then
            local count=$(find "$plugin_dir" -type f | wc -l)
            if [ "$count" -gt 0 ]; then
                log_info "发现 $plugin_type 插件: $count 个"
                plugin_count=$((plugin_count + count))
            fi
        fi
    done
    
    log_info "总计插件数量: $plugin_count"
    log_success "插件检查完成"
}

# 等待依赖服务
wait_for_services() {
    log_info "等待依赖服务..."
    
    # 等待Redis
    if [ -n "${MUSICFOX_REDIS_URL:-}" ]; then
        log_info "等待Redis服务..."
        local redis_host=$(echo "$MUSICFOX_REDIS_URL" | sed 's|redis://||' | cut -d':' -f1)
        local redis_port=$(echo "$MUSICFOX_REDIS_URL" | sed 's|redis://||' | cut -d':' -f2 | cut -d'/' -f1)
        
        local retry_count=0
        local max_retries=30
        
        while [ $retry_count -lt $max_retries ]; do
            if nc -z "$redis_host" "$redis_port" 2>/dev/null; then
                log_success "Redis服务已就绪"
                break
            fi
            
            retry_count=$((retry_count + 1))
            log_info "等待Redis服务... ($retry_count/$max_retries)"
            sleep 2
        done
        
        if [ $retry_count -eq $max_retries ]; then
            log_warning "Redis服务连接超时，继续启动"
        fi
    fi
    
    # 等待数据库
    if [ -n "${MUSICFOX_DB_URL:-}" ]; then
        log_info "等待数据库服务..."
        local db_host=$(echo "$MUSICFOX_DB_URL" | sed 's|.*@||' | cut -d':' -f1)
        local db_port=$(echo "$MUSICFOX_DB_URL" | sed 's|.*@||' | cut -d':' -f2 | cut -d'/' -f1)
        
        local retry_count=0
        local max_retries=30
        
        while [ $retry_count -lt $max_retries ]; do
            if nc -z "$db_host" "$db_port" 2>/dev/null; then
                log_success "数据库服务已就绪"
                break
            fi
            
            retry_count=$((retry_count + 1))
            log_info "等待数据库服务... ($retry_count/$max_retries)"
            sleep 2
        done
        
        if [ $retry_count -eq $max_retries ]; then
            log_warning "数据库服务连接超时，继续启动"
        fi
    fi
}

# 运行数据库迁移
run_migrations() {
    if [ "$MUSICFOX_ENV" = "production" ] && [ -n "${MUSICFOX_DB_URL:-}" ]; then
        log_info "运行数据库迁移..."
        
        if /app/bin/go-musicfox migrate --config="$MUSICFOX_CONFIG_DIR/config.yaml" 2>/dev/null; then
            log_success "数据库迁移完成"
        else
            log_warning "数据库迁移失败或不需要迁移"
        fi
    fi
}

# 健康检查
health_check() {
    if [ "$1" = "--health-check" ]; then
        log_info "执行健康检查..."
        
        # 检查进程是否运行
        if ! pgrep -f "go-musicfox" > /dev/null; then
            log_error "go-musicfox进程未运行"
            exit 1
        fi
        
        # 检查端口是否监听
        if ! nc -z localhost 8080 2>/dev/null; then
            log_error "端口8080未监听"
            exit 1
        fi
        
        log_success "健康检查通过"
        exit 0
    fi
}

# 构建启动参数
build_args() {
    local args=""
    
    # 基础参数
    args="$args --config=$MUSICFOX_CONFIG_DIR/config.yaml"
    args="$args --plugin-dir=$MUSICFOX_PLUGIN_DIR"
    args="$args --data-dir=$MUSICFOX_DATA_DIR"
    args="$args --log-dir=$MUSICFOX_LOG_DIR"
    args="$args --log-level=$MUSICFOX_LOG_LEVEL"
    
    # 环境特定参数
    if [ "$MUSICFOX_ENV" = "development" ]; then
        args="$args --dev"
        if [ "${MUSICFOX_HOT_RELOAD:-}" = "true" ]; then
            args="$args --hot-reload"
        fi
    fi
    
    # 守护进程模式
    if [ "$1" = "--daemon" ] || [ "$MUSICFOX_ENV" = "production" ]; then
        args="$args --daemon"
    fi
    
    # 其他参数
    for arg in "$@"; do
        case $arg in
            --daemon) ;; # 已处理
            --health-check) ;; # 特殊处理
            *) args="$args $arg" ;;
        esac
    done
    
    echo "$args"
}

# 启动应用
start_app() {
    local args=$(build_args "$@")
    
    log_info "启动go-musicfox..."
    log_info "启动参数: $args"
    
    # 启动应用
    exec /app/bin/go-musicfox $args &
    PID=$!
    
    log_success "go-musicfox已启动 (PID: $PID)"
    
    # 等待进程
    wait $PID
}

# 显示版本信息
show_version() {
    if [ "$1" = "--version" ] || [ "$1" = "-v" ]; then
        /app/bin/go-musicfox --version
        exit 0
    fi
}

# 显示帮助信息
show_help() {
    if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
        cat << EOF
go-musicfox v2 容器启动脚本

用法: $0 [选项]

选项:
  --daemon          以守护进程模式运行
  --health-check    执行健康检查
  --version, -v     显示版本信息
  --help, -h        显示帮助信息

环境变量:
  MUSICFOX_ENV           运行环境 (development|production)
  MUSICFOX_LOG_LEVEL     日志级别 (debug|info|warn|error)
  MUSICFOX_CONFIG_DIR    配置目录路径
  MUSICFOX_PLUGIN_DIR    插件目录路径
  MUSICFOX_DATA_DIR      数据目录路径
  MUSICFOX_LOG_DIR       日志目录路径
  MUSICFOX_REDIS_URL     Redis连接URL
  MUSICFOX_DB_URL        数据库连接URL
  MUSICFOX_HOT_RELOAD    启用热重载 (true|false)

示例:
  $0                     # 默认启动
  $0 --daemon            # 守护进程模式
  $0 --health-check      # 健康检查

EOF
        exit 0
    fi
}

# 主函数
main() {
    # 处理特殊参数
    show_help "$@"
    show_version "$@"
    health_check "$@"
    
    # 显示启动信息
    show_startup_info
    
    # 执行启动检查
    check_directories
    check_binary
    init_config
    check_plugins
    wait_for_services
    run_migrations
    
    # 启动应用
    start_app "$@"
}

# 执行主函数
main "$@"