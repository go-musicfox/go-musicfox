# Audio Backend Architecture

音频播放器后端架构是go-musicfox v2的核心组件，提供了统一的音频播放接口和多播放器后端支持。

## 架构概述

### 核心组件

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  PlayerManager  │────│  PlayerFactory  │────│  PlayerBackend  │
│                 │    │                 │    │   (Interface)   │
│ - 生命周期管理   │    │ - 后端注册      │    │                 │
│ - 播放控制      │    │ - 后端创建      │    │ - 播放控制      │
│ - 事件处理      │    │ - 能力检测      │    │ - 状态管理      │
│ - 后端切换      │    │ - 优先级管理    │    │ - 事件通知      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  ConfigManager  │    │   BaseBackend   │    │ Concrete Backends│
│                 │    │                 │    │                 │
│ - 配置管理      │    │ - 基础实现      │    │ - BeepBackend   │
│ - 热重载        │    │ - 事件系统      │    │ - MPVBackend    │
│ - 验证          │    │ - 状态管理      │    │ - OSXBackend    │
│ - 持久化        │    │ - 音量控制      │    │ - MPDBackend    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 设计原则

1. **接口统一**: 所有播放器后端实现相同的`PlayerBackend`接口
2. **插件化**: 支持动态注册和发现播放器后端
3. **优先级管理**: 根据平台和能力自动选择最佳后端
4. **事件驱动**: 完整的事件系统支持状态变化通知
5. **配置热更新**: 支持运行时配置变更和后端切换
6. **错误恢复**: 完善的错误处理和恢复机制

## 快速开始

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/go-musicfox/go-musicfox/v2/pkg/audio"
)

func main() {
    // 创建播放器管理器
    manager := audio.NewPlayerManager()
    
    // 初始化
    ctx := context.Background()
    if err := manager.Initialize(ctx); err != nil {
        panic(err)
    }
    defer manager.Shutdown(ctx)
    
    // 播放音乐
    if err := manager.Play("http://example.com/song.mp3"); err != nil {
        fmt.Printf("播放失败: %v\n", err)
        return
    }
    
    // 设置音量
    manager.SetVolume(0.8)
    
    // 等待播放
    time.Sleep(10 * time.Second)
    
    // 暂停播放
    manager.Pause()
    
    fmt.Println("播放完成")
}
```

### 后端切换

```go
// 获取可用后端
available := manager.GetAvailableBackends()
fmt.Printf("可用后端: %v\n", available)

// 切换到指定后端
config := &audio.BackendConfig{
    Name:          "mpv",
    Enabled:       true,
    DefaultVolume: 0.7,
    Settings: map[string]interface{}{
        "audio_driver": "pulse",
        "cache":        true,
    },
}

if err := manager.SwitchBackend("mpv", config); err != nil {
    fmt.Printf("切换后端失败: %v\n", err)
} else {
    fmt.Printf("已切换到后端: %s\n", manager.GetCurrentBackendName())
}
```

### 事件处理

```go
// 添加状态变化事件处理器
manager.AddEventHandler(audio.EventStateChanged, func(event *audio.Event) {
    fmt.Printf("状态变化: %s -> %s\n", 
        event.Data["old_state"], 
        event.Data["new_state"])
})

// 添加音量变化事件处理器
manager.AddEventHandler(audio.EventVolumeChanged, func(event *audio.Event) {
    fmt.Printf("音量变化: %.2f -> %.2f\n", 
        event.Data["old_volume"], 
        event.Data["new_volume"])
})

// 添加后端切换事件处理器
manager.AddEventHandler("backend_switched", func(event *audio.Event) {
    fmt.Printf("后端切换: %s -> %s\n", 
        event.Data["from"], 
        event.Data["to"])
})
```

## 配置管理

### 配置文件结构

```json
{
  "default_backend": "beep",
  "hot_reload": true,
  "global_settings": {
    "default_volume": 0.8,
    "buffer_size": 4096,
    "sample_rate": 44100,
    "channels": 2,
    "auto_switch_backend": true,
    "retry_attempts": 3,
    "retry_delay": "1s"
  },
  "backends": {
    "beep": {
      "name": "beep",
      "enabled": true,
      "priority": 5,
      "buffer_size": 4096,
      "sample_rate": 44100,
      "channels": 2,
      "default_volume": 0.8,
      "settings": {}
    },
    "mpv": {
      "name": "mpv",
      "enabled": true,
      "priority": 7,
      "buffer_size": 8192,
      "sample_rate": 44100,
      "channels": 2,
      "default_volume": 0.8,
      "settings": {
        "audio_driver": "auto",
        "cache": true
      }
    }
  }
}
```

### 配置热更新

```go
// 创建配置管理器
configManager, err := audio.NewConfigManager("/path/to/config.json")
if err != nil {
    panic(err)
}
defer configManager.StopWatching()

// 启动配置监听
ctx := context.Background()
if err := configManager.StartWatching(ctx); err != nil {
    panic(err)
}

// 添加配置变化回调
configManager.AddConfigChangeCallback(func(oldConfig, newConfig *audio.AudioConfig) error {
    fmt.Printf("配置已更新: %s -> %s\n", 
        oldConfig.DefaultBackend, 
        newConfig.DefaultBackend)
    return nil
})

// 更新配置
config := configManager.GetConfig()
config.DefaultBackend = "mpv"
config.GlobalSettings.DefaultVolume = 0.9

if err := configManager.UpdateConfig(config); err != nil {
    fmt.Printf("更新配置失败: %v\n", err)
}
```

## 支持的播放器后端

### Beep Backend
- **平台**: 跨平台 (Linux, macOS, Windows)
- **依赖**: Go原生音频库
- **格式**: MP3, WAV, FLAC, OGG
- **特点**: 纯Go实现，无外部依赖
- **优先级**: 5

### MPV Backend
- **平台**: 跨平台 (Linux, macOS, Windows)
- **依赖**: 需要安装mpv
- **格式**: 支持mpv支持的所有格式
- **特点**: 功能强大，格式支持最全
- **优先级**: 7

### OSX Backend
- **平台**: macOS
- **依赖**: 系统原生AVAudioPlayer
- **格式**: MP3, WAV, M4A, AAC, FLAC, OGG
- **特点**: 系统集成度高，性能优秀
- **优先级**: 10

### Windows Backend
- **平台**: Windows
- **依赖**: 系统原生Windows Media Player API
- **格式**: MP3, WAV, WMA, M4A, AAC
- **特点**: 系统集成度高
- **优先级**: 10

### MPD Backend
- **平台**: Linux/Unix
- **依赖**: Music Player Daemon
- **格式**: MP3, WAV, FLAC, OGG, M4A, AAC
- **特点**: 支持远程控制，适合服务器环境
- **优先级**: 8

## 自定义播放器后端

### 实现PlayerBackend接口

```go
type CustomBackend struct {
    *audio.BaseBackend
    // 自定义字段
    customPlayer *CustomPlayer
}

func NewCustomBackend(config *audio.BackendConfig) *CustomBackend {
    capabilities := &audio.BackendCapabilities{
        SupportedFormats:   []string{"mp3", "wav"},
        SupportedPlatforms: []string{"linux", "darwin"},
        Features: map[string]bool{
            "seek":      true,
            "streaming": true,
            "volume":    true,
        },
        MaxVolume:        1.0,
        MinVolume:        0.0,
        SeekSupport:      true,
        StreamingSupport: true,
    }
    
    return &CustomBackend{
        BaseBackend: audio.NewBaseBackend("custom", "1.0.0", capabilities),
        customPlayer: NewCustomPlayer(),
    }
}

// 实现播放控制方法
func (c *CustomBackend) Play(url string) error {
    if err := c.customPlayer.LoadAndPlay(url); err != nil {
        return err
    }
    
    c.SetState(audio.StatePlaying)
    return nil
}

func (c *CustomBackend) Pause() error {
    if err := c.customPlayer.Pause(); err != nil {
        return err
    }
    
    c.SetState(audio.StatePaused)
    return nil
}

// 实现其他必需方法...

func (c *CustomBackend) IsAvailable() bool {
    // 检查自定义播放器是否可用
    return c.customPlayer.IsAvailable()
}
```

### 注册自定义后端

```go
// 创建后端信息
customInfo := &audio.BackendInfo{
    Name:        "custom",
    Version:     "1.0.0",
    Description: "Custom audio backend",
    Capabilities: &audio.BackendCapabilities{
        SupportedFormats:   []string{"mp3", "wav"},
        SupportedPlatforms: []string{"linux", "darwin"},
        Features: map[string]bool{
            "seek":      true,
            "streaming": true,
            "volume":    true,
        },
        SeekSupport:      true,
        StreamingSupport: true,
    },
    Platforms: []string{"linux", "darwin"},
    Priority:  6,
    Creator: func(config *audio.BackendConfig) (audio.PlayerBackend, error) {
        return NewCustomBackend(config), nil
    },
}

// 注册到工厂
manager := audio.NewPlayerManager()
if err := manager.factory.RegisterBackend(customInfo); err != nil {
    panic(err)
}
```

## 事件系统

### 支持的事件类型

- `EventStateChanged`: 播放状态变化
- `EventPositionChanged`: 播放位置变化
- `EventVolumeChanged`: 音量变化
- `EventTrackChanged`: 曲目变化
- `EventError`: 错误事件
- `EventBuffering`: 缓冲事件
- `backend_switched`: 后端切换事件

### 事件数据结构

```go
type Event struct {
    Type      EventType              `json:"type"`
    Timestamp time.Time              `json:"timestamp"`
    Data      map[string]interface{} `json:"data"`
    Source    string                 `json:"source"`
}
```

### 事件处理示例

```go
// 监听所有事件
manager.AddEventHandler(audio.EventStateChanged, func(event *audio.Event) {
    switch event.Type {
    case audio.EventStateChanged:
        handleStateChange(event)
    case audio.EventVolumeChanged:
        handleVolumeChange(event)
    case audio.EventError:
        handleError(event)
    }
})

func handleStateChange(event *audio.Event) {
    oldState := event.Data["old_state"].(string)
    newState := event.Data["new_state"].(string)
    
    fmt.Printf("播放状态: %s -> %s\n", oldState, newState)
    
    // 根据状态变化执行相应操作
    switch newState {
    case "playing":
        // 开始播放时的处理
        updateUI("playing")
    case "paused":
        // 暂停时的处理
        updateUI("paused")
    case "stopped":
        // 停止时的处理
        updateUI("stopped")
    }
}
```

## 错误处理

### 错误类型

- 后端不可用错误
- 播放失败错误
- 配置错误
- 网络错误
- 格式不支持错误

### 错误处理策略

```go
// 添加错误事件处理器
manager.AddEventHandler(audio.EventError, func(event *audio.Event) {
    errorType := event.Data["error_type"].(string)
    errorMessage := event.Data["message"].(string)
    
    switch errorType {
    case "backend_unavailable":
        // 尝试切换到其他可用后端
        handleBackendUnavailable()
    case "playback_failed":
        // 重试播放或跳过当前曲目
        handlePlaybackFailed()
    case "network_error":
        // 检查网络连接，可能需要重试
        handleNetworkError()
    default:
        fmt.Printf("未知错误: %s\n", errorMessage)
    }
})

func handleBackendUnavailable() {
    // 获取可用后端列表
    available := manager.GetAvailableBackends()
    if len(available) > 0 {
        // 切换到第一个可用后端
        config := &audio.BackendConfig{
            Name:    available[0],
            Enabled: true,
        }
        
        if err := manager.SwitchBackend(available[0], config); err != nil {
            fmt.Printf("切换后端失败: %v\n", err)
        } else {
            fmt.Printf("已自动切换到后端: %s\n", available[0])
        }
    }
}
```

## 性能优化

### 缓冲区配置

```go
// 根据系统性能调整缓冲区大小
config := &audio.BackendConfig{
    Name:       "beep",
    BufferSize: 8192, // 增大缓冲区减少延迟
    SampleRate: 44100,
    Channels:   2,
}
```

### 异步操作

```go
// 异步播放
go func() {
    if err := manager.Play("http://example.com/song.mp3"); err != nil {
        fmt.Printf("播放失败: %v\n", err)
    }
}()

// 异步后端切换
go func() {
    config := &audio.BackendConfig{Name: "mpv", Enabled: true}
    if err := manager.SwitchBackend("mpv", config); err != nil {
        fmt.Printf("切换失败: %v\n", err)
    }
}()
```

## 测试

### 运行单元测试

```bash
# 运行所有测试
go test ./pkg/audio/

# 运行特定测试
go test ./pkg/audio/ -run TestPlayerManager

# 运行基准测试
go test ./pkg/audio/ -bench=.

# 运行集成测试
go test ./pkg/audio/ -run TestIntegration
```

### 测试覆盖率

```bash
# 生成覆盖率报告
go test ./pkg/audio/ -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## 最佳实践

### 1. 后端选择策略

- 优先使用平台原生后端（OSX、Windows）
- 功能需求优先（MPV支持最多格式）
- 性能要求优先（原生后端性能最佳）
- 兼容性要求优先（Beep跨平台兼容性最好）

### 2. 错误处理

- 始终检查操作返回的错误
- 实现适当的错误恢复机制
- 使用事件系统监听错误状态
- 提供用户友好的错误信息

### 3. 资源管理

- 及时调用Cleanup()清理资源
- 使用context控制操作超时
- 避免长时间阻塞操作
- 合理设置缓冲区大小

### 4. 配置管理

- 使用配置文件管理后端设置
- 启用配置热重载提升用户体验
- 验证配置参数的有效性
- 提供合理的默认配置

### 5. 事件处理

- 使用异步事件处理避免阻塞
- 合理设计事件数据结构
- 避免在事件处理器中执行耗时操作
- 及时移除不需要的事件处理器

## 故障排除

### 常见问题

1. **后端不可用**
   - 检查依赖是否安装（如mpv）
   - 验证平台兼容性
   - 查看系统音频设备状态

2. **播放失败**
   - 检查音频文件格式是否支持
   - 验证网络连接（对于流媒体）
   - 检查文件路径是否正确

3. **配置加载失败**
   - 验证配置文件JSON格式
   - 检查文件权限
   - 查看配置参数是否有效

4. **事件不触发**
   - 确认事件处理器已正确注册
   - 检查事件类型是否匹配
   - 验证播放器状态是否正常

### 调试技巧

```go
// 启用详细日志
manager.AddEventHandler(audio.EventStateChanged, func(event *audio.Event) {
    fmt.Printf("[DEBUG] Event: %+v\n", event)
})

// 定期健康检查
go func() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if err := manager.HealthCheck(); err != nil {
                fmt.Printf("[WARNING] Health check failed: %v\n", err)
            }
        }
    }
}()
```

## 贡献指南

### 添加新的播放器后端

1. 实现`PlayerBackend`接口
2. 继承`BaseBackend`获得基础功能
3. 实现平台特定的播放逻辑
4. 添加完整的单元测试
5. 更新文档和示例

### 代码规范

- 遵循Go代码规范
- 添加适当的注释和文档
- 编写全面的测试用例
- 确保向后兼容性
- 使用语义化版本号

## 许可证

本项目采用MIT许可证，详见LICENSE文件。