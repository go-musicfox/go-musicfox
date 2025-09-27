# go-musicfox v2 微内核插件架构文档

欢迎使用 go-musicfox v2！这是一个基于微内核架构的现代化音乐播放器，采用插件系统实现高度的模块化和可扩展性。

## 🚀 快速开始

### 环境要求

- **Go**: 1.21 或更高版本
- **操作系统**: Linux, macOS, Windows
- **内存**: 最少 512MB RAM
- **存储**: 最少 100MB 可用空间

### 安装和运行

```bash
# 克隆项目
git clone https://github.com/go-musicfox/go-musicfox.git
cd go-musicfox/v2

# 安装依赖
go mod download

# 构建项目
make build

# 运行程序
./build/musicfox
```

### 快速体验

```bash
# 加载示例插件
./musicfox --load-plugin=./examples/basic-plugin/plugin.so

# 启动 Web 界面
./musicfox --web --port=8080

# 访问 http://localhost:8080
```

## 📚 文档导航

### 核心文档

#### [API 文档](api/README.md)
详细的 API 接口文档，包含所有核心组件的接口定义和使用说明。

- [微内核 API](api/kernel.md) - 系统核心接口
- [插件管理器 API](api/plugin-manager.md) - 插件管理接口
- [事件总线 API](api/event-bus.md) - 事件通信接口
- [服务注册表 API](api/service-registry.md) - 服务发现接口
- [安全管理器 API](api/security-manager.md) - 安全控制接口

#### [开发指南](guides/README.md)
完整的插件开发指南，从入门到精通。

- [插件开发快速入门](guides/plugin-quickstart.md) - 新手必读
- [动态链接库插件开发](guides/dynamic-library.md) - 高性能插件
- [RPC 插件开发](guides/rpc-plugin.md) - 分布式插件
- [WebAssembly 插件开发](guides/webassembly.md) - 安全沙箱插件
- [热加载插件开发](guides/hot-reload.md) - 可热更新插件
- [插件配置和部署](guides/plugin-config.md) - 配置管理
- [插件测试指南](guides/plugin-testing.md) - 测试最佳实践

#### [架构设计](architecture/README.md)
深入了解系统架构和设计理念。

- [微内核架构概述](architecture/microkernel.md) - 核心架构设计
- [插件系统设计](architecture/plugin-system.md) - 插件系统详解
- [安全机制说明](architecture/security.md) - 安全设计
- [性能优化指南](architecture/performance.md) - 性能优化

#### [示例代码](examples/README.md)
丰富的示例代码，快速上手开发。

- [基础插件示例](examples/basic-plugin/) - 最简单的插件
- [音乐源插件示例](examples/music-source-plugin/) - RPC 插件示例
- [音频处理插件示例](examples/audio-processor/) - 高性能插件示例
- [UI 扩展插件示例](examples/ui-extension/) - 界面扩展示例

#### [开发工具](tools/README.md)
提高开发效率的工具和方法。

- [开发工具使用说明](tools/development.md) - 开发工具链
- [调试和诊断工具](tools/debugging.md) - 问题排查
- [性能分析工具](tools/performance.md) - 性能优化

## 🏗️ 架构概览

### 微内核设计

go-musicfox v2 采用微内核架构，将系统分为最小化的内核和可扩展的插件层：

```
┌─────────────────────────────────────────────────────────────┐
│                        应用层                                │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────────────┐    │
│  │   CLI   │ │   TUI   │ │   GUI   │ │    Web UI       │    │
│  └─────────┘ └─────────┘ └─────────┘ └─────────────────┘    │
├─────────────────────────────────────────────────────────────┤
│                      微内核层                                │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────────────┐    │
│  │ 微内核  │ │插件管理 │ │事件总线 │ │   安全管理      │    │
│  └─────────┘ └─────────┘ └─────────┘ └─────────────────┘    │
├─────────────────────────────────────────────────────────────┤
│                       插件层                                │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────────────┐    │
│  │动态库   │ │RPC插件  │ │WASM插件 │ │   热加载插件    │    │
│  │插件     │ │         │ │         │ │                 │    │
│  └─────────┘ └─────────┘ └─────────┘ └─────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### 核心特性

- **🔧 高度模块化**: 核心功能最小化，通过插件扩展
- **🔌 混合插件系统**: 支持动态库、RPC、WebAssembly、热加载四种插件类型
- **🛡️ 安全可靠**: 插件沙箱隔离，权限控制机制
- **⚡ 高性能**: 异步事件通信，资源池化管理
- **🔄 热更新**: 支持插件的动态加载、卸载和热更新
- **🌐 跨平台**: 支持 Linux、macOS、Windows

### 插件类型

| 插件类型 | 特点 | 适用场景 | 示例 |
|---------|------|----------|------|
| **动态链接库** | 高性能，直接内存访问 | CPU密集型任务 | 音频处理、编解码 |
| **RPC插件** | 进程隔离，多语言支持 | 外部服务集成 | 音乐源、云服务 |
| **WebAssembly** | 安全沙箱，跨平台 | 第三方插件 | 社区扩展、实验功能 |
| **热加载** | 无需重启更新 | 频繁迭代组件 | UI扩展、主题 |

## 🎯 核心概念

### 微内核 (Kernel)
系统的核心，负责：
- 生命周期管理
- 组件协调
- 资源管理
- 依赖注入

### 插件管理器 (PluginManager)
管理所有插件，负责：
- 插件加载和卸载
- 生命周期管理
- 依赖解析
- 健康监控

### 事件总线 (EventBus)
异步通信机制，提供：
- 事件发布订阅
- 事件路由和过滤
- 优先级管理
- 批量处理

### 服务注册表 (ServiceRegistry)
服务发现机制，支持：
- 服务注册和发现
- 健康检查
- 负载均衡
- 依赖管理

### 安全管理器 (SecurityManager)
安全控制机制，包括：
- 插件验证
- 权限控制
- 沙箱管理
- 审计日志

## 🛠️ 开发流程

### 1. 选择插件类型

根据需求选择合适的插件类型：

```
性能要求高 → 动态链接库插件
需要进程隔离 → RPC 插件
第三方开发 → WebAssembly 插件
频繁更新 → 热加载插件
```

### 2. 创建插件项目

```bash
# 使用脚手架创建插件
musicfox-cli plugin create --name=my-plugin --type=dynamic

# 或手动创建
mkdir my-plugin && cd my-plugin
go mod init my-plugin
go get github.com/go-musicfox/go-musicfox/v2/pkg/plugin
```

### 3. 实现插件接口

```go
package main

import (
    "github.com/go-musicfox/go-musicfox/v2/pkg/plugin"
)

type MyPlugin struct {
    info *plugin.PluginInfo
}

func (p *MyPlugin) GetInfo() *plugin.PluginInfo {
    return &plugin.PluginInfo{
        Name:    "My Plugin",
        Version: "1.0.0",
        Author:  "Your Name",
    }
}

func (p *MyPlugin) Initialize(ctx plugin.PluginContext) error {
    // 初始化逻辑
    return nil
}

// 实现其他必需接口...
```

### 4. 构建和测试

```bash
# 构建插件
go build -buildmode=plugin -o my-plugin.so

# 运行测试
go test ./...

# 加载插件测试
musicfox --load-plugin=./my-plugin.so
```

## 📖 学习路径

### 初学者
1. 阅读 [插件开发快速入门](guides/plugin-quickstart.md)
2. 运行 [基础插件示例](examples/basic-plugin/)
3. 了解 [微内核架构概述](architecture/microkernel.md)
4. 尝试修改示例代码

### 进阶开发者
1. 深入学习 [API 文档](api/README.md)
2. 研究 [插件系统设计](architecture/plugin-system.md)
3. 实践不同类型的插件开发
4. 学习 [性能优化指南](architecture/performance.md)

### 高级开发者
1. 研究源码实现
2. 贡献核心功能
3. 设计复杂的插件架构
4. 参与社区建设

## 🤝 社区和支持

### 获取帮助
- **GitHub Issues**: [提交问题和建议](https://github.com/go-musicfox/go-musicfox/issues)
- **Discussions**: [参与讨论](https://github.com/go-musicfox/go-musicfox/discussions)
- **Wiki**: [查看常见问题](https://github.com/go-musicfox/go-musicfox/wiki)

### 贡献代码
1. Fork 项目仓库
2. 创建功能分支
3. 编写代码和测试
4. 提交 Pull Request

### 插件市场
- **官方插件**: [核心插件集合](https://plugins.go-musicfox.com/official)
- **社区插件**: [社区贡献插件](https://plugins.go-musicfox.com/community)
- **插件开发**: [发布您的插件](https://plugins.go-musicfox.com/publish)

## 📋 版本信息

- **当前版本**: v2.0.0
- **API 版本**: v2.0.0
- **最低 Go 版本**: 1.21
- **发布日期**: 2024年

### 版本兼容性
- **向后兼容**: 在主版本内保持 API 兼容性
- **废弃通知**: 废弃的 API 会提前一个版本通知
- **迁移指南**: 重大变更会提供详细的迁移指南

## 📄 许可证

本项目采用 [MIT 许可证](../LICENSE)。

---

**开始您的插件开发之旅吧！** 🎵

如果您有任何问题或建议，欢迎通过 [GitHub Issues](https://github.com/go-musicfox/go-musicfox/issues) 与我们联系。

欢迎使用 go-musicfox v2 微内核插件架构！本文档提供了完整的开发指南、API参考和最佳实践。

## 📚 文档导航

### 🚀 快速开始
- [插件开发快速入门](guides/plugin-quickstart.md) - 5分钟创建你的第一个插件
- [架构概述](architecture/microkernel.md) - 了解微内核设计理念
- [环境搭建](guides/development-setup.md) - 开发环境配置

### 📖 API 参考
- [微内核 API](api/kernel.md) - 核心内核接口
- [插件管理器 API](api/plugin-manager.md) - 插件生命周期管理
- [事件总线 API](api/event-bus.md) - 事件驱动通信
- [服务注册表 API](api/service-registry.md) - 服务发现与注册
- [安全管理器 API](api/security-manager.md) - 安全沙箱机制

### 🔧 开发指南
- [动态链接库插件](guides/dynamic-library.md) - 高性能核心插件开发
- [RPC插件开发](guides/rpc-plugin.md) - 分布式插件架构
- [WebAssembly插件](guides/webassembly.md) - 安全沙箱插件
- [热加载插件](guides/hot-reload.md) - 实时更新插件
- [插件配置管理](guides/plugin-config.md) - 配置和部署
- [插件测试指南](guides/plugin-testing.md) - 测试最佳实践

### 🏗️ 架构文档
- [微内核架构](architecture/microkernel.md) - 核心架构设计
- [插件系统设计](architecture/plugin-system.md) - 混合插件架构
- [安全机制](architecture/security.md) - 沙箱和权限控制
- [性能优化](architecture/performance.md) - 性能调优指南
- [错误处理](architecture/error-handling.md) - 错误处理策略

### 💡 示例和教程
- [音频处理插件示例](examples/audio-processor.md) - 完整的音频插件开发
- [音乐源插件示例](examples/music-source.md) - 网易云音乐插件实现
- [UI扩展插件示例](examples/ui-extension.md) - 自定义界面插件
- [第三方插件示例](examples/third-party.md) - WebAssembly插件开发
- [最佳实践](examples/best-practices.md) - 设计模式和最佳实践

### 🛠️ 开发工具
- [插件开发工具](tools/plugin-dev-tools.md) - 开发辅助工具
- [调试和诊断](tools/debugging.md) - 问题排查工具
- [性能分析](tools/performance-analysis.md) - 性能监控工具
- [插件打包工具](tools/packaging.md) - 插件分发工具

## 🎯 核心特性

### 微内核架构
- **轻量级内核**: 最小化核心功能，最大化扩展性
- **插件化设计**: 所有功能通过插件实现
- **热插拔支持**: 运行时动态加载/卸载插件
- **依赖注入**: 基于 dig 的依赖管理

### 混合插件系统
- **动态链接库**: 高性能核心功能插件
- **RPC通信**: 分布式插件架构
- **WebAssembly**: 安全沙箱第三方插件
- **热加载**: 实时更新的UI扩展插件

### 安全机制
- **沙箱隔离**: 插件运行环境隔离
- **权限控制**: 细粒度权限管理
- **资源限制**: CPU、内存、网络限制
- **安全审计**: 插件行为监控

## 🚀 快速开始

### 1. 环境准备

```bash
# 克隆项目
git clone https://github.com/go-musicfox/go-musicfox.git
cd go-musicfox/v2

# 安装依赖
go mod download

# 运行测试
go test ./...
```

### 2. 创建第一个插件

```go
package main

import (
    "context"
    "github.com/go-musicfox/go-musicfox/v2/pkg/plugin"
)

type HelloPlugin struct {
    plugin.BasePlugin
}

func (p *HelloPlugin) GetInfo() *plugin.PluginInfo {
    return &plugin.PluginInfo{
        Name:        "hello-plugin",
        Version:     "1.0.0",
        Description: "A simple hello world plugin",
        Author:      "Your Name",
    }
}

func (p *HelloPlugin) Start() error {
    p.Logger().Info("Hello from plugin!")
    return nil
}

// 插件入口点
func NewPlugin() plugin.Plugin {
    return &HelloPlugin{}
}
```

### 3. 加载和运行插件

```go
package main

import (
    "context"
    "github.com/go-musicfox/go-musicfox/v2/pkg/kernel"
    "github.com/go-musicfox/go-musicfox/v2/pkg/plugin"
)

func main() {
    // 创建微内核
    k := kernel.NewMicroKernel()
    
    // 初始化内核
    ctx := context.Background()
    if err := k.Initialize(ctx); err != nil {
        panic(err)
    }
    
    // 加载插件
    pm := k.GetPluginManager()
    if err := pm.LoadPlugin("./hello-plugin.so", plugin.PluginTypeDynamicLibrary); err != nil {
        panic(err)
    }
    
    // 启动插件
    if err := pm.StartPlugin("hello-plugin"); err != nil {
        panic(err)
    }
    
    // 启动内核
    if err := k.Start(ctx); err != nil {
        panic(err)
    }
}
```

## 📋 系统要求

- **Go版本**: 1.21+
- **操作系统**: Linux, macOS, Windows
- **内存**: 最小 512MB，推荐 2GB+
- **磁盘空间**: 100MB+

## 🤝 贡献指南

我们欢迎社区贡献！请查看以下资源：

- [贡献指南](CONTRIBUTING.md)
- [代码规范](CODE_STYLE.md)
- [问题报告](https://github.com/go-musicfox/go-musicfox/issues)
- [功能请求](https://github.com/go-musicfox/go-musicfox/discussions)

## 📄 许可证

本项目采用 MIT 许可证。详情请查看 [LICENSE](../../LICENSE) 文件。

## 🔗 相关链接

- [项目主页](https://github.com/go-musicfox/go-musicfox)
- [在线文档](https://docs.go-musicfox.com)
- [社区讨论](https://github.com/go-musicfox/go-musicfox/discussions)
- [更新日志](../../CHANGELOG.md)

---

**开始你的插件开发之旅吧！** 🎵