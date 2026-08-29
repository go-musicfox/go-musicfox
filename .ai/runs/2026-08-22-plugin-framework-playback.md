# Execution plan: plugin-framework-playback

- Date: 2026-08-22
- Goal: 按 spec 实现 go-musicfox 插件化——Phase 0 证伪原型 → Phase 1 播放链路插件化（URL 中间件链 + source provider）→ Phase 2 核心框架层（cordis 语义，Go 手写零依赖）。行为约束：重构不改变任何用户可见行为。
- Scope: `internal/ui/player.go`（播放链）、`internal/track`（三源解析抽象）、`utils/netease`（UNM 胶水）、`cmd/musicfox.go`（UNM 配置注入）、`internal/player/player.go`（可选：引擎注册）、新增 `internal/framework`（Phase 2 框架包）。
- Source doc: `.ai/specs/2026-08-22-plugin-framework-playback.md`（spec PR #645，design-only，本实现 PR 引用之）
- Non-goals: 菜单/页面 provider、WebUI/GUI/TUI 前端抽象、外部插件加载机制、运行时动态加载、引入外部 DI 框架、修改 vendored SDK 行为、netease.go 核心协调器一次性迁移（Phase 2 步骤 4 标注为可选独立 PR，不在本实现 PR 内）。
- Risks: UNM 三层交互约束（SkipUNM/ProxyURL 强制关/banned-path 拦截）顺序稍错即改变播放行为；等价性靠样本集 URL 对比 + 门禁保障。

## Implementation Plan

### Phase 0: 证伪原型

1. 编写最小 UNM 中间件原型：对已知被封锁歌曲直接调用 vendored `github.com/cnsilvan/UnblockNeteaseMusic/processor` 的 `RequestBefore`/`Request`，产出 URL，与今天 SDK 路径（`utils/netease/songinfo.go` FetchPlayableInfo，`SkipUNM: true`）产出对比 → 验证: 原型可编译运行；对比结论记录在原型文档/注释中
2. 记录原型结论（可行/不可行 + 理由）到 plan/spec 决策点 → 验证: 决策点文档化；不可行则停止并回炉讨论（决策点不得跳过）

### Phase 1: 播放链路插件化

1. 定义 `URLMiddleware` 类型与中间件链容器（有序 func 切片，支持追加）→ 验证: 单元测试覆盖空链/单链/多链
2. 将现有 UNM 逻辑抽为 UNM 中间件插件（cmd/musicfox.go 配置拷贝、songinfo.go SkipUNM 主路径、ProxyURL 改写、player.go HasBannedPathSuffix 拦截，复刻三层交互约束），注册进链 → 验证: 等价性测试样本集 URL 对比一致（单元测试）
3. 定义 `PlayableSourceProvider` 接口，将 `track.Manager.resolveSongSource`（downloaded/cached/remote 三源）抽象为网易云 provider 实现，保留 Fetcher/option 注入 → 验证: 三源优先级行为与现状一致（单元测试）
4. 播放链接入：`internal/ui/player.go` 播放链改为经中间件链 + provider 链 → 验证: `make lint` / `make test` / `make build` 全绿 + 手动回归要点记录（播放/切歌/代理场景）

### Phase 2: 核心框架层

1. 实现 `internal/framework` 包：Context 容器 + 服务按名注册/解析/覆写（`map[string]any` + 类型断言）→ 验证: 单元测试覆盖注册/解析/覆写/缺失
2. 实现作用域生命周期：插件 start/stop/dispose 递归清理 → 验证: 单元测试覆盖嵌套作用域清理
3. 实现事件链：listener / middleware / parallel / serial 四类 → 验证: 单元测试覆盖四类派发顺序与错误传播

## Progress

PR: #646

> Convention: `- [ ]` pending, `- [x]` done. Append ` — <commit sha>` when a step lands. Do not rename step titles.

### Phase 0: 证伪原型

- [x] 0.1 编写最小 UNM 中间件原型（processor 直调 vs SDK 路径 URL 对比） — 3cf49b32
- [x] 0.2 记录原型结论（可行/不可行决策点） — 3cf49b32

### Phase 1: 播放链路插件化

- [x] 1.1 定义 URLMiddleware 类型与中间件链容器 — 685648c2
- [x] 1.2 UNM 中间件插件（三层交互约束复刻） — ad52348d
- [x] 1.3 PlayableSourceProvider 接口 + 网易云 provider 实现 — 64b89420
- [x] 1.4 播放链接入 + 全量门禁 — 64b89420

### Phase 2: 核心框架层

- [x] 2.1 Context + 服务注册/解析/覆写 — 48746903
- [x] 2.2 作用域生命周期（start/stop/dispose 递归） — 48746903
- [x] 2.3 事件链（listener/middleware/parallel/serial） — 48746903
