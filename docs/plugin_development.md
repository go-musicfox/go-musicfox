# 插件开发指南（plugin development）

本文档描述 go-musicfox 的对外插件边界表面（Phase 3.4 接口层预留）：菜单/页面 provider 注册表、服务解析、生命周期，以及第三方插件当前如何接入（编译期注册）。

## 概述

从 Phase 3.2 起，go-musicfox 的全部菜单与页面统一走 **provider 注册表**（key → 参数化工厂），跳转不再硬编码构造函数；从 Phase 3.1 起，全部业务能力注册为 framework 命名服务，组件按名解析。这套机制为未来的第三方插件生态预留了边界：**opts 结构体（参数契约）即插件契约**。

- 本文档适用对象：想为 go-musicfox 贡献/接入一个「菜单」或「页面」的开发者。
- 当前形态：**编译期注册**（import + `init()` 注册）。**运行时动态加载不在本期支持范围**（spec Non-goals，见「当前形态」一节）。
- 行为约束：插件不得改变核心行为，错误经 `(Menu, error)` + toast 暴露（见「行为保持契约」）。

## 插件边界表面

| 能力 | API | 位置 |
|------|-----|------|
| 菜单注册 | `RegisterMenu[T](key, factory)` | `internal/ui/registry.go` |
| 页面注册 | `RegisterPage[T](key, factory)` | `internal/ui/registry.go` |
| 菜单跳转 | `BuildMenu[T]` / `MustBuildNoArg` / `MustBuild` / `BuildMenuOrToast`（导出形式，供插件使用） | `internal/ui/registry.go` |
| 页面跳转 | `BuildPage[T]` / `BuildPageOrToast[T]` | `internal/ui/registry.go` |
| 主菜单入口 | `RegisterMainMenuItem(key, title)` / `RegisterMainMenuItemWith(key, title, build)` / `MainMenuPluginItems()` | `internal/ui/registry.go` |
| 启动钩子 | `RegisterStartupHook(fn)` | `internal/ui/registry.go` |
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

// 主菜单入口（Phase 3.9）：key 必须已在菜单注册表中注册（主菜单在全部内置项
// 之后追加该入口并按 key 构建菜单；未注册的 key 在 NewMainMenu 启动时 panic，
// 作为启动完整性信号）。无 Build 的入口 key 必须是无参菜单 provider（经
// mustBuildNoArg 构建）；带 Build 的入口由插件以自身 options 构造菜单
// （参数化 provider 主菜单入口）。
func RegisterMainMenuItem(key, title string)                            // 便捷形式，Build = nil
func RegisterMainMenuItemWith(key, title string, build func(base BaseMenu) Menu) // 空 key/title 或重复 key 会 panic
func MainMenuPluginItems() []MainMenuItem                               // 快照，供 NewMainMenu 构造

// 启动钩子（Phase 3.9）：注册启动任务。shell 在 InitHook 中用户/登录就绪后
// 按注册序调用，每个 hook 带 panic 隔离（recover + 日志，不阻断启动）。
func RegisterStartupHook(fn func()) // nil 会 panic
```

规则：

- 注册时机：编译期 `init()`。key 为空、factory 为 nil 或 key 重复注册都会 panic（程序员错误，启动即暴露）。
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

## 当前形态：编译期注册（import + init()）

第三方插件**今天**的接入方式：

1. 插件包实现菜单/页面类型与注册函数，在包 `init()` 中调用 `RegisterMenu` / `RegisterPage`。
2. 插件包被聚合器（`internal/plugins/plugins.go`）空导入，入口（`cmd/musicfox.go`）空导入聚合器，`init()` 在启动完整性断言（`AssertMenuRegistryComplete` / `AssertPageRegistryComplete`）之前完成。
3. 启动断言锁定内置注册清单；插件 key 不得与内置 key 冲突（`expectedMenuKeys` 不含插件 key）。

**明确不支持（spec Non-goals）**：

- **运行时动态加载**（Go `plugin` 包 / 共享库热加载）当前不支持；外部边界只在接口层预留，机制落地需要独立设计。
- 不改造 tea 消息循环、不引入外部 DI 框架、不修改 vendored SDK。

**当前边界注意**：

- `BaseMenu` 已导出，注册闭包可用 `BaseMenu` 书写（`baseMenu` 是别名，二者等价），插件菜单类型嵌入 `ui.BaseMenu` 即可——**可以在 `ui` 包之外实现与注册**（编译期边界验证见 `internal/ui/plugin_boundary_external_test.go`，`package ui_test`；首个真实插件 `internal/plugins/checkupdate` 即此形态）。
- 插件经聚合器接入：`internal/plugins/plugins.go` 空导入各插件包，`cmd/musicfox.go` 空导入聚合器。插件 key 不得与内置 key 冲突；`expectedMenuKeys` / `expectedPageKeys` 不包含插件 key（完整性断言只校验内置清单，`check_update` / `last_fm` / `lastfm_auth` / `lastfm_custom_api`、整个 DJ/电台集群的 `dj_*` / `radio_dj_type`、整个专辑集群的 `album_menu` / `album_*` / `album_detail`、整个歌手集群的 `artist_detail` / `artist_*` / `hot_artists` / `artists_sub_list` 以及整个推荐集群的 `daily_songs` / `daily_playlists` / `personal_fm` / `recent_songs` / `ranks` 由插件注册即可通过）。
- `internal/ui` 仍是 internal 包：按 Go internal 规则，插件代码必须位于 go-musicfox 模块树内（如 `internal/plugins/checkupdate`）才能导入；独立仓库插件需 `replace` 到模块树内或以子包形式落地，纯外部模块直接导入 `internal/ui` 仍被 Go 拒绝（属预留边界）。
- `Netease` 薄壳仍未导出：外部插件经 `base.BaseMenu.Netease()` 逃生口访问，不应直接构造。

## 工作示例

> 以下代码与 `internal/ui/plugin_boundary_external_test.go`（包外编译校验）对应；示例一/四/五/六/七/八为已合入仓库的真实插件（`internal/plugins/checkupdate` / `internal/plugins/lastfm` / `internal/plugins/dj` / `internal/plugins/album` / `internal/plugins/artist` / `internal/plugins/recommend`），示例二/三为最小演示形态。

### 示例一：检查更新插件（首个真实提取示例）

> `internal/plugins/checkupdate` 是把内置「检查更新」菜单提取为外部式插件的第一个真实插件，完整走通编译期插件链路：`BaseMenu` 嵌入 + `RegisterMenu` 注册 + `RegisterMainMenuItem` 主菜单入口 + `RegisterStartupHook` 启动钩子 + 聚合器 + 空导入。菜单 key 仍为 `check_update`，行为与提取前一致。

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

`registry.go`（编译期注册入口，`init()` 触发）：

```go
// 文件：internal/plugins/checkupdate/registry.go
func init() {
	ui.RegisterMenu("check_update", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return &CheckUpdateMenu{BaseMenu: base}, nil
	})
	// 声明主菜单入口：NewMainMenu 在全部内置项之后追加「检查更新」。
	ui.RegisterMainMenuItem("check_update", "检查更新")
	// 注册启动自动检查（原 shell 级硬编码启动检查，见 startup.go）。
	ui.RegisterStartupHook(startupCheck)
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
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/recommend"
)

// 文件：cmd/musicfox.go（入口空导入聚合器）
import (
	_ "github.com/go-musicfox/go-musicfox/internal/plugins"
)
```

主菜单不再对「检查更新」做索引特判（`mainMenuCheckUpdateIndex` 已随插件化移除）：`NewMainMenu` 读取 `MainMenuPluginItems()`，在全部内置项之后追加插件入口并按 key 构建菜单；选中后进入插件菜单，由插件自身的 `BeforeEnterMenuHook` / `Action` 承担检查与通知。

**启动钩子调用点**：shell 的 `InitHook`（`internal/ui/netease.go`）在用户/登录恢复之后、自动播放之前的位置调用 `runStartupHooks()`（即原 shell 级启动自动检查所在位置，启动序第 10 步）。此时 services 已注册、toast 已接线；每个 hook 带 recover 隔离，panic 仅记日志不阻断启动。

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

> `internal/plugins/lastfm` 是第二个真实插件，把内置 Last.fm 菜单/页面整体提取为外部式插件。它补全了示例一的插件能力矩阵：**服务访问**（`svc.Lastfm()`——访问器解析 `ServiceLastfm` 客户端）、**页面插件**（`lastfm_auth` / `lastfm_custom_api` 经 `RegisterPage` 注册，opts 携带 `ui.MenuServices`）、**主菜单项**（`RegisterMainMenuItem("last_fm", "LastFM")`——原内置入口已从 `menu_main.go` 移除，现排在全部内置项之后）。页面通过 `ui.BuildPageOrToast` 打开，渲染复用 ui 导出的自定义页面布局助手。

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
	// 主菜单项：NewMainMenu 在全部内置项之后追加「LastFM」（内置入口已移除）。
	ui.RegisterMainMenuItem("last_fm", "LastFM")
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
	// 主播电台主菜单入口：原为内置索引 12，现为插件主菜单项排在全部内置项之后。
	ui.RegisterMainMenuItem("radio_dj_type", "主播电台")
}
```

集群内共享的 opts 契约类型（`DjCategoryDetailOpts` / `DjHotOpts`，仅集群内部使用）随菜单移入插件包；被 ui 侧共享的 `DjRadioDetailOpts` 留在 ui（跳入该菜单的调用点在 ui）。参数化菜单的 `GetMenuKey()` 仍返回动态形式（如 `dj_radio_detail_<id>`），注册 key 保持静态前缀。集群中唯一被 ui 反向引用的能力是「排序切换」（`OpToggleSortOrder` 键处理对 `DjRadioDetailMenu` 做类型断言）——ui 不能反向导入插件包，故改经导出接口 `ui.DjRadioDetailSortable`（`ToggleSortOrder` / `Reload`）访问，插件菜单实现该接口即保持键处理行为不变。

### 示例六：专辑集群（批量菜单提取 + 主菜单项）

> `internal/plugins/album` 是第四个真实插件：把「专辑列表」整个集群（8 个菜单）整体提取为外部式插件，与示例五同为集群批量提取示范。所有 provider key 与提取前**逐一相同**（`album_menu` / `album_new_area` / `album_top_area` / `album_new_hot` / `album_new` / `album_top` / `album_sub_list` / `album_detail`），因此 ui 侧跳入集群的调用点（`menu_search_result.go` / `menu_artist_album.go` 跳 `album_detail`、`menu_user_collection.go` 构建 `album_sub_list` 子菜单）全部经注册表按 key 跳转、无需改动。`album_menu`（专辑列表入口菜单）声明主菜单入口「专辑列表」（原内置索引 5，现为插件主菜单项排在全部内置项之后，帮助索引随之 12→11）：

```go
// 文件：internal/plugins/album/registry.go（节选）——入口菜单声明主菜单项
func init() {
	// ... RegisterMenu("album_top", ...) 等 8 个注册，key 与原内置注册一致
	ui.RegisterMainMenuItem("album_menu", "专辑列表")
}
```

集群内共享的 opts 契约（`AlbumTopOpts` / `AlbumNewOpts`，仅集群内部使用）随菜单移入插件包；被 ui 侧共享的 `AlbumDetailOpts` 留在 ui（`menu_search_result.go` / `menu_artist_album.go` / `operate.go` 均跳 `album_detail`）。参数化菜单的 `GetMenuKey()` 仍返回动态形式（如 `album_top_<area>` / `album_new_<area>`），注册 key 保持静态前缀。集群中唯一被 ui 反向引用的能力是「查看歌曲所属专辑的去重判断」（`operate.go` 对 `AlbumDetailMenu` 做类型断言取 `albumID`）——ui 不能反向导入插件包，故改经导出接口 `ui.AlbumDetailIDGetter`（`AlbumID() int64`）访问，插件菜单实现该接口即保持行为不变。

### 示例七：歌手集群（第五个真实插件）

> `internal/plugins/artist` 是第五个真实插件：把「热门歌手 / 歌手详情」整个集群（6 个菜单）整体提取为外部式插件，与示例五/六同为集群批量提取示范。所有 provider key 与提取前**逐一相同**（`hot_artists` / `artist_detail` / `artist_song` / `artist_album` / `artist_of_song` / `artists_sub_list`），因此 ui 侧跳入集群的调用点（`menu_search_result.go` / `operate.go` 跳 `artist_detail`、`operate.go` 跳 `artist_of_song`、`menu_user_collection.go` 构建 `artists_sub_list` 子菜单）全部经注册表按 key 跳转、无需改动。`hot_artists`（热门歌手入口菜单）声明主菜单入口「热门歌手」（原内置索引 8，现为插件主菜单项排在全部内置项之后，帮助索引随之 11→10）：

```go
// 文件：internal/plugins/artist/registry.go（节选）——入口菜单声明主菜单项
func init() {
	// ... RegisterMenu("hot_artists", ...) 等 6 个注册，key 与原内置注册一致
	ui.RegisterMainMenuItem("hot_artists", "热门歌手")
}
```

集群内共享的 opts 契约（`ArtistAlbumOpts` / `ArtistSongOpts`，仅集群内部使用）随菜单移入插件包；被 ui 侧共享的 `ArtistDetailOpts` / `ArtistsOfSongOpts` 留在 ui（`menu_search_result.go` / `operate.go` 均跳 `artist_detail`，operate.go 的 `goToArtistOfSong` 以 `ArtistsOfSongOpts` 携带歌曲载荷）。参数化菜单的 `GetMenuKey()` 仍返回动态形式（如 `artist_detail_<id>` / `artist_song_<id>` / `artist_album_<id>`），注册 key 保持静态前缀。集群中被 ui 反向引用的能力是「查看歌曲所属歌手的去重判断」（`operate.go` 对 `ArtistDetailMenu` / `ArtistsOfSongMenu` 做类型断言取 id）——ui 不能反向导入插件包，故改经导出接口 `ui.ArtistDetailIDGetter`（`ArtistID() int64`）与 `ui.ArtistsOfSongSongIDGetter`（`SongID() int64`）访问，插件菜单实现该接口即保持行为不变。

### 示例八：推荐集群（第六个真实插件 + 批量主菜单项）

> `internal/plugins/recommend` 是第六个真实插件：把「推荐/播放历史」集群（每日推荐歌曲、每日推荐歌单、私人FM、最近播放歌曲、排行榜）5 个菜单整体提取为外部式插件，与示例五/六/七同为集群批量提取示范，且是**一次声明 5 个主菜单项**的示范。所有 provider key 与提取前**逐一相同**（`daily_songs` / `daily_playlists` / `personal_fm` / `recent_songs` / `ranks`），ui 侧跳入集群的调用点零改动；集群内跳转（`ranks` / `daily_playlists` 的 SubMenu 跳 `playlist_detail`）经 `ui.BuildMenu` 按 key 解析，`playlist_detail` 留在 ui。5 个入口菜单各自声明主菜单项「每日推荐歌曲 / 每日推荐歌单 / 私人FM / 最近播放歌曲 / 排行榜」（原内置索引 0/1/4/6/8，现为插件主菜单项排在全部内置项之后，帮助索引随之 10→5）：

```go
// 文件：internal/plugins/recommend/registry.go（节选）——入口菜单声明主菜单项
func init() {
	// ... RegisterMenu("daily_songs", ...) 等 5 个注册，key 与原内置注册一致
	ui.RegisterMainMenuItem("daily_songs", "每日推荐歌曲")
	ui.RegisterMainMenuItem("daily_playlists", "每日推荐歌单")
	ui.RegisterMainMenuItem("personal_fm", "私人FM")
	ui.RegisterMainMenuItem("recent_songs", "最近播放歌曲")
	ui.RegisterMainMenuItem("ranks", "排行榜")
}
```

本集群还是**登录门控 + 播放器服务经访问器**的示范：`daily_songs` / `daily_playlists` / `recent_songs` 的 `BeforeEnterMenuHook` 经 `m.User()` / `m.ToLoginPage(enterMenuCallback(main))` 走 `ui.BaseMenu` 转发（`enterMenuCallback` 在插件内镜像 ui 未导出的 `EnterMenuCallback`）；`personal_fm` 的 `BottomOutHook` 经 `m.Player()` 更新播放列表（`ReinitializePlaylist` / `MarkPlaylistUpdated`，Phase 3.6 播放列表 API）。

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
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/recommend"
)

// cmd/musicfox.go 或入口处空导入聚合器，触发全部插件 init() 注册：
import (
	_ "github.com/go-musicfox/go-musicfox/internal/plugins"
)
```

> 注：插件子包必须位于 go-musicfox 模块树内（`internal` 导入规则）；`BaseMenu` 已导出、`baseMenu` 是其别名，注册闭包可在 `ui` 包外以 `ui.BaseMenu` 类型书写。`ui` 不得反向导入插件包（import cycle），shell 需要插件能力时通过注册机制调用：插件能力经 `RegisterMainMenuItem` / `RegisterStartupHook` 声明，由 shell 在构建主菜单 / `InitHook` 时统一消费（不再需要内联插件逻辑）。

## 行为保持契约

- 插件**不得改变核心行为**：播放控制、导航、启动序、渲染、主题、桌面歌词、登录链路等用户可见行为必须保持现状；插件的菜单/页面是「新增贡献点」，不得覆写或替换内置菜单的行为。
- **错误经 `(Menu, error)` + toast 暴露**：构建失败返回 error，跳转处经 `buildMenuOrToast` / `buildPageOrToast` toast 报错并降级（返回 nil），**不 panic**。`mustBuildNoArg` 的 panic 语义只适用于静态代码中的注册表编程错误。
- **服务解析不得丢弃 bool**：`framework.ServiceOf[T]` 的第二个返回值必须处理（记录错误 + 降级路径），禁止裸断言。
- 插件 key 全局唯一：不得与内置 key（`expectedMenuKeys` / `expectedPageKeys`）或其它插件冲突。
- **启动钩子不得 panic**：`RegisterStartupHook` 注册的任务在 `InitHook` 中执行，每个 hook 带 recover 隔离——panic 仅记日志、跳过该 hook，不得阻断启动。主菜单入口 key 必须已在菜单注册表中注册（`RegisterMainMenuItem` 的 key 在 `NewMainMenu` 构建，未注册 key 会 panic——属启动完整性信号，而非运行时错误）；无 `Build` 的入口 key 必须是无参菜单 provider（经 `mustBuildNoArg` 构建），参数化菜单入口需用 `RegisterMainMenuItemWith` 提供 `Build`。

## 未来演进（预留边界，未实现）

- 运行时动态加载（Go `plugin` / 共享库热加载）与插件清单/配置化启停。
- 突破 `internal` 导入规则的外部仓库形态（独立插件仓库直接导入边界包）；`Netease` 薄壳的导出适配层（当前仅经 `BaseMenu.Netease()` 逃生口）。
- 插件元数据（名称/版本/作者）与贡献点声明。
