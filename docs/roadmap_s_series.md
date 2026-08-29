# S 系列后续立项：S1（WASM view 独立页面）· S5（WebUI-connect 客户端模式）· GUI（Wails v2 原生窗口）

> 本文档为 go-musicfox 后续三阶段立项设计文档（ticket 级拆票规格），供实施团队直接使用。风格与粒度对齐 `docs/frontend_plugin.md`（每票含目标 / 涉及文件表 / 签名级改动 / 依赖 / 验证标准 / 风险 / 写权限）。
>
> 目标分支：`feat/plugin-framework-playback` 及其后续。
>
> 前提：P1–P8 已全部落地（`internal/frontend` 前端注册表、轨 B 命令契约、core 事件总线、headless daemon 订阅协议、webui standalone 前端、WASM 迁轨 B）。本文档所有「现状事实」均已逐文件核实。
>
> 来源：本文档由两次独立立项设计 reconcile 而成（设计作者为三层 cordis 化 P1-P8 的同一 oracle 会话族），关键裁决编号 D-S1-1 … D-GUI-3，含备选方案与采用理由。

## 0. 概述

### 0.1 背景

P8 之后的现状：

- **轨 B 命令契约**（`internal/frontend/command.go`）已含 `"view"` action，但 TUI 消费端（`internal/ui/command_menu.go:159-161`）把 `"toast", "view"` 合并为多行 toast——**view 无独立页面**（`docs/plugin_development.md:875` 明示「MVP 以多行 toast 呈现」）。长文本内容在 toast 中不可滚动、不可持久查看。
- **WebUI 前端**（`internal/webui/`）是「自建 engine 的 standalone 服务」：每次 `--frontend=webui` 都新建 `core.Engine`（新音频输出、新登录态、新播放列表）。**同一机器上「常驻播放 + 随时浏览器控制」没有轻量路径**——headless daemon 只有 `musicfox ctrl` 的窄命令面，没有事件驱动的富页面。
- **前端注册表**（`internal/frontend`）已能承载任意前端（TUI/headless/webui），但还没有原生窗口形态；WebUI 的页面/WS/API 资产已被验证为完整可用的「对后端契约」。

### 0.2 目标

| 阶段 | 目标 | 一句话价值 |
|------|------|-----------|
| **S1** | `CommandResult{Action:"view"}` 的 TUI 消费端升级为独立可滚动文本页（`CommandViewPage`） | WASM / 轨 B 命令的长文本输出从「一次性 toast」变为「可滚动持久页面」，并为 S8「view 交互协议」预留接口 |
| **S5** | WebUI 客户端模式：`LaunchOptions.Mode=connect`，WebUI 作为客户端 Dial 本地 headless daemon，**不建 engine** | 「headless 常驻播放 + 浏览器富控制面板」的单实例形态，页面/JS 零改动、后端换源 |
| **GUI** | Wails v2 原生窗口前端（`frontend.Register` 加 `"wails"`），复用 WebUI 资产 | 无浏览器依赖的原生客户端形态，macOS/Windows 优先，Linux 无桌面环境自动回退浏览器 |
| **S6（TUI-connect）** | TUI 遥控壳：`--frontend=tui --mode=connect` 连接本地 headless daemon，控制经 Call 转发、状态经订阅驱动，浏览/搜索留本地 | 无浏览器依赖的终端遥控形态；与 webui-connect 共享 SubscribeClient 数据面，对齐「本机单实例」边界哲学（D-TC-1 方案 B：轻量遥控壳，非整体换源） |

### 0.3 实施优先级与依赖图

**优先级：S1 → S5 → S6（TUI-connect）→ GUI**（S6 依赖 S5 已落地资产、由用户需求直接驱动，先于 GUI；GUI 仍依赖 S5-2，不受 S6 影响）。

```
S1（view 独立页面）────────── 独立，无前置
    │
S5（connect 客户端模式）────── 独立（S5-2 的 Server 后端抽象 + 认证可配置是 GUI 的前置）
    │                          │
    │                          └────────────┐
    │                                     GUI（Wails v2）── 依赖 S5-2
    │
S6（TUI-connect）────────── 依赖 S5-3（SubscribeClient）/ P7 订阅协议 / Dispatcher 命令集
                              （不依赖 webui.Backend，D-TC-2；先于 GUI，D-TC-5）
```

- **S1 / S5 完全并行**（文件集不重叠：S1 在 `internal/ui` + foxful-cli，S5 在 `internal/frontend`/`internal/webui`/`internal/headless`）。
- **GUI 依赖 S5**：GUI 复用 `webui.Server`（经 Backend 抽象注入 `localBackend`），且认证层必须可配置（见 §3.4 的 D-GUI-2）。
- 每票独立 commit，Conventional Commits（`feat(ui): ...` / `feat(webui): ...` / `feat(gui): ...`）。

### 0.4 裁决清单（作者已定，实施时如有新证据可推翻但需注释理由）

| 编号 | 裁决 | 采用 | 备选方案（未采用理由） |
|------|------|------|------------------------|
| D-S1-1 | view 跳转机制 = **foxful-cli 极小扩展 `Options.UnknownMsgHandler`**（nil 时行为零变化）+ `commandActionCmd` 运行时判定（WASM 与编译期命令统一生效），toast 照发 + 投递 `commandViewMsg` | ✅ | ① foxful `App.SetPage` 导出（更薄但直接暴露底层 setPage，绕过 Main 分发语义）；② 声明式 `Command.ResultAction`（仅编译期命令可用，WASM 无法预判，功能打折） |
| D-S1-2 | WebUI 侧 view **不升级**（JS 已渲染 message/data），留待 GUI 复用资产时统一裁决 | ✅ | — |
| D-S5-1 | `CtrlClient`/`DialAddr`/`SubscribeClient` **保留在 `internal/headless`**，webui → headless 依赖可接受（无环、headless 包 UI-free、与 `commands/ctrl.go` 先例一致） | ✅ | 新建 `internal/daemonclient` 共享包（依赖方向更纯、与 core 命令面语义对齐，但迁移 server.go/ctrl.go/测试改动面大、风险高；本项目「简单性优先」准则下不取） |
| D-S5-2 | connect 模式**功能边界**：播放控制/状态/事件全可用；命令面（`/api/commands`）为空、登录端点为 503、`/api/albumart` 404、`/api/lyrics` 空结构、WS `quit` 不转发 | ✅ | daemon 端扩展 `commands`/`run_command` 传输层命令使 connect 命令面可用（体验更丰富，但需 daemon 端新命令 + exec 安全边界下沉，记入 S5 后「可选扩展」） |
| D-S5-3 | `--mode` 仅 CLI 旗标（不做配置项，YAGNI）；非 webui 前端忽略并警告 | ✅ | `--connect` 布尔 flag（更短但不可扩展为更多模式） |
| D-GUI-1 | 采用**方式 B**（同 module 共存 + manual build，跳过 wails CLI 直接 `go build -tags desktop,production`）；`internal/frontend/gui/` 实现 Frontend 接口 | ✅ | 独立嵌套 module（隔离干净但双 go.mod、依赖重复、需 go.work） |
| D-GUI-2 | 窗口加载路径**优先 GUI-B（Navigate 外部 http URL，认证链路零改动）**，spike 验证；失败回退 GUI-A（AssetServer + `ServerOptions.Auth=false`） | ✅ | — |
| D-GUI-3 | Linux 无桌面环境检测 → 回退「系统浏览器打开 WebUI」；GUI 优先 macOS/Windows 交付 | ✅ | — |
| D-TC-1 | TUI-connect 采用**轻量遥控壳**（方案 B：新增 `RemotePlayer` + `ui.Player` 遮蔽转发 + 降级分支），**不做整体换源**（方案 A） | ✅ | 方案 A（core.Player 全接口远程化 + daemon 快照传输完整 structs.Song）：15–25 人日、TUI 全菜单行为回归面大，且与 S5「精简快照」设计矛盾；方案 B 6–10 人日、复用 SubscribeClient 零协议改动 |
| D-TC-2 | TUI **直接消费 `headless.SubscribeClient`**，**不复用 `webui.Backend`/`remoteBackend`** | ✅ | TUI import webui（兄弟前端包反向耦合，语义混乱）；把 Backend 提升到中立包（S5 刚定稿的包边界重构，风险与收益不成比例） |
| D-TC-3 | TUI-connect 功能边界：控制/状态/事件全可用；播放队列只读精简；选歌播放、登录、歌词、封面、频谱、智能模式、命令面全部降级（对齐 D-S5-2 哲学） | ✅ | daemon 端扩展 play_song/歌词/封面协议（体验完整，但需 daemon 协议大扩展，记入 S6-P2 扩展清单） |
| D-TC-4 | 断线**不自动重连**（MVP），事件通道关闭 → ready=false + toast 降级 | ✅ | 自动重连 + 状态重同步（对齐 webui-connect，记入 S6-P2 扩展清单） |
| D-TC-5 | **TUI-connect 先于 GUI**（用户需求驱动 + 依赖已就绪 + 工作量 6–10 人日 vs 15–25 人日） | ✅ | GUI 先行（GUI 无前置依赖 S6，但交付节奏上用户价值 S6 更直接） |
| D-TC-6 | 歌词/封面/选歌播放 = **S6-P2 扩展**（需 daemon 快照/命令扩展），MVP 明确降级隐藏 | ✅ | MVP 即做（歌词拉取 + daemon 快照加 PicUrl + play_song 命令：拉大票面 3–4 人日，挤占 MVP 收敛目标） |

---

## 1. S1：WASM view 独立页面渲染（TUI 可滚动文本页）

### 1.1 现状事实（已核实）

- 轨 B 契约 `internal/frontend/command.go:29-38`：`CommandResult.Action` 含 `"" | "toast" | "view" | "open_url" | "exec" | "data"`；`view` 语义 = Title + Message 文本。
- TUI 消费端 `internal/ui/command_menu.go:159-161`：`case "toast", "view": a.Notify(commandResultSpec(res))` 合并；`commandResultSpec`（76-85）只取 Level/Title/Message 做多行 toast。`commandActionCmd`（121-167）在 bubbletea goroutine 异步执行命令（WASM 最长 5s），结果经 `app.Notify` 线程安全投递。
- WASM 侧 `internal/wasm/contract.go:38-46`：`Response.Action "view" = Title+Message`；`sink.go` 的 `callWasm` 已映射为 `frontend.CommandResult`——**WASM 契约零改动**。
- 页面基建：`internal/ui/page_layout.go` 提供 `pageMenuTitleViewWithBack`（带返回按钮标题行）、`PageMenuTitleRow`/`FinishCustomPageView` 等导出助手；页面走 `RegisterPage[T]`/`BuildPage`/`buildPageOrToast`（`internal/ui/registry.go:159-218`）。
- `model.Page` 接口（`vendor/.../foxful-cli/model/page.go:14-34`）：`IgnoreQuitKeyMsg` / `Type` / `Update(msg, *App) (Page, tea.Cmd)` / `View(*App) string` / `Msg()`。
- 封面图清理：独立页面跳转点须 `coverRenderer.ClearDisplayed()`（`internal/ui/netease.go:194,201` 的 ToLoginPage/ToSearchPage 先例；AGENTS.md 硬性规则）。
- **关键约束（已核实）**：`model.App.Update` 对非 Key/Mouse 消息在 434 行转给当前页面的 `Update`；`model.Main.Update` 对未知消息返回 `(m, nil)`（`main.go:470`），**无扩展点**。因此「异步命令完成后从 Main 页面跳转新页面」**必须**给 foxful-cli 加极小钩子（D-S1-1），或退化为命令作者声明式预判。
- 未来锚点：`docs/plugin_ecosystem.md §三 3.3` 的 S8「view 交互协议」（页面向插件回传交互）——本阶段只留接口形状，不实现。

### 1.2 目标

`Action:"view"` 的结果在 TUI 中以**独立可滚动文本页**呈现：标题 = `Title`，正文 = `Message` 分行的可滚动文本；支持键盘（↑/↓/PgUp/PgDn/Home/End/←/Esc）与鼠标滚轮；toast 照发（信息即时可见），页面提供持久查看。页面抽象为 S8 交互协议留扩展口（本阶段不接线）。

### 1.3 Ticket 拆分

```
S1-1（CommandViewPage 页面类型 + provider 注册）──────── 无前置
S1-2（commandActionCmd 分派 + foxful OnUnknownMsg 钩子）┐ 依赖 S1-1（页面可构建）
S1-3（测试 + 文档）────────────────────────────────────┘ 依赖 S1-1 + S1-2
```

S1-1 与 S1-2 的页面/分派两部分可并行开发（S1-2 的编译需要 S1-1 的 `RegisterPage` 注册点，但代码编写可并行，合并在同一里程碑合并）。

---

### S1-1｜`CommandViewPage`：view 可滚动文本页面类型 + `command_view` provider 注册

**目标**：新增 TUI 独立页面 `CommandViewPage`（纯展示、可滚动），经页面注册表注册为 `"command_view"`，为 S8 交互协议留出 `ViewPageContent`/`ViewPageHooks` 接口形状。

**涉及文件**：

| 文件 | 动作 |
|---|---|
| 新增 `internal/ui/command_view_page.go` | 页面类型 + 构造 opts + `init()` 注册 |
| 新增 `internal/ui/command_view_page_test.go` | 滚动边界/超宽行/渲染不 panic 单测 |

**改动内容（签名级）**：

```go
// command_view_page.go
package ui

// CommandViewOpts 是 "command_view" 页面的构造参数（RegisterPage 契约）。
// 内容由命令执行结果产生（命令只在 bubbletea goroutine 跑一次，页面纯渲染）。
type CommandViewOpts struct {
    Title string
    Lines []string // Message 按 \n 分行；空行保留
}

// CommandViewPage 渲染可滚动文本视图。
type CommandViewPage struct {
    opts    CommandViewOpts
    scroll  int   // 首行偏移（0-based），恒 ≥ 0
    hovered bool  // 返回按钮 hover（pageBackButtonIcon 样式切换）
}
```

- 实现 `model.Page`：
  - `Type() model.PageType`：返回 `"command_view"`（自定义 `PageType` 常量 `PtCommandView model.PageType = "command_view"`）。
  - `IgnoreQuitKeyMsg(msg)`：`msg` 为 `"q"` 时返回 `false`（保持全局退出；文本页无输入框，无需吞 q）。
  - `Msg()`：返回 `nil`（内容已在构造时给定，无异步加载）。
  - `Update(msg tea.Msg, a *model.App) (model.Page, tea.Cmd)`：
    - 键盘：`up`/`down`（±1 行）、`pgup`/`pgdn`（±半页，`EffectiveWindowHeight/2`）、`home`/`end`（首/末行）、`left`/`esc`/`backspace`（返回 `a.Main()`）；滚动范围 clamp 到 `[0, len(Lines)-可见行数]`。
    - 鼠标：`MouseWheelUp`/`MouseWheelDown`（±3 行，镜像菜单滚轮习惯）；左键点击命中返回按钮（`PageMenuTitleRow` 行 + `pageBackButtonWidth` 列）→ 返回 Main；`MouseMotionMsg` 更新 `hovered`（命中返回按钮时变化 → `RerenderCmd`）。
    - 命中坐标记录约定（AGENTS.md「自定义表单页面」规则）：返回按钮命中坐标按 `pageMenuTitleViewWithBack` 的实际渲染位置记录。
    - 其他消息：忽略，返回 `(p, nil)`。
  - `View(a *model.App) string`：
    - 复用 `pageMenuTitleViewWithBack(a, main, &top, 标题MenuItem, hovered)` 渲染带返回按钮的标题行。
    - 正文区：从 `opts.Lines[scroll:]` 逐行渲染，可用行数 = `a.EffectiveWindowHeight() - 已占用行`（镜像 `pageMenuTitleView` 的行计算）；超宽行用 `lipgloss.NewStyle().MaxWidth(maxWidth)` 截断（防超宽行破坏布局）；底部以 `FinishCustomPageView` 填充（`fillPageHeight` + `RenderAppBackground`）。
    - 返回按钮 hover 样式走 `pageBackButtonIcon(hovered)`，与 Search/Login 页一致。
- **S8 交互协议留口（本票只定义接口形状，不接线、不实现）**：

```go
// ViewPageContent 抽象 view 页面的数据源（S8 交互协议扩展口：页面内容可
// 由命令动态提供而非构造时快照）。本阶段 CommandViewPage 直接持
// CommandViewOpts，未实现本接口；S8 引入交互协议时按此接口扩展。
type ViewPageContent interface {
    Title() string
    Lines() []string
}

// ViewPageHooks 可选交互回调（S8：页面向插件回传交互，如行选择/按键事件
// 上报）。本阶段为空接口，仅作锚点。
type ViewPageHooks interface{}
```
- 注册（本文件 `init()`，镜像 `registry_registrations.go` 模式）：

```go
func init() {
    RegisterPage("command_view", func(opts CommandViewOpts) (model.Page, error) {
        return &CommandViewPage{opts: opts}, nil
    })
}
```

**依赖**：无前置，与 S1-2 并行编写。

**验证**：
- `go build ./... && go vet ./internal/ui/...`；`go test ./internal/ui/... -run CommandViewPage`。
- 单测覆盖：空 Lines 不 panic；滚动 clamp（scroll 越界上/下）；超宽行截断后总宽度 ≤ 窗口宽；`View` 返回字符串行数 = 窗口高（fillPageHeight 保证）；返回键/滚轮消息返回 Main 页。
- 冒烟（S1-2 合入后）：见 S1-3 冒烟清单。

**风险**：低。纯新增页面，不触碰既有页面/菜单路径。命中坐标与 Search/Login 页共用约定，注意 `pageMenuTitleRow` 与 `pageMenuTitleViewWithBack` 的 `top` 参数传递。

**写权限**：fixer。

---

### S1-2｜`commandActionCmd` 分派升级 + foxful-cli `Options.UnknownMsgHandler` 极小扩展

**目标**：`view` 结果从「仅 toast」升级为「toast + 独立页面跳转」。`commandActionCmd` 对运行时判定为 `view` 的结果返回自定义 `tea.Msg`（非 nil）；foxful-cli 的 `Main.Update` 增加可选 `UnknownMsgHandler` 钩子兜底分发未知消息（nil 时行为零变化），由 TUI 装配处把 `commandViewMsg` 转成 `CommandViewPage` 跳转。

**涉及文件**：

| 文件 | 动作 |
|---|---|
| `vendor/github.com/anhoder/foxful-cli/model/options.go` | 加 `UnknownMsgHandler` 字段 |
| `vendor/github.com/anhoder/foxful-cli/model/main.go` | `Main.Update` 默认分支（470 行）经钩子分发 |
| `internal/ui/command_menu.go` | `commandActionCmd` 的 `case "toast", "view"` 拆分 + `commandViewMsg` 类型 + `splitLines` |
| `internal/ui/frontend.go`（TUI 装配处） | `app.With` 设置 `UnknownMsgHandler` |
| `internal/ui/command_view_page.go` | 封面清理引用（跳转点 `ClearDisplayed`） |

**改动内容（签名级）**：

1. foxful-cli（vendor，走 AGENTS.md 的 foxful 依赖更新流程：本地 `replace` → commit/tag → 升版本 → `make vendor`）：

```go
// options.go（foxful-cli）
// UnknownMsgHandler 可选：Main.Update 对未识别 tea.Msg 的兜底分发。
// nil 时保持现状（忽略并返回自身）。返回的 Page 非 nil 时由 App 切换
// 到该页面（App.Update 434 行 setPage 语义）。用于「异步命令结果 →
// 导航新页面」等前端扩展场景（go-musicfox S1）。
UnknownMsgHandler func(msg tea.Msg, a *App) (Page, tea.Cmd)
```

```go
// main.go  Main.Update 末尾（470 行 `return m, nil` 之前）
if m.options.UnknownMsgHandler != nil {
    if page, cmd := m.options.UnknownMsgHandler(msg, a); page != nil || cmd != nil {
        return page, cmd
    }
}
return m, nil
```

2. `internal/ui/command_menu.go`：

```go
// commandViewMsg 由 commandActionCmd 对 Action=="view" 的结果投递
// （非 nil tea.Msg）。Main.Update 经 foxful Options.UnknownMsgHandler
// 兜底分发给 TUI 装配处的处理器 → 构建 command_view 页面。
type commandViewMsg struct {
    Title string
    Lines []string
}

// splitLines 把 Message 按 \n 分行（保留空行）。\r\n 归一为 \n。
func splitLines(message string) []string

// commandActionCmd 的 ④ 改为：
switch res.Action {
case "toast":
    a.Notify(commandResultSpec(res))
case "view":
    a.Notify(commandResultSpec(res)) // toast 照发：信息即时可见
    return commandViewMsg{Title: res.Title, Lines: splitLines(res.Message)}
case "open_url", "exec":
    a.Notify(runCommandSideEffects(res))
}
return nil
```

- `commandResultSpec`（76-85）**零改动**（view 的 toast 兜底路径保留）。
- `runCommandSideEffects` 零改动。

3. `internal/ui/frontend.go`（`tuiFrontend.Run` 装配处，`model.App` 创建之后）：

```go
// D-S1-1：view 结果经 foxful UnknownMsgHandler 分发为独立页面。
// 页面跳转点必须清理封面图（AGENTS.md 规则）。
app.With(model.WithUnknownMsgHandler(func(msg tea.Msg, a *model.App) (model.Page, tea.Cmd) {
    vm, ok := msg.(commandViewMsg)
    if !ok {
        return nil, nil
    }
    n.coverRenderer.ClearDisplayed()
    if page := buildPageOrToast("command_view", CommandViewOpts{Title: vm.Title, Lines: vm.Lines}); page != nil {
        return page, nil
    }
    return nil, nil
}))
```

- `model.WithUnknownMsgHandler` 为 foxful `WithOption` 新增构造器（与既有 `With*` 同名风格）。

**运行语义（已推演）**：
- `commandActionCmd` 的 msg 在 Main 页面活动时进入 `Main.Update` → 默认分支 → 钩子 → 返回 `commandViewPage` → `App.Update` 434 行 `setPage` 跳转。
- 非 Main 页面活动时（理论上命令菜单只在 Main 触发，防御性）：msg 进当前 page 的 `Update`，不认识则忽略——安全无副作用。
- 热重载语义不变：`commandActionCmd` 按 key 运行时解析当前命令（`frontend.CommandByKey`），view 判定基于**运行时返回的 `res.Action`**，WASM 命令与编译期命令统一生效，无需命令作者声明。

**依赖**：**S1-1**（`RegisterPage("command_view", ...)` 必须已注册，否则 `buildPageOrToast` 报「页面加载失败」toast）；foxful-cli 依赖更新流程（本项目既有能力）。

**验证**：
- `go build ./... && go vet ./internal/ui/...`；`go test ./internal/ui/... ./internal/frontend/...`。
- 单测：`splitLines`（`\r\n` 归一、空行保留）；`commandActionCmd` 对假命令（`Run` 返回 `view`）返回 `commandViewMsg`、对 `toast` 返回 nil 且 Notify 被调用（用现有 Notify 测试基建）。
- 冒烟（手工，见 S1-3 清单）。

**风险（S1 最高风险票）**：
- **foxful-cli vendor 扩展**是框架改动：① 必须保证 `UnknownMsgHandler == nil` 时 `Main.Update` 行为逐字节不变（现有全部 TUI 行为由测试守护）；② 需走 foxful 依赖更新流程（`replace` → tag → 升级），单独 commit 便于 review；③ **回退方案（D-S1-1 回退）**：若框架改动不被接受，改为 `frontend.Command` 增加声明式 `ResultAction string` 字段（缺省 `"toast"`），`CommandMenu.Action` 对 `ResultAction=="view"` 的命令**先跳转页面**、页面内经 `Msg()` 触发加载 cmd 执行命令（页面内复制 gating），此方案仅编译期命令可用（WASM 命令无法预判），功能打折。
- `commandViewMsg` 字段命名/位置是 ui 内部细节，不进轨 B 契约（`frontend.CommandResult` 零改动）。

**写权限**：fixer（foxful-cli 改动建议 reviewer 重点复核 nil 分支等价性）。

---

### S1-3｜测试 + 文档（契约注释、AGENTS.md、docs）

**目标**：补齐 S1 的行为测试与文档同步义务（项目文档维护准则硬性要求）。

**涉及文件**：

| 文件 | 动作 |
|---|---|
| 新增 `internal/ui/command_view_flow_test.go` | 端到端消息流测试（菜单 Action → 命令 → view msg → 页面跳转） |
| `docs/plugin_development.md`（875 行附近） | view 呈现方式更新 |
| `docs/frontend_plugin.md` | 轨 B 消费端段落补 view 独立页面 |
| `docs/plugin_ecosystem.md`（§3.3） | 标注 S8 留口接口形状（`ViewPageContent`/`ViewPageHooks`）已就位 |
| `AGENTS.md` | 页面基建/插件段落补 `command_view` |

**改动内容**：
- 测试：镜像 `menu_to_page_test.go`/`foxful_integration_test.go` 基建，构造 `CommandMenu` + 假命令（`Run` 返回 `view`），经 `model.App` 消息循环验证：① 命令菜单 Action 返回 cmd；② cmd 执行后产出 `commandViewMsg`；③ 钩子构建页面；④ `App.Update` 切换到 `CommandViewPage`；⑤ 页面滚动/返回 Main。**不真实执行 WASM**（用假 `frontend.Command` 注册/替换即可，参考 `ReplaceCommand` 测试）。
- 文档：
  - `plugin_development.md:875`：「view：MVP 以多行 toast 呈现」→「view：TUI 以独立可滚动文本页呈现（`command_view` 页面，toast 同步提示）；S8 交互协议预留 `ViewPageContent`/`ViewPageHooks` 接口」。
  - `frontend_plugin.md`：轨 B「一个抽象三个消费方」段补 TUI 侧 view 消费端升级说明。
  - `plugin_ecosystem.md §3.3`：S8 锚点处标注接口形状已就位（未接线）。
  - `AGENTS.md`：页面注册表段落、WASM 插件段落补 `command_view`。

**依赖**：S1-1 + S1-2。

**验证**：`go build/vet/test ./...` 全绿；`go test ./internal/ui/...`；冒烟清单：

```
冒烟清单（S1）：
1. 构造一个返回 view 的测试命令（或临时 WASM view 插件）挂到主菜单
2. 触发命令 → toast 弹出（title/message）→ 同时跳转 command_view 页面
3. 页面标题 = Title，正文逐行渲染、空行保留
4. ↑/↓ 滚动、PgUp/PgDn 半页、Home/End、鼠标滚轮
5. ←/Esc/点击返回按钮 → 回到主菜单
6. 超宽行截断不破坏布局；窗口 resize 后行数自适应
7. 封面图在页面跳转后已清理（无残留 Kitty 图像）
8. toast 类命令行为与 S1 前完全一致（不跳页面）
```

**风险**：低。文档腐化是主要风险——三份 docs + AGENTS.md 必须同票收尾。

**写权限**：fixer。

---

## 2. S5：WebUI-connect 客户端模式

### 2.1 现状事实（已核实）

- `internal/frontend/frontend.go:20-24`：`LaunchOptions{Once, Debug, Pprof}` 仅 3 字段；`Frontend` 接口 `ID()/Name()/Run(ctx, opts)`。
- `internal/webui/run.go`：standalone 流程 = `core.NewEngine` → `wasm scope`（`LoadIntoScope` + `webuiWasmSink`）→ `engine.Startup(webuiNoopObserver{})` → `NewServer(engine)` → `server.Serve` → `open.Start("/token?token=...")` → 阻塞到信号/`ShutdownCh`。
- `internal/webui/server.go` **强耦合 `*core.Engine`**：`engine` 字段、`core.NewDispatcher(engine)`（59 行）、`framework.ServiceOf[*framework.EventEmitter](engine.Ctx(), core.ServiceEventBus)` 订阅事件面（73-76，engine nil 跳过）；`mux` 注册 4 组路由（静态/token/ws/api）。
- `internal/webui/api.go`：`handleStatus` 用 `s.dispatcher.Dispatch`；`handleAlbumArt`/`handleLyrics` 直用 `s.engine.Player().PlayingInfo()` / `s.engine.LyricService().State()`。
- `internal/webui/commands.go`：`handleCommandsList`/`handleCommandExec` 用 `s.engine.Player().CommandContext()` + `frontend.Commands()`；`commandExecAllowed=false`。
- `internal/webui/api_login.go`：QR 登录端点用 `completeQRLogin(s.engine, jar)`；`s.engine == nil` 时 500。
- `internal/webui/ws.go`：`handleWS` 校验 → snapshot（`buildSnapshot` = Dispatcher status + engine playlist）→ `serveWS`（Dispatch + quit 传输层拦截）。
- `internal/webui/events.go`：`eventWireToFrame` 映射（`core.EvSongChanged`→`song_changed` 等 5 个）+ `subscribeEmitter` + `positionThrottle`。
- `internal/headless/client.go`：`CtrlClient{Dial(), Call(ctx, cmd, args)}`——每次 Call 新建连接、一请求一响应、3s 截止；`DialAddr()` 解析 unix socket（`DataDir/musicfox.sock`，0600）或 Windows TCP（`musicfox.port`）。
- `internal/headless/server.go`（P7 订阅协议）：首请求非 `subscribe` 走一请求一响应；`subscribe` 进入长连接会话（快照帧 = Dispatcher status + playlist 精简字段，先于注册；ack；随后事件帧 `{"type":"event",...}` 按订阅集过滤；`unsubscribe`/`quit` 处理）。`frontend.EventSink` 为通用事件扇出基座。
- `internal/commands/netease.go:40-52`：`resolveFrontendID()` → `frontend.ByID` → `fe.Run(ctx, LaunchOptions{...})`；`--once` 仅 headless。
- **依赖方向事实**：`commands/ctrl.go` 已 import `internal/headless` 使用 `CtrlClient`（先例）；`internal/headless` 源级禁止导入 ui/bubbletea/foxful-cli，是 UI-free 包。

### 2.2 目标与定位

`--frontend=webui --mode=connect` 时，WebUI 不建 engine、不建 wasm scope，而是 Dial 本地 headless daemon（`musicfox --headless` 常驻），经 daemon 的 subscribe 协议获得状态/事件、经 `Call` 转发控制命令。**定位明确为「本机单实例」**：unix socket（0600）与 Windows 127.0.0.1 均不可跨机器访问，connect 模式解决的是「常驻播放 + 随时浏览器控制」而非远程访问（远程访问不在本阶段范围）。

页面/JS **零改动**（同一套 `/token` + `/api/*` + `/ws` 契约，后端换源）。

### 2.3 Ticket 拆分

```
S5-1（LaunchOptions.Mode + CLI --mode + Run 分发）─── 无前置
S5-2（Server 后端抽象 Backend + 认证可配置）───────── 无前置（GUI 的前置！）
S5-3（daemon 订阅客户端 + remoteBackend + connect.go）┐ 依赖 S5-1 + S5-2
S5-4（测试 + 文档）──────────────────────────────────┘ 依赖 S5-2 + S5-3
```

S5-1 / S5-2 并行；S5-3 依赖两者；S5-4 收尾。

---

### S5-1｜`LaunchOptions.Mode` + CLI `--mode` 旗标 + webui Run 分发

**目标**：前端契约增加运行模式；`--frontend=webui --mode=connect` 进入客户端模式（S5-3 实现），其余前端忽略 `--mode`。

**涉及文件**：

| 文件 | 动作 |
|---|---|
| `internal/frontend/frontend.go` | `LaunchOptions` 加 `Mode` 字段 + `Mode` 类型与常量 |
| `internal/commands/options.go` | `GlobalOptions` 加 `Mode string` |
| `cmd/musicfox.go` | `--mode` 旗标 |
| `internal/commands/netease.go` | `runPlayer` 传 `Mode`；非法值报错 |
| `internal/webui/register.go` | `Run` 接收 `opts` |
| `internal/webui/run.go` | 按 mode 分发（standalone 现流程 / connect 调 `connectRun`） |

**改动内容（签名级）**：

```go
// frontend.go
// Mode 是前端运行模式（当前仅 webui 前端消费；其它前端忽略）。
type Mode string

const (
    ModeStandalone Mode = "standalone" // 默认：自建 engine + 本地服务
    ModeConnect    Mode = "connect"    // 连接本地 headless daemon，不建 engine
)

type LaunchOptions struct {
    Once  string
    Debug bool
    Pprof bool
    Mode  Mode
}
```

- `commands/options.go`：`GlobalOptions` 加 `Mode string`。
- `cmd/musicfox.go`：`gf.StrOpt(&GlobalOptions.Mode, "mode", "", "", "frontend run mode: standalone|connect (default standalone; webui only)")`。
- `netease.go` `runPlayer`：

```go
mode := frontend.Mode(GlobalOptions.Mode)
switch mode {
case "", frontend.ModeStandalone, frontend.ModeConnect:
default:
    return fmt.Errorf("未知模式 %q（可用: standalone|connect）", GlobalOptions.Mode)
}
return fe.Run(context.Background(), frontend.LaunchOptions{
    Once:  GlobalOptions.Once,
    Debug: GlobalOptions.DebugMode,
    Pprof: GlobalOptions.PProfMode,
    Mode:  mode,
})
```

- `webui/register.go`：`Run(ctx, opts)` 转调 `RunWithOptions(ctx, opts)`。
- `webui/run.go`：现有 `Run` 提为 `runStandalone(ctx, opts)`；入口：

```go
// RunWithOptions 按 mode 分发：standalone 自建 engine（现状）；connect
// 连接本地 headless daemon（S5-3）。
func RunWithOptions(ctx context.Context, opts frontend.LaunchOptions) error {
    if opts.Mode == frontend.ModeConnect {
        return connectRun(ctx)
    }
    return runStandalone(ctx, opts)
}
```

**依赖**：无前置，与 S5-2 并行。

**验证**：
- `go build/vet ./...`；`go test ./internal/commands/... ./internal/frontend/...`。
- 冒烟：`--frontend=webui`（默认 standalone）行为零变化；`--mode=connect`（S5-3 前）报「connect 模式未实现」或走 connectRun 占位错误；`--mode=bogus` 报未知模式且退出码非 0；`--frontend=tui --mode=connect` 忽略 mode 正常进 TUI。

**风险**：低。`--mode` 是纯增量旗标；run.go 的拆分需保证 standalone 流程逐行等价（建议拆分为纯重构 commit 先行，mode 分发后置）。

**写权限**：fixer。

---

### S5-2｜`webui.Server` 后端抽象：`Backend` 接口 + `localBackend` + 认证可配置

**目标**：把 `Server` 从「绑死 `*core.Engine`」重构为「依赖 `Backend` 接口」，使同一 Server 能服务本地 engine（standalone/GUI）与远程 daemon 客户端（connect）。**认证层可配置**（`ServerOptions.Auth`）为 GUI 的 AssetServer 方案留口（D-GUI-2 回退路径）。

**涉及文件**：

| 文件 | 动作 |
|---|---|
| 新增 `internal/webui/backend.go` | `Backend` 接口 + `localBackend` + `ServerOptions` |
| `internal/webui/server.go` | 删 `engine`/`dispatcher` 字段；`NewServerWithBackend`/`NewServerWithOptions`；路由按 `Auth` 条件注册 |
| `internal/webui/ws.go` | `buildSnapshot`/`serveWS` 改走 `backend` |
| `internal/webui/api.go` | `handleAlbumArt`/`handleLyrics` 改走 `backend` |
| `internal/webui/commands.go` | engine 直用 → `backend.Ready()`/`backend.CommandContext()` |
| `internal/webui/api_login.go` | engine nil 检查 → `backend.Ready()`（connect 模式 503） |

**改动内容（签名级）**：

```go
// backend.go
package webui

// Backend 抽象 WebUI 的数据源：本地 engine（standalone / GUI）或远程
// headless daemon 客户端（connect）。Server 只依赖本接口，不持有
// *core.Engine。全部方法均为纯查询/转发，不维护播放状态。
type Backend interface {
    // Ready 报告后端可用（engine 非 nil / daemon 已连接）。不可用时
    // 辅助端点降级（404/503/空数据），HTTP 层不 panic。
    Ready() bool
    // Dispatch 执行控制命令（本地 Dispatcher 或远程 CtrlClient.Call）。
    Dispatch(ctx context.Context, cmd string, args map[string]any) (any, error)
    // SubscribeEvents 注册事件监听并返回退订函数。handler 的 name 为
    // WebUI 帧名（song_changed 等），payload 为已序列化事件帧
    // （{"type":"event",...}，与 frontend.EventFrame 同构）。
    SubscribeEvents(handler func(name string, payload []byte)) func()
    // Playlist 返回精简播放列表（id/name/artist/album），快照用。
    Playlist() []map[string]any
    // PlayingInfo 返回当前歌曲封面 URL；ok=false 表示不可用（connect 模式
    // 恒 false，daemon status 快照不含 PicUrl）。
    PlayingInfo() (picURL string, ok bool)
    // LyricState 返回歌词状态；connect 模式返回空结构。
    LyricState() (fragments []lyricFragment, translated map[int64]string, currentIndex int, offsetMs int64)
    // CommandContext 返回命令上下文快照（命令端点用）。
    CommandContext() frontend.CommandContext
}

// localBackend：绑定 *core.Engine 的实现，行为与现状逐字一致。
type localBackend struct{ engine *core.Engine }

func (b *localBackend) Ready() bool                     // engine != nil
func (b *localBackend) Dispatch(...) (any, error)       // core.NewDispatcher(b.engine).Dispatch
func (b *localBackend) SubscribeEvents(handler) func()  // 复用 subscribeEmitter（events.go）
func (b *localBackend) Playlist() []map[string]any      // engine.Player().Playlist() 精简映射
func (b *localBackend) PlayingInfo() (string, bool)     // engine.Player().PlayingInfo().PicUrl
func (b *localBackend) LyricState() (...)               // engine.LyricService().State() 映射
func (b *localBackend) CommandContext() frontend.CommandContext // engine.Player().CommandContext()

// ServerOptions 是 Server 构造选项。
type ServerOptions struct {
    // Auth 控制认证层（token 交换 + cookie + Origin 校验）是否启用。
    // true=现状（standalone/connect）；false=关闭（GUI AssetServer
    // 方案：页面经内置 scheme 加载，无法做 cookie 交换）。
    Auth bool
}
```

- `server.go`：
  - `NewServer(engine *core.Engine) *Server` 保留（`NewServerWithOptions(&localBackend{engine}, ServerOptions{Auth: true})` 的便捷包装，standalone 调用点零改动）。
  - 新增 `NewServerWithBackend(backend Backend) *Server` 与 `NewServerWithOptions(backend Backend, opts ServerOptions) *Server`。
  - 删字段 `engine *core.Engine`、`dispatcher *core.Dispatcher`，加 `backend Backend`。
  - `mux` 注册：`Auth==true` 时 `/api/*` 走 `authMiddleware`、`/ws` 走 `verifyWSRequest`（现状）；`Auth==false` 时直接挂 handler。
  - 事件面订阅：Server 在 `NewServerWithOptions` 内调用 `backend.SubscribeEvents`（把事件转发到 broadcaster）并保存返回的 unsubscribe 供 `Close` 幂等清理；若 backend 自身已维护订阅生命周期则返回 no-op。**具体接线语义留 S5-3 与 remoteBackend 一起定稿，本票以接口契约为准**。
- `ws.go` `buildSnapshot`：`status` 经 `s.backend.Dispatch(ctx, "status", nil)`；`playlist` 经 `s.backend.Playlist()`；删除 engine 直用。
- `api.go`：`handleAlbumArt` → `picURL, ok := s.backend.PlayingInfo()`，`!ok` → 404；`handleLyrics` → `s.backend.LyricState()`。
- `commands.go`：`!s.backend.Ready()` → 500（现状语义）；`ctx := s.backend.CommandContext()`。
- `api_login.go`：`!s.backend.Ready()` → 503「connect 模式不支持登录」（D-S5-2）。

**依赖**：无前置（独立重构票）。**注意**：本票是 GUI 的硬前置（GUI 复用 `NewServerWithOptions` + `localBackend` + `Auth` 可配置）。

**验证**：
- `go build/vet ./...`；`go test ./internal/webui/...`（现有 webui 测试全绿 = 行为零变化守护）。
- 冒烟：`--frontend=webui` 全功能（登录/播放/事件/命令/专辑图/歌词）与重构前一致。
- `Auth=false` 分支单测：`httptest` 断言路由可达、`verifyWSRequest` 跳过。

**风险（S5 最高风险票）**：重构面 6 文件。对策：① 分步提交（每步编译绿）；② `NewServer(engine)` 便捷包装保 standalone 调用点零改动；③ 现有 webui 测试 + 冒烟兜底；④ `SubscribeEvents` 接线语义（谁负责订阅/退订）在 S5-3 与 remoteBackend 一起定稿，本票以接口契约为准。

**写权限**：fixer（reviewer 复核 handler 注册的 Auth 条件分支与事件订阅的 Close 清理）。

---

### S5-3｜daemon 订阅客户端（`headless.SubscribeClient`）+ `remoteBackend` + `connect.go`

**目标**：实现 connect 模式的完整链路。`internal/headless` 新增订阅式长连接客户端（复用 `DialAddr` + subscribe 协议）；`internal/webui` 新增 `remoteBackend`（`Backend` 的远程实现）与 `connectRun`。

**涉及文件**：

| 文件 | 动作 |
|---|---|
| 新增 `internal/headless/subscribe_client.go` | `SubscribeClient`（长连接 + 快照缓存 + 请求/响应关联 + 事件投递） |
| 新增 `internal/webui/remote_backend.go` | `remoteBackend`（Backend 接口远程实现） |
| 新增 `internal/webui/connect.go` | `connectRun`（Dial → remoteBackend → Server → 浏览器 → 阻塞） |
| `internal/webui/run.go` | 抽公共「serve + 打开浏览器 + 等待」供 standalone/connect 共用 |
| `internal/webui/ws.go` | `buildSnapshot` 已走 backend（S5-2），本票核对 connect 快照路径 |

**改动内容（签名级）**：

```go
// headless/subscribe_client.go
package headless

// SubscribeClient 连接 headless daemon 的控制通道并维持订阅会话：首个请求
// subscribe（快照帧 + ack），随后事件帧与请求/响应在同一条连接上交错
// （daemon 写锁保证帧不交错）。比 CtrlClient 多出「长连接 + 事件流」，
// 是 CtrlClient 的订阅态升级（CtrlClient 保留不动，供 musicfox ctrl 使用）。
type SubscribeClient struct {
    network, addr string
    mu     sync.Mutex // 写锁：subscribe/Call/unsubscribe 与读循环不交错
    conn   net.Conn
    enc    *json.Encoder
    dec    *json.Decoder
    idSeq  atomic.Int64
    // events 投递通道（buffered，满则丢帧并记日志——镜像 daemon 侧
    // 「慢连接只丢自己帧」纪律）；关闭时通道关闭。
    events chan []byte
    // snapshot 最近一次快照帧的 data（status + playlist 精简字段）。
    snapshotMu sync.RWMutex
    snapshot   map[string]any
    // pending 进行中的请求 id → 响应通道（Call 阻塞等待）。
    pendingMu sync.Mutex
    pending   map[int64]chan *core.Response
}

// DialSubscribe 解析 daemon 地址（DialAddr）、探测存活并以 events 发起
// subscribe（wire 名见 core.Ev* 常量）。无 daemon 时报错（文案对齐
// CtrlClient：「headless daemon 未在运行」）。
func DialSubscribe(events []string) (*SubscribeClient, error)

// Call 在订阅连接上执行一条控制命令（镜像 daemon 的 dispatch 语义），
// 响应按 ID 关联，3s 截止（对齐 CtrlClient）。未订阅的"quit"仍由
// 传输层语义处理：本客户端不发送 quit（D-S5-2）。
func (c *SubscribeClient) Call(ctx context.Context, cmd string, args map[string]any) (*core.Response, error)

// Snapshot 返回最近一次快照 data（可能为 nil：订阅 ack 前）。
func (c *SubscribeClient) Snapshot() map[string]any

// Events 返回事件帧投递通道（快照帧也经此投递一次）。
func (c *SubscribeClient) Events() <-chan []byte

// Close 关闭连接、关闭 Events、唤醒全部 pending caller（返回错误）。
func (c *SubscribeClient) Close() error
```

- 读循环：`json.Decoder` 逐帧解码——`core.Response`（ID 匹配 pending）、`{"type":"snapshot","data":{...}}`（缓存 + 投递）、`{"type":"event",...}`（投递）。
- 断线：读错误 → `Close` 语义（Events 关闭、pending 唤醒）。

```go
// webui/remote_backend.go
package webui

// remoteBackend 是 Backend 的远程实现：数据源 = 本地 headless daemon。
// 控制面经 SubscribeClient.Call；事件面经订阅长连接；快照/状态从
// SubscribeClient.Snapshot 缓存读取。
type remoteBackend struct {
    client    *headless.SubscribeClient
    unsub     func() // SubscribeEvents 的退订（= 停读循环）
    posThrottle *positionThrottle // 复用 events.go
}

// newRemoteBackend 构造并启动事件消费循环。
func newRemoteBackend(client *headless.SubscribeClient) *remoteBackend

func (b *remoteBackend) Ready() bool              // client 未关闭
func (b *remoteBackend) Dispatch(ctx, cmd, args) (any, error) // client.Call → Data/Error
func (b *remoteBackend) SubscribeEvents(handler func(name string, payload []byte)) func()
func (b *remoteBackend) Playlist() []map[string]any        // Snapshot()["playlist"]
func (b *remoteBackend) PlayingInfo() (string, bool)       // 恒 ("", false)（D-S5-2）
func (b *remoteBackend) LyricState() (...)                 // 空结构
func (b *remoteBackend) CommandContext() frontend.CommandContext // Snapshot() 的 status 字段映射
```

- 事件映射：订阅 wire 名 = `eventWireToFrame` 的全部 key（`core.EvSongChanged` 等 5 个）；转发时 `eventWireToFrame[wire]` 映射到 webui 帧名；`position` 经 `positionThrottle` 节流（复用 events.go 现成实现）。
- 快照路径（裁决）：webui 的 `buildSnapshot` 保持「`Dispatch("status")` + `Playlist()`」同构组装（两实现一致，避免快照双源）；`remoteBackend.Playlist` 从 `Snapshot()` 缓存取——**订阅 ack 前 WS 可能已连接**，此时 playlist 为空数组（前端幂等，不崩）。

```go
// webui/connect.go
package webui

// connectRun 以客户端模式运行 WebUI：连接本地 headless daemon，不建
// engine、不加载 WASM。播放/状态/事件经 daemon 转发；命令面为空、登录
// 503、albumart 404、lyrics 空结构、WS quit 不转发（D-S5-2）。
func connectRun(ctx context.Context) error {
    client, err := headless.DialSubscribe(eventWireNames()) // wire 名集合
    if err != nil {
        return fmt.Errorf("connect 模式需要 headless daemon 正在运行: %w", err)
    }
    defer client.Close()
    backend := newRemoteBackend(client)
    server := NewServerWithOptions(backend, ServerOptions{Auth: true})
    return runServer(ctx, server) // 抽自 run.go 的公共「Serve + open browser + wait」
}
```

- `run.go`：抽 `runServer(ctx, server) error`（`server.Serve` goroutine + `waitReady` + `open.Start(token URL)` + 信号/`ShutdownCh` select + `server.Close`）。standalone 流程 = 现 Run 主体（engine + wasm scope + startup + `runServer`）；connect 流程 = `connectRun`。

**依赖**：S5-1（`--mode` 入口）+ S5-2（`NewServerWithOptions`/`Backend`）。

**验证**：
- `go build/vet ./...`；`go test ./internal/headless/... ./internal/webui/...`。
- 单测/集成测试：`SubscribeClient`（假 daemon：临时 socket + 现有 `NewServerWithAddr` 起 server，验证快照缓存/事件投递/Call 关联/断线 Close）；`remoteBackend`（事件名映射、position 节流、Snapshot 数据源）。
- 冒烟清单见 S5-4。

**风险**：
- **断线语义**：daemon 重启后 `SubscribeClient` 读错误 → `Ready=false` → 辅助端点降级 + 浏览器事件静止。**重连不做**（MVP，文档注明；`musicfox ctrl` 也是每次调用新建连接，天然容错，connect 长连接的重连留后续）。
- 快照与订阅的先后：daemon 侧「快照先于注册」保证首个事件不早于快照；webui 侧 WS 快照独立组装（`Dispatch("status")`），与事件流竞态由前端幂等兜底（现状既有语义）。
- `headless` 包新增导出 API 面扩大：`SubscribeClient` 仅被 webui 消费，命名与注释标明用途。

**写权限**：fixer。

---

### S5-4｜测试 + 文档（connect 功能边界、AGENTS.md、docs）

**目标**：S5 收尾——Backend 双实现测试、connect 集成测试、文档同步。

**涉及文件**：

| 文件 | 动作 |
|---|---|
| 新增 `internal/webui/backend_test.go` | Backend 接口双实现（local/remote）语义测试 + httptest 端点降级测试 |
| 新增 `internal/webui/connect_integration_test.go` | headless server（临时 socket）+ connect server 全链路 |
| `internal/webui/run.go` / `connect.go` 注释 | 功能边界表落注释 |
| `AGENTS.md` | webui/headless/frontend 段补 Mode、Backend、connect |
| `docs/frontend_plugin.md` | 新增 C10「WebUI connect 客户端模式」段 |
| `README.md` | `--mode` 旗标 |

**改动内容**：
- 测试要点：
  1. `localBackend` 与 `remoteBackend` 对 `Dispatch("status")`/`Playlist`/`CommandContext` 的语义一致性（共享断言集）。
  2. `Auth=false` 端点可达性（httptest）。
  3. connect 集成：`NewServerWithAddr(engine, "unix", tmpSocket)` 起 daemon → `DialSubscribe` → `NewServerWithOptions(remoteBackend, Auth:true)` → httptest 断言 `/api/status` 反映 daemon 状态、WS 快照 + 事件帧（切歌 → `song_changed`）到达、`/api/commands` 空、`/api/albumart` 404、`/api/lyrics` 空结构、`/api/login/qr/*` 503。
- 功能边界表（写入 connect.go 头注释 + 文档）：

| 能力 | standalone | connect |
|---|---|---|
| 播放控制（play/next/seek/volume/...） | Dispatcher | daemon 转发 ✅ |
| 状态/快照 | engine | daemon status + playlist ✅ |
| 事件推送（song/state/position/startup/login） | 本地 EventEmitter | daemon 订阅映射 ✅ |
| 命令面 `/api/commands` | 本地命令 | 空列表（daemon 不注册命令） |
| 登录（QR） | 本地 | 503 |
| `/api/albumart` | engine | 404（快照无 PicUrl） |
| `/api/lyrics` | engine | 空结构 |
| WS `quit` | 关自身 | 不转发（防误关 daemon） |

- 文档：`frontend_plugin.md` C10 段（形态/功能边界/与 `musicfox ctrl` 的关系）；`AGENTS.md`（webui 段补 Mode 分发与 Backend 抽象、headless 段补 SubscribeClient、frontend 段补 `LaunchOptions.Mode`）；README 补 `--mode`。

**依赖**：S5-2 + S5-3。

**验证**：
```
冒烟清单（S5）：
1. musicfox --headless &（daemon 常驻）
2. musicfox --frontend=webui --mode=connect
3. 浏览器打开 → 页面显示 daemon 当前播放状态/播放列表
4. 页面控制 play/next/pause/volume → daemon 播放器实际变化
5. 手工切歌（musicfox ctrl next）→ 页面实时刷新（song_changed）
6. /api/albumart 404（前端封面降级不崩）；/api/commands 空列表
7. 无 daemon 时 --mode=connect 报错退出（非 0），提示 daemon 未运行
8. --frontend=webui（standalone）全功能回归（登录/播放/事件）
9. Windows 平台：daemon TCP 端口文件路径冒烟
```

**风险**：低（行为由测试锁定）。文档腐化是主要风险——功能边界表必须与实现一致。

**写权限**：fixer。

---

## 3. GUI：Wails v2 原生窗口前端

### 3.1 调研结论（lib-1 已核实，直接采用）

- **版本**：Wails v2 为当前稳定版（v2.15.0，2026-08-20），官方承诺继续接收修复；v3 仍 beta 且为 port 级迁移——**本项目选 v2**。
- **Go 兼容**：v2.15 模块 go 1.25，项目 go 1.26 兼容 ✅。
- **三平台**：macOS 需 Xcode CLT（10.15+/11.0+）；Linux 需 gcc + GTK3 + WebKit2GTK 4.1（**依赖重是最大坑**；无桌面环境可构建不可运行）；Windows 需 WebView2 Runtime（Win11 自带）。**CGO 强依赖，交叉编译受限**（每平台原生构建 / CI matrix）。
- **产物**：单二进制（go:embed 前端资源 + Windows .syso）；`wails build` 默认 `-tags desktop,production`；goreleaser 经 `builds[].hooks.post` 调 wails build。
- **集成模式**：AssetServer.Handler 接受任意 `http.Handler`（官方 vanilla + Dynamic Assets 指南即此模式）。**外部 URL 页面拿不到 Wails runtime/IPC（issue #4686）→ 非 Bind 模式：页面独立走 HTTP/WS**。
- **项目组织**：v2 硬约束「main 包须与 wails.json 同目录」仅在 **wails CLI** 路径下生效；**Manual Builds 官方支持跳过 wails CLI 直接 `go build`**——方式 B 与 go-musicfox 架构最契合（`internal/frontend/gui/` 实现 Frontend 接口，`--frontend=wails` 选择，不动 commands/注册表/播放引擎）。

### 3.2 目标与定位

新增 `"wails"` 前端：原生窗口内加载 WebUI 页面（复用 S5 的 Backend 抽象 + webui.Server + 全套页面/WS/API 资产）。**定位**：浏览器体验增强形态（原生窗口、常驻托盘、后续可加 Bind），macOS/Windows 优先交付；Linux 无桌面环境自动回退系统浏览器 WebUI。

### 3.3 Ticket 拆分

```
GUI-1（spike + gui 骨架）────────── 无前置
GUI-2（Wails 窗口集成 + Linux 回退）┐ 依赖 GUI-1（spike 结论）
GUI-3（构建与分发）────────────────┤ 依赖 GUI-1
GUI-4（测试 + 文档）───────────────┘ 依赖 GUI-2 + GUI-3
```

### 3.4 窗口加载路径裁决（D-GUI-2，spike 验证点）

- **GUI-B（推荐）**：窗口 Navigate 到本地 server 的 `http://127.0.0.1:<port>/token?token=...`（与 standalone 打开浏览器同一 URL），认证链路**零改动**，`ServerOptions{Auth: true}`。
- **GUI-A（回退）**：AssetServer.Handler = server mux，页面经内置 scheme 加载；认证降级 `ServerOptions{Auth: false}`（S5-2 已留口），token/cookie 链路绕过（窗口内页面无跨进程暴露风险）。
- **spike 判定点**：Wails v2 主窗口/`WebviewWindow` 能否 Navigate 外部 http URL（API 形状与事件语义）；`wails.Run` 在 `Assets:nil` 时是否可运行；manual build（`go build -tags desktop,production`）不依赖 wails CLI 生成物。

---

### GUI-1｜spike + `internal/frontend/gui` 骨架

**目标**：验证 Wails v2 + manual build 在 go-musicfox 架构下的关键不确定点（D-GUI-1/D-GUI-2），产出骨架（`Frontend` 接口注册 + 平台文件布局），spike 结论落文档。

**涉及文件**：

| 文件 | 动作 |
|---|---|
| 新增 `internal/frontend/gui/register.go` | `guiFrontend`（ID `"wails"`），`Run` 占位（spike 完成前返回「未实现」） |
| 新增 `internal/frontend/gui/spike.md`（或 docs 内） | spike 结论记录 |
| `internal/frontend/registration/registration.go` | 空导入 `gui` |
| 新增 `internal/frontend/gui/wails.json`（spike 定） | 若 manual build 需要 |
| `cmd/musicfox.go` | 无改动（聚合器已空导入） |

**spike 验证清单（写进票，逐项过）**：
1. Wails v2.15 在本机（Go 1.26）`go build -tags desktop,production ./internal/frontend/gui` 可编译、可运行（最小窗口）。
2. `wails.Run` 无 AssetServer（`Assets:nil, Handler:nil`）时的行为（panic 与否）——决定 GUI-B 是否可只用窗口管理。
3. 窗口 Navigate 外部 http URL：`WebviewWindow`/主窗口的 API 形状、`wails.Run` 内何时 Navigate（`OnStartup`?）、页面能否正常加载本地 HTTP（WKWebView/WebView2 的 localhost 访问限制）。
4. macOS 无 Xcode 环境、Linux 无 DISPLAY 时构建/运行行为（构建可过、运行报错 → 回退判定）。
5. wails.json 在 manual build 下的作用（版本注入缺失是否影响 `wails.Run`）。
6. 与既有前端共存：`--frontend=wails` 经 `resolveFrontendID` 选择后，`bootstrap()` 与 `--once` fail-fast 约束（`--once` 非 headless fail-fast，GUI 天然满足）。

**改动内容（签名级）**：

```go
// register.go
package gui

// guiFrontend 是 Wails v2 原生窗口前端（Frontend 接口实现）。
// 窗口内加载 WebUI 页面（复用 webui.Server + Backend 抽象），
// 非 Bind 模式：页面经 HTTP/WS 与后端通信，不依赖 Wails runtime。
type guiFrontend struct{}

func (guiFrontend) ID() string   { return "wails" }
func (guiFrontend) Name() string { return "GUI" }
func (guiFrontend) Run(ctx context.Context, opts frontend.LaunchOptions) error {
    // GUI-1 占位：spike 结论落定后由 GUI-2 实现。
    return errors.New("wails frontend not wired yet")
}

func init() { frontend.Register(guiFrontend{}) }
```

**依赖**：无前置（GUI 的骨架票）；Wails v2 依赖引入（`go.mod` + vendor）在本票完成。

**验证**：spike 清单逐项过、结论落 `spike.md`；`go build ./...`；`go test ./internal/frontend/...`（注册表含 `wails`）。

**风险**：spike 任一关键点失败 → 回到 3.4 的回退路径或调整方式 B 细节；依赖引入（wails + 平台绑定）需 `make vendor` 更新。

**写权限**：fixer。

---

### GUI-2｜Wails 窗口集成 + Linux 回退

**目标**：`guiFrontend.Run` 完整实现：core engine + Startup + `localBackend` 注入的 webui Server + Wails 窗口（GUI-B 优先）；Linux 无桌面环境回退「系统浏览器打开 WebUI」。

**涉及文件**：

| 文件 | 动作 |
|---|---|
| 新增 `internal/frontend/gui/run.go` | `Run(ctx)` 主流程（引擎 + server + 窗口生命周期） |
| 新增 `internal/frontend/gui/run_linux.go` | Linux 桌面环境检测与回退 |
| 新增 `internal/frontend/gui/run_darwin.go` / `run_windows.go` | 平台窗口细化（如需要） |
| `internal/webui/server.go` | `Handler() http.Handler` 导出（GUI-A 时取 mux）——**若 GUI-A 成立** |
| 新增 `internal/frontend/gui/run_test.go` | 可测逻辑（回退判定等） |

**改动内容（签名级）**：

```go
// run.go
package gui

// Run 以 Wails 原生窗口运行 go-musicfox：
// 1. 平台检测（Linux 无 DISPLAY/WAYLAND_DISPLAY → 回退系统浏览器 WebUI）
// 2. core.NewEngine + engine.Startup（noop observer，镜像 webui）
// 3. webui.NewServerWithOptions(localBackend, ServerOptions{Auth: 依路径})
// 4. Wails 窗口：GUI-B = Navigate 到 /token URL；GUI-A = AssetServer.Handler
// 5. 窗口关闭 / ctx 取消 → server.Close + engine.Close
func Run(ctx context.Context) error
```

- 平台检测（`run_linux.go`，镜像 `login_webview_linux.go` 的降级哲学）：
```go
// 无桌面环境：回退系统浏览器打开 WebUI（复用 standalone 行为）。
func desktopAvailable() bool // DISPLAY != "" || WAYLAND_DISPLAY != ""
```
- 回退实现：`desktopAvailable() == false` 时 `Run` 转调 `webui.RunWithOptions(ctx, frontend.LaunchOptions{Mode: frontend.ModeStandalone})`——**复用现有「系统浏览器打开」路径，零新代码**。
- 窗口生命周期：`wails.Run(&wails.Options{Width, Height, Title, AssetServer: <依 spike>, OnStartup/OnShutdown, OnDomReady: Navigate（GUI-B）})`；`OnShutdown` 里 `server.Close()` + `engine.Close()`。
- 引擎装配与 webui standalone 共用模式：`core.NewEngine(core.EngineOptions{})`；wasm scope 加载 WASM 插件命令（`LoadIntoScope` + 复用 `webuiWasmSink`（已导出）或 gui 包内实现同形 sink；命令出现在页面 `/api/commands`）。

**依赖**：GUI-1 + **S5-2**（`NewServerWithOptions`/`localBackend`/`Auth`）。

**验证**：
```
冒烟清单（GUI）：
1. macOS：--frontend=wails → 原生窗口打开 → 登录/播放/事件/命令可用
2. Windows：同（WebView2 Runtime）
3. Linux 桌面：窗口打开；Linux 无桌面（CI/SSH）：回退系统浏览器 WebUI
4. 窗口关闭 → engine/server 干净退出（无泄漏日志）
5. --frontend=wails --once xxx → fail-fast（--once 仅 headless）
6. GUI-A 路径（若回退）：认证降级后页面正常、无 401
```

**风险**：Wails 窗口生命周期与 engine 生命周期的耦合（窗口关闭回调与 `engine.Close` 的竞态——参考 webui standalone 的 `ShutdownCh` 模式，Server 的 `ShutdownCh` 与窗口关闭事件做 select）；Linux WebKitGTK 依赖缺失时的启动失败兜底（回退判定 + 错误提示）。

**写权限**：fixer。

---

### GUI-3｜构建与分发（Makefile / goreleaser / CI matrix）

**目标**：把 GUI 纳入可发布构建链。**方式 B 语义**：不引入 wails CLI 作为必需构建工具（开发可手动 `go build`），但发布构建允许 wails CLI。

**涉及文件**：

| 文件 | 动作 |
|---|---|
| `Makefile` | `build-gui` target（`go build -tags desktop,production`，按 OS 条件分支） |
| `internal/frontend/gui/wails.json` | 确认/修正（manual build 与 CLI build 双路径） |
| `.goreleaser.yaml`（若有） | builds[].hooks.post 或直接 go build；平台 matrix |
| `docs/`（README / 构建说明） | GUI 构建与运行依赖说明 |

**改动内容**：
- `Makefile`：新增 `build-gui`（`go build -tags desktop,production -o bin/musicfox-gui ./cmd/musicfox`）与平台条件（Windows `.syso` 资源、`-ldflags` 版本注入）；**遵循 AGENTS.md 的 Windows 兼容清单**（`$(OS)`/`Windows_NT` 条件、`where` vs `which` 等）。
- 产物：单二进制（Wails 运行时 + go:embed 前端资源）；NSIS 安装器可选（不默认）。
- goreleaser：CGO 强依赖 → **每平台原生 runner 构建**（macOS runner / Windows runner / Linux runner），不承诺交叉编译；`builds[].hooks.post` 调 `wails build`（CLI 路径）或 `go build`（manual 路径），二选一并文档化。
- Linux 分发：运行时依赖 `libgtk-3-0` + `libwebkit2gtk-4.1-0`（Debian/Ubuntu 包名），无托盘需求不装 libayatana-appindicator；文档化「无桌面环境回退浏览器」。

**依赖**：GUI-1（wails.json/依赖落地）。

**验证**：三平台原生 `make build-gui` 成功；产物冒烟（GUI-2 清单）；CI matrix 绿态。

**风险**：Wails 构建的 `.syso`/`-tags` 与项目既有 Makefile 结构冲突；版本注入与 `wails.json` 的不一致（manual build 时 wails 运行时版本信息缺失——spike 已定）。

**写权限**：fixer（Makefile 改动需 reviewer 按 Windows 兼容清单复核）。

---

### GUI-4｜测试 + 文档

**目标**：GUI 收尾——逻辑测试、冒烟矩阵落文档、AGENTS.md/docs 同步。

**涉及文件**：

| 文件 | 动作 |
|---|---|
| 新增 `internal/frontend/gui/run_test.go` 扩展 | 回退判定、引擎装配可测部分 |
| `AGENTS.md` | 前端注册表段落 + GUI 前端说明（含依赖与回退） |
| `docs/frontend_plugin.md` | 新增 C11「Wails 原生窗口前端」段 |
| `README.md` | `--frontend=wails`、构建/依赖说明 |

**改动内容**：测试覆盖 `desktopAvailable` 判定（env 注入）、回退分支（`Run` 在无桌面环境转调 webui）；文档记录三平台依赖表、构建命令、回退行为；冒烟矩阵（GUI-2 清单）落文档。

**依赖**：GUI-2 + GUI-3。

**验证**：`go build/vet/test ./...` 全绿；三平台冒烟记录。

**风险**：低。GUI 的逻辑面（Backend 抽象、Server）已由 S5 测试覆盖，GUI 层主要是窗口与生命周期，测试以冒烟为主——文档需如实标注「GUI 逻辑层测试薄、冒烟为重」的边界。

**写权限**：fixer。

---

## 4. 风险与对策总表

| # | 风险 | 阶段 | 影响 | 对策 |
|---|------|------|------|------|
| 1 | **foxful-cli vendor 扩展**（`UnknownMsgHandler`）：框架改动，需独立 review 与依赖更新流程 | S1 | 中 | ① nil 分支逐字节等价（现有测试守护）；② 单独 commit + 走 foxful 更新流程；③ 回退 D-S1-1：声明式 `Command.ResultAction`（仅编译期命令） |
| 2 | **view 长文本渲染**：超长 Message / 超宽行拖慢 View | S1 | 低 | 行截断（`MaxWidth`）+ 滚动分页渲染（只渲染可见行），不做整页 lipgloss 大字符串 |
| 3 | **S5 重构回归**：Server 从 engine 解耦波及 6 文件 | S5 | 高 | ① 分步提交每步编译绿；② `NewServer(engine)` 便捷包装保调用点零改动；③ 现有 webui 测试 + 冒烟兜底 |
| 4 | **connect 依赖 unix socket（仅本机）**：跨机器不可达 | S5 | 低 | 定位明确为「本机单实例」；远程访问不在范围（文档标注）；Windows TCP 已有 |
| 5 | **connect 断线/重连**：daemon 重启后事件静止 | S5 | 中 | MVP 不重连（文档注明）；`Ready=false` 时辅助端点降级；重连留后续立项 |
| 6 | **Wails 外部 URL 导航 API 不确定**（GUI-B） | GUI | 高 | **spike 先行**（GUI-1 清单）；回退 GUI-A（AssetServer + `ServerOptions.Auth=false`，S5-2 已留口） |
| 7 | **Wails Linux 依赖重**（GTK3 + WebKit2GTK 4.1） | GUI | 中 | 无桌面环境回退系统浏览器（复用 standalone）；文档化依赖；Linux 非核心交付 |
| 8 | **Wails CGO 交叉编译受限** | GUI | 中 | 每平台原生构建 / CI matrix；不承诺交叉编译；`make build-gui` 平台条件分支 |
| 9 | **Wails 版本升级（v2→v3）**：官方 v2 仍维护但终将迁移 | GUI | 低 | 采用 v2 稳定版；gui 包内隔离 wails 依赖（`internal/frontend/gui` 单包），未来 port 时影响面可控 |
| 10 | **文档腐化**（三阶段 × 多份 docs） | 全部 | 中 | 每票收尾含文档同步义务（见 §6），PR review 含文档检查 |

## 5. 跨阶段依赖与里程碑

```
里程碑       内容                       依赖           验证规划
─────────────────────────────────────────────────────────────────────
M1（S1）     view 独立页面               S1-1→S1-2→S1-3   绿态：go build/vet/test ./...
             （toast + 可滚动页面 +                                   冒烟清单（§1.3）
              S8 留口）
M2（S5）     WebUI connect 模式          S5-1/S5-2 并行           绿态：go build/vet/test ./...
                                        → S5-3 → S5-4    冒烟清单（§2.4）+ standalone 回归
M3（GUI）    Wails 原生窗口前端          GUI-1 → GUI-2/3 → GUI-4   绿态：go build/vet/test ./...
             依赖 S5-2（Backend）                                冒烟矩阵（§3 GUI-2 清单）
```

- **并行建议**：S1 与 S5 完全并行（不同文件集）；S5-1/S5-2 并行；GUI-1（spike）可与 S1/S5 任意阶段并行（唯一前置是 S5-2 的接口形状——spike 不依赖 S5 实现，但 GUI-2 的集成代码依赖 S5-2）。
- **每阶段绿态门槛**：`go build ./...` + `go vet ./...` + `go test ./...` 全绿 + 对应冒烟清单全过，才进入下一阶段。
- **每票 commit 粒度**：单票单 commit（Conventional Commits），重构票（S5-2、S1-2 的 foxful 扩展）建议 reviewer 复核 diff 等价性。

## 6. 文档同步义务（项目文档维护准则硬性要求）

每阶段收尾**必须**完成对应文档更新（PR 包含文档检查项，违反 AGENTS.md 维护准则）：

| 阶段 | 必须更新的文档 | 内容 |
|------|--------------|------|
| S1 | `AGENTS.md`；`docs/plugin_development.md`（875 行）；`docs/frontend_plugin.md`；`docs/plugin_ecosystem.md` | 页面基建段补 `command_view`；WASM view 呈现方式更新；轨 B 消费端补 view 独立页面；S8 锚点标注 `ViewPageContent`/`ViewPageHooks` 就位 |
| S5 | `AGENTS.md`；`docs/frontend_plugin.md`；`README.md` | webui 段补 Mode 分发与 Backend 抽象；headless 段补 `SubscribeClient`；frontend 段补 `LaunchOptions.Mode`；新增 C10 段；`--mode` 旗标 |
| GUI | `AGENTS.md`；`docs/frontend_plugin.md`；`README.md` | 前端注册表段补 `"wails"`；新增 C11 段（依赖表/构建/回退）；README 补 `--frontend=wails` 与构建说明 |

---

## 7. 可选扩展（S5 后续，非本阶段范围）

**daemon 命令面扩展（connect 模式下 `/api/commands` 可用）**：在 `internal/headless/server.go` 传输层特判（仿 subscribe/quit 模式，不进 core.Dispatcher）增加：

- `commands`：返回 `frontend.Commands()` 快照（过滤 `configs.IsPluginEnabled`）——connect 的 `/api/commands` 列表来源。
- `run_command <key>`：`frontend.CommandByKey` → Show 门控 → Run；**exec 拒绝在 daemon 端执行**（Web 面永远禁 exec，与 WebUI standalone 的 `commandExecAllowed=false` 策略一致）；open_url 原样返回（由 WebUI 端 `open.Start`）。

本扩展将把 D-S5-2 的功能边界从「命令面为空」升级为「命令面可用」，并需在 connect 文档中同步安全边界说明。仅在用户对 connect 命令面有明确需求时立项。

---

## 8. TUI-connect（S6）：TUI 遥控 headless daemon

> ✅ **S6 已实施**（TC-1..TC-4 + S6-R1 菜单面修复 + S6-R2 收尾全部落地，分支 `feat/plugin-framework-playback`）。本节「现状事实/约束分析/裁决」在实施时以代码为准；功能边界表（§8.4，R1 后已按「本地浏览照常 / 需登录菜单 toast 降级」拆分，§8.10 TC-5..TC-8 完整能力扩展后再按「登录遥控扫码 / 选歌播放投递」升级）与 `connect.go` 头注释、`plugin_scope.go` 挂载集一致，行为由 `go test ./internal/ui/...`（含 `connect_integration_test.go` 全链路集成测试与 `connect_test.go` 链/装配测试）锁定。MVP 明确降级的扩展项落档见 §8.9 P2 扩展清单。

### 8.1 背景与用户需求

用户提出：「单独的 `--mode=connect` 不能使用 TUI 进行连接吗」。当前 `--mode=connect` 只有 webui 前端消费（D-S5-3：TUI 忽略 `opts.Mode`）。用户诉求明确：**用 TUI 界面遥控一个 headless daemon（常驻播放器）**，而不是每次起一个本地 TUI engine——与 webui-connect 的动机同构，但形态是终端（无浏览器依赖、无 HTTP 层，纯 socket + 订阅协议）。

### 8.2 现状约束分析（TUI 对本地 engine 的绑定点，逐条核实）

TUI-connect ≠ 简单把 `Mode` 传给 `tuiFrontend`。TUI 的渲染面、播放列表面、菜单面都假设本地 engine 存在。以下为逐条绑定点与裁决：

| # | 绑定点 | 位置（已核实） | 性质 | 裁决 |
|---|--------|--------------|------|------|
| B1 | **渲染事件面**：`ui.Player` 经 `core.Observer` 接缝（OnSongChanged/OnStateChanged/OnPosition）驱动重绘 + `renderTicker` | `internal/ui/player.go:212-230`（`NewPlayer` 里 `corePlayer.SetObserver(p)`，player.go:56） | 事件驱动 | **必须换源**：订阅 daemon 事件（快照 + `song_changed`/`state_changed`/`position`）替代本地 Observer 回调；渲染回调从事件消费 goroutine 触发（与 core 播放 goroutine 调 `Rerender` 的既有模式同构，线程安全前提一致） |
| B2 | **播放器状态面**：renderer 组（Lyric/SongInfo/Progress/Cover/Spectrum）读 `core.Player` 的 CurSong/PassedTime/State/Volume/Mode/Playlist | `internal/ui/player.go:28-36`（`playerRendererState` 接口） | 状态查询 | **必须换源**：从 `SubscribeClient.Snapshot()` 缓存 + 事件增量维护「遥控状态缓存」 |
| B3 | **播放控制面**：菜单/操作/快捷键调 `m.Player().Ctrl*`/`PlaySong`/`ReinitializePlaylist`/`SetMode` 等 | 全菜单体系 + `operate.go` + `event_handler.go` | 控制命令 | **必须换源（转发）**：Ctrl* 经 `Call` 转发 daemon（`core.Dispatcher` 命令集已覆盖）；`PlaySong`/`ReinitializePlaylist`/智能模式等**本地建列表类操作**无 daemon 对应命令 → **降级**（见 8.4） |
| B4 | **播放列表面**：CurPlaylist 菜单、正在播放菜单、`LocatePlayingSong` 深度依赖本地 `core.Player.Playlist()` | `internal/ui/player.go:81-138` | 数据面 | **降级**：daemon 快照 playlist 只有精简字段（id/name/artist/album，`headless/server.go:510-518`），无完整 `structs.Song`（PicUrl/Duration/Album 图）——TUI 播放队列 UI 只能显示精简只读列表，**不可用本地完整列表** |
| B5 | **歌词**：LyricRenderer 读 `lyric.Service`（engine 经 `core.ServiceLyric` 注册） | `menu_accessor.go` `LyricService()`；`netease.go` Components | 数据面 | **降级（P2 扩展）**：daemon 事件总线无歌词事件、快照无歌词；TUI-connect 无 engine 即无 lyric 服务。MVP 隐藏歌词区；P2 由 TUI 本地拉取当前歌歌词 + position 事件驱动推进（`lyric.Service` 可脱离 engine 独立构造，需新适配器） |
| B6 | **封面**：CoverRenderer 读 PlayingInfo().PicUrl | `webui/localBackend.PlayingInfo` 同源逻辑；TUI `coverRenderer` | 数据面 | **降级（P2 扩展）**：daemon 快照无 PicUrl（webui-connect 的 albumart 即 404，D-S5-2 同哲学）。P2 可给 daemon 快照加 PicUrl 字段（低成本 daemon 改动） |
| B7 | **频谱**：`MTAudioProcessingTap` PCM 在本地播放引擎 | `spectrum_renderer.go`（仅本地 osx 引擎有 PCM） | 本地能力 | **降级**：connect 模式无本地 PCM → 隐藏频谱组件（renderer 装配层跳过） |
| B8 | **登录态**：`menuServices.User()` 从 engine 的 UserService 解析；菜单登录门控（`CheckUserInfo(User())` → ToLoginPage） | `menu_accessor.go:117-120`；`player.go:159` 等 | 状态查询 | **换源 + 降级**：显示 daemon 用户昵称（`status.user`）；**TUI 侧登录门控禁用**（TUI 无法登录 daemon，webui-connect 同为 503 哲学）；需登录的菜单（我的歌单/收藏）在 connect 模式 toast 降级 |
| B9 | **启动序列**：`InitHook` 跑 `engine.Startup`（jar→用户→播放列表→登录→hooks→自动播放） | `internal/ui/netease.go:205-218` | 生命周期 | **不执行**：connect 不建 engine 不跑 Startup；InitHook 改为「连接 daemon + 建立订阅」 |
| B10 | **命令面（轨 B/WASM）**：TUI 前端 scope 加载 WASM，`CommandContext` 取 `Player().CommandContext()` | `frontend.go`/`command_menu.go:142-145` | 命令面 | **降级（MVP 禁用）**：`CommandContext.UserID` 无法从 daemon 获得（`status.user` 仅昵称）；与 webui-connect「命令面为空」哲学对齐。WASM 不加载、命令菜单不注册；P2 扩展。**S6-R1 收敛语义**：「命令面禁用」≠「业务插件不加载」——connect 壳仍经 `NewConnectFrontendScope` 挂载 8/9 业务插件以保本地浏览菜单树，仅 lastfm 排除（Deps 依赖 engine 服务，connect 无 engine） |

**关键架构裁决（D-TC-1）**：TUI-connect **不做「TUI 整体换源为 remote 数据面」**（方案 A），而是 **「轻量遥控壳」**（方案 B）：

| 维度 | 方案 A：整体换源（TUI 播放层远程化） | 方案 B：轻量遥控壳（MVP，推荐） |
|------|-----------------------------------|-------------------------------|
| 改造面 | `core.Player` 全接口远程化（约 30+ 方法逐一转发或缓存）+ renderer 全量换源 + 全部菜单行为回归验证 | 新增 `ui/remote_player.go`（遥控状态缓存 + Call 转发 + 事件消费）+ `ui.Player` 显式遮蔽 connect 方法 + renderer/访问器降级分支 |
| daemon 协议 | **必须大扩展**：快照需传输完整 `structs.Song`（PicUrl/Duration/Album…），与 webui-connect 的「精简快照」设计矛盾 | **零改动**：复用现快照（status + 精简 playlist）+ 现 Dispatcher 命令集 |
| 工作量 | 15–25 人日 | 6–10 人日 |
| 风险 | 高（TUI 全部菜单/操作行为回归面；缓存一致性难） | 中（降级体验 + 遮蔽方法漏网风险，由测试守护） |
| 用户体验 | 完整（接近本地） | MVP 有明确降级（歌词/封面/选歌播放/频谱），核心遥控能力完整 |

**推荐方案 B 的理由**：① 用户需求是「遥控常驻播放器」——控制 + 状态是核心，浏览/搜索是网易云 API 网络层（与引擎无关），本就该留在本地；② 方案 A 的 daemon 协议大扩展是高投入且与 S5 已定稿的「精简快照」设计方向矛盾；③ 方案 B 与 webui-connect 共享 `SubscribeClient` 数据面（复用最大化）；④ 方案 B 遵循项目「简单性优先」准则——不动既有 TUI 菜单/播放器主路径，改动收敛到新增适配 + 显式降级。

### 8.3 目标与定位

**一句话价值**：`--frontend=tui --mode=connect` 时，TUI 以「遥控壳」形态连接本地 headless daemon——播放控制经 `Call` 转发、播放状态经订阅驱动，浏览/搜索留在本地，无浏览器依赖的终端遥控形态。

**与 webui-connect 的关系**：同是 daemon 客户端（共享 `SubscribeClient` 数据面、共享「本机单实例」定位，对齐 D-S5-2 边界哲学）；差异在消费端——webui 是浏览器富页面（HTTP/WS 层），TUI 是终端（纯 socket + 订阅协议，无 HTTP 层）。

### 8.4 功能边界表（standalone TUI vs TUI-connect 完整能力）

> ✅ TC-1..TC-4 + S6-R1/R2 落地 MVP；§8.10 **TC-5..TC-8 完整能力扩展已落地**——登录（daemon 侧 QR 登录、TUI 遥控扫码）与选歌播放（`play_list` 整列表投递）从 toast 降级升级为可用。TUI-connect 列即 §8.10.2「本扩展后」列的等价表述。

| 能力 | TUI standalone | TUI-connect（完整能力） |
|------|---------------|----------------------|
| 播放控制（next/prev/pause/resume/toggle/stop/seek/volume/repeat/shuffle/like/dislike） | 本地 engine | **`Call` 转发 daemon** ✅ |
| 播放状态 / 当前歌曲 / 进度 | 本地 | **订阅事件 + 快照缓存** ✅ |
| 搜索 / 排行榜 / 精选歌单 / 专辑 / 歌手 / DJ 浏览（无需登录） | 本地网易云 API | **本地照常** ✅（8 个业务插件挂载，菜单完整，S6-R1） |
| 需登录浏览（收藏 / 我的歌单 / 云盘 / 每日推荐 / 最近播放 / 私人FM） | 本地网易云 API + 登录 | **无条件 toast 降级**（本地 API 无登录 cookie，跨进程登录态墙；对齐 B8。P2：cookie 回拉） |
| 业务插件菜单面 | 9 插件全挂载（frontend scope） | **8/9 挂载**（lastfm 除外——Deps 依赖 engine 服务，connect 无 engine；S6-R1） |
| 浏览菜单的播放动作（选中 → PlaySong） | 本地建列表 + 播放 | **`play_list` 整列表投递 daemon** ✅（`ui.Player.PlaySong` → `CallPlayList([]Song{song},0,true)`，TC-7） |
| 播放队列显示 | 本地完整列表 | 快照**精简只读**列表（id/name/artist/album）△ + **投递后响应同步**（`play_list` 响应 playlist 写回缓存，next/prev 正确） |
| 选歌播放（菜单选中 → PlaySong） | 本地 | **`play_list` 整列表投递 daemon** ✅（PlaySong/ReinitializePlaylist/StartPlay 遮蔽升级，单曲点播为 index=0 特例，D-TC-9） |
| 播放模式/音量显示 | 本地 | 快照字段 ✅ |
| 登录 | 本地 | **daemon 侧 QR 登录、TUI 遥控扫码** ✅（ToLoginPage → connect LoginPage 只显 QR 入口 → CallQRKey/CallQRStatus；EvLogin 驱动用户态刷新，D-TC-7/TC-6） |
| 用户态展示 | 本地 user | **昵称 + UserId**（`status.userId` 快照幂等 + EvLogin 增量；门控仍剥离 UserId，D-TC-8） |
| 播放队列编辑 / 智能模式 / 心动 / 桌面歌词 | 本地 | **禁用 toast**（P2 扩展） |
| 歌词 | 本地 LyricService | **降级**：隐藏（P2：本地拉取 + position 推进） |
| 封面 | 本地 | **降级**：无（P2：daemon 快照加 PicUrl） |
| 频谱 | 本地 PCM | **不可用**（组件隐藏） |
| 命令面（轨 B / WASM） | 本地执行 | **禁用**（`CommandContext.UserID` 不可得，对齐 webui-connect 空命令面；P2 扩展） |
| 断线语义 | — | daemon 断开 → 订阅 Events 关闭 → toast 提示 + 状态降级；**不自动重连**（MVP，对齐 webui-connect） |

### 8.5 Ticket 拆分

```
TC-1（Mode 接入 tuiFrontend + connect 装配骨架）───── 无前置
TC-2（RemotePlayer 数据面：订阅+缓存+转发+事件投递）──┐ 独立可并行（被 TC-3 装配调用）
TC-3（装配 + 降级：renderer/访问器/遮蔽方法）────────┘ 依赖 TC-1 + TC-2
TC-4（测试 + 文档 + P2 扩展清单）───────────────────── 依赖 TC-2 + TC-3
```

---

### TC-1｜`--mode=connect` 接入 `tuiFrontend` + connect 装配骨架

**目标**：D-S5-3 修订——TUI 不再忽略 `opts.Mode`；`--frontend=tui --mode=connect` 进入遥控壳装配（先探测 daemon，失败 fail-fast 提示）；standalone 路径逐字节不变。

**涉及文件**：

| 文件 | 动作 |
|---|---|
| `internal/ui/frontend.go` | `tuiFrontend.Run` 读 `opts.Mode`，分发 connect 装配 |
| 新增 `internal/ui/connect.go` | `RunConnect` 骨架（daemon 探测 + 装配入口占位） |
| `internal/ui/netease.go` | `NewNetease` 增加 connect 变体构造（TC-3 完成前为占位） |
| `internal/commands/netease.go` | `--mode` 校验文案扩展（tui 消费 mode） |

**改动内容（签名级）**：

```go
// frontend.go
func (tuiFrontend) Run(ctx context.Context, opts frontend.LaunchOptions) error {
    if opts.Mode == frontend.ModeConnect {
        return RunConnect(ctx) // D-TC-1：遥控壳，不建 engine
    }
    // 以下为现状装配，逐字节不动
    ...
}

// connect.go（骨架）
// RunConnect 以遥控壳形态运行 TUI：连接本地 headless daemon（DialSubscribe
// 探测），不建 engine、不跑 Startup（D-TC-1/B9）。装配细节由 TC-2/TC-3 落地。
func RunConnect(ctx context.Context) error {
    client, err := headless.DialSubscribe(remoteEventWireNames()) // core.Ev* 常量集合
    if err != nil {
        return fmt.Errorf("connect 模式需要 headless daemon 正在运行: %w", err)
    }
    defer client.Close()
    // TC-3 起：NewNeteaseRemote → 事件消费 → 主循环装配（镜像 tuiFrontend.Run）
    return errors.New("tui connect not wired yet") // TC-3 替换
}
```

- `internal/commands/netease.go`：`--mode` 校验保持（standalone|connect 枚举不变），文案补「tui 前端自 S6 起消费 connect」。
- **无 daemon 行为**：`DialSubscribe` 报「headless daemon is not running」→ `RunConnect` 返回错误 → 退出码非 0（对齐 webui-connect 冒烟 7）。

**依赖**：S5-3（`headless.DialSubscribe`，已落地）。

**验证**：
- `go build/vet ./...`；`go test ./internal/ui/... ./internal/commands/...`。
- 冒烟：`--frontend=tui`（standalone）逐字节不变；`--frontend=tui --mode=connect`（无 daemon）报错退出非 0；有 daemon 时进入骨架（TC-3 前显示占位 toast 或错误）。

**风险**：低。纯分发 + 探测；standalone 路径由现有测试守护。

**写权限**：fixer。

---

### TC-2｜`RemotePlayer` 数据面：订阅消费 + 状态缓存 + 控制转发 + 渲染事件投递

**目标**：TUI-connect 的播放器数据面——复用 `headless.SubscribeClient`（D-TC-2），状态经快照 + 事件流缓存、控制经 `Call` 转发、渲染事件镜像 `ui.Player` 的 Observer 回调模式投递。

**涉及文件**：

| 文件 | 动作 |
|---|---|
| 新增 `internal/ui/remote_player.go` | `RemotePlayer` 类型（状态缓存 + 控制转发 + 事件消费） |
| 新增 `internal/ui/remote_player_test.go` | 缓存/事件映射/转发语义单测 |

**改动内容（签名级）**：

```go
// remote_player.go
package ui

// RemotePlayer 是 TUI-connect 的播放器数据面（D-TC-1 方案 B）：
// 状态 = SubscribeClient 快照 + 事件流增量缓存；控制 = Call 转发 daemon；
// 渲染事件经 renderTicker/Rerender 投递（与 ui.Player 的 Observer 回调
// 同构——core 播放 goroutine 调 Rerender 是既有线程安全前提，事件消费
// goroutine 沿用同一模式，B1）。
type RemotePlayer struct {
    client *headless.SubscribeClient
    netease *Netease
    renderTicker *tickerByPlayer

    mu sync.Mutex // 状态缓存锁
    // 以下字段从快照（Dispatcher status + 精简 playlist）与事件帧增量维护。
    ready    bool
    song     structs.Song // 快照 song 精简映射（Id/Name/Artist/Album）
    state    types.State
    passed   time.Duration
    volume   int
    mode     types.Mode
    playlist []structs.Song // 快照 playlist 精简映射
    user     *structs.User  // status.user 昵称（仅昵称，B8）
}

// NewRemotePlayer 构造遥控播放器并启动事件消费 goroutine。
func NewRemotePlayer(n *Netease, client *headless.SubscribeClient) *RemotePlayer

// --- playerRendererState 接口（B2）：renderer 组读取面 ---
func (p *RemotePlayer) CurSong() structs.Song
func (p *RemotePlayer) CurSongIndex() int
func (p *RemotePlayer) PassedTime() time.Duration
func (p *RemotePlayer) State() types.State
func (p *RemotePlayer) Volume() int
func (p *RemotePlayer) Mode() types.Mode
func (p *RemotePlayer) Playlist() []structs.Song

// --- 控制面（B3）：Call 转发；daemon 无对应命令的降级见 TC-3 ---
func (p *RemotePlayer) CtrlPlay/CtrlPause/CtrlResume/CtrlToggle()          // Call("play"|"pause"|"resume"|"toggle")
func (p *RemotePlayer) CtrlNext/CtrlPrevious/CtrlStop()                    // Call("next"|"prev"|"stop")
func (p *RemotePlayer) CtrlSeek(d time.Duration)                           // Call("seek", {seconds})
func (p *RemotePlayer) CtrlSetVolume(v int) / Volume()                     // Call("volume", {value}) / 缓存
func (p *RemotePlayer) CtrlSetRepeat(m int) / CtrlSetShuffle(on int)       // Call("repeat"|"shuffle")
func (p *RemotePlayer) CtrlLikeNowPlaying() / CtrlDislikeNowPlaying()      // Call("like"|"dislike")

// --- 查询面 ---
func (p *RemotePlayer) User() *structs.User                    // 快照 user 昵称（B8）
func (p *RemotePlayer) CommandContext() frontend.CommandContext // 快照映射（B10：MVP 命令面禁用，本方法保留供 P2）
func (p *RemotePlayer) RenderTicker() model.Ticker              // 复用 tickerByPlayer
func (p *RemotePlayer) Ready() bool

// consumeEvents 是事件消费 goroutine（TC-2 核心）：
// 逐帧读取 client.Events()：
//   - {"type":"snapshot"} → 全量刷新缓存（song/state/volume/mode/playlist/user）
//   - {"type":"event","event":"song_changed"|"state_changed"|"position"|"startup_phase"|"login"}
//     → 增量更新缓存（position 更新 passed + 喂 renderTicker；song/state 更新后 Rerender）
// 通道关闭（断线）→ ready=false → Rerender + 状态降级（D-TC-4）。
func (p *RemotePlayer) consumeEvents()
```

**事件名映射**：订阅 `core.Ev*` wire 名（`player.song_changed` 等），帧 `event` 字段即 wire 名（daemon 广播原样透传，无 webui 的帧名重映射层——TUI 直接消费 wire 名，比 webui 少一层映射）。

**实现要点**：
- 位置事件节流：镜像 `webui/events.go` 的 `positionThrottle`（250ms）——订阅流已是 daemon 限频后的，TUI 侧再按渲染 ticker 消化。
- 渲染投递：`song_changed`/`state_changed` → `netease.Rerender(false)`（镜像 `ui.Player.OnSongChanged/OnStateChanged`，B1）；`position` → 喂 `renderTicker.c`（非阻塞 select，镜像 `OnPosition`）。
- 控制转发失败（断线/超时）：`Call` 返回错误 → toast（`app.Notify`，线程安全）。

**依赖**：S5-3（`SubscribeClient`）。与 TC-1 并行编写。

**验证**：
- `go build/vet ./...`；`go test ./internal/ui/... -run RemotePlayer`。
- 单测：快照 → 缓存映射（song/state/volume/mode/playlist/user）；事件增量（song_changed 换歌、position 推进 passed + ticker）；`Call` 转发参数映射（seek 秒、repeat off/one/all）；断线（client.Close → ready=false + 通道关闭语义）。
- 集成测试（TC-4 合并或本票附带）：临时 socket daemon（`headless.NewServerWithAddr`）+ `DialSubscribe` + 事件断言。

**风险（TC 最高风险票）**：
- **渲染线程投递**：`Rerender` 从事件消费 goroutine 调用——与 core 播放 goroutine 调 `Rerender` 是同一既有前提（`ui.Player` 各 Observer 方法即此模式），风险低；但 `renderTicker` 的 channel 语义必须镜像 `OnPosition` 的非阻塞 select（防事件消费 goroutine 阻塞）。
- **遮蔽方法漏网**（TC-3 处理）：`RemotePlayer` 被 `ui.Player` 遮蔽引用，TUI 实际调用点集合必须 grep 锁定并以测试守护（见 TC-3 风险）。

**写权限**：fixer。

---

### TC-3｜装配 + 降级：`ui.Player` 遮蔽方法、renderer/访问器换源、登录门控禁用

**目标**：遥控壳完整装配。connect 模式不建 engine、不跑 Startup；`ui.Player` 在 connect 模式下遮蔽转发到 `RemotePlayer`（菜单代码零改动）；renderer 降级（歌词/封面/频谱隐藏）；`menuServices` 换源（Player/User）；登录门控与需登录菜单降级。

**涉及文件**：

| 文件 | 动作 |
|---|---|
| `internal/ui/connect.go` | `RunConnect` 完整装配（TC-1 骨架替换） |
| `internal/ui/netease.go` | `NewNeteaseRemote`（共享装配、跳过 engine/Startup）；`InitHook` connect 分支 |
| `internal/ui/player.go` | `Player` 加 `remote *RemotePlayer` 字段 + **显式遮蔽方法**（Go 遮蔽内嵌提升） |
| `internal/ui/menu_accessor.go` | `Player()`/`User()` connect 分支 |
| `internal/ui/netease.go`（Components） | renderer 组 connect 分支（Lyric/Spectrum 跳过，Cover 降级） |
| `internal/ui/frontend.go` | Ticker/StatusBar 装配 connect 分支 |

**改动内容（签名级）**：

```go
// connect.go（完整装配，镜像 tuiFrontend.Run 的装配序）
func RunConnect(ctx context.Context) error {
    client, err := headless.DialSubscribe(remoteEventWireNames())
    if err != nil { return ... } // TC-1 同文案

    opts := model.DefaultOptions()
    configs.AppConfig.FillToModelOpts(opts)
    model.Submit = types.SubmitText // 全局赋值保持（对齐 standalone 装配序）
    SetupI18n(configs.AppConfig.Main.Locale)

    netease := NewNeteaseRemote(model.NewApp(opts), client)
    // WithHook(InitHook=connect 装配+订阅, CloseHook)、WithMainMenu、Components、
    // Ticker=remote.RenderTicker()、KBControllers/MouseControllers=EventHandler ——
    // 与 standalone 同构，仅 Player/User/renderer 换源。
    return netease.Run()
}

// netease.go
// NewNeteaseRemote 是 connect 变体：不建 engine、不注册业务服务、不跑
// Startup（B9）；ctx 为 nil（menuServices 服务解析走 connect 分支）。
func NewNeteaseRemote(app *model.App, client *headless.SubscribeClient) *Netease

// player.go —— ui.Player 遮蔽方法（connect 分支，Go 显式方法遮蔽内嵌提升）：
// 遮蔽清单 = TUI 实际调用点集合（grep m.Player()/p.Player() 锁定，约 20 个：
// PlaySong/Playlist/CurSong/CurSongIndex/PassedTime/State/Volume/Mode/User/
// Ctrl*/ReinitializePlaylist/SetMode/CompareWithCurPlaylist/CommandContext/
// RenderTicker/PlayingInfo...）。connect 模式（p.remote != nil）转发
// RemotePlayer；standalone 走内嵌 *core.Player（现状零改动）。
func (p *Player) PlaySong(song structs.Song, dir PlayDirection) // connect：toast「遥控模式不支持」或 play<搜索>（D-TC-3）
func (p *Player) ReinitializePlaylist(index int, songs []structs.Song) // connect：no-op + toast（本地建列表不可达 daemon）
func (p *Player) CtrlSeek(d time.Duration)   // 转发 p.remote.CtrlSeek
func (p *Player) CtrlSetRepeat(m int)        // 转发
... // 其余遮蔽方法同模式

// menu_accessor.go
func (s *menuServices) Player() *Player {
    // connect：s.n.player.remote != nil → 返回带 remote 的 *Player（现状签名不变）
}
func (s *menuServices) User() *structs.User {
    // connect：s.n.player.remote.User()（快照昵称）；否则现状 engine 服务解析
}

// netease.go Components()（connect 分支）：
// LyricRenderer/SpectrumRenderer 不加入 Components；CoverRenderer 加入但
// 无数据源（内部降级为空渲染）；SongInfoRenderer/ProgressRenderer 保留
// （读 RemotePlayer 状态）。
```

**降级细则（D-TC-3，落 `connect.go` 头注释）**：
- **登录门控**：`RequestLogin`/`ToLoginPage` 在 connect 模式被菜单触发时 toast「遥控模式：登录由 daemon 管理」；需登录菜单（user_playlist/user_collect/云盘等）入口 toast 降级。
- **播放动作**：菜单选中歌曲 → `PlaySong` 遮蔽 → toast（MVP）或 `play <搜索词>`（近似，ticket 内二选一并注释）。
- **命令面**：connect 模式不加载 WASM（`loadWasmPlugins` 跳过）、`registerCommandMenus` 不执行（命令菜单不注册，B10）。

**依赖**：TC-1 + TC-2。

**验证**：
- `go build/vet ./...`；`go test ./internal/ui/...`。
- **遮蔽完整性测试**（关键守护）：断言 connect 模式下所有遮蔽方法**不落入内嵌占位 core.Player**（占位 core.Player 的方法若被意外调用即 panic/记录——用测试枚举遮蔽清单 + `remote != nil` 时占位方法 panic 的哨兵设计）。
- 冒烟清单（TC-4 合并执行）。

**风险**：
- **遮蔽清单漏网**：漏遮蔽的方法静默走占位 `core.Player`（nil/零值行为错）。对策：① grep 锁定调用点集合；② 占位 core.Player 用「connect 模式方法调用即 panic」的哨兵包装（`panic("tui-connect: unimplemented player method ...")`），测试与运行期双保险。
- **装配序回归**：`model.Submit`/`SetupI18n` 必须在 `NewNeteaseRemote` 之前（对齐 standalone 装配序注释，frontend.go:27-30）。
- renderer 降级分支漏改导致 nil 解引用（LyricRenderer 无服务）——Components connect 分支集中处理，单测覆盖空渲染不 panic。

**写权限**：fixer（遮蔽清单建议 reviewer 用 grep 复核）。

---

### TC-4｜测试 + 文档 + P2 扩展清单

> ✅ 已实施：新增 `internal/ui/connect_integration_test.go`（daemon 全链路集成测试：快照状态映射 / ctrl next 事件到达 + 触发重绘 / `Call("volume")` 在 daemon 生效 / 断线 `server.Close` → ready=false + 降级路径；测试另暴露并修复了 TC-2/TC-3 的两处渲染数据竞争——`n.App` 赋值晚于消费 goroutine 启动、消费 goroutine 早于 `app.Run()` 触碰 `App.program`）；文档同步（AGENTS.md / `docs/frontend_plugin.md` C12 / README.md）；P2 扩展清单落档 §8.9。

**目标**：S6 收尾——遥控壳端到端测试、文档同步、P2 扩展清单落档（含「选歌播放/歌词/封面」的 daemon 侧改动点）。

**涉及文件**：

| 文件 | 动作 |
|---|---|
| 新增 `internal/ui/connect_integration_test.go` | daemon（临时 socket）+ TUI-connect 全链路（状态/事件/控制转发） |
| `docs/frontend_plugin.md` | 新增 C12「TUI-connect」段 |
| `AGENTS.md` | TUI 前端段、headless 段、frontend 段补 connect 消费方 |
| `README.md` | `--frontend=tui --mode=connect` 用法 |
| `docs/roadmap_s_series.md`（本文件） | §8 标注已实施状态 + P2 清单 |

**改动内容**：
- 集成测试：`headless.NewServerWithAddr(engine, "unix", tmpSocket)` 起 daemon → `DialSubscribe` → 构造 `NewNeteaseRemote` 装配（不跑完整 TUI 主循环，用 `model.App` 消息循环驱动）→ 断言：快照状态映射；`ctrl next` 事件到达并触发重绘；`Call("volume")` 在 daemon 生效；断线（server.Close）→ ready=false + 降级路径。
- **P2 扩展清单（落文档，不进 MVP）**：
  1. **选歌播放**：daemon `Dispatcher` 新增 `play_song` 命令（wire 形状：id/name/artist/album 或歌曲 URL）→ TUI 菜单选中歌曲可精确投递（需 daemon 协议小扩展 + TUI `PlaySong` 遮蔽升级）。
  2. **歌词**：TUI 本地按 CurSong 拉取 + position 事件驱动推进（`lyric.Service` 脱离 engine 独立构造，需新适配器注入 renderer）。
  3. **封面**：daemon 快照 `status` 加 `picUrl` 字段（低成本）→ TUI CoverRenderer 恢复。
  4. **命令面**：daemon 命令面下沉（`CommandContext.UserID` 可得后）或 TUI 侧本地执行（UserID 降级）。
  5. **重连**：断线后自动重连（`DialSubscribe` 重试 + 状态重同步）。

**依赖**：TC-2 + TC-3。

**验证**：
```
冒烟清单（S6）：
1. musicfox --headless &（daemon 常驻）
2. musicfox --frontend=tui --mode=connect
3. TUI 启动显示 daemon 当前歌曲/状态/进度（快照）
4. 快捷键 next/prev/pause/seek/volume/repeat/shuffle → daemon 实际变化
5. musicfox ctrl next 切歌 → TUI 实时刷新（song_changed 事件）
6. 歌词/封面/频谱区域降级（隐藏或占位，不崩）
7. 需登录菜单（我的歌单）→ toast 降级；登录页不可达
8. 无 daemon 时 --mode=connect 报错退出（非 0）
9. daemon 断开 → TUI toast 提示 + 状态降级，TUI 不崩（Ctrl+C 正常退出）
10. --frontend=tui（standalone）全功能回归
```

**风险**：低。集成测试的 daemon 生命周期管理（临时 socket 清理，镜像 `headless/server_client_test.go` 基建）。

**写权限**：fixer。

### 8.6 风险与对策总表

| # | 风险 | 影响 | 对策 |
|---|------|------|------|
| 1 | **TUI 数据面改造面广**（renderer 5 组件 + 访问器 + Player 遮蔽） | 高 | 只改「数据源注入点」：renderer 读 `playerRendererState` 接口不动，换源集中在 `RemotePlayer` + `ui.Player` 遮蔽 + 访问器 connect 分支（D-TC-1 方案 B 的收敛设计） |
| 2 | **订阅事件 → TUI 渲染线程投递** | 中 | 与 core 播放 goroutine 调 `Rerender` 的既有模式同构（`ui.Player` Observer 方法先例）；`renderTicker` 沿用 `OnPosition` 的非阻塞 select（B1） |
| 3 | **遮蔽方法漏网**（静默走占位 core.Player） | 高 | grep 锁定调用点集合 + 占位 core.Player 的「connect 模式调用即 panic」哨兵 + 测试守护（TC-3 风险） |
| 4 | **断线语义**（daemon 重启 → TUI 状态静止） | 中 | 事件通道关闭 → ready=false + toast + 降级（D-TC-4）；不自动重连（MVP），P2 扩展 |
| 5 | **登录态错位**（daemon 已登录 vs TUI 显示） | 中 | 仅显示昵称（status.user）；登录门控禁用 + 需登录菜单降级；不实现 TUI 侧登录（对齐 webui-connect 503 哲学） |
| 6 | **与 webui-connect 功能一致性** | 中 | 共享 `SubscribeClient` 数据面与边界哲学（D-S5-2/D-TC-3）；集成测试共用 daemon 基建，行为差异以文档边界表锁定 |
| 7 | **播放列表撕裂**（本地浏览歌单 ≠ daemon 队列） | 中 | 操作语义明确：浏览/搜索本地；播放动作降级（toast / play<搜索>）；播放队列 UI 只读精简（8.4 边界表） |

### 8.7 与 S5 / GUI 的关系与优先级

- **是否依赖 S5 的 Backend 抽象（裁决 D-TC-2）**：**不依赖、不复用** `webui.Backend`/`remoteBackend`。理由：① `webui.Backend` 是 webui 包内为 HTTP Server 设计的查询接口（`Playlist` 返回 `[]map[string]any`、`LyricState` 是 wire 形状），TUI 需要 `structs.Song` 级状态缓存，形状不匹配；② TUI（`internal/ui`）import webui 语义混乱（兄弟前端包反向耦合），且会给 S5 已定稿的包边界引入新依赖方向；③ **TUI 直接消费 `headless.SubscribeClient`**（与 webui 同源，D-S5-1 的依赖方向先例：ui → headless 无环、headless UI-free）。
- **依赖 S5 的哪些产物**：`headless.SubscribeClient`（S5-3）、daemon subscribe 协议（P7）、`core.Dispatcher` 命令集——均为已落地资产，**零新协议**。
- **相对 GUI 的优先级（裁决 D-TC-5）**：**TUI-connect 先于 GUI**。理由：① 用户需求直接驱动（用户主动询问）；② 依赖已全部就绪（S5 落地），工作量小（6–10 人日 vs GUI 15–25 人日）；③ GUI 无前置依赖 TUI-connect，但 TUI-connect 的「遥控壳」模式（数据面复用、降级哲学）为 GUI 的 connect 变体提供先例；④ 交付节奏上，S6 可与 GUI-1（spike）并行。

### 8.8 文档同步义务

| 阶段 | 必须更新的文档 | 内容 |
|------|--------------|------|
| S6 收尾 | `AGENTS.md`；`docs/frontend_plugin.md`（C12）；`README.md`；`docs/roadmap_s_series.md`（§8 状态标记） | TUI 前端段补 connect 遥控壳；headless 段补 SubscribeClient 的 TUI 消费方；`--frontend=tui --mode=connect` 用法；P2 扩展清单 |

### 8.9 P2 扩展清单（MVP 明确降级的扩展项，落档）

以下 5 项为 S6 MVP 明确降级的扩展（D-TC-3 / D-TC-4 / D-TC-6），MVP 不实施。每项含 daemon 侧与 TUI 侧改动点，供后续独立立项。**第 1 项（选歌播放）已由 §8.10 TC-5..TC-8 提前落地**（落地形式与下文改动点不同——改为 `play_list` 整列表投递而非 `play_song` 单曲命令，D-TC-9）；其余 4 项仍为 P2 落档。

| # | 扩展 | daemon 侧改动点 | TUI 侧改动点 |
|---|------|----------------|--------------|
| 1 | **选歌播放** | ✅ **已落地（§8.10 TC-5..TC-8）**：`Dispatcher` 新增 `play_list` 命令（wire `{songs:[{id,name,artist,album}...], index, play}`，`cmdPlayList` 重建 daemon 队列 + 可选 StartPlay，护栏 5000）；单曲点播为 index=0 特例，**不单独新增 `play_song`**（D-TC-9） | ✅ **已落地（TC-7）**：`ui.Player.PlaySong`/`ReinitializePlaylist`/`StartPlay` 遮蔽升级为 `play_list` 投递（`CallPlayList`，响应 playlist 写回缓存同步队列） |
| 2 | **歌词** | 无协议改动（position 事件已在订阅面） | TUI 本地按 `CurSong` 拉取歌词 + position 事件驱动推进（`lyric.Service` 脱离 engine 独立构造，需新适配器注入 renderer；当前 `LyricRenderer` 被跳过） |
| 3 | **封面** | daemon 快照 `status` 加 `picUrl` 字段（低成本单字段扩展） | TUI `CoverRenderer` 恢复（`connectCoverState` 的剥 PicUrl 降级可移除；当前封面空渲染） |
| 4 | **命令面（轨 B / WASM）** | daemon 命令面下沉（`CommandContext.UserID` 可得后）或 TUI 本地执行（UserID 降级） | 恢复 WASM 加载与 `registerCommandMenus`（当前 `CommandContext` 已保留供 P2 复用）；对齐 webui-connect「空命令面」哲学 |
| 5 | **重连** | 无 | 断线后自动重连（`DialSubscribe` 重试 + 状态重同步），替换 D-TC-4 的「事件通道关闭 → `ready=false` + 不自动重连」 |

裁决背景：MVP 收敛优先，daemon 协议保持「精简快照 + 现 Dispatcher 命令集」零扩展（对齐 webui-connect 的 D-S5-2 哲学）；1/3 为低成本 daemon 改动，2/4 为 TUI 侧适配，5 为可靠性增强。（第 1 项已由 §8.10 TC-5..TC-8 落地。）

### 8.10 完整能力扩展：登录 + 选歌播放（D-TC-7..9，用户诉求驱动）

> 用户反馈：connect 模式提示「登录由 daemon 管理」且不能选歌播放，要求完整能力。本小节将 §8.9 扩展清单第 1 项（选歌播放）与新增的登录项提前落地为 **TC-5..TC-8**，推翻 D-TC-3 中「登录/选歌播放 toast 降级」的相关表述（落地后以代码为准）。

> ✅ **TC-5..TC-8 已实施**（分支 `feat/plugin-framework-playback`）：daemon 命令面扩展（`login_qr_key`/`login_qr_status`/`play_list` + `status.userId`，`internal/core/dispatcher.go`）；TUI 登录遥控（`ToLoginPage` connect 分支 → 只显 QR 入口的 LoginPage → `RemotePlayer.CallQRKey`/`CallQRStatus` 数据源，`EvLogin` 驱动用户态刷新）；选歌播放（`PlaySong`/`ReinitializePlaylist`/`StartPlay` 遮蔽升级为 `play_list` 整列表投递，响应 playlist 写回缓存同步队列）。边界表见 §8.4（完整版）与 §8.10.2，行为由 `go test ./internal/core/... ./internal/ui/...` 锁定。

#### 8.10.1 核心裁决

| 编号 | 裁决 | 理由 |
|------|------|------|
| **D-TC-7** | **登录态归 daemon（唯一宿主），TUI 遥控扫码**：daemon 扩展 `login_qr_key`/`login_qr_status` 命令（复用 `core/qrlogin` + `engine.CompleteQRLogin`），TUI 在 connect 登录页显示二维码、轮询状态，`EvLogin` 事件经订阅驱动 TUI 刷新 | ① 用户诉求「在 TUI 里登录」=「在 TUI 界面发起登录」，非「登录发生在 TUI 进程」；② daemon 是登录态正确宿主（LoginService.InitJar → SetGlobalCookieJar → startup 恢复）；③ 播放链路 URL 解析需 cookie，daemon 侧持有结构性消除跨进程 cookie 同步；④ 与 webui standalone「进程内 QR 登录」同构。否决方案 B（TUI 本地登录 + cookie 同步 daemon）：双宿主不一致 + wire 传输凭据安全评审 + 覆盖 daemon 已有登录态语义混乱 |
| **D-TC-7a** | **cookie 回拉 = P2**（`login_qr_status` 803 响应带回 cookie 串，TUI 恢复本地需登录浏览） | 需登录浏览（收藏/我的歌单/云盘/每日推荐/私人FM）在 connect 下**无条件降级**（TUI 本地进程无 cookie，跨进程登录态墙）；cookie 副本引入双宿主不一致。用户核心诉求（登录 + 播放）不依赖它 |
| **D-TC-8** | **用户态同步 = 快照幂等 + EvLogin 事件增量双写**：daemon `status` 加 `userId` 字段（快照幂等，重连即恢复）；`EvLogin` data 已含 userId（事件增量，会话中即时刷新） | 两路径覆盖正交场景，成本均低（daemon 3 行 + TUI 5 行）。`User()` 访问器继续返回剥离 UserId 的副本（门控语义保持）；新增 `UserID()`/`UserLoggedIn()` 供展示与 P2 命令面 |
| **D-TC-9** | **播放 = `play_list` 整列表投递**（单曲点播为其特例，不单独新增 `play_song`）：daemon 新增 `play_list` 命令（wire `{songs:[{id,name,artist,album}...], index, play}`），`cmdPlayList` 按 wire songs 重建 daemon 队列 + 可选 `StartPlay`，与 `cmdPlay` 搜索先例同构 | ① TUI 菜单选中语义就是「整列表 + 定位 index」（`playOrToggle`），`play_list` 与 `ReinitializePlaylist(index,songs)`+`StartPlay` 逐参数对齐，遮蔽升级最小；② 单曲点播造成 next/prev 队列语义撕裂；③ daemon 端实现与 cmdPlay 同构，复杂度不增。护栏 `playListWireLimit=5000`。队列缓存经命令响应同步（`EvPlaylistChanged` 事件 = P2） |

#### 8.10.2 功能边界表（§8.4 完整版）

| 能力 | TUI standalone | TUI-connect（MVP） | TUI-connect（本扩展后） |
|------|---------------|----------------------|------------------------|
| 播放控制 / 状态 / 进度 | 本地 engine | Call 转发 / 订阅快照 ✅ | ✅（不变） |
| 搜索 / 排行 / 精选歌单 / 专辑 / 歌手 / DJ 浏览 | 本地 API | 本地照常 ✅ | ✅（不变） |
| **登录** | 本地（账号/QR/WebView） | toast「登录由 daemon 管理」 | **daemon 侧 QR 登录，TUI 遥控扫码** ✅（ToLoginPage → connect LoginPage 只显 QR 入口 → Call 数据源；EvLogin 驱动刷新） |
| **选歌播放** | 本地建列表 + 播放 | toast 降级 | **`play_list` 整列表投递 daemon** ✅（PlaySong/ReinitializePlaylist/StartPlay 遮蔽升级） |
| 播放队列显示 | 本地完整列表 | 快照精简只读 △ | 精简只读 △ + 投递后响应同步（next/prev 正确） |
| 播放队列编辑 | 本地 | toast | 仍降级（P2：daemon 队列编辑命令） |
| 需登录浏览（收藏/我的歌单/云盘/每日推荐/最近播放/私人FM） | 本地 API + 登录 | toast | **仍降级（无条件）**——TUI 本地进程无 cookie（D-TC-7 边界）；P2：cookie 回拉 |
| 用户态展示 | 本地 user | 昵称（无 UserId） | 昵称 + UserId（快照幂等 + EvLogin 增量）；门控仍剥离 UserId |
| 歌词 / 封面 / 频谱 / 智能模式 / 命令面 / 重连 | 本地 | 隐藏/禁用 | 仍 P2（歌词本地拉取、封面 daemon 快照加 PicUrl、频谱结构性不可用、命令面 UserId 可得后需 WASM 加载裁决、重连） |

#### 8.10.3 Ticket 拆分

```
TC-5（daemon 命令面扩展：login_qr_key/login_qr_status/play_list + status 加 userId）─ 无前置（公共底座）
TC-6（TUI 登录遥控：ToLoginPage connect 分支 + QR 页数据源换 Call + 用户态展示）────┐ 依赖 TC-5（可与 TC-7 并行）
TC-7（TUI 播放转发：ui.Player 遮蔽升级 + CallPlayList + 队列缓存同步）──────────────┘ 依赖 TC-5（可与 TC-6 并行）
TC-8（测试 + 文档 + 边界表更新）──────────────────────────────────────────────────── 依赖 TC-6 + TC-7
```

- **TC-5**（`internal/core/dispatcher.go` + 测试）：`login_qr_key`（qrlogin.GetKey → {uniKey,qrcodeUrl}）、`login_qr_status {key}`（CheckStatus → code；803 时 engine.CompleteQRLogin(jar) → {code,user}；包级 overridable 镜像 webui/api_login.go）、`play_list {songs,index,play}`（ReinitializePlaylist + 可选 StartPlay → {playlist,index,started}；护栏 5000）、`cmdStatus` 加 `userId`。`ProtocolVersion` 不 bump（命令名空间扩展）。
- **TC-6**（`internal/ui/netease.go`/`login_page.go`/`login_qr_page.go`/`remote_player.go`）：`ToLoginPage` connect 分支 toast → 返回 LoginPage（只显 QR 入口，隐藏 tabs/webview——`n.ConnectMode()` 守卫防 nil engine panic）；`QRLoginPage` 加 `remote` 字段（generateQRCodeCmd/pollQRStatusCmd/loginSuccessHandle 三处 connect 分支）；`RemotePlayer` 加 `callResult`/`CallQRKey`/`CallQRStatus`/`UserID()`/`UserLoggedIn()`。
- **TC-7**（`internal/ui/player.go`/`remote_player.go`）：`PlaySong`→`CallPlayList([]Song{song},0,true)`、`ReinitializePlaylist`→`CallPlayList(s,i,false)`、`StartPlay`→`CtrlResume()`（daemon resumeOrStart）；`CallPlayList` 响应 playlist 更新缓存。`event_handler.go` 零改动（playOrToggle 三步遮蔽已在 Player 层）。
- **TC-8**（✅ 已实施）：集成测试评估——登录 stub 流程由 TC-5（core `dispatcher_test.go` 803 成功路径 + `login_qr_key`/`login_qr_status` 参数/错误路径）与 TC-6（`connect_integration_test.go`：CallQRStatus 空 key 校验、EvLogin→UserID、快照 userId）覆盖；play_list 全链路由 TC-7（play=false 建队列 + 响应同步、遮蔽转发、空 song 不 panic）覆盖，play=true 真实播放由 core `TestDispatcherPlayList`（fake track provider）覆盖（集成侧已有注释确认不重复）；边界回归（断线降级 / 无 daemon fail-fast）由既有测试覆盖（`TestConnectIntegrationDisconnect`、headless `TestDialSubscribeNoDaemon`）。TC-8 新增 `TestQRLoginPageConnectLoginSuccess`（QR 页 connect 分支：loginSuccessHandle 只跑 AfterLogin、不调本地 CompleteQRLogin）。文档（AGENTS.md/docs/roadmap §8.4/README/frontend_plugin C12）。

**并行策略**：TC-5 先行（公共底座）→ TC-6 / TC-7 并行 → TC-8 收尾。播放（TC-7）不依赖登录完成（play_list 对游客同样可用）。

**风险要点**：`login_qr_status` 803 路径阻塞 Dispatcher mutex ≤ 数秒（低频可接受）；LoginPage connect 渲染分支漏禁本地登录路径会 nil 解引用（高，`ConnectMode()` 守卫 + 单测）；`StartPlay`→`CtrlResume` 依赖 daemon `resumeOrStart`（已核实 Stopped 恢复）；`play_list` 大数组（护栏 + 精简字段）。

---

*本文档所有代码块为签名级规格（供实施参考），非最终实现；个别签名（如 `Backend.SubscribeEvents` 的精确返回形状、`SubscribeClient` 的断线语义）在实施时可微调，但**契约形状与裁决编号（D-S1-1 … D-GUI-3）不可静默变更**——如需推翻，须在对应 ticket 注释理由。*
