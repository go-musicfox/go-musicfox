# 为 go-musicfox 设计插件化方案（cordis 式核心框架 + 播放链路插件化，三阶段）

- Date: 2026-08-22
- Category: refactor
- Priority signal: high — 双驱动动机（核心协调器难改 + 第三方插件生态诉求），但体量大、风险高，必须分阶段
- Risk signal: high — 重构播放链路并引入新框架层；UNM 行为保真涉及 vendored SDK 边界与三层交互约束
- Routing: Next: om-auto-write-spec "为 go-musicfox 设计插件化方案（cordis 式核心框架 + 播放链路插件化，三阶段）— brief: .ai/specs/briefs/2026-08-22-plugin-framework-playback.md"

## Problem

go-musicfox 的核心协调器 `internal/ui/netease.go` 将 login、search、player、lyricService、trackManager 等全部挂载其上，加功能要动太多地方，改动成本高、风险大；同时维护者希望未来能形成第三方插件生态。两个动机均被用户明确确认。当前"加某个功能要动 N 处、花 M 天"的具体成本尚未量化，spec 开头必须列出 2-3 个真实改动场景（如新增一个音源 provider、新增一个播放前 hook），逐个对比现状改动点与目标态改动点，作为成本锚点。

## Agreed direction

参考 `~/Desktop/cordis`（koishi 的元框架：Context 容器、服务按名注册/解析/可覆写、作用域隔离、事件链、fiber 树生命周期）设计 go-musicfox 的插件能力。**先内部解耦为主线，同时预留外部插件边界**。第一刀范围 = 核心框架最小集 + 播放链路插件化（菜单/页面 provider、WebUI/GUI 前端抽象、外部加载机制均后置）。

一份 spec，内部明确三阶段，每阶段独立 PR、独立可交付/可回滚：

- **Phase 0**：30 行证伪原型——写一个最小中间件，对一首已知被封锁的歌曲直接调用 vendored SDK 的 UnblockNeteaseMusic processor，与今天 SDK 路径产出的 URL 对比。对得上 → 中间件形态可行，进入 Phase 1；对不上 → 中间件形态选错，回炉讨论（此决策点不得跳过）。
- **Phase 1**：播放链路插件化——URL 中间件链（UNM 注入点）+ playable source provider 接口落地。**不依赖核心框架**，先用有序 func 切片 + 现有接口落地。验收双示范：UNM 中间件插件（现有 UNM 逻辑抽成插件，行为一致）+ source provider 接口（`ResolvePlayableSource` 背后抽象化，网易云音源成为第一个 provider 实现）。
- **Phase 2**：核心框架层——cordis 语义较完整移植，Go 手写**零外部依赖**：Context + 服务按名注册/解析/可覆写 + 作用域生命周期（插件 start/stop/dispose 递归清理）+ 事件链（listener/middleware/parallel/serial）。spec 中必须显式记录对 uber fx / dig 等现成 DI 框架的否决理由（Go 无法移植 cordis 的 Proxy/原型链机制，"ctx 属性访问解析服务"退化为 `map[string]any` + 类型断言，即自研轻量 DI 容器——接受自维护、零生态成本）。

行为约束：**重构不改变任何用户可见行为**（UNM 作为插件后表现与今天完全一致），靠 `make lint` / `make test` / `make build` 门禁 + 手动回归保障。插件第一刀采用编译期注册（import + 注册），不做运行时动态加载；外部边界仅在接口层预留（未来可加进程/脚本适配器）。

## Resolved unknowns

| Question | Answer (from the conversation) |
|----------|--------------------------------|
| 插件开放给谁 | 先内部解耦为主线，同时预留外部插件边界（用户选 C） |
| 动机 | 双驱动：netease.go 核心协调器难改 + 想要第三方插件生态 |
| 第一刀范围 | 核心框架最小集 + 播放链路插件化（用户从 A 菜单 provider / B 播放链路 / C 全量中选 B） |
| 典型插件场景 | 新增完整菜单/页面、替换音频播放源、菜单项 provider、WebUI/GUI/TUI、注入 UNM 插件（后三者后置） |
| 框架深度 | cordis 语义较完整移植（Context/注册表/覆写/作用域生命周期/事件链），Go 手写，不引外部 DI 框架（用户选 B） |
| 行为约束 | 行为完全不变（用户选 1A） |
| 验收示范 | UNM 中间件插件 + source provider 接口双示范（用户选 2B） |
| UNM 中间件范围 | 先跑 30 行证伪原型再定（challenger CRITICAL 1 的解决方案；原型结论决定 Phase 1 只包胶水还是搬引擎） |
| 路由形态 | 一份 spec 内部三阶段，每阶段独立 PR（challenger CRITICAL 2 的解决方案） |
| 框架路线定价 | 维持手写零依赖，spec 中记录对 fx/dig 的显式否决理由（challenger CRITICAL 3 的解决方案） |
| UNM 现状（challenger 勘误） | UNM 引擎本体在 vendored SDK 包 `github.com/go-musicfox/netease-music/util`（request.go:219 UNMFlag、config.go:13-20 全局变量、config.go:40-42 ProxyURL 非空强制关 UNM）；go-musicfox 自有侧仅四块薄胶水：cmd/musicfox.go:73-78 配置拷贝、utils/netease/songinfo.go:36 主路径 SkipUNM: true、songinfo.go:77-81 ProxyURL 前缀改写、internal/ui/player.go:275-277 HasBannedPathSuffix 拦截 |
| 播放链路现状 | Player.PlaySong → getPlayInfo → trackManager.ResolvePlayableSource(ctx, song) → UNM skip 检查 → p.Play(player.URLMusic{URL, Song, Type})（internal/ui/player.go:271-306、538-546）；Player 引擎已接口化但 NewPlayerFromConfig 是硬编码 switch |

## Non-goals

- 菜单/页面 provider 机制（后置切片）
- WebUI/GUI/TUI 前端抽象（后置切片；当前仅保证接口设计不堵死该方向）
- 外部插件加载机制（进程/脚本适配器；第一刀仅编译期注册 + 接口层预留）
- 运行时动态加载插件（go plugin 动态库，跨平台问题大）
- 引入外部 DI 框架（uber fx / dig 等；spec 中记录否决理由）
- 修改 vendored SDK 行为（Phase 0 证伪原型仅只读调用验证；除非原型证伪后用户重新决策）

## Affected areas (if known)

- `internal/ui/player.go` — 播放链（271-306、538-546）：中间件注入点位于 ResolvePlayableSource 之后、Play 之前；UNM skip 检查（275-277）
- `internal/track` — manager.go（downloaded/cached/remote 三源逻辑 288-332）、fetcher.go（已有 Fetcher 接口，Manager 支持 WithFetcher 注入）
- `cmd/musicfox.go` — 73-78 UNM 配置注入胶水
- `utils/netease/songinfo.go` — 36（SkipUNM: true 主路径）、77-81（ProxyURL 改写）
- `internal/player/player.go` — NewPlayerFromConfig 硬编码 switch（引擎注册 provider 化的候选点）
- `vendor/github.com/go-musicfox/netease-music/util` — UNM 引擎所在，只读边界/最小触碰

## 约束与风险提示（给 spec 作者）

- UNM 三层交互约束必须逐条复刻：V1 主路径 SkipUNM、ProxyURL 非空强制关引擎、banned-path 拦截——顺序和条件稍错即改变现有用户播放行为
- "行为保真"需要等价性测试标准（歌曲样本、网络条件、回退路径判定），不能只是口号
- netease.go 难改的成本需量化（spec 开头列真实改动场景对比）
