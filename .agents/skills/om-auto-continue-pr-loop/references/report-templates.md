# Report templates — user-facing output (step 11)

How `om-auto-continue-pr-loop` reports back to the user at the end of a resume.
Reporting style contract: `references/rules.md` (Reporting style) — full
sentences, explain the why, never compress; emojis structure the sections, the
text carries the meaning. The PR summary comment has its own template:
`references/summary-comment-template.md`.

## Final run report

```markdown
## 🔁 om-auto-continue-pr-loop — PR #{prNumber}: {title}

**Status:** {✅ complete | 🔁 still in-progress — re-run `/om-auto-continue-pr-loop {prNumber}`} — {one full sentence: every Tasks row now `done` and the final gate green, or which Step, blocker, or timeout stopped this resume}
**Resume point:** {phase.step} — {one sentence: how it was found (first non-`done` Tasks row, HANDOFF.md, or the `--from` override) and how far this resume got from there}
**PR:** #{prNumber} ({url}) — {flipped from draft to ready via mark-pr-ready | still a draft because the run remains in-progress}.
**Branch:** `{branch}` — {one sentence: resumed at the PR head, history untouched}
**Run folder:** `{run folder path}` — {one sentence: which Tasks rows this resume flipped to `done`; PLAN.md, HANDOFF.md, and NOTIFY.md were updated and pushed so the next resume can re-enter}

### 🎯 What this resume changed
{Short paragraph: the Steps landed on top of the previous state (one lean commit per Step), the notable autonomous decisions made, and anything deliberately left for a later resume.}

### 🧪 Tests, checkpoints & final gate
{Full sentences: which tests were added or updated per Step, which checkpoints fired and what their `checkpoint-<N>-checks.md` recorded, the final-gate results when reached (`validation.commands`, the integration suite via `om-integration-tests`, the style pass), the `om-auto-review-pr` verdict and what fixes it landed, and any failure with what was done about it. 📸 Mention the checkpoint/final-gate screenshot evidence posted to the PR when UI was touched, or the logged reason it was skipped.}

### 🏷️ Labels
{One or two sentences: which labels were preserved, added, or raised and the reasoning (resume semantics: preserve the pipeline state, `needs-qa`/`skip-qa` never both, preserve priority/risk) — or that labels are disabled in config.}

### {✅ Done | 🔁 Re-enter | ⛔ Handed off}
{Complete: what happens next — review, the QA gate when `needs-qa`, merge hand-off. In-progress: the first remaining `todo` Step that HANDOFF.md names, why the resume stopped there, and the exact re-entry command `/om-auto-continue-pr-loop {prNumber}`. Spec-only guard: explain that implementation ships on its own PR and the hand-off to `om-auto-implement-spec {SPEC_PATH}`.}
```

End the report with the chaining reference lines on their own lines, exact
shape (the one part never decorated or reworded):

```text
PR: #<number> (link: <full PR URL>)
Issue: #<number> (link: <full issue URL>)   <- only when the run has a subject issue
```
