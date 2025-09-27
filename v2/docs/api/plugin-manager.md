# 插件管理器 API 文档

插件管理器是微内核架构的核心组件，负责插件的生命周期管理、类型适配和运行时监控。

## 接口定义

### PluginManager 接口

```go
type PluginManager interface {
    // 插件生命周期管理
    LoadPlugin(path string, pluginType PluginType) error
    UnloadPlugin(name string) error
    StartPlugin(name string) error
    StopPlugin(name string) error
    RestartPlugin(name string) error
    ReloadPlugin(name string) error
    
    // 插件查询
    GetPlugin(name string) (Plugin, error)
    GetPlugins() map[string]Plugin
    GetPluginInfo(name string) (*PluginInfo, error)
    GetPluginState(name string) (PluginState, error)
    
    // 插件发现
    ScanPlugins(directories []string) ([]*PluginInfo, error)
    ValidatePlugin(path string) error
    
    // 依赖管理
    ResolveDependencies(plugin Plugin) error
    GetDependencyGraph() *DependencyGraph
    
    // 监控和诊断
    GetPluginStats(name string) (*PluginStats, error)
    HealthCheck() error
    
    // 事件处理
    Subscribe(eventType string, handler EventHandler)
    Unsubscribe(eventType string, handler EventHandler)
}
```

## 核心方法

### 插件生命周期管理

#### LoadPlugin(path string, pluginType PluginType) error

加载指定路径的插件。

**参数:**
- `path`: 插件文件路径
- `pluginType`: 插件类型（dynamic_library, rpc, webassembly, hot_reload）

**返回值:**
- `error`: 加载失败时返回错误

**示例:**
```go
pm := kernel.GetPluginManager()

// 加载动态链接库插件
if err := pm.LoadPlugin("./plugins/audio-processor.so", plugin.PluginTypeDynamicLibrary); err != nil {
    log.Printf("Failed to load plugin: %v", err)
}

// 加载RPC插件
if err := pm.LoadPlugin("./plugins/netease-music", plugin.PluginTypeRPC); err != nil {
    log.Printf("Failed to load RPC plugin: %v", err)
}

// 加载WebAssembly插件
if err := pm.LoadPlugin("./plugins/custom-filter.wasm", plugin.PluginTypeWebAssembly); err != nil {
    log.Printf("Failed to load WASM plugin: %v", err)
}
```

**加载流程:**
1. 验证插件文件存在性和权限
2. 安全检查和签名验证
3. 根据插件类型选择对应的加载器
4. 创建插件实例
5. 初始化插件上下文
6. 解析插件依赖
7. 注册插件服务
8. 发送插件加载事件

#### StartPlugin(name string) error

启动已加载的插件。

**参数:**
- `name`: 插件名称

**返回值:**
- `error`: 启动失败时返回错误

**示例:**
```go
if err := pm.StartPlugin("audio-processor"); err != nil {
    log.Printf("Failed to start plugin: %v", err)
}
```

#### StopPlugin(name string) error

停止运行中的插件。

**参数:**
- `name`: 插件名称

**返回值:**
- `error`: 停止失败时返回错误

**示例:**
```go
if err := pm.StopPlugin("audio-processor"); err != nil {
    log.Printf("Failed to stop plugin: %v", err)
}
```

#### UnloadPlugin(name string) error

卸载插件，释放所有资源。

**参数:**
- `name`: 插件名称

**返回值:**
- `error`: 卸载失败时返回错误

**示例:**
```go
if err := pm.UnloadPlugin("audio-processor"); err != nil {
    log.Printf("Failed to unload plugin: %v", err)
}
```

#### RestartPlugin(name string) error

重启插件（先停止再启动）。

**参数:**
- `name`: 插件名称

**返回值:**
- `error`: 重启失败时返回错误

#### ReloadPlugin(name string) error

重新加载插件（支持热更新的插件类型）。

**参数:**
- `name`: 插件名称

**返回值:**
- `error`: 重载失败时返回错误

### 插件查询

#### GetPlugin(name string) (Plugin, error)

获取插件实例。

**参数:**
- `name`: 插件名称

**返回值:**
- `Plugin`: 插件接口实例
- `error`: 插件不存在时返回错误

**示例:**
```go
plugin, err := pm.GetPlugin("audio-processor")
if err != nil {
    log.Printf("Plugin not found: %v", err)
    return
}

// 类型断言获取具体插件接口
if audioPlugin, ok := plugin.(AudioProcessorPlugin); ok {
    result, err := audioPlugin.ProcessAudio(audioData, 44100, 2)
    if err != nil {
        log.Printf("Audio processing failed: %v", err)
    }
}
```

#### GetPlugins() map[string]Plugin

获取所有已加载的插件。

**返回值:**
- `map[string]Plugin`: 插件名称到插件实例的映射

**示例:**
```go
plugins := pm.GetPlugins()
for name, plugin := range plugins {
    info := plugin.GetInfo()
    log.Printf("Plugin: %s, Version: %s, State: %s", 
        name, info.Version, plugin.GetState())
}
```

#### GetPluginInfo(name string) (*PluginInfo, error)

获取插件元信息。

**参数:**
- `name`: 插件名称

**返回值:**
- `*PluginInfo`: 插件信息结构
- `error`: 插件不存在时返回错误

**示例:**
```go
info, err := pm.GetPluginInfo("audio-processor")
if err != nil {
    log.Printf("Failed to get plugin info: %v", err)
    return
}

log.Printf("Plugin Info: Name=%s, Version=%s, Author=%s", 
    info.Name, info.Version, info.Author)
```

### 插件发现

#### ScanPlugins(directories []string) ([]*PluginInfo, error)

扫描指定目录中的插件。

**参数:**
- `directories`: 要扫描的目录列表

**返回值:**
- `[]*PluginInfo`: 发现的插件信息列表
- `error`: 扫描失败时返回错误

**示例:**
```go
directories := []string{
    "./plugins",
    "/usr/local/lib/go-musicfox/plugins",
    "~/.go-musicfox/plugins",
}

pluginInfos, err := pm.ScanPlugins(directories)
if err != nil {
    log.Printf("Failed to scan plugins: %v", err)
    return
}

for _, info := range pluginInfos {
    log.Printf("Found plugin: %s v%s at %s", 
        info.Name, info.Version, info.Path)
}
```

#### ValidatePlugin(path string) error

验证插件文件的有效性。

**参数:**
- `path`: 插件文件路径

**返回值:**
- `error`: 验证失败时返回错误

**示例:**
```go
if err := pm.ValidatePlugin("./plugins/suspicious-plugin.so"); err != nil {
    log.Printf("Plugin validation failed: %v", err)
    return
}
```

### 依赖管理

#### ResolveDependencies(plugin Plugin) error

解析插件依赖关系。

**参数:**
- `plugin`: 要解析依赖的插件

**返回值:**
- `error`: 依赖解析失败时返回错误

**示例:**
```go
plugin, _ := pm.GetPlugin("music-visualizer")
if err := pm.ResolveDependencies(plugin); err != nil {
    log.Printf("Failed to resolve dependencies: %v", err)
}
```

#### GetDependencyGraph() *DependencyGraph

获取插件依赖图。

**返回值:**
- `*DependencyGraph`: 依赖关系图

**示例:**
```go
graph := pm.GetDependencyGraph()
for _, node := range graph.GetNodes() {
    deps := graph.GetDependencies(node.Name)
    log.Printf("Plugin %s depends on: %v", node.Name, deps)
}
```

## 插件类型和加载器

### 插件类型

```go
type PluginType string

const (
    PluginTypeDynamicLibrary PluginType = "dynamic_library"
    PluginTypeRPC           PluginType = "rpc"
    PluginTypeWebAssembly   PluginType = "webassembly"
    PluginTypeHotReload     PluginType = "hot_reload"
)
```

### 加载器接口

```go
type PluginLoader interface {
    LoadPlugin(path string) (Plugin, error)
    UnloadPlugin(plugin Plugin) error
    ValidatePlugin(path string) error
    GetSupportedExtensions() []string
}
```

### 动态链接库加载器

```go
type DynamicLibraryLoader struct {
    logger *slog.Logger
}

func (dl *DynamicLibraryLoader) LoadPlugin(path string) (Plugin, error) {
    // 使用 plugin.Open 加载 .so 文件
    p, err := plugin.Open(path)
    if err != nil {
        return nil, fmt.Errorf("failed to open plugin: %w", err)
    }
    
    // 查找插件入口点
    newPluginSymbol, err := p.Lookup("NewPlugin")
    if err != nil {
        return nil, fmt.Errorf("plugin entry point not found: %w", err)
    }
    
    // 类型断言并调用构造函数
    newPlugin, ok := newPluginSymbol.(func() Plugin)
    if !ok {
        return nil, fmt.Errorf("invalid plugin entry point signature")
    }
    
    return newPlugin(), nil
}
```

### RPC插件加载器

```go
type RPCPluginLoader struct {
    logger *slog.Logger
}

func (rpc *RPCPluginLoader) LoadPlugin(path string) (Plugin, error) {
    // 启动插件进程
    cmd := exec.Command(path)
    
    // 设置环境变量和管道
    cmd.Env = append(os.Environ(), "PLUGIN_MODE=rpc")
    
    stdin, err := cmd.StdinPipe()
    if err != nil {
        return nil, err
    }
    
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return nil, err
    }
    
    // 启动进程
    if err := cmd.Start(); err != nil {
        return nil, err
    }
    
    // 创建RPC客户端
    client := rpc.NewClient(stdin, stdout)
    
    return &RPCPlugin{
        client: client,
        cmd:    cmd,
    }, nil
}
```

## 插件状态管理

### 插件状态

```go
type PluginState int

const (
    PluginStateUnloaded PluginState = iota
    PluginStateLoading
    PluginStateLoaded
    PluginStateStarting
    PluginStateRunning
    PluginStateStopping
    PluginStateStopped
    PluginStateError
    PluginStateReloading
)

func (s PluginState) String() string {
    switch s {
    case PluginStateUnloaded:
        return "unloaded"
    case PluginStateLoading:
        return "loading"
    case PluginStateLoaded:
        return "loaded"
    case PluginStateStarting:
        return "starting"
    case PluginStateRunning:
        return "running"
    case PluginStateStopping:
        return "stopping"
    case PluginStateStopped:
        return "stopped"
    case PluginStateError:
        return "error"
    case PluginStateReloading:
        return "reloading"
    default:
        return "unknown"
    }
}
```

### 状态转换

```go
type StateTransition struct {
    From      PluginState
    To        PluginState
    Action    string
    Timestamp time.Time
}

// 有效的状态转换
var validTransitions = map[PluginState][]PluginState{
    PluginStateUnloaded:  {PluginStateLoading},
    PluginStateLoading:   {PluginStateLoaded, PluginStateError},
    PluginStateLoaded:    {PluginStateStarting, PluginStateUnloaded},
    PluginStateStarting:  {PluginStateRunning, PluginStateError},
    PluginStateRunning:   {PluginStateStopping, PluginStateReloading, PluginStateError},
    PluginStateStopping:  {PluginStateStopped, PluginStateError},
    PluginStateStopped:   {PluginStateStarting, PluginStateUnloaded},
    PluginStateError:     {PluginStateUnloaded, PluginStateLoading},
    PluginStateReloading: {PluginStateRunning, PluginStateError},
}
```

## 事件系统

### 插件事件

```go
type PluginEvent struct {
    Type      string      `json:"type"`
    Plugin    string      `json:"plugin"`
    Timestamp time.Time   `json:"timestamp"`
    Data      interface{} `json:"data,omitempty"`
}

// 事件类型常量
const (
    EventPluginLoaded    = "plugin.loaded"
    EventPluginUnloaded  = "plugin.unloaded"
    EventPluginStarted   = "plugin.started"
    EventPluginStopped   = "plugin.stopped"
    EventPluginError     = "plugin.error"
    EventPluginReloaded  = "plugin.reloaded"
)
```

### 事件处理

```go
// 订阅插件事件
pm.Subscribe(EventPluginLoaded, func(event *PluginEvent) {
    log.Printf("Plugin loaded: %s", event.Plugin)
})

pm.Subscribe(EventPluginError, func(event *PluginEvent) {
    if errorData, ok := event.Data.(map[string]interface{}); ok {
        log.Printf("Plugin error: %s - %v", event.Plugin, errorData["error"])
        
        // 自动重启插件
        if errorData["recoverable"].(bool) {
            go func() {
                time.Sleep(5 * time.Second)
                if err := pm.RestartPlugin(event.Plugin); err != nil {
                    log.Printf("Failed to restart plugin: %v", err)
                }
            }()
        }
    }
})
```

## 监控和统计

### 插件统计

```go
type PluginStats struct {
    Name         string        `json:"name"`
    State        PluginState   `json:"state"`
    LoadTime     time.Time     `json:"load_time"`
    StartTime    time.Time     `json:"start_time"`
    Uptime       time.Duration `json:"uptime"`
    RestartCount int           `json:"restart_count"`
    ErrorCount   int           `json:"error_count"`
    
    // 资源使用统计
    MemoryUsage  int64 `json:"memory_usage"`
    CPUUsage     float64 `json:"cpu_usage"`
    
    // 调用统计
    CallCount    int64         `json:"call_count"`
    AvgLatency   time.Duration `json:"avg_latency"`
    ErrorRate    float64       `json:"error_rate"`
}

// 获取插件统计信息
stats, err := pm.GetPluginStats("audio-processor")
if err != nil {
    log.Printf("Failed to get plugin stats: %v", err)
    return
}

log.Printf("Plugin Stats: Uptime=%v, Calls=%d, Errors=%d, Memory=%dMB",
    stats.Uptime, stats.CallCount, stats.ErrorCount, stats.MemoryUsage/1024/1024)
```

### 健康检查

```go
func (pm *HybridPluginManager) HealthCheck() error {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()
    
    var errors []string
    
    for name, loadedPlugin := range pm.plugins {
        if err := loadedPlugin.Plugin.HealthCheck(); err != nil {
            errors = append(errors, fmt.Sprintf("%s: %v", name, err))
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("plugin health check failed: %s", strings.Join(errors, ", "))
    }
    
    return nil
}
```

## 配置选项

```yaml
# plugin-manager.yaml
plugin_manager:
  # 插件目录
  plugin_directories:
    - "./plugins"
    - "/usr/local/lib/go-musicfox/plugins"
    - "~/.go-musicfox/plugins"
  
  # 自动发现和加载
  auto_discovery: true
  auto_load: true
  
  # 加载配置
  load_timeout: "30s"
  start_timeout: "10s"
  stop_timeout: "10s"
  
  # 并发控制
  max_concurrent_loads: 5
  max_concurrent_starts: 3
  
  # 重试配置
  retry_count: 3
  retry_delay: "1s"
  
  # 健康检查
  health_check_interval: "30s"
  health_check_timeout: "5s"
  
  # 依赖管理
  dependency_resolution: true
  circular_dependency_check: true
  
  # 安全配置
  signature_verification: true
  trusted_sources:
    - "github.com/go-musicfox"
    - "plugins.go-musicfox.com"
  
  # 监控配置
  metrics_enabled: true
  metrics_interval: "10s"
  
  # 日志配置
  log_plugin_events: true
  log_level: "info"
```

## 错误处理

### 错误类型

```go
type PluginManagerError struct {
    Type      PluginManagerErrorType `json:"type"`
    Code      string                 `json:"code"`
    Message   string                 `json:"message"`
    Plugin    string                 `json:"plugin,omitempty"`
    Cause     error                  `json:"cause,omitempty"`
    Timestamp time.Time              `json:"timestamp"`
}

type PluginManagerErrorType int

const (
    PluginManagerErrorTypeLoad PluginManagerErrorType = iota
    PluginManagerErrorTypeStart
    PluginManagerErrorTypeStop
    PluginManagerErrorTypeDependency
    PluginManagerErrorTypeSecurity
    PluginManagerErrorTypeValidation
)
```

### 常见错误

| 错误代码 | 描述 | 解决方案 |
|---------|------|----------|
| `PLUGIN_NOT_FOUND` | 插件文件不存在 | 检查插件路径 |
| `PLUGIN_LOAD_FAILED` | 插件加载失败 | 检查插件格式和依赖 |
| `PLUGIN_START_TIMEOUT` | 插件启动超时 | 增加启动超时时间 |
| `DEPENDENCY_NOT_FOUND` | 依赖插件不存在 | 安装依赖插件 |
| `CIRCULAR_DEPENDENCY` | 循环依赖 | 重新设计插件依赖关系 |
| `SECURITY_VIOLATION` | 安全检查失败 | 检查插件签名和权限 |

## 最佳实践

### 1. 插件生命周期管理

```go
func managePluginLifecycle(pm PluginManager, pluginPath string, pluginType PluginType) error {
    // 1. 验证插件
    if err := pm.ValidatePlugin(pluginPath); err != nil {
        return fmt.Errorf("plugin validation failed: %w", err)
    }
    
    // 2. 加载插件
    if err := pm.LoadPlugin(pluginPath, pluginType); err != nil {
        return fmt.Errorf("plugin load failed: %w", err)
    }
    
    // 3. 获取插件信息
    info, err := pm.GetPluginInfo(pluginName)
    if err != nil {
        return fmt.Errorf("failed to get plugin info: %w", err)
    }
    
    // 4. 解析依赖
    plugin, _ := pm.GetPlugin(info.Name)
    if err := pm.ResolveDependencies(plugin); err != nil {
        return fmt.Errorf("dependency resolution failed: %w", err)
    }
    
    // 5. 启动插件
    if err := pm.StartPlugin(info.Name); err != nil {
        return fmt.Errorf("plugin start failed: %w", err)
    }
    
    return nil
}
```

### 2. 错误恢复

```go
func setupPluginErrorRecovery(pm PluginManager) {
    pm.Subscribe(EventPluginError, func(event *PluginEvent) {
        log.Printf("Plugin error detected: %s", event.Plugin)
        
        // 获取错误详情
        if errorData, ok := event.Data.(map[string]interface{}); ok {
            errorType := errorData["type"].(string)
            recoverable := errorData["recoverable"].(bool)
            
            if recoverable {
                // 尝试重启插件
                go func() {
                    time.Sleep(5 * time.Second)
                    
                    if err := pm.RestartPlugin(event.Plugin); err != nil {
                        log.Printf("Failed to restart plugin %s: %v", event.Plugin, err)
                        
                        // 如果重启失败，尝试重新加载
                        if err := pm.ReloadPlugin(event.Plugin); err != nil {
                            log.Printf("Failed to reload plugin %s: %v", event.Plugin, err)
                        }
                    }
                }()
            } else {
                log.Printf("Plugin %s error is not recoverable: %s", event.Plugin, errorType)
            }
        }
    })
}
```

### 3. 性能监控

```go
func monitorPluginPerformance(pm PluginManager) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        plugins := pm.GetPlugins()
        
        for name := range plugins {
            stats, err := pm.GetPluginStats(name)
            if err != nil {
                continue
            }
            
            // 检查性能指标
            if stats.MemoryUsage > 100*1024*1024 { // 100MB
                log.Printf("Plugin %s high memory usage: %dMB", name, stats.MemoryUsage/1024/1024)
            }
            
            if stats.ErrorRate > 0.1 { // 10%
                log.Printf("Plugin %s high error rate: %.2f%%", name, stats.ErrorRate*100)
            }
            
            if stats.AvgLatency > time.Second {
                log.Printf("Plugin %s high latency: %v", name, stats.AvgLatency)
            }
        }
    }
}
```

## 相关文档

- [插件接口规范](../guides/plugin-interface.md)
- [插件开发指南](../guides/plugin-quickstart.md)
- [插件类型详解](../guides/plugin-types.md)
- [依赖管理](../guides/dependency-management.md)
- [插件安全](../architecture/security.md)