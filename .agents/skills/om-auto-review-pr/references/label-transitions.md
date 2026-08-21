# Pipeline-label mechanics, priority/risk inference, and comment text (step 10)

The label machinery `om-auto-review-pr` uses in step 10: the `set_pipeline_label` helper's behavior, the priority/risk inference rules, and the suggested comment strings. The verdict→label decisions and QA safety rules live in the skill body.

## The `set_pipeline_label` helper

Pipeline-label transitions go through the `set_pipeline_label` helper (usage: `set_pipeline_label <prNumber> <newLabel>`), which is one of the label guards from the tracker descriptor — do not redefine it here. It operates over the pipeline group:

```bash
PIPELINE_LABELS="review changes-requested qa qa-failed merge-queue blocked do-not-merge"
```

When `LABELS_ENABLED` is not `true`, the guard skips the pipeline label change (log that labels are disabled in config).

The helper:

- adds `newLabel`
- removes every other pipeline label from the list above
- preserves category labels (`bug`, `feature`, `refactor`, `security`, `dependencies`, `documentation`), meta labels (`needs-qa`, `skip-qa`, `qa-approved`, `qa-self-verified`, `in-progress`, `ci-monitoring`), priority labels (`priority-low`, `priority-medium`, `priority-high`, `priority-extreme`), and risk labels (`risk-low`, `risk-medium`, `risk-high`)

Every label change is reflected in the **single consolidated `🏷️ label rationale` comment** (template below) — updated in place, never a new comment per change.

## Priority label (always ensure exactly one, when labels are enabled)

- If the PR carries no priority label, infer one from the diff and the linked issue, apply it through the guard, and record the choice and reason in the label-rationale comment. Inference rule: outage, data loss, or a security incident → `priority-extreme`; security hardening, a release-blocking regression, or fixes touching auth/session/data-scoping/money/event-reliability → `priority-high`; ordinary bug or feature → `priority-medium`; cosmetic, docs, dependency bumps, or cleanup → `priority-low`.
- If the PR already has a priority label, keep it unless the review reveals the scope is clearly mis-rated (e.g. a "cleanup" PR that actually touches auth) — then adjust it and explain why in the comment.
- Priority is mutually exclusive: when changing it, remove the other three priority labels.

## Risk label (always ensure exactly one, when labels are enabled)

- If the PR carries no risk label, infer one from the diff and the linked issue, apply it through the guard, and record the choice and reason in the label-rationale comment. Inference rule: auth/session/data scoping/money, migrations or schema, encryption, event reliability, shared contract surfaces, or broad cross-cutting edits → `risk-high`; ordinary single-area change with tests → `risk-medium`; docs, dependency bumps, test-only, typo, or isolated cleanup → `risk-low`.
- If the PR already has a risk label, keep it unless the review reveals the scope is clearly mis-rated (e.g. a "docs" PR that actually changes a migration) — then adjust it and explain why in the comment. A `risk-high` rating reinforces the case for `needs-qa` and deeper review even when the PR would otherwise look routine.
- Risk is mutually exclusive: when changing it, remove the other two risk labels.

## The consolidated label-rationale comment

Exactly **one** marker-idempotent comment from this skill per PR describes the whole current label state: **one label per line**, each with its emoji from the map below and a full-sentence reason. On every transition (verdict change, priority/risk adjustment, re-review), find the `` 🤖 `om-auto-review-pr` — 🏷️ label rationale `` marker via **list-issue-comments** and rewrite that comment via **update-comment**; create it when absent. Never post an additional per-change comment (when the descriptor lacks **update-comment**, post a replacement stating it supersedes the previous rationale).

```markdown
🤖 `om-auto-review-pr` — 🏷️ label rationale

- 🚀 `merge-queue` — code review passed; `needs-qa` stays on, so the QA-approval gate holds the merge until a QA reviewer adds `qa-approved`.
- 🧪 `needs-qa` — the change alters user-facing behavior that a human should exercise.
- 🔺 `priority-high` — {why this priority, one full sentence}.
- 🟡 `risk-medium` — {why this risk, one full sentence}.
```

Suggested per-line reasons: `review` — "this PR is ready for code review"; `changes-requested` — "review found actionable issues: {short list}"; `merge-queue` (no QA required) — "the required review gates passed and QA is not required (or `qa-approved` is already present)"; `blocked` — "progress depends on an external blocker: {which}"; `do-not-merge` — "this PR must not merge yet: {why}". Write them as full sentences with the concrete reason — never bare labels.

Label emoji map (decoration only — parsers key on the backticked label text): 🔍 `review` · ❌ `changes-requested` / `qa-failed` · 🧪 `qa` / `needs-qa` · 🚀 `merge-queue` · ⛔ `blocked` / `do-not-merge` · 🐛 `bug` · ✨ `feature` · ♻️ `refactor` · 🔒 `security` · 📦 `dependencies` · 📚 `documentation` · ⏭️ `skip-qa` · ✅ `qa-approved` · 📸 `qa-self-verified` · 🔥 `priority-extreme` · 🔺 `priority-high` · 🔹 `priority-medium` · 🔽 `priority-low` · ⚠️ `risk-high` · 🟡 `risk-medium` · 🟢 `risk-low` · 🤖 `in-progress` · ⏱️ `ci-monitoring`.
