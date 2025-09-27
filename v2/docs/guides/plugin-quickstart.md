# 插件开发快速入门指南

本指南将帮助您快速开始为 go-musicfox v2 微内核架构开发插件。通过本指南，您将学会如何创建、构建、测试和部署插件。

## 目录

- [环境准备](#环境准备)
- [创建第一个插件](#创建第一个插件)
- [插件类型选择](#插件类型选择)
- [基本插件结构](#基本插件结构)
- [实现插件接口](#实现插件接口)
- [构建和测试](#构建和测试)
- [部署和调试](#部署和调试)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)

## 环境准备

### 系统要求

- Go 1.21 或更高版本
- Git
- 支持的操作系统：Linux、macOS、Windows

### 安装开发工具

```bash
# 安装 Go（如果尚未安装）
# 访问 https://golang.org/dl/ 下载安装

# 验证 Go 安装
go version

# 安装插件开发工具
go install github.com/go-musicfox/plugin-tools/cmd/plugin-gen@latest
go install github.com/go-musicfox/plugin-tools/cmd/plugin-test@latest
```

### 获取项目模板

```bash
# 克隆插件模板项目
git clone https://github.com/go-musicfox/plugin-template.git my-plugin
cd my-plugin

# 初始化模块
go mod init github.com/your-username/my-plugin
go mod tidy
```

## 创建第一个插件

让我们创建一个简单的音频效果插件作为示例。

### 1. 使用插件生成器

```bash
# 生成插件骨架
plugin-gen create --name="reverb-effect" --type="audio-processor" --author="Your Name"
```

这将创建以下目录结构：

```
reverb-effect/
├── cmd/
│   └── plugin/
│       └── main.go          # 插件入口点
├── internal/
│   ├── processor/
│   │   └── reverb.go        # 核心处理逻辑
│   └── config/
│       └── config.go        # 配置管理
├── pkg/
│   └── api/
│       └── interfaces.go    # 公共接口定义
├── configs/
│   └── plugin.yaml          # 插件配置文件
├── tests/
│   └── integration/
│       └── plugin_test.go   # 集成测试
├── docs/
│   └── README.md            # 插件文档
├── go.mod
├── go.sum
├── Makefile
└── plugin.json              # 插件元数据
```

### 2. 编辑插件元数据

编辑 `plugin.json` 文件：

```json
{
  "id": "reverb-effect",
  "name": "Reverb Audio Effect",
  "version": "1.0.0",
  "description": "A high-quality reverb audio effect plugin",
  "author": "Your Name",
  "license": "MIT",
  "type": "audio-processor",
  "category": "effect",
  "tags": ["audio", "effect", "reverb"],
  "api_version": "v2.0.0",
  "min_kernel_version": "2.0.0",
  "dependencies": [],
  "permissions": [
    {
      "id": "audio-processing",
      "description": "Process audio data",
      "required": true
    }
  ],
  "resources": {
    "memory": "50MB",
    "cpu": "30%",
    "disk": "10MB"
  },
  "entry_points": {
    "main": "cmd/plugin/main.go"
  },
  "build": {
    "output": "reverb-effect.so",
    "flags": ["-buildmode=plugin"]
  }
}
```

## 插件类型选择

go-musicfox v2 支持四种插件类型：

### 1. 动态库插件 (Dynamic Library)

**适用场景：**
- 高性能音频处理
- 系统级功能扩展
- 需要直接内存访问的场景

**优点：**
- 性能最佳
- 完整的系统访问权限
- 与主程序共享内存空间

**缺点：**
- 安全性较低
- 崩溃可能影响主程序
- 平台相关性强

### 2. RPC 插件 (Remote Procedure Call)

**适用场景：**
- 网络服务集成
- 跨语言插件开发
- 需要独立进程的场景

**优点：**
- 进程隔离，安全性高
- 支持多种编程语言
- 可以独立部署和更新

**缺点：**
- 通信开销较大
- 延迟相对较高
- 需要额外的进程管理

### 3. WebAssembly 插件 (WASM)

**适用场景：**
- 跨平台兼容性要求高
- 安全性要求严格
- 轻量级功能扩展

**优点：**
- 平台无关性
- 安全沙箱环境
- 体积小，加载快

**缺点：**
- 性能有一定损失
- 功能受限
- 调试相对困难

### 4. 热重载插件 (Hot Reload)

**适用场景：**
- 开发和调试阶段
- 需要频繁更新的功能
- 实验性功能

**优点：**
- 无需重启即可更新
- 开发效率高
- 便于调试和测试

**缺点：**
- 稳定性相对较低
- 内存使用可能增加
- 不适合生产环境

## 基本插件结构

### 插件接口定义

所有插件都必须实现基本的 `Plugin` 接口：

```go
// pkg/api/interfaces.go
package api

import (
    "context"
    "time"
)

// Plugin 基础插件接口
type Plugin interface {
    // 基本信息
    ID() string
    Name() string
    Version() string
    Description() string
    
    // 生命周期管理
    Initialize(ctx context.Context, config Config) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Cleanup() error
    
    // 状态查询
    Status() PluginStatus
    HealthCheck() error
    
    // 配置管理
    GetConfig() Config
    UpdateConfig(config Config) error
    
    // 元数据
    GetMetadata() PluginMetadata
}

// AudioProcessor 音频处理插件接口
type AudioProcessor interface {
    Plugin
    
    // 音频处理
    ProcessAudio(input *AudioBuffer) (*AudioBuffer, error)
    ProcessAudioStream(input <-chan *AudioBuffer, output chan<- *AudioBuffer) error
    
    // 参数控制
    GetParameters() []Parameter
    SetParameter(name string, value interface{}) error
    GetParameter(name string) (interface{}, error)
    
    // 格式支持
    GetSupportedFormats() []AudioFormat
    SetAudioFormat(format AudioFormat) error
}

// MusicSource 音乐源插件接口
type MusicSource interface {
    Plugin
    
    // 搜索功能
    Search(query string, options SearchOptions) ([]*Song, error)
    SearchAlbums(query string, options SearchOptions) ([]*Album, error)
    SearchArtists(query string, options SearchOptions) ([]*Artist, error)
    SearchPlaylists(query string, options SearchOptions) ([]*Playlist, error)
    
    // 内容获取
    GetSong(id string) (*Song, error)
    GetAlbum(id string) (*Album, error)
    GetArtist(id string) (*Artist, error)
    GetPlaylist(id string) (*Playlist, error)
    
    // 流媒体
    GetStreamURL(songID string, quality Quality) (string, error)
    GetLyrics(songID string) (*Lyrics, error)
    
    // 认证
    Login(credentials Credentials) error
    Logout() error
    IsLoggedIn() bool
}
```

### 数据结构定义

```go
// 基础数据结构
type PluginStatus int

const (
    StatusUnknown PluginStatus = iota
    StatusInitializing
    StatusRunning
    StatusStopped
    StatusError
)

type PluginMetadata struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Version     string            `json:"version"`
    Description string            `json:"description"`
    Author      string            `json:"author"`
    License     string            `json:"license"`
    Tags        []string          `json:"tags"`
    Capabilities []string         `json:"capabilities"`
    Metadata    map[string]interface{} `json:"metadata"`
}

type Config map[string]interface{}

// 音频相关数据结构
type AudioBuffer struct {
    Data       [][]float32 `json:"data"`        // 多声道音频数据
    SampleRate int         `json:"sample_rate"` // 采样率
    Channels   int         `json:"channels"`    // 声道数
    Frames     int         `json:"frames"`      // 帧数
    Timestamp  time.Time   `json:"timestamp"`   // 时间戳
}

type AudioFormat struct {
    SampleRate int    `json:"sample_rate"`
    Channels   int    `json:"channels"`
    BitDepth   int    `json:"bit_depth"`
    Format     string `json:"format"` // "pcm", "float32", etc.
}

type Parameter struct {
    Name        string      `json:"name"`
    DisplayName string      `json:"display_name"`
    Type        string      `json:"type"`
    Value       interface{} `json:"value"`
    MinValue    interface{} `json:"min_value,omitempty"`
    MaxValue    interface{} `json:"max_value,omitempty"`
    DefaultValue interface{} `json:"default_value"`
    Description string      `json:"description"`
    Unit        string      `json:"unit,omitempty"`
}

// 音乐相关数据结构
type Song struct {
    ID          string            `json:"id"`
    Title       string            `json:"title"`
    Artist      string            `json:"artist"`
    Album       string            `json:"album"`
    Duration    time.Duration     `json:"duration"`
    Genre       string            `json:"genre"`
    Year        int               `json:"year"`
    TrackNumber int               `json:"track_number"`
    CoverURL    string            `json:"cover_url"`
    Metadata    map[string]interface{} `json:"metadata"`
}

type SearchOptions struct {
    Limit  int    `json:"limit"`
    Offset int    `json:"offset"`
    Type   string `json:"type"`   // "song", "album", "artist", "playlist"
    Sort   string `json:"sort"`   // "relevance", "popularity", "date"
    Filter map[string]interface{} `json:"filter"`
}
```

## 实现插件接口

### 音频处理插件示例

```go
// internal/processor/reverb.go
package processor

import (
    "context"
    "fmt"
    "math"
    "sync"
    "time"
    
    "github.com/your-username/my-plugin/pkg/api"
)

type ReverbProcessor struct {
    // 基本信息
    id          string
    name        string
    version     string
    description string
    
    // 状态管理
    status      api.PluginStatus
    config      api.Config
    mutex       sync.RWMutex
    
    // 音频处理参数
    roomSize    float64
    damping     float64
    wetLevel    float64
    dryLevel    float64
    
    // 内部状态
    delayLines  [][]float32
    bufferSize  int
    sampleRate  int
    channels    int
    
    // 生命周期
    ctx         context.Context
    cancel      context.CancelFunc
}

// 创建新的混响处理器
func NewReverbProcessor() *ReverbProcessor {
    return &ReverbProcessor{
        id:          "reverb-effect",
        name:        "Reverb Audio Effect",
        version:     "1.0.0",
        description: "A high-quality reverb audio effect plugin",
        status:      api.StatusUnknown,
        
        // 默认参数
        roomSize:    0.5,
        damping:     0.5,
        wetLevel:    0.3,
        dryLevel:    0.7,
        
        bufferSize:  1024,
    }
}

// 实现 Plugin 接口
func (r *ReverbProcessor) ID() string {
    return r.id
}

func (r *ReverbProcessor) Name() string {
    return r.name
}

func (r *ReverbProcessor) Version() string {
    return r.version
}

func (r *ReverbProcessor) Description() string {
    return r.description
}

func (r *ReverbProcessor) Initialize(ctx context.Context, config api.Config) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    r.ctx, r.cancel = context.WithCancel(ctx)
    r.config = config
    
    // 从配置中读取参数
    if roomSize, ok := config["room_size"].(float64); ok {
        r.roomSize = roomSize
    }
    if damping, ok := config["damping"].(float64); ok {
        r.damping = damping
    }
    if wetLevel, ok := config["wet_level"].(float64); ok {
        r.wetLevel = wetLevel
    }
    if dryLevel, ok := config["dry_level"].(float64); ok {
        r.dryLevel = dryLevel
    }
    
    // 初始化延迟线
    r.initializeDelayLines()
    
    r.status = api.StatusInitializing
    return nil
}

func (r *ReverbProcessor) Start(ctx context.Context) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    if r.status != api.StatusInitializing {
        return fmt.Errorf("plugin not initialized")
    }
    
    r.status = api.StatusRunning
    return nil
}

func (r *ReverbProcessor) Stop(ctx context.Context) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    if r.cancel != nil {
        r.cancel()
    }
    
    r.status = api.StatusStopped
    return nil
}

func (r *ReverbProcessor) Cleanup() error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    // 清理资源
    r.delayLines = nil
    
    return nil
}

func (r *ReverbProcessor) Status() api.PluginStatus {
    r.mutex.RLock()
    defer r.mutex.RUnlock()
    return r.status
}

func (r *ReverbProcessor) HealthCheck() error {
    r.mutex.RLock()
    defer r.mutex.RUnlock()
    
    if r.status != api.StatusRunning {
        return fmt.Errorf("plugin not running")
    }
    
    return nil
}

func (r *ReverbProcessor) GetConfig() api.Config {
    r.mutex.RLock()
    defer r.mutex.RUnlock()
    return r.config
}

func (r *ReverbProcessor) UpdateConfig(config api.Config) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    r.config = config
    
    // 更新参数
    if roomSize, ok := config["room_size"].(float64); ok {
        r.roomSize = roomSize
    }
    if damping, ok := config["damping"].(float64); ok {
        r.damping = damping
    }
    if wetLevel, ok := config["wet_level"].(float64); ok {
        r.wetLevel = wetLevel
    }
    if dryLevel, ok := config["dry_level"].(float64); ok {
        r.dryLevel = dryLevel
    }
    
    return nil
}

func (r *ReverbProcessor) GetMetadata() api.PluginMetadata {
    return api.PluginMetadata{
        ID:          r.id,
        Name:        r.name,
        Version:     r.version,
        Description: r.description,
        Author:      "Your Name",
        License:     "MIT",
        Tags:        []string{"audio", "effect", "reverb"},
        Capabilities: []string{"audio-processing", "real-time"},
        Metadata: map[string]interface{}{
            "category":     "effect",
            "latency":      "low",
            "cpu_usage":    "medium",
            "memory_usage": "low",
        },
    }
}

// 实现 AudioProcessor 接口
func (r *ReverbProcessor) ProcessAudio(input *api.AudioBuffer) (*api.AudioBuffer, error) {
    r.mutex.RLock()
    defer r.mutex.RUnlock()
    
    if r.status != api.StatusRunning {
        return nil, fmt.Errorf("plugin not running")
    }
    
    // 检查音频格式
    if input.Channels != r.channels {
        return nil, fmt.Errorf("channel count mismatch: expected %d, got %d", r.channels, input.Channels)
    }
    
    // 创建输出缓冲区
    output := &api.AudioBuffer{
        Data:       make([][]float32, input.Channels),
        SampleRate: input.SampleRate,
        Channels:   input.Channels,
        Frames:     input.Frames,
        Timestamp:  input.Timestamp,
    }
    
    // 为每个声道分配内存
    for ch := 0; ch < input.Channels; ch++ {
        output.Data[ch] = make([]float32, input.Frames)
    }
    
    // 处理每个声道
    for ch := 0; ch < input.Channels; ch++ {
        r.processChannel(input.Data[ch], output.Data[ch], ch)
    }
    
    return output, nil
}

func (r *ReverbProcessor) ProcessAudioStream(input <-chan *api.AudioBuffer, output chan<- *api.AudioBuffer) error {
    for {
        select {
        case <-r.ctx.Done():
            return r.ctx.Err()
        case buffer, ok := <-input:
            if !ok {
                return nil // 输入流关闭
            }
            
            processed, err := r.ProcessAudio(buffer)
            if err != nil {
                return fmt.Errorf("audio processing failed: %w", err)
            }
            
            select {
            case output <- processed:
            case <-r.ctx.Done():
                return r.ctx.Err()
            }
        }
    }
}

func (r *ReverbProcessor) GetParameters() []api.Parameter {
    return []api.Parameter{
        {
            Name:         "room_size",
            DisplayName:  "Room Size",
            Type:         "float",
            Value:        r.roomSize,
            MinValue:     0.0,
            MaxValue:     1.0,
            DefaultValue: 0.5,
            Description:  "Size of the virtual room",
            Unit:         "",
        },
        {
            Name:         "damping",
            DisplayName:  "Damping",
            Type:         "float",
            Value:        r.damping,
            MinValue:     0.0,
            MaxValue:     1.0,
            DefaultValue: 0.5,
            Description:  "High frequency damping",
            Unit:         "",
        },
        {
            Name:         "wet_level",
            DisplayName:  "Wet Level",
            Type:         "float",
            Value:        r.wetLevel,
            MinValue:     0.0,
            MaxValue:     1.0,
            DefaultValue: 0.3,
            Description:  "Reverb signal level",
            Unit:         "dB",
        },
        {
            Name:         "dry_level",
            DisplayName:  "Dry Level",
            Type:         "float",
            Value:        r.dryLevel,
            MinValue:     0.0,
            MaxValue:     1.0,
            DefaultValue: 0.7,
            Description:  "Original signal level",
            Unit:         "dB",
        },
    }
}

func (r *ReverbProcessor) SetParameter(name string, value interface{}) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    floatValue, ok := value.(float64)
    if !ok {
        return fmt.Errorf("parameter value must be float64")
    }
    
    switch name {
    case "room_size":
        if floatValue < 0.0 || floatValue > 1.0 {
            return fmt.Errorf("room_size must be between 0.0 and 1.0")
        }
        r.roomSize = floatValue
    case "damping":
        if floatValue < 0.0 || floatValue > 1.0 {
            return fmt.Errorf("damping must be between 0.0 and 1.0")
        }
        r.damping = floatValue
    case "wet_level":
        if floatValue < 0.0 || floatValue > 1.0 {
            return fmt.Errorf("wet_level must be between 0.0 and 1.0")
        }
        r.wetLevel = floatValue
    case "dry_level":
        if floatValue < 0.0 || floatValue > 1.0 {
            return fmt.Errorf("dry_level must be between 0.0 and 1.0")
        }
        r.dryLevel = floatValue
    default:
        return fmt.Errorf("unknown parameter: %s", name)
    }
    
    return nil
}

func (r *ReverbProcessor) GetParameter(name string) (interface{}, error) {
    r.mutex.RLock()
    defer r.mutex.RUnlock()
    
    switch name {
    case "room_size":
        return r.roomSize, nil
    case "damping":
        return r.damping, nil
    case "wet_level":
        return r.wetLevel, nil
    case "dry_level":
        return r.dryLevel, nil
    default:
        return nil, fmt.Errorf("unknown parameter: %s", name)
    }
}

func (r *ReverbProcessor) GetSupportedFormats() []api.AudioFormat {
    return []api.AudioFormat{
        {SampleRate: 44100, Channels: 2, BitDepth: 32, Format: "float32"},
        {SampleRate: 48000, Channels: 2, BitDepth: 32, Format: "float32"},
        {SampleRate: 96000, Channels: 2, BitDepth: 32, Format: "float32"},
    }
}

func (r *ReverbProcessor) SetAudioFormat(format api.AudioFormat) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    r.sampleRate = format.SampleRate
    r.channels = format.Channels
    
    // 重新初始化延迟线
    r.initializeDelayLines()
    
    return nil
}

// 私有方法
func (r *ReverbProcessor) initializeDelayLines() {
    // 根据房间大小计算延迟线长度
    delayLengths := []int{
        int(r.roomSize * 0.03 * float64(r.sampleRate)), // 30ms
        int(r.roomSize * 0.05 * float64(r.sampleRate)), // 50ms
        int(r.roomSize * 0.07 * float64(r.sampleRate)), // 70ms
        int(r.roomSize * 0.09 * float64(r.sampleRate)), // 90ms
    }
    
    r.delayLines = make([][]float32, len(delayLengths))
    for i, length := range delayLengths {
        r.delayLines[i] = make([]float32, length)
    }
}

func (r *ReverbProcessor) processChannel(input, output []float32, channel int) {
    for i, sample := range input {
        // 简化的混响算法
        reverbSample := r.calculateReverb(sample, channel)
        
        // 混合干湿信号
        output[i] = float32(r.dryLevel)*sample + float32(r.wetLevel)*reverbSample
    }
}

func (r *ReverbProcessor) calculateReverb(sample float32, channel int) float32 {
    var reverbSum float32
    
    // 简化的多延迟线混响
    for i, delayLine := range r.delayLines {
        if len(delayLine) > 0 {
            // 获取延迟样本
            delayedSample := delayLine[0]
            
            // 移动延迟线
            copy(delayLine, delayLine[1:])
            
            // 添加新样本（带反馈）
            feedback := sample + delayedSample*float32(r.damping)
            delayLine[len(delayLine)-1] = feedback
            
            // 累加到混响信号
            reverbSum += delayedSample * float32(0.25) // 每个延迟线贡献25%
        }
    }
    
    return reverbSum
}
```

### 插件入口点

```go
// cmd/plugin/main.go
package main

import (
    "C"
    "context"
    "log"
    
    "github.com/your-username/my-plugin/internal/processor"
    "github.com/your-username/my-plugin/pkg/api"
)

// 全局插件实例
var pluginInstance api.AudioProcessor

// 插件入口点 - 动态库插件必须导出这些函数

//export PluginCreate
func PluginCreate() uintptr {
    pluginInstance = processor.NewReverbProcessor()
    return uintptr(0) // 返回插件实例的指针或ID
}

//export PluginDestroy
func PluginDestroy() {
    if pluginInstance != nil {
        pluginInstance.Cleanup()
        pluginInstance = nil
    }
}

//export PluginInitialize
func PluginInitialize(configJSON *C.char) C.int {
    if pluginInstance == nil {
        return -1
    }
    
    // 解析配置
    config := make(api.Config)
    // TODO: 解析 JSON 配置
    
    ctx := context.Background()
    if err := pluginInstance.Initialize(ctx, config); err != nil {
        log.Printf("Plugin initialization failed: %v", err)
        return -1
    }
    
    return 0
}

//export PluginStart
func PluginStart() C.int {
    if pluginInstance == nil {
        return -1
    }
    
    ctx := context.Background()
    if err := pluginInstance.Start(ctx); err != nil {
        log.Printf("Plugin start failed: %v", err)
        return -1
    }
    
    return 0
}

//export PluginStop
func PluginStop() C.int {
    if pluginInstance == nil {
        return -1
    }
    
    ctx := context.Background()
    if err := pluginInstance.Stop(ctx); err != nil {
        log.Printf("Plugin stop failed: %v", err)
        return -1
    }
    
    return 0
}

//export PluginProcessAudio
func PluginProcessAudio(inputPtr uintptr, outputPtr uintptr) C.int {
    if pluginInstance == nil {
        return -1
    }
    
    // TODO: 从指针转换为 AudioBuffer
    // input := (*api.AudioBuffer)(unsafe.Pointer(inputPtr))
    
    // output, err := pluginInstance.ProcessAudio(input)
    // if err != nil {
    //     log.Printf("Audio processing failed: %v", err)
    //     return -1
    // }
    
    // TODO: 将输出写入到 outputPtr
    
    return 0
}

// 主函数 - 动态库插件不需要，但保留用于测试
func main() {
    // 测试代码
    plugin := processor.NewReverbProcessor()
    
    ctx := context.Background()
    config := api.Config{
        "room_size": 0.7,
        "damping":   0.6,
        "wet_level": 0.4,
        "dry_level": 0.6,
    }
    
    if err := plugin.Initialize(ctx, config); err != nil {
        log.Fatalf("Failed to initialize plugin: %v", err)
    }
    
    if err := plugin.Start(ctx); err != nil {
        log.Fatalf("Failed to start plugin: %v", err)
    }
    
    log.Printf("Plugin %s v%s started successfully", plugin.Name(), plugin.Version())
    
    // 测试音频处理
    testAudio := &api.AudioBuffer{
        Data:       [][]float32{{0.1, 0.2, 0.3}, {0.1, 0.2, 0.3}},
        SampleRate: 44100,
        Channels:   2,
        Frames:     3,
    }
    
    output, err := plugin.ProcessAudio(testAudio)
    if err != nil {
        log.Fatalf("Failed to process audio: %v", err)
    }
    
    log.Printf("Processed audio: %v", output.Data)
    
    plugin.Stop(ctx)
    plugin.Cleanup()
}
```

## 构建和测试

### 构建插件

创建 `Makefile`：

```makefile
# Makefile
.PHONY: build test clean install

# 变量定义
PLUGIN_NAME := reverb-effect
VERSION := 1.0.0
BUILD_DIR := build
DIST_DIR := dist

# Go 构建参数
GO_BUILD_FLAGS := -ldflags "-X main.version=$(VERSION)"
PLUGIN_BUILD_FLAGS := -buildmode=plugin

# 默认目标
all: build

# 构建动态库插件
build-plugin:
	@echo "Building plugin..."
	@mkdir -p $(BUILD_DIR)
	go build $(GO_BUILD_FLAGS) $(PLUGIN_BUILD_FLAGS) -o $(BUILD_DIR)/$(PLUGIN_NAME).so ./cmd/plugin

# 构建可执行文件（用于测试）
build-exe:
	@echo "Building executable..."
	@mkdir -p $(BUILD_DIR)
	go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(PLUGIN_NAME) ./cmd/plugin

# 构建所有目标
build: build-plugin build-exe

# 运行测试
test:
	@echo "Running tests..."
	go test -v ./...

# 运行集成测试
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./tests/integration/...

# 运行基准测试
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# 代码检查
lint:
	@echo "Running linter..."
	golangci-lint run

# 格式化代码
fmt:
	@echo "Formatting code..."
	go fmt ./...

# 清理构建文件
clean:
	@echo "Cleaning build files..."
	rm -rf $(BUILD_DIR) $(DIST_DIR)

# 安装依赖
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# 创建发布包
package: build
	@echo "Creating package..."
	@mkdir -p $(DIST_DIR)
	tar -czf $(DIST_DIR)/$(PLUGIN_NAME)-$(VERSION).tar.gz -C $(BUILD_DIR) .
	cp plugin.json $(DIST_DIR)/
	cp configs/plugin.yaml $(DIST_DIR)/
	cp docs/README.md $(DIST_DIR)/

# 安装插件到本地
install: build
	@echo "Installing plugin..."
	mkdir -p ~/.go-musicfox/plugins/$(PLUGIN_NAME)
	cp $(BUILD_DIR)/$(PLUGIN_NAME).so ~/.go-musicfox/plugins/$(PLUGIN_NAME)/
	cp plugin.json ~/.go-musicfox/plugins/$(PLUGIN_NAME)/
	cp configs/plugin.yaml ~/.go-musicfox/plugins/$(PLUGIN_NAME)/

# 卸载插件
uninstall:
	@echo "Uninstalling plugin..."
	rm -rf ~/.go-musicfox/plugins/$(PLUGIN_NAME)

# 开发模式（监听文件变化并重新构建）
dev:
	@echo "Starting development mode..."
	air -c .air.toml

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  build          - Build plugin and executable"
	@echo "  build-plugin   - Build plugin only"
	@echo "  build-exe      - Build executable only"
	@echo "  test           - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  benchmark      - Run benchmarks"
	@echo "  lint           - Run code linter"
	@echo "  fmt            - Format code"
	@echo "  clean          - Clean build files"
	@echo "  deps           - Install dependencies"
	@echo "  package        - Create release package"
	@echo "  install        - Install plugin locally"
	@echo "  uninstall      - Uninstall plugin"
	@echo "  dev            - Start development mode"
	@echo "  help           - Show this help"
```

### 构建命令

```bash
# 安装依赖
make deps

# 构建插件
make build

# 运行测试
make test

# 代码检查
make lint

# 创建发布包
make package
```

### 单元测试

```go
// tests/integration/plugin_test.go
package integration

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "github.com/your-username/my-plugin/internal/processor"
    "github.com/your-username/my-plugin/pkg/api"
)

func TestReverbProcessor_BasicFunctionality(t *testing.T) {
    // 创建插件实例
    plugin := processor.NewReverbProcessor()
    require.NotNil(t, plugin)
    
    // 测试基本信息
    assert.Equal(t, "reverb-effect", plugin.ID())
    assert.Equal(t, "Reverb Audio Effect", plugin.Name())
    assert.Equal(t, "1.0.0", plugin.Version())
    
    // 测试初始化
    ctx := context.Background()
    config := api.Config{
        "room_size": 0.7,
        "damping":   0.6,
        "wet_level": 0.4,
        "dry_level": 0.6,
    }
    
    err := plugin.Initialize(ctx, config)
    require.NoError(t, err)
    
    // 测试启动
    err = plugin.Start(ctx)
    require.NoError(t, err)
    assert.Equal(t, api.StatusRunning, plugin.Status())
    
    // 测试健康检查
    err = plugin.HealthCheck()
    assert.NoError(t, err)
    
    // 测试停止
    err = plugin.Stop(ctx)
    require.NoError(t, err)
    assert.Equal(t, api.StatusStopped, plugin.Status())
    
    // 测试清理
    err = plugin.Cleanup()
    assert.NoError(t, err)
}

func TestReverbProcessor_AudioProcessing(t *testing.T) {
    plugin := processor.NewReverbProcessor()
    
    ctx := context.Background()
    config := api.Config{
        "room_size": 0.5,
        "damping":   0.5,
        "wet_level": 0.3,
        "dry_level": 0.7,
    }
    
    err := plugin.Initialize(ctx, config)
    require.NoError(t, err)
    
    err = plugin.Start(ctx)
    require.NoError(t, err)
    
    // 设置音频格式
    format := api.AudioFormat{
        SampleRate: 44100,
        Channels:   2,
        BitDepth:   32,
        Format:     "float32",
    }
    err = plugin.SetAudioFormat(format)
    require.NoError(t, err)
    
    // 创建测试音频数据
    testAudio := &api.AudioBuffer{
        Data: [][]float32{
            {0.1, 0.2, 0.3, 0.4, 0.5},
            {0.1, 0.2, 0.3, 0.4, 0.5},
        },
        SampleRate: 44100,
        Channels:   2,
        Frames:     5,
        Timestamp:  time.Now(),
    }
    
    // 处理音频
    output, err := plugin.ProcessAudio(testAudio)
    require.NoError(t, err)
    require.NotNil(t, output)
    
    // 验证输出
    assert.Equal(t, testAudio.SampleRate, output.SampleRate)
    assert.Equal(t, testAudio.Channels, output.Channels)
    assert.Equal(t, testAudio.Frames, output.Frames)
    assert.Len(t, output.Data, 2)
    assert.Len(t, output.Data[0], 5)
    assert.Len(t, output.Data[1], 5)
    
    // 验证音频数据已被处理（不应该完全相同）
    for ch := 0; ch < output.Channels; ch++ {
        for i := 0; i < output.Frames; i++ {
            // 由于混响效果，输出应该与输入不同
            // 但在初始几个样本中，差异可能很小
            assert.True(t, output.Data[ch][i] != 0.0, "Output should not be zero")
        }
    }
    
    plugin.Stop(ctx)
    plugin.Cleanup()
}

func TestReverbProcessor_Parameters(t *testing.T) {
    plugin := processor.NewReverbProcessor()
    
    ctx := context.Background()
    config := api.Config{}
    
    err := plugin.Initialize(ctx, config)
    require.NoError(t, err)
    
    // 测试获取参数
    params := plugin.GetParameters()
    assert.Len(t, params, 4)
    
    paramNames := make(map[string]bool)
    for _, param := range params {
        paramNames[param.Name] = true
    }
    
    assert.True(t, paramNames["room_size"])
    assert.True(t, paramNames["damping"])
    assert.True(t, paramNames["wet_level"])
    assert.True(t, paramNames["dry_level"])
    
    // 测试设置参数
    err = plugin.SetParameter("room_size", 0.8)
    assert.NoError(t, err)
    
    value, err := plugin.GetParameter("room_size")
    require.NoError(t, err)
    assert.Equal(t, 0.8, value)
    
    // 测试无效参数
    err = plugin.SetParameter("invalid_param", 0.5)
    assert.Error(t, err)
    
    // 测试参数范围验证
    err = plugin.SetParameter("room_size", 1.5) // 超出范围
    assert.Error(t, err)
    
    plugin.Cleanup()
}

func TestReverbProcessor_StreamProcessing(t *testing.T) {
    plugin := processor.NewReverbProcessor()
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    config := api.Config{
        "room_size": 0.5,
        "wet_level": 0.3,
    }
    
    err := plugin.Initialize(ctx, config)
    require.NoError(t, err)
    
    err = plugin.Start(ctx)
    require.NoError(t, err)
    
    // 设置音频格式
    format := api.AudioFormat{
        SampleRate: 44100,
        Channels:   2,
        BitDepth:   32,
        Format:     "float32",
    }
    err = plugin.SetAudioFormat(format)
    require.NoError(t, err)
    
    // 创建输入输出通道
    input := make(chan *api.AudioBuffer, 10)
    output := make(chan *api.AudioBuffer, 10)
    
    // 启动流处理
    go func() {
        err := plugin.ProcessAudioStream(input, output)
        if err != nil && err != context.Canceled {
            t.Errorf("Stream processing failed: %v", err)
        }
    }()
    
    // 发送测试数据
    for i := 0; i < 5; i++ {
        testBuffer := &api.AudioBuffer{
            Data: [][]float32{
                {float32(i) * 0.1, float32(i) * 0.2},
                {float32(i) * 0.1, float32(i) * 0.2},
            },
            SampleRate: 44100,
            Channels:   2,
            Frames:     2,
            Timestamp:  time.Now(),
        }
        
        select {
        case input <- testBuffer:
        case <-ctx.Done():
            t.Fatal("Timeout sending input")
        }
    }
    
    // 接收处理结果
    for i := 0; i < 5; i++ {
        select {
        case result := <-output:
            assert.NotNil(t, result)
            assert.Equal(t, 2, result.Channels)
            assert.Equal(t, 2, result.Frames)
        case <-ctx.Done():
            t.Fatal("Timeout receiving output")
        }
    }
    
    close(input)
    plugin.Stop(ctx)
    plugin.Cleanup()
}

func BenchmarkReverbProcessor_ProcessAudio(b *testing.B) {
    plugin := processor.NewReverbProcessor()
    
    ctx := context.Background()
    config := api.Config{
        "room_size": 0.5,
        "wet_level": 0.3,
    }
    
    plugin.Initialize(ctx, config)
    plugin.Start(ctx)
    
    format := api.AudioFormat{
        SampleRate: 44100,
        Channels:   2,
        BitDepth:   32,
        Format:     "float32",
    }
    plugin.SetAudioFormat(format)
    
    // 创建较大的音频缓冲区进行基准测试
    frames := 1024
    testAudio := &api.AudioBuffer{
        Data: [][]float32{
            make([]float32, frames),
            make([]float32, frames),
        },
        SampleRate: 44100,
        Channels:   2,
        Frames:     frames,
        Timestamp:  time.Now(),
    }
    
    // 填充测试数据
    for ch := 0; ch < 2; ch++ {
        for i := 0; i < frames; i++ {
            testAudio.Data[ch][i] = float32(i) / float32(frames)
        }
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := plugin.ProcessAudio(testAudio)
        if err != nil {
            b.Fatalf("Audio processing failed: %v", err)
        }
    }
    
    plugin.Stop(ctx)
    plugin.Cleanup()
}
```

## 部署和调试

### 本地安装和测试

```bash
# 构建并安装插件
make install

# 启动 go-musicfox 并测试插件
go-musicfox --plugin-dir ~/.go-musicfox/plugins

# 查看插件日志
tail -f ~/.go-musicfox/logs/plugins.log
```

### 调试技巧

1. **使用日志记录**

```go
import (
    "log/slog"
    "os"
)

// 在插件中添加结构化日志
type ReverbProcessor struct {
    // ... 其他字段
    logger *slog.Logger
}

func NewReverbProcessor() *ReverbProcessor {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    }))
    
    return &ReverbProcessor{
        // ... 其他初始化
        logger: logger,
    }
}

func (r *ReverbProcessor) ProcessAudio(input *api.AudioBuffer) (*api.AudioBuffer, error) {
    r.logger.Debug("Processing audio",
        "channels", input.Channels,
        "frames", input.Frames,
        "sample_rate", input.SampleRate,
    )
    
    // ... 处理逻辑
    
    r.logger.Debug("Audio processing completed",
        "processing_time", time.Since(start),
    )
    
    return output, nil
}
```

2. **性能分析**

```go
// 添加性能监控
import (
    "runtime"
    "time"
)

func (r *ReverbProcessor) ProcessAudio(input *api.AudioBuffer) (*api.AudioBuffer, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        if duration > 10*time.Millisecond {
            r.logger.Warn("Slow audio processing detected",
                "duration", duration,
                "frames", input.Frames,
            )
        }
    }()
    
    // 内存使用监控
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    r.logger.Debug("Memory usage",
        "alloc", m.Alloc,
        "total_alloc", m.TotalAlloc,
        "sys", m.Sys,
    )
    
    // ... 处理逻辑
}
```

3. **错误处理和恢复**

```go
func (r *ReverbProcessor) ProcessAudio(input *api.AudioBuffer) (output *api.AudioBuffer, err error) {
    // 恢复 panic
    defer func() {
        if r := recover(); r != nil {
            r.logger.Error("Panic in audio processing",
                "panic", r,
                "stack", string(debug.Stack()),
            )
            err = fmt.Errorf("audio processing panic: %v", r)
        }
    }()
    
    // 输入验证
    if input == nil {
        return nil, fmt.Errorf("input buffer is nil")
    }
    
    if input.Channels == 0 || input.Frames == 0 {
        return nil, fmt.Errorf("invalid audio format: channels=%d, frames=%d", 
            input.Channels, input.Frames)
    }
    
    // ... 处理逻辑
    
    return output, nil
}
```

## 最佳实践

### 1. 性能优化

- **内存池使用**：重用音频缓冲区以减少 GC 压力
- **SIMD 优化**：对于密集计算使用向量化指令
- **并发处理**：合理使用 goroutine 处理多声道音频
- **缓存友好**：优化数据结构的内存布局

### 2. 错误处理

- **优雅降级**：在出错时提供备用处理方案
- **详细日志**：记录足够的上下文信息用于调试
- **资源清理**：确保在错误情况下正确释放资源

### 3. 配置管理

- **参数验证**：严格验证所有配置参数
- **热更新支持**：支持运行时配置更新
- **默认值**：为所有参数提供合理的默认值

### 4. 测试策略

- **单元测试**：测试所有公共方法和边界条件
- **集成测试**：测试与内核的交互
- **性能测试**：确保满足实时处理要求
- **压力测试**：测试长时间运行的稳定性

## 常见问题

### Q: 插件加载失败怎么办？

A: 检查以下几点：
1. 插件文件是否存在且有执行权限
2. 插件元数据文件是否正确
3. 依赖库是否完整
4. 查看详细的错误日志

### Q: 音频处理延迟过高怎么优化？

A: 优化建议：
1. 减少内存分配，使用对象池
2. 优化算法复杂度
3. 使用更小的缓冲区大小
4. 考虑使用 SIMD 指令

### Q: 如何调试插件崩溃问题？

A: 调试步骤：
1. 启用详细日志记录
2. 使用 race detector：`go build -race`
3. 添加 panic 恢复机制
4. 使用 pprof 进行性能分析

### Q: 插件如何与其他插件通信？

A: 通信方式：
1. 通过事件总线发送消息
2. 使用服务注册表共享服务
3. 通过内核提供的 API 进行交互

## 下一步

完成第一个插件后，您可以：

1. 阅读 [四种插件类型开发指南](dynamic-library.md)
2. 学习 [插件配置和部署指南](plugin-config.md)
3. 查看 [插件测试指南](plugin-testing.md)
4. 参考 [API 文档](../api/README.md)
5. 浏览 [示例项目](../examples/README.md)

## 相关资源

- [微内核架构概述](../architecture/microkernel.md)
- [插件系统设计](../architecture/plugin-system.md)
- [开发者工具文档](../tools/README.md)
- [社区论坛](https://github.com/go-musicfox/go-musicfox/discussions)
- [问题反馈](https://github.com/go-musicfox/go-musicfox/issues)