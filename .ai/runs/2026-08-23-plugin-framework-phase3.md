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

### Phase 3.7: framework lifecycle hardening (round 2)

framework 生命周期语义加固（#646 评审后续第二轮）：Scope 状态机 + Stop/Dispose 错误聚合 + 事件 handler panic 隔离。正确用法行为不变，误用语义显式化。

- [ ] 3.7.1 Scope 生命周期状态机 + Stop/Dispose 错误聚合 — 验证: 状态机/聚合单测 + make lint/test/build 绿
- [ ] 3.7.2 事件 handler panic 隔离 — 验证: 四类 handler panic 单测（-race）+ make lint/test/build 绿
- [ ] 3.7.3 测试覆盖 — 验证: 全部门禁绿

### Phase 3.8: decoupling round 2 — renderer interface refactor

5 个 renderer 目前持有 `netease *Netease`，仅用于 (a) 嵌入式 `*model.App` 方法、(b) 访问器已有成员、(c) 5 个 shell 持有的跨 renderer 几何方法。目标：renderer 依赖 `*menuServices`（访问器）+ 自有业务服务，不再持有 `*Netease`。仅访问路径变更，布局/渲染逻辑不动。

- [x] 3.8.1 访问器几何转发：menuServices 新增 EffectiveWindowHeight/SpectrumLines/GetCoverWidth/GetCoverEndColumn/GetLyricPosition — 验证: 编译 + 单测绿 — e2c369fc
- [x] 3.8.2 renderer 去壳：5 个 renderer 的 netease 字段 → svc，构造器签名同步；netease.go 组合接线与 renderer 测试更新 — 验证: grep renderer netease 为 0 + make lint/test/build 绿 — ae33e2f8
- [x] 3.8.3 lastfm 页面 opts 去壳：LastfmCustomAPIPageOpts/LastfmAuthPageOpts 携带 svc 访问器取代 *Netease；lastfm 三页构造器收 svc — 验证: svc.Netease() 降 2 + make lint/test/build 绿 — a4c1fb3a
- [x] 3.8.4 bootstrap 构造器收窄：NewMainMenu/NewLocalSearchMenu 改收 baseMenu；新增导出 NewBaseMenu 桥接 commands 入口；删除 neteaseFromBase — 验证: neteaseFromBase 为 0 + make lint/test/build 绿 — 579d3505
- [x] 3.8.5 player 桌面歌词走访问器：p.svc.DesktopLyrics()/GetDesktopLyricsLines()（menuServices 转发）— 验证: 编译 + 单测绿 — ae382fe5

### Phase 3.9: plugin ecosystem practicalization — main-menu items and startup hooks

两个机制让编译期插件生态真正可用：插件可以 (a) 声明自己的主菜单入口、(b) 注册启动任务；并把 check_update 插件改造为同时使用两者（移除其硬编码主菜单入口与 shell 级启动检查）。

- **Mechanism A（主菜单插件项，registry.go + menu_main.go）**：`MainMenuItem{Key, Title}` + `RegisterMainMenuItem(key, title)`（追加，重复 key panic）+ `MainMenuPluginItems()`（快照）。`NewMainMenu` 在全部内置项之后追加插件项：`{Title: item.Title}` 入 `menus`、`mustBuildNoArg(item.Key, base)` 入 `menuList`（插件主菜单项 MUST 是无参菜单；key 未注册时启动 panic 作为完整性信号，先显式断言再构建以给出清晰错误）。`mainMenuCheckUpdateIndex`（15）与其 Action 特判分支随 check_update 插件化移除；`mainMenuHelpIndex`（14）保留（内置项末尾，插件项只追加在其后、不会使其漂移）。触发等价性：以前选中「检查更新」经 Action 特判直接触发检查；现在 Action 落到默认分支（返回 nil/nil）→ SubMenu 进入插件菜单 → 插件菜单自身的 `BeforeEnterMenuHook`（进入即触发检查并弹回主页面，单次 Enter）承担检查与通知。
- **Mechanism B（启动钩子，registry.go + netease.go InitHook）**：`var startupHooks []func()` + `RegisterStartupHook(fn)`（追加，nil 拒绝）。`runStartupHooks()` 在 InitHook 启动序第 10 步调用（即原 shell 级「检查更新」自动检查位置：用户/登录恢复后、自动播放前，`errorx.Go` goroutine 内；此时 services 已注册、toast 已接线）。每个 hook 按注册序调用并包 recover——panic 经 slog/slogx 记日志、跳过该 hook，不阻断启动（framework panic 隔离 house style）。
- **check_update 改造（internal/plugins/checkupdate）**：init() 追加 `ui.RegisterMainMenuItem("check_update", "检查更新")` 与 `ui.RegisterStartupHook(startupCheck)`；启动自动检查逻辑（config-gated `config.Startup.CheckUpdate`）自 netease.go 移入新文件 startup.go。`CheckUpdateMenu` 新增 `BeforeEnterMenuHook`（后台 goroutine 检查 + `app.Notify` 弹 TUI 通知 + 返回 (false, main) 弹回主页面）；Action 保留原检查命令路径。netease.go 删除硬编码启动检查块。
- **文档**：docs/plugin_development.md 记录两个 API（边界表面表 + 规则 + 行为保持契约新增启动钩子不得 panic / 主菜单项 key 必须无参两条）+ 示例一按新形态更新（进入钩子 / 主菜单入口 / startup.go / 调用点）；AGENTS.md 同步 registry.go 表项与插件开发段落。
- **测试**：registry_test 新增注册校验（重复/空 key panic、快照拷贝隔离）、NewMainMenu 追加插件项（内置项后、帮助索引不漂移、未注册 key panic）、runStartupHooks panic 隔离（顺序保持 + 中间 hook panic 被跳过）；ui 集成测试改为 TestMainMenuPluginItemRoutesToPluginMenu（Action 落默认分支 + SubMenu 路由到插件菜单）；plugins 聚合器测试断言 MainMenuPluginItems 含 check_update；checkupdate 测试新增进入钩子安装与 startupCheck 配置门控。
- **3.9.5 Last.fm 提取为第二个真实插件**：`menu_accessor.go` 导出 `MenuServices = *menuServices` 别名 + `NewMenuServices(ctx)`；`BaseMenu.Services()` 取访问器；ui 导出页面布局助手（`PageTitleView`/`PageMenuTitleView[WithBack]`/`PageInput*`/`PageSubmit*`/`PageButton*`/`FinishCustomPageView`/`PageBreadcrumb*`/`PageBackButtonWidth`/`PageMenuTitleRow`/`SetPageInputCursor`）与 `TickLogin`/`ShowConfirmPopup`/`BuildPageOrToast`。`internal/plugins/lastfm`：`Lastfm` 菜单（嵌入 `ui.BaseMenu`，经 `m.Lastfm()` 服务访问）+ `LastfmProfile` + 三页（`NewLastfmAuthPage(svc ui.MenuServices)` 等，QR 页不注册、仅由 auth 页 `authByQRCode` 内部直构）+ `registry.go`（注册 `last_fm`/`lastfm_auth`/`lastfm_custom_api` + `RegisterMainMenuItem("last_fm","LastFM")`；页面 opts 类型随页面移入插件，字段 `Svc ui.MenuServices`）。ui 侧：删除 5 个 lastfm 文件 + 其测试；`registry_registrations.go` 删除三项注册；`menu_main.go` 删除内置 LastFM 入口（索引 13，帮助索引 14→13）——**主菜单位置变更：LastFM 从内置索引 13 变为插件项，排在全部内置项之后**；`expectedMenuKeys`/`expectedPageKeys` 移除 `last_fm`/`lastfm_auth`/`lastfm_custom_api`（插件 key 不参与内置完整性断言，与 check_update 一致）；`registry_test.go`/`custom_page_background_test.go` 移除已移动用例（lastfm 页背景渲染仍经共享 `FinishCustomPageView` 覆盖，插件侧不可构造 shell 故不重复）。测试移到 `internal/plugins/lastfm/`（菜单/页面经 registry 构建 + 服务经测试 context 解析）。聚合器 `plugins.go` 空导入 lastfm。— 验证: make lint 0 issues · make test 绿（avcore 为既有环境性偶发）· make build 绿 · gofmt 干净
- **3.9.6 DJ/电台集群提取为第三个真实插件（批量菜单提取示范）**：`internal/plugins/dj`——「主播电台」集群 10 个菜单整体移出（`menu_radio_dj_type` + 9 个 `menu_dj_*.go`），每菜单嵌入 `ui.BaseMenu`、构造器收 `ui.BaseMenu`。**key 逐一保持与内置注册相同**：`dj_radio_detail`/`dj_category_detail`/`dj_category`/`dj_program_rank`/`dj_program_hour_rank`/`dj_hot`/`dj_sub`/`dj_recommend`/`dj_today_recommend`/`radio_dj_type`——ui 侧跳入集群的调用点（`menu_search_result.go` 跳 `dj_radio_detail`）零改动。集群内跨菜单跳转改走新导出的 `ui.BuildMenuOrToast`/`ui.MustBuild`/`ui.MustBuildNoArg`（语义与内置 `buildMenuOrToast`/`mustBuild`/`mustBuildNoArg` 逐一等价；`BuildPageOrToast` 的菜单对应物）。仅集群内部使用的 opts 契约 `DjCategoryDetailOpts`/`DjHotOpts`（含 `DjHotType`/`DjHot`/`DjNotHot`）随菜单移入插件；被 ui 共享的 `DjRadioDetailOpts` 留在 ui。`radio_dj_type` 声明主菜单入口 `RegisterMainMenuItem("radio_dj_type", "主播电台")`。ui 侧：删除 10 个文件；`registry_registrations.go` 删除 `dj_radio_detail` 与 batch 1 注册；`expectedMenuKeys` 移除 10 个 dj key；`menu_main.go` 删除内置「主播电台」入口（索引 12，帮助索引 13→12）——**主菜单位置变更：主播电台从内置索引 12 变为插件项，排在全部内置项之后**；`event_handler.go` 的 `OpToggleSortOrder` 类型断言 `*DjRadioDetailMenu` → 新导出接口 `ui.DjRadioDetailSortable`（`ToggleSortOrder`/`Reload`，ui 不得反向导入插件包）；`registry_test.go` 移除 dj checkSub 用例（SearchResultMenu→dj_radio_detail 边）。测试移到 `internal/plugins/dj/`（10 key 经 `ui.BuildMenu` 构建 + key 断言 + 集群内跳转边 + 入口菜单 8 子项 + 主菜单项注册断言）；聚合器 `plugins.go` 空导入 dj + `plugins_test.go` 新增 `TestBlankImportRegistersDjPlugin`。插件内 ST1003 修复（whole-file lint）：`djRadioId`/`categoryId` → `djRadioID`/`categoryID`（含 `DjRadioId()` 访问器 → `DjRadioID()`，无外部调用方）。— 验证: make lint 0 issues · make test 绿（23 包 ok）· make build 绿 · gofmt 干净
- **3.9.7 专辑集群提取为第四个真实插件（批量菜单提取 + 主菜单项 + 共享 opts 判定示范）**：`internal/plugins/album`——「专辑列表」集群 8 个菜单整体移出（`menu_album_list`/`menu_album_new`/`menu_album_new_area`/`menu_album_newest`/`menu_album_subsribe_list`/`menu_album_top`/`menu_album_top_area` + 演示迁移的 `menu_album_detail`），每菜单嵌入 `ui.BaseMenu`、构造器收 `ui.BaseMenu`。**key 逐一保持与内置注册相同**：`album_menu`/`album_new_area`/`album_top_area`/`album_new_hot`/`album_new`/`album_top`/`album_sub_list`/`album_detail`——ui 侧跳入集群的调用点零改动（`menu_search_result.go`/`menu_artist_album.go` 跳 `album_detail`、`menu_user_collection.go` 构建 `album_sub_list` 子菜单，均为 key 级跳转）。集群内跨菜单跳转改走导出的 `ui.BuildMenuOrToast`/`ui.MustBuildNoArg`。仅集群内部使用的 opts 契约 `AlbumTopOpts`/`AlbumNewOpts` 随菜单移入插件；被 ui 共享的 `AlbumDetailOpts` 留在 ui（search_result/artist_album/operate 均使用）。`album_detail` 判定为可提取：其 key 与 opts 留在 ui、具体类型随集群移出。`album_menu` 声明主菜单入口 `RegisterMainMenuItem("album_menu", "专辑列表")`。ui 侧：删除 8 个文件；`registry_registrations.go` 删除 `album_detail` 与 batch 2 全部 8 个注册；`expectedMenuKeys` 移除 8 个 album key；`menu_main.go` 删除内置「专辑列表」入口（索引 5，帮助索引 12→11）——**主菜单位置变更：专辑列表从内置索引 5 变为插件项，排在全部内置项之后**；`operate.go` 的 `goToAlbumOfSong` 类型断言 `*AlbumDetailMenu` + `detail.albumID` → 新导出接口 `ui.AlbumDetailIDGetter`（`AlbumID() int64`，ui 不得反向导入插件包）；`registry_test.go` 移除 album checkSub 用例（SearchResultMenu→album_detail 边）+ `builtinMenuCount` 13→12；ui 测试二进制无法链接插件（import cycle），`foxful_integration_test.go` 以包级 init() 注册 `album_sub_list` 行为等价 test-double（内置 `NewUserCollectionMenu` 经 `mustBuildNoArg` 构建该 key，无 double 则 ui 测试启动即 panic）。测试移到 `internal/plugins/album/`（8 key 经 `ui.BuildMenu` 构建 + key 断言 + 集群内跳转边 album_top→album_detail/album_new_hot→album_detail + 区域选择菜单子项 + 入口菜单 3 子项 + 主菜单项注册断言）；聚合器 `plugins.go` 空导入 album + `plugins_test.go` 新增 `TestBlankImportRegistersAlbumPlugin`。— 验证: make lint 0 issues · make test 绿 · make build 绿 · gofmt 干净
- **3.9.8 歌手集群提取为第五个真实插件（批量菜单提取 + 主菜单项 + 共享 opts 判定 + 多去重接口示范）**：`internal/plugins/artist`——「热门歌手 / 歌手详情」集群 6 个菜单整体移出（`menu_artist_list`/`menu_artist_detail`/`menu_artist_album`/`menu_artist_song`/`menu_artists_subscribe_list`/`menu_hot_artists`），每菜单嵌入 `ui.BaseMenu`、构造器收 `ui.BaseMenu`。**key 逐一保持与内置注册相同**：`hot_artists`/`artist_detail`/`artist_song`/`artist_album`/`artist_of_song`/`artists_sub_list`——ui 侧跳入集群的调用点零改动（`menu_search_result.go` 跳 `artist_detail`、`menu_user_collection.go` 构建 `artists_sub_list` 子菜单、operate.go `goToArtistOfSong` 跳 `artist_detail`/`artist_of_song`，均为 key 级跳转）。集群内跨菜单跳转改走导出的 `ui.BuildMenu`/`ui.BuildMenuOrToast`。仅集群内部使用的 opts 契约 `ArtistAlbumOpts`/`ArtistSongOpts` 随菜单移入插件；被 ui 共享的 `ArtistDetailOpts`/`ArtistsOfSongOpts` 留在 ui（search_result/operate 均使用）。`artist_detail` 是 Phase 3.0 原型样本菜单之一，本次判定为可提取：其 key 与 opts 留在 ui、具体类型随集群移出（与 `album_detail` 同判据）。`hot_artists` 声明主菜单入口 `RegisterMainMenuItem("hot_artists", "热门歌手")`。ui 侧：删除 6 个文件；`registry_registrations.go` 删除 `artist_detail`（原型样本注册）与 batch 3 全部 6 个注册；`expectedMenuKeys` 移除 6 个 artist key；`menu_main.go` 删除内置「热门歌手」入口（索引 8，帮助索引 11→10）——**主菜单位置变更：热门歌手从内置索引 8 变为插件项，排在全部内置项之后**；`operate.go` 的 `goToArtistOfSong` 类型断言 `*ArtistDetailMenu`（`detail.artistId`）与 `*ArtistsOfSongMenu`（`artists.song.Id`）→ 新导出接口 `ui.ArtistDetailIDGetter`（`ArtistID() int64`）与 `ui.ArtistsOfSongSongIDGetter`（`SongID() int64`，ui 不得反向导入插件包）；`registry_test.go` 移除 artist checkSub 用例（SearchResultMenu→artist_detail 边）+ `builtinMenuCount` 12→11；`foxful_integration_test.go` 以包级 init() 注册 `artists_sub_list` 行为等价 test-double（内置 `NewUserCollectionMenu` 经 `mustBuildNoArg` 构建该 key，无 double 则 ui 测试启动即 panic）。测试移到 `internal/plugins/artist/`（6 key 经 `ui.BuildMenu` 构建 + key 断言 + 集群内跳转边 hot_artists→artist_detail/artists_sub_list→artist_detail + artist_detail 双子菜单 + 主菜单项注册断言）；聚合器 `plugins.go` 空导入 artist + `plugins_test.go` 新增 `TestBlankImportRegistersArtistPlugin`。— 验证: make lint 0 issues · make test 绿 · make build 绿 · gofmt 干净
- **3.9.9 主菜单项参数化构建（registry.go + menu_main.go）**：`MainMenuItem` 新增可选 `Build func(base BaseMenu) Menu` 字段——nil 时主菜单经 `mustBuildNoArg(Key, base)` 构建（无参 provider）；非 nil 时由插件以自身 options 构造菜单（参数化 provider 主菜单入口，如 user_playlist 携带 `UserPlaylistOpts{UserID}` 的场景）。`RegisterMainMenuItem(key, title)` 保持便捷形式（Build=nil），新增 `RegisterMainMenuItemWith(key, title, build)`；两者对空 key/title 或重复 key 均 panic。`NewMainMenu` 追加插件项时改用 `item.Build(base)`（nil 回退 `mustBuildNoArg`），保留「key 必须已注册」显式启动断言（先断言再构建给出清晰错误；Build 以插件自身 options 构造）。现有插件（checkupdate/lastfm/dj/album/artist）用便捷形式，零改动。— 验证: registry_test 新增 `TestRegisterMainMenuItemWithBuilder`（builder 入口构建参数化菜单 + key 已注册断言）+ make lint/test/build 绿
- **3.9.10 推荐集群提取为第六个真实插件（批量主菜单项 + 登录门控经 BaseMenu 转发示范）**：`internal/plugins/recommend`——「推荐/播放历史」集群 5 个菜单整体移出（`menu_daily_recommend_songs`/`menu_daily_recommend_playlists`/`menu_personal_fm`/`menu_recent_songs`/`menu_ranks`），每菜单嵌入 `ui.BaseMenu`、构造器收 `ui.BaseMenu`。**key 逐一保持与内置注册相同**：`daily_songs`/`daily_playlists`/`personal_fm`/`recent_songs`/`ranks`——ui 侧跳入集群的调用点零改动（仅主菜单入口，均为 key 级跳转）。`m.svc.*` → BaseMenu 导出转发（`m.User()`/`m.ToLoginPage(enterMenuCallback(main))`/`m.Player()`；`enterMenuCallback` 插件内镜像 ui 未导出的 `EnterMenuCallback`）。集群内跨菜单跳转改走 `ui.BuildMenu`（`ranks`/`daily_playlists` 的 SubMenu 跳 `playlist_detail`，后者留在 ui）。5 个入口菜单各自声明主菜单入口。ui 侧：删除 5 个文件；`registry_registrations.go` 删除 5 个注册；`expectedMenuKeys` 移除 5 个 recommend key；`menu_main.go` 删除内置「每日推荐歌曲/每日推荐歌单/私人FM/排行榜/最近播放歌曲」入口（索引 0/1/4/6/8，帮助索引 10→5）——**主菜单位置变更：5 个推荐入口从内置索引 0/1/4/6/8 变为插件项，排在全部内置项之后**；`registry_test.go` 重接 Ranks 引用（`TestMustBuildNoArg`/`TestBuildMenuOptsTypeMismatch`/`TestMenuNavigationSmoke` → 内置 `user_collect`/`high_quality_playlists`，Ranks→PlaylistDetail 边移入插件测试）+ `builtinMenuCount` 11→6。测试移到 `internal/plugins/recommend/`（5 key 经 `ui.BuildMenu` 构建 + key 断言 + 集群内跳转边 ranks/daily_playlists→playlist_detail + 5 主菜单项注册断言）；聚合器 `plugins.go` 空导入 recommend + `plugins_test.go` 新增 `TestBlankImportRegistersRecommendPlugin`。— 验证: make lint 0 issues · make test 绿（avcore 为既有环境性偶发）· make build 绿 · gofmt 干净

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
- [x] 3.6.3 首个真实插件提取（check_update）— 08b5d7dd（`internal/plugins/checkupdate`：`CheckUpdateMenu` 嵌入 `ui.BaseMenu`，检查/通知逻辑自 `menu_check_update.go` 原样迁入；`registry.go` init() 注册 `"check_update"`；聚合器 `internal/plugins/plugins.go` 空导入 + `cmd/musicfox.go` 空导入聚合器；menu_main 检查更新入口经 `buildMenuOrToast("check_update")` 路由并执行插件菜单 Action；删除 `menu_check_update.go`；`check_update` 原本就不在内置注册与 `expectedMenuKeys` 中（无需删除），完整性断言不受影响、无重复 key panic。启动自动检查（netease.go config 门控）保留 shell 级：无启动钩子机制，ui 不得反向导入插件包，notify 内容在 shell 内联。`BuildToastNotificationSpec` 导出供插件复用；ui 测试二进制无法链接插件（import cycle），以行为等价 test-double 注册 `check_update`。测试：插件菜单经 registry 工厂构建/key/MenuViews、通知 spec 三态、聚合器空导入路径注册断言）

验证：`make lint` 0 issues · `make test` 绿 · `make build` 绿 · 改动文件 `gofmt -l` 干净（`song_info_renderer.go` 为 HEAD 既有的非 gofmt 债务，非本阶段引入，未触碰）。avcore 测试在 `make test` 全量运行时偶发 FAIL、单独/重跑即绿（与本阶段改动无关）。

### Phase 3.8: decoupling round 2 — renderer interface refactor

- [x] 3.8.1 访问器几何转发 — e2c369fc
- [x] 3.8.2 renderer 去壳 — ae33e2f8
- [x] 3.8.3 lastfm 页面 opts 去壳 — a4c1fb3a
- [x] 3.8.4 bootstrap 构造器收窄 — 579d3505
- [x] 3.8.5 player 桌面歌词访问器 — ae382fe5

### Phase 3.9: plugin ecosystem practicalization — main-menu items and startup hooks

- [x] 3.9.1 Mechanism A：主菜单插件项（RegisterMainMenuItem/MainMenuPluginItems + NewMainMenu 追加构建 + 完整性断言）— 见提交（上方 Implementation Plan 3.9 节）
- [x] 3.9.2 Mechanism B：启动钩子（RegisterStartupHook + runStartupHooks 于 InitHook 第 10 步调用，recover 隔离）
- [x] 3.9.3 check_update 插件化改造：主菜单入口 + 启动检查钩子 + BeforeEnterMenuHook 单次 Enter 触发；移除 mainMenuCheckUpdateIndex 特判与 netease.go 硬编码启动检查
- [x] 3.9.4 测试与门禁：make lint 0 issues · make test 绿 · make build 绿 · gofmt 干净
- [x] 3.9.5 Last.fm 提取为第二个真实插件（服务访问 + 页面插件 + 主菜单项）— 见提交（上方 Implementation Plan 3.9.5；主菜单位置变更：LastFM 由内置索引 13 → 插件项排在全部内置项之后）
- [x] 3.9.6 DJ/电台集群提取为第三个真实插件（批量菜单提取 + 跨菜单跳转 + 主菜单项）— 见提交（上方 Implementation Plan 3.9.6；主菜单位置变更：主播电台由内置索引 12 → 插件项排在全部内置项之后；帮助索引 13→12）
- [x] 3.9.7 专辑集群提取为第四个真实插件（批量菜单提取 + 主菜单项 + 共享 opts 判定）— 见提交（上方 Implementation Plan 3.9.7；主菜单位置变更：专辑列表由内置索引 5 → 插件项排在全部内置项之后；帮助索引 12→11）
- [x] 3.9.8 歌手集群提取为第五个真实插件（批量菜单提取 + 主菜单项 + 共享 opts 判定 + 多去重接口）— 1c52c9fa / 8ffec4bf（上方 Implementation Plan 3.9.8；主菜单位置变更：热门歌手由内置索引 8 → 插件项排在全部内置项之后；帮助索引 11→10）
- [x] 3.9.9 主菜单项参数化构建（`MainMenuItem.Build` + `RegisterMainMenuItemWith`；`NewMainMenu` 走 `item.Build(base)`，nil 回退 `mustBuildNoArg`，key 已注册断言保留）— f7ed63cd（上方 Implementation Plan 3.9.9）
- [x] 3.9.10 推荐集群提取为第六个真实插件（批量主菜单项 + 登录门控经 BaseMenu 转发）— 4f029d67（上方 Implementation Plan 3.9.10；主菜单位置变更：每日推荐歌曲/每日推荐歌单/私人FM/排行榜/最近播放歌曲 由内置索引 0/1/4/6/8 → 插件项排在全部内置项之后；帮助索引 10→5）

验证：`make lint` 0 issues · `make test` 绿（无 FAIL）· `make build` 绿 · 改动文件 `gofmt -l` 干净。

### Phase 3.6 handler-services（解耦调查 work package 2）

operate 辅助函数 / 操作执行器 / 事件处理器全面去 `*Netease`（仅访问路径变更，行为保持）：

- [x] 3.6.4 操作执行器去耦合 — a2151f29（`NewOperation(svc *menuServices, coreFunc CoreFunc)`，`CoreFunc` 改为 `func(svc *menuServices) model.Page`；认证经 `svc.User()`、登录跳转经 `svc.ToLoginPage`、loading 经 `svc.MustMain`；operate.go 全部 NewOperation 调用点传入访问器，coreLogic 闭包改收 `*menuServices`，去除闭包内 `newMenuServices(n)` 二次派生；顺带修复 executor.go 被 whole-file lint 暴露的 gci import 分组）
- [x] 3.6.5 operate 辅助函数迁移 — 2521d7e4（~30 个辅助函数签名 `func(n *Netease, …)` → `func(svc *menuServices, …)`：getTargetSong/getSelectedPlaylist/likeSong/trashSong/confirmTrashSong/trashSongWithConfirm/downloadSong/handleSongDownload/downloadSongLrc/handleLyricDownload/findSimilarSongs/goToAlbumOfSong/goToArtistOfSong/openInWeb/collectSelectedPlaylist/subscribeAlbum/subscribeArtist/appendSongsToCurPlaylist/openAddSongToUserPlaylistMenu/addSongToUserPlaylist/delSongFromPlaylist/clearSongCache/action/shareItem/searchSong 及 getTargetSong 家族；`newBaseMenu(n)` → `newBaseMenuFromSvc(svc)`；action_select 的 actionItemsForMenu/buildSongActions/buildPlaylistActions 同迁，buildActionItems 走 `m.svc`；menu.go 右键/动作路径走 `e.svc`；event_handler 调用点走 `h.svc`；playOrToggle 收 `*menuServices`；顺带修复 action_select 被 whole-file lint 暴露的 gci import 分组 + ST1021 ActionItem 文档注释）
- [x] 3.6.6 事件处理器字段清除 — ad4c6784（`EventHandler` 删除 `netease *Netease` 字段，仅保留 `svc` 访问器；`NewEventHandler` 参数保留以构造访问器。handler-services 解耦完成：辅助函数/执行器/处理器全部只经 `menuServices` 工作）
- **剩余 `.Netease()` 逃生口（4 处，均有依据）**：`menu.go` BaseMenu.Netease() 方法体（对外插件边界逃生口，docs/plugin_development.md 已文档化，外部插件经 `base.Netease()` 访问薄壳）· `registry_registrations.go` neteaseFromBase（main_menu/local_search bootstrap 直构，构造器先于 baseMenu 签名，3.3.5 裁决保留）· `lastfm_profile.go` ×2（页面 opts 携带 `*Netease` 壳引用，3.3.5 合理残留「页面由 shell 构建并持有 shell 引用」）

验证：`make lint` 0 issues · `make test` 绿 · `make build` 绿 · 改动文件 `gofmt -l` 干净。验证：`make lint` 0 issues · `make test` 绿 · `make build` 绿 · 改动文件 `gofmt -l` 干净（`foxful_integration_test.go`/`song_info_renderer.go` 为 HEAD 既有的非 gofmt 债务，非本阶段引入，未触碰）。

### Phase 3.6 player playlist API（解耦调查 work package 3）

Player 播放列表变更 API（最深耦合：外部代码直接变更 Player 内部播放列表状态）。仅访问路径变更，行为保持：

- [x] 3.6.7 播放列表状态封装 — 44ed2603（新增 Player 方法：`ReinitializePlaylist(index, playlist)`（void 包装 `playlistManager.Initialize`，不引入 cancelGaplessPreload 行为差异）、`RemoveSong(index)`、`LoadPlaylistState()`、`SetPlayingMenu(key, menu model.Menu)`（内部断言为 ui.Menu，非 ui.Menu 时清空 playingMenu 保留 key，与原逻辑等价）、`MarkPlaylistModified()`（playingMenu=nil + key += "modified"）、`MarkPlaylistUpdated()`（playlistUpdateAt=now）、读取 getter `PlayingMenuKey()`/`PlayingMenu()`/`PlaylistUpdateAt()`。`InitSongManager` 委托 `ReinitializePlaylist`。调用点全量迁移：event_handler.go playOrToggle + cur_playlist 副标题读取、operate.go appendSongsToCurPlaylist/delSongFromPlaylist（含 RemoveSong 路径）、cur_playlist.go BottomOutHook、menu_personal_fm.go、menu_similar_songs.go、netease.go 启动加载 LoadState。顺带修复 whole-file lint 暴露的既有 ST1003：menu_similar_songs.go 的 relateSongId/songId → relateSongID/songID（含 operate.go:378 引用点））
- [x] 3.6.8 Player 播放列表 API 测试 — 7613ca49（新增 `TestPlayerPlaylistStateAPI`：ReinitializePlaylist/SetPlayingMenu/MarkPlaylistModified/MarkPlaylistUpdated/RemoveSong 状态转移断言；player_gapless_test 的 Initialize 迁移到 `ReinitializePlaylist` seam；RemoveSong 返回值为下一曲语义与 playlist manager 一致）

**最终耦合盘点**：`.Netease()` 逃生口 4 处（不变，均有依据：menu.go BaseMenu.Netease() 方法体 / registry_registrations.go neteaseFromBase / lastfm_profile.go ×2）；`playlistManager`/`playingMenuKey`/`playingMenu`/`playlistUpdateAt` 在 player.go 之外的直接变更 **0**（`rg '\.playingMenuKey|\.playingMenu\b|\.playlistUpdateAt|\.playlistManager' internal/ui --glob '!player*.go' --glob '!*_test.go'` 为空；白盒 gapless 测试仍直接访问 playlistManager.SetPlayMode，无 Player 包装且 SetMode 含存储副作用，不适用单测隔离，属测试内部细节）。

验证：`make lint` 0 issues · `make test` 绿（21 包 ok，无 FAIL）· `make build` 绿 · 改动文件 `gofmt -l` 干净（`song_info_renderer.go` 为 HEAD 既有的非 gofmt 债务，非本阶段引入，未触碰）。

### Phase 3.7: framework lifecycle hardening (round 2)

framework 生命周期语义加固（#646 评审后续第二轮）。**正确用法行为不变**，误用语义显式化：

- **Scope 状态机**：`started`/`disposed` 双布尔；`Start` 对已启动/已 dispose 作用域返回显式错误（非静默双启动），失败回滚后保持未启动可重试；`Stop` 对未启动作用域为 nil no-op（`defer scope.Stop()` 恒安全）；`Dispose` 幂等且 final——已启动作用域先隐式 Stop（子作用域与插件收到 Stop 再 Dispose，cordis dispose 语义，防仅实现 Stop 的插件泄漏），dispose 后 `Start`/`Add` 返回显式错误；状态机按 Scope 独立，无父耦合（子作用域可直接 Start）。
- **Stop/Dispose 错误聚合**：`Stop` 对插件/子作用域 Stop 失败时记录并继续逆序停止其余，`errors.Join` 聚合返回（Go stdlib，framework 保持零依赖）；`Dispose` 同样跨子作用域→插件聚合，且状态清理（children/plugins=nil、parent 解绑）在出错时仍执行；`rollback` 保持 best-effort 但同样聚合停止错误随原始 Start 错误一并返回。
- **事件 handler panic 隔离**：`Emit` 四类 handler 每次调用均包 recover，panic 转为 `framework: <name> handler panicked: <r>` + `debug.Stack()` 截断 ~1KB 的错误；语义与返回错误一致——listener/middleware/serial 中断链，parallel 经 errCh 传递且全部 handler 仍跑完；emitter 自身永不因 handler panic 而 panic。

- [x] 3.7.1 Scope 生命周期状态机 + Stop/Dispose 错误聚合 — 4280f454（`Scope` 增加 `started`/`disposed`；`Start` 双启动/已 dispose 返回显式错误、失败回滚后保持未启动可重试；`Stop` 未启动为 nil no-op、失败 `errors.Join` 聚合且继续逆序停止其余、停止后清除 started；`Dispose` 幂等且 final——已启动作用域先隐式 Stop、状态清理出错仍执行、dispose 后 Start/Add 拒绝；`rollback` 聚合停止错误随原始 Start 错误返回；`Add` 签名改为返回 error，`newAppScope` 接线点同步处理）
- [x] 3.7.2 事件 handler panic 隔离 — 5e0d4bc2（`Emit` 四类 handler 每次调用经 `invokeHandler` recover，panic 转 `framework: <name> handler panicked: <r>` + `debug.Stack()` 截断 ~1KB 错误；语义与返回错误一致——listener/middleware/serial 中断链，parallel 经 errCh 传递且全部 handler 仍跑完；emitter 自身永不因 handler panic 而 panic）
- [x] 3.7.3 测试覆盖 — b52926b2（状态机：双启动拒绝/Stop 前置 no-op/Dispose 隐式 Stop 已启动作用域/Start 与 Add 于 dispose 后拒绝/失败启动保持未启动可重试；错误聚合：Stop 聚合插件与子作用域失败仍全停、Dispose 聚合子+插件失败仍清状态、Dispose 已启动作用域透出隐式 Stop 失败；panic 隔离：listener/middleware/serial panic 中断链返回框架错误、parallel panic 经 errCh 且其余 handler 仍跑完（-race 下全绿）；既有 Dispose 测试断言更新为隐式 Stop 顺序）

验证：`make lint` 0 issues · `make test` 绿 · `make build` 绿 · 改动文件 `gofmt -l` 干净。

### Phase 3.8: decoupling round 2 — renderer interface refactor

5 个 renderer（lyric/songInfo/progress/cover/composite）不再持有 `*Netease`，改为依赖 `*menuServices` 访问器 + 自有业务服务：

- [x] 3.8.1 访问器几何转发 — e2c369fc（menuServices 新增 5 个薄壳几何转发：EffectiveWindowHeight/SpectrumLines/GetCoverWidth/GetCoverEndColumn/GetLyricPosition，各转发至 Netease 薄壳方法并 nil 安全）
- [x] 3.8.2 renderer 去壳 — ae33e2f8（lyric/songInfo/progress/cover/composite 的 `netease *Netease` 字段 → `svc *menuServices`；构造器首参 netease → svc；`WindowWidth` → `svc.App().WindowWidth()`，`Player()` → `svc.Player()`，`playbarHoveredElement` → `svc.PlaybarHoveredElement()`，跨 renderer 几何走新转发；netease.go Components 构造点与 renderer 测试同步；`grep -n netease internal/ui/*renderer*.go` 为 0；顺带修复 whole-file lint 暴露的既有债务：ST1003 initialism 重命名（currentSongId→currentSongID/picUrl→picURL/getCoverUrl→getCoverURL/cachedSongId→cachedSongID）、SA1019（util.SetFg*Style → style.FG/FGBG）、移除 dead 字段 lastAngle/lastViewTime）
- [x] 3.8.3 lastfm 页面 opts 去壳 — a4c1fb3a（`LastfmCustomAPIPageOpts`/`LastfmAuthPageOpts` 携带 `svc *menuServices` 取代 `Netease *Netease`；lastfm_custom_api_account/lastfm_auth/lastfm_qr_auth 三页构造器收 svc 并删 netease 字段，导航改经 `svc.MustMain()`/`svc.App().RerenderCmd(true)`；lastfm_profile 跳转点直传 `m.svc`；lastfm 三页 `grep netease` 为 0；`svc.Netease()` 计数 3 → 1；配套 registry_test/custom_page_background_test/lastfm_custom_api_account_test 同步）
- [x] 3.8.4 bootstrap 构造器收窄 — 579d3505（`NewMainMenu`/`NewLocalSearchMenu` 签名 `*Netease` → `baseMenu`，函数体不再经 `newBaseMenu` 派生；新增导出 `NewBaseMenu(n *Netease) BaseMenu` 桥接 internal/commands 入口（`newBaseMenuFromSvc` 收 `*menuServices` 不可导出）；删除 `neteaseFromBase`（唯一使用方随 3.8.4 注册改直传 base 而消亡）；顺带修复 commands/netease.go 被 whole-file lint 暴露的既有 gci import 分组（neteaseutil 移入 default 段）与 ST1000 包注释）
- [x] 3.8.5 player 桌面歌词访问器 — ae382fe5（`updateDesktopLyrics` 的 `p.netease.DesktopLyrics()` → `p.svc.DesktopLyrics()`、`p.netease.GetDesktopLyricsLines()` → `p.svc.GetDesktopLyricsLines()`；menuServices 新增 `GetDesktopLyricsLines` 转发至 shell 方法并 nil 安全返回零值/-1；player.go 不再直连 `p.netease` 服务方法）

验证：`make lint` 0 issues · `make test` 绿（无 FAIL）· `make build` 绿 · 改动文件 `gofmt -l` 干净。

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
