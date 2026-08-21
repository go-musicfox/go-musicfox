---
name: ask-anhoder
description: 自动路由并调用合适的 open-mercato/skills 技能（om-*）。当用户描述一个任务但不确定该用哪个技能、或需要端到端工作流串联时使用。确定路由后自动加载目标技能并按其工作流执行，不只给建议。
disable-model-invocation: true
---

# Ask Anhoder

你不必记住全部 36 个技能，按场景问这里。确定路由后**自动调用**目标技能并执行（见「自动调用协议」），而不是只输出一条 `Next:` 建议。

技能本体（`SKILL.md`）是权威参考；本技能是路由地图。完整速查表见仓库根目录 `SKILLS_WORKFLOW.md`。

## 主流程：想法 → 合并的 PR

大多数工作走这条主线。按现状选入口，后续环节由技能内部链与标记行自动串联，无需手动接续。

1. **已有 issue / 问题描述** → `om-auto-fix-issue <issue-id|问题描述>` — 自动分类后端到端执行：bug 走 `om-verify-in-repo` → `om-root-cause` → `om-fix` → `om-open-pr` → `om-auto-review-pr` 链（UI 修复附带 `om-auto-qa-pr` 截图证据）；feature 走写 spec → 实现链。常用变体 `--no-ui`（跳过 UI QA）、`--loop`（长计划）、`--interactive`（每步确认）、`--force`（接管他人认领）。
2. **只有任务 brief / 模糊想法** → `om-auto-create-pr "<brief>"` — 隔离 worktree 中规划 → 分阶段实现 → 跑验证门禁 → 打开带标签 PR；Step 数超 20 自动转 loop 版（`--loop` 强制）。若想法还太模糊、值得先想清楚再动手 → `om-brainstorm "<主题>"`（交互式发散收敛，产出移交 brief 与 `Next:` 路由行）。
3. **已有 spec / spec PR** → `om-auto-implement-spec <spec 路径|名称|issue|PR 号>` — 实施 spec，交付已验证、已评审、UI 已截图的 ready PR。
4. **想法清晰，先建档** → `om-prepare-issue "<想法>"` — 查重、贴 spec、打 SDLC 标签，建 issue 不实施；之后可再路由 `om-auto-fix-issue`。

## On-ramps（汇入主流程的入口）

- **PR 推进与合并**
  - "这个 PR 下一步该干什么 / 推进到合并" → `om-pr-autopilot <pr>` — 诊断十信号 → 自动编排对应技能链 → 状态报告，默认停在 merge-ready。
  - "恢复中断的 PR" → `om-auto-continue-pr <pr>` — 无计划则从 PR 上下文收养重建；计划过长自动移交 loop 版。
  - "把这个 PR 弄到能合并" → `om-auto-fix-pr <pr>` — 合入最新 base → 循环修评审发现/CI/UI QA；`--ci-only` 只修 CI。不自己合并。
  - "合并这个 PR" → `om-approve-merge-pr <pr>` — 复核全部门禁（含 QA 门禁）→ approving review → squash 合并；有阻塞信号时拒绝。
- **评审**
  - "评审这个 PR" → `om-auto-review-pr <pr>`（单 PR）；"评审所有未评审 PR" → `om-review-prs`（批量，最新在前，尊重 in-progress 锁）。
- **提交分支** → `om-check-and-commit` — 先跑完整验证门禁，全绿且用户明确要求才提交推送。

## Standalone（独立技能）

- **QA 与测试**：`om-auto-qa-pr <pr>`（浏览器 UI QA + 截图证据）；`om-prepare-test-env`（一次性准备可复用测试环境）；`om-integration-tests`（探索运行中的应用编写 E2E 测试）。
- **Issue 卫生**：`om-auto-manage-issues`（批量清洗：补标签/澄清/标注缺 spec，`--write-missing-specs` 顺带补 spec）；`om-close-fixed-issues`（按 fixes/closes/resolves 关键字关闭已修复 issue）；`om-followup-issue-from-pr <链接>`（PR 遗留诉求 → 跟踪 issue）。
- **UX 设计**：`om-ux-setup`（抽取仓库设计契约到 `.uxproof/`，一次）；`om-ux-shape "<想法>"`（塑形决策，Shape/Review/Handoff 三模式）；`om-ux-review-pr <pr>`（证据优先的 PR UI 设计评审）。
- **管道与技能维护**：`om-setup-agent-pipeline`（重新配置管道，仓库已配置过，重跑前会先询问）；`om-apply-upgrade-notes`（技能升级后同步描述符，`--dry-run` 先预览）；`om-pipeline-retro`（管道运行成本复盘，只读）；`om-create-skill "<brief>"`（写新技能/拆分大 SKILL.md）。
- **发布收尾**：`om-auto-update-changelog --version <x.y.z>`（为上次 release 以来合并的 PR 起草 CHANGELOG 条目并落 docs PR）。
- **只读咨询**（不认领、不改动）：`om-merge-buddy`（哪些 PR 现在能合并）、`om-root-cause`（issue 根因分析）、`om-code-review`（diff/分支审查）、`om-verify-in-repo`（缺陷分流验证）。

## 自动调用协议

路由不是终点。确定目标技能后**立即自动调用并执行**，不要只输出建议：

1. **加载目标技能**：读取 `../om-<skill>/SKILL.md`（ask-anhoder 与 om-* 同处 `.agents/skills/`；`.claude/skills/om-<skill>/SKILL.md` 是同一文件的符号链接）。工作流中引用的 `references/*.md` 一并读取。配置加载、认领协议等共同契约由目标技能自行处理。
2. **解析参数**：按目标技能的 Arguments 从用户描述中提取所需参数（issue/PR 号、brief、spec 路径、`--no-ui`/`--loop`/`--force` 等标志）。关键参数缺失时先问用户（见路由规则 2），不要猜测 issue 号或 PR 号。
3. **verbatim 执行**：按目标技能 SKILL.md 的 Workflow 逐条执行，遵循其 Rules 与 Security boundaries；链内技能的输出（如 `om-root-cause` 的 brief）原样传递给下一环节。不自行发明步骤，不绕过其门禁。
4. **保留标记行**：汇报结束时转述目标技能产出的机器可解析链式行（`PR:` / `Issue:` / `Spec:` / `Next:` / `NO_ACTION_NEEDED`），供下一环节自动串联或让用户知道停在哪。
5. **停下等待用户的两种情况**：(a) 目标技能自身要求人类介入——`om-brainstorm` 的路由确认、认领冲突询问、`--interactive` 模式、`om-ux-shape`/`om-ux-review-pr` 的对话环节；(b) 目标技能产出阻塞结论（`NO_ACTION_NEEDED`、`Status: blocked`、需用户决策的守卫）——向用户报告原因与证据，不自行绕行。

## 路由规则

1. **一次一个路由**：给出路由的同时就自动调用（见自动调用协议），不罗列所有可能。
2. **先澄清再路由**：用户描述模糊（没有 issue 号、没有 PR 号、不知道是 bug 还是 feature）时，先问一个问题补齐关键信息再路由。
3. **长诉求拆链**：多个阶段的诉求按顺序给主入口即可——如"实现功能并合并"只路由 `om-auto-fix-issue` 或 `om-auto-create-pr`，后续环节由技能自动串联（标记行 `PR:`/`Issue:`/`Spec:` 驱动）。
4. **只读咨询不认领**：用户只是了解情况（"哪些 PR 能合并"、"issue 根因"、"管道成本"）时，路由到只读技能（`om-merge-buddy`、`om-root-cause`、`om-code-review`、`om-pipeline-retro`），同样自动调用，但不产生认领与变更。
5. **兜底**：拿不准时读 `SKILLS_WORKFLOW.md` 速查表；流程细节查 `SDLC.md`（标签状态机/QA 门禁/认领协议）、审查规则查 `CODE_REVIEW.md`、契约表面查 `BACKWARD_COMPATIBILITY.md`。
6. **交互式技能例外**：目标技能本身是对话型（`om-brainstorm`、`om-ux-shape` 等）时，按其 workflow 进行对话，确认路由后由它决定下一技能，本技能不越过其确认门。

## 本仓库环境提示

- 配置：`.ai/agentic.config.json`（tracker github、browser agent-browser、门禁 `make lint`/`make test`/`make build`、QA 门禁开启）。
- 沟通用中文；代码注释与提交信息用英文；提交遵循 Conventional Commits。
- 用户可见改动建议带 `needs-qa`；纯文档/依赖/CI 改动用 `skip-qa`。
