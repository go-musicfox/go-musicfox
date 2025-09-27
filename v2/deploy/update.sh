#!/bin/bash
# go-musicfox v2 更新脚本
# 智能更新系统，支持版本检查、备份回滚和平滑升级

set -euo pipefail

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_PREFIX="${INSTALL_PREFIX:-/usr/local}"
CONFIG_DIR="${CONFIG_DIR:-}"
DATA_DIR="${DATA_DIR:-}"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/go-musicfox}"
SERVICE_USER="${SERVICE_USER:-musicfox}"
UPDATE_CHANNEL="${UPDATE_CHANNEL:-stable}"

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

# 检测操作系统
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
        if command -v systemctl &> /dev/null; then
            INIT_SYSTEM="systemd"
        elif command -v service &> /dev/null; then
            INIT_SYSTEM="sysv"
        else
            INIT_SYSTEM="unknown"
        fi
        
        if command -v apt-get &> /dev/null; then
            DISTRO="debian"
        elif command -v yum &> /dev/null; then
            DISTRO="rhel"
        elif command -v pacman &> /dev/null; then
            DISTRO="arch"
        else
            DISTRO="unknown"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="darwin"
        DISTRO="macos"
        INIT_SYSTEM="launchd"
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        OS="windows"
        DISTRO="windows"
        INIT_SYSTEM="windows"
    else
        error_exit "不支持的操作系统: $OSTYPE"
    fi
    
    # 检测架构
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        armv7l) ARCH="arm" ;;
        *) error_exit "不支持的架构: $ARCH" ;;
    esac
    
    log_info "检测到系统: $OS/$DISTRO ($ARCH)"
}

# 显示帮助信息
show_help() {
    cat << EOF
go-musicfox v2 更新脚本

用法: $0 [选项] [版本]

选项:
  -h, --help              显示帮助信息
  -c, --check             仅检查更新，不执行升级
  -f, --force             强制更新，跳过版本检查
  -b, --backup            创建备份
  -r, --rollback          回滚到上一版本
  --channel CHANNEL       更新渠道 (stable|beta|dev)
  --backup-dir DIR        备份目录
  --no-backup             跳过备份
  --no-restart            更新后不重启服务
  --dry-run               预览更新操作

版本:
  latest                  最新稳定版
  beta                    最新测试版
  dev                     最新开发版
  vX.Y.Z                  指定版本号

更新渠道:
  stable                  稳定版本 (默认)
  beta                    测试版本
  dev                     开发版本

示例:
  $0                      # 检查并更新到最新稳定版
  $0 -c                   # 仅检查更新
  $0 v2.1.0               # 更新到指定版本
  $0 --channel beta       # 更新到最新测试版
  $0 --rollback           # 回滚到上一版本

EOF
}

# 解析命令行参数
parse_args() {
    CHECK_ONLY=false
    FORCE_UPDATE=false
    CREATE_BACKUP=true
    ROLLBACK=false
    NO_RESTART=false
    DRY_RUN=false
    TARGET_VERSION=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help) show_help; exit 0 ;;
            -c|--check) CHECK_ONLY=true; shift ;;
            -f|--force) FORCE_UPDATE=true; shift ;;
            -b|--backup) CREATE_BACKUP=true; shift ;;
            -r|--rollback) ROLLBACK=true; shift ;;
            --channel) UPDATE_CHANNEL="$2"; shift 2 ;;
            --backup-dir) BACKUP_DIR="$2"; shift 2 ;;
            --no-backup) CREATE_BACKUP=false; shift ;;
            --no-restart) NO_RESTART=true; shift ;;
            --dry-run) DRY_RUN=true; shift ;;
            -*) error_exit "未知选项: $1" ;;
            *) TARGET_VERSION="$1"; shift ;;
        esac
    done
    
    # 设置默认目录
    if [[ -z "$CONFIG_DIR" ]]; then
        if [[ $OS == "darwin" ]]; then
            CONFIG_DIR="/usr/local/etc/go-musicfox"
        else
            CONFIG_DIR="/etc/go-musicfox"
        fi
    fi
    
    if [[ -z "$DATA_DIR" ]]; then
        if [[ $OS == "darwin" ]]; then
            DATA_DIR="/usr/local/var/lib/go-musicfox"
        else
            DATA_DIR="/var/lib/go-musicfox"
        fi
    fi
    
    # 根据渠道设置默认版本
    if [[ -z "$TARGET_VERSION" ]]; then
        case $UPDATE_CHANNEL in
            stable) TARGET_VERSION="latest" ;;
            beta) TARGET_VERSION="beta" ;;
            dev) TARGET_VERSION="dev" ;;
            *) error_exit "无效的更新渠道: $UPDATE_CHANNEL" ;;
        esac
    fi
}

# 检查权限
check_permissions() {
    if [[ $EUID -ne 0 ]] && [[ "$INSTALL_PREFIX" == "/usr/local" || "$INSTALL_PREFIX" == "/usr" ]]; then
        if [[ $FORCE_UPDATE == false ]]; then
            error_exit "需要root权限更新系统安装，请使用sudo运行"
        fi
    fi
}

# 获取当前版本
get_current_version() {
    local binary_path="$INSTALL_PREFIX/bin/go-musicfox"
    
    if [[ ! -f "$binary_path" ]]; then
        log_warning "未找到已安装的go-musicfox"
        CURRENT_VERSION="none"
        return 0
    fi
    
    if [[ ! -x "$binary_path" ]]; then
        log_error "二进制文件不可执行: $binary_path"
        CURRENT_VERSION="unknown"
        return 0
    fi
    
    local version_output
    if version_output=$("$binary_path" --version 2>/dev/null); then
        CURRENT_VERSION=$(echo "$version_output" | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?' | head -1)
        if [[ -z "$CURRENT_VERSION" ]]; then
            CURRENT_VERSION="unknown"
        fi
    else
        CURRENT_VERSION="unknown"
    fi
    
    log_info "当前版本: $CURRENT_VERSION"
}

# 获取最新版本信息
get_latest_version() {
    log_info "检查最新版本..."
    
    local api_url="https://api.github.com/repos/go-musicfox/go-musicfox/releases"
    
    case $TARGET_VERSION in
        latest)
            # 获取最新稳定版
            LATEST_VERSION=$(curl -s "$api_url/latest" | grep '"tag_name"' | cut -d'"' -f4)
            ;;
        beta)
            # 获取最新预发布版
            LATEST_VERSION=$(curl -s "$api_url" | grep '"tag_name"' | head -1 | cut -d'"' -f4)
            ;;
        dev)
            # 获取最新开发版（通常是最新的预发布版）
            LATEST_VERSION=$(curl -s "$api_url" | grep '"tag_name"' | grep -E '(alpha|beta|rc|dev)' | head -1 | cut -d'"' -f4)
            if [[ -z "$LATEST_VERSION" ]]; then
                LATEST_VERSION=$(curl -s "$api_url" | grep '"tag_name"' | head -1 | cut -d'"' -f4)
            fi
            ;;
        v*)
            # 指定版本
            LATEST_VERSION="$TARGET_VERSION"
            ;;
        *)
            error_exit "无效的目标版本: $TARGET_VERSION"
            ;;
    esac
    
    if [[ -z "$LATEST_VERSION" ]]; then
        error_exit "无法获取版本信息，请检查网络连接"
    fi
    
    log_info "目标版本: $LATEST_VERSION"
}

# 比较版本
compare_versions() {
    if [[ "$CURRENT_VERSION" == "$LATEST_VERSION" ]]; then
        VERSION_COMPARISON="equal"
    elif [[ "$CURRENT_VERSION" == "none" || "$CURRENT_VERSION" == "unknown" ]]; then
        VERSION_COMPARISON="upgrade"
    else
        # 简单的版本比较（可以改进为更复杂的语义版本比较）
        if [[ "$CURRENT_VERSION" < "$LATEST_VERSION" ]]; then
            VERSION_COMPARISON="upgrade"
        else
            VERSION_COMPARISON="downgrade"
        fi
    fi
    
    log_info "版本比较结果: $VERSION_COMPARISON"
}

# 检查更新
check_update() {
    get_current_version
    get_latest_version
    compare_versions
    
    case $VERSION_COMPARISON in
        equal)
            log_success "已是最新版本: $CURRENT_VERSION"
            if [[ $FORCE_UPDATE == false ]]; then
                exit 0
            fi
            ;;
        upgrade)
            log_info "发现新版本: $CURRENT_VERSION -> $LATEST_VERSION"
            ;;
        downgrade)
            log_warning "目标版本较旧: $CURRENT_VERSION -> $LATEST_VERSION"
            if [[ $FORCE_UPDATE == false ]]; then
                echo -n "确认降级? [y/N] "
                read -r response
                case $response in
                    [yY]|[yY][eE][sS]) ;;
                    *) log_info "取消更新"; exit 0 ;;
                esac
            fi
            ;;
    esac
    
    if [[ $CHECK_ONLY == true ]]; then
        exit 0
    fi
}

# 创建备份
create_backup() {
    if [[ $CREATE_BACKUP == false ]]; then
        log_info "跳过备份"
        return 0
    fi
    
    log_info "创建备份..."
    
    local backup_timestamp=$(date +"%Y%m%d_%H%M%S")
    local backup_path="$BACKUP_DIR/$backup_timestamp"
    
    if [[ $DRY_RUN == true ]]; then
        log_info "[DRY-RUN] 将创建备份到: $backup_path"
        return 0
    fi
    
    mkdir -p "$backup_path"
    
    # 备份二进制文件
    if [[ -f "$INSTALL_PREFIX/bin/go-musicfox" ]]; then
        cp "$INSTALL_PREFIX/bin/go-musicfox" "$backup_path/"
        log_info "已备份二进制文件"
    fi
    
    # 备份配置文件
    if [[ -d "$CONFIG_DIR" ]]; then
        cp -r "$CONFIG_DIR" "$backup_path/config"
        log_info "已备份配置文件"
    fi
    
    # 备份插件
    if [[ -d "$DATA_DIR/plugins" ]]; then
        mkdir -p "$backup_path/plugins"
        cp -r "$DATA_DIR/plugins"/* "$backup_path/plugins/" 2>/dev/null || true
        log_info "已备份插件"
    fi
    
    # 创建备份信息文件
    cat > "$backup_path/backup_info.txt" << EOF
备份时间: $(date)
原版本: $CURRENT_VERSION
目标版本: $LATEST_VERSION
系统信息: $OS/$DISTRO ($ARCH)
安装前缀: $INSTALL_PREFIX
配置目录: $CONFIG_DIR
数据目录: $DATA_DIR
EOF
    
    # 创建最新备份链接
    ln -sfn "$backup_path" "$BACKUP_DIR/latest"
    
    BACKUP_PATH="$backup_path"
    log_success "备份完成: $backup_path"
}

# 停止服务
stop_service() {
    log_info "停止服务..."
    
    if [[ $DRY_RUN == true ]]; then
        log_info "[DRY-RUN] 将停止服务"
        return 0
    fi
    
    case $INIT_SYSTEM in
        systemd)
            if systemctl is-active --quiet go-musicfox 2>/dev/null; then
                systemctl stop go-musicfox
                log_success "systemd服务已停止"
                SERVICE_WAS_RUNNING=true
            else
                log_info "systemd服务未运行"
                SERVICE_WAS_RUNNING=false
            fi
            ;;
        launchd)
            if launchctl list | grep -q com.go-musicfox 2>/dev/null; then
                launchctl stop com.go-musicfox 2>/dev/null || true
                log_success "launchd服务已停止"
                SERVICE_WAS_RUNNING=true
            else
                log_info "launchd服务未运行"
                SERVICE_WAS_RUNNING=false
            fi
            ;;
        *)
            log_warning "未知的服务系统，请手动停止服务"
            SERVICE_WAS_RUNNING=false
            ;;
    esac
}

# 下载新版本
download_new_version() {
    local binary_name="go-musicfox-$OS-$ARCH"
    if [[ $OS == "windows" ]]; then
        binary_name="${binary_name}.exe"
    fi
    
    local download_url="https://github.com/go-musicfox/go-musicfox/releases/download/$LATEST_VERSION/go-musicfox-$LATEST_VERSION-$OS.tar.gz"
    
    log_info "下载新版本: $LATEST_VERSION"
    log_info "下载地址: $download_url"
    
    if [[ $DRY_RUN == true ]]; then
        log_info "[DRY-RUN] 将下载: $download_url"
        return 0
    fi
    
    local temp_dir=$(mktemp -d)
    local archive_file="$temp_dir/go-musicfox.tar.gz"
    
    if ! curl -L -o "$archive_file" "$download_url"; then
        error_exit "下载失败: $download_url"
    fi
    
    log_info "解压新版本..."
    tar -xzf "$archive_file" -C "$temp_dir"
    
    # 查找解压后的目录
    local extract_dir=$(find "$temp_dir" -maxdepth 1 -type d -name "go-musicfox-*" | head -1)
    if [[ -z "$extract_dir" ]]; then
        error_exit "未找到解压后的目录"
    fi
    
    NEW_BINARY_DIR="$extract_dir"
    log_success "下载完成: $NEW_BINARY_DIR"
}

# 安装新版本
install_new_version() {
    log_info "安装新版本..."
    
    if [[ $DRY_RUN == true ]]; then
        log_info "[DRY-RUN] 将安装新版本"
        return 0
    fi
    
    # 安装二进制文件
    local binary_name="go-musicfox"
    if [[ $OS == "windows" ]]; then
        binary_name="${binary_name}.exe"
    fi
    
    cp "$NEW_BINARY_DIR/bin/$binary_name" "$INSTALL_PREFIX/bin/"
    chmod +x "$INSTALL_PREFIX/bin/$binary_name"
    log_success "二进制文件已更新"
    
    # 更新插件（如果有新插件）
    if [[ -d "$NEW_BINARY_DIR/plugins" ]]; then
        mkdir -p "$DATA_DIR/plugins"
        cp -r "$NEW_BINARY_DIR/plugins"/* "$DATA_DIR/plugins/" || true
        log_success "插件已更新"
    fi
    
    # 更新配置文件（仅新增，不覆盖现有配置）
    if [[ -d "$NEW_BINARY_DIR/config" ]]; then
        for config_file in "$NEW_BINARY_DIR/config"/*; do
            local filename=$(basename "$config_file")
            local target_file="$CONFIG_DIR/$filename"
            
            if [[ ! -f "$target_file" ]]; then
                cp "$config_file" "$target_file"
                log_info "新增配置文件: $filename"
            fi
        done
    fi
    
    # 更新文档
    if [[ -d "$NEW_BINARY_DIR/docs" ]]; then
        mkdir -p "$INSTALL_PREFIX/share/doc/go-musicfox"
        cp -r "$NEW_BINARY_DIR/docs"/* "$INSTALL_PREFIX/share/doc/go-musicfox/" || true
        log_success "文档已更新"
    fi
}

# 启动服务
start_service() {
    if [[ $NO_RESTART == true ]]; then
        log_info "跳过服务重启"
        return 0
    fi
    
    if [[ $SERVICE_WAS_RUNNING == false ]]; then
        log_info "服务之前未运行，跳过启动"
        return 0
    fi
    
    log_info "启动服务..."
    
    if [[ $DRY_RUN == true ]]; then
        log_info "[DRY-RUN] 将启动服务"
        return 0
    fi
    
    case $INIT_SYSTEM in
        systemd)
            systemctl start go-musicfox
            if systemctl is-active --quiet go-musicfox; then
                log_success "systemd服务已启动"
            else
                log_error "systemd服务启动失败"
                return 1
            fi
            ;;
        launchd)
            launchctl start com.go-musicfox
            sleep 2
            if launchctl list | grep -q com.go-musicfox; then
                log_success "launchd服务已启动"
            else
                log_error "launchd服务启动失败"
                return 1
            fi
            ;;
        *)
            log_warning "未知的服务系统，请手动启动服务"
            ;;
    esac
}

# 验证更新
verify_update() {
    log_info "验证更新..."
    
    if [[ $DRY_RUN == true ]]; then
        log_info "[DRY-RUN] 跳过验证"
        return 0
    fi
    
    local binary_path="$INSTALL_PREFIX/bin/go-musicfox"
    
    if [[ ! -f "$binary_path" ]]; then
        log_error "二进制文件不存在: $binary_path"
        return 1
    fi
    
    if [[ ! -x "$binary_path" ]]; then
        log_error "二进制文件不可执行: $binary_path"
        return 1
    fi
    
    # 检查版本
    local new_version
    if new_version=$("$binary_path" --version 2>/dev/null); then
        local installed_version=$(echo "$new_version" | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?' | head -1)
        if [[ "$installed_version" == "$LATEST_VERSION" ]]; then
            log_success "更新验证成功: $installed_version"
        else
            log_warning "版本不匹配: 期望 $LATEST_VERSION，实际 $installed_version"
        fi
    else
        log_error "版本检查失败"
        return 1
    fi
    
    # 检查服务状态
    if [[ $SERVICE_WAS_RUNNING == true && $NO_RESTART == false ]]; then
        case $INIT_SYSTEM in
            systemd)
                if systemctl is-active --quiet go-musicfox; then
                    log_success "服务运行正常"
                else
                    log_warning "服务未运行"
                fi
                ;;
            launchd)
                if launchctl list | grep -q com.go-musicfox; then
                    log_success "服务运行正常"
                else
                    log_warning "服务未运行"
                fi
                ;;
        esac
    fi
}

# 回滚到上一版本
rollback() {
    log_info "开始回滚..."
    
    local latest_backup="$BACKUP_DIR/latest"
    
    if [[ ! -L "$latest_backup" ]]; then
        error_exit "未找到备份，无法回滚"
    fi
    
    local backup_path=$(readlink "$latest_backup")
    
    if [[ ! -d "$backup_path" ]]; then
        error_exit "备份目录不存在: $backup_path"
    fi
    
    log_info "从备份恢复: $backup_path"
    
    if [[ $DRY_RUN == true ]]; then
        log_info "[DRY-RUN] 将从备份恢复"
        return 0
    fi
    
    # 停止服务
    stop_service
    
    # 恢复二进制文件
    if [[ -f "$backup_path/go-musicfox" ]]; then
        cp "$backup_path/go-musicfox" "$INSTALL_PREFIX/bin/"
        chmod +x "$INSTALL_PREFIX/bin/go-musicfox"
        log_success "二进制文件已恢复"
    fi
    
    # 恢复配置文件
    if [[ -d "$backup_path/config" ]]; then
        rm -rf "$CONFIG_DIR"
        cp -r "$backup_path/config" "$CONFIG_DIR"
        log_success "配置文件已恢复"
    fi
    
    # 恢复插件
    if [[ -d "$backup_path/plugins" ]]; then
        rm -rf "$DATA_DIR/plugins"
        mkdir -p "$DATA_DIR/plugins"
        cp -r "$backup_path/plugins"/* "$DATA_DIR/plugins/" 2>/dev/null || true
        log_success "插件已恢复"
    fi
    
    # 启动服务
    start_service
    
    log_success "回滚完成"
}

# 清理临时文件
cleanup() {
    if [[ -n "${NEW_BINARY_DIR:-}" ]]; then
        rm -rf "$(dirname "$NEW_BINARY_DIR")"
    fi
}

# 显示更新摘要
show_update_summary() {
    if [[ $DRY_RUN == true ]]; then
        log_info "预览模式完成，未实际执行更新"
        return 0
    fi
    
    if [[ $ROLLBACK == true ]]; then
        log_success "回滚完成！"
        return 0
    fi
    
    log_success "go-musicfox v2 更新完成！"
    echo
    
    log_info "更新摘要:"
    log_info "  原版本: $CURRENT_VERSION"
    log_info "  新版本: $LATEST_VERSION"
    log_info "  更新渠道: $UPDATE_CHANNEL"
    
    if [[ $CREATE_BACKUP == true ]]; then
        log_info "  备份位置: $BACKUP_PATH"
    fi
    
    echo
    log_info "验证安装:"
    log_info "  运行: go-musicfox --version"
    
    if [[ $SERVICE_WAS_RUNNING == true ]]; then
        echo
        log_info "服务状态:"
        case $INIT_SYSTEM in
            systemd)
                log_info "  检查状态: sudo systemctl status go-musicfox"
                log_info "  查看日志: sudo journalctl -u go-musicfox -f"
                ;;
            launchd)
                log_info "  检查状态: sudo launchctl list | grep go-musicfox"
                ;;
        esac
    fi
    
    echo
    log_info "如果遇到问题，可以使用以下命令回滚:"
    log_info "  $0 --rollback"
}

# 主函数
main() {
    log_info "go-musicfox v2 更新程序"
    
    detect_os
    parse_args "$@"
    
    if [[ $ROLLBACK == true ]]; then
        rollback
        show_update_summary
        exit 0
    fi
    
    check_permissions
    check_update
    
    # 设置清理函数
    trap cleanup EXIT
    
    create_backup
    stop_service
    download_new_version
    install_new_version
    start_service
    verify_update
    
    show_update_summary
}

# 执行主函数
main "$@"