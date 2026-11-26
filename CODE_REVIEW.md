# 代码审查规则 — go-musicfox

由 `om-code-review`（因此也由 `om-auto-review-pr`）在每个 PR 上执行。这是仓库级检查清单；它补充技能内置的检查清单。

## 验证门禁

每次审查都假定作者已运行门禁；红色门禁本身就是一个 blocker 发现，与代码发现一起报告，绝不代替它们：

- `make lint` — golangci-lint（`internal/...`、`utils/...` 及 `.golangci.yml` 范围内的全部包）
- `make test` — `go test ./internal/... ./utils/...` 并生成覆盖率报告
- `make build` — 带 `enable_global_hotkey,purego` 构建标签的完整构建

## 审查优先级

1. **正确性** — 改动是否实现了其声称的行为，包括错误路径与并发。
2. **跨平台兼容性** — go-musicfox 发布 macOS/Linux/Windows 三平台。在作者平台上可构建、却在其他平台构建失败的改动是 blocker。
3. **安全与凭据** — token、cookie、API key 与用户数据绝不能泄漏到日志、注释或配置中。网络代码（netease API 客户端、`req/v3` 用法）必须遵循仓库的反检测与 cookie-jar 约定。
4. **契约表面** — 见 `BACKWARD_COMPATIBILITY.md`；触及受保护表面的改动需要走要求的路径，而不是静默破坏。
5. **性能** — UI 刷新与音频路径按帧运行；热路径（歌词渲染、频谱分析、状态栏）上避免分配与阻塞工作。

## 仓库专属检查项

- **UI 层（bubbletea + foxful-cli）** — 新菜单嵌入 `baseMenu` 并实现 `internal/ui/menu.go` 的接口（`Menu`、`SongsMenu`、`PlaylistsMenu`）；新渲染器在 `netease.go:Components()` 注册。键盘映射必须走 `internal/keybindings` + `event_handler.go`，且 README 的快捷键表必须在同一 PR 中更新。
- **配置兼容性** — TOML 编辑走 `configs.SetTOMLValue`（按键路径保真）。新配置键必须添加到内嵌默认 `configs/config.toml`，使 `configs.UpgradeConfig` 能补全。未经迁移不得重排、重命名或改变既有键的类型。
- **播放器引擎** — 严格实现 `internal/player.Player`；在 `player.go:NewPlayerFromConfig()` 注册。macOS 框架桥接（`internal/macdriver/*`、`internal/webkitgtk`）在旧系统版本上必须静默降级（class 查找为 nil → 功能禁用），绝不 panic。
- **存储** — BoltDB 用法必须保留既有 bucket；schema 变更需要在 `internal/storage` 或 `internal/runtime` 中提供迁移路径。
- **错误处理** — 遵循 `utils/errorx` 约定；给错误附加上下文；除仓库文档明确说明的有意降级外，不静默吞错。
- **并发** — 音频、歌词、频谱与远程控制路径使用 goroutine；检查竞态（`go vet`）、channel 泄漏以及后台 goroutine 访问 UI 线程的问题。macOS 主线程派发必须遵循既有的 `performSelectorOnMainThread` helper 模式。
- **依赖** — 更新走 `make vendor`（go mod tidy + vendor + CGo 头文件）。优先在上游更新 foxful-cli（打 tag + 升版本），而不是本地打 vendor 补丁。
- **构建脚本** — `Makefile` / `hack/` 的改动必须保持 Windows 兼容（见 AGENTS.md 检查清单）：`$(OS)` 分支、`.sh` 的 PowerShell 对应实现、`nul` 与 `/dev/null` 的替代。
- **文档维护** — 按 AGENTS.md，核心文件、API、配置格式或快捷键的变更必须在同一 PR 中更新 AGENTS.md。
- **风格** — Go 代码遵循 golangci-lint 默认；代码注释用英文；提交信息遵循 Conventional Commits 规范（英文）；面向用户的交流用中文。

## 严重度指引

- **Blocker** — 破坏某个平台、契约表面、构建或门禁；数据丢失；凭据暴露；UI 死锁。必须在合并前修复。
- **Major** — 真实流程中的错误行为、失败路径缺少错误处理、竞态条件、新行为缺少测试、缺少文档更新。应在合并前修复；仅在明确超出范围且已建档跟踪时可转为后续 issue。
- **Minor** — 风格细节、命名、死代码、可选改进。可不阻塞地转为后续 issue。
- 结论判定：任何 blocker 或未修复的 major → `changes-requested`。只有 blocker/major 需要复审；复审覆盖修复内容加任何新增行。
