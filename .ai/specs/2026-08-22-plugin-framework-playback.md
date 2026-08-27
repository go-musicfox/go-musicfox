# go-musicfox 插件化方案（cordis 式核心框架 + 播放链路插件化，三阶段）

## 📝 TLDR

go-musicfox 的核心协调器 `internal/ui/netease.go` 将 login、search、player、lyricService、trackManager 等全部挂载其上，加功能需动多处、风险高；同时维护者希望形成第三方插件生态。本 spec 参考 koishi 的 cordis 元框架，分三阶段落地：Phase 0 用 30 行证伪原型验证 UNM 中间件形态可行性；Phase 1 先以零框架依赖的有序 func 切片 + 现有接口落地播放链路插件化（URL 中间件链 + playable source provider）；Phase 2 再移植 cordis 语义的核心框架层（Context、服务注册/覆写、作用域生命周期、事件链），Go 手写零外部依赖。行为约束：重构不改变任何用户可见行为，插件第一刀采用编译期注册（import + 注册），外部边界仅在接口层预留。

## 📝 Problem Statement

### 双驱动动机

1. **核心协调器难改**：`internal/ui/netease.go` 是巨型协调器，login、search、player、lyricService、trackManager 等全部挂载其上。新增功能（如新音源 provider、播放前 hook、新菜单）需要改动多处既有代码，改动成本高、回归风险大。
2. **第三方插件生态**：维护者希望未来能形成第三方插件生态，让社区能力以插件形式接入，而不必 fork 主仓库。

两个动机均已在前期对话中由用户明确确认。

### 现状改动成本锚点（真实场景对比）

> 以下对比基于当前代码现状（`internal/ui/player.go`、`internal/track`、`cmd/musicfox.go`、`utils/netease/songinfo.go`、`internal/player/player.go`）。行号以本 spec 撰写时 master 分支为准，实现时以实际为准。

**场景 A：新增一个播放源 provider（如本地文件/自定义音源）**

- 现状改动点：修改 `internal/track/manager.go` 的 `ResolvePlayableSource`（downloaded/cached/remote 三源 if-else 逻辑，约 288-332 行）加入新分支；同步修改 `internal/ui/player.go` 播放链（271-306、538-546 行）中可能涉及的调用；需要理解 trackManager 全貌才能不破坏三源优先级。
- 目标态改动点：实现一个 `PlayableSourceProvider` 接口的 provider，注册进 provider 链（Phase 1 后约 1 个新文件 + 1 行注册）。

**场景 B：新增一个播放前 hook（如音量归一化、URL 改写、日志）**

- 现状改动点：在 `internal/ui/player.go` 的 `ResolvePlayableSource` 之后、`p.Play(...)` 之前（271-306 行区间）插入代码；若涉及 UNM 相关逻辑还需同步检查 `utils/netease/songinfo.go`（36 行 SkipUNM、77-81 行 ProxyURL 改写）与 `internal/ui/player.go`（275-277 行 HasBannedPathSuffix 拦截）的顺序约束。
- 目标态改动点：实现一个 `URLMiddleware` 函数，追加到中间件链（Phase 1 后约 1 个新文件 + 1 行注册），无需触碰既有调用顺序。

**场景 C：接入第三方插件（未来生态）**

- 现状：不存在插件边界，第三方只能 fork 后直接改源码。
- 目标态：实现 `Plugin` 接口 + 编译期注册表，第三方以 import + 注册方式接入（Phase 2 后）。

### 当前播放链路（已核实，`internal/ui/player.go`）

```
PlaySong(254) → getPlayInfo(271) → trackManager.ResolvePlayableSource(ctx, song)
  → [SkipInvalidTracks 时 HasBannedPathSuffix 拦截(275-277)]
  → p.Play(player.URLMusic{URL, Song, Type})(302-306)
```

- `getPlayInfo`（538-546）核心一行即 `trackManager.ResolvePlayableSource(context.Background(), song)`（539），返回 `track.PlayableSource`（Type/Path/Info 三字段）→ 映射为 URL + MusicType。
- `ResolvePlayableSource`（`internal/track/manager.go:137-150`）→ `resolveSongSource`（274-339）三源顺序：① 已下载（downloaded，279-288）→ ② 缓存（cached，301-305）→ ③ 远程（remote，322-328，经 `m.fetcher.FetchPlayableInfo`）；经 `sfGroup.Do` 去重，Remote 且缓存开启时后台预热。
- `Manager` 自身是 option 注入协调器：`WithFetcher`（102-107）、`WithCacher`、`WithNameGenerator`、`WithSongQuality`；`Fetcher` 接口定义于 `internal/track/fetcher.go:16-21`（FetchPlayableInfo / FetchStream / FetchLyric / FetchCloudLyric）。
- `HasBannedPathSuffix` 定义于 `utils/netease/filter.go:14`。

UNM 引擎（processor）本体在 vendored `github.com/cnsilvan/UnblockNeteaseMusic/processor`，可被外部直接调用：`RequestBefore(request) *Netease`（processor.go:90）、`RequestAfter(request, response, netease)`（166）、`Request(request, remoteUrl)`（154）、`type Netease`（75-88）；SDK 侧 `util/request.go:219` UNMFlag、`util/config.go:12-24` 全局变量、`util/config.go:40-42` ProxyURL 非空强制关 UNM（`UNMSwitch = false`）。当前主仓库对 processor 无任何直接 import（全走 SDK 内部调用）。go-musicfox 自有侧仅四块薄胶水：

- `cmd/musicfox.go:72-78` — 6 个 UNM 全局变量静态赋值（UNMSwitch/Sources/SearchLimit/EnableLocalVip/UnlockSoundEffects/UNMProxyURL）
- `utils/netease/songinfo.go:36` — 主路径 SkipUNM: true（V1 URL 服务经 SDK 时跳过 UNM 引擎）
- `utils/netease/songinfo.go:77-81` — ProxyURL 前缀改写（`music.163.com/package/` → `UNM.ProxyURL/package/`）
- `internal/ui/player.go:275-277` — HasBannedPathSuffix 拦截（SkipInvalidTracks 配置开关）

现有可复用的扩展点：`track.Manager` 的 option 注入、`Fetcher` 接口、foxful-cli 的 `model.Hook`（菜单生命周期钩子，非播放中间件）。仓库中**不存在**任何播放链中间件/事件总线/拦截器机制，需要从零建立。

## 📝 Proposed Solution

### 总体路线

参考 `~/Desktop/cordis`（koishi 的元框架）设计插件能力：**先内部解耦为主线，同时预留外部插件边界**。第一刀范围 = 核心框架最小集 + 播放链路插件化（菜单/页面 provider、WebUI/GUI 前端抽象、外部加载机制均后置）。

一份 spec，内部明确三阶段，每阶段独立 PR、独立可交付/可回滚：

- **Phase 0**：30 行证伪原型——写一个最小中间件，对一首已知被封锁的歌曲直接调用 vendored SDK 的 UnblockNeteaseMusic processor，与今天 SDK 路径产出的 URL 对比。对得上 → 中间件形态可行，进入 Phase 1；对不上 → 中间件形态选错，回炉讨论（此决策点不得跳过）。
- **Phase 1**：播放链路插件化——URL 中间件链（UNM 注入点）+ playable source provider 接口落地。**不依赖核心框架**，先用有序 func 切片 + 现有接口落地。验收双示范：UNM 中间件插件（现有 UNM 逻辑抽成插件，行为一致）+ source provider 接口（`ResolvePlayableSource` 背后抽象化，网易云音源成为第一个 provider 实现）。
- **Phase 2**：核心框架层——cordis 语义较完整移植，Go 手写**零外部依赖**：Context + 服务按名注册/解析/可覆写 + 作用域生命周期（插件 start/stop/dispose 递归清理）+ 事件链（listener/middleware/parallel/serial）。

### 否决的替代方案

| 方案 | 否决理由 |
|------|----------|
| 引入 uber fx / dig 等现成 DI 框架 | Go 无法移植 cordis 的 Proxy/原型链机制，"ctx 属性访问解析服务"退化为 `map[string]any` + 类型断言，即自研轻量 DI 容器——接受自维护、零生态成本 |
| 运行时动态加载插件（go plugin） | 跨平台问题大（Windows/darwin 支持不完整），第一刀不做 |
| 外部插件加载机制（进程/脚本适配器） | 后置；第一刀仅编译期注册 + 接口层预留 |
| 全量插件化（菜单/页面 provider + 播放链路 + 前端抽象一起） | 体量过大、风险过高，用户从 A 菜单 provider / B 播放链路 / C 全量中选择 B |

### 行为约束

- **重构不改变任何用户可见行为**（UNM 作为插件后表现与今天完全一致），靠 `make lint` / `make test` / `make build` 门禁 + 手动回归保障。
- 插件第一刀采用编译期注册（import + 注册），不做运行时动态加载。
- 外部边界仅在接口层预留（未来可加进程/脚本适配器）。

## 📝 Architecture

### Phase 1 架构（播放链路插件化）

```
Player.PlaySong → getPlayInfo → trackManager.ResolvePlayableSource(ctx, song)
  → [URLMiddleware 链: UNM 中间件(可选)、其他] → p.Play(player.URLMusic{URL, Song, Type})
```

- **URL 中间件链**：有序 func 切片。中间件签名遵循标准中间件模式（`func(next ...)` 或简单顺序包装，以 Phase 0/1 实现验证为准）。UNM 注入点位于 `ResolvePlayableSource` 之后、`Play` 之前。
- **PlayableSourceProvider 接口**：`ResolvePlayableSource` 背后抽象化。网易云音源成为第一个 provider 实现（现状 downloaded/cached/remote 三源逻辑整体搬迁为 provider 内部实现，行为不变）。
- **注册方式**：编译期注册——`init()` 或显式注册函数，import 即生效。

### Phase 2 架构（cordis 语义核心框架）

- **Context 容器**：`map[string]any` + 类型断言解析服务（cordis Proxy 的退化形态，brief 已确认接受）。
- **服务注册/解析/覆写**：服务按名注册，可按作用域覆写。
- **作用域生命周期**：插件 start/stop/dispose 递归清理（子作用域随父作用域销毁）。
- **事件链**：listener / middleware / parallel / serial 四类事件处理方式。
- **零外部依赖**：Go 手写，不引外部 DI 框架。

### 边界与复用

- 修改 vendored SDK 行为是 Non-goal（Phase 0 证伪原型仅只读调用验证；除非原型证伪后用户重新决策）。
- 菜单/页面 provider、WebUI/GUI/TUI 前端抽象、外部加载机制后置切片；当前仅保证接口设计不堵死该方向。

### 为什么一份 spec 承载三阶段（而非三个独立 spec）

- **同一战略目标**：三阶段是"cordis 式插件化"这一单一战略的连续切片（先证伪形态 → 播放链路落地 → 框架层），不是三个互不相关的功能；拆成三份会导致动机、行为约束、UNM 三层约束、等价性标准重复叙述且难以保持一致。
- **用户已明确决策**：前期对话中用户选择"一份 spec 内部三阶段，每阶段独立 PR（challenger CRITICAL 2 的解决方案）"；每阶段独立 PR/独立可回滚的编排已保证独立交付能力，scope cohesion 通过编排而非文档拆分达成。
- **决策点贯通**：Phase 0 的证伪结论直接决定 Phase 1 的 UNM 搬迁深度（只包胶水 vs 搬引擎），跨阶段依赖需要单一文档承载。

## 📝 Data Model

无新增持久化数据模型。播放状态、播放列表快照等既有 BoltDB 存储不变。插件注册表为进程内存态（编译期注册，无运行时配置）。

## 📝 API Contracts

### Phase 1

**URLMiddleware**（草案，以 Phase 0/1 实现验证为准；置于 `getPlayInfo` 之后、`p.Play` 之前，作用于已解析出的 URL/URLMusic）：

```go
// 中间件签名草案：ctx 携带播放上下文，urlMusic 为可改写目标，next 调用链中下一环
type URLMiddleware func(ctx context.Context, urlMusic *player.URLMusic, next func(context.Context, *player.URLMusic) error) error
```

**PlayableSourceProvider**（草案）：

```go
// 将现状 track.Manager.resolveSongSource（downloaded→cached→remote 三源）
// 抽象化；网易云音源成为第一个（默认）实现，返回类型沿用 track.PlayableSource
type PlayableSourceProvider interface {
    ResolvePlayableSource(ctx context.Context, song structs.Song) (track.PlayableSource, error)
}
```

注册点：`internal/track`（fetcher.go 已有 Fetcher 接口与 WithFetcher 注入模式，provider 注册沿用类似模式）与 `internal/player/player.go`（NewPlayerFromConfig 硬编码 switch 为引擎注册 provider 化的候选点，但引擎注册 provider 化属于 Phase 1 可选范围，视实现难度决定是否后置）。

### Phase 2

**Service 注册/解析**（cordis 退化形态）：

```go
type Context struct { services map[string]any }
func (c *Context) Service(name string) any                    // 解析
func (c *Context) Provide(name string, svc any)               // 注册
func (c *Context) Override(name string, svc any)              // 覆写
```

**Plugin 生命周期**：

```go
type Plugin interface {
    Start(ctx *Context) error
    Stop() error
    Dispose() error
}
```

**事件链**：listener / middleware / parallel / serial 四类注册与派发。

> 具体签名在 Phase 2 设计时细化；本 spec 只锁定语义与边界。

## 📝 UI/UX

本 spec 为纯内部重构/框架层设计，**无用户可见界面变化**（行为保真约束）。不涉及 mockup 制作。

当前 app 截图非必需：重构不改变任何用户可见表面，视觉证据无增量价值。

## 📝 Edge Cases & Failure Scenarios

### UNM 三层交互约束（必须逐条复刻）

1. **V1 主路径 SkipUNM**：`utils/netease/songinfo.go:36` 主路径 `SkipUNM: true`——UNM 中间件必须保持该语义，不得在非预期路径触发 UNM。
2. **ProxyURL 非空强制关引擎**：`vendor/.../util/config.go:40-42`——ProxyURL 非空时 UNM 强制关闭；`utils/netease/songinfo.go:77-81` ProxyURL 前缀改写也必须保持。
3. **banned-path 拦截**：`internal/ui/player.go:275-277` `HasBannedPathSuffix` 拦截——顺序和条件稍错即改变现有用户播放行为。

### 等价性测试标准（不能只是口号）

- **歌曲样本集**：固定歌曲样本，覆盖：普通歌曲、被封锁歌曲、云盘歌曲、代理场景（ProxyURL 非空）各若干。
- **断言内容**：UNM 插件开关前后，URL 输出一致；回退路径（UNM 失败 → 原 URL）判定一致；播放行为一致。
- **执行方式**：Phase 1 验收时对样本集逐一对比（中间件链前后 URL 快照对比），配合 `make lint` / `make test` / `make build` 门禁 + 手动回归。

### 其他失败场景

- **中间件链空/未注册**：行为退化为现状（无中间件时原链路直接 Play）。
- **provider 解析失败**：回退现状逻辑或返回错误，用户侧表现为现有失败提示，不新增用户可见差异。
- **Phase 2 作用域泄漏**：dispose 递归清理必须覆盖子作用域，防止插件 stop 后残留 goroutine/监听器。

## 📝 Risks & Impact Review

### 风险

- **行为保真风险（高）**：UNM 三层交互约束顺序/条件稍错即改变用户播放行为。缓解：等价性测试样本集 + 三阶段独立回滚。
- **体量风险（高）**：重构播放链路并引入新框架层。缓解：三阶段拆分，每阶段独立 PR、独立可交付/可回滚；Phase 0 证伪原型先行验证形态。
- **vendored SDK 边界（中）**：UNM 引擎在 vendor 内，只读边界/最小触碰。缓解：Phase 0 仅只读调用验证；Phase 1 只包胶水不搬引擎（原型结论为"中间件可行"时）。

### 兼容性与回滚

- 无用户可见行为变化，无配置格式变化（`configs.UpgradeConfig` 不受影响）。
- 每阶段独立 PR 可独立回滚：Phase 1 回滚 = 恢复 `ResolvePlayableSource` 直调；Phase 2 回滚 = 框架层仅新包引入，旧路径保留。

### Blast radius

- Phase 0：无生产代码改动（原型独立于主程序）。
- Phase 1：`internal/track`、`internal/ui/player.go` 播放链、UNM 胶水四处。
- Phase 2：新增框架包 + `netease.go` 逐步接入（接入动作本身可后置到独立 PR）。

## 📋 Phasing

### Phase 0 — 30 行证伪原型

- **目标**：验证"UNM 中间件形态"可行。
- **做法**：写一个最小中间件，对一首已知被封锁的歌曲直接调用 vendored SDK 的 UnblockNeteaseMusic processor，与今天 SDK 路径产出的 URL 对比。
- **决策点（不得跳过）**：对得上 → 中间件形态可行，进入 Phase 1；对不上 → 中间件形态选错，回炉讨论。
- **产物**：原型代码（独立于主程序，或最小集成验证）+ 对比结论。

### Phase 1 — 播放链路插件化

- **目标**：URL 中间件链（UNM 注入点）+ playable source provider 接口落地。
- **做法**：不依赖核心框架，先用有序 func 切片 + 现有接口落地。
- **验收**：双示范——UNM 中间件插件（现有 UNM 逻辑抽成插件，行为一致）+ source provider 接口（`ResolvePlayableSource` 背后抽象化，网易云音源成为第一个 provider 实现）。
- **行为保真**：等价性测试样本集通过；`make lint` / `make test` / `make build` 门禁通过。

### Phase 2 — 核心框架层

- **目标**：cordis 语义较完整移植，Go 手写零外部依赖。
- **范围**：Context + 服务按名注册/解析/可覆写 + 作用域生命周期（插件 start/stop/dispose 递归清理）+ 事件链（listener/middleware/parallel/serial）。
- **接入**：框架层落地后，`netease.go` 核心协调器的接入动作可作为独立 PR 渐进执行（本 spec 不强制一次性迁移全部组件）。

## 📋 Implementation Plan

### Phase 0 步骤

1. 编写最小 UNM 中间件原型：对已知被封锁歌曲，直接调用 vendored `github.com/cnsilvan/UnblockNeteaseMusic/processor` 的 `RequestBefore` / `Request`（或经 SDK `util.Request` 开启 UNMFlag 的对照路径），产出 URL → 验证: 与今天 SDK 路径（`utils/netease/songinfo.go` FetchPlayableInfo，`SkipUNM: true`）产出的 URL 对比一致
2. 记录原型结论（可行/不可行 + 理由）→ 验证: 决策点文档化；不可行则停止并回炉讨论

### Phase 1 步骤

1. 定义 URLMiddleware 类型与中间件链容器（有序 func 切片，支持追加）→ 验证: 单元测试覆盖空链/单链/多链
2. 将现有 UNM 逻辑抽为 UNM 中间件插件（四块薄胶水 + 三层交互约束复刻），注册进链 → 验证: 等价性测试样本集 URL 对比一致
3. 定义 PlayableSourceProvider 接口，将 `ResolvePlayableSource` downloaded/cached/remote 三源逻辑抽象为网易云 provider 实现 → 验证: 三源优先级行为与现状一致（单元测试）
4. 播放链接入：`internal/ui/player.go` 播放链改为经中间件链 + provider 链 → 验证: `make lint` / `make test` / `make build` 全绿 + 手动回归（播放/切歌/代理场景）

### Phase 2 步骤

1. 实现 Context 容器 + 服务按名注册/解析/覆写（`map[string]any` + 类型断言）→ 验证: 单元测试覆盖注册/解析/覆写/缺失
2. 实现作用域生命周期：插件 start/stop/dispose 递归清理 → 验证: 单元测试覆盖嵌套作用域清理
3. 实现事件链：listener / middleware / parallel / serial 四类 → 验证: 单元测试覆盖四类派发顺序与错误传播
4. （可选独立 PR）`netease.go` 核心协调器组件逐步接入框架层 → 验证: 行为不变，门禁全绿

## Resolved assumptions (autonomous defaults)

> 本 spec 由 `om-auto-write-spec` 以 autonomous 模式撰写。brief（`.ai/specs/briefs/2026-08-22-plugin-framework-playback.md`）的 Resolved unknowns 表已预答核心问题（插件开放给谁、动机、第一刀范围、框架深度、行为约束、验收示范等），以下仅列出 brief 遗留的开放点及本次采用的保守默认值：

| # | 问题 | 采用的默认值 | 理由 | 确认? |
|---|------|-------------|------|-------|
| Q1 | 中间件签名具体形态（函数式 vs 结构体式） | 标准函数式中间件签名：`URLMiddleware func(ctx, *player.URLMusic, next) error`（草案见 API Contracts），作用于 `getPlayInfo` 之后、`p.Play` 之前 | 最可逆、Go 生态惯例、与有序 func 切片要求一致 | ok |
| Q2 | UNM 胶水搬迁深度（Phase 1 只包胶水 vs 搬引擎） | Phase 1 只包胶水：引擎留在 vendored SDK 内，中间件仅复刻三层交互约束做调用编排 | 最小改动面、vendored SDK 只读边界、行为保真最易验证；原型结论为"中间件可行"时采用 | ok |
| Q3 | 引擎注册 provider 化（NewPlayerFromConfig switch）是否纳入 Phase 1 | 纳入 Phase 1 可选范围：接口先行，实际 switch 重构视实现难度决定是否后置到 Phase 2 | 避免 Phase 1 面过大；接口设计不堵死该方向 | ok |
| Q4 | 等价性测试是否引入自动化测试框架 | Phase 1 先以样本集 URL 快照对比 + 现有测试门禁执行，不新增测试框架依赖 | 最小新增依赖；自动化框架可后置 | ok |
| Q5 | Phase 2 框架包名/目录位置 | 新增独立包（如 `internal/framework`），与既有 `internal/ui` 平级 | 边界清晰、可独立回滚 | ok |

> 无 ⚠ NEEDS HUMAN CONFIRMATION 项。所有默认值均为保守可逆选择，可在评审时通过评论或直接编辑本 spec 覆盖。
