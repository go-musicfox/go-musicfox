# Report templates — tracking-issue body (step 9) and final report (step 10)

The authoritative tracking-issue body and the final run report for
`om-followup-issue-from-pr`. Reporting style contract: `references/rules.md`
(Reporting style) — full sentences, explain the why, never compress; the
glossary emojis structure the sections, the text carries the meaning.

## Tracking-issue body (design-doc mode, step 9)

The issue title stays exactly `Implement: <feature title>` — no emoji in the
title, so dedupe searches on `Implement:` keep matching. The body:

```markdown
## 📝 Design doc
- Document: `<path>`
- Design PR: <pr-url>

## 🎯 Summary
- 2–4 lines describing what the document proposes (from its overview/goal).

## 📋 How to implement
- Once the design PR merges, pick this issue up — for example with `om-auto-create-pr`, using the document as the brief.
- Do not start implementation until the design PR is merged into the configured base branch (`$BASE_BRANCH`).

Related: #<num>
```

## Final report (step 10)

Report in full sentences — not a one-liner — so a reader who did not watch
the run understands what was filed and why. For **each issue created**, cover:

- **What was filed and where:** the issue title and its full URL.
- **Which mode produced it:** a follow-up issue extracted from a review
  comment (comment mode) or an `Implement:` tracking issue for a design doc
  (design-doc mode).
- **What was extracted:** for comment mode, the actionable ask you distilled
  from the comment (and whose comment it was); for design-doc mode, the
  document path and the feature it proposes.
- **Who owns it and why:** the assignee, and whether they came from an
  @-mention or fell back to the PR author — plus a note when assignment
  failed so the user can fix it.
- **Labels applied,** or that labels were skipped (`labels.enabled: false` or
  no matching labels in the target repo).

When a tracking issue **already existed**, say so explicitly: name the
existing issue and its URL, state that no duplicate was created, and mention
the cross-link comment if one was added.

End the report with the chaining reference line on its own line, exact shape
(never decorated, reworded, or wrapped in extra formatting) — one line per
issue this run created:

```text
Issue: #<number> (link: <full issue URL>)
```
