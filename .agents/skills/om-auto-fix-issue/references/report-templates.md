# Report templates — user-facing output (step 12)

How `om-auto-fix-issue` reports back to the user. Reporting style contract:
`references/rules.md` (Reporting style) — full sentences, explain the why,
never compress; emojis structure the sections, the text carries the meaning.

## Final run report

```markdown
## 🎯 om-auto-fix-issue — issue #{issueId}: {title}

**Status:** {✅ fixed | ✅ no action needed | ⚠️ already in progress | ⛔ blocked} — {one full sentence: what the run delivered, or exactly why and where it stopped}
**Issue mode:** {existing | filed from brief (new | reused duplicate)} — {one sentence: the issue was passed in directly, or brief mode filed a new issue / reused a duplicate via om-prepare-issue}
**Route:** {bug chain | feature route} — {one sentence: which signals (labels, title/body wording) drove the classification}
**Branch:** {branch | — (the run stopped before implementation)} — {one sentence: where the work lives, or why no branch was created}
**PR:** {#{number} ({url}) | —} — {one sentence: opened fresh, continued on an existing PR via the reuse guard, or none and why}

### 🔍 Review
{The om-auto-review-pr outcome in full sentences: the final verdict, how many autofix iterations ran, what was fixed, and any documented non-actionable findings left for a human. When the loop was skipped, say `Review: skipped` plus the concrete reason, and note that the PR stays in the `review` pipeline state for a human or a later sweep.}

### 📸 UI verification
{The om-auto-qa-pr outcome in full sentences: what was exercised in a real browser and the verdict, noting that `needs-qa` stays on for the human QA gate. When it did not run, say which case applied and why: `UI: n/a` (purely backend/API/docs fix), `UI: skipped (--no-ui)`, or `UI: skipped — {reason}` (e.g. no test env or browser provider — noted on the PR, not fatal).}

### 🧪 Tests
{Which regression tests were added and what behavior they pin down; which validation commands ran and their results, calling out any failure and what was done about it. Carry the om-root-cause `LOW_CONFIDENCE` flag here when it was raised, so a human reviewer looks harder.}

### {⚠️ Needs human | ✅ Done}
{When something needs a person: describe it concretely and state what decision or action is needed. When clean: state what happens next — the pipeline's review and QA gates own the merge; this run never merges and never adds `qa-approved`.}
```

When the run stopped at the step-4 triage gate (`om-verify-in-repo` returned
`NO_ACTION_NEEDED`), there is no branch, PR, review, or UI outcome to report:
set **Status:** `✅ no action needed` and cite the gate's evidence in full
sentences — the existing PR, commit hash, file path, or explanation showing the
issue is already handled — instead of a branch and PR.

End the report with the chaining reference lines on their own lines, exact
shape (the one part never decorated or reworded):

```text
PR: #<number> (link: <full PR URL>)
Issue: #<number> (link: <full issue URL>)   <- when the run has a subject issue
```
