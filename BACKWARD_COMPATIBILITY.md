# 向后兼容 — go-musicfox

go-musicfox 是一个网易云音乐 TUI 客户端，具有用户安装的二进制、用户编辑的 TOML 配置与持久化的本地状态。以下表面是**受保护契约**：破坏用户既有配置即构成破坏性变更。审查技能依据本文件检查 PR；实现技能在改动违反本文件时必须提出警告。

## 1. CLI 命令与参数

- 表面：`cmd/musicfox.go` — 子命令 `player`（默认）、`config`、`upgrade-config`、`reset`。
- 破坏性变更：删除或重命名子命令或参数；改变参数语义；改变文档化失败模式的退出码。
- 路径：废弃别名至少保留一个 minor 版本；在 CHANGELOG 中宣布移除。

## 2. 配置文件格式（TOML）

- 表面：配置路径下用户编辑的 TOML；`configs/config.toml` 中的内嵌默认值；`configs.SetTOMLValue(path, keyPath, value)` 保留无关注释、布局、已有值、未知键与文件权限；`configs.UpgradeConfig` 只追加缺失的叶子键。
- 破坏性变更：重命名或删除既有键；改变值的类型（如 bool→string）而不处理旧类型；破坏按键路径保真的重写；改变 `theme.activeTheme` 语义（名称原样持久化并在启动时重新加载）。
- 路径：新键加入内嵌默认值以便升级时补全；重命名时在弃用窗口期内读旧键并写新键；在 CHANGELOG 中记录迁移。

## 3. 快捷键

- 表面：`internal/keybindings`（`OperateType`）、`internal/ui/event_handler.go` 映射以及 README 的快捷键表。
- 破坏性变更：删除某个 `OperateType` 或其默认键；在无可配置替代的情况下改变文档化键的行为。
- 路径：弃用窗口期内保持绑定可用，或在改变默认前使其可配置；同一 PR 中更新 README。

## 4. Player 接口

- 表面：`internal/player.Player`（`Play`、`Pause`、`Resume`、`Stop`、`Toggle`、`Seek`、`PassedTime`、`PlayedTime`、`Volume`、`SetVolume`、`UpVolume`、`DownVolume`、`State`、`Close`）与 `NewPlayerFromConfig()` 分发。
- 破坏性变更：改变接口（破坏所有引擎与调用方）；改变 `types.State` 语义。
- 路径：接口变更是对所有引擎加 `internal/ui` 调用方的协调重构，随单个版本发布并在 CHANGELOG 中说明；只在引擎可以默认实现时才倾向新增方法。

## 5. 菜单接口

- 表面：`internal/ui/menu.go` — `Menu`（含 `IsPlayable`、`IsLocatable`）、`SongsMenu`、`PlaylistsMenu`；通过 `baseMenu` 嵌入与导航系统注册。
- 破坏性变更：破坏第三方菜单或导航分发的签名变化。
- 路径：仅做增量扩展；在 AGENTS.md 的"添加新菜单"指南中记录。

## 6. 状态栏组件

- 表面：`model.DefaultStatusBar.Components` — `StatusBarComponent`（仅视图）与 `InteractiveStatusBarComponent`（在局部坐标命中测试）。
- 破坏性变更：改变组件契约或状态栏负责的布局/边界语义。
- 路径：增量式；坐标命中测试必须保持既有交互组件可用。

## 7. BoltDB 存储 schema

- 表面：持久化 bucket，用于用户信息、播放状态、播放列表快照与桌面歌词窗口位置/显示器。
- 破坏性变更：重命名 bucket、未经迁移改变序列化值形状、损坏既有数据库。
- 路径：打开时在 `internal/storage` / `internal/runtime` 中运行 schema 迁移；未经迁移绝不丢弃用户数据。

## 8. 远程控制表面

- 表面：MPRIS（Linux）、Now Playing（macOS `internal/macdriver/mediaplayer`）、System Media（Windows）。框架加载失败必须静默降级（class 查找为 nil → 功能禁用）——旧版 macOS（MediaPlayer < 10.12.2、UserNotifications < 10.14）依赖这一行为。
- 破坏性变更：框架缺失时导致应用无法启动；改变暴露的元数据（标题/歌手/封面/播放状态）契约。
- 路径：保留静默降级路径；只在框架版本支持且经运行时检测后新增功能。

## 9. 桌面歌词持久化

- 表面：窗口位置与显示器选择持久化在 BoltDB 中，下次启动时恢复。
- 破坏性变更：忽略或重置已存几何数据。
- 路径：字段缺失时读取旧条目；合理兜底，绝不抹除。

## 10. TOML 编辑与升级语义

- 表面：`configs.UpgradeConfig`（以及 `upgrade-config` 子命令）必须保持幂等，保留未知键/注释/布局/权限，并以原子替换写回。
- 破坏性变更：任何以有损方式重写用户文件的行为。
- 路径：把文件保真当作带测试的正确性属性，而非优化。

## 通用规则

- 每个破坏性变更都需要 CHANGELOG 条目，并在相关处说明 `upgrade-config` 行为。
- 弃用窗口期：最短一个 minor 版本，在 changelog 中宣布，旧表面仍可用。
- `AGENTS.md` 记录架构；契约变更时在同一 PR 中更新它。
- 不确定某表面是否受保护时，按受保护处理。
