# go-musicfox 项目架构文档

> 本文档描述**系统架构、模块职责与开发约定**。变更历史、bug 修复过程、临时规避方案与版本演进细节不写入本文档，判定标准见「文档记录边界」。

## 沟通语言准则

### 重要准则：与用户交流必须使用中文

**所有贡献者、开发者以及 AI 助手在与用户交流时，必须使用中文。**

#### 适用场景
- 代码审查意见和反馈
- Issue 回复和讨论
- Pull Request 描述和评论
- 文档编写（README、CHANGELOG 等除外，代码注释和提交信息仍使用英文）
- AI 助手与用户的对话

#### 例外情况
- **代码注释**：使用英文（保持代码可读性和国际化标准）
- **Git Commit Message**：使用英文（遵循 Conventional Commits 规范）
- **代码变量/函数命名**：使用英文（遵循 Go 语言惯例）
- **API 文档和错误消息**：根据实际情况使用中文或英文

#### 示例

```markdown
✓ 正确示例（中文交流）：
"这个 PR 优化了歌词封面图圆角渲染，CPU 占用从 10% 降至 2%。"

✗ 错误示例（英文交流）：
"This PR optimizes the lyric cover cornerRadius rendering..."
```

## 编码行为准则

本准则旨在减少 AI 编码常见的错误。**权衡原则**：这些准则偏向谨慎而非速度，对于简单任务可自行判断。

> 本准则改编自 [andrej-karpathy-skills/CLAUDE.md](https://github.com/forrestchang/andrej-karpathy-skills/blob/main/CLAUDE.md)

### 1. 三思而后行

**不要假设。不要隐藏困惑。明确权衡。**

在实现之前：
- 明确陈述你的假设。如果不确定，先问。
- 如果存在多种解释方案，全部提出——不要默默选择一个。
- 如果存在更简单的方案，指出它。有理由时可以反驳。
- 如果有不清楚的地方，停下来。说出哪里不清楚，然后问。

### 2. 简单性优先

**最小代码解决问题。不要 speculative。**

- 不添加需求之外的功能。
- 一次性使用的代码不抽象。
- 不添加未被请求的"灵活性"或"可配置性"。
- 不处理不可能发生的错误场景。
- 如果写了 200 行而可以用 50 行完成，重写。

问自己："高级工程师会觉得这过于复杂吗？"如果是，简化。

### 3. 精准修改

**只触碰必须改的。只清理自己造成的混乱。**

编辑现有代码时：
- 不"改进"相邻代码、注释或格式。
- 不重构没坏的东西。
- 匹配现有风格，即使你可能用不同方式实现。
- 如果发现无关的死代码，指出它——不要删除它。

当你的修改产生孤儿（不再使用的代码）时：
- 移除因你的修改而不再使用的 import/变量/函数。
- 不要移除预先存在的死代码，除非被要求。

**检验标准**：每一行修改都应该能直接追溯到用户的请求。

### 4. 目标驱动执行

**定义成功标准。循环验证直到完成。**

将任务转化为可验证的目标：

| 模糊表述 | 可验证目标 |
|---|---|
| "添加验证" | "为无效输入编写测试，然后让测试通过" |
| "修复 bug" | "编写能复现问题的测试，然后让测试通过" |
| "重构 X" | "确保重构前后测试都通过" |

对于多步骤任务，简要陈述计划：

```
1. [步骤] → 验证: [检查点]
2. [步骤] → 验证: [检查点]
3. [步骤] → 验证: [检查点]
```

明确的目标让你能独立循环验证。模糊的目标（"让它能工作"）需要不断确认。

---

**这些准则生效的标志**：diff 中不必要的修改更少，因过度复杂而返工更少，澄清问题在错误之前而非错误之后提出。

## 项目概述

go-musicfox 是基于 Go 和 bubbletea 的网易云音乐 TUI 客户端，支持 macOS/Linux/Windows。

**技术栈**：
- **UI 框架**：bubbletea + foxful-cli（部分定制）
- **音频处理**：beep、go-mp3、go-flac
- **存储**：BoltDB
- **配置**：TOML + mapstructure；`configs.SetTOMLValue(path, keyPath, value)` 按键路径保真编辑既有 TOML，`configs.UpgradeConfig(path)` 将内嵌默认 TOML 中缺失的叶子项补入对应 table（`upgrade-config` 子命令调用），两者均保留注释、布局、已有值、未知键与原文件权限
- **API**：netease-music SDK

**项目结构**：`cmd/` 入口 | `internal/` 核心业务（22 个包） | `utils/` 工具 | `configs/` 嵌入式配置

## 核心架构

### 应用入口与初始化

**入口**：`cmd/musicfox.go` → `runtime.Run()` → 加载配置 → 数据迁移 → 启动 TUI

### UI 协调器

**文件**：`internal/ui/netease.go`

核心结构包含：login、search、player、lyricService、trackManager 等组件。

### 核心接口

**Menu 接口**（`internal/ui/menu.go`）：
```go
type Menu interface {
    model.Menu
    IsPlayable() bool
    IsLocatable() bool
}
type SongsMenu interface { Songs() []structs.Song }
type PlaylistsMenu interface { Playlists() []structs.Playlist }
```

**Player 接口**（`internal/player/player.go`）：
```go
type Player interface {
    Play(music URLMusic)
    Pause()/Resume()/Stop()/Toggle()
    Seek(duration time.Duration)
    PassedTime()/PlayedTime() time.Duration
    Volume()/SetVolume()/UpVolume()/DownVolume()
    State() types.State
    Close()
}
```

### 播放引擎

| 引擎 | 平台 | 特点 |
|------|------|------|
| Beep（默认） | 跨平台 | MP3/FLAC/OGG/WAV；可选 MP3 无缝播放 |
| DLNA | 跨平台 | 设备投送 |
| MPV | 跨平台 | IPC 控制；自然结束时暂停并保留播放计时供切歌上报 |
| MPD | Linux | 远程服务器 |
| AVFoundation | macOS | 原生集成 |
| MediaPlayer | Windows | WinRT API |

### 事件处理

**文件**：`internal/ui/event_handler.go`

支持 40+ 键盘操作、鼠标事件（单击/双击/滚轮/右键）、可配置快捷键。

### 功能模块

- **歌词**：LRC/YRC 格式，支持 smooth/wave/glow 渲染模式。`lyric.Service` 异步加载歌词并支持取消（新请求作废旧请求，`Stop` 清理状态）；`track.Manager.GetLyric` 用 `singleflight` 合并并发请求。未匹配的云盘歌曲优先获取云盘内嵌歌词，失败回退普通歌词接口。
- **播放列表**：列表循环/顺序/单曲循环/随机/无限随机/智能心动模式。随机序由 `internal/playlist` 维护，队列不变时前后切歌在随机序内移动。
- **远程控制**：MPRIS(Linux)、Now Playing(macOS)、System Media(Windows)。macOS 侧 `internal/macdriver` 在框架不可用时静默降级，禁用对应功能而不影响启动。
- **状态栏组件**：`model.DefaultStatusBar.Components` 可注入 `StatusBarComponent`；需要交互的组件额外实现 `InteractiveStatusBarComponent`，以局部坐标处理命中。状态栏只负责布局与指针，不持有业务回调。
- **存储**：BoltDB，保存用户信息、播放状态、播放列表快照、桌面歌词窗口位置/显示器。
- **扫码登录**：`internal/ui/login_qr_page.go` 经 `internal/ui/qr_login_client.go` 直接请求网易云接口（非 `service.LoginQRService`），以 `req/v3` 的 `SetTLSFingerprintChrome()` 与完整浏览器 Headers 规避反爬检测；轮询每次绑定最新全局 CookieJar（`neteaseutil.GetGlobalCookieJar()`）。
- **网页登录**：页面层 `internal/ui/login_webview_page.go` 平台无关，三平台实现均零 CGO——macOS WKWebView（`internal/macdriver/webkit`）、Windows WebView2（`internal/ui/login_webview_windows.go`，依赖 `13thgoutham/go-webview2` fork）、Linux WebKitGTK（`internal/webkitgtk`，优先 6.0 栈、回退 4.1）。通过原生 cookie API 轮询 HttpOnly 的 `MUSIC_U`；WebView 运行时不可用时「网页登录」入口自动隐藏。
- **实时频谱**：仅 macOS `osx` 引擎支持；`MTAudioProcessingTap` 经 PureGo 取 PCM，`internal/player/spectrum.go` 异步分析，`harmonica` 弹簧按帧插值。配置项为 `[main].frameRate` 与 `[main.visualizer]`（`enable`、`maxHeight`、三种单元字符）。
- **主题**：`[dark.app]` / `[light.app]` 的显式非透明 `background` 作为应用与自定义页面整窗背景；未设置或 `transparent` 时保持终端透明。语义色、状态栏、播放栏、菜单圆角分隔符（`menuSelectedSepLeft`/`menuSelectedSepRight`）与进度条渐变均可覆写，缺省回落到主题调色板。切换主题时以名称写入 `theme.activeTheme`，系统外观变化只切换当前主题的明暗变体。
- **自定义表单页面**：`internal/ui/page_layout.go` 统一提供标题返回按钮、面包屑、输入框焦点/hover 样式与按钮文本；Search、Login、Last.fm API account 页面按其渲染结果记录命中坐标，并用 `pageSubmitButton`/`pageButton` 同步键盘与鼠标状态。
- **封面图**：Kitty 协议图像渲染在应用背景预留的透明区域，位于文本与单元格背景之下。tmux 内通过 DCS 透传（需 tmux ≥3.3 且 `allow-passthrough on`），启用还要求 `main.lyric.cover.tmuxPassthrough=true`（默认关闭）；开启后强制静态封面、缩图并以 768 KiB 单包 / 1 MiB/s 预算限流，慢写触发退避冷却。herdr pane 使用裸 Kitty 定位序列与静态封面。独立页面清理封面须在 deferred `MenuToPage.BeforeEnterMenuHook` 中调用 `coverRenderer.ClearDisplayed()`。

## 开发指南

### 添加新菜单

1. 创建 `internal/ui/menu_new_feature.go`，嵌入 `baseMenu`
2. 实现 `Menu` 接口
3. 注册到导航系统

### 添加新播放器引擎

1. 实现 `internal/player.Player` 接口
2. 在 `player.go:NewPlayerFromConfig()` 添加 case
3. 添加配置支持

### 修改快捷键

1. 在 `internal/keybindings/keybindings.go` 定义 `OperateType`
2. 在 `event_handler.go` 添加键映射
3. 在配置文件中添加自定义绑定

### 添加新渲染器

1. 实现 `Update()` 和 `View()` 方法
2. 在 `netease.go:Components()` 注册

### 更新依赖

使用 `make vendor` 更新 vendor 目录，该命令会依次执行 `go mod tidy`、`go mod vendor` 并复制 CGo 依赖的头文件。

```bash
make vendor
```

#### foxful-cli 依赖更新流程

foxful-cli 是本项目的核心 UI 框架，位于 `vendor/github.com/anhoder/foxful-cli/`。更新流程：

1. 在 `go.mod` 中使用 `replace` 指向本地 foxful-cli 仓库开发
2. foxful-cli 改动完成后，在本地仓库 commit 并推送
3. 在 foxful-cli 仓库打 tag（如 `v1.0.5`）
4. 在本项目 `go.mod` 中注释掉本地 `replace`，将版本升为新 tag
5. 执行 `make vendor` 同步 vendor 目录

### 跨平台构建兼容性

**修改 `Makefile` 或 `hack/` 目录下的构建脚本时，必须确保 Windows 系统兼容。**

#### 原因
go-musicfox 支持 macOS/Linux/Windows 三平台，构建脚本需要在不同环境下正确执行。

#### 检查清单
- [ ] 新增的 target 是否包含 Unix 特有命令（`which`、`cp`、`mkdir -p`、`chmod`、`tar`、`awk` 等）
- [ ] 是否使用 `$(OS)` / `Windows_NT` 条件分支提供 Windows 替代逻辑
- [ ] Shell 脚本（`.sh`）是否有对应的 PowerShell（`.ps1`）实现
- [ ] 路径分隔符是否兼容（使用 `$(PACKAGE_ROOT)` 而非 `` `pwd` ``）
- [ ] 重定向和设备文件是否有 Windows 替代（`nul` 对应 `/dev/null`，`where` 对应 `which`）

#### Windows 兼容模式参考

```makefile
# 使用 OS 条件分支
ifeq ($(OS),Windows_NT)
    # Windows 逻辑
else
    # Unix 逻辑
endif
```

| Unix 写法 | Windows 替代 |
|---|---|
| `which <cmd>` | `where <cmd>` |
| `/dev/null` | `nul` |
| `` `pwd` `` | `$(PACKAGE_ROOT)`（Make 变量） |
| `<cmd> >/dev/null 2>&1` | `<cmd> >nul 2>&1` |
| `<cmd>\|\|{ cmd2; }` | `<cmd>\|\|( cmd2 )` |
| `hack/*.sh` | `hack/*.ps1`（PowerShell 实现） |

**例外**：纯交叉编译/CI 内部工具（如 `hack/init_linux_env.sh`、`hack/init_windows_env.sh` 等 Docker 内部脚本），仅在 Linux Docker 容器中执行，无需适配 Windows。

## 文档维护准则

### 重要准则：修改代码后需维护 AGENTS.md

**所有贡献者在修改代码后，必须检查并更新 AGENTS.md 文档，防止文档腐化。**

#### 何时需要更新文档

| 变更类型 | 动作 |
|---|---|
| 架构变更 | 必须立即更新 |
| 新增/删除核心模块或组件 | 必须补充或移除对应说明 |
| 新增播放器引擎、菜单类型、渲染模式 | 必须更新核心架构章节与开发指南 |
| 公共接口或配置格式变化 | 必须更新接口定义与配置语义 |
| 核心文件或目录重命名、迁移 | 及时更新路径与项目结构 |
| 快捷键变更 | 必须同步 README 的快捷键说明（help 菜单由 keybindings 动态生成，无需手动修改） |
| 细微修复、参数调优、bugfix | **不更新**（见「文档记录边界」） |

### 重要准则：只记录架构与功能，不记录修复细节（文档记录边界）

**AGENTS.md 是架构说明书，不是变更日志。** 任何写入的内容都必须描述系统"当前长什么样、为什么这样组织"，而不是"某次改动为什么被做"。

#### 应当记录
- 模块划分、职责边界与依赖方向
- 公共接口、扩展点与注册方式（新增菜单/引擎/渲染器的步骤）
- 配置项的语义与默认行为
- 平台差异与降级策略（如某功能仅在特定平台可用）
- 关键数据流（入口 → 初始化 → 运行 → 渲染）
- 构建、依赖与发布流程

#### 不应记录
- **Bug 修复的因果叙事**：症状、根因、修复手法
- **Issue / PR 编号引用**及其讨论上下文
- **小改动与调优**：参数微调、格式化、日志、单点性能数字
- **临时规避手段**：workaround、watchdog、冷却、重试次数的实现细节
- **测试细节**：具体测试文件名、测试替身策略、回归用例清单
- **变更历史**：与旧实现的对比、"此前错误/现在修正"的表述
- **计划与限制清单**："后续将…"、"已知限制…"、"不代表…已根治"
- **Agent 工作日志**：任务过程、调试结论、临时假设、已尝试但放弃的方案
- **文件级实现内幕**：私有函数名、锁粒度、goroutine 编排等需要读代码才能理解的细节

以上内容分别归属：变更历史 → `CHANGELOG.md` / Git log；用户可见行为 → `README.md`；特定子系统的深度设计 → 独立 `docs/*.md` 或代码注释。

#### 写法要求
- 使用现在时的陈述句描述现状，例如"tmux 内通过 DCS 透传…"，不写"本次修复了…"
- 每个功能模块最多 3–5 条要点，单条不超过 3 行
- 优先写配置键、接口名、包路径等稳定锚点，避免行号与临时符号
- 同一事实只出现一次；细节已存在于 README/`docs/` 时只保留链接或一句摘要
- 新增条目必须能回答"读者是否会因此改变对系统的理解"；答案为否即删除

#### 判断标准
提交前对每一条新增内容自问：

1. 它是**现状描述**还是**变更说明**？后者不写入。
2. 删掉它，新维护者是否仍能定位代码并正确扩展系统？能则删掉。
3. 它是否只对某次修复有意义、对后续开发无指导价值？是则删除。
4. 它是否依赖具体 issue、PR、commit 才能理解？是则删除。

#### 检查清单
- [ ] 目录结构与核心文件路径与实际代码一致
- [ ] 接口定义与代码一致，无已删除 API
- [ ] 新增功能只以"能力 + 入口 + 配置"的形式描述
- [ ] 无 issue/PR 编号、无 bugfix 过程叙述、无测试文件清单
- [ ] 无重复描述；与 README/CHANGELOG 无矛盾

### 文档腐化警告信号
- 文件路径与实际不符
- 过时的 API 接口描述
- 已删除功能仍出现在文档中
- 章节被按时间顺序堆叠的改动说明填充
- 出现"本次/此前/修复了/临时"等变更叙事词汇
- 与 README/CHANGELOG 存在矛盾

违反此准则可能导致：新贡献者无法理解项目结构、文档失去参考价值、维护成本持续增加。

## Git 提交规范

### 重要准则：Git Commit Message 必须遵循 Conventional Commits 规范

**所有贡献者在提交代码时，必须遵循 [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) 规范。**

**Git Commit Message 必须全部使用英文**：标题、body、footer（含 `BREAKING CHANGE` 描述）不得出现中文或其他非英文自然语言。中文仅用于 PR 描述、Issue 讨论与代码审查意见。

#### 提交格式

```
<type>(<scope>)!: <description>

<body>

<footer>
```

- 标题行：`type` 必填，`scope` 可选，`!` 标记破坏性变更，冒号后接一个空格再接描述
- 标题与 body、body 与 footer 之间各空一行
- 只写标题的提交同样合法；无 body 时不要在标题后追加句点或列表符号

#### Type 类型

| Type | 说明 |
|---|---|
| `feat` | 新功能开发 |
| `fix` | Bug 修复 |
| `docs` | 文档更新 |
| `style` | 代码格式调整（不影响功能） |
| `refactor` | 代码重构（既不修复 bug 也不添加功能） |
| `perf` | 性能优化 |
| `test` | 测试相关 |
| `chore` | 构建工具、辅助工具、配置变更 |
| `revert` | 回滚提交 |

`type` 必须与实际变更一致：行为变化不得用 `chore`/`docs` 隐藏，纯文档改动不得用 `feat`。

#### Scope 约定

可选、小写、单个模块。跨模块改动可省略 scope，或取影响最大的模块，不写 `ui,player` 这类组合。

| Scope | 覆盖范围 |
|---|---|
| `ui` | `internal/ui` 页面、菜单、事件与渲染 |
| `player` | `internal/player` 播放引擎与控制 |
| `lyric` | 歌词获取、解析与渲染 |
| `playlist` | 播放列表模式与随机序 |
| `track` | `internal/track` 歌曲、缓存与目录管理 |
| `login` | 扫码、网页登录与 Cookie 流程 |
| `keybindings` | 快捷键定义与映射 |
| `storage` | BoltDB 持久化与迁移 |
| `configs` | 配置定义、升级与 TOML 编辑 |
| `deps` | `go.mod` 与 vendor 更新 |
| `build` | `Makefile`、`hack/` 与 CI |
| `docs` | 文档 |

#### 标题规则

1. 全部英文，祈使句、现在时：`add`、`fix`、`remove`，不写 `added`、`fixes`、`updating`
2. 不超过 50 字符（硬上限 72），结尾不加句号
3. 描述具体行为：`fix(player): resume mpv after seek`，而不是 `fix: bug`、`update code`、`changes`
4. 不写 issue 编号、不加 `[tag]` 前缀或 emoji
5. scope 后描述使用小写开头（专有名词如 `MPV`、`Kitty` 保持原样）
6. 禁止中文字符：标题、body、footer 与示例提交文本均不得包含中文

#### Body 与 Footer

- **Body**：说明动机与影响（为什么改、改变了哪些行为、兼容性后果），不复述 diff；每行不超过 72 字符，可用 `-` 列表
- **`BREAKING CHANGE: <描述>`**：紧随 `!` 使用，写明破坏点与迁移方式；仅有 `!` 无法让人判断影响范围
- **`Closes #<n>` / `Refs #<n>`**：关联 issue/PR；纯文档或内部重构不必强行关联
- **`Co-authored-by: Name <email>`**：多作者协作
- 依赖升级用 `deps` scope；仅调整构建脚本用 `build` scope；两者混在同一提交时以主要意图为准

#### 提交粒度

- 逻辑变更、格式化/重命名、vendor 同步分开提交，便于 review 与回滚
- 不提交无关文件、调试输出、临时脚本与本地生成物
- 回滚使用 `revert: <被回滚提交的标题>`，在 body 中引用原 commit hash 与原因

#### 示例

功能（含 body 与 footer）：

```
feat(player): add mpv playback engine support

Add an IPC-controlled backend so existing mpv users can reuse their
audio filters and output configuration.

Closes #123
```

修复（单行标题）：

```
fix(lyric): keep active line when switching render mode
```

破坏性变更：

```
refactor(configs)!: move visualizer options under [main.visualizer]

BREAKING CHANGE: keys under `[main].spectrum` are no longer read; migrate
them to `[main.visualizer]` before upgrading.
```

#### 提交前检查

- [ ] 标题与 body、footer 全部为英文，无中文字符
- [ ] 标题符合 `<type>(<scope>): <description>`，≤50 字符，英文祈使句，无句号
- [ ] `type` 与实际变更一致，未用 `chore` 掩盖行为变化
- [ ] body（如有）说明动机与影响，每行 ≤72 字符
- [ ] 破坏性变更同时使用 `!` 与 `BREAKING CHANGE:` footer
- [ ] 无无关文件、调试残留、临时产物；vendor 变更与代码变更已拆分
- [ ] issue 关联写在 footer，而非标题
