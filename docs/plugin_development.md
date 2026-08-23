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
| 菜单跳转 | `BuildMenu[T]` / `mustBuildNoArg` / `buildMenuOrToast` | `internal/ui/registry.go` |
| 页面跳转 | `BuildPage[T]` / `buildPageOrToast` | `internal/ui/registry.go` |
| 菜单基座（可嵌入） | `BaseMenu`（含导出转发方法） | `internal/ui/menu.go` |
| 服务解析 | `framework.Context` / `framework.ServiceOf[T]` | `internal/framework/context.go` |
| 类型安全访问器 | `menuServices`（`svc.Player()` 等） | `internal/ui/menu_accessor.go` |
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

// 页面 provider：返回 model.Page。
func RegisterPage[T any](key string, f func(opts T) (model.Page, error))
func BuildPage[T any](key string, opts T) (model.Page, error)
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

### 类型安全访问器：menuServices

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
- 插件经聚合器接入：`internal/plugins/plugins.go` 空导入各插件包，`cmd/musicfox.go` 空导入聚合器。插件 key 不得与内置 key 冲突；`expectedMenuKeys` 不包含插件 key（完整性断言只校验内置清单，`check_update` 由插件注册即可通过）。
- `internal/ui` 仍是 internal 包：按 Go internal 规则，插件代码必须位于 go-musicfox 模块树内（如 `internal/plugins/checkupdate`）才能导入；独立仓库插件需 `replace` 到模块树内或以子包形式落地，纯外部模块直接导入 `internal/ui` 仍被 Go 拒绝（属预留边界）。
- `Netease` 薄壳仍未导出：外部插件经 `base.BaseMenu.Netease()` 逃生口访问，不应直接构造。

## 工作示例

> 以下代码与 `internal/ui/plugin_boundary_external_test.go`（包外编译校验）对应；示例一为已合入仓库的真实插件（`internal/plugins/checkupdate`），示例二/三为最小演示形态。

### 示例一：检查更新插件（首个真实提取示例）

> `internal/plugins/checkupdate` 是把内置「检查更新」菜单提取为外部式插件的第一个真实插件，完整走通编译期插件链路：`BaseMenu` 嵌入 + `RegisterMenu` 注册 + `BuildMenu` 导航 + 聚合器 + 空导入。菜单 key 仍为 `check_update`，行为与提取前一致。

`menu.go`（菜单类型 + 检查/通知逻辑，与提取前 `internal/ui/menu_check_update.go` 一致）：

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

// Action 触发检查更新并留在当前页面（非 nil 的 page/cmd 会跳过子菜单导航）。
func (m *CheckUpdateMenu) Action(a *model.App, _ int) (model.Page, tea.Cmd) {
	return a.MustMain(), checkUpdateCmd()
}

func checkUpdateCmd() tea.Cmd {
	return func() tea.Msg {
		hasUpdate, latestVersion := version.CheckUpdate()
		return checkUpdateNotificationMsg(hasUpdate, latestVersion)
	}
}
// ... checkUpdateNotificationMsg / newVersionNotifyContent 与提取前逐字一致；
// 「发现新版本」toast 经 ui.BuildToastNotificationSpec 构建（ui 导出的
// toast-spec 助手，插件对 ui 的唯一依赖）。
```

`registry.go`（编译期注册入口，`init()` 触发）：

```go
// 文件：internal/plugins/checkupdate/registry.go
func init() {
	ui.RegisterMenu("check_update", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return &CheckUpdateMenu{BaseMenu: base}, nil
	})
}
```

聚合器 + 空导入（插件被链接进二进制的入口）：

```go
// 文件：internal/plugins/plugins.go（聚合器：空导入每个插件包，触发其 init()）
package plugins

import (
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/checkupdate"
)

// 文件：cmd/musicfox.go（入口空导入聚合器）
import (
	_ "github.com/go-musicfox/go-musicfox/internal/plugins"
)
```

主菜单跳转点经注册表构建并执行插件菜单的 Action（`internal/ui/menu_main.go`）：

```go
checkUpdate := buildMenuOrToast("check_update", m.baseMenu, NoArgMenuOpts{})
if checkUpdate == nil {
	return app.MustMain(), nil
}
return checkUpdate.Action(app, 0)
```

**启动钩子限制**：插件只接管菜单触发的检查；启动时的自动检查（配置 `[startup] checkUpdate`）仍留在 shell（`netease.go`）直连执行——当前没有启动钩子机制，且 `ui` 不得反向导入插件包（避免 import cycle）。启动检查的 notify 内容在 shell 内联，与插件内 `newVersionNotifyContent` 文案保持一致。

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

### 接入二进制（聚合器）

正式插件的标准接入方式：每个插件子包被聚合器空导入（见示例一），入口只需空导入聚合器即可，不必逐个列插件：

```go
// 文件：internal/plugins/plugins.go —— 聚合器：空导入每个插件包
package plugins

import (
	_ "github.com/go-musicfox/go-musicfox/internal/plugins/checkupdate"
)

// cmd/musicfox.go 或入口处空导入聚合器，触发全部插件 init() 注册：
import (
	_ "github.com/go-musicfox/go-musicfox/internal/plugins"
)
```

> 注：插件子包必须位于 go-musicfox 模块树内（`internal` 导入规则）；`BaseMenu` 已导出、`baseMenu` 是其别名，注册闭包可在 `ui` 包外以 `ui.BaseMenu` 类型书写。`ui` 不得反向导入插件包（import cycle），因此 shell 侧需要复用插件逻辑时须内联（如启动自动检查）。

## 行为保持契约

- 插件**不得改变核心行为**：播放控制、导航、启动序、渲染、主题、桌面歌词、登录链路等用户可见行为必须保持现状；插件的菜单/页面是「新增贡献点」，不得覆写或替换内置菜单的行为。
- **错误经 `(Menu, error)` + toast 暴露**：构建失败返回 error，跳转处经 `buildMenuOrToast` / `buildPageOrToast` toast 报错并降级（返回 nil），**不 panic**。`mustBuildNoArg` 的 panic 语义只适用于静态代码中的注册表编程错误。
- **服务解析不得丢弃 bool**：`framework.ServiceOf[T]` 的第二个返回值必须处理（记录错误 + 降级路径），禁止裸断言。
- 插件 key 全局唯一：不得与内置 key（`expectedMenuKeys` / `expectedPageKeys`）或其它插件冲突。

## 未来演进（预留边界，未实现）

- 运行时动态加载（Go `plugin` / 共享库热加载）与插件清单/配置化启停。
- 突破 `internal` 导入规则的外部仓库形态（独立插件仓库直接导入边界包）；`Netease` 薄壳的导出适配层（当前仅经 `BaseMenu.Netease()` 逃生口）。
- 插件元数据（名称/版本/作者）与贡献点声明。
