# Report templates — user-facing output (steps 9–10)

How `om-auto-update-changelog` reports back to the user. Reporting style
contract: `references/rules.md` (Reporting style) — full sentences, explain the
why, never compress; emojis structure the sections, the text carries the
meaning. These templates cover the **run report only**; the CHANGELOG entry
and line formats themselves are the product format and stay authoritative in
SKILL.md steps 5–6.

## Final run report (step 10)

```markdown
## 🚀 om-auto-update-changelog — {version} ({sinceDate} → {date})

**Result:** {✅ entry drafted and shipped as a docs PR via `om-auto-create-pr` | ⚠️ stopped — {reason}} — {one full sentence on the overall outcome}
**PRs consumed:** {count} — {one sentence: how the window was built (reachability from the release ref, or the documented fallback), how many PRs were enumerated before filtering, whether pagination was needed, and what was excluded (runs-only PRs, prior changelog PRs, branch-sync plumbing) and why}
**Supersede detections:** {count} — {one sentence naming each carried-forward PR (Paths A–C) and the original author it was credited to, or stating none were detected}
**Merge-capture corrections:** {count} — {one sentence naming each umbrella merge (Path D) and free-text hand-off (Path E), who the credit moved from and to, and which PRs were coalesced into a single bullet}
**Credit verification:** {count checked} — {one sentence: how many credits were compared against commit authorship, how many mismatches were reviewed by hand, what each resolved to, and any credit left marked unverified}
**Contributors:** {count} — {one sentence: how the credit list was built and that bot accounts and AI coding agents were skipped}

### 📝 CHANGELOG entry preview
{The first ~10 lines of the new block, quoted verbatim, so the reader sees the heading, the Highlights TODO marker, and the first section without opening the file.}

### 📋 What happens next
{Full sentences: which PR `om-auto-create-pr` opened or reused, that the Highlights paragraph is deliberately left as a TODO for a human author, and that the entry awaits maintainer review before merge.}
```

End the report with the chaining reference line on its own line, exact shape
(the one part never decorated, reworded, or wrapped in extra formatting) —
the PR number and URL come from the `om-auto-create-pr` output:

```text
PR: #<number> (link: <full PR URL>)
```

## Dry-run report (step 9)

With `--dry-run`, print the **full drafted entry** (not a preview — the whole
block that would be prepended to `CHANGELOG.md`), then this per-PR table so
the reader can audit every categorization and credit decision, then a closing
paragraph in full sentences summarizing what a real run would do (edit
`CHANGELOG.md`, invoke `om-auto-create-pr`) and confirming that nothing was
changed:

```markdown
| PR | Category | Line emoji | Primary author | Via | Path | Commits by credited author | Notes |
|----|----------|-----------|----------------|-----|------|---------------------------|-------|
| #1555 | fix | 🐛 | @contributor-a | @reviewer-b | A+B | 0 of 4 | supersedes #1421 — template present, verified ✅ |
| #1550 | fix | 🔧 | @reviewer-b | — | fallback | 6 of 6 | — |
| #1546 | fix | 🐛 | @contributor-c | — | fallback | 3 of 5 | 2 review-fixup commits by @reviewer-b, verified ✅ |
| #1820 | feat | ✨ | @contributor-d | — | D | 100 of 192 | umbrella merge by @maintainer-c (0 commits); coalesced with #1655 |
```

The `Path` and `Commits by credited author` columns are what make the credit
auditable — a reader must be able to see, per row, why the credited handle is
the right one. Never drop them to shorten the table.

A dry run emits **no** `PR:` chaining line — no PR exists.
