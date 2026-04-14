# go-musicfox 项目架构文档

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
"这个 PR 修复了歌词封面图圆角功能的性能问题，优化后 CPU 占用从 10% 降至 2%。"

✗ 错误示例（英文交流）：
"This PR fixes the performance issue with lyric cover cornerRadius feature..."
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
|---------|-----------|
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
- **配置**：TOML + mapstructure
- **API**：netease-music SDK

**项目结构**：`cmd/` 入口 | `internal/` 核心业务（22个包） | `utils/` 工具 | `configs/` 嵌入式配置

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
| Beep（默认） | 跨平台 | MP3/FLAC/OGG/WAV |
| DLNA | 跨平台 | 设备投送 |
| MPV | 跨平台 | IPC 控制 |
| MPD | Linux | 远程服务器 |
| AVFoundation | macOS | 原生集成 |
| MediaPlayer | Windows | WinRT API |

### 事件处理

**文件**：`internal/ui/event_handler.go`

支持 40+ 键盘操作、鼠标事件（单击/双击/滚轮/右键）、可配置快捷键。

### 其他模块

- **歌词**：LRC/YRC 格式，支持 smooth/wave/glow 渲染模式
- **播放列表**：列表循环/顺序/单曲循环/随机/无限随机/智能心动模式
- **远程控制**：MPRIS(linux)、Now Playing(macOS)、System Media(Windows)
- **存储**：BoltDB，存储用户信息、播放状态、播放列表快照

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

## 文档维护准则

### 重要准则：修改代码后需维护 AGENTS.md

**所有贡献者在修改代码后，必须检查并更新 AGENTS.md 文档，防止文档腐化。**

#### 何时需要更新文档
- 添加、删除或重命名核心文件或目录
- 新增功能模块或组件
- 修改项目结构或架构
- 更改 API 接口或配置格式
- 添加新的播放器引擎、菜单类型、渲染模式
- 修改关键路径或重要流程
- **快捷键变更**：新增、删除或修改快捷键后，必须同步更新 README 中的快捷键说明部分（help 菜单是动态从 keybindings 配置生成的，无需手动修改）

#### 更新检查清单
- [ ] 目录结构是否准确反映当前项目结构
- [ ] 核心文件路径是否正确
- [ ] 接口定义是否与代码一致
- [ ] 新增功能的说明是否完整
- [ ] 开发指南是否需要补充
- [ ] 关键文件路径表格是否需要更新

#### 更新优先级
| 变更类型 | 优先级 | 说明 |
|---------|--------|------|
| 架构变更 | **高** | 必须立即更新 |
| 新增核心模块 | **高** | 必须添加说明 |
| 文件路径变更 | **中** | 及时更新路径表 |
| 细微修复 | **低** | 可批量更新 |

#### 维护建议
- 保持文档与代码同步，避免技术债务积累
- 使用一致的术语和格式
- 添加代码示例时确保与实际代码匹配
- 定期（如每月）审查文档完整性
- PR 审查时应包含文档检查

#### 文档腐化警告信号
- 文件路径与实际不符
- 过时的 API 接口描述
- 已删除功能仍出现在文档中
- 章节结构混乱或重复
- 与 README/CHANGELOG 存在矛盾

违反此准则可能导致：新贡献者无法理解项目结构、开发效率降低、文档失去参考价值、维护成本增加。

## Git 提交规范

### 重要准则：Git Commit Message 必须遵循 Conventional Commits 规范

**所有贡献者在提交代码时，必须遵循 [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) 规范。**

#### 提交格式

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### Type 类型

| Type | 说明 |
|------|------|
| `feat` | 新功能开发 |
| `fix` | Bug 修复 |
| `docs` | 文档更新 |
| `style` | 代码格式调整（不影响功能） |
| `refactor` | 代码重构（既不修复bug也不添加功能） |
| `perf` | 性能优化 |
| `test` | 测试相关 |
| `chore` | 构建工具、辅助工具、配置变更 |
| `revert` | 回滚提交 |

#### 示例

```
feat(player): 添加 MPV 播放引擎支持

- 支持多种音频格式
- 实现播放进度控制
- 优化内存使用

Closes #123
```

#### 规范要求

1. **必须使用英文**：提交信息、描述均使用英文
2. **动词开头**：描述部分以动词开头，使用现在时态
3. **长度限制**：标题不超过 50 字符
4. **Body 可选**：复杂变更可添加详细说明，每行不超过 72 字符
5. **引用 Issues**：在 footer 中使用 `Closes #xxx` 关联 Issue
