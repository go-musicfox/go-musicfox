# Report templates — user-facing output (step 14)

How `om-auto-create-pr-loop` reports back to the user at the end of a run.
Reporting style contract: `references/rules.md` (Reporting style) — full
sentences, explain the why, never compress; emojis structure the sections, the
text carries the meaning. The PR summary comment has its own template:
`references/summary-comment-template.md`.

## Final run report

```markdown
## 🚀 om-auto-create-pr-loop — {brief, one line}

**Status:** {✅ complete | 🔁 partial — resume with `om-auto-continue-pr-loop {prNumber}`} — {one full sentence: every Tasks row done and the final gate green, or which Step, blocker, or timeout stopped the run}
**Mode:** {Simple run | Spec-implementation run} — {one sentence: which classification rule from step 1 decided it}
**PR:** #{number} ({url}) — {flipped to ready for review | left a draft because the run is unfinished}, opened against the configured base branch.
**Branch:** `{branch}` — {one sentence: why `fix/` or `feat/` was chosen for this work}
**Run folder:** `{RUNS_DIR}/{DATE}-{SLUG}/` — {one sentence: PLAN.md carries the authoritative Tasks table, HANDOFF.md the resume snapshot, NOTIFY.md the event log — the contract `om-auto-continue-pr-loop` parses to resume} (Simple runs: no run folder — say so and why.)

### 🎯 Goal & scope
{Short paragraph: the goal in one sentence, the Phases/Steps landed (one lean commit per Step), the notable autonomous decisions made along the way, and any explicit Non-goals deliberately left untouched.}

### 🧪 Tests, checkpoints & final gate
{Full sentences: which tests were added or updated per Step, which checkpoints fired and what their `checkpoint-<N>-checks.md` recorded, the final-gate results (`validation.commands`, the integration suite via `om-integration-tests`, the style pass), the `om-auto-review-pr` verdict and what fixes it landed, and any failure with what was done about it. 📸 Mention the checkpoint/final-gate screenshot evidence posted to the PR when UI was touched, or the logged reason it was skipped.}

### 🏷️ Labels
{One or two sentences: which pipeline, QA-meta, priority, and risk labels were applied and the reasoning — or that labels are disabled in config and label work was skipped.}

### {✅ Done | 🔁 Resume}
{Complete: what happens next — review, the QA gate when `needs-qa`, merge hand-off. Partial: the first `todo` Step that HANDOFF.md points at, why the run stopped there, and the exact resume command `om-auto-continue-pr-loop {prNumber}`.}
```

End the report with the chaining reference lines on their own lines, exact
shape (the one part never decorated or reworded):

```text
PR: #<number> (link: <full PR URL>)
Issue: #<number> (link: <full issue URL>)   <- only when the run has a subject issue
```
