# Report templates — user-facing output (step 8)

How `om-auto-write-spec` reports back to the user. Reporting style contract:
`references/rules.md` (Reporting style) — full sentences, explain the why,
never compress; emojis structure the sections, the text carries the meaning.

## Final run report

```markdown
## 🎯 om-auto-write-spec — {spec title}

**Outcome:** {✅ spec PR published ready for review | ⚠️ spec PR left as draft — assumptions gated | ⛔ blocked} — {one full sentence: what was delivered, or what stopped the run and what is needed to unblock it}
**📝 Spec:** `{repo-relative spec path}` — {one sentence: what the spec covers and which brief or issue it answers}
**🌿 Branch:** `spec/{slug}`
**🚀 PR:** #{n} ({url}) — {ready for review | draft because {which ⚠ NEEDS HUMAN CONFIRMATION assumptions gate the merge}}

### ⚠️ Assumptions posted
{One short paragraph: how many Open Questions the autonomous run resolved, that
the resolved-assumptions table was posted on the PR (and on the issue when
issue-driven) for human override, and which answers — if any — carry the
⚠ NEEDS HUMAN CONFIRMATION marker and why. When the spec had no Open
Questions, say so in one sentence.}

### 📸 Evidence
{One short paragraph: which current-app screenshots and proposed-UI mockups
were attached to the PR and what they show, or why visuals were skipped
(--no-mockups, no browser provider or test-env descriptor, or a spec with no
user-facing surface).}

### 🏷️ Labels
{One label per line with its emoji and a full-sentence reason — mirror the
rationale comments posted on the PR; when labels are disabled in config, say so
here instead of listing labels:}
- 🔍 `review` — {full-sentence reason}
- 📝 `documentation` — {full-sentence reason}
- ✅ `skip-qa` — {full-sentence reason — a design-only document has no runtime surface to QA}
- 🔥/🔺/🔹/🔽 `{priority-*}` — {full-sentence reason}
- ⚠️/🟡/🟢 `{risk-*}` — {full-sentence reason}

### 🔁 Hand-off
{One sentence: the spec PR stays design-only — implement it via
`om-auto-implement-spec` (or `om-auto-fix-issue` when issue-driven), which
ships the implementation on its own PR referencing this one.}
```

End the report with the chaining reference lines on their own lines, exact
shape — the one part never decorated or reworded (`Issue:` only when
issue-driven):

```text
Issue: #<issue number> (link: <full issue URL>)
PR: #<PR number> (link: <full PR URL>)
Spec: <repo-relative spec path>
```
