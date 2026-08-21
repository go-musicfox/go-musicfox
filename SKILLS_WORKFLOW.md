# 代理技能使用工作流（open-mercato/skills）

本文档是 go-musicfox 仓库当前安装的 [open-mercato/skills](https://github.com/open-mercato/skills) 代理技能集合（36 个 `om-*` 技能 + `ask-anhoder` 路由入口）的使用指南：哪些场景用哪个技能、如何串联成端到端工作流、有哪些共同约定。技能本身（`SKILL.md` 及 `references/`）是权威参考，本文档是路由地图——不确定时先看对应技能的 SKILL.md。

## 路由入口：ask-anhoder

不确定该用哪个技能时，直接说你的诉求，`ask-anhoder` 会给出唯一的路由并**自动调用**目标技能执行：加载该技能的 `SKILL.md`，按其工作流 verbatim 运行，而非只输出建议。长诉求只给主入口，后续环节由技能间的 `PR:`/`Issue:`/`Spec:` 标记行自动串联。本文件速查表是它的兜底索引。

## 安装与配置

- **安装位置**：`.claude/skills/om-*/`（符号链接指向 `.agents/skills/om-*/`），两处为同一集合。
- **管道配置**：`.ai/agentic.config.json`（本仓库已配置：`tracker: github`、`browser: agent-browser`、验证门禁 `make lint` / `make test` / `make build`、完整标签分类法、`qaGate: true`）。
- **描述文件**：`.ai/trackers/github.md`（gh CLI 实现所有 tracker 操作）、`.ai/browsers/agent-browser.md`（浏览器自动化契约）。
- **流程文档**：`SDLC.md`（交付流程与标签状态机）、`CODE_REVIEW.md`（审查规则，`om-code-review` 自动套用）、`BACKWARD_COMPATIBILITY.md`（受保护契约表面）、`AGENTS.md`（架构约定）。

## 共同契约（所有技能）

- **Step 0 配置加载**：每个技能先加载 `.ai/agentic.config.json` 与 tracker 描述符；配置缺失时自动指向 `om-setup-agent-pipeline` 重新配置。
- **标记行**：PR 产出技能在报告末尾输出机器可解析的链式引用行 `PR: #<number> (link: <url>)`、`Issue: #<number> (link: <url>)`、`Spec: <路径>`，供下一个技能接手。
- **认领协议**：三信号认领（assignee + `in-progress` 标签 + `🤖` 认领注释）；`ci-monitoring` 不是认领，只表示 CI 结果跟进注释欠着。
- **验证门禁**：每个 PR 合并前必须依次通过 `make lint`、`make test`、`make build`（见 `SDLC.md` 验证门禁章节）。
- **QA 门禁**（本仓库开启）：带 `needs-qa` 的 PR 必须同时带 `qa-approved` 才能合并；自动化从不施加 `qa-approved`。
- **只读技能**（不认领、不改动）：`om-verify-in-repo`、`om-root-cause`、`om-code-review`、`om-merge-buddy`、`om-pipeline-retro`、`om-brainstorm`、`om-ux-review-pr`。

## 端到端工作流

### 1. 从 issue 到修复/功能 PR（推荐入口）

```bash
om-auto-fix-issue <issue-id|问题描述>
```

一条命令完成分类与全链路：bug → `om-verify-in-repo`（分流）→ `om-root-cause`（根因）→ `om-fix`（最小改动 + 回归测试 + 门禁）→ `om-open-pr` → `om-auto-review-pr`（评审循环）→ UI 修复附带 `om-auto-qa-pr`（浏览器截图证据）；feature → 自动写 spec → 实现链。产出带完整标签的 ready PR。也可直接给问题描述（先经 `om-prepare-issue` 建档）。

常用变体：`--interactive`（每步确认）、`--no-ui`（跳过 UI QA）、`--loop`（长计划走 loop 引擎）。

### 2. 任务 brief 直接出 PR

```bash
om-auto-create-pr "实现 XXX"                # 普通引擎，Step 数超过 loopStepThreshold(20) 自动转 loop
om-auto-create-pr "实现 XXX" --loop        # 强制 loop 引擎（1 Step = 1 commit，每 ~5 Step checkpoint）
```

在隔离 worktree 中规划 → 分阶段实现 → 跑验证门禁 → 打开带标签 PR。中断后可恢复：`om-auto-continue-pr <pr>`（无计划则从 PR 上下文收养重建）。

### 3. 功能想法 → spec → 实现

```bash
om-brainstorm "新功能想法"                 # 发散收敛（交互式），输出移交 brief 到 .ai/specs/briefs/
om-prepare-issue "新功能想法"              # 建 issue（可选，先建档）
om-auto-write-spec <brief|issue-id>       # 写出 spec 并落 design-only spec PR（含 mockup/截图证据）
om-auto-implement-spec <spec 路径|名称|issue|PR 号>   # 实施 spec，交付已验证、已评审的 ready PR
```

spec 文件位于 `.ai/specs/{YYYY-MM-DD}-{kebab}.md`，由 `om-spec-writing` 引擎产出（骨架先行 + Open Questions 硬门禁）。

### 4. PR 驱动到合并（自动化闭环）

```bash
om-pr-autopilot <pr>                      # 统一入口：诊断状态 → 自动编排下一步 → 状态报告
```

手动链等价于：`om-auto-review-pr <pr>`（评审）→ `om-auto-fix-pr <pr>`（合入 base、修 CI、循环修复到 merge-ready）→ `om-auto-qa-pr <pr>`（浏览器 UI QA + 截图）→ `om-approve-merge-pr <pr>`（复核门禁 → squash 合并）。`om-auto-fix-pr --ci-only` 只修 CI。

### 5. 只读状态检查

```bash
om-merge-buddy          # 所有 open PR 分类：Ready / Almost ready / Blocked（不合并，只报告）
om-review-prs           # 批量评审全部未评审 PR（最新在前，尊重 in-progress 锁）
om-auto-manage-issues   # 批量清洗 issue 清单（补标签、澄清描述、标注缺 spec）
```

### 6. 合并后清理

```bash
om-close-fixed-issues           # 关闭被合并 PR 权威修复的 issue
om-followup-issue-from-pr <链接>  # 把 PR/评论里的遗留诉求转成跟踪 issue
om-auto-update-changelog        # 为上次 release 以来合并的 PR 起草 CHANGELOG 条目并落 docs PR
```

### 7. 测试环境与集成测试（QA 支撑）

```bash
om-prepare-test-env     # 一次性准备可复用测试环境（.ai/scripts/test-env-up.sh，浏览器自装）
om-integration-tests    # 探索运行中的应用编写/运行 E2E 测试
om-auto-qa-pr <pr>      # PR 的 UI 改动浏览器 QA（截图 + pass/fail 报告）
```

### 8. UX 工作流

```bash
om-ux-setup             # 抽取仓库设计契约到 .uxproof/（一次）
om-ux-shape "模糊想法"    # 塑形决策（Shape/Review/Handoff 三模式）
om-ux-review-pr <pr>    # 证据优先的 PR UI 设计评审
```

## 技能速查表

| 技能 | 职责 | 典型用法 |
|---|---|---|
| `ask-anhoder` | 总路由：按场景路由并自动调用目标技能执行 | `ask-anhoder "我想把想法做成 PR"` |
| `om-auto-fix-issue` | 一条命令端到端修 issue | `om-auto-fix-issue 123` |
| `om-auto-create-pr` | 任务 brief → 计划 → 分阶段实现 → PR | `om-auto-create-pr "brief" --loop` |
| `om-auto-continue-pr` / `-loop` | 恢复未完成的 PR（无计划则收养） | `om-auto-continue-pr 456` |
| `om-auto-write-spec` | brief/issue → spec 文档 + design-only PR | `om-auto-write-spec 789` |
| `om-auto-implement-spec` | 实施 spec → 已验证 ready PR | `om-auto-implement-spec .ai/specs/2026-08-21-x.md` |
| `om-auto-review-pr` | 隔离 worktree 评审 PR（引擎：om-code-review） | `om-auto-review-pr 456 --autofix` |
| `om-auto-fix-pr` | PR 驱动到 merge-ready（修冲突/CI/评审发现） | `om-auto-fix-pr 456` |
| `om-auto-qa-pr` | 浏览器 UI QA + 截图证据 | `om-auto-qa-pr 456` |
| `om-approve-merge-pr` | 复核门禁 → approving review → squash 合并 | `om-approve-merge-pr 456` |
| `om-pr-autopilot` | 单 PR 状态诊断 + 自动编排 | `om-pr-autopilot 456` |
| `om-merge-buddy` | 只读：哪些 PR 现在能合并 | `om-merge-buddy` |
| `om-review-prs` | 批量评审未评审 PR | `om-review-prs` |
| `om-prepare-issue` | 从 brief 建格式良好的 issue（不实施） | `om-prepare-issue "想法" --priority high` |
| `om-auto-manage-issues` | 批量清洗 issue 清单 | `om-auto-manage-issues --limit 25` |
| `om-verify-in-repo` / `om-root-cause` | autofix 链只读分流 / 根因分析 | 链内步骤，一般不单独调用 |
| `om-fix` | 按根因 brief 实施最小改动（不提交） | 链内步骤 |
| `om-open-pr` | 共享 PR 开启器（提交推送 + 标签） | 链内步骤 |
| `om-check-and-commit` | 跑门禁验证分支并提交推送 | `om-check-and-commit` |
| `om-close-fixed-issues` | 关闭已修复 issue | `om-close-fixed-issues` |
| `om-followup-issue-from-pr` | PR/评论 → 跟踪 issue | `om-followup-issue-from-pr <URL>` |
| `om-auto-update-changelog` | 起草 CHANGELOG 发布条目 + docs PR | `om-auto-update-changelog --version x.y.z` |
| `om-prepare-test-env` | 准备可复用测试环境 | `om-prepare-test-env` |
| `om-integration-tests` | 探索式编写/运行 E2E 测试 | `om-integration-tests` |
| `om-brainstorm` | 发散对话收敛决策 | `om-brainstorm "主题"` |
| `om-spec-writing` | 编写/审查规格（引擎） | 一般经 `om-auto-write-spec` 调用 |
| `om-ux-setup` / `om-ux-shape` / `om-ux-review-pr` | 设计契约抽取 / 塑形 / UI 评审 | 见 UX 工作流 |
| `om-pipeline-retro` | 管道运行成本复盘（只读） | `om-pipeline-retro` |
| `om-apply-upgrade-notes` | 技能升级后同步描述符 | `om-apply-upgrade-notes --dry-run` |
| `om-create-skill` | 编写新 OM 技能 / 拆分大 SKILL.md | `om-create-skill "brief"` |
| `om-setup-agent-pipeline` | 管道配置器（本仓库已配置） | 工具链/标签变化时重跑 |

## 本仓库注意点

- **标签**：完整分类法已安装（pipeline 7 + category 6 + meta 6 + priority 4 + risk 3 + `do-not-close`）；`bug`/`dependencies`/`documentation` 为仓库原有标签，语义沿用。
- **验证命令**：`make lint` / `make test` / `make build`，运行前确认本机装有 `golangci-lint`（`make lint` 依赖）。
- **QA 门禁开启**：用户可见改动（UI、新功能）建议带 `needs-qa`；纯文档/依赖/CI 改动用 `skip-qa` 豁免。
- **中文交流**：所有技能汇报与评论默认中文（AGENTS.md 约定）；代码注释与提交信息用英文。
- **未安装的提及项**：`om-auto-update-changelog` 可选提及 `om-sync-merged-pr-issues`（本集合未安装，缺失时自动降级）；`om-create-skill` 依赖仓库内 `om-filozofia.md`（未提供时该技能部分功能受限）。

## 扩展与自定义

- **仓库级技能覆写**：`.ai/skills/<skill-name>/SKILL.md` 可扩展任意技能（本地规则优先，但不能突破已安装技能的安全规则）。
- **tracker 行为**：编辑 `.ai/trackers/github.md` 可覆写任意 tracker 操作。
- **浏览器契约**：编辑 `.ai/browsers/agent-browser.md` 可覆写浏览器操作。
- **流程与门禁**：修改 `SDLC.md` 与 `.ai/agentic.config.json` 需同步；变更后重跑 `om-setup-agent-pipeline`。
