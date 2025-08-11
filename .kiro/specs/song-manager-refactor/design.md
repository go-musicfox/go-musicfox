# 播放列表管理系统重构设计文档

## 概述

本设计文档描述了将现有的songManager接口及其实现从UI包重构为独立的播放列表管理系统的架构设计。重构的目标是创建一个更清晰、解耦、可测试且易于维护的播放列表管理系统。

### 当前问题分析

通过对现有代码的分析，发现以下主要问题：

1. **命名不清晰**：`songManager`接口名称不够描述性
2. **UI耦合**：播放列表管理逻辑与UI关注点紧密耦合，位于`internal/ui`包中
3. **复杂的模式切换**：每个播放模式都实现了所有其他模式的转换方法，导致代码重复
4. **可选类型使用**：使用自定义的`optionalSong`类型而非标准Go错误处理
5. **缺乏单元测试**：当前实现难以进行单元测试
6. **职责混乱**：播放列表管理与播放器状态管理混合在一起

## 架构设计

### 整体架构

新的架构将采用分层设计，将播放列表管理从UI层分离出来：

```
┌─────────────────────────────────────┐
│              UI Layer               │
│         (internal/ui)               │
└─────────────────┬───────────────────┘
                  │
┌─────────────────▼───────────────────┐
│          Playlist Service           │
│       (internal/playlist)           │
├─────────────────────────────────────┤
│  • PlaylistManager Interface        │
│  • PlayMode Strategy Interface      │
│  • Concrete PlayMode Implementations│
└─────────────────┬───────────────────┘
                  │
┌─────────────────▼───────────────────┐
│            Domain Layer             │
│        (internal/structs)           │
└─────────────────────────────────────┘
```

### 设计原则

1. **单一职责原则**：每个组件只负责一个特定的功能
2. **依赖倒置原则**：高层模块不依赖低层模块，都依赖抽象
3. **策略模式**：使用策略模式替代复杂的模式切换逻辑 <mcreference link="https://refactoring.guru/design-patterns/strategy" index="2">2</mcreference>
4. **标准Go实践**：使用标准Go错误处理模式替代自定义可选类型 <mcreference link="https://go.dev/blog/error-handling-and-go" index="1">1</mcreference>

## 组件和接口

### 核心接口定义

#### PlaylistManager接口

```go
// PlaylistManager 播放列表管理器接口
type PlaylistManager interface {
    // 初始化播放列表
    Initialize(index int, playlist []structs.Song) error
    
    // 获取当前播放列表
    GetPlaylist() []structs.Song
    
    // 获取当前歌曲索引
    GetCurrentIndex() int
    
    // 获取当前歌曲
    GetCurrentSong() (structs.Song, error)
    
    // 下一首歌曲
    NextSong(manual bool) (structs.Song, error)
    
    // 上一首歌曲
    PreviousSong(manual bool) (structs.Song, error)
    
    // 删除指定索引的歌曲
    RemoveSong(index int) error
    
    // 设置播放模式
    SetPlayMode(mode types.Mode) error
    
    // 获取当前播放模式
    GetPlayMode() types.Mode
    
    // 获取播放模式名称
    GetPlayModeName() string
}
```

#### PlayMode策略接口

```go
// PlayMode 播放模式策略接口
type PlayMode interface {
    // 下一首歌曲
    NextSong(current int, playlist []structs.Song, manual bool) (int, error)
    
    // 上一首歌曲
    PreviousSong(current int, playlist []structs.Song, manual bool) (int, error)
    
    // 初始化模式（用于需要特殊初始化的模式，如随机模式）
    Initialize(current int, playlist []structs.Song) error
    
    // 获取模式类型
    GetMode() types.Mode
    
    // 获取模式名称
    GetModeName() string
    
    // 处理播放列表变更（添加/删除歌曲时调用）
    OnPlaylistChanged(current int, playlist []structs.Song) error
}
```

### 具体实现

#### PlaylistManager实现

```go
type playlistManager struct {
    currentIndex int
    playlist     []structs.Song
    playMode     PlayMode
    playModes    map[types.Mode]PlayMode
    mu           sync.RWMutex
}
```

#### 播放模式实现

1. **OrderedPlayMode** - 顺序播放
2. **ListLoopPlayMode** - 列表循环
3. **SingleLoopPlayMode** - 单曲循环
4. **ListRandomPlayMode** - 列表随机
5. **InfiniteRandomPlayMode** - 无限随机

每个播放模式都实现`PlayMode`接口，专注于自己的播放逻辑。

## 数据模型

### 错误类型定义

```go
// 播放列表相关错误
var (
    ErrEmptyPlaylist     = errors.New("playlist is empty")
    ErrInvalidIndex      = errors.New("invalid song index")
    ErrInvalidPlayMode   = errors.New("invalid play mode")
    ErrNoNextSong        = errors.New("no next song available")
    ErrNoPreviousSong    = errors.New("no previous song available")
)

// PlaylistError 自定义错误类型
type PlaylistError struct {
    Op   string // 操作名称
    Err  error  // 原始错误
    Code int    // 错误代码
}

func (e *PlaylistError) Error() string {
    return fmt.Sprintf("playlist %s: %v", e.Op, e.Err)
}

func (e *PlaylistError) Unwrap() error {
    return e.Err
}
```

### 播放模式状态

```go
// RandomPlayState 随机播放状态
type RandomPlayState struct {
    Order       []int // 随机播放顺序
    CurrentPos  int   // 当前在随机序列中的位置
}

// InfiniteRandomState 无限随机播放状态
type InfiniteRandomState struct {
    History     []int // 播放历史
    MaxHistory  int   // 最大历史记录数
}
```

## 错误处理

### 错误处理策略

1. **标准Go错误处理**：使用`error`接口替代自定义可选类型 <mcreference link="https://www.jetbrains.com/guide/go/tutorials/handle_errors_in_go/best_practices/" index="2">2</mcreference>
2. **错误分类**：定义特定的错误类型用于不同的错误场景
3. **错误包装**：使用`fmt.Errorf`和`%w`动词包装错误以保留上下文 <mcreference link="https://medium.com/hprog99/mastering-error-handling-in-go-a-comprehensive-guide-fac34079833f" index="3">3</mcreference>
4. **优雅降级**：在遇到边界条件时提供合理的默认行为

### 边界条件处理

- **空播放列表**：返回`ErrEmptyPlaylist`错误
- **无效索引**：验证索引范围，返回`ErrInvalidIndex`错误
- **播放列表末尾**：根据播放模式决定行为（循环、停止等）
- **并发访问**：使用读写锁保护共享状态

## 测试策略

### 单元测试覆盖

1. **接口测试**：测试所有公共方法的正确性
2. **播放模式测试**：验证每种播放模式的行为
3. **边界条件测试**：测试空播放列表、边界索引等情况
4. **并发测试**：验证并发访问的安全性
5. **错误处理测试**：确保错误情况得到正确处理

### 测试结构

```go
// 测试套件结构
type PlaylistManagerTestSuite struct {
    manager  PlaylistManager
    testSongs []structs.Song
}

// 播放模式测试
func TestPlayModes(t *testing.T) {
    modes := []types.Mode{
        types.PmOrdered,
        types.PmListLoop,
        types.PmSingleLoop,
        types.PmListRandom,
        types.PmInfRandom,
    }
    
    for _, mode := range modes {
        t.Run(fmt.Sprintf("Mode_%v", mode), func(t *testing.T) {
            // 测试特定播放模式
        })
    }
}
```

### 模拟和依赖注入

使用接口和依赖注入使测试更容易：

```go
// 可测试的构造函数
func NewPlaylistManager(modes map[types.Mode]PlayMode) PlaylistManager {
    return &playlistManager{
        playModes: modes,
    }
}

// 测试中使用模拟播放模式
func TestWithMockPlayMode(t *testing.T) {
    mockMode := &MockPlayMode{}
    modes := map[types.Mode]PlayMode{
        types.PmOrdered: mockMode,
    }
    manager := NewPlaylistManager(modes)
    // 测试逻辑
}
```

## 迁移计划

### 阶段1：创建新包结构
1. 创建`internal/playlist`包
2. 定义核心接口
3. 实现基础的`PlaylistManager`

### 阶段2：实现播放模式
1. 实现各种播放模式策略
2. 添加单元测试
3. 验证功能完整性

### 阶段3：集成和重构
1. 修改UI层以使用新的播放列表管理器
2. 移除旧的`songManager`实现
3. 更新相关的调用代码

### 阶段4：优化和清理
1. 性能优化
2. 代码清理
3. 文档更新

## 性能考虑

1. **内存效率**：避免不必要的数据复制
2. **并发安全**：使用适当的锁机制
3. **算法优化**：随机播放使用高效的洗牌算法
4. **缓存策略**：缓存计算结果以提高性能

## 向后兼容性

在迁移过程中，将保持现有API的兼容性，通过适配器模式包装新的实现，确保现有代码能够继续工作。迁移完成后，将逐步废弃旧的接口。

## 总结

这个设计通过以下方式解决了现有问题：

1. **清晰的命名**：使用`PlaylistManager`替代`songManager`
2. **关注点分离**：将播放列表管理从UI层分离 <mcreference link="https://actvst.com/flutter/development/architecture/2025/02/20/decoupling-business-logic-from-ui.html" index="4">4</mcreference>
3. **策略模式**：简化播放模式的实现和切换
4. **标准错误处理**：使用Go标准错误处理模式
5. **可测试性**：通过接口和依赖注入提高可测试性
6. **可维护性**：清晰的架构和职责分离

新的设计将提供一个更加健壮、可维护和可扩展的播放列表管理系统。