# 插件开发指南（plugin development）

本文档描述 go-musicfox 的对外插件边界表面（Phase 3.4 接口层预留）：菜单/页面 provider 注册表、服务解析、生命周期，以及第三方插件当前如何接入（编译期注册）。

## 概述

从 Phase 3.2 起，go-musicfox 的全部菜单与页面统一走 **provider 注册表**（key → 参数化工厂），跳转不再硬编码构造函数；从 Phase 3.1 起，全部业务能力注册为 framework 命名服务，组件按名解析。这套机制为未来的第三方插件生态预留了边界：**opts 结构体（参数契约）即插件契约**。

- 本文档适用对象：想为 go-musicfox 贡献/接入一个「菜单」或「页面」的开发者。
- 当前形态：**编译期注册构造器 + 运行时激活**（P5 cordis 化：包 `init()` 只做 `framework.RegisterPlugin(id, 构造器)`，实际注册移入插件 `Start`，由前端 scope 挂载启动；见「当前形态」一节）。**运行时动态加载**的落地形态是 **WASM 插件**（不重编译即启用，见「WASM 插件（实验性）」）；Go `plugin` 共享库 / 子进程形态不在支持范围（spec Non-goals）。
- 行为约束：插件不得改变核心行为，错误经 `(Menu, error)` + toast 暴露（见「行为保持契约」）。

## 插件边界表面

| 能力 | API | 位置 |
|------|-----|------|
| 菜单注册 | `RegisterMenu[T](key, factory)` | `internal/ui/registry.go` |
| 页面注册 | `RegisterPage[T](key, factory)` | `internal/ui/registry.go` |
| 菜单跳转 | `BuildMenu[T]` / `MustBuildNoArg` / `MustBuild` / `BuildMenuOrToast`（导出形式，供插件使用） | `internal/ui/registry.go` |
| 页面跳转 | `BuildPage[T]` / `BuildPageOrToast[T]` | `internal/ui/registry.go` |
| 主菜单入口 | `RegisterMainMenuItem(key, title)` / `RegisterMainMenuItemWith(key, title, build)` / `RegisterMainMenuItemAfter(key, title, after, build)` / `MainMenuPluginItems()` | `internal/ui/registry.go` |
| 启动钩子 | `RegisterStartupHook(fn)`（捕获插件 id 后委托 `internal/framework` 注册表，执行在 `core.Engine.Startup` 启动序第 10 步） | `internal/ui/registry.go` |
| 快捷键操作 | `keybindings.RegisterOperate(name, desc, keys)` → `ui.RegisterOperateHandler(op, fn)` | `internal/keybindings/keybindings.go` / `internal/ui/event_handler_operate.go` |
| 右键菜单项 | `RegisterContextMenuContrib(contrib)` / `ContextMenuContribs()` | `internal/ui/context_menu.go` |
| 状态栏组件 | `RegisterStatusBarComponent(comp)` / `StatusBarComponents()` | `internal/ui/status_bar.go` |
| 轨 B 命令（UI-agnostic，跨前端） | `ui.RegisterCommand(cmd)`（插件归属注入）→ `frontend.RegisterCommand` / `Commands()` / `CommandByKey` | `internal/ui/register_command.go` / `internal/frontend/command.go` |
| 菜单基座（可嵌入） | `BaseMenu`（含导出转发方法 + `Services()`） | `internal/ui/menu.go` |
| 服务解析 | `framework.Context` / `framework.ServiceOf[T]` | `internal/framework/context.go` |
| 类型安全访问器 | `menuServices`（`svc.Player()` 等；导出别名 `MenuServices` + `NewMenuServices(ctx)`） | `internal/ui/menu_accessor.go` |
| 自定义页面布局助手 | `PageTitleView` / `PageMenuTitleView[WithBack]` / `PageInput*` / `FinishCustomPageView` 等（页面插件渲染用） | `internal/ui/page_layout.go` |
| 确认弹窗 | `ShowConfirmPopup(app, title, content, onConfirm)` | `internal/ui/confirm_popup.go` |
| 服务名常量 | `ServicePlayer` 等 | `internal/ui/services.go` |
| 生命周期 | `framework.Scope` / `framework.Plugin` | `internal/framework/plugin.go` |

### 注册表：key → 参数化工厂

```go
// 菜单 provider：opts 类型 T 即该菜单的参数契约。factory 的 base 参数用
// baseMenu（BaseMenu 的别名），等价地可写为 BaseMenu——注册签名对
// BaseMenu 类型工厂直接兼容（别名可互换）。
func RegisterMenu[T any](key string, f func(base baseMenu, opts T) (Menu, error))

// 跳转：唯一运行时类型断言藏在 registry 内部，调用侧零断言。
func BuildMenu[T any](key string, base baseMenu, opts T) (Menu, error)

// 无参菜单的紧凑辅助（共享占位类型 NoArgMenuOpts{}）。
func mustBuildNoArg(key string, base baseMenu) Menu

// 跳转失败经 toast 降级（不 panic），返回 nil。
func buildMenuOrToast[T any](key string, base baseMenu, opts T) Menu

// 插件内的菜单跳转用导出形式（语义与上面的内置形式逐一等价）：
func MustBuild[T any](key string, base baseMenu, opts T) Menu        // = mustBuild：静态 menuList 构建失败 panic（编程错误）
func MustBuildNoArg(key string, base baseMenu) Menu                  // = mustBuildNoArg：无参菜单的紧凑形式
func BuildMenuOrToast[T any](key string, base baseMenu, opts T) Menu // = buildMenuOrToast：SubMenu 跳转失败 toast 降级

// 页面 provider：返回 model.Page。
func RegisterPage[T any](key string, f func(opts T) (model.Page, error))
func BuildPage[T any](key string, opts T) (model.Page, error)

// 跳转失败经 toast 降级（不 panic），返回 nil。插件内页面打开点用导出形式。
func BuildPageOrToast[T any](key string, opts T) model.Page

// 主菜单入口（Phase 3.9）：key 必须已在菜单注册表中注册（未注册的 key 在
// NewMainMenu 启动时 panic，作为启动完整性信号）。无 Build 的入口 key 必须
// 是无参菜单 provider（经 mustBuildNoArg 构建）；带 Build 的入口由插件以自身
// options 构造菜单（参数化 provider 主菜单入口）。**顺序（after-anchor 链）**：
// 每个入口声明其前驱项 key（`After` = 前驱项 key），NewMainMenu 从
// `ui.MainMenuStart`（`_main_start`，第一项的哨兵锚点）沿链走序，复现插件化
// 前的主菜单原始顺序（每日推荐歌曲 → 每日推荐歌单 → … → 帮助 → 检查更新）。
// 插入一个菜单 = 声明一个锚点，其余项不漂移（无需重排编号）；空 After 的项
// （便捷形式）追加在链尾（注册序保持，既有"追加在末尾"行为）。链完整性由
// NewMainMenu 断言：每个 After 目标必须存在、每项恰好可达一次、链长 == 总项
// 数（孤儿/环 panic）。
func RegisterMainMenuItem(key, title string)                                             // 便捷形式，Build = nil，末尾追加
func RegisterMainMenuItemWith(key, title string, build func(base BaseMenu) Menu)         // 便捷形式，末尾追加
func RegisterMainMenuItemAfter(key, title string, after string, build func(base BaseMenu) Menu) // 声明前驱项 key（after 非空；锚点存在性由 NewMainMenu 断言）
func MainMenuPluginItems() []MainMenuItem                                                // 快照，供 NewMainMenu 构造

// 启动钩子（Phase 3.9）：注册启动任务。核心引擎在启动序列第 10 步
// （用户/登录就绪后、自动播放前）经 framework.RunStartupHooks 按注册序调用，
// 每个 hook 带 panic 隔离（recover + 日志，不阻断启动）。
func RegisterStartupHook(fn func()) // nil 会 panic
```

规则：

- 注册时机：**P5 cordis 化后**为「`init()` 只注册构造器 + `Start` 内实际注册」（前端 scope 挂载时执行；`init()` 时序早于一切 ctx 的旧形态仅存于测试与迁移窗口）。key 为空、factory 为 nil 或 key 重复注册都会 panic（程序员错误，启动即暴露）。
- 参数契约进类型系统：如 `PlaylistDetailOpts{PlaylistID int64}`、`ArtistDetailOpts{ArtistID, Name}`；与注册类型不匹配的构建在 registry 内部报错。
- 注册表本身是 framework 服务（`ServiceMenuRegistry` / `ServicePageRegistry`），提供 `Registered(key)` / `Keys()` 供完整性断言与测试使用。

### 菜单基座：BaseMenu

`BaseMenu`（`internal/ui/menu.go`）是菜单的导出基座，外部插件菜单**嵌入它**即获得 `ui.Menu` 接口的默认实现（`IsPlayable`/`IsLocatable`/`Action`/`ContextMenuItems` 等）与 `menuServices` 访问器。`baseMenu` 是它的别名（alias），内置菜单与注册签名不受影响，二者可互换。

`BaseMenu` 导出以下转发方法（每个都转发到 `menuServices`，nil 安全）：

```go
// 服务解析：
base.Player()       // *Player
base.User()         // *structs.User（未登录为 nil）
base.TrackManager() // *track.Manager
base.LyricService() // *lyric.Service
base.DesktopLyrics()
base.CoverRenderer()
base.ShareSvc()
base.Lastfm()
base.Ctx()          // *framework.Context

// 薄壳/导航：
base.App()                       // *model.App
base.MustMain()                  // *model.Main
base.Rerender()                  // tea.Cmd（强制重绘）
base.Search()                    // *SearchPage（shell 单例）
base.ToLoginPage(callback)       // (model.Page, tea.Cmd)
base.ToSearchPage(searchType)    // (model.Page, tea.Cmd)

// 访问器本体（Phase 3.9）：把菜单自身的 menuServices 访问器传给页面 opts
// 或构造函数（页面插件的 opts 字段类型就是 MenuServices，见示例四）。
base.Services()                  // ui.MenuServices

// 逃生口（旧辅助函数仍收 *Netease；新插件代码优先用上面的访问器）：
base.Netease()                   // *Netease
```

外部菜单只需实现/覆写少量方法（`GetMenuKey`、`MenuViews`、`SubMenu`、数据加载的 `BeforeEnterMenuHook` 等），即可被注册并导航。

### 服务解析：Context / ServiceOf

```go
// 服务容器（cordis 语义退化实现）：按名注册、解析、覆写。
type Context struct{ ... }
func (c *Context) Provide(name string, svc any)   // 重名 panic
func (c *Context) Override(name string, svc any)  // 未注册 panic
func (c *Context) Service(name string) any

// 类型安全解析。bool 必须处理（缺失时记录 + 降级），禁止裸丢弃。
func ServiceOf[T any](c *Context, name string) (T, bool)
```

业务能力（player / lyricService / trackManager / desktopLyrics / coverRenderer / userService / loginService / shareSvc / lastfm 以及两个注册表）在启动时经 `registerServices`（`internal/ui/services.go`）注册进容器。

### 类型安全访问器：menuServices / MenuServices

菜单代码不直接持有 `*Netease` 字段，而是经 `baseMenu.svc`（`*menuServices`）访问：

```go
// 服务解析（nil 安全，缺失时告警日志 + 返回零值）：
svc.Player()       // *Player
svc.User()         // *structs.User（未登录为 nil）
svc.TrackManager() // *track.Manager
svc.LyricService() // *lyric.Service
svc.DesktopLyrics()
svc.CoverRenderer()
svc.ShareSvc()
svc.Lastfm()

// 薄壳方法转发（导航/渲染/状态）：
svc.MustMain() / svc.Rerender(force) / svc.SaveActiveTheme(name) / ...

// 逃生口（迁移窗口用，新代码避免）：
svc.Netease() // *Netease 薄壳
```

插件边界（Phase 3.9）：`MenuServices` 是 `*menuServices` 的导出别名（alias），外部包可引用访问器类型于签名/opts 字段/页面构造函数（如 `NewLastfmAuthPage(svc ui.MenuServices)`）。外部获取访问器有两个入口：

- `base.Services()`（`BaseMenu` 方法）——菜单把自身访问器传给页面 opts；
- `ui.NewMenuServices(ctx *framework.Context)`——只挂 context 不挂 shell 的访问器（shell 相关转发降级为 nil/零值），插件测试与 shell 无关的页面流使用。

页面插件渲染自定义页面（标题/返回按钮/输入框/按钮/面包屑）复用 ui 导出的布局助手（`PageTitleView`、`PageMenuTitleView[WithBack]`、`PageInputStyles`、`FocusPageInput`/`BlurPageInput`、`PageInputView`、`PageSubmitButton`、`PageButton`、`PageButtonHoverView`、`PageSubmitText`、`SetPageInputCursor`、`PageMenuTitleRow`、`PageBackButtonWidth`、`PageBreadcrumbMotion`/`PageBreadcrumbClick`、`FinishCustomPageView`），确认弹窗用 `ui.ShowConfirmPopup`——示例四（Last.fm 插件）即此形态。

### 服务名常量

全部服务名集中定义于 `internal/ui/services.go`：`ServicePlayer`、`ServiceLyricService`、`ServiceTrackManager`、`ServiceDesktopLyrics`、`ServiceCoverRenderer`、`ServiceUserService`、`ServiceLoginService`、`ServiceShareSvc`、`ServiceLastfm`、`ServiceMenuRegistry`、`ServicePageRegistry`。插件解析服务时应引用常量而非裸字符串。

### 生命周期：Scope / Plugin

```go
type Plugin interface {
    Start(ctx *Context) error
    Stop() error
    Dispose() error
}
type PluginWithDeps interface { // 可选：Start 前注入依赖
    Plugin
    Deps(ctx *Context) error
}
type Scope struct{ ... }
func NewScope() *Scope
func (s *Scope) Add(plugin Plugin)
func (s *Scope) NewScope() *Scope // 子作用域随父作用域启停，Dispose 先子后父
func (s *Scope) Start(ctx *Context) error  // 按注册序
func (s *Scope) Stop() error               // 逆序（先子后父）
func (s *Scope) Dispose() error            // 递归清理，幂等
```

当前生产接入小切片：shareSvc、lastfm 经 `servicePlugin` 适配器在 `newAppScope`（`internal/ui/services_scope.go`）接入生命周期，`Start` 把既有实例注册进共享 Context。第三方插件的资源清理应实现 `Dispose`。

### 快捷键操作：RegisterOperate / RegisterOperateHandler

```go
// 第一步（keybindings 包）：注册唯一操作名、用户可见描述与默认按键
// （可为空）。返回动态分配的 OperateType（固定高位 1000 起递增，与内置
// iota 序列永不冲突）。name 冲突时 panic。默认按键并入 defaultOtherOperateToKeys，
// InitDefaults(true) 自然包含；操作名同步进 opNameToOperateMap，用户配置
// 解析（ProcessUserBindings）无需任何改动即可覆盖插件操作。
op := keybindings.RegisterOperate("my_plugin_action", "我的插件动作", []string{"ctrl+m"})

// 第二步（ui 包）：注册该操作的处理函数，快捷键触发时经 svc 执行，返回
// 页面与命令。nil handler 或重复 op 时 panic。
ui.RegisterOperateHandler(op, func(svc ui.MenuServices, app *model.App) (model.Page, tea.Cmd) {
    // 经 svc 访问服务与导航；返回 nil, nil 表示无页面跳转与命令
    return nil, nil
})
```

- 注册时机：编译期 `init()`（先于 configs 加载与事件循环，默认键在 `InitDefaults(true)` 合并进生效绑定）。用户可在配置 `[keybindings]` 中以操作名自定义或解绑按键，与内置操作行为一致。
- 分发：`EventHandler.handle` 的内置 switch 未命中时，`default` 分支按 op 查注册表，命中则调用 handler（未命中保持原 `(false, nil, nil)` 语义）。handler 收到与 EventHandler 相同的 `svc` 与 `app`。
- 行为保持契约：插件操作是「新增贡献点」，不得覆写内置操作（内置操作的 handler 永不触发——内置 switch 分支优先）；handler 内同样禁止 panic 破坏主循环。

### 右键菜单项：RegisterContextMenuContrib

```go
// 插件在右键菜单末尾贡献一个操作项。全部注册项归入一个 "插件" 分组
// （Header: "插件"），追加在 generic 全局组之后，按注册序显示。空 Title
// 或 nil Action 时 panic 拒绝（编程错误）。
ui.RegisterContextMenuContrib(ui.ContextMenuContrib{
    Title: "我的插件动作", // 菜单项文案（不含图标）
    // Show 决定该项在当前右键上下文是否显示（nil 表示恒显示）。
    Show: func(ctx ui.ContextMenuContext) bool {
        return ctx.SelectedIndex >= 0 // 仅在存在选中行时显示
    },
    // Action 在用户点击时执行，返回页面与命令（导航自行处理）。
    Action: func(svc ui.MenuServices, ctx ui.ContextMenuContext) (model.Page, tea.Cmd) {
        // ctx.Menu 为当前菜单（ui.Menu）、ctx.SelectedIndex 为选中行索引
        // （-1 表示无选中）、ctx.Playing 表示是否针对当前播放。
        return nil, nil
    },
})
```

- 分发：右键菜单项 ID 为 `plugin:<注册序号>`，`BaseMenu.ContextMenuAction` 在 generic/sel/play 分支之后按此 ID 查注册表调用对应 `Action`；序号越界/未注册时返回 `nil, nil`。
- 行为保持契约：插件项是「新增贡献点」，只能追加在菜单末尾，不参与内置 sel/play/generic 组的排序；`Show`/`Action` 内同样禁止 panic 破坏主循环。

### 状态栏组件：RegisterStatusBarComponent

```go
// 插件注册一个状态栏组件，追加到 DefaultStatusBar 居中区域的内置队列/音质
// 组件之后（按注册序）。nil 组件 panic 拒绝（编程错误）。组件只需实现
// model.StatusBarComponent 的 View(a *model.App, m *model.Main) string；
// 可点击组件额外实现 model.InteractiveStatusBarComponent（HandleMouse /
// IsMouseOver，坐标为组件自身渲染文本的局部坐标）。
ui.RegisterStatusBarComponent(myStatusBarComponent{})
```

- 分发：状态栏构造（`NewQueueQualityStatusBar`）读取 `StatusBarComponents()` 快照，按注册序追加在队列/音质组件后；组件只负责渲染（与可选鼠标处理），不持有业务回调。
- 行为保持契约：插件组件是「新增贡献点」，追加在居中区域末尾，不参与内置组件排序；`View`/`HandleMouse` 内同样禁止 panic 破坏主循环。

## 当前形态：编译期注册构造器 + 运行时激活（P5 cordis 化）

第三方插件**今天**的接入方式（9 个内置业务插件即此形态）：

1. 插件包实现一个 `framework.Plugin` 类型（生命周期三件套 `Start`/`Stop`/`Dispose`，零样板可嵌入 `framework.NoopPlugin`；需要观察自身启用状态时嵌入 `framework.PluginBase`）。**声明期只做一件事**：包 `init()` 中 `framework.RegisterPlugin(id, func() framework.Plugin { return &Plugin{} })`。
2. 实际注册（`RegisterMenu` / `RegisterPage` / `RegisterMainMenuItem*` / `RegisterStartupHook` / `RegisterCommand`）移入插件 `Start`，并包在 `ui.WithPlugin(id, name, func(){...})` 作用域内完成归属盖章（`PluginInfo` 记录）。
3. 插件包被聚合器（`internal/plugins/plugins.go`）空导入，入口（`cmd/musicfox.go`）空导入聚合器。TUI 前端 scope 构建时（`ui.NewFrontendScope`）按 `configs.IsPluginEnabled(id)` 过滤、以 `AddWithEnabled` 挂载插件；`Start` 在启动完整性断言（`AssertMenuRegistryComplete` / `AssertPageRegistryComplete`）之前完成。
4. 启动断言锁定内置注册清单；插件 key 不得与内置 key 冲突（`expectedMenuKeys` 不含插件 key）。

**明确不支持（spec Non-goals）**：

- **运行时动态加载**已部分落地：**WASM 插件**（`internal/wasm` + `examples/wasm/hello`，见「WASM 插件（实验性）」）支持用户**不重编译**地安装/启用/禁用插件（MVP：菜单动作 + 文本结果）。Go `plugin` 包 / 共享库热加载仍不支持。
- 不改造 tea 消息循环、不引入外部 DI 框架、不修改 vendored SDK。

**当前边界注意**：

- `BaseMenu` 已导出，注册闭包可用 `BaseMenu` 书写（`baseMenu` 是别名，二者等价），插件菜单类型嵌入 `ui.BaseMenu` 即可——**可以在 `ui` 包之外实现与注册**（编译期边界验证见 `internal/ui/plugin_boundary_external_test.go`，`package ui_test`；首个真实插件 `internal/plugins/checkupdate` 即此形态）。
- 插件经聚合器接入：`internal/plugins/plugins.go` 空导入各插件包，`cmd/musicfox.go` 空导入聚合器。插件 key 不得与内置 key 冲突；`expectedMenuKeys` / `expectedPageKeys` 不包含插件 key（完整性断言只校验内置清单，`check_update` / `last_fm` / `lastfm_auth` / `lastfm_custom_api`、整个 DJ/电台集群的 `dj_*` / `radio_dj_type`、整个专辑集群的 `album_menu` / `album_*` / `album_detail`、整个歌手集群的 `artist_detail` / `artist_*` / `hot_artists` / `artists_sub_list`、整个推荐集群的 `daily_songs` / `daily_playlists` / `personal_fm` / `recent_songs` / `ranks`、整个歌单/云盘集群的 `user_playlist` / `user_collect` / `high_quality_playlists` / `could` / `playlist_detail`、整个搜索集群的 `search_type` / `search_result` / `search` 页以及整个单曲集群的 `simi_songs` / `add_to_user_playlist` 由插件注册即可通过）。
- `internal/ui` 仍是 internal 包：按 Go internal 规则，插件代码必须位于 go-musicfox 模块树内（如 `internal/plugins/checkupdate`）才能导入；独立仓库插件需 `replace` 到模块树内或以子包形式落地，纯外部模块直接导入 `internal/ui` 仍被 Go 拒绝（属预留边界）。
- `Netease` 薄壳仍未导出：外部插件经 `base.BaseMenu.Netease()` 逃生口访问，不应直接构造。

### 插件配置化启停：WithPlugin + `framework.RegisterPlugin` / `[plugins] disabled`

每个插件以 cordis 形态接入：包 `init()` 只注册构造器，`Start` 把全部注册包进 **`ui.WithPlugin(id, name, register)`**（`internal/ui/plugin_registry.go`；id 为插件目录名，同 id 可多文件多次声明幂等合并），使作用域内的 `RegisterMenu` / `RegisterPage` / `RegisterMainMenuItem*` / `RegisterStartupHook` / `RegisterCommand` 归属到该插件：

```go
// 文件：internal/plugins/example/registry.go（P5 cordis 形态）
// Plugin 嵌入 NoopPlugin（生命周期零样板）；Start 内完成全部注册。
type Plugin struct {
	framework.NoopPlugin
}

func (p *Plugin) Start(_ *framework.Context) error {
	ui.WithPlugin("example", "示例", func() {
		ui.RegisterMenu("example", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
			return NewExampleMenu(base), nil
		})
		ui.RegisterMainMenuItemAfter("example", "示例", ui.MainMenuStart, nil)
	})
	return nil
}

// 声明期只注册构造器；实际注册在 Start（前端 scope 挂载时执行）。
func init() {
	framework.RegisterPlugin("example", func() framework.Plugin { return &Plugin{} })
}
```

用户经配置 `[plugins] disabled = ["search", "checkupdate"]` 禁用插件后（`configs.PluginsConfig.Disabled`，空配置视为全部启用）：

- **禁用 = 不存在**（P5 契约切换）：禁用插件**不 Start**，其菜单/页面/主菜单项/启动钩子/命令**全部不注册**——"按 key 跳入禁用插件菜单"不再可能（`BuildMenu` 报缺失 key，跳转点经 `buildMenuOrToast` 降级 toast）。这是全量 cordis 化后生命周期自洽的形态。
- **锚点完整性**：被禁用插件的主菜单项从不进入 `NewMainMenu` 的 After 锚点链（因为不注册），但其它项仍可声明以它为锚点——锚点存在性断言基于**注册序**校验，禁用插件的项不在注册集中，后继项自然前移、不 panic。
- **残余消费点过滤**（非禁用=不存在的注册）：WASM 命令菜单项由 `registerCommandMenus` 对**每个已加载 manifest** 无条件适配，经 `IsPluginEnabled` 在 `NewMainMenu` 显示与 `commandActionCmd` 执行两层过滤；`framework.RunStartupHooks` 同样按插件 id 过滤包级钩子。

查询 API：`ui.PluginInfos()`（自 P8 从**前端 scope**收集实际启用插件集——插件须暴露 `framework.PluginIdentity`，9 个业务插件经 ui 的 `identifiedPlugin` 装饰器携带 id，WASM 适配器自身实现；含归属的菜单/页面/主菜单项 key 与启动钩子数）、`ui.IsPluginEnabled(id)`（读 `configs.AppConfig.Plugins`，nil 配置返回 true）。不在任何 `WithPlugin` 作用域内的注册（内置 bootstrap、测试二进制）为空归属，不受启停过滤。

## 工作示例

> 以下代码与 `internal/ui/plugin_boundary_external_test.go`（包外编译校验）对应；示例一/四/五/六/七/八/九/十/十一为已合入仓库的真实插件（`internal/plugins/checkupdate` / `internal/plugins/lastfm` / `internal/plugins/dj` / `internal/plugins/album` / `internal/plugins/artist` / `internal/plugins/recommend` / `internal/plugins/playlist` / `internal/plugins/search` / `internal/plugins/song`），示例二/三为最小演示形态。
>
> 注（P5 cordis 化）：示例中的 `func init() { ui.RegisterMenu(...) }` 为**注册调用演示**（各 `RegisterXxx` 签名与行为不变）；真实插件的实际接线是「`init()` 只 `framework.RegisterPlugin` + `Start` 内 `ui.WithPlugin` 包裹这些注册」（见示例一与「插件配置化启停」），禁用语义相应变为「禁用 = 不存在」。

### 示例一：检查更新插件（首个真实提取示例）

> `internal/plugins/checkupdate` 是把内置「检查更新」菜单提取为外部式插件的第一个真实插件，完整走通编译期插件链路：`BaseMenu` 嵌入 + `RegisterMenu` 注册 + `RegisterMainMenuItemAfter` 主菜单入口 + `RegisterStartupHook` 启动钩子 + 聚合器 + 空导入。菜单 key 仍为 `check_update`，行为与提取前一致。

`menu.go`（菜单类型 + 检查/通知逻辑 + 进入钩子）：

```go
// 文件：internal/plugins/checkupdate/menu.go
package checkupdate

import (
	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"

	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// CheckUpdateMenu 嵌入导出基座 ui.BaseMenu（不是未导出的 baseMenu），
// 在 ui 包之外即可实现 ui.Menu 接口。
type CheckUpdateMenu struct {
	ui.BaseMenu
}

func (m *CheckUpdateMenu) GetMenuKey() string { return "check_update" }
func (m *CheckUpdateMenu) IsPlayable() bool   { return false }
func (m *CheckUpdateMenu) IsLocatable() bool  { return false }

// MenuViews 检查更新为动作菜单；返回一个静态项，不会作为子菜单被渲染。
func (m *CheckUpdateMenu) MenuViews() []model.MenuItem {
	return []model.MenuItem{{Title: "检查更新"}}
}

func (m *CheckUpdateMenu) SubMenu(_ *model.App, _ int) model.Menu { return nil }

// BeforeEnterMenuHook 进入即触发检查并弹回主页面（单次 Enter 完成检查，
// 等价于提取前 main_menu 的 index-15 特判）。检查在后台 goroutine 执行，
// 结果经 app.Notify 安全投递。
func (m *CheckUpdateMenu) BeforeEnterMenuHook() model.Hook {
	return func(main *model.Main) (bool, model.Page) {
		go func() {
			hasUpdate, latestVersion := version.CheckUpdate()
			if app := m.App(); app != nil {
				app.Notify(checkUpdateNotificationSpec(hasUpdate, latestVersion))
			}
		}()
		return false, main
	}
}

// Action 触发检查更新并留在当前页面（非 nil 的 page/cmd 会跳过子菜单导航）。
func (m *CheckUpdateMenu) Action(a *model.App, _ int) (model.Page, tea.Cmd) {
	return a.MustMain(), checkUpdateCmd()
}
// ... checkUpdateNotificationSpec / checkUpdateNotificationMsg /
// newVersionNotifyContent 与提取前逐字一致；「发现新版本」toast 经
// ui.BuildToastNotificationSpec 构建（ui 导出的 toast-spec 助手）。
```

`registry.go`（P5 cordis 形态：`init()` 只注册构造器，`Start` 完成全部注册）：

```go
// 文件：internal/plugins/checkupdate/registry.go
// Plugin 嵌入 NoopPlugin（Start/Stop/Dispose 零样板）；Start 经 WithPlugin
// 作用域归属盖章后注册全部贡献。
type Plugin struct {
	framework.NoopPlugin
}

func (p *Plugin) Start(_ *framework.Context) error {
	ui.WithPlugin("checkupdate", "检查更新", func() {
		ui.RegisterMenu("check_update", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
			return &CheckUpdateMenu{BaseMenu: base}, nil
		})
		// 声明主菜单入口：NewMainMenu 经 After 锚点链归并——检查更新跟在帮助
		// （内置项）之后，位于主菜单链尾（复现插件化前顺序）。
		ui.RegisterMainMenuItemAfter("check_update", "检查更新", "help", nil)
		// 注册启动自动检查（原 shell 级硬编码启动检查，见 startup.go）。
		ui.RegisterStartupHook(startupCheck)
	})
	return nil
}

// init() 是编译期注册入口：只声明插件构造器（实际注册在 Start，前端 scope
// 挂载时执行）。
func init() {
	framework.RegisterPlugin("checkupdate", func() framework.Plugin { return &Plugin{} })
}
```

`startup.go`（启动钩子：配置门控的自动检查）：

```go
// 文件：internal/plugins/checkupdate/startup.go
func startupCheck() {
	if !configs.AppConfig.Startup.CheckUpdate {
		return
	}
	if ok, newVersion := version.CheckUpdate(); ok {
		notify.Notify(notify.NotifyContent{
			Title:       "发现新版本: " + newVersion,
			Text:        "去看看呗",
			Url:         types.AppGithubUrl + "/releases/tag/" + newVersion,
			ActionLabel: "前往 GitHub",
		})
	}
}
```

聚合器 + 空导入（插件被链接进二进制的入口）：

```go
// 文件：internal/plugins/plugins.go（聚合器：空导入每个插件包，触发其 init()）
package plugins

import (
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/album"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/artist"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/checkupdate"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/dj"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/lastfm"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/playlist"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/recommend"
)

// 文件：cmd/musicfox.go（入口空导入聚合器）
import (
	_ "github.com/go-musicfox/go-musicfox/internal/plugins"
)
```

主菜单不再对「检查更新」做索引特判（`mainMenuCheckUpdateIndex` 已随插件化移除）：`NewMainMenu` 读取 `MainMenuPluginItems()`，按 After 锚点链归并插件与内置入口并按 key 构建菜单；选中后进入插件菜单，由插件自身的 `BeforeEnterMenuHook` / `Action` 承担检查与通知。

**启动钩子调用点**：核心引擎 `core.Engine.Startup`（`internal/core/startup.go`，前端无关）在用户/登录恢复之后、自动播放之前的位置调用 `framework.RunStartupHooks(configs.IsPluginEnabled)`（启动序第 10 步；即原 shell 级启动自动检查所在位置）。此时 services 已注册、用户/登录已就绪；每个 hook 带 recover 隔离，panic 仅记日志不阻断启动，被禁用插件的 hook 被 `configs.IsPluginEnabled` 过滤。TUI 与 headless 两种前端都执行同一序列。

### 示例二：hello 菜单插件（静态菜单项）

```go
// 文件：plugins/example_hello/plugin.go（go-musicfox 模块树内的插件子包）
package example_hello

import (
	"github.com/anhoder/foxful-cli/model"

	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// ExampleHelloMenu 展示一个静态菜单项的示例插件菜单：嵌入 ui.BaseMenu，
// 无需在外部包中接触未导出类型。
type ExampleHelloMenu struct {
	ui.BaseMenu
	menus []model.MenuItem
}

func (m *ExampleHelloMenu) GetMenuKey() string { return "example_hello" }

func (m *ExampleHelloMenu) MenuViews() []model.MenuItem { return m.menus }

// 静态菜单不产生子菜单。
func (m *ExampleHelloMenu) SubMenu(_ *model.App, _ int) model.Menu { return nil }

func (m *ExampleHelloMenu) BeforeEnterMenuHook() model.Hook {
	return func(_ *model.Main) (bool, model.Page) {
		m.menus = []model.MenuItem{
			{Title: "你好，插件世界！", Subtitle: "这是静态菜单项 A"},
			{Title: "插件菜单项 B", Subtitle: "不触发任何跳转"},
		}
		return true, nil
	}
}

// init() 是编译期注册入口：包被链接进二进制后自动执行。factory 的 base
// 参数类型写成 ui.BaseMenu（注册签名用其别名 baseMenu，二者可互换）。
func init() {
	ui.RegisterMenu("example_hello", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return &ExampleHelloMenu{BaseMenu: base}, nil
	})
}
```

在任一父菜单的 `SubMenu` 中经注册表跳转：

```go
func (m *SomeParentMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index != 0 {
		return nil
	}
	helloMenu, err := BuildMenu("example_hello", m.baseMenu, NoArgMenuOpts{})
	if err != nil {
		return nil // 需要提示时改用 buildMenuOrToast("example_hello", m.baseMenu, NoArgMenuOpts{})
	}
	return helloMenu
}
```

### 示例三：页面插件

```go
// 文件：plugins/example_hello/page.go（插件子包）
package example_hello

import (
	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"

	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// ExampleHelloPageOpts 是页面插件的参数契约（可按需携带任意字段）。
type ExampleHelloPageOpts struct{}

// ExampleHelloPage 展示一个最小页面插件。
type ExampleHelloPage struct{}

func (p *ExampleHelloPage) IgnoreQuitKeyMsg(tea.KeyMsg) bool { return false }

func (p *ExampleHelloPage) Type() model.PageType { return "example_hello_page" }

func (p *ExampleHelloPage) Update(tea.Msg, *model.App) (model.Page, tea.Cmd) { return p, nil }

func (p *ExampleHelloPage) View(*model.App) string { return "你好，插件页面！" }

func (p *ExampleHelloPage) Msg() tea.Msg { return nil }

func init() {
	ui.RegisterPage("example_hello_page", func(_ ExampleHelloPageOpts) (model.Page, error) {
		return &ExampleHelloPage{}, nil
	})
}
```

页面导航经 `BuildPage` / `buildPageOrToast`：

```go
page, err := ui.BuildPage("example_hello_page", ExampleHelloPageOpts{})
```

### 示例四：Last.fm 插件（服务访问 + 页面插件 + 主菜单项）

> `internal/plugins/lastfm` 是第二个真实插件，把内置 Last.fm 菜单/页面整体提取为外部式插件。它补全了示例一的插件能力矩阵：**服务访问**（`svc.Lastfm()`——访问器解析 `ServiceLastfm` 客户端）、**页面插件**（`lastfm_auth` / `lastfm_custom_api` 经 `RegisterPage` 注册，opts 携带 `ui.MenuServices`）、**主菜单项**（`RegisterMainMenuItemAfter("last_fm", "LastFM", "radio_dj_type", nil)`——原内置入口已从 `menu_main.go` 移除，经 After 锚点回到原内置位置）。页面通过 `ui.BuildPageOrToast` 打开，渲染复用 ui 导出的自定义页面布局助手。

```go
// 文件：internal/plugins/lastfm/page_auth.go（节选）——页面 opts 携带访问器
// LastfmAuthPageOpts / LastfmCustomAPIPageOpts 是页面插件的参数契约：
type LastfmAuthPageOpts struct {
	Svc ui.MenuServices // 访问器：页面经 svc.MustMain()/svc.App()/svc.Lastfm() 解析 shell 与服务
}

// 文件：internal/plugins/lastfm/registry.go——编译期注册入口
func init() {
	ui.RegisterMenu("last_fm", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return NewLastfm(base), nil
	})
	ui.RegisterPage("lastfm_auth", func(opts LastfmAuthPageOpts) (model.Page, error) {
		return NewLastfmAuthPage(opts.Svc), nil
	})
	ui.RegisterPage("lastfm_custom_api", func(opts LastfmCustomAPIPageOpts) (model.Page, error) {
		return NewLastfmCustomAPIPage(opts.Svc), nil
	})
	// 主菜单项：NewMainMenu 经 After 锚点链归并——LastFM 跟在主播电台
	// （dj 插件）之后、帮助（内置）之前（复现插件化前顺序）。
	ui.RegisterMainMenuItemAfter("last_fm", "LastFM", "radio_dj_type", nil)
}

// 文件：internal/plugins/lastfm/profile.go（节选）——菜单内打开页面：
// 访问器经 base.Services() 传给页面 opts，失败经 ui.BuildPageOrToast 降级。
page := ui.BuildPageOrToast("lastfm_custom_api", LastfmCustomAPIPageOpts{Svc: m.Services()})
if page == nil {
	return nil
}
return ui.NewMenuToPage(m.BaseMenu, page, m.CoverRenderer().ClearDisplayed)
```

Last.fm 菜单/页面的服务访问全部走访问器（`m.Lastfm()` / `svc.Lastfm()`），不触碰任何未导出 ui 符号——这正是插件边界的目标形态。

### 示例五：DJ / 电台集群（批量菜单提取）

> `internal/plugins/dj` 是第三个真实插件：把「主播电台」整个集群（10 个菜单）整体提取为外部式插件，是最大的可见集群批量提取示范。所有 provider key 与提取前**逐一相同**（`dj_radio_detail` / `dj_category_detail` / `dj_category` / `dj_program_rank` / `dj_program_hour_rank` / `dj_hot` / `dj_sub` / `dj_recommend` / `dj_today_recommend` / `radio_dj_type`），因此 ui 侧跳入集群的调用点（如 `menu_search_result.go` 跳 `dj_radio_detail`）无需改动。集群内部跨菜单跳转（子菜单互相跳转）改走导出的 `ui.BuildMenuOrToast` / `ui.MustBuild` / `ui.MustBuildNoArg`，行为与提取前的 `buildMenuOrToast` / `mustBuild` / `mustBuildNoArg` 逐一等价：

```go
// 文件：internal/plugins/dj/menu_dj_recommend.go（节选）——集群内跳转
func (m *DjRecommendMenu) SubMenu(_ *model.App, index int) model.Menu {
	if index >= len(m.radios) {
		return nil
	}
	return ui.BuildMenuOrToast("dj_radio_detail", m.BaseMenu, ui.DjRadioDetailOpts{DjRadioID: m.radios[index].Id})
}

// 文件：internal/plugins/dj/registry.go（节选）——入口菜单声明主菜单项
func init() {
	// ... RegisterMenu("dj_recommend", ...) 等 10 个注册，key 与原内置注册一致
	// 主播电台主菜单入口：声明前驱项 key（云盘，playlist 插件）回到原内置位置。
	ui.RegisterMainMenuItemAfter("radio_dj_type", "主播电台", "could", nil)
}
```

集群内共享的 opts 契约类型（`DjCategoryDetailOpts` / `DjHotOpts`，仅集群内部使用）随菜单移入插件包；被 ui 侧共享的 `DjRadioDetailOpts` 留在 ui（跳入该菜单的调用点在 ui）。参数化菜单的 `GetMenuKey()` 仍返回动态形式（如 `dj_radio_detail_<id>`），注册 key 保持静态前缀。集群中唯一被 ui 反向引用的能力是「排序切换」（`OpToggleSortOrder` 键处理对 `DjRadioDetailMenu` 做类型断言）——ui 不能反向导入插件包，故改经导出接口 `ui.DjRadioDetailSortable`（`ToggleSortOrder` / `Reload`）访问，插件菜单实现该接口即保持键处理行为不变。

### 示例六：专辑集群（批量菜单提取 + 主菜单项）

> `internal/plugins/album` 是第四个真实插件：把「专辑列表」整个集群（8 个菜单）整体提取为外部式插件，与示例五同为集群批量提取示范。所有 provider key 与提取前**逐一相同**（`album_menu` / `album_new_area` / `album_top_area` / `album_new_hot` / `album_new` / `album_top` / `album_sub_list` / `album_detail`），因此 ui 侧跳入集群的调用点（`menu_search_result.go` / `menu_artist_album.go` 跳 `album_detail`、`menu_user_collection.go` 构建 `album_sub_list` 子菜单）全部经注册表按 key 跳转、无需改动。`album_menu`（专辑列表入口菜单）声明主菜单入口「专辑列表」（经 After 锚点声明前驱项 key = 私人FM，回到插件化前的原始位置）：

```go
// 文件：internal/plugins/album/registry.go（节选）——入口菜单声明主菜单项
func init() {
	// ... RegisterMenu("album_top", ...) 等 8 个注册，key 与原内置注册一致
	// 专辑列表主菜单入口：声明前驱项 key（私人FM，recommend 插件）回到原内置位置。
	ui.RegisterMainMenuItemAfter("album_menu", "专辑列表", "personal_fm", nil)
}
```

集群内共享的 opts 契约（`AlbumTopOpts` / `AlbumNewOpts`，仅集群内部使用）随菜单移入插件包；被 ui 侧共享的 `AlbumDetailOpts` 留在 ui（`menu_search_result.go` / `menu_artist_album.go` / `operate.go` 均跳 `album_detail`）。参数化菜单的 `GetMenuKey()` 仍返回动态形式（如 `album_top_<area>` / `album_new_<area>`），注册 key 保持静态前缀。集群中唯一被 ui 反向引用的能力是「查看歌曲所属专辑的去重判断」（`operate.go` 对 `AlbumDetailMenu` 做类型断言取 `albumID`）——ui 不能反向导入插件包，故改经导出接口 `ui.AlbumDetailIDGetter`（`AlbumID() int64`）访问，插件菜单实现该接口即保持行为不变。

### 示例七：歌手集群（第五个真实插件）

> `internal/plugins/artist` 是第五个真实插件：把「热门歌手 / 歌手详情」整个集群（6 个菜单）整体提取为外部式插件，与示例五/六同为集群批量提取示范。所有 provider key 与提取前**逐一相同**（`hot_artists` / `artist_detail` / `artist_song` / `artist_album` / `artist_of_song` / `artists_sub_list`），因此 ui 侧跳入集群的调用点（`menu_search_result.go` / `operate.go` 跳 `artist_detail`、`operate.go` 跳 `artist_of_song`、`menu_user_collection.go` 构建 `artists_sub_list` 子菜单）全部经注册表按 key 跳转、无需改动。`hot_artists`（热门歌手入口菜单）声明主菜单入口「热门歌手」（经 After 锚点声明前驱项 key = 精选歌单，回到插件化前的原始位置）：

```go
// 文件：internal/plugins/artist/registry.go（节选）——入口菜单声明主菜单项
func init() {
	// ... RegisterMenu("hot_artists", ...) 等 6 个注册，key 与原内置注册一致
	// 热门歌手主菜单入口：声明前驱项 key（精选歌单，playlist 插件）回到原内置位置。
	ui.RegisterMainMenuItemAfter("hot_artists", "热门歌手", "high_quality_playlists", nil)
}
```

集群内共享的 opts 契约（`ArtistAlbumOpts` / `ArtistSongOpts`，仅集群内部使用）随菜单移入插件包；被 ui 侧共享的 `ArtistDetailOpts` / `ArtistsOfSongOpts` 留在 ui（`menu_search_result.go` / `operate.go` 均跳 `artist_detail`，operate.go 的 `goToArtistOfSong` 以 `ArtistsOfSongOpts` 携带歌曲载荷）。参数化菜单的 `GetMenuKey()` 仍返回动态形式（如 `artist_detail_<id>` / `artist_song_<id>` / `artist_album_<id>`），注册 key 保持静态前缀。集群中被 ui 反向引用的能力是「查看歌曲所属歌手的去重判断」（`operate.go` 对 `ArtistDetailMenu` / `ArtistsOfSongMenu` 做类型断言取 id）——ui 不能反向导入插件包，故改经导出接口 `ui.ArtistDetailIDGetter`（`ArtistID() int64`）与 `ui.ArtistsOfSongSongIDGetter`（`SongID() int64`）访问，插件菜单实现该接口即保持行为不变。

### 示例八：推荐集群（第六个真实插件 + 批量主菜单项）

> `internal/plugins/recommend` 是第六个真实插件：把「推荐/播放历史」集群（每日推荐歌曲、每日推荐歌单、私人FM、最近播放歌曲、排行榜）5 个菜单整体提取为外部式插件，与示例五/六/七同为集群批量提取示范，且是**一次声明 5 个主菜单项**的示范。所有 provider key 与提取前**逐一相同**（`daily_songs` / `daily_playlists` / `personal_fm` / `recent_songs` / `ranks`），ui 侧跳入集群的调用点零改动；集群内跳转（`ranks` / `daily_playlists` 的 SubMenu 跳 `playlist_detail`）经 `ui.BuildMenu` 按 key 解析（`playlist_detail` 由 playlist 插件注册）。5 个入口菜单各自声明主菜单项「每日推荐歌曲 / 每日推荐歌单 / 私人FM / 最近播放歌曲 / 排行榜」（经 After 锚点声明各自的前驱项 key，回到插件化前的原始位置）：

```go
// 文件：internal/plugins/recommend/registry.go（节选）——入口菜单声明主菜单项
func init() {
	// ... RegisterMenu("daily_songs", ...) 等 5 个注册，key 与原内置注册一致
	// 5 个入口各声明主菜单入口（前驱项 key）：每日推荐歌曲跟链首（MainMenuStart）、
	// 每日推荐歌单跟每日推荐歌曲、私人FM 跟我的收藏（playlist 插件）、排行榜跟搜索
	// （search 插件）、最近播放歌曲跟热门歌手（artist 插件）。
	ui.RegisterMainMenuItemAfter("daily_songs", "每日推荐歌曲", ui.MainMenuStart, nil)
	ui.RegisterMainMenuItemAfter("daily_playlists", "每日推荐歌单", "daily_songs", nil)
	ui.RegisterMainMenuItemAfter("personal_fm", "私人FM", "user_collect", nil)
	ui.RegisterMainMenuItemAfter("recent_songs", "最近播放歌曲", "hot_artists", nil)
	ui.RegisterMainMenuItemAfter("ranks", "排行榜", "search_type", nil)
}
```

本集群还是**登录门控 + 播放器服务经访问器**的示范：`daily_songs` / `daily_playlists` / `recent_songs` 的 `BeforeEnterMenuHook` 经 `m.User()` / `m.ToLoginPage(enterMenuCallback(main))` 走 `ui.BaseMenu` 转发（`enterMenuCallback` 在插件内镜像 ui 未导出的 `EnterMenuCallback`）；`personal_fm` 的 `BottomOutHook` 经 `m.Player()` 更新播放列表（`ReinitializePlaylist` / `MarkPlaylistUpdated`，Phase 3.6 播放列表 API）。

### 示例九：歌单/云盘集群（第七个真实插件 + 参数化主菜单项）

> `internal/plugins/playlist` 是第七个真实插件：把「歌单/云盘」集群（我的歌单、我的收藏、精选歌单、云盘）4 个菜单整体提取为外部式插件，`playlist_detail`（歌单详情）随后并入（示例九补全），与示例五/六/七/八同为集群批量提取示范，且是**参数化主菜单项**（`RegisterMainMenuItemWith`，Phase 3.9.9 机制）的生产示范。所有 provider key 与提取前**逐一相同**（`user_playlist` / `user_collect` / `high_quality_playlists` / `could` / `playlist_detail`），ui 侧跳入集群的调用点零改动（`operate.go` 仍按 key 跳 `user_playlist` / `playlist_detail`，`UserPlaylistOpts` / `PlaylistDetailOpts` 留在 ui）；集群内跳转（`user_playlist` / `high_quality_playlists` 的 SubMenu 跳 `playlist_detail`）经 `ui.BuildMenu` 按 key 解析。4 个入口菜单各自声明主菜单项「我的歌单 / 我的收藏 / 精选歌单 / 云盘」（经 After 锚点声明各自的前驱项 key，回到插件化前的原始位置），`playlist_detail` 为纯跳转目标不声明主菜单项。`user_playlist` 是参数化 provider（`ui.UserPlaylistOpts` 携带 userID），其主菜单入口经 builder 以 `UserID: ui.CurUser` 构造——与内置入口行为一致（`ui.CurUser` 常量留在 ui，song 插件仍引用）：

```go
// 文件：internal/plugins/playlist/registry.go（节选）——参数化主菜单项 + 无参主菜单项
func init() {
	// ... RegisterMenu("user_playlist", ...) 等 4 个注册，key 与原内置注册一致
	// 4 个入口各声明主菜单入口（前驱项 key）：我的歌单跟每日推荐歌单（recommend
	// 插件）、我的收藏跟我的歌单、精选歌单跟排行榜（recommend 插件）、云盘跟最近
	// 播放歌曲（recommend 插件）。user_playlist 经参数化 builder 构造（UserID =
	// ui.CurUser，与内置入口行为一致）。
	ui.RegisterMainMenuItemAfter("user_playlist", "我的歌单", "daily_playlists", func(base ui.BaseMenu) ui.Menu {
		return ui.MustBuild("user_playlist", base, ui.UserPlaylistOpts{UserID: ui.CurUser})
	})
	ui.RegisterMainMenuItemAfter("user_collect", "我的收藏", "user_playlist", nil)
	ui.RegisterMainMenuItemAfter("high_quality_playlists", "精选歌单", "ranks", nil)
	ui.RegisterMainMenuItemAfter("could", "云盘", "recent_songs", nil)
}
```

`user_collect` 的子菜单（收藏专辑 / 收藏歌手）经 `ui.MustBuildNoArg` 按 `album_sub_list` / `artists_sub_list` key 构建——这两个 key 由 album / artist 插件注册，构成**跨插件按 key 协作**（插件之间也通过注册表交互，不 import 彼此）。集群中唯一被 ui 反向引用的共享符号是 `ui.UserPlaylistOpts`（`operate.go` 跳 `user_playlist` 时携带用户 ID），其具体类型 `UserPlaylistMenu` 随集群移入插件。

### 示例十：搜索集群（第八个真实插件 + 页面注册转发）

> `internal/plugins/search` 是第八个真实插件：把「搜索」集群（搜索类型、搜索结果 2 个菜单 + 搜索页注册）整体提取为外部式插件，且是**页面注册转发**示范——`SearchPage` 类型与 wordsInput/result/searchType 状态是 shell 持有的单例（`operate.go` 与 `SearchResultMenu` 共享），页面类型留在 ui，插件只做注册转发。所有 provider key 与提取前**逐一相同**（`search_type` / `search_result`，页面 key `search`）；`search_type` 声明主菜单入口「搜索」（After 锚点 `album_menu`——album 插件，保持插件化前的原位置），内置「搜索」入口随之从 `menu_main.go` 移除。`SearchType` / `St*` 常量与共享 opts（`SearchResultOpts` / `SearchPageOpts`）留在 ui，`SearchResultMenu` 经 `ui` 为 `SearchPage` 新增的导出访问器（`WordsInput()` / `Result()` / `SearchType()`）读取共享状态；ui 侧 `operate.go` 跳 `search_result` 的调用点零改动：

```go
// 文件：internal/plugins/search/registry.go（节选）——菜单 + 页面注册转发 + 主菜单项
func init() {
	ui.RegisterMenu("search_type", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return NewSearchTypeMenu(base), nil
	})
	ui.RegisterMenu("search_result", func(base ui.BaseMenu, opts ui.SearchResultOpts) (ui.Menu, error) {
		return NewSearchResultMenu(base, opts.SearchType), nil
	})
	// 页面类型留在 ui（shell 单例，状态与 operate / SearchResultMenu 共享），
	// 插件只做注册转发——shell 在 NewNetease 经 BuildPage("search") 构建单例。
	ui.RegisterPage("search", func(opts ui.SearchPageOpts) (model.Page, error) {
		return ui.NewSearchPage(opts.Netease), nil
	})
	// 「搜索」主菜单入口：跟专辑列表（album 插件）后，保持插件化前的原位置。
	ui.RegisterMainMenuItemAfter("search_type", "搜索", "album_menu", nil)
}
```

### 示例十一：单曲集群（第九个真实插件 + 纯跳转目标）

> `internal/plugins/song` 是第九个真实插件：把「单曲」集群（相似歌曲、添加到歌单 2 个菜单）整体提取为外部式插件，均为**参数化纯跳转目标**（不声明主菜单项，从右键菜单/搜索页跳入）。所有 provider key 与提取前**逐一相同**（`simi_songs` / `add_to_user_playlist`）；共享符号留在 ui：`SimiSongsOpts` / `AddToUserPlaylistOpts`（`operate.go` 使用）与 `ui.CurUser` 常量（`AddToUserPlaylistMenu` 以 `userID == CurUser` 解析当前登录用户）。ui 侧经**接口**访问插件具体类型（`SimilarSongsRelateSongIDGetter` / `AddToUserPlaylistGetter`），不直接引用插件包：

```go
// 文件：internal/plugins/song/registry.go（节选）——两个参数化纯跳转目标
func init() {
	ui.RegisterMenu("simi_songs", func(base ui.BaseMenu, opts ui.SimiSongsOpts) (ui.Menu, error) {
		return NewSimilarSongsMenu(base, opts.Song), nil
	})
	ui.RegisterMenu("add_to_user_playlist", func(base ui.BaseMenu, opts ui.AddToUserPlaylistOpts) (ui.Menu, error) {
		return NewAddToUserPlaylistMenu(base, opts.UserID, opts.Song, opts.IsAdd), nil
	})
}
```

### 接入二进制（聚合器）

正式插件的标准接入方式：每个插件子包被聚合器空导入（见示例一），入口只需空导入聚合器即可，不必逐个列插件：

```go
// 文件：internal/plugins/plugins.go —— 聚合器：空导入每个插件包
package plugins

import (
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/album"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/artist"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/checkupdate"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/dj"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/lastfm"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/playlist"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/recommend"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/search"
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/song"
)

// cmd/musicfox.go 或入口处空导入聚合器，触发全部插件 init() 注册构造器：
import (
	_ "github.com/go-musicfox/go-musicfox/internal/plugins"
)
```

> 注：插件子包必须位于 go-musicfox 模块树内（`internal` 导入规则）；`BaseMenu` 已导出、`baseMenu` 是其别名，注册闭包可在 `ui` 包外以 `ui.BaseMenu` 类型书写。`ui` 不得反向导入插件包（import cycle），shell 需要插件能力时通过注册机制调用：插件能力经 `RegisterMainMenuItem[After]` / `RegisterStartupHook` 声明，由 shell 在构建主菜单 / `InitHook` 时统一消费（不再需要内联插件逻辑）。

### 插入一个主菜单项（after-anchor 单锚点改动）

主菜单顺序由 **after-anchor 链**驱动：每个入口声明其前驱项 key（`After`），`NewMainMenu` 从 `ui.MainMenuStart` 沿链走序。**插入一个菜单 = 声明一个锚点，其余项不漂移**——不再像数值 Order 那样需要重排所有后续项：

```go
// 现状：新插件菜单「我的电台」想插在「云盘」之后（其当前后继是「主播电台」）。
// 改动一：把「主播电台」的锚点从 could 改为 my_radio（唯一需要改的既有项）。
func init() {
	ui.RegisterMenu("my_radio", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return NewMyRadioMenu(base), nil
	})
	ui.RegisterMainMenuItemAfter("my_radio", "我的电台", "could", nil) // 新项：跟在云盘后
}

// 文件：internal/plugins/dj/registry.go（既有插件，唯一改动行）
ui.RegisterMainMenuItemAfter("radio_dj_type", "主播电台", "my_radio", nil) // 锚点改为新项

// 插入后主菜单片段：… → 云盘 → 我的电台 → 主播电台 → LastFM → …，
// 其余所有项（daily_songs … could / last_fm … check_update）的锚点与位置零改动。
```

追加在链尾的简单场景则完全是**单锚点零改动**：新插件只声明 `RegisterMainMenuItemAfter("new_menu", "标题", "check_update", nil)`（跟在检查更新后）或便捷形式 `RegisterMainMenuItem("new_menu", "标题")`（末尾追加）。链完整性由 `NewMainMenu` 启动断言（After 目标存在 / 每项恰好可达 / 无孤儿环），插错锚点在启动时即 panic 报错。

## 插件脚手架（hack/new-plugin）

`hack/new-plugin` 是一个 Go 生成器（标准库，跨平台），帮助开发者快速生成新插件骨架——直接从上面的工作示例形态起步，替换占位符后即可编译。

**位置**：

- 生成器：`hack/new-plugin/main.go`
- 模板：`hack/new-plugin/templates/plugin_name/`（`registry.go.tmpl` / `menu.go.tmpl` / `plugin_name_test.go.tmpl` / `README.md.tmpl`，占位符 `{{plugin_name}}` / `{{MenuType}}` / `{{menu_key}}` / `{{menu_title}}` / `{{menu_after}}`）

**生成命令**（在仓库根执行）：

```bash
go run ./hack/new-plugin -name example -menu Example -key example -title 示例 -after MainMenuStart
```

| Flag | 说明 | 默认 |
|------|------|------|
| `-name` | 插件名：包名与目录名（小写 snake_case，**必填**） | — |
| `-menu` | 菜单类型名前缀（CamelCase，如 `Example` → `ExampleMenu`，**必填**） | — |
| `-key` | 菜单注册 key（全局唯一） | = `-name` |
| `-title` | 主菜单显示标题 | = `-menu` |
| `-after` | 主菜单 after 锚点（`MainMenuStart` = 主菜单链首） | `MainMenuStart` |
| `-dir` | 输出插件目录（相对当前工作目录） | `internal/plugins` |
| `-templates` | 模板目录（相对当前工作目录） | `hack/new-plugin/templates` |
| `-force` | 目标目录已存在时强制覆盖 | 关闭（已存在即拒绝） |
| `-skip-fmt` | 跳过 `gofmt -w`（默认自动格式化生成的 `.go` 文件） | 关闭 |

**生成结果**：`<dir>/<name>/{registry.go, menu.go, <name>_test.go, README.md}`。其中：

- `registry.go`——编译期注册入口（`ui.RegisterMenu` 无参菜单 + `ui.RegisterMainMenuItemAfter` 主菜单入口），并附**可选扩展的注释示例**（按需取消注释）：启动钩子（`ui.RegisterStartupHook`）、右键菜单项（`ui.RegisterContextMenuContrib`，见「右键菜单项」）、状态栏组件（`ui.RegisterStatusBarComponent`，见「状态栏组件」）、快捷键操作（`keybindings.RegisterOperate` + `ui.RegisterOperateHandler`，见「快捷键操作」）。
- `menu.go`——`<MenuType>Menu` 最小菜单实现（嵌入 `ui.BaseMenu`，覆写 `GetMenuKey` / `MenuViews` / `SubMenu` / `IsPlayable` / `IsLocatable`）。
- `<name>_test.go`——注册表构建 + 菜单接口断言骨架（模式同 `internal/plugins/checkupdate/checkupdate_test.go`）。
- `README.md`——生成参数记录与接入步骤速查。

**接入步骤**：

1. 在 `internal/plugins/plugins.go` 添加一行空导入（生成器会提示），使本包 `init()` 注册在链接时生效。
2. 按需调整 `registry.go` 的主菜单锚点/标题与可选扩展（右键菜单 / 状态栏 / 快捷键操作等）。
3. `go build ./...` 与 `go test ./internal/plugins/<name>/...`（生成器已默认 `gofmt -w`）。
4. 用真实业务逻辑替换 `menu.go` 的静态项。

> 注：脚手架生成的插件包落在 `internal/plugins/<name>`（或 `-dir` 指定位置），key 必须全局唯一（不得与内置 `expectedMenuKeys` 或其它插件冲突）；生成后仍需手动完成聚合器空导入——脚手架不修改任何 `internal/` 文件。

## WASM 插件（实验性，运行时动态加载）

> 从该功能落地起，go-musicfox 支持**不重新编译主程序**的插件形态：用户把插件目录放入配置目录（默认 `<配置目录>/wasm-plugins`，即 `~/.config/go-musicfox/wasm-plugins/`），启动时宿主（`internal/wasm`）经 **wazero** 沙箱加载 `.wasm` 插件并注册其命令（**轨 B UI-agnostic 命令**，见「轨 B 命令」章节）——这是「运行时动态加载」的首个落地形态（MVP：**命令动作 + 文本结果**）。WASM 命令经 `wasm.RegistrySink` 管线注册进 `frontend` 命令注册表：TUI 前端经 `tuiWasmSink` + `WithPlugin` 归属加载、命令在 `registerCommandMenus()` 适配为 `CommandMenu`（主菜单项，禁用插件的入口隐藏 + 执行被拒）；WebUI 前端经 `webuiWasmSink` 加载、命令出现在 `/api/commands`（`exec` action 被策略拒绝）。Go `plugin` 包 / 共享库热加载仍不支持。示例插件见 `examples/wasm/hello/`。

### 插件形态

每个子目录是一个插件，含 `manifest.toml` + 一个 wasm reactor 文件：

```text
~/.config/go-musicfox/wasm-plugins/
  hello/
    manifest.toml
    main.wasm
```

```toml
# manifest.toml
id = "hello"            # 插件 id（唯一；用于 [plugins] disabled 启停）
name = "Hello WASM"     # 显示名
version = "0.1.0"
author = "you"
description = "示例 WASM 插件"
sha256 = ""             # 可选：main.wasm 的 64 位十六进制 SHA-256；非空时启动校验，不匹配拒绝加载
wasm = "main.wasm"      # 可选，默认 "main.wasm"

[[menus]]
key = "wasm_hello"      # 全局唯一菜单注册 key（不得与内置/其它插件 key 冲突）
title = "你好 WASM"      # 主菜单项标题
after = ""              # 可选：主菜单 after-anchor 前驱项 key；空 = 追加在链尾
export = "run"          # 可选：调用的 wasm 导出函数，默认 "run"
args = {}               # 可选：静态参数，随请求 JSON 传给插件
```

### 契约（host ↔ guest 的 JSON）

请求（宿主 → 插件 `export` 函数）：

```json
{ "version": 1,
  "action": "wasm_hello",
  "args": { "name": "musicfox" },
  "context": { "userId": 0, "userName": "", "playing": true,
               "song": { "id": 1, "name": "...", "artist": "...", "album": "..." } } }
```

响应（插件 → 宿主），`action` 决定宿主行为：

| action | 字段 | 宿主行为 |
|--------|------|----------|
| `toast` | Title / Message / Level（`info`/`success`/`warning`/`error`） | TUI 内通知 |
| `view` | Title / Message | TUI 以独立可滚动文本页呈现（`command_view` 页面，toast 同步提示）；S8 交互协议预留 `ViewPageContent`/`ViewPageHooks` 接口 |
| `open_url` | URL | 系统浏览器打开链接（`open.Start`） |
| `exec` | Command / Args | 执行命令（**无 shell 包装**，`exec.Command(command, args...)`） |

未知/空 action 忽略；插件调用失败、响应解析失败均以「WASM 插件执行失败」错误 toast 暴露。

### 开发插件（Go wasmexport）

示例插件是标准 Go wasip1 reactor，编译命令：

```bash
GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o main.wasm .
```

必须满足：

- **导出 `alloc(size uint32) uint32` 与 `dealloc(ptr uint32, size uint32)`**：宿主经 alloc 写入请求、经 dealloc 释放缓冲区。guest 侧需自行持有分配引用防 GC 回收（Go GC 非移动，`uintptr` 在持有引用期间稳定，见 `examples/wasm/hello/main.go` 的 `allocs` map 模式）。
- **每个菜单的 `export` 函数**（默认 `run`），签名 `(reqPtr, reqLen uint32) uint64`：读请求（`unsafe.Slice((*byte)(unsafe.Pointer(uintptr(reqPtr))), reqLen)`），返回**打包的单 uint64 结果** `(uint64(outPtr) << 32) | uint64(outLen)`——Go 的 wasmexport ABI 只允许一个结果值（多值提案未落地）。
- **无需 `main` 逻辑**：reactor 构建忽略 `main`（占位空 `func main(){}` 即可）；宿主经 `_initialize` 初始化运行时（wazero 不会自动调用，宿主显式 `WithStartFunctions("_initialize")`）。
- 语言不限于 Go：宿主只用 wazero 加载 wasm，遵守上述 ABI 的任何语言（Rust `wasm32-wasip1`、TinyGo 等）均可。

### 行为与安全

- **注册与启停**：插件命令经 `wasm.RegistrySink` 管线注册（TUI 侧 `ui.WithPlugin(manifest.id, ...)` 归属盖章）——用户可用既有 `[plugins] disabled = ["hello"]` 配置禁用 WASM 插件（主菜单入口隐藏、命令执行被拒；key 注册仍保留）。
- **失败隔离**：单个插件加载失败（manifest 非法、sha256 不匹配、wasm 缺导出、目录不存在）只记日志并跳过，**不阻断启动**；注册冲突/坏锚点经 recover 隔离。
- **沙箱加固**：wazero 运行时，线性内存上限 128 页（8 MiB）、无文件系统、stdin EOF / stdout-stderr 丢弃、关闭即随 context 取消；单次调用默认超时 5s，超时由 watchdog 关闭实例并报「插件调用超时」（实例随后不可复用）。
- **已知限制**：纯计算死循环由超时 watchdog 尽力中断（Go wasm 在无限 CPU 循环中不响应 context 取消，关闭实例是兜底手段）；wasm 单线程；`exec` 执行用户自装插件声明的命令——**安装即运行其代码**，仅安装可信来源插件，并建议在 manifest 填写 `sha256` 校验文件完整性。
- 契约版本：请求带 `version`（当前 1），插件可据此拒绝不支持的版本。

## 行为保持契约

- 插件**不得改变核心行为**：播放控制、导航、启动序、渲染、主题、桌面歌词、登录链路等用户可见行为必须保持现状；插件的菜单/页面是「新增贡献点」，不得覆写或替换内置菜单的行为。
- **错误经 `(Menu, error)` + toast 暴露**：构建失败返回 error，跳转处经 `buildMenuOrToast` / `buildPageOrToast` toast 报错并降级（返回 nil），**不 panic**。`mustBuildNoArg` 的 panic 语义只适用于静态代码中的注册表编程错误。
- **服务解析不得丢弃 bool**：`framework.ServiceOf[T]` 的第二个返回值必须处理（记录错误 + 降级路径），禁止裸断言。
- 插件 key 全局唯一：不得与内置 key（`expectedMenuKeys` / `expectedPageKeys`）或其它插件冲突。
- **启动钩子不得 panic**：`RegisterStartupHook` 注册的任务在核心引擎启动序列（`core.Engine.Startup`）第 10 步执行，每个 hook 带 recover 隔离——panic 仅记日志、跳过该 hook，不得阻断启动。主菜单入口 key 必须已在菜单注册表中注册（`RegisterMainMenuItem` 的 key 在 `NewMainMenu` 构建，未注册 key 会 panic——属启动完整性信号，而非运行时错误）；无 `Build` 的入口 key 必须是无参菜单 provider（经 `mustBuildNoArg` 构建），参数化菜单入口需用 `RegisterMainMenuItemWith` / `RegisterMainMenuItemAfter` 提供 `Build`。
- **主菜单顺序保持**：插件项经 **after-anchor 链**（每个入口声明其前驱项 key）复现插件化前的原始顺序（用户可见行为）。`After` 目标必须存在（`MainMenuStart` 或已注册入口 key），同一锚点只能被一个入口声明（重复锚点使另一入口成为孤儿，NewMainMenu 链完整性断言 panic 报错）；声明顺序不依赖注册时序——链在 `NewMainMenu` 构建时统一走序并断言完整性。

## 未来演进（预留边界，未实现）

- WASM 插件深化：`view` 渲染为独立页面/popup；插件哈希强制校验与签名分发；插件 API 版本协商。
- **热重载**（P8 已落地基础）：命令集可按代际刷新——WebUI 完整支持（`/api/commands` 每次现查注册表），TUI 命令**执行**按 key 现查刷新（`CommandMenu` 构建/动作时解析当前命令），但主菜单项标题/位置在启动时定格，完整 TUI 热重载（含主菜单重建）仍不规划。Go `plugin` 共享库 / 子进程（go-plugin）形态仍不规划。
- 突破 `internal` 导入规则的外部仓库形态（独立插件仓库直接导入边界包）；`Netease` 薄壳的导出适配层（当前仅经 `BaseMenu.Netease()` 逃生口）。
- 插件元数据（名称/版本/作者）与贡献点声明的进一步扩展（WASM 插件 manifest 已含基本信息）。
