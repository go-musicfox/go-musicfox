# Report templates — user-facing output (step 13)

How `om-auto-create-pr` reports back to the user at the end of a run. Reporting
style contract: `references/rules.md` (Reporting style) — full sentences,
explain the why, never compress; emojis structure the sections, the text
carries the meaning. The PR summary comment has its own template:
`references/summary-comment-template.md`.

## Final run report

```markdown
## 🚀 om-auto-create-pr — {brief, one line}

**Status:** {✅ complete | 🔁 partial — resume with `om-auto-continue-pr {prNumber}`} — {one full sentence: what state the run reached and why — every Progress step done and the gate green, or which blocker or timeout stopped it}
**PR:** #{number} ({url}) — {flipped to ready for review | left a draft because the run is unfinished}, opened against the configured base branch.
**Branch:** `{branch}` — {one sentence: why `fix/` or `feat/` was chosen for this work}
**Plan:** `{RUNS_DIR}/{DATE}-{SLUG}.md` — {one sentence: what the plan covers; its Progress checklist is what `om-auto-continue-pr` resumes from}
**Engine:** `Engine: <name> (steps: <N>, --loop: <yes|no>)` — the exact shape from `references/engine-selection.md`, never reworded; on a handoff this report relays the `om-auto-create-pr-loop` report with this line prefixed.

### 🎯 Goal & scope
{Short paragraph: the goal in one sentence, the Phases implemented, the notable autonomous decisions made along the way, and any explicit Non-goals deliberately left untouched.}

### 🧪 Tests & validation
{Full sentences: which tests were added or updated and why, which `validation.commands` ran and their results, the `om-auto-review-pr` verdict and what fixes it landed, and any failure with what was done about it. 📸 Mention UI evidence attached to the PR when UI was touched.}

### 🏷️ Labels
{One or two sentences: which pipeline, QA-meta, priority, and risk labels were applied and the reasoning — or that labels are disabled in config and label work was skipped.}

### {✅ Done | 🔁 Resume}
{Complete: what happens next — review, the QA gate when `needs-qa`, merge hand-off. Partial: which Progress step is next, why the run stopped there, and the exact resume command `om-auto-continue-pr {prNumber}`.}
```

End the report with the chaining reference lines on their own lines, exact
shape (the one part never decorated or reworded):

```text
PR: #<number> (link: <full PR URL>)
Issue: #<number> (link: <full issue URL>)   <- only when the run has a subject issue
```
