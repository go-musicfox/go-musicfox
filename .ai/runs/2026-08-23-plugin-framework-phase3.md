# Execution plan: plugin-framework-phase3

- Date: 2026-08-23
- Goal: 按 spec 实现 go-musicfox 插件化 Phase 3——核心协调器薄壳化（netease.go 业务能力全部服务化）+ 菜单/页面统一 provider 注册表（key → 参数化工厂），行为不变。先 Phase 3.0 证伪原型定工厂形态，再全量迁移。
- Scope: `internal/ui/netease.go`（薄壳化）、`internal/ui/event_handler.go`（70 处引用）、`internal/ui/operate.go`（硬编码跳转）、`internal/ui/menu_*.go`（43 菜单）、页面（login/search/lastfm/QR）、`internal/framework`（服务容器复用）、新增 `internal/ui/services.go` / `internal/ui/registry.go`、`cmd/musicfox.go`（组装）。
- Source doc: `.ai/specs/2026-08-23-plugin-framework-phase3.md`（spec PR #647，design-only，本实现 PR 引用之）
- Base branch: `feat/plugin-framework-playback`（#646 实现 PR 分支，未合并；internal/framework 仅在此分支，P0 闸门前 Phase 3 开发基于此）
- Non-goals: 运行时动态加载、WebUI/GUI 前端抽象、把 tea 消息循环改造成 framework 事件、外部 DI 框架、修改 vendored SDK、一次性大 PR（合并时按批次拆 PR）。
- Risks: 启动序重排（InitHook 280 行副作用）、运行时类型断言面（ServiceOf bool 不得丢弃）、43 菜单 15 种异构参数、无 E2E（导航冒烟测试兜底）、#635-638 修复 PR 重叠、#646 未合并（P0）。

## P0 前置闸门

- [ ] P0: `#646` 完成 review 并合入 master（`#645` spec 同步收口）；3.x 批 PR 在 `#646` 合入后从本分支按批次切出。开发先行，但任何 3.x PR 不得在 `#646` 合入前打开 — 验证: #646/#645 合入记录

## Implementation Plan

### Phase 3.0: 证伪原型（spec 硬性前置，临时分支）

- [ ] 3.0.1 双形态最小注册表原型：4 个异构菜单（PlaylistDetail(playlistId)、ArtistDetail(artistId, name)、SearchResult(searchType)、AddToUserPlaylist(userId, song, action)）+ 1 无参菜单（Ranks）+ 1 页面（login），分别实现候选形态 A（变参 Build(base, args ...any) (Menu, error)）与 B（泛型 Register[T](key, factory)）两套最小注册表，各记录迁移行数/调用侧类型安全/可读性；注明 MainMenu 引导路径 — 验证: 两套原型可编译；跳转与现状等价
- [ ] 3.0.2 按证据裁决 A/B（判定：迁移 ≤N 行、零类型断言、跳转点可读），记录决策点 — 验证: 决策点文档化；形态未定不得进入 3.2

### Phase 3.1: 服务化

- [ ] 3.1.1 framework 生命周期小切片：shareSvc、lastfm 经 Scope/Plugin 真实接入 — 验证: 接入单测 + 启动/退出无泄漏
- [ ] 3.1.2 服务名常量 + 注册点：services.go 集中定义，启动注册 — 验证: 注册表完整性单测
- [ ] 3.1.3 InitHook 拆分：顺序约束逐条枚举（cookie → 用户 → playmode → 音量 → 播放列表 → extInfo → likelist → 签到 → 检查更新 → 自动播放 → changelog），跨服务依赖显式（loginService jar 先于 userService 回调）— 验证: 启动序冒烟测试（游客 + 登录两路径）
- [ ] 3.1.4 baseMenu 访问器：baseMenu.netease → 类型安全访问器 — 验证: 全部菜单编译 + 现有测试绿
- [ ] 3.1.5 服务化迁移：业务能力注册为服务，引用点经 ServiceOf 解析（不丢弃 bool）— 验证: make lint/test/build 全绿 + 手动回归

### Phase 3.2: provider 机制

- [ ] 3.2.1 按原型结论实现 MenuProvider/PageProvider 接口 + 注册表 + init() 注册点 + 完整性断言 — 验证: 注册表单测（注册/构建/缺失/参数错误）
- [ ] 3.2.2 跳转 API（BuildMenu/BuildPage）+ 示范迁移（2-3 菜单 + 1 页面）— 验证: 示范跳转行为等价
- [ ] 3.2.3 导航冒烟测试骨架：menu→menu 全链路 + 登录回调时序 — 验证: 冒烟测试绿

### Phase 3.3: 全量迁移

- [ ] 3.3.1 菜单迁移（分批 8-10 个）：硬编码跳转 → BuildMenu — 验证: 每批后冒烟 + lint/test 绿
- [ ] 3.3.2 页面迁移：ToXxxPage → BuildPage — 验证: 页面跳转冒烟
- [ ] 3.3.3 深耦合点手术：event_handler/operate + player_controller/player_gapless/cur_playlist/status_bar/lastfm*/qr_login_client/toast/theme_persistence — 验证: 全部导航测试绿
- [ ] 3.3.4 旧路径清理：旧构造函数移除 — 验证: lint/test/build 绿 + 手动回归清单全过
- [ ] 3.3.5 薄壳收尾：Netease 字段收敛 — 验证: 全量门禁 + 手动回归
- [ ] 3.3.6 文档维护：AGENTS.md（UI 协调器/开发指南/路径表）+ 本 runs 文档进度 — 验证: 文档与代码同步

### Phase 3.4: 接口层预留（可选）

- [ ] 3.4.1 对外插件边界文档 + 示例插件（编译期注册示范）— 验证: 示例编译 + 文档评审

## Progress

> Convention: `- [ ]` pending, `- [x]` done. Append ` — <commit sha>` when a step lands. Do not rename step titles.

### P0 前置闸门

- [ ] P0: #646/#645 review 并合入 master

### Phase 3.0: 证伪原型

- [ ] 3.0.1 双形态最小注册表原型
- [ ] 3.0.2 裁决 A/B 记录决策点

### Phase 3.1: 服务化

- [ ] 3.1.1 framework 生命周期小切片
- [ ] 3.1.2 服务名常量 + 注册点
- [ ] 3.1.3 InitHook 拆分
- [ ] 3.1.4 baseMenu 访问器
- [ ] 3.1.5 服务化迁移

### Phase 3.2: provider 机制

- [ ] 3.2.1 接口 + 注册表 + 注册点
- [ ] 3.2.2 跳转 API + 示范迁移
- [ ] 3.2.3 导航冒烟测试骨架

### Phase 3.3: 全量迁移

- [ ] 3.3.1 菜单迁移（分批）
- [ ] 3.3.2 页面迁移
- [ ] 3.3.3 深耦合点手术
- [ ] 3.3.4 旧路径清理
- [ ] 3.3.5 薄壳收尾
- [ ] 3.3.6 文档维护

### Phase 3.4: 接口层预留（可选）

- [ ] 3.4.1 对外插件边界文档 + 示例
