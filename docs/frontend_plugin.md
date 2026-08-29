# 前端插件化架构（Frontend Plugin Architecture）

> 本文档描述 go-musicfox 的前端插件化架构方案：把现有 TUI 前端做成可拔插的前端插件，并支持 GUI/WebUI 等新前端形态接入。目标分支：`feat/plugin-framework-playback` 及其后续。
>
> 本方案是设计文档（Phase 5 预留边界），当前处于「待实施」状态，P0/P1 实施 ticket 见文末。

## 概述

从 Phase 3.2 起，go-musicfox 的菜单/页面统一走 provider 注册表；从 Phase 3.x 起，核心引擎抽取为 UI-free 的 `internal/core`，前端（TUI `internal/ui` / headless `internal/headless`）经 `core.Observer` 事件接缝 + `framework` 服务注册表接入。本方案在此基础上前进两步：

1. **前端形态可插拔**：TUI / headless / WebUI 都是可注册的前端（`internal/frontend` 注册表），`runPlayer` 不再硬编码二选一。
2. **插件贡献点 UI-agnostic 化（新通道优先）**：新增「轨 B = UI-agnostic 命令贡献」，GUI/WebUI 插件走轨 B；现有 TUI 插件生态（轨 A）零改动、冻结不再新增。

## 设计原则

- **行为零破坏**：现有 9 个编译期插件 + WASM 插件在 TUI 上的行为保持现状（行为保持契约）。
- **core 保持 UI-free**：`internal/core` 不引入任何前端概念；前端注册表独立成包。
- **单前端互斥**：同一时刻只有一个前端运行，不设计多前端共存（YAGNI）。
- **新通道优先**：不迁移现有 Menu 体系（约 63 文件、40+ 菜单深嵌 foxful），轨 A 冻结、轨 B 是唯一新入口。
- **接口即契约**：opts 结构体 / 命令描述 / 前端接口即插件契约，类型进系统。

## 现状架构要点（依据）

- 依赖单向图：`plugins → ui → core → configs/framework`、`headless → core`；core/headless 禁止导入 bubbletea/foxful-cli/ui（已实测无违反）。
- `core.Engine`（`internal/core/engine.go`）组装 8 个业务服务进 `framework.Context`，`Startup(ctx, observer)` 跑 12 步启动序列，`Close()` 统一清理。
- `core.Observer`（`internal/core/observer.go`）是前端事件接缝（8 事件 + 可选 `LoadingIndicator`/`SongLocator`，nil-safe）。
- 前端选择硬编码在 `commands.runPlayer`（`internal/commands/netease.go:35-129`）：`headlessEnabled()` if 分支二选一；TUI 侧强引用 `ui.NewNetease`/`ui.NewMainMenu`。
- 插件扩展点（全部绑 bubbletea）：`RegisterMenu[T]`/`RegisterPage[T]`/`RegisterMainMenuItem(With/After)`/`RegisterStartupHook`/`RegisterContextMenuContrib`/`RegisterOperate+Handler`。opts 即契约；`WithPlugin` 归属启停；ui 经小接口反转访问插件类型。
- WASM 插件（`internal/wasm`）：manifest + wasm reactor，alloc/dealloc/export 调用协议，response 四种 action（toast/view/open_url/exec）——已是 UI-agnostic 的雏形；加载点耦合在 TUI 前端 `NewNetease`。

## 核心架构决策

### C1 前端契约：独立 `internal/frontend` 包

```go
// internal/frontend/frontend.go（零业务依赖，仅可 import core）
type LaunchOptions struct {
    Once  string   // 仅 headless 有意义
    Debug bool
    Pprof bool
    // ... CLI 旗标（取代 commands.GlobalOptions 直读）
}
type Frontend interface {
    ID() string
    Name() string
    Run(ctx context.Context, opts LaunchOptions) error
}
// 包级注册表（镜像 ui/registry.go 的 menuRegistry 模式），不进 framework.Context
func Register(f Frontend)                  // 重名 panic
func ByID(id string) (Frontend, bool)
func Registered() []string
```

- **明确不用 `Start(engine)` 形状**：engine 构建留在前端内部（`Run` 内部构建），避免 engine 生命周期反转触碰 TUI 的 user 槽共享（`ui/netease.go:69-76`）、DesktopLyrics 选项、WASM 加载时机三处微妙顺序。
- headless 的 `Run(once string)`（`headless/run.go:26-69`）已是此形状，改造量≈0。
- 注册方式镜像 `internal/plugins/plugins.go`：`ui` / `headless` / `webui` 各自 `init()` 注册，`cmd/musicfox.go` 空导入聚合器触发。
- ✅ P1-1 已实施（Run 占位，装配待 P1-2）。

### C2 前端选择：选项驱动

- 优先级：`--frontend` CLI ＞ `[main] frontend` 配置 ＞ 缺省 `tui`（镜像 `headlessEnabled` 语义）。
- `--headless` 保留为 legacy 别名（`--frontend=headless`）；`--once` 选非 headless 前端时 fail-fast。
- `bootstrap()`（`commands/netease.go:97-123`）前端无关，保持在 runPlayer 先跑再分发。
- ✅ P1-2 已实施。

### C3 双轨贡献点

> ✅ P2 已实施：轨 B 命令契约（`internal/frontend/command.go`）+ `ui.RegisterCommand` 归属 + TUI `CommandMenu` 适配器 + WASM 迁轨 B（`wasm/sink.go`）+ WebUI `/api/commands` 端点（exec 禁用）+ parity 测试。

**轨 A（现状，TUI 专属）**：`ui.RegisterMenu` 等现有注册表。冻结不再新增（TUI 专属 widget 菜单例外）。

**轨 B（新增，UI-agnostic 命令贡献）**：

```go
// internal/frontend（或 core）—— 轨 B 命令
type CommandContext struct {
    UserID int64; UserName string; Playing bool
    Song   *SongInfo   // 快照，复用并扩充 wasm.RequestContext
}
type CommandResult struct {
    Action  string         // "toast"|"view"|"open_url"|"exec"|"data"
    Title, Message, Level  string
    URL     string
    Command string; Args []string
    Data    any            // 新增：结构化 payload，GUI/WebUI 原生渲染
}
type Command struct {
    Key    string
    Title  string
    After  string                              // 复用 MainMenuItem.After 锚点语义
    Show   func(ctx CommandContext) bool       // 上下文过滤器
    Run    func(ctx CommandContext) CommandResult
}
```

**一个抽象三个消费方**（P2 已落地）：TUI 适配器 = `CommandMenu`（`internal/ui/command_menu.go`，泛化原 `WasmPluginMenu`）；WebUI = `GET/POST /api/commands` 端点（`internal/webui/commands.go`，exec 禁用）；headless = 无命令消费方，不加载 WASM（文档化非目标）。TUI 侧 `view` 结果自 S1 起升级为独立可滚动文本页：`commandActionCmd` 产出 `commandViewMsg`（`internal/ui/command_menu.go`）→ `command_view` 页面（`internal/ui/command_view_page.go`），toast 同步提示。

**双轨收敛纪律**：① WASM 插件注册立即迁轨 B；② 新插件/新贡献默认走轨 B；③ parity 测试断言「每个轨 B 命令在 TUI 有对应菜单入口」。三条不能同时落实则砍掉轨 B，WebUI 直接消费 core 命令面。

### C4 Observer 拆分

- 必需三件套：`OnSongChanged` / `OnStateChanged` / `OnPosition`。
- 可选接口（镜像 `LoadingIndicator`/`SongLocator` 模式）：`RequestLogin` / `OnPlaylistExhausted` / `OnRerender` / `OnStartupPhase`，core 内「断言+分派」。
- 影响面：`ui.Player` 全量实现零改动；`HeadlessObserver` 删 4 个方法。core 内部约 8 处调用点必须同步改。✅ P1-3 已实施。

### C5 menuServices 最小接口化

`type MenuServices = *menuServices`（`menu_accessor.go:36`）改为 `type MenuServices interface`（接口面 = 现有全部导出方法，结构体隐式实现），插件签名类型名不变零改动（影响仅 lastfm 插件 4 文件 8 处）。能力分层（纯业务窄接口 vs foxful 方法）留到 WebUI 真正需要时再做。✅ P1-4 已实施。

### C6 Dispatcher 提升到 core

headless 的 `Dispatcher`（`headless/ipc.go:48-102`）已 UI-free、engine 绑定、mutex 串行——提升到 `core`，headless unix socket/TCP server 退化为纯传输层，WebUI HTTP handler 复用同一 Dispatcher。`musicfox ctrl`（out-of-process）与 WebUI（in-process）共享同一命令语义。`ErrQuit` 语义由传输层决定。并发纪律：WebUI handler 所有控制/查询必须经 Dispatcher，禁止直调 Player mutating 方法。✅ P0-1 已实施

### C7 登录成为 core 能力

QR 登录客户端（`internal/ui/qr_login_client.go`）提升到 core 侧包，补 `Engine` 级登录流程编排（`SetAppCookieJar` + `UserService.Login` + `engine.LoginCallback`）。**WebUI 可行性的前置依赖**。✅ P0-2 已实施

### C8 WASM RegistrySink + side-effect policy

抽象 `wasm.RegistrySink` 接口（`RegisterMenu(decl)` / `RegisterMainMenuItem(...)`），TUI 与 WebUI 各自实现；保留 `WithPlugin` 归属语义让 `[plugins] disabled` 生效。sink 带 side-effect policy 能力位：WebUI 暴露 WASM 命令时 `exec` 必须禁用，`open_url` 降级为页面内链接。

### C9 WebUI 形态

> ✅ P3 已实施：`internal/webui` 前端已落地（HTTP/WS server + 安全四层 + 事件推送 + QR 登录 + 辅助端点 + vanilla JS 页面），`--frontend=webui` 可用。

**本地 HTTP/WebSocket server（127.0.0.1 随机端口）+ 自动打开浏览器**，作为 headless 的一等公民演进：

- **传输**：`net/http`（Go 1.22+ ServeMux）+ `github.com/coder/websocket`（gorilla 停更、x/net 废弃；并发安全、context 贯穿）。
- **协议**：复用 headless `Request/Response v1`，JSONL → WS 帧；新增 `{"type":"event",...}` 事件帧。
- **事件推送**：① 快照 + delta——WS 连接建立先推全量（复用 Dispatcher `status` + `PlayingInfo()` 快照），再增量推事件；② `OnPosition` 双重限频——core 已有 MPRIS throttle，WS 侧再限 2–4Hz，前端 rAF 插值。
- **静态资源**：`embed.FS` 内嵌，vanilla JS 零 Node 构建链。
- **辅助端点**：`/api/status`、`/api/albumart`、`/api/lyrics`（照 myMPD 路由模式）。
- **安全四层**（缺一不可）：仅 `127.0.0.1:0` 绑定；每次启动 256-bit token（URL→HttpOnly/SameSite=Strict cookie 交换，`subtle.ConstantTimeCompare`）；Host 白名单（防 DNS rebinding）；Origin/CORS 白名单（拒绝 `Origin: null`，不反射 ACAO，命令强制非 GET）。
- **与 headless 关系**：并存、共享 Dispatcher——unix socket 保留给脚本/`musicfox ctrl`，WebUI 走 HTTP/WS，同一 daemon 两通道互不干扰。

### C10 WebUI connect 客户端模式

> ✅ S5 已实施：`--frontend=webui --mode=connect` 让 WebUI 作为**客户端**连接本地 headless daemon（`musicfox --headless` 常驻），**不建 engine、不加载 WASM scope**——「headless 常驻播放 + 浏览器富控制面板」的单实例形态（本机单实例定位：unix socket 0600 / Windows 127.0.0.1 均不可跨机器访问，远程访问不在本阶段范围）。

**形态**：`connectRun`（`internal/webui/connect.go`）经 `headless.DialSubscribe(eventWireNames())` 连接 daemon 订阅通道 → `newRemoteBackend` 包成 `Backend` 的远程实现（`internal/webui/remote_backend.go`）→ `NewServerWithOptions(remoteBackend, ServerOptions{Auth: true})` 起常规 WebUI HTTP/WS 面 → 复用 `runServer`（自 S5-3 从 `run.go` 抽出，standalone/connect 共用「Serve + 打开浏览器 + 等待」）。页面/JS 零改动，后端换源；standalone 流程 = 现 `runStandalone`（engine + wasm scope + Startup）。

**功能边界**（S5-4，与 `connect.go` 头注释一致，`go test ./internal/webui/...` 锁定）：

| 能力 | standalone | connect |
|------|-----------|---------|
| 播放控制（play/next/seek/volume/...） | 本地 Dispatcher | daemon 转发（`SubscribeClient.Call`）✅ |
| 状态/快照 | 本地 engine | daemon status + playlist 快照缓存 ✅ |
| 事件推送（song/state/position/startup/login） | 本地 EventEmitter | daemon 订阅映射（wire→帧名）✅ |
| 命令面 `/api/commands` | 本地命令注册表 | 空（不加载 WASM，无命令贡献） |
| 登录（QR） | 本地 engine | 503（无本地 engine，扫码完成 803 路径拒） |
| `/api/albumart` | 本地 engine | 404（daemon 快照无 PicUrl） |
| `/api/lyrics` | 本地 engine | 空结构 |
| WS `quit` | 关自身 | 不转发（防误关 daemon） |

**与 `musicfox ctrl` 的关系**：二者都是 headless daemon 的客户端，形态不同——`musicfox ctrl` 每次调用新建连接、一请求一响应（窄命令面，脚本友好，`CtrlClient` 零改动保持兼容）；`--mode=connect` 持长连接订阅会话（快照 + 事件流 + 同连接 `Call`），面向浏览器富控制面板，`SubscribeClient` 即 `CtrlClient` 的长连接升级。断线语义：daemon 重启后 connect 长连接关闭，`Ready()=false` 使辅助端点降级、事件静止；**重连不做**（MVP，文档注明；`ctrl` 每次新建连接天然容错）。

### C12 TUI-connect 遥控壳

> ✅ S6 已实施：`--frontend=tui --mode=connect` 让 TUI 作为**遥控壳**连接本地 headless daemon（`musicfox --headless` 常驻）——控制经 `SubscribeClient.Call` 转发、状态经订阅驱动，**不建 engine、不跑 Startup（B9）**，与 webui-connect 共享 SubscribeClient 数据面，对齐「本机单实例」边界哲学（D-TC-1 方案 B：轻量遥控壳，非整体换源）。

**形态**：`tuiFrontend.Run` 对 `--mode=connect` 分发到 `RunConnect`（`internal/ui/connect.go`）：`headless.DialSubscribe(remoteEventWireNames())` 遥控 daemon → `NewNeteaseRemote`（`internal/ui/netease.go`）装配遥控壳——`ui.Player` 内嵌的 `*core.Player` 永不构造（B9），约 30 个**遮蔽方法**（`internal/ui/player.go`）在 connect 模式转发 `RemotePlayer`（`internal/ui/remote_player.go`：快照 + 事件流增量缓存、`Call` 转发、渲染 ticker），菜单/操作/快捷键调用点零改动；renderer 降级（歌词/频谱跳过、封面剥 PicUrl）；`menuServices.Player()/User()` 走 connect 分支；**connect 前端 scope**（S6-R1，`NewConnectFrontendScope`）挂载 **8/9 业务插件**（checkupdate/search/dj/album/artist/recommend/playlist/song；lastfm 排除——Deps 依赖 engine 服务），本地浏览菜单树完整；search 页回退注册（`registerConnectProviders`，仅 search 插件被禁用时兜底）。InitHook 只完成壳装配（不跑 Startup），并标记 `RemotePlayer` 运行态以开启渲染 poke（消费 goroutine 在构造时即启动，渲染事件在 App 运行后才投递）。

**功能边界**（S6 TC-4 + R1，与 `connect.go` 头注释一致，`go test ./internal/ui/...` 锁定）：

| 能力 | TUI standalone | TUI-connect（S6 MVP） |
|------|---------------|----------------------|
| 播放控制（next/prev/pause/resume/toggle/stop/seek/volume/repeat/shuffle/like/dislike） | 本地 engine | **`Call` 转发 daemon** ✅ |
| 播放状态 / 当前歌曲 / 进度 | 本地 | **订阅事件 + 快照缓存** ✅ |
| 搜索 / 排行榜 / 精选歌单 / 专辑 / 歌手 / DJ 浏览（无需登录） | 本地网易云 API | **本地照常** ✅（8 个业务插件挂载，菜单完整） |
| 收藏 / 我的歌单 / 云盘 / 每日推荐 / 最近播放 / 私人FM（需登录） | 本地网易云 API + 登录 | **toast 降级**（本地 API 无登录 cookie + 登录门控 → `ToLoginPage` → connect toast，对齐 B8） |
| 业务插件菜单面 | 9 插件全挂载（frontend scope） | **8/9 挂载**（lastfm 除外——Deps 依赖 engine 服务，connect 无 engine） |
| 浏览菜单的播放动作（选中 → PlaySong） | 本地建列表 + 播放 | **Player 遮蔽 toast**（`ui.Player.PlaySong` connect 分支） |
| 播放队列显示 | 本地完整列表 | 快照**精简只读**列表（id/name/artist/album）△ |
| 选歌播放（菜单选中 → PlaySong） | 本地 | **降级**：toast「遥控模式：daemon 不支持该操作」（P2：daemon `play_song` 命令扩展） |
| 播放模式/音量显示 | 本地 | 快照字段 ✅ |
| 登录 | 本地 | **daemon 登录态**（status.user 昵称）；TUI 侧登录禁用、需登录菜单 toast 降级 |
| 歌词 | 本地 LyricService | **降级**：隐藏（P2：本地拉取 + position 推进） |
| 封面 | 本地 | **降级**：无（P2：daemon 快照加 PicUrl） |
| 频谱 | 本地 PCM | **不可用**（组件隐藏） |
| 智能模式 / 心动 / 桌面歌词 | 本地 | **禁用**（P2 扩展） |
| 命令面（轨 B / WASM） | 本地执行 | **禁用**（`CommandContext.UserID` 不可得，对齐 webui-connect 空命令面；P2 扩展） |
| 断线语义 | — | daemon 断开 → 订阅 Events 关闭 → `ready=false` + 状态降级；**不自动重连**（MVP，对齐 webui-connect） |

**与 webui-connect 的关系**：同是 daemon 客户端，**共享 `headless.SubscribeClient` 数据面**（快照缓存 + 事件流 + 同连接 `Call`，D-TC-2 裁决）与「本机单实例」定位；差异在消费端——webui 是浏览器富页面（HTTP/WS 层 + `Backend` 抽象），TUI 是终端（纯 socket + 订阅协议，直接消费 core wire 名，无 webui 的帧名重映射层；`ui.Player` 遮蔽转发使菜单代码零改动）。`musicfox ctrl` 仍是窄命令面一次性客户端，三者并存。

**降级清单**（D-TC-3/B8/B10，`go test ./internal/ui/...` 守护）：选歌播放 / 播放队列编辑（`ReinitializePlaylist` 等本地建列表操作）/ 智能模式 / 登录 / 命令面全部 toast 禁用；需登录浏览（收藏/我的歌单/云盘/每日推荐/最近播放/私人FM）toast 降级；歌词 / 封面 / 频谱渲染隐藏或空渲染；断线（`server.Close`）→ 事件通道关闭 → `ready=false` + 状态冻结渲染，不自动重连。**注意**：「命令面禁用」≠「业务插件不加载」——8/9 业务插件仍挂载以保本地浏览菜单树（lastfm 除外，S6-R1）。

## TUI 作为前端插件的打包形态

- 注册：`internal/ui` 内 `init()` 调 `frontend.Register(tuiFrontend{})`，构造器 `Run` 内部搬入当前 runPlayer TUI 分支的全部装配（`commands/netease.go:51-89`）。CLI 旗标经 `LaunchOptions` 传入，**ui 保持不 import commands**。
- foxful-cli 依赖随 TUI 前端包携带可接受：只有 ui import foxful，frontend 注册表不 import，依赖图单向。
- headless 注册为内置前端：`Run(once)` 形状已匹配，改造最小。
- TUI 专属服务（coverRenderer/menuRegistry/pageRegistry）、renderer 组合、SearchPage 单例全部留在 ui 包内部——TUI 前端插件的私有实现，不进 core。

## 分阶段路线

| 阶段 | 内容 | 验证标准 |
|------|------|----------|
| **P0**（纯移动） | Dispatcher 提升 core（C6）；QR login client 提升 core 侧包 + Engine 级登录流程（C7） | 零行为变化；`go build/test` 全绿；`musicfox ctrl` 冒烟 |
| **P1**（纯重构） | `internal/frontend` 包 + `Run(ctx, LaunchOptions)` 契约 + 注册表 + runPlayer 改造（C1/C2）+ Observer 拆分（C4）+ menuServices 最小接口化（C5） | `go build/vet/test ./...` 全绿；行为测试全过；TUI/headless/`musicfox ctrl` 手工冒烟 |
| **P3**（WebUI 核心路径，**先于 P2**） | HTTP/WS server + Dispatcher 复用 + 快照/事件扇出 + 登录 + 安全四层 | httptest 测控制命令与快照；并发测试；curl/浏览器冒烟 ✅ 已实施 |
| **P2**（可与 P3 并行） | 轨 B 命令贡献 + TUI 适配器 + WASM 迁轨 B（C3/C8）+ `[plugins] disabled` gating | parity 测试；WASM 插件 TUI 行为不变 |

**价值排序**：P1 最高（解锁一切）＞ P3（第一个真实受益者 + 验证 Frontend 契约）＞ P2。

## 风险与对策

1. **engine 生命周期反转**（C1 原始 `Start(engine)` 形状）：已通过改为 `Run(ctx, LaunchOptions)` 消除。
2. **双轨永久分叉**（轨 B 无消费方 → 腐烂）：对策三件套（WASM 立即迁轨 B + 新贡献默认轨 B + parity 测试），三条不能同时落实则砍掉轨 B。

## 遗留设计点（当前决定）

| 点 | 决定 |
|----|------|
| `[plugins] disabled` 对轨 B | 轨 B 命令注册纳入 `WithPlugin` 归属 + `IsPluginEnabled` 过滤 |
| 状态同步 | 单前端互斥 + WebUI 快照 + 既有 `LoadPlaylistState` |
| 主题 | 保持 TUI 专属；WebUI 用浏览器 CSS 或映射主题名 |
| 前端元数据/启停 | 前端是单选非插件，不复用 WithPlugin；`musicfox frontend list` 自省留到 P3 之后 |
| 前端注册表 | 包级 map 即可，不进 framework.Context（YAGNI） |

## P0/P1 实施 tickets

> 可执行实施 ticket 清单（每票一个 commit 粒度，含文件/签名级规格、依赖、验证标准）。已逐文件核对代码（`internal/headless/`、`internal/core/`、`internal/ui/`、`internal/commands/`、`cmd/musicfox.go`、`internal/configs/`、`embed/config.toml` 及全部相关测试）。

### 总览与并行关系

```
P0-1 (Dispatcher→core) ──┐
P0-2 (QR login→core)   ──┼── 全部互相独立，可全并行
P1-3 (Observer 拆分)    ──┤
P1-4 (MenuServices 接口) ──┘
P1-1 (frontend 包+注册表) ──┐ 依赖链
P1-2 (runPlayer 改造)   ────┘ 依赖 P1-1（唯一前置）
```

- **P0-1 / P0-2 / P1-1 / P1-3 / P1-4**：五票互不触碰同一文件，可全部并行。
- **P1-2 必须等 P1-1**（需要 `internal/frontend` 注册表可用）。
- 每票独立 commit，Conventional Commits（`refactor(core): ...` / `feat(frontend): ...`）。

---

### P0-1｜Dispatcher 与线协议提升到 core，headless server 退化为纯传输层

**目标**：把 UI-free 的控制命令语义（Dispatcher）与线协议（Request/Response）提升进 core，WebUI（未来）与 headless 共享同一命令面；headless 只保留 socket/TCP 传输与 quit 生命周期语义。

**涉及文件**：

| 文件 | 动作 |
|---|---|
| `internal/headless/ipc.go` | 整体删除（内容迁往 core + server.go/run.go） |
| 新增 `internal/core/dispatcher.go` | 承接 Dispatcher + cmd\* + helpers |
| 新增 `internal/core/control.go` | 承接 Request/Response/ProtocolVersion |
| `internal/headless/server.go` | 改造：quit 拦截移入传输层；`ErrQuit` 移入本文件 |
| `internal/headless/run.go` | 改造：`core.NewDispatcher` + quit 拦截 |
| `internal/headless/client.go` | `Request`/`Response`/`ProtocolVersion` 改 `core.` 前缀 |
| `internal/headless/ipc_test.go` | Dispatcher 单测迁 `internal/core/dispatcher_test.go`；协议测试迁 `control_test.go` |
| 新增 `internal/core/dispatcher_test.go` / `control_test.go` | 承接测试（含 TestMain 基建） |

**改动内容**：
- `internal/core/control.go`：移动 `ProtocolVersion = 1`、`Request`、`Response`（字段与 json tag 原样保留）。
- `internal/core/dispatcher.go`：移动 `Dispatcher` 结构（`engine *core.Engine` + `mu sync.Mutex`）、`NewDispatcher`、`Dispatch`（**删除 `case "quit"`**）、`cmdStatus/cmdPlay/cmdSeek/cmdVolume/cmdRepeat/cmdShuffle`、`stateName/ctrlFloat/ctrlInt`——函数体零改动；imports 整理为 core-legal（`netease-music/service` + `internal/types` + `utils/struct` 均为合法）。
- `server.go`：`ErrQuit` 移入本文件（传输层哨兵）；`handleConn` 在 Dispatch 之前 `if req.Cmd == "quit"` → 写 `Response{Ok:true}` → `s.Close()` → return；`NewServerWithAddr` 改 `core.NewDispatcher`。
- `run.go`：`runOnce` 改 `core.NewDispatcher`；Dispatch 前拦截 `quit`——严格复刻当前 `--once quit` 行为（stdout `{"ok":false,"error":"quit"}` 并返回错误）。
- 测试：`TestRequestResponseJSONRoundTrip` → `control_test.go`；4 个 Dispatcher 单测 → `dispatcher_test.go`；**`TestDispatcherQuitReturnsErrQuit` 删除**（quit 已非 core 职责，由 `TestServerClientIntegration` 覆盖）。core 测试需自建 TestMain（镜像 headless ipc_test.go:31-59：临时 MUSICFOX_ROOT、beep 配置、`storage.LocalDBManager`、`sync.Once` 共享 `core.NewEngine`）。

**依赖**：无前置，与 P0-2 / P1-1 / P1-3 / P1-4 并行。

**验证**：`go build ./... && go vet ./internal/core/... ./internal/headless/...`；`go test ./internal/core/... ./internal/headless/...`；冒烟 `musicfox --headless &` + `musicfox ctrl status` + `ctrl quit`；`--once status` / `--once "play 周杰伦"` / `--once quit` 输出不变。

**风险**：core import 面扩张（均为合法依赖）；quit 两处拦截点（server.handleConn、runOnce）必须同时改，漏一处则行为回归；dispatcher 单测依赖真实 beep engine（照抄现有 TestMain）。

**写权限**：fixer。

---

### P0-2｜QR 登录客户端提升到 core 侧包 + Engine 级登录流程编排

**目标**：把 TLS 指纹二维码登录 HTTP 客户端从 `internal/ui` 移到 UI-free 层，Engine 上提供扫码成功后的登录完成编排（WebUI 前置依赖）。

**涉及文件**：`internal/ui/qr_login_client.go` 整体删除；新增 `internal/core/qrlogin/qr_login.go`（包 `qrlogin`）；新增 `internal/core/engine_qrlogin.go`；`internal/ui/login_qr_page.go` 调用点改造。

**改动内容**：
- `internal/core/qrlogin/qr_login.go`（包注释注明「UI-free QR login client，供 TUI/WebUI/headless 复用」）：移动 `qrLoginUserAgent`/`qrKeyClient`/`newQRLoginClient`（保持 unexported）；导出包级函数（零状态、签名与现有一致）：
  ```go
  func GetKey(cookieJar http.CookieJar) (uniKey string, qrcodeUrl string, err error)
  func CheckStatus(uniKey string, cookieJar http.CookieJar) (code float64, respBytes []byte, err error)
  ```
  函数体（`util.ApiParamsEncode`/`GenerateChainID`/`ApplyRequestStrategy`、全部 Header、TLS 指纹、错误文案）逐字保留。
- `internal/core/engine_qrlogin.go`：
  ```go
  // CompleteQRLogin 完成扫码登录后半程：替换 app 级 cookie jar、同步全局
  // jar、持久化并刷新用户资料。jar 为 nil 时跳过 jar 操作（nil-safe）。
  func (e *Engine) CompleteQRLogin(jar *cookiejar.Jar) error {
      if jar != nil {
          SetAppCookieJar(jar)
          neteaseutil.SetGlobalCookieJar(jar)
          if err := jar.Save(); err != nil { slog.Warn("持久化 Cookie 失败", slogx.Error(err)) }
      }
      return e.LoginCallback()
  }
  ```
  （镜像 login_qr_page.go:326-341 与 login_page.go:772-788 成功路径。）
- `login_qr_page.go`：`qrGetKey` → `qrlogin.GetKey`；`qrCheckStatus` → `qrlogin.CheckStatus`；`loginSuccessHandle` 替换为 `n.engine.CompleteQRLogin(core.AppCookieJar())`。

**归属决策（明确推荐）**：客户端放 `internal/core/qrlogin` **子包**而非 core 主包——core 主包不引入 `req/v3`（当前零 HTTP 客户端依赖）；WebUI/headless 可独立复用；编排方法在 core 主包，单向依赖子包无环。

**依赖**：无前置，可全并行。

**验证**：`go build ./... && go vet ./internal/core/... ./internal/ui/...`；`go test` 全绿（建议新增 `qr_login_test.go`：`GetKey` 空 jar 不 panic、`CheckStatus("")` 返回 (0, nil, nil) 短路）；冒烟 TUI 登录页扫码 → 回主界面标题刷新为登录昵称；`ls ~/.config/go-musicfox/cookie` 非空。

**风险**：`neteaseutil.GetGlobalCookieJar()` 每次轮询获取最新 jar 的语义必须保留（子包内**不要**缓存 jar）；`core.AppCookieJar()` 可能为 nil → nil-safe 分支覆盖（旧代码会 nil panic，新代码属可接受的行为硬化）；462 反爬与 TLS 指纹行为完全不动。

**写权限**：fixer。

---

### P1-1｜新建 `internal/frontend` 包：Frontend 契约 + 包级注册表 + 聚合器

**目标**：建立 UI-agnostic 前端注册机制（镜像 `internal/plugins/plugins.go` 聚合器模式）。

**涉及文件**：新增 `internal/frontend/frontend.go`、`frontend_test.go`、`internal/frontend/registration/registration.go`、`internal/headless/register.go`、`internal/ui/frontend.go`；`cmd/musicfox.go` 加空导入。

**改动内容**：
- `internal/frontend/frontend.go`（零业务依赖，仅 import context）：
  ```go
  type LaunchOptions struct {
      Once  string // 仅 headless 有意义
      Debug bool
      Pprof bool
  }
  type Frontend interface {
      ID() string
      Name() string
      Run(ctx context.Context, opts LaunchOptions) error
  }
  func Register(f Frontend)            // nil 或重复 ID panic
  func ByID(id string) (Frontend, bool)
  func Registered() []string           // 按注册序返回 ID 列表
  ```
  实现：`map[string]Frontend` + `sync.RWMutex`。**不做 `Start(engine)` 形状**（定稿 C1 已否决）。
- `internal/frontend/registration/registration.go`：**必须放子包而非 frontend 自身**（`ui → frontend`，若 `frontend → ui` 成环）：
  ```go
  package registration
  import (
      _ "github.com/go-musicfox/go-musicfox/internal/headless"
      _ "github.com/go-musicfox/go-musicfox/internal/ui"
  )
  ```
- `internal/headless/register.go`：`headlessFrontend{}.Run` 转调现有 `Run(opts.Once)`；`init()` 注册 `"headless"`。
- `internal/ui/frontend.go`：`tuiFrontend{}.Run` **P1-1 阶段占位** `errors.New("tui frontend not wired yet")` + 注释指向 P1-2（避免新旧两份装配重复）；`init()` 注册 `"tui"`。
- `cmd/musicfox.go`：空导入 `_ "github.com/go-musicfox/go-musicfox/internal/frontend/registration"`。

**依赖**：无前置（P1-2 的唯一前置），与 P0-1 / P0-2 / P1-3 / P1-4 并行。

**验证**：`go build ./...`；`go test ./internal/frontend/...`（Register 后 ByID 命中、Registered 顺序、重复 ID panic、nil panic）；`go vet`；冒烟 `musicfox --headless &` + `ctrl status` 仍正常。

**风险**：import cycle 是硬约束（聚合器必须在子包）；ui 的 `init()`（registry_registrations.go）是幂等注册无副作用；替代方案（直接空导入 ui+headless）代码更少但违反聚合器约定，不推荐。

**写权限**：fixer。

---

### P1-2｜runPlayer 改造：`--frontend` 选项驱动 + TUI 装配迁入 ui.Run + `--once` fail-fast

**目标**：用 frontend 注册表替换 `headlessEnabled()` 布尔二选一；TUI 装配整体搬进 `tuiFrontend.Run`；`--headless` 降级为 legacy 别名；`--once` 仅 headless 有效。

**涉及文件**：`internal/commands/netease.go`（35-91、125-129）、`options.go`（GlobalOptions 加 `Frontend string`）、`netease_test.go`（重写为 `TestResolveFrontend`）、`cmd/musicfox.go`（加 `--frontend` 旗标）、`internal/configs/main.go`（`MainConfig` 加 `Frontend`）、`embed/config.toml`（`[main]` 加 `frontend = "tui"`）、`main_config_test.go`、`internal/ui/frontend.go`（Run 占位换真实装配）。

**改动内容**：
- `cmd/musicfox.go`：`gf.BoolOpt(&GlobalOptions.Headless, "headless", ...)` 注释改 legacy 别名；`gf.StrOpt(&GlobalOptions.Frontend, "frontend", "", "", "select frontend: tui|headless (default tui)")`。
- `internal/configs/main.go`：`Frontend string \`koanf:"frontend"\``；`Headless bool` 字段保留（legacy 配置仍生效）。`embed/config.toml` `[main]` 加 `frontend = "tui"`（`UpgradeConfig` 会把该叶子项合并进老用户文件——隐藏依赖，必须加，否则老用户永不获得该键）。
- `commands/netease.go`：删 `headlessEnabled()`；删 TUI 装配块（43-89，含 `model.Submit` 等全局赋值）整体搬入 `internal/ui/frontend.go` 的 `tuiFrontend.Run`（**`ui.SetupI18n` 与 `model.Submit` 必须保持在 `ui.NewNetease` 之前**，勿重排）；`runPlayer` 改为注册表分发 + `--once` fail-fast；新增纯函数：
  ```go
  // resolveFrontendID 优先级：--frontend CLI ＞ --headless（legacy 别名）
  // ＞ [main] frontend ＞ [main] headless（legacy 配置）＞ 缺省 tui。
  func resolveFrontendID() string { ... }
  ```
  决策：`[main] headless = true` 保留为 legacy 配置并纳入解析（P1 零行为变化）；`[main] frontend` 优先于 `[main] headless`。
- import 调整：commands 删 `internal/headless`（ctrl.go 仍保留）、`tea`/`foxful model`/`runewidth`；加 `context`、`errors`、`internal/frontend`。

**依赖**：**P1-1**（唯一前置）。

**验证**：`go build/vet ./...`；`go test ./internal/commands/... ./internal/configs/...`；冒烟：默认进 TUI / `--headless` + `ctrl status` + `ctrl quit` / `--frontend=headless --once status` / `--once status`（无 headless）报错 `--once 仅支持 headless 前端` 且退出码非 0 / 配置 `[main] frontend = "headless"` 无 TUI 常驻 / 配置 `[main] headless = true` 仍进 headless。

**风险（最高风险票）**：`--once` 语义从「TUI 下被静默忽略」变为「非 headless fail-fast」是定稿明确的行为变化，必须在文档/CHANGELOG 标注；装配搬移顺序敏感；`bootstrap()` 留在 commands 不搬；commands 对 ui 依赖清零（已确认仅 netease.go:21 一处），`commands/ctrl.go:13` 对 headless 的 import 是有意保留。

**写权限**：fixer（建议 reviewer 重点复核 diff 等价性）。

---

### P1-3｜core.Observer 拆分：必需三件套 + 可选接口「断言+分派」

**目标**：Observer 接口瘦身为播放必需三件套，其余 4 个回调改为可选能力接口（镜像 LoadingIndicator/SongLocator 模式）。

**涉及文件**：`internal/core/observer.go`（接口瘦身 + 新增可选接口）、`player.go`（6 处调用点）、`player_gapless.go`（1 处 OnRerender）、`startup.go`（emitStartupPhase）、`internal/headless/frontend.go`（删 4 方法）、`frontend_test.go`、新增 `internal/core/observer_test.go`。

**改动内容**：
- `Observer` 只保留 `OnSongChanged` / `OnStateChanged` / `OnPosition`；新增可选接口：
  ```go
  type LoginRequester interface { RequestLogin(afterLogin func()) }
  type PlaylistExhaustedObserver interface { OnPlaylistExhausted(dir PlayDirection) }
  type RerenderObserver interface { OnRerender() }
  type StartupPhaseObserver interface { OnStartupPhase(phase StartupPhase) }
  ```
  保留现有 `LoadingIndicator`/`SongLocator` 不动。
- 8 处调用点：三件套直调不变；OnPlaylistExhausted（player.go:386,406）、OnRerender（player.go:599 + player_gapless.go:113）、emitStartupPhase（startup.go:43-47）改「`if o, ok := p.observer.(X); ok { o.X() }`」断言+分派。**接口方法删除后任何残留直调编译失败，天然防漏**。
- `internal/headless/frontend.go`：删 4 个可选方法；`var _ core.Observer` 断言保留（只断言三件套）。测试同步（`TestHeadlessObserverZeroValueNoPanic` 删可选方法调用；`TestHeadlessObserverRequestLoginDoesNotInvokeCallback` 删除）。
- 新增 `observer_test.go`：用 `NewEmptyPlayer()` + `SetObserver` 验证三件套-only 实现不 panic（CtrlRerender/NextSong 底部/emitStartupPhase）、全接口实现被正确分派。
- `internal/ui/player.go`（212-291）：**零改动**（方法全保留，自动满足三件套 + 4 可选接口）。

**依赖**：无前置，全并行。与 P1-1 文件边界不冲突（P1-1 新增 register.go，P1-3 改 frontend.go）。

**验证**：`go build/vet ./...`；`go test ./internal/core/... ./internal/headless/... ./internal/ui/...`；冒烟：TUI 启动（OnStartupPhase 标题刷新）、切歌、列表底部翻页、CtrlR/CtrlL 重绘、headless 常驻 + ctrl status。

**风险**：core 对 RequestLogin **零调用点**（已确认，仅为预留），headless 删除后无实际行为变化；`ui/netease.go:197` 传 `*ui.Player` 实现全部可选接口，断言全命中零改动。

**写权限**：fixer。

---

### P1-4｜`ui.MenuServices` 从类型别名改为接口

**目标**：插件边界类型从 `*menuServices` 指针别名改为显式接口，解除插件对具体实现的依赖（WebUI 可提供替代实现）。

**涉及文件**：`internal/ui/menu_accessor.go:36` 改造；`internal/plugins/lastfm/` 等 **零改动**（类型名不变）；`internal/ui/menu.go:282` 零改动（自动兼容）。

**改动内容**：
- `type MenuServices = *menuServices` → `type MenuServices interface { ... }`，接口面 = menu_accessor.go 现有**全部导出方法**（65-290 行共 24 个：Player/User/TrackManager/LyricService/DesktopLyrics/CoverRenderer/ShareSvc/Lastfm/Ctx/Netease/ToLoginPage/ToSearchPage/App/Main/MustMain/Rerender/Search/SaveActiveTheme/NotifyThemeSwitch/PlaybarHoveredElement/SetPlaybarHoveredElement/EffectiveWindowHeight/SpectrumLines/GetCoverWidth/GetCoverEndColumn/GetLyricPosition，一个不落）。`*menuServices` 隐式实现；`NewMenuServices(ctx)` 签名不变。
- 加编译期断言 `var _ MenuServices = (*menuServices)(nil)`。
- 已核实：全仓 69 处使用点全部是参数/字段/方法调用，无 nil 比较、无 `.(*menuServices)` 断言；**若后续发现需在改动前单独评估**（interface typed-nil 语义差异）。
- 能力分层（纯业务窄接口 vs foxful 方法）**明确不做**（定稿 C5）。

**依赖**：无前置，全并行。

**验证**：`go build/vet ./...`；`go test ./internal/plugins/lastfm/... ./internal/ui/... ./internal/commands/...`；编译期断言通过。

**风险**：interface vs 指针的 nil 语义差异是唯一理论风险（当前无调用点做 nil 比较，接口定义处加注释警示）。

**写权限**：fixer（建议 reviewer 用 grep 复核 nil 比较与类型断言清单）。

---

### 补充前置项（实施前必读）

1. **`embed/config.toml` 与 `configs.UpgradeConfig` 联动**（P1-2 隐藏依赖）：默认 TOML `[main]` 必须加 `frontend = "tui"` 叶子项，否则老用户配置永不获得该键。
2. **`internal/core` 新增测试基建（TestMain）**（P0-1 内部件，易被低估）：dispatcher 单测需要真实 `NewEngine` 的 beep speaker，core 测试包尚无 TestMain，需镜像 headless ipc_test.go:31-59 新建。
3. **`frontend` 聚合器必须放子包防环**（P1-1 硬约束）：`internal/frontend/registration/` 空导入 ui/headless，放 frontend 自身会形成 `ui → frontend → ui` 编译环。
4. **`--headless` 与 `[main] headless` 的 legacy 保底**（P1-2 决策点）：全部保留并纳入 `resolveFrontendID` 解析，保证 P1 阶段零行为变化；若要废弃 `[main] headless` 需单独立 ticket（文档 + CHANGELOG），不要混进 P1-2。
5. **`Request/Response` 归属**（P0-1 决策点）：随 Dispatcher 一起进 core（`control.go`）——WebUI 复用同一 JSON 协议，留在 headless 则 WebUI 被迫 import 前端包（依赖方向错误）；`ErrQuit` 则必须留 headless（传输层 shutdown 语义）。
6. **AGENTS.md 文档腐化控制**：五票收尾均包含对应文档段更新（AGENTS.md 的 core/headless 描述段、docs/frontend_plugin.md 的 C6/C7/C1/C2/C4/C5 标记为已实施），是项目文档维护准则的硬性要求。
