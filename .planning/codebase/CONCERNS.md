# CONCERNS.md — 关注点、风险与技术债务

> 基于 go-musicfox 代码库分析（239 个源文件，32 个测试文件，29MB vendor 依赖）
> 生成时间: 2026-06-15

---

## 🔴 高优先级

### 1. pprof 服务监听所有网络接口 — 安全风险

- **文件**: `internal/commands/netease.go:34`
- **问题**: pprof 调试端点在 `:9876`（所有接口）监听，暴露内存、goroutine、CPU profile 等敏感信息。未设置 `localhost` 绑定。
- **风险**: 局域网内的攻击者或恶意软件可读取运行时数据。
- **修复**: 将 `ListenAndServe` 地址改为 `"127.0.0.1:9876"`

### 2. 播放器初始化大量使用 panic — 启动崩溃风险

- **文件**: `internal/player/player.go:43,47,64,71,74,78,81`
- **问题**: `NewPlayerFromConfig()` 对配置缺失、依赖缺失直接 `panic`，导致应用无法优雅降级启动。
- **影响**: 用户若配置了不支持的播放引擎或外部依赖缺失，应用会直接崩溃，无法显示任何错误提示。
- **恢复影响行**: `internal/player/beep_player.go:100,130,133,139,194` — 文件操作失败也 panic。
- **建议**: 改为返回 error，在 UI 层展示优雅的错误提示。

### 3. Fork 依赖的无追踪维护 — 供应链风险

- **文件**: `go.mod:99-113`
- **问题**: 项目 forked 并替换了 9 个关键依赖：
  - `bubbletea` → `go-musicfox/bubbletea v0.25.0-foxful`
  - `beep` → `go-musicfox/beep v1.4.1`
  - `go-mp3` → `go-musicfox/go-mp3 v0.3.3`
  - `gohook` → `go-musicfox/gohook v0.41.1`
  - `gcli` → `anhoder/gcli/v2 v2.3.5`
  - `goflac` → `go-musicfox/goflac v0.1.5`
  - `tag` → `go-musicfox/tag v1.0.2`
  - `winrt-go` → `go-musicfox/winrt-go v0.1.4`
  - `lastfm-go` → `go-musicfox/lastfm-go v0.0.2`
- **风险**: fork 版本的修复和安全补丁可能滞后上游，长期维护成本高。
- **建议**: 定期同步上游，或逐步向上游提交 PR 合并修改。

### 4. API 密钥/敏感信息明文存储

- **文件**: 
  - `internal/types/constants.go:11-12` — 编译时注入的 Last.fm Key/Secret
  - `internal/configs/reporter.go:11,40` — 配置文件中的 Last.fm Secret
  - `utils/filex/embed/config.toml:250-251` — 明文 key/secret 字段
  - `internal/storage/lastfm_api_account.go:11` — BoltDB 中存储 Secret
- **风险**: 用户 Last.fm API credentials 以明文形式存储在 config.toml 和 BoltDB 数据库中。
- **建议**: 考虑使用系统 keychain（macOS Keychain、Linux Secret Service、Windows Credential Manager），或至少进行简单的混淆/加密存储。

### 5. 网易云 Cookie 明文存储

- **文件**: `utils/filex/embed/config.toml:82-83` — `neteaseCookie = ""`
- **问题**: 用户网易云登录 cookie 明文存储在配置文件中。
- **风险**: 若配置文件泄露，攻击者可直接使用 cookie 进行会话劫持。
- **建议**: 使用加密存储，或利用操作系统安全存储。

---

## 🟡 中优先级

### 6. FIXME 未解决的已知问题

| 位置 | 问题描述 |
|------|---------|
| `internal/player/beep_player.go:127-128` | "先这样处理，暂时没想到更好的办法" — 缓存文件处理逻辑粗糙 |
| `internal/player/beep_player.go:287-291` | FLAC 格式 Seek 会卡住 20-40 秒，仅支持 MP3 跳转 |
| `internal/keybindings/keybindings.go:551,555` | 键绑定解析器："," 字符的快捷键冲突，Unicode 空格可能被错误 trim |
| `internal/ui/operate.go:415` | 类型断言逻辑应进一步通用化 |
| `cmd/musicfox.go:39` | "后续版本移除" — migrate 命令的临时兼容处理待清理 |

### 7. 超大文件 — 可维护性风险

| 文件 | 行数 | 风险 |
|------|------|------|
| `internal/ui/login_page.go` | 902 行 | 登录逻辑与 UI 耦合，难以测试 |
| `internal/ui/operate.go` | 849 行 | 大量函数平铺，无子包拆分 |
| `internal/ui/event_handler.go` | 711 行 | 40+ 键盘操作集中处理 |
| `internal/ui/player.go` | 684 行 | 播放逻辑复杂，状态管理分散 |
| `internal/ui/cover_renderer.go` | 678 行 | 渲染、缓存、动画逻辑混合 |
| `internal/commands/migrate.go` | 642 行 | 配置迁移逻辑过于庞大 |
| `internal/player/dlna_player.go` | 591 行 | DLNA 协议处理复杂 |
| `internal/player/mpv_player.go` | 584 行 | IPC 控制逻辑集中 |
| `internal/keybindings/keybindings.go` | 581 行 | 键解析和操作定义混合 |

### 8. Goroutine 管理复杂 — 潜在泄漏风险

- **文件**: `internal/ui/cover_renderer.go:284-303`
- **问题**: 手动管理 `context.Cancel()` 和 goroutine 生命周期，worker pool 模式下的 goroutine 泄漏难以追踪。
- **文件**: `internal/player/beep_player.go:149` — 边下载边播放的 goroutine，依赖于 context 取消。
- **文件**: `internal/track/cache.go:106` — `prune()` 在 goroutine 中异步执行，无超时控制。
- **文件**: `internal/reporter/reporter.go:63,83` — reporter goroutine 生命周期管理。
- **建议**: 考虑使用 `errgroup` 统一管理 goroutine，添加显式的 shutdown 流程。

### 9. 不完全的 MPRIS 接口实现

- **文件**: `internal/remote_control/remote_control_linux.go:264`
- **问题**: MPRIS `org.mpris.MediaPlayer2.TrackList` 接口注释了但未实现。
- **影响**: Linux 桌面环境下，某些音乐控制插件可能无法正常显示播放列表。

### 10. 测试覆盖率不足

- **数据**: 239 个源文件 vs 32 个测试文件（~13%）
- **缺失关键的测试**:
  - `internal/ui/*` — 核心 UI 逻辑（事件处理、操作、播放器页面）完全没有测试
  - `internal/player/*` — beep、mpv、mpd 播放器无单元测试
  - `internal/lyric/*` — 歌词解析无完整测试
  - `internal/remote_control/*` — Linux/Windows 远程控制无测试
  - `internal/track/*` — 缓存和获取逻辑无测试
- **已有测试**: 仅 playlist（较完善）和 macdriver 包有测试。

### 11. 并发安全 — 锁顺序和竞争

- **文件**: `internal/player/beep_player.go:35` — 使用 `sync.Mutex`，但多处 panic 可能导致锁未释放
- **文件**: `internal/lyric/service.go:80` — `sync.RWMutex` 保护，但 SetSong 中的切换逻辑复杂
- **文件**: `internal/playlist/manager.go:18` — 有 recover 保护的 goroutine，但错误只是打印日志
- **文件**: `internal/ui/kitty/image.go:89,96` — 多个 RWMutex，锁粒度可能过大

### 12. 平台特定代码维护负担

- **数据**: 77 个文件使用了 `//go:build` 标签
- **Darwin 独占**: 47 个文件（`macdriver/`, `remote_control/`, `avcore/` 等）
- **Windows 独占**: `internal/player/win_media_player_windows.go`, `remote_control/remote_control_windows.go`
- **Linux 独占**: `remote_control/remote_control_linux.go`, `mpris_player_linux.go`
- **全局热键**: 通过 `enable_global_hotkey` build tag 控制
- **风险**: Darwin 特定代码占比过高（~20%），macOS 变更可能导致大量代码需要同步修改。

### 13. Deprecated 方法未设定移除时间线

- **文件**: `internal/ui/player_controller.go:19-156` — 15 个方法标记为 Deprecated，但仍在使用
- **文件**: `internal/playlist/interfaces.go:46,75`
- **建议**: 添加 `// Deprecated: will be removed in v4.0` 之类的注释，设定移除计划。

---

## 🟢 低优先级

### 14. TODO 改进项

| 位置 | 内容 |
|------|------|
| `internal/lyric/lrc.go:21` | 重构歌词解析以简化 Service.SetSong |
| `internal/ui/operate.go:127` | 提取获取歌单 ID 为独立函数 |
| `internal/ui/action_select.go:13` | 自适应添加操作项 |
| `utils/app/app.go:129` | 考虑移除 runtime 目录改为系统临时目录 |
| `internal/remote_control/remote_control_linux.go:264` | MPRIS TrackList 接口未完全实现 |

### 15. 中文注释影响国际贡献者

- Go 项目中代码注释约定使用英文，但项目中大量使用中文 FIXME/TODO/函数注释。
- **影响**: 非中文开发者难以理解代码意图。
- **文件范围**: `internal/ui/`, `internal/player/beep_player.go` 等

### 16. os.Stdout 写入忽略错误

- **文件**: `internal/ui/cover_renderer.go:224-225,254-259,416-417` 等多处
- **问题**: `os.Stdout.WriteString()` 和 `os.Stdout.Sync()` 的返回错误被 `_` 忽略。
- **风险**: 管道断开或终端关闭时无法检测，可能导致数据丢失或无限循环。

### 17. HTTP 客户端无 TLS/证书验证配置

- **文件**:
  - `internal/player/beep_player.go:72` — `http.Client{}` 无任何配置
  - `internal/player/dlna_player.go:123` — 仅设置了 1s timeout
  - `internal/track/fetcher.go:31` — 仅设置了 60s timeout
  - `internal/track/tagger.go:29` — 基础的 HTTP client
- **风险**: 未设置 `TLSClientConfig` 等安全配置。虽然网易云音乐 API 走 HTTPS，但未启用 HTTP/2 或证书固定。

### 18. 依赖版本的可升级性

- `github.com/pkg/errors` → 已归档，Go 1.13+ 标准库 `fmt.Errorf("%w")` 已替代
- `github.com/buger/jsonparser v1.1.2` → 版本较旧
- `gopkg.in/errgo.v1 v1.0.1` / `gopkg.in/retry.v1 v1.0.3` → 较老的 gopkg.in 路径
- `github.com/icza/bitio v1.1.0` → 4 年未更新
- `github.com/skratchdot/open-golang` → 5+ 年未更新

### 19. 大测试文件

- `internal/playlist/integration_test.go:504` — 集成测试过大
- `internal/playlist/manager_test.go:381` — 管理测试复杂
- `internal/playlist/infinite_random_test.go:380` — 随机测试数据

### 20. composer 包 panic 使用

- **文件**: `internal/composer/filename.go:36`, `internal/composer/share.go:47`
- **问题**: 模板加载失败直接 panic，无错误恢复路径。
- **建议**: 返回 error 或使用默认模板降级。

---

## 📊 总结统计

| 类别 | 高 | 中 | 低 | 合计 |
|------|-----|-----|-----|------|
| 安全 | 5 | 0 | 1 | 6 |
| 可维护性 | 0 | 4 | 3 | 7 |
| 技术债务 | 0 | 3 | 2 | 5 |
| 性能/并发 | 0 | 2 | 0 | 2 |
| 测试覆盖 | 0 | 1 | 0 | 1 |
| 依赖风险 | 1 | 0 | 1 | 2 |
| 平台 | 1 | 1 | 0 | 2 |

**总计**: 239 个源文件 | 32 个测试文件 | 77 个平台特定文件 | 12 个 FIXME/TODO | 29MB vendor 依赖
