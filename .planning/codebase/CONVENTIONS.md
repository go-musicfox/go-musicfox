
# go-musicfox 代码规范文档

> 最后更新：2026-06-15 | 基于代码库当前状态

## 目录

1. [Go 版本与模块](#1-go-版本与模块)
2. [项目结构规范](#2-项目结构规范)
3. [命名规范](#3-命名规范)
4. [代码格式化](#4-代码格式化)
5. [Import 组织](#5-import-组织)
6. [注释与文档规范](#6-注释与文档规范)
7. [错误处理模式](#7-错误处理模式)
8. [接口设计规范](#8-接口设计规范)
9. [并发编程规范](#9-并发编程规范)
10. [配置管理规范](#10-配置管理规范)
11. [Git 提交规范](#11-git-提交规范)
12. [项目特定约定](#12-项目特定约定)

---

## 1. Go 版本与模块

- **Go 版本**：`go 1.24.0`（go.mod 声明）
- **模块路径**：`github.com/go-musicfox/go-musicfox`
- **依赖管理**：Go Modules + vendoring（`vendor/` 目录随代码提交）
- **replace 指令**：多处 fork 依赖的 replace（如 bubbletea、beep、gohook 等）

示例（`go.mod:99-113`）：
```go
replace (
    github.com/charmbracelet/bubbletea v0.25.0 => github.com/go-musicfox/bubbletea v0.25.0-foxful
    github.com/cocoonlife/goflac v0.0.0-20170210142907-50ea06ed5a9d => github.com/go-musicfox/goflac v0.1.5
    // ...
)
```

---

## 2. 项目结构规范

### 顶层目录

| 目录 | 用途 | 说明 |
|------|------|------|
| `cmd/` | 应用入口 | 仅 `musicfox.go`，`main()` 调用 `runtime.Run()` |
| `internal/` | 核心业务 | 22 个子包，外部不可 import |
| `utils/` | 工具库 | 可对外暴露，按功能分子包 |
| `configs/` | 嵌入配置 | 默认 TOML 配置文件 |
| `vendor/` | 依赖 vendoring | 随代码提交 |

### internal 子包组织

```
internal/
├── automator/     # 自动播放
├── commands/      # CLI 命令定义
├── composer/      # 分享文本模板
├── configs/       # 配置结构体与加载
├── keybindings/   # 快捷键定义
├── lastfm/        # Last.fm API
├── lyric/         # 歌词服务
├── macdriver/     # macOS 原生 API 封装
├── netease/       # 网易云 API 错误类型
├── player/        # 播放引擎（多平台多引擎）
├── playlist/      # 播放列表管理（策略模式）
├── remote_control/# 远程控制（MPRIS/NowPlaying）
├── reporter/      # 播放上报（Last.fm/网易）
├── runtime/       # 运行时初始化
├── storage/       # BoltDB 存储层
├── structs/       # 数据模型
├── track/         # 音频轨道管理
├── types/         # 全局常量与类型
└── ui/            # TUI 界面（60+ 文件，最大包）
```

### 文件组织约定

- **一个 struct/接口 = 一个文件**：如 `internal/playlist/ordered.go` 只含 `OrderedPlayMode`
- **平台特定文件**：使用 `_darwin.go`、`_linux.go`、`_windows.go` 后缀
  - 示例：`osx_player_darwin.go`、`win_media_player_windows.go`
- **测试文件**：与源文件同目录，`_test.go` 后缀
- **菜单文件命名**：`menu_<功能>.go`（如 `menu_main.go`、`menu_album_list.go`）

---

## 3. 命名规范

### 包命名

- **全小写**，不使用下划线或驼峰
- 简短、描述性强：`configs`、`structs`、`player`、`storage`
- 特殊：`utils/_struct` 因与关键字冲突使用 underscore 前缀
- 注意 `package _struct` 在代码中使用 `_struct.XXX` 引用

### 文件命名

- 全小写 + 下划线分隔：`beep_player.go`、`event_handler.go`
- 测试文件：`原文件名_test.go`
- 平台特定：`文件名_GOOS.go` 或 `文件名_GOOS_GOARCH.go`

### 类型命名

- **接口**：单方法接口常用 `-er` 后缀；多方法接口直接名词如 `Player`、`Menu`
  ```go
  type Player interface { ... }    // internal/player/player.go:13
  type PlaylistManager interface { ... } // internal/playlist/interfaces.go:10
  type Model interface { ... }      // internal/storage/model.go:3
  ```
- **结构体**：PascalCase，如 `Netease`、`ListLoopPlayMode`、`LocalDBManager`
- **私有结构体**：小写开头，如 `baseMenu`、`playlistManager`

### 函数命名

- **导出函数**：PascalCase，如 `NewPlaylistManager()`、`Play()`
- **未导出函数**：camelCase，如 `saveStateAsync()`、`registerPlayModes()`
- **构造函数**：统一使用 `NewXxx()` 模式
  ```go
  func NewPlaylistManager() PlaylistManager       // internal/playlist/manager.go:26
  func NewListLoopPlayMode() PlayMode             // internal/playlist/list_loop.go:13
  func NewConfigFromTomlFile(tomlPath string) (*Config, error) // internal/configs/loader.go:25
  ```

### 变量命名

- 包级变量：`AppConfig`、`DBManager`（全局变量）
- 局部变量：短名称（Go 惯例），如 `err`、`cfg`、`pm`
- 常量：PascalCase，如 `AppName`、`BeepPlayer`、`MaxPlayErrCount`
- 错误变量：`Err` 前缀，如 `ErrEmptyPlaylist`、`ErrInvalidIndex`

### 常量

定义在 `internal/types/constants.go`（全局常量）和各包内：
```go
const AppName = "musicfox"     // types/constants.go:15
const BeepPlayer = "beep"      // types/constants.go:42
const MaxPlayErrCount = 3       // types/constants.go:52
```

---

## 4. 代码格式化

### 基础工具

- **gofmt**：标准 Go 格式化
- **goimports**：import 自动管理
- **gci**：import 分组排序

### golangci-lint 配置

`.golangci.yml` 配置了：

**Linters（代码检查）**：
- `govet` - Go 官方检查
- `errcheck` - 错误检查
- `ineffassign` - 无效赋值检测
- `staticcheck` - 静态分析（all checks，排除 SA4006、SA1029）
- `unused` - 未使用代码检测

**Formatters（格式化）**：
- `gci` - import 分组（标准库 → 第三方 → go-musicfox）
- `gofmt` - 代码格式化
- `goimports` - import 管理

**Pre-commit hook**（`githooks/pre-commit`）：
```sh
make lint-fix
git add .
```

### Import 分组规则

按以下顺序分组，组间空行：
1. **标准库**
2. **第三方库**
3. **go-musicfox 内部包**（前缀 `github.com/go-musicfox/go-musicfox`）

示例（`internal/playlist/manager.go:3-14`）：
```go
import (
    "encoding/json"
    "log/slog"
    "maps"
    "slices"
    "sync"
    "time"

    "github.com/go-musicfox/go-musicfox/internal/storage"
    "github.com/go-musicfox/go-musicfox/internal/structs"
    "github.com/go-musicfox/go-musicfox/internal/types"
)
```

---

## 5. 注释与文档规范

### 注释语言

根据 `AGENTS.md` 规定：
- **代码注释**：使用英文
- **用户交互/文档**：使用中文
- **Git Commit Message**：使用英文

### 注释风格

- **包注释**：不需要（Go 1.22+）
- **导出类型/函数**：应有文档注释
  ```go
  // PlaylistManager 播放列表管理器接口
  // 提供播放列表的核心管理功能，包括播放控制、模式切换等
  type PlaylistManager interface { ... }
  ```
- **接口方法**：应有清晰的中文注释
  ```go
  // NextSong 切换到下一首歌曲
  // manual 参数表示是否为手动切换
  NextSong(manual bool) (structs.Song, error)
  ```
- **结构体字段**：可选中英文注释
- **实现细节注释**：英文，简洁
  ```go
  // 列表循环播放模式无需特殊初始化逻辑
  func (l *ListLoopPlayMode) Initialize(...) error { ... }
  ```

### Deprecated 标记

使用标准格式：
```go
// GetPlayModeName 获取当前播放模式的名称
//
// Deprecated: please use GetPlayMode().Name() instead.
GetPlayModeName() string
```

---

## 6. 错误处理模式

### 错误类型设计

**自定义错误结构体**（`internal/playlist/errors.go`）：
```go
type PlaylistError struct {
    Op  string // 操作名称
    Err error  // 底层错误
}

func (e *PlaylistError) Error() string {
    if e.Op == "" {
        return e.Err.Error()
    }
    return fmt.Sprintf("playlist %s: %v", e.Op, e.Err)
}

func (e *PlaylistError) Unwrap() error {
    return e.Err
}
```

### 哨兵错误（Sentinel Errors）

在包级别定义标准错误变量：
```go
var (
    ErrEmptyPlaylist    = errors.New("playlist is empty")
    ErrInvalidIndex     = errors.New("invalid index")
    ErrInvalidPlayMode  = errors.New("invalid play mode")
    ErrNoNextSong       = errors.New("no next song")
    ErrNoPreviousSong   = errors.New("no previous song")
)
```

### 错误包装

两种模式并存：

1. **自定义包装**（playlist 包）：`newPlaylistError("next song", err)`
2. **pkg/errors**（configs/loader.go）：
   ```go
   errors.Wrap(err, "failed to read embedded default config")
   errors.Wrapf(err, "error loading TOML config file '%s'", tomlPath)
   ```

### 错误处理辅助

`utils/errorx/` 提供了：

- `Must(err)` - panic on error
- `Must1[T any](a T, err error) T` - 泛型 Must
- `Must2[T, S any](a T, b S, err error) (T, S)` - 双返回值 Must
- `Recover(ignore bool)` - panic 恢复
- `Go(f func(), ignorePanic ...bool)` - goroutine + panic 恢复
- `WaitGoStart(f func(), ignorePanic ...bool)` - 同步等待 goroutine 启动

### 错误处理风格

- 总是检查 error 返回值
- 使用 `if err != nil` 模式
- 关键路径使用 panic（如配置加载失败 `cmd/musicfox.go:100`）
- 异步路径使用 `defer recover` + `slog.Error`

---

## 7. 接口设计规范

### 接口隔离

接口保持小而专注：
- `Player` 接口（`internal/player/player.go:13-31`）：全部播放控制方法
- `PlaylistManager` 接口：播放列表管理方法
- `PlayMode` 接口（策略模式）：单个播放模式行为
- `Menu` 接口：嵌入 `model.Menu` + 扩展方法

### 接口组合

```go
type SongsMenu interface {
    Menu
    Songs() []structs.Song
}
```

### 构造函数返回接口

```go
func NewPlaylistManager() PlaylistManager { ... }
func NewListLoopPlayMode() PlayMode { ... }
func NewPlayerFromConfig() Player { ... }
```

### 策略模式

`internal/playlist/` 是典型的策略模式实现：
- `PlayMode` 接口定义策略
- `OrderedPlayMode`、`ListLoopPlayMode`、`SingleLoopPlayMode` 等具体实现
- `playlistManager` 持有当前策略并通过 `SetPlayMode()` 切换

---

## 8. 并发编程规范

### RWMutex 保护

多 goroutine 访问的数据结构使用 `sync.RWMutex`：
```go
type playlistManager struct {
    mu           sync.RWMutex
    currentIndex int
    playlist     []structs.Song
    playMode     PlayMode
}
```

- 读操作：`mu.RLock(); defer mu.RUnlock()`
- 写操作：`mu.Lock(); defer mu.Unlock()`
- 拷贝返回：避免竞态，读操作返回数据副本

### Goroutine 管理

- 使用 `errorx.Go(f, ignorePanic)` 启动 goroutine
- 内部 goroutine 使用 `defer recover` 保护

---

## 9. 配置管理规范

### 配置框架

使用 Koanf + TOML：
```go
// internal/configs/loader.go
k := koanf.New(".")
k.Load(rawbytes.Provider(defaultTomlBytes), toml.Parser())
k.Load(file.Provider(tomlPath), toml.Parser())
k.UnmarshalWithConf("", finalConfig, unmarshalConf)
```

### 配置结构体

- 使用 `koanf` struct tag 映射配置键
- mapstructure decode hooks 处理类型转换
```go
type Config struct {
    Startup     StartupConfig     `koanf:"startup"`
    Main        MainConfig        `koanf:"main"`
    Theme       ThemeConfig       `koanf:"theme"`
    Player      PlayerConfig      `koanf:"player"`
    // ...
}
```

### 全局配置变量

```go
var AppConfig *Config                                       // internal/configs/loader.go:19
var EffectiveKeybindings map[keybindings.OperateType][]string // internal/configs/loader.go:22
```

---

## 10. Git 提交规范

### Conventional Commits

严格遵循 Conventional Commits 规范（`AGENTS.md:249-296`）：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Type 类型**：
| Type | 说明 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `docs` | 文档更新 |
| `style` | 格式调整 |
| `refactor` | 代码重构 |
| `perf` | 性能优化 |
| `test` | 测试相关 |
| `chore` | 构建/工具/配置 |
| `revert` | 回滚 |

**示例**：
```
feat(player): 添加 MPV 播放引擎支持

- 支持多种音频格式
- 实现播放进度控制

Closes #123
```

**要求**：英文、标题 ≤ 50 字符、body 每行 ≤ 72 字符

### Commit Hook

Pre-commit hook 自动运行 `make lint-fix`（`githooks/pre-commit`）

### PR 规范

CI 检查（`.github/workflows/pr-standards.yml`）：
- PR 标题必须符合 Conventional Commits 格式
- fix/chore/test 类型 PR 需要关联 issue
- PR 描述必须包含完整模板内容

---

## 11. 项目特定约定

### 菜单模式

所有菜单文件遵循统一模式：
- 嵌入 `baseMenu`（`internal/ui/menu.go:45-48`）
- 实现 `Menu` 接口
- 文件命名：`menu_<功能>.go`
- 构造函数通过 `netease *Netease` 获取依赖

### 播放模式策略

- 新增播放模式实现 `PlayMode` 接口
- 在 `manager.go:registerPlayModes()` 注册
- 提供对应测试文件

### 平台特定代码

- macOS：`_darwin.go` 文件（`macdriver/`、`osx_player_darwin.go`）
- Linux：`_linux.go` 文件（MPRIS、全局快捷键）
- Windows：`_windows.go` 文件（WinRT MediaPlayer）
- 跨平台公共代码不使用后缀

### 日志

- 使用 `log/slog` 结构化日志
- `utils/slogx/slog.go` 提供日志初始化
- 错误路径：`slog.Error("message", slog.Any("key", val))`
- 格式化：避免使用 `fmt.Sprintf` 构建日志消息

### embed 文件

- 使用 `//go:embed` 嵌入默认配置文件
- `filex.ReadFileFromEmbed()` 读取嵌入资源

---

## 现状评估

### 优势

1. **一致的命名规范**：PascalCase/camelCase 严格执行，函数命名语义清晰
2. **良好的接口抽象**：Player/PlayMode/Menu 接口设计合理，策略模式应用恰当
3. **完整的错误处理**：哨兵错误 + 自定义错误类型 + 错误包装，层次分明
4. **规范的并发控制**：RWMutex 保护共享状态，goroutine recover 保护
5. **严格的 PR/Commit 规范**：CI 自动检查 Conventional Commits 和模板完整性

### 改进机会

1. **注释语言混用**：部分注释使用中文（接口方法），部分使用英文（实现细节），缺乏一致性
2. **配置结构体过大**：`Config` 聚合了所有子配置，可考虑更扁平的结构
3. **import 分组**：部分文件未严格遵循 gci 规则（标准库/第三方/内部 三组）
4. **错误变量命名**：个别文件使用 `errors.New()` 返回局部变量而非包级哨兵错误
