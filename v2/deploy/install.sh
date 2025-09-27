#!/bin/bash
# go-musicfox v2 安装脚本
# 支持多平台自动安装和配置

set -euo pipefail

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VERSION="${MUSICFOX_VERSION:-latest}"
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
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        OS="windows"
        DISTRO="windows"
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
go-musicfox v2 安装脚本

用法: $0 [选项]

选项:
  -h, --help              显示帮助信息
  -v, --version VERSION   指定安装版本 (默认: latest)
  -p, --prefix PREFIX     安装前缀 (默认: /usr/local)
  -c, --config-dir DIR    配置目录
  -d, --data-dir DIR      数据目录
  -u, --user USER         服务用户 (默认: musicfox)
  -s, --service           安装为系统服务
  -f, --force             强制安装
  --no-deps               跳过依赖安装
  --no-service            不安装系统服务
  --uninstall             卸载

环境变量:
  MUSICFOX_VERSION        安装版本
  INSTALL_PREFIX          安装前缀
  CONFIG_DIR              配置目录
  DATA_DIR                数据目录
  SERVICE_USER            服务用户

示例:
  $0                      # 默认安装
  $0 -v v2.1.0            # 安装指定版本
  $0 -p /opt/musicfox     # 安装到指定目录
  $0 -s                   # 安装为系统服务
  $0 --uninstall          # 卸载

EOF
}

# 解析命令行参数
parse_args() {
    INSTALL_SERVICE=false
    FORCE_INSTALL=false
    SKIP_DEPS=false
    NO_SERVICE=false
    UNINSTALL=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help) show_help; exit 0 ;;
            -v|--version) VERSION="$2"; shift 2 ;;
            -p|--prefix) INSTALL_PREFIX="$2"; shift 2 ;;
            -c|--config-dir) CONFIG_DIR="$2"; shift 2 ;;
            -d|--data-dir) DATA_DIR="$2"; shift 2 ;;
            -u|--user) SERVICE_USER="$2"; shift 2 ;;
            -s|--service) INSTALL_SERVICE=true; shift ;;
            -f|--force) FORCE_INSTALL=true; shift ;;
            --no-deps) SKIP_DEPS=true; shift ;;
            --no-service) NO_SERVICE=true; shift ;;
            --uninstall) UNINSTALL=true; shift ;;
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
        if [[ $FORCE_INSTALL == false ]]; then
            error_exit "需要root权限安装到系统目录，请使用sudo运行或指定其他安装目录"
        fi
    fi
}

# 安装依赖
install_dependencies() {
    if [[ $SKIP_DEPS == true ]]; then
        log_info "跳过依赖安装"
        return 0
    fi
    
    log_info "安装系统依赖..."
    
    case $DISTRO in
        debian)
            apt-get update
            apt-get install -y \
                curl \
                wget \
                tar \
                gzip \
                ca-certificates \
                libasound2 \
                libportaudio2 \
                libflac8 \
                libtag1v5
            ;;
        rhel)
            yum update -y
            yum install -y \
                curl \
                wget \
                tar \
                gzip \
                ca-certificates \
                alsa-lib \
                portaudio \
                flac \
                taglib
            ;;
        arch)
            pacman -Syu --noconfirm
            pacman -S --noconfirm \
                curl \
                wget \
                tar \
                gzip \
                ca-certificates \
                alsa-lib \
                portaudio \
                flac \
                taglib
            ;;
        macos)
            if command -v brew &> /dev/null; then
                brew install portaudio flac taglib
            else
                log_warning "未检测到Homebrew，请手动安装依赖: portaudio flac taglib"
            fi
            ;;
        windows)
            log_info "Windows版本使用纯Go实现，无需额外依赖"
            ;;
        *)
            log_warning "未知发行版，跳过依赖安装"
            ;;
    esac
    
    log_success "依赖安装完成"
}

# 创建用户
create_user() {
    if [[ $OS == "windows" ]]; then
        return 0
    fi
    
    if [[ $INSTALL_SERVICE == true ]] && ! id "$SERVICE_USER" &>/dev/null; then
        log_info "创建服务用户: $SERVICE_USER"
        
        case $DISTRO in
            debian|arch)
                useradd -r -s /bin/false -d "$DATA_DIR" "$SERVICE_USER"
                ;;
            rhel)
                useradd -r -s /sbin/nologin -d "$DATA_DIR" "$SERVICE_USER"
                ;;
            macos)
                # macOS使用launchd，通常不需要专门的用户
                log_info "macOS系统，跳过用户创建"
                ;;
        esac
    fi
}

# 下载二进制文件
download_binary() {
    local binary_name="go-musicfox-$OS-$ARCH"
    if [[ $OS == "windows" ]]; then
        binary_name="${binary_name}.exe"
    fi
    
    local download_url
    if [[ $VERSION == "latest" ]]; then
        download_url="https://github.com/go-musicfox/go-musicfox/releases/latest/download/go-musicfox-$VERSION-$OS.tar.gz"
    else
        download_url="https://github.com/go-musicfox/go-musicfox/releases/download/$VERSION/go-musicfox-$VERSION-$OS.tar.gz"
    fi
    
    log_info "下载 go-musicfox $VERSION for $OS/$ARCH..."
    log_info "下载地址: $download_url"
    
    local temp_dir=$(mktemp -d)
    local archive_file="$temp_dir/go-musicfox.tar.gz"
    
    if ! curl -L -o "$archive_file" "$download_url"; then
        error_exit "下载失败: $download_url"
    fi
    
    log_info "解压安装包..."
    tar -xzf "$archive_file" -C "$temp_dir"
    
    # 查找解压后的目录
    local extract_dir=$(find "$temp_dir" -maxdepth 1 -type d -name "go-musicfox-*" | head -1)
    if [[ -z "$extract_dir" ]]; then
        error_exit "未找到解压后的目录"
    fi
    
    BINARY_DIR="$extract_dir"
    log_success "下载完成: $BINARY_DIR"
}

# 安装文件
install_files() {
    log_info "安装文件到 $INSTALL_PREFIX..."
    
    # 创建目录
    mkdir -p "$INSTALL_PREFIX/bin"
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$DATA_DIR"
    mkdir -p "$DATA_DIR/plugins"
    mkdir -p "$DATA_DIR/logs"
    
    # 安装二进制文件
    local binary_name="go-musicfox"
    if [[ $OS == "windows" ]]; then
        binary_name="${binary_name}.exe"
    fi
    
    cp "$BINARY_DIR/bin/$binary_name" "$INSTALL_PREFIX/bin/"
    chmod +x "$INSTALL_PREFIX/bin/$binary_name"
    
    # 安装插件
    if [[ -d "$BINARY_DIR/plugins" ]]; then
        cp -r "$BINARY_DIR/plugins"/* "$DATA_DIR/plugins/" || true
    fi
    
    # 安装配置文件
    if [[ -d "$BINARY_DIR/config" ]]; then
        cp -r "$BINARY_DIR/config"/* "$CONFIG_DIR/" || true
    fi
    
    # 安装文档
    if [[ -d "$BINARY_DIR/docs" ]]; then
        mkdir -p "$INSTALL_PREFIX/share/doc/go-musicfox"
        cp -r "$BINARY_DIR/docs"/* "$INSTALL_PREFIX/share/doc/go-musicfox/" || true
    fi
    
    # 复制其他文件
    for file in README.md LICENSE CHANGELOG.md VERSION; do
        if [[ -f "$BINARY_DIR/$file" ]]; then
            cp "$BINARY_DIR/$file" "$INSTALL_PREFIX/share/doc/go-musicfox/" || true
        fi
    done
    
    log_success "文件安装完成"
}

# 设置权限
set_permissions() {
    log_info "设置文件权限..."
    
    # 设置配置目录权限
    if [[ $INSTALL_SERVICE == true ]]; then
        chown -R "$SERVICE_USER:$SERVICE_USER" "$CONFIG_DIR" "$DATA_DIR" || true
        chmod -R 755 "$CONFIG_DIR"
        chmod -R 755 "$DATA_DIR"
    else
        chmod -R 755 "$CONFIG_DIR"
        chmod -R 755 "$DATA_DIR"
    fi
    
    log_success "权限设置完成"
}

# 安装系统服务
install_service() {
    if [[ $INSTALL_SERVICE == false || $NO_SERVICE == true ]]; then
        return 0
    fi
    
    log_info "安装系统服务..."
    
    case $DISTRO in
        debian|rhel|arch)
            install_systemd_service
            ;;
        macos)
            install_launchd_service
            ;;
        windows)
            log_warning "Windows服务安装需要手动配置"
            ;;
        *)
            log_warning "不支持的系统，跳过服务安装"
            ;;
    esac
}

# 安装systemd服务
install_systemd_service() {
    local service_file="/etc/systemd/system/go-musicfox.service"
    
    cat > "$service_file" << EOF
[Unit]
Description=go-musicfox v2 Music Player
After=network.target
Wants=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
ExecStart=$INSTALL_PREFIX/bin/go-musicfox --daemon --config=$CONFIG_DIR/config.yaml
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=5
Environment=MUSICFOX_CONFIG_DIR=$CONFIG_DIR
Environment=MUSICFOX_DATA_DIR=$DATA_DIR
Environment=MUSICFOX_LOG_DIR=$DATA_DIR/logs
WorkingDirectory=$DATA_DIR

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$DATA_DIR $CONFIG_DIR

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    systemctl enable go-musicfox
    
    log_success "systemd服务安装完成"
    log_info "使用以下命令管理服务:"
    log_info "  启动: sudo systemctl start go-musicfox"
    log_info "  停止: sudo systemctl stop go-musicfox"
    log_info "  状态: sudo systemctl status go-musicfox"
}

# 安装launchd服务
install_launchd_service() {
    local plist_file="/Library/LaunchDaemons/com.go-musicfox.plist"
    
    cat > "$plist_file" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.go-musicfox</string>
    <key>ProgramArguments</key>
    <array>
        <string>$INSTALL_PREFIX/bin/go-musicfox</string>
        <string>--daemon</string>
        <string>--config=$CONFIG_DIR/config.yaml</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>WorkingDirectory</key>
    <string>$DATA_DIR</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>MUSICFOX_CONFIG_DIR</key>
        <string>$CONFIG_DIR</string>
        <key>MUSICFOX_DATA_DIR</key>
        <string>$DATA_DIR</string>
        <key>MUSICFOX_LOG_DIR</key>
        <string>$DATA_DIR/logs</string>
    </dict>
</dict>
</plist>
EOF
    
    launchctl load "$plist_file"
    
    log_success "launchd服务安装完成"
    log_info "使用以下命令管理服务:"
    log_info "  启动: sudo launchctl start com.go-musicfox"
    log_info "  停止: sudo launchctl stop com.go-musicfox"
}

# 创建符号链接
create_symlinks() {
    if [[ "$INSTALL_PREFIX/bin" != "/usr/local/bin" && "$INSTALL_PREFIX/bin" != "/usr/bin" ]]; then
        log_info "创建符号链接..."
        
        local link_target="/usr/local/bin/go-musicfox"
        if [[ -d "/usr/local/bin" ]]; then
            ln -sf "$INSTALL_PREFIX/bin/go-musicfox" "$link_target" || true
            log_success "符号链接已创建: $link_target"
        fi
    fi
}

# 验证安装
verify_installation() {
    log_info "验证安装..."
    
    local binary_path="$INSTALL_PREFIX/bin/go-musicfox"
    if [[ ! -f "$binary_path" ]]; then
        error_exit "二进制文件不存在: $binary_path"
    fi
    
    if [[ ! -x "$binary_path" ]]; then
        error_exit "二进制文件不可执行: $binary_path"
    fi
    
    # 测试版本
    local version_output
    if version_output=$("$binary_path" --version 2>&1); then
        log_success "安装验证成功: $version_output"
    else
        log_warning "版本检查失败，但文件已安装"
    fi
}

# 卸载
uninstall() {
    log_info "开始卸载 go-musicfox..."
    
    # 停止服务
    case $DISTRO in
        debian|rhel|arch)
            if systemctl is-active --quiet go-musicfox 2>/dev/null; then
                systemctl stop go-musicfox
            fi
            if systemctl is-enabled --quiet go-musicfox 2>/dev/null; then
                systemctl disable go-musicfox
            fi
            rm -f /etc/systemd/system/go-musicfox.service
            systemctl daemon-reload
            ;;
        macos)
            launchctl unload /Library/LaunchDaemons/com.go-musicfox.plist 2>/dev/null || true
            rm -f /Library/LaunchDaemons/com.go-musicfox.plist
            ;;
    esac
    
    # 删除文件
    rm -f "$INSTALL_PREFIX/bin/go-musicfox"
    rm -f "/usr/local/bin/go-musicfox"
    rm -rf "$INSTALL_PREFIX/share/doc/go-musicfox"
    
    # 询问是否删除配置和数据
    if [[ $FORCE_INSTALL == false ]]; then
        echo -n "是否删除配置和数据目录? [y/N] "
        read -r response
        case $response in
            [yY]|[yY][eE][sS])
                rm -rf "$CONFIG_DIR"
                rm -rf "$DATA_DIR"
                log_info "配置和数据已删除"
                ;;
            *)
                log_info "保留配置和数据目录"
                ;;
        esac
    fi
    
    # 删除用户
    if id "$SERVICE_USER" &>/dev/null && [[ "$SERVICE_USER" != "root" ]]; then
        if [[ $FORCE_INSTALL == false ]]; then
            echo -n "是否删除服务用户 $SERVICE_USER? [y/N] "
            read -r response
            case $response in
                [yY]|[yY][eE][sS])
                    userdel "$SERVICE_USER" 2>/dev/null || true
                    log_info "用户 $SERVICE_USER 已删除"
                    ;;
            esac
        fi
    fi
    
    log_success "卸载完成"
}

# 显示安装信息
show_install_info() {
    log_success "go-musicfox v2 安装完成！"
    echo
    log_info "安装信息:"
    log_info "  版本: $VERSION"
    log_info "  二进制文件: $INSTALL_PREFIX/bin/go-musicfox"
    log_info "  配置目录: $CONFIG_DIR"
    log_info "  数据目录: $DATA_DIR"
    
    if [[ $INSTALL_SERVICE == true ]]; then
        log_info "  系统服务: 已安装"
    fi
    
    echo
    log_info "使用方法:"
    log_info "  命令行运行: go-musicfox"
    log_info "  查看帮助: go-musicfox --help"
    log_info "  查看版本: go-musicfox --version"
    
    if [[ $INSTALL_SERVICE == true ]]; then
        echo
        log_info "服务管理:"
        case $DISTRO in
            debian|rhel|arch)
                log_info "  启动服务: sudo systemctl start go-musicfox"
                log_info "  停止服务: sudo systemctl stop go-musicfox"
                log_info "  查看状态: sudo systemctl status go-musicfox"
                log_info "  查看日志: sudo journalctl -u go-musicfox -f"
                ;;
            macos)
                log_info "  启动服务: sudo launchctl start com.go-musicfox"
                log_info "  停止服务: sudo launchctl stop com.go-musicfox"
                ;;
        esac
    fi
    
    echo
    log_info "配置文件位置: $CONFIG_DIR/config.yaml"
    log_info "插件目录: $DATA_DIR/plugins"
    log_info "日志目录: $DATA_DIR/logs"
    
    echo
    log_info "文档和支持:"
    log_info "  用户手册: https://github.com/go-musicfox/go-musicfox/tree/v2/docs"
    log_info "  问题反馈: https://github.com/go-musicfox/go-musicfox/issues"
}

# 主函数
main() {
    log_info "go-musicfox v2 安装程序"
    
    detect_os
    parse_args "$@"
    
    if [[ $UNINSTALL == true ]]; then
        uninstall
        exit 0
    fi
    
    check_permissions
    install_dependencies
    create_user
    download_binary
    install_files
    set_permissions
    install_service
    create_symlinks
    verify_installation
    
    # 清理临时文件
    if [[ -n "${BINARY_DIR:-}" ]]; then
        rm -rf "$(dirname "$BINARY_DIR")"
    fi
    
    show_install_info
}

# 执行主函数
main "$@"