#!/bin/bash
# go-musicfox v2 打包脚本
# 创建发布包，支持多种格式和平台

set -euo pipefail

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"
DIST_DIR="$PROJECT_ROOT/dist"
DOCS_DIR="$PROJECT_ROOT/docs"
CONFIGS_DIR="$PROJECT_ROOT/configs"

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
go-musicfox v2 打包脚本

用法: $0 [选项] [版本号]

选项:
  -h, --help          显示帮助信息
  -v, --verbose       详细输出
  -c, --clean         打包前清理
  -f, --format        指定打包格式 (tar.gz|zip|deb|rpm|dmg|msi|all)
  -p, --platforms     指定平台列表 (用逗号分隔)
  -o, --output        指定输出目录
  --include-source    包含源代码
  --include-docs      包含文档
  --sign              对包进行签名
  --checksum          生成校验和文件

打包格式:
  tar.gz              通用压缩包 (默认)
  zip                 Windows压缩包
  deb                 Debian/Ubuntu包
  rpm                 RedHat/CentOS包
  dmg                 macOS磁盘镜像
  msi                 Windows安装包
  all                 所有格式

示例:
  $0 v2.1.0                           # 打包v2.1.0版本
  $0 -f zip -p windows v2.1.0         # 仅打包Windows ZIP
  $0 --include-docs --checksum v2.1.0 # 包含文档并生成校验和

EOF
}

# 默认配置
VERBOSE=false
CLEAN=false
FORMAT="tar.gz"
PLATFORMS="linux,darwin,windows"
OUTPUT_DIR="$DIST_DIR"
INCLUDE_SOURCE=false
INCLUDE_DOCS=true
SIGN=false
CHECKSUM=true
VERSION=""

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help) show_help; exit 0 ;;
        -v|--verbose) VERBOSE=true; shift ;;
        -c|--clean) CLEAN=true; shift ;;
        -f|--format) FORMAT="$2"; shift 2 ;;
        -p|--platforms) PLATFORMS="$2"; shift 2 ;;
        -o|--output) OUTPUT_DIR="$2"; shift 2 ;;
        --include-source) INCLUDE_SOURCE=true; shift ;;
        --include-docs) INCLUDE_DOCS=true; shift ;;
        --sign) SIGN=true; shift ;;
        --checksum) CHECKSUM=true; shift ;;
        -*) error_exit "未知选项: $1" ;;
        *) VERSION="$1"; shift ;;
    esac
done

# 验证版本号
if [[ -z "$VERSION" ]]; then
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    log_info "使用自动检测的版本号: $VERSION"
fi

# 验证格式
case $FORMAT in
    tar.gz|zip|deb|rpm|dmg|msi|all) ;;
    *) error_exit "无效的打包格式: $FORMAT" ;;
esac

# 获取构建信息
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION=$(go version | cut -d' ' -f3)

# 准备打包环境
prepare_package() {
    log_info "准备打包环境..."
    
    mkdir -p "$OUTPUT_DIR"
    
    if [[ $CLEAN == true ]]; then
        log_info "清理旧的发布包..."
        rm -rf "$OUTPUT_DIR"/*
    fi
    
    # 检查构建文件
    if [[ ! -d "$BUILD_DIR" ]] || [[ -z "$(ls -A "$BUILD_DIR" 2>/dev/null)" ]]; then
        log_warning "构建目录为空，开始构建..."
        "$SCRIPT_DIR/cross-compile.sh" -p "$PLATFORMS"
    fi
    
    # 检查必要工具
    check_packaging_tools
}

# 检查打包工具
check_packaging_tools() {
    local missing_tools=()
    
    case $FORMAT in
        deb|all)
            if ! command -v dpkg-deb &> /dev/null; then
                missing_tools+=("dpkg-deb")
            fi
            ;;
        rpm|all)
            if ! command -v rpmbuild &> /dev/null; then
                missing_tools+=("rpmbuild")
            fi
            ;;
        dmg|all)
            if [[ "$(uname)" == "Darwin" ]] && ! command -v hdiutil &> /dev/null; then
                missing_tools+=("hdiutil")
            fi
            ;;
        msi|all)
            if ! command -v wixl &> /dev/null && ! command -v candle &> /dev/null; then
                missing_tools+=("wix-toolset")
            fi
            ;;
    esac
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        log_warning "缺少打包工具: ${missing_tools[*]}"
        log_warning "将跳过相应格式的打包"
    fi
}

# 创建基础包结构
create_base_package() {
    local platform="$1"
    local package_dir="$2"
    
    log_info "创建基础包结构: $platform"
    
    mkdir -p "$package_dir"/{bin,config,docs,plugins,scripts}
    
    # 复制二进制文件
    local binary_pattern="go-musicfox-$platform-*"
    find "$BUILD_DIR" -name "$binary_pattern" -type f | while read -r binary; do
        cp "$binary" "$package_dir/bin/"
        # 创建通用名称的符号链接
        local generic_name="go-musicfox"
        if [[ $platform == "windows" ]]; then
            generic_name="${generic_name}.exe"
        fi
        ln -sf "$(basename "$binary")" "$package_dir/bin/$generic_name"
    done
    
    # 复制插件
    if [[ -d "$BUILD_DIR/plugins" ]]; then
        find "$BUILD_DIR/plugins" -type f | while read -r plugin; do
            local plugin_type=$(basename "$(dirname "$plugin")")
            mkdir -p "$package_dir/plugins/$plugin_type"
            cp "$plugin" "$package_dir/plugins/$plugin_type/"
        done
    fi
    
    # 复制配置文件
    if [[ -d "$CONFIGS_DIR" ]]; then
        cp -r "$CONFIGS_DIR"/* "$package_dir/config/" 2>/dev/null || true
    fi
    
    # 复制文档
    if [[ $INCLUDE_DOCS == true && -d "$DOCS_DIR" ]]; then
        cp -r "$DOCS_DIR"/* "$package_dir/docs/" 2>/dev/null || true
    fi
    
    # 复制项目文件
    cp "$PROJECT_ROOT/README.md" "$package_dir/" 2>/dev/null || true
    cp "$PROJECT_ROOT/LICENSE" "$package_dir/" 2>/dev/null || true
    cp "$PROJECT_ROOT/CHANGELOG.md" "$package_dir/" 2>/dev/null || true
    
    # 创建安装脚本
    create_install_scripts "$platform" "$package_dir"
    
    # 创建版本信息文件
    create_version_file "$package_dir"
}

# 创建安装脚本
create_install_scripts() {
    local platform="$1"
    local package_dir="$2"
    
    case $platform in
        linux)
            create_linux_install_script "$package_dir"
            ;;
        darwin)
            create_macos_install_script "$package_dir"
            ;;
        windows)
            create_windows_install_script "$package_dir"
            ;;
    esac
}

# 创建Linux安装脚本
create_linux_install_script() {
    local package_dir="$1"
    
    cat > "$package_dir/scripts/install.sh" << 'EOF'
#!/bin/bash
# go-musicfox v2 Linux安装脚本

set -e

INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/go-musicfox"
PLUGIN_DIR="$HOME/.local/share/go-musicfox/plugins"

echo "安装 go-musicfox v2..."

# 创建目录
mkdir -p "$CONFIG_DIR" "$PLUGIN_DIR"

# 安装二进制文件
sudo cp bin/go-musicfox "$INSTALL_DIR/"
sudo chmod +x "$INSTALL_DIR/go-musicfox"

# 复制配置文件
cp -r config/* "$CONFIG_DIR/" 2>/dev/null || true

# 复制插件
cp -r plugins/* "$PLUGIN_DIR/" 2>/dev/null || true

echo "安装完成！"
echo "运行 'go-musicfox' 开始使用"
EOF
    
    chmod +x "$package_dir/scripts/install.sh"
}

# 创建macOS安装脚本
create_macos_install_script() {
    local package_dir="$1"
    
    cat > "$package_dir/scripts/install.sh" << 'EOF'
#!/bin/bash
# go-musicfox v2 macOS安装脚本

set -e

INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/go-musicfox"
PLUGIN_DIR="$HOME/Library/Application Support/go-musicfox/plugins"

echo "安装 go-musicfox v2..."

# 创建目录
mkdir -p "$CONFIG_DIR" "$PLUGIN_DIR"

# 安装二进制文件
sudo cp bin/go-musicfox "$INSTALL_DIR/"
sudo chmod +x "$INSTALL_DIR/go-musicfox"

# 复制配置文件
cp -r config/* "$CONFIG_DIR/" 2>/dev/null || true

# 复制插件
cp -r plugins/* "$PLUGIN_DIR/" 2>/dev/null || true

echo "安装完成！"
echo "运行 'go-musicfox' 开始使用"
EOF
    
    chmod +x "$package_dir/scripts/install.sh"
}

# 创建Windows安装脚本
create_windows_install_script() {
    local package_dir="$1"
    
    cat > "$package_dir/scripts/install.bat" << 'EOF'
@echo off
REM go-musicfox v2 Windows安装脚本

echo 安装 go-musicfox v2...

set INSTALL_DIR=%ProgramFiles%\go-musicfox
set CONFIG_DIR=%APPDATA%\go-musicfox
set PLUGIN_DIR=%LOCALAPPDATA%\go-musicfox\plugins

REM 创建目录
mkdir "%INSTALL_DIR%" 2>nul
mkdir "%CONFIG_DIR%" 2>nul
mkdir "%PLUGIN_DIR%" 2>nul

REM 复制文件
copy bin\go-musicfox.exe "%INSTALL_DIR%\"
xcopy config "%CONFIG_DIR%" /E /I /Q 2>nul
xcopy plugins "%PLUGIN_DIR%" /E /I /Q 2>nul

REM 添加到PATH
setx PATH "%PATH%;%INSTALL_DIR%" /M

echo 安装完成！
echo 运行 'go-musicfox' 开始使用
pause
EOF
}

# 创建版本信息文件
create_version_file() {
    local package_dir="$1"
    
    cat > "$package_dir/VERSION" << EOF
go-musicfox v2
==============

版本: $VERSION
提交: $COMMIT
构建时间: $BUILD_TIME
Go版本: $GO_VERSION

微内核插件架构
支持动态链接库、RPC、WebAssembly、热加载插件
EOF
}

# 创建tar.gz包
create_tar_package() {
    local platform="$1"
    local package_dir="$2"
    
    local package_name="go-musicfox-$VERSION-$platform"
    local archive_name="${package_name}.tar.gz"
    
    log_info "创建tar.gz包: $archive_name"
    
    cd "$(dirname "$package_dir")"
    tar -czf "$OUTPUT_DIR/$archive_name" "$(basename "$package_dir")"
    
    log_success "tar.gz包创建完成: $archive_name"
}

# 创建zip包
create_zip_package() {
    local platform="$1"
    local package_dir="$2"
    
    local package_name="go-musicfox-$VERSION-$platform"
    local archive_name="${package_name}.zip"
    
    log_info "创建zip包: $archive_name"
    
    cd "$(dirname "$package_dir")"
    zip -r "$OUTPUT_DIR/$archive_name" "$(basename "$package_dir")" >/dev/null
    
    log_success "zip包创建完成: $archive_name"
}

# 创建DEB包
create_deb_package() {
    local platform="$1"
    local package_dir="$2"
    
    if [[ $platform != "linux" ]] || ! command -v dpkg-deb &> /dev/null; then
        return 0
    fi
    
    local package_name="go-musicfox-$VERSION-amd64"
    local deb_dir="$OUTPUT_DIR/deb-build"
    
    log_info "创建DEB包: ${package_name}.deb"
    
    mkdir -p "$deb_dir/DEBIAN"
    mkdir -p "$deb_dir/usr/local/bin"
    mkdir -p "$deb_dir/etc/go-musicfox"
    mkdir -p "$deb_dir/usr/share/go-musicfox/plugins"
    
    # 复制文件
    cp "$package_dir/bin/go-musicfox"* "$deb_dir/usr/local/bin/"
    cp -r "$package_dir/config"/* "$deb_dir/etc/go-musicfox/" 2>/dev/null || true
    cp -r "$package_dir/plugins"/* "$deb_dir/usr/share/go-musicfox/plugins/" 2>/dev/null || true
    
    # 创建控制文件
    cat > "$deb_dir/DEBIAN/control" << EOF
Package: go-musicfox
Version: ${VERSION#v}
Section: sound
Priority: optional
Architecture: amd64
Maintainer: go-musicfox Team <team@go-musicfox.com>
Description: 基于微内核插件架构的音乐播放器
 支持多种音乐源和插件扩展的终端音乐播放器
EOF
    
    # 构建DEB包
    dpkg-deb --build "$deb_dir" "$OUTPUT_DIR/${package_name}.deb"
    
    rm -rf "$deb_dir"
    
    log_success "DEB包创建完成: ${package_name}.deb"
}

# 生成校验和文件
generate_checksums() {
    log_info "生成校验和文件..."
    
    local checksum_file="$OUTPUT_DIR/checksums.txt"
    
    cd "$OUTPUT_DIR"
    
    # 生成SHA256校验和
    find . -name "*.tar.gz" -o -name "*.zip" -o -name "*.deb" -o -name "*.rpm" -o -name "*.dmg" -o -name "*.msi" | while read -r file; do
        if command -v sha256sum &> /dev/null; then
            sha256sum "$file" >> "$checksum_file"
        elif command -v shasum &> /dev/null; then
            shasum -a 256 "$file" >> "$checksum_file"
        fi
    done
    
    if [[ -f "$checksum_file" ]]; then
        log_success "校验和文件已生成: checksums.txt"
    fi
}

# 签名包文件
sign_packages() {
    if [[ $SIGN != true ]]; then
        return 0
    fi
    
    log_info "签名包文件..."
    
    if ! command -v gpg &> /dev/null; then
        log_warning "GPG未安装，跳过签名"
        return 0
    fi
    
    cd "$OUTPUT_DIR"
    
    find . -name "*.tar.gz" -o -name "*.zip" -o -name "*.deb" -o -name "*.rpm" -o -name "*.dmg" -o -name "*.msi" | while read -r file; do
        gpg --detach-sign --armor "$file"
        log_info "已签名: $file"
    done
    
    log_success "包文件签名完成"
}

# 生成发布说明
generate_release_notes() {
    local release_notes="$OUTPUT_DIR/RELEASE_NOTES.md"
    
    log_info "生成发布说明..."
    
    cat > "$release_notes" << EOF
# go-musicfox v2 $VERSION 发布说明

## 版本信息

- **版本**: $VERSION
- **提交**: $COMMIT
- **构建时间**: $BUILD_TIME
- **Go版本**: $GO_VERSION

## 新特性

- 微内核插件架构
- 支持四种插件类型：动态链接库、RPC、WebAssembly、热加载
- 多平台支持：Linux、macOS、Windows
- 完整的插件开发工具链
- 容器化部署支持

## 安装说明

### Linux/macOS

1. 下载对应平台的tar.gz包
2. 解压: \`tar -xzf go-musicfox-$VERSION-<platform>.tar.gz\`
3. 运行安装脚本: \`./scripts/install.sh\`

### Windows

1. 下载Windows zip包
2. 解压到目标目录
3. 运行 \`scripts/install.bat\` (需要管理员权限)

### 包管理器

- **Debian/Ubuntu**: \`sudo dpkg -i go-musicfox-$VERSION-amd64.deb\`
- **RedHat/CentOS**: \`sudo rpm -i go-musicfox-$VERSION-x86_64.rpm\`

## 文件校验

请使用 \`checksums.txt\` 文件验证下载文件的完整性：

\`\`\`bash
sha256sum -c checksums.txt
\`\`\`

## 支持

- 文档: https://github.com/go-musicfox/go-musicfox/tree/v2/docs
- 问题反馈: https://github.com/go-musicfox/go-musicfox/issues
- 插件开发: https://github.com/go-musicfox/go-musicfox/tree/v2/docs/guides

EOF
    
    log_success "发布说明已生成: RELEASE_NOTES.md"
}

# 主函数
main() {
    log_info "开始打包 go-musicfox v2 $VERSION..."
    
    prepare_package
    
    local platforms=(${PLATFORMS//,/ })
    local temp_dir="$OUTPUT_DIR/temp"
    
    mkdir -p "$temp_dir"
    
    for platform in "${platforms[@]}"; do
        local package_dir="$temp_dir/go-musicfox-$VERSION-$platform"
        
        create_base_package "$platform" "$package_dir"
        
        case $FORMAT in
            tar.gz)
                create_tar_package "$platform" "$package_dir"
                ;;
            zip)
                create_zip_package "$platform" "$package_dir"
                ;;
            deb)
                create_deb_package "$platform" "$package_dir"
                ;;
            all)
                create_tar_package "$platform" "$package_dir"
                create_zip_package "$platform" "$package_dir"
                create_deb_package "$platform" "$package_dir"
                ;;
        esac
    done
    
    # 清理临时目录
    rm -rf "$temp_dir"
    
    # 后处理
    if [[ $CHECKSUM == true ]]; then
        generate_checksums
    fi
    
    sign_packages
    generate_release_notes
    
    log_success "打包完成！输出目录: $OUTPUT_DIR"
    
    # 显示生成的文件
    if [[ $VERBOSE == true ]]; then
        log_info "生成的文件:"
        find "$OUTPUT_DIR" -type f -name "*" | sort | while read -r file; do
            local size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo "0")
            local human_size=$(numfmt --to=iec-i --suffix=B "$size" 2>/dev/null || echo "${size}B")
            echo "  $(basename "$file") ($human_size)"
        done
    fi
}

# 执行主函数
main "$@"