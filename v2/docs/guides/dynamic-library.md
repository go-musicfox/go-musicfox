# 动态库插件开发指南

动态库插件是 go-musicfox v2 中性能最高的插件类型，通过共享库（.so/.dll/.dylib）的形式加载到主程序中。本指南将详细介绍如何开发、构建和部署动态库插件。

## 目录

- [概述](#概述)
- [技术原理](#技术原理)
- [开发环境准备](#开发环境准备)
- [项目结构](#项目结构)
- [接口实现](#接口实现)
- [构建配置](#构建配置)
- [内存管理](#内存管理)
- [性能优化](#性能优化)
- [调试技巧](#调试技巧)
- [部署指南](#部署指南)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)

## 概述

### 什么是动态库插件

动态库插件是编译为共享库的 Go 程序，在运行时被主程序动态加载。它们与主程序运行在同一进程空间中，可以直接访问主程序的内存和资源。

### 优势

- **最高性能**：无进程间通信开销，直接内存访问
- **低延迟**：函数调用开销最小
- **完整访问**：可以访问所有系统资源和 Go 运行时
- **简单集成**：与主程序紧密集成，调用方式简单

### 劣势

- **安全风险**：插件崩溃可能导致主程序崩溃
- **平台依赖**：需要为每个平台编译不同的二进制文件
- **版本兼容**：Go 版本和依赖版本必须匹配
- **调试困难**：调试和错误隔离相对困难

### 适用场景

- 高性能音频处理
- 实时音频效果
- 系统级功能扩展
- 性能关键的算法实现
- 需要直接硬件访问的功能

## 技术原理

### Go Plugin 机制

Go 的 plugin 包提供了动态加载共享库的能力：

```go
import "plugin"

// 加载插件
p, err := plugin.Open("plugin.so")
if err != nil {
    return err
}

// 查找符号
sym, err := p.Lookup("PluginCreate")
if err != nil {
    return err
}

// 类型断言并调用
createFunc := sym.(func() interface{})
instance := createFunc()
```

### 符号导出

动态库插件必须导出特定的符号供主程序调用：

```go
// 导出的全局变量
var PluginInstance MyPlugin

// 导出的函数
func PluginCreate() interface{} {
    return &PluginInstance
}

func PluginDestroy() {
    PluginInstance.Cleanup()
}
```

### 内存共享

动态库插件与主程序共享内存空间，可以直接传递指针：

```go
// 直接传递结构体指针
func ProcessAudio(input *AudioBuffer) *AudioBuffer {
    // 直接操作内存，无需序列化
    return processAudioData(input)
}
```

## 开发环境准备

### 系统要求

```bash
# Go 版本要求
go version  # 需要 1.21+

# 构建工具
sudo apt-get install build-essential  # Linux
# 或
brew install gcc  # macOS
```

### 项目初始化

```bash
# 创建项目目录
mkdir audio-enhancer-plugin
cd audio-enhancer-plugin

# 初始化 Go 模块
go mod init github.com/your-username/audio-enhancer-plugin

# 创建目录结构
mkdir -p {cmd/plugin,internal/{processor,config},pkg/api,configs,tests,docs}
```

### 依赖管理

```go
// go.mod
module github.com/your-username/audio-enhancer-plugin

go 1.21

require (
    github.com/go-musicfox/kernel v2.0.0
    github.com/stretchr/testify v1.8.4
)

require (
    github.com/davecgh/go-spew v1.1.1 // indirect
    github.com/pmezard/go-difflib v1.0.0 // indirect
    gopkg.in/yaml.v3 v3.0.1 // indirect
)
```

## 项目结构

### 标准目录布局

```
audio-enhancer-plugin/
├── cmd/
│   └── plugin/
│       ├── main.go              # 插件入口点
│       └── exports.go           # 导出函数定义
├── internal/
│   ├── processor/
│   │   ├── enhancer.go          # 核心处理逻辑
│   │   ├── algorithms.go        # 算法实现
│   │   └── filters.go           # 滤波器实现
│   ├── config/
│   │   ├── config.go            # 配置管理
│   │   └── validation.go        # 配置验证
│   └── utils/
│       ├── memory.go            # 内存管理工具
│       └── math.go              # 数学工具函数
├── pkg/
│   └── api/
│       ├── interfaces.go        # 公共接口定义
│       ├── types.go             # 数据类型定义
│       └── constants.go         # 常量定义
├── configs/
│   ├── plugin.yaml              # 插件配置
│   └── presets/                 # 预设配置
│       ├── vocal.yaml
│       ├── instrumental.yaml
│       └── bass-boost.yaml
├── tests/
│   ├── unit/
│   │   └── processor_test.go    # 单元测试
│   ├── integration/
│   │   └── plugin_test.go       # 集成测试
│   └── benchmarks/
│       └── performance_test.go  # 性能测试
├── docs/
│   ├── README.md                # 插件文档
│   ├── API.md                   # API 文档
│   └── CHANGELOG.md             # 变更日志
├── scripts/
│   ├── build.sh                 # 构建脚本
│   ├── test.sh                  # 测试脚本
│   └── install.sh               # 安装脚本
├── .gitignore
├── .golangci.yml                # 代码检查配置
├── Makefile                     # 构建配置
├── plugin.json                  # 插件元数据
├── go.mod
└── go.sum
```

## 接口实现

### 核心插件接口

```go
// pkg/api/interfaces.go
package api

import (
    "context"
    "time"
)

// AudioEnhancer 音频增强插件接口
type AudioEnhancer interface {
    Plugin
    AudioProcessor
    
    // 增强功能
    EnhanceAudio(input *AudioBuffer, preset string) (*AudioBuffer, error)
    GetPresets() []PresetInfo
    LoadPreset(name string) error
    SavePreset(name string, config EnhancerConfig) error
    
    // 实时处理
    StartRealTimeProcessing(config RealTimeConfig) error
    StopRealTimeProcessing() error
    
    // 分析功能
    AnalyzeAudio(input *AudioBuffer) (*AudioAnalysis, error)
    GetSpectrum(input *AudioBuffer) (*SpectrumData, error)
}

// EnhancerConfig 增强器配置
type EnhancerConfig struct {
    // 均衡器设置
    Equalizer EqualizerConfig `json:"equalizer"`
    
    // 压缩器设置
    Compressor CompressorConfig `json:"compressor"`
    
    // 限制器设置
    Limiter LimiterConfig `json:"limiter"`
    
    // 立体声增强
    StereoEnhancer StereoEnhancerConfig `json:"stereo_enhancer"`
    
    // 低音增强
    BassBoost BassBoostConfig `json:"bass_boost"`
    
    // 高音增强
    TrebleBoost TrebleBoostConfig `json:"treble_boost"`
    
    // 总体设置
    MasterGain   float64 `json:"master_gain"`
    BypassAll    bool    `json:"bypass_all"`
    ProcessingQuality string `json:"processing_quality"` // "draft", "good", "best"
}

type EqualizerConfig struct {
    Enabled bool      `json:"enabled"`
    Bands   []EQBand  `json:"bands"`
}

type EQBand struct {
    Frequency float64 `json:"frequency"` // Hz
    Gain      float64 `json:"gain"`      // dB
    Q         float64 `json:"q"`         // Quality factor
    Type      string  `json:"type"`      // "peak", "lowpass", "highpass", "lowshelf", "highshelf"
}

type CompressorConfig struct {
    Enabled     bool    `json:"enabled"`
    Threshold   float64 `json:"threshold"`   // dB
    Ratio       float64 `json:"ratio"`       // 1:ratio
    Attack      float64 `json:"attack"`      // ms
    Release     float64 `json:"release"`     // ms
    Knee        float64 `json:"knee"`        // dB
    MakeupGain  float64 `json:"makeup_gain"` // dB
}

type AudioAnalysis struct {
    RMS          float64           `json:"rms"`           // RMS level
    Peak         float64           `json:"peak"`          // Peak level
    DynamicRange float64           `json:"dynamic_range"` // dB
    Spectrum     []float64         `json:"spectrum"`      // Frequency spectrum
    Features     map[string]float64 `json:"features"`     // Additional features
    Timestamp    time.Time         `json:"timestamp"`
}
```

### 核心处理器实现

```go
// internal/processor/enhancer.go
package processor

import (
    "context"
    "fmt"
    "math"
    "sync"
    "time"
    
    "github.com/your-username/audio-enhancer-plugin/pkg/api"
)

type AudioEnhancerProcessor struct {
    // 基本信息
    id          string
    name        string
    version     string
    description string
    
    // 状态管理
    status      api.PluginStatus
    config      api.Config
    enhancerConfig api.EnhancerConfig
    mutex       sync.RWMutex
    
    // 音频处理组件
    equalizer      *Equalizer
    compressor     *Compressor
    limiter        *Limiter
    stereoEnhancer *StereoEnhancer
    bassBoost      *BassBoost
    trebleBoost    *TrebleBoost
    
    // 分析器
    analyzer       *AudioAnalyzer
    spectrumAnalyzer *SpectrumAnalyzer
    
    // 音频格式
    sampleRate     int
    channels       int
    bufferSize     int
    
    // 内存池
    bufferPool     *BufferPool
    
    // 实时处理
    realTimeEnabled bool
    realTimeConfig  api.RealTimeConfig
    
    // 预设管理
    presets        map[string]api.EnhancerConfig
    currentPreset  string
    
    // 生命周期
    ctx            context.Context
    cancel         context.CancelFunc
    
    // 性能监控
    processingTime time.Duration
    processedFrames int64
    
    // 日志
    logger         Logger
}

// 创建新的音频增强处理器
func NewAudioEnhancerProcessor() *AudioEnhancerProcessor {
    return &AudioEnhancerProcessor{
        id:          "audio-enhancer",
        name:        "Audio Enhancer",
        version:     "1.0.0",
        description: "Professional audio enhancement plugin with EQ, compression, and stereo enhancement",
        status:      api.StatusUnknown,
        
        // 初始化组件
        equalizer:      NewEqualizer(),
        compressor:     NewCompressor(),
        limiter:        NewLimiter(),
        stereoEnhancer: NewStereoEnhancer(),
        bassBoost:      NewBassBoost(),
        trebleBoost:    NewTrebleBoost(),
        analyzer:       NewAudioAnalyzer(),
        spectrumAnalyzer: NewSpectrumAnalyzer(),
        
        // 默认配置
        enhancerConfig: getDefaultEnhancerConfig(),
        presets:        loadDefaultPresets(),
        
        // 内存池
        bufferPool:     NewBufferPool(1024, 4096),
        
        logger:         NewLogger("audio-enhancer"),
    }
}

// 实现 Plugin 接口
func (a *AudioEnhancerProcessor) ID() string {
    return a.id
}

func (a *AudioEnhancerProcessor) Name() string {
    return a.name
}

func (a *AudioEnhancerProcessor) Version() string {
    return a.version
}

func (a *AudioEnhancerProcessor) Description() string {
    return a.description
}

func (a *AudioEnhancerProcessor) Initialize(ctx context.Context, config api.Config) error {
    a.mutex.Lock()
    defer a.mutex.Unlock()
    
    a.ctx, a.cancel = context.WithCancel(ctx)
    a.config = config
    
    // 解析增强器配置
    if enhancerConfig, ok := config["enhancer"].(api.EnhancerConfig); ok {
        a.enhancerConfig = enhancerConfig
    }
    
    // 初始化音频格式
    if sampleRate, ok := config["sample_rate"].(int); ok {
        a.sampleRate = sampleRate
    } else {
        a.sampleRate = 44100 // 默认采样率
    }
    
    if channels, ok := config["channels"].(int); ok {
        a.channels = channels
    } else {
        a.channels = 2 // 默认立体声
    }
    
    if bufferSize, ok := config["buffer_size"].(int); ok {
        a.bufferSize = bufferSize
    } else {
        a.bufferSize = 1024 // 默认缓冲区大小
    }
    
    // 初始化所有处理组件
    if err := a.initializeComponents(); err != nil {
        return fmt.Errorf("failed to initialize components: %w", err)
    }
    
    a.status = api.StatusInitializing
    a.logger.Info("Audio enhancer initialized",
        "sample_rate", a.sampleRate,
        "channels", a.channels,
        "buffer_size", a.bufferSize,
    )
    
    return nil
}

func (a *AudioEnhancerProcessor) Start(ctx context.Context) error {
    a.mutex.Lock()
    defer a.mutex.Unlock()
    
    if a.status != api.StatusInitializing {
        return fmt.Errorf("plugin not initialized")
    }
    
    // 启动所有组件
    if err := a.startComponents(); err != nil {
        return fmt.Errorf("failed to start components: %w", err)
    }
    
    a.status = api.StatusRunning
    a.logger.Info("Audio enhancer started")
    
    return nil
}

func (a *AudioEnhancerProcessor) Stop(ctx context.Context) error {
    a.mutex.Lock()
    defer a.mutex.Unlock()
    
    if a.cancel != nil {
        a.cancel()
    }
    
    // 停止实时处理
    if a.realTimeEnabled {
        a.stopRealTimeProcessing()
    }
    
    // 停止所有组件
    a.stopComponents()
    
    a.status = api.StatusStopped
    a.logger.Info("Audio enhancer stopped")
    
    return nil
}

func (a *AudioEnhancerProcessor) Cleanup() error {
    a.mutex.Lock()
    defer a.mutex.Unlock()
    
    // 清理所有组件
    a.cleanupComponents()
    
    // 清理内存池
    if a.bufferPool != nil {
        a.bufferPool.Cleanup()
    }
    
    a.logger.Info("Audio enhancer cleaned up")
    return nil
}

// 实现 AudioProcessor 接口
func (a *AudioEnhancerProcessor) ProcessAudio(input *api.AudioBuffer) (*api.AudioBuffer, error) {
    start := time.Now()
    defer func() {
        a.processingTime = time.Since(start)
        a.processedFrames += int64(input.Frames)
    }()
    
    a.mutex.RLock()
    defer a.mutex.RUnlock()
    
    if a.status != api.StatusRunning {
        return nil, fmt.Errorf("plugin not running")
    }
    
    // 验证输入
    if err := a.validateInput(input); err != nil {
        return nil, fmt.Errorf("invalid input: %w", err)
    }
    
    // 如果全部旁路，直接返回输入
    if a.enhancerConfig.BypassAll {
        return a.copyAudioBuffer(input), nil
    }
    
    // 从内存池获取输出缓冲区
    output := a.bufferPool.Get(input.Channels, input.Frames)
    defer a.bufferPool.Put(output)
    
    // 复制基本信息
    output.SampleRate = input.SampleRate
    output.Timestamp = input.Timestamp
    
    // 复制音频数据
    for ch := 0; ch < input.Channels; ch++ {
        copy(output.Data[ch], input.Data[ch])
    }
    
    // 应用音频处理链
    if err := a.applyProcessingChain(output); err != nil {
        return nil, fmt.Errorf("processing failed: %w", err)
    }
    
    // 应用主增益
    if a.enhancerConfig.MasterGain != 0.0 {
        gain := float32(math.Pow(10, a.enhancerConfig.MasterGain/20.0))
        a.applyGain(output, gain)
    }
    
    return a.copyAudioBuffer(output), nil
}

// 实现 AudioEnhancer 接口
func (a *AudioEnhancerProcessor) EnhanceAudio(input *api.AudioBuffer, preset string) (*api.AudioBuffer, error) {
    // 如果指定了预设，临时加载它
    if preset != "" && preset != a.currentPreset {
        if err := a.LoadPreset(preset); err != nil {
            return nil, fmt.Errorf("failed to load preset %s: %w", preset, err)
        }
    }
    
    return a.ProcessAudio(input)
}

func (a *AudioEnhancerProcessor) GetPresets() []api.PresetInfo {
    a.mutex.RLock()
    defer a.mutex.RUnlock()
    
    presets := make([]api.PresetInfo, 0, len(a.presets))
    for name, config := range a.presets {
        presets = append(presets, api.PresetInfo{
            Name:        name,
            Description: getPresetDescription(name),
            Category:    getPresetCategory(name),
            Tags:        getPresetTags(name),
            Config:      config,
        })
    }
    
    return presets
}

func (a *AudioEnhancerProcessor) LoadPreset(name string) error {
    a.mutex.Lock()
    defer a.mutex.Unlock()
    
    preset, exists := a.presets[name]
    if !exists {
        return fmt.Errorf("preset %s not found", name)
    }
    
    a.enhancerConfig = preset
    a.currentPreset = name
    
    // 更新所有组件的配置
    if err := a.updateComponentConfigs(); err != nil {
        return fmt.Errorf("failed to update component configs: %w", err)
    }
    
    a.logger.Info("Preset loaded", "name", name)
    return nil
}

func (a *AudioEnhancerProcessor) SavePreset(name string, config api.EnhancerConfig) error {
    a.mutex.Lock()
    defer a.mutex.Unlock()
    
    a.presets[name] = config
    
    // 保存到文件
    if err := a.savePresetToFile(name, config); err != nil {
        return fmt.Errorf("failed to save preset to file: %w", err)
    }
    
    a.logger.Info("Preset saved", "name", name)
    return nil
}

func (a *AudioEnhancerProcessor) AnalyzeAudio(input *api.AudioBuffer) (*api.AudioAnalysis, error) {
    a.mutex.RLock()
    defer a.mutex.RUnlock()
    
    if a.analyzer == nil {
        return nil, fmt.Errorf("analyzer not initialized")
    }
    
    return a.analyzer.Analyze(input)
}

func (a *AudioEnhancerProcessor) GetSpectrum(input *api.AudioBuffer) (*api.SpectrumData, error) {
    a.mutex.RLock()
    defer a.mutex.RUnlock()
    
    if a.spectrumAnalyzer == nil {
        return nil, fmt.Errorf("spectrum analyzer not initialized")
    }
    
    return a.spectrumAnalyzer.GetSpectrum(input)
}

// 私有方法
func (a *AudioEnhancerProcessor) initializeComponents() error {
    // 初始化均衡器
    if err := a.equalizer.Initialize(a.sampleRate, a.channels); err != nil {
        return fmt.Errorf("failed to initialize equalizer: %w", err)
    }
    
    // 初始化压缩器
    if err := a.compressor.Initialize(a.sampleRate, a.channels); err != nil {
        return fmt.Errorf("failed to initialize compressor: %w", err)
    }
    
    // 初始化限制器
    if err := a.limiter.Initialize(a.sampleRate, a.channels); err != nil {
        return fmt.Errorf("failed to initialize limiter: %w", err)
    }
    
    // 初始化立体声增强器
    if err := a.stereoEnhancer.Initialize(a.sampleRate, a.channels); err != nil {
        return fmt.Errorf("failed to initialize stereo enhancer: %w", err)
    }
    
    // 初始化低音增强
    if err := a.bassBoost.Initialize(a.sampleRate, a.channels); err != nil {
        return fmt.Errorf("failed to initialize bass boost: %w", err)
    }
    
    // 初始化高音增强
    if err := a.trebleBoost.Initialize(a.sampleRate, a.channels); err != nil {
        return fmt.Errorf("failed to initialize treble boost: %w", err)
    }
    
    // 初始化分析器
    if err := a.analyzer.Initialize(a.sampleRate, a.channels, a.bufferSize); err != nil {
        return fmt.Errorf("failed to initialize analyzer: %w", err)
    }
    
    // 初始化频谱分析器
    if err := a.spectrumAnalyzer.Initialize(a.sampleRate, a.channels, a.bufferSize); err != nil {
        return fmt.Errorf("failed to initialize spectrum analyzer: %w", err)
    }
    
    return nil
}

func (a *AudioEnhancerProcessor) applyProcessingChain(buffer *api.AudioBuffer) error {
    // 处理链顺序：EQ -> Compressor -> Bass Boost -> Treble Boost -> Stereo Enhancer -> Limiter
    
    // 1. 均衡器
    if a.enhancerConfig.Equalizer.Enabled {
        if err := a.equalizer.Process(buffer); err != nil {
            return fmt.Errorf("equalizer processing failed: %w", err)
        }
    }
    
    // 2. 压缩器
    if a.enhancerConfig.Compressor.Enabled {
        if err := a.compressor.Process(buffer); err != nil {
            return fmt.Errorf("compressor processing failed: %w", err)
        }
    }
    
    // 3. 低音增强
    if a.enhancerConfig.BassBoost.Enabled {
        if err := a.bassBoost.Process(buffer); err != nil {
            return fmt.Errorf("bass boost processing failed: %w", err)
        }
    }
    
    // 4. 高音增强
    if a.enhancerConfig.TrebleBoost.Enabled {
        if err := a.trebleBoost.Process(buffer); err != nil {
            return fmt.Errorf("treble boost processing failed: %w", err)
        }
    }
    
    // 5. 立体声增强（仅立体声）
    if a.channels == 2 && a.enhancerConfig.StereoEnhancer.Enabled {
        if err := a.stereoEnhancer.Process(buffer); err != nil {
            return fmt.Errorf("stereo enhancer processing failed: %w", err)
        }
    }
    
    // 6. 限制器（最后处理，防止削波）
    if a.enhancerConfig.Limiter.Enabled {
        if err := a.limiter.Process(buffer); err != nil {
            return fmt.Errorf("limiter processing failed: %w", err)
        }
    }
    
    return nil
}

func (a *AudioEnhancerProcessor) validateInput(input *api.AudioBuffer) error {
    if input == nil {
        return fmt.Errorf("input buffer is nil")
    }
    
    if input.Channels != a.channels {
        return fmt.Errorf("channel count mismatch: expected %d, got %d", a.channels, input.Channels)
    }
    
    if input.SampleRate != a.sampleRate {
        return fmt.Errorf("sample rate mismatch: expected %d, got %d", a.sampleRate, input.SampleRate)
    }
    
    if input.Frames <= 0 {
        return fmt.Errorf("invalid frame count: %d", input.Frames)
    }
    
    for ch := 0; ch < input.Channels; ch++ {
        if len(input.Data[ch]) != input.Frames {
            return fmt.Errorf("data length mismatch for channel %d: expected %d, got %d", 
                ch, input.Frames, len(input.Data[ch]))
        }
    }
    
    return nil
}

func (a *AudioEnhancerProcessor) copyAudioBuffer(src *api.AudioBuffer) *api.AudioBuffer {
    dst := &api.AudioBuffer{
        Data:       make([][]float32, src.Channels),
        SampleRate: src.SampleRate,
        Channels:   src.Channels,
        Frames:     src.Frames,
        Timestamp:  src.Timestamp,
    }
    
    for ch := 0; ch < src.Channels; ch++ {
        dst.Data[ch] = make([]float32, src.Frames)
        copy(dst.Data[ch], src.Data[ch])
    }
    
    return dst
}

func (a *AudioEnhancerProcessor) applyGain(buffer *api.AudioBuffer, gain float32) {
    for ch := 0; ch < buffer.Channels; ch++ {
        for i := 0; i < buffer.Frames; i++ {
            buffer.Data[ch][i] *= gain
        }
    }
}
```

### 导出函数定义

```go
// cmd/plugin/exports.go
package main

import (
    "C"
    "context"
    "encoding/json"
    "log"
    "unsafe"
    
    "github.com/your-username/audio-enhancer-plugin/internal/processor"
    "github.com/your-username/audio-enhancer-plugin/pkg/api"
)

// 全局插件实例
var pluginInstance api.AudioEnhancer

// 插件生命周期函数

//export PluginCreate
func PluginCreate() uintptr {
    pluginInstance = processor.NewAudioEnhancerProcessor()
    log.Printf("Plugin created: %s v%s", pluginInstance.Name(), pluginInstance.Version())
    return uintptr(unsafe.Pointer(&pluginInstance))
}

//export PluginDestroy
func PluginDestroy() {
    if pluginInstance != nil {
        pluginInstance.Cleanup()
        pluginInstance = nil
        log.Printf("Plugin destroyed")
    }
}

//export PluginInitialize
func PluginInitialize(configJSON *C.char) C.int {
    if pluginInstance == nil {
        log.Printf("Plugin not created")
        return -1
    }
    
    // 解析配置
    configStr := C.GoString(configJSON)
    var config api.Config
    if err := json.Unmarshal([]byte(configStr), &config); err != nil {
        log.Printf("Failed to parse config: %v", err)
        return -1
    }
    
    ctx := context.Background()
    if err := pluginInstance.Initialize(ctx, config); err != nil {
        log.Printf("Plugin initialization failed: %v", err)
        return -1
    }
    
    log.Printf("Plugin initialized successfully")
    return 0
}

//export PluginStart
func PluginStart() C.int {
    if pluginInstance == nil {
        log.Printf("Plugin not created")
        return -1
    }
    
    ctx := context.Background()
    if err := pluginInstance.Start(ctx); err != nil {
        log.Printf("Plugin start failed: %v", err)
        return -1
    }
    
    log.Printf("Plugin started successfully")
    return 0
}

//export PluginStop
func PluginStop() C.int {
    if pluginInstance == nil {
        log.Printf("Plugin not created")
        return -1
    }
    
    ctx := context.Background()
    if err := pluginInstance.Stop(ctx); err != nil {
        log.Printf("Plugin stop failed: %v", err)
        return -1
    }
    
    log.Printf("Plugin stopped successfully")
    return 0
}

// 音频处理函数

//export PluginProcessAudio
func PluginProcessAudio(inputPtr uintptr, outputPtr uintptr) C.int {
    if pluginInstance == nil {
        return -1
    }
    
    // 从指针转换为 AudioBuffer
    input := (*api.AudioBuffer)(unsafe.Pointer(inputPtr))
    if input == nil {
        log.Printf("Input buffer is null")
        return -1
    }
    
    // 处理音频
    output, err := pluginInstance.ProcessAudio(input)
    if err != nil {
        log.Printf("Audio processing failed: %v", err)
        return -1
    }
    
    // 将输出写入到指定位置
    outputBuffer := (*api.AudioBuffer)(unsafe.Pointer(outputPtr))
    *outputBuffer = *output
    
    return 0
}

//export PluginEnhanceAudio
func PluginEnhanceAudio(inputPtr uintptr, preset *C.char, outputPtr uintptr) C.int {
    if pluginInstance == nil {
        return -1
    }
    
    input := (*api.AudioBuffer)(unsafe.Pointer(inputPtr))
    if input == nil {
        return -1
    }
    
    presetName := C.GoString(preset)
    
    output, err := pluginInstance.EnhanceAudio(input, presetName)
    if err != nil {
        log.Printf("Audio enhancement failed: %v", err)
        return -1
    }
    
    outputBuffer := (*api.AudioBuffer)(unsafe.Pointer(outputPtr))
    *outputBuffer = *output
    
    return 0
}

// 配置管理函数

//export PluginGetPresets
func PluginGetPresets(presetsPtr uintptr, count *C.int) C.int {
    if pluginInstance == nil {
        return -1
    }
    
    presets := pluginInstance.GetPresets()
    *count = C.int(len(presets))
    
    // 将预设信息写入到指定位置
    // 注意：这里需要根据实际的内存布局来实现
    // 简化示例，实际实现需要更复杂的内存管理
    
    return 0
}

//export PluginLoadPreset
func PluginLoadPreset(name *C.char) C.int {
    if pluginInstance == nil {
        return -1
    }
    
    presetName := C.GoString(name)
    if err := pluginInstance.LoadPreset(presetName); err != nil {
        log.Printf("Failed to load preset %s: %v", presetName, err)
        return -1
    }
    
    return 0
}

// 分析函数

//export PluginAnalyzeAudio
func PluginAnalyzeAudio(inputPtr uintptr, analysisPtr uintptr) C.int {
    if pluginInstance == nil {
        return -1
    }
    
    input := (*api.AudioBuffer)(unsafe.Pointer(inputPtr))
    if input == nil {
        return -1
    }
    
    analysis, err := pluginInstance.AnalyzeAudio(input)
    if err != nil {
        log.Printf("Audio analysis failed: %v", err)
        return -1
    }
    
    analysisResult := (*api.AudioAnalysis)(unsafe.Pointer(analysisPtr))
    *analysisResult = *analysis
    
    return 0
}

// 工具函数

//export PluginGetMetadata
func PluginGetMetadata(metadataPtr uintptr) C.int {
    if pluginInstance == nil {
        return -1
    }
    
    metadata := pluginInstance.GetMetadata()
    metadataResult := (*api.PluginMetadata)(unsafe.Pointer(metadataPtr))
    *metadataResult = metadata
    
    return 0
}

//export PluginHealthCheck
func PluginHealthCheck() C.int {
    if pluginInstance == nil {
        return -1
    }
    
    if err := pluginInstance.HealthCheck(); err != nil {
        log.Printf("Health check failed: %v", err)
        return -1
    }
    
    return 0
}

//export PluginGetStatus
func PluginGetStatus() C.int {
    if pluginInstance == nil {
        return int(api.StatusUnknown)
    }
    
    return int(pluginInstance.Status())
}
```

## 构建配置

### Makefile 配置

```makefile
# Makefile
.PHONY: build-plugin build-test test clean install package

# 变量定义
PLUGIN_NAME := audio-enhancer
VERSION := 1.0.0
BUILD_DIR := build
DIST_DIR := dist

# Go 构建参数
GO_VERSION := $(shell go version | cut -d' ' -f3)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)"
PLUGIN_FLAGS := -buildmode=plugin
OPTIMIZE_FLAGS := -gcflags="-N -l" # 调试版本
# OPTIMIZE_FLAGS := -ldflags="-s -w" # 发布版本

# 平台检测
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    PLUGIN_EXT := .so
    CGO_ENABLED := 1
endif
ifeq ($(UNAME_S),Darwin)
    PLUGIN_EXT := .so
    CGO_ENABLED := 1
endif
ifeq ($(UNAME_S),Windows)
    PLUGIN_EXT := .dll
    CGO_ENABLED := 1
endif

# 默认目标
all: build-plugin

# 构建动态库插件
build-plugin:
	@echo "Building plugin for $(UNAME_S)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) go build $(LDFLAGS) $(PLUGIN_FLAGS) $(OPTIMIZE_FLAGS) \
		-o $(BUILD_DIR)/$(PLUGIN_NAME)$(PLUGIN_EXT) ./cmd/plugin
	@echo "Plugin built: $(BUILD_DIR)/$(PLUGIN_NAME)$(PLUGIN_EXT)"

# 构建测试可执行文件
build-test:
	@echo "Building test executable..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) $(OPTIMIZE_FLAGS) -o $(BUILD_DIR)/$(PLUGIN_NAME)-test ./cmd/plugin

# 交叉编译
build-linux:
	@echo "Cross-compiling for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-linux-gnu-gcc \
		go build $(LDFLAGS) $(PLUGIN_FLAGS) -o $(BUILD_DIR)/$(PLUGIN_NAME)-linux.so ./cmd/plugin

build-darwin:
	@echo "Cross-compiling for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 \
		go build $(LDFLAGS) $(PLUGIN_FLAGS) -o $(BUILD_DIR)/$(PLUGIN_NAME)-darwin.so ./cmd/plugin

build-windows:
	@echo "Cross-compiling for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
		go build $(LDFLAGS) $(PLUGIN_FLAGS) -o $(BUILD_DIR)/$(PLUGIN_NAME)-windows.dll ./cmd/plugin

# 构建所有平台
build-all: build-linux build-darwin build-windows

# 运行测试
test:
	@echo "Running tests..."
	go test -v -race ./...

# 运行基准测试
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./...

# 代码检查
lint:
	@echo "Running linter..."
	golangci-lint run --config .golangci.yml

# 格式化代码
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# 清理构建文件
clean:
	@echo "Cleaning build files..."
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	rm -f *.prof

# 安装依赖
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	go mod verify

# 更新依赖
update-deps:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# 创建发布包
package: build-all
	@echo "Creating release package..."
	@mkdir -p $(DIST_DIR)
	# 创建各平台的压缩包
	tar -czf $(DIST_DIR)/$(PLUGIN_NAME)-$(VERSION)-linux-amd64.tar.gz \
		-C $(BUILD_DIR) $(PLUGIN_NAME)-linux.so
	tar -czf $(DIST_DIR)/$(PLUGIN_NAME)-$(VERSION)-darwin-amd64.tar.gz \
		-C $(BUILD_DIR) $(PLUGIN_NAME)-darwin.so
	zip -j $(DIST_DIR)/$(PLUGIN_NAME)-$(VERSION)-windows-amd64.zip \
		$(BUILD_DIR)/$(PLUGIN_NAME)-windows.dll
	# 复制配置和文档
	cp plugin.json $(DIST_DIR)/
	cp -r configs $(DIST_DIR)/
	cp -r docs $(DIST_DIR)/
	cp README.md $(DIST_DIR)/
	cp LICENSE $(DIST_DIR)/

# 本地安装
install: build-plugin
	@echo "Installing plugin locally..."
	mkdir -p ~/.go-musicfox/plugins/$(PLUGIN_NAME)
	cp $(BUILD_DIR)/$(PLUGIN_NAME)$(PLUGIN_EXT) ~/.go-musicfox/plugins/$(PLUGIN_NAME)/
	cp plugin.json ~/.go-musicfox/plugins/$(PLUGIN_NAME)/
	cp -r configs ~/.go-musicfox/plugins/$(PLUGIN_NAME)/
	@echo "Plugin installed to ~/.go-musicfox/plugins/$(PLUGIN_NAME)/"

# 卸载
uninstall:
	@echo "Uninstalling plugin..."
	rm -rf ~/.go-musicfox/plugins/$(PLUGIN_NAME)

# 开发模式（文件监听）
dev:
	@echo "Starting development mode..."
	air -c .air.toml

# 性能分析
profile: benchmark
	@echo "Analyzing performance profiles..."
	go tool pprof -http=:8080 cpu.prof &
	go tool pprof -http=:8081 mem.prof &

# 代码覆盖率
coverage:
	@echo "Generating code coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 安全扫描
security:
	@echo "Running security scan..."
	gosec ./...

# 依赖漏洞检查
vuln-check:
	@echo "Checking for vulnerabilities..."
	govulncheck ./...

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  build-plugin    - Build plugin for current platform"
	@echo "  build-test      - Build test executable"
	@echo "  build-all       - Build for all platforms"
	@echo "  test            - Run unit tests"
	@echo "  benchmark       - Run benchmarks"
	@echo "  lint            - Run code linter"
	@echo "  fmt             - Format code"
	@echo "  clean           - Clean build files"
	@echo "  deps            - Install dependencies"
	@echo "  package         - Create release package"
	@echo "  install         - Install plugin locally"
	@echo "  uninstall       - Uninstall plugin"
	@echo "  dev             - Start development mode"
	@echo "  profile         - Analyze performance"
	@echo "  coverage        - Generate coverage report"
	@echo "  security        - Run security scan"
	@echo "  help            - Show this help"
```

### 构建脚本

```bash
#!/bin/bash
# scripts/build.sh

set -e

# 配置
PLUGIN_NAME="audio-enhancer"
VERSION="1.0.0"
BUILD_DIR="build"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 Go 环境
check_go_env() {
    log_info "Checking Go environment..."
    
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi
    
    GO_VERSION=$(go version | cut -d' ' -f3)
    log_info "Go version: $GO_VERSION"
    
    # 检查 Go 版本
    if [[ "$GO_VERSION" < "go1.21" ]]; then
        log_error "Go 1.21 or higher is required"
        exit 1
    fi
    
    # 检查 CGO
    if [[ "$CGO_ENABLED" != "1" ]]; then
        log_warn "CGO is not enabled, enabling it for plugin build"
        export CGO_ENABLED=1
    fi
}

# 检查依赖
check_dependencies() {
    log_info "Checking dependencies..."
    
    # 检查 C 编译器
    if ! command -v gcc &> /dev/null && ! command -v clang &> /dev/null; then
        log_error "C compiler (gcc or clang) is required for CGO"
        exit 1
    fi
    
    # 检查 Go 模块
    if [[ ! -f "go.mod" ]]; then
        log_error "go.mod not found, please run 'go mod init' first"
        exit 1
    fi
    
    # 下载依赖
    log_info "Downloading dependencies..."
    go mod download
    go mod tidy
}

# 运行测试
run_tests() {
    log_info "Running tests..."
    
    if ! go test -v ./...; then
        log_error "Tests failed"
        exit 1
    fi
    
    log_info "All tests passed"
}

# 构建插件
build_plugin() {
    local platform=$1
    local output_name=$2
    
    log_info "Building plugin for $platform..."
    
    mkdir -p "$BUILD_DIR"
    
    # 设置构建参数
    local ldflags="-ldflags='-X main.version=$VERSION -X main.buildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')'"
    local buildmode="-buildmode=plugin"
    
    # 根据平台设置环境变量
    case $platform in
        "linux")
            GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
                go build $ldflags $buildmode -o "$BUILD_DIR/$output_name" ./cmd/plugin
            ;;
        "darwin")
            GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 \
                go build $ldflags $buildmode -o "$BUILD_DIR/$output_name" ./cmd/plugin
            ;;
        "windows")
            GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
                go build $ldflags $buildmode -o "$BUILD_DIR/$output_name" ./cmd/plugin
            ;;
        "current")
            go build $ldflags $buildmode -o "$BUILD_DIR/$output_name" ./cmd/plugin
            ;;
        *)
            log_error "Unsupported platform: $platform"
            exit 1
            ;;
    esac
    
    if [[ $? -eq 0 ]]; then
        log_info "Plugin built successfully: $BUILD_DIR/$output_name"
    else
        log_error "Plugin build failed"
        exit 1
    fi
}

# 验证插件
validate_plugin() {
    local plugin_file=$1
    
    log_info "Validating plugin: $plugin_file"
    
    if [[ ! -f "$plugin_file" ]]; then
        log_error "Plugin file not found: $plugin_file"
        exit 1
    fi
    
    # 检查文件大小
    local file_size=$(stat -f%z "$plugin_file" 2>/dev/null || stat -c%s "$plugin_file" 2>/dev/null)
    if [[ $file_size -lt 1000 ]]; then
        log_error "Plugin file seems too small: $file_size bytes"
        exit 1
    fi
    
    log_info "Plugin validation passed"
}

# 主函数
main() {
    local platform="current"
    local run_tests_flag=true
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --platform)
                platform="$2"
                shift 2
                ;;
            --skip-tests)
                run_tests_flag=false
                shift
                ;;
            --help)
                echo "Usage: $0 [--platform current|linux|darwin|windows] [--skip-tests]"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    log_info "Starting build process..."
    
    # 执行构建步骤
    check_go_env
    check_dependencies
    
    if [[ "$run_tests_flag" == true ]]; then
        run_tests
    fi
    
    # 确定输出文件名
    local output_name
    case $platform in
        "linux")
            output_name="${PLUGIN_NAME}-linux.so"
            ;;
        "darwin")
            output_name="${PLUGIN_NAME}-darwin.so"
            ;;
        "windows")
            output_name="${PLUGIN_NAME}-windows.dll"
            ;;
        "current")
            case "$(uname -s)" in
                Linux)
                    output_name="${PLUGIN_NAME}.so"
                    ;;
                Darwin)
                    output_name="${PLUGIN_NAME}.so"
                    ;;
                MINGW*|CYGWIN*)
                    output_name="${PLUGIN_NAME}.dll"
                    ;;
                *)
                    log_error "Unsupported platform: $(uname -s)"
                    exit 1
                    ;;
            esac
            ;;
    esac
    
    build_plugin "$platform" "$output_name"
    validate_plugin "$BUILD_DIR/$output_name"
    
    log_info "Build completed successfully!"
    log_info "Plugin location: $BUILD_DIR/$output_name"
}

# 执行主函数
main "$@"
```

## 内存管理

### 内存池实现

```go
// internal/utils/memory.go
package utils

import (
    "sync"
    
    "github.com/your-username/audio-enhancer-plugin/pkg/api"
)

// BufferPool 音频缓冲区内存池
type BufferPool struct {
    pools   map[string]*sync.Pool // 按大小分组的池
    maxSize int                   // 最大缓冲区大小
    mutex   sync.RWMutex
}

// NewBufferPool 创建新的缓冲区池
func NewBufferPool(minSize, maxSize int) *BufferPool {
    return &BufferPool{
        pools:   make(map[string]*sync.Pool),
        maxSize: maxSize,
    }
}

// Get 从池中获取缓冲区
func (bp *BufferPool) Get(channels, frames int) *api.AudioBuffer {
    key := bp.getPoolKey(channels, frames)
    
    bp.mutex.RLock()
    pool, exists := bp.pools[key]
    bp.mutex.RUnlock()
    
    if !exists {
        bp.mutex.Lock()
        // 双重检查
        if pool, exists = bp.pools[key]; !exists {
            pool = &sync.Pool{
                New: func() interface{} {
                    return bp.createBuffer(channels, frames)
                },
            }
            bp.pools[key] = pool
        }
        bp.mutex.Unlock()
    }
    
    buffer := pool.Get().(*api.AudioBuffer)
    
    // 重置缓冲区
    bp.resetBuffer(buffer, channels, frames)
    
    return buffer
}

// Put 将缓冲区返回到池中
func (bp *BufferPool) Put(buffer *api.AudioBuffer) {
    if buffer == nil {
        return
    }
    
    // 检查缓冲区大小是否超过限制
    totalSize := buffer.Channels * buffer.Frames
    if totalSize > bp.maxSize {
        return // 不回收过大的缓冲区
    }
    
    key := bp.getPoolKey(buffer.Channels, buffer.Frames)
    
    bp.mutex.RLock()
    pool, exists := bp.pools[key]
    bp.mutex.RUnlock()
    
    if exists {
        pool.Put(buffer)
    }
}

// Cleanup 清理内存池
func (bp *BufferPool) Cleanup() {
    bp.mutex.Lock()
    defer bp.mutex.Unlock()
    
    bp.pools = make(map[string]*sync.Pool)
}

// 私有方法
func (bp *BufferPool) getPoolKey(channels, frames int) string {
    return fmt.Sprintf("%d_%d", channels, frames)
}

func (bp *BufferPool) createBuffer(channels, frames int) *api.AudioBuffer {
    buffer := &api.AudioBuffer{
        Data:     make([][]float32, channels),
        Channels: channels,
        Frames:   frames,
    }
    
    for ch := 0; ch < channels; ch++ {
        buffer.Data[ch] = make([]float32, frames)
    }
    
    return buffer
}

func (bp *BufferPool) resetBuffer(buffer *api.AudioBuffer, channels, frames int) {
    buffer.Channels = channels
    buffer.Frames = frames
    buffer.SampleRate = 0
    buffer.Timestamp = time.Time{}
    
    // 清零音频数据
    for ch := 0; ch < channels; ch++ {
        for i := 0; i < frames; i++ {
            buffer.Data[ch][i] = 0.0
        }
    }
}
```

## 性能优化

### SIMD 优化

```go
// internal/utils/simd.go
// +build amd64

package utils

import (
    "unsafe"
)

// 使用 SIMD 指令进行向量化计算
// 注意：这需要汇编代码或使用 CGO

// VectorAdd 向量加法（SIMD 优化）
func VectorAdd(a, b, result []float32) {
    if len(a) != len(b) || len(a) != len(result) {
        panic("vector lengths must match")
    }
    
    // 对齐到 4 的倍数进行 SIMD 处理
    simdLen := len(a) & ^3
    
    // SIMD 处理（这里简化，实际需要汇编实现）
    for i := 0; i < simdLen; i += 4 {
        // 模拟 SIMD 操作
        result[i] = a[i] + b[i]
        result[i+1] = a[i+1] + b[i+1]
        result[i+2] = a[i+2] + b[i+2]
        result[i+3] = a[i+3] + b[i+3]
    }
    
    // 处理剩余元素
    for i := simdLen; i < len(a); i++ {
        result[i] = a[i] + b[i]
    }
}

// VectorMultiply 向量乘法（SIMD 优化）
func VectorMultiply(a, b, result []float32) {
    if len(a) != len(b) || len(a) != len(result) {
        panic("vector lengths must match")
    }
    
    simdLen := len(a) & ^3
    
    for i := 0; i < simdLen; i += 4 {
        result[i] = a[i] * b[i]
        result[i+1] = a[i+1] * b[i+1]
        result[i+2] = a[i+2] * b[i+2]
        result[i+3] = a[i+3] * b[i+3]
    }
    
    for i := simdLen; i < len(a); i++ {
        result[i] = a[i] * b[i]
    }
}
```

### 并发处理

```go
// internal/processor/parallel.go
package processor

import (
    "runtime"
    "sync"
    
    "github.com/your-username/audio-enhancer-plugin/pkg/api"
)

// ParallelProcessor 并行音频处理器
type ParallelProcessor struct {
    numWorkers int
    workerPool chan chan *api.AudioBuffer
    workers    []*Worker
    wg         sync.WaitGroup
    quit       chan bool
}

// Worker 工作协程
type Worker struct {
    id         int
    workerPool chan chan *api.AudioBuffer
    jobChannel chan *api.AudioBuffer
    processor  AudioProcessor
    quit       chan bool
}

// NewParallelProcessor 创建并行处理器
func NewParallelProcessor(numWorkers int, processor AudioProcessor) *ParallelProcessor {
    if numWorkers <= 0 {
        numWorkers = runtime.NumCPU()
    }
    
    pp := &ParallelProcessor{
        numWorkers: numWorkers,
        workerPool: make(chan chan *api.AudioBuffer, numWorkers),
        quit:       make(chan bool),
    }
    
    // 创建工作协程
    pp.workers = make([]*Worker, numWorkers)
    for i := 0; i < numWorkers; i++ {
        worker := &Worker{
            id:         i,
            workerPool: pp.workerPool,
            jobChannel: make(chan *api.AudioBuffer),
            processor:  processor,
            quit:       make(chan bool),
        }
        pp.workers[i] = worker
    }
    
    return pp
}

// Start 启动并行处理器
func (pp *ParallelProcessor) Start() {
    for _, worker := range pp.workers {
        worker.Start()
    }
}

// Stop 停止并行处理器
func (pp *ParallelProcessor) Stop() {
    for _, worker := range pp.workers {
        worker.Stop()
    }
    close(pp.quit)
}

// ProcessAudioParallel 并行处理音频
func (pp *ParallelProcessor) ProcessAudioParallel(buffers []*api.AudioBuffer) ([]*api.AudioBuffer, error) {
    if len(buffers) == 0 {
        return nil, nil
    }
    
    results := make([]*api.AudioBuffer, len(buffers))
    resultChan := make(chan struct {
        index int
        buffer *api.AudioBuffer
        err    error
    }, len(buffers))
    
    // 分发任务
    for i, buffer := range buffers {
        go func(index int, buf *api.AudioBuffer) {
            // 获取可用的工作协程
            select {
            case jobChannel := <-pp.workerPool:
                jobChannel <- buf
                
                // 等待处理结果（简化实现）
                // 实际实现需要更复杂的结果收集机制
                result := buf // 假设处理完成
                resultChan <- struct {
                    index int
                    buffer *api.AudioBuffer
                    err    error
                }{index, result, nil}
                
            case <-pp.quit:
                resultChan <- struct {
                    index int
                    buffer *api.AudioBuffer
                    err    error
                }{index, nil, fmt.Errorf("processor stopped")}
            }
        }(i, buffer)
    }
    
    // 收集结果
    for i := 0; i < len(buffers); i++ {
        result := <-resultChan
        if result.err != nil {
            return nil, result.err
        }
        results[result.index] = result.buffer
    }
    
    return results, nil
}

// Worker 方法
func (w *Worker) Start() {
    go func() {
        for {
            // 将工作通道注册到池中
            w.workerPool <- w.jobChannel
            
            select {
            case job := <-w.jobChannel:
                // 处理音频
                w.processor.ProcessAudio(job)
                
            case <-w.quit:
                return
            }
        }
    }()
}

func (w *Worker) Stop() {
    go func() {
        w.quit <- true
    }()
}
```

## 调试技巧

### 1. 日志记录

```go
// internal/utils/logger.go
package utils

import (
    "fmt"
    "log/slog"
    "os"
    "path/filepath"
    "time"
)

type Logger struct {
    *slog.Logger
    component string
}

func NewLogger(component string) Logger {
    // 创建日志文件
    logDir := filepath.Join(os.TempDir(), "go-musicfox", "plugins")
    os.MkdirAll(logDir, 0755)
    
    logFile := filepath.Join(logDir, fmt.Sprintf("%s.log", component))
    file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        file = os.Stdout
    }
    
    handler := slog.NewJSONHandler(file, &slog.HandlerOptions{
        Level: slog.LevelDebug,
        AddSource: true,
    })
    
    logger := slog.New(handler)
    
    return Logger{
        Logger:    logger,
        component: component,
    }
}

func (l Logger) WithContext(ctx ...interface{}) Logger {
    args := []interface{}{"component", l.component, "timestamp", time.Now()}
    args = append(args, ctx...)
    
    return Logger{
        Logger:    l.Logger.With(args...),
        component: l.component,
    }
}
```

### 2. 性能监控

```go
// internal/utils/profiler.go
package utils

import (
    "runtime"
    "sync"
    "time"
)

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
    metrics map[string]*Metric
    mutex   sync.RWMutex
}

type Metric struct {
    Name         string
    Count        int64
    TotalTime    time.Duration
    MinTime      time.Duration
    MaxTime      time.Duration
    LastTime     time.Duration
    AverageTime  time.Duration
}

func NewPerformanceMonitor() *PerformanceMonitor {
    return &PerformanceMonitor{
        metrics: make(map[string]*Metric),
    }
}

func (pm *PerformanceMonitor) StartTimer(name string) func() {
    start := time.Now()
    
    return func() {
        duration := time.Since(start)
        pm.RecordMetric(name, duration)
    }
}

func (pm *PerformanceMonitor) RecordMetric(name string, duration time.Duration) {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()
    
    metric, exists := pm.metrics[name]
    if !exists {
        metric = &Metric{
            Name:    name,
            MinTime: duration,
            MaxTime: duration,
        }
        pm.metrics[name] = metric
    }
    
    metric.Count++
    metric.TotalTime += duration
    metric.LastTime = duration
    metric.AverageTime = metric.TotalTime / time.Duration(metric.Count)
    
    if duration < metric.MinTime {
        metric.MinTime = duration
    }
    if duration > metric.MaxTime {
        metric.MaxTime = duration
    }
}

func (pm *PerformanceMonitor) GetMetrics() map[string]*Metric {
    pm.mutex.RLock()
    defer pm.mutex.RUnlock()
    
    result := make(map[string]*Metric)
    for name, metric := range pm.metrics {
        result[name] = &Metric{
            Name:        metric.Name,
            Count:       metric.Count,
            TotalTime:   metric.TotalTime,
            MinTime:     metric.MinTime,
            MaxTime:     metric.MaxTime,
            LastTime:    metric.LastTime,
            AverageTime: metric.AverageTime,
        }
    }
    
    return result
}

func (pm *PerformanceMonitor) GetMemoryStats() runtime.MemStats {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    return m
}
```

## 部署指南

### 1. 插件打包

```bash
#!/bin/bash
# scripts/package.sh

set -e

PLUGIN_NAME="audio-enhancer"
VERSION="1.0.0"
DIST_DIR="dist"

# 创建发布目录
mkdir -p "$DIST_DIR"

# 构建所有平台
make build-all

# 创建插件包结构
for platform in linux darwin windows; do
    case $platform in
        "linux")
            ext=".so"
            ;;
        "darwin")
            ext=".so"
            ;;
        "windows")
            ext=".dll"
            ;;
    esac
    
    package_dir="$DIST_DIR/${PLUGIN_NAME}-${VERSION}-${platform}"
    mkdir -p "$package_dir"
    
    # 复制插件文件
    cp "build/${PLUGIN_NAME}-${platform}${ext}" "$package_dir/"
    
    # 复制配置文件
    cp plugin.json "$package_dir/"
    cp -r configs "$package_dir/"
    
    # 复制文档
    cp README.md "$package_dir/"
    cp LICENSE "$package_dir/"
    cp -r docs "$package_dir/"
    
    # 创建安装脚本
    cat > "$package_dir/install.sh" << EOF
#!/bin/bash
set -e

PLUGIN_DIR="\$HOME/.go-musicfox/plugins/$PLUGIN_NAME"

echo "Installing $PLUGIN_NAME plugin..."

# 创建插件目录
mkdir -p "\$PLUGIN_DIR"

# 复制文件
cp "${PLUGIN_NAME}-${platform}${ext}" "\$PLUGIN_DIR/${PLUGIN_NAME}${ext}"
cp plugin.json "\$PLUGIN_DIR/"
cp -r configs "\$PLUGIN_DIR/"

echo "Plugin installed successfully to \$PLUGIN_DIR"
EOF
    
    chmod +x "$package_dir/install.sh"
    
    # 创建压缩包
    if [[ "$platform" == "windows" ]]; then
        (cd "$DIST_DIR" && zip -r "${PLUGIN_NAME}-${VERSION}-${platform}.zip" "${PLUGIN_NAME}-${VERSION}-${platform}")
    else
        tar -czf "$DIST_DIR/${PLUGIN_NAME}-${VERSION}-${platform}.tar.gz" -C "$DIST_DIR" "${PLUGIN_NAME}-${VERSION}-${platform}"
    fi
    
    # 清理临时目录
    rm -rf "$package_dir"
done

echo "All packages created in $DIST_DIR"
```

### 2. 自动安装脚本

```bash
#!/bin/bash
# scripts/install.sh

set -e

# 配置
PLUGIN_NAME="audio-enhancer"
PLUGIN_DIR="$HOME/.go-musicfox/plugins/$PLUGIN_NAME"

# 检测平台
detect_platform() {
    case "$(uname -s)" in
        Linux)
            echo "linux"
            ;;
        Darwin)
            echo "darwin"
            ;;
        MINGW*|CYGWIN*)
            echo "windows"
            ;;
        *)
            echo "Unsupported platform: $(uname -s)" >&2
            exit 1
            ;;
    esac
}

# 检测架构
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)
            echo "amd64"
            ;;
        arm64|aarch64)
            echo "arm64"
            ;;
        *)
            echo "Unsupported architecture: $(uname -m)" >&2
            exit 1
            ;;
    esac
}

# 主安装函数
install_plugin() {
    local platform=$(detect_platform)
    local arch=$(detect_arch)
    
    echo "Installing $PLUGIN_NAME for $platform-$arch..."
    
    # 创建插件目录
    mkdir -p "$PLUGIN_DIR"
    
    # 确定文件扩展名
    local ext
    case $platform in
        "linux"|"darwin")
            ext=".so"
            ;;
        "windows")
            ext=".dll"
            ;;
    esac
    
    # 复制插件文件
    if [[ -f "build/${PLUGIN_NAME}${ext}" ]]; then
        cp "build/${PLUGIN_NAME}${ext}" "$PLUGIN_DIR/"
    elif [[ -f "build/${PLUGIN_NAME}-${platform}${ext}" ]]; then
        cp "build/${PLUGIN_NAME}-${platform}${ext}" "$PLUGIN_DIR/${PLUGIN_NAME}${ext}"
    else
        echo "Plugin binary not found. Please build the plugin first." >&2
        exit 1
    fi
    
    # 复制配置文件
    cp plugin.json "$PLUGIN_DIR/"
    
    if [[ -d "configs" ]]; then
        cp -r configs "$PLUGIN_DIR/"
    fi
    
    # 设置权限
    chmod +x "$PLUGIN_DIR/${PLUGIN_NAME}${ext}"
    
    echo "Plugin installed successfully to $PLUGIN_DIR"
    echo "You can now restart go-musicfox to load the plugin."
}

# 卸载函数
uninstall_plugin() {
    if [[ -d "$PLUGIN_DIR" ]]; then
        echo "Uninstalling $PLUGIN_NAME..."
        rm -rf "$PLUGIN_DIR"
        echo "Plugin uninstalled successfully."
    else
        echo "Plugin is not installed."
    fi
}

# 检查安装状态
check_installation() {
    if [[ -d "$PLUGIN_DIR" ]]; then
        echo "Plugin is installed at: $PLUGIN_DIR"
        
        # 显示插件信息
        if [[ -f "$PLUGIN_DIR/plugin.json" ]]; then
            echo "Plugin information:"
            cat "$PLUGIN_DIR/plugin.json" | grep -E '"(name|version|description)"' | sed 's/^/  /'
        fi
    else
        echo "Plugin is not installed."
    fi
}

# 解析命令行参数
case "${1:-install}" in
    "install")
        install_plugin
        ;;
    "uninstall")
        uninstall_plugin
        ;;
    "status")
        check_installation
        ;;
    "help")
        echo "Usage: $0 [install|uninstall|status|help]"
        echo "  install   - Install the plugin (default)"
        echo "  uninstall - Uninstall the plugin"
        echo "  status    - Check installation status"
        echo "  help      - Show this help message"
        ;;
    *)
        echo "Unknown command: $1" >&2
        echo "Use '$0 help' for usage information." >&2
        exit 1
        ;;
esac
```

## 最佳实践

### 1. 错误处理

- 使用 `recover()` 捕获 panic
- 提供详细的错误信息
- 实现优雅降级
- 记录错误日志

### 2. 内存管理

- 使用内存池减少 GC 压力
- 及时释放大对象
- 避免内存泄漏
- 监控内存使用

### 3. 性能优化

- 使用 SIMD 指令
- 实现并发处理
- 缓存计算结果
- 优化热点代码路径

### 4. 兼容性

- 保持 API 向后兼容
- 处理版本差异
- 测试多平台兼容性
- 文档化依赖要求

## 常见问题

### Q: 插件加载失败，提示符号未找到？

A: 检查导出函数名称和签名是否正确，确保使用了正确的 `//export` 注释。

### Q: 插件运行时崩溃？

A: 检查内存访问是否越界，使用 race detector 检测并发问题。

### Q: 性能不如预期？

A: 使用 pprof 进行性能分析，优化热点代码，考虑使用 SIMD 指令。

### Q: 如何调试动态库插件？

A: 可以先构建为可执行文件进行调试，然后再构建为动态库。

## 相关文档

- [插件开发快速入门](plugin-quickstart.md)
- [RPC 插件开发指南](rpc-plugin.md)
- [WebAssembly 插件开发指南](webassembly.md)
- [插件测试指南](plugin-testing.md)
- [API 文档](../api/README.md)