# go-musicfox 项目结构文档

> 生成时间: 2026-06-15 | 代码版本: v3.7.0 | Go 版本: 1.24.0

## 概述

go-musicfox 是一个标准的 Go 项目，模块名为 `github.com/go-musicfox/go-musicfox`。项目采用"约定优于配置"的目录布局。

---

## 1. 顶层目录布局

| 目录/文件 | 用途 | 类型 |
|-----------|------|------|
| `cmd/` | 应用入口 (`musicfox.go`) | Go 源码 |
| `internal/` | 核心业务代码 (19 个子包) | Go 源码 |
| `utils/` | 通用工具函数 (15 个子包) | Go 源码 |
| `configs/` | 内嵌默认配置文件 | 配置资源 |
| `external/` | 外部辅助应用 (macOS notifier) | 辅助代码 |
| `docs/` | 项目文档 | 文档 |
| `previews/` | 预览截图/GIF | 静态资源 |
| `deploy/` | 部署配置 | 部署 |
| `hack/` | 构建辅助脚本 | 脚本 |
| `script/` | 打包脚本 | 脚本 |
| `githooks/` | Git hooks | 配置 |
| `sysroot/` | 系统根文件 | 系统资源 |
| `bin/` | 编译输出 | 构建产物 |
| `dist/` | 发布包 | 构建产物 |
| `vendor/` | Go vendor 目录 | 依赖缓存 |
| `testdata/` | 测试数据 | 测试资源 |
| `.github/` | GitHub Actions/模板 | CI/CD |
| `.kiro/` | Kiro 配置 | 工具配置 |
| `.opencode/` | OpenCode 配置 | AI 工具 |
| `.claude/` | Claude Code 规则 | AI 工具 |
| `.rules/` | 项目规则 | AI 规则 |
| `.vscode/` | VS Code 配置 | IDE 配置 |
| `Makefile` | 构建命令 | 构建 |
| `go.mod` / `go.sum` | Go 模块定义 | 依赖管理 |
| `flake.nix` / `flake.lock` | Nix 包定义 | 包管理 |
| `.goreleaser.yaml` | GoReleaser 配置 | 发布 |
| `.golangci.yml` | golangci-lint 配置 | 代码质量 |
| `AGENTS.md` | AI 助手文档 | 文档 |
| `CHANGELOG.md` | 变更日志 | 文档 |
| `README.md` | 项目说明 | 文档 |

---

## 2. 包依赖关系

### 外部核心依赖

| 依赖 | 版本 | 用途 |
|------|------|------|
| `github.com/go-musicfox/bubbletea` | v0.25.0-foxful | TUI 框架 (fork) |
| `github.com/anhoder/foxful-cli` | v0.5.0 | TUI UI 抽象层 |
| `github.com/go-musicfox/netease-music` | v1.6.0 | 网易云音乐 API SDK |
| `github.com/go-musicfox/beep` | v1.4.1 | 音频播放库 (fork) |
| `github.com/gopxl/beep` | v1.4.0 | (被 fork 替换) |
| `github.com/go-musicfox/gohook` | v0.41.1 | 全局热键 (fork) |
| `github.com/knadh/koanf` | v2.3.0 | 配置解析 (TOML) |
| `go.etcd.io/bbolt` | v1.3.7 | 嵌入式 KV 数据库 |
| `github.com/gookit/gcli/v2` | v2.3.4 | CLI 框架 |
| `github.com/charmbracelet/lipgloss` | v0.8.0 | TUI 样式 |
| `github.com/charmbracelet/bubbles` | v0.16.1 | TUI 组件库 |
| `github.com/shkh/lastfm-go` | v0.0.0 | (被 fork 替换) |
| `github.com/skratchdot/open-golang` | v0.0.0 | 浏览器打开 |
| `github.com/juju/persistent-cookiejar` | v1.0.0 | Cookie 持久化 |
| `github.com/gen2brain/beeep` | v0.0.0 | 桌面通知 |

### Fork 关系

项目对多个上游库进行了 fork 以满足定制需求:

```
github.com/charmbracelet/bubbletea → github.com/go-musicfox/bubbletea
github.com/gopxl/beep              → github.com/go-musicfox/beep
github.com/robotn/gohook           → github.com/go-musicfox/gohook
github.com/shkh/lastfm-go          → github.com/go-musicfox/lastfm-go
github.com/hajimehoshi/go-mp3      → github.com/go-musicfox/go-mp3
github.com/gookit/gcli/v2          → github.com/anhoder/gcli/v2
```

### internal 包内部依赖关系

```
cmd/musicfox.go
  ├── internal/commands     (CLI 命令注册)
  │     ├── internal/ui     (TUI 核心)
  │     │     ├── internal/player      (播放引擎)
  │     │     ├── internal/playlist    (播放列表管理)
  │     │     ├── internal/lyric       (歌词服务)
  │     │     ├── internal/track       (音轨管理)
  │     │     ├── internal/composer    (文件名生成/分享)
  │     │     ├── internal/storage     (持久化)
  │     │     ├── internal/structs     (数据模型)
  │     │     ├── internal/types       (常量/类型)
  │     │     ├── internal/configs     (配置)
  │     │     ├── internal/keybindings (按键映射)
  │     │     ├── internal/lastfm      (Last.fm 客户端)
  │     │     ├── internal/reporter    (播放上报)
  │     │     ├── internal/remote_control (远程控制)
  │     │     ├── internal/automator   (自动播放)
  │     │     ├── internal/macdriver   (macOS 驱动)
  │     │     └── internal/ui/kitty    (Kitty 图形协议)
  │     ├── internal/runtime   (平台运行时)
  │     ├── internal/types     (常量定义)
  │     └── internal/configs   (配置)
  ├── utils/*      (各 UI 包均依赖 utils)
  └── configs/     (内嵌配置文件)
```

`utils/` 下的包相对独立，被 `internal/` 各包依赖但不反向依赖。

---

## 3. 关键包内文件组织

### internal/ui/ - TUI 核心 (73 文件)

最大且最重要的包。文件组织:

| 分类 | 关键文件 | 说明 |
|------|----------|------|
| **核心协调** | `netease.go` (463行) | Netease 根组件，包含所有子组件引用、初始化、生命周期 |
| **菜单系统** | `menu.go` (62行) | Menu 接口定义、baseMenu 基类 |
| | `menu_main.go` | 主菜单 |
| | `menu_*.go` (50+) | 具体菜单: 歌单/专辑/歌手/电台/云盘/搜索/帮助等 |
| **事件处理** | `event_handler.go` (711行) | 键盘/鼠标事件处理 |
| **播放控制** | `player.go` (684行) | UI 层播放器封装 (播放/暂停/切歌等) |
| | `player_controller.go` (164行) | 远程控制接口实现 |
| **操作逻辑** | `operate.go` (849行) | 业务操作函数: like/下载/分享/打开网页等 |
| | `executor.go` (63行) | Operation 操作单元 (认证检查+加载状态) |
| | `action_select.go` | 操作菜单 |
| **渲染器** | `lyric_renderer.go` | 歌词渲染 |
| | `song_info_renderer.go` | 歌曲信息渲染 |
| | `progress_renderer.go` | 进度条渲染 |
| | `cover_renderer.go` | 封面渲染 (Kitty 协议) |
| | `composite_renderer.go` (168行) | 多列组合渲染器 |
| | `progress_color.go` | 进度条颜色 |
| | `lyric_color.go` | 歌词颜色 |
| | `ticker_by_player.go` (31行) | 基于播放器的渲染时钟 |
| **搜索** | `search_page.go` | 搜索界面 |
| | `menu_search_result.go` | 搜索结果菜单 |
| | `menu_search_type.go` | 搜索类型选择 |
| | `menu_local_search.go` | 本地搜索 |
| **登录** | `login_page.go` | 登录页面 |
| | `login_qr_page.go` | 二维码登录 |
| | `login_callback.go` | 登录回调 |
| **Last.fm** | `lastfm.go` | Last.fm 集成 |
| | `lastfm_auth_page.go` | Last.fm 认证页 |
| | `lastfm_qr_auth_page.go` | Last.fm 二维码认证 |
| | `lastfm_profile.go` | Last.fm 用户档案 |
| | `lastfm_custom_api_account.go` | Last.fm API 账号配置 |
| **全局热键** | `global_hotkey_enabled.go` | 启用全局热键 (含构建标签) |
| | `global_hotkey_disabled.go` | 禁用全局热键 (空实现) |
| **Kitty 协议** | `kitty/protocol.go` | Kitty 图形协议实现 |
| | `kitty/image.go` | Kitty 图像编码 |
| | `kitty/detector.go` | Kitty 终端检测 |

### internal/player/ - 播放引擎 (13 文件)

| 文件 | 说明 |
|------|------|
| `player.go` (85行) | Player 接口定义 + 工厂函数 |
| `song.go` | URLMusic 曲目模型 |
| `beep_player.go` | Beep (跨平台) 引擎实现 |
| `beep_decoder.go` | Beep 解码器桥接 |
| `mpv_player.go` | MPV (IPC) 引擎实现 |
| `mpd_player.go` | MPD (服务控制) 引擎实现 |
| `dlna_player.go` (591行) | DLNA 投送引擎 (UPnP SOAP 协议) |
| `osx_player.go` + `osx_player_handler.go` + `osx_player_darwin.go` | macOS AVFoundation 引擎 |
| `win_media_player.go` + `win_media_player_windows.go` + `win_media_player_windows_test.go` | Windows WinRT 引擎 |

### internal/playlist/ - 播放列表管理 (18 文件)

采用策略模式，六种播放模式各有独立实现:

| 文件 | 说明 |
|------|------|
| `interfaces.go` (81行) | PlaylistManager + PlayMode 接口定义 |
| `manager.go` (373行) | PlaylistManager 实现 |
| `list_loop.go` + `list_loop_test.go` | 列表循环模式 |
| `ordered.go` + `ordered_test.go` | 顺序播放模式 |
| `single_loop.go` + `single_loop_test.go` | 单曲循环模式 |
| `list_random.go` + `list_random_test.go` | 列表随机模式 |
| `infinite_random.go` + `infinite_random_test.go` | 无限随机模式 |
| `intelligent.go` + `intelligent_test.go` | 心动模式 (服务器推荐) |
| `errors.go` | 错误定义 |
| `benchmark_test.go` + `integration_test.go` + `manager_test.go` | 基准/集成/管理测试 |

### internal/storage/ - 持久化存储 (13 文件)

| 文件 | 说明 |
|------|------|
| `local_db.go` | BoltDB 初始化和管理器 |
| `table.go` | Table 接口 (KV 模型操作) |
| `model.go` | 数据模型接口 |
| `ky_model.go` | KV 模型辅助类型 |
| `user.go` | 用户数据模型 |
| `play_mode.go` | 播放模式存储 |
| `volume.go` | 音量存储 |
| `player_snapshot.go` | 播放列表快照 |
| `ext_info.go` | 扩展信息 (含版本号) |
| `last_signin_date.go` | 签到日期 |
| `lastfm_user.go` / `lastfm_api_account.go` / `lastfm_scrobble.go` | Last.fm 数据 |

### internal/structs/ - 数据模型 (9 文件)

纯数据模型，无业务逻辑:

| 文件 | 说明 |
|------|------|
| `song.go` | 歌曲模型 |
| `album.go` | 专辑模型 |
| `artist.go` | 艺术家模型 |
| `playlist.go` | 歌单模型 |
| `lyric.go` | 歌词模型 |
| `user.go` | 用户模型 |
| `rank.go` | 排行榜模型 |
| `dj_radio.go` | 电台模型 |
| `dj_category.go` | 电台分类模型 |

### internal/configs/ - 配置管理 (13 文件)

基于 `koanf` 的 TOML 配置解析:

| 文件 | 说明 |
|------|------|
| `config.go` (39行) | Config 根结构体 |
| `loader.go` | TOML 文件加载 (`koanf`) |
| `main.go` | 主界面配置 |
| `startup.go` | 启动页配置 |
| `theme.go` | 主题配置 |
| `player.go` | 播放器引擎配置 |
| `storage.go` | 存储配置 |
| `unm.go` | UNM 配置 |
| `reporter.go` | 上报配置 |
| `keybindings.go` | 快捷键配置 |
| `autoplayer.go` | 自动播放配置 |
| `framerate.go` | 帧率配置 |
| `hooks.go` | 配置钩子 |

### internal/lyric/ - 歌词服务 (4 文件)

| 文件 | 说明 |
|------|------|
| `service.go` (367行) | 歌词服务核心 (获取/解析/状态管理) |
| `fetcher.go` | 歌词获取 (LRC/YRC) |
| `lrc.go` | LRC 格式解析 |
| `yrc.go` | YRC 逐字歌词格式解析 |

### internal/track/ - 音轨管理 (5 文件)

| 文件 | 说明 |
|------|------|
| `manager.go` (432行) | 统一入口 (下载/缓存/歌词) |
| `fetcher.go` | 音频流获取 |
| `cache.go` | 缓存管理 |
| `tagger.go` | ID3/FLAC 标签写入 |
| `type.go` | 类型定义 |

---

## 4. 配置管理结构

### 配置文件位置

- **内嵌默认**: `utils/filex/embed/config.toml` → 编译时嵌入
- **用户配置**: `~/.config/go-musicfox/config.toml` (Linux) / `~/Library/Application Support/go-musicfox/config.toml` (macOS) / `%AppData%\go-musicfox\config.toml` (Windows)
- **环境变量**: `MUSICFOX_ROOT` 覆盖配置目录

### 配置结构层级

```toml
[startup]       # 启动页 (动画/签到/更新检查)
[main]          # 主界面 (altScreen/mouse/frameRate/debug)
  [main.notification]  # 桌面通知
  [main.lyric]         # 歌词显示
    [main.lyric.cover] # 封面图显示 (Kitty)
  [main.pprof]         # 性能分析
  [main.account]       # 账号 Cookie
[theme]         # 主题 (颜色/双列/居中)
  [theme.progress]     # 进度条样式
[storage]       # 存储 (下载目录/缓存)
  [storage.cache]      # 缓存设置
[player]        # 播放器 (引擎/音质/错误重试)
  [player.beep]        # Beep 引擎
  [player.mpd]         # MPD 引擎
  [player.mpv]         # MPV 引擎
  [player.dlna]        # DLNA 引擎
[autoplay]      # 自动播放 (歌单/偏移/模式)
[unm]           # UNM (解锁灰色歌曲)
[reporter]      # 播放上报
  [reporter.netease]   # 网易云上报
  [reporter.lastfm]    # Last.fm 上报
[keybindings]   # 快捷键
  [keybindings.global] # 全局热键
  [keybindings.app]    # 应用内快捷键
[share]         # 自定义分享模板
```

### 配置加载流程

1. `cmd/musicfox.go:loadConfig()` 被调用
2. 检查用户配置目录下的 `config.toml`
3. 不存在 → 从 `embed/config.toml` 复制
4. 调用 `configs.NewConfigFromTomlFile(path)` 解析
5. 配置存储到 `configs.AppConfig` 全局变量
6. 启动时通过 `configs.AppConfig.FillToModelOpts()` 注入 foxful-cli

---

## 5. 测试文件组织

### 测试覆盖

| 包 | 测试文件数 | 说明 |
|------|----------|------|
| `internal/playlist/` | 9 | 最完整: 6 种模式各有单元测试 + 集成测试 + 基准测试 |
| `internal/player/` | 1 | `win_media_player_windows_test.go` |
| `internal/remote_control/` | 1 | `remote_command_handler_darwin_test.go` |
| `internal/ui/` | 0 | **无测试** (最大包, 73 文件) |
| `internal/storage/` | 0 | **无测试** |
| `internal/lyric/` | 0 | **无测试** |
| `internal/track/` | 0 | **无测试** |
| `internal/keybindings/` | 0 | **无测试** |
| `utils/` | 0 | 整体无测试 |

### 测试数据

`testdata/` 目录存放测试用 flac 音频文件等。

---

## 6. Key Files 映射表

| 文件路径 | 行数 | 核心职责 |
|----------|------|----------|
| `cmd/musicfox.go` | 103 | 应用入口，CLI 初始化，配置加载 |
| `internal/runtime/runtime.go` | 13 | 非 Darwin 平台运行时 (panic 恢复) |
| `internal/runtime/runtime_darwin.go` | ~30 | macOS 平台运行时 (含应用支持目录设置) |
| `internal/commands/netease.go` | 76 | TUI 启动命令，组件装配 |
| `internal/commands/config.go` | ~60 | 配置管理命令 |
| `internal/commands/migrate.go` | ~50 | 配置迁移命令 (INI → TOML) |
| `internal/ui/netease.go` | 463 | **核心协调器**，所有 UI 组件的容器 |
| `internal/ui/menu.go` | 62 | Menu 接口定义 |
| `internal/ui/menu_main.go` | ~200 | 主菜单 (首页) |
| `internal/ui/event_handler.go` | 711 | 键盘/鼠标事件处理 |
| `internal/ui/player.go` | 684 | UI 层播放器 (播放控制+状态管理) |
| `internal/ui/player_controller.go` | 164 | 远程控制接口实现 |
| `internal/ui/operate.go` | 849 | 业务操作函数集 (like/下载/分享等) |
| `internal/ui/executor.go` | 63 | 操作执行器 (认证检查+加载态) |
| `internal/ui/action_select.go` | ~100 | 操作菜单 |
| `internal/ui/lyric_renderer.go` | ~300 | 歌词渲染 |
| `internal/ui/song_info_renderer.go` | ~200 | 歌曲信息渲染 |
| `internal/ui/progress_renderer.go` | ~150 | 进度条渲染 |
| `internal/ui/cover_renderer.go` | ~200 | Kitty 封面图渲染 |
| `internal/ui/composite_renderer.go` | 168 | 多列布局组合渲染器 |
| `internal/ui/ticker_by_player.go` | 31 | 播放器时钟适配器 |
| `internal/ui/search_page.go` | ~150 | 搜索界面 |
| `internal/ui/login_page.go` | ~100 | 登录页面 |
| `internal/ui/login_qr_page.go` | ~80 | 二维码登录 |
| `internal/ui/global_hotkey_enabled.go` | 28 | 全局热键 (带构建标签) |
| `internal/ui/kitty/protocol.go` | ~100 | Kitty 图形协议核心 |
| `internal/ui/kitty/image.go` | ~80 | Kitty 图像编码 |
| `internal/ui/kitty/detector.go` | ~30 | Kitty 终端检测 |
| `internal/player/player.go` | 85 | Player 接口 + 工厂 |
| `internal/player/beep_player.go` | ~300 | Beep 跨平台引擎 |
| `internal/player/mpv_player.go` | ~200 | MPV IPC 引擎 |
| `internal/player/mpd_player.go` | ~200 | MPD 服务引擎 |
| `internal/player/dlna_player.go` | 591 | DLNA UPnP 投送引擎 |
| `internal/player/osx_player.go` | ~100 | macOS AVFoundation 引擎 |
| `internal/player/win_media_player.go` | ~100 | Windows WinRT 引擎 |
| `internal/playlist/interfaces.go` | 81 | 播放列表管理器 + 播放模式接口 |
| `internal/playlist/manager.go` | 373 | 播放列表管理器实现 |
| `internal/playlist/list_loop.go` | ~50 | 列表循环策略 |
| `internal/playlist/ordered.go` | ~50 | 顺序播放策略 |
| `internal/playlist/single_loop.go` | ~50 | 单曲循环策略 |
| `internal/playlist/list_random.go` | ~70 | 随机播放策略 |
| `internal/playlist/infinite_random.go` | ~80 | 无限随机策略 |
| `internal/playlist/intelligent.go` | ~80 | 心动模式策略 |
| `internal/storage/local_db.go` | ~40 | BoltDB 初始化 |
| `internal/storage/table.go` | ~60 | KV 模型操作接口 |
| `internal/storage/model.go` | ~20 | 数据模型接口 |
| `internal/lyric/service.go` | 367 | 歌词服务 (获取/解析/状态) |
| `internal/lyric/lrc.go` | ~150 | LRC 格式解析 |
| `internal/lyric/yrc.go` | ~120 | YRC 逐字歌词解析 |
| `internal/lyric/fetcher.go` | ~80 | 歌词 API 获取 |
| `internal/track/manager.go` | 432 | 音轨管理 (下载/缓存) |
| `internal/track/cache.go` | ~120 | 文件缓存管理 |
| `internal/track/fetcher.go` | ~100 | 音频流获取 |
| `internal/track/tagger.go` | ~150 | ID3/FLAC 标签 |
| `internal/reporter/reporter.go` | 101 | 播放上报聚合器 |
| `internal/reporter/netease.go` | ~40 | 网易云播放上报 |
| `internal/reporter/lastfm.go` | ~60 | Last.fm 播放上报 |
| `internal/remote_control/remote_control.go` | 21 | 远程控制接口 (空实现) |
| `internal/remote_control/play_controller.go` | ~30 | Controller 接口 |
| `internal/remote_control/playing_info.go` | ~30 | PlayingInfo 接口 |
| `internal/remote_control/remote_control_darwin.go` | ~150 | macOS NowPlaying |
| `internal/remote_control/remote_control_linux.go` | ~100 | Linux MPRIS |
| `internal/remote_control/remote_control_windows.go` | ~80 | Windows SMTC |
| `internal/automator/autoplayer.go` | ~100 | 自动播放逻辑 |
| `internal/composer/share.go` | ~80 | 分享模板引擎 |
| `internal/composer/filename.go` | ~80 | 文件名生成模板 |
| `internal/composer/common.go` | ~30 | 公共类型 |
| `internal/composer/manager.go` | ~30 | 管理器 |
| `internal/lastfm/api.go` | ~80 | Last.fm API 客户端 |
| `internal/lastfm/track.go` | ~50 | Last.fm 曲目跟踪 |
| `internal/configs/config.go` | 39 | 配置根结构体 |
| `internal/configs/loader.go` | ~50 | TOML 配置加载 |
| `internal/types/constants.go` | 72 | 应用常量定义 |
| `internal/types/player.go` | 49 | 播放模式/状态类型定义 |
| `internal/keybindings/keybindings.go` | ~200 | 操作类型 + 按键映射构建 |
| `internal/structs/song.go` | ~80 | Song 数据模型 |
| `internal/structs/album.go` | ~40 | Album 数据模型 |
| `configs/config.toml` | 299 | 默认 TOML 配置 (详细注释) |
| `utils/app/app.go` | ~80 | 应用路径辅助 |
| `utils/notify/notification.go` | ~100 | 桌面通知 |
| `utils/errorx/` | ~30 | Panic 恢复/Goroutine 包装 |
| `utils/version/version.go` | ~40 | 版本检查 |
| `utils/netease/netease.go` | ~60 | 网易云 URL 工具 |
| `utils/slogx/` | ~30 | 结构化日志初始化 |
| `utils/clipboard/` | ~20 | 剪贴板操作 |
| `utils/likelist/` | ~30 | 喜欢列表缓存 |
| `utils/menux/` | ~40 | 菜单视图辅助 |
| `Makefile` | ~80 | 构建/安装/测试/发布命令 |

---

## 现状评估

### 优势

1. **模块化清晰**: `internal/` 下 19 个包职责明确，依赖关系合理
2. **接口驱动开发**: Player/PlaylistManager/PlayMode 等核心组件均面向接口编程
3. **构建标签隔离**: 平台相关代码通过文件级构建标签干净隔离
4. **测试先行 (部分)**: `playlist` 包测试覆盖全面，包含基准测试和集成测试
5. **配置驱动**: TOML 配置灵活，支持嵌套结构和详细注释

### 潜在改进

1. **ui 包文件过多 (73)**: 可按功能拆分为子包 (如 `ui/menus/`, `ui/renderers/`)
2. **测试覆盖不均**: 大部分包缺少测试，特别是 `ui` (最大包) 和 `storage`
3. **utils 包职责模糊**: 15 个子包中有部分可归入 `internal/` (如 `netease/`, `notify/`)
4. **configs 包与 utils/filex/embed 的配置重复**: 默认配置值在两个地方定义
5. **缺少 wire/di 容器**: 依赖关系通过手动构造函数建立，深层嵌套的组件创建链较长
