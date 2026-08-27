# go-musicfox 项目架构文档

## 沟通语言准则

### 重要准则：与用户交流必须使用中文

**所有贡献者、开发者以及 AI 助手在与用户交流时，必须使用中文。**

#### 适用场景
- 代码审查意见和反馈
- Issue 回复和讨论
- Pull Request 描述和评论
- 文档编写（README、CHANGELOG 等除外，代码注释和提交信息仍使用英文）
- AI 助手与用户的对话

#### 例外情况
- **代码注释**：使用英文（保持代码可读性和国际化标准）
- **Git Commit Message**：使用英文（遵循 Conventional Commits 规范）
- **代码变量/函数命名**：使用英文（遵循 Go 语言惯例）
- **API 文档和错误消息**：根据实际情况使用中文或英文

#### 示例

```markdown
✓ 正确示例（中文交流）：
"这个 PR 修复了歌词封面图圆角功能的性能问题，优化后 CPU 占用从 10% 降至 2%。"

✗ 错误示例（英文交流）：
"This PR fixes the performance issue with lyric cover cornerRadius feature..."
```

## 编码行为准则

本准则旨在减少 AI 编码常见的错误。**权衡原则**：这些准则偏向谨慎而非速度，对于简单任务可自行判断。

> 本准则改编自 [andrej-karpathy-skills/CLAUDE.md](https://github.com/forrestchang/andrej-karpathy-skills/blob/main/CLAUDE.md)

### 1. 三思而后行

**不要假设。不要隐藏困惑。明确权衡。**

在实现之前：
- 明确陈述你的假设。如果不确定，先问。
- 如果存在多种解释方案，全部提出——不要默默选择一个。
- 如果存在更简单的方案，指出它。有理由时可以反驳。
- 如果有不清楚的地方，停下来。说出哪里不清楚，然后问。

### 2. 简单性优先

**最小代码解决问题。不要 speculative。**

- 不添加需求之外的功能。
- 一次性使用的代码不抽象。
- 不添加未被请求的"灵活性"或"可配置性"。
- 不处理不可能发生的错误场景。
- 如果写了 200 行而可以用 50 行完成，重写。

问自己："高级工程师会觉得这过于复杂吗？"如果是，简化。

### 3. 精准修改

**只触碰必须改的。只清理自己造成的混乱。**

编辑现有代码时：
- 不"改进"相邻代码、注释或格式。
- 不重构没坏的东西。
- 匹配现有风格，即使你可能用不同方式实现。
- 如果发现无关的死代码，指出它——不要删除它。

当你的修改产生孤儿（不再使用的代码）时：
- 移除因你的修改而不再使用的 import/变量/函数。
- 不要移除预先存在的死代码，除非被要求。

**检验标准**：每一行修改都应该能直接追溯到用户的请求。

### 4. 目标驱动执行

**定义成功标准。循环验证直到完成。**

将任务转化为可验证的目标：

| 模糊表述 | 可验证目标 |
|---------|-----------|
| "添加验证" | "为无效输入编写测试，然后让测试通过" |
| "修复 bug" | "编写能复现问题的测试，然后让测试通过" |
| "重构 X" | "确保重构前后测试都通过" |

对于多步骤任务，简要陈述计划：

```
1. [步骤] → 验证: [检查点]
2. [步骤] → 验证: [检查点]
3. [步骤] → 验证: [检查点]
```

明确的目标让你能独立循环验证。模糊的目标（"让它能工作"）需要不断确认。

---

**这些准则生效的标志**：diff 中不必要的修改更少，因过度复杂而返工更少，澄清问题在错误之前而非错误之后提出。

## 项目概述

go-musicfox 是基于 Go 和 bubbletea 的网易云音乐 TUI 客户端，支持 macOS/Linux/Windows。

**技术栈**：
- **UI 框架**：bubbletea + foxful-cli（部分定制）
- **音频处理**：beep、go-mp3、go-flac
- **存储**：BoltDB
- **配置**：TOML + mapstructure；`configs.SetTOMLValue(path, keyPath, value)` 使用 `github.com/pelletier/go-toml/v2/unstable/edit` 按键路径保真编辑既有 TOML。`configs.UpgradeConfig(path)` 将内嵌默认 TOML 中用户文件缺失的叶子项追加到对应 table；`upgrade-config` 子命令调用它。两者均保留无关注释、布局、已有值、未知键与原文件权限，并以原子替换写回。
- **API**：netease-music SDK

**项目结构**：`cmd/` 入口 | `internal/` 核心业务（含 `plugins/` 编译期插件聚合器与插件子包） | `utils/` 工具 | `configs/` 嵌入式配置

## 核心架构

### 应用入口与初始化

**入口**：`cmd/musicfox.go` → `runtime.Run()` → 加载配置 → 数据迁移 → 启动前端（TUI 或 headless，由 `--headless`/`[main] headless` 选择；headless 为无界面模式，可经 `musicfox ctrl` 控制）

### UI 协调器

**文件**：`internal/ui/netease.go`

`Netease` 是薄壳协调器（TUI 前端适配），只承担四类职责：

- **model.App 集成**：嵌入 `*model.App`，实现 foxful-cli 的 `InitHook`/`Update`/`CloseHook` 事件分派；
- **导航状态**：持有 search 页单例（`wordsInput`/`result`/`searchType` 被 SearchResultMenu 与 operate 共享）、页面跳转入口（`ToLoginPage`/`ToSearchPage`）；
- **renderer 组合**：歌词/歌曲信息/进度/封面/频谱 renderer 的创建与 `Components()` 组合排序；
- **TUI 专属服务注册**：业务实例由 `core.Engine` 持有（见下），shell 只把 TUI 专属服务（`coverRenderer`/菜单注册表/页面注册表）经 `registerUIExtraServices` 注册进 `engine.Ctx()` 的 framework 容器；启动序列经 `InitHook` 委托 `engine.Startup(ctx, observer)`（observer 为 `ui.Player` 包装，处理 TUI 重绘/定位/登录页等回调），changelog 弹窗留在 TUI 侧。

菜单与页面统一走 provider 注册表（`internal/ui/registry.go`）：key → 参数化工厂，各文件 `init()` 注册（`internal/ui/registry_registrations.go`），跳转经 `BuildMenu`/`BuildPage` 等 API。`baseMenu` 经 `menuServices` 类型安全访问器（`internal/ui/menu_accessor.go`）解析服务与薄壳方法（`MustMain`/`Rerender`/`SaveActiveTheme` 等）。

**前端插件化架构**：核心引擎（`internal/core`，UI-free）与前端解耦。`core.Engine` 组装全部业务服务（lastfm/trackManager/lyricService/desktopLyrics/shareSvc/core.Player/user 槽）与 framework ctx/scope，`Startup(ctx, observer)` 跑完整启动序列，`Close()` 统一清理。前端经 `core.Observer`（必需三件套 `OnSongChanged`/`OnStateChanged`/`OnPosition` + 可选接口 `LoginRequester`/`PlaylistExhaustedObserver`/`RerenderObserver`/`StartupPhaseObserver`，core 内断言+分派）消费引擎事件：TUI 前端（`internal/ui`）实现为重绘/定位/登录页，headless 前端（`internal/headless`）仅实现必需三件套。`ui.Player` 内嵌 `*core.Player` 方法提升转发，插件/commands 调用点不变。前端经 `internal/frontend` 注册表可拔插（`Frontend` 接口 `ID`/`Name`/`Run(ctx, LaunchOptions)` + 包级 `Register`/`ByID`/`Registered`；`cmd/musicfox.go` 空导入 `internal/frontend/registration` 聚合器触发 ui/headless/webui 的 `init()` 注册）。前端选择在 `commands.runPlayer`：`resolveFrontendID()`（CLI `--frontend` ＞ `--headless`（legacy 别名）＞ `[main] frontend` ＞ `[main] headless`（legacy）＞ 缺省 `tui`）从注册表解析并 `fe.Run(...)`，`--once` 仅 headless 前端有效（其它前端 fail-fast）；TUI 装配（model.App/EventHandler/MainMenu/Components/StatusBar/Ticker）已迁入 `ui/frontend.go` 的 `tuiFrontend.Run`。

### 核心文件路径

| 文件 | 说明 |
|------|------|
| `internal/ui/netease.go` | 薄壳协调器（App 集成/导航/事件分派/renderer 组合/TUI 专属服务注册，业务实例由 core.Engine 持有） |
| `internal/ui/services.go` | TUI 专属服务常量（`ServiceCoverRenderer`/`ServiceMenuRegistry`/`ServicePageRegistry`）+ 8 个 core 服务名的别名常量（core 定义在 `internal/core/services.go`）+ `registerUIExtraServices` |
| `internal/ui/registry.go` | provider 注册表：`RegisterMenu[T]`/`BuildMenu`/`mustBuildNoArg`/`buildMenuOrToast`、`RegisterPage[T]`/`BuildPage`、opts 契约类型；插件主菜单入口（`RegisterMainMenuItem`/`RegisterMainMenuItemWith`（带 Build 参数化入口）/`RegisterMainMenuItemAfter`（After 锚点链）/`MainMenuPluginItems`）与启动钩子（`RegisterStartupHook` 捕获插件 id 后委托 `internal/framework` 注册表，执行在 `core.Engine.Startup` 启动序第 10 步经 `framework.RunStartupHooks(configs.IsPluginEnabled)`，`runStartupHooks` 已随迁出 ui） |
| `internal/ui/context_menu.go` | 插件右键菜单扩展点：`RegisterContextMenuContrib`/`ContextMenuContribs` 注册表 + `ContextMenuContrib`/`ContextMenuContext` 契约 + `buildPluginContextMenuItems`/`handlePluginContextAction`（ID `plugin:<注册序号>` 分发） |
| `internal/ui/plugin_registry.go` | 插件归属声明与启停过滤：`WithPlugin(id, name, register)` 把作用域内的 `RegisterMenu`/`RegisterPage`/`RegisterMainMenuItem*`/`RegisterStartupHook`/`RegisterCommand` 记入该插件（`PluginInfo`/`PluginInfos` 快照，含 `CommandKeys`，同 id 多次声明幂等合并）；`IsPluginEnabled(id)` 委托 `configs.IsPluginEnabled`（读 `[plugins] disabled`，nil 配置按启用）；被禁用插件的**主菜单入口隐藏、启动钩子不执行、轨 B 命令执行被拒**，菜单 key 注册与 `BuildMenu` 跳转不受影响 |
| `internal/ui/registry_registrations.go` | 内置菜单/页面 provider 的 `init()` 注册 |
| `internal/ui/menu_accessor.go` | `menuServices` 类型安全访问器（服务解析 + 薄壳方法转发） |
| `internal/ui/menu.go` | `Menu` 接口与 `baseMenu`（经访问器接线） |
| `internal/core/` | UI-free 播放协调器（核心引擎层）：`player.go`（`Player` 结构 + `NewPlayer`/`NewEmptyPlayer` + 播放列表/播放模式/音量/歌词/遥控 API）、`player_gapless.go`（无缝播放预载/提交/取消）、`player_controller.go`（`remote_control.Controller` 实现）、`ctrl.go`（`PlayDirection`/`CtrlType`/`CtrlSignal`）、`observer.go`（前端观察者接缝：`Observer`/`LoadingIndicator`/`SongLocator`）、`events.go`（P4 事件面双写：事件名常量 `EvSongChanged`/`EvStateChanged`/`EvPosition`/`EvPlaylistEnd`/`EvRerender`/`EvLogin`/`EvStartupPhase` + `Player.emit`/`Engine.emit` 辅助 + `songEventPayload`/`loginEventPayload` 帧 data 构造器）、`mpris_throttle.go`（MPRIS Position 限流）、`engine.go`（`Engine`/`EngineOptions`：服务组装 + 用户槽 + `Startup(ctx, observer)`/`Close()`/`LoginCallback` 末尾发 `EvLogin`）、`services.go`（10 个业务服务名常量 + `UserService`/`LoginService`/`appCookieJar`/`ProvideIfAbsent`；`ServiceEventBus`/`ServiceDispatcher` 新增）、`services_plugins.go`（root Scope 服务构造器插件：10 个 `framework.Plugin` 按依赖序注册，Deps 经 `ServiceOf` 解析、Start 构造 + Provide、Dispose 真实清理；`NewEngine` 退化为纯装配器，`Engine.Close` 委托 scope.Stop/Dispose；`eventBusPlugin` 先于 `playerPlugin` 注册，player 经 `PlayerOptions.EventEmitter` 注入总线）、`startup.go`（完整启动序列：jar→用户恢复→cookie 登录→播放模式/音量/播放列表恢复→like list→签到→插件启动钩子→自动播放，经 `Engine.emitStartupPhase`（observer + 总线双写）通知前端）、`dispatcher.go` + `control.go`（控制命令面：`Dispatcher`（`Dispatch(ctx, cmd, args)` 经 Ctrl* 通道下发/直读查询，命令集不含 quit）与线协议 `Request`/`Response`/`ProtocolVersion`，自 headless 提升，WebUI 与 `musicfox ctrl` 共享同一命令语义）、`engine_qrlogin.go`（`CompleteQRLogin(jar)` 扫码登录后半程编排：替换 app 级 jar + 同步全局 jar + 持久化 + `LoginCallback`，jar 为 nil 时跳过；`EvLogin` 经 `LoginCallback` 单点发射）、`command_context.go`（`(*Player).CommandContext()` 产出 UI-agnostic 命令上下文快照，供 TUI CommandMenu 与 WebUI 命令端点共享，消灭各前端重复实现）、`qrlogin/` 子包（UI-free 二维码登录客户端 `GetKey`/`CheckStatus`，TLS 指纹反爬，供 TUI/WebUI/headless 复用）。播放事件双写纪律：每个 observer 调用点并排 `emit`（observer 先行、emitter 同步随后），listener 必须「投递后即返回」（见 docs/plugin_ecosystem.md §四）。源级禁止导入 `internal/ui`/bubbletea/foxful-cli；TUI 的 `ui.Player` 经内嵌 `*core.Player` 方法提升转发，插件/commands 调用点不变 |
| `internal/headless/` | 无 TUI 前端（headless 模式）：`frontend.go`（`HeadlessObserver` 仅实现 `core.Observer` 必需三件套，不实现可选接口）、`register.go`（`headlessFrontend` 经 `frontend.Register` 注册为内置前端）、`run.go`（`Run(once string)`：`core.NewEngine` → `Startup` → 常驻模式启动控制通道 server 并阻塞等待 SIGINT/SIGTERM 或 `quit` 命令；`--once "<cmd> [args]"` 单次模式执行一条控制命令并以紧凑 JSON 输出后退出，不启动 server；quit 由传输层拦截，`ErrQuit` 哨兵移至 server.go）、`server.go`（控制通道 server：非 Windows 用 unix socket（`DataDir/musicfox.sock`，监听前探测 stale socket 并移除，chmod 0600），Windows 用 TCP `127.0.0.1:0` 并把端口持久化到 `DataDir/musicfox.port`；`Serve(ctx)`/`Close()`/`ShutdownCh()`，每连接一请求一响应后关闭，连接内 panic 不杀 accept loop）、`client.go`（`CtrlClient`：`Dial()`/`Call(ctx, cmd, args)`（每次调用新建连接），包级原子计数器生成 ID）。源级禁止导入 `internal/ui`/bubbletea/foxful-cli（仅经 `internal/configs` 传递依赖可接受）、`internal/commands`（`--once` 值由 `commands` 作为参数传入 `headless.Run`）；`commands.runPlayer` 经 `resolveFrontendID()` 从 `internal/frontend` 注册表解析前端，`commands.NewCtrlCommand()` 注册 `musicfox ctrl <cmd> [args...]` 子命令 |
| `internal/frontend/` | UI-agnostic 前端注册表与轨 B 命令契约：`frontend.go`（`Frontend` 接口 `ID`/`Name`/`Run(ctx, LaunchOptions)`、`LaunchOptions`、包级 `Register`/`ByID`/`Registered`）、`command.go`（轨 B UI-agnostic 命令贡献：`Command{Key,Title,After,PluginID,Show,Run}`、`CommandContext`/`SongInfo`/`CommandResult` 纯数据契约（Action: toast/view/open_url/exec/data）、`RegisterCommand`/`Commands`/`CommandByKey` 注册表；**零业务依赖不变量**——不 import 任何内部业务包，契约形状与 `wasm.Request/Response` JSON 对齐）、`registration/` 聚合器（空导入 ui/headless/webui 触发注册，子包防 `ui → frontend → ui` 环）；前端经 `commands.runPlayer` 的 `resolveFrontendID()` 选择，轨 B 命令经 `ui/command_menu.go`（TUI）与 `webui/commands.go`（WebUI）两前端消费 |
| `internal/webui/` | WebUI 前端（源级禁止导入 `internal/ui`/bubbletea/foxful-cli）：`register.go`（`webuiFrontend` 经 `frontend.Register` 注册，`--frontend=webui` 启动）、`run.go`（`Run(ctx)`：`core.NewEngine` → `Startup(ctx, webuiNoopObserver{})` → 本地 HTTP/WS server → `open.Start` 打开浏览器 → 阻塞到 quit → `engine.Close`）、`server.go`（`Server`：`127.0.0.1:0` 监听 + `crypto/rand` token + ServeMux 路由 + `core.Dispatcher` 控制面 + WS 连接集合 + 幂等 `Close`；NewServer 从 engine ctx 解析 `ServiceEventBus` 并注册事件面订阅，Close 退订）、`security.go`（安全四层：token cookie 交换（`/token?token=` → HttpOnly/SameSite=Strict cookie）→ Host 白名单（防 DNS rebinding）→ Origin 校验（拒绝 `null`/跨源，不反射 CORS）→ `authMiddleware`/`verifyWSRequest`）、`ws.go`（`GET /ws`：`websocket.Accept`（coder/websocket，OriginPatterns 纵深防御）+ `wsjson` 复用 `core.Request/Response` 线协议 + `Dispatcher.Dispatch` 串行控制 + quit 传输层拦截 + 写锁/快照帧 + recover 隔离）、`events.go` + `broadcast.go`（事件面订阅：`subscribeEmitter` 注册 core `EventEmitter` listener（`webuiNoopObserver` 满足 Startup 的 `core.Observer` 签名，事件经总线而非 observer 消费），事件 → 帧名映射 `player.song_changed→song_changed`/`state_changed`/`position`（250ms 节流）/`startup.phase→startup_phase`/`auth.login_succeeded→login`，listener 只做 `broadcast` 投递；连接先注册后发快照（status + playlist 精简字段），增量事件 `{"type":"event",...}` 帧经 broadcaster 逐连接写锁广播）、`api_login.go`（QR 登录端点 `/api/login/qr/key|image|status`，复用 `core/qrlogin`（包级可覆写变量）+ `Engine.CompleteQRLogin`，803 成功后 `login` 帧由 core `EvLogin` 经事件面广播，handler 不再手工广播）、`api.go`（`/api/status`（Dispatcher status 快照）、`/api/albumart`（PicUrl 后端代理 + `.music.163.com` 域名白名单防 SSRF + Referer 防盗链）、`/api/lyrics`（`LyricService().State()` 结构化））、`commands.go`（轨 B 命令端点 `GET /api/commands`（按 `IsPluginEnabled` + `Show` 过滤列表）+ `POST /api/commands/{key}`（执行；禁用插件 404、`exec` action 403 拒绝、open_url 经 `open.Start`）、`wasm_sink.go`（`webuiWasmSink` 经 `wasm.LoadIntoScope` 加载 WASM 插件命令进 frontend 注册表（Replace 语义），WebUI 页面显示「插件命令」按钮组）、`static/`（`//go:embed` 内嵌 vanilla JS 单页：index.html + app.js + player.js + style.css，零 Node 构建链，WS 帧分发/控制命令/QR 登录/歌词高亮/封面） |
| `internal/plugins/plugins.go` | 编译期插件聚合器（空导入各插件子包；`cmd/musicfox.go` 空导入触发注册） |
| `internal/plugins/checkupdate/` | 首个真实插件：`CheckUpdateMenu`（嵌入 `ui.BaseMenu`，`init()` 注册 `check_update`，声明主菜单入口 + 启动钩子） |
| `internal/plugins/lastfm/` | 第二个真实插件：Last.fm 菜单/页面整体提取（`last_fm` 菜单 + `lastfm_auth`/`lastfm_custom_api` 页面，opts 携带 `ui.MenuServices`；`RegisterMainMenuItem("last_fm", "LastFM")` 主菜单入口） |
| `internal/plugins/dj/` | 第三个真实插件：DJ/电台集群整体提取（10 个菜单，key 与内置注册逐一相同；`radio_dj_type` 声明「主播电台」主菜单入口，主菜单项经 After 锚点 `could` 保持原内置位置；集群内跳转经 `ui.BuildMenuOrToast`/`ui.MustBuild*`；ui 侧排序切换经 `ui.DjRadioDetailSortable` 接口访问 `DjRadioDetailMenu`） |
| `internal/plugins/album/` | 第四个真实插件：专辑集群整体提取（8 个菜单，key 与内置注册逐一相同；`album_menu` 声明「专辑列表」主菜单入口，主菜单项经 After 锚点 `personal_fm` 保持原内置位置；集群内跳转经 `ui.BuildMenuOrToast`/`ui.MustBuildNoArg`；被 ui 共享的 `AlbumDetailOpts` 留在 ui，ui 侧去重判断经 `ui.AlbumDetailIDGetter` 接口访问 `AlbumDetailMenu`） |
| `internal/plugins/artist/` | 第五个真实插件：歌手集群整体提取（6 个菜单，key 与内置注册逐一相同；`hot_artists` 声明「热门歌手」主菜单入口，主菜单项经 After 锚点 `high_quality_playlists` 保持原内置位置；集群内跳转经 `ui.BuildMenu`/`ui.BuildMenuOrToast`；被 ui 共享的 `ArtistDetailOpts`/`ArtistsOfSongOpts` 留在 ui，ui 侧去重判断经 `ui.ArtistDetailIDGetter`/`ui.ArtistsOfSongSongIDGetter` 接口访问 `ArtistDetailMenu`/`ArtistsOfSongMenu`） |
| `internal/plugins/recommend/` | 第六个真实插件：推荐集群整体提取（5 个菜单，key 与内置注册逐一相同；`daily_songs`/`daily_playlists`/`personal_fm`/`recent_songs`/`ranks` 各自声明主菜单入口「每日推荐歌曲 / 每日推荐歌单 / 私人FM / 最近播放歌曲 / 排行榜」，主菜单项经 After 锚点 `MainMenuStart`/`daily_songs`/`user_collect`/`hot_artists`/`search_type` 保持原内置位置；登录门控经 `BaseMenu` 的 `ToLoginPage` 转发，`personal_fm` 的 BottomOutHook 经 `m.Player()` 访问播放列表 API） |
| `internal/plugins/playlist/` | 第七个真实插件：歌单/云盘集群整体提取（5 个菜单，key 与内置注册逐一相同；`user_playlist`（参数化 provider，`UserPlaylistOpts` 携带 userID，主菜单入口经 `RegisterMainMenuItemAfter` 以 After 锚点 `daily_playlists` + `UserID: ui.CurUser` 构造）/`user_collect`/`high_quality_playlists`/`could` 各自声明主菜单入口「我的歌单 / 我的收藏 / 精选歌单 / 云盘」，`playlist_detail`（参数化，`PlaylistDetailOpts` 携带 playlistID）为纯跳转目标不声明主菜单入口，主菜单项经 After 锚点 `daily_playlists`/`user_playlist`/`ranks`/`recent_songs` 保持原内置位置；`user_collect` 的子菜单经 `ui.MustBuildNoArg` 按 `album_sub_list`/`artists_sub_list` key 构建（由 album/artist 插件注册）；集群内及 ui 侧跳转经 `ui.BuildMenu` 按 key 解析 `playlist_detail`；被 ui 共享的 `PlaylistDetailOpts`/`ui.CurUser` 常量留在 ui（`operate.go`/`player.go`/`menu_add_to_user_playlist.go` 使用）） |
| `internal/plugins/search/` | 第八个真实插件：搜索集群整体提取（2 个菜单，key 与内置注册逐一相同；`search_type` 声明主菜单入口「搜索」，主菜单项经 After 锚点 `album_menu` 保持原内置位置；`search_result` 为参数化 provider（`ui.SearchResultOpts` 携带 searchType），集群内跳转经 `ui.BuildMenu`/`ui.BuildMenuOrToast` 按 key 解析 `search_result`/`album_detail`/`playlist_detail`/`artist_detail`/`user_playlist`/`dj_radio_detail`；搜索页注册转发 `ui.NewSearchPage`——`SearchPage` 类型与 wordsInput/result/searchType 状态留在 ui（shell 单例，ui 侧需为 `SearchPage` 提供 `WordsInput()`/`Result()`/`SearchType()` 导出访问器供 `SearchResultMenu` 读取）；`SearchType`/`St*` 常量与共享 opts 留在 ui） |
| `internal/plugins/song/` | 第九个真实插件：单曲集群整体提取（2 个菜单，key 与内置注册逐一相同；`simi_songs`/`add_to_user_playlist` 均为参数化 provider（`ui.SimiSongsOpts` 携带 song、`ui.AddToUserPlaylistOpts` 携带 userID/song/IsAdd），纯跳转目标不声明主菜单入口；`ui.CurUser` 常量留在 ui，本集群以 `ui.CurUser` 引用） |
| `internal/wasm/` | WASM 插件运行时（实验性）：契约（`contract.go` JSON Request/Response，guest 线协议冻结）、manifest（`manifest.go`）、wazero 封装（`plugin.go`：alloc/run/dealloc 调用协议、单 uint64 打包结果、超时 watchdog、实例互斥串行化）、目录扫描 + SHA-256 校验 + 生命周期（`manager.go`）、**sink 管线**（`sink.go`：`RegistrySink` 接口（`RegisterCommands`，Replace 语义）、`CommandsOf`（MenuDecl→frontend.Command，PluginID 盖章）、`callWasm`（CommandContext→wasm.Request 映射 + `p.Run` + Response→CommandResult）、`LoadAndRegister` 便捷入口）、**cordis 化**（`cordis.go`：`ManagerPlugin`（root 插件，Start=NewManager + provide `ServiceWasmManager`，Dispose=Close）+ 动态 `wasmPlugin`（Start=CommandsOf+sink 注册、Stop=逐命令 Unregister、Dispose=关实例）+ `LoadIntoScope`（扫描目录→每插件 wasmPlugin→scope.AddAndStart，热加载能力））。WASM 插件已从专属 `WasmPluginMenu` 迁轨 B 并 cordis 化：TUI 在 frontend scope 下挂 wasm 子 scope（`ui/plugin_scope.go`）加载、命令在 `registerCommandMenus()` 适配为 CommandMenu；WebUI 建独立 wasm scope 加载、命令出现在 `/api/commands`；headless 不加载（无命令消费方，文档化非目标） |
| `examples/wasm/hello/` | WASM 插件示例：标准 Go wasip1 reactor（`//go:wasmexport` alloc/dealloc/run/hang），编译 `GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared`，见 docs/plugin_development.md「WASM 插件（实验性）」 |
| `hack/new-plugin/` | 插件开发脚手架：Go 生成器（`main.go`，标准库跨平台）+ 模板（`templates/plugin_name/`，占位符 `{{plugin_name}}`/`{{MenuType}}`/`{{menu_key}}`/`{{menu_title}}`/`{{menu_after}}`），`go run ./hack/new-plugin -name X -menu Y` 生成 `internal/plugins/<name>/` 骨架（registry/menu/test/README），不写死路径（`-dir`/`-templates` 可覆盖），详见 docs/plugin_development.md「插件脚手架」 |
| `internal/ui/event_handler.go` | 键盘/鼠标事件处理 |
| `internal/ui/operate.go` | 右键操作表 |

### 核心接口

**Menu 接口**（`internal/ui/menu.go`）：
```go
type Menu interface {
    model.Menu
    IsPlayable() bool
    IsLocatable() bool
}
type SongsMenu interface { Songs() []structs.Song }
type PlaylistsMenu interface { Playlists() []structs.Playlist }
```

**Player 接口**（`internal/player/player.go`）：
```go
type Player interface {
    Play(music URLMusic)
    Pause()/Resume()/Stop()/Toggle()
    Seek(duration time.Duration)
    PassedTime()/PlayedTime() time.Duration
    Volume()/SetVolume()/UpVolume()/DownVolume()
    State() types.State
    Close()
}
```

### 播放引擎

| 引擎 | 平台 | 特点 |
|------|------|------|
| Beep（默认） | 跨平台 | MP3/FLAC/OGG/WAV；可选 MP3 无缝播放，使用 go-mp3 时额外支持编码器延迟和尾部填充裁剪 |
| DLNA | 跨平台 | 设备投送 |
| MPV | 跨平台 | IPC 控制（命令连接带排空 goroutine 防 mpv 发送缓冲堆积，切歌带 file-loaded 超时 watchdog 兜底） |
| MPD | Linux | 远程服务器 |
| AVFoundation | macOS | 原生集成 |
| MediaPlayer | Windows | WinRT API |

### 事件处理

**文件**：`internal/ui/event_handler.go`

支持 40+ 键盘操作、鼠标事件（单击/双击/滚轮/右键）、可配置快捷键。

### 其他模块

- **歌词**：LRC/YRC 格式，支持 smooth/wave/glow 渲染模式；未匹配的云盘歌曲（含旧播放快照）优先通过云盘歌词接口获取内嵌歌词，失败时回退普通歌曲歌词接口
- **播放列表**：列表循环/顺序/单曲循环/随机/无限随机/智能心动模式
- **远程控制**：MPRIS(linux)、Now Playing(macOS)、System Media(Windows)。macOS 侧 `internal/macdriver/mediaplayer` 与 `internal/macdriver/cocoa/unnotifications.go` 的框架加载（Dlopen）失败不 panic 静默降级：MediaPlayer.framework 需 macOS 10.12.2+、UserNotifications.framework 需 10.14+，旧系统上 class 查找得 nil、objc 消息发送安全返回 0，对应功能（Now Playing 远程控制/系统通知）自动禁用而不影响主应用启动
- **状态栏组件**：`model.DefaultStatusBar.Components` 可注入任意 `StatusBarComponent` 到居中区域；展示项只实现 `View`，可点击项额外实现 `InteractiveStatusBarComponent`，以自身局部坐标处理命中和事件。状态栏负责布局、边界与指针，不持有业务回调。播放队列适配器的 `musicfox` 前缀自行打开 `https://github.com/go-musicfox/go-musicfox`，队列位置与音质文本不触发。
- **存储**：BoltDB，存储用户信息、播放状态、播放列表快照和桌面歌词窗口位置/显示器
- **扫码登录**：二维码登录（`internal/ui/login_qr_page.go`）通过 `internal/core/qrlogin` 包的 `qrlogin.GetKey` / `qrlogin.CheckStatus` 直接请求网易云接口（客户端已从 ui 提升至 core 侧子包，扫码成功经 `engine.CompleteQRLogin` 编排完成登录），而非 `service.LoginQRService`。客户端基于 `github.com/imroc/req/v3` 的 `SetTLSFingerprintChrome()` 模拟 Chrome 的 TLS ClientHello 指纹并附带完整浏览器 Headers，规避服务端对非浏览器客户端的反爬检测（462）。`qrKeyClient` 请求 unikey 时不带 Cookie；轮询用 `newQRLoginClient(cookieJar)` 每次绑定最新的全局 CookieJar（`neteaseutil.GetGlobalCookieJar()`），避免登录 Cookie（`MUSIC_U`）写入过期 jar。轮询前 `util.ApplyRequestStrategy` 注入反风控 Cookie。462 拦截为网易云动态风控行为，与 IP/地域/频率相关，并非必现。
- **实时频谱**：仅 macOS `osx` 播放引擎支持；`MTAudioProcessingTap` 经 PureGo 获取 PCM，由 `internal/player/spectrum.go` 异步分析。PCM 更新频谱目标值，`github.com/charmbracelet/harmonica` 的临界阻尼弹簧以每帧推进，避免 PCM 回调间隙使动画停顿。频谱分析与 UI 刷新均使用 `[main].frameRate`。`[main.visualizer]` 支持 `enable`、`maxHeight`（默认 `0` 为不限制；正数限制频谱行数）及 `fullCharHalfBlock`、`fullCharFullBlock`、`emptyCharBlock` 字符配置（各取首个 Unicode 字符，默认分别为 `▌`、`█`、空格）。`SpectrumRenderer` 将相邻频段分组为由低至高的横向进度条（低频在底部），以三态字符显示：满单元、使用前景/背景双色渐变的半单元、无样式空白单元，并以半单元为粒度提供双倍横向幅度分辨率。未限制时频谱会占满歌词与歌曲信息之间的可用行，顶部及歌曲信息前各保留一行空白。
- **主题**：`[dark.app]` / `[light.app]` 的显式非透明 `background` 作为应用、菜单行、Startup 页面（普通内容和全屏特效）及 Search、Login、QR Login、Last.fm 授权/二维码授权/API account 等自定义页面完整窗口的终端单元格背景；未设置或 `transparent` 时 Startup 与自定义页面均不写入背景，Default 与 Transparent 保持终端透明。`[*.highlights]` 不提供菜单背景覆写。内置主题状态栏的面包屑与时间字体继承 Default 主题。仅 Transparent 主题通过 `statusBar.nuggetLabelFg` / `nuggetLabelBg` 将 `»` 配置为 Primary 前景、透明背景，并以 `statusBar.breadcrumbHover` 将面包屑 hover 前景设为 Primary；右键菜单 hover 文本也使用 `Primary`。菜单选中项的 `menuSelectedSepLeft` / `menuSelectedSepRight` 配置圆角分隔符；右侧符号优先替换原有末尾填充空格，无可替换空格时增加独立符号宽度，不压缩或截断正常内容。居中菜单始终按标称宽度计算起点，额外圆角单元格不会改变标题列；双列模式下左列额外单元格从右列结构性前导空格中吸收，右列起点保持不变。用户通过快捷键或右键菜单切换主题时，会将名称保真写入 `theme.activeTheme`，下次启动加载该主题；系统外观变化仅切换当前主题的明暗变体。
- **自定义表单页面**：`internal/ui/page_layout.go` 统一提供标题返回按钮、面包屑委托、输入框焦点/hover 样式、按钮本地化和文本光标定位；Search、Login 与 Last.fm API account 页面必须按其渲染结果记录命中坐标，并以 `pageSubmitButton` / `pageButton` 保持键盘和鼠标 active 状态一致。
- **封面图**：Kitty 封面图在应用背景预留的透明区域渲染，位于文本和单元格背景之下；popup 仅覆盖与其重叠的部分。背景排除矩形仅覆盖图像实际填充的前 `rows-1` 行，Kitty 未填充的末行仍由主题 `AppBackground` 绘制。独立页面若需清理封面，必须在 deferred `MenuToPage.BeforeEnterMenuHook` 中调用 `coverRenderer.ClearDisplayed()`，使删除发生在 loading 的 Main 帧之后、目标页面切换之前；Last.fm API account 入口遵循此规则。
- **网页登录**：macOS / Windows / Linux 三平台（非三平台 `login_webview.go` 桩）。Login 页「网页登录」入口在 WebView 检测不可用时自动隐藏：`webviewLoginAvailable()` 按 build tag 实现运行时检测——macOS WebKit 系统自带恒可用；Windows 经 `webviewloader.GetAvailableCoreWebView2BrowserVersionString("")` 检测 WebView2 Runtime；Linux 以 `webkitgtk.WebKitGTKVersion() == 0` 判定 WebKitGTK 库缺失；不可用时按钮不渲染，鼠标 hover/点击与键盘 right 导航同步跳过。检测结果在 `NewLoginPage` 缓存于 `LoginPage.webviewAvailable`。可用时弹出原生 WebView 窗口加载 `https://music.163.com/#/login`，通过原生 cookie API 每 ~1s 轮询检测 `MUSIC_U` cookie（HttpOnly，`document.cookie` 读不到）。页面层 `internal/ui/login_webview_page.go` 平台无关，仿 QR 登录页用 `tea.Tick` 轮询消费事件（`WebviewLoginEvent{CookieString, WindowClosed}`，缓冲 4 channel select-default 永不 close）；拿到 cookie 后复用 `utils/app.ParseCookieFromStr` → `neteaseutil.SetGlobalCookieJar` → `apputils.RefreshCookieJar()` 链路校验，失败则保持窗口继续轮询。共享常量在 `login_webview_consts.go`。三平台实现：
  - **macOS**：`internal/macdriver/webkit`（purego + objc 桥接，无 CGO）弹出 WKWebView，使用 `nonPersistentDataStore`（内存不落盘）；`WKHTTPCookieStore getAllCookies:`（macOS 27 改名，init 时经 `instancesRespondToSelector:` 兼容 `getAllCookiesWithCompletionHandler:`）轮询；block 回调运行在 WebKit 私有队列，只做数据 marshal 经 channel 回传。控制器 `internal/ui/login_webview_darwin.go` + 主线程派发 `login_webview_helper_darwin.go`（`MusicfoxWebviewLoginHelper`，仿 `internal/desktop_lyrics/helper_darwin.go` 的 performSelectorOnMainThread 模式）。窗口显示期间将全局 `SetActivationPolicy(Prohibited)`（`internal/runtime/runtime_darwin.go:26`）临时切换为 `Accessory` 并激活，关闭时恢复；所有退出路径（成功/用户关窗/取消）都恢复。
  - **Windows**：`internal/ui/login_webview_windows.go`，依赖 `github.com/13thgoutham/go-webview2` v1.0.25（wailsapp/go-webview2 v1.0.23 的修复 fork，零 CGO，`pkg/webview2` 生成 COM 绑定 + `webviewloader` 从内存加载 WebView2Loader.dll）。单 UI 线程模型：messageLoop goroutine（`LockOSThread` + CoInitializeEx）承载建窗（RegisterClassExW/CreateWindowExW）→ 环境/控制器创建（回调经消息泵派发）→ `SetTimer` WM_TIMER 轮询 `GetCookieManager().GetCookies("")`（空 uri 取全部 cookie 含 HttpOnly）→ 消息循环退出后 `cleanup()`（`controller.Close()` + 经 IUnknown vtable Release 槽释放 COM 对象）。Close() 幂等 = 置 closed + `PostMessage(WM_CLOSE)`。WebView2 Runtime：Win11 自带，Win10 绝大多数已装。**依赖替换原因**：wailsapp/go-webview2 v1.0.23 生成的 `ICoreWebView2AddScriptToExecuteOnDocumentCreatedCompletedHandler` / `ICoreWebView2CallDevToolsProtocolMethodCompletedHandler` / `ICoreWebView2ExecuteScriptCompletedHandler` 三个 handler 的 Invoke 回调以 `string`（16 字节）承载 LPCWSTR 结果，新版 Go（≥1.22）`compileCallback` 逐参数限制 ≤ uintptr（8 字节）导致包 init 时 panic；上游无修复版本（最新即 v1.0.23，PR #36/#37 未合并），故改用 13thgoutham fork v1.0.25（含 PR #36/#37 的生成器级修复）；`make vendor` 后无需额外打补丁。
  - **Linux**：`internal/webkitgtk`（purego 直调 WebKitGTK，零 CGO）运行时优先加载 6.0 栈（`libwebkitgtk-6.0.so.4` + `libgtk-4.so.1`），失败自动回退 4.1 栈（`libwebkit2gtk-4.1.so.0` + `libgtk-3.so.0`，覆盖 Debian 11/12、Ubuntu 22.04 等旧发行版）。6.0 使用 `WebKitNetworkSession` 与 `webkit_cookie_manager_get_all_cookies`，WebView 经 `g_object_new` 的 `"network-session"` 构造属性绑定会话（WebKitGTK 6.0 起移除了 `webkit_web_view_new_with_*` 便利构造函数，`webkit_web_view_new_with_network_session` 符号在任意版本都不存在）；4.1/4.0 使用旧版 `WebContext` / `CookieManager` API，其中 4.1 链接 libsoup 3、4.0 链接 libsoup 2，并通过带 URI 的 `webkit_cookie_manager_get_cookies` 查询 Cookie。所有支持栈都不可用不 panic（`WebKitGTKVersion()` 返回 0，导出函数降级 no-op，controller 走 WindowClosed 回退，且 Login 页「网页登录」入口自动隐藏）。**符号缺失降级**：库可加载（RTLD_LAZY）但缺少对应 API 符号时，`registerSymbol`（`purego.Dlsym` 失败静默跳过而非 `RegisterLibFunc` panic）使对应函数指针保持 nil，栈可用性检查判定该栈不可用并继续回退，最终 `WebKitGTKVersion()` 返回 0 走表单登录降级，不影响主应用启动。版本差异分发：`gtk_window_new` 带参/无参、`gtk_window_set_child` vs `gtk_container_add`、`gtk_widget_show_all`、窗口关闭信号 "close-request"（GTK4）vs "delete-event"（GTK3，均经 `GSignalConnectCloseRequest` 统一连接，回调统一 `func(window, extra uintptr) int32` 返回 0 允许关闭）。`internal/ui/login_webview_linux.go`：GTK 线程模型——`runGTK` goroutine（`LockOSThread`）内 `gtk_init_check` → 创建对应 API 栈的 WebView → `gtk_window_new` → `g_timeout_add` 轮询 Cookie → `gtk_main` 阻塞；Close() 幂等 = `gtk_main_quit`（跨线程安全）。无 DISPLAY/WAYLAND_DISPLAY 时直接发 WindowClosed 降级二维码登录。每次轮询释放 GList/SoupCookie/GError，退出时 `g_object_unref` window/webView/session。

## 开发指南

### 添加新菜单

1. 创建 `internal/ui/menu_new_feature.go`，嵌入 `baseMenu`，实现 `Menu` 接口
2. 在该文件 `init()` 注册 `MenuProvider`：`RegisterMenu[T](key, factory)`（参数契约用结构体 `T` 表达，如 `PlaylistDetailOpts`；无参菜单用共享的 `NoArgMenuOpts`）
3. 跳转点改走注册表：`BuildMenu`/`mustBuildNoArg`/`buildMenuOrToast`（`registry.go`），key 为菜单 `GetMenuKey()`；父菜单与 operate 不再硬编码构造函数

### 添加新页面

1. 创建页面类型（实现 `model.Page`），在 `init()` 注册 `RegisterPage[T](key, factory)`（如 `login`、`search`、`lastfm_custom_api`）
2. 导航经 `BuildPage`/`buildPageOrToast`（`registry.go`）；shell 持有引用的单例页（search）在 `NewNetease` 经 `BuildPage` 构建
3. 页面持有 shell 引用用于导航（`MustMain`/`RerenderCmd`），业务能力经 `menuServices` 访问器解析（如 `svc.Lastfm()`），不直连 shell 服务字段

### 插件开发（外部边界）

对外插件边界（注册表 API、`framework.Context`/`ServiceOf` 服务解析、`Scope`/`Plugin` 生命周期、快捷键/操作扩展点（`keybindings.RegisterOperate` + `ui.RegisterOperateHandler`）、右键菜单扩展点（`ui.RegisterContextMenuContrib`）、编译期注册示例与行为保持契约、**WASM 插件（实验性）**）见 `docs/plugin_development.md`。当前插件形态：编译期注册（import + `init()`）+ 运行时动态加载的 WASM 插件（MVP：菜单动作 + 文本结果，宿主 `internal/wasm` 经 wazero 沙箱执行，不重编译即启用）。Go `plugin` 共享库 / 子进程形态不支持。

**WASM 插件（实验性）**：用户把插件目录（`manifest.toml` + `.wasm` reactor）放入 `[plugins] wasmDir`（默认 `<配置目录>/wasm-plugins`），启动时 `NewNetease` 经 `wasm.NewManager().LoadDir` 扫描、SHA-256 校验（manifest `sha256` 非空时）并用 wazero 加载（`WithStartFunctions("_initialize")`、内存上限 128 页、无文件系统、单调用 5s 超时 watchdog 关闭实例），随后 `ui.WithPlugin(manifest.id, ...)` 归属注册其菜单（`WasmPluginMenu` 嵌入 `ui.BaseMenu`，Action 调 wasm 导出并解析 `wasm.Response`：`toast`/`view`（MVP 以多行 toast 呈现）/`open_url`（`open.Start`）/`exec`（无 shell），经 `app.Notify` 线程安全投递）与主菜单项（`after` 锚点），`[plugins] disabled` 按 manifest id 生效。调用协议：guest 导出 `alloc`/`dealloc`/`export`（默认 `run`），`export(reqPtr, reqLen) uint64` 返回打包 `(outPtr<<32)|outLen`（Go wasmexport 仅单结果）。单个插件加载/注册失败仅记日志不阻断启动（recover 隔离）。示例见 `examples/wasm/hello/`（编译 `GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared`）。**加载发生在 TUI 前端（`NewNetease`）**——headless 模式无菜单，不加载 WASM 插件。

**插件配置化启停**：每个插件把所有注册包进 `ui.WithPlugin(id, name, func(){ ... })`（id 为插件目录名，同 id 可多文件多次声明幂等合并），使作用域内的 `RegisterMenu`/`RegisterPage`/`RegisterMainMenuItem*`/`RegisterStartupHook` 归属到该插件。用户经配置 `[plugins] disabled = ["search", "checkupdate"]` 禁用插件后：被禁用插件的**主菜单入口隐藏、启动钩子不执行**；菜单 key 注册与 `BuildMenu` 跳转不受影响（禁用插件菜单仍可被按 key 跳入）。锚点完整性处理：被禁用插件的主菜单项仍保留在 `NewMainMenu` 的 After 锚点链 entries 中（其 key 仍是其它项的合法锚点，链完整性基于全部注册项校验），仅在显示阶段被跳过——不 panic、后继项位置自然前移。

插件经聚合器接入：`internal/plugins/plugins.go` 空导入各插件子包（如 `internal/plugins/checkupdate`，首个真实插件——`CheckUpdateMenu` 嵌入 `ui.BaseMenu`，init() 经 `ui.RegisterMenu` 注册 `"check_update"`，并声明主菜单入口 `RegisterMainMenuItemAfter("check_update", "检查更新", "help", nil)` 与启动钩子 `RegisterStartupHook(startupCheck)`；`internal/plugins/lastfm` 为第二个真实插件——Last.fm 菜单/页面整体提取，init() 注册 `"last_fm"` 菜单与 `"lastfm_auth"`/`"lastfm_custom_api"` 页面（opts 携带 `ui.MenuServices`，`MenuServices` 是 `*menuServices` 的导出别名），并声明主菜单入口 `RegisterMainMenuItemAfter("last_fm", "LastFM", "radio_dj_type", nil)`；`internal/plugins/dj` 为第三个真实插件——「主播电台」DJ/电台集群整体提取（10 个菜单，key 与内置注册逐一相同，集群内跳转经导出的 `ui.BuildMenuOrToast`/`ui.MustBuild`/`ui.MustBuildNoArg`，`radio_dj_type` 声明主菜单入口「主播电台」）；`internal/plugins/album` 为第四个真实插件——「专辑列表」专辑集群整体提取（8 个菜单，key 与内置注册逐一相同，`album_menu` 声明主菜单入口「专辑列表」，被 ui 共享的 `AlbumDetailOpts` 留在 ui、去重判断经 `ui.AlbumDetailIDGetter` 接口访问）；`internal/plugins/artist` 为第五个真实插件——「热门歌手」歌手集群整体提取（6 个菜单，key 与内置注册逐一相同，`hot_artists` 声明主菜单入口「热门歌手」，被 ui 共享的 `ArtistDetailOpts`/`ArtistsOfSongOpts` 留在 ui、去重判断经 `ui.ArtistDetailIDGetter`/`ui.ArtistsOfSongSongIDGetter` 接口访问）；`internal/plugins/recommend` 为第六个真实插件——「推荐/播放历史」集群整体提取（5 个菜单，key 与内置注册逐一相同，`daily_songs`/`daily_playlists`/`personal_fm`/`recent_songs`/`ranks` 各自声明主菜单入口）；`internal/plugins/playlist` 为第七个真实插件——「歌单/云盘」集群整体提取（5 个菜单，key 与内置注册逐一相同，`user_playlist` 经 `RegisterMainMenuItemAfter` 参数化主菜单入口（After `daily_playlists`）以 `UserID: ui.CurUser` 构造，`user_collect`/`high_quality_playlists`/`could` 各自声明主菜单入口，`playlist_detail` 为纯跳转目标）；`internal/plugins/search` 为第八个真实插件——「搜索」集群整体提取（2 个菜单，key 与内置注册逐一相同，`search_type` 声明主菜单入口「搜索」，搜索页注册转发 `ui.NewSearchPage`）；`internal/plugins/song` 为第九个真实插件——「单曲」集群整体提取（2 个菜单，key 与内置注册逐一相同，`simi_songs`/`add_to_user_playlist` 均为参数化纯跳转目标，`ui.CurUser` 常量留在 ui 供本集群引用）），`cmd/musicfox.go` 空导入聚合器触发注册。插件 key 不得与内置 key（`expectedMenuKeys`）冲突；`ui` 不得反向导入插件包（import cycle）。插件主菜单项在 `NewMainMenu` 经 **After 锚点链**归并（每个入口声明前驱项 key，`MainMenuStart` 为第一项、空 After 追加在末尾），复现插件化前的原始顺序（key 必须已注册，未注册时启动 panic 作为完整性信号；无 `Build` 的入口 key 必须是无参菜单，参数化入口经 `RegisterMainMenuItemWith(key, title, build)` 以插件自身 options 构造，显式位置经 `RegisterMainMenuItemAfter(key, title, after, build)` 声明 After 锚点——插入菜单只需声明一个锚点，其后各项不漂移；链完整性——After 目标缺失/重复锚点/环/孤儿——在 NewMainMenu 构建时断言并列出违规 key；内置项仅帮助（After `last_fm`）仍声明在链内（「搜索」入口已随 search 插件化移除，由插件声明，After 锚点 `album_menu` 不变）；内置 LastFM 入口已随 lastfm 插件化移除，内置「主播电台」入口已随 dj 插件化移除，内置「专辑列表」入口已随 album 插件化移除，内置「热门歌手」入口已随 artist 插件化移除，内置「每日推荐歌曲 / 每日推荐歌单 / 私人FM / 排行榜 / 最近播放歌曲」入口已随 recommend 插件化移除，内置「我的歌单 / 我的收藏 / 精选歌单 / 云盘」入口已随 playlist 插件化移除，内置「搜索」入口已随 search 插件化移除，现均为插件主菜单项，经各自 After 锚点回到原内置位置）；启动钩子在 `InitHook` 启动序第 10 步（用户/登录就绪后、自动播放前）经 `runStartupHooks` 按注册序调用，每个 hook 带 recover 隔离，panic 仅记日志不阻断启动。

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

### 更新依赖

使用 `make vendor` 更新 vendor 目录，该命令会依次执行 `go mod tidy`、`go mod vendor` 并复制 CGo 依赖的头文件。

```bash
make vendor
```

#### foxful-cli 依赖更新流程

foxful-cli 是本项目的核心 UI 框架，位于 `vendor/github.com/anhoder/foxful-cli/`。更新流程：

1. 在 `go.mod` 中使用 `replace` 指向本地 foxful-cli 仓库开发
2. foxful-cli 改动完成后，在 `~/Desktop/foxful-cli` 中 commit 并推送
3. 在 foxful-cli 仓库打 tag（如 `v1.0.5`）
4. 在本项目 `go.mod` 中注释掉本地 `replace`，将版本升为新 tag
5. 执行 `make vendor` 同步 vendor 目录

### 跨平台构建兼容性

**修改 `Makefile` 或 `hack/` 目录下的构建脚本时，必须确保 Windows 系统兼容。**

#### 原因
go-musicfox 支持 macOS/Linux/Windows 三平台，构建脚本需要在不同环境下正确执行。

#### 检查清单
- [ ] 新增的 target 是否包含 Unix 特有命令（`which`、`cp`、`mkdir -p`、`chmod`、`tar`、`awk` 等）
- [ ] 是否使用 `$(OS)` / `Windows_NT` 条件分支提供 Windows 替代逻辑
- [ ] Shell 脚本（`.sh`）是否有对应的 PowerShell（`.ps1`）实现
- [ ] 路径分隔符是否兼容（使用 `$(PACKAGE_ROOT)` 而非 `` `pwd` ``）
- [ ] 重定向和设备文件是否有 Windows 替代（`nul` 对应 `/dev/null`，`where` 对应 `which`）

#### Windows 兼容模式参考

```makefile
# 使用 OS 条件分支
ifeq ($(OS),Windows_NT)
    # Windows 逻辑
else
    # Unix 逻辑
endif
```

| Unix 写法 | Windows 替代 |
|-----------|-------------|
| `which <cmd>` | `where <cmd>` |
| `/dev/null` | `nul` |
| `` `pwd` `` | `$(PACKAGE_ROOT)`（Make 变量） |
| `<cmd> >/dev/null 2>&1` | `<cmd> >nul 2>&1` |
| `<cmd> || { cmd2; }` | `<cmd> || ( cmd2 )` |
| `hack/*.sh` | `hack/*.ps1`（PowerShell 实现） |

**例外**：纯交叉编译/CI 内部工具（如 `hack/init_linux_env.sh`、`hack/init_windows_env.sh` 等 Docker 内部脚本），仅在 Linux Docker 容器中执行，无需适配 Windows。

## 文档维护准则

### 重要准则：修改代码后需维护 AGENTS.md

**所有贡献者在修改代码后，必须检查并更新 AGENTS.md 文档，防止文档腐化。**

#### 何时需要更新文档
- 添加、删除或重命名核心文件或目录
- 新增功能模块或组件
- 修改项目结构或架构
- 更改 API 接口或配置格式
- 添加新的播放器引擎、菜单类型、渲染模式
- 修改关键路径或重要流程
- **快捷键变更**：新增、删除或修改快捷键后，必须同步更新 README 中的快捷键说明部分（help 菜单是动态从 keybindings 配置生成的，无需手动修改）

#### 更新检查清单
- [ ] 目录结构是否准确反映当前项目结构
- [ ] 核心文件路径是否正确
- [ ] 接口定义是否与代码一致
- [ ] 新增功能的说明是否完整
- [ ] 开发指南是否需要补充
- [ ] 关键文件路径表格是否需要更新

#### 更新优先级
| 变更类型 | 优先级 | 说明 |
|---------|--------|------|
| 架构变更 | **高** | 必须立即更新 |
| 新增核心模块 | **高** | 必须添加说明 |
| 文件路径变更 | **中** | 及时更新路径表 |
| 细微修复 | **低** | 可批量更新 |

#### 维护建议
- 保持文档与代码同步，避免技术债务积累
- 使用一致的术语和格式
- 添加代码示例时确保与实际代码匹配
- 定期（如每月）审查文档完整性
- PR 审查时应包含文档检查

#### 文档腐化警告信号
- 文件路径与实际不符
- 过时的 API 接口描述
- 已删除功能仍出现在文档中
- 章节结构混乱或重复
- 与 README/CHANGELOG 存在矛盾

违反此准则可能导致：新贡献者无法理解项目结构、开发效率降低、文档失去参考价值、维护成本增加。

## Git 提交规范

### 重要准则：Git Commit Message 必须遵循 Conventional Commits 规范

**所有贡献者在提交代码时，必须遵循 [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) 规范。**

#### 提交格式

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### Type 类型

| Type | 说明 |
|------|------|
| `feat` | 新功能开发 |
| `fix` | Bug 修复 |
| `docs` | 文档更新 |
| `style` | 代码格式调整（不影响功能） |
| `refactor` | 代码重构（既不修复bug也不添加功能） |
| `perf` | 性能优化 |
| `test` | 测试相关 |
| `chore` | 构建工具、辅助工具、配置变更 |
| `revert` | 回滚提交 |

#### 示例

```
feat(player): 添加 MPV 播放引擎支持

- 支持多种音频格式
- 实现播放进度控制
- 优化内存使用

Closes #123
```

#### 规范要求

1. **必须使用英文**：提交信息、描述均使用英文
2. **动词开头**：描述部分以动词开头，使用现在时态
3. **长度限制**：标题不超过 50 字符
4. **Body 可选**：复杂变更可添加详细说明，每行不超过 72 字符
5. **引用 Issues**：在 footer 中使用 `Closes #xxx` 关联 Issue
