# Report templates — final report (step 6)

How `om-check-and-commit` reports back to the user. Reporting style contract:
`references/rules.md` (Reporting style) — full sentences, explain the why,
never compress; emojis structure the sections, the text carries the meaning.
This skill defines no chaining reference lines.

## Final run report

```markdown
## ✅ om-check-and-commit — {branch}

**Result:** {✅ all gates green — committed and pushed | ✅ all gates green — verification only, publication was not requested | ❌ stopped — a required gate still fails} — {one full sentence on the outcome and why}

### 🧪 Validation gates

| Configured command | Result | Notes |
|---|---|---|
| `{command}` | {✅ pass | ❌ fail} | {Full sentence: passed unchanged, or what failed and what was done about it — including re-runs after fixes.} |

{One row per command in `validation.commands`, in the configured order. When a gate failed and was fixed, keep the row ✅ and explain the fix in Notes; a ❌ row means the run stopped on it.}

### 📋 Fixes applied
{Full sentences, one bullet per fix: what was broken, what you changed, and why the fix is minimal and in scope. Call out locale-file updates explicitly when the repo checks locales (which keys were added/removed, across which locales). When nothing needed fixing, say so in one sentence.}

### 🚀 Publication
{When pushed: a full sentence with the commit SHA, the conventional-commit subject, and the branch name. When not pushed: why — the user did not ask for publication, or which gate blocks it — and what to do next.}
```

When a required gate still fails, the report replaces the 🚀 section with a
⚠️ section describing the exact blocker — the failing command, the first
error, and what a human needs to decide — instead of committing.
