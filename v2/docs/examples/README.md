# 示例代码索引

本目录包含 go-musicfox v2 微内核插件架构的完整示例代码，帮助开发者快速理解和上手插件开发。

## 示例分类

### [基础插件示例](basic-plugin/)
最简单的插件实现，展示插件的基本结构和生命周期。

**包含内容：**
- 基础插件接口实现
- 插件配置文件
- 构建脚本
- 单元测试

**适合人群：**
- 插件开发新手
- 了解插件基本概念
- 学习插件开发流程

### [音乐源插件示例](music-source-plugin/)
完整的音乐源插件实现，展示 RPC 插件的开发模式。

**包含内容：**
- RPC 服务实现
- gRPC 协议定义
- 客户端集成
- 错误处理和重试
- 配置管理

**适合人群：**
- 需要集成外部音乐服务
- 学习 RPC 插件开发
- 了解分布式架构

### [音频处理插件示例](audio-processor/)
高性能音频处理插件，展示动态链接库插件的开发。

**包含内容：**
- 动态库插件实现
- 音频数据处理
- 性能优化技巧
- 跨平台编译
- 基准测试

**适合人群：**
- 需要高性能音频处理
- 学习动态库插件开发
- 了解性能优化

### [UI 扩展插件示例](ui-extension/)
用户界面扩展插件，展示热加载插件的开发。

**包含内容：**
- 热加载插件实现
- UI 组件开发
- 主题系统
- 状态管理
- 热更新机制

**适合人群：**
- 需要扩展用户界面
- 学习热加载插件开发
- 了解前端技术集成

## 快速开始

### 环境准备

```bash
# 确保 Go 版本
go version  # 需要 Go 1.21+

# 克隆项目
git clone https://github.com/go-musicfox/go-musicfox.git
cd go-musicfox/v2

# 安装依赖
go mod download
```

### 运行示例

#### 1. 基础插件示例

```bash
cd docs/examples/basic-plugin

# 构建插件
make build

# 运行测试
make test

# 加载插件
go run main.go
```

#### 2. 音乐源插件示例

```bash
cd docs/examples/music-source-plugin

# 生成 gRPC 代码
make proto

# 构建服务端
make build-server

# 构建客户端插件
make build-plugin

# 启动服务
./bin/music-source-server &

# 测试插件
go run client/main.go
```

#### 3. 音频处理插件示例

```bash
cd docs/examples/audio-processor

# 构建动态库
make build

# 运行基准测试
make benchmark

# 测试插件加载
go run test/main.go
```

#### 4. UI 扩展插件示例

```bash
cd docs/examples/ui-extension

# 构建插件
make build

# 启动开发服务器
make dev

# 测试热更新
make hot-reload
```

## 示例架构

### 目录结构

```
examples/
├── basic-plugin/           # 基础插件示例
│   ├── main.go            # 插件主文件
│   ├── plugin.json        # 插件配置
│   ├── Makefile          # 构建脚本
│   ├── README.md         # 说明文档
│   └── test/             # 测试文件
├── music-source-plugin/   # 音乐源插件示例
│   ├── proto/            # gRPC 协议定义
│   ├── server/           # RPC 服务端
│   ├── client/           # 插件客户端
│   ├── config/           # 配置文件
│   └── docker/           # Docker 配置
├── audio-processor/       # 音频处理插件示例
│   ├── src/              # 源代码
│   ├── include/          # 头文件
│   ├── test/             # 测试代码
│   ├── benchmark/        # 性能测试
│   └── build/            # 构建输出
└── ui-extension/          # UI 扩展插件示例
    ├── components/       # UI 组件
    ├── themes/           # 主题文件
    ├── assets/           # 静态资源
    └── hot-reload/       # 热更新逻辑
```

### 通用模式

每个示例都遵循以下模式：

1. **插件接口实现**
   ```go
   type ExamplePlugin struct {
       info   *plugin.PluginInfo
       config map[string]interface{}
       logger *slog.Logger
   }
   
   func (p *ExamplePlugin) GetInfo() *plugin.PluginInfo {
       return p.info
   }
   
   func (p *ExamplePlugin) Initialize(ctx plugin.PluginContext) error {
       p.logger = ctx.GetLogger()
       return nil
   }
   ```

2. **配置文件格式**
   ```json
   {
     "id": "example-plugin",
     "name": "Example Plugin",
     "version": "1.0.0",
     "type": "dynamic_library",
     "description": "An example plugin",
     "author": "go-musicfox",
     "config": {
       "setting1": "value1",
       "setting2": "value2"
     }
   }
   ```

3. **构建脚本**
   ```makefile
   .PHONY: build test clean
   
   build:
   	go build -buildmode=plugin -o plugin.so
   
   test:
   	go test ./...
   
   clean:
   	rm -f *.so
   ```

## 开发指南

### 选择合适的示例

根据您的需求选择合适的示例：

```
简单功能扩展 → 基础插件示例
外部服务集成 → 音乐源插件示例
高性能计算 → 音频处理插件示例
UI 界面扩展 → UI 扩展插件示例
```

### 修改示例代码

1. **复制示例目录**
   ```bash
   cp -r docs/examples/basic-plugin my-plugin
   cd my-plugin
   ```

2. **修改插件信息**
   ```json
   {
     "id": "my-plugin",
     "name": "My Plugin",
     "version": "1.0.0"
   }
   ```

3. **实现业务逻辑**
   ```go
   func (p *MyPlugin) DoSomething() error {
       // 实现您的业务逻辑
       return nil
   }
   ```

4. **添加测试**
   ```go
   func TestMyPlugin(t *testing.T) {
       plugin := &MyPlugin{}
       err := plugin.DoSomething()
       assert.NoError(t, err)
   }
   ```

### 调试技巧

1. **启用详细日志**
   ```go
   logger.Debug("Plugin operation", "step", "initialization")
   ```

2. **使用调试工具**
   ```bash
   # 使用 delve 调试
   dlv exec ./plugin-test
   ```

3. **性能分析**
   ```bash
   # 生成性能报告
   go test -cpuprofile=cpu.prof -memprofile=mem.prof
   go tool pprof cpu.prof
   ```

## 最佳实践

### 1. 错误处理

```go
func (p *ExamplePlugin) ProcessData(data []byte) error {
    if len(data) == 0 {
        return fmt.Errorf("empty data provided")
    }
    
    if err := p.validateData(data); err != nil {
        return fmt.Errorf("data validation failed: %w", err)
    }
    
    return nil
}
```

### 2. 资源管理

```go
func (p *ExamplePlugin) Start() error {
    // 获取资源
    conn, err := p.createConnection()
    if err != nil {
        return err
    }
    p.connection = conn
    
    return nil
}

func (p *ExamplePlugin) Cleanup() error {
    // 清理资源
    if p.connection != nil {
        p.connection.Close()
        p.connection = nil
    }
    return nil
}
```

### 3. 配置验证

```go
func (p *ExamplePlugin) ValidateConfig(config map[string]interface{}) error {
    required := []string{"host", "port", "timeout"}
    
    for _, key := range required {
        if _, exists := config[key]; !exists {
            return fmt.Errorf("missing required config: %s", key)
        }
    }
    
    return nil
}
```

### 4. 并发安全

```go
type SafePlugin struct {
    mutex sync.RWMutex
    data  map[string]interface{}
}

func (p *SafePlugin) GetData(key string) interface{} {
    p.mutex.RLock()
    defer p.mutex.RUnlock()
    return p.data[key]
}

func (p *SafePlugin) SetData(key string, value interface{}) {
    p.mutex.Lock()
    defer p.mutex.Unlock()
    p.data[key] = value
}
```

## 常见问题

### Q: 如何选择插件类型？

A: 根据以下标准选择：
- **性能要求高**：动态链接库插件
- **需要进程隔离**：RPC 插件
- **第三方开发**：WebAssembly 插件
- **频繁更新**：热加载插件

### Q: 插件如何与核心系统通信？

A: 通过以下方式：
- **事件总线**：异步事件通信
- **服务注册表**：服务发现和调用
- **插件上下文**：访问系统服务

### Q: 如何处理插件依赖？

A: 在插件配置中声明依赖：
```json
{
  "dependencies": [
    "audio-codec",
    "network-client"
  ]
}
```

### Q: 如何进行性能优化？

A: 参考以下建议：
- 使用合适的数据结构
- 避免频繁的内存分配
- 使用连接池和缓存
- 进行性能测试和分析

## 贡献示例

欢迎贡献新的示例代码：

1. **Fork 项目仓库**
2. **创建示例目录**
3. **编写完整的示例代码**
4. **添加详细的文档**
5. **提交 Pull Request**

### 示例要求

- 代码清晰易懂
- 包含完整的文档
- 提供构建和测试脚本
- 遵循项目编码规范
- 包含错误处理和测试

---

更多详细信息请查看具体的示例代码和文档。