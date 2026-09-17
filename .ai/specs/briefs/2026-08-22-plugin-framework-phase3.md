# netease.go 薄壳化 + 菜单/页面 provider 化（插件化 Phase 3）

- Date: 2026-08-23
- Category: refactor
- Priority signal: high — 双驱动动机（netease.go 核心协调器难改 + 第三方插件生态愿景），是既定插件化战略（#645 spec / #646 实现）的下一阶段，用户明确要继续
- Risk signal: high — 大规模重构：netease.go 枢纽迁移 + 43 个菜单 + 事件分发 + 全部页面；无 E2E 门禁，"行为不变"难以机器验证
- Routing: om-auto-write-spec "netease.go 薄壳化 + 菜单/页面 provider 化（Phase 3，先原型后 spec，同分支分批开发） — brief: .ai/specs/briefs/2026-08-22-plugin-framework-phase3.md"

## Problem

`internal/ui/netease.go`（734 行、19 个方法）是枢纽型 god-object：`Netease` struct 持有 login、search、player、lyricService、trackManager、desktopLyrics、coverRenderer、user 状态等全部业务组件；`InitHook`（280 行）串起 cookie → 用户 → 播放模式 → 音量 → 播放列表 → 签到 → 自动播放等有序副作用；`Update` 统一分派全部事件。43 个 `menu_*.go` 都嵌入 `baseMenu` 持有 `*Netease` 引用，菜单跳转全是硬编码构造函数调用（`operate.go` 1013 行）。`event_handler.go` 有 70 处 netease 引用。最近 3 个月落在 netease.go 的特性提交（桌面歌词 149 行、主题 71 行、toast 45 行）每次都要动它——"加一个功能要动 N 处"成立，但成本从未量化。用户动机：内部解耦（主线，当前耦合严重），未来第三方插件生态（远期愿景）。

## Agreed direction

- **方向 B（薄壳 + provider）**：Netease 退化为薄壳（导航/组装/事件分派），全部业务能力（player、lyricService、trackManager、desktopLyrics、coverRenderer、user/login 状态等）注册为 framework 服务、按名解析，组件间不再互相持有引用。
- **菜单 + 页面统一 provider 机制**：菜单（43 个）与 Page 类入口（登录/搜索/Last.fm 授权/二维码等）共用同一套注册机制（key → 参数化工厂），跳转不再硬编码构造函数。已明确否决"只做菜单 provider"与"两套机制分开"。
- **先原型后 spec**：参数化工厂是方向 B 的核心机制但形态未定义（43 菜单约 15 种异构参数形态：纯 base、playlistId、artistId+name、SearchType、structs.Song、bool、组合等）。**Phase 3.0 硬性前置**：先写最小注册表证伪原型——3-5 个异构菜单（PlaylistDetail(playlistId)、ArtistDetail(artistId, name)、SearchResult(searchType)、AddToUserPlaylist(userId, song, action)）+ 1 个页面（login），在临时分支跑通注册表原型、记录结论，再定 spec 的工厂形态。复制 Phase 0 的证伪原型模式，决策点不得跳过。
- **开发节奏**：在现有 #646 分支上继续开发（用户已选，不先合并 #646），但**提交按 Phase 3 内部子项分批**（如服务化 / 菜单 provider 机制 / 页面 provider / 菜单迁移等批次），**合并时按批次拆独立 PR**——已批准治理"每阶段独立 PR、独立可交付/可回滚"仍生效，否决"一次性大 PR"交付形态。
- **文档形态**：独立新 spec（本 brief 对应的 plugin-framework-phase3 spec），引用 #645 为前置文档，不追加修改未合并的 #645。
- 行为约束（沿用既有决定）：重构不改变任何用户可见行为；编译期注册（import + 注册），不做运行时动态加载；外部插件边界仅在接口层预留；零依赖手写框架不引入外部 DI 框架。
- **成本锚点**：spec 开头必须列出 2-3 个真实改动场景（如新增一个菜单、新增一个页面入口、改播放源解析），逐个对比现状改动点与目标态改动点（上轮 brief 就要求、至今未兑现）。

## Resolved unknowns

| Question | Answer (from the conversation) |
|----------|--------------------------------|
| 动机 | 内部解耦（主线，"当前代码设计实现耦合严重"），第三方插件生态为远期愿景 |
| 解耦终态 | B：薄壳 + 菜单/页面 provider 统一机制（非仅服务化） |
| 服务化范围 | 全部业务能力服务化，无明确例外（player/lyricService/trackManager/desktopLyrics/coverRenderer/user 等） |
| 页面范围 | 菜单 + Page 类入口（登录/搜索/Last.fm 授权/二维码）同一套 provider 机制 |
| 开发节奏 | 在 #646 分支继续开发，提交按子项分批，合并时拆独立 PR（否决一次性大 PR） |
| 核心机制 | 参数化工厂（key → factory with args）；先 Phase 3.0 证伪原型定形态，spec 不得留白 |
| 文档形态 | 独立新 spec，引用 #645 为前置（否决追加到未合并的 #645） |
| 框架验证 | framework 生命周期（Scope/Plugin Start/Stop/Dispose）零生产使用，spec 须含一小块真实接入切片先验证框架本身 |
| 事件模型 | 不把 tea 消息循环改造成 framework 事件（TUI Update 在主 goroutine，避免并发回归面） |
| 全局状态 | appCookieJar 等包级全局变量（登录/网页登录/二维码登录共享）的服务化归属由 spec 明确 |
| 行为不变验证 | make lint/test/build 全绿 + 新增导航冒烟测试（注册表上 menu→menu 全链路 + 登录回调时序）+ 逐项手动回归清单；现有 86 个测试全是渲染/popup/i18n 类，无导航测试 |
| 耦合面 | 43 个菜单仅约 15 个直接摸 `.netease.`（浅而广）；event_handler/operate/InitHook 是深耦合点（深而窄），spec 应利用此不对称 |

## Non-goals

- 运行时动态加载插件（go plugin 动态库）
- WebUI/GUI/TUI 前端抽象（接口层不堵死该方向即可）
- 把 tea 消息循环改造成 framework 事件链
- 修改 vendored SDK 行为
- 引入外部 DI 框架（uber fx/dig 等）
- "一次性大 PR"交付形态（已否决，改同分支分批 + 合并拆 PR）
- 外部插件加载机制（进程/脚本适配器；仅编译期注册 + 接口层预留）

## Affected areas (if known)

- `internal/ui/netease.go` — 枢纽：InitHook（280 行有序副作用，顺序约束需逐条枚举：cookie 在用户前、用户前 likelist 前、自动播放等）、Update 事件分派、19 方法、Netease struct 字段
- `internal/ui/event_handler.go` — 70 处 netease 引用（深耦合点）
- `internal/ui/operate.go`（1013 行）— 硬编码菜单跳转构造函数
- `internal/ui/menu_*.go` — 43 个菜单，15 种异构构造参数形态
- `internal/ui/` 页面 — login_page、search、lastfm_auth、qr login 等 Page 类入口
- `internal/framework` — 现有 Context/服务注册/Scope 生命周期/事件链（Phase 3 的构建块，尚未生产使用）
- `internal/ui/player.go`、`internal/track` — 服务化对象（player/lyricService/trackManager）
- `cmd/musicfox.go` — 启动组装与编译期注册点

## 约束与风险提示（给 spec 作者）

- 参数化工厂证伪原型（Phase 3.0）必须先行并记录结论，决策点不得跳过；原型选 3-5 个异构菜单 + 1 页面覆盖主要参数形态
- InitHook 280 行有序副作用（cookie → 用户 → playmode → 音量 → 播放列表状态 → extInfo → likelist → 签到 → 检查更新 → 自动播放 → changelog）必须逐条枚举顺序约束，迁移不得静默重排启动序
- 成本锚点：spec 开头列 2-3 个真实改动场景现状 vs 目标态对比，论证全量服务化优于最小痛点方案
- 验证：新增导航冒烟测试（menu→menu 全链路 + 登录回调时序）+ 逐项手动回归清单；"行为不变"不能只是口号
- 事件模型：不把 tea 消息循环改造成 framework 事件
- framework 生命周期先小切片真实接入验证，再全量铺开
- 菜单迁移是"浅而广"（多数只要 player/user/ToLoginPage/search 薄访问器），深耦合点（event_handler/operate/InitHook）做定点手术
- 注意 #635-638 等未合并修复 PR 与本次改动面重叠（event_handler/player 引擎/songinfo），分批提交时留意 rebase 冲突面
