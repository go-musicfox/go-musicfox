
# go-musicfox 测试文档

> 最后更新：2026-06-15 | 基于代码库当前状态

## 目录

1. [测试框架与工具](#1-测试框架与工具)
2. [测试文件组织](#2-测试文件组织)
3. [测试覆盖范围](#3-测试覆盖范围)
4. [测试类型](#4-测试类型)
5. [测试编写规范](#5-测试编写规范)
6. [Mock 与 Stub 方式](#6-mock-与-stub-方式)
7. [Benchmark 测试](#7-benchmark-测试)
8. [CI 与测试执行](#8-ci-与测试执行)
9. [测试质量评估](#9-测试质量评估)
10. [测试缺口分析](#10-测试缺口分析)

---

## 1. 测试框架与工具

### 标准库

项目使用 Go 标准库 `testing` 包，不引入第三方测试框架。

### 辅助工具

- **`testing.T`/`testing.B`**：单元测试和 benchmark
- **`t.Run()`**：子测试分组
- **`sync.WaitGroup`**：并发测试同步
- **`go test -coverprofile`**：覆盖率收集

### Makefile 测试命令

```makefile
# Makefile:60-63
test:
    go test ./internal/... ./utils/... \
        -coverpkg=./internal/...,./utils/... \
        -covermode=atomic -coverprofile=coverage.txt
```

- 覆盖范围：`./internal/...` 和 `./utils/...`
- 覆盖模式：atomic（并发安全）
- 输出：`coverage.txt`（可用于 CI 展示）

### Lint 工具

`.golangci.yml` 中 `tests: true` 表示 lint 也检查测试文件。

---

## 2. 测试文件组织

### 命名规则

- 测试文件：`<源文件名>_test.go`，与源文件同目录
- 包名：**与源文件相同**（white-box testing），不使用 `_test` 后缀包名

### 文件分布

```
internal/
├── macdriver/
│   ├── avcore/avplayer_test.go          # AVFoundation 播放器测试
│   ├── cocoa/nsapplication_test.go       # NSApplication 测试
│   ├── cocoa/nsimage_test.go             # NSImage 测试
│   ├── cocoa/nsworkspace_test.go         # NSWorkspace 测试
│   ├── cocoa/unnotifications_test.go     # 通知测试
│   ├── core/nsarray_test.go              # NSArray 测试
│   ├── core/nsmutabledictionary_test.go  # NSMutableDictionary 测试
│   ├── core/nsnumber_test.go             # NSNumber 测试
│   ├── core/nsstring_test.go             # NSString 测试
│   ├── core/nsurl_test.go                # NSURL 测试
│   ├── mediaplayer/mpchangeplaybackpositioncommand_test.go
│   ├── mediaplayer/mpchangeplaybackratecommand_test.go
│   ├── mediaplayer/mpchangerepeatmodecommand_test.go
│   ├── mediaplayer/mpchangeshufflemodecommand_test.go
│   ├── mediaplayer/mpfeedbackcommand_test.go
│   ├── mediaplayer/mpmediaitemartwork_test.go
│   ├── mediaplayer/mpnowplayinginfocenter_test.go
│   ├── mediaplayer/mpratingcommand_test.go
│   ├── mediaplayer/mpremotecommand_test.go
│   ├── mediaplayer/mpremotecommandcenter_test.go
│   └── mediaplayer/mpskipintervalcommand_test.go
├── player/
│   └── win_media_player_windows_test.go  # Windows 播放器测试
├── playlist/
│   ├── benchmark_test.go                 # 性能基准测试
│   ├── infinite_random_test.go           # 无限随机模式测试
│   ├── integration_test.go               # 集成测试
│   ├── intelligent_test.go               # 心动模式测试
│   ├── list_loop_test.go                 # 列表循环测试
│   ├── list_random_test.go               # 列表随机测试
│   ├── manager_test.go                   # 管理器测试
│   ├── ordered_test.go                   # 顺序播放测试
│   └── single_loop_test.go               # 单曲循环测试
└── remote_control/
    └── remote_command_handler_darwin_test.go
```

总计：**32 个测试文件**

---

## 3. 测试覆盖范围

### 有测试的模块

| 模块 | 测试文件数 | 覆盖程度 |
|------|-----------|---------|
| `macdriver/avcore` | 1 | 低 |
| `macdriver/cocoa` | 4 | 低 |
| `macdriver/core` | 5 | 中 |
| `macdriver/mediaplayer` | 11 | 中 |
| `player` | 1 | 极低 |
| `playlist` | 9 | **高** |
| `remote_control` | 1 | 低 |

### 无测试的模块

| 模块 | 说明 |
|------|------|
| `cmd/` | 应用入口，难以单元测试 |
| `internal/automator/` | 自动播放器 |
| `internal/commands/` | CLI 命令 |
| `internal/composer/` | 分享模板 |
| `internal/configs/` | 配置加载 |
| `internal/keybindings/` | 快捷键 |
| `internal/lastfm/` | API 客户端 |
| `internal/lyric/` | 歌词服务 |
| `internal/netease/` | 网易云错误 |
| `internal/reporter/` | 上报服务 |
| `internal/runtime/` | 运行时 |
| `internal/storage/` | BoltDB 存储 |
| `internal/structs/` | 数据模型 |
| `internal/track/` | 轨道管理 |
| `internal/types/` | 常量（无测试必要） |
| `internal/ui/` | TUI 界面（60+ 文件，无测试） |
| `utils/*` | 全部 utils 子包 |

---

## 4. 测试类型

### 单元测试

**playlist 包**是最完善的单元测试示例，针对每个播放模式：

1. **独立函数测试**：每个方法独立测试
   ```go
   func TestOrderedPlayMode_NextSong(t *testing.T) { ... }
   func TestOrderedPlayMode_PreviousSong(t *testing.T) { ... }
   func TestOrderedPlayMode_GetMode(t *testing.T) { ... }
   ```

2. **表驱动测试（Table-Driven Tests）**：在复杂场景中使用
   ```go
   // infinite_random_test.go:25-50
   tests := []struct {
       name         string
       currentIndex int
       playlist     []structs.Song
       wantErr      bool
   }{ ... }
   for _, tt := range tests {
       t.Run(tt.name, func(t *testing.T) { ... })
   }
   ```

3. **子测试（Subtests）**：使用 `t.Run()` 分组
   ```go
   // ordered_test.go:26-91
   t.Run("空播放列表", func(t *testing.T) { ... })
   t.Run("正常情况 - 从第一首到第二首", func(t *testing.T) { ... })
   t.Run("边界情况 - 最后一首歌的下一首", func(t *testing.T) { ... })
   ```

### 集成测试

`playlist/integration_test.go` 提供端到端集成测试：

```go
// TestPlaylistManagerIntegration 端到端集成测试
// 验证完整的播放流程和所有播放模式的正确性
func TestPlaylistManagerIntegration(t *testing.T) {
    // 测试所有播放模式的完整流程
    t.Run("OrderedPlayMode", func(t *testing.T) { ... })
    t.Run("ListLoopPlayMode", func(t *testing.T) { ... })
    t.Run("SingleLoopPlayMode", func(t *testing.T) { ... })
    t.Run("ListRandomPlayMode", func(t *testing.T) { ... })
    t.Run("InfiniteRandomPlayMode", func(t *testing.T) { ... })
}
```

集成测试覆盖：
- 完整播放流程（初始化 → 播放 → 切换 → 结束）
- 所有播放模式
- 状态一致性验证
- UI 层交互模拟（`TestPlaylistManagerWithUIInteraction`）

### 并发测试

```go
// manager_test.go:306
func TestConcurrentAccess(t *testing.T) {
    // 并发读取
    for i := 0; i < 10; i++ { go func() { ... }() }
    // 并发写入
    for i := 0; i < 5; i++ { go func() { ... }() }
    // 并发 NextSong/PreviousSong
    for i := 0; i < 3; i++ { go func() { ... }() }
    wg.Wait()
}
```

### Benchmark 测试

`playlist/benchmark_test.go` 包含 7 个 benchmark：

| Benchmark | 说明 |
|-----------|------|
| `BenchmarkPlaylistManagerInitialize` | 初始化 1000 首歌 |
| `BenchmarkPlaylistManagerNextSong` | 下一首操作 |
| `BenchmarkPlaylistManagerPreviousSong` | 上一首操作 |
| `BenchmarkPlaylistManagerSetPlayMode` | 切换播放模式 |
| `BenchmarkPlaylistManagerRemoveSong` | 删除歌曲 |
| `BenchmarkPlaylistManagerRandomMode` | 随机模式性能 |
| `BenchmarkPlaylistManagerLargePlaylist` | 10000 首歌大列表 |

---

## 5. 测试编写规范

### 命名

- 测试函数：`Test<Type>_<Method>` 或 `Test<Function>`
  - `TestOrderedPlayMode_NextSong`
  - `TestNewPlaylistManager`
  - `TestConcurrentAccess`
- Benchmark 函数：`Benchmark<Type><Method>`
  - `BenchmarkPlaylistManagerNextSong`

### 测试结构

1. **Arrange**：创建测试数据（辅助函数如 `createTestPlaylist(size)`）
2. **Act**：执行被测方法
3. **Assert**：使用 `t.Errorf` / `t.Fatalf` / `t.Error`

```go
// 辅助函数
func createTestPlaylist(size int) []structs.Song {
    playlist := make([]structs.Song, size)
    for i := 0; i < size; i++ {
        playlist[i] = structs.Song{
            Id:   int64(i + 1),
            Name: fmt.Sprintf("Song %d", i+1),
        }
    }
    return playlist
}
```

### 错误断言模式

```go
if err != nil {
    t.Errorf("unexpected error: %v", err)       // 非致命
    t.Fatalf("Initialize failed: %v", err)       // 致命（终止当前子测试）
}
if err == nil {
    t.Error("Expected error but got none")
}
```

### 测试的语言

测试函数名和断言消息混用英文（代码）和中文（描述）：
- 简短测试：英文 `"expected ErrEmptyPlaylist, got %v"`
- 中文描述：`"期望索引为0，实际为%d"`
- 子测试名：中文 `t.Run("空播放列表", ...)`

---

## 6. Mock 与 Stub 方式

### 当前状态

**项目不使用 mock 框架或 interface mock**（如 gomock、testify/mock）。

原因分析：
- `playlist` 包测试直接使用具体实现类型，因为播放模式是无依赖的纯逻辑
- `macdriver` 包测试使用真实 macOS API（仅在 macOS 上运行）

### 测试替身方式

1. **测试辅助函数**：`createTestPlaylist()` / `createTestSongs()` 创建测试固定数据
2. **直接构造 struct**：手工创建 `structs.Song{Id: 1, Name: "Song 1"}` 等测试数据
3. **类型断言访问内部状态**：
   ```go
   mode := NewIntelligentPlayMode().(*IntelligentPlayMode)  // 白盒测试
   mode.maxHistory = 3  // 直接修改内部字段
   ```

4. **storage 隔离**：`playlistManager.LoadState()` 使用 defer recover 在无 storage 环境友好降级

---

## 7. Benchmark 测试

### 执行方式

```bash
go test -bench=. ./internal/playlist/ -benchmem
```

### 测试场景

| Benchmark | 数据规模 | 操作 |
|-----------|---------|------|
| Initialize | 1000 首 | 创建管理器 + 初始化 |
| NextSong | 1000 首 | 获取下一首 |
| PreviousSong | 1000 首 | 获取上一首（中间位置） |
| SetPlayMode | 1000 首 | 循环切换 5 种模式 |
| RemoveSong | 1000 首 | 删除第 100 首 |
| RandomMode | 1000 首 | 随机取下一首 |
| LargePlaylist | 10000 首 | 综合操作 |

### Benchmark 最佳实践

- 使用 `b.ResetTimer()` 排除初始化开销
- 使用 `b.StopTimer()` / `b.StartTimer()` 排除重建设置开销
- 每个 benchmark 独立创建测试数据

---

## 8. CI 与测试执行

### 当前状态

**CI pipeline 不包含自动化测试步骤**。

`.github/workflows/release.yaml` 流程：
1. Checkout → Setup Go → Sync vendor
2. `go mod tidy` + `go mod vendor`
3. 运行 goreleaser dry-run 或 release
4. 无 `go test` 步骤

**pre-commit hook**（`githooks/pre-commit`）：
```sh
make lint-fix    # 仅 lint，不运行测试
git add .
```

### 可用的测试命令

开发者需要手动运行：
```bash
make test                                            # 完整测试 + 覆盖率
go test ./internal/playlist/ -v                      # 单包测试
go test -bench=. ./internal/playlist/               # 基准测试
```

---

## 9. 测试质量评估

### playlist 包（★★★★☆ 优秀）

| 维度 | 评分 | 说明 |
|------|------|------|
| 覆盖率 | ★★★★☆ | 核心逻辑全覆盖 |
| 边界条件 | ★★★★★ | 空列表、无效索引、越界、单元素 |
| 并发安全 | ★★★★☆ | 读写并发测试 |
| 错误路径 | ★★★★★ | 所有错误路径有对应测试 |
| 集成测试 | ★★★★★ | 完整流程 + 状态一致性 |
| 性能基准 | ★★★★★ | 7 个 benchmark，覆盖常规和大数据 |
| 表驱动 | ★★★★☆ | 复杂场景使用 |

**亮点**：
- `TestPlaylistManagerIntegration` 覆盖所有播放模式的完整流程
- `TestPlaylistManagerErrorHandling` 系统性地测试错误和边界
- `TestPlaylistManagerStateConsistency` 验证状态机正确性
- `TestConcurrentAccess` 验证并发安全（多 goroutine 读写）

### macdriver 包（★★☆☆☆ 基础）

| 维度 | 评分 | 说明 |
|------|------|------|
| 覆盖率 | ★★☆☆☆ | 仅基本创建/初始化测试 |
| 边界条件 | ★☆☆☆☆ | 缺少异常路径 |
| 平台依赖 | -- | 仅 macOS 可运行 |

### 其他包（★☆☆☆☆ 无测试）

`player/`、`storage/`、`ui/`、`configs/`、`lyric/` 等包**完全没有单元测试**。

---

## 10. 测试缺口分析

### 严重缺口

| 模块 | 风险 | 建议 |
|------|------|------|
| **`internal/player/`** | 多播放引擎无测试，播放错误难复现 | 编写 beep player 单元测试，mock 音频解码 |
| **`internal/storage/`** | BoltDB 存储层无测试，数据一致性风险 | 编写 CRUD 测试，使用临时数据库 |
| **`internal/configs/`** | 配置加载无测试，TOML 解析错误难发现 | 编写 loader 测试，使用 fixture TOML |
| **`internal/lyric/`** | LRC/YRC 解析无测试 | 编写解析器测试，使用 fixture 文件 |
| **`internal/ui/`** | TUI 界面无测试，交互逻辑靠手动验证 | 编写核心组件（player_controller、lyric_renderer）的单元测试 |

### 中度缺口

| 模块 | 建议 |
|------|------|
| `internal/track/` | 轨道管理逻辑可测试 |
| `internal/reporter/` | 上报逻辑可 mock |
| `internal/keybindings/` | 键绑定处理逻辑可单元测试 |
| `utils/errorx/` | 工具函数易于测试 |
| `utils/netease/` | 数据解析函数可测试（使用 fixture JSON） |

### 结构性改进建议

1. **引入 mock 框架**：对于依赖外部服务的包（storage、player、reporter），使用 `gomock` 或 `testify/mock` 隔离依赖
2. **CI 集成测试**：在 release workflow 中添加 `make test` 步骤
3. **测试覆盖率门槛**：设置最低覆盖率要求（如 60%），防止覆盖率退步
4. **接口抽象**：为 storage、player 等层引入接口，方便 mock 测试
5. **多平台测试矩阵**：在 CI 中测试 macOS/Linux/Windows 交叉编译，确保平台特定代码正确
6. **fixture 文件**：`testdata/` 目录已有但未充分利用，可添加歌词文件、配置文件等测试固定数据
