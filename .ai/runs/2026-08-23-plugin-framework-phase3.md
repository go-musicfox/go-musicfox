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

- [x] 3.3.1 菜单迁移（分批 8-10 个）：硬编码跳转 → BuildMenu — 验证: 每批后冒烟 + lint/test 绿 — 21fbb303 / fd38c384 / ad9b3ea4 / 7b45fc6a
- [x] 3.3.2 页面迁移：ToXxxPage → BuildPage — 验证: 页面跳转冒烟
- [x] 3.3.3 深耦合点手术：event_handler/operate + player_controller/player_gapless/cur_playlist/status_bar/lastfm*/qr_login_client/toast/theme_persistence — 验证: 全部导航测试绿
- [x] 3.3.4 旧路径清理：旧构造函数移除 — 验证: lint/test/build 绿 + 手动回归清单全过
- [x] 3.3.5 薄壳收尾：Netease 字段收敛 — 验证: 全量门禁 + 手动回归
- [x] 3.3.6 文档维护：AGENTS.md（UI 协调器/开发指南/路径表）+ 本 runs 文档进度 — 验证: 文档与代码同步

### Phase 3.4: 接口层预留（可选）

- [x] 3.4.1 对外插件边界文档 + 示例插件（编译期注册示范）— 验证: 示例编译 + 文档评审

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

- [x] 3.3.1 菜单迁移（分批 8-10 个）：硬编码跳转 → BuildMenu — 验证: 每批后冒烟 + lint/test 绿 — 21fbb303 / fd38c384 / ad9b3ea4 / 7b45fc6a（32 菜单注册 + 全部内部导航调用点迁移；main_menu/local_search 因构造器签名保留 bootstrap 直构，注册仅为完整性断言覆盖）
- [x] 3.3.2 页面迁移：ToXxxPage → BuildPage — 验证: 页面跳转冒烟 — 5f18d89a（注册 search / lastfm_custom_api 页面 provider；ToLoginPage 经 buildPageOrToast 每次构建新实例并接线 AfterLogin（登录页无跨组件状态，shell 不再持有 login 单例）；search 页保持 shell 单例（wordsInput/result/searchType 与 SearchResultMenu/operate.searchSong 共享），在 NewNetease 经 BuildPage 构建；lastfm_profile 自定义 API 页入口走注册表；新增页面构建/导航冒烟测试）
- [x] 3.3.3 深耦合点手术：event_handler/operate + player_controller/player_gapless/cur_playlist/status_bar/lastfm*/qr_login_client/toast/theme_persistence — 验证: 全部导航测试绿 — b72871f1 / efb0aefc（menuServices 增加薄壳访问器 App/Main/MustMain/Rerender/Search/SaveActiveTheme/NotifyThemeSwitch/PlaybarHoveredElement + newBaseMenuFromSvc；event_handler 全部 h.netease.* 走 h.svc，cur_playlist 菜单经注册表构建；operate 全部 helper 经 svc 解析、n.* 裸访问清零，action_menu 菜单经注册表构建；last_fm 注册为无参 provider，MainMenu 用 mustBuildNoArg；Player 增 svc 访问器字段，player_gapless/player_controller 经 p.svc；lastfm 菜单 RefreshMenuList 走 m.svc.MustMain；status_bar/qr_login_client 已无 *Netease 耦合无需改动）
- [x] 3.3.4 旧路径清理 — 2455bbba（删除原型文件 registry_proto_a.go/registry_proto_b.go（证据文档 .ai/runs/proto-3.0-evidence.md 保留）；构造函数清点：所有 menu/page 构造函数均被 registry 工厂包裹或由 bootstrap（NewMainMenu/NewLocalSearchMenu 直构）或测试引用，唯一死构造函数 NewCompositeRenderer 删除——composite_renderer.go 其余类型/方法为迁移前已存在的死代码，按「只删构造函数」保留并标记）
- [x] 3.3.5 薄壳收尾 — 0dc7de18（见下方「3.3.5 薄壳最终形态」）
- [x] 3.3.6 文档维护 — 见提交（AGENTS.md UI 协调器/开发指南（添加新菜单/新页面）/核心文件路径表 + 本 runs 进度）

### Phase 3.4: 接口层预留（可选）

- [x] 3.4.1 对外插件边界文档 + 示例

### Phase 3.5: framework 加固（#646 评审后续）

#646 评审登记的 Phase 3 后续项（framework 生命周期/并发契约/unmproto 标注）在本阶段落地：

- [x] 3.5.1 `Scope.Start` 失败回滚 — 372eab1b（任何插件/子作用域 Start 失败时，按逆序 Stop 已启动的插件与子作用域后再返回错误；测试：三插件中第二个失败 → 第一个收到 Stop 且错误返回、Deps 失败回滚、子作用域失败回滚兄弟与父插件、嵌套作用域失败回滚父作用域）
- [x] 3.5.2 parallel 事件链并发契约文档 — eabe4b6f（parallel handler 各自 goroutine 共享 `*Context` 只读；Service/ServiceOf 并发解析安全，Provide/Override 并发写为 race 且不支持；契约落在 `Parallel` 文档注释、parallel emit 分支与包文档；新增 `TestParallelReadOnlyServiceResolutionIsRaceSafe`（32 个 parallel handler 并发解析，-race 下通过）；未加锁（按 AGENTS.md 不做投机加固））
- [x] 3.5.3 unmproto 标注 — 0b7c549c（header 说明为 Phase 0 证伪原型常驻 harness，引用 `.ai/runs/2026-08-22-plugin-framework-playback.md` 决策记录；保留用于复现 UNM 中间件等价性检查；如不再需要可在后续清理中移除；顺带修复该文件被 whole-file lint 门禁暴露的既有 ST1003 `linuxApiData` → `linuxAPIData`）

验证：`make lint` 0 issues · `make test` 绿（framework 全量含 -race）· `make build` 绿 · `gofmt -l` 干净。macdriver cocoa/mediaplayer 的 2 项网络依赖测试（TestNSImage/TestMPMediaItemArtwork）为既有环境性偶发（master 同样失败，与本阶段无关）。

### Phase 3.6: 解耦调查 + 插件边界（BaseMenu 导出）

解耦调查 work package 1（访问器导航 + Search）+ 插件边界的 BaseMenu 导出，两组逻辑单元：

- [x] 3.6.1 访问器导航 + Search — bde4a9ac（`menuServices` 新增 `ToLoginPage`/`ToSearchPage` 转发；登录门控菜单 25 处 + player.go 2 处 `ToLoginPage` 改走 `svc`；`menu_search_result` 的 `svc.Netease().search.*` 改经 `svc.Search()`；`.Netease()` 逃生口 internal/ui 43 → 11（剩余为 menu.go 6 / action_select 2 / lastfm_profile 2 / registry_registrations 1，均为收 `*Netease` 的旧辅助函数）；顺带修复 whole-file lint 暴露的既有债务：menu_add_to_user_playlist/menu_album_detail/menu_playlist_detail 的 ST1003（userId/albumId/playlistId → userID/albumID/playlistID，含 operate/player/registry_test 引用点）与 menu_accessor QF1008（`s.n.App.Rerender` → `s.n.Rerender`））
- [x] 3.6.2 BaseMenu 导出 — 72cf880a（`type BaseMenu struct{ model.DefaultMenu; svc *menuServices }` + `type baseMenu = BaseMenu` 别名，全部既有代码与注册零改动（嵌入式别名字段名保持 `baseMenu`，40+ 处 `m.baseMenu` 引用不受影响）；BaseMenu 新增 17 个导出转发方法（服务解析/薄壳/导航，nil 安全）；外部包编译校验 `internal/ui/plugin_boundary_external_test.go`（`package ui_test`，嵌入 `ui.BaseMenu`、注册闭包 `func(base ui.BaseMenu, _ ui.NoArgMenuOpts)` 编译通过、BuildMenu 零值基座构建 + 转发方法零值降级）；docs/plugin_development.md 更新 BaseMenu 基座/转发方法/外部形态示例）

验证：`make lint` 0 issues · `make test` 绿 · `make build` 绿 · 改动文件 `gofmt -l` 干净（`foxful_integration_test.go`/`song_info_renderer.go` 为 HEAD 既有的非 gofmt 债务，非本阶段引入，未触碰）。

### 3.3.2/3.3.3 决策与合理残留（`.netease.` 引用清单）

- **页面导航形态（3.3.2）**：login 页在 ToLoginPage 每次经 `buildPageOrToast("login")` 构建新实例（登录页无跨组件状态；webviewAvailable 检测缓存在 NewLoginPage 实例内，回调/返回语义不变），shell 移除 `n.login` 字段；search 页保留 shell 单例（`n.search`），其 wordsInput/result/searchType 被 SearchResultMenu.BeforeEnterMenuHook 与 operate.searchSong 共享，必须保持同一实例。
- **深耦合迁移形态（3.3.3）**：`NewCurPlaylist`/`NewActionMenu`/`NewLastfm` 均注册为菜单 provider（`cur_playlist`/`action_menu`/`last_fm`），构造点改走 `buildMenuOrToast`/`mustBuildNoArg`；operate/event_handler 仍以 `*Netease` 为载体参数（承载 newMenuServices 构造），内部全部经 `svc` 访问器解析，不再出现裸 `n.*`/`h.netease.*` 访问。menu_main/local_search bootstrap 直构保留（3.3.5 裁决）。
- **合理残留 `.netease.` 引用**（`grep -rn "\.netease\." internal/ui/` 其余匹配均属此类，非深耦合）：
  - **renderer 组合**（lyric/cover/songInfo/progress/composite_renderer）：spec Q3 决议 renderer 留薄壳组合，持有 shell 引用属组合职责；
  - **页面**（login/search/lastfm_auth/lastfm_custom_api/login_qr/lastfm_qr/login_webview）：页面由 shell 构建并持有 shell 引用（`l.netease.MustMain()`/`RerenderCmd` 等），随 3.3.5 字段收敛评估；
  - **player.go**：Player 为 shell 自有组件，保留 `p.netease` 组合引用（player_gapless/player_controller 已改走 `p.svc`）。
- **status_bar.go / qr_login_client.go**：已无 `*Netease` 耦合（纯访问器/包内工具），无需改动。
- **toast.go / theme_persistence.go**：`registerToastHook`/`saveActiveTheme`/`notifyThemeSwitch` 为 shell 自有薄壳方法，消费方（event_handler/menu.go）已改经访问器 `svc.SaveActiveTheme`/`svc.NotifyThemeSwitch` 调用。

### 3.3.5 薄壳最终形态（0dc7de18）

- **Netease 字段清单**（全部保留，职责已收敛）：
  - model.App 集成：`*model.App`
  - 导航状态：`search *SearchPage`（单例，wordsInput/result/searchType 跨组件共享）
  - renderer 组合：`lyricRenderer`/`songInfoRenderer`/`progressRenderer`/`coverRenderer`/`spectrumRenderer`/`spectrogramRenderer`
  - 服务实例持有（registerServices 的来源）：`player`/`lyricService`/`trackManager`/`desktopLyrics`/`coverRenderer`/`shareSvc`/`lastfm`/`user`
  - framework：`ctx *framework.Context`/`scope *framework.Scope`
  - 小 UI 状态：`playbarHoveredElement`、`themeNotifID`/`themeNotifTimer`
- **`.netease.` 引用收敛**：全部 102 处剩余引用均为 shell 方法调用（`MustMain`/`RerenderCmd`/`Tick`/`WindowWidth`/`EffectiveWindowHeight`/`SpectrumLines`/`GetCoverWidth`/`GetCoverEndColumn`/`GetLyricPosition`/`Player()`/`DesktopLyrics()`/`GetDesktopLyricsLines()`/`ToLoginPage`/`ToSearchPage`/`playbarHoveredElement`/`Ctx()`），来自 renderer 组合（5 文件）与 shell 构建的页面（7 文件）及 player.go 的导航/桌面歌词方法；**不再有任何 `X.netease.<服务实例字段>` 直连**（服务实例仅经访问器/Context 解析）。player.go 的 `p.netease.player.Mode()`/`trackManager`/`user` 与 lastfm 三个授权页的 `svc.Lastfm()` 迁移为本阶段重点。
- **lint 债务**：仓库 lint 配置为 `issues.new: true` + `whole-files: true`（改动文件全量检查），本阶段触碰 lastfm 页面文件导致其历史债务暴露——已一并修复：ST1003 初始ism 重命名（`LastfmCustomAPIPage` 族/`reloadAPIAccount`/`getAuthURLWithToken`）、`util.SetFgStyle` → `style.FG`、ui 包文档注释（随 3.3.4 原型文件删除而丢失，ST1000）、composite_renderer.go gci import 排序、registry.go ST1021 组注释；`registry_test.go` 的 TestRegisterAndBuildLastfmCustomAPIPage 因 lastfm 改经访问器解析而需注册服务到 Context。
- **bootstrap 直构保留**：`NewMainMenu`/`NewLocalSearchMenu`（internal/commands 入口，经 `neteaseFromBase` 兼容注册完整性断言）不变。
- **composite_renderer.go**：整个文件为迁移前已存在的死代码（无任何引用），仅删除死构造函数 `NewCompositeRenderer`，其余按 AGENTS.md「不删除预存在死代码」保留（建议后续单独清理）。

### 手动回归清单（3.3.4/3.3.5 门禁后必须重检）

服务访问路径已全部收敛到注册表/访问器，以下用户可见行为必须手动回归确认行为不变：

- **播放控制**：播放/暂停、上一首/下一首、进度 seek、音量（滚轮/快捷键）
- **右键菜单**：任意菜单右键 → 播放控制分组 / 选中项分组（收藏/下载/喜欢/相似/网页打开/分享）
- **主题切换**：快捷键与右键切换、toast 通知、`theme.activeTheme` 持久化（重启后生效）
- **桌面歌词**：开关切换、拖动窗口
- **搜索**：进入搜索、搜索结果加载更多（BottomOut）
- **登录/Last.fm**：登录流程、Last.fm 授权/主页/队列、Last.fm API account 页（key/secret 保存/重载/清空）
- **喜欢/收藏**：点赞/取消、收藏歌单/专辑/歌手、添加至歌单
- **远程控制**：macOS Now Playing（含喜欢/取消喜欢）、Linux MPRIS
- **鼠标事件**：单击/双击/滚轮/hover 状态栏元素

## Phase 3.0 裁决（决策点，2026-08-23）

**结论：候选形态 B（泛型注册 `RegisterMenuB[T](key, factory)`）胜出**，原型证据见 `.ai/runs/proto-3.0-evidence.md`（commit 1b32ef71）：

- 迁移成本持平：A = B = 15 调用点 +54 行，调用点类型断言均为 0（唯一断言在 login bootstrap，A/B 相同）
- B 优势：参数契约进类型系统（`ArtistDetailOpts{ArtistID, Name}` 消除 `(int64, string)` 歧义）；注册闭包零变参断言（唯一断言藏在 registry 内部，`.(menuFactory[T])`）；opts 结构体成为显式插件契约（未来第三方插件边界，与 kopia 研究结论一致）
- B 代价（可缓解）：无参菜单需 `NoArgMenuOpts{}`（20+/43 菜单无参，用 `mustBuildNoArg` 紧凑辅助形式解决）；单字段 struct 轻微噪音

**3.2 按形态 B 实现**；原型文件 `registry_proto_a.go` / `registry_proto_b.go` 于 3.3.4 移除（证据文档 `.ai/runs/proto-3.0-evidence.md` 保留）。

## 节奏变更（用户决策，2026-08-23）

- 用户选择：**在当前分支 `feat/plugin-framework-phase3` 持续开发，功能稳定后再合入 master**（不按 P0 先合并 #646 再拆 PR 的路径）。
- P0 闸门对 PR 打开的限制相应**顺延**：3.x PR 暂不开，等用户判定功能稳定后统一处理合并（届时按批次拆 PR 或整支合入，由用户决定）。
- #645/#646 的合并动作由用户自行掌握（#646 QA 门禁待处理）。
- 本 runs 文档与分支提交持续作为开发基线。
