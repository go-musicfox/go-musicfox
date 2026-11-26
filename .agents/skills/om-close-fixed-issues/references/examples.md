# Worked examples

Illustrative examples for `om-close-fixed-issues` — a dry-run preview and the
three comment templates rendered with concrete values. The authoritative
templates live in `references/report-templates.md` (referenced from workflow
steps 4a/4b/4c and 7) and the dry-run behavior in step 5; these are here to
show the filled-in shape.

## Dry-run preview

```text
$ /om-close-fixed-issues --since 2026-04-01 --dry-run

Window: 2026-04-01 → 2026-04-17
Repo:   acme/widgets
Base branch: main

DRY-RUN: would close #1350 with link to PR #1421 (merged into main)
DRY-RUN: would comment on #1288 about PR #1419 (merged into release/0.5.0, not closing)
DRY-RUN: would comment on #1299 about PR #1412 (closed unmerged; superseded by #1415)
DRY-RUN: would skip #1270 — carries `do-not-close`
DRY-RUN: would skip #1260 — already closed

Summary: would-close 1, would-comment 2, would-skip 2.
```

## Unmatched-mentions section (non-English repository, no `closeKeywords` yet)

The shape a run prints when every PR in the window references issues in a
language the built-in keywords do not cover. Nothing was mutated; the section
is what turns an apparently quiet run into an actionable one:

```markdown
### ⚠️ Issue mentions without a recognized closing keyword

| PR | Mentions | Why it was skipped |
|----|----------|--------------------|
| #91 | #88 | The PR body says `Zamyka #88.`, which is not a recognized close keyword, and the tracker returned no `closingIssuesReferences`, so this run had no authority to close the issue. |
| #90 | #62 | The PR body says `Naprawia #62.`, which is not a recognized close keyword, so the issue stays open. |

The close-keyword vocabulary in effect for this run was the built-in English
list, with no `closeKeywords` configured in `.ai/agentic.config.json`. Both
the tracker's own parser and the built-in list recognize English only, so a
repository whose PR bodies are written in another language reports every PR
here until the local phrasing is added — for example
`"closeKeywords": ["zamyka", "naprawia"]`. Add the terms your team actually
writes, then re-run; nothing in this section was mutated.
```

With `"closeKeywords": ["zamyka", "naprawia"]` set, the same window produces
two ordinary `✅ closed` rows in the per-pair table and no unmatched-mentions
section at all.

## Close comment template (merged)

```markdown
✅ Fixed by #1421 (https://github.com/acme/widgets/pull/1421) — merged at 2026-04-15T14:02:31Z (commit `8a60110`).

Closed automatically by the `om-close-fixed-issues` skill. Credit to @alice (or the original author when the PR is a carry-forward — see the PR body for credit details).

If this is incorrect, reopen the issue and add the `do-not-close` label so future runs leave it alone.
```

## Informational comment template (closed unmerged + superseded)

```markdown
ℹ️ #1412 (https://github.com/acme/widgets/pull/1412) referenced this issue but was closed **without merging** on 2026-04-10T09:15:00Z. It was superseded by #1415 (https://github.com/acme/widgets/pull/1415).

This issue remains open. Posted automatically by the `om-close-fixed-issues` skill.
```

## Informational comment template (merged to non-base branch)

```markdown
ℹ️ #1419 (https://github.com/acme/widgets/pull/1419) references this issue and was merged into `release/0.5.0`, which is not the configured base branch (`main`). Leaving this issue open until the change lands on `main`.

Posted automatically by the `om-close-fixed-issues` skill.
```
