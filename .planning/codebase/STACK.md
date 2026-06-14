# STACK.md - go-musicfox 技术栈文档

> 生成日期: 2026-06-15 | 基于代码库实际分析

## 概述

go-musicfox 是使用 Go 编写的网易云音乐 TUI（终端用户界面）客户端。项目采用 bubbletea 弹性框架构建 UI，支持六大音频播放引擎，跨三大桌面平台（macOS/Linux/Windows），并集成了平台原生的媒体控制功能。

---

## 1. 编程语言

| 语言 | 用途 | 文件分布 |
|------|------|---------|
| **Go** | 主语言，所有核心业务逻辑 | `internal/` (19 个子包), `cmd/`, `utils/` (15 个子包), `configs/` |
| **Bash** | 构建脚本、发行辅助 | `hack/*.sh` (6 个脚本), `deploy/`, `githooks/pre-commit` |
| **Nix** | NixOS 发行与开发环境 | `flake.nix`, `deploy/nix/` |
| **C** (间接) | CGO 音频编解码依赖 | 通过 beep/FLAC/MP3 等库间接触发 |

---

## 2. Go 版本与模块

- **Go 版本**: `go 1.24.0` (go.mod)
- **模块路径**: `github.com/go-musicfox/go-musicfox`
- **CGO 必需**: 是 (`CGO_ENABLED=1`)，因为音频解码需要 C 库（FLAC、ALSA 等）
- **vendor 目录**: 存在，通过 CI 工作流自动同步（`vendor-sync.yml`）

---

## 3. 核心框架与 UI 库

| 依赖 | 版本 | 用途 |
|------|------|------|
| `github.com/charmbracelet/bubbletea` | v0.25.0 *(fork 版)* | TUI 框架核心 — 通过 replace 指令使用项目 fork `go-musicfox/bubbletea v0.25.0-foxful` |
| `github.com/anhoder/foxful-cli` | v0.5.0 | 基于 bubbletea 的 CLI 应用框架，提供菜单/路由/组件系统 |
| `github.com/charmbracelet/bubbles` | v0.16.1 | bubbletea UI 组件库（文本框、分页器等） |
| `github.com/charmbracelet/lipgloss` | v0.8.0 | TUI 样式/颜色/布局引擎 |
| `github.com/muesli/termenv` | v0.15.2 | 终端颜色与样式抽象 |
| `github.com/mattn/go-runewidth` | v0.0.15 | Unicode 字符宽度计算 |

**评估**: bubbletea 和 foxful-cli 均使用项目 fork，存在与上游的维护差距风险。但 fork 仅做必要定制（如 `go-musicfox/bubbletea`），变更可控。

---

## 4. 音频处理依赖

| 依赖 | 版本 | 用途 |
|------|------|------|
| `github.com/gopxl/beep` | v1.4.0 *(fork 版)* | 跨平台音频播放核心 — 通过 replace 使用 `go-musicfox/beep v1.4.1` |
| `github.com/ebitengine/purego` | v0.10.0 | macOS 端使用纯 Go 调用 ObjC API (AVFoundation) |
| `github.com/ebitengine/oto/v3` | v3.1.0 | beep 的低级音频输出驱动 (间接依赖) |
| `github.com/go-flac/go-flac` | v1.0.0 | FLAC 音频解码 (间接) |
| `github.com/go-flac/flacpicture` | v0.3.0 | FLAC 封面图读取 |
| `github.com/mewkiz/flac` | v1.0.8 | FLAC 编解码器 (间接) |
| `github.com/hajimehoshi/go-mp3` | v0.3.4 *(fork 版)* | MP3 解码器 — 通过 replace 使用 `go-musicfox/go-mp3 v0.3.3` |
| `github.com/tosone/minimp3` | v1.0.2 | 备选 MP3 解码器 — minimp3 (C 移植, CPU 占用更低) |
| `github.com/jfreymuth/oggvorbis` | v1.0.5 | OGG Vorbis 解码 (间接) |
| `github.com/jfreymuth/vorbis` | v1.0.2 | Vorbis 解码 (间接) |
| `github.com/icza/bitio` | v1.1.0 | 音频比特流读取 (间接) |

**评估**: 音频解码链路复杂，依赖大量 C 库。多个关键库（beep, go-mp3, goflac）使用 fork 版本，需注意上游更新。

---

## 5. 网易云音乐 API 集成

| 依赖 | 版本 | 用途 |
|------|------|------|
| `github.com/go-musicfox/netease-music` | v1.6.0 | 网易云音乐 API 客户端 — 封装所有 API 服务 |
| `github.com/go-musicfox/requests` | v0.2.3 | HTTP 请求客户端 (间接，netease-music 内部使用) |
| `github.com/buger/jsonparser` | v1.1.2 | 高性能 JSON 解析 |
| `github.com/tidwall/gjson` | v1.17.1 | JSON 路径查询 (间接) |
| `github.com/cnsilvan/UnblockNeteaseMusic` | *(fork 版)* | UNM — 解锁灰色/无版权歌曲 — 使用 `go-musicfox/UnblockNeteaseMusic v0.1.6` |
| `github.com/forgoer/openssl` | v1.6.0 | 加密/解密网易云请求 (间接) |

**评估**: netease-music SDK 和 UNM 均由项目维护，与上游 API 变更有潜在的滞后风险。

---

## 6. 播放引擎集成

| 依赖 | 版本 | 用途 |
|------|------|------|
| `github.com/fhs/gompd/v2` | v2.3.0 | MPD (Music Player Daemon) 协议客户端 — Linux 播放引擎 |
| `github.com/godbus/dbus/v5` | v5.1.0 | D-Bus 通信 — Linux MPRIS 远程控制与 MPD |
| `github.com/go-ole/go-ole` | v1.3.0 | Windows COM/OLE 接口 — 用于 WinRT API 调用 |
| `github.com/saltosystems/winrt-go` | *(fork 版)* | Windows Runtime API — SystemMediaTransportControls 集成 — 使用 `go-musicfox/winrt-go v0.1.4` |

---

## 7. 其他功能依赖

| 依赖 | 版本 | 用途 |
|------|------|------|
| `github.com/shkh/lastfm-go` | *(fork 版)* | Last.fm API 集成 — 使用 `go-musicfox/lastfm-go v0.0.2` |
| `go.etcd.io/bbolt` | v1.3.7 | 嵌入式 KV 数据库 — 本地持久化存储 |
| `github.com/knadh/koanf/v2` | v2.3.0 | TOML/文件配置文件管理 |
| `github.com/knadh/koanf/parsers/toml` | v0.1.0 | TOML 解析器 |
| `github.com/knadh/koanf/providers/file` | v1.2.0 | 文件配置提供者 |
| `github.com/knadh/koanf/providers/rawbytes` | v1.0.0 | 内嵌文件配置提供者 |
| `github.com/go-viper/mapstructure/v2` | v2.4.0 | 配置反序列化 |
| `github.com/adrg/xdg` | v0.5.3 | XDG 基础目录规范支持 (Linux) |
| `github.com/gookit/gcli/v2` | *(fork 版)* | CLI 命令框架 — 使用 `anhoder/gcli/v2 v2.3.5` |
| `github.com/gookit/ini/v2` | v2.2.2 | 旧版 INI 格式配置文件迁移用 |
| `github.com/robotn/gohook` | *(fork 版)* | 全局热键 — 使用 `go-musicfox/gohook v0.41.1` |
| `github.com/gen2brain/beeep` | *(fork 版)* | 桌面通知 |
| `github.com/go-musicfox/notificator` | v0.1.2 | 备用桌面通知 (项目自有) |
| `github.com/skip2/go-qrcode` | v0.0.0 | QR 码生成 — 登录、Last.fm 授权 |
| `github.com/mdp/qrterminal/v3` | v3.2.1 | 终端 QR 码显示 |
| `github.com/skratchdot/open-golang` | v0.0.0 | 跨平台 URL 打开 |
| `github.com/atotto/clipboard` | v0.1.4 | 剪贴板操作 |
| `github.com/bogem/id3v2/v2` | v2.1.4 | ID3 标签读写 — 下载歌曲标签 |
| `github.com/frolovo22/tag` | *(fork 版)* | 通用音频标签 — 使用 `go-musicfox/tag v1.0.2` |
| `github.com/juju/persistent-cookiejar` | v1.0.0 | Cookie 持久化 — 登录状态保持 |
| `github.com/pkg/errors` | v0.9.1 | 错误包装与堆栈追踪 |
| `golang.org/x/sync` | v0.18.0 | 并发控制 |
| `golang.org/x/sys` | v0.38.0 | 系统调用 (间接) |

---

## 8. Fork/Replace 依赖汇总

项目通过 `go.mod` 的 `replace` 指令使用以下 fork 版本（共 10 项）：

| 原始库 | fork 版本 | 原因推测 |
|--------|-----------|---------|
| `bubbletea` v0.25.0 | `go-musicfox/bubbletea v0.25.0-foxful` | TUI 框架定制 |
| `beep` v1.4.0 | `go-musicfox/beep v1.4.1` | 音频引擎修复/增强 |
| `gcli/v2` v2.3.4 | `anhoder/gcli/v2 v2.3.5` | CLI 命令框架定制 |
| `go-mp3` v0.3.4 | `go-musicfox/go-mp3 v0.3.3` | MP3 解码器修复 |
| `gohook` v0.41.0 | `go-musicfox/gohook v0.41.1` | 全局热键修复 |
| `winrt-go` | `go-musicfox/winrt-go v0.1.4` | Windows RT API 扩展 |
| `lastfm-go` | `go-musicfox/lastfm-go v0.0.2` | Last.fm API 修复 |
| `goflac` | `go-musicfox/goflac v0.1.5` | FLAC 解码修复 |
| `tag` v0.0.2 | `go-musicfox/tag v1.0.2` | 音频标签增强 |
| `UnblockNeteaseMusic` | `go-musicfox/UnblockNeteaseMusic v0.1.6` | UNM 库修复 |

**评估**: fork 数量较多 (10 个)，长期维护成本较高。建议定期检查上游更新并评估是否需要回迁。

---

## 9. 构建工具链

### 构建脚本 (`hack/build.sh`)
- CGO 编译 (`CGO_ENABLED=1`)
- 通过 `-ldflags` 注入版本信息：`-X internal/types.AppVersion`, `-X internal/types.LastfmKey/LastfmSecret`, `-X internal/types.BuildTags`
- 支持构建标签：`enable_global_hotkey`, `macapp`

### Makefile
| 目标 | 功能 |
|------|------|
| `build` | 调用 `hack/build.sh build`，输出到 `bin/musicfox` |
| `build-macapp` | macOS .app Bundle 打包 (带 `enable_global_hotkey` 标签) |
| `install` | 调用 `hack/build.sh install` |
| `lint` | `golangci-lint run -v` |
| `lint-fix` | `golangci-lint run --fix -v` |
| `test` | `go test ./internal/... ./utils/...` 生成覆盖率报告 |
| `release` | Docker 中运行 goreleaser 交叉编译发布 |
| `release-dry-run` | 预演发布流程 |
| `scoop-config-gen` | 生成 Scoop 包配置 |
| `changelog-gen` | 生成 CHANGELOG |
| `sysroot-pack/unpack` | 交叉编译系统 root 打包 |
| `init` | 设置 githooks (`core.hooksPath githooks`) |

### goreleaser 配置 (`.goreleaser.yaml`)
- **版本**: goreleaser v2
- **目标平台 (9 个构建)**:
  - Linux: amd64, arm64, arm
  - Windows: amd64, arm64 (均带 `enable_global_hotkey`)
  - macOS: amd64, arm64 (均带 `enable_global_hotkey`)
  - macOS App: amd64, arm64 (带 `enable_global_hotkey` + `macapp`)
- **打包格式**: zip (所有平台)
- **Linux 包**: APK, DEB, RPM, Arch Linux (通过 nFPM)
- **包管理器发布**: Homebrew (Formula + Cask), Winget
- **交叉编译**: 使用自定义 Docker 镜像 `goreleaser-musicfox`，内置 MinGW (Windows) 和交叉 GCC

### CI/CD (`.github/workflows/`)
| 工作流文件 | 触发器 | 功能 |
|-----------|--------|------|
| `release.yaml` | push tag v* | goreleaser 构建/发布/push scoop/changelog |
| `vendor-sync.yml` | push go.mod/go.sum 变更 | 自动 `go mod tidy && go mod vendor` + PR |
| `close-issues.yml` | 标签事件 | 自动关闭过时 issue |
| `duplicate-issues.yml` | issue 事件 | 重复 issue 检测 |
| `opencode.yml` | push/pull_request | OpenCode AI 工作流 |
| `pr-management.yml` | PR 事件 | PR 管理自动化 |
| `pr-standards.yml` | PR 事件 | PR 规范检查 |
| `review.yml` | issue_comment | 代码审查流程 |
| `triage.yml` | issue 事件 | 问题分类 |

---

## 10. 包管理方式

- **Go Modules** (`go.mod` / `go.sum`)
- 使用 `vendor` 目录提交依赖（CI 自动同步）
- 依赖下载: `go mod download` 或 `go mod vendor`

---

## 11. 开发工具

| 工具 | 配置 | 功能 |
|------|------|------|
| **golangci-lint** | `.golangci.yml` | Linter 套件：govet, errcheck, ineffassign, staticcheck, unused |
| **gci** | `.golangci.yml` | 导入排序 (standard → default → 项目前缀) |
| **gofmt** | `.golangci.yml` | 代码格式化 |
| **goimports** | `.golangci.yml` | 导入管理+格式化 |
| **githooks** | `githooks/pre-commit` | Git pre-commit hook |
| **Nix Flake** | `flake.nix` | Nix 开发环境 (go, alsa-lib, flac, pkg-config, devenv) |
| **mission-control** | Nix devShell | 开发命令快捷方式 (`run`, `build` 等) |

---

## 12. 跨平台支持

| 平台 | 架构 | 播放引擎默认值 | 远程控制 | 打包方式 |
|------|------|---------------|---------|---------|
| **macOS** | amd64, arm64 | OSX (AVFoundation) | Now Playing Center + MediaRemote | Homebrew Cask, .app Bundle |
| **Linux** | amd64, arm64, arm | Beep (PulseAudio/ALSA) | MPRIS2 (D-Bus) | Homebrew, APK/DEB/RPM/Arch, Flatpak, Nix |
| **Windows** | amd64, arm64 | WinMedia (WinRT) | SystemMediaTransportControls | Scoop, Winget |

**平台特定代码** (通过 `//go:build` 标签隔离):
- `darwin`: `runtime/runtime_darwin.go`, `player/osx_player_handler.go`, `remote_control/remote_control_darwin.go` 等
- `linux`: `remote_control/remote_control_linux.go` 等
- `windows`: `player/win_media_player_windows.go`, `remote_control/remote_control_windows.go` 等
- 全局热键: `enable_global_hotkey` 构建标签（Windows/macOS 默认启用，Linux 需手动编译）
- macOS App: `macapp` 构建标签

---

## 13. 现状评估

### 优势
- 成熟的 Go 生态，类型安全，编译为单一二进制
- bubbletea/foxful-cli 提供一致的 TUI 开发体验
- 跨平台构建流水线完整 (goreleaser + CI/CD)
- TOML 配置替代旧的 INI 格式，更结构化

### 风险
- **10 个 fork 依赖**维护成本高，上游更新可能遗漏
- CGO 强依赖导致编译复杂（需要交叉编译 sysroot）
- 音频解码库版本老旧（go-mp3 fork 比上游版本低）
- Go 版本升级到 1.24 较激进（许多 Linux 发行版仓库未跟进）
- pprof 未在生产构建中剥离（虽然默认 disabled）
