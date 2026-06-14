# INTEGRATIONS.md - go-musicfox 外部集成文档

> 生成日期: 2026-06-15 | 基于代码库实际分析

## 概述

go-musicfox 集成了多个外部 API、本地服务、平台特定框架和远程控制协议，实现网易云音乐播放、跨平台媒体控制和外部服务上报等功能。本文档详细记录每一项集成。

---

## 1. 网易云音乐 API 集成

### 集成方式
通过项目自有 Go SDK `github.com/go-musicfox/netease-music` (v1.6.0) 调用网易云音乐服务端 API。

**关键文件**:
- `internal/netease/song.go` — 歌曲/播放列表 API 封装
- `internal/netease/playlist.go` — 歌单 API 封装
- `internal/netease/error.go` — API 错误处理
- `cmd/musicfox.go:58-64` — UNM 配置注入到 SDK

### API 功能覆盖
| 功能 | API 服务 | 文件 |
|------|---------|------|
| 每日推荐歌曲 | `RecommendSongsService` | `song.go:12` |
| 歌单详情/歌曲列表 | `PlaylistDetailService`, `PlaylistTrackAllService` | `song.go:24-30`, `playlist.go` |
| 用户登录 (多种方式) | 手机/邮箱/QR/验证码 | `ui/login_page.go`, `ui/login_qr_page.go` |
| 搜索 | 搜索 API | `ui/menu_search_type.go`, `ui/menu_search_result.go` |
| 歌词获取 | 歌词 API | `ui/lyric_renderer.go` |
| 用户播放列表 | 用户歌单 API | `ui/menu_user_playlist.go` |
| 专辑/歌手/电台 | 对应 API 服务 | `ui/menu_album_*.go`, `ui/menu_artist_*.go`, `ui/menu_dj_*.go` |
| 每日签到 | 签到 API | `configs/config.toml` 中 `startup.signIn` |
| 听歌排行上报 | 播放上报 API | `internal/reporter/netease.go` |

### 配置项
```toml
# config.toml
[main.account]
neteaseCookie = ""            # 登录 Cookie 持久化

[player]
songLevel = "higher"          # 音质级别
showAllSongsOfPlaylist = false # 是否获取全部歌单歌曲
```

### UNM (Unblock NetEase Music)
解锁灰色/无版权歌曲功能，通过 `github.com/go-musicfox/UnblockNeteaseMusic` (fork v0.1.6) 实现：
- **多源匹配**: 酷我(kuwo)、酷狗(kugou)、咪咕(migu)、QQ 音乐(qq)
- **配置**: `configs/config.toml` → `[unm]` 段
- **选项**: 代理模式 (`proxyURL`)、会员解锁 (`enableLocalVip`)、音质解锁 (`unlockSoundEffects`)
- **注入点**: `cmd/musicfox.go:59-64` — 启动时设置 SDK 全局开关

### 现状评估
- SDK 为自有 fork，API 变更响应可控
- 登录协议可能随网易云更新而失效（行业常见问题）
- UNM 依赖于第三方音乐平台的匹配准确性
- API 限流/风控风险：签到、上报等功能需注意使用频率

---

## 2. Last.fm API 集成

### 集成方式
通过 `github.com/go-musicfox/lastfm-go` (fork v0.0.2) 调用 Last.fm API v2.0。

**关键文件**:
- `internal/lastfm/api.go` — API 客户端封装，OAuth 授权流程
- `internal/lastfm/track.go` — Scrobble/NowPlaying 上报
- `internal/reporter/lastfm.go` — 播放状态自动上报器
- `ui/lastfm.go`, `ui/lastfm_auth_page.go` — 授权 UI 流程
- `internal/storage/lastfm_api_account.go` — API 账号存储
- `internal/storage/lastfm_scrobble.go` — Scrobble 记录存储
- `internal/storage/lastfm_user.go` — 用户信息存储

### 功能
- **OAuth 授权**: 通过 Web URL + QR 码完成授权
- **Now Playing**: 实时上报正在播放的歌曲
- **Scrobble**: 播放进度达阈值后上报听歌记录
- **API Key 注入**: 编译时通过 `-ldflags` 注入 (`types.LastfmKey`, `types.LastfmSecret`)

### 配置项
```toml
# config.toml
[reporter.lastfm]
enable = false              # 启用开关
key = ""                    # API Key (编译时或运行时)
secret = ""                 # API Secret
scrobblePoint = 50          # Scrobble 触发百分比 (50-100)
onlyFirstArtist = false     # 仅上报第一艺术家
skipDjRadio = false         # 跳过电台节目
```

### 现状评估
- 到 Last.fm 的 scrobbling 功能成熟稳定
- API Key 编译时注入方式不够灵活，建议支持运行时配置
- Last.fm API 已有多年未更新，稳定性好但功能有限

---

## 3. 本地存储 — BoltDB

### 集成方式
通过 `go.etcd.io/bbolt` (v1.3.7) 嵌入式的键值数据库。

**关键文件**:
- `internal/storage/local_db.go` — 数据库生命周期管理 (打开/关闭/临时实例)
- `internal/storage/table.go` — 数据桶定义和初始化
- `internal/storage/model.go` — 基础 Model 接口和通用 CRUD

### 存储数据 (12 个桶)
| 桶名 | 存储内容 | 文件 |
|------|---------|------|
| `user` | 用户基本信息 (ID, 昵称, 头像) | `user.go` |
| `play_mode` | 播放模式 (列表循环/随机等) | `play_mode.go` |
| `volume` | 系统音量 | `volume.go` |
| `last_signin_date` | 最后签到日期 | `last_signin_date.go` |
| `player_snapshot` | 播放器快照 (当前歌曲/进度) | `player_snapshot.go` |
| `lastfm_user` | Last.fm 授权用户信息 | `lastfm_user.go` |
| `lastfm_api_account` | Last.fm API 账号 | `lastfm_api_account.go` |
| `lastfm_scrobble` | Last.fm Scrobble 记录 | `lastfm_scrobble.go` |
| `ext_info` | 扩展信息 | `ext_info.go` |
| `ky_model` | 扩展数据桶 (通用键值) | `ky_model.go` |

### 数据库路径
- macOS: `$HOME/Library/Application Support/go-musicfox/musicfox.db`
- Linux: `$XDG_DATA_HOME/go-musicfox/musicfox.db`
- Windows: `%AppData%/go-musicfox/musicfox.db`

### 现状评估
- BoltDB 选择合理，嵌入式 DB 适合桌面应用
- 数据模型扁平，无迁移机制（新增字段不会自动创建）
- 无备份/恢复机制
- BoltDB 是只读事务并发模型，写事务串行化 — 高并发场景下需注意锁竞争

---

## 4. 音频播放引擎

go-musicfox 支持六种播放引擎，通过 `config.toml` 的 `player.engine` 配置选择。

| 引擎 | 配置值 | 平台 | 实现文件 | 技术原理 |
|------|--------|------|---------|---------|
| **Beep** | `beep` | 跨平台 | `beep_player.go`, `beep_decoder.go` | beep 库 + PulseAudio/ALSA/Windows WASAPI 输出，内置 MP3/FLAC/OGG/WAV 解码 |
| **OSX/AVFoundation** | `osx` | macOS | `osx_player.go` (通用桩), `osx_player_darwin.go`, `osx_player_handler.go` | 通过 purego/objc 桥接调用 AVFoundation API |
| **WinMedia** | `win_media` | Windows | `win_media_player.go` (通用桩), `win_media_player_windows.go` | Windows.Media.Playback API |
| **MPV** | `mpv` | 跨平台 | `mpv_player.go` | 通过 IPC (Unix socket) 控制 mpv 进程 |
| **MPD** | `mpd` | Linux | `mpd_player.go` | MPD 协议 (Unix socket/TCP)，通过 `gompd` 库 |
| **DLNA** | `dlna` | 跨平台 | `dlna_player.go` | UPnP MediaRenderer — 内置 HTTP Server 提供音频流，SOAP 控制 |

### 引擎初始化
`internal/player/player.go:33-85` — `NewPlayerFromConfig()` 工厂函数，根据配置创建对应播放器。

### Beep 引擎解码器
- **MP3**: `go-mp3` (默认) 或 `minimp3` (备选，CPU 占用更低)
- **FLAC**: `goflac` (fork 版)
- **OGG Vorbis**: beep 内置 OGG/Vorbis 解码

### OSX (AVFoundation) 引擎
- **文件**: `internal/player/osx_player_handler.go` — `//go:build darwin`
- **技术**: 通过 `ebitengine/purego/objc` 注册 `AVPlayerHandler` ObjC 类
- **依赖**: `internal/macdriver/avcore/`, `internal/macdriver/core/`
- **事件**: 播放完成 (`handleFinish`)、播放失败 (`handleFailed`)

### WinMedia 引擎
- **文件**: `internal/player/win_media_player_windows.go` — `//go:build windows`
- **技术**: Windows.Media.Playback.MediaPlayer API
- **平台集成**: 与 SystemMediaTransportControls 共享 MediaPlayer 实例

### MPV 引擎
- **依赖**: 系统安装 `mpv` 可执行文件
- **配置**: `player.mpv.bin` — MPV 路径
- **通信**: IPC Unix socket (JSON RPC 协议)
- **文件**: `mpv_player.go` (584 行)

### MPD 引擎
- **依赖**: 系统安装 `mpd` 可执行文件
- **配置**: `player.mpd.bin/network/addr/configFile/autoStart`
- **通信**: MPD 协议 (TCP/Unix socket)
- **自动启动**: 支持自动启动 mpd 进程

### DLNA 引擎
- **依赖**: DLNA/UPnP 兼容设备
- **配置**: `player.dlna.deviceUrl/localIP`
- **原理**: 内建 HTTP 服务器提供音频流，SOAP 协议控制远程设备播放
- **发现**: 通过 `gupnp-universal-cp` 扫描设备

### 现状评估
- 六大引擎覆盖全面，满足不同场景
- OSX/WinMedia 引擎与平台深度绑定，不可移植
- Beep 引擎依赖多个 fork 音频库，升级路径不明确
- MPD 配置较复杂，对普通用户不友好
- DLNA 引擎仅支持基本控制，不支持元数据同步

---

## 5. 平台特定集成

### 5.1 macOS 集成

**Now Playing Center + MediaRemote**
- **文件**: `internal/remote_control/remote_control_darwin.go` — `//go:build darwin`
- **框架**: `internal/macdriver/mediaplayer/` — MPNowPlayingInfoCenter, MPRemoteCommandCenter
- **功能**:
  - 菜单栏控制中心显示歌曲信息
  - 系统 Now Playing Widget
  - Remote Command 响应 (播放/暂停/上一首/下一首/快进/快退/进度控制/歌词开关等)
  - 播放状态同步 (Playing/Paused/Stopped/Interrupted)

**macOS 应用生命周期**
- **文件**: `internal/runtime/runtime_darwin.go` — `//go:build darwin`
- **框架**: `internal/macdriver/cocoa/` — NSApplication/NSApp
- **功能**: macOS 应用激活 (ActivationPolicyProhibited 避免 Dock 图标), 生命周期管理

**AVFoundation 音频**
- **文件**: `internal/player/osx_player_handler.go`, `internal/macdriver/avcore/`
- **功能**: 通过纯 Go 的 ObjC 运行时调用 AVAsset, AVPlayer 等 API

**系统事件响应**
- 睡眠/唤醒事件 (通过 macOS 通知)
- 蓝牙耳机连接/断开
- 支持 LyricsX fork 版同步显示歌词

### 5.2 Linux 集成

**MPRIS2 (D-Bus)**
- **文件**: `internal/remote_control/remote_control_linux.go` — `//go:build linux`
- **依赖**: `github.com/godbus/dbus/v5` (v5.1.0)
- **协议**: MPRIS2 (Media Player Remote Interfacing Specification)
- **功能**:
  - 注册 `org.mpris.MediaPlayer2.musicfox.instance<PID>` 到 D-Bus Session Bus
  - 提供 `org.mpris.MediaPlayer2` 接口 (Identity, CanQuit, HasTrackList 等属性)
  - 提供 `org.mpris.MediaPlayer2.Player` 接口 (播放控制/进度/音量/元数据)
  - 播放列表 TrackList 接口
  - D-Bus introspection (267 行完整属性注册)
- **集成客户端**: KDE Connect, GNOME Shell Extension, playerctl 等

**音频输出**
- PulseAudio / ALSA (通过 beep 引擎)
- PipeWire (通过 PulseAudio 兼容层)

### 5.3 Windows 集成

**SystemMediaTransportControls (SMTC)**
- **文件**: `internal/remote_control/remote_control_windows.go` — `//go:build windows`
- **依赖**: `github.com/saltosystems/winrt-go` (fork), `go-ole`
- **API**: Windows.Media.SystemMediaTransportControls
- **功能**:
  - 播放状态同步到系统媒体弹出窗口
  - 播放控制按钮事件响应 (Play/Pause/Next/Previous)
  - 缩略图/艺术家/歌曲名显示
  - 时间线进度显示

### 现状评估
- 三大平台媒体控制均已实现，覆盖主流集成场景
- macOS 的 macdriver 使用纯 Go 的 ObjC 运行时 (purego/objc)，无 CGO 依赖 — 是项目架构亮点
- Linux MPRIS2 实现功能完整 (337 行)，符合规范
- Windows SMTC 使用 COM/WinRT API，初始化需要 `ole.RoInitialize`

---

## 6. 桌面通知

| 库 | 平台 | 文件 |
|----|------|------|
| `gen2brain/beeep` (fork) | Windows/Linux 基础通知 | `utils/notify/` |
| `go-musicfox/notificator` | 备用通知 | `utils/notify/` |
| macOS NSUserNotification | macOS 原生通知 | 通过 `internal/remote_control/remote_control_darwin.go` |

**配置**: `configs/config.toml` → `[main.notification]` 段

---

## 7. 全局热键

**依赖**: `robotn/gohook` (fork `go-musicfox/gohook v0.41.1`)
**文件**: `internal/ui/global_hotkey_enabled.go`, `internal/ui/global_hotkey_disabled.go`
**构建标签**: `enable_global_hotkey`
**平台支持**:
- macOS/Windows: 默认构建包含
- Linux: 需手动编译 (`BUILD_TAGS=enable_global_hotkey make build`) — 需要额外系统依赖

**配置**: `configs/config.toml` → `[keybindings.global]` 段

---

## 8. 内建 HTTP 服务

- **DLNA 内建 HTTP 服务器**: `dlna_player.go` 启动 HTTP server，为 DLNA 设备提供音频流
- **QR 码登录回调**: 本地 HTTP 服务器接收登录回调
- **pprof**: `configs/config.toml` → `[main.pprof]` — 开发调试用 (`--pprof` 命令行选项)

---

## 9. 其他外部依赖

| 功能 | 依赖 | 文件 |
|------|------|------|
| **QR 码生成** | `skip2/go-qrcode`, `mdp/qrterminal/v3` | 登录、Last.fm 授权 |
| **剪贴板** | `atotto/clipboard` | 分享/复制操作 |
| **URL 打开** | `skratchdot/open-golang` | 在浏览器中打开歌曲/专辑链接 |
| **Cookie 持久化** | `juju/persistent-cookiejar` | 网易云登录状态保持 |
| **音频标签** | `bogem/id3v2/v2`, `go-musicfox/tag` (fork) | 下载歌曲的 ID3/元数据标签 |
| **文件编码解码** | `forgoer/openssl` | 网易云请求加密 (间接) |
| **运行环境检测** | `adrg/xdg` | XDG 目录规范 |

---

## 10. 集成架构总览

```
go-musicfox
├── 网易云 API (netease-music SDK)
│   ├── UNM 多源匹配 (UnblockNeteaseMusic)
│   ├── 歌词获取
│   └── 播放上报
├── 播放引擎 (Player 接口)
│   ├── Beep (跨平台, fork beep)
│   ├── AVFoundation (macOS, purego/objc)
│   ├── WinRT Media (Windows, winrt-go)
│   ├── MPV IPC (Unix socket)
│   ├── MPD Protocol (Linux, gompd)
│   └── DLNA/UPnP (HTTP Server + SOAP)
├── 远程控制 (RemoteControl)
│   ├── macOS Now Playing (macdriver/mediaplayer)
│   ├── Linux MPRIS2 (godbus/dbus v5)
│   └── Windows SMTC (winrt-go, go-ole)
├── 外部服务
│   ├── Last.fm Scrobbling (fork lastfm-go)
│   └── 桌面通知 (beeep/notificator)
├── 本地存储
│   └── BoltDB (bbolt)
└── 辅助功能
    ├── QR Code (go-qrcode/qrterminal)
    ├── 全局热键 (gohook)
    ├── 剪贴板 (atotto/clipboard)
    └── Cookie 持久化 (persistent-cookiejar)
```

---

## 11. 现状评估

### 优势
- 六大播放引擎 + 三大平台远程控制，集成深度和广度优于同类 CLI 项目
- 平台 API 集成使用 `//go:build` 标签良好隔离，编译时自动选择
- macOS 的 macdriver (纯 Go ObjC 运行时) 是创新性的技术选型，减少 CGO 依赖
- BoltDB 轻量嵌入式存储，无外部服务依赖

### 风险
- 网易云 API 非官方接口，协议可能随时变更
- UNM 依赖第三方音源，匹配率无法保证
- 过多的 fork 依赖（10 个）增加安全漏洞跟踪和上游更新合并成本
- Last.fm API Key 硬编码编译时注入，Runtime 不可配置
- 无 API 请求失败重试/降级机制
- DLNA 引擎无自动设备发现，需手动配置
- MPD 配置对普通用户不友好
