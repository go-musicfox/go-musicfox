# 插件生态架构（Plugin Ecosystem Architecture）

> 本文档描述 go-musicfox 的插件生态统一架构方案：**三层全量 cordis 化**——① 内部业务插件、② WASM 插件、③ headless daemon server，全部以 `internal/framework` 的 cordis 语义（`Scope`/`Plugin` 生命周期 + `Context` DI + `EventEmitter`）为统一底座。目标分支：`feat/plugin-framework-playback` 及其后续。
>
> 本文档是设计文档（Phase 6 预留边界），当前处于「待实施」状态，分阶段迁移路线见文末。

## 概述

从 Phase 3.x 起，go-musicfox 建立了前端可拔插（`internal/frontend` 注册表）、UI-free 核心引擎（`internal/core`）、轨 B 命令贡献（跨前端共享）、WASM 动态加载（`internal/wasm` sink 管线）等能力。本方案在这些基础上做最后一步统一：**用 cordis 的 Scope 生命周期模型统辖"运行时激活"层**。

核心矛盾：现有插件是**声明期贡献**（编译期 `init()` 注册菜单/页面/命令，早于一切 ctx），cordis 是**运行时激活**（`Scope.Start` 时启动插件）。全量化 = **把注册时机从 `init()` 迁移到 `Scope.Start`**，使插件具备统一的生命周期（Start/Stop/Dispose）、依赖注入（Deps）、事件通信（EventEmitter）与运行时启停语义。

## 现状基础

- `internal/framework` 的 cordis 语义**已完整**：`Context` DI（`context.go:20-82`）、`Scope`/`Plugin`/`PluginWithDeps` 生命周期 + rollback + 递归 Dispose（`plugin.go:39-203`）、`EventEmitter` 四类 handler（listener/middleware/parallel/serial，panic 隔离，`events.go:27-162`）——但**生产接入几乎为零**（仅 shareSvc/lastfm 经 `servicePlugin`，`core/services_scope.go:14-77`）。
- 9 个业务插件全部 `init()` + `ui.WithPlugin` 归属记录（`plugins/*/registry.go`），聚合器空导入（`plugins/plugins.go`）。
- WASM 加载点散在 TUI（`ui/load_wasm_plugins.go`）与 WebUI（`webui/run.go`），headless 不加载。
- headless 控制通道一连接一请求一响应（`headless/server.go:173-209`），事件面 no-op。
- 事件是单消费者 `core.Observer`（`core/observer.go:19-63`）；WebUI 用 observer→broadcaster 变相多播。

**三个共同断点**：① 事件无多订阅者总线；② WASM 加载点分散、无生命周期归属、命令注册表无 Unregister；③ headless 无事件推送能力。

## 一、统一 Scope 拓扑

**单一全局 Context + 分层 Scope**。Context（服务注册表）保持进程级单例（沿用 `ProvideIfAbsent` 的"同名服务全局唯一"约定）；Scope 只负责生命周期（谁先 Start、谁先 Stop、谁的子插件随谁清理）。服务解析跨 scope 直通全局 Context，不做命名空间隔离。

```
root Scope（进程级，core.Engine 创建，core/engine.go 持有）
│  Add 顺序约束：[R0]
│
├── [root 直属插件] core 服务构造器插件 ×8
│      order: lastfm → trackManager → lyricService → desktopLyrics
│             → loginService → userService → player → eventBus
│      （顺序由 Deps 依赖图强制，Add 序即注册序）
├── [root 直属插件] wasmManagerPlugin              ← wazero runtime 唯一持有者
│      （Dispose = Manager.Close）
│
└── frontend Scope（前端创建，TUI/WebUI/headless 各自构建，互斥）
    │
    ├── [frontend 直属插件] 前端服务插件
    │      TUI:    uiServicesPlugin（coverRenderer/menuRegistry/pageRegistry）
    │      WebUI:  webuiServicesPlugin（token/静态资源/broadcaster 基座）
    │      headless: daemonPlugin
    │
    ├── [frontend 直属插件] 9 个编译期业务插件（TUI 专属；WebUI/headless 不 Add）
    │      经 framework.RegisterPlugin 由 plugins 聚合器注册，
    │      TUI 前端按 configs.IsPluginEnabled 过滤 AddWithEnabled
    │
    └── [frontend 动态子 Scope] wasm 插件子 Scope
           ├── Add 时机：scope.Start 之后运行时 AddAndStart（动态插件）
           └── 每个 wasm 插件目录 → 一个 wasmPlugin adapter
```

### Add 顺序约束

- **[R0] root Scope**：① 服务构造器插件先于任何消费方（player 依赖 5 服务、userService 依赖 slot）；② `wasmManagerPlugin` 先于前端 wasm 加载；③ `eventBusPlugin` 先于 daemon/WebUI 订阅方。约束强制手段：**Deps 失败即 Start 失败**（`plugin.go:90-97`），错序显式报错而非静默 nil——与现状"顺序靠注释约定"（`core/engine.go:56-59`）的关键差异。
- **[R1] frontend Scope**：① 前端服务插件先于业务插件（菜单工厂需要 `menuServices`/注册器服务）；② 9 个业务插件按聚合器声明序 Add（替代 init() 无序注册，保证主菜单锚点链与命令序确定性）；③ 禁用插件**不 Add**，从源头消除"入口隐藏但 key 可跳"中间态。
- **[R2] 业务插件内部**：插件间保持注册表解耦（跳转经 `BuildMenu`/`BuildMenuOrToast` 按 key 解析），cordis 化只改"谁何时注册"，不改"注册表是唯一耦合面"。

## 二、framework 增强

全部落在 `internal/framework/`，保持零业务依赖。

```go
// plugin.go 新增 —— 支持 disabled 与运行时 Add
func (s *Scope) AddWithEnabled(p Plugin, enabled bool) error
// 注册插件并在 disabled 时跳过 Start（保留在切片中，Stop/Dispose 仍按序执行）。

func (s *Scope) AddAndStart(ctx *Context, p Plugin) error
// scope 已 Start 后注册并立即启动插件（动态插件/WASM 热加载用）；scope 未 Start 时
// 退化为 Add。返回 err 时插件已回滚（Stop + 从切片移除）。

func (s *Scope) Plugins() []Plugin
// 快照（调试/断言/PluginInfo 收集用）。

// noop.go 新增 —— 纯菜单/纯命令插件零样板基座
type NoopPlugin struct{}
func (NoopPlugin) Start(*Context) error { return nil }
func (NoopPlugin) Stop() error          { return nil }
func (NoopPlugin) Dispose() error       { return nil }

// contrib.go 新增 —— 贡献接口（生命周期锚点 + 断言/文档）
type PluginBase struct { Enabled bool }
type StartupHookContributor interface { StartupHook() func() }   // 替代包级 startup_hooks 注册表
type MenuContributor       interface { MenuKeys() []string }
type PageContributor       interface { PageKeys() []string }
type MainMenuContributor   interface { MainMenuKeys() []string }
type CommandContributor    interface { CommandKeys() []string }
type ContextMenuContributor interface { ContextMenuKeys() []string }
```

> 泛型限制：`RegisterMenu` 是包级泛型函数（Go 禁止泛型方法），贡献接口**无法**承载泛型方法。因此接口只做"生命周期锚点 + 断言"，真实注册仍在插件 `Start` 内调用现有注册函数。

### command 注册表扩展（WASM Stop 的硬前提）

```go
// internal/frontend/command.go 扩展（仍零业务依赖）
func UnregisterCommand(key string)   // 删除 key 对应命令，不存在时 no-op
func ReplaceCommand(cmd Command)     // 替换 key 现有命令；key 不存在等价 RegisterCommand
// 内部 cmdOrder 变可删除结构（保持 Commands() 注册序快照语义）
```

### startup hooks 收编

`framework/startup_hooks.go` 的包级 `startupHooks` 改为由 root Scope 收集：`RunStartupHooks` 签名不变（兼容壳），内部遍历 scope 内实现 `StartupHookContributor` 的插件并按 `Enabled` 过滤。`ui.RegisterStartupHook` 保留为注册转发壳。

## 三、三层 cordis 化形态

### 3.1 内部业务插件（9 个）

`init()` + `ui.WithPlugin(id, name, func(){...})` → 插件类型 + `framework.RegisterPlugin(id, 构造器)`：

```go
// internal/plugins/checkupdate/plugin.go（示意，其余 8 个同构）
type Plugin struct {
    framework.PluginBase
    svc *ui.MenuServices   // Deps 注入
}
func (p *Plugin) Deps(ctx *framework.Context) error {
    p.svc = ui.NewMenuServices(ctx)
    return nil
}
func (p *Plugin) Start(ctx *framework.Context) error {
    ui.RegisterMenu("check_update", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
        return &CheckUpdateMenu{BaseMenu: base}, nil
    })
    ui.RegisterMainMenuItemAfter("check_update", "检查更新", "help", nil)
    return nil
}
func (p *Plugin) StartupHook() func() { return startupCheck }
func (p *Plugin) MenuKeys() []string  { return []string{"check_update"} }
func (p *Plugin) MainMenuKeys() []string { return []string{"check_update"} }
```

- **lastfm 插件**经 `Deps` 用 `framework.ServiceOf[*lastfm.Client](ctx, core.ServiceLastfm)` 解析服务；页面 opts 照旧携带 `ui.MenuServices`（形状不变，测试零改动）。`ui.WithPlugin` 的归属盖章保留——从"init 时"变"Start 时"。
- **聚合器**（`plugins/plugins.go`）：`init()` 保留但只做 `framework.RegisterPlugin(id, 构造器)`。
- **禁用语义**：`AddWithEnabled(p, configs.IsPluginEnabled(id))`。禁用 = 不 Start = 不注册贡献 = 不执行 startup hook。**契约变化**：禁用插件菜单 key 不在注册表，"按 key 跳入禁用插件菜单"不再可能（原契约 `ui/plugin_registry.go:15-18`），需 CHANGELOG/文档声明。

### 3.2 core 8 服务构造器插件

`NewEngine` 手写组装（`core/engine.go:60-125`）+ `registerServices` → 8 个构造器插件：

```go
// core/services_plugins.go（示意：playerPlugin）
type playerPlugin struct {
    e *Engine
    opts PlayerOptions
    player *Player
}
func (p *playerPlugin) Deps(ctx *framework.Context) error {
    // ServiceOf 解析 lyricService/trackManager/desktopLyrics/lastfm，缺失即返回错误
    return nil
}
func (p *playerPlugin) Start(ctx *framework.Context) error {
    p.player = NewPlayer(p.opts)
    return provideIfAbsent(ctx, ServicePlayer, p.player)
}
func (p *playerPlugin) Dispose() error { return p.player.Close() }
```

- **构造与生命周期合一**：8 个插件各自持有实例，`NewEngine` 退化为纯装配器（建 ctx/scope → Add 8 插件 → Add wasmManagerPlugin → Start）。配置读取移入插件构造。
- **所有权迁移顺序**：第一阶段 `Dispose` 委托 `Engine.Close`（`engine.go:223-238`）现有清理，引擎对外 API（`engine.go:128-174` 访问器）不动，前端零改动；最终阶段 `Dispose` 直接持有清理，`Engine.Close` 只做 `scope.Dispose()`。
- **`registeredServiceNames` 断言**（`services.go:56-65`）：从"对比 registerServices 结果"改为"对比 scope Start 后 `Context.Names()`"。
- **新增服务**：`ServiceDispatcher`、`ServiceEventBus`（`*framework.EventEmitter`）。

### 3.3 WASM 插件

```go
// internal/wasm/cordis.go
type ManagerPlugin struct { mgr *Manager }          // root 插件
func (p *ManagerPlugin) Start(ctx) error {          // NewManager + provideIfAbsent(ServiceWasmManager)
    ...
}
func (p *ManagerPlugin) Dispose() error { return p.mgr.Close(context.Background()) }

type wasmPlugin struct {                             // 每 wasm 目录一个
    framework.PluginBase
    mgr *wasm.Manager       // Deps: ServiceWasmManager
    p   *wasm.Plugin        // LoadDir 产物
    sink RegistrySink       // TUI: tuiWasmSink / WebUI: webuiWasmSink
    cmds []frontend.Command
}
func (p *wasmPlugin) Start(ctx) error {             // RegisterCommands（Replace 语义）
    p.cmds = wasm.CommandsOf(p.p)
    return p.sink.RegisterCommands(p.p, p.cmds)
}
func (p *wasmPlugin) Stop() error {                 // 逐命令 Unregister
    for _, c := range p.cmds { frontend.UnregisterCommand(c.Key) }
    return nil
}
func (p *wasmPlugin) Dispose() error { return p.p.Close(ctx) }
```

- **动态加载 + 生命周期**：前端 scope 构建时扫描 `wasmDir` → 每目录 `LoadDir` → 建 `wasmPlugin` → **`AddAndStart` 挂入已 Start 的 scope**。Scope 运行时 Add 是硬需求。
- **热加载（S8）自然化**：目录变更 → 新目录经 `AddAndStart` 挂入、旧插件 `Stop`（Unregister）+ `Dispose`（Close）。生命周期由 scope 自动管理。
- **sink 归并**：TUI/WebUI 两个 sink 统一为 `RegisterCommands` 内部用 `ReplaceCommand`，不再需要 recover-dedup 防御（key 冲突变替换而非 panic）。
- headless 仍不加载（无命令消费方，维持文档化非目标）。

### 3.4 daemon server 插件

```go
// internal/headless/daemon_plugin.go
type DaemonPlugin struct {
    framework.PluginBase
    dispatcher *core.Dispatcher          // Deps: ServiceDispatcher
    emitter    *framework.EventEmitter   // Deps: ServiceEventBus
    opts       DaemonOpts                // network/addr（unix socket / win 临时 TCP）
    server     *Server                   // 复用现有 Server，生命周期由插件驱动
}
func (p *DaemonPlugin) Start(ctx) error {
    // server.Listen() + go server.AcceptLoop(ctx)
    // 订阅事件：emitter.Listener("player.song_changed", p.enqueue) 等
    //   —— listener 只投递到内部队列，绝不阻塞播放 goroutine
    return nil
}
func (p *DaemonPlugin) Stop() error    { p.server.Close(); return nil }
func (p *DaemonPlugin) Dispose() error { return p.server.CloseAllConns() /* + 退订 emitter */ }
```

- **依赖面**：Dispatcher 命令面（现状已有）、EventEmitter 事件源（取代 no-op `HeadlessObserver`）、通用 `EventSink` 服务（从 `webui/broadcast.go` 提取，daemon 与 WebUI 共用）。
- **事件循环**：`enqueue` 只做 channel 投递（buffer 有上限，满则丢帧——与 WebUI"慢连接只丢自己帧"哲学一致），独立 goroutine 从队列取帧经 broadcaster 分发。
- **订阅模型**：headless 协议升级为**长连接双能力**——请求-响应（兼容 `musicfox ctrl`，`headless/client.go` 零改动）之外新增 `subscribe`/`unsubscribe` 帧；订阅集合是 daemon 内部状态（`map[int64]map[string]bool`），**不进 Scope/Context**（Scope 管生命周期，订阅管数据面，正交）。

### 3.5 前端接线

- **TUI**（`ui/netease.go:66-128` 重构）：engine（root scope Start 完）→ 建 frontend scope → Add `uiServicesPlugin` + 9 业务插件（按启用过滤）→ Start → WASM 动态挂载 → `registerCommandMenus` → 断言 → `NewMainMenu` → `engine.Startup`。
- **WebUI**（`webui/run.go` 重构）：同构建 frontend scope → Add `webuiServicesPlugin`（含 broadcaster 基座）→ WASM 挂载 → Start。`WebUIObserver` 删除，改在 webui 服务插件 Start 时注册 emitter listener（事件帧格式不变，静态资源零改动）。
- **headless**（`headless/run.go` 重构）：NewEngine → frontend scope（仅 daemonPlugin）→ Start → 阻塞。`--once` 模式不建 daemon，直接 `NewDispatcher(engine).Dispatch`（现状保持）。

## 四、事件面设计

### 事件名常量（wire 名与 webui 帧名对齐）

```go
// core/events.go
EvSongChanged   = "player.song_changed"
EvStateChanged  = "player.state_changed"
EvPosition      = "player.position"        // 已限流
EvPlaylistEnd   = "player.playlist_exhausted"
EvRerender      = "player.rerender"
EvLogin         = "auth.login_succeeded"
EvStartupPhase  = "startup.phase"
```

### 双写点（全部在 core，与现有 observer 调用并排）

- `core/player.go`：OnStateChanged / OnPosition / OnSongChanged / PlaylistExhausted / Rerender 各 emit 点。
- `startup.go:43-47`（`emitStartupPhase`）、`engine.go:179-219`（`LoginCallback` 末尾发 `EvLogin`）。
- 实现 `emit(ctx, name, payload)`：先 observer（现有调用不变），再 `emitter.Emit`。

### 职责边界

| 维度 | core.Observer（保留） | EventEmitter（新增总线） |
|---|---|---|
| 消费者 | 单消费者：TUI 渲染（`ui.Player`） | 多订阅者：daemon、WebUI、插件 |
| 语义 | 进程内渲染驱动、同步、nil-safe | 事件总线、广播、panic 隔离 |
| 性能 | OnPosition 高帧率（渲染 ticker） | 低频事件为主；position 由订阅方二次限流 |
| 错误 | 无返回值 | 返回 error（listener 错误中止链） |

**并发契约**：emitter listener 在播放 goroutine 内同步执行——**所有订阅方 listener 必须是"投递后即返回"**（enqueue 只写 channel），严禁在 listener 内直接写 socket/阻塞。`Parallel` 分支用于高开销订阅方。

### 插件间通信事件

| 事件 | 生产者 | 消费者示例 |
|---|---|---|
| `auth.login_succeeded` | `LoginCallback` / `CompleteQRLogin` | lastfm 重载用户、WebUI 广播 login 帧 |
| `playlist.changed` | `ReinitializePlaylist` 等 | 未来插件监听队列变更 |
| `plugin.started` / `plugin.stopped` | frontend scope 包装 emit | daemon 通知客户端插件集变化 |
| `theme.switched` | TUI | 未来插件响应主题 |

**需补能力**：`EventEmitter.Unregister(name)`（daemon/WebUI 插件 Dispose 时退订；按 name 清空四类 handler，粗粒度足够）。

## 五、注册时机迁移后的启动序列

### TUI 启动序列（`NewNetease` 重构后）

```
0.  [进程 init] 各包 init() 只做 RegisterPlugin（聚合器）+ 前端 Register（frontend）
1.  NewEngine(opts)                    ← core.NewEngine 纯装配
      a. 创建 ctx + root scope
      b. Add 8 服务构造器插件 + wasmManagerPlugin（顺序见 [R0]）
      c. root scope.Start(ctx)（Deps → Start → Provide；失败 rollback）
2.  前端建 frontend scope（TUI）
      a. Add uiServicesPlugin
      b. Add 9 个业务插件（AddWithEnabled(p, configs.IsPluginEnabled(id))）
      c. frontend scope.Start(ctx)（插件 Start = 注册菜单/页面/主菜单项/命令）
3.  WASM 加载（原 ui/netease.go:120 位置，语义不变但机制变）
      a. 扫 wasmDir → LoadDir 每目录
      b. 每目录建 wasmPlugin → wasmScope.AddAndStart(ctx, wp)
4.  registerCommandMenus()             ← 读 frontend.Commands()（WASM + 业务插件命令都在）
5.  断言（原 netease.go:114-115 后移至此，语义从"init 后"变"Start 后"）
      AssertMenuRegistryComplete / AssertPageRegistryComplete
6.  NewMainMenu(base)（天然在全部注册之后）
7.  engine.Startup(ctx, observer)（第 10 步 RunStartupHooks 改 scope 收集）
8.  CloseHook → engine.Close → root scope.Dispose()（逆序：wasm 子 scope → 业务插件 → 服务构造器）
```

### 关键时序裁决

1. **断言位置**：从 init 后 → frontend scope.Start 完成后、NewMainMenu 前。`expectedMenuKeys` 语义不变（只锁内置集，不含插件 key）。
2. **WASM 先于 registerCommandMenus**：硬约束（`command_menu.go:164` 注释），步骤 3 在 4 前。
3. **headless/WebUI 同理**：headless = 步骤 1 + daemonPlugin；WebUI = 步骤 1 + webui 服务插件 + WASM + Startup（无 TUI 断言）。

### 测试改造策略

| 测试类别 | 改造 |
|---|---|
| framework 单测 | 补 `AddWithEnabled`/`AddAndStart`/`Unregister` 用例，其余不动 |
| core 服务完备性 | 改调"scope Start"，断言从 `Context.Names()` 对比 |
| 插件注册测试 | 统一 fixture：`framework.RegisterPlugin` 收集 + `startPluginForTest(t, id)` helper 触发 Start |
| 启动断言测试 | 断言函数抽出为 `assertRegistriesComplete(t, ctx)` 纯函数 |
| init() 时序敏感测试 | 显式注册序（现状已基本显式） |

## 六、分阶段迁移路线

| 阶段 | 内容 | 依赖 | 风险 |
|------|------|------|------|
| **P1** | framework 增强（AddWithEnabled/AddAndStart/Plugins/NoopPlugin/PluginBase/贡献接口/EventEmitter.Unregister/startup hooks 收编） | 起点 | 低 |
| **P2** | command 注册表 Unregister/Replace | 可与 P1 并行，须先于 P4 | 低 |
| **P3** | core 8 服务构造器插件化（NewEngine 纯装配、Engine.Close 委托 scope、新增 ServiceDispatcher/ServiceEventBus） | 可与 P1 并行，P5 依赖 | 中 |
| **P4** | 播放事件双写 EventEmitter；WebUI 从 observer 迁 emitter | P3 后 | 中 |
| **P5** | 9 个业务插件 init→Start 迁移 + disabled 语义切换 | P1+P3 后 | **高** |
| **P6** | WASM cordis 化（ManagerPlugin + wasmPlugin + 动态挂载 + sink 归并） | P2+P4 后 | 中 |
| **P7** | daemon 插件化 + 协议订阅升级（subscribe/unsubscribe，兼容旧帧） | P3+P4 后，与 P5/P6 可并行 | 中 |
| **P8** | 热加载收尾 + PluginInfo 换源 + 文档同步 | P6+P7 后 | 低 |

**可并行**：P1/P2/P3 起始；P5/P6/P7 在 P3+P4 后并行。

## 七、风险清单与对策

| # | 风险 | 等级 | 对策 |
|---|---|---|---|
| R1 | 注册时机后移（init→Start）：Start 前读到空注册集 | 高 | 启动序列固化；断言函数后移；注释 + 断言 + 测试 fixture 三层防线 |
| R2 | disabled 语义变化：禁用插件菜单 key 不再注册，"按 key 跳入"契约破坏 | 高 | 明确为**有意行为变更**写入 CHANGELOG；`buildMenuOrToast` 兜底（跳转变 toast）；判定为全量 cordis 的必要代价（禁用=不存在 才是生命周期自洽形态） |
| R3 | Start 顺序敏感性（Add 序即注册序） | 中 | Deps 显式声明使错序显式失败；Add 序文档化（[R0]/[R1]）；插件间注册表解耦已保证互不依赖 |
| R4 | emitter listener 阻塞播放 goroutine | 高 | 订阅方纪律（enqueue-only）；代码评审强制项；Parallel handler 供高开销方 |
| R5 | WASM 动态 Add 的并发 | 中 | Scope 加简单 mutex；单测锁定"Start 后 AddAndStart" |
| R6 | 测试面广（依赖 init 时序） | 中 | 统一 fixture + 断言纯函数化；P5 集中处理 |
| R7 | P7 协议升级向后兼容 | 中 | subscribe 是新增 cmd（未知 cmd 走 Dispatcher 默认分支报错）；ctrl 每次新建连接天然不受影响 |

## 八、全量 vs 分层：诚实权衡

**真正产生价值的全量部分**（非形式成本）：
1. **P4 事件双写**——消灭 WebUI observer→broadcaster 变相多播，daemon 事件面就绪。
2. **P6 WASM cordis 化**——加载点统一 + Unregister 使热加载成为生命周期自然推论。
3. **P7 daemon 插件化**——headless 从"哑终端"升级为可订阅 daemon。
4. **P3 服务插件化**——把"顺序靠注释"变为"Deps 显式强制"。

**偏形式成本的部分**（全量化的主要代价）：
1. **9 个业务插件的 Start/Deps 样板**——`NoopPlugin` 基座 + 注册器服务注入可压到 ~30 行/插件。纯形式。
2. **P5 的 init→Start 迁移 + disabled 语义变化（R2）**——行为变更 + 最大测试面，全量最贵的部分。唯一收益是语义自洽（禁用=不存在），无功能增量。当前 `[plugins] disabled` 是启动静态配置，今日无运行时启停需求。
3. **P8 热加载**——S8 是实验性 WASM 的未来项，P6 机制已覆盖。

**投入优先级建议**：P4/P6/P7（真实断点）＞ P3（顺序安全网）＞ P5（语义自洽，纯形式+行为变更）。若资源紧张，P5 的"注册时机后移"可暂缓（保留 init 注册、仅加 scope 生命周期壳），不影响其它三层 cordis 化——这是全量前提内唯一的弹性点。
