#!/bin/bash
# go-musicfox v2 卸载脚本
# 完全卸载go-musicfox及其相关文件和服务

set -euo pipefail

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_PREFIX="${INSTALL_PREFIX:-/usr/local}"
CONFIG_DIR="${CONFIG_DIR:-}"
DATA_DIR="${DATA_DIR:-}"
SERVICE_USER="${SERVICE_USER:-musicfox}"

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
    
    log_info "检测到系统: $OS/$DISTRO ($INIT_SYSTEM)"
}

# 显示帮助信息
show_help() {
    cat << EOF
go-musicfox v2 卸载脚本

用法: $0 [选项]

选项:
  -h, --help              显示帮助信息
  -p, --prefix PREFIX     安装前缀 (默认: /usr/local)
  -c, --config-dir DIR    配置目录
  -d, --data-dir DIR      数据目录
  -u, --user USER         服务用户 (默认: musicfox)
  -f, --force             强制卸载，不询问确认
  --keep-config           保留配置文件
  --keep-data             保留数据文件
  --keep-user             保留服务用户
  --dry-run               预览卸载操作，不实际删除

环境变量:
  INSTALL_PREFIX          安装前缀
  CONFIG_DIR              配置目录
  DATA_DIR                数据目录
  SERVICE_USER            服务用户

示例:
  $0                      # 交互式卸载
  $0 -f                   # 强制卸载
  $0 --keep-config        # 卸载但保留配置
  $0 --dry-run            # 预览卸载操作

EOF
}

# 解析命令行参数
parse_args() {
    FORCE_UNINSTALL=false
    KEEP_CONFIG=false
    KEEP_DATA=false
    KEEP_USER=false
    DRY_RUN=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help) show_help; exit 0 ;;
            -p|--prefix) INSTALL_PREFIX="$2"; shift 2 ;;
            -c|--config-dir) CONFIG_DIR="$2"; shift 2 ;;
            -d|--data-dir) DATA_DIR="$2"; shift 2 ;;
            -u|--user) SERVICE_USER="$2"; shift 2 ;;
            -f|--force) FORCE_UNINSTALL=true; shift ;;
            --keep-config) KEEP_CONFIG=true; shift ;;
            --keep-data) KEEP_DATA=true; shift ;;
            --keep-user) KEEP_USER=true; shift ;;
            --dry-run) DRY_RUN=true; shift ;;
            -*) error_exit "未知选项: $1" ;;
            *) error_exit "无效参数: $1" ;;
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
}

# 检查权限
check_permissions() {
    if [[ $EUID -ne 0 ]] && [[ "$INSTALL_PREFIX" == "/usr/local" || "$INSTALL_PREFIX" == "/usr" ]]; then
        if [[ $FORCE_UNINSTALL == false ]]; then
            error_exit "需要root权限卸载系统安装，请使用sudo运行"
        fi
    fi
}

# 确认卸载
confirm_uninstall() {
    if [[ $FORCE_UNINSTALL == true || $DRY_RUN == true ]]; then
        return 0
    fi
    
    echo
    log_warning "即将卸载 go-musicfox v2"
    log_info "安装前缀: $INSTALL_PREFIX"
    log_info "配置目录: $CONFIG_DIR"
    log_info "数据目录: $DATA_DIR"
    
    if [[ $KEEP_CONFIG == false ]]; then
        log_warning "配置文件将被删除"
    fi
    
    if [[ $KEEP_DATA == false ]]; then
        log_warning "数据文件将被删除"
    fi
    
    echo
    echo -n "确认卸载? [y/N] "
    read -r response
    case $response in
        [yY]|[yY][eE][sS]) ;;
        *) log_info "取消卸载"; exit 0 ;;
    esac
}

# 安全删除函数
safe_remove() {
    local path="$1"
    local description="$2"
    
    if [[ ! -e "$path" ]]; then
        log_info "跳过不存在的路径: $path"
        return 0
    fi
    
    if [[ $DRY_RUN == true ]]; then
        log_info "[DRY-RUN] 将删除 $description: $path"
        return 0
    fi
    
    log_info "删除 $description: $path"
    
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

# 停止服务
stop_service() {
    log_info "停止 go-musicfox 服务..."
    
    case $INIT_SYSTEM in
        systemd)
            if systemctl is-active --quiet go-musicfox 2>/dev/null; then
                if [[ $DRY_RUN == false ]]; then
                    systemctl stop go-musicfox
                    log_success "systemd服务已停止"
                else
                    log_info "[DRY-RUN] 将停止systemd服务"
                fi
            else
                log_info "systemd服务未运行"
            fi
            ;;
        launchd)
            if launchctl list | grep -q com.go-musicfox 2>/dev/null; then
                if [[ $DRY_RUN == false ]]; then
                    launchctl stop com.go-musicfox 2>/dev/null || true
                    log_success "launchd服务已停止"
                else
                    log_info "[DRY-RUN] 将停止launchd服务"
                fi
            else
                log_info "launchd服务未运行"
            fi
            ;;
        sysv)
            if service go-musicfox status &>/dev/null; then
                if [[ $DRY_RUN == false ]]; then
                    service go-musicfox stop
                    log_success "SysV服务已停止"
                else
                    log_info "[DRY-RUN] 将停止SysV服务"
                fi
            else
                log_info "SysV服务未运行"
            fi
            ;;
        windows)
            # Windows服务停止逻辑
            log_info "请手动停止Windows服务"
            ;;
        *)
            log_warning "未知的初始化系统，请手动停止服务"
            ;;
    esac
}

# 卸载服务
uninstall_service() {
    log_info "卸载系统服务..."
    
    case $INIT_SYSTEM in
        systemd)
            # 禁用服务
            if systemctl is-enabled --quiet go-musicfox 2>/dev/null; then
                if [[ $DRY_RUN == false ]]; then
                    systemctl disable go-musicfox
                    log_success "systemd服务已禁用"
                else
                    log_info "[DRY-RUN] 将禁用systemd服务"
                fi
            fi
            
            # 删除服务文件
            safe_remove "/etc/systemd/system/go-musicfox.service" "systemd服务文件"
            
            # 重新加载systemd
            if [[ $DRY_RUN == false ]]; then
                systemctl daemon-reload
                systemctl reset-failed go-musicfox 2>/dev/null || true
            else
                log_info "[DRY-RUN] 将重新加载systemd"
            fi
            ;;
        launchd)
            # 卸载launchd服务
            if [[ $DRY_RUN == false ]]; then
                launchctl unload /Library/LaunchDaemons/com.go-musicfox.plist 2>/dev/null || true
            else
                log_info "[DRY-RUN] 将卸载launchd服务"
            fi
            
            safe_remove "/Library/LaunchDaemons/com.go-musicfox.plist" "launchd配置文件"
            ;;
        sysv)
            # 删除SysV服务
            safe_remove "/etc/init.d/go-musicfox" "SysV服务脚本"
            
            # 删除服务链接
            for runlevel in 0 1 2 3 4 5 6; do
                safe_remove "/etc/rc${runlevel}.d/"*go-musicfox "SysV服务链接"
            done
            ;;
        windows)
            log_info "请手动卸载Windows服务"
            ;;
        *)
            log_warning "未知的初始化系统，跳过服务卸载"
            ;;
    esac
}

# 删除二进制文件
remove_binaries() {
    log_info "删除二进制文件..."
    
    # 主二进制文件
    safe_remove "$INSTALL_PREFIX/bin/go-musicfox" "主二进制文件"
    
    # 符号链接
    safe_remove "/usr/local/bin/go-musicfox" "符号链接"
    safe_remove "/usr/bin/go-musicfox" "符号链接"
    
    # Windows可执行文件
    if [[ $OS == "windows" ]]; then
        safe_remove "$INSTALL_PREFIX/bin/go-musicfox.exe" "Windows可执行文件"
    fi
}

# 删除配置文件
remove_config() {
    if [[ $KEEP_CONFIG == true ]]; then
        log_info "保留配置文件"
        return 0
    fi
    
    log_info "删除配置文件..."
    
    safe_remove "$CONFIG_DIR" "配置目录"
    
    # 用户配置目录
    local user_config_dirs=(
        "$HOME/.config/go-musicfox"
        "$HOME/.go-musicfox"
    )
    
    if [[ $OS == "darwin" ]]; then
        user_config_dirs+=("$HOME/Library/Application Support/go-musicfox")
    elif [[ $OS == "windows" ]]; then
        user_config_dirs+=("$HOME/AppData/Roaming/go-musicfox")
    fi
    
    for config_dir in "${user_config_dirs[@]}"; do
        if [[ -d "$config_dir" ]]; then
            if [[ $FORCE_UNINSTALL == false && $DRY_RUN == false ]]; then
                echo -n "删除用户配置目录 $config_dir? [y/N] "
                read -r response
                case $response in
                    [yY]|[yY][eE][sS]) safe_remove "$config_dir" "用户配置目录" ;;
                    *) log_info "保留用户配置目录: $config_dir" ;;
                esac
            else
                safe_remove "$config_dir" "用户配置目录"
            fi
        fi
    done
}

# 删除数据文件
remove_data() {
    if [[ $KEEP_DATA == true ]]; then
        log_info "保留数据文件"
        return 0
    fi
    
    log_info "删除数据文件..."
    
    safe_remove "$DATA_DIR" "数据目录"
    
    # 用户数据目录
    local user_data_dirs=(
        "$HOME/.local/share/go-musicfox"
    )
    
    if [[ $OS == "darwin" ]]; then
        user_data_dirs+=("$HOME/Library/Application Support/go-musicfox")
    elif [[ $OS == "windows" ]]; then
        user_data_dirs+=("$HOME/AppData/Local/go-musicfox")
    fi
    
    for data_dir in "${user_data_dirs[@]}"; do
        if [[ -d "$data_dir" ]]; then
            if [[ $FORCE_UNINSTALL == false && $DRY_RUN == false ]]; then
                echo -n "删除用户数据目录 $data_dir? [y/N] "
                read -r response
                case $response in
                    [yY]|[yY][eE][sS]) safe_remove "$data_dir" "用户数据目录" ;;
                    *) log_info "保留用户数据目录: $data_dir" ;;
                esac
            else
                safe_remove "$data_dir" "用户数据目录"
            fi
        fi
    done
}

# 删除文档
remove_documentation() {
    log_info "删除文档..."
    
    safe_remove "$INSTALL_PREFIX/share/doc/go-musicfox" "文档目录"
    safe_remove "$INSTALL_PREFIX/share/man/man1/go-musicfox.1" "手册页"
}

# 删除日志文件
remove_logs() {
    log_info "删除日志文件..."
    
    # 系统日志目录
    local log_dirs=(
        "/var/log/go-musicfox"
        "$DATA_DIR/logs"
    )
    
    for log_dir in "${log_dirs[@]}"; do
        safe_remove "$log_dir" "日志目录"
    done
    
    # 用户日志目录
    if [[ -d "$HOME/.cache/go-musicfox" ]]; then
        safe_remove "$HOME/.cache/go-musicfox" "用户缓存目录"
    fi
}

# 删除服务用户
remove_user() {
    if [[ $KEEP_USER == true || $OS == "windows" ]]; then
        log_info "保留服务用户"
        return 0
    fi
    
    if ! id "$SERVICE_USER" &>/dev/null; then
        log_info "服务用户不存在: $SERVICE_USER"
        return 0
    fi
    
    if [[ "$SERVICE_USER" == "root" ]]; then
        log_info "跳过删除root用户"
        return 0
    fi
    
    if [[ $FORCE_UNINSTALL == false && $DRY_RUN == false ]]; then
        echo -n "删除服务用户 $SERVICE_USER? [y/N] "
        read -r response
        case $response in
            [yY]|[yY][eE][sS]) ;;
            *) log_info "保留服务用户: $SERVICE_USER"; return 0 ;;
        esac
    fi
    
    log_info "删除服务用户: $SERVICE_USER"
    
    if [[ $DRY_RUN == false ]]; then
        case $DISTRO in
            debian|arch)
                userdel "$SERVICE_USER" 2>/dev/null || true
                groupdel "$SERVICE_USER" 2>/dev/null || true
                ;;
            rhel)
                userdel "$SERVICE_USER" 2>/dev/null || true
                ;;
            macos)
                dscl . -delete /Users/"$SERVICE_USER" 2>/dev/null || true
                dscl . -delete /Groups/"$SERVICE_USER" 2>/dev/null || true
                ;;
        esac
        
        log_success "服务用户已删除: $SERVICE_USER"
    else
        log_info "[DRY-RUN] 将删除服务用户: $SERVICE_USER"
    fi
}

# 清理包管理器
cleanup_package_manager() {
    log_info "清理包管理器缓存..."
    
    case $DISTRO in
        debian)
            if [[ $DRY_RUN == false ]]; then
                apt-get autoremove -y 2>/dev/null || true
                apt-get autoclean 2>/dev/null || true
            else
                log_info "[DRY-RUN] 将清理APT缓存"
            fi
            ;;
        rhel)
            if [[ $DRY_RUN == false ]]; then
                yum autoremove -y 2>/dev/null || true
                yum clean all 2>/dev/null || true
            else
                log_info "[DRY-RUN] 将清理YUM缓存"
            fi
            ;;
        arch)
            if [[ $DRY_RUN == false ]]; then
                pacman -Sc --noconfirm 2>/dev/null || true
            else
                log_info "[DRY-RUN] 将清理Pacman缓存"
            fi
            ;;
        macos)
            if command -v brew &> /dev/null; then
                if [[ $DRY_RUN == false ]]; then
                    brew cleanup 2>/dev/null || true
                else
                    log_info "[DRY-RUN] 将清理Homebrew缓存"
                fi
            fi
            ;;
    esac
}

# 验证卸载
verify_uninstall() {
    if [[ $DRY_RUN == true ]]; then
        log_info "预览模式，跳过验证"
        return 0
    fi
    
    log_info "验证卸载结果..."
    
    local remaining_files=()
    
    # 检查二进制文件
    if [[ -f "$INSTALL_PREFIX/bin/go-musicfox" ]]; then
        remaining_files+=("$INSTALL_PREFIX/bin/go-musicfox")
    fi
    
    # 检查服务文件
    case $INIT_SYSTEM in
        systemd)
            if [[ -f "/etc/systemd/system/go-musicfox.service" ]]; then
                remaining_files+=("/etc/systemd/system/go-musicfox.service")
            fi
            ;;
        launchd)
            if [[ -f "/Library/LaunchDaemons/com.go-musicfox.plist" ]]; then
                remaining_files+=("/Library/LaunchDaemons/com.go-musicfox.plist")
            fi
            ;;
    esac
    
    # 检查配置和数据目录（如果应该被删除）
    if [[ $KEEP_CONFIG == false && -d "$CONFIG_DIR" ]]; then
        remaining_files+=("$CONFIG_DIR")
    fi
    
    if [[ $KEEP_DATA == false && -d "$DATA_DIR" ]]; then
        remaining_files+=("$DATA_DIR")
    fi
    
    if [[ ${#remaining_files[@]} -gt 0 ]]; then
        log_warning "以下文件未完全删除:"
        for file in "${remaining_files[@]}"; do
            log_warning "  $file"
        done
    else
        log_success "卸载验证通过，所有文件已删除"
    fi
}

# 显示卸载摘要
show_uninstall_summary() {
    if [[ $DRY_RUN == true ]]; then
        log_info "预览模式完成，未实际删除任何文件"
        return 0
    fi
    
    log_success "go-musicfox v2 卸载完成！"
    echo
    
    log_info "卸载摘要:"
    log_info "  ✓ 服务已停止并卸载"
    log_info "  ✓ 二进制文件已删除"
    log_info "  ✓ 文档已删除"
    log_info "  ✓ 日志文件已删除"
    
    if [[ $KEEP_CONFIG == false ]]; then
        log_info "  ✓ 配置文件已删除"
    else
        log_info "  ⚠ 配置文件已保留"
    fi
    
    if [[ $KEEP_DATA == false ]]; then
        log_info "  ✓ 数据文件已删除"
    else
        log_info "  ⚠ 数据文件已保留"
    fi
    
    if [[ $KEEP_USER == false ]]; then
        log_info "  ✓ 服务用户已删除"
    else
        log_info "  ⚠ 服务用户已保留"
    fi
    
    echo
    log_info "感谢使用 go-musicfox v2！"
    
    if [[ $KEEP_CONFIG == true || $KEEP_DATA == true ]]; then
        echo
        log_info "保留的文件:"
        [[ $KEEP_CONFIG == true ]] && log_info "  配置目录: $CONFIG_DIR"
        [[ $KEEP_DATA == true ]] && log_info "  数据目录: $DATA_DIR"
        log_info "如需完全清理，请手动删除这些目录"
    fi
}

# 主函数
main() {
    log_info "go-musicfox v2 卸载程序"
    
    detect_os
    parse_args "$@"
    check_permissions
    confirm_uninstall
    
    # 执行卸载步骤
    stop_service
    uninstall_service
    remove_binaries
    remove_config
    remove_data
    remove_documentation
    remove_logs
    remove_user
    cleanup_package_manager
    
    verify_uninstall
    show_uninstall_summary
}

# 执行主函数
main "$@"