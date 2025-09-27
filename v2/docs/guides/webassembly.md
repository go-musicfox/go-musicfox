# WebAssembly 插件开发指南

WebAssembly (WASM) 插件为 go-musicfox v2 提供了跨平台、安全的插件执行环境。WASM 插件运行在沙箱环境中，提供了良好的安全性和可移植性。

## 目录

- [概述](#概述)
- [技术原理](#技术原理)
- [开发环境准备](#开发环境准备)
- [项目结构](#项目结构)
- [Go WASM 插件开发](#go-wasm-插件开发)
- [Rust WASM 插件开发](#rust-wasm-插件开发)
- [C/C++ WASM 插件开发](#cc-wasm-插件开发)
- [主机接口设计](#主机接口设计)
- [内存管理](#内存管理)
- [性能优化](#性能优化)
- [调试技巧](#调试技巧)
- [部署指南](#部署指南)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)

## 概述

### 什么是 WebAssembly 插件

WebAssembly 插件是编译为 WASM 字节码的程序，运行在 WebAssembly 虚拟机中。它们通过定义好的接口与主程序进行交互，提供了安全的沙箱执行环境。

### 优势

- **跨平台兼容**：一次编译，到处运行
- **安全沙箱**：隔离的执行环境，无法直接访问系统资源
- **多语言支持**：支持 Go、Rust、C/C++、AssemblyScript 等
- **体积小巧**：编译后的文件通常比较小
- **快速加载**：启动速度快，内存占用少

### 劣势

- **性能开销**：相比原生代码有一定性能损失
- **功能限制**：无法直接访问系统 API
- **调试困难**：调试工具相对有限
- **生态限制**：某些库可能不支持 WASM

### 适用场景

- 跨平台插件
- 安全要求高的场景
- 轻量级功能扩展
- 第三方开发者插件
- 实验性功能

## 技术原理

### WASM 运行时

go-musicfox v2 使用 Wasmtime 作为 WASM 运行时：

```go
// WASM 运行时管理器
type WASMRuntime struct {
    engine   *wasmtime.Engine
    store    *wasmtime.Store
    module   *wasmtime.Module
    instance *wasmtime.Instance
    
    // 导出函数
    exports  map[string]*wasmtime.Func
    
    // 主机函数
    imports  map[string]*wasmtime.Func
}
```

### 主机接口 (Host Interface)

主机程序向 WASM 模块提供的接口：

```go
// 主机接口定义
type HostInterface struct {
    // 日志接口
    Log func(level int32, message string)
    
    // 配置接口
    GetConfig func(key string) string
    SetConfig func(key, value string)
    
    // 网络接口
    HttpRequest func(method, url, body string) (int32, string)
    
    // 存储接口
    ReadFile func(path string) []byte
    WriteFile func(path string, data []byte) int32
    
    // 事件接口
    EmitEvent func(eventType, data string)
    SubscribeEvent func(eventType string) int32
}
```

### WASM 模块接口

WASM 模块需要导出的标准接口：

```go
// WASM 模块导出接口
type WASMExports struct {
    // 生命周期
    Initialize func(configPtr, configLen int32) int32
    Start      func() int32
    Stop       func() int32
    Cleanup    func() int32
    
    // 插件功能
    ProcessAudio func(inputPtr, inputLen int32, outputPtr int32) int32
    HandleEvent  func(eventPtr, eventLen int32) int32
    
    // 内存管理
    Malloc func(size int32) int32
    Free   func(ptr int32)
    
    // 元数据
    GetMetadata func() int32
    GetVersion  func() int32
}
```

## 开发环境准备

### 工具安装

```bash
# 安装 Wasmtime
curl https://wasmtime.dev/install.sh -sSf | bash

# 安装 WASI SDK (C/C++)
wget https://github.com/WebAssembly/wasi-sdk/releases/download/wasi-sdk-20/wasi-sdk-20.0-linux.tar.gz
tar xvf wasi-sdk-20.0-linux.tar.gz
export WASI_SDK_PATH=/path/to/wasi-sdk-20.0

# 安装 Rust WASM 工具链
rustup target add wasm32-wasi
cargo install wasm-pack

# 安装 Go WASM 工具
# Go 1.21+ 内置支持

# 安装 AssemblyScript
npm install -g assemblyscript
```

### 项目初始化

```bash
# 创建项目目录
mkdir audio-filter-wasm-plugin
cd audio-filter-wasm-plugin

# 创建多语言目录结构
mkdir -p {go,rust,c,assemblyscript}/{src,build}
mkdir -p {host,tests,docs,examples}
```

## 项目结构

```
audio-filter-wasm-plugin/
├── go/
│   ├── src/
│   │   ├── main.go              # Go WASM 插件源码
│   │   ├── audio.go             # 音频处理逻辑
│   │   └── host.go              # 主机接口封装
│   ├── build/
│   │   └── plugin.wasm          # 编译输出
│   └── go.mod
├── rust/
│   ├── src/
│   │   ├── lib.rs               # Rust WASM 插件源码
│   │   ├── audio.rs             # 音频处理模块
│   │   └── host.rs              # 主机接口绑定
│   ├── build/
│   │   └── plugin.wasm
│   └── Cargo.toml
├── c/
│   ├── src/
│   │   ├── main.c               # C WASM 插件源码
│   │   ├── audio.c              # 音频处理实现
│   │   └── host.h               # 主机接口头文件
│   ├── build/
│   │   └── plugin.wasm
│   └── Makefile
├── assemblyscript/
│   ├── src/
│   │   ├── index.ts             # AssemblyScript 源码
│   │   └── audio.ts             # 音频处理模块
│   ├── build/
│   │   └── plugin.wasm
│   └── package.json
├── host/
│   ├── runtime.go               # WASM 运行时封装
│   ├── interface.go             # 主机接口实现
│   └── loader.go                # 插件加载器
├── tests/
│   ├── integration/             # 集成测试
│   └── benchmarks/              # 性能测试
├── docs/
│   └── API.md                   # API 文档
├── examples/
│   └── simple-filter/           # 示例插件
├── plugin.json                  # 插件元数据
└── README.md
```

## Go WASM 插件开发

### 基础插件结构

```go
// go/src/main.go
package main

import (
    "encoding/json"
    "syscall/js"
    "unsafe"
)

// 插件元数据
type PluginMetadata struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Version     string   `json:"version"`
    Description string   `json:"description"`
    Author      string   `json:"author"`
    Tags        []string `json:"tags"`
}

// 音频缓冲区
type AudioBuffer struct {
    Data       [][]float32 `json:"data"`
    SampleRate int         `json:"sample_rate"`
    Channels   int         `json:"channels"`
    Frames     int         `json:"frames"`
}

// 全局状态
var (
    pluginConfig map[string]interface{}
    isRunning    bool
    filterParams FilterParams
)

type FilterParams struct {
    Cutoff    float32 `json:"cutoff"`
    Resonance float32 `json:"resonance"`
    Type      string  `json:"type"` // "lowpass", "highpass", "bandpass"
    Enabled   bool    `json:"enabled"`
}

// 主机接口函数声明
//go:wasmimport env host_log
func hostLog(level int32, msgPtr, msgLen int32)

//go:wasmimport env host_get_config
func hostGetConfig(keyPtr, keyLen int32) int32

//go:wasmimport env host_emit_event
func hostEmitEvent(typePtr, typeLen, dataPtr, dataLen int32)

// 导出函数

//export initialize
func initialize(configPtr, configLen int32) int32 {
    // 从内存读取配置
    configData := readMemory(configPtr, configLen)
    
    var config map[string]interface{}
    if err := json.Unmarshal(configData, &config); err != nil {
        logError("Failed to parse config: " + err.Error())
        return -1
    }
    
    pluginConfig = config
    
    // 解析滤波器参数
    if params, ok := config["filter_params"].(map[string]interface{}); ok {
        filterParams = FilterParams{
            Cutoff:    float32(params["cutoff"].(float64)),
            Resonance: float32(params["resonance"].(float64)),
            Type:      params["type"].(string),
            Enabled:   params["enabled"].(bool),
        }
    } else {
        // 默认参数
        filterParams = FilterParams{
            Cutoff:    1000.0,
            Resonance: 0.7,
            Type:      "lowpass",
            Enabled:   true,
        }
    }
    
    logInfo("Audio filter plugin initialized")
    return 0
}

//export start
func start() int32 {
    isRunning = true
    logInfo("Audio filter plugin started")
    return 0
}

//export stop
func stop() int32 {
    isRunning = false
    logInfo("Audio filter plugin stopped")
    return 0
}

//export cleanup
func cleanup() int32 {
    pluginConfig = nil
    isRunning = false
    logInfo("Audio filter plugin cleaned up")
    return 0
}

//export process_audio
func processAudio(inputPtr, inputLen int32, outputPtr int32) int32 {
    if !isRunning || !filterParams.Enabled {
        // 直接复制输入到输出
        copyAudioBuffer(inputPtr, inputLen, outputPtr)
        return 0
    }
    
    // 读取输入音频数据
    inputData := readMemory(inputPtr, inputLen)
    
    var audioBuffer AudioBuffer
    if err := json.Unmarshal(inputData, &audioBuffer); err != nil {
        logError("Failed to parse audio buffer: " + err.Error())
        return -1
    }
    
    // 应用音频滤波
    processedBuffer := applyAudioFilter(audioBuffer)
    
    // 序列化输出数据
    outputData, err := json.Marshal(processedBuffer)
    if err != nil {
        logError("Failed to serialize output: " + err.Error())
        return -1
    }
    
    // 写入输出内存
    writeMemory(outputPtr, outputData)
    
    return int32(len(outputData))
}

//export get_metadata
func getMetadata() int32 {
    metadata := PluginMetadata{
        ID:          "audio-filter-wasm",
        Name:        "Audio Filter WASM Plugin",
        Version:     "1.0.0",
        Description: "WebAssembly audio filter plugin with configurable parameters",
        Author:      "Your Name",
        Tags:        []string{"audio", "filter", "wasm"},
    }
    
    data, _ := json.Marshal(metadata)
    ptr := allocateMemory(int32(len(data)))
    writeMemory(ptr, data)
    
    return ptr
}

//export malloc
func malloc(size int32) int32 {
    return allocateMemory(size)
}

//export free
func free(ptr int32) {
    freeMemory(ptr)
}

// 音频处理函数
func applyAudioFilter(input AudioBuffer) AudioBuffer {
    output := AudioBuffer{
        Data:       make([][]float32, input.Channels),
        SampleRate: input.SampleRate,
        Channels:   input.Channels,
        Frames:     input.Frames,
    }
    
    // 为每个声道分配内存
    for ch := 0; ch < input.Channels; ch++ {
        output.Data[ch] = make([]float32, input.Frames)
    }
    
    // 应用滤波器到每个声道
    for ch := 0; ch < input.Channels; ch++ {
        applyFilterToChannel(input.Data[ch], output.Data[ch], input.SampleRate)
    }
    
    return output
}

func applyFilterToChannel(input, output []float32, sampleRate int) {
    // 简化的双二阶滤波器实现
    // 实际实现应该使用更复杂的滤波器算法
    
    nyquist := float32(sampleRate) / 2.0
    normalizedCutoff := filterParams.Cutoff / nyquist
    
    // 限制频率范围
    if normalizedCutoff > 0.99 {
        normalizedCutoff = 0.99
    }
    if normalizedCutoff < 0.01 {
        normalizedCutoff = 0.01
    }
    
    // 简单的一阶低通滤波器
    alpha := normalizedCutoff
    
    var prevOutput float32 = 0
    
    for i, sample := range input {
        switch filterParams.Type {
        case "lowpass":
            filtered := alpha*sample + (1-alpha)*prevOutput
            output[i] = filtered
            prevOutput = filtered
            
        case "highpass":
            filtered := alpha * (prevOutput + sample - input[max(0, i-1)])
            output[i] = filtered
            prevOutput = filtered
            
        default:
            output[i] = sample // 直通
        }
    }
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

// 内存管理函数
var memoryPool = make(map[int32][]byte)
var nextPtr int32 = 1024 // 从 1KB 开始分配

func allocateMemory(size int32) int32 {
    ptr := nextPtr
    nextPtr += size + 8 // 添加一些对齐空间
    
    memoryPool[ptr] = make([]byte, size)
    return ptr
}

func freeMemory(ptr int32) {
    delete(memoryPool, ptr)
}

func readMemory(ptr, length int32) []byte {
    if data, exists := memoryPool[ptr]; exists {
        if int32(len(data)) >= length {
            return data[:length]
        }
    }
    
    // 从 WASM 线性内存读取
    // 这里需要使用 unsafe 包直接访问内存
    return (*[1 << 30]byte)(unsafe.Pointer(uintptr(ptr)))[:length:length]
}

func writeMemory(ptr int32, data []byte) {
    if pool, exists := memoryPool[ptr]; exists {
        copy(pool, data)
    } else {
        // 直接写入 WASM 线性内存
        dst := (*[1 << 30]byte)(unsafe.Pointer(uintptr(ptr)))[:len(data):len(data)]
        copy(dst, data)
    }
}

func copyAudioBuffer(inputPtr, inputLen, outputPtr int32) {
    inputData := readMemory(inputPtr, inputLen)
    writeMemory(outputPtr, inputData)
}

// 日志函数
func logInfo(message string) {
    logMessage(0, message) // INFO level
}

func logError(message string) {
    logMessage(2, message) // ERROR level
}

func logMessage(level int32, message string) {
    msgBytes := []byte(message)
    ptr := allocateMemory(int32(len(msgBytes)))
    writeMemory(ptr, msgBytes)
    
    hostLog(level, ptr, int32(len(msgBytes)))
    
    freeMemory(ptr)
}

// 主函数（WASM 不需要，但保留用于测试）
func main() {
    // 在 WASM 环境中，这个函数不会被调用
    // 所有的交互都通过导出函数进行
}
```

### 构建脚本

```bash
#!/bin/bash
# go/build.sh

set -e

echo "Building Go WASM plugin..."

# 设置环境变量
export GOOS=wasip1
export GOARCH=wasm

# 构建 WASM 模块
go build -o build/plugin.wasm src/main.go src/audio.go src/host.go

# 优化 WASM 文件大小
if command -v wasm-opt &> /dev/null; then
    echo "Optimizing WASM file..."
    wasm-opt -Oz build/plugin.wasm -o build/plugin.wasm
fi

# 显示文件信息
echo "Build completed:"
ls -lh build/plugin.wasm

# 验证 WASM 文件
if command -v wasmtime &> /dev/null; then
    echo "Validating WASM file..."
    wasmtime --invoke initialize build/plugin.wasm
fi
```

## Rust WASM 插件开发

### Cargo.toml 配置

```toml
# rust/Cargo.toml
[package]
name = "audio-filter-wasm"
version = "1.0.0"
edition = "2021"

[lib]
crate-type = ["cdylib"]

[dependencies]
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
wasm-bindgen = "0.2"

[dependencies.web-sys]
version = "0.3"
features = [
  "console",
]

[profile.release]
opt-level = "s"  # 优化文件大小
lto = true       # 链接时优化
codegen-units = 1
panic = "abort"
```

### Rust 插件实现

```rust
// rust/src/lib.rs
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::ffi::{CStr, CString};
use std::os::raw::{c_char, c_int};
use wasm_bindgen::prelude::*;

// 插件元数据
#[derive(Serialize, Deserialize)]
struct PluginMetadata {
    id: String,
    name: String,
    version: String,
    description: String,
    author: String,
    tags: Vec<String>,
}

// 音频缓冲区
#[derive(Serialize, Deserialize)]
struct AudioBuffer {
    data: Vec<Vec<f32>>,
    sample_rate: i32,
    channels: i32,
    frames: i32,
}

// 滤波器参数
#[derive(Serialize, Deserialize)]
struct FilterParams {
    cutoff: f32,
    resonance: f32,
    filter_type: String,
    enabled: bool,
}

// 全局状态
static mut PLUGIN_CONFIG: Option<HashMap<String, serde_json::Value>> = None;
static mut IS_RUNNING: bool = false;
static mut FILTER_PARAMS: Option<FilterParams> = None;

// 主机接口声明
extern "C" {
    fn host_log(level: i32, msg_ptr: *const u8, msg_len: i32);
    fn host_get_config(key_ptr: *const u8, key_len: i32) -> i32;
    fn host_emit_event(type_ptr: *const u8, type_len: i32, data_ptr: *const u8, data_len: i32);
}

// 导出函数

#[no_mangle]
pub extern "C" fn initialize(config_ptr: *const u8, config_len: i32) -> i32 {
    unsafe {
        // 读取配置数据
        let config_slice = std::slice::from_raw_parts(config_ptr, config_len as usize);
        let config_str = match std::str::from_utf8(config_slice) {
            Ok(s) => s,
            Err(_) => {
                log_error("Invalid UTF-8 in config");
                return -1;
            }
        };
        
        // 解析配置
        let config: HashMap<String, serde_json::Value> = match serde_json::from_str(config_str) {
            Ok(c) => c,
            Err(e) => {
                log_error(&format!("Failed to parse config: {}", e));
                return -1;
            }
        };
        
        // 解析滤波器参数
        let filter_params = if let Some(params) = config.get("filter_params") {
            match serde_json::from_value(params.clone()) {
                Ok(p) => p,
                Err(_) => FilterParams {
                    cutoff: 1000.0,
                    resonance: 0.7,
                    filter_type: "lowpass".to_string(),
                    enabled: true,
                },
            }
        } else {
            FilterParams {
                cutoff: 1000.0,
                resonance: 0.7,
                filter_type: "lowpass".to_string(),
                enabled: true,
            }
        };
        
        PLUGIN_CONFIG = Some(config);
        FILTER_PARAMS = Some(filter_params);
        
        log_info("Audio filter plugin initialized");
        0
    }
}

#[no_mangle]
pub extern "C" fn start() -> i32 {
    unsafe {
        IS_RUNNING = true;
        log_info("Audio filter plugin started");
        0
    }
}

#[no_mangle]
pub extern "C" fn stop() -> i32 {
    unsafe {
        IS_RUNNING = false;
        log_info("Audio filter plugin stopped");
        0
    }
}

#[no_mangle]
pub extern "C" fn cleanup() -> i32 {
    unsafe {
        PLUGIN_CONFIG = None;
        FILTER_PARAMS = None;
        IS_RUNNING = false;
        log_info("Audio filter plugin cleaned up");
        0
    }
}

#[no_mangle]
pub extern "C" fn process_audio(input_ptr: *const u8, input_len: i32, output_ptr: *mut u8) -> i32 {
    unsafe {
        if !IS_RUNNING {
            return -1;
        }
        
        let filter_params = match &FILTER_PARAMS {
            Some(p) => p,
            None => return -1,
        };
        
        if !filter_params.enabled {
            // 直接复制输入到输出
            std::ptr::copy_nonoverlapping(input_ptr, output_ptr, input_len as usize);
            return input_len;
        }
        
        // 读取输入数据
        let input_slice = std::slice::from_raw_parts(input_ptr, input_len as usize);
        let input_str = match std::str::from_utf8(input_slice) {
            Ok(s) => s,
            Err(_) => {
                log_error("Invalid UTF-8 in input");
                return -1;
            }
        };
        
        // 解析音频缓冲区
        let mut audio_buffer: AudioBuffer = match serde_json::from_str(input_str) {
            Ok(b) => b,
            Err(e) => {
                log_error(&format!("Failed to parse audio buffer: {}", e));
                return -1;
            }
        };
        
        // 应用音频滤波
        apply_audio_filter(&mut audio_buffer, filter_params);
        
        // 序列化输出
        let output_str = match serde_json::to_string(&audio_buffer) {
            Ok(s) => s,
            Err(e) => {
                log_error(&format!("Failed to serialize output: {}", e));
                return -1;
            }
        };
        
        let output_bytes = output_str.as_bytes();
        let output_len = output_bytes.len();
        
        // 写入输出数据
        std::ptr::copy_nonoverlapping(output_bytes.as_ptr(), output_ptr, output_len);
        
        output_len as i32
    }
}

#[no_mangle]
pub extern "C" fn get_metadata() -> *const u8 {
    let metadata = PluginMetadata {
        id: "audio-filter-rust-wasm".to_string(),
        name: "Audio Filter Rust WASM Plugin".to_string(),
        version: "1.0.0".to_string(),
        description: "Rust WebAssembly audio filter plugin".to_string(),
        author: "Your Name".to_string(),
        tags: vec!["audio".to_string(), "filter".to_string(), "rust".to_string(), "wasm".to_string()],
    };
    
    let json_str = serde_json::to_string(&metadata).unwrap();
    let c_str = CString::new(json_str).unwrap();
    c_str.into_raw() as *const u8
}

// 音频处理函数
fn apply_audio_filter(buffer: &mut AudioBuffer, params: &FilterParams) {
    let nyquist = buffer.sample_rate as f32 / 2.0;
    let mut normalized_cutoff = params.cutoff / nyquist;
    
    // 限制频率范围
    normalized_cutoff = normalized_cutoff.clamp(0.01, 0.99);
    
    for channel in &mut buffer.data {
        apply_filter_to_channel(channel, normalized_cutoff, &params.filter_type);
    }
}

fn apply_filter_to_channel(samples: &mut [f32], cutoff: f32, filter_type: &str) {
    let alpha = cutoff;
    let mut prev_output = 0.0f32;
    let mut prev_input = 0.0f32;
    
    for i in 0..samples.len() {
        let input = samples[i];
        
        let output = match filter_type {
            "lowpass" => {
                let filtered = alpha * input + (1.0 - alpha) * prev_output;
                prev_output = filtered;
                filtered
            }
            "highpass" => {
                let filtered = alpha * (prev_output + input - prev_input);
                prev_output = filtered;
                prev_input = input;
                filtered
            }
            _ => input, // 直通
        };
        
        samples[i] = output;
    }
}

// 日志函数
fn log_info(message: &str) {
    log_message(0, message); // INFO level
}

fn log_error(message: &str) {
    log_message(2, message); // ERROR level
}

fn log_message(level: i32, message: &str) {
    unsafe {
        let msg_bytes = message.as_bytes();
        host_log(level, msg_bytes.as_ptr(), msg_bytes.len() as i32);
    }
}

// 内存分配函数
#[no_mangle]
pub extern "C" fn malloc(size: i32) -> *mut u8 {
    let layout = std::alloc::Layout::from_size_align(size as usize, 1).unwrap();
    unsafe { std::alloc::alloc(layout) }
}

#[no_mangle]
pub extern "C" fn free(ptr: *mut u8, size: i32) {
    let layout = std::alloc::Layout::from_size_align(size as usize, 1).unwrap();
    unsafe { std::alloc::dealloc(ptr, layout) }
}
```

### 构建脚本

```bash
#!/bin/bash
# rust/build.sh

set -e

echo "Building Rust WASM plugin..."

# 构建 WASM 模块
cargo build --target wasm32-wasi --release

# 复制到 build 目录
cp target/wasm32-wasi/release/audio_filter_wasm.wasm build/plugin.wasm

# 优化 WASM 文件
if command -v wasm-opt &> /dev/null; then
    echo "Optimizing WASM file..."
    wasm-opt -Oz build/plugin.wasm -o build/plugin.wasm
fi

# 显示文件信息
echo "Build completed:"
ls -lh build/plugin.wasm

echo "Rust WASM plugin built successfully!"
```

## C/C++ WASM 插件开发

### Makefile 配置

```makefile
# c/Makefile
CC = $(WASI_SDK_PATH)/bin/clang
CFLAGS = -O2 -Wall -Wextra -std=c11
LDFLAGS = -Wl,--export-all -Wl,--no-entry

SRCDIR = src
BUILDDIR = build
SOURCES = $(wildcard $(SRCDIR)/*.c)
OBJECTS = $(SOURCES:$(SRCDIR)/%.c=$(BUILDDIR)/%.o)
TARGET = $(BUILDDIR)/plugin.wasm

.PHONY: all clean

all: $(TARGET)

$(TARGET): $(OBJECTS) | $(BUILDDIR)
	$(CC) $(LDFLAGS) -o $@ $^
	if command -v wasm-opt >/dev/null 2>&1; then \
		wasm-opt -Oz $@ -o $@; \
	fi

$(BUILDDIR)/%.o: $(SRCDIR)/%.c | $(BUILDDIR)
	$(CC) $(CFLAGS) -c -o $@ $<

$(BUILDDIR):
	mkdir -p $(BUILDDIR)

clean:
	rm -rf $(BUILDDIR)

info:
	@echo "Build completed:"
	@ls -lh $(TARGET)
```

### C 插件实现

```c
// c/src/main.c
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>
#include <stdbool.h>
#include "host.h"

// 插件状态
static bool is_running = false;
static FilterParams filter_params = {
    .cutoff = 1000.0f,
    .resonance = 0.7f,
    .type = FILTER_LOWPASS,
    .enabled = true
};

// 滤波器状态
static float prev_output[MAX_CHANNELS] = {0};
static float prev_input[MAX_CHANNELS] = {0};

// 导出函数实现

__attribute__((export_name("initialize")))
int32_t initialize(const uint8_t* config_ptr, int32_t config_len) {
    // 解析配置（简化实现）
    if (config_len > 0) {
        // 这里应该解析 JSON 配置
        // 为简化，使用默认参数
        log_info("Using default filter parameters");
    }
    
    // 重置滤波器状态
    memset(prev_output, 0, sizeof(prev_output));
    memset(prev_input, 0, sizeof(prev_input));
    
    log_info("Audio filter plugin initialized");
    return 0;
}

__attribute__((export_name("start")))
int32_t start() {
    is_running = true;
    log_info("Audio filter plugin started");
    return 0;
}

__attribute__((export_name("stop")))
int32_t stop() {
    is_running = false;
    log_info("Audio filter plugin stopped");
    return 0;
}

__attribute__((export_name("cleanup")))
int32_t cleanup() {
    is_running = false;
    memset(prev_output, 0, sizeof(prev_output));
    memset(prev_input, 0, sizeof(prev_input));
    log_info("Audio filter plugin cleaned up");
    return 0;
}

__attribute__((export_name("process_audio")))
int32_t process_audio(const uint8_t* input_ptr, int32_t input_len, uint8_t* output_ptr) {
    if (!is_running || !filter_params.enabled) {
        // 直接复制输入到输出
        memcpy(output_ptr, input_ptr, input_len);
        return input_len;
    }
    
    // 解析音频数据（简化实现）
    AudioBuffer* input_buffer = (AudioBuffer*)input_ptr;
    AudioBuffer* output_buffer = (AudioBuffer*)output_ptr;
    
    // 复制元数据
    output_buffer->sample_rate = input_buffer->sample_rate;
    output_buffer->channels = input_buffer->channels;
    output_buffer->frames = input_buffer->frames;
    
    // 应用滤波器到每个声道
    for (int ch = 0; ch < input_buffer->channels && ch < MAX_CHANNELS; ch++) {
        apply_filter_to_channel(
            input_buffer->data[ch],
            output_buffer->data[ch],
            input_buffer->frames,
            input_buffer->sample_rate,
            ch
        );
    }
    
    return input_len; // 返回处理的字节数
}

__attribute__((export_name("get_metadata")))
const char* get_metadata() {
    static const char metadata[] = 
        "{"
        "\"id\":\"audio-filter-c-wasm\","
        "\"name\":\"Audio Filter C WASM Plugin\","
        "\"version\":\"1.0.0\","
        "\"description\":\"C WebAssembly audio filter plugin\","
        "\"author\":\"Your Name\","
        "\"tags\":[\"audio\",\"filter\",\"c\",\"wasm\"]"
        "}";
    
    return metadata;
}

// 内存管理函数
__attribute__((export_name("malloc")))
void* wasm_malloc(size_t size) {
    return malloc(size);
}

__attribute__((export_name("free")))
void wasm_free(void* ptr) {
    free(ptr);
}

// 音频处理函数
void apply_filter_to_channel(const float* input, float* output, int frames, int sample_rate, int channel) {
    float nyquist = (float)sample_rate / 2.0f;
    float normalized_cutoff = filter_params.cutoff / nyquist;
    
    // 限制频率范围
    if (normalized_cutoff > 0.99f) normalized_cutoff = 0.99f;
    if (normalized_cutoff < 0.01f) normalized_cutoff = 0.01f;
    
    float alpha = normalized_cutoff;
    
    for (int i = 0; i < frames; i++) {
        float sample = input[i];
        float filtered;
        
        switch (filter_params.type) {
            case FILTER_LOWPASS:
                filtered = alpha * sample + (1.0f - alpha) * prev_output[channel];
                prev_output[channel] = filtered;
                break;
                
            case FILTER_HIGHPASS:
                filtered = alpha * (prev_output[channel] + sample - prev_input[channel]);
                prev_output[channel] = filtered;
                prev_input[channel] = sample;
                break;
                
            default:
                filtered = sample; // 直通
                break;
        }
        
        output[i] = filtered;
    }
}
```

### 头文件定义

```c
// c/src/host.h
#ifndef HOST_H
#define HOST_H

#include <stdint.h>
#include <stdbool.h>

#define MAX_CHANNELS 8
#define MAX_FRAMES 4096

// 滤波器类型
typedef enum {
    FILTER_LOWPASS,
    FILTER_HIGHPASS,
    FILTER_BANDPASS
} FilterType;

// 滤波器参数
typedef struct {
    float cutoff;
    float resonance;
    FilterType type;
    bool enabled;
} FilterParams;

// 音频缓冲区
typedef struct {
    float data[MAX_CHANNELS][MAX_FRAMES];
    int32_t sample_rate;
    int32_t channels;
    int32_t frames;
} AudioBuffer;

// 主机接口函数声明
__attribute__((import_module("env"), import_name("host_log")))
void host_log(int32_t level, const uint8_t* msg_ptr, int32_t msg_len);

__attribute__((import_module("env"), import_name("host_get_config")))
int32_t host_get_config(const uint8_t* key_ptr, int32_t key_len);

__attribute__((import_module("env"), import_name("host_emit_event")))
void host_emit_event(const uint8_t* type_ptr, int32_t type_len, const uint8_t* data_ptr, int32_t data_len);

// 便利函数
void log_info(const char* message);
void log_error(const char* message);
void apply_filter_to_channel(const float* input, float* output, int frames, int sample_rate, int channel);

#endif // HOST_H
```

```c
// c/src/host.c
#include "host.h"
#include <string.h>

void log_info(const char* message) {
    host_log(0, (const uint8_t*)message, strlen(message));
}

void log_error(const char* message) {
    host_log(2, (const uint8_t*)message, strlen(message));
}
```

## 主机接口设计

### WASM 运行时封装

```go
// host/runtime.go
package host

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/bytecodealliance/wasmtime-go/v14"
)

type WASMRuntime struct {
    engine   *wasmtime.Engine
    store    *wasmtime.Store
    module   *wasmtime.Module
    instance *wasmtime.Instance
    
    // 导出函数
    initialize   *wasmtime.Func
    start        *wasmtime.Func
    stop         *wasmtime.Func
    cleanup      *wasmtime.Func
    processAudio *wasmtime.Func
    getMetadata  *wasmtime.Func
    malloc       *wasmtime.Func
    free         *wasmtime.Func
    
    // 内存
    memory       *wasmtime.Memory
    
    // 配置
    config       *WASMConfig
    
    // 状态
    mutex        sync.RWMutex
    running      bool
    
    // 主机接口
    hostInterface *HostInterface
}

type WASMConfig struct {
    MaxMemoryPages int
    MaxTableSize   int
    MaxInstances   int
    Timeout        time.Duration
    EnableWASI     bool
    AllowedHosts   []string
}

func NewWASMRuntime(wasmBytes []byte, config *WASMConfig) (*WASMRuntime, error) {
    // 创建引擎
    engine := wasmtime.NewEngine()
    
    // 创建存储
    store := wasmtime.NewStore(engine)
    
    // 设置资源限制
    if config.MaxMemoryPages > 0 {
        store.Limiter(config.MaxMemoryPages*65536, config.MaxTableSize, config.MaxInstances, 1000, 1000)
    }
    
    // 编译模块
    module, err := wasmtime.NewModule(engine, wasmBytes)
    if err != nil {
        return nil, fmt.Errorf("failed to compile WASM module: %w", err)
    }
    
    runtime := &WASMRuntime{
        engine: engine,
        store:  store,
        module: module,
        config: config,
    }
    
    // 创建主机接口
    runtime.hostInterface = NewHostInterface(runtime)
    
    // 创建导入对象
    imports, err := runtime.createImports()
    if err != nil {
        return nil, fmt.Errorf("failed to create imports: %w", err)
    }
    
    // 实例化模块
    instance, err := wasmtime.NewInstance(store, module, imports)
    if err != nil {
        return nil, fmt.Errorf("failed to instantiate WASM module: %w", err)
    }
    
    runtime.instance = instance
    
    // 获取导出函数
    if err := runtime.bindExports(); err != nil {
        return nil, fmt.Errorf("failed to bind exports: %w", err)
    }
    
    return runtime, nil
}

func (r *WASMRuntime) createImports() ([]wasmtime.AsExtern, error) {
    var imports []wasmtime.AsExtern
    
    // 主机日志函数
    hostLog := wasmtime.NewFunc(
        r.store,
        wasmtime.NewFuncType(
            []*wasmtime.ValType{
                wasmtime.NewValType(wasmtime.KindI32), // level
                wasmtime.NewValType(wasmtime.KindI32), // msg_ptr
                wasmtime.NewValType(wasmtime.KindI32), // msg_len
            },
            []*wasmtime.ValType{},
        ),
        func(caller *wasmtime.Caller, args []wasmtime.Val) ([]wasmtime.Val, *wasmtime.Trap) {
            level := args[0].I32()
            msgPtr := args[1].I32()
            msgLen := args[2].I32()
            
            memory := caller.GetExport("memory").Memory()
            data := memory.UnsafeData(caller)
            
            if msgPtr >= 0 && msgLen > 0 && int(msgPtr+msgLen) <= len(data) {
                message := string(data[msgPtr : msgPtr+msgLen])
                r.hostInterface.Log(level, message)
            }
            
            return []wasmtime.Val{}, nil
        },
    )
    imports = append(imports, hostLog.AsExtern())
    
    // 主机配置函数
    hostGetConfig := wasmtime.NewFunc(
        r.store,
        wasmtime.NewFuncType(
            []*wasmtime.ValType{
                wasmtime.NewValType(wasmtime.KindI32), // key_ptr
                wasmtime.NewValType(wasmtime.KindI32), // key_len
            },
            []*wasmtime.ValType{
                wasmtime.NewValType(wasmtime.KindI32), // result_ptr
            },
        ),
        func(caller *wasmtime.Caller, args []wasmtime.Val) ([]wasmtime.Val, *wasmtime.Trap) {
            keyPtr := args[0].I32()
            keyLen := args[1].I32()
            
            memory := caller.GetExport("memory").Memory()
            data := memory.UnsafeData(caller)
            
            if keyPtr >= 0 && keyLen > 0 && int(keyPtr+keyLen) <= len(data) {
                key := string(data[keyPtr : keyPtr+keyLen])
                value := r.hostInterface.GetConfig(key)
                
                // 将结果写入内存（简化实现）
                valueBytes := []byte(value)
                resultPtr := r.allocateMemory(int32(len(valueBytes)))
                if resultPtr > 0 {
                    copy(data[resultPtr:], valueBytes)
                    return []wasmtime.Val{wasmtime.ValI32(resultPtr)}, nil
                }
            }
            
            return []wasmtime.Val{wasmtime.ValI32(0)}, nil
        },
    )
    imports = append(imports, hostGetConfig.AsExtern())
    
    // 主机事件函数
    hostEmitEvent := wasmtime.NewFunc(
        r.store,
        wasmtime.NewFuncType(
            []*wasmtime.ValType{
                wasmtime.NewValType(wasmtime.KindI32), // type_ptr
                wasmtime.NewValType(wasmtime.KindI32), // type_len
                wasmtime.NewValType(wasmtime.KindI32), // data_ptr
                wasmtime.NewValType(wasmtime.KindI32), // data_len
            },
            []*wasmtime.ValType{},
        ),
        func(caller *wasmtime.Caller, args []wasmtime.Val) ([]wasmtime.Val, *wasmtime.Trap) {
            typePtr := args[0].I32()
            typeLen := args[1].I32()
            dataPtr := args[2].I32()
            dataLen := args[3].I32()
            
            memory := caller.GetExport("memory").Memory()
            data := memory.UnsafeData(caller)
            
            if typePtr >= 0 && typeLen > 0 && int(typePtr+typeLen) <= len(data) &&
               dataPtr >= 0 && dataLen > 0 && int(dataPtr+dataLen) <= len(data) {
                eventType := string(data[typePtr : typePtr+typeLen])
                eventData := string(data[dataPtr : dataPtr+dataLen])
                r.hostInterface.EmitEvent(eventType, eventData)
            }
            
            return []wasmtime.Val{}, nil
        },
    )
    imports = append(imports, hostEmitEvent.AsExtern())
    
    return imports, nil
}

func (r *WASMRuntime) bindExports() error {
    // 获取内存
    memoryExport := r.instance.GetExport(r.store, "memory")
    if memoryExport == nil {
        return fmt.Errorf("memory export not found")
    }
    r.memory = memoryExport.Memory()
    
    // 绑定导出函数
    exports := map[string]**wasmtime.Func{
        "initialize":    &r.initialize,
        "start":         &r.start,
        "stop":          &r.stop,
        "cleanup":       &r.cleanup,
        "process_audio": &r.processAudio,
        "get_metadata":  &r.getMetadata,
        "malloc":        &r.malloc,
        "free":          &r.free,
    }
    
    for name, funcPtr := range exports {
        export := r.instance.GetExport(r.store, name)
        if export == nil {
            return fmt.Errorf("export %s not found", name)
        }
        *funcPtr = export.Func()
    }
    
    return nil
}

// 插件生命周期方法
func (r *WASMRuntime) Initialize(config map[string]interface{}) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    configJSON, err := json.Marshal(config)
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }
    
    // 分配内存并写入配置
    configPtr := r.allocateMemory(int32(len(configJSON)))
    if configPtr == 0 {
        return fmt.Errorf("failed to allocate memory for config")
    }
    defer r.freeMemory(configPtr)
    
    r.writeMemory(configPtr, configJSON)
    
    // 调用初始化函数
    result, err := r.initialize.Call(r.store, configPtr, int32(len(configJSON)))
    if err != nil {
        return fmt.Errorf("initialize call failed: %w", err)
    }
    
    if result.(int32) != 0 {
        return fmt.Errorf("initialize returned error code: %d", result.(int32))
    }
    
    return nil
}

func (r *WASMRuntime) Start() error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    result, err := r.start.Call(r.store)
    if err != nil {
        return fmt.Errorf("start call failed: %w", err)
    }
    
    if result.(int32) != 0 {
        return fmt.Errorf("start returned error code: %d", result.(int32))
    }
    
    r.running = true
    return nil
}

func (r *WASMRuntime) Stop() error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    if !r.running {
        return nil
    }
    
    result, err := r.stop.Call(r.store)
    if err != nil {
        return fmt.Errorf("stop call failed: %w", err)
    }
    
    if result.(int32) != 0 {
        return fmt.Errorf("stop returned error code: %d", result.(int32))
    }
    
    r.running = false
    return nil
}

func (r *WASMRuntime) ProcessAudio(input []byte) ([]byte, error) {
    r.mutex.RLock()
    defer r.mutex.RUnlock()
    
    if !r.running {
        return nil, fmt.Errorf("plugin not running")
    }
    
    // 分配输入内存
    inputPtr := r.allocateMemory(int32(len(input)))
    if inputPtr == 0 {
        return nil, fmt.Errorf("failed to allocate input memory")
    }
    defer r.freeMemory(inputPtr)
    
    // 分配输出内存
    outputPtr := r.allocateMemory(int32(len(input) * 2)) // 预留更多空间
    if outputPtr == 0 {
        return nil, fmt.Errorf("failed to allocate output memory")
    }
    defer r.freeMemory(outputPtr)
    
    // 写入输入数据
    r.writeMemory(inputPtr, input)
    
    // 调用处理函数
    result, err := r.processAudio.Call(r.store, inputPtr, int32(len(input)), outputPtr)
    if err != nil {
        return nil, fmt.Errorf("process_audio call failed: %w", err)
    }
    
    outputLen := result.(int32)
    if outputLen < 0 {
        return nil, fmt.Errorf("process_audio returned error code: %d", outputLen)
    }
    
    // 读取输出数据
    output := r.readMemory(outputPtr, outputLen)
    return output, nil
}

// 内存管理方法
func (r *WASMRuntime) allocateMemory(size int32) int32 {
    if r.malloc == nil {
        return 0
    }
    
    result, err := r.malloc.Call(r.store, size)
    if err != nil {
        return 0
    }
    
    return result.(int32)
}

func (r *WASMRuntime) freeMemory(ptr int32) {
    if r.free != nil && ptr != 0 {
        r.free.Call(r.store, ptr)
    }
}

func (r *WASMRuntime) readMemory(ptr, length int32) []byte {
    if r.memory == nil || ptr == 0 || length <= 0 {
        return nil
    }
    
    data := r.memory.UnsafeData(r.store)
    if int(ptr+length) > len(data) {
        return nil
    }
    
    result := make([]byte, length)
    copy(result, data[ptr:ptr+length])
    return result
}

func (r *WASMRuntime) writeMemory(ptr int32, data []byte) {
    if r.memory == nil || ptr == 0 || len(data) == 0 {
        return
    }
    
    memData := r.memory.UnsafeData(r.store)
    if int(ptr)+len(data) <= len(memData) {
        copy(memData[ptr:], data)
    }
}

func (r *WASMRuntime) Close() error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    if r.running {
        r.Stop()
    }
    
    if r.cleanup != nil {
        r.cleanup.Call(r.store)
    }
    
    return nil
}
```

### 主机接口实现

```go
// host/interface.go
package host

import (
    "encoding/json"
    "fmt"
    "log/slog"
    "net/http"
    "os"
    "sync"
)

type HostInterface struct {
    runtime *WASMRuntime
    config  map[string]string
    logger  *slog.Logger
    
    // 事件处理
    eventHandlers map[string][]EventHandler
    eventMutex    sync.RWMutex
}

type EventHandler func(eventType, data string)

func NewHostInterface(runtime *WASMRuntime) *HostInterface {
    return &HostInterface{
        runtime:       runtime,
        config:        make(map[string]string),
        logger:        slog.Default(),
        eventHandlers: make(map[string][]EventHandler),
    }
}

// 日志接口
func (h *HostInterface) Log(level int32, message string) {
    switch level {
    case 0: // INFO
        h.logger.Info("[WASM Plugin]", "message", message)
    case 1: // WARN
        h.logger.Warn("[WASM Plugin]", "message", message)
    case 2: // ERROR
        h.logger.Error("[WASM Plugin]", "message", message)
    default:
        h.logger.Debug("[WASM Plugin]", "message", message)
    }
}

// 配置接口
func (h *HostInterface) GetConfig(key string) string {
    if value, exists := h.config[key]; exists {
        return value
    }
    return ""
}

func (h *HostInterface) SetConfig(key, value string) {
    h.config[key] = value
}

// 网络接口
func (h *HostInterface) HttpRequest(method, url, body string) (int32, string) {
    client := &http.Client{}
    
    var req *http.Request
    var err error
    
    if body != "" {
        req, err = http.NewRequest(method, url, strings.NewReader(body))
    } else {
        req, err = http.NewRequest(method, url, nil)
    }
    
    if err != nil {
        h.logger.Error("Failed to create HTTP request", "error", err)
        return -1, err.Error()
    }
    
    resp, err := client.Do(req)
    if err != nil {
        h.logger.Error("HTTP request failed", "error", err)
        return -1, err.Error()
    }
    defer resp.Body.Close()
    
    responseBody, err := io.ReadAll(resp.Body)
    if err != nil {
        h.logger.Error("Failed to read response body", "error", err)
        return int32(resp.StatusCode), ""
    }
    
    return int32(resp.StatusCode), string(responseBody)
}

// 文件接口
func (h *HostInterface) ReadFile(path string) []byte {
    // 安全检查：只允许读取特定目录下的文件
    if !h.isPathAllowed(path) {
        h.logger.Warn("Unauthorized file access attempt", "path", path)
        return nil
    }
    
    data, err := os.ReadFile(path)
    if err != nil {
        h.logger.Error("Failed to read file", "path", path, "error", err)
        return nil
    }
    
    return data
}

func (h *HostInterface) WriteFile(path string, data []byte) int32 {
    // 安全检查
    if !h.isPathAllowed(path) {
        h.logger.Warn("Unauthorized file write attempt", "path", path)
        return -1
    }
    
    err := os.WriteFile(path, data, 0644)
    if err != nil {
        h.logger.Error("Failed to write file", "path", path, "error", err)
        return -1
    }
    
    return 0
}

// 事件接口
func (h *HostInterface) EmitEvent(eventType, data string) {
    h.eventMutex.RLock()
    handlers := h.eventHandlers[eventType]
    h.eventMutex.RUnlock()
    
    for _, handler := range handlers {
        go handler(eventType, data)
    }
}

func (h *HostInterface) SubscribeEvent(eventType string, handler EventHandler) {
    h.eventMutex.Lock()
    defer h.eventMutex.Unlock()
    
    h.eventHandlers[eventType] = append(h.eventHandlers[eventType], handler)
}

// 安全检查
func (h *HostInterface) isPathAllowed(path string) bool {
    // 简化的路径检查，实际应该更严格
    allowedPrefixes := []string{
        "./data/",
        "./config/",
        "./temp/",
    }
    
    for _, prefix := range allowedPrefixes {
        if strings.HasPrefix(path, prefix) {
            return true
        }
    }
    
    return false
}
```

## 内存管理

### 内存池优化

```go
// host/memory.go
package host

import (
    "sync"
    "unsafe"
)

type MemoryPool struct {
    pools map[int]*sync.Pool
    mutex sync.RWMutex
}

func NewMemoryPool() *MemoryPool {
    return &MemoryPool{
        pools: make(map[int]*sync.Pool),
    }
}

func (mp *MemoryPool) Get(size int) []byte {
    // 向上取整到最近的 2 的幂
    poolSize := nextPowerOfTwo(size)
    
    mp.mutex.RLock()
    pool, exists := mp.pools[poolSize]
    mp.mutex.RUnlock()
    
    if !exists {
        mp.mutex.Lock()
        if pool, exists = mp.pools[poolSize]; !exists {
            pool = &sync.Pool{
                New: func() interface{} {
                    return make([]byte, poolSize)
                },
            }
            mp.pools[poolSize] = pool
        }
        mp.mutex.Unlock()
    }
    
    buf := pool.Get().([]byte)
    return buf[:size] // 返回请求的大小
}

func (mp *MemoryPool) Put(buf []byte) {
    if cap(buf) == 0 {
        return
    }
    
    poolSize := cap(buf)
    
    mp.mutex.RLock()
    pool, exists := mp.pools[poolSize]
    mp.mutex.RUnlock()
    
    if exists {
        // 清零缓冲区
        for i := range buf {
            buf[i] = 0
        }
        pool.Put(buf[:cap(buf)])
    }
}

func nextPowerOfTwo(n int) int {
    if n <= 0 {
        return 1
    }
    
    n--
    n |= n >> 1
    n |= n >> 2
    n |= n >> 4
    n |= n >> 8
    n |= n >> 16
    n++
    
    return n
}
```

### WASM 内存监控

```go
// host/monitor.go
package host

import (
    "sync/atomic"
    "time"
)

type MemoryMonitor struct {
    runtime *WASMRuntime
    
    // 统计信息
    allocCount   int64
    freeCount    int64
    totalAlloc   int64
    currentAlloc int64
    maxAlloc     int64
    
    // 监控配置
    maxMemory    int64
    checkInterval time.Duration
    
    // 控制
    stopChan     chan struct{}
}

func NewMemoryMonitor(runtime *WASMRuntime, maxMemory int64) *MemoryMonitor {
    return &MemoryMonitor{
        runtime:       runtime,
        maxMemory:     maxMemory,
        checkInterval: 5 * time.Second,
        stopChan:      make(chan struct{}),
    }
}

func (mm *MemoryMonitor) Start() {
    ticker := time.NewTicker(mm.checkInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            mm.checkMemoryUsage()
        case <-mm.stopChan:
            return
        }
    }
}

func (mm *MemoryMonitor) Stop() {
    close(mm.stopChan)
}

func (mm *MemoryMonitor) RecordAlloc(size int64) {
    atomic.AddInt64(&mm.allocCount, 1)
    atomic.AddInt64(&mm.totalAlloc, size)
    
    current := atomic.AddInt64(&mm.currentAlloc, size)
    
    // 更新最大分配量
    for {
        max := atomic.LoadInt64(&mm.maxAlloc)
        if current <= max || atomic.CompareAndSwapInt64(&mm.maxAlloc, max, current) {
            break
        }
    }
}

func (mm *MemoryMonitor) RecordFree(size int64) {
    atomic.AddInt64(&mm.freeCount, 1)
    atomic.AddInt64(&mm.currentAlloc, -size)
}

func (mm *MemoryMonitor) GetStats() MemoryStats {
    return MemoryStats{
        AllocCount:   atomic.LoadInt64(&mm.allocCount),
        FreeCount:    atomic.LoadInt64(&mm.freeCount),
        TotalAlloc:   atomic.LoadInt64(&mm.totalAlloc),
        CurrentAlloc: atomic.LoadInt64(&mm.currentAlloc),
        MaxAlloc:     atomic.LoadInt64(&mm.maxAlloc),
    }
}

func (mm *MemoryMonitor) checkMemoryUsage() {
    current := atomic.LoadInt64(&mm.currentAlloc)
    
    if current > mm.maxMemory {
        // 内存使用超限，触发垃圾回收或其他处理
        mm.runtime.hostInterface.Log(1, fmt.Sprintf("Memory usage exceeded limit: %d > %d", current, mm.maxMemory))
        
        // 可以在这里触发强制垃圾回收或限制新的内存分配
    }
}

type MemoryStats struct {
    AllocCount   int64 `json:"alloc_count"`
    FreeCount    int64 `json:"free_count"`
    TotalAlloc   int64 `json:"total_alloc"`
    CurrentAlloc int64 `json:"current_alloc"`
    MaxAlloc     int64 `json:"max_alloc"`
}
```

## 性能优化

### 1. 编译优化

```bash
# Go WASM 优化
GOOS=wasip1 GOARCH=wasm go build -ldflags="-s -w" -o plugin.wasm

# Rust 优化
cargo build --target wasm32-wasi --release
wasm-opt -Oz plugin.wasm -o plugin.wasm

# C/C++ 优化
clang -O3 -flto -Wl,--lto-O3 --target=wasm32-wasi -o plugin.wasm main.c
wasm-opt -O3 plugin.wasm -o plugin.wasm
```

### 2. 批处理优化

```go
// 批量处理音频数据
func (r *WASMRuntime) ProcessAudioBatch(inputs [][]byte) ([][]byte, error) {
    outputs := make([][]byte, len(inputs))
    
    // 预分配内存
    totalInputSize := 0
    for _, input := range inputs {
        totalInputSize += len(input)
    }
    
    inputPtr := r.allocateMemory(int32(totalInputSize))
    if inputPtr == 0 {
        return nil, fmt.Errorf("failed to allocate batch input memory")
    }
    defer r.freeMemory(inputPtr)
    
    outputPtr := r.allocateMemory(int32(totalInputSize * 2))
    if outputPtr == 0 {
        return nil, fmt.Errorf("failed to allocate batch output memory")
    }
    defer r.freeMemory(outputPtr)
    
    // 批量写入输入数据
    offset := int32(0)
    for i, input := range inputs {
        r.writeMemory(inputPtr+offset, input)
        
        // 处理单个音频块
        result, err := r.processAudio.Call(r.store, inputPtr+offset, int32(len(input)), outputPtr+offset)
        if err != nil {
            return nil, fmt.Errorf("batch process failed at index %d: %w", i, err)
        }
        
        outputLen := result.(int32)
        if outputLen < 0 {
            return nil, fmt.Errorf("batch process error at index %d: %d", i, outputLen)
        }
        
        outputs[i] = r.readMemory(outputPtr+offset, outputLen)
        offset += int32(len(input))
    }
    
    return outputs, nil
}
```

### 3. 缓存优化

```go
// 结果缓存
type ResultCache struct {
    cache map[string][]byte
    mutex sync.RWMutex
    maxSize int
    ttl     time.Duration
}

func (rc *ResultCache) Get(key string) ([]byte, bool) {
    rc.mutex.RLock()
    defer rc.mutex.RUnlock()
    
    result, exists := rc.cache[key]
    return result, exists
}

func (rc *ResultCache) Set(key string, value []byte) {
    rc.mutex.Lock()
    defer rc.mutex.Unlock()
    
    if len(rc.cache) >= rc.maxSize {
        // 简单的 LRU 清理
        for k := range rc.cache {
            delete(rc.cache, k)
            if len(rc.cache) < rc.maxSize/2 {
                break
            }
        }
    }
    
    rc.cache[key] = value
}
```

## 调试技巧

### 1. 调试信息输出

```go
// 调试模式下的详细日志
func (r *WASMRuntime) EnableDebugMode() {
    r.hostInterface.SetConfig("debug_mode", "true")
    
    // 添加调试拦截器
    r.addDebugInterceptors()
}

func (r *WASMRuntime) addDebugInterceptors() {
    // 拦截所有函数调用
    originalProcessAudio := r.processAudio
    
    r.processAudio = wasmtime.NewFunc(
        r.store,
        originalProcessAudio.Type(r.store),
        func(caller *wasmtime.Caller, args []wasmtime.Val) ([]wasmtime.Val, *wasmtime.Trap) {
            start := time.Now()
            
            r.hostInterface.Log(0, fmt.Sprintf("ProcessAudio called with args: %v", args))
            
            result, trap := originalProcessAudio.Call(caller, args...)
            
            duration := time.Since(start)
            r.hostInterface.Log(0, fmt.Sprintf("ProcessAudio completed in %v, result: %v", duration, result))
            
            return result, trap
        },
    )
}
```

### 2. 性能分析

```go
// 性能分析器
type WASMProfiler struct {
    runtime *WASMRuntime
    samples []ProfileSample
    mutex   sync.Mutex
}

type ProfileSample struct {
    Function string
    Duration time.Duration
    Args     []interface{}
    Result   interface{}
    Error    error
    Timestamp time.Time
}

func (wp *WASMProfiler) ProfileFunction(name string, fn func() (interface{}, error)) (interface{}, error) {
    start := time.Now()
    result, err := fn()
    duration := time.Since(start)
    
    wp.mutex.Lock()
    wp.samples = append(wp.samples, ProfileSample{
        Function:  name,
        Duration:  duration,
        Result:    result,
        Error:     err,
        Timestamp: start,
    })
    wp.mutex.Unlock()
    
    return result, err
}

func (wp *WASMProfiler) GetReport() ProfileReport {
    wp.mutex.Lock()
    defer wp.mutex.Unlock()
    
    report := ProfileReport{
        TotalSamples: len(wp.samples),
        Functions:    make(map[string]FunctionStats),
    }
    
    for _, sample := range wp.samples {
        stats := report.Functions[sample.Function]
        stats.CallCount++
        stats.TotalDuration += sample.Duration
        
        if sample.Duration > stats.MaxDuration {
            stats.MaxDuration = sample.Duration
        }
        
        if stats.MinDuration == 0 || sample.Duration < stats.MinDuration {
            stats.MinDuration = sample.Duration
        }
        
        if sample.Error != nil {
            stats.ErrorCount++
        }
        
        report.Functions[sample.Function] = stats
    }
    
    // 计算平均值
    for name, stats := range report.Functions {
        if stats.CallCount > 0 {
            stats.AvgDuration = stats.TotalDuration / time.Duration(stats.CallCount)
            report.Functions[name] = stats
        }
    }
    
    return report
}

type ProfileReport struct {
    TotalSamples int                    `json:"total_samples"`
    Functions    map[string]FunctionStats `json:"functions"`
}

type FunctionStats struct {
    CallCount     int           `json:"call_count"`
    ErrorCount    int           `json:"error_count"`
    TotalDuration time.Duration `json:"total_duration"`
    AvgDuration   time.Duration `json:"avg_duration"`
    MinDuration   time.Duration `json:"min_duration"`
    MaxDuration   time.Duration `json:"max_duration"`
}
```

## 部署指南

### 1. 插件打包

```bash
#!/bin/bash
# scripts/package.sh

set -e

PLUGIN_NAME="audio-filter-wasm"
VERSION="1.0.0"
DIST_DIR="dist"

echo "Packaging WASM plugin..."

# 创建发布目录
mkdir -p "$DIST_DIR"

# 构建所有语言版本
echo "Building Go version..."
(cd go && ./build.sh)

echo "Building Rust version..."
(cd rust && ./build.sh)

echo "Building C version..."
(cd c && make clean && make)

# 创建插件包
package_dir="$DIST_DIR/${PLUGIN_NAME}-${VERSION}"
mkdir -p "$package_dir"

# 复制 WASM 文件
cp go/build/plugin.wasm "$package_dir/plugin-go.wasm"
cp rust/build/plugin.wasm "$package_dir/plugin-rust.wasm"
cp c/build/plugin.wasm "$package_dir/plugin-c.wasm"

# 复制配置和文档
cp plugin.json "$package_dir/"
cp README.md "$package_dir/"
cp -r docs "$package_dir/"

# 创建安装脚本
cat > "$package_dir/install.sh" << 'EOF'
#!/bin/bash
set -e

PLUGIN_DIR="$HOME/.go-musicfox/plugins/audio-filter-wasm"

echo "Installing WASM audio filter plugin..."

# 创建插件目录
mkdir -p "$PLUGIN_DIR"

# 复制文件
cp *.wasm "$PLUGIN_DIR/"
cp plugin.json "$PLUGIN_DIR/"

echo "Plugin installed successfully to $PLUGIN_DIR"
EOF

chmod +x "$package_dir/install.sh"

# 创建压缩包
tar -czf "$DIST_DIR/${PLUGIN_NAME}-${VERSION}.tar.gz" -C "$DIST_DIR" "${PLUGIN_NAME}-${VERSION}"

# 清理临时目录
rm -rf "$package_dir"

echo "Package created: $DIST_DIR/${PLUGIN_NAME}-${VERSION}.tar.gz"
```

### 2. 配置文件

```json
{
  "id": "audio-filter-wasm",
  "name": "Audio Filter WASM Plugin",
  "version": "1.0.0",
  "description": "WebAssembly audio filter plugin with multiple language implementations",
  "author": "Your Name",
  "license": "MIT",
  "type": "wasm",
  "category": "audio-effect",
  "tags": ["audio", "filter", "wasm", "cross-platform"],
  "api_version": "v2.0.0",
  "min_kernel_version": "2.0.0",
  "wasm": {
    "implementations": {
      "go": {
        "file": "plugin-go.wasm",
        "runtime": "wasmtime",
        "memory_limit": "50MB",
        "timeout": "30s"
      },
      "rust": {
        "file": "plugin-rust.wasm",
        "runtime": "wasmtime",
        "memory_limit": "30MB",
        "timeout": "30s"
      },
      "c": {
        "file": "plugin-c.wasm",
        "runtime": "wasmtime",
        "memory_limit": "20MB",
        "timeout": "30s"
      }
    },
    "preferred": "rust",
    "host_functions": [
      "host_log",
      "host_get_config",
      "host_emit_event"
    ]
  },
  "permissions": [
    {
      "id": "audio-processing",
      "description": "Process audio data",
      "required": true
    }
  ],
  "config_schema": {
    "type": "object",
    "properties": {
      "filter_params": {
        "type": "object",
        "properties": {
          "cutoff": {
            "type": "number",
            "minimum": 20,
            "maximum": 20000,
            "default": 1000
          },
          "resonance": {
            "type": "number",
            "minimum": 0.1,
            "maximum": 10.0,
            "default": 0.7
          },
          "type": {
            "type": "string",
            "enum": ["lowpass", "highpass", "bandpass"],
            "default": "lowpass"
          },
          "enabled": {
            "type": "boolean",
            "default": true
          }
        }
      }
    }
  }
}
```

## 最佳实践

### 1. 安全性

- 限制 WASM 模块的内存和 CPU 使用
- 验证所有输入数据
- 使用沙箱环境运行
- 限制主机函数的访问权限

### 2. 性能

- 使用编译器优化选项
- 实现内存池和对象复用
- 避免频繁的内存分配
- 使用批处理减少函数调用开销

### 3. 兼容性

- 使用标准的 WASI 接口
- 避免平台特定的代码
- 提供多语言实现选择
- 保持 API 向后兼容

### 4. 调试

- 提供详细的错误信息
- 实现性能监控
- 使用结构化日志
- 支持调试模式

## 常见问题

### Q: WASM 插件性能如何？

A: WASM 性能通常是原生代码的 80-90%，对于大多数应用场景足够使用。

### Q: 如何调试 WASM 插件？

A: 可以使用 wasmtime 的调试功能，或者在主机端添加详细日志。

### Q: WASM 插件可以访问哪些系统资源？

A: 只能通过主机提供的接口访问，无法直接访问系统 API。

### Q: 如何处理 WASM 插件崩溃？

A: WASM 运行在沙箱中，崩溃不会影响主程序，可以重新加载插件。

## 相关文档

- [插件开发快速入门](plugin-quickstart.md)
- [动态库插件开发指南](dynamic-library.md)
- [RPC 插件开发指南](rpc-plugin.md)
- [插件测试指南](plugin-testing.md)
- [API 文档](../api/README.md)