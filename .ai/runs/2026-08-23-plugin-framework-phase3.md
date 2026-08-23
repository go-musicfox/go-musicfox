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

- [x] 3.1.1 framework 生命周期小切片：shareSvc、lastfm 经 Scope/Plugin 真实接入 — 验证: 接入单测 + 启动/退出无泄漏 — 82eb8322
- [x] 3.1.2 服务名常量 + 注册点：services.go 集中定义，启动注册 — 验证: 注册表完整性单测 — 9b9eb953
- [x] 3.1.3 InitHook 拆分：顺序约束逐条枚举（cookie → 用户 → playmode → 音量 → 播放列表 → extInfo → likelist → 签到 → 检查更新 → 自动播放 → changelog），跨服务依赖显式（loginService jar 先于 userService 回调）— 验证: 启动序冒烟测试（游客 + 登录两路径）— e1c91719
- [x] 3.1.4 baseMenu 访问器：baseMenu.netease → 类型安全访问器 — 验证: 全部菜单编译 + 现有测试绿 — e7919c86
- [x] 3.1.5 服务化迁移：业务能力注册为服务，引用点经 ServiceOf 解析（不丢弃 bool）— 验证: make lint/test/build 全绿 + 手动回归 — d567a2c2 / 883f1a3e / c0d6d9ad

### Phase 3.2: provider 机制

- [x] 3.2.1 按原型结论实现 MenuProvider/PageProvider 接口 + 注册表 + init() 注册点 + 完整性断言 — 验证: 注册表单测（注册/构建/缺失/参数错误）— c59a60f8
- [x] 3.2.2 跳转 API（BuildMenu/BuildPage）+ 示范迁移（2-3 菜单 + 1 页面）— 验证: 示范跳转行为等价 — d8714c36
- [x] 3.2.3 导航冒烟测试骨架：menu→menu 全链路 + 登录回调时序 — 验证: 冒烟测试绿 — 797f2de3

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

- [x] 3.0.1 双形态最小注册表原型 — evidence: `.ai/runs/proto-3.0-evidence.md` (both shapes compile + tests green; call sites wired to B)
- [ ] 3.0.2 裁决 A/B 记录决策点

### Phase 3.1: 服务化

- [x] 3.1.1 framework 生命周期小切片 — 82eb8322
- [x] 3.1.2 服务名常量 + 注册点
- [x] 3.1.3 InitHook 拆分 — e1c91719（cookie-jar 生命周期 → LoginService.InitJar；用户恢复 + cookie 登录流 → UserService.LoadFromStorage/LoginWithCookie；InitHook 保留 12 步顺序注释 + 序列测试）
- [x] 3.1.4 baseMenu 访问器 — e7919c86（menuServices 类型安全访问器；baseMenu 自带代码迁移；netease 字段保留至 3.1.5）
- [x] 3.1.5 服务化迁移 — d567a2c2 / 883f1a3e / c0d6d9ad（43 菜单 + lastfm/lastfm_profile/cur_playlist/action_select → svc 访问器；event_handler/operate/executor/player_controller → 服务解析；baseMenu.netease 移除，编译强制收尾）

### 手动回归要点（3.1.5 迁移后必须重检）

3.1.5 只改服务访问路径（`*.netease.<服务字段>` → `svc` 访问器 / ServiceOf 解析），不改任何逻辑。但以下用户可见行为全部经过服务解析链路，需手动回归确认行为不变：

- **播放控制**：播放/暂停、上一首/下一首、进度 seek、音量滚轮/快捷键（event_handler + operate 的 player 解析）
- **右键菜单**：任意菜单右键 → 播放控制分组 / 选中项分组（收藏/下载/喜欢/相似/网页打开/分享），确认分组完整、点击行为与以前一致（action_select + operate 闭包经 svc.Netease() 逃生口）
- **主题切换**：快捷键 OpSwitchTheme 与右键「切换主题」的 toast 通知、theme.activeTheme 持久化（Netease().saveActiveTheme/notifyThemeSwitch 逃生口）
- **桌面歌词开关**：菜单内开关、Player().DesktopLyrics() 逃生口路径
- **搜索**：进入搜索、搜索结果的加载更多（BottomOut）、搜索结果页 before/back hook 中 `netease.search` 逃生口
- **登录流程**：进入需登录菜单（云盘/每日推荐/歌单收藏等）→ ToLoginPage 逃生口；登录成功回调链路（Operation.NeedsAuth → UserService）；Last.fm 授权/主页/队列（svc.Lastfm()）
- **喜欢/收藏操作**：点赞/取消、收藏歌单/专辑/歌手、添加至歌单（operate 中 svc.User()/svc.Player()/svc.TrackManager()/svc.ShareSvc() 解析）
- **远程控制**：macOS Now Playing 的喜欢/取消喜欢（player_controller 的 svc.User() 实时槽读取）

若上述任一行为异常，优先怀疑对应服务的注册时序（`registerServices` 在 NewNetease 内、任何菜单渲染前执行）。

### Phase 3.2: provider 机制

- [x] 3.2.1 接口 + 注册表 + 注册点 — c59a60f8（production registry.go 按形态 B 实现；5 样本菜单 + login page 的 init() 注册；menuRegistry/pageRegistry 服务化；bootstrap 完整性断言；原型调用点全部转生产 API；proto_b 仅保留机制与注册作证据）
- [x] 3.2.2 跳转 API + 示范迁移 — d8714c36（buildMenuOrToast/buildPageOrToast；SearchResultMenu 五个 submenu 分支走注册表；album_detail/user_playlist/dj_radio_detail 新注册；Lastfm 授权页入口走 BuildPage）
- [x] 3.2.3 导航冒烟测试骨架 — 797f2de3（registry_test.go：注册表单测 + Ranks→PlaylistDetail、SearchType→SearchResult→demo 菜单链 + 完整性断言 + 服务解析；登录回调时序引用 user_service_test.go）

### Phase 3.3: 全量迁移

- [ ] 3.3.1 菜单迁移（分批）
- [ ] 3.3.2 页面迁移
- [ ] 3.3.3 深耦合点手术
- [ ] 3.3.4 旧路径清理
- [ ] 3.3.5 薄壳收尾
- [ ] 3.3.6 文档维护

### Phase 3.4: 接口层预留（可选）

- [ ] 3.4.1 对外插件边界文档 + 示例

## Phase 3.0 裁决（决策点，2026-08-23）

**结论：候选形态 B（泛型注册 `RegisterMenuB[T](key, factory)`）胜出**，原型证据见 `.ai/runs/proto-3.0-evidence.md`（commit 1b32ef71）：

- 迁移成本持平：A = B = 15 调用点 +54 行，调用点类型断言均为 0（唯一断言在 login bootstrap，A/B 相同）
- B 优势：参数契约进类型系统（`ArtistDetailOpts{ArtistID, Name}` 消除 `(int64, string)` 歧义）；注册闭包零变参断言（唯一断言藏在 registry 内部，`.(menuFactory[T])`）；opts 结构体成为显式插件契约（未来第三方插件边界，与 kopia 研究结论一致）
- B 代价（可缓解）：无参菜单需 `NoArgMenuOpts{}`（20+/43 菜单无参，用 `mustBuildNoArg` 紧凑辅助形式解决）；单字段 struct 轻微噪音

**3.2 按形态 B 实现**；原型文件 `registry_proto_a.go` / `registry_proto_b.go` 保留为证据（标记 PROTO），3.2 正式实现时由 B 演进，3.3.4 清理旧路径时移除。
