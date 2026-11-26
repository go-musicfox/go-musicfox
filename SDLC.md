# 软件交付流程

## 目的

本文档说明本仓库中从 ticket 到合并 PR 的工作流转方式。`.ai/agentic.config.json` 中配置的代理技能负责执行该流程，人类在此阅读。PR 以 `master` 为目标分支；issue 与 PR 托管在 GitHub，技能运行的所有 tracker 操作由 `.ai/trackers/github.md` 定义（编辑该文件可扩展或覆写 tracker 行为）。

工作通过两条路径进入：直接交给代理的开放式任务简报，或已提交的 ticket。两者汇聚到相同的审查循环、相同的验证门禁与相同的合并门禁。

## 角色

- **作者（Author）** — 编写改动的人类或代理。从认领 ticket 到产出可合并的 PR，全程负责。
- **审查者（Reviewer）** — 阅读 diff 并批准或要求修改。可以是人类，也可以是 `om-auto-review-pr` 技能；无论哪种方式都适用 `om-code-review` 检查清单。
- **QA 审查者（QA reviewer）** — 在合并前手动验证用户可见的改动。始终以角色称呼，绝不指名道姓：人员会变动。
- **维护者（Maintainer）** — 负责分支保护、标签分类法、配置与本文档；门禁冲突时进行仲裁。

## Ticket 生命周期

| 阶段 | 发生什么 | 由谁驱动 | 完成标志 |
|---|---|---|---|
| 录入 | 在 GitHub 提交信息充足的 ticket 或任务简报。 | 任何人 | ticket 存在 |
| 分诊 | 确认 issue 真实存在、在 `master` 上仍未修复、未被认领且未被已有 PR 覆盖。只读操作；无事可做时干净地终止链路。 | `om-verify-in-repo` 或人类 | 确认可处理，或按无操作关闭 |
| 认领 | 作者认领 ticket，使并发代理退避。见下方认领协议。 | `om-fix` / `om-auto-create-pr`，或人类 | ticket 上可见认领标记 |
| 实现 | 定位最小改动面（`om-root-cause`，只读），实现改动并补充回归测试，运行验证门禁。无 ticket 的任务简报走 `om-auto-create-pr`，它在隔离 worktree 中分阶段规划、实现并运行同一门禁。 | `om-root-cause` + `om-fix`、`om-auto-create-pr`，或人类作者 | 改动完成，验证门禁通过 |
| PR | 提交、推送并向 `master` 打开带规范标签的 PR。手工分支上，`om-check-and-commit` 运行门禁、修复明显偏差并在通过后推送。 | `om-open-pr`、`om-auto-create-pr` 或 `om-check-and-commit` | PR 已打开并打好标签 |
| 审查循环 | 审查者按 `om-code-review` 检查清单阅读 diff，批准或要求修改。被要求修改的改动由 `om-auto-continue-pr` 按跟踪计划继续推进（无计划的 PR 则从 PR 自身上下文重建计划后接手），并重新审查直至批准。 | `om-auto-review-pr`（单个 PR）、`om-review-prs`（批量），或人类 | 已提交批准审查 |
| QA | 带 `needs-qa` 的 PR 等待人工 QA。QA 审查者测试并记录结果。见下方 QA 门禁。 | QA 审查者（人工） | 打上 `qa-approved`，或 `qa-failed` 将其退回 |
| 合并 | `om-merge-buddy` 只读报告哪些 PR 现在可合并、哪些接近但受阻。`om-approve-merge-pr` 复核每个门禁后批准并 squash 合并。 | `om-merge-buddy` + `om-approve-merge-pr`，或人类 | PR 以 squash 方式合并进 `master` |
| 合并后清理 | 关闭被合并 PR 修复的 issue；对未合并即关闭的 PR 关联 issue 留言；把遗留诉求或审查评论转为已跟踪的后续 issue。 | `om-close-fixed-issues`、`om-followup-issue-from-pr` | tracker 已对账，后续事项已建档 |

## 标签状态机

Pipeline 标签互斥：一个 PR 至多携带一个，标识 PR 当前所处的流程位置。

- 就绪的非 draft PR 携带 `review`。
- 审查者移动它：要求修改 → `changes-requested`；修复后回到 `review`；批准 → `merge-queue`。
- `merge-queue` 只是路由，不是 QA 证明：带 `needs-qa` 的 PR 在 QA 签字前合法地停留于此。
- 只有 QA 审查者设置 `qa` 这个 pipeline 标签。测试期间把已入队的 `needs-qa` PR 从 `merge-queue` 移到 `qa`，通过后带着 `qa-approved` 回到 `merge-queue`，失败则移入 `qa-failed`。自动化技能用 `needs-qa` 请求 QA；它们从不设置 `qa`。
- `blocked` 与 `do-not-merge` 由人类设置和清除，无论流程进行到哪一步都会叫停。

| 分组 | 标签 | 互斥性 | 含义 |
|---|---|---|---|
| Pipeline | `review`、`changes-requested`、`qa`、`qa-failed`、`merge-queue`、`blocked`、`do-not-merge` | 同一时刻一个 | 工作流状态 |
| Category | `bug`、`feature`、`refactor`、`security`、`dependencies`、`documentation` | 可叠加 | 变更类型 |
| Meta | `needs-qa`、`skip-qa`、`qa-approved`、`qa-self-verified`、`in-progress`、`ci-monitoring` | 可叠加 | 流程信号 |
| Priority | `priority-low`、`priority-medium`、`priority-high`、`priority-extreme` | 同一时刻一个；未设视为 medium | 工作紧急程度 |
| Risk | `risk-low`、`risk-medium`、`risk-high` | 同一时刻一个；未设视为 medium | 变更影响范围 |

Priority 是工作有多紧急；risk 是改动上线有多危险。给故障的一次性修复可以是 `priority-extreme` + `risk-low`；可缓行的大型认证重构可以是 `priority-low` + `risk-high`。PR 从来源 issue 继承两者，除非范围明显改变。自动化技能添加或变更 pipeline / meta 标签时，会留下简短注释说明原因。

未设置 priority 标签时按以下规则推断：

- `priority-extreme` — 生产故障、数据丢失或正在发生的安全事件。
- `priority-high` — 安全加固或阻塞发布的回归。
- `priority-medium` — 普通 bug 修复与全新功能（未设置时的默认解读）。
- `priority-low` — 外观、纯文档、依赖升级、后续清理。

未设置 risk 标签时按以下规则推断：

- `risk-high` — 认证、会话、数据边界、资金、schema 迁移、共享契约表面，或大范围跨切面改动。
- `risk-medium` — 带测试上线的普通单领域改动（未设置时的默认解读）。
- `risk-low` — 纯文档、纯测试、错别字或孤立的视觉改动。

信号冲突时取更高一级的标签，并在标签注释中说明理由。`risk-high` 的 PR 即使看起来常规，也应加强 `needs-qa` 与深入审查的理由。

一个标签位于本分类法之外：`do-not-close`，由人类施加到整理类技能绝不自动关闭的 issue 上。技能只读取它。

## QA 门禁

本流程唯一一条硬规则：**携带 `needs-qa` 的 PR 必须同时携带 `qa-approved` 才能合并，即使其余所有检查全绿。** `om-merge-buddy` 将此类 PR 归类为受阻；`om-approve-merge-pr` 拒绝合并它。

- 对 UI 改动、新功能以及其他需要人工操作的用户可见行为施加 `needs-qa`。
- `skip-qa` 是显式豁免，适用于纯文档、纯依赖、纯 CI、纯测试等低风险的、非用户可见的改动。切勿与 `needs-qa` 同时使用。
- `qa-failed`、`do-not-merge` 和 `blocked` 是硬阻断，无视其余任何信号。活跃的 `qa` pipeline 标签意味着测试者此刻正在处理该 PR——绝不在测试者活跃时合并。
- 门禁满足的条件：QA 审查者测试 PR 并施加 `qa-approved`。
- **自测例外（Self-QA exception）**：当 QA 审查者暂无时间时，任何工程师都可以代为签字——但必须满足 (1) 检出 PR 并在本地运行，(2) 操作受影响的流程，(3) 在 PR 上附上证据：一张工作正常的截图，或一份书面说明记录操作了什么、观察到了什么。然后同时施加 `qa-approved`（让门禁通过）与 `qa-self-verified`（让例外可审计）。无证据，无 `qa-approved`。

## 认领协议

在变更 issue 或 PR 之前，代理用三个信号认领：把自己设为 assignee、添加 `in-progress` 标签、发布说明正在做什么的认领注释。任何发现已有认领的代理退避，避免冲突。带 `in-progress` 的 PR 也会被合并工具跳过。

`in-progress` 表示**正在积极处理**。代理的工作完成并完整汇报后——标签已打、审查已提交、注释已发布——若仍打算报告 CI 结果，就把 `in-progress` 换成 `ci-monitoring`。`ci-monitoring` **不是**认领，不阻碍任何人：它只说明 CI 结果的跟进注释仍然欠着，因此其他代理或人类可以自由处理该 PR。这一区分很重要，因为 CI 运行时间很长：一个汇报完工作后在看跑批期间挂掉的代理，留下的是诚实、自描述的现场状态，而不是一把无人持有的锁。跟进注释落地、或代理在 `ci.maxWaitMinutes` 超时并明确声明后，该标签摘除。

工作完成时释放认领——成功与失败都一样。维护者可以清除长期无活动迹象的过期 `in-progress`。

### 汇报与 CI 解耦

代理在**工作完成的第一时间**打标签、提交审查、发布注释，不等待 CI 变绿。审查提交时若检查仍在运行，会在正文中说明这一点：分支保护加上 QA 批准门禁把守着真正的合并，批准针对的是代码，不是绿色运行。CI 结果随后以跟进注释的形式到达，若结果改变结论，也会一并纠正 pipeline 标签。

等待该结果有上限：`ci.maxWaitMinutes`（默认 40）。超时且检查仍在运行时，代理停止等待，以本地验证门禁作为自己的证据，连同仍未结束的检查名与"不再跟进"的明确声明一起发布，摘除 `ci-monitoring`，然后结束。

红色信号同样不会短路审查。失败的必要检查或冲突的头部被收集为 **blocker 发现**，与完整代码审查一起报告，绝不代替它：一次审查循环同时把失败的检查、冲突和所有代码发现交给作者，而不是先报最便宜的红旗、再花一轮循环发现其余问题。这种结论仍然是 `changes-requested`——变的是完整性，门禁没变。

以上均不影响合并门禁。提前汇报是安全的；提前合并不安全——必要检查仍把关每一次合并，合并工具在它们真正变绿前拒绝执行。

## 自动化契约

`om-auto-*` 技能无人值守地运行本流程且可链式调用：每个技能接受前一个产出的工件（issue id、spec 路径，或每个产出 PR 的技能发出的 `PR: #<number> (link: <url>)` 引用行中的 PR 编号），并检测已开始的工作——引用该 issue 或计划的开放 PR——在其上继续而不是重复打开。一次完成的自主运行留下一个 **ready**（非 draft）、标签完整的 PR——一个 pipeline 标签、一个 category、QA meta、一个 priority、一个 risk——附带运行摘要注释，用户可见的改动还附带工作应用截图的 PR 证据。Draft PR 保留给明确未完成的状态：仅 spec 的设计 PR、被打断的交接、或标记为需人工确认的自主默认行为。自动化从不施加 `qa-approved`。

## 验证门禁

每个 PR 在审查签字前按顺序通过完整验证门禁：

- `make lint`
- `make test`
- `make build`

任何非零退出都会使门禁失败并阻塞 PR。实现类技能在打开 PR 前运行门禁；`om-check-and-commit` 在推送手工分支前运行它。命令列表位于 `.ai/agentic.config.json`；变更时需同时更新该文件与本节的命令列表。

## 修订本流程

本文档与 `.ai/agentic.config.json` 描述的是同一流程：两者需一起修改；工具链或标签分类法变化时，重新运行 `om-setup-agent-pipeline` 技能。各技能的偏差——额外审查规则、不同的 PR 正文模板、附加的门禁步骤——属于 `.ai/skills/<skill-name>/SKILL.md` 中的仓库级同名技能，它优先于已安装技能（并可 `@`-import 或引用它以扩展而非替换）；本地规则优先，但仓库级技能不能授予已安装技能安全规则所禁止的能力。
