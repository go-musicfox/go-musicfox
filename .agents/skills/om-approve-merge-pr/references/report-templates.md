# Report templates — user-facing output (step 6)

How `om-approve-merge-pr` reports back to the user. Reporting style contract:
`references/rules.md` (Reporting style) — full sentences, explain the why,
never compress; emojis structure the sections, the text carries the meaning.

## Final run report

```markdown
## 🚀 om-approve-merge-pr — PR #{prNumber}: {title}

**Outcome:** {✅ merged (squash) | 🔁 queued for auto-merge once checks pass | ⛔ refused | ⚠️ stopped for confirmation} — {one full sentence: why this outcome — the merge went through cleanly, it was queued because required checks were still running or the user asked to merge on green, or which gate refused it}
**Approval:** {approving review submitted | self-approval rejected by the tracker — {what was done instead, per the user's answer}} — {one sentence of context}
**Gates:** {🏷️ label gates checked: {list} — all clear | ⛔ hard block: {`qa-failed` | `do-not-merge` | `blocked`} | 🧪 QA gate: `needs-qa` without `qa-approved` | skipped — labels disabled in config} — {one full sentence per gate that mattered: what it means and, when it blocked the merge, how to satisfy it}
**Follow-up:** {issue #{number} ({url}) filed and assigned to {assignee} | none requested} — {one sentence: what the follow-up tracks, or that no follow-up was asked for}

### {🚀 Merged | ⛔ Blocker and route}
{When merged or queued: confirm what landed (or will land on green) and note branch deletion only if the user asked for it. When refused: describe the blocker in full sentences and the route forward — failing required checks → offer `om-auto-fix-pr {prNumber} --ci-only`; conflicts, unresolved review feedback, or several blockers at once → offer `om-auto-fix-pr {prNumber}` (the full merge-ready loop, which hands back here); hard label blocks and the QA gate never route to automation — explain the human path (QA sign-off, the evidenced self-QA exception, or `skip-qa` where genuinely appropriate).}
```

End the report with the chaining reference lines on their own lines, exact
shape (the one part never decorated or reworded):

```text
PR: #<number> (link: <full PR URL>)
Issue: #<number> (link: <full issue URL>)   <- only when the run has a subject issue (e.g. the follow-up issue filed in step 5)
```
