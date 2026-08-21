---
name: ask-anhoder
description: 自动路由到合适的 open-mercato/skills 技能（om-*）。当用户描述一个任务但不确定该用哪个技能、或需要端到端工作流串联建议时使用。只做路由推荐，不执行工作。
disable-model-invocation: true
---

# Ask Anhoder

你不必记住全部 36 个技能，按场景问这里。每次只给出**一个明确的路由建议**：`Next: om-<skill> <参数>`，附一句话理由与用法示例。当用户描述模糊或存在多条可行路径时，先问一个澄清问题再路由。

技能本体（`SKILL.md`）是权威参考；本技能是路由地图。完整速查表见仓库根目录 `SKILLS_WORKFLOW.md`。

## 按场景路由

### 修复与交付（主流程）

- **"修这个 issue"** → `om-auto-fix-issue <issue-id>` — 一条命令端到端：自动分类（bug 走 verify→root-cause→fix→open-pr→review→QA；feature 走 spec→实现链）。附 `--no-ui` 跳过 UI QA、`--loop` 长计划。
- **"把这个任务做成 PR"** → `om-auto-create-pr "<brief>"` — 计划→worktree→分阶段实现→验证门禁→带标签 PR；Step 数超 20 自动转 loop 版（`--loop` 强制）。
- **"功能想法，先想清楚/写 spec"** → 三选一：
  - 想法还很模糊 → `om-brainstorm "<主题>"`（交互式发散收敛）
  - 想法清晰，先建档 → `om-prepare-issue "<想法>"`（建 issue，不实施）
  - 直接要 spec + 设计 PR → `om-auto-write-spec <brief|issue-id>`
- **"实施这个 spec"** → `om-auto-implement-spec <spec 路径|名称|issue|PR 号>` — 交付已验证、已评审、UI 已截图的 ready PR。

### PR 推进与合并

- **"这个 PR 下一步该干什么 / 推进到合并"** → `om-pr-autopilot <pr>` — 诊断十信号 → 自动编排对应技能链 → 状态报告。默认停在 merge-ready。
- **"恢复中断的 PR"** → `om-auto-continue-pr <pr>` — 无计划则从 PR 上下文收养重建；计划过长自动移交 `om-auto-continue-pr-loop`。
- **"评审这个 PR"** → `om-auto-review-pr <pr>`（单 PR）；**"评审所有未评审 PR"** → `om-review-prs`（批量，尊重 in-progress 锁）。
- **"把这个 PR 弄到能合并"** → `om-auto-fix-pr <pr>` — 合入最新 base → 循环修评审发现/CI/UI QA；`--ci-only` 只修 CI。不自己合并。
- **"哪些 PR 现在能合并"** → `om-merge-buddy`（只读报告，绝不合并）。
- **"合并这个 PR"** → `om-approve-merge-pr <pr>` — 复核全部门禁（含 QA 门禁）→ approving review → squash 合并；有阻塞信号时拒绝。
- **"提交/推送当前分支"** → `om-check-and-commit` — 先跑完整验证门禁，全绿且你明确要求才提交推送。

### QA 与测试

- **"给 PR 的 UI 做 QA（截图证据）"** → `om-auto-qa-pr <pr>` — 浏览器走查 + pass/fail 报告 + 截图证据。
- **"准备可复用测试环境"** → `om-prepare-test-env` — 生成 `.ai/scripts/test-env-up.sh` 等。
- **"写/跑集成测试"** → `om-integration-tests` — 探索运行中的应用编写 E2E 测试。

### Issue 卫生

- **"清洗 issue 清单（补标签/澄清/标注缺 spec）"** → `om-auto-manage-issues` — 批量或单个，`--write-missing-specs` 可顺带补 spec。
- **"把想法建成规范的 issue"** → `om-prepare-issue "<brief>"` — 查重、贴 spec、打 SDLC 标签。

### 合并后收尾

- **"关闭已修复的 issue"** → `om-close-fixed-issues`（仅按 fixes/closes/resolves 关键字，绝不因裸 #N 动作）。
- **"把 PR 里的遗留诉求转成 issue"** → `om-followup-issue-from-pr <链接>`。
- **"起草 CHANGELOG 发布条目"** → `om-auto-update-changelog --version <x.y.z>` — 落为 docs PR。

### UX 设计

- **"抽取仓库设计契约"** → `om-ux-setup`（一次，产出 `.uxproof/`）。
- **"模糊产品/UI 想法塑形"** → `om-ux-shape "<想法>"`。
- **"评审 PR 的 UI 设计"** → `om-ux-review-pr <pr>` — 证据优先，含截图。

### 管道与技能维护

- **"重新配置管道 / 工具链或标签变了"** → `om-setup-agent-pipeline`（仓库已配置过，重跑前会先询问）。
- **"升级技能集合后同步描述符"** → `om-apply-upgrade-notes`（`--dry-run` 先预览）。
- **"管道运行复盘"** → `om-pipeline-retro`（只读成本排名）。
- **"写一个新技能 / 拆分大 SKILL.md"** → `om-create-skill "<brief>"`。

## 路由规则

1. **一次一个建议**：给出 `Next: om-<skill> <参数>` 后停，不罗列所有可能。
2. **先澄清再路由**：用户描述模糊（没有 issue 号、没有 PR 号、不知道是 bug 还是 feature）时，先问一个问题补齐关键信息。
3. **长诉求拆链**：多个阶段的诉求按顺序给出主入口即可——如"实现功能并合并"只路由 `om-auto-fix-issue` 或 `om-auto-create-pr`，后续环节由技能自动串联（标记行 `PR:`/`Issue:`/`Spec:` 驱动）。
4. **只读咨询**：用户只是了解情况（"某功能怎么实现"、"哪些 PR 能合并"、"issue 根因"）时，路由到只读技能（`om-merge-buddy`、`om-root-cause`、`om-code-review`、`om-pipeline-retro`），不产生认领与变更。
5. **兜底**：拿不准时读 `SKILLS_WORKFLOW.md` 速查表；流程细节查 `SDLC.md`（标签状态机/QA 门禁/认领协议）、审查规则查 `CODE_REVIEW.md`、契约表面查 `BACKWARD_COMPATIBILITY.md`。
6. **本技能只路由不执行**：给完建议即停，不代替目标技能工作。

## 本仓库环境提示

- 配置：`.ai/agentic.config.json`（tracker github、browser agent-browser、门禁 `make lint`/`make test`/`make build`、QA 门禁开启）。
- 沟通用中文；代码注释与提交信息用英文；提交遵循 Conventional Commits。
- 用户可见改动建议带 `needs-qa`；纯文档/依赖/CI 改动用 `skip-qa`。
