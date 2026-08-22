# go-musicfox 插件化 Phase 3：核心协调器薄壳化与菜单/页面 Provider 机制

## 📝 TLDR

netease.go 是 734 行的枢纽型协调器，持有 player、lyricService、trackManager、桌面歌词、全部 renderer 等业务组件；43 个菜单都嵌入持有 `*Netease` 引用的 baseMenu，菜单跳转全是硬编码构造函数调用。Phase 3 的目标是把 Netease 退化为薄壳（导航/组装/事件分派），全部业务能力注册为 framework 服务按名解析，并引入统一的菜单/页面 provider 注册机制（key → 参数化工厂），使跳转不再硬编码。行为约束：重构不改变任何用户可见行为。第一刀编译期注册，接口层预留未来外部插件边界。核心机制（参数化工厂）形态先经 Phase 3.0 证伪原型确定。

## 📝 Problem Statement

`internal/ui/netease.go`（734 行、19 个方法）是枢纽型 god-object：`Netease` struct 直接持有 user、lastfm、login、search、lyricService、六个 renderer、player、shareSvc、trackManager、desktopLyrics 等全部业务组件；`InitHook`（约 280 行）串起 cookie → 用户 → 播放模式 → 音量 → 播放列表 → 签到 → 自动播放等有序副作用；`Update` 统一分派全部事件。耦合的下游形态：`event_handler.go` 70 处 netease 引用、`operate.go`（1013 行）硬编码全部菜单跳转构造函数、43 个 `menu_*.go` 全部嵌入持有 `*Netease` 引用的 `baseMenu`、8 个页面（login/search/lastfm/QR 等）同样持有 netease 引用。每次加功能都要动 Netease 或其周边（近 3 个月实例：桌面歌词 149 行、主题切换 71 行、toast 45 行均落在 netease.go 上）。

**成本锚点——三个真实改动场景（现状 vs 目标态）**：

| 场景 | 现状改动点 | 目标态改动点 |
|------|-----------|-------------|
| A. 新增一个菜单（如"新歌速递"） | 新建 menu_xxx.go 嵌入 baseMenu（引用 n）→ 在父菜单 `SubMenu()` 或 operate.go 动作表硬编码 `NewXxxMenu(newBaseMenu(n), params...)` → 父菜单与子菜单构造器强耦合；多处入口（如搜索结果跳歌手）需在每处硬编码跳转，共动 2-3 个文件 | 实现 `MenuProvider`（key + 构建函数）并注册 → 跳转处统一按 key 查注册表构建；新菜单只动新文件 + 一行注册，父菜单零改动 |
| B. 新增一个页面入口（如新的第三方授权页） | Netease 加 `ToXxxPage()` 方法（19 方法之一）→ 页面持有 n 引用 → event_handler/菜单调用点硬编码 → 动 3+ 文件且 Netease 膨胀 | 页面实现 PageProvider 注册 → 导航统一经注册表；Netease 不再新增导航方法 |
| C. 替换播放源解析逻辑 | trackManager 由 Netease 内部实例化，组件经 `n.trackManager` 访问；替换实现需改 Netease 构造（Phase 1 已抽象 PlayableSourceProvider，但解析入口仍在 Netease 手中） | trackManager 注册为 named service，任意组件按名解析；替换 provider 注册即可，调用方零改动 |

动机（用户确认）：内部解耦为主线（"当前代码设计实现耦合严重"），第三方插件生态为远期愿景。当前 Phase 2（internal/framework 零依赖 cordis 风格框架）已落地但生命周期（Scope/Plugin）零生产使用。

## 📝 Proposed Solution

方向 B（薄壳 + 统一 provider 机制）经用户确认，分解为四个可独立交付阶段（见 Phasing）。核心机制——**参数化工厂**——形态未定，由 Phase 3.0 证伪原型在候选集中裁决（见 Q1 与 Resolved assumptions）。外部研究（koishi/cordis 服务模式、Go 注册表实践、VS Code contribution points）结果整合进本节与 Architecture。

**外部研究结论（lib-1，2026-08-23）**：业界没有主流 Go 注册表混合"泛型 + 异构参数"——两个可行参考是 **kopia 式**（每个 provider 一个 options 类型 `T`，`Register[T any](key, factory func(T) (Provider, error))`，调用方构造具体 T，JSON round-trip 解码异构选项）与 **go-git 式**（类型化 key 绑定工厂输出类型 + 闭包捕获参数，`func() T` 固定形状）。VS Code `contributes.menus` 贡献点**没有 args 字段**（参数来自上下文推断或命令注册时闭包捕获）；k9s 插件参数是 string 模板；koishi 配置是 schema 驱动的单一对象。**对 Phase 3.0 原型的启示**：候选形态收敛为三种——A. 变参工厂 `Build(base, args ...any) (Menu, error)`（最贴近现有构造函数，零类型安全）；B. kopia/go-git 式泛型注册（每个菜单一个 options 类型或闭包工厂，注册侧类型安全，调用侧构造参数对象/变参透传）；C. `map[string]any` + 工厂内断言（最不推荐，challenger 已指出类型安全损失）。原型在 A/B 之间裁决，C 仅作对照。

## 📝 Architecture

**服务容器**：复用 `internal/framework.Context`（Provide/Override/ServiceOf，cordis 语义退化实现，Phase 2 已落地并有单测）。服务名常量集中定义（见 API Contracts）。`Netease` 薄壳持有 Context，各业务能力注册为服务。

**Provider 注册表**：菜单/页面统一注册表，key 为菜单 `GetMenuKey()` 字符串；注册项为参数化构建函数；构建结果返回 `Menu`/`Page`。注册表自身作为 framework 服务暴露。

**薄壳 Netease**：保留 foxful-cli 要求的 `model.App` 集成、页面导航状态（login/search 页持有）、`Update` 事件分派与 renderer 组合；所有业务能力改为经 Context 解析。**注意**：薄壳即使对已服务化的 coverRenderer 也保留组合引用（`Components()` 的渲染排序依赖它），服务化只改跨组件访问路径，不改薄壳内部的组合与排序逻辑。

**baseMenu 访问器**：`baseMenu` 不再持有 `*Netease`，改为持有访问器（或直接持 Context），菜单代码从 `e.netease.player` 改为访问器/服务解析——访问器提供类型安全的前向兼容（详见 API Contracts 的迁移策略）。

## 📝 Data Model

无新增持久化模型。`framework.Context` 内部为 `map[string]any` 服务表（已有，`context.go`），服务名常量收敛为集中定义（见 API Contracts）。BoltDB 存储层（用户/播放状态/快照）不变，仅访问入口从 Netease 字段转移至服务。`appCookieJar` 包级全局的归属收敛为：**注册为 `loginService` 的服务状态**，三个读写方（login_page.go、login_webview_page.go、qr 登录路径）均经服务解析；SDK 侧 `neteaseutil.SetGlobalCookieJar` 全局同步义务由 loginService 承担（jar 变更即同步），防止三处各自维护。

## 📝 API Contracts

**服务命名表**（Phase 3.1 落地，常量集中定义于 `internal/ui/services.go`）：

| 服务名 | 类型 | 说明 |
|--------|------|------|
| `player` | `*Player` | 播放器（含播放列表状态） |
| `lyricService` | `*lyric.Service` | 歌词服务 |
| `trackManager` | `*track.Manager` | 三源解析 + PlayableSourceProvider |
| `desktopLyrics` | `desktop_lyrics.Controller` | 桌面歌词 |
| `coverRenderer` | `*CoverRenderer` | 封面渲染（被多页面跨组件引用） |
| `userService` | 用户状态服务（包装 `*structs.User` + 登录回调） | user 状态迁移 |
| `loginService` | 登录流程（cookie jar 归属） | 含 appCookieJar 收敛 |
| `shareSvc` | `*composer.ShareService` | 分享服务 |
| `lastfm` | `*lastfm.Client` | Last.fm 客户端 |
| `menuRegistry` / `pageRegistry` | provider 注册表（见下） | Phase 3.2 起 |

内部 UI 部件（lyricRenderer/songInfoRenderer/progressRenderer/spectrumRenderer/spectrogramRenderer）留在薄壳组合中，不服务化（Resolved assumptions Q3）。

**Provider 接口**（草案形态，**以 Phase 3.0 原型裁决结论为准**；研究参考见 Proposed Solution）：

```go
// 候选形态 A：变参工厂（最贴近现有构造函数，零类型安全）
type MenuProvider interface {
    Key() string
    Build(base baseMenu, args ...any) (Menu, error)
}
// 候选形态 B（kopia/go-git 式）：泛型注册，注册侧类型安全
//   Register[T](key string, factory func(base baseMenu, opts T) (Menu, error))
//   —— 每个异构菜单一个 options 类型 T；调用侧构造 T 后按 key 构建
// 候选形态 C（对照）：map[string]any 参数 + 工厂内类型断言（最不推荐）
// 页面 provider 同理（返回 model.Page）
```

**注册表 API**（草案）：`Register(provider MenuProvider)`（编译期调用，`init()` 注册 + 启动时注册表完整性断言）、`BuildMenu(key string, args ...any) (Menu, error)`、`BuildPage(key string, args ...any) (model.Page, error)`。跳转失败按现有 toast 错误路径降级，不 panic。

**兼容面**：现有菜单构造函数（`NewXxxMenu(base, params...)`）在迁移期间保留（新路径就绪前旧路径可用）；迁移完成后移除。`Menu`/`Page`/`baseMenu` 对外签名不变（`baseMenu` 内部引用由 `*Netease` 改为访问器）。

## 📝 UI/UX

**行为约束（硬性）**：重构不改变任何用户可见行为——布局、快捷键、右键菜单、播放/切歌/代理行为、启动序全部保持现状。菜单列表顺序与内容不变（迁移前后菜单渲染结果逐屏对比）。不把 tea 消息循环改造成 framework 事件链（TUI Update 在主 goroutine，事件分派保持现状，framework 事件链仅用于服务间解耦，不得改变 Update 的并发模型）。

## 📝 Edge Cases & Failure Scenarios

- **服务缺失 / 类型断言失败**：`ServiceOf[T]` 返回 `(T, bool)`，迁移期间禁止丢弃 bool——所有解析点必须处理缺失（记录错误 + 降级路径），禁止裸 panic。与今天的编译期保证（`baseMenu.netease.player` 非 nil）相比这是**运行时风险的引入面**，靠 Phase 3.1 的注册表完整性单测 + 导航冒烟测试兜底。
- **注册冲突**：`Provide` 对重复名字 panic（framework 既有语义）。注册为每文件 `init()`（Q4）；重复注册由 `Provide` 的 panic 语义在启动时暴露；启动时注册表完整性断言锁定注册清单（key 全集 + 类型正确性），冲突/遗漏在启动即炸。
- **菜单参数错误**：已定为 `(Menu, error)` 显式返回，跳转处 toast 报错降级（与现有菜单出错行为一致），不 panic。
- **启动序重排**：InitHook 的顺序约束（cookie → 用户 → playmode → 音量 → 播放列表 → extInfo → likelist → 签到 → 检查更新 → 自动播放 → changelog）必须在拆分前逐条枚举并写成测试（启动序冒烟），任何服务不得提前消费未就绪依赖。
- **桌面歌词/播放器在登录前启动**：游客模式启动路径（无 cookie）不得因服务化引入 nil 服务访问。
- **框架生命周期零生产使用**：Scope/Plugin Start/Stop/Dispose 首次真实接入——Phase 3.1 以小切片验证（1-2 个服务先接入），暴露问题在铺开前解决。

## 📝 Risks & Impact Review

- **回归风险（高）**：迁移爆炸半径是启动序 + 43 菜单导航 + 事件分发 + 全部页面，现有 86 个测试全是渲染/popup/i18n 类，无导航测试。缓解：新增导航冒烟测试（注册表上 menu→menu 全链路转换 + 登录回调时序）+ 逐项手动回归清单（播放/切歌/代理/右键/主题切换/桌面歌词/扫码登录）作为每 Phase 交付门槛。
- **rebase 冲突面**：4 个未合并修复 PR（#635 event_handler、#637 player 引擎、#636 songinfo/UNM）与本次改动面重叠；同一分支分批提交 + 合并拆 PR，落地时按批次 rebase 并优先合入修复 PR 减小冲突。
- **评审面**：用户已确认同分支分批提交、合并拆 PR（否决一次性大 PR），每批次对应一个独立评审单元。
- **回滚**：每 Phase 独立可回滚；服务化与 provider 机制均保持"新增为主、迁移为辅"（旧路径保留至新路径冒烟通过），单 Phase 内失败可 revert 该批次提交。
- **行为漂移**：代理（ProxyURL）与 UNM 三层交互（SkipUNM/banned-path）是 Phase 1 已锁定的行为面，服务化不得改动其判定顺序。

## 📋 Phasing

每个 Phase 独立可交付/可回滚，对应独立合并批次（同一分支分批提交，合并时按批次拆 PR）：

- **Phase 3.0（证伪原型，spec 硬性前置）**：参数化工厂形态原型——在临时分支实现双形态最小注册表（覆盖 4 个异构菜单 + 1 个无参菜单 + 1 个页面），跑通跳转并对比行为等价性，结论记录为本 spec 的决策点。形态未定不得进入 3.2。
- **Phase 3.1（服务化）**：framework 生命周期小切片真实接入（先 1-2 个服务验证 Scope/Plugin Start/Stop/Dispose）→ 业务能力全量注册为 named services → InitHook 按顺序约束拆分到各服务 → baseMenu 与页面改为经服务表访问。
- **Phase 3.2（provider 机制）**：菜单/页面统一 provider 注册表（key → 参数化工厂）+ 跳转 API + 编译期注册点，机制带最小示范（2-3 个菜单 + 1 页面迁移）。
- **Phase 3.3（全量迁移）**：43 菜单 + 8 页面 + event_handler/operate 定点手术，薄壳收尾（Netease 只留导航/组装/事件分派），导航冒烟测试全绿。
- **Phase 3.4（接口层预留，可选）**：对外插件边界的接口文档化与示例（编译期注册示范插件），不改机制。

## 📋 Implementation Plan

> 每 Phase 独立可交付/可回滚；同一分支（#646）分批提交，合并时按批次拆独立 PR。步骤标题不得改名。

### Phase 3.0 前置闸门（P0）

- [ ] P0: `#646` 完成 review 并合入 master（`#645` spec 同步收口）；此后 Phase 3 各批 PR 从新的 Phase 3 分支按批次切出，且优先合入 #635/#637 修复 PR 再 rebase 迁移面。Phase 3 分支上的开发可在 #646 未合时先行开始，但**任何 3.x 批次的 PR 不得在 #646 合入前打开** → 验证: #646/#645 合入记录；3.1 批 PR 基于新 master

### Phase 3.0: 证伪原型（spec 硬性前置，临时分支）

- [ ] 3.0.1 双形态最小注册表原型：选 4 个异构菜单（`PlaylistDetail(playlistId int64)`、`ArtistDetail(artistId int64, name string)`、`SearchResult(searchType SearchType)`、`AddToUserPlaylist(userId int64, song structs.Song, action bool)`）+ 1 个无参菜单（`Ranks`）+ 1 页面（login），**分别实现候选形态 A（变参 `Build(base, args ...any) (Menu, error)`）与候选形态 B（kopia/go-git 式泛型 `Register[T](key, factory)`）两套最小注册表**，各记录：迁移改动行数、调用侧类型安全（类型断言 vs 泛型）、跳转代码可读性；注明 MainMenu 引导路径（`NewMainMenu(netease)`）的构建方式 → 验证: 两套原型均可编译；跳转路径与现状等价（行为对比）
- [ ] 3.0.2 按证据裁决：若 A 在 4 个异构 + 1 无参菜单上类型安全/可维护性可接受（判定标准：每个菜单迁移 ≤N 行、零类型断言、跳转点可读），选 A；否则选 B；记录裁决结论与理由到本 spec 决策点 → 验证: 决策点文档化；形态未定不得进入 3.2

### Phase 3.1: 服务化

- [ ] 3.1.1 framework 生命周期小切片：1-2 个服务（shareSvc、lastfm）经 Scope/Plugin 真实接入，验证 Start/Stop/Dispose → 验证: 接入单测 + 启动/退出无泄漏（framework 首次生产使用）
- [ ] 3.1.2 服务名常量 + 注册点：`services.go` 集中定义，启动注册 → 验证: 注册表完整性单测（清单与预期一致，无重复/缺失）
- [ ] 3.1.3 InitHook 拆分：逐条枚举顺序约束（cookie → 用户 → playmode → 音量 → 播放列表状态 → extInfo → likelist → 签到 → 检查更新 → 自动播放 → changelog）→ 各服务按需自初始化，薄壳只做编排；**显式覆盖跨服务依赖**：`LoginCallback`（userService）内部调用 `appCookieJar.Save()`——loginService 的 jar 必须先注册/初始化，再允许 userService 回调 → 验证: 启动序冒烟测试（游客模式 + 登录模式两条路径，顺序断言）
- [ ] 3.1.4 baseMenu 访问器：`baseMenu.netease` → 类型安全访问器 → 验证: 全部菜单编译通过 + 现有测试绿
- [ ] 3.1.5 服务化迁移：业务能力注册为服务，引用点经 `ServiceOf` 解析（**不丢弃 bool**）→ 验证: `make lint` / `make test` / `make build` 全绿 + 手动回归清单

### Phase 3.2: provider 机制

- [ ] 3.2.1 按原型结论实现 `MenuProvider`/`PageProvider` 接口 + 注册表 + `init()` 注册点 + 完整性断言 → 验证: 注册表单测（注册/构建/缺失/参数错误）
- [ ] 3.2.2 跳转 API（`BuildMenu`/`BuildPage`）+ 示范迁移（2-3 菜单 + 1 页面）→ 验证: 示范跳转行为等价
- [ ] 3.2.3 导航冒烟测试骨架：注册表上 menu→menu 全链路转换 + 登录回调时序 → 验证: 冒烟测试绿（后续迁移的回归网）

### Phase 3.3: 全量迁移

- [ ] 3.3.1 菜单迁移（分批，每批 8-10 个）：硬编码跳转 → `BuildMenu` → 验证: 每批后冒烟测试 + lint/test 绿
- [ ] 3.3.2 页面迁移：`ToXxxPage` → `BuildPage` → 验证: 页面跳转冒烟
- [ ] 3.3.3 深耦合点定点手术：event_handler.go（70 处引用）、operate.go（1013 行操作表）及其余引用 netease 的非菜单文件——`player_controller.go`、`player_gapless.go`、`cur_playlist.go`、`status_bar.go`、`lastfm.go`/`lastfm_profile.go`、`qr_login_client.go`、`toast.go`、`theme_persistence.go` → 验证: 全部导航测试绿
- [ ] 3.3.4 旧路径清理：迁移完成的旧构造函数移除 → 验证: lint/test/build 绿 + 手动回归清单全过
- [ ] 3.3.5 薄壳收尾：Netease 字段收敛（只留 App/导航状态/Update/renderer 组合）→ 验证: 全量门禁 + 手动回归清单
- [ ] 3.3.6 文档维护（AGENTS.md 硬规则）：更新 UI 协调器章节（Netease 薄壳化描述）、开发指南「添加新菜单/新页面」（注册 MenuProvider/PageProvider 替代"注册到导航系统"）、核心文件路径表（services.go/registry.go）、运行文档 `.ai/runs/2026-08-23-plugin-framework-phase3.md`（进度追踪，P0 时创建）→ 验证: 文档与代码同步（架构变更=高优先级）

### Phase 3.4: 接口层预留（可选）

- [ ] 3.4.1 对外插件边界文档 + 示例插件（编译期注册示范）→ 验证: 示例编译通过 + 文档评审

**手动回归清单**（每 Phase 交付门槛）：播放/切歌/暂停/音量与播放模式切换（列表循环/随机/心动）；代理模式与 UNM 行为；右键菜单（含 generic 动作）；鼠标单击/双击/滚轮（event_handler 迁移后）；全局快捷键；远程控制（MPRIS/Now Playing/System Media）；主题切换（含持久化）；桌面歌词开/关与拖动；Kitty 封面清理路径（`BeforeEnterMenuHook` → `ClearDisplayed`）；扫码/网页登录与 Last.fm 授权；断网/接口错误降级路径；changelog 弹窗与检查更新；搜索与各菜单逐项导航；游客模式启动。

## Resolved assumptions (autonomous defaults)

| # | Question | Applied default | Why |
|---|----------|-----------------|-----|
| Q1 | 参数化工厂候选形态与原型判据 | 候选形态 A（变参 `Build(base, args ...any) (Menu, error)`）与 B（kopia/go-git 式泛型注册 `Register[T](key, factory)`）为原型主裁决对象，C（map[string]any）仅对照；原型必须覆盖 4 个异构菜单 + 1 页面，裁决结论写入 spec 决策点后才进入 3.2 | 研究（lib-1）确认业界无"泛型 + 异构参数"混合先例，A/B 覆盖两个可行方向；原型是硬前置 gate，不是可跳过的假设 |
| Q2 | 服务化 + 菜单 provider + 页面 provider + 迁移是一个 spec 还是 bundle | 保持单 spec | 三 capability 共享服务容器 + baseMenu 签名互锁（3.1 决定访问器形状，3.2 工厂签名依赖它）+ 单一行为保真验证故事；独立 spec 会分裂机制设计。**3.1 先行理由**：先收窄 baseMenu 依赖面，注册表工厂签名才稳定（否则 3.2 建在旧 baseMenu 上，3.1 换访问器后工厂要返工）。用户已确认同分支分批 + 合并拆 PR，Phasing 提供独立交付边界 |
| Q3 | 服务化清单边界 | 核心业务 + 跨组件共享组件（player/lyricService/trackManager/desktopLyrics/coverRenderer/user/login/search/shareSvc/lastfm/cookieJar）服务化；内部 UI 部件（lyric/songInfo/progress/spectrum/spectrogram renderer）留薄壳 | 用户选的"全部业务能力"指跨组件业务能力；renderer 是薄壳 UI 组合内部件，服务化零收益且增加 Context 噪音 |
| Q4 | 注册时机与位置 | 每菜单/页面文件 `init()` 注册 + 启动时注册表完整性断言 | 符合"编译期注册（import + 注册）"；新文件自带注册行，对未来外部插件（import 即注册）是标准心智；完整性断言兜底遗漏 |
| Q5 | 注册表包位置 | `internal/ui/registry.go`，注册表作为 framework 服务暴露，类型定义在 ui 包 | Menu/Page 属 ui 包；framework 保持通用不依赖 ui 类型 |

> Q1-Q5 已由对话裁决；Phase 3.0 原型结论将作为决策点追加到本 spec。
