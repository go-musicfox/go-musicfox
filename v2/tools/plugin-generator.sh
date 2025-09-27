#!/bin/bash
# go-musicfox v2 插件生成器
# 快速生成各种类型插件的模板代码

set -euo pipefail

# 脚本配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TEMPLATES_DIR="$SCRIPT_DIR/templates"
PLUGINS_DIR="$PROJECT_ROOT/plugins"

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
go-musicfox v2 插件生成器

用法: $0 [选项] <插件名称>

选项:
  -h, --help              显示帮助信息
  -t, --type TYPE         插件类型 (shared|rpc|wasm|hotload)
  -l, --language LANG     编程语言 (go|rust|c|cpp)
  -c, --category CAT      插件分类 (audio|ui|network|storage|misc)
  -a, --author AUTHOR     作者名称
  -e, --email EMAIL       作者邮箱
  -d, --description DESC  插件描述
  -v, --version VERSION   插件版本 (默认: 0.1.0)
  -o, --output DIR        输出目录 (默认: plugins/)
  --license LICENSE       许可证 (MIT|Apache-2.0|GPL-3.0|BSD-3-Clause)
  --interactive           交互式模式
  --overwrite             覆盖已存在的插件

插件类型:
  shared                  动态链接库插件 (默认)
  rpc                     RPC插件
  wasm                    WebAssembly插件
  hotload                 热加载插件

编程语言:
  go                      Go语言 (默认)
  rust                    Rust语言 (仅支持wasm类型)
  c                       C语言 (仅支持shared类型)
  cpp                     C++语言 (仅支持shared类型)

插件分类:
  audio                   音频处理插件
  ui                      用户界面插件
  network                 网络服务插件
  storage                 存储插件
  misc                    其他插件 (默认)

示例:
  $0 my-plugin                                    # 生成基础Go插件
  $0 -t rpc -c audio audio-processor             # 生成音频处理RPC插件
  $0 -t wasm -l rust wasm-filter                 # 生成Rust WebAssembly插件
  $0 --interactive                               # 交互式生成插件

EOF
}

# 解析命令行参数
parse_args() {
    PLUGIN_TYPE="shared"
    LANGUAGE="go"
    CATEGORY="misc"
    AUTHOR=""
    EMAIL=""
    DESCRIPTION=""
    VERSION="0.1.0"
    OUTPUT_DIR="$PLUGINS_DIR"
    LICENSE="MIT"
    INTERACTIVE=false
    OVERWRITE=false
    PLUGIN_NAME=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help) show_help; exit 0 ;;
            -t|--type) PLUGIN_TYPE="$2"; shift 2 ;;
            -l|--language) LANGUAGE="$2"; shift 2 ;;
            -c|--category) CATEGORY="$2"; shift 2 ;;
            -a|--author) AUTHOR="$2"; shift 2 ;;
            -e|--email) EMAIL="$2"; shift 2 ;;
            -d|--description) DESCRIPTION="$2"; shift 2 ;;
            -v|--version) VERSION="$2"; shift 2 ;;
            -o|--output) OUTPUT_DIR="$2"; shift 2 ;;
            --license) LICENSE="$2"; shift 2 ;;
            --interactive) INTERACTIVE=true; shift ;;
            --overwrite) OVERWRITE=true; shift ;;
            -*) error_exit "未知选项: $1" ;;
            *) PLUGIN_NAME="$1"; shift ;;
        esac
    done
    
    # 验证插件类型
    case $PLUGIN_TYPE in
        shared|rpc|wasm|hotload) ;;
        *) error_exit "无效的插件类型: $PLUGIN_TYPE" ;;
    esac
    
    # 验证编程语言
    case $LANGUAGE in
        go|rust|c|cpp) ;;
        *) error_exit "无效的编程语言: $LANGUAGE" ;;
    esac
    
    # 验证语言和类型的兼容性
    case $PLUGIN_TYPE in
        shared)
            case $LANGUAGE in
                go|c|cpp) ;;
                *) error_exit "$LANGUAGE 不支持 $PLUGIN_TYPE 类型插件" ;;
            esac
            ;;
        rpc|hotload)
            case $LANGUAGE in
                go) ;;
                *) error_exit "$LANGUAGE 不支持 $PLUGIN_TYPE 类型插件" ;;
            esac
            ;;
        wasm)
            case $LANGUAGE in
                go|rust) ;;
                *) error_exit "$LANGUAGE 不支持 $PLUGIN_TYPE 类型插件" ;;
            esac
            ;;
    esac
    
    # 验证插件分类
    case $CATEGORY in
        audio|ui|network|storage|misc) ;;
        *) error_exit "无效的插件分类: $CATEGORY" ;;
    esac
    
    # 验证许可证
    case $LICENSE in
        MIT|Apache-2.0|GPL-3.0|BSD-3-Clause) ;;
        *) error_exit "无效的许可证: $LICENSE" ;;
    esac
}

# 交互式输入
interactive_input() {
    echo "=== go-musicfox v2 插件生成器 ==="
    echo
    
    # 插件名称
    if [[ -z "$PLUGIN_NAME" ]]; then
        echo -n "插件名称: "
        read -r PLUGIN_NAME
        if [[ -z "$PLUGIN_NAME" ]]; then
            error_exit "插件名称不能为空"
        fi
    fi
    
    # 插件类型
    echo "插件类型:"
    echo "  1) shared - 动态链接库插件"
    echo "  2) rpc - RPC插件"
    echo "  3) wasm - WebAssembly插件"
    echo "  4) hotload - 热加载插件"
    echo -n "选择插件类型 [1-4] (默认: 1): "
    read -r type_choice
    case $type_choice in
        1|"") PLUGIN_TYPE="shared" ;;
        2) PLUGIN_TYPE="rpc" ;;
        3) PLUGIN_TYPE="wasm" ;;
        4) PLUGIN_TYPE="hotload" ;;
        *) error_exit "无效选择: $type_choice" ;;
    esac
    
    # 编程语言
    echo "编程语言:"
    case $PLUGIN_TYPE in
        shared)
            echo "  1) go - Go语言"
            echo "  2) c - C语言"
            echo "  3) cpp - C++语言"
            echo -n "选择编程语言 [1-3] (默认: 1): "
            read -r lang_choice
            case $lang_choice in
                1|"") LANGUAGE="go" ;;
                2) LANGUAGE="c" ;;
                3) LANGUAGE="cpp" ;;
                *) error_exit "无效选择: $lang_choice" ;;
            esac
            ;;
        rpc|hotload)
            LANGUAGE="go"
            echo "  自动选择: Go语言"
            ;;
        wasm)
            echo "  1) go - Go语言"
            echo "  2) rust - Rust语言"
            echo -n "选择编程语言 [1-2] (默认: 1): "
            read -r lang_choice
            case $lang_choice in
                1|"") LANGUAGE="go" ;;
                2) LANGUAGE="rust" ;;
                *) error_exit "无效选择: $lang_choice" ;;
            esac
            ;;
    esac
    
    # 插件分类
    echo "插件分类:"
    echo "  1) audio - 音频处理插件"
    echo "  2) ui - 用户界面插件"
    echo "  3) network - 网络服务插件"
    echo "  4) storage - 存储插件"
    echo "  5) misc - 其他插件"
    echo -n "选择插件分类 [1-5] (默认: 5): "
    read -r cat_choice
    case $cat_choice in
        1) CATEGORY="audio" ;;
        2) CATEGORY="ui" ;;
        3) CATEGORY="network" ;;
        4) CATEGORY="storage" ;;
        5|"") CATEGORY="misc" ;;
        *) error_exit "无效选择: $cat_choice" ;;
    esac
    
    # 插件描述
    if [[ -z "$DESCRIPTION" ]]; then
        echo -n "插件描述: "
        read -r DESCRIPTION
    fi
    
    # 作者信息
    if [[ -z "$AUTHOR" ]]; then
        echo -n "作者名称: "
        read -r AUTHOR
    fi
    
    if [[ -z "$EMAIL" ]]; then
        echo -n "作者邮箱: "
        read -r EMAIL
    fi
    
    # 版本号
    echo -n "插件版本 (默认: $VERSION): "
    read -r version_input
    if [[ -n "$version_input" ]]; then
        VERSION="$version_input"
    fi
    
    # 许可证
    echo "许可证:"
    echo "  1) MIT"
    echo "  2) Apache-2.0"
    echo "  3) GPL-3.0"
    echo "  4) BSD-3-Clause"
    echo -n "选择许可证 [1-4] (默认: 1): "
    read -r license_choice
    case $license_choice in
        1|"") LICENSE="MIT" ;;
        2) LICENSE="Apache-2.0" ;;
        3) LICENSE="GPL-3.0" ;;
        4) LICENSE="BSD-3-Clause" ;;
        *) error_exit "无效选择: $license_choice" ;;
    esac
}

# 验证插件名称
validate_plugin_name() {
    if [[ -z "$PLUGIN_NAME" ]]; then
        if [[ $INTERACTIVE == false ]]; then
            error_exit "插件名称不能为空"
        fi
    fi
    
    # 检查名称格式
    if [[ ! "$PLUGIN_NAME" =~ ^[a-z][a-z0-9-]*[a-z0-9]$ ]]; then
        error_exit "插件名称格式无效，应该使用小写字母、数字和连字符，以字母开头和结尾"
    fi
    
    # 检查是否已存在
    PLUGIN_DIR="$OUTPUT_DIR/$PLUGIN_TYPE/$PLUGIN_NAME"
    if [[ -d "$PLUGIN_DIR" ]]; then
        if [[ $OVERWRITE == false ]]; then
            error_exit "插件目录已存在: $PLUGIN_DIR，使用 --overwrite 强制覆盖"
        else
            log_warning "将覆盖已存在的插件目录: $PLUGIN_DIR"
        fi
    fi
}

# 创建插件目录结构
create_plugin_structure() {
    log_info "创建插件目录结构..."
    
    # 创建主目录
    mkdir -p "$PLUGIN_DIR"
    
    # 根据插件类型和语言创建子目录
    case $LANGUAGE in
        go)
            mkdir -p "$PLUGIN_DIR"/{cmd,internal,pkg,test,docs,examples}
            ;;
        rust)
            mkdir -p "$PLUGIN_DIR"/{src,tests,examples,docs}
            ;;
        c|cpp)
            mkdir -p "$PLUGIN_DIR"/{src,include,test,docs,examples}
            ;;
    esac
    
    log_success "插件目录结构已创建: $PLUGIN_DIR"
}

# 生成Go插件代码
generate_go_plugin() {
    log_info "生成Go插件代码..."
    
    # 生成go.mod
    cat > "$PLUGIN_DIR/go.mod" << EOF
module github.com/go-musicfox/go-musicfox/v2/plugins/$PLUGIN_TYPE/$PLUGIN_NAME

go 1.21

require (
    github.com/go-musicfox/go-musicfox/v2 v2.0.0
)
EOF

    # 生成主插件文件
    case $PLUGIN_TYPE in
        shared)
            generate_go_shared_plugin
            ;;
        rpc)
            generate_go_rpc_plugin
            ;;
        wasm)
            generate_go_wasm_plugin
            ;;
        hotload)
            generate_go_hotload_plugin
            ;;
    esac
    
    # 生成测试文件
    cat > "$PLUGIN_DIR/test/${PLUGIN_NAME}_test.go" << EOF
package main

import (
    "testing"
)

func TestPlugin(t *testing.T) {
    // TODO: 添加插件测试
    t.Log("插件测试")
}

func BenchmarkPlugin(b *testing.B) {
    // TODO: 添加性能测试
    for i := 0; i < b.N; i++ {
        // 性能测试代码
    }
}
EOF
}

# 生成Go动态链接库插件
generate_go_shared_plugin() {
    cat > "$PLUGIN_DIR/main.go" << EOF
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/go-musicfox/go-musicfox/v2/pkg/plugin"
)

// Plugin 插件实现
type Plugin struct {
    name        string
    version     string
    description string
}

// NewPlugin 创建插件实例
func NewPlugin() plugin.Plugin {
    return &Plugin{
        name:        "$PLUGIN_NAME",
        version:     "$VERSION",
        description: "$DESCRIPTION",
    }
}

// Name 返回插件名称
func (p *Plugin) Name() string {
    return p.name
}

// Version 返回插件版本
func (p *Plugin) Version() string {
    return p.version
}

// Description 返回插件描述
func (p *Plugin) Description() string {
    return p.description
}

// Category 返回插件分类
func (p *Plugin) Category() string {
    return "$CATEGORY"
}

// Initialize 初始化插件
func (p *Plugin) Initialize(ctx context.Context, config map[string]interface{}) error {
    log.Printf("初始化插件: %s v%s", p.name, p.version)
    
    // TODO: 添加初始化逻辑
    
    return nil
}

// Execute 执行插件功能
func (p *Plugin) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    log.Printf("执行插件: %s", p.name)
    
    // TODO: 添加插件执行逻辑
    
    return fmt.Sprintf("Hello from %s plugin!", p.name), nil
}

// Cleanup 清理插件资源
func (p *Plugin) Cleanup() error {
    log.Printf("清理插件: %s", p.name)
    
    // TODO: 添加清理逻辑
    
    return nil
}

// 导出符号（动态链接库需要）
var PluginInstance = NewPlugin()
EOF
}

# 生成Go RPC插件
generate_go_rpc_plugin() {
    cat > "$PLUGIN_DIR/main.go" << EOF
package main

import (
    "context"
    "log"
    "net"
    "net/rpc"
    
    "github.com/go-musicfox/go-musicfox/v2/pkg/plugin"
)

// RPCPlugin RPC插件实现
type RPCPlugin struct {
    name        string
    version     string
    description string
    listener    net.Listener
}

// PluginArgs RPC调用参数
type PluginArgs struct {
    Input interface{} \`json:"input"\`
}

// PluginReply RPC调用返回
type PluginReply struct {
    Output interface{} \`json:"output"\`
    Error  string      \`json:"error,omitempty"\`
}

// NewRPCPlugin 创建RPC插件实例
func NewRPCPlugin() *RPCPlugin {
    return &RPCPlugin{
        name:        "$PLUGIN_NAME",
        version:     "$VERSION",
        description: "$DESCRIPTION",
    }
}

// Name 返回插件名称
func (p *RPCPlugin) Name() string {
    return p.name
}

// Version 返回插件版本
func (p *RPCPlugin) Version() string {
    return p.version
}

// Description 返回插件描述
func (p *RPCPlugin) Description() string {
    return p.description
}

// Category 返回插件分类
func (p *RPCPlugin) Category() string {
    return "$CATEGORY"
}

// Initialize 初始化RPC插件
func (p *RPCPlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
    log.Printf("初始化RPC插件: %s v%s", p.name, p.version)
    
    // 启动RPC服务器
    rpc.Register(p)
    
    listener, err := net.Listen("tcp", ":0")
    if err != nil {
        return err
    }
    
    p.listener = listener
    
    go func() {
        for {
            conn, err := listener.Accept()
            if err != nil {
                log.Printf("RPC连接错误: %v", err)
                continue
            }
            go rpc.ServeConn(conn)
        }
    }()
    
    log.Printf("RPC服务器启动: %s", listener.Addr().String())
    
    return nil
}

// Execute RPC方法
func (p *RPCPlugin) Execute(args *PluginArgs, reply *PluginReply) error {
    log.Printf("执行RPC插件: %s", p.name)
    
    // TODO: 添加RPC插件执行逻辑
    
    reply.Output = map[string]interface{}{
        "message": "Hello from RPC plugin!",
        "plugin":  p.name,
        "version": p.version,
    }
    
    return nil
}

// Cleanup 清理RPC插件资源
func (p *RPCPlugin) Cleanup() error {
    log.Printf("清理RPC插件: %s", p.name)
    
    if p.listener != nil {
        return p.listener.Close()
    }
    
    return nil
}

func main() {
    plugin := NewRPCPlugin()
    
    ctx := context.Background()
    config := make(map[string]interface{})
    
    if err := plugin.Initialize(ctx, config); err != nil {
        log.Fatalf("插件初始化失败: %v", err)
    }
    
    // 保持运行
    select {}
}
EOF
}

# 生成Go WebAssembly插件
generate_go_wasm_plugin() {
    cat > "$PLUGIN_DIR/main.go" << EOF
//go:build wasm
// +build wasm

package main

import (
    "context"
    "encoding/json"
    "fmt"
    "syscall/js"
    
    "github.com/go-musicfox/go-musicfox/v2/pkg/plugin"
)

// WASMPlugin WebAssembly插件实现
type WASMPlugin struct {
    name        string
    version     string
    description string
}

// NewWASMPlugin 创建WASM插件实例
func NewWASMPlugin() *WASMPlugin {
    return &WASMPlugin{
        name:        "$PLUGIN_NAME",
        version:     "$VERSION",
        description: "$DESCRIPTION",
    }
}

// Name 返回插件名称
func (p *WASMPlugin) Name() string {
    return p.name
}

// Version 返回插件版本
func (p *WASMPlugin) Version() string {
    return p.version
}

// Description 返回插件描述
func (p *WASMPlugin) Description() string {
    return p.description
}

// Category 返回插件分类
func (p *WASMPlugin) Category() string {
    return "$CATEGORY"
}

// Initialize 初始化WASM插件
func (p *WASMPlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
    fmt.Printf("初始化WASM插件: %s v%s\n", p.name, p.version)
    
    // TODO: 添加初始化逻辑
    
    return nil
}

// Execute 执行WASM插件功能
func (p *WASMPlugin) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    fmt.Printf("执行WASM插件: %s\n", p.name)
    
    // TODO: 添加WASM插件执行逻辑
    
    return map[string]interface{}{
        "message": "Hello from WASM plugin!",
        "plugin":  p.name,
        "version": p.version,
    }, nil
}

// Cleanup 清理WASM插件资源
func (p *WASMPlugin) Cleanup() error {
    fmt.Printf("清理WASM插件: %s\n", p.name)
    
    // TODO: 添加清理逻辑
    
    return nil
}

// JavaScript接口函数
func jsInitialize(this js.Value, args []js.Value) interface{} {
    plugin := NewWASMPlugin()
    
    configJSON := args[0].String()
    var config map[string]interface{}
    json.Unmarshal([]byte(configJSON), &config)
    
    err := plugin.Initialize(context.Background(), config)
    if err != nil {
        return map[string]interface{}{
            "error": err.Error(),
        }
    }
    
    return map[string]interface{}{
        "success": true,
    }
}

func jsExecute(this js.Value, args []js.Value) interface{} {
    plugin := NewWASMPlugin()
    
    inputJSON := args[0].String()
    var input interface{}
    json.Unmarshal([]byte(inputJSON), &input)
    
    output, err := plugin.Execute(context.Background(), input)
    if err != nil {
        return map[string]interface{}{
            "error": err.Error(),
        }
    }
    
    return output
}

func jsCleanup(this js.Value, args []js.Value) interface{} {
    plugin := NewWASMPlugin()
    
    err := plugin.Cleanup()
    if err != nil {
        return map[string]interface{}{
            "error": err.Error(),
        }
    }
    
    return map[string]interface{}{
        "success": true,
    }
}

func main() {
    // 注册JavaScript函数
    js.Global().Set("pluginInitialize", js.FuncOf(jsInitialize))
    js.Global().Set("pluginExecute", js.FuncOf(jsExecute))
    js.Global().Set("pluginCleanup", js.FuncOf(jsCleanup))
    
    // 保持WASM模块运行
    select {}
}
EOF
}

# 生成Go热加载插件
generate_go_hotload_plugin() {
    cat > "$PLUGIN_DIR/main.go" << EOF
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"
    
    "github.com/go-musicfox/go-musicfox/v2/pkg/plugin"
)

// HotloadPlugin 热加载插件实现
type HotloadPlugin struct {
    name        string
    version     string
    description string
    mu          sync.RWMutex
    state       map[string]interface{}
    stopCh      chan struct{}
}

// NewHotloadPlugin 创建热加载插件实例
func NewHotloadPlugin() *HotloadPlugin {
    return &HotloadPlugin{
        name:        "$PLUGIN_NAME",
        version:     "$VERSION",
        description: "$DESCRIPTION",
        state:       make(map[string]interface{}),
        stopCh:      make(chan struct{}),
    }
}

// Name 返回插件名称
func (p *HotloadPlugin) Name() string {
    return p.name
}

// Version 返回插件版本
func (p *HotloadPlugin) Version() string {
    return p.version
}

// Description 返回插件描述
func (p *HotloadPlugin) Description() string {
    return p.description
}

// Category 返回插件分类
func (p *HotloadPlugin) Category() string {
    return "$CATEGORY"
}

// Initialize 初始化热加载插件
func (p *HotloadPlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
    log.Printf("初始化热加载插件: %s v%s", p.name, p.version)
    
    p.mu.Lock()
    defer p.mu.Unlock()
    
    // 初始化状态
    p.state["initialized_at"] = time.Now()
    p.state["config"] = config
    
    // 启动后台任务
    go p.backgroundTask()
    
    // TODO: 添加初始化逻辑
    
    return nil
}

// Execute 执行热加载插件功能
func (p *HotloadPlugin) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    log.Printf("执行热加载插件: %s", p.name)
    
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    // TODO: 添加热加载插件执行逻辑
    
    return map[string]interface{}{
        "message":     "Hello from hotload plugin!",
        "plugin":      p.name,
        "version":     p.version,
        "state":       p.state,
        "executed_at": time.Now(),
    }, nil
}

// Reload 重新加载插件
func (p *HotloadPlugin) Reload(ctx context.Context, config map[string]interface{}) error {
    log.Printf("重新加载插件: %s", p.name)
    
    p.mu.Lock()
    defer p.mu.Unlock()
    
    // 保存旧状态
    oldState := make(map[string]interface{})
    for k, v := range p.state {
        oldState[k] = v
    }
    
    // 更新配置
    p.state["config"] = config
    p.state["reloaded_at"] = time.Now()
    p.state["previous_state"] = oldState
    
    // TODO: 添加重新加载逻辑
    
    return nil
}

// GetState 获取插件状态
func (p *HotloadPlugin) GetState() map[string]interface{} {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    state := make(map[string]interface{})
    for k, v := range p.state {
        state[k] = v
    }
    
    return state
}

// SetState 设置插件状态
func (p *HotloadPlugin) SetState(key string, value interface{}) {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    p.state[key] = value
}

// backgroundTask 后台任务
func (p *HotloadPlugin) backgroundTask() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            p.SetState("last_heartbeat", time.Now())
            log.Printf("插件心跳: %s", p.name)
        case <-p.stopCh:
            log.Printf("停止后台任务: %s", p.name)
            return
        }
    }
}

// Cleanup 清理热加载插件资源
func (p *HotloadPlugin) Cleanup() error {
    log.Printf("清理热加载插件: %s", p.name)
    
    close(p.stopCh)
    
    p.mu.Lock()
    defer p.mu.Unlock()
    
    p.state["cleaned_up_at"] = time.Now()
    
    // TODO: 添加清理逻辑
    
    return nil
}

// 导出符号
var PluginInstance = NewHotloadPlugin()
EOF
}

# 生成Rust WebAssembly插件
generate_rust_wasm_plugin() {
    log_info "生成Rust WebAssembly插件代码..."
    
    # 生成Cargo.toml
    cat > "$PLUGIN_DIR/Cargo.toml" << EOF
[package]
name = "$PLUGIN_NAME"
version = "$VERSION"
authors = ["$AUTHOR <$EMAIL>"]
edition = "2021"
description = "$DESCRIPTION"
license = "$LICENSE"

[lib]
crate-type = ["cdylib"]

[dependencies]
wasm-bindgen = "0.2"
web-sys = "0.3"
js-sys = "0.3"
serde = { version = "1.0", features = ["derive"] }
serde-wasm-bindgen = "0.4"
console_error_panic_hook = "0.1"
wee_alloc = "0.4"

[dependencies.web-sys]
version = "0.3"
features = [
  "console",
  "Window",
  "Document",
]
EOF
    
    # 生成src/lib.rs
    cat > "$PLUGIN_DIR/src/lib.rs" << EOF
use wasm_bindgen::prelude::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

// 当 \`wee_alloc\` 特性启用时，使用 \`wee_alloc\` 作为全局分配器
#[cfg(feature = "wee_alloc")]
#[global_allocator]
static ALLOC: wee_alloc::WeeAlloc = wee_alloc::WeeAlloc::INIT;

// 插件配置结构
#[derive(Serialize, Deserialize)]
pub struct PluginConfig {
    pub enabled: bool,
    pub settings: HashMap<String, String>,
}

// 插件输入结构
#[derive(Serialize, Deserialize)]
pub struct PluginInput {
    pub data: String,
    pub metadata: HashMap<String, String>,
}

// 插件输出结构
#[derive(Serialize, Deserialize)]
pub struct PluginOutput {
    pub result: String,
    pub status: String,
    pub metadata: HashMap<String, String>,
}

// 插件错误结构
#[derive(Serialize, Deserialize)]
pub struct PluginError {
    pub message: String,
    pub code: u32,
}

// 插件主结构
pub struct Plugin {
    name: String,
    version: String,
    description: String,
    initialized: bool,
}

impl Plugin {
    pub fn new() -> Self {
        Self {
            name: "$PLUGIN_NAME".to_string(),
            version: "$VERSION".to_string(),
            description: "$DESCRIPTION".to_string(),
            initialized: false,
        }
    }
    
    pub fn initialize(&mut self, config: PluginConfig) -> Result<(), PluginError> {
        console_error_panic_hook::set_once();
        
        web_sys::console::log_1(&format!("初始化插件: {} v{}", self.name, self.version).into());
        
        if !config.enabled {
            return Err(PluginError {
                message: "插件未启用".to_string(),
                code: 1,
            });
        }
        
        // TODO: 添加初始化逻辑
        
        self.initialized = true;
        Ok(())
    }
    
    pub fn execute(&self, input: PluginInput) -> Result<PluginOutput, PluginError> {
        if !self.initialized {
            return Err(PluginError {
                message: "插件未初始化".to_string(),
                code: 2,
            });
        }
        
        web_sys::console::log_1(&format!("执行插件: {}", self.name).into());
        
        // TODO: 添加插件执行逻辑
        
        let mut metadata = HashMap::new();
        metadata.insert("plugin".to_string(), self.name.clone());
        metadata.insert("version".to_string(), self.version.clone());
        
        Ok(PluginOutput {
            result: format!("Hello from {} WASM plugin!", self.name),
            status: "success".to_string(),
            metadata,
        })
    }
    
    pub fn cleanup(&mut self) -> Result<(), PluginError> {
        web_sys::console::log_1(&format!("清理插件: {}", self.name).into());
        
        // TODO: 添加清理逻辑
        
        self.initialized = false;
        Ok(())
    }
}

// 全局插件实例
static mut PLUGIN_INSTANCE: Option<Plugin> = None;

// WebAssembly导出函数
#[wasm_bindgen]
pub fn plugin_initialize(config_json: &str) -> Result<(), JsValue> {
    let config: PluginConfig = serde_wasm_bindgen::from_value(
        js_sys::JSON::parse(config_json)?
    )?;
    
    unsafe {
        if PLUGIN_INSTANCE.is_none() {
            PLUGIN_INSTANCE = Some(Plugin::new());
        }
        
        if let Some(ref mut plugin) = PLUGIN_INSTANCE {
            plugin.initialize(config).map_err(|e| {
                JsValue::from_str(&format!("初始化失败: {}", e.message))
            })?;
        }
    }
    
    Ok(())
}

#[wasm_bindgen]
pub fn plugin_execute(input_json: &str) -> Result<JsValue, JsValue> {
    let input: PluginInput = serde_wasm_bindgen::from_value(
        js_sys::JSON::parse(input_json)?
    )?;
    
    unsafe {
        if let Some(ref plugin) = PLUGIN_INSTANCE {
            let output = plugin.execute(input).map_err(|e| {
                JsValue::from_str(&format!("执行失败: {}", e.message))
            })?;
            
            Ok(serde_wasm_bindgen::to_value(&output)?)
        } else {
            Err(JsValue::from_str("插件未初始化"))
        }
    }
}

#[wasm_bindgen]
pub fn plugin_cleanup() -> Result<(), JsValue> {
    unsafe {
        if let Some(ref mut plugin) = PLUGIN_INSTANCE {
            plugin.cleanup().map_err(|e| {
                JsValue::from_str(&format!("清理失败: {}", e.message))
            })?;
        }
    }
    
    Ok(())
}

#[wasm_bindgen]
pub fn plugin_info() -> JsValue {
    let info = serde_json::json!({
        "name": "$PLUGIN_NAME",
        "version": "$VERSION",
        "description": "$DESCRIPTION",
        "category": "$CATEGORY",
        "type": "wasm",
        "language": "rust"
    });
    
    JsValue::from_str(&info.to_string())
}
EOF
}

# 生成C动态链接库插件
generate_c_shared_plugin() {
    log_info "生成C动态链接库插件代码..."
    
    # 生成Makefile
    cat > "$PLUGIN_DIR/Makefile" << EOF
CC = gcc
CFLAGS = -Wall -Wextra -fPIC -O2
LDFLAGS = -shared
TARGET = lib$PLUGIN_NAME.so
SOURCES = src/plugin.c
OBJECTS = \$(SOURCES:.c=.o)

.PHONY: all clean install

all: \$(TARGET)

\$(TARGET): \$(OBJECTS)
	\$(CC) \$(LDFLAGS) -o \$@ \$^

%.o: %.c
	\$(CC) \$(CFLAGS) -c \$< -o \$@

clean:
	rm -f \$(OBJECTS) \$(TARGET)

install: \$(TARGET)
	cp \$(TARGET) /usr/local/lib/go-musicfox/plugins/

test:
	gcc -o test_plugin test/test_plugin.c -L. -l$PLUGIN_NAME
	./test_plugin
EOF
    
    # 生成头文件
    cat > "$PLUGIN_DIR/include/plugin.h" << EOF
#ifndef ${PLUGIN_NAME^^}_PLUGIN_H
#define ${PLUGIN_NAME^^}_PLUGIN_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <stddef.h>

// 插件信息结构
typedef struct {
    const char* name;
    const char* version;
    const char* description;
    const char* category;
} plugin_info_t;

// 插件接口函数
int plugin_initialize(const char* config);
int plugin_execute(const char* input, char** output, size_t* output_size);
int plugin_cleanup(void);
const plugin_info_t* plugin_get_info(void);

#ifdef __cplusplus
}
#endif

#endif // ${PLUGIN_NAME^^}_PLUGIN_H
EOF
    
    # 生成源文件
    cat > "$PLUGIN_DIR/src/plugin.c" << EOF
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "../include/plugin.h"

// 插件信息
static const plugin_info_t plugin_info = {
    .name = "$PLUGIN_NAME",
    .version = "$VERSION",
    .description = "$DESCRIPTION",
    .category = "$CATEGORY"
};

// 插件状态
static int initialized = 0;

// 初始化插件
int plugin_initialize(const char* config) {
    printf("初始化插件: %s v%s\n", plugin_info.name, plugin_info.version);
    
    if (config) {
        printf("配置: %s\n", config);
    }
    
    // TODO: 添加初始化逻辑
    
    initialized = 1;
    return 0;
}

// 执行插件功能
int plugin_execute(const char* input, char** output, size_t* output_size) {
    if (!initialized) {
        fprintf(stderr, "插件未初始化\n");
        return -1;
    }
    
    printf("执行插件: %s\n", plugin_info.name);
    
    if (input) {
        printf("输入: %s\n", input);
    }
    
    // TODO: 添加插件执行逻辑
    
    // 生成输出
    const char* result = "Hello from C plugin!";
    *output_size = strlen(result) + 1;
    *output = malloc(*output_size);
    
    if (*output == NULL) {
        return -1;
    }
    
    strcpy(*output, result);
    
    return 0;
}

// 清理插件资源
int plugin_cleanup(void) {
    printf("清理插件: %s\n", plugin_info.name);
    
    // TODO: 添加清理逻辑
    
    initialized = 0;
    return 0;
}

// 获取插件信息
const plugin_info_t* plugin_get_info(void) {
    return &plugin_info;
}
EOF
    
    # 生成测试文件
    cat > "$PLUGIN_DIR/test/test_plugin.c" << EOF
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "../include/plugin.h"

int main() {
    printf("测试插件: $PLUGIN_NAME\n");
    
    // 获取插件信息
    const plugin_info_t* info = plugin_get_info();
    printf("插件信息:\n");
    printf("  名称: %s\n", info->name);
    printf("  版本: %s\n", info->version);
    printf("  描述: %s\n", info->description);
    printf("  分类: %s\n", info->category);
    
    // 初始化插件
    if (plugin_initialize("{\"enabled\": true}") != 0) {
        fprintf(stderr, "插件初始化失败\n");
        return 1;
    }
    
    // 执行插件
    char* output = NULL;
    size_t output_size = 0;
    
    if (plugin_execute("test input", &output, &output_size) != 0) {
        fprintf(stderr, "插件执行失败\n");
        plugin_cleanup();
        return 1;
    }
    
    printf("插件输出: %s\n", output);
    free(output);
    
    // 清理插件
    if (plugin_cleanup() != 0) {
        fprintf(stderr, "插件清理失败\n");
        return 1;
    }
    
    printf("插件测试完成\n");
    return 0;
}
EOF
}

# 生成通用文件
generate_common_files() {
    log_info "生成通用文件..."
    
    # 生成README.md
    cat > "$PLUGIN_DIR/README.md" << EOF
# $PLUGIN_NAME

$DESCRIPTION

## 插件信息

- **名称**: $PLUGIN_NAME
- **版本**: $VERSION
- **类型**: $PLUGIN_TYPE
- **语言**: $LANGUAGE
- **分类**: $CATEGORY
- **作者**: $AUTHOR
- **邮箱**: $EMAIL
- **许可证**: $LICENSE

## 功能特性

- TODO: 描述插件功能
- TODO: 列出主要特性

## 安装和使用

### 构建插件

\`\`\`bash
# 进入插件目录
cd $PLUGIN_DIR

# 构建插件
EOF
    
    case $LANGUAGE in
        go)
            cat >> "$PLUGIN_DIR/README.md" << EOF
go build -buildmode=plugin -o $PLUGIN_NAME.so .
EOF
            ;;
        rust)
            cat >> "$PLUGIN_DIR/README.md" << EOF
wasm-pack build --target web
EOF
            ;;
        c|cpp)
            cat >> "$PLUGIN_DIR/README.md" << EOF
make
EOF
            ;;
    esac
    
    cat >> "$PLUGIN_DIR/README.md" << EOF
\`\`\`

### 安装插件

\`\`\`bash
# 复制插件到插件目录
cp $PLUGIN_NAME.* /path/to/go-musicfox/plugins/$PLUGIN_TYPE/
\`\`\`

### 配置插件

在go-musicfox配置文件中添加插件配置：

\`\`\`yaml
plugins:
  $PLUGIN_TYPE:
    $PLUGIN_NAME:
      enabled: true
      # 其他配置选项
\`\`\`

## 开发

### 项目结构

\`\`\`
$PLUGIN_NAME/
EOF
    
    case $LANGUAGE in
        go)
            cat >> "$PLUGIN_DIR/README.md" << EOF
├── main.go              # 主插件文件
├── go.mod               # Go模块文件
├── test/                # 测试文件
├── docs/                # 文档
├── examples/            # 示例代码
EOF
            ;;
        rust)
            cat >> "$PLUGIN_DIR/README.md" << EOF
├── src/
│   └── lib.rs           # 主插件文件
├── Cargo.toml           # Rust项目文件
├── tests/               # 测试文件
├── docs/                # 文档
├── examples/            # 示例代码
EOF
            ;;
        c|cpp)
            cat >> "$PLUGIN_DIR/README.md" << EOF
├── src/
│   └── plugin.c         # 主插件文件
├── include/
│   └── plugin.h         # 头文件
├── Makefile             # 构建文件
├── test/                # 测试文件
├── docs/                # 文档
├── examples/            # 示例代码
EOF
            ;;
    esac
    
    cat >> "$PLUGIN_DIR/README.md" << EOF
└── README.md            # 说明文档
\`\`\`

### 测试

\`\`\`bash
EOF
    
    case $LANGUAGE in
        go)
            cat >> "$PLUGIN_DIR/README.md" << EOF
go test ./test/...
EOF
            ;;
        rust)
            cat >> "$PLUGIN_DIR/README.md" << EOF
cargo test
EOF
            ;;
        c|cpp)
            cat >> "$PLUGIN_DIR/README.md" << EOF
make test
EOF
            ;;
    esac
    
    cat >> "$PLUGIN_DIR/README.md" << EOF
\`\`\`

## 贡献

欢迎提交Issue和Pull Request！

## 许可证

本项目采用 $LICENSE 许可证。详见 [LICENSE](LICENSE) 文件。
EOF
    
    # 生成LICENSE文件
    generate_license_file
    
    # 生成.gitignore
    cat > "$PLUGIN_DIR/.gitignore" << EOF
# 构建产物
*.so
*.dll
*.dylib
*.wasm
*.o
*.obj
*.exe

# 临时文件
*.tmp
*.temp
*~

# 日志文件
*.log

# IDE文件
.vscode/
.idea/
*.swp
*.swo

# 操作系统文件
.DS_Store
Thumbs.db
EOF
    
    case $LANGUAGE in
        go)
            cat >> "$PLUGIN_DIR/.gitignore" << EOF

# Go相关
vendor/
go.sum
EOF
            ;;
        rust)
            cat >> "$PLUGIN_DIR/.gitignore" << EOF

# Rust相关
target/
Cargo.lock
pkg/
EOF
            ;;
        c|cpp)
            cat >> "$PLUGIN_DIR/.gitignore" << EOF

# C/C++相关
*.a
*.lib
EOF
            ;;
    esac
}

# 生成许可证文件
generate_license_file() {
    case $LICENSE in
        MIT)
            cat > "$PLUGIN_DIR/LICENSE" << EOF
MIT License

Copyright (c) $(date +%Y) $AUTHOR

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
EOF
            ;;
        Apache-2.0)
            cat > "$PLUGIN_DIR/LICENSE" << EOF
Apache License
Version 2.0, January 2004
http://www.apache.org/licenses/

Copyright $(date +%Y) $AUTHOR

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
EOF
            ;;
        # 其他许可证可以继续添加
    esac
}

# 显示生成摘要
show_generation_summary() {
    log_success "插件生成完成！"
    echo
    
    log_info "插件信息:"
    log_info "  名称: $PLUGIN_NAME"
    log_info "  类型: $PLUGIN_TYPE"
    log_info "  语言: $LANGUAGE"
    log_info "  分类: $CATEGORY"
    log_info "  版本: $VERSION"
    log_info "  作者: $AUTHOR"
    log_info "  许可证: $LICENSE"
    
    echo
    log_info "插件目录: $PLUGIN_DIR"
    
    echo
    log_info "下一步操作:"
    log_info "  1. 进入插件目录: cd $PLUGIN_DIR"
    log_info "  2. 查看README文件: cat README.md"
    
    case $LANGUAGE in
        go)
            log_info "  3. 构建插件: go build -buildmode=plugin -o $PLUGIN_NAME.so ."
            log_info "  4. 运行测试: go test ./test/..."
            ;;
        rust)
            log_info "  3. 构建插件: wasm-pack build --target web"
            log_info "  4. 运行测试: cargo test"
            ;;
        c|cpp)
            log_info "  3. 构建插件: make"
            log_info "  4. 运行测试: make test"
            ;;
    esac
    
    echo
    log_info "插件开发文档: $PROJECT_ROOT/docs/guides/plugin-development.md"
    log_info "API文档: $PROJECT_ROOT/docs/api/plugin-api.md"
}

# 主函数
main() {
    log_info "go-musicfox v2 插件生成器"
    
    parse_args "$@"
    
    if [[ $INTERACTIVE == true ]]; then
        interactive_input
    fi
    
    validate_plugin_name
    create_plugin_structure
    
    # 根据语言生成代码
    case $LANGUAGE in
        go)
            generate_go_plugin
            ;;
        rust)
            if [[ $PLUGIN_TYPE == "wasm" ]]; then
                generate_rust_wasm_plugin
            else
                error_exit "Rust仅支持WebAssembly插件类型"
            fi
            ;;
        c)
            if [[ $PLUGIN_TYPE == "shared" ]]; then
                generate_c_shared_plugin
            else
                error_exit "C仅支持动态链接库插件类型"
            fi
            ;;
        cpp)
            if [[ $PLUGIN_TYPE == "shared" ]]; then
                # C++插件生成逻辑与C类似
                generate_c_shared_plugin
            else
                error_exit "C++仅支持动态链接库插件类型"
            fi
            ;;
    esac
    
    generate_common_files
    show_generation_summary
}

# 执行主函数
main "$@"