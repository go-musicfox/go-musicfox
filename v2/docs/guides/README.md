# 插件开发指南索引

本目录包含 go-musicfox v2 微内核插件架构的完整开发指南，帮助开发者快速上手插件开发。

## 快速入门

### [插件开发快速入门](plugin-quickstart.md)
新手必读的插件开发入门指南，包含环境搭建、基础概念和第一个插件的开发。

**内容包括：**
- 开发环境准备
- 插件系统概述
- 创建第一个插件
- 基本调试技巧

## 插件类型开发指南

### [动态链接库插件开发](dynamic-library.md)
适用于性能敏感的核心功能，如音频处理、编解码等。

**特点：**
- 高性能，接近原生代码速度
- 直接内存访问
- 适合 CPU 密集型任务
- 平台相关（.so/.dll/.dylib）

**适用场景：**
- 音频处理插件
- 编解码插件
- 数学计算插件

### [RPC 插件开发](rpc-plugin.md)
适用于独立服务和外部系统集成，如音乐源、云服务等。

**特点：**
- 进程隔离，稳定性高
- 支持多语言开发
- 网络通信，支持分布式
- 易于调试和维护

**适用场景：**
- 音乐源插件（网易云、Spotify等）
- 云存储插件
- 外部 API 集成

### [WebAssembly 插件开发](webassembly.md)
适用于第三方插件和社区贡献，提供安全的沙箱环境。

**特点：**
- 安全沙箱执行
- 跨平台兼容
- 多语言支持（Rust、C/C++、Go等）
- 资源限制和监控

**适用场景：**
- 社区插件
- 第三方扩展
- 实验性功能

### [热加载插件开发](hot-reload.md)
适用于需要频繁更新的 UI 组件和用户界面扩展。

**特点：**
- 无需重启即可更新
- 状态保持
- 快速迭代开发
- 版本管理和回滚

**适用场景：**
- UI 扩展插件
- 主题插件
- 布局组件

## 配置和部署

### [插件配置和部署](plugin-config.md)
详细介绍插件的配置管理、打包和部署流程。

**内容包括：**
- 插件配置文件格式
- 资源限制配置
- 安全策略配置
- 打包和分发

### [插件测试指南](plugin-testing.md)
插件开发的测试最佳实践和工具使用。

**内容包括：**
- 单元测试编写
- 集成测试策略
- 性能测试方法
- 调试技巧

## 开发流程

### 1. 选择插件类型

根据您的需求选择合适的插件类型：

```
性能要求高 → 动态链接库插件
需要进程隔离 → RPC 插件
第三方开发 → WebAssembly 插件
频繁更新 → 热加载插件
```

### 2. 环境准备

```bash
# 安装 Go 1.21+
go version

# 克隆项目
git clone https://github.com/go-musicfox/go-musicfox.git
cd go-musicfox/v2

# 安装依赖
go mod download
```

### 3. 创建插件项目

```bash
# 创建插件目录
mkdir my-plugin
cd my-plugin

# 初始化 Go 模块
go mod init my-plugin

# 添加依赖
go get github.com/go-musicfox/go-musicfox/v2/pkg/plugin
```

### 4. 实现插件接口

```go
package main

import (
    "github.com/go-musicfox/go-musicfox/v2/pkg/plugin"
)

type MyPlugin struct {
    // 插件状态
}

func (p *MyPlugin) GetInfo() *plugin.PluginInfo {
    return &plugin.PluginInfo{
        Name:    "My Plugin",
        Version: "1.0.0",
        // ...
    }
}

// 实现其他必需的接口方法
```

### 5. 构建和测试

```bash
# 构建插件
go build -buildmode=plugin -o my-plugin.so

# 运行测试
go test ./...
```

### 6. 部署和使用

```bash
# 复制插件到插件目录
cp my-plugin.so ~/.go-musicfox/plugins/

# 更新配置文件
vim ~/.go-musicfox/config.yaml
```

## 最佳实践

### 错误处理

```go
func (p *MyPlugin) SomeMethod() error {
    if err := p.doSomething(); err != nil {
        return fmt.Errorf("plugin operation failed: %w", err)
    }
    return nil
}
```

### 日志记录

```go
func (p *MyPlugin) Initialize(ctx plugin.PluginContext) error {
    logger := ctx.GetLogger()
    logger.Info("Initializing plugin", "name", p.GetInfo().Name)
    return nil
}
```

### 配置管理

```go
func (p *MyPlugin) UpdateConfig(config map[string]interface{}) error {
    if value, ok := config["my_setting"]; ok {
        p.mySetting = value.(string)
    }
    return nil
}
```

### 资源清理

```go
func (p *MyPlugin) Cleanup() error {
    // 清理资源
    if p.connection != nil {
        p.connection.Close()
    }
    return nil
}
```

## 调试技巧

### 1. 启用调试日志

```yaml
# config.yaml
log:
  level: debug
  format: json
```

### 2. 使用调试工具

```bash
# 使用 delve 调试
dlv exec ./musicfox
```

### 3. 插件状态检查

```bash
# 检查插件状态
curl http://localhost:8080/api/plugins/status
```

## 常见问题

### Q: 插件加载失败怎么办？

A: 检查以下几点：
1. 插件文件路径是否正确
2. 插件是否实现了必需的接口
3. 依赖是否满足
4. 查看错误日志获取详细信息

### Q: 如何处理插件间通信？

A: 使用事件总线进行异步通信：

```go
// 发布事件
ctx.GetEventBus().Publish("my.event", data)

// 订阅事件
ctx.GetEventBus().Subscribe("my.event", handler)
```

### Q: 如何进行性能优化？

A: 参考以下建议：
1. 使用合适的插件类型
2. 避免频繁的内存分配
3. 使用连接池和缓存
4. 进行性能测试和分析

## 社区资源

- [GitHub 仓库](https://github.com/go-musicfox/go-musicfox)
- [问题反馈](https://github.com/go-musicfox/go-musicfox/issues)
- [讨论区](https://github.com/go-musicfox/go-musicfox/discussions)
- [插件市场](https://plugins.go-musicfox.com)

## 贡献指南

欢迎贡献插件和改进文档：

1. Fork 项目仓库
2. 创建功能分支
3. 提交更改
4. 创建 Pull Request

---

更多详细信息请查看具体的开发指南文档。