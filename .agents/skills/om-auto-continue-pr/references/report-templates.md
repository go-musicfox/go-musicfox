# Report templates — user-facing output (step 10)

How `om-auto-continue-pr` reports back to the user at the end of a resume.
Reporting style contract: `references/rules.md` (Reporting style) — full
sentences, explain the why, never compress; emojis structure the sections, the
text carries the meaning. The PR summary comment has its own template:
`references/summary-comment-template.md`.

## Final run report

```markdown
## 🔁 om-auto-continue-pr — PR #{prNumber}: {title}

**Status:** {✅ complete | 🔁 still in-progress — re-run `/om-auto-continue-pr {prNumber}`} — {one full sentence: every Progress step now `- [x]` and the gate green, or which step, blocker, or timeout stopped this resume}
**Resume point:** {phase.step} — {one sentence: how it was found (first unchecked Progress line, or the `--from` override) and how far this resume got from there}
**PR:** #{prNumber} ({url}) — {flipped from draft to ready via mark-pr-ready | still a draft because the run remains in-progress}.
**Branch:** `{branch}` — {one sentence: resumed at the PR head, history untouched}
**Plan:** `{plan path}` — {one sentence: which Progress checkboxes this resume flipped, with their commit SHAs}

### 🎯 What this resume changed
{Short paragraph: what was implemented on top of the previous state, the notable autonomous decisions made, and anything deliberately left for a later resume.}

### 🧪 Tests & validation
{Full sentences: which tests were added or updated and why, which `validation.commands` ran and their results, the `om-auto-review-pr` verdict and what fixes it landed, and any failure with what was done about it. 📸 Mention UI evidence attached to the PR when UI was touched.}

### 🏷️ Labels
{One or two sentences: which labels were preserved, added, or raised and the reasoning (resume semantics: keep non-terminal pipeline states, `needs-qa` for newly user-facing work, preserve priority/risk) — or that labels are disabled in config.}

### {✅ Done | 🔁 Re-enter | ⛔ Handed off}
{Complete: what happens next — review, the QA gate when `needs-qa`, merge hand-off. In-progress: which Progress step is next, why the resume stopped there, and the exact re-entry command `/om-auto-continue-pr {prNumber}`. Spec-only guard: explain that implementation ships on its own PR and the hand-off to `om-auto-implement-spec {SPEC_PATH}`.}
```

End the report with the chaining reference lines on their own lines, exact
shape (the one part never decorated or reworded):

```text
PR: #<number> (link: <full PR URL>)
Issue: #<number> (link: <full issue URL>)   <- only when the run has a subject issue
```
