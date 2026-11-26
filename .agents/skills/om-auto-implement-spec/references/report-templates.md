# Report templates — user-facing output (step 4)

How `om-auto-implement-spec` reports back to the user. Reporting style contract:
`references/rules.md` (Reporting style) — full sentences, explain the why,
never compress; emojis structure the sections, the text carries the meaning.

## Final run report

```markdown
## 🎯 om-auto-implement-spec — {spec title or slug}

**Outcome:** {✅ implemented, reviewed, ready | ⚠️ implemented with caveats | ⛔ blocked} — {one full sentence: what shipped, or what stopped the run and what is needed to unblock it}
**📝 Spec:** `{repo-relative spec path}`{ (spec PR #{n}, kept design-only)} — {one sentence: how the spec was resolved (path, name, issue, or spec-PR number) and what it covers}
**🌿 Branch:** `{branch}` — {fresh implementation run | resumed the existing implementation PR, one sentence why}
**🚀 PR:** #{n} ({url}) — {ready for review | draft ({why — e.g. assumptions gated under ⚠ NEEDS HUMAN CONFIRMATION})}

### ⚙️ Engine
{One short paragraph: which engine ran and why per the selection rule — plan
size, --loop flag, or an existing implementation PR to resume — and what the
engine owned (worktree, commits, validation gate, labels, review loop).}
```

Inside the ⚙️ Engine section, state the selection decision on its own line in
the exact shape defined by the `om-auto-create-pr` skill's engine selection,
relayed verbatim from the engine's report — never reworded:

```text
Engine: <name> (steps: <N>, --loop: <yes|no>)
```

```markdown
### 🧪 Validation & 🔍 review
{One short paragraph: the validation gate outcome as the engine reported it,
the review verdict from om-auto-review-pr, what the autofix loop fixed if it
ran, and anything handed back for human judgment.}

### 📸 UI verification
{One short paragraph: what om-auto-qa-pr verified and where its screenshot
evidence lives (the PR comment link), `UI: n/a` with the reason for a purely
backend/API/docs spec, that --no-ui skipped it, or why the verify could not run
(no test env, checks not green) — noted on the PR, not fatal.}
```

End the report with the chaining reference lines on their own lines, exact
shape — the one part never decorated or reworded (`Issue:` only when an issue
drives the run):

```text
Issue: #<issue number> (link: <full issue URL>)
PR: #<PR number> (link: <full PR URL>)
Spec: <repo-relative spec path>
```
