# Audio Plugin

音频处理插件是go-musicfox v2架构中的核心插件，提供多播放器后端支持和运行时切换功能。

## 功能特性

### 多播放器后端支持

- **Beep Player** - Go原生音频库，跨平台支持
- **MPV Player** - 基于MPV媒体播放器，跨平台支持
- **OSX Player** - macOS原生AVAudioPlayer（仅macOS）
- **Windows Player** - Windows Media Player API（仅Windows）
- **MPD Player** - Music Player Daemon（Linux/Unix）

### 核心功能

- 播放控制（播放、暂停、停止、跳转）
- 音量控制
- 播放状态管理
- 运行时播放器后端切换
- 事件通知系统
- 多种音频格式支持

### 支持的音频格式

- MP3
- WAV
- FLAC
- OGG
- M4A
- AAC
- WMA
- APE

## 使用方法

### 基本使用

```go
package main

import (
    "context"
    "log"
    
    "github.com/go-musicfox/go-musicfox/v2/plugins/audio"
    "github.com/go-musicfox/go-musicfox/v2/pkg/event"
    "github.com/go-musicfox/go-musicfox/v2/pkg/model"
)

func main() {
    // 创建事件总线
    eventBus := event.NewDefaultEventBus()
    eventBus.Start(context.Background())
    
    // 创建音频插件
    audioPlugin := audio.NewAudioPlugin(eventBus)
    
    // 初始化插件
    err := audioPlugin.Initialize(context.Background(), nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // 启动插件
    err = audioPlugin.Start(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    
    // 播放音乐
    song := &model.Song{
        ID:     "test-song",
        Name:   "Test Song",
        Artist: "Test Artist",
        URL:    "path/to/audio/file.mp3",
    }
    
    err = audioPlugin.Play(song)
    if err != nil {
        log.Fatal(err)
    }
    
    // 设置音量
    audioPlugin.SetVolume(80)
    
    // 获取播放状态
    state := audioPlugin.GetState()
    log.Printf("Playing: %s by %s", state.CurrentSong.Name, state.CurrentSong.Artist)
}
```

### 配置选项

```go
config := &core.PluginConfig{
    Settings: map[string]interface{}{
        "backend":      "beep",    // 默认播放器后端
        "volume":       80,        // 默认音量 (0-100)
        "buffer_size":  4096,      // 缓冲区大小
        "sample_rate":  44100,     // 采样率
        "channels":     2,         // 声道数
    },
}

audioPlugin.Initialize(context.Background(), config)
```

### 切换播放器后端

```go
// 获取可用的播放器后端
backends := audioPlugin.GetAvailableBackends()
log.Printf("Available backends: %v", backends)

// 切换到MPV播放器
err := audioPlugin.SwitchBackend("mpv")
if err != nil {
    log.Printf("Failed to switch backend: %v", err)
}

// 获取当前播放器后端
currentBackend := audioPlugin.GetCurrentBackend()
log.Printf("Current backend: %s", currentBackend)
```

### 事件监听

```go
// 监听播放状态变化事件
eventBus.Subscribe(event.EventPlayerPlay, func(ctx context.Context, e event.Event) error {
    log.Printf("Song started playing: %v", e.GetData())
    return nil
})

// 监听音量变化事件
eventBus.Subscribe(event.EventPlayerVolumeChanged, func(ctx context.Context, e event.Event) error {
    log.Printf("Volume changed: %v", e.GetData())
    return nil
})
```

## 播放器后端详情

### Beep Player

- **平台**: 跨平台
- **依赖**: 无外部依赖
- **格式**: MP3, WAV, FLAC, OGG
- **特点**: 纯Go实现，轻量级

### MPV Player

- **平台**: 跨平台
- **依赖**: 需要安装mpv
- **格式**: 支持mpv支持的所有格式
- **特点**: 功能强大，格式支持最全

### OSX Player

- **平台**: macOS
- **依赖**: 系统原生
- **格式**: MP3, WAV, M4A, AAC, FLAC, OGG
- **特点**: 系统集成度高，性能优秀

### Windows Player

- **平台**: Windows
- **依赖**: 系统原生
- **格式**: MP3, WAV, WMA, M4A, AAC
- **特点**: 系统集成度高

### MPD Player

- **平台**: Linux/Unix
- **依赖**: 需要运行MPD服务
- **格式**: 支持MPD支持的所有格式
- **特点**: 适合服务器环境，支持网络控制

## 开发和测试

### 运行测试

```bash
cd plugins/audio
go test -v
```

### 运行基准测试

```bash
go test -bench=.
```

### 添加新的播放器后端

1. 实现 `PlayerBackend` 接口
2. 在工厂中注册新后端
3. 添加相应的测试

```go
// 实现新的播放器后端
type MyPlayer struct {
    *BasePlayer
    // 自定义字段
}

// 实现PlayerBackend接口的所有方法
func (p *MyPlayer) Initialize(ctx context.Context) error {
    // 初始化逻辑
    return nil
}

// ... 其他方法实现

// 注册到工厂
factory.RegisterBackend("my_player", func(config map[string]interface{}) (PlayerBackend, error) {
    return NewMyPlayer(config), nil
})
```

## 依赖项

- `github.com/faiface/beep` - Beep音频库
- `github.com/stretchr/testify` - 测试框架

## 许可证

本项目采用与go-musicfox主项目相同的许可证。