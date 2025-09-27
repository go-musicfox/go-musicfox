#!/bin/bash
# go-musicfox v2 插件打包器
# 将插件打包成标准的分发格式

set -euo pipefail

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
PLUGINS_DIR="$PROJECT_ROOT/plugins"
DIST_DIR="$PROJECT_ROOT/dist/plugins"

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
go-musicfox v2 插件打包器

用法: $0 [选项] <插件路径>

选项:
  -h, --help              显示帮助信息
  -o, --output DIR        输出目录 (默认: dist/plugins)
  -f, --format FORMAT     打包格式 (tar.gz|zip|plugin|all)
  -v, --version VERSION   插件版本号
  -n, --name NAME         插件包名称
  -c, --compress LEVEL    压缩级别 (1-9, 默认: 6)
  -s, --sign              对插件包进行签名
  -k, --key FILE          签名密钥文件
  --checksum              生成校验和文件
  --metadata FILE         插件元数据文件
  --include-docs          包含文档
  --include-examples      包含示例代码
  --exclude PATTERN       排除文件模式
  --verbose               详细输出
  --dry-run               预览打包操作

打包格式:
  tar.gz                  压缩包格式 (默认)
  zip                     ZIP压缩包
  plugin                  go-musicfox插件包格式
  all                     所有格式

示例:
  $0 plugins/shared/my-plugin/                    # 打包插件目录
  $0 -f zip -v 1.0.0 plugins/shared/my-plugin/   # 打包为ZIP格式
  $0 --sign --checksum plugins/shared/my-plugin/ # 签名并生成校验和
  $0 -f all plugins/shared/my-plugin/             # 生成所有格式

EOF
}

# 解析命令行参数
parse_args() {
    OUTPUT_DIR="$DIST_DIR"
    PACKAGE_FORMAT="tar.gz"
    PLUGIN_VERSION=""
    PACKAGE_NAME=""
    COMPRESS_LEVEL=6
    SIGN_PACKAGE=false
    SIGN_KEY=""
    GENERATE_CHECKSUM=false
    METADATA_FILE=""
    INCLUDE_DOCS=false
    INCLUDE_EXAMPLES=false
    EXCLUDE_PATTERNS=()
    VERBOSE=false
    DRY_RUN=false
    PLUGIN_PATH=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help) show_help; exit 0 ;;
            -o|--output) OUTPUT_DIR="$2"; shift 2 ;;
            -f|--format) PACKAGE_FORMAT="$2"; shift 2 ;;
            -v|--version) PLUGIN_VERSION="$2"; shift 2 ;;
            -n|--name) PACKAGE_NAME="$2"; shift 2 ;;
            -c|--compress) COMPRESS_LEVEL="$2"; shift 2 ;;
            -s|--sign) SIGN_PACKAGE=true; shift ;;
            -k|--key) SIGN_KEY="$2"; shift 2 ;;
            --checksum) GENERATE_CHECKSUM=true; shift ;;
            --metadata) METADATA_FILE="$2"; shift 2 ;;
            --include-docs) INCLUDE_DOCS=true; shift ;;
            --include-examples) INCLUDE_EXAMPLES=true; shift ;;
            --exclude) EXCLUDE_PATTERNS+=("$2"); shift 2 ;;
            --verbose) VERBOSE=true; shift ;;
            --dry-run) DRY_RUN=true; shift ;;
            -*) error_exit "未知选项: $1" ;;
            *) PLUGIN_PATH="$1"; shift ;;
        esac
    done
    
    # 验证参数
    if [[ -z "$PLUGIN_PATH" ]]; then
        error_exit "请指定插件路径"
    fi
    
    if [[ ! -e "$PLUGIN_PATH" ]]; then
        error_exit "插件路径不存在: $PLUGIN_PATH"
    fi
    
    case $PACKAGE_FORMAT in
        tar.gz|zip|plugin|all) ;;
        *) error_exit "无效的打包格式: $PACKAGE_FORMAT" ;;
    esac
    
    if [[ $COMPRESS_LEVEL -lt 1 || $COMPRESS_LEVEL -gt 9 ]]; then
        error_exit "压缩级别必须在1-9之间"
    fi
    
    if [[ $SIGN_PACKAGE == true && -z "$SIGN_KEY" ]]; then
        log_warning "未指定签名密钥，将使用默认密钥"
    fi
}

# 初始化打包环境
init_packaging() {
    log_info "初始化打包环境..."
    
    # 创建输出目录
    mkdir -p "$OUTPUT_DIR"
    
    # 创建临时目录
    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT
    
    # 检查必要工具
    local missing_tools=()
    
    if ! command -v tar &> /dev/null; then
        missing_tools+=("tar")
    fi
    
    if [[ $PACKAGE_FORMAT == "zip" || $PACKAGE_FORMAT == "all" ]] && ! command -v zip &> /dev/null; then
        missing_tools+=("zip")
    fi
    
    if [[ $SIGN_PACKAGE == true ]] && ! command -v gpg &> /dev/null; then
        missing_tools+=("gpg")
    fi
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        error_exit "缺少必要工具: ${missing_tools[*]}"
    fi
    
    log_success "打包环境初始化完成"
}

# 分析插件信息
analyze_plugin() {
    log_info "分析插件信息..."
    
    PLUGIN_PATH=$(realpath "$PLUGIN_PATH")
    
    if [[ -f "$PLUGIN_PATH" ]]; then
        # 单个文件
        PLUGIN_TYPE="file"
        PLUGIN_NAME=$(basename "$PLUGIN_PATH" | sed 's/\.[^.]*$//')
        PLUGIN_DIR=$(dirname "$PLUGIN_PATH")
        PLUGIN_FILES=("$PLUGIN_PATH")
    elif [[ -d "$PLUGIN_PATH" ]]; then
        # 目录
        PLUGIN_TYPE="directory"
        PLUGIN_NAME=$(basename "$PLUGIN_PATH")
        PLUGIN_DIR="$PLUGIN_PATH"
        
        # 查找插件文件
        PLUGIN_FILES=()
        while IFS= read -r -d '' file; do
            PLUGIN_FILES+=("$file")
        done < <(find "$PLUGIN_PATH" -type f \( -name "*.so" -o -name "*.dll" -o -name "*.dylib" -o -name "*.wasm" -o -executable \) -print0 2>/dev/null)
    else
        error_exit "无效的插件路径: $PLUGIN_PATH"
    fi
    
    # 设置默认包名
    if [[ -z "$PACKAGE_NAME" ]]; then
        PACKAGE_NAME="$PLUGIN_NAME"
    fi
    
    # 检测插件版本
    if [[ -z "$PLUGIN_VERSION" ]]; then
        detect_plugin_version
    fi
    
    # 检测插件类型
    detect_plugin_category
    
    log_info "插件名称: $PLUGIN_NAME"
    log_info "插件版本: $PLUGIN_VERSION"
    log_info "插件类型: $PLUGIN_CATEGORY"
    log_info "插件文件数: ${#PLUGIN_FILES[@]}"
}

# 检测插件版本
detect_plugin_version() {
    # 从文件名检测
    if [[ "$PLUGIN_NAME" =~ v?([0-9]+\.[0-9]+\.[0-9]+) ]]; then
        PLUGIN_VERSION="${BASH_REMATCH[1]}"
        return
    fi
    
    # 从元数据文件检测
    local metadata_files=("$PLUGIN_DIR/plugin.yaml" "$PLUGIN_DIR/plugin.json" "$PLUGIN_DIR/package.json")
    
    for metadata_file in "${metadata_files[@]}"; do
        if [[ -f "$metadata_file" ]]; then
            if command -v yq &> /dev/null && [[ "$metadata_file" == *.yaml ]]; then
                PLUGIN_VERSION=$(yq eval '.version' "$metadata_file" 2>/dev/null || echo "")
            elif command -v jq &> /dev/null && [[ "$metadata_file" == *.json ]]; then
                PLUGIN_VERSION=$(jq -r '.version // empty' "$metadata_file" 2>/dev/null || echo "")
            fi
            
            if [[ -n "$PLUGIN_VERSION" && "$PLUGIN_VERSION" != "null" ]]; then
                break
            fi
        fi
    done
    
    # 从二进制文件检测
    if [[ -z "$PLUGIN_VERSION" && ${#PLUGIN_FILES[@]} -gt 0 ]]; then
        for plugin_file in "${PLUGIN_FILES[@]}"; do
            if command -v strings &> /dev/null; then
                local version=$(strings "$plugin_file" 2>/dev/null | grep -oE 'v?[0-9]+\.[0-9]+\.[0-9]+' | head -1)
                if [[ -n "$version" ]]; then
                    PLUGIN_VERSION="$version"
                    break
                fi
            fi
        done
    fi
    
    # 默认版本
    if [[ -z "$PLUGIN_VERSION" ]]; then
        PLUGIN_VERSION="1.0.0"
        log_warning "未检测到版本信息，使用默认版本: $PLUGIN_VERSION"
    fi
}

# 检测插件分类
detect_plugin_category() {
    PLUGIN_CATEGORY="unknown"
    
    # 从文件扩展名检测
    for plugin_file in "${PLUGIN_FILES[@]}"; do
        case "$plugin_file" in
            *.so|*.dll|*.dylib)
                PLUGIN_CATEGORY="shared"
                break
                ;;
            *.wasm)
                PLUGIN_CATEGORY="wasm"
                break
                ;;
        esac
    done
    
    # 从目录结构检测
    if [[ "$PLUGIN_CATEGORY" == "unknown" ]]; then
        case "$PLUGIN_PATH" in
            */shared/*) PLUGIN_CATEGORY="shared" ;;
            */rpc/*) PLUGIN_CATEGORY="rpc" ;;
            */wasm/*) PLUGIN_CATEGORY="wasm" ;;
            */hotload/*) PLUGIN_CATEGORY="hotload" ;;
        esac
    fi
    
    # 从元数据检测
    if [[ "$PLUGIN_CATEGORY" == "unknown" && -n "$METADATA_FILE" && -f "$METADATA_FILE" ]]; then
        if command -v yq &> /dev/null && [[ "$METADATA_FILE" == *.yaml ]]; then
            PLUGIN_CATEGORY=$(yq eval '.type // .category' "$METADATA_FILE" 2>/dev/null || echo "unknown")
        elif command -v jq &> /dev/null && [[ "$METADATA_FILE" == *.json ]]; then
            PLUGIN_CATEGORY=$(jq -r '.type // .category // "unknown"' "$METADATA_FILE" 2>/dev/null)
        fi
    fi
}

# 准备打包内容
prepare_package_content() {
    log_info "准备打包内容..."
    
    PACKAGE_DIR="$TEMP_DIR/$PACKAGE_NAME-$PLUGIN_VERSION"
    mkdir -p "$PACKAGE_DIR"
    
    # 复制插件文件
    if [[ $PLUGIN_TYPE == "file" ]]; then
        cp "$PLUGIN_PATH" "$PACKAGE_DIR/"
    else
        # 复制目录内容，排除不需要的文件
        rsync -av --exclude='.git' --exclude='.svn' --exclude='*.tmp' --exclude='*.log' \
              "$PLUGIN_PATH/" "$PACKAGE_DIR/"
        
        # 应用排除模式
        for pattern in "${EXCLUDE_PATTERNS[@]}"; do
            find "$PACKAGE_DIR" -name "$pattern" -delete 2>/dev/null || true
        done
    fi
    
    # 生成插件清单
    generate_plugin_manifest
    
    # 生成元数据
    generate_plugin_metadata
    
    # 包含文档
    if [[ $INCLUDE_DOCS == true ]]; then
        include_documentation
    fi
    
    # 包含示例
    if [[ $INCLUDE_EXAMPLES == true ]]; then
        include_examples
    fi
    
    # 生成安装脚本
    generate_install_script
    
    log_success "打包内容准备完成"
}

# 生成插件清单
generate_plugin_manifest() {
    log_info "生成插件清单..."
    
    local manifest_file="$PACKAGE_DIR/MANIFEST"
    
    cat > "$manifest_file" << EOF
# go-musicfox v2 插件清单
# 生成时间: $(date -u +%Y-%m-%dT%H:%M:%SZ)

[plugin]
name = "$PLUGIN_NAME"
version = "$PLUGIN_VERSION"
category = "$PLUGIN_CATEGORY"
package_format = "$PACKAGE_FORMAT"
package_time = "$(date -u +%Y-%m-%dT%H:%M:%SZ)"

[files]
EOF
    
    # 列出所有文件
    find "$PACKAGE_DIR" -type f -not -name "MANIFEST" | while read -r file; do
        local rel_path=$(realpath --relative-to="$PACKAGE_DIR" "$file")
        local file_size=$(stat -c%s "$file" 2>/dev/null || stat -f%z "$file" 2>/dev/null)
        local file_hash=$(sha256sum "$file" 2>/dev/null | cut -d' ' -f1 || shasum -a 256 "$file" 2>/dev/null | cut -d' ' -f1)
        
        echo "$rel_path = { size = $file_size, sha256 = \"$file_hash\" }" >> "$manifest_file"
    done
}

# 生成插件元数据
generate_plugin_metadata() {
    log_info "生成插件元数据..."
    
    local metadata_file="$PACKAGE_DIR/plugin.json"
    
    # 如果已有元数据文件，则合并
    local existing_metadata="{}"
    if [[ -n "$METADATA_FILE" && -f "$METADATA_FILE" ]]; then
        if command -v jq &> /dev/null; then
            existing_metadata=$(cat "$METADATA_FILE")
        fi
    elif [[ -f "$PACKAGE_DIR/plugin.json" ]]; then
        existing_metadata=$(cat "$PACKAGE_DIR/plugin.json")
    fi
    
    # 生成完整元数据
    jq -n \
        --arg name "$PLUGIN_NAME" \
        --arg version "$PLUGIN_VERSION" \
        --arg category "$PLUGIN_CATEGORY" \
        --arg package_time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        --argjson existing "$existing_metadata" \
        '$existing + {
            name: $name,
            version: $version,
            category: $category,
            package_time: $package_time,
            go_musicfox_version: "v2.0.0",
            api_version: "1.0"
        }' > "$metadata_file"
}

# 包含文档
include_documentation() {
    log_info "包含文档..."
    
    local docs_dir="$PACKAGE_DIR/docs"
    mkdir -p "$docs_dir"
    
    # 查找文档文件
    local doc_files=()
    
    if [[ $PLUGIN_TYPE == "directory" ]]; then
        while IFS= read -r -d '' file; do
            doc_files+=("$file")
        done < <(find "$PLUGIN_DIR" -name "*.md" -o -name "*.txt" -o -name "*.rst" -print0 2>/dev/null)
    fi
    
    # 复制文档文件
    for doc_file in "${doc_files[@]}"; do
        cp "$doc_file" "$docs_dir/"
    done
    
    # 生成基础README
    if [[ ! -f "$docs_dir/README.md" ]]; then
        cat > "$docs_dir/README.md" << EOF
# $PLUGIN_NAME

版本: $PLUGIN_VERSION
类型: $PLUGIN_CATEGORY

## 安装

运行安装脚本:

\`\`\`bash
./install.sh
\`\`\`

## 使用

请参考插件文档了解具体使用方法。

## 支持

如有问题，请联系插件作者或提交Issue。
EOF
    fi
}

# 包含示例
include_examples() {
    log_info "包含示例..."
    
    local examples_dir="$PACKAGE_DIR/examples"
    
    if [[ $PLUGIN_TYPE == "directory" && -d "$PLUGIN_DIR/examples" ]]; then
        cp -r "$PLUGIN_DIR/examples" "$PACKAGE_DIR/"
    else
        mkdir -p "$examples_dir"
        
        # 生成基础示例
        cat > "$examples_dir/basic_usage.md" << EOF
# $PLUGIN_NAME 基础使用示例

## 配置

在go-musicfox配置文件中添加:

\`\`\`yaml
plugins:
  $PLUGIN_CATEGORY:
    $PLUGIN_NAME:
      enabled: true
      # 其他配置选项
\`\`\`

## 使用

插件安装后会自动加载，无需额外操作。
EOF
    fi
}

# 生成安装脚本
generate_install_script() {
    log_info "生成安装脚本..."
    
    cat > "$PACKAGE_DIR/install.sh" << 'EOF'
#!/bin/bash
# 插件安装脚本

set -e

PLUGIN_NAME="__PLUGIN_NAME__"
PLUGIN_VERSION="__PLUGIN_VERSION__"
PLUGIN_CATEGORY="__PLUGIN_CATEGORY__"

echo "安装插件: $PLUGIN_NAME v$PLUGIN_VERSION"

# 检测go-musicfox安装路径
MUSICFOX_HOME="${MUSICFOX_HOME:-$HOME/.local/share/go-musicfox}"
PLUGIN_DIR="$MUSICFOX_HOME/plugins/$PLUGIN_CATEGORY"

# 创建插件目录
mkdir -p "$PLUGIN_DIR"

# 复制插件文件
echo "复制插件文件..."
find . -name "*.so" -o -name "*.dll" -o -name "*.dylib" -o -name "*.wasm" | while read -r file; do
    cp "$file" "$PLUGIN_DIR/"
    echo "已安装: $file"
done

# 复制配置文件
if [ -f "plugin.json" ]; then
    cp plugin.json "$PLUGIN_DIR/$PLUGIN_NAME.json"
fi

echo "插件安装完成！"
echo "插件目录: $PLUGIN_DIR"
echo "请重启go-musicfox以加载新插件。"
EOF
    
    # 替换占位符
    sed -i "s/__PLUGIN_NAME__/$PLUGIN_NAME/g" "$PACKAGE_DIR/install.sh"
    sed -i "s/__PLUGIN_VERSION__/$PLUGIN_VERSION/g" "$PACKAGE_DIR/install.sh"
    sed -i "s/__PLUGIN_CATEGORY__/$PLUGIN_CATEGORY/g" "$PACKAGE_DIR/install.sh"
    
    chmod +x "$PACKAGE_DIR/install.sh"
}

# 创建tar.gz包
create_tar_package() {
    local package_file="$OUTPUT_DIR/$PACKAGE_NAME-$PLUGIN_VERSION.tar.gz"
    
    log_info "创建tar.gz包: $(basename "$package_file")"
    
    if [[ $DRY_RUN == true ]]; then
        log_info "[DRY-RUN] 将创建: $package_file"
        return 0
    fi
    
    cd "$TEMP_DIR"
    tar -czf "$package_file" "$PACKAGE_NAME-$PLUGIN_VERSION"
    
    log_success "tar.gz包创建完成: $package_file"
    CREATED_PACKAGES+=("$package_file")
}

# 创建ZIP包
create_zip_package() {
    local package_file="$OUTPUT_DIR/$PACKAGE_NAME-$PLUGIN_VERSION.zip"
    
    log_info "创建ZIP包: $(basename "$package_file")"
    
    if [[ $DRY_RUN == true ]]; then
        log_info "[DRY-RUN] 将创建: $package_file"
        return 0
    fi
    
    cd "$TEMP_DIR"
    zip -r "$package_file" "$PACKAGE_NAME-$PLUGIN_VERSION" >/dev/null
    
    log_success "ZIP包创建完成: $package_file"
    CREATED_PACKAGES+=("$package_file")
}

# 创建插件包格式
create_plugin_package() {
    local package_file="$OUTPUT_DIR/$PACKAGE_NAME-$PLUGIN_VERSION.plugin"
    
    log_info "创建插件包: $(basename "$package_file")"
    
    if [[ $DRY_RUN == true ]]; then
        log_info "[DRY-RUN] 将创建: $package_file"
        return 0
    fi
    
    # 插件包格式是特殊的tar.gz，包含额外的头信息
    cd "$TEMP_DIR"
    
    # 创建插件包头
    local header_file="$TEMP_DIR/plugin_header"
    cat > "$header_file" << EOF
#!/bin/bash
# go-musicfox v2 插件包
# 这是一个自解压的插件包

PLUGIN_MAGIC="GOMUSICFOX_PLUGIN_V2"
PLUGIN_NAME="$PLUGIN_NAME"
PLUGIN_VERSION="$PLUGIN_VERSION"
PLUGIN_CATEGORY="$PLUGIN_CATEGORY"

if [ "\$1" = "--extract" ]; then
    # 提取插件内容
    tail -n +__ARCHIVE_LINE__ "\$0" | tar -xzf -
    exit 0
fi

if [ "\$1" = "--install" ]; then
    # 安装插件
    TEMP_DIR=\$(mktemp -d)
    tail -n +__ARCHIVE_LINE__ "\$0" | tar -xzf - -C "\$TEMP_DIR"
    cd "\$TEMP_DIR/$PACKAGE_NAME-$PLUGIN_VERSION"
    ./install.sh
    rm -rf "\$TEMP_DIR"
    exit 0
fi

echo "go-musicfox v2 插件包: \$PLUGIN_NAME v\$PLUGIN_VERSION"
echo "使用方法:"
echo "  \$0 --extract    # 提取插件内容"
echo "  \$0 --install    # 安装插件"
exit 0

__ARCHIVE_BELOW__
EOF
    
    # 创建tar.gz内容
    local archive_file="$TEMP_DIR/plugin_archive.tar.gz"
    tar -czf "$archive_file" "$PACKAGE_NAME-$PLUGIN_VERSION"
    
    # 合并头和内容
    local archive_line=$(wc -l < "$header_file")
    archive_line=$((archive_line + 1))
    
    sed "s/__ARCHIVE_LINE__/$archive_line/g" "$header_file" > "$package_file"
    cat "$archive_file" >> "$package_file"
    
    chmod +x "$package_file"
    
    log_success "插件包创建完成: $package_file"
    CREATED_PACKAGES+=("$package_file")
}

# 生成校验和
generate_checksums() {
    if [[ $GENERATE_CHECKSUM == false || ${#CREATED_PACKAGES[@]} -eq 0 ]]; then
        return 0
    fi
    
    log_info "生成校验和文件..."
    
    local checksum_file="$OUTPUT_DIR/$PACKAGE_NAME-$PLUGIN_VERSION.checksums"
    
    if [[ $DRY_RUN == true ]]; then
        log_info "[DRY-RUN] 将生成校验和文件: $checksum_file"
        return 0
    fi
    
    > "$checksum_file"  # 清空文件
    
    for package_file in "${CREATED_PACKAGES[@]}"; do
        local filename=$(basename "$package_file")
        
        # SHA256
        if command -v sha256sum &> /dev/null; then
            echo "SHA256 ($filename) = $(sha256sum "$package_file" | cut -d' ' -f1)" >> "$checksum_file"
        elif command -v shasum &> /dev/null; then
            echo "SHA256 ($filename) = $(shasum -a 256 "$package_file" | cut -d' ' -f1)" >> "$checksum_file"
        fi
        
        # MD5
        if command -v md5sum &> /dev/null; then
            echo "MD5 ($filename) = $(md5sum "$package_file" | cut -d' ' -f1)" >> "$checksum_file"
        elif command -v md5 &> /dev/null; then
            echo "MD5 ($filename) = $(md5 -q "$package_file")" >> "$checksum_file"
        fi
    done
    
    log_success "校验和文件已生成: $checksum_file"
}

# 签名包文件
sign_packages() {
    if [[ $SIGN_PACKAGE == false || ${#CREATED_PACKAGES[@]} -eq 0 ]]; then
        return 0
    fi
    
    log_info "签名包文件..."
    
    if [[ $DRY_RUN == true ]]; then
        log_info "[DRY-RUN] 将签名包文件"
        return 0
    fi
    
    for package_file in "${CREATED_PACKAGES[@]}"; do
        local sig_file="${package_file}.sig"
        
        if [[ -n "$SIGN_KEY" ]]; then
            gpg --detach-sign --armor --local-user "$SIGN_KEY" --output "$sig_file" "$package_file"
        else
            gpg --detach-sign --armor --output "$sig_file" "$package_file"
        fi
        
        log_info "已签名: $(basename "$package_file")"
    done
    
    log_success "包文件签名完成"
}

# 显示打包摘要
show_packaging_summary() {
    log_success "插件打包完成！"
    echo
    
    log_info "插件信息:"
    log_info "  名称: $PLUGIN_NAME"
    log_info "  版本: $PLUGIN_VERSION"
    log_info "  类型: $PLUGIN_CATEGORY"
    log_info "  格式: $PACKAGE_FORMAT"
    
    if [[ $DRY_RUN == false ]]; then
        echo
        log_info "生成的包文件:"
        for package_file in "${CREATED_PACKAGES[@]}"; do
            local file_size=$(stat -c%s "$package_file" 2>/dev/null || stat -f%z "$package_file" 2>/dev/null)
            local human_size=$(numfmt --to=iec-i --suffix=B "$file_size" 2>/dev/null || echo "${file_size}B")
            log_info "  $(basename "$package_file") ($human_size)"
        done
        
        if [[ $GENERATE_CHECKSUM == true ]]; then
            log_info "  $PACKAGE_NAME-$PLUGIN_VERSION.checksums"
        fi
        
        if [[ $SIGN_PACKAGE == true ]]; then
            for package_file in "${CREATED_PACKAGES[@]}"; do
                log_info "  $(basename "$package_file").sig"
            done
        fi
        
        echo
        log_info "输出目录: $OUTPUT_DIR"
    fi
    
    echo
    log_info "安装方法:"
    case $PACKAGE_FORMAT in
        tar.gz)
            log_info "  tar -xzf $PACKAGE_NAME-$PLUGIN_VERSION.tar.gz"
            log_info "  cd $PACKAGE_NAME-$PLUGIN_VERSION"
            log_info "  ./install.sh"
            ;;
        zip)
            log_info "  unzip $PACKAGE_NAME-$PLUGIN_VERSION.zip"
            log_info "  cd $PACKAGE_NAME-$PLUGIN_VERSION"
            log_info "  ./install.sh"
            ;;
        plugin)
            log_info "  ./$PACKAGE_NAME-$PLUGIN_VERSION.plugin --install"
            ;;
    esac
}

# 主函数
main() {
    log_info "go-musicfox v2 插件打包器"
    
    parse_args "$@"
    init_packaging
    analyze_plugin
    prepare_package_content
    
    # 初始化包列表
    CREATED_PACKAGES=()
    
    # 根据格式创建包
    case $PACKAGE_FORMAT in
        tar.gz)
            create_tar_package
            ;;
        zip)
            create_zip_package
            ;;
        plugin)
            create_plugin_package
            ;;
        all)
            create_tar_package
            create_zip_package
            create_plugin_package
            ;;
    esac
    
    # 后处理
    generate_checksums
    sign_packages
    
    show_packaging_summary
}

# 执行主函数
main "$@"