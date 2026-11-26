# Report templates — user-facing output (step 6 / CI-only mode)

How `om-auto-fix-pr` reports back to the user. Reporting style contract:
`references/rules.md` (Reporting style) — full sentences, explain the why,
never compress; emojis structure the sections, the text carries the meaning.

## Final run report (full merge-ready mode, step 6)

```markdown
## 🚀 om-auto-fix-pr — PR #{prNumber}: {title}

**Merge-readiness:** {✅ merge-ready | ⚠️ needs human | ⛔ blocked after {n} iterations} — {one full sentence: what makes the PR ready, or which blocker remains and why the loop could not clear it}
**Base merge:** {merged the latest base cleanly | resolved conflicts in {files} | handed to the fork carry-forward flow → replacement PR #{n} with supersede + credit} — {one sentence: why, so the reader knows review and CI judged the real merge result}
**Review:** {✅ approvable after {n} autofix iterations | ❌ changes remain} — {one sentence: what om-auto-review-pr fixed and what, if anything, it deliberately left}
**Reviewer comments:** {✅ all {n} asks already on the PR addressed | {n} fixed, {m} deferred to follow-ups, {k} declined | none were open} — {one sentence: whose feedback it was and how each ask ended up, so a reviewer sees their comment was not ignored}
**CI:** {✅ all required checks green | ❌ {check} still failing} — {one sentence: how each failure was classified (real bug / test bug / flake / infra) and resolved — never by weakening a test or disabling a check}
**UI:** {✅ om-auto-qa-pr verdict | n/a — no user-facing surface in the diff | skipped (--no-ui) | skipped: {reason}} — {one sentence: what the evidence shows, or why UI verification did not apply}
**Labels:** {🚀 `merge-queue` (+ 🧪 `needs-qa` while the QA gate is on) | ⛔ `blocked` / ❌ `changes-requested` | labels disabled in config} — {one sentence: why this set reflects the PR's real state; `qa-approved` is never added here}
**Draft state:** {promoted to ready via mark-pr-ready | left draft ({reason: spec-only design PR | ⚠ NEEDS HUMAN CONFIRMATION guard}) | already ready}

### 🔁 Follow-ups filed
- {issue #{number} ({url})} — {the non-blocking finding it tracks and why it was deferred to a follow-up rather than fixed in-loop}
{Or one sentence stating that every finding was fixed in-loop and no follow-ups were needed.}

### {🚀 Hand-off | ⚠️ Remaining blockers}
{When merge-ready: state that this skill never merges — the PR is handed to om-approve-merge-pr / om-merge-buddy behind the QA gate — and what a human still owes (e.g. QA sign-off while `needs-qa` is on). When blocked: describe each remaining blocker concretely and what decision or fix is needed.}
```

## CI-only run report (`--ci-only`)

```markdown
## 🧪 om-auto-fix-pr --ci-only — {PR #{prNumber}: {title} | branch {name}}

**CI result:** {✅ all required checks green | ❌ still red after {n} fix→push→re-check cycles} — {one full sentence: what got the checks green, or which failure survived the loop and why}

### 🧪 Failures handled
- {check/test} — classified as {real bug | test bug | flake | infra} — {what was done about it and why, one full sentence per failure}

{In plain-branch mode, note that there was no PR comment or label mutation — the pushed branch and this report are the deliverable.}
```

End every report with the chaining reference lines on their own lines, exact
shape (the one part never decorated or reworded). In plain-branch CI-only mode
there is no PR, so the lines are omitted:

```text
PR: #<number> (link: <full PR URL>)
Issue: #<number> (link: <full issue URL>)   <- only when the run has a subject issue
```
