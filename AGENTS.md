# go-musicfox 项目架构文档

## 项目概述

go-musicfox 是一个基于 Go 语言开发的网易云音乐命令行客户端，支持 macOS、Linux 和 Windows 平台。应用采用 TUI（文本用户界面）架构，提供了丰富的音乐播放功能。

### 核心特性

- **多平台支持**：macOS、Linux、Windows
- **多种音频引擎**：Beep（默认）、MPV、MPD、macOS AVFoundation、Windows Media Player
- **音频格式支持**：MP3、FLAC、OGG、WAV
- **歌词显示**：支持 LRC 和 YRC（逐字歌词）格式
- **播放模式**：列表循环、顺序播放、单曲循环、随机播放、无限随机、智能心动模式
- **系统集成**：MPRIS（Linux）、Now Playing（macOS）、系统媒体控制（Windows）
- **Last.fm 集成**：播放记录和收藏功能
- **UnblockNeteaseMusic**：解锁灰色歌曲

## 项目结构

```
go-musicfox/
├── cmd/
│   └── musicfox.go              # 程序入口点
├── internal/                    # 核心业务逻辑（22个包）
│   ├── automator/               # 自动播放功能
│   ├── commands/                # CLI 命令定义
│   ├── composer/                # 文件命名和分享模板
│   ├── configs/                 # 配置管理
│   ├── keybindings/             # 快捷键配置
│   ├── lastfm/                  # Last.fm 集成
│   ├── lyric/                   # 歌词解析服务
│   ├── macdriver/               # macOS 原生集成
│   ├── migration/               # 数据迁移
│   ├── netease/                 # 网易云音乐 API 封装
│   ├── player/                  # 音频播放引擎
│   ├── playlist/                # 播放列表管理
│   ├── remote_control/          # 远程控制集成
│   ├── reporter/                # 播放报告服务
│   ├── runtime/                 # 平台运行时
│   ├── storage/                 # 本地存储
│   ├── structs/                 # 数据模型
│   ├── track/                   # 歌曲管理
│   └── types/                   # 类型定义
├── utils/                       # 工具函数包
├── v2/                          # v2 开发工作区
├── configs/                     # 嵌入式配置
└── vendor/                      #  vendored 依赖
```

## 核心架构

### 1. 应用入口与初始化

**入口文件**：`cmd/musicfox.go`

```go
func main() {
    runtime.Run(musicfox)  // 平台特定运行时
}

func musicfox() {
    app := gcli.NewApp()
    // 1. 加载配置
    loadConfig()
    // 2. 检查并执行数据迁移
    if needsMigration { migration.RunAndReport() }
    // 3. 注册命令并运行
    app.Add(playerCommand)
    app.Run()
}
```

**初始化流程**：
1. 加载配置（TOML 或迁移自旧版 INI）
2. 检查数据迁移需求
3. 初始化 Netease 应用
4. 设置网易云音乐 API 参数
5. 启动 TUI

### 2. UI 架构（bubbletea）

**核心协调器**：`internal/ui/netease.go`

```go
type Netease struct {
    *model.App                    // foxful-cli 应用框架
    login      *LoginPage          // 登录页面
    search     *SearchPage         // 搜索页面
    player     *Player             // 播放器控制器
    lyricService *lyric.Service    // 歌词服务
    lyricRenderer *LyricRenderer   // 歌词渲染器
    progressRenderer *ProgressRenderer  // 进度条渲染器
    songInfoRenderer *SongInfoRenderer  // 歌曲信息渲染器
    trackManager *track.Manager    // 歌曲管理器
}
```

**组件渲染顺序**：
1. 主菜单渲染
2. 歌曲信息栏
3. 播放进度条
4. 歌词显示

### 3. 菜单系统

**核心接口**（`internal/ui/menu.go`）：
```go
type Menu interface {
    model.Menu  // 基础菜单接口
    IsPlayable() bool
    IsLocatable() bool
}

type SongsMenu interface { Songs() []structs.Song }
type PlaylistsMenu interface { Playlists() []structs.Playlist }
type AlbumsMenu interface { Albums() []structs.Album }
type ArtistsMenu interface { Artists() []structs.Artist }
```

**主要菜单类型**（72个 UI 文件）：
- `MenuMain` - 主菜单
- `LoginPage` - 登录页面
- `SearchPage` - 搜索页面
- `CurPlaylist` - 当前播放列表
- `PlaylistDetailMenu` - 歌单详情
- `AlbumDetailMenu` - 专辑详情
- `ArtistDetailMenu` - 歌手详情
- 等等...

### 4. 事件处理

**事件处理器**：`internal/ui/event_handler.go`

- **键盘事件**：40+ 操作映射
- **鼠标事件**：单击、双击、滚轮、右键
- **快捷键系统**：可配置 + 内置

**内置快捷键**：
| 按键 | 功能 |
|------|------|
| `h/j/k/l` 或方向键 | 导航 |
| `g` / `G` | 跳至顶部/底部 |
| `n` / `Enter` | 进入 |
| `b` / `Esc` | 返回 |
| `/` | 搜索 |
| `q` | 退出 |
| `r` | 重新渲染 |

### 5. 音频播放引擎

**核心接口**（`internal/player/player.go`）：
```go
type Player interface {
    Play(music URLMusic)
    Pause()/Resume()/Stop()/Toggle()
    Seek(duration time.Duration)
    PassedTime()/PlayedTime() time.Duration
    TimeChan() <-chan time.Duration
    State() types.State
    StateChan() <-chan types.State
    Volume()/SetVolume()/UpVolume()/DownVolume()
    Close()
}
```

**引擎实现**：

| 引擎 | 文件 | 平台 | 特点 |
|------|------|------|------|
| **Beep** | `beep_player.go` | 跨平台 | 默认，依赖 go-mp3/go-flac |
| **MPV** | `mpv_player.go` | 跨平台 | IPC 控制，音视频分离 |
| **MPD** | `mpd_player.go` | Linux | 远程 MPD 服务器 |
| **AVFoundation** | `osx_player_darwin.go` | macOS | 原生系统集成 |
| **MediaPlayer** | `win_media_player_windows.go` | Windows | WinRT API |

**音频格式支持**（`beep_decoder.go`）：
- MP3（支持 minimp3 轻量解码器）
- FLAC（无损）
- OGG/Vorbis
- WAV（PCM）

### 6. 播放列表管理

**播放模式**（`internal/playlist/`）：

| 模式 | 文件 | 说明 |
|------|------|------|
| 列表循环 | `list_loop.go` | 循环播放整个列表 |
| 顺序播放 | `ordered.go` | 按顺序播放 |
| 单曲循环 | `single_loop.go` | 重复当前歌曲 |
| 随机播放 | `list_random.go` | 随机打乱列表 |
| 无限随机 | `infinite_random.go` | 持续随机播放 |
| 智能模式 | `intelligent.go` | 网易云心动模式 |

### 7. 歌词系统

**歌词服务**：`internal/lyric/service.go`

**支持格式**：
- **LRC**：标准时间轴歌词
- **YRC**：逐字歌词（带精确时间戳）

**渲染模式**（`lyric_renderer.go`）：
- 平滑（smooth）
- 波浪（wave）
- 发光（glow）

### 8. 网易云音乐 API 集成

**API 客户端**：使用 `github.com/go-musicfox/netease-music` 包

**认证方式**：
- 手机号登录
- 邮箱登录
- QR 码登录
- Cookie 持久化

**主要 API**：
- 歌曲详情/播放链接
- 歌单管理
- 歌手/专辑搜索
- 每日推荐
- 用户收藏

### 9. 本地存储

**数据库**：BoltDB（`internal/storage/`）

**存储内容**：
- 用户信息
- 播放状态（当前歌曲、播放模式、音量）
- 播放列表快照
- Last.fm 凭证
- 最后登录日期

### 10. 系统集成

**远程控制**（`internal/remote_control/`）：

| 平台 | 实现 | 功能 |
|------|------|------|
| Linux | `mpris_player_linux.go` | MPRIS D-Bus |
| macOS | `remote_control_darwin.go` | Now Playing、Remote Command Center |
| Windows | `remote_control_windows.go` | System Media Transport |

**macDriver**（`internal/macdriver/`）：
- Cocoa 框架集成
- 通知中心
- 工作空间
- 睡眠/唤醒检测

## 关键文件路径

### 入口与配置
| 文件 | 功能 |
|------|------|
| `cmd/musicfox.go` | 程序入口 |
| `internal/configs/config.go` | 配置结构 |
| `internal/types/constants.go` | 常量定义 |

### UI 组件
| 文件 | 功能 |
|------|------|
| `internal/ui/netease.go` | UI 协调器 |
| `internal/ui/event_handler.go` | 事件处理 |
| `internal/ui/menu_*.go` | 各菜单页面 |
| `internal/ui/*_renderer.go` | 渲染器 |

### 播放引擎
| 文件 | 功能 |
|------|------|
| `internal/player/player.go` | 接口与工厂 |
| `internal/player/beep_player.go` | Beep 引擎 |
| `internal/player/mpv_player.go` | MPV 引擎 |
| `internal/player/mpd_player.go` | MPD 引擎 |

### 数据管理
| 文件 | 功能 |
|------|------|
| `internal/playlist/manager.go` | 播放列表 |
| `internal/track/manager.go` | 歌曲管理 |
| `internal/storage/local_db.go` | 本地存储 |

## 开发指南

### 添加新菜单

1. 创建 `internal/ui/menu_new_feature.go`
2. 嵌入 `baseMenu`
3. 实现 `Menu` 接口
4. 注册到导航系统

### 添加新播放器引擎

1. 实现 `internal/player.Player` 接口
2. 在 `player.go:NewPlayerFromConfig()` 添加 case
3. 添加配置支持

### 修改快捷键

1. 在 `internal/keybindings/keybindings.go` 定义 `OperateType`
2. 在 `event_handler.go` 添加键映射
3. 在配置文件中添加自定义绑定

### 添加新渲染器

1. 实现 `Update()` 和 `View()` 方法
2. 在 `netease.go:Components()` 注册

## 文档维护准则

### 重要准则：修改代码后需维护 AGENTS.md

**所有贡献者在修改代码后，必须检查并更新 AGENTS.md 文档，防止文档腐化。**

#### 1. 何时需要更新文档
- 添加、删除或重命名核心文件或目录
- 新增功能模块或组件
- 修改项目结构或架构
- 更改 API 接口或配置格式
- 添加新的播放器引擎、菜单类型、渲染模式
- 修改关键路径或重要流程

#### 2. 更新检查清单
- [ ] 目录结构是否准确反映当前项目结构
- [ ] 核心文件路径是否正确
- [ ] 接口定义是否与代码一致
- [ ] 新增功能的说明是否完整
- [ ] 开发指南是否需要补充
- [ ] 关键文件路径表格是否需要更新

#### 3. 更新优先级
| 变更类型 | 优先级 | 说明 |
|---------|--------|------|
| 架构变更 | **高** | 必须立即更新 |
| 新增核心模块 | **高** | 必须添加说明 |
| 文件路径变更 | **中** | 及时更新路径表 |
| 细微修复 | **低** | 可批量更新 |

#### 4. 维护建议
- 保持文档与代码同步，避免技术债务积累
- 使用一致的术语和格式
- 添加代码示例时确保与实际代码匹配
- 定期（如每月）审查文档完整性
- PR 审查时应包含文档检查

#### 5. 文档腐化警告信号
- 文件路径与实际不符
- 过时的 API 接口描述
- 已删除功能仍出现在文档中
- 章节结构混乱或重复
- 与 README/CHANGELOG 存在矛盾

违反此准则可能导致：
- 新贡献者无法理解项目结构
- 开发效率降低
- 文档失去参考价值
- 维护成本增加

## 配置文件

**路径**：
- macOS: `~/Library/Application Support/go-musicfox/config.toml`
- Linux: `~/.config/go-musicfox/config.toml`
- Windows: `%AppData%\go-musicfox\config.toml`

**主要配置段**：
```toml
[startup]      # 启动选项
[main]         # 主界面设置
[theme]        # 主题颜色
[storage]      # 存储路径
[player]       # 播放器配置
[autoplay]     # 自动播放
[unm]          # UnblockNeteaseMusic
[reporter]     # 播放报告
[keybindings]  # 快捷键
[share]        # 分享模板
```

## 构建与发布

**构建命令**：
```sh
make build      # 编译到 bin/ 目录
make install    # 安装到 $GOPATH/bin
make test       # 运行测试
make lint       # 代码检查
```

**发布**：
- 使用 GoReleaser
- 支持跨平台编译
- 提供 Homebrew、Scoop、Flatpak 等包管理器支持

## 技术栈

- **UI 框架**：bubbletea + foxful-cli
- **音频处理**：beep、go-mp3、go-flac
- **存储**：BoltDB
- **配置**：TOML + mapstructure
- **API**：netease-music SDK
- **系统集成**：CGO、WinRT、D-Bus

---

这份文档涵盖了 go-musicfox 的完整架构，包括新增的**文档维护准则**，可以帮助开发者和贡献者快速理解项目结构并开始开发工作，同时确保文档与代码保持同步，避免腐化。
