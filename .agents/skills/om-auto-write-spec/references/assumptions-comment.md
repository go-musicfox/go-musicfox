# Assumptions comment (step 7)

The override-invitation comment posted when `om-spec-writing --autonomous` resolved Open Questions. Post on the PR via **comment-pr** (body file), and on the FR issue via **comment-issue** in issue-driven runs. Use the stable marker line verbatim so a re-run finds and **updates** the existing comment instead of posting a duplicate.

```markdown
🤖 `om-auto-write-spec` — Open Questions answered with autonomous defaults

This spec had open questions and the run is autonomous, so I applied conservative
defaults to keep moving. **Please review and override before merge if any is wrong.**

| # | Question | Applied default | Why | Confirm? |
|---|----------|-----------------|-----|----------|
| Q1 | {question} | {default} | {one-line rationale} | {ok / ⚠ needs human} |
| Q2 | … | … | … | … |

To change any answer: reply here (or edit the spec's "Resolved assumptions"
section) and re-run `om-auto-write-spec` — or implement with the corrected spec
via `om-auto-implement-spec {SPEC_PATH}`.
```

High-stakes guard: any `⚠ needs human` row ⇒ the PR stays a draft and the body states merge is gated on confirming those assumptions. Never `qa-approved` from this skill.
